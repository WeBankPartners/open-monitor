package agent

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxMetricLineBytes = 1 << 20
const appMetricFreshness = 5 * time.Minute

// Read only the bounded tail and parse the newest complete record. This works
// after restarts, truncation and rotation without replaying cumulative values.
func readLatestMetricLine(f *os.File) (string, error) {
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	size := stat.Size()
	start := size - int64(2*maxMetricLineBytes+2)
	if start < 0 {
		start = 0
	}
	raw := make([]byte, size-start)
	n, err := f.ReadAt(raw, start)
	if err != nil && err != io.EOF {
		return "", err
	}
	raw = raw[:n]
	end := bytes.LastIndexByte(raw, '\n')
	if end < 0 {
		return "", fmt.Errorf("no complete metric record")
	}
	previous := bytes.LastIndexByte(raw[:end], '\n')
	if previous < 0 && start > 0 {
		return "", fmt.Errorf("metric record exceeds limit")
	}
	line := raw[previous+1 : end]
	if len(line) > maxMetricLineBytes {
		return "", fmt.Errorf("metric record exceeds limit")
	}
	return string(line), nil
}
func (a *Agent) readAppMetrics(f *os.File, t Target) error {
	line, err := readLatestMetricLine(f)
	if err != nil {
		return err
	}
	samples, err := parseOTLPMetricLine(line)
	if err != nil {
		return err
	}
	source := sha256.Sum256([]byte(f.Name()))
	sourceID := hex.EncodeToString(source[:8])
	now := time.Now().Unix()
	prefix := t.ID + "\x00"
	count := 0
	for key := range a.appMetrics {
		if strings.HasPrefix(key, prefix) {
			count++
		}
	}
	if count+len(samples) > 4096 {
		return fmt.Errorf("target metric series limit exceeded")
	}
	for _, sample := range samples {
		if !freshAppMetric(sample, now) {
			continue
		}
		sample.Labels["source_id"] = sourceID
		a.appMetrics[appMetricKey(t.ID, sample)] = sample
	}
	return nil
}
func freshAppMetric(s appMetric, now int64) bool {
	return s.Timestamp > 0 && s.Timestamp >= now-int64(appMetricFreshness/time.Second) && s.Timestamp <= now+5
}
func appLabel(value string) string {
	// Prometheus text format supports only backslash, newline and quote escapes.
	return `"` + strings.NewReplacer(`\`, `\\`, "\n", `\n`, `"`, `\"`).Replace(value) + `"`
}
func (a *Agent) writeAppMetrics(b *strings.Builder, t Target, cluster, node string) {
	// Metric files are optional. A target with no matching file must be silent;
	// availability and read-error series are useful only after a file exists.
	if a.missing[t.ID] {
		return
	}
	now := time.Now().Unix()
	keys := []string{}
	latest := int64(0)
	for key, sample := range a.appMetrics {
		if strings.HasPrefix(key, t.ID+"\x00") && freshAppMetric(sample, now) {
			keys = append(keys, key)
			if sample.Timestamp > latest {
				latest = sample.Timestamp
			}
		}
	}
	sort.Strings(keys)
	labels := map[string]string{"cluster_id": cluster, "namespace": t.Namespace, "pod": t.PodName, "pod_uid": t.PodUID, "workload_name": t.WorkloadName, "workload_type": t.WorkloadType, "container": t.Container, "instance_id": t.InstanceID, "node": node, "target": t.ID}
	write := func(name string, value float64, extra map[string]string) {
		all := cloneLabels(labels)
		for k, v := range extra {
			all[k] = v
		}
		names := make([]string, 0, len(all))
		for k := range all {
			names = append(names, k)
		}
		sort.Strings(names)
		fmt.Fprintf(b, "%s{", name)
		for i, k := range names {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(b, "%s=%s", k, appLabel(all[k]))
		}
		fmt.Fprintf(b, "} %s\n", strconv.FormatFloat(value, 'g', -1, 64))
	}
	available := 0.0
	if a.healthy[t.ID] && len(keys) > 0 {
		available = 1
	}
	write("k8s_app_metric_available", available, nil)
	write("k8s_app_metric_read_error_total", float64(a.readErrors[t.ID]), nil)
	if latest > 0 {
		write("k8s_app_metric_sample_timestamp_seconds", float64(latest), nil)
	}
	if available == 0 {
		return
	}
	for _, key := range keys {
		s := a.appMetrics[key]
		write("k8s_app_"+s.Name, s.Value, s.Labels)
	}
}
