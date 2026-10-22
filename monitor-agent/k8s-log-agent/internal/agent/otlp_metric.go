package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type appMetric struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp int64
}

// Each entry fixes the output name and permitted dimensions; resource attributes
// and arbitrary incoming attributes never become Prometheus labels.
type metricSpec struct {
	Name, Kind, Unit string
	Labels           map[string]string
}

var appMetricWhitelist = buildAppMetricWhitelist()

func buildAppMetricWhitelist() map[string]metricSpec {
	out := map[string]metricSpec{}
	add := func(source, name, kind, unit string, labels map[string]string) {
		out[source] = metricSpec{name, kind, unit, labels}
	}
	memory := map[string]string{"area": "area", "jvm.memory.type": "area", "id": "pool", "jvm.memory.pool.name": "pool"}
	pool := map[string]string{"pool": "pool", "pool.name": "pool", "name": "pool"}
	for _, field := range []string{"used", "committed", "max", "limit"} {
		add("jvm.memory."+field, "jvm_memory_"+field+"_bytes", "gauge", "bytes", memory)
	}
	for _, field := range []string{"allocated", "promoted"} {
		add("jvm.gc.memory."+field, "jvm_gc_memory_"+field+"_bytes_total", "counter", "bytes", nil)
	}
	for _, field := range []string{"live", "daemon", "peak", "states"} {
		add("jvm.threads."+field, "jvm_threads_"+field, "gauge", "", map[string]string{"state": "state"})
	}
	add("jvm.thread.count", "jvm_thread_count", "gauge", "", map[string]string{"jvm.thread.state": "state", "jvm.thread.daemon": "daemon"})
	for _, source := range []string{"jvm.gc.pause", "jvm.gc.duration"} {
		add(source, strings.ReplaceAll(source, ".", "_")+"_seconds", "histogram", "seconds", map[string]string{"action": "action", "cause": "cause", "gc": "gc", "jvm.gc.action": "action", "jvm.gc.name": "gc"})
	}
	for _, source := range []string{"process.cpu.usage", "jvm.cpu.recent_utilization"} {
		add(source, strings.ReplaceAll(source, ".", "_"), "gauge", "", nil)
	}
	add("process.cpu.time", "process_cpu_seconds_total", "counter", "seconds", nil)
	add("jvm.cpu.time", "jvm_cpu_seconds_total", "counter", "seconds", nil)
	for _, field := range []string{"rss", "vss"} {
		add("process.memory."+field, "process_memory_"+field+"_bytes", "gauge", "bytes", nil)
	}
	for _, field := range []string{"uptime", "start.time"} {
		add("process."+field, "process_"+strings.ReplaceAll(field, ".", "_")+"_seconds", "gauge", "seconds", nil)
	}
	add("jvm.running.time", "jvm_running_time_seconds", "gauge", "seconds", nil)
	for _, field := range []string{"open", "max"} {
		add("process.files."+field, "process_files_"+field, "gauge", "", nil)
	}
	for _, field := range []string{"active", "idle", "max", "min", "pending"} {
		add("hikaricp.connections."+field, "hikaricp_connections_"+field, "gauge", "", pool)
	}
	add("hikaricp.connections.timeout", "hikaricp_connections_timeout_total", "counter", "", pool)
	for _, field := range []string{"acquire", "creation", "usage"} {
		add("hikaricp.connections."+field, "hikaricp_connections_"+field+"_seconds", "histogram", "seconds", pool)
	}
	for _, field := range []string{"active", "idle", "min", "max"} {
		add("jdbc.connections."+field, "jdbc_connections_"+field, "gauge", "", pool)
	}
	for _, field := range []string{"usage", "max", "idle.min", "idle.max", "pending_requests"} {
		labels := cloneLabels(pool)
		labels["state"] = "state"
		add("db.client.connections."+field, "db_client_connections_"+strings.ReplaceAll(field, ".", "_"), "gauge", "", labels)
	}
	for _, field := range []string{"active", "pool.size", "pool.core", "pool.max", "queued", "queue.remaining", "completed"} {
		kind := "gauge"
		name := "executor_" + strings.ReplaceAll(field, ".", "_")
		if field == "completed" {
			kind = "counter"
			name += "_total"
		}
		add("executor."+field, name, kind, "", pool)
	}
	for _, field := range []string{"threads", "work.queue"} {
		labels := cloneLabels(pool)
		labels["type"] = "type"
		add("jdk.thread.pool."+field, "jdk_thread_pool_"+strings.ReplaceAll(field, ".", "_"), "gauge", "", labels)
	}
	add("http.server.requests", "http_server_requests_seconds", "histogram", "seconds", map[string]string{"method": "method", "status": "status", "outcome": "outcome"})
	add("logback.events", "logback_events_total", "counter", "", map[string]string{"level": "level"})
	add("jvm.trace.exception.events", "jvm_trace_exception_events_total", "counter", "", nil)
	return out
}

func parseOTLPMetricLine(line string) ([]appMetric, error) {
	line = strings.TrimSpace(line)
	if len(line) > maxMetricLineBytes {
		return nil, fmt.Errorf("metric line exceeds limit")
	}
	timestamp := parseMetricLineTime(line)
	if i := strings.IndexByte(line, '{'); i >= 0 {
		line = line[i:]
	} else {
		return nil, fmt.Errorf("metric line has no JSON object")
	}
	var document struct {
		ResourceMetrics []struct {
			ScopeMetrics []struct {
				Metrics []struct {
					Name, Unit            string
					Gauge, Sum, Histogram json.RawMessage
				} `json:"metrics"`
			} `json:"scopeMetrics"`
		} `json:"resourceMetrics"`
	}
	if err := json.Unmarshal([]byte(line), &document); err != nil || len(document.ResourceMetrics) == 0 {
		return nil, fmt.Errorf("invalid OTLP metric document")
	}
	result := map[string]appMetric{}
	ambiguous := map[string]bool{}
	for _, resource := range document.ResourceMetrics {
		for _, scope := range resource.ScopeMetrics {
			for _, metric := range scope.Metrics {
				spec, ok := appMetricWhitelist[metric.Name]
				if !ok {
					continue
				}
				kind := "gauge"
				raw := metric.Gauge
				if len(metric.Sum) > 0 {
					kind = "sum"
					raw = metric.Sum
				}
				if len(metric.Histogram) > 0 {
					kind = "histogram"
					raw = metric.Histogram
				}
				parsed, err := parseOTLPData(spec, kind, metric.Unit, raw, timestamp)
				if err != nil {
					return nil, fmt.Errorf("invalid metric %s: %w", metric.Name, err)
				}
				// Dropped dimensions (e.g. URI) must not collapse distinct counters. An
				// opaque source identity preserves rate/reset semantics without leaking values.
				for _, sample := range parsed {
					key := appMetricKey("", sample)
					if ambiguous[key] {
						continue
					}
					if existing, ok := result[key]; ok {
						if existing.Timestamp == sample.Timestamp && existing.Value != sample.Value {
							// The same metric can be emitted by two instrumentation scopes.
							// Without a stable scope dimension, retaining either value would
							// create a false time series, so drop the ambiguous series.
							delete(result, key)
							ambiguous[key] = true
							continue
						}
						if existing.Timestamp > sample.Timestamp {
							continue
						}
					}
					result[key] = sample
				}
				if len(result) > 4096 {
					return nil, fmt.Errorf("metric series limit exceeded")
				}
			}
		}
	}
	keys := make([]string, 0, len(result))
	for key := range result {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]appMetric, 0, len(keys))
	for _, key := range keys {
		out = append(out, result[key])
	}
	return out, nil
}

type otlpPoint struct {
	AsDouble, AsInt, Sum, Count, TimeUnixNano, StartTimeUnixNano json.RawMessage
	Attributes                                                   []struct {
		Key   string
		Value map[string]json.RawMessage
	}
	ExplicitBounds []float64
	BucketCounts   []json.RawMessage
}

func cumulative(raw json.RawMessage) bool {
	return string(raw) == "2" || string(raw) == `"AGGREGATION_TEMPORALITY_CUMULATIVE"`
}
func parseOTLPData(spec metricSpec, kind, unit string, raw json.RawMessage, fallback int64) ([]appMetric, error) {
	var data struct {
		DataPoints             []otlpPoint
		AggregationTemporality json.RawMessage
		IsMonotonic            bool
	}
	if len(raw) == 0 || json.Unmarshal(raw, &data) != nil {
		return nil, fmt.Errorf("invalid metric payload")
	}
	if spec.Kind == "histogram" && kind != "histogram" {
		return nil, fmt.Errorf("expected histogram")
	}
	if spec.Kind == "gauge" && kind != "gauge" && (kind != "sum" || data.IsMonotonic) {
		return nil, fmt.Errorf("expected gauge")
	}
	if spec.Kind == "counter" && (kind != "sum" || !data.IsMonotonic) {
		return nil, fmt.Errorf("expected monotonic sum")
	}
	if kind == "sum" || kind == "histogram" {
		if !cumulative(data.AggregationTemporality) {
			return nil, fmt.Errorf("only cumulative temporality is supported")
		}
	}
	factor := 1.0
	if spec.Unit == "seconds" {
		switch unit {
		case "s", "seconds":
		case "ms", "milliseconds":
			factor = 0.001
		case "ns":
			factor = 1e-9
		default:
			return nil, fmt.Errorf("unsupported duration unit")
		}
	}
	if spec.Unit == "bytes" && unit != "bytes" && unit != "By" {
		return nil, fmt.Errorf("unsupported byte unit")
	}
	var out []appMetric
	for _, p := range data.DataPoints {
		labels := map[string]string{}
		identity := map[string]string{}
		for _, a := range p.Attributes {
			raw, _ := json.Marshal(a.Value)
			identity[a.Key] = string(raw)
			label, allowed := spec.Labels[a.Key]
			if !allowed {
				continue
			}
			var value string
			if v, ok := a.Value["stringValue"]; ok {
				if json.Unmarshal(v, &value) != nil {
					return nil, fmt.Errorf("invalid attribute")
				}
			} else if v, ok := a.Value["boolValue"]; ok {
				value = string(v)
			} else {
				continue
			}
			if len(value) > 128 {
				return nil, fmt.Errorf("metric label exceeds limit")
			}
			labels[label] = value
		}
		if len(identity) > 0 {
			encoded, _ := json.Marshal(identity)
			hash := sha256.Sum256(encoded)
			labels["series_id"] = hex.EncodeToString(hash[:8])
		}
		// Label validation and allowlisting prevent resource attributes overriding scope.
		timestamp := fallback
		if len(p.TimeUnixNano) > 0 {
			n, err := strconv.ParseInt(strings.Trim(string(p.TimeUnixNano), `"`), 10, 64)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("invalid timestamp")
			}
			timestamp = n / int64(time.Second)
		}
		appendSample := func(name string, v float64, l map[string]string) {
			out = append(out, appMetric{Name: name, Value: v, Labels: l, Timestamp: timestamp})
		}
		if kind != "histogram" {
			rawValue := p.AsDouble
			if len(rawValue) == 0 {
				rawValue = p.AsInt
			}
			value, ok := metricRawNumber(rawValue)
			if !ok {
				return nil, fmt.Errorf("invalid point value")
			}
			if spec.Kind == "counter" && value < 0 {
				return nil, fmt.Errorf("negative counter")
			}
			appendSample(spec.Name, value*factor, labels)
			continue
		}
		count, ok := metricRawNumber(p.Count)
		if len(p.Count) == 0 {
			count = 0
			ok = true
		}
		sum, valid := metricRawNumber(p.Sum)
		if !ok || !valid || count < 0 || count != math.Trunc(count) || sum < 0 {
			return nil, fmt.Errorf("invalid histogram totals")
		}
		appendSample(spec.Name+"_sum", sum*factor, labels)
		appendSample(spec.Name+"_count", count, labels)
		if len(p.BucketCounts) == 0 {
			if len(p.ExplicitBounds) > 0 {
				return nil, fmt.Errorf("missing buckets")
			}
			continue
		}
		if len(p.BucketCounts) != len(p.ExplicitBounds)+1 {
			return nil, fmt.Errorf("invalid bucket count")
		}
		total := 0.0
		previous := math.Inf(-1)
		for i, rawCount := range p.BucketCounts {
			v, ok := metricRawNumber(rawCount)
			if !ok || v < 0 || v != math.Trunc(v) {
				return nil, fmt.Errorf("invalid bucket value")
			}
			total += v
			l := cloneLabels(labels)
			l["le"] = "+Inf"
			if i < len(p.ExplicitBounds) {
				bound := p.ExplicitBounds[i]
				if math.IsNaN(bound) || math.IsInf(bound, 0) || bound <= previous {
					return nil, fmt.Errorf("invalid bucket bounds")
				}
				previous = bound
				l["le"] = strconv.FormatFloat(bound*factor, 'g', -1, 64)
			}
			appendSample(spec.Name+"_bucket", total, l)
		}
		if total != count {
			return nil, fmt.Errorf("histogram count mismatch")
		}
	}
	return out, nil
}

func cloneLabels(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
func metricRawNumber(raw json.RawMessage) (float64, bool) {
	var v float64
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	if json.Unmarshal(raw, &v) == nil {
		return v, math.IsNaN(v) == false && math.IsInf(v, 0) == false
	}
	v, e := strconv.ParseFloat(strings.Trim(string(raw), `"`), 64)
	return v, e == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
}
func parseMetricLineTime(line string) int64 {
	if len(line) < 22 || line[0] != '[' {
		return 0
	}
	end := strings.IndexByte(line, ']')
	if end < 0 {
		return 0
	}
	raw := line[1:end]
	for _, layout := range []string{"2006-01-02 15:04:05.000", "2006-01-02 15:04:05:000"} {
		if t, e := time.ParseInLocation(layout, raw, time.Local); e == nil {
			return t.Unix()
		}
	}
	return 0
}
