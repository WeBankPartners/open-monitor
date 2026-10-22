package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func metricLine(metrics string) string {
	return `[2026-10-20 10:00:00.000]{"resourceMetrics":[{"scopeMetrics":[{"metrics":[` + metrics + `]}]}]}`
}
func TestOTLPNormalizesUnitsAndNonMonotonicSum(t *testing.T) {
	line := metricLine(`{"name":"jvm.memory.used","unit":"By","sum":{"aggregationTemporality":2,"dataPoints":[{"asInt":"123","attributes":[{"key":"jvm.memory.pool.name","value":{"stringValue":"heap-pool"}},{"key":"host.id","value":{"stringValue":"private-host"}}]}]}},{"name":"process.cpu.time","unit":"ns","sum":{"isMonotonic":true,"aggregationTemporality":"AGGREGATION_TEMPORALITY_CUMULATIVE","dataPoints":[{"asDouble":2000000000}]}}`)
	samples, err := parseOTLPMetricLine(line)
	if err != nil || len(samples) != 2 {
		t.Fatalf("%+v %v", samples, err)
	}
	if samples[0].Name != "jvm_memory_used_bytes" || samples[0].Value != 123 || samples[0].Labels["pool"] != "heap-pool" || len(samples[0].Labels) != 2 {
		t.Fatalf("bad memory: %+v", samples[0])
	}
	if samples[1].Name != "process_cpu_seconds_total" || samples[1].Value != 2 {
		t.Fatalf("bad CPU: %+v", samples[1])
	}
}
func TestOTLPHistogramConvertsBucketsAndDoesNotInventQuantiles(t *testing.T) {
	samples, err := parseOTLPMetricLine(metricLine(`{"name":"jvm.gc.duration","unit":"s","histogram":{"aggregationTemporality":2,"dataPoints":[{"count":"3","sum":0.5,"bucketCounts":["1","2","0"],"explicitBounds":[0.1,1]}]}}`))
	if err != nil || len(samples) != 5 {
		t.Fatalf("%+v %v", samples, err)
	}
	found := false
	for _, s := range samples {
		if s.Labels["le"] == "1" {
			found = s.Value == 3
		}
	}
	if !found {
		t.Fatal("OTLP per-bucket counts not accumulated")
	}
	samples, err = parseOTLPMetricLine(metricLine(`{"name":"http.server.requests","unit":"milliseconds","histogram":{"aggregationTemporality":2,"dataPoints":[{"count":"2","sum":500}]}}`))
	if err != nil || len(samples) != 2 || samples[1].Value != 0.5 {
		t.Fatalf("count/sum conversion failed: %+v %v", samples, err)
	}
}
func TestOTLPRejectsMalformedValuesAndSkipsUnapprovedMetrics(t *testing.T) {
	for _, data := range []string{`{"name":"jvm.memory.used","unit":"bytes","gauge":{"dataPoints":[{"asDouble":"NaN"}]}}`, `{"name":"jvm.memory.used","unit":"bytes","gauge":{"dataPoints":[{}]}}`} {
		if samples, err := parseOTLPMetricLine(metricLine(data)); err == nil && len(samples) != 0 {
			t.Fatal("invalid value emitted")
		}
	}
	samples, err := parseOTLPMetricLine(metricLine(`{"name":"spring.security.filterchains","gauge":{"dataPoints":[{"asDouble":1}]}}`))
	if err != nil || len(samples) != 0 {
		t.Fatal("non-whitelist metric emitted")
	}
}
func TestMetricFileScanPublishesLatestAndExpires(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "logs"), 0700)
	a, err := New(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	target := Target{ID: "metric", Namespace: "sit", PodUID: "pod", PodName: "app-0", InstanceID: "guid", Directory: "logs", Glob: "*_metric.log", Start: "end", Format: "otlp_metric"}
	if err = a.Apply(Config{Version: 1, Targets: []Target{target}}); err != nil {
		t.Fatal(err)
	}
	write := func(stamp int64, value int) {
		line := metricLine(fmt.Sprintf(`{"name":"jvm.memory.used","unit":"bytes","gauge":{"dataPoints":[{"timeUnixNano":"%d","asInt":"%d"}]}}`, stamp, value))
		if err := os.WriteFile(filepath.Join(root, "logs", "app_metric.log"), []byte(line+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(time.Now().UnixNano(), 123)
	for i := 0; i < 2; i++ {
		if err = a.Scan(); err != nil {
			t.Fatal(err)
		}
		text := a.MetricsText("cluster", "node")
		if !strings.Contains(text, "k8s_app_jvm_memory_used_bytes{") || !strings.Contains(text, "} 123\n") || !strings.Contains(text, `instance_id="guid"`) {
			t.Fatal(text)
		}
	}
	write(time.Now().Add(-time.Hour).UnixNano(), 999)
	a.Scan()
	if strings.Contains(a.MetricsText("cluster", "node"), "k8s_app_jvm_memory_used_bytes{") {
		t.Fatal("expired metric emitted")
	}
}

func TestMetricSnapshotTailPartialRotationAndKeywordIsolation(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "logs"), 0700)
	a, err := New(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	metric := Target{ID: "metric", Namespace: "sit", PodUID: "pod", PodName: "app-0", Directory: "logs", Glob: "*_metric.log", Start: "end", Format: "otlp_metric"}
	ordinary := metric
	ordinary.ID = "ordinary"
	ordinary.Format = ""
	ordinary.Glob = "*.log"
	ordinary.Start = "beginning"
	ordinary.Rules = []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}
	if err = a.Apply(Config{Version: 1, Targets: []Target{metric, ordinary}}); err != nil {
		t.Fatal(err)
	}
	line := metricLine(fmt.Sprintf(`{"name":"logback.events","sum":{"isMonotonic":true,"aggregationTemporality":2,"dataPoints":[{"timeUnixNano":"%d","asInt":"4","attributes":[{"key":"level","value":{"stringValue":"ERROR"}}]}]}}`, time.Now().UnixNano()))
	path := filepath.Join(root, "logs", "app_metric.log")
	os.WriteFile(path, []byte(strings.Repeat("old record\n", 250000)+line+"\npartial"), 0600)
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("ordinary", "error") != 0 {
		t.Fatal("metric log counted as ERROR")
	}
	if !strings.Contains(a.MetricsText("cluster", "node"), "k8s_app_logback_events_total{") {
		t.Fatal("latest complete tail record missing")
	}
	os.WriteFile(path, []byte("invalid\n"), 0600)
	a.Scan()
	if strings.Contains(a.MetricsText("cluster", "node"), "k8s_app_logback_events_total{") {
		t.Fatal("malformed snapshot retained stale values")
	}
	os.WriteFile(path, []byte(line+"\n"), 0600)
	a.Scan()
	a.Apply(Config{Version: 2, Targets: nil})
	if strings.Contains(a.MetricsText("cluster", "node"), "k8s_app_") {
		t.Fatal("revoked metrics retained")
	}
}
func TestOTLPRejectsDeltaAndInvalidBuckets(t *testing.T) {
	for _, payload := range []string{
		`{"name":"jvm.gc.duration","unit":"s","histogram":{"aggregationTemporality":1,"dataPoints":[{"count":"1","sum":1}]}}`,
		`{"name":"jvm.gc.duration","unit":"s","histogram":{"aggregationTemporality":2,"dataPoints":[{"count":"2","sum":1,"explicitBounds":[1],"bucketCounts":["1","0"]}]}}`,
	} {
		if _, err := parseOTLPMetricLine(metricLine(payload)); err == nil {
			t.Fatal("invalid histogram accepted")
		}
	}
}

func TestOTLPAmbiguousSeriesDoesNotDiscardOtherMetrics(t *testing.T) {
	line := metricLine(`{"name":"jvm.trace.exception.events","sum":{"isMonotonic":true,"aggregationTemporality":2,"dataPoints":[{"asInt":"10"}]}},{"name":"jvm.trace.exception.events","sum":{"isMonotonic":true,"aggregationTemporality":2,"dataPoints":[{"asInt":"2"}]}},{"name":"jvm.memory.used","unit":"By","gauge":{"dataPoints":[{"asInt":"100"}]}}`)
	samples, err := parseOTLPMetricLine(line)
	if err != nil || len(samples) != 1 || samples[0].Name != "jvm_memory_used_bytes" {
		t.Fatalf("unexpected samples: %+v, %v", samples, err)
	}
}

func TestOTLPRejectsIncompatibleGaugeType(t *testing.T) {
	for _, payload := range []string{`"histogram":{"aggregationTemporality":2,"dataPoints":[{"count":"1","sum":1}]}`, `"sum":{"isMonotonic":true,"aggregationTemporality":2,"dataPoints":[{"asInt":"1"}]}`} {
		if _, err := parseOTLPMetricLine(metricLine(`{"name":"jvm.memory.used","unit":"By",` + payload + `}`)); err == nil {
			t.Fatal("incompatible gauge type accepted")
		}
	}
}

func TestOptionalMetricFileAppearsAndDisappearsSilently(t *testing.T) {
	root := t.TempDir()
	a, err := New(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	target := Target{ID: "metric", Namespace: "sit", PodUID: "pod", PodName: "app-0", InstanceID: "guid", Directory: ".", Glob: "*_metric.log", Start: "end", Format: "otlp_metric"}
	if err = a.Apply(Config{Version: 1, Targets: []Target{target}}); err != nil {
		t.Fatal(err)
	}
	checkAbsent := func() {
		t.Helper()
		if err := a.Scan(); err != nil {
			t.Fatalf("optional metric file caused scan failure: %v", err)
		}
		if a.readErrors[target.ID] != 0 {
			t.Fatal("missing optional file counted as read error")
		}
		if text := a.MetricsText("cluster", "node"); strings.Contains(text, "k8s_app_") || strings.Contains(text, "k8s_log_file_missing{") {
			t.Fatalf("absent optional target exported metrics: %s", text)
		}
	}
	checkAbsent()
	path := filepath.Join(root, "app_metric.log")
	line := metricLine(fmt.Sprintf(`{"name":"jvm.memory.used","unit":"By","gauge":{"dataPoints":[{"timeUnixNano":"%d","asInt":"123"}]}}`, time.Now().UnixNano()))
	if err = os.WriteFile(path, []byte(line+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.MetricsText("cluster", "node"), "k8s_app_jvm_memory_used_bytes{") {
		t.Fatal("new optional metric file was not discovered")
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	checkAbsent()
}

func TestExistingMetricFileReadFailureRemainsObservable(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside.log"), filepath.Join(root, "app_metric.log")); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	target := Target{ID: "metric", Namespace: "sit", PodUID: "pod", Directory: ".", Glob: "*_metric.log", Start: "end", Format: "otlp_metric"}
	if err = a.Apply(Config{Version: 1, Targets: []Target{target}}); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	text := a.MetricsText("cluster", "node")
	if !strings.Contains(text, "k8s_app_metric_available{") || !strings.Contains(text, "k8s_app_metric_read_error_total{") || a.readErrors[target.ID] != 1 {
		t.Fatalf("existing unreadable metric file was silently ignored: %s", text)
	}
	if strings.Contains(text, "k8s_app_jvm_memory_used_bytes{") {
		t.Fatal("unreadable file exported application values")
	}
}
