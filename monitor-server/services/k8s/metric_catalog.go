package k8s

import "strings"

// MetricDefinition is the history query allowlist, also used by the Web picker.
// Application names match the Agent exporter; histogram views use count/sum.
type MetricDefinition struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
	Scope    string `json:"scope"`
	Name     string `json:"name"`
	kind     string
}

var historyMetricCatalog = buildMetricCatalog()
var historyMetricDefinitions = func() map[string]MetricDefinition {
	definitions := make(map[string]MetricDefinition, len(historyMetricCatalog))
	for _, metric := range historyMetricCatalog {
		definitions[metric.ID] = metric
	}
	return definitions
}()

func MetricCatalog() []MetricDefinition {
	return append([]MetricDefinition(nil), historyMetricCatalog...)
}

func buildMetricCatalog() []MetricDefinition {
	out := []MetricDefinition{}
	add := func(id, category, unit, scope, name, kind string) {
		out = append(out, MetricDefinition{id, category, unit, scope, name, kind})
	}
	for _, item := range []struct{ id, unit, scope string }{
		{"cpu", "cores", "container"}, {"memory", "bytes", "container"},
		{"cpu_request", "cores", "container"}, {"cpu_limit", "cores", "container"},
		{"memory_request", "bytes", "container"}, {"memory_limit", "bytes", "container"},
		{"cpu_throttled", "s/s", "container"}, {"cpu_throttled_periods", "events/s", "container"},
		{"network_receive", "bytes/s", "pod"}, {"network_transmit", "bytes/s", "pod"},
		{"block_read", "bytes/s", "container"}, {"block_write", "bytes/s", "container"},
		{"restart_count", "count", "container"}, {"ready", "state", "pod"}, {"phase", "state", "pod"},
		{"container_ready", "state", "container"}, {"last_termination_oom", "state", "container"},
		{"oom_events", "events/window", "container"}, {"memory_failcnt", "count", "container"},
		{"quality", "state", "namespace"},
	} {
		add(item.id, "resource", item.unit, item.scope, item.id, "")
	}
	add("log_matches", "log", "events/window", "container", "log_matches", "")
	add("log_health", "log", "state", "container", "log_health", "")
	// Keep this list aligned with buildAppMetricWhitelist in the independent Agent.
	gauges := strings.Fields(`jvm_memory_used_bytes jvm_memory_committed_bytes jvm_memory_max_bytes jvm_memory_limit_bytes
		jvm_threads_live jvm_threads_daemon jvm_threads_peak jvm_threads_states jvm_thread_count
		process_cpu_usage jvm_cpu_recent_utilization process_memory_rss_bytes process_memory_vss_bytes
		process_uptime_seconds process_start_time_seconds jvm_running_time_seconds process_files_open process_files_max
		hikaricp_connections_active hikaricp_connections_idle hikaricp_connections_max hikaricp_connections_min hikaricp_connections_pending
		jdbc_connections_active jdbc_connections_idle jdbc_connections_min jdbc_connections_max
		db_client_connections_usage db_client_connections_max db_client_connections_idle_min db_client_connections_idle_max db_client_connections_pending_requests
		executor_active executor_pool_size executor_pool_core executor_pool_max executor_queued executor_queue_remaining
		jdk_thread_pool_threads jdk_thread_pool_work_queue`)
	for _, name := range gauges {
		unit := "count"
		switch {
		case strings.HasSuffix(name, "_bytes"):
			unit = "bytes"
		case strings.HasSuffix(name, "_seconds"):
			unit = "seconds"
		case name == "process_cpu_usage" || name == "jvm_cpu_recent_utilization":
			unit = "ratio"
		}
		add("app_"+name, "application", unit, "container", "k8s_app_"+name, "gauge")
	}
	for _, name := range strings.Fields(`jvm_gc_memory_allocated_bytes_total jvm_gc_memory_promoted_bytes_total process_cpu_seconds_total jvm_cpu_seconds_total hikaricp_connections_timeout_total executor_completed_total logback_events_total jvm_trace_exception_events_total`) {
		unit := "events/s"
		if strings.Contains(name, "_bytes_total") {
			unit = "bytes/s"
		} else if strings.Contains(name, "_seconds_total") {
			unit = "cores"
		}
		add("app_"+name, "application", unit, "container", "k8s_app_"+name, "counter")
	}
	for _, name := range strings.Fields(`jvm_gc_pause_seconds jvm_gc_duration_seconds hikaricp_connections_acquire_seconds hikaricp_connections_creation_seconds hikaricp_connections_usage_seconds http_server_requests_seconds`) {
		add("app_"+name+"_rate", "application", "events/s", "container", "k8s_app_"+name, "histogram_rate")
		add("app_"+name+"_mean", "application", "seconds", "container", "k8s_app_"+name, "histogram_mean")
	}
	add("app_metric_available", "application", "state", "container", "k8s_app_metric_available", "diagnostic")
	add("app_metric_read_error_total", "application", "events/s", "container", "k8s_app_metric_read_error_total", "diagnostic_counter")
	return out
}
