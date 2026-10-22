package k8s

import (
	"context"
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

func AlertRuntimeTiming(policy Policy) (time.Duration, time.Duration) {
	interval := policy.MetricsInterval()
	timeout := interval - time.Second
	if timeout < time.Second {
		timeout = time.Second
	}
	return interval, timeout
}

func RunAlerts(ctx context.Context, policy Policy, store AlertRuntimeStore, snapshots AlertSnapshotSource, alertStore AlertSink, outbox OutboxStore, sender NotificationSender, reader MetricReader, report func(error)) {
	if !policy.Enabled {
		return
	}
	if store == nil || snapshots == nil || alertStore == nil || outbox == nil || sender == nil || reader == nil {
		if report != nil {
			report(fmt.Errorf("alert runtime is not configured"))
		}
		return
	}
	dispatcher := AlertDispatcher{Store: outbox, Sender: sender, Owner: "monitor-alert", BatchSize: 32, Lease: time.Minute, MaxAttempts: 5}
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
