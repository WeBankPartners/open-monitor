package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMetricServiceUsesIndependentClusterCredentialAndDeduplicatesNodes(t *testing.T) {
	index := NewIndex()
	exporter := NewExporter(index)
	a := metricSnapshot()
	index.Configure("cluster-a", "sit-a", 1, true)
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{a}); err != nil {
		t.Fatal(err)
	}
	b := metricSnapshot()
	b.Containers[0].Key.Namespace = "sit-b"
	index.Configure("cluster-a", "sit-b", 1, true)
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-b", 1, staticScanner{b}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	registered := []NamespaceState{{Registration: Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true, Version: 1}}, {Registration: Registration{ClusterID: "cluster-a", Namespace: "sit-b", Enabled: true, Version: 1}}}
	svc := MetricService{Index: index, Exporter: exporter, Factory: func(r Registration) (MetricSource, func(), error) {
		if r.Credentials.Guid != "cluster-credential" {
			t.Fatal("namespace credential used")
		}
		return metricSourceFunc(func(context.Context, string) (string, error) {
			calls++
			return strings.ReplaceAll(metricPayload(), "namespace=\"sit-a\"", "namespace=\"sit-a\"") + "container_memory_working_set_bytes{namespace=\"sit-b\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 4\n", nil
		}), func() {}, nil
	}}
	if err := svc.Cycle(context.Background(), registered, []Registration{{ClusterID: "cluster-a", Enabled: true, Version: 1, Credentials: Credentials{Guid: "cluster-credential"}}}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("node fetched %d times", calls)
	}
	if len(exporter.samples[scope{"cluster-a", "sit-a"}]) != 2 {
		t.Fatal("samples not published")
	}
	_ = time.Second
}

func TestRunMetricsDisabledDoesNotAccessDependencies(t *testing.T) {
	for _, policy := range []Policy{{}, {Enabled: true}, {LocalCollectionEnabled: true}} {
		// Nil dependencies are deliberate: neither database nor Kubernetes may be
		// accessed when either deployment switch disables local collection.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		done := make(chan struct{})
		go func() {
			defer close(done)
			RunMetrics(ctx, policy, Store{}, nil, nil, func(err error) { t.Errorf("disabled worker: %v", err) })
		}()
		select {
		case <-done:
		case <-ctx.Done():
			t.Error("disabled worker did not return")
		}
		cancel()
	}
}

func TestRunMetricsEnabledLoadsScopes(t *testing.T) {
	store, _ := newLogStoreTest(t, []logStoreStep{{match: "FROM k8s_namespace_source", err: errors.New("test database unavailable")}})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	index := NewIndex()
	reported := false
	RunMetrics(ctx, Policy{Enabled: true, LocalCollectionEnabled: true}, store, index, NewExporter(index), func(err error) {
		reported = true
		cancel()
	})
	if !reported {
		t.Fatal("enabled worker did not load scopes")
	}
}
