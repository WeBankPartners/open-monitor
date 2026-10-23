package agent

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadRuntimeConfigFromEnv(t *testing.T) {
	c, e := LoadRuntimeConfig(func(k string) string {
		return map[string]string{"LOG_AGENT_AUTH_TOKEN": "test-token", "LOG_AGENT_ROOT": "/logs", "LOG_AGENT_STATE_DIR": "/state", "LOG_AGENT_MONITOR_URL": "https://monitor.example", "LOG_AGENT_CLUSTER_ID": "sit", "LOG_AGENT_NODE_NAME": "node-a", "LOG_AGENT_ADVERTISE_ADDRESS": "10.0.0.2", "LOG_AGENT_POLL_INTERVAL": "2s"}[k]
	})
	if e != nil || c.PollInterval != 2*time.Second {
		t.Fatalf("%+v %v", c, e)
	}
}

func TestLoadRuntimeConfigAcceptsHTTPMonitor(t *testing.T) {
	base := map[string]string{"LOG_AGENT_AUTH_TOKEN": "test-token", "LOG_AGENT_ROOT": "/logs", "LOG_AGENT_STATE_DIR": "/state", "LOG_AGENT_MONITOR_URL": "http://monitor.example/monitor", "LOG_AGENT_CLUSTER_ID": "sit", "LOG_AGENT_NODE_NAME": "node-a", "LOG_AGENT_ADVERTISE_ADDRESS": "10.0.0.2"}
	if _, err := LoadRuntimeConfig(func(k string) string { return base[k] }); err != nil {
		t.Fatal(err)
	}
}
func TestFetchConfigEncodesScopeAndAuth(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("node") != "node/a" || r.Header.Get("Authorization") != "Bearer token" {
			t.Fatal(r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"config":{"version":2,"targets":[]}}`))
	}))
	defer s.Close()
	c := RuntimeConfig{MonitorURL: s.URL, ClusterID: "sit", NodeName: "node/a", AuthToken: "token"}
	got, e := FetchConfig(s.Client(), c, 1)
	if e != nil || got.Version != 2 {
		t.Fatalf("%+v %v", got, e)
	}
}

func TestFetchConfigDistinguishesWaitingFromHTTPFailures(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden, http.StatusServiceUnavailable} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
			defer server.Close()
			_, err := FetchConfig(server.Client(), RuntimeConfig{MonitorURL: server.URL, ClusterID: "sit", NodeName: "node-a", AuthToken: "token"}, 0)
			if status == http.StatusNoContent {
				if !errors.Is(err, ErrConfigNotPublished) {
					t.Fatalf("expected unpublished config, got %v", err)
				}
			} else if err == nil || errors.Is(err, ErrConfigNotPublished) || !strings.Contains(err.Error(), fmt.Sprint(status)) {
				t.Fatalf("HTTP failure must retain status %d, got %v", status, err)
			}
		})
	}
}

func TestRuntimeRejectsUnsafeBootstrap(t *testing.T) {
	for _, tc := range []struct{ key, value string }{{"LOG_AGENT_AUTH_TOKEN", ""}, {"LOG_AGENT_MONITOR_URL", "ftp://monitor.example"}, {"LOG_AGENT_MONITOR_URL", "https://monitor.example?token=hidden"}, {"LOG_AGENT_ROOT", "relative"}} {
		env := map[string]string{"LOG_AGENT_ROOT": "/logs", "LOG_AGENT_STATE_DIR": "/state", "LOG_AGENT_MONITOR_URL": "https://monitor.example", "LOG_AGENT_CLUSTER_ID": "sit", "LOG_AGENT_NODE_NAME": "node-a", "LOG_AGENT_AUTH_TOKEN": "test-token", "LOG_AGENT_ADVERTISE_ADDRESS": "10.0.0.2"}
		env[tc.key] = tc.value
		if _, err := LoadRuntimeConfig(func(k string) string { return env[k] }); err == nil {
			t.Errorf("accepted unsafe %s", tc.key)
		}
	}
}

func TestLogMetricsShareObjectLabelsAndReportScanHealth(t *testing.T) {
	root := t.TempDir()
	a, err := New(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	cfg := Config{Version: 1, Targets: []Target{{ID: "target", Namespace: "sit-a", PodName: "app-0", PodUID: "pod-a", Container: "app", WorkloadName: "app", Directory: root, Glob: "*.log", Start: "beginning", Rules: []Rule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	if err = a.Apply(cfg); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	text := a.MetricsText("cluster-a", "node-a")
	if !strings.Contains(text, `cluster_id="cluster-a"`) || !strings.Contains(text, `workload_name="app"`) || !strings.Contains(text, "k8s_log_target_healthy") {
		t.Fatal(text)
	}
	if err = os.WriteFile(filepath.Join(root, "app.log"), []byte("ERROR test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(a.MetricsText("cluster-a", "node-a"), "\n") {
		if strings.HasPrefix(line, "k8s_log_target_healthy{") && !strings.HasSuffix(line, " 1") {
			t.Fatal(line)
		}
	}
	if err = os.Remove(filepath.Join(root, "app.log")); err != nil {
		t.Fatal(err)
	}
	if err = a.Scan(); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(a.MetricsText("cluster-a", "node-a"), "\n") {
		if strings.HasPrefix(line, "k8s_log_target_healthy{") && !strings.HasSuffix(line, " 0") {
			t.Fatal(line)
		}
	}
}
