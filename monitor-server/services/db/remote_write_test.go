package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"gopkg.in/yaml.v2"
)

func TestWriteRemoteWriteConfigAtomicallyKeepsOriginalOnFailure(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "prometheus.yml")
	if err := os.WriteFile(config, []byte("old config"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeRemoteWriteConfig(config, []byte("new config")); err != nil {
		t.Fatal(err)
	}
	if content, err := os.ReadFile(config); err != nil || string(content) != "new config" {
		t.Fatalf("new config was not installed: %v", err)
	}
	if err := writeRemoteWriteConfig(dir, []byte("invalid replacement")); err == nil {
		t.Fatal("replacing a directory unexpectedly succeeded")
	}
	if content, err := os.ReadFile(config); err != nil || string(content) != "new config" {
		t.Fatalf("failed replacement damaged existing config: %v", err)
	}
	if candidates, _ := filepath.Glob(filepath.Join(dir, ".remote-write-*")); len(candidates) != 0 {
		t.Fatal("temporary files were not cleaned up")
	}
}

func TestRenderRemoteWriteConfigDeduplicatesExistingRows(t *testing.T) {
	const prefix = "scrape_configs:\n  - job_name: k8s_monitor\nremote_write:\n#Remote_write_start\n"
	const suffix = "#Remote_write_end\n# unrelated config\n"
	rows := []*models.RemoteWriteConfigTable{
		{Id: "config-a", Address: "https://receiver.example.invalid/write"},
		{Id: "config-b", Address: "https://receiver.example.invalid/write"},
		{Id: "config-c", Address: " https://receiver.example.invalid/write "},
		{Id: "config-d", Address: "  "},
		{Id: "config-e", Address: "https://other.example.invalid/write"},
	}
	config, duplicates, err := renderRemoteWriteConfig(prefix+"  - url: obsolete\n"+suffix, "  - url: '{{remote_write_url}}'", rows)
	if err != nil {
		t.Fatal(err)
	}
	if duplicates != 2 || strings.Count(config, "  - url:") != 2 {
		t.Fatalf("unexpected render result: duplicates=%d", duplicates)
	}
	if !strings.HasPrefix(config, prefix) || !strings.HasSuffix(config, suffix) || strings.Contains(config, "obsolete") {
		t.Fatal("renderer changed unrelated sections or retained old entries")
	}
	var parsed struct {
		RemoteWrite []struct {
			URL string `yaml:"url"`
		} `yaml:"remote_write"`
	}
	if err := yaml.Unmarshal([]byte(config), &parsed); err != nil || len(parsed.RemoteWrite) != 2 {
		t.Fatalf("rendered config is not valid: %v", err)
	}
}

func TestRenderRemoteWriteConfigEscapesQuotedAddress(t *testing.T) {
	address := "https://receiver.example.invalid/write?tenant=team'example"
	config, _, err := renderRemoteWriteConfig("remote_write:\n#Remote_write_start\n#Remote_write_end\n", "  - url: '{{remote_write_url}}'", []*models.RemoteWriteConfigTable{{Address: address}})
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		RemoteWrite []struct {
			URL string `yaml:"url"`
		} `yaml:"remote_write"`
	}
	if err := yaml.Unmarshal([]byte(config), &parsed); err != nil || len(parsed.RemoteWrite) != 1 || parsed.RemoteWrite[0].URL != address {
		t.Fatalf("address did not round trip: %v", err)
	}
}

func TestRenderRemoteWriteConfigRejectsBrokenMarkers(t *testing.T) {
	for _, content := range []string{"remote_write: []\n", "#Remote_write_end\n#Remote_write_start\n", "#Remote_write_start\n#Remote_write_start\n#Remote_write_end\n"} {
		if _, _, err := renderRemoteWriteConfig(content, "  - url: '{{remote_write_url}}'", nil); err == nil {
			t.Fatal("invalid managed region accepted")
		}
	}
}
