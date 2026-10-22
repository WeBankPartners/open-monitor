package prom

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteLogAgentTargetsDeduplicatesAndReplacesAtomically(t *testing.T) {
	dir := t.TempDir()
	if err := WriteLogAgentTargets(dir, []string{"agent-b:9105", "agent-a:9105", "agent-a:9105"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "k8s_log_agents.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `[{"targets":["agent-a:9105","agent-b:9105"]}]` {
		t.Fatalf("%s", raw)
	}
}
