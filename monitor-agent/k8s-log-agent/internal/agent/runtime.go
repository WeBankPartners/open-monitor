package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RuntimeConfig struct {
	Root, StateDir, MonitorURL, ClusterID, NodeName, AuthToken, HTTPAddr, AdvertiseAddress string
	PollInterval                                                                           time.Duration
}

var ErrConfigNotPublished = errors.New("configuration is not published")

func LoadRuntimeConfig(get func(string) string) (RuntimeConfig, error) {
	c := RuntimeConfig{Root: get("LOG_AGENT_ROOT"), StateDir: get("LOG_AGENT_STATE_DIR"), MonitorURL: strings.TrimRight(get("LOG_AGENT_MONITOR_URL"), "/"), ClusterID: get("LOG_AGENT_CLUSTER_ID"), NodeName: get("LOG_AGENT_NODE_NAME"), AuthToken: get("LOG_AGENT_AUTH_TOKEN"), HTTPAddr: get("LOG_AGENT_HTTP_ADDR"), AdvertiseAddress: get("LOG_AGENT_ADVERTISE_ADDRESS"), PollInterval: 10 * time.Second}
	if v := get("LOG_AGENT_POLL_INTERVAL"); v != "" {
		d, e := time.ParseDuration(v)
		if e != nil || d < time.Second {
			return c, errors.New("invalid poll interval")
		}
		c.PollInterval = d
	}
	if c.Root == "" || c.StateDir == "" || c.MonitorURL == "" || c.ClusterID == "" || c.NodeName == "" || c.AuthToken == "" {
		return c, errors.New("required agent environment is missing")
	}
	if c.HTTPAddr == "" {
		c.HTTPAddr = ":9105"
	}
	if c.AdvertiseAddress == "" {
		return c, errors.New("agent advertise address is required")
	}
	if !strings.HasPrefix(c.Root, "/") || !strings.HasPrefix(c.StateDir, "/") {
		return c, errors.New("absolute agent directories required")
	}
	u, e := url.Parse(c.MonitorURL)
	if e != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return c, errors.New("invalid monitor URL")
	}
	return c, nil
}
func Register(client *http.Client, c RuntimeConfig) error {
	u, err := url.Parse(c.MonitorURL + "/internal/k8s/log-agent/register")
	if err != nil {
		return errors.New("invalid registration URL")
	}
	port := c.HTTPAddr
	if i := strings.LastIndex(port, ":"); i >= 0 {
		port = port[i+1:]
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return errors.New("invalid agent port")
	}
	payload, err := json.Marshal(struct {
		ClusterID string `json:"clusterId"`
		Node      string `json:"node"`
		Address   string `json:"address"`
		Port      int    `json:"port"`
	}{c.ClusterID, c.NodeName, c.AdvertiseAddress, portNumber})
	if err != nil {
		return errors.New("invalid registration payload")
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("agent registration failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("agent registration status %d", resp.StatusCode)
	}
	return nil
}
func FetchConfig(client *http.Client, c RuntimeConfig, current uint64) (Config, error) {
	u, err := url.Parse(c.MonitorURL + "/internal/k8s/log-agent/config")
	if err != nil {
		return Config{}, errors.New("invalid config URL")
	}
	q := u.Query()
	q.Set("clusterId", c.ClusterID)
	q.Set("node", c.NodeName)
	q.Set("version", strconv.FormatUint(current, 10))
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return Config{}, errors.New("invalid config request")
	}
	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	resp, err := client.Do(req)
	if err != nil {
		return Config{}, errors.New("config request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return Config{}, ErrConfigNotPublished
		}
		return Config{}, fmt.Errorf("config request status %d", resp.StatusCode)
	}
	var out struct {
		Config Config `json:"config"`
	}
	if err = json.NewDecoder(http.MaxBytesReader(nil, resp.Body, 4<<20)).Decode(&out); err != nil {
		return Config{}, errors.New("invalid config response")
	}
	return out.Config, nil
}

func EncodeConfig(c Config) ([]byte, error) { return json.Marshal(c) }

func (a *Agent) MetricsText(cluster, node string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var b strings.Builder
	fmt.Fprintf(&b, "k8s_log_agent_up{cluster_id=%q,node=%q} 1\n", cluster, node)
	fmt.Fprintf(&b, "k8s_log_agent_config_version{cluster_id=%q,node=%q} %d\n", cluster, node, a.config.Version)
	stateAvailable := 0
	if a.stateAvailable {
		stateAvailable = 1
	}
	fmt.Fprintf(&b, "k8s_log_agent_state_available{cluster_id=%q,node=%q} %d\n", cluster, node, stateAvailable)
	fmt.Fprintf(&b, "k8s_log_agent_state_resets_total{cluster_id=%q,node=%q} %d\n", cluster, node, a.stateResets)
	fmt.Fprintf(&b, "k8s_log_agent_state_persist_errors_total{cluster_id=%q,node=%q} %d\n", cluster, node, a.statePersistErrors)
	fmt.Fprintf(&b, "k8s_log_sample_limit_hits_total{cluster_id=%q,node=%q} %d\n", cluster, node, a.saved.SampleLimitHits)
	for _, t := range a.config.Targets {
		if t.Format == "otlp_metric" {
			a.writeAppMetrics(&b, t, cluster, node)
			continue
		}
		healthy := 0
		if a.healthy[t.ID] {
			healthy = 1
		}
		fmt.Fprintf(&b, "k8s_log_target_healthy{cluster_id=%q,namespace=%q,pod=%q,pod_uid=%q,workload_name=%q,workload_type=%q,container=%q,target=%q,node=%q} %d\n", cluster, t.Namespace, t.PodName, t.PodUID, t.WorkloadName, t.WorkloadType, t.Container, t.ID, node, healthy)
		missing := 0.0
		if a.missing[t.ID] {
			missing = 1
		}
		fmt.Fprintf(&b, "k8s_log_file_missing{cluster_id=%q,namespace=%q,pod=%q,pod_uid=%q,instance_id=%q,workload_name=%q,workload_type=%q,container=%q,target=%q,node=%q} %g\n", cluster, t.Namespace, t.PodName, t.PodUID, t.InstanceID, t.WorkloadName, t.WorkloadType, t.Container, t.ID, node, missing)
		fmt.Fprintf(&b, "k8s_log_read_error_total{cluster_id=%q,namespace=%q,pod=%q,pod_uid=%q,instance_id=%q,workload_name=%q,workload_type=%q,container=%q,target=%q,node=%q} %d\n", cluster, t.Namespace, t.PodName, t.PodUID, t.InstanceID, t.WorkloadName, t.WorkloadType, t.Container, t.ID, node, a.readErrors[t.ID])
		if ts := a.lastRead[t.ID]; ts > 0 {
			fmt.Fprintf(&b, "k8s_log_last_read_timestamp{cluster_id=%q,namespace=%q,pod_uid=%q,target=%q,node=%q} %d\n", cluster, t.Namespace, t.PodUID, t.ID, node, ts)
		}
		for _, r := range t.Rules {
			v := uint64(0)
			lastMatch := int64(0)
			prefix := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00", t.ID, t.PodUID, r.ID, r.Version)
			for k, c := range a.saved.Files {
				if strings.HasPrefix(k, prefix) {
					v += c.Counts[r.ID]
					if c.LastMatch > lastMatch {
						lastMatch = c.LastMatch
					}
				}
			}
			if total, ok := a.saved.Totals[strings.TrimSuffix(prefix, "\x00")]; ok {
				v += total.Counts[r.ID]
				if total.LastMatch > lastMatch {
					lastMatch = total.LastMatch
				}
			}
			fmt.Fprintf(&b, "k8s_log_match_total{cluster_id=%q,namespace=%q,pod=%q,pod_uid=%q,instance_id=%q,workload_name=%q,workload_type=%q,container=%q,target=%q,rule=%q,rule_version=%q,node=%q} %d\n", cluster, t.Namespace, t.PodName, t.PodUID, t.InstanceID, t.WorkloadName, t.WorkloadType, t.Container, t.ID, r.ID, strconv.FormatUint(r.Version, 10), node, v)
			fmt.Fprintf(&b, "k8s_log_last_match_timestamp{cluster_id=%q,namespace=%q,pod=%q,pod_uid=%q,instance_id=%q,workload_name=%q,workload_type=%q,container=%q,target=%q,rule=%q,rule_version=%q,node=%q} %d\n", cluster, t.Namespace, t.PodName, t.PodUID, t.InstanceID, t.WorkloadName, t.WorkloadType, t.Container, t.ID, r.ID, strconv.FormatUint(r.Version, 10), node, lastMatch)
			a.writeLogSample(&b, t, r, cluster, node, time.Now().Unix())
		}

	}
	return b.String()
}
