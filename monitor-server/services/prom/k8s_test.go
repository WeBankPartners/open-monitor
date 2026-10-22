package prom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestKubernetesTargetTogglePreservesLegacyTargets(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "sd_file_10.json")
	os.WriteFile(legacy, []byte("legacy"), 0600)
	if err := WriteKubernetesTarget(dir, "8080", true); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "k8s_monitor.json"))
	var targets []struct {
		Targets []string          `json:"targets"`
		Labels  map[string]string `json:"labels"`
	}
	if json.Unmarshal(raw, &targets) != nil || len(targets) != 1 || targets[0].Targets[0] != "localhost:8080" || targets[0].Labels["__metrics_path__"] != "/monitor/internal/k8s/metrics" {
		t.Fatalf("bad target %s", raw)
	}
	if err := WriteKubernetesTarget(dir, "8080", false); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(filepath.Join(dir, "k8s_monitor.json"))
	if string(raw) != "[]" {
		t.Fatalf("disabled target remains: %s", raw)
	}
	raw, _ = os.ReadFile(legacy)
	if string(raw) != "legacy" {
		t.Fatal("legacy targets overwritten")
	}
}
