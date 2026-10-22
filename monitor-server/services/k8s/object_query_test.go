package k8s

import (
	"strings"
	"testing"
	"time"
)

func TestApplicationMetricHistoryUsesWhitelistAndFreshTargets(t *testing.T) {
	for _, tc := range []struct{ metric, fragment string }{
		{"app_jvm_memory_used_bytes", "k8s_app_jvm_memory_used_bytes{"},
		{"app_executor_completed_total", "rate(k8s_app_executor_completed_total{"},
		{"app_http_server_requests_seconds_rate", "rate(k8s_app_http_server_requests_seconds_count{"},
		{"app_http_server_requests_seconds_mean", "rate(k8s_app_http_server_requests_seconds_sum{"},
	} {
		query, err := ObjectMetricQuery("cluster-a", "sit-a", tc.metric, MetricObjectFilter{}, 60*time.Second)
		if err != nil || !strings.Contains(query, tc.fragment) || !strings.Contains(query, "k8s_app_metric_available") || !strings.Contains(query, `up{job="k8s_log_agent"}`) {
			t.Fatalf("application history %s: %s %v", tc.metric, query, err)
		}
		if strings.Contains(query, "rate(") && !strings.Contains(query, "[240s]") {
			t.Fatal("application rate window does not follow scrape interval")
		}
	}
	for _, metric := range []string{"app_arbitrary", "app_http_server_requests_seconds_bucket", `app_jvm_memory_used_bytes} or up{`} {
		if _, err := ObjectMetricQuery("cluster-a", "sit-a", metric, MetricObjectFilter{}, 30*time.Second); err == nil {
			t.Fatalf("unapproved application query accepted: %s", metric)
		}
	}
}

func TestObjectMetricQueryScopesSourcesWithoutLosingPodMetricsOrHistory(t *testing.T) {
	object := MetricObjectFilter{WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "main"}
	for _, metric := range []string{"cpu", "oom_events", "log_matches", "app_jvm_memory_used_bytes"} {
		query, err := ObjectMetricQuery("cluster-a", "sit-a", metric, object, 30*time.Second)
		if err != nil || !strings.Contains(query, `pod="app-0",container="main"`) || !strings.Contains(query, `workload_name="app"`) || strings.Contains(query, `pod_uid=`) {
			t.Fatalf("object scope missing or tied to a single incarnation: %s %v", query, err)
		}
		if metric == "oom_events" && !strings.Contains(query, `k8s_source_good{cluster_id="cluster-a",namespace="sit-a"}`) {
			t.Fatalf("namespace quality incorrectly filtered by pod: %s", query)
		}
	}
	for _, metric := range []string{"network_receive", "network_transmit", "ready", "phase", "quality"} {
		query, err := ObjectMetricQuery("cluster-a", "sit-a", metric, object, 30*time.Second)
		if err != nil || strings.Contains(query, `container="main"`) {
			t.Fatalf("pod or namespace metric incorrectly filtered by container: %s %v", query, err)
		}
	}
	for _, bad := range []MetricObjectFilter{{PodName: `x"} or up{`}, {Container: "x\n"}, {WorkloadName: strings.Repeat("a", 254)}} {
		if _, err := ObjectMetricQuery("cluster-a", "sit-a", "cpu", bad, 30*time.Second); err == nil {
			t.Fatalf("unsafe object accepted: %+v", bad)
		}
	}
}

func TestMetricCatalogContainsUniqueQueryableMetrics(t *testing.T) {
	seen := map[string]bool{}
	for _, metric := range MetricCatalog() {
		if seen[metric.ID] || metric.Unit == "" || metric.Scope == "" {
			t.Fatalf("invalid catalog entry: %+v", metric)
		}
		seen[metric.ID] = true
		if _, err := ObjectMetricQuery("cluster-a", "sit-a", metric.ID, MetricObjectFilter{}, 30*time.Second); err != nil {
			t.Fatalf("catalog entry cannot be queried: %s %v", metric.ID, err)
		}
	}
}

func TestApplicationHistoryDoesNotEnableAmbiguousThresholdAlerts(t *testing.T) {
	rule := AlertRule{ID: "memory", Metric: "app_jvm_memory_used_bytes", Operator: ">", Threshold: 1}
	if err := ValidateAlertRule(rule); err == nil {
		t.Fatal("application dimensions must not be evaluated as a single arbitrary threshold value")
	}
}
