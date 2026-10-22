package k8s

import (
	"context"
	"testing"
	"time"
)

func TestMetricRuntimePublishesSamplesForCurrentSnapshot(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-a", ContainerID: "runtime-a", Node: "node-a"}}}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{snapshot: snapshot}); err != nil {
		t.Fatal(err)
	}
	exporter := NewExporter(index)
	source := metricSourceFunc(func(context.Context, string) (string, error) {
		return "container_memory_working_set_bytes{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"runtime-a\"} 4\n", nil
	})
	runtime := MetricRuntime{Index: index, Collector: Collector{Source: source}, Exporter: exporter, Freshness: time.Minute}
	if err := runtime.Collect(context.Background(), "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	if got := len(exporter.samples[scope{"cluster-a", "sit-a"}]); got != 1 {
		t.Fatalf("got %d samples", got)
	}
}

func TestMetricRuntimeRejectsMissingSnapshot(t *testing.T) {
	runtime := MetricRuntime{Index: NewIndex(), Collector: Collector{Source: metricSourceFunc(func(context.Context, string) (string, error) { return "", nil })}, Exporter: NewExporter(NewIndex())}
	if err := runtime.Collect(context.Background(), "cluster-a", "sit-a"); err == nil {
		t.Fatal("missing snapshot accepted")
	}
}

func TestMetricRuntimeDoesNotTreatObjectMetricsAsSuccessfulCAdvisorCollection(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	s := metricSnapshot()
	s.Containers[0].Resources.Requests = map[string]string{"cpu": "250m"}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{s}); err != nil {
		t.Fatal(err)
	}
	e := NewExporter(index)
	r := MetricRuntime{Index: index, Exporter: e, Collector: Collector{Source: metricSourceFunc(func(context.Context, string) (string, error) { return "", nil })}}
	if err := r.Collect(context.Background(), "cluster-a", "sit-a"); err == nil {
		t.Fatal("empty cAdvisor collection accepted")
	}
}
