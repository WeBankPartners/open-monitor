package db

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const legacyScrapeFixture = `scrape_configs:
  - job_name: k8s_monitor
#Kubernetes_start
  - job_name: k8s-cadvisor-cluster-a
    metrics_path: /metrics/cadvisor
#Kubernetes_end
#Snmp_start
  - job_name: snmp-example
#Snmp_end
remote_write: []
`

func legacyScrapeFiles(t *testing.T, valid bool) (string, string, string, string) {
	t.Helper()
	dir := t.TempDir()
	config := filepath.Join(dir, "prometheus.yml")
	if err := os.WriteFile(config, []byte(legacyScrapeFixture), 0600); err != nil {
		t.Fatal(err)
	}
	promtool := filepath.Join(dir, "promtool")
	exitCode := "0"
	if !valid {
		exitCode = "1"
	}
	if err := os.WriteFile(promtool, []byte("#!/bin/sh\nexit "+exitCode+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("../../../build/disable-legacy-k8s-scrape.sh")
	if err != nil {
		t.Fatal(err)
	}
	defaults, err := filepath.Abs("../../../build/conf/prometheus.yml")
	if err != nil {
		t.Fatal(err)
	}
	return config, script, promtool, defaults
}

func TestDisableLegacyKubernetesScrapeReloadsCleanConfig(t *testing.T) {
	config, script, promtool, defaults := legacyScrapeFiles(t, true)
	reloads := 0
	reload := func() error {
		reloads++
		content, err := os.ReadFile(config)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), "k8s-cadvisor-cluster-a") {
			t.Error("reload still sees the retired job")
		}
		if !strings.Contains(string(content), "snmp-example") || !strings.Contains(string(content), "k8s_monitor") {
			t.Error("cleanup removed unrelated jobs")
		}
		return nil
	}
	for i := 0; i < 2; i++ {
		if err := disableLegacyKubernetesScrape(config, script, promtool, defaults, reload); err != nil {
			t.Fatal(err)
		}
	}
	if reloads != 2 {
		t.Fatalf("want reload on startup even if disk was already cleaned, got %d", reloads)
	}
	backup, err := os.ReadFile(config + ".legacy-k8s.bak")
	if err != nil || string(backup) != legacyScrapeFixture {
		t.Fatalf("original backup was not retained: %v", err)
	}
}

func TestDisableLegacyKubernetesScrapeRestoresOnReloadFailure(t *testing.T) {
	config, script, promtool, defaults := legacyScrapeFiles(t, true)
	reloads := 0
	err := disableLegacyKubernetesScrape(config, script, promtool, defaults, func() error {
		reloads++
		if reloads == 1 {
			return errors.New("reload rejected")
		}
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "reload rejected") {
		t.Fatalf("reload error lost: %v", err)
	}
	content, readErr := os.ReadFile(config)
	if readErr != nil || string(content) != legacyScrapeFixture || reloads != 2 {
		t.Fatalf("failed reload did not restore original config: reloads=%d err=%v", reloads, readErr)
	}
	if err := disableLegacyKubernetesScrape(config, script, promtool, defaults, func() error { return nil }); err != nil {
		t.Fatalf("cleanup could not recover after retry: %v", err)
	}
}

func TestDisableLegacyKubernetesScrapeRejectsInvalidConfigBeforeReload(t *testing.T) {
	config, script, promtool, defaults := legacyScrapeFiles(t, false)
	err := disableLegacyKubernetesScrape(config, script, promtool, defaults, func() error {
		t.Fatal("invalid config reached reload")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "config validation") {
		t.Fatalf("invalid config did not preserve a diagnostic reason: %v", err)
	}
	content, readErr := os.ReadFile(config)
	if readErr != nil || string(content) != legacyScrapeFixture {
		t.Fatalf("validation failure changed config: %v", readErr)
	}
}

func TestDisableLegacyKubernetesScrapeSerializesConcurrentCalls(t *testing.T) {
	config, script, promtool, defaults := legacyScrapeFiles(t, true)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := disableLegacyKubernetesScrape(config, script, promtool, defaults, func() error { return nil }); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	content, err := os.ReadFile(config)
	if err != nil || strings.Contains(string(content), "k8s-cadvisor-cluster-a") {
		t.Fatalf("concurrent cleanup left the old job: %v", err)
	}
}
