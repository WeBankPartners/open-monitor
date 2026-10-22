package k8s

import (
	"strings"
	"testing"
)

func TestLogQueryUsesWindowAndHealthForCurrentObject(t *testing.T) {
	q, err := MetricQuery("cluster-a", "sit-a", "log_matches")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"increase(k8s_log_match_total", "[300s]", "pod_uid", "k8s_log_target_healthy", "up{job=\"k8s_log_agent\"}", "timestamp(", "cluster_id=\"cluster-a\"", "namespace=\"sit-a\""} {
		if !strings.Contains(q, required) {
			t.Fatalf("missing %s in %s", required, q)
		}
	}
	if err := ValidateAlertRule(AlertRule{ID: "log", Metric: "log_matches", Operator: ">"}); err != nil {
		t.Fatal(err)
	}
}
