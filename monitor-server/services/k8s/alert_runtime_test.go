package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAlertRuntimeTimingFollowsMetricsInterval(t *testing.T) {
	interval, timeout := AlertRuntimeTiming(Policy{MetricsIntervalSeconds: 60})
	if interval != 60*time.Second || timeout != 59*time.Second {
		t.Fatalf("unexpected alert timing: interval=%s timeout=%s", interval, timeout)
	}
	interval, timeout = AlertRuntimeTiming(Policy{MetricsIntervalSeconds: 15})
	if interval != 15*time.Second || timeout != 14*time.Second {
		t.Fatalf("unexpected minimum alert timing: interval=%s timeout=%s", interval, timeout)
	}
}

type alertRuntimeStoreFake struct {
	listCalls int
	panicList bool
}

func (s *alertRuntimeStoreFake) List(context.Context) ([]NamespaceState, error) {
	s.listCalls++
	if s.panicList {
		panic("test worker panic")
	}
	return nil, nil
}

func (*alertRuntimeStoreFake) ListAlertRules(context.Context, string, string) ([]AlertRule, error) {
	return nil, nil
}

type alertRuntimeSnapshotFake struct{}

func (alertRuntimeSnapshotFake) Snapshot(context.Context, string, string) (*Snapshot, error) {
	return nil, errors.New("not expected")
}

type alertRuntimeSinkFake struct{}

func (alertRuntimeSinkFake) EvaluateAndPersist(context.Context, AlertEvaluation, time.Time) (Transition, error) {
	return Transition{}, nil
}

type alertRuntimeOutboxFake struct{}

func (alertRuntimeOutboxFake) ClaimOutbox(context.Context, string, int, time.Duration, time.Time) ([]OutboxEvent, error) {
	return nil, nil
}
func (alertRuntimeOutboxFake) MarkOutboxSent(context.Context, string, string, time.Time) error {
	return nil
}
func (alertRuntimeOutboxFake) RetryOutbox(context.Context, string, string, time.Time, error) error {
	return nil
}
func (alertRuntimeOutboxFake) FailOutbox(context.Context, string, string, time.Time, error) error {
	return nil
}

func TestRunAlertsDoesNotStartWhenDisabled(t *testing.T) {
	store := &alertRuntimeStoreFake{}
	RunAlerts(context.Background(), Policy{}, store, alertRuntimeSnapshotFake{}, alertRuntimeSinkFake{}, alertRuntimeOutboxFake{}, NotificationSenderFunc(func(context.Context, Notification) error { return nil }), MetricReadFunc(func(context.Context, AlertRule, Container) (*float64, string, bool) { return nil, "", false }), nil)
	if store.listCalls != 0 {
		t.Fatalf("disabled runtime performed work: %d", store.listCalls)
	}
}

func TestRunAlertsReportsWorkerPanicAndReturnsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	store := &alertRuntimeStoreFake{panicList: true}
	reported := make(chan error, 1)
	go RunAlerts(ctx, Policy{Enabled: true, LocalCollectionEnabled: true, MetricsIntervalSeconds: 15}, store, alertRuntimeSnapshotFake{}, alertRuntimeSinkFake{}, alertRuntimeOutboxFake{}, NotificationSenderFunc(func(context.Context, Notification) error { return nil }), MetricReadFunc(func(context.Context, AlertRule, Container) (*float64, string, bool) { return nil, "", false }), func(err error) { reported <- err })
	select {
	case err := <-reported:
		if err == nil {
			t.Fatal("panic was not reported")
		}
	case <-time.After(time.Second):
		cancel()
		t.Fatal("panic was not reported")
	}
	cancel()
}

type unavailableAlertRuntimeStore struct{ alertRuntimeStoreFake }

func (unavailableAlertRuntimeStore) List(context.Context) ([]NamespaceState, error) {
	return nil, errors.New("scope query failed")
}

type observedRuntimeOutbox struct {
	alertRuntimeOutboxFake
	called chan struct{}
}

func (s observedRuntimeOutbox) ClaimOutbox(ctx context.Context, _ string, _ int, _ time.Duration, _ time.Time) ([]OutboxEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	close(s.called)
	return nil, nil
}

func TestRunAlertsDispatchesPendingNotificationsWhenEvaluationFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	outbox := observedRuntimeOutbox{called: make(chan struct{})}
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunAlerts(ctx, Policy{Enabled: true, LocalCollectionEnabled: true}, &unavailableAlertRuntimeStore{}, alertRuntimeSnapshotFake{}, alertRuntimeSinkFake{}, outbox, NotificationSenderFunc(func(context.Context, Notification) error { return nil }), testMetricReader{}, func(error) {})
	}()
	select {
	case <-outbox.called:
	case <-time.After(time.Second):
		t.Error("evaluation failure blocked pending notifications")
	}
	cancel()
	<-done
}

type countingDisabledAlerts struct {
	alertRuntimeOutboxFake
	lists, rules, snapshots, evaluations, claims, sends int
}

func (s *countingDisabledAlerts) List(context.Context) ([]NamespaceState, error) {
	s.lists++
	return []NamespaceState{{Registration: Registration{Enabled: true}}}, nil
}
func (s *countingDisabledAlerts) ListAlertRules(context.Context, string, string) ([]AlertRule, error) {
	s.rules++
	return []AlertRule{{ID: "cpu", Enabled: true}}, nil
}
func (s *countingDisabledAlerts) Snapshot(context.Context, string, string) (*Snapshot, error) {
	s.snapshots++
	return &Snapshot{Containers: []Container{{PodUID: "pod-a"}}}, nil
}
func (s *countingDisabledAlerts) EvaluateAndPersist(context.Context, AlertEvaluation, time.Time) (Transition, error) {
	s.evaluations++
	return Transition{}, nil
}
func (s *countingDisabledAlerts) ClaimOutbox(context.Context, string, int, time.Duration, time.Time) ([]OutboxEvent, error) {
	s.claims++
	return []OutboxEvent{{EventID: "pending-event"}}, nil
}
func (s *countingDisabledAlerts) Send(context.Context, Notification) error {
	s.sends++
	return nil
}

func TestRunAlertsInactivePerformsNoEvaluationOrOutboxWork(t *testing.T) {
	for _, policy := range []Policy{{}, {Enabled: true}, {LocalCollectionEnabled: true}} {
		s := &countingDisabledAlerts{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		RunAlerts(ctx, policy, s, s, s, s, s, testMetricReader{}, nil)
		if s.lists+s.rules+s.snapshots+s.evaluations+s.claims+s.sends != 0 {
			t.Fatalf("inactive alerts performed work for policy %+v: %+v", policy, s)
		}
	}
}

func TestLocalCollectionNotificationSenderRejectsBeforeStrategyBranch(t *testing.T) {
	for _, policy := range []Policy{{}, {Enabled: true}, {LocalCollectionEnabled: true}} {
		for _, strategy := range []int{0, 7} {
			calls := 0
			sender := LocalCollectionNotificationSender(policy, NotificationSenderFunc(func(context.Context, Notification) error {
				calls++
				return nil
			}))
			err := sender.Send(context.Background(), Notification{StrategyID: strategy})
			if err == nil || !strings.Contains(err.Error(), "local collection disabled") || calls != 0 {
				t.Fatalf("inactive notification accepted: policy=%+v strategy=%d calls=%d err=%v", policy, strategy, calls, err)
			}
		}
	}
	upstream := errors.New("notification unavailable")
	sender := LocalCollectionNotificationSender(Policy{Enabled: true, LocalCollectionEnabled: true}, NotificationSenderFunc(func(context.Context, Notification) error { return upstream }))
	if err := sender.Send(context.Background(), Notification{}); !errors.Is(err, upstream) {
		t.Fatalf("active sender lost notification error: %v", err)
	}
}

type cancellationResult struct {
	count      int64
	err        error
	panicValue any
}

type cancellationRuntimeFake struct {
	results    []cancellationResult
	limits     []int
	afterBatch func()
}

func (s *cancellationRuntimeFake) CancelPendingOutbox(_ context.Context, limit int) (int64, error) {
	s.limits = append(s.limits, limit)
	if len(s.results) == 0 {
		panic("unexpected cancellation batch")
	}
	r := s.results[0]
	s.results = s.results[1:]
	if s.afterBatch != nil {
		s.afterBatch()
	}
	if r.panicValue != nil {
		panic(r.panicValue)
	}
	return r.count, r.err
}

func TestDisabledAlertMaintenanceDrainsBatchesOnEveryInactiveStartup(t *testing.T) {
	for _, policy := range []Policy{{}, {Enabled: true}, {LocalCollectionEnabled: true}} {
		s := &cancellationRuntimeFake{results: []cancellationResult{{count: 500}, {count: 1}, {}, {}}}
		var messages []string
		report := func(message string, err error) {
			if err != nil {
				t.Fatalf("unexpected maintenance error: %v", err)
			}
			messages = append(messages, message)
		}
		if err := RunDisabledAlertMaintenance(context.Background(), policy, s, report); err != nil {
			t.Fatal(err)
		}
		if len(s.limits) != 3 || !strings.Contains(messages[len(messages)-1], "total=501") {
			t.Fatalf("maintenance did not finish all batches: limits=%v messages=%v", s.limits, messages)
		}
		if err := RunDisabledAlertMaintenance(context.Background(), policy, s, report); err != nil || len(s.limits) != 4 {
			t.Fatalf("same-mode restart skipped maintenance: limits=%v err=%v", s.limits, err)
		}
		for _, limit := range s.limits {
			if limit <= 0 || limit > 500 {
				t.Fatalf("unbounded batch: %d", limit)
			}
		}
	}
}

func TestDisabledAlertMaintenanceActivePreservesState(t *testing.T) {
	s := &cancellationRuntimeFake{}
	if err := RunDisabledAlertMaintenance(context.Background(), Policy{Enabled: true, LocalCollectionEnabled: true}, s, nil); err != nil || len(s.limits) != 0 {
		t.Fatalf("active startup performed cancellation: limits=%v err=%v", s.limits, err)
	}
}

func TestDisabledAlertMaintenanceRetriesErrorsAndPanics(t *testing.T) {
	upstream := errors.New("batch commit failed")
	panicErr := errors.New("test panic")
	s := &cancellationRuntimeFake{results: []cancellationResult{{err: upstream}, {panicValue: panicErr}, {count: 2}, {}}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var failures []error
	if err := RunDisabledAlertMaintenance(ctx, Policy{}, s, func(_ string, err error) {
		if err != nil {
			failures = append(failures, err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if len(s.limits) != 4 || len(failures) != 2 || !errors.Is(failures[0], upstream) || !errors.Is(failures[1], panicErr) || !strings.Contains(failures[1].Error(), "goroutine") {
		t.Fatalf("retry or panic stack missing: batches=%v errors=%v", s.limits, failures)
	}
}

func TestDisabledAlertMaintenanceBackoffStopsOnCancellationAndCanRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	s := &cancellationRuntimeFake{results: []cancellationResult{{err: errors.New("database unavailable")}, {count: 1}, {}}}
	if err := RunDisabledAlertMaintenance(ctx, Policy{}, s, nil); !errors.Is(err, context.DeadlineExceeded) || len(s.limits) != 1 {
		t.Fatalf("failed maintenance did not back off: batches=%v err=%v", s.limits, err)
	}
	if err := RunDisabledAlertMaintenance(context.Background(), Policy{}, s, nil); err != nil || len(s.limits) != 3 {
		t.Fatalf("restart did not retry remaining pending events: batches=%v err=%v", s.limits, err)
	}
}

func TestDisabledAlertMaintenanceRestartsAfterPartialDrain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &cancellationRuntimeFake{results: []cancellationResult{{count: 500}, {count: 3}, {}}, afterBatch: cancel}
	if err := RunDisabledAlertMaintenance(ctx, Policy{}, s, nil); !errors.Is(err, context.Canceled) || len(s.limits) != 1 {
		t.Fatalf("partial drain did not stop: batches=%v err=%v", s.limits, err)
	}
	s.afterBatch = nil
	if err := RunDisabledAlertMaintenance(context.Background(), Policy{}, s, nil); err != nil || len(s.limits) != 3 {
		t.Fatalf("restart did not resume after committed batch: batches=%v err=%v", s.limits, err)
	}
}
