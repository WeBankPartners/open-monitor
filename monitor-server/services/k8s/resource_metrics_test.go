package k8s

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResourceQuantitiesRejectInvalidAndConvertUnits(t *testing.T) {
	for _, tc := range []struct {
		resource, raw string
		value         float64
		valid         bool
	}{
		{"cpu", "250m", .25, true}, {"cpu", "1000000n", .001, true}, {"memory", "64Mi", 67108864, true},
		{"memory", "1e6", 1000000, true}, {"memory", "1Gi", 1073741824, true}, {"memory", "1k", 1000, true},
		{"cpu", "NaN", 0, false}, {"memory", "-1", 0, false}, {"memory", "Inf", 0, false}, {"ephemeral-storage", "1Gi", 0, false},
	} {
		value, valid := quantityValue(tc.resource, tc.raw)
		if valid != tc.valid || valid && value != tc.value {
			t.Errorf("%s %s: %v %v", tc.resource, tc.raw, value, valid)
		}
	}
}

func TestExporterObjectMetricsSurviveCAdvisorFailure(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	s := metricSnapshot()
	s.Containers[0].Resources.Requests = map[string]string{"cpu": "250m"}
	s.Containers[0].Phase = "Running"
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{s}); err != nil {
		t.Fatal(err)
	}
	e := NewExporter(index)
	e.Fail("cluster-a", "sit-a")
	out := httptest.NewRecorder()
	e.ServeHTTP(out, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(out.Body.String(), "k8s_pod_cpu_request_cores{") || !strings.Contains(out.Body.String(), "} 0.25\n") {
		t.Fatal(out.Body.String())
	}
}

func TestObjectMetricsRemainVisibleUntilDiscoverySnapshotChanges(t *testing.T) {
	index := NewIndex()
	index.Configure("cluster-a", "sit-a", 1, true)
	s := metricSnapshot()
	s.ObservedAt = time.Now().Add(-10 * time.Minute)
	s.Containers[0].Resources.Requests = map[string]string{"cpu": "250m"}
	if err := index.Reconcile(context.Background(), "cluster-a", "sit-a", 1, staticScanner{s}); err != nil {
		t.Fatal(err)
	}
	e := NewExporter(index)
	e.Set("cluster-a", "sit-a", []MetricSample{{Name: "k8s_pod_cpu_usage_seconds_total", Value: 1, Labels: map[string]string{"pod_uid": "pod-a", "container_id": "containerd://runtime-a"}}})
	out := httptest.NewRecorder()
	e.ServeHTTP(out, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(out.Body.String(), "k8s_pod_cpu_request_cores") {
		t.Fatal(out.Body.String())
	}
}
