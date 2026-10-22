package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckpointRestartAndPartialLine(t *testing.T) {
	root := t.TempDir()
	state := t.TempDir()
	file := filepath.Join(root, "app.log")
	os.WriteFile(file, []byte("ERROR first\nERR"), 0600)
	config := Config{Version: 1, Targets: []Target{{ID: "app", Namespace: "sit", PodUID: "pod-a", Directory: ".", Glob: "*.log", Start: "beginning", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Apply(config); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 1 {
		t.Fatal("missing match")
	}
	a.Close()
	f, _ := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("OR second\n")
	f.Close()
	a, err = New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Apply(config); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 2 {
		t.Fatalf("restart count %v", a.Count("app", "error"))
	}
}
func TestInvalidConfigKeepsPreviousAndLocksState(t *testing.T) {
	root := t.TempDir()
	state := t.TempDir()
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if b, err := New(root, state); err == nil {
		b.Close()
		t.Fatal("duplicate agent owns checkpoint")
	}
	bad := Config{Version: 1, Targets: []Target{{ID: "escape", Directory: "../other", Glob: "*.log"}}}
	if a.Apply(bad) == nil {
		t.Fatal("path escape accepted")
	}
	if a.Version() != 0 {
		t.Fatal("invalid config activated")
	}
}

func TestRotatedFileStartsAtBeginning(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	file := filepath.Join(root, "app.log")
	os.WriteFile(file, []byte("ERROR old\n"), 0600)
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	c := Config{Version: 1, Targets: []Target{{ID: "app", Namespace: "sit", PodUID: "pod-a", Directory: ".", Glob: "*.log", Start: "end", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	os.Rename(file, file+".1")
	os.WriteFile(file, []byte("ERROR new\n"), 0600)
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 1 {
		t.Fatal("new file skipped")
	}
}

func TestMetricsExposeLastMatchAndErrors(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(root, "app.log"), []byte("ERROR\n"), 0600)
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	c := Config{Version: 1, Targets: []Target{{ID: "app", Namespace: "sit", PodUID: "pod-a", Directory: ".", Glob: "*.log", Start: "beginning", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	out := a.MetricsText("cluster", "node")
	if !strings.Contains(out, "k8s_log_last_match_timestamp") || !strings.Contains(out, "k8s_log_file_missing") {
		t.Fatal(out)
	}
}

func TestAbsoluteDirectoryAndIdentityLabels(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	logDir := filepath.Join(root, "var", "log")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(logDir, "app.log"), []byte("ERROR\n"), 0600)
	a, err := New("/", state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	// The test path is absolute only to exercise the same contract as CMDB deploy_path.
	path := filepath.Join(root, "var", "log")
	c := Config{Version: 1, Targets: []Target{{ID: "app", Namespace: "sit", PodUID: "pod-a", PodName: "orders-0", InstanceID: "app_instance-a", Container: "app", WorkloadName: "orders", Directory: path, Glob: "*.log", Start: "beginning", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	out := a.MetricsText("cluster", "node")
	if !strings.Contains(out, `instance_id="app_instance-a"`) || !strings.Contains(out, `pod="orders-0"`) {
		t.Fatal(out)
	}
}

func TestConfigRemovalDropsOldCursor(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(root, "app.log"), []byte("ERROR\n"), 0600)
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	c := Config{Version: 1, Targets: []Target{{ID: "app", Namespace: "sit", PodUID: "pod-a", Directory: ".", Glob: "*.log", Start: "beginning", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	a.Apply(c)
	a.Scan()
	if a.Count("app", "error") != 1 {
		t.Fatal("initial cursor missing")
	}
	if err = a.Apply(Config{Version: 2, Targets: []Target{}}); err != nil {
		t.Fatal(err)
	}
	if len(a.saved.Files) != 0 {
		t.Fatal("removed target cursor retained")
	}
}
