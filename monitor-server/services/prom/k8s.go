package prom

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// WriteKubernetesTarget owns a dedicated file; legacy sd_file_10.json is untouched.
func WriteKubernetesTarget(directory, port string, enabled bool) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return errors.New("invalid Monitor HTTP port")
	}
	targets := []map[string]interface{}{}
	if enabled {
		targets = append(targets, map[string]interface{}{"targets": []string{net.JoinHostPort("localhost", port)}, "labels": map[string]string{"__metrics_path__": "/monitor/internal/k8s/metrics"}})
	}
	raw, err := json.Marshal(targets)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(directory, 0750); err != nil {
		return err
	}
	name := filepath.Join(directory, "k8s_monitor.json")
	if existing, err := os.ReadFile(name); err == nil && strings.TrimSpace(string(existing)) == string(raw) {
		return nil
	}
	f, err := os.CreateTemp(directory, ".k8s-target-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), name)
}
