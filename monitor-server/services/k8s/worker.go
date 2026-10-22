package k8s

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

type DiscoverySource interface {
	Scanner
	Watch(context.Context, string, string, func()) error
}
type Worker struct {
	Index                      *Index
	Source                     DiscoverySource
	Cluster, Namespace         string
	Version                    uint64
	RetryDelay, ResyncInterval time.Duration
	Report                     func(error)
}

// Run reconciles before watching and relists after changes, expiration or periodic resync.
// Its supervisor must cancel and join it when the namespace version or task ownership changes.
func (w Worker) Run(ctx context.Context) error {
	if w.Index == nil || w.Source == nil {
		return errors.New("discovery worker dependencies are required")
	}
	retry, resync := w.RetryDelay, w.ResyncInterval
	if retry <= 0 {
		retry = time.Second
	}
	if resync <= 0 {
		resync = 5 * time.Minute
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := w.cycle(ctx, resync)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			w.Index.markDiscoveryFailed(w.Cluster, w.Namespace, w.Version)
		}
		if err != nil && w.Report != nil {
			w.Report(err)
		}
		timer := time.NewTimer(retry)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
func (w Worker) cycle(ctx context.Context, resync time.Duration) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("discovery worker panic (%T)\n%s", p, debug.Stack())
		}
	}()
	if err = w.Index.Reconcile(ctx, w.Cluster, w.Namespace, w.Version, w.Source); err != nil {
		return err
	}
	snapshot, ok := w.Index.Current(w.Cluster, w.Namespace)
	if !ok {
		return errors.New("namespace inventory unavailable")
	}
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	changed := make(chan struct{}, 1)
	done := make(chan error, 2)
	notify := func() {
		select {
		case changed <- struct{}{}:
		default:
		}
	}
	watch := func(resource, version string) {
		defer func() {
			if p := recover(); p != nil {
				done <- fmt.Errorf("Kubernetes watch panic (%T)\n%s", p, debug.Stack())
			}
		}()
		done <- w.Source.Watch(watchCtx, resource, version, notify)
	}
	go watch("pods", snapshot.PodVersion)
	go watch("statefulsets", snapshot.WorkloadVersion)
	timer := time.NewTimer(resync)
	defer timer.Stop()
	finished := 0
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-changed:
	case err = <-done:
		finished++
	case <-timer.C:
	}
	cancel()
	for finished < 2 {
		<-done
		finished++
	}
	return err
}
