package k8s

import (
	"context"
	"errors"
	"testing"
)

type scanFunc func(context.Context, string) (*Snapshot, error)

func (f scanFunc) Snapshot(ctx context.Context, c string) (*Snapshot, error) { return f(ctx, c) }
func TestReconcileFailurePreservesPreviousSnapshot(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	good := scanFunc(func(context.Context, string) (*Snapshot, error) { return &Snapshot{PodVersion: "first"}, nil })
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, good); err != nil {
		t.Fatal(err)
	}
	bad := scanFunc(func(context.Context, string) (*Snapshot, error) { return nil, errors.New("unavailable") })
	if index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, bad) == nil {
		t.Fatal("scan failure ignored")
	}
	s, ok := index.Current("cluster-a", "sit-a")
	if !ok || s.PodVersion != "first" {
		t.Fatal("scan failure removed previous inventory")
	}
	if !s.DiscoveryFailed {
		t.Fatal("failed discovery still marked healthy")
	}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, good); err != nil {
		t.Fatal(err)
	}
	s, _ = index.Current("cluster-a", "sit-a")
	if s.DiscoveryFailed {
		t.Fatal("successful relist did not restore discovery health")
	}
}
func TestDisabledNamespaceRejectsInflightScan(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	scan := scanFunc(func(context.Context, string) (*Snapshot, error) {
		index.Configure("cluster-a", "sit-a", 2, false)
		return &Snapshot{PodVersion: "stale"}, nil
	})
	if index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, scan) == nil {
		t.Fatal("old scan published after disable")
	}
	if _, ok := index.Current("cluster-a", "sit-a"); ok {
		t.Fatal("disabled namespace has inventory")
	}
	index.Configure("cluster-a", "sit-a", 1, true)
	if index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, scan) == nil {
		t.Fatal("old configuration reenabled namespace")
	}
}
func TestConcurrentReconcileCannotOverwriteNewerScan(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	slow := scanFunc(func(context.Context, string) (*Snapshot, error) {
		close(entered)
		<-release
		return &Snapshot{PodVersion: "old"}, nil
	})
	go func() { done <- index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, slow) }()
	<-entered
	fast := scanFunc(func(context.Context, string) (*Snapshot, error) { return &Snapshot{PodVersion: "new"}, nil })
	if index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, fast) == nil {
		t.Error("concurrent scan accepted")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, fast); err != nil {
		t.Fatal(err)
	}
	s, _ := index.Current("cluster-a", "sit-a")
	if s.PodVersion != "new" {
		t.Fatal("latest scan not published")
	}
}
