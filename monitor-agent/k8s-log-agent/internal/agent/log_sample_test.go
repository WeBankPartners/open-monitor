package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

const sampleMetricName = "k8s_log_match_sample_timestamp_seconds"

func sampleTestAgent(t *testing.T, content string) (*Agent, string, Config) {
	t.Helper()
	root, state := t.TempDir(), t.TempDir()
	path := filepath.Join(root, "app.log")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	cfg := stateTestConfig()
	cfg.Targets[0].Start = "beginning"
	cfg.Targets[0].PodName = "app-0"
	cfg.Targets[0].Container = "app"
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	return a, path, cfg
}

func sampleMetricLabels(t *testing.T, text string) map[string]string {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, sampleMetricName+"{") {
			continue
		}
		labels := map[string]string{}
		for _, match := range regexp.MustCompile(`([a-z_]+)=("(?:[^"\\]|\\.)*")`).FindAllStringSubmatch(line, -1) {
			value, err := strconv.Unquote(match[2])
			if err != nil {
				t.Fatalf("invalid sample label %s: %v", match[1], err)
			}
			labels[match[1]] = value
		}
		if len(labels) == 0 {
			t.Fatal("sample has no labels")
		}
		return labels
	}
	t.Fatal("matched log did not export a representative sample")
	return nil
}

func TestLogSampleExportsRawBytePositionAndCleanPreview(t *testing.T) {
	prefix := "INFO 正常运行\n"
	a, _, _ := sampleTestAgent(t, prefix+"ERROR \x1b[31mfailed\x1b[0m \"quoted\" C:\\logs\\app\t详情\x00\r\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	metrics := a.MetricsText("cluster-a", "node-a")
	labels := sampleMetricLabels(t, metrics)
	if labels["file"] != "app.log" || labels["byte_offset"] != strconv.Itoa(len(prefix)) || labels["file_id"] == "" {
		t.Fatalf("sample location is incomplete or counts characters instead of bytes: %v", labels)
	}
	if labels["preview"] != "ERROR failed \"quoted\" C:\\logs\\app 详情" || labels["truncated"] != "false" {
		t.Fatalf("unexpected cleaned preview: %v", labels)
	}
	for _, name := range []string{"cluster_id", "namespace", "pod", "pod_uid", "container", "target", "rule", "rule_version", "node"} {
		if labels[name] == "" {
			t.Errorf("sample is missing identity label %s", name)
		}
	}
	for _, line := range strings.Split(metrics, "\n") {
		if strings.HasPrefix(line, "k8s_log_match_total{") && (strings.Contains(line, "preview=") || strings.Contains(line, "byte_offset=") || !strings.HasSuffix(line, " 1")) {
			t.Fatal("sample changed counter identity or value")
		}
	}
}

func TestLogSampleTruncatesUnicodeWithoutChangingCount(t *testing.T) {
	a, _, _ := sampleTestAgent(t, "ERROR "+strings.Repeat("测", 1100)+"\nERROR second\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	labels := sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
	if len([]rune(labels["preview"])) != 1024 || labels["truncated"] != "true" {
		t.Fatalf("preview did not truncate at 1024 Unicode characters: length=%d truncated=%s", len([]rune(labels["preview"])), labels["truncated"])
	}
	if a.Count("app", "error") != 2 {
		t.Fatal("sample selection lost matching lines")
	}
}

func TestLogSampleSurvivesRestartWithoutReplayingLogs(t *testing.T) {
	a, path, cfg := sampleTestAgent(t, "ERROR first\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	before := sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
	a.Close()
	b, err := New(filepath.Dir(path), a.state)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err = b.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	appendStateTestLog(t, path)
	if err = b.Scan(); err != nil {
		t.Fatal(err)
	}
	after := sampleMetricLabels(t, b.MetricsText("cluster-a", "node-a"))
	if after["preview"] != before["preview"] || after["byte_offset"] != before["byte_offset"] || b.Count("app", "error") != 2 {
		t.Fatal("restart bypassed sample rate limit or changed the counter")
	}
}

func TestLogSampleIgnoresExcludedLines(t *testing.T) {
	a, _, cfg := sampleTestAgent(t, "ERROR expected\nERROR unexpected\n")
	cfg.Version++
	cfg.Targets[0].Rules[0].Exclude = []Exclusion{{Mode: "literal", Pattern: "ERROR expected"}}
	if err := a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	labels := sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
	if labels["preview"] != "ERROR unexpected" || a.Count("app", "error") != 1 {
		t.Fatal("excluded line appeared in sample or counter")
	}
}

func TestLogSampleBudgetDoesNotDropCountsAndSurvivesRestart(t *testing.T) {
	a, path, cfg := sampleTestAgent(t, "ERROR first\n")
	base := cfg.Targets[0]
	cfg.Version++
	cfg.Targets = nil
	for i := 0; i < 21; i++ {
		target := base
		target.ID = fmt.Sprintf("target-%02d", i)
		cfg.Targets = append(cfg.Targets, target)
	}
	if err := a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(a.saved.Samples) != 20 || a.Count("target-20", "error") != 1 || a.saved.SampleLimitHits == 0 {
		t.Fatal("global sample budget did not preserve all match counters")
	}
	a.Close()
	b, err := New(filepath.Dir(path), a.state)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	cfg.Version++
	cfg.Targets = cfg.Targets[20:]
	if err = b.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	appendStateTestLog(t, path)
	if err = b.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(b.saved.Samples) != 0 || b.Count("target-20", "error") != 2 {
		t.Fatal("restart bypassed global sample budget or lost counts")
	}
	b.saved.SampleBudget.StartedAt = time.Now().Unix() - 301
	appendStateTestLog(t, path)
	if err = b.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(b.saved.Samples) != 1 || b.Count("target-20", "error") != 3 {
		t.Fatal("sample publication did not recover after budget window")
	}
}

func TestLogSampleInvalidBudgetCannotDisableLimit(t *testing.T) {
	state := checkpoint{Samples: map[string]logSample{}, SampleBudget: logSampleBudget{StartedAt: time.Now().Unix(), Updates: -100}}
	now := time.Now().Unix()
	pruneLogSamples(&state, now)
	for i := 0; i < 21; i++ {
		key := strconv.Itoa(i)
		if sample := newLogSample(&state, key, "app.log", "device:inode", 0, []byte("ERROR test"), now); sample != nil {
			commitLogSample(&state, key, *sample)
		}
	}
	if len(state.Samples) != 20 {
		t.Fatalf("invalid persisted budget bypassed publication limit: samples=%d", len(state.Samples))
	}
}

func TestLogSampleExpiresWithoutScrapeRefreshingItsAge(t *testing.T) {
	a, _, _ := sampleTestAgent(t, "ERROR first\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	for key, sample := range a.saved.Samples {
		sample.ObservedAt = time.Now().Unix() - 901
		a.saved.Samples[key] = sample
	}
	if strings.Contains(a.MetricsText("cluster-a", "node-a"), sampleMetricName+"{") {
		t.Fatal("scrape exported an expired sample")
	}
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(a.saved.Samples) != 0 || a.Count("app", "error") != 1 {
		t.Fatal("expired sample was not pruned independently of match counter")
	}
}

func TestLogSampleCapacityDoesNotEvictOtherActiveSamples(t *testing.T) {
	now := time.Now().Unix()
	state := checkpoint{Samples: map[string]logSample{}}
	for i := 0; i < 128; i++ {
		state.Samples[strconv.Itoa(i)] = logSample{File: "app.log", FileID: "device:inode", Preview: "ERROR old", ObservedAt: now - 301}
	}
	if sample := newLogSample(&state, "new-target", "app.log", "device:inode", 0, []byte("ERROR new"), now); sample != nil {
		t.Fatal("sample capacity was exceeded")
	}
	if sample := newLogSample(&state, "0", "app.log", "device:inode", 0, []byte("ERROR refreshed"), now); sample == nil {
		t.Fatal("capacity prevented refresh of an existing sample")
	}
	pruneLogSamples(&state, now+600)
	if len(state.Samples) != 0 {
		t.Fatal("expired samples did not release capacity")
	}
}

func TestLogSampleOldCheckpointKeepsCountersAndWaitsForNewMatch(t *testing.T) {
	a, path, cfg := sampleTestAgent(t, "ERROR historical\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	a.Close()
	raw, err := os.ReadFile(filepath.Join(a.state, "checkpoint.json"))
	if err != nil {
		t.Fatal(err)
	}
	var legacy map[string]json.RawMessage
	if err = json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	delete(legacy, "Samples")
	delete(legacy, "SampleBudget")
	delete(legacy, "SampleLimitHits")
	raw, err = json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(a.state, "checkpoint.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	b, err := New(filepath.Dir(path), a.state)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err = b.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = b.Scan(); err != nil {
		t.Fatal(err)
	}
	if b.Count("app", "error") != 1 || b.stateResets != 0 || strings.Contains(b.MetricsText("cluster-a", "node-a"), sampleMetricName+"{") {
		t.Fatal("upgrade reset state or replayed a historical sample")
	}
	appendStateTestLog(t, path)
	if err = b.Scan(); err != nil {
		t.Fatal(err)
	}
	labels := sampleMetricLabels(t, b.MetricsText("cluster-a", "node-a"))
	if b.Count("app", "error") != 2 || labels["byte_offset"] != "17" {
		t.Fatal("upgrade did not continue from the old byte cursor")
	}
}

func TestLogSampleLocationTracksRotationAndTruncation(t *testing.T) {
	a, path, _ := sampleTestAgent(t, "INFO old\nERROR first\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	first := sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
	age := func() {
		for key, sample := range a.saved.Samples {
			sample.ObservedAt = time.Now().Unix() - 301
			a.saved.Samples[key] = sample
		}
	}
	age()
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("ERROR rotated\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	rotated := sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
	if rotated["byte_offset"] != "0" || rotated["file_id"] == first["file_id"] || a.Count("app", "error") != 2 {
		t.Fatal("rotation did not preserve file identity and counts")
	}
	age()
	if err := os.WriteFile(path, []byte("ERROR x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	truncated := sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
	if truncated["preview"] != "ERROR x" || truncated["byte_offset"] != "0" || a.Count("app", "error") != 3 {
		t.Fatal("truncation did not reset sample location")
	}
}

func TestLogSampleRemovedOnScopeOrRuleChange(t *testing.T) {
	for _, change := range []string{"rule", "pod", "directory", "glob", "remove"} {
		t.Run(change, func(t *testing.T) {
			a, path, cfg := sampleTestAgent(t, "ERROR first\n")
			if err := a.Scan(); err != nil {
				t.Fatal(err)
			}
			cfg.Version++
			switch change {
			case "rule":
				cfg.Targets[0].Rules[0].Version++
			case "pod":
				cfg.Targets[0].PodUID = "pod-b"
			case "directory":
				cfg.Targets[0].Directory = filepath.Dir(path)
			case "glob":
				cfg.Targets[0].Glob = "app*.log"
			case "remove":
				cfg.Targets = nil
			}
			if err := a.Apply(cfg); err != nil {
				t.Fatal(err)
			}
			if len(a.saved.Samples) != 0 || strings.Contains(a.MetricsText("cluster-a", "node-a"), sampleMetricName+"{") {
				t.Fatal("old sample survived a scope or rule change")
			}
		})
	}
}

func TestLogSampleInvalidUTF8IsNormalized(t *testing.T) {
	a, _, _ := sampleTestAgent(t, "ERROR bad \xff byte\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	text := a.MetricsText("cluster-a", "node-a")
	labels := sampleMetricLabels(t, text)
	if !utf8.ValidString(text) || labels["preview"] != "ERROR bad \ufffd byte" {
		t.Fatal("invalid UTF-8 reached exposition output")
	}
}

func TestLogSampleInvalidUTF8FileNameKeepsMetricsValid(t *testing.T) {
	a, path, _ := sampleTestAgent(t, "ERROR first\n")
	if err := os.Rename(path, filepath.Join(filepath.Dir(path), "bad-\xff.log")); err != nil {
		t.Fatal(err)
	}
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	metrics := a.MetricsText("cluster-a", "node-a")
	if !utf8.ValidString(metrics) {
		t.Fatal("invalid UTF-8 in a file name invalidated the metrics response")
	}
	labels := sampleMetricLabels(t, metrics)
	if labels["file"] != "bad-\ufffd.log" || labels["file_id"] == "" || a.Count("app", "error") != 1 {
		t.Fatal("invalid file name lost file identity or count")
	}
}

func TestLogSampleConcurrentScanAndScrape(t *testing.T) {
	a, _, _ := sampleTestAgent(t, "ERROR first\n")
	if err := a.Scan(); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		for i := 0; i < 20; i++ {
			if err := a.Scan(); err != nil {
				t.Error(err)
			}
		}
	}()
	go func() {
		defer workers.Done()
		for i := 0; i < 20; i++ {
			sampleMetricLabels(t, a.MetricsText("cluster-a", "node-a"))
		}
	}()
	workers.Wait()
}

func TestLogSampleWaitsForCompleteLineAndCommittedCursor(t *testing.T) {
	for _, content := range []string{"ERROR incomplete", "ERROR first\n" + strings.Repeat("x", 65536) + "\n"} {
		a, _, _ := sampleTestAgent(t, content)
		if err := a.Scan(); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(a.MetricsText("cluster-a", "node-a"), sampleMetricName+"{") {
			t.Fatal("sample published before the matching line's cursor committed")
		}
	}
}

func TestLogSampleProvidedFiles(t *testing.T) {
	directory := os.Getenv("LOG_AGENT_SAMPLE_LOG_DIR")
	if directory == "" {
		t.Skip("LOG_AGENT_SAMPLE_LOG_DIR is not set")
	}
	source, err := os.ReadFile("../../../../monitor-server/services/k8s/cmdb_log_target.go")
	if err != nil {
		t.Fatal(err)
	}
	definition := regexp.MustCompile("Pattern: `([^`]+)`").FindSubmatch(source)
	if len(definition) != 2 {
		t.Fatal("default log pattern is not available")
	}
	pattern := regexp.MustCompile(string(definition[1]))
	paths, err := filepath.Glob(filepath.Join(directory, "*.log"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("sample files are not available: %v", err)
	}
	lines, matches, truncated, oversized := 0, 0, 0, 0
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		reader := bufio.NewReaderSize(file, 256*1024)
		offset := int64(0)
		for {
			line, readErr := reader.ReadBytes('\n')
			if len(line) == 0 && readErr == io.EOF {
				break
			}
			if readErr != nil && readErr != io.EOF {
				file.Close()
				t.Fatal(readErr)
			}
			lines++
			if len(line) > 65536 {
				oversized++
			}
			if pattern.Match(line) {
				matches++
				state := checkpoint{Samples: map[string]logSample{}}
				sample := newLogSample(&state, "sample", filepath.Base(path), "device:inode", offset, line, 1)
				if sample == nil || !utf8.ValidString(sample.Preview) || utf8.RuneCountInString(sample.Preview) > 1024 || sample.Offset != offset {
					file.Close()
					t.Fatalf("invalid preview or byte position in sample file %s", filepath.Base(path))
				}
				if sample.Truncated {
					truncated++
				}
			}
			offset += int64(len(line))
			if readErr == io.EOF {
				break
			}
		}
		file.Close()
	}
	if matches == 0 {
		t.Fatal("default rule did not match any provided sample lines")
	}
	t.Logf("files=%d lines=%d default_rule_matches=%d truncated_previews=%d lines_over_64KiB=%d; offline preview verification does not bypass the reader's line-size limit", len(paths), lines, matches, truncated, oversized)
}
