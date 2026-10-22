package agent

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func stateTestConfig() Config {
	return Config{Version: 1, Targets: []Target{{ID: "app", Namespace: "test", PodUID: "pod-a", Directory: ".", Glob: "*.log", Start: "end", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
}

func appendStateTestLog(t *testing.T, path string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = f.WriteString("ERROR new\n"); err != nil {
		t.Fatal(err)
	}
}

func TestCorruptCheckpointRestartsAtEnd(t *testing.T) {
	for _, invalid := range []string{`{`, `null`, `{"Files":{"bad":{"Offset":-1,"Counts":{}}}}`, `{"Files":{"bad":{"Offset":0}}}`} {
		t.Run(invalid, func(t *testing.T) {
			root, state := t.TempDir(), t.TempDir()
			file := filepath.Join(root, "app.log")
			os.WriteFile(file, []byte("ERROR old\n"), 0600)
			os.WriteFile(filepath.Join(state, "checkpoint.json"), []byte(invalid), 0600)
			a, err := New(root, state)
			if err != nil {
				t.Fatalf("corrupt state must not prevent startup: %v", err)
			}
			defer a.Close()
			if err = a.Apply(stateTestConfig()); err != nil {
				t.Fatal(err)
			}
			if err = a.Scan(); err != nil || a.Count("app", "error") != 0 {
				t.Fatalf("recovery must skip old logs: %v", err)
			}
			appendStateTestLog(t, file)
			if err = a.Scan(); err != nil || a.Count("app", "error") != 1 {
				t.Fatalf("new logs must be counted after recovery: %v", err)
			}
			if !strings.Contains(a.MetricsText("cluster", "node"), "k8s_log_agent_state_resets_total{cluster_id=\"cluster\",node=\"node\"} 1") {
				t.Fatal("state reset must be observable")
			}
		})
	}
}

func TestUnavailableStateDoesNotStopScanning(t *testing.T) {
	root, parent := t.TempDir(), t.TempDir()
	state := filepath.Join(parent, "unavailable")
	os.WriteFile(state, []byte("not a directory"), 0600)
	file := filepath.Join(root, "app.log")
	os.WriteFile(file, []byte("ERROR old\n"), 0600)
	a, err := New(root, state)
	if err != nil {
		t.Fatalf("state access failure must allow memory mode: %v", err)
	}
	defer a.Close()
	a.Apply(stateTestConfig())
	if err = a.Scan(); err != nil || a.Count("app", "error") != 0 {
		t.Fatalf("memory mode must start at end: %v", err)
	}
	appendStateTestLog(t, file)
	if err = a.Scan(); err != nil || a.Count("app", "error") != 1 {
		t.Fatalf("state failure stopped collection: %v", err)
	}
	if err = a.Scan(); err != nil || a.Count("app", "error") != 1 {
		t.Fatal("failed checkpoint must not replay logs")
	}
	if err = os.Remove(state); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(state, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(state, "checkpoint.json"), []byte(`{"Files":{},"Paths":{},"Config":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	a.Close()
	b, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	b.Apply(stateTestConfig())
	if err = b.Scan(); err != nil || b.Count("app", "error") != 1 {
		t.Fatalf("recovered storage must save the current memory state: %v", err)
	}
}

func TestPersistenceFailureKeepsMemoryCounters(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	file := filepath.Join(root, "app.log")
	os.WriteFile(file, []byte("ERROR first\n"), 0600)
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	cfg := stateTestConfig()
	cfg.Targets[0].Start = "beginning"
	a.Apply(cfg)
	// An existing directory at the checkpoint destination makes rename fail.
	os.Mkdir(filepath.Join(state, "checkpoint.json"), 0700)
	if err = a.Scan(); err != nil || a.Count("app", "error") != 1 {
		t.Fatalf("checkpoint failure must not fail a scan: %v", err)
	}
	appendStateTestLog(t, file)
	if err = a.Scan(); err != nil || a.Count("app", "error") != 2 {
		t.Fatal("memory cursor must advance while persistence fails")
	}
	if !strings.Contains(a.MetricsText("cluster", "node"), "k8s_log_agent_state_persist_errors_total{cluster_id=\"cluster\",node=\"node\"} 2") {
		t.Fatal("persistence errors must be observable")
	}
}

func TestRotationPrunesOldInodeCursorWithoutResettingCounters(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	file := filepath.Join(root, "app.log")
	if err := os.WriteFile(file, []byte("ERROR first\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	cfg := stateTestConfig()
	cfg.Targets[0].Start = "beginning"
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil || a.Count("app", "error") != 1 {
		t.Fatalf("initial scan failed: %v", err)
	}
	for i := 1; i <= 3; i++ {
		rotated := filepath.Join(root, "app.log-rotation-"+strconv.Itoa(i))
		if err = os.Rename(file, rotated); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(file, []byte("ERROR next\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err = a.Scan(); err != nil || a.Count("app", "error") != uint64(i+1) {
			t.Fatalf("counter must include each rotated file exactly once: iteration=%d err=%v count=%d", i, err, a.Count("app", "error"))
		}
		if len(a.saved.Files) != 1 || len(a.saved.Paths) != 1 || len(a.saved.Totals) != 1 {
			t.Fatalf("checkpoint must retain only active file and fixed rule totals, got files=%d paths=%d totals=%d", len(a.saved.Files), len(a.saved.Paths), len(a.saved.Totals))
		}
	}
}

func TestUnavailableTargetDirectoryDoesNotPruneCursors(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	logDir := filepath.Join(root, "logs")
	if err := os.Mkdir(logDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "app.log"), []byte("ERROR retained\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	cfg := stateTestConfig()
	cfg.Targets[0].Directory = "logs"
	cfg.Targets[0].Start = "beginning"
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(logDir, filepath.Join(root, "logs-unavailable")); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if len(a.saved.Files) != 1 || a.Count("app", "error") != 1 {
		t.Fatal("unavailable directory must not delete saved cursor or cumulative count")
	}
}
