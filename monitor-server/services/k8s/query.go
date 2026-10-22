package k8s

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func MetricQuery(cluster, namespace, metric string, scrapeIntervals ...time.Duration) (string, error) {
	// Alert rules require one value per container. Application dimensions are
	// available only through the history query until aggregation is defined.
	if historyMetricDefinitions[metric].Category == "application" {
		return "", errors.New("application metric thresholds are not supported")
	}
	return objectMetricQuery(cluster, namespace, metric, MetricObjectFilter{}, scrapeIntervals...)
}

type MetricObjectFilter struct {
	WorkloadType string `json:"workloadType"`
	WorkloadName string `json:"workloadName"`
	PodName      string `json:"podName"`
	Container    string `json:"container"`
}

var metricObjectName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

func ObjectMetricQuery(cluster, namespace, metric string, object MetricObjectFilter, interval time.Duration) (string, error) {
	return objectMetricQuery(cluster, namespace, metric, object, interval)
}

func objectMetricQuery(cluster, namespace, metric string, object MetricObjectFilter, scrapeIntervals ...time.Duration) (string, error) {
	if len(cluster) > 64 || !clusterPattern.MatchString(cluster) || len(namespace) > 63 || !namespacePattern.MatchString(namespace) {
		return "", errors.New("invalid metric scope")
	}
	interval := (Policy{}).MetricsInterval()
	if len(scrapeIntervals) > 0 {
		interval = (Policy{MetricsIntervalSeconds: int(scrapeIntervals[0] / time.Second)}).MetricsInterval()
	}
	seconds := int(interval / time.Second)
	rateWindow := strconv.Itoa(4*seconds) + "s"
	countWindowSeconds := 4 * seconds
	if countWindowSeconds < 300 {
		countWindowSeconds = 300
	}
	countWindow := strconv.Itoa(countWindowSeconds) + "s"
	logFreshness := strconv.Itoa(3 * seconds)
	scopeSelector := "{cluster_id=" + strconv.Quote(cluster) + ",namespace=" + strconv.Quote(namespace) + "}"
	definition, supported := historyMetricDefinitions[metric]
	if !supported {
		return "", errors.New("unsupported metric")
	}
	selector := strings.TrimSuffix(scopeSelector, "}")
	for _, field := range []struct{ label, value string }{
		{"workload_type", object.WorkloadType}, {"workload_name", object.WorkloadName}, {"pod", object.PodName}, {"container", object.Container},
	} {
		if field.value == "" {
			continue
		}
		if len(field.value) > 253 || !metricObjectName.MatchString(field.value) {
			return "", errors.New("invalid metric object selector")
		}
		if definition.Scope == "namespace" || (definition.Scope == "pod" && field.label == "container") {
			continue
		}
		selector += "," + field.label + "=" + strconv.Quote(field.value)
	}
	selector += "}"
	if definition.Category == "application" {
		sampleName := definition.Name
		if strings.HasPrefix(definition.kind, "histogram") {
			sampleName += "_count"
		}
		value := definition.Name + selector
		switch definition.kind {
		case "counter", "diagnostic_counter":
			value = "rate(" + value + "[" + rateWindow + "])"
		case "histogram_rate":
			value = "rate(" + definition.Name + "_count" + selector + "[" + rateWindow + "])"
		case "histogram_mean":
			count := "rate(" + definition.Name + "_count" + selector + "[" + rateWindow + "])"
			value = "(rate(" + definition.Name + "_sum" + selector + "[" + rateWindow + "]) / (" + count + " > 0))"
		}
		if !strings.HasPrefix(definition.kind, "diagnostic") {
			value += " and on (cluster_id,namespace,pod,pod_uid,container,target,node,instance,job) (k8s_app_metric_available" + selector + " == 1)"
		}
		return value + " and (time() - timestamp(" + sampleName + selector + ") < " + logFreshness + ") and on (instance,job) (up{job=\"k8s_log_agent\"} == 1)", nil
	}
	containerGroup := " by (cluster_id, namespace, workload_type, workload_name, pod, pod_uid, container) "
	podGroup := " by (cluster_id, namespace, workload_type, workload_name, pod, pod_uid) "
	switch metric {
	case "log_matches":
		// Filter each target before aggregation. A failed or stale target must
		// never contribute a healthy zero or hide behind another target.
		target := "cluster_id,namespace,pod,pod_uid,container,target,node,instance,job"
		health := "(k8s_log_target_healthy" + selector + " == 1) and (time() - timestamp(k8s_log_target_healthy" + selector + ") < " + logFreshness + ")"
		counts := "increase(k8s_log_match_total" + selector + "[" + countWindow + "]) and on (" + target + ") (" + health + ") and on (instance,job) (up{job=\"k8s_log_agent\"} == 1)"
		bad := "k8s_log_target_healthy" + selector + " unless (" + health + " and on (instance,job) (up{job=\"k8s_log_agent\"} == 1))"
		return "sum" + containerGroup + "(" + counts + ") unless on (cluster_id,namespace,pod,pod_uid,container) (" + bad + ")", nil
	case "log_health":
		return "min" + containerGroup + "(k8s_log_target_healthy" + selector + " and (time() - timestamp(k8s_log_target_healthy" + selector + ") < " + logFreshness + ") and on (instance,job) (up{job=\"k8s_log_agent\"} == 1))", nil
	case "cpu":
		return "sum" + containerGroup + "(rate(k8s_pod_cpu_usage_seconds_total" + selector + "[" + rateWindow + "]))", nil
	case "memory":
		return "sum" + containerGroup + "(k8s_pod_memory_working_set_bytes" + selector + ")", nil
	case "network_receive":
		return "sum" + containerGroup + "(rate(k8s_pod_network_receive_bytes_total" + selector + "[" + rateWindow + "]))", nil
	case "network_transmit":
		return "sum" + containerGroup + "(rate(k8s_pod_network_transmit_bytes_total" + selector + "[" + rateWindow + "]))", nil
	case "block_read":
		return "sum" + containerGroup + "(rate(k8s_pod_block_read_bytes_total" + selector + "[" + rateWindow + "]))", nil
	case "block_write":
		return "sum" + containerGroup + "(rate(k8s_pod_block_write_bytes_total" + selector + "[" + rateWindow + "]))", nil
	case "cpu_throttled":
		return "sum" + containerGroup + "(rate(k8s_pod_cpu_cfs_throttled_seconds_total" + selector + "[" + rateWindow + "]))", nil
	case "cpu_throttled_periods":
		return "sum" + containerGroup + "(rate(k8s_pod_cpu_cfs_throttled_periods_total" + selector + "[" + rateWindow + "]))", nil
	case "cpu_request":
		return "sum" + containerGroup + "(k8s_pod_cpu_request_cores" + selector + ")", nil
	case "cpu_limit":
		return "sum" + containerGroup + "(k8s_pod_cpu_limit_cores" + selector + ")", nil
	case "memory_request":
		return "sum" + containerGroup + "(k8s_pod_memory_request_bytes" + selector + ")", nil
	case "memory_limit":
		return "sum" + containerGroup + "(k8s_pod_memory_limit_bytes" + selector + ")", nil
	case "restart_count":
		return "sum" + containerGroup + "(k8s_pod_restart_count" + selector + ")", nil
	case "ready":
		return "max" + podGroup + "(k8s_pod_ready" + selector + ")", nil
	case "phase":
		return "k8s_pod_phase" + selector, nil
	case "oom_events":
		return "sum" + containerGroup + "(increase(k8s_pod_oom_events_total" + selector + "[" + countWindow + "]) and k8s_pod_oom_events_total" + selector + ") and on (cluster_id,namespace) (k8s_source_good" + scopeSelector + " == 1)", nil
	case "memory_failcnt":
		return "sum" + containerGroup + "(k8s_pod_memory_failcnt_total" + selector + ")", nil
	case "container_ready":
		return "max" + containerGroup + "(k8s_container_ready" + selector + ")", nil
	case "last_termination_oom":
		return "max" + containerGroup + "(k8s_container_last_termination_oom" + selector + ")", nil
	case "quality":
		return "k8s_source_good" + selector, nil
	default:
		return "", errors.New("unsupported metric")
	}
}
