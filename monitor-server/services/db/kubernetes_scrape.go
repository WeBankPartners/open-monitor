package db

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"go.uber.org/zap"
)

func disableLegacyKubernetesScrape(config, script, promtool, defaults string, reload func() error) error {
	// Shared with SNMP, remote-write and custom scrape config writers.
	prometheusConfigMutex.Lock()
	defer prometheusConfigMutex.Unlock()
	original, err := os.ReadFile(config)
	if err != nil {
		return fmt.Errorf("read legacy Kubernetes scrape config: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", script, config, promtool, defaults)
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	if err = cmd.Run(); err != nil {
		// The helper suppresses promtool output and emits only safe stage errors.
		return fmt.Errorf("disable legacy Kubernetes scrape config: %w: %s", err, strings.TrimSpace(diagnostic.String()))
	}
	// Reload even when startup has already cleaned the disk. A previous process
	// may have exited between replacing the file and reloading Prometheus.
	if err = reload(); err != nil {
		current, readErr := os.ReadFile(config)
		if readErr != nil {
			return fmt.Errorf("disable legacy Kubernetes scrape reload failed: %w; read rollback config: %v", err, readErr)
		}
		if !bytes.Equal(original, current) {
			if restoreErr := restoreLegacyScrapeConfig(config, original); restoreErr != nil {
				return fmt.Errorf("disable legacy Kubernetes scrape reload failed: %w; restore config: %v", err, restoreErr)
			}
			if restoreErr := reload(); restoreErr != nil {
				return fmt.Errorf("disable legacy Kubernetes scrape reload failed: %w; reload restored config: %v", err, restoreErr)
			}
		}
		return fmt.Errorf("disable legacy Kubernetes scrape reload failed: %w", err)
	}
	log.Info(nil, log.LOGGER_APP, "Legacy Kubernetes scrape disabled and Prometheus reload confirmed",
		zap.Strings("job_prefixes", []string{"k8s-kubelet-", "k8s-cadvisor-"}))
	return nil
}

func restoreLegacyScrapeConfig(config string, content []byte) error {
	file, err := os.CreateTemp(filepath.Dir(config), ".legacy-k8s-rollback-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(content); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), config)
}
