package prom

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

// WriteLogAgentTargets atomically replaces the dedicated file_sd file. The
// caller supplies addresses discovered from deployment configuration; this
// function does not infer node addresses from names.
func WriteLogAgentTargets(directory string, addresses []string) error {
	unique := map[string]bool{}
	for _, address := range addresses {
		if address == "" {
			return errors.New("empty log agent address")
		}
		unique[address] = true
	}
	ordered := make([]string, 0, len(unique))
	for address := range unique {
		ordered = append(ordered, address)
	}
	sort.Strings(ordered)
	targets := []map[string]interface{}{}
	if len(ordered) > 0 {
		targets = append(targets, map[string]interface{}{"targets": ordered})
	}
	raw, err := json.Marshal(targets)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(directory, 0750); err != nil {
		return err
	}
	name := filepath.Join(directory, "k8s_log_agents.json")
	if existing, readErr := os.ReadFile(name); readErr == nil && string(existing) == string(raw) {
		return nil
	}
	f, err := os.CreateTemp(directory, ".k8s-log-agents-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(raw); err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, name)
}
