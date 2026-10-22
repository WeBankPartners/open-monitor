package k8s

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRuntimeCollectionReachesHTTPAndFailureRemovesOldValues(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	snapshot := metricSnapshot()
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{snapshot}); err != nil {
		t.Fatal(err)
	}
	exporter := NewExporter(index)
	fail := false
	source := metricSourceFunc(func(context.Context, string) (string, error) {
		if fail {
			return "", errors.New("unavailable")
		}
		return metricPayload(), nil
	})
	runtime := MetricRuntime{Index: index, Collector: Collector{Source: source}, Exporter: exporter}
	if err := runtime.Collect(context.Background(), "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	out := httptest.NewRecorder()
	exporter.ServeHTTP(out, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(out.Body.String(), "k8s_pod_memory_working_set_bytes{") {
		t.Fatalf("missing data: %s", out.Body.String())
	}
	fail = true
	for n := 0; n < 3; n++ {
		if err := runtime.Collect(context.Background(), "cluster-a", "sit-a"); err == nil {
			t.Fatal("failure hidden")
		}
	}
	out = httptest.NewRecorder()
	exporter.ServeHTTP(out, httptest.NewRequest("GET", "/metrics", nil))
	if strings.Contains(out.Body.String(), "k8s_pod_memory_working_set_bytes{") || !strings.Contains(out.Body.String(), "k8s_source_failed{") || !strings.Contains(out.Body.String(), "} 3\n") {
		t.Fatalf("stale data or missing failure status: %s", out.Body.String())
	}
}
func metricSnapshot() *Snapshot {
	return &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-a", ContainerID: "containerd://runtime-a", Node: "node-a"}}}
}
func metricPayload() string {
	return "# TYPE container_cpu_usage_seconds_total counter\ncontainer_cpu_usage_seconds_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",cpu=\"total\",id=\"/kubepods/runtime-a\"} 12\n# TYPE container_memory_working_set_bytes gauge\ncontainer_memory_working_set_bytes{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 4\n"
}
func TestExporterExpiresAndRejectsChangedRuntime(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	s := metricSnapshot()
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{s}); err != nil {
		t.Fatal(err)
	}
	exporter := NewExporter(index)
	now := time.Now()
	exporter.now = func() time.Time { return now }
	r := MetricRuntime{Index: index, Exporter: exporter, Collector: Collector{Source: metricSourceFunc(func(context.Context, string) (string, error) { return metricPayload(), nil })}}
	if err := r.Collect(context.Background(), "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	out := httptest.NewRecorder()
	exporter.ServeHTTP(out, httptest.NewRequest("GET", "/metrics", nil))
	if strings.Contains(out.Body.String(), "k8s_pod_memory_working_set_bytes{") {
		t.Fatal("expired cache exposed")
	}
	now = time.Now()
	if err := r.Collect(context.Background(), "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	replacement := metricSnapshot()
	replacement.Containers[0].PodUID = "replacement"
	replacement.Containers[0].ContainerID = "containerd://replacement"
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{replacement}); err != nil {
		t.Fatal(err)
	}
	out = httptest.NewRecorder()
	exporter.ServeHTTP(out, httptest.NewRequest("GET", "/metrics", nil))
	if strings.Contains(out.Body.String(), "k8s_pod_memory_working_set_bytes{") {
		t.Fatal("old runtime exposed")
	}
}

func TestFailedDiscoveryStopsExportingOldObjectStateAndMarksAlertsUnknown(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{metricSnapshot()}); err != nil {
		t.Fatal(err)
	}
	exporter := NewExporter(index)
	runtime := MetricRuntime{Index: index, Exporter: exporter, Collector: Collector{Source: metricSourceFunc(func(context.Context, string) (string, error) { return metricPayload(), nil })}}
	if err := runtime.Collect(context.Background(), "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	_ = index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, scanFunc(func(context.Context, string) (*Snapshot, error) { return nil, errors.New("API unavailable") }))
	output := httptest.NewRecorder()
	exporter.ServeHTTP(output, httptest.NewRequest("GET", "/metrics", nil))
	if strings.Contains(output.Body.String(), "k8s_pod_ready{") || strings.Contains(output.Body.String(), "k8s_pod_memory_working_set_bytes{") || !strings.Contains(output.Body.String(), "k8s_source_good{cluster_id=\"cluster-a\",namespace=\"sit-a\"} 0") {
		t.Fatalf("stale state exported: %s", output.Body.String())
	}
	snapshot, _ := index.Current("cluster-a", "sit-a")
	sink := &testAlertSink{}
	reader := MetricReadFunc(func(context.Context, AlertRule, Container) (*float64, string, bool) {
		t.Fatal("queried metrics for failed inventory")
		return nil, "", false
	})
	if err := EvaluateSnapshot(context.Background(), snapshot, []AlertRule{{ID: "ready", Metric: "ready", Enabled: true}}, reader, sink, time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(sink.inputs) != 1 || sink.inputs[0].Observation.Valid {
		t.Fatalf("failed inventory did not become unknown: %+v", sink.inputs)
	}
}
