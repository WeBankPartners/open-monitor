package k8s

import (
	"context"
	"errors"
	"testing"
	"time"
)

type dispatcherStore struct {
	events                []OutboxEvent
	sent, retried, failed int
	retryAt               time.Time
}

func (s *dispatcherStore) ClaimOutbox(context.Context, string, int, time.Duration, time.Time) ([]OutboxEvent, error) {
	return s.events, nil
}
func (s *dispatcherStore) MarkOutboxSent(context.Context, string, string, time.Time) error {
	s.sent++
	return nil
}
func (s *dispatcherStore) RetryOutbox(_ context.Context, _ string, _ string, at time.Time, _ error) error {
	s.retried++
	s.retryAt = at
	return nil
}
func (s *dispatcherStore) FailOutbox(context.Context, string, string, time.Time, error) error {
	s.failed++
	return nil
}

type dispatcherSender struct {
	err   error
	calls int
}

func (s *dispatcherSender) Send(context.Context, Notification) error { s.calls++; return s.err }

func TestAlertDispatcherSendsAndMarksOutbox(t *testing.T) {
	store := &dispatcherStore{events: []OutboxEvent{{EventID: "e", LeaseToken: "l", Event: AlertEventFiring}}}
	sender := &dispatcherSender{}
	if err := (AlertDispatcher{Store: store, Sender: sender, Owner: "o", Now: func() time.Time { return time.Unix(1, 0) }}).DispatchOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.calls != 1 || store.sent != 1 || store.retried != 0 {
		t.Fatalf("unexpected delivery: %+v calls=%d", store, sender.calls)
	}
}

func TestAlertDispatcherRetriesFailure(t *testing.T) {
	store := &dispatcherStore{events: []OutboxEvent{{EventID: "e", LeaseToken: "l", Attempts: 2}}}
	sender := &dispatcherSender{err: errors.New("temporary")}
	if err := (AlertDispatcher{Store: store, Sender: sender, Owner: "o", RetryBase: time.Second, Now: func() time.Time { return time.Unix(1, 0) }}).DispatchOnce(context.Background()); err == nil {
		t.Fatal("expected send failure")
	}
	if store.sent != 0 || store.retried != 1 || !store.retryAt.Equal(time.Unix(3, 0)) {
		t.Fatalf("unexpected retry: %+v", store)
	}
}

func TestAlertDispatcherMarksPermanentFailureAfterMaximumAttempts(t *testing.T) {
	store := &dispatcherStore{events: []OutboxEvent{{EventID: "e", LeaseToken: "l", Attempts: 5}}}
	sender := &dispatcherSender{err: errors.New("permanent")}
	if err := (AlertDispatcher{Store: store, Sender: sender, Owner: "o", MaxAttempts: 5, Now: func() time.Time { return time.Unix(1, 0) }}).DispatchOnce(context.Background()); err == nil {
		t.Fatal("expected send failure")
	}
	if store.failed != 1 || store.retried != 0 {
		t.Fatalf("expected permanent failure state: %+v", store)
	}
}

func TestAlertDispatcherStopsSendingAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := &dispatcherStore{events: []OutboxEvent{{EventID: "first"}, {EventID: "second"}}}
	calls := 0
	sender := NotificationSenderFunc(func(context.Context, Notification) error { calls++; cancel(); return nil })
	err := (AlertDispatcher{Store: store, Sender: sender, Owner: "worker"}).DispatchOnce(ctx)
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("sent after cancellation: calls=%d err=%v", calls, err)
	}
}
