package k8s

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Notification struct {
	EventID    string
	Event      AlertEvent
	Key        AlertKey
	Payload    map[string]any
	StrategyID int
}

type NotificationSender interface {
	Send(context.Context, Notification) error
}
type NotificationSenderFunc func(context.Context, Notification) error

func (f NotificationSenderFunc) Send(ctx context.Context, notification Notification) error {
	return f(ctx, notification)
}

type OutboxStore interface {
	ClaimOutbox(context.Context, string, int, time.Duration, time.Time) ([]OutboxEvent, error)
	MarkOutboxSent(context.Context, string, string, time.Time) error
	RetryOutbox(context.Context, string, string, time.Time, error) error
	FailOutbox(context.Context, string, string, time.Time, error) error
}

type AlertDispatcher struct {
	Store       OutboxStore
	Sender      NotificationSender
	Owner       string
	BatchSize   int
	Lease       time.Duration
	RetryBase   time.Duration
	RetryMax    time.Duration
	MaxAttempts uint
	Now         func() time.Time
}

func (d AlertDispatcher) DispatchOnce(ctx context.Context) error {
	if d.Store == nil || d.Sender == nil || d.Owner == "" {
		return errors.New("alert dispatcher is not configured")
	}
	limit := d.BatchSize
	if limit <= 0 {
		limit = 32
	}
	lease := d.Lease
	if lease <= 0 {
		lease = time.Minute
	}
	now := time.Now()
	if d.Now != nil {
		now = d.Now()
	}
	events, err := d.Store.ClaimOutbox(ctx, d.Owner, limit, lease, now)
	if err != nil {
		return err
	}
	var first error
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		strategyID := 0
		if raw, ok := event.Payload["notify_strategy_id"].(float64); ok {
			strategyID = int(raw)
		}
		err = d.Sender.Send(ctx, Notification{EventID: event.EventID, Event: event.Event, Key: event.Key, Payload: event.Payload, StrategyID: strategyID})
		if err == nil {
			err = d.Store.MarkOutboxSent(ctx, event.EventID, event.LeaseToken, now)
		} else {
			maxAttempts := d.MaxAttempts
			if maxAttempts == 0 {
				maxAttempts = 5
			}
			var retryErr error
			if event.Attempts >= maxAttempts {
				retryErr = d.Store.FailOutbox(ctx, event.EventID, event.LeaseToken, now, err)
			} else {
				retryAt := now.Add(d.backoff(event.Attempts))
				retryErr = d.Store.RetryOutbox(ctx, event.EventID, event.LeaseToken, retryAt, err)
			}
			if retryErr != nil {
				err = fmt.Errorf("notification failed: %w; retry state failed: %v", err, retryErr)
			}
		}
		if err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (d AlertDispatcher) backoff(attempt uint) time.Duration {
	base := d.RetryBase
	if base <= 0 {
		base = time.Second
	}
	max := d.RetryMax
	if max <= 0 {
		max = 5 * time.Minute
	}
	delay := base
	for i := uint(1); i < attempt && delay < max; i++ {
		delay *= 2
	}
	if delay > max {
		return max
	}
	return delay
}
