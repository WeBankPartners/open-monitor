package k8s

import (
	"math"
	"regexp"
	"strconv"
)

// Object metrics use the Pod UID, not the container runtime ID: Kubernetes
// restartCount and lastState survive container restarts within the same Pod.
func SnapshotMetricSamples(snapshot *Snapshot) []MetricSample {
	if snapshot == nil {
		return nil
	}
	result := []MetricSample{}
	pods := map[string]bool{}
	for _, c := range snapshot.Containers {
		labels := map[string]string{"cluster_id": c.Key.ClusterID, "namespace": c.Key.Namespace, "workload_type": c.Key.WorkloadType, "workload_name": c.Key.WorkloadName, "pod": c.Key.PodName, "container": c.Key.ContainerName, "pod_uid": c.PodUID}
		add := func(name string, value float64, labels map[string]string) {
			result = append(result, MetricSample{Name: name, Value: value, Labels: labels})
		}
		for _, kind := range []string{"request", "limit"} {
			resources := c.Resources.Requests
			if kind == "limit" {
				resources = c.Resources.Limits
			}
			for _, resource := range []string{"cpu", "memory"} {
				if value, ok := quantityValue(resource, resources[resource]); ok {
					unit := "bytes"
					if resource == "cpu" {
						unit = "cores"
					}
					add("k8s_pod_"+resource+"_"+kind+"_"+unit, value, labels)
				}
			}
		}
		if c.StatusKnown {
			add("k8s_pod_restart_count", float64(c.RestartCount), labels)
			add("k8s_container_ready", boolFloat(c.Ready), labels)
			add("k8s_container_last_termination_oom", boolFloat(c.LastTerminationReason == "OOMKilled"), labels)
		}
		if !pods[c.PodUID] {
			pods[c.PodUID] = true
			podLabels := cloneLabels(labels)
			delete(podLabels, "container")
			for _, phase := range []string{"Pending", "Running", "Succeeded", "Failed", "Unknown"} {
				if c.Phase == phase {
					for _, state := range []string{"Pending", "Running", "Succeeded", "Failed", "Unknown"} {
						stateLabels := cloneLabels(podLabels)
						stateLabels["phase"] = state
						add("k8s_pod_phase", boolFloat(c.Phase == state), stateLabels)
					}
					break
				}
			}
			if c.PodReady == "True" || c.PodReady == "False" {
				add("k8s_pod_ready", boolFloat(c.PodReady == "True"), podLabels)
			}
		}
	}
	return result
}
func cloneLabels(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
func boolFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

// Kubernetes decimal, binary and exponent quantities; only CPU/memory are exported.
var resourceQuantity = regexp.MustCompile(`^([+]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+))([eE][+-]?[0-9]+|[numkKMGTPE]|[KMGTPE]i)?$`)

func quantityValue(resource, raw string) (float64, bool) {
	if resource != "cpu" && resource != "memory" {
		return 0, false
	}
	parts := resourceQuantity.FindStringSubmatch(raw)
	if parts == nil {
		return 0, false
	}
	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, false
	}
	suffix := parts[2]
	multipliers := map[string]float64{"": 1, "n": 1e-9, "u": 1e-6, "m": 1e-3, "k": 1e3, "K": 1e3, "M": 1e6, "G": 1e9, "T": 1e12, "P": 1e15, "E": 1e18, "Ki": 1 << 10, "Mi": 1 << 20, "Gi": 1 << 30, "Ti": 1 << 40, "Pi": 1 << 50, "Ei": 1 << 60}
	if factor, ok := multipliers[suffix]; ok {
		value *= factor
	} else {
		value, err = strconv.ParseFloat(parts[1]+suffix, 64)
	}
	return value, err == nil && value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
