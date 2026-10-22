package k8s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExporterServesOnlyFreshSamplesForEnabledScope(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-a", ContainerID: "runtime-a", Node: "node-a"}}}
	scanner := staticScanner{snapshot: snapshot}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, scanner); err != nil {
		t.Fatal(err)
	}
	exporter := NewExporter(index)
	exporter.Set("cluster-a", "sit-a", []MetricSample{{Name: "k8s_pod_memory_working_set_bytes", Value: 4, Labels: map[string]string{"namespace": "sit-a", "pod": "orders-0", "pod_uid": "pod-a", "container_id": "runtime-a"}}})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	exporter.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "k8s_pod_memory_working_set_bytes") {
		t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
	}
	index.Configure("cluster-a", "sit-a", 2, false)
	recorder = httptest.NewRecorder()
	exporter.ServeHTTP(recorder, req)
	if strings.Contains(recorder.Body.String(), "k8s_pod_memory_working_set_bytes") {
		t.Fatal("disabled scope still exposed")
	}
}

func TestExporterRejectsInvalidMetricSamples(t *testing.T) {
	exporter := NewExporter(NewIndex())
	if err := exporter.Set("cluster-a", "sit-a", []MetricSample{{Name: "bad name", Value: 1}}); err == nil {
		t.Fatal("invalid metric accepted")
	}
}

type staticScanner struct{ snapshot *Snapshot }

func (s staticScanner) Snapshot(context.Context, string) (*Snapshot, error) { return s.snapshot, nil }
