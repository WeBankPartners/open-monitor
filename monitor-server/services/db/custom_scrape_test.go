package db

import (
	"strings"
	"testing"
)

func TestNormalizeCustomScrapeYamlList(t *testing.T) {
	content := `- job_name: 'CLUSTER-CORE'
  metrics_path: '/assemble'
  static_configs:
    - targets: ['10.0.128.162:19997']
  relabel_configs:
    - source_labels: [job]
      target_label: __param_clusterName
`
	fileContent, jobNames, err := normalizeCustomScrapeYaml(content)
	if err != nil {
		t.Fatalf("normalize fail: %v", err)
	}
	if len(jobNames) != 1 || jobNames[0] != "CLUSTER-CORE" {
		t.Fatalf("unexpected job names: %v", jobNames)
	}
	if fileContent == "" {
		t.Fatal("file content empty")
	}
}

func TestNormalizeCustomScrapeYamlSingleJob(t *testing.T) {
	content := `job_name: demo
metrics_path: /metrics
static_configs:
  - targets: ['127.0.0.1:9100']
`
	_, jobNames, err := normalizeCustomScrapeYaml(content)
	if err != nil {
		t.Fatalf("normalize fail: %v", err)
	}
	if len(jobNames) != 1 || jobNames[0] != "demo" {
		t.Fatalf("unexpected job names: %v", jobNames)
	}
}

func TestNormalizeCustomScrapeYamlRejectFullConfig(t *testing.T) {
	content := `global:
  scrape_interval: 10s
scrape_configs:
  - job_name: demo
    static_configs:
      - targets: ['127.0.0.1:9100']
`
	_, _, err := normalizeCustomScrapeYaml(content)
	if err == nil {
		t.Fatal("expected reject full prometheus.yml")
	}
}

func TestNormalizeCustomScrapeYamlReservedJob(t *testing.T) {
	content := `- job_name: prometheus
  static_configs:
    - targets: ['127.0.0.1:9090']
`
	_, _, err := normalizeCustomScrapeYaml(content)
	if err == nil {
		t.Fatal("expected reserved job_name error")
	}
}

func TestInsertCustomScrapeConfigFiles(t *testing.T) {
	src := "scrape_configs:\n  - job_name: prometheus\n\nremote_write:\n"
	out := insertCustomScrapeConfigFiles(src, "/tmp/custom/*.yml")
	if !strings.Contains(out, "scrape_config_files:") {
		t.Fatalf("missing scrape_config_files:\n%s", out)
	}
	if !strings.Contains(out, "/tmp/custom/*.yml") {
		t.Fatalf("missing glob:\n%s", out)
	}
	if !strings.Contains(out, "remote_write:") {
		t.Fatalf("missing remote_write:\n%s", out)
	}
}
