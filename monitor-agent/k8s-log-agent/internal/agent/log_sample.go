package agent

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	logSampleMaxCharacters = 1024
	logSampleInterval      = int64(5 * time.Minute / time.Second)
	logSampleRetention     = int64(15 * time.Minute / time.Second)
	logSampleMaxEntries    = 128
	logSampleMaxUpdates    = 20
)

var logSampleANSI = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

type logSample struct {
	File       string
	FileID     string
	Offset     int64
	ObservedAt int64
	Preview    string
	Truncated  bool
}

type logSampleBudget struct {
	StartedAt int64
	Updates   int
}

func logSampleKey(t Target, r Rule) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s", t.ID, t.PodUID, r.ID, r.Version, t.Directory, t.Glob)
}

func freshLogSample(s logSample, now int64) bool {
	return s.ObservedAt > 0 && s.ObservedAt <= now+5 && now-s.ObservedAt < logSampleRetention
}

func pruneLogSamples(c *checkpoint, now int64) {
	if c.SampleBudget.StartedAt < 0 || c.SampleBudget.StartedAt > now+5 || c.SampleBudget.Updates < 0 {
		c.SampleBudget = logSampleBudget{}
	}
	if c.SampleBudget.Updates > logSampleMaxUpdates {
		c.SampleBudget.Updates = logSampleMaxUpdates
	}
	if c.Samples == nil {
		c.Samples = map[string]logSample{}
	}
	for key, sample := range c.Samples {
		if !freshLogSample(sample, now) || sample.Offset < 0 || sample.File == "" || sample.FileID == "" || !utf8.ValidString(sample.Preview) || utf8.RuneCountInString(sample.Preview) > logSampleMaxCharacters {
			delete(c.Samples, key)
		}
	}
	if len(c.Samples) > logSampleMaxEntries {
		keys := make([]string, 0, len(c.Samples))
		for key := range c.Samples {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if c.Samples[keys[i]].ObservedAt == c.Samples[keys[j]].ObservedAt {
				return keys[i] < keys[j]
			}
			return c.Samples[keys[i]].ObservedAt > c.Samples[keys[j]].ObservedAt
		})
		for _, key := range keys[logSampleMaxEntries:] {
			delete(c.Samples, key)
		}
	}
}

// Build at most one candidate per file/rule scan. Commit it only when that
// rule's cursor advances, so a later read error cannot publish an uncommitted hit.
func newLogSample(c *checkpoint, key, file, fileID string, offset int64, line []byte, now int64) *logSample {
	previous, exists := c.Samples[key]
	if exists && now-previous.ObservedAt < logSampleInterval {
		return nil
	}
	budget := c.SampleBudget
	if (!exists && len(c.Samples) >= logSampleMaxEntries) || (now >= budget.StartedAt && now-budget.StartedAt < logSampleInterval && budget.Updates >= logSampleMaxUpdates) {
		c.SampleLimitHits++
		return nil
	}
	preview := strings.ToValidUTF8(logSampleANSI.ReplaceAllString(string(line), ""), "\ufffd")
	preview = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, preview))
	runes := []rune(preview)
	truncated := len(runes) > logSampleMaxCharacters
	if truncated {
		preview = string(runes[:logSampleMaxCharacters])
	}
	return &logSample{File: strings.ToValidUTF8(filepath.Base(file), "\ufffd"), FileID: fileID, Offset: offset, ObservedAt: now, Preview: preview, Truncated: truncated}
}

func commitLogSample(c *checkpoint, key string, sample logSample) {
	if sample.ObservedAt < c.SampleBudget.StartedAt || sample.ObservedAt-c.SampleBudget.StartedAt >= logSampleInterval {
		c.SampleBudget = logSampleBudget{StartedAt: sample.ObservedAt}
	}
	c.SampleBudget.Updates++
	c.Samples[key] = sample
}

func (a *Agent) writeLogSample(b *strings.Builder, t Target, r Rule, cluster, node string, now int64) {
	sample, exists := a.saved.Samples[logSampleKey(t, r)]
	if !exists || !freshLogSample(sample, now) {
		return
	}
	labels := map[string]string{
		"cluster_id": cluster, "namespace": t.Namespace, "pod": t.PodName, "pod_uid": t.PodUID,
		"instance_id": t.InstanceID, "workload_name": t.WorkloadName, "workload_type": t.WorkloadType,
		"container": t.Container, "target": t.ID, "rule": r.ID, "rule_version": strconv.FormatUint(r.Version, 10), "node": node,
		"file": sample.File, "file_id": sample.FileID, "byte_offset": strconv.FormatInt(sample.Offset, 10),
		"preview": sample.Preview, "truncated": strconv.FormatBool(sample.Truncated),
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	b.WriteString("k8s_log_match_sample_timestamp_seconds{")
	for i, key := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(b, "%s=%s", key, appLabel(labels[key]))
	}
	fmt.Fprintf(b, "} %d\n", sample.ObservedAt)
}
