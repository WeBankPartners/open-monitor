package k8s

import (
	"context"
	"errors"
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
	go RunAlerts(ctx, Policy{Enabled: true, MetricsIntervalSeconds: 15}, store, alertRuntimeSnapshotFake{}, alertRuntimeSinkFake{}, alertRuntimeOutboxFake{}, NotificationSenderFunc(func(context.Context, Notification) error { return nil }), MetricReadFunc(func(context.Context, AlertRule, Container) (*float64, string, bool) { return nil, "", false }), func(err error) { reported <- err })
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
		RunAlerts(ctx, Policy{Enabled: true}, &unavailableAlertRuntimeStore{}, alertRuntimeSnapshotFake{}, alertRuntimeSinkFake{}, outbox, NotificationSenderFunc(func(context.Context, Notification) error { return nil }), testMetricReader{}, func(error) {})
	}()
	select {
	case <-outbox.called:
	case <-time.After(time.Second):
		t.Error("evaluation failure blocked pending notifications")
	}
	cancel()
	<-done
}
