package k8s

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type liveSource struct {
	mu      sync.Mutex
	calls   int
	expired bool
	notify  chan struct{}
}

func (s *liveSource) Snapshot(ctx context.Context, cluster string) (*Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return &Snapshot{PodVersion: "pod-version", WorkloadVersion: "sts-version"}, nil
}
func (s *liveSource) Watch(ctx context.Context, resource, version string, changed func()) error {
	s.mu.Lock()
	first := resource == "pods" && !s.expired
	if first {
		s.expired = true
	}
	s.mu.Unlock()
	if first {
		return &StatusError{Code: 410}
	}
	select {
	case s.notify <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return ctx.Err()
}
func TestWorkerRelistsAfterExpiredWatchAndStops(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	source := &liveSource{notify: make(chan struct{}, 10)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker := Worker{Index: index, Cluster: "cluster-a", Namespace: "sit-a", Version: 1, Source: source, RetryDelay: time.Millisecond, ResyncInterval: time.Hour}
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-source.notify:
			source.mu.Lock()
			calls := source.calls
			source.mu.Unlock()
			if calls >= 2 {
				cancel()
				if err := <-done; !errors.Is(err, context.Canceled) {
					t.Fatalf("unexpected stop: %v", err)
				}
				return
			}
		case <-deadline:
			t.Fatal("watch expiration did not trigger relist")
		}
	}
}
func TestIndexCopiesPublishedData(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	original := &Snapshot{Containers: []Container{{Resources: Resources{Requests: map[string]string{"cpu": "100m"}}}}}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, scanFunc(func(context.Context, string) (*Snapshot, error) { return original, nil })); err != nil {
		t.Fatal(err)
	}
	original.Containers[0].Resources.Requests["cpu"] = "200m"
	first, _ := index.Current("cluster-a", "sit-a")
	first.Containers[0].Resources.Requests["cpu"] = "300m"
	second, _ := index.Current("cluster-a", "sit-a")
	if second.Containers[0].Resources.Requests["cpu"] != "100m" {
		t.Fatal("caller mutated inventory")
	}
}
