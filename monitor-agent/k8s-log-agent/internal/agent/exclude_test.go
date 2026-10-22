package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func exclusionConfig(t *testing.T, exclusions string) Config {
	t.Helper()
	var c Config
	err := json.Unmarshal([]byte(`{"version":1,"targets":[{"id":"app","namespace":"sit","podUid":"pod-a","directory":".","glob":"*.log","start":"beginning","rules":[{"id":"error","version":1,"mode":"regex","pattern":"(?i)\\b(ERROR|FATAL|PANIC)\\b","exclude":`+exclusions+`},{"id":"all-errors","version":1,"mode":"literal","pattern":"ERROR"}]}]}`), &c)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestExclusionsFilterOnlyTheirRuleAndPersistCursor(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	path := filepath.Join(root, "app.log")
	if err := os.WriteFile(path, []byte("ERROR real\nERROR EXPECTED_ERROR\nfatal healthcheck timeout\nERROR expected_error\nINFO ignored\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c := exclusionConfig(t, `[{"mode":"literal","pattern":"EXPECTED_ERROR"},{"mode":"regex","pattern":"(?i)healthcheck.*timeout"}]`)
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 2 || a.Count("app", "all-errors") != 3 {
		t.Fatalf("exclusions leaked or ignored: %s", a.MetricsText("cluster-a", "node-a"))
	}
	a.Close()
	a, err = New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 2 {
		t.Fatal("excluded lines replayed after restart")
	}
	if strings.Contains(a.MetricsText("cluster-a", "node-a"), "EXPECTED_ERROR") {
		t.Fatal("exclusion text exposed as a label")
	}
}

func TestInvalidExclusionKeepsAgentConfiguration(t *testing.T) {
	for _, exclude := range []string{`[{"mode":"regex","pattern":"["}]`, `[{"mode":"literal","pattern":""}]`, `[{"mode":"unknown","pattern":"ERROR"}]`} {
		a, err := New(t.TempDir(), t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		if err = a.Apply(exclusionConfig(t, `[]`)); err != nil {
			t.Fatal(err)
		}
		invalid := exclusionConfig(t, exclude)
		invalid.Version = 2
		if err = a.Apply(invalid); err == nil {
			t.Errorf("invalid exclusion accepted: %s", exclude)
		}
		if a.Version() != 1 {
			t.Error("invalid configuration activated")
		}
		a.Close()
	}
}

func TestExclusionChangeStartsNewCounterAtExistingReadOffset(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	path := filepath.Join(root, "app.log")
	if err := os.WriteFile(path, []byte("ERROR old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	c := exclusionConfig(t, `[]`)
	c.Targets[0].Start = "end"
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("ERROR EXPECTED_ERROR\nERROR new\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	c = exclusionConfig(t, `[{"mode":"literal","pattern":"EXPECTED_ERROR"}]`)
	c.Version = 2
	c.Targets[0].Start = "end"
	c.Targets[0].Rules[0].Version = 2
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 1 {
		t.Fatal("rule update skipped unread lines or replayed old lines")
	}
	if len(a.saved.Paths) != 2 {
		t.Fatal("obsolete rule version paths retained")
	}
}

func TestExclusionUpdateAfterRestartRetainsCommittedOffset(t *testing.T) {
	root, state := t.TempDir(), t.TempDir()
	path := filepath.Join(root, "app.log")
	if err := os.WriteFile(path, []byte("ERROR old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	c := exclusionConfig(t, `[]`)
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	a.Close()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("ERROR EXPECTED_ERROR\nERROR new\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	a, err = New(root, state)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	c = exclusionConfig(t, `[{"mode":"literal","pattern":"EXPECTED_ERROR"}]`)
	c.Version = 2
	c.Targets[0].Start = "end"
	c.Targets[0].Rules[0].Version = 2
	if err = a.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	if a.Count("app", "error") != 1 {
		t.Fatal("rule update after restart lost the committed offset")
	}
}
