package k8s

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMetricQueryRatesBeforeAggregationAndRejectsInjection(t *testing.T) {
	q, err := MetricQuery("cluster-a", "sit-a", "cpu")
	if err != nil || !strings.Contains(q, "(rate(k8s_pod_cpu_usage_seconds_total{") || !strings.Contains(q, "[120s]))") {
		t.Fatalf("invalid CPU query %s %v", q, err)
	}
	q, err = MetricQuery("cluster-a", "sit-a", "memory")
	if err != nil || strings.Contains(q, "rate(") {
		t.Fatal("memory treated as counter")
	}
	if _, err = MetricQuery("cluster-a", "sit-a\"}", "cpu"); err == nil {
		t.Fatal("injection accepted")
	}
}

func TestMetricQueryWindowsFollowScrapeInterval(t *testing.T) {
	for _, seconds := range []int{30, 60, 300} {
		interval := time.Duration(seconds) * time.Second
		query, err := MetricQuery("cluster-a", "sit-a", "cpu", interval)
		if err != nil || !strings.Contains(query, "["+strconv.Itoa(seconds*4)+"s]") {
			t.Fatalf("rate window does not retain four scrape intervals: %s %v", query, err)
		}
		query, err = MetricQuery("cluster-a", "sit-a", "log_health", interval)
		if err != nil || !strings.Contains(query, "< "+strconv.Itoa(seconds*3)) {
			t.Fatalf("log freshness does not follow scrape interval: %s %v", query, err)
		}
		query, err = MetricQuery("cluster-a", "sit-a", "log_matches", interval)
		window := seconds * 4
		if window < 300 {
			window = 300
		}
		if err != nil || !strings.Contains(query, "["+strconv.Itoa(window)+"s]") {
			t.Fatalf("log count window cannot accommodate scrape interval: %s %v", query, err)
		}
		query, err = MetricQuery("cluster-a", "sit-a", "oom_events", interval)
		if err != nil || !strings.Contains(query, "["+strconv.Itoa(window)+"s]") {
			t.Fatalf("OOM count window cannot accommodate scrape interval: %s %v", query, err)
		}
	}
}

func TestMetricQuerySupportsNetworkAndBlockRates(t *testing.T) {
	for _, metric := range []string{"network_receive", "network_transmit", "block_read", "block_write"} {
		q, err := MetricQuery("cluster-a", "sit-a", metric)
		if err != nil || !strings.Contains(q, "rate(") || !strings.Contains(q, "[120s])") {
			t.Fatalf("%s: %s %v", metric, q, err)
		}
	}
}

func TestMetricQuerySupportsResourceHealthMetrics(t *testing.T) {
	cases := []string{"cpu_throttled", "cpu_throttled_periods", "cpu_request", "cpu_limit", "memory_request", "memory_limit", "restart_count", "ready", "phase", "oom_events", "memory_failcnt", "container_ready", "last_termination_oom"}
	for _, metric := range cases {
		if q, err := MetricQuery("cluster-a", "sit-a", metric); err != nil || q == "" {
			t.Fatalf("%s: %s %v", metric, q, err)
		}
	}
}

func TestMetricQueryPreservesPodUIDForAlertIdentity(t *testing.T) {
	q, err := MetricQuery("cluster-a", "sit-a", "cpu")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(q, "pod_uid") {
		t.Fatalf("cpu query must preserve pod_uid: %s", q)
	}
	q, err = MetricQuery("cluster-a", "sit-a", "ready")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(q, "pod_uid") || strings.Contains(q, "container) ") {
		t.Fatalf("ready query must be pod-scoped and preserve pod_uid: %s", q)
	}
}

func TestMetricQueryUsesRecentOOMIncrease(t *testing.T) {
	q, err := MetricQuery("cluster-a", "sit-a", "oom_events")
	if err != nil || !strings.Contains(q, "increase(k8s_pod_oom_events_total") || !strings.Contains(q, "[300s])") {
		t.Fatalf("OOM query must use a recent counter increase: %s", q)
	}
}

func TestOOMIncreaseRequiresCurrentSamples(t *testing.T) {
	q, err := MetricQuery("cluster-a", "sit-a", "oom_events")
	if err != nil || !strings.Contains(q, "and k8s_pod_oom_events_total") || !strings.Contains(q, "k8s_source_good") {
		t.Fatalf("old range samples can trigger OOM after source failure: %s", q)
	}
}
