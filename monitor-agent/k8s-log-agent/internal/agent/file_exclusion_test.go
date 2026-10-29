package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileExclusionsSkipArchivesAndPruneCheckpointWithoutResettingCounts(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	live := filepath.Join(root, "app.log")
	archive := filepath.Join(root, "app.log.2026-01-01_08-1.log")
	for _, name := range []string{
		live, archive,
		filepath.Join(root, "app.log-archive-20260101-08.log"),
		filepath.Join(root, "app.log.2026_01_01.log"),
		filepath.Join(root, "app.log.2026.01.01.log"),
		filepath.Join(root, "app.log.worker.log"),
		filepath.Join(root, "app.log.backup.log"),
		filepath.Join(root, "service.2026-01-01_08-1.log"),
		filepath.Join(root, "server-2026-01-01-08-2.log"),
		filepath.Join(root, "admin-2026-01-01.0.log"),
		filepath.Join(root, "access.2026-01-01.log"),
		filepath.Join(root, "account.2026-01-01.log"),
		filepath.Join(root, "access_log.2026-01-01.log"),
		filepath.Join(root, "app-2026-01-01.log"),
		filepath.Join(root, "20260101-080000.log"),
	} {
		if err := os.WriteFile(name, []byte("ERROR initial\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if a != nil {
			a.Close()
		}
	}()
	cfg := stateTestConfig()
	cfg.Targets[0].Start = "beginning"
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil || a.Count("app", "error") != 15 {
		t.Fatalf("initial count=%d err=%v", a.Count("app", "error"), err)
	}
	if err = json.Unmarshal([]byte(`{"version":2,"targets":[{"id":"app","namespace":"test","podUid":"pod-a","directory":".","glob":"*.log","excludeGlobs":["*[12][0-9][0-9][0-9]-[01][0-9]-[0-3][0-9]*.log","*[12][0-9][0-9][0-9][._][01][0-9][._][0-3][0-9]*.log","*[12][0-9][0-9][0-9][01][0-9][0-3][0-9]*.log"],"start":"beginning","rules":[{"id":"error","version":1,"mode":"literal","pattern":"ERROR"}]}]}`), &cfg); err != nil {
		t.Fatal(err)
	}
	// An excluded symlink must not be opened or block cursor cleanup.
	if err = os.Symlink("missing", filepath.Join(root, "blocked.log.2026-01-01.log")); err != nil {
		t.Fatal(err)
	}
	appendStateTestLog(t, live)
	appendStateTestLog(t, archive)
	appendStateTestLog(t, filepath.Join(root, "access_log.2026-01-01.log"))
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if got := a.Count("app", "error"); got != 16 {
		t.Fatalf("archive must be skipped and historical count retained: %d", got)
	}
	if len(a.saved.Files) != 3 || len(a.saved.Paths) != 3 || a.readErrors["app"] != 0 || !a.healthy["app"] {
		t.Fatalf("excluded state retained or file opened: files=%d paths=%d errors=%d healthy=%t", len(a.saved.Files), len(a.saved.Paths), a.readErrors["app"], a.healthy["app"])
	}
	a.Close()
	a = nil
	a, err = New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil || a.Count("app", "error") != 16 {
		t.Fatalf("restart lost counts or replayed data: count=%d err=%v", a.Count("app", "error"), err)
	}
}

func TestArchivedFilesDoNotExhaustMatchedFileLimit(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 10001; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("app.log.2026-01-01_%d.log", i)), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "app.log"), []byte("ERROR live\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app_metric.log"), []byte(metricLine(fmt.Sprintf(`{"name":"jvm.memory.used","unit":"By","gauge":{"dataPoints":[{"timeUnixNano":"%d","asInt":"42"}]}}`, time.Now().UnixNano()))+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	var cfg Config
	if err = json.Unmarshal([]byte(`{"version":1,"targets":[{"id":"app","namespace":"test","podUid":"pod-a","directory":".","glob":"*.log","excludeGlobs":["*.log*[12][0-9][0-9][0-9]-[01][0-9]-[0-3][0-9]*.log","*.log*[12][0-9][0-9][0-9][._][01][0-9][._][0-3][0-9]*.log","*.log*[12][0-9][0-9][0-9][01][0-9][0-3][0-9]*.log"],"start":"beginning","rules":[{"id":"error","version":1,"mode":"literal","pattern":"ERROR"}]},{"id":"metric","namespace":"test","podUid":"pod-a","directory":".","glob":"*_metric.log","start":"end","format":"otlp_metric"}]}`), &cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil || a.Count("app", "error") != 1 || !a.healthy["app"] || !a.healthy["metric"] || len(a.appMetrics) != 1 {
		t.Fatalf("archives blocked live logs or metrics: count=%d healthy=%v metrics=%d err=%v", a.Count("app", "error"), a.healthy, len(a.appMetrics), err)
	}
	// The bound still applies when more than 10000 files actually match.
	cfg.Version++
	cfg.Targets[0].ExcludeGlobs = nil
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil || a.healthy["app"] || a.readErrors["app"] != 1 || a.Count("app", "error") != 1 {
		t.Fatalf("matching file limit was not enforced safely: healthy=%t errors=%d count=%d err=%v", a.healthy["app"], a.readErrors["app"], a.Count("app", "error"), err)
	}
}

func TestAgentRejectsInvalidFileExclusions(t *testing.T) {
	a, err := New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	for _, pattern := range []string{"", "[", "../*.log", "nested/*.log", `nested\*.log`} {
		t.Run(pattern, func(t *testing.T) {
			exclusions, _ := json.Marshal([]string{pattern})
			var cfg Config
			if err := json.Unmarshal([]byte(fmt.Sprintf(`{"version":1,"targets":[{"id":"app","namespace":"test","podUid":"pod-a","directory":".","glob":"*.log","excludeGlobs":%s,"start":"end"}]}`, exclusions)), &cfg); err != nil {
				t.Fatal(err)
			}
			if err := a.Apply(cfg); err == nil || a.Version() != 0 {
				t.Fatal("invalid file exclusion activated")
			}
		})
	}
}
