package k8s

import (
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"regexp"
	"strings"
)

type LogAgentEndpoint struct {
	ReportedAddress string `json:"reportedAddress"`
	ReportedPort    int    `json:"reportedPort"`
	ClusterID       string `json:"clusterId"`
	Node            string `json:"node"`
	Address         string `json:"address"`
	Port            int    `json:"port"`
	PodUID          string `json:"podUid,omitempty"`
	LastSeen        int64  `json:"lastSeen"`
	Enabled         bool   `json:"enabled"`
	ManualOverride  bool   `json:"manualOverride"`
	Online          bool   `json:"online"`
	ConfigVersion   uint64 `json:"configVersion"`
	TargetCount     int    `json:"targetCount"`
}

var nodeNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

func ValidateLogAgentEndpoint(e LogAgentEndpoint) error {
	if !clusterPattern.MatchString(e.ClusterID) || len(e.ClusterID) > 64 || len(e.Node) > 253 || !nodeNamePattern.MatchString(e.Node) || net.ParseIP(e.Address) == nil || e.Port < 1 || e.Port > 65535 {
		return errors.New("invalid log agent endpoint")
	}
	return nil
}

type LogRule struct {
	ID      string         `json:"id"`
	Version uint64         `json:"version"`
	Mode    string         `json:"mode"`
	Pattern string         `json:"pattern"`
	Exclude []LogExclusion `json:"exclude,omitempty"`
}
type LogExclusion struct {
	Mode    string `json:"mode"`
	Pattern string `json:"pattern"`
}
type LogTarget struct {
	ID           string    `json:"id"`
	Namespace    string    `json:"namespace"`
	WorkloadType string    `json:"workloadType"`
	WorkloadName string    `json:"workloadName"`
	WorkloadUID  string    `json:"workloadUid"`
	PodName      string    `json:"podName"`
	PodUID       string    `json:"podUid"`
	Container    string    `json:"container"`
	InstanceID   string    `json:"instanceId"`
	Directory    string    `json:"directory"`
	Glob         string    `json:"glob"`
	ExcludeGlobs []string  `json:"excludeGlobs,omitempty"`
	Start        string    `json:"start"`
	Rules        []LogRule `json:"rules"`
	Format       string    `json:"format,omitempty"`
}
type LogAgentConfig struct {
	Version uint64      `json:"version"`
	Targets []LogTarget `json:"targets"`
}

func ValidateLogAgentConfig(c LogAgentConfig) error {
	if c.Version == 0 || len(c.Targets) > 1000 {
		return errors.New("invalid log agent configuration")
	}
	ids := map[string]bool{}
	for _, t := range c.Targets {
		if t.ID == "" || ids[t.ID] || t.Namespace == "" || t.PodUID == "" || t.Directory == "" || strings.Contains(t.Directory, "..") || t.Glob == "" || (t.Start != "beginning" && t.Start != "end") || (t.Format != "" && t.Format != "otlp_metric") || (t.Format == "otlp_metric" && len(t.Rules) > 0) {
			return errors.New("invalid log target")
		}
		ids[t.ID] = true
		if len(t.ExcludeGlobs) > 100 {
			return errors.New("too many log file exclusions")
		}
		for _, pattern := range t.ExcludeGlobs {
			if pattern == "" || len(pattern) > 4096 || strings.ContainsAny(pattern, "/\\") {
				return errors.New("invalid log file exclusion")
			}
			if _, err := filepath.Match(pattern, ""); err != nil {
				return errors.New("invalid log file exclusion glob")
			}
		}
		if len(t.Rules) > 100 {
			return errors.New("too many log rules")
		}
		rules := map[string]bool{}
		for _, r := range t.Rules {
			if r.ID == "" || rules[r.ID] || r.Version == 0 || r.Pattern == "" || len(r.Pattern) > 4096 {
				return errors.New("invalid log rule")
			}
			rules[r.ID] = true
			if r.Mode != "literal" && r.Mode != "regex" {
				return errors.New("invalid log rule mode")
			}
			if r.Mode == "regex" {
				if _, err := regexp.Compile(r.Pattern); err != nil {
					return errors.New("invalid log rule regex")
				}
			}
			if len(r.Exclude) > 100 {
				return errors.New("too many log rule exclusions")
			}
			for _, exclude := range r.Exclude {
				if exclude.Pattern == "" || len(exclude.Pattern) > 4096 || (exclude.Mode != "literal" && exclude.Mode != "regex") {
					return errors.New("invalid log rule exclusion")
				}
				if exclude.Mode == "regex" {
					if _, err := regexp.Compile(exclude.Pattern); err != nil {
						return errors.New("invalid log rule exclusion regex")
					}
				}
			}
		}
	}
	return nil
}

func sameLogRuleDefinition(a, b LogRule) bool {
	if a.Mode != b.Mode || a.Pattern != b.Pattern || len(a.Exclude) != len(b.Exclude) {
		return false
	}
	for i := range a.Exclude {
		if a.Exclude[i] != b.Exclude[i] {
			return false
		}
	}
	return true
}

// Rule versions belong to the server; retries of an unchanged definition keep
// the existing counter series. Call while holding the cluster configuration lock.
func versionLogRules(current, incoming LogAgentConfig) LogAgentConfig {
	incoming = cloneLogAgentConfig(incoming)
	for i := range incoming.Targets {
		for j := range incoming.Targets[i].Rules {
			rule := &incoming.Targets[i].Rules[j]
			rule.Version = 1
			for _, target := range current.Targets {
				if target.ID != incoming.Targets[i].ID {
					continue
				}
				for _, previous := range target.Rules {
					if previous.ID != rule.ID {
						continue
					}
					rule.Version = previous.Version
					if !sameLogRuleDefinition(previous, *rule) {
						rule.Version++
					}
				}
			}
		}
	}
	return incoming
}
func EncodeLogAgentConfig(c LogAgentConfig) ([]byte, error) {
	if err := ValidateLogAgentConfig(c); err != nil {
		return nil, err
	}
	return json.Marshal(c)
}
