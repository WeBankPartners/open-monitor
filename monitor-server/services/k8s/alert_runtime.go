package k8s

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

type AlertSnapshotSource interface {
	Snapshot(context.Context, string, string) (*Snapshot, error)
}
type AlertRuleSource interface {
	ListAlertRules(context.Context, string, string) ([]AlertRule, error)
}
type AlertRuntimeStore interface {
	AlertRuleSource
	List(context.Context) ([]NamespaceState, error)
}

// LocalCollectionNotificationSender guards the actual notification entry point,
// including notifications without a strategy, before downstream side effects.
func LocalCollectionNotificationSender(policy Policy, next NotificationSender) NotificationSender {
	return NotificationSenderFunc(func(ctx context.Context, notification Notification) error {
		if !policy.LocalCollectionActive() {
			return errors.New("local collection disabled")
		}
		return next.Send(ctx, notification)
	})
}

type PendingOutboxCanceller interface {
	CancelPendingOutbox(context.Context, int) (int64, error)
}

// RunDisabledAlertMaintenance returns only after every pending event has been
// cancelled, or the context ends. Call it after database initialization even
// when the Kubernetes master switch is off. Each startup retries maintenance.
func RunDisabledAlertMaintenance(ctx context.Context, policy Policy, store PendingOutboxCanceller, report func(message string, err error)) error {
	if policy.LocalCollectionActive() {
		return nil
	}
	if store == nil {
		return errors.New("alert cancellation store is not configured")
	}
	var total int64
	delay := time.Second
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		count, err := cancelPendingAlertBatch(ctx, store)
		if err != nil {
			if report != nil {
				report(fmt.Sprintf("Kubernetes alert cancellation batch failed: total=%d retry_delay=%s", total, delay), err)
			}
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			continue
		}
		delay = time.Second
		total += count
		if report != nil {
			report(fmt.Sprintf("Kubernetes alert cancellation batch committed: cancelled=%d total=%d", count, total), nil)
		}
		if count == 0 {
			if report != nil {
				report(fmt.Sprintf("Kubernetes alert cancellation completed: total=%d", total), nil)
			}
			return nil
		}
	}
}

func cancelPendingAlertBatch(ctx context.Context, store PendingOutboxCanceller) (count int64, err error) {
	defer func() {
		if p := recover(); p != nil {
			if cause, ok := p.(error); ok {
				err = fmt.Errorf("alert cancellation worker panic: %w\n%s", cause, debug.Stack())
			} else {
				err = fmt.Errorf("alert cancellation worker panic (%T)\n%s", p, debug.Stack())
			}
		}
	}()
	batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	count, err = store.CancelPendingOutbox(batchCtx, 500)
	if err != nil {
		err = fmt.Errorf("cancel pending alert outbox batch: %w", err)
	}
	return count, err
}

func AlertRuntimeTiming(policy Policy) (time.Duration, time.Duration) {
	interval := policy.MetricsInterval()
	timeout := interval - time.Second
	if timeout < time.Second {
		timeout = time.Second
	}
	return interval, timeout
}

func RunAlerts(ctx context.Context, policy Policy, store AlertRuntimeStore, snapshots AlertSnapshotSource, alertStore AlertSink, outbox OutboxStore, sender NotificationSender, reader MetricReader, report func(error)) {
	if !policy.LocalCollectionActive() {
		return
	}
	if store == nil || snapshots == nil || alertStore == nil || outbox == nil || sender == nil || reader == nil {
		if report != nil {
			report(fmt.Errorf("alert runtime is not configured"))
		}
		return
	}
	dispatcher := AlertDispatcher{Store: outbox, Sender: LocalCollectionNotificationSender(policy, sender), Owner: "monitor-alert", BatchSize: 32, Lease: time.Minute, MaxAttempts: 5}
	interval, timeout := AlertRuntimeTiming(policy)
	// The sender's bounded SMTP operation can finish shortly after cancellation.
	// Keep claims valid through the dispatch budget and that final operation.
	dispatcher.Lease = timeout + 15*time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		func() {
			defer func() {
				if p := recover(); p != nil && report != nil {
					report(fmt.Errorf("alert worker panic (%T)\n%s", p, debug.Stack()))
				}
			}()
			cycleCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			rows, err := store.List(cycleCtx)
			if err == nil {
				for _, row := range rows {
					if !row.Enabled {
						continue
					}
					snapshot, scanErr := snapshots.Snapshot(cycleCtx, row.ClusterID, row.Namespace)
					if scanErr != nil || snapshot == nil {
						continue
					}
					rules, ruleErr := store.ListAlertRules(cycleCtx, row.ClusterID, row.Namespace)
					if ruleErr != nil {
						if report != nil {
							report(ruleErr)
						}
						continue
					}
					if evalErr := EvaluateSnapshot(cycleCtx, snapshot, rules, reader, alertStore, time.Now()); evalErr != nil {
						if report != nil {
							report(evalErr)
						}
					}
				}
			}
			if err != nil && report != nil {
				report(err)
			}
			// Previously persisted events are independent of this evaluation's
			// outcome and deadline, and must still be delivered after a failure.
			if !policy.LocalCollectionActive() {
				return
			}
			dispatchCtx, dispatchCancel := context.WithTimeout(ctx, timeout)
			defer dispatchCancel()
			if err = dispatcher.DispatchOnce(dispatchCtx); err != nil && report != nil {
				report(err)
			}
		}()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
