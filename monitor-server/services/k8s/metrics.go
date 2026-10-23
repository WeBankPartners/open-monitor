package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/common/expfmt"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
)

type MetricSample struct {
	Name      string
	Value     float64
	Timestamp int64
	Labels    map[string]string
}

var cadvisorMetricNames = map[string]string{
	"container_cpu_usage_seconds_total":         "k8s_pod_cpu_usage_seconds_total",
	"container_memory_working_set_bytes":        "k8s_pod_memory_working_set_bytes",
	"container_network_receive_bytes_total":     "k8s_pod_network_receive_bytes_total",
	"container_network_transmit_bytes_total":    "k8s_pod_network_transmit_bytes_total",
	"container_blkio_device_usage_total":        "k8s_pod_block_io_bytes_total",
	"container_cpu_cfs_periods_total":           "k8s_pod_cpu_cfs_periods_total",
	"container_cpu_cfs_throttled_periods_total": "k8s_pod_cpu_cfs_throttled_periods_total",
	"container_cpu_cfs_throttled_seconds_total": "k8s_pod_cpu_cfs_throttled_seconds_total",
	"container_oom_events_total":                "k8s_pod_oom_events_total",
	"container_memory_failcnt":                  "k8s_pod_memory_failcnt_total",
}

func ParseCAdvisor(payload string, snapshot *Snapshot) ([]MetricSample, error) {
	if snapshot == nil {
		return nil, errors.New("snapshot is required")
	}
	var parser expfmt.TextParser
	families, err := parser.TextToMetricFamilies(strings.NewReader(payload))
	if err != nil {
		return nil, errors.New("invalid node metrics exposition")
	}
	allowed := map[string]Container{}
	for _, c := range snapshot.Containers {
		if c.PodUID != "" {
			allowed[c.Key.Namespace+"\x00"+c.Key.PodName+"\x00"+c.Key.ContainerName] = c
		}
	}
	result := []MetricSample{}
	seen := map[string]bool{}
	for name, mapped := range cadvisorMetricNames {
		family := families[name]
		if family == nil {
			continue
		}
		for _, m := range family.Metric {
			labels := map[string]string{}
			for _, p := range m.Label {
				labels[p.GetName()] = p.GetValue()
			}
			network := strings.HasPrefix(name, "container_network_")
			c, ok := allowed[labels["namespace"]+"\x00"+labels["pod"]+"\x00"+labels["container"]]
			if network {
				// Network counters are pod-scoped. Only accept the pod/sandbox source;
				// assigning them to each business container would multiply traffic.
				if labels["container"] != "" && labels["container"] != "POD" {
					continue
				}
				ok = false
				for _, candidate := range snapshot.Containers {
					if candidate.Key.Namespace == labels["namespace"] && candidate.Key.PodName == labels["pod"] && networkCgroupMatches(labels["id"], candidate.PodUID) {
						c, ok = candidate, true
						break
					}
				}
			}
			if !ok || c.PodUID == "" {
				continue
			}
			if !network {
				id := c.ContainerID
				if _, v, ok := strings.Cut(id, "://"); ok {
					id = v
				}
				leaf := path.Base(labels["id"])
				if id == "" || !(leaf == id || leaf == "cri-containerd-"+id+".scope" || leaf == "crio-"+id+".scope" || leaf == "docker-"+id+".scope") {
					continue
				}
			}
			if name == "container_cpu_usage_seconds_total" && labels["cpu"] != "" && labels["cpu"] != "total" {
				continue
			}
			if name == "container_blkio_device_usage_total" && (labels["device"] == "" || (labels["operation"] != "Read" && labels["operation"] != "Write")) {
				continue
			}
			var v float64
			switch {
			case m.Counter != nil:
				v = m.Counter.GetValue()
			case m.Gauge != nil:
				v = m.Gauge.GetValue()
			case m.Untyped != nil:
				v = m.Untyped.GetValue()
			default:
				continue
			}
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
				continue
			}
			safe := map[string]string{"cluster_id": c.Key.ClusterID, "namespace": c.Key.Namespace, "workload_type": c.Key.WorkloadType, "workload_name": c.Key.WorkloadName, "pod": c.Key.PodName, "container": c.Key.ContainerName, "pod_uid": c.PodUID, "container_id": c.ContainerID}
			if network {
				safe["container"] = ""
				safe["container_id"] = ""
				safe["interface"] = labels["interface"]
				safe["network_source"] = labels["id"]
				if safe["interface"] == "" {
					continue
				}
			} else if name == "container_blkio_device_usage_total" {
				safe["device"] = labels["device"]
				safe["operation"] = labels["operation"]
			}
			key := mapped + "\x00" + c.Key.Namespace + "\x00" + c.Key.PodName + "\x00" + c.Key.ContainerName + "\x00" + labels["interface"] + "\x00" + labels["device"] + "\x00" + labels["operation"]
			if seen[key] {
				continue
			}
			seen[key] = true
			if name == "container_blkio_device_usage_total" {
				if labels["operation"] == "Read" {
					mapped = "k8s_pod_block_read_bytes_total"
				} else {
					mapped = "k8s_pod_block_write_bytes_total"
				}
			}
			result = append(result, MetricSample{Name: mapped, Value: v, Timestamp: m.GetTimestampMs(), Labels: safe})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
func networkCgroupMatches(cgroup, podUID string) bool {
	if cgroup == "" || podUID == "" {
		return false
	}
	normalized := strings.ToLower(strings.ReplaceAll(podUID, "-", "_"))
	for _, part := range strings.Split(strings.ToLower(cgroup), "/") {
		part = strings.TrimSuffix(part, ".slice")
		for _, podPart := range []string{"pod" + normalized, "pod" + strings.ToLower(podUID)} {
			if part == podPart || strings.HasSuffix(part, "-"+podPart) {
				return true
			}
		}
	}
	return false
}

func RenderMetricSamples(samples []MetricSample) (string, error) {
	var b strings.Builder
	for _, s := range samples {
		if !validMetricName(s.Name) {
			return "", fmt.Errorf("invalid metric name")
		}
		b.WriteString(s.Name)
		if len(s.Labels) > 0 {
			keys := make([]string, 0, len(s.Labels))
			for k := range s.Labels {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			b.WriteByte('{')
			for i, k := range keys {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteString(k)
				b.WriteString("=\"")
				b.WriteString(strconv.Quote(s.Labels[k])[1 : len(strconv.Quote(s.Labels[k]))-1])
				b.WriteByte('"')
			}
			b.WriteByte('}')
		}
		b.WriteByte(' ')
		b.WriteString(strconv.FormatFloat(s.Value, 'g', -1, 64))
		if s.Timestamp != 0 {
			b.WriteByte(' ')
			b.WriteString(strconv.FormatInt(s.Timestamp, 10))
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}
func validMetricName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if !(r == '_' || r == ':' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

type MetricSource interface {
	Fetch(context.Context, string) (string, error)
}
type Collector struct{ Source MetricSource }

func (c Collector) Collect(ctx context.Context, snapshot *Snapshot) ([]MetricSample, error) {
	if c.Source == nil || snapshot == nil {
		return nil, errors.New("metric collector is not configured")
	}
	nodes := map[string]bool{}
	for _, item := range snapshot.Containers {
		if item.Node != "" {
			nodes[item.Node] = true
		}
	}
	result := []MetricSample{}
	for node := range nodes {
		payload, err := c.Source.Fetch(ctx, node)
		if err != nil {
			return nil, err
		}
		nodeSnapshot := &Snapshot{}
		for _, item := range snapshot.Containers {
			if item.Node == node {
				nodeSnapshot.Containers = append(nodeSnapshot.Containers, item)
			}
		}
		samples, err := ParseCAdvisor(payload, nodeSnapshot)
		if err != nil {
			return nil, err
		}
		result = append(result, samples...)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
