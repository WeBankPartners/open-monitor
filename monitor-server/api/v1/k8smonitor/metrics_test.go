package k8smonitor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type logAgentTestStore struct {
	config             k8s.LogAgentConfig
	err                error
	manual             []bool
	listErr            error
	saved              map[string]k8s.LogAgentEndpoint
	listCalls          int
	released, disabled int
}

func (s *logAgentTestStore) SaveLogAgentConfig(context.Context, string, string, k8s.LogAgentConfig) error {
	return nil
}
func (s *logAgentTestStore) UpsertLogAgentEndpoint(context.Context, k8s.LogAgentEndpoint) error {
	return nil
}

func (s *logAgentTestStore) SetLogAgentEndpoint(_ context.Context, endpoint k8s.LogAgentEndpoint, manual bool) error {
	s.manual = append(s.manual, manual)
	if s.saved == nil {
		s.saved = map[string]k8s.LogAgentEndpoint{}
	}
	s.saved[endpoint.ClusterID+"/"+endpoint.Node] = endpoint
	return nil
}
func (s *logAgentTestStore) ListLiveLogAgentEndpoints(context.Context, string, time.Duration) ([]k8s.LogAgentEndpoint, error) {
	s.listCalls++
	rows := []k8s.LogAgentEndpoint{}
	for _, endpoint := range s.saved {
		rows = append(rows, endpoint)
	}
	return rows, s.listErr
}

func (s *logAgentTestStore) ListLiveLogAgentEndpointsAll(ctx context.Context, ttl time.Duration) ([]k8s.LogAgentEndpoint, error) {
	return s.ListLiveLogAgentEndpoints(ctx, "", ttl)
}

func (s *logAgentTestStore) ListLogAgentEndpoints(context.Context, string) ([]k8s.LogAgentEndpoint, error) {
	return nil, nil
}
func (s *logAgentTestStore) ReleaseLogAgentEndpoint(context.Context, string, string) error {
	s.released++
	return nil
}
func (s *logAgentTestStore) DisableLogAgentEndpoint(context.Context, string, string) error {
	s.disabled++
	return nil
}

func (s *logAgentTestStore) GetLogAgentConfig(context.Context, string, string) (k8s.LogAgentConfig, error) {
	return s.config, s.err
}

func TestLogAgentConfigRequiresScopedBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Handler{Policy: k8s.Policy{Enabled: true, Environment: "sit", LogAgentToken: "agent-token", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {}}}, LogStore: &logAgentTestStore{config: k8s.LogAgentConfig{Version: 2}}}
	r := gin.New()
	r.GET("/config", h.LogAgentConfig)
	for _, tc := range []struct {
		name, auth string
		code       int
	}{{"missing", "", 401}, {"wrong", "Bearer wrong", 401}, {"valid", "Bearer agent-token", 200}} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/config?clusterId=cluster-a&node=node-a", nil)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			out := httptest.NewRecorder()
			r.ServeHTTP(out, req)
			if out.Code != tc.code {
				t.Fatalf("status=%d body=%s", out.Code, out.Body.String())
			}
			if tc.code == 200 {
				var body struct {
					Config k8s.LogAgentConfig `json:"config"`
				}
				if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil || body.Config.Version != 2 {
					t.Fatalf("bad config response: %s", out.Body.String())
				}
			}
		})
	}
}

func TestLogAgentConfigNotPublishedAndScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Handler{Policy: k8s.Policy{Enabled: true, Environment: "sit", LogAgentToken: "agent-token", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {}}}, LogStore: &logAgentTestStore{err: sql.ErrNoRows}}
	r := gin.New()
	r.GET("/config", h.LogAgentConfig)
	for _, path := range []string{"/config?clusterId=cluster-b&node=node-a", "/config?clusterId=cluster-a&node=../bad"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer agent-token")
		out := httptest.NewRecorder()
		r.ServeHTTP(out, req)
		if out.Code != http.StatusForbidden {
			t.Fatalf("status=%d for %s", out.Code, path)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/config?clusterId=cluster-a&node=node-a", nil)
	req.Header.Set("Authorization", "Bearer agent-token")
	out := httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if out.Code != http.StatusNoContent {
		t.Fatalf("status=%d", out.Code)
	}
}

func TestMetricsServesExporterOutput(t *testing.T) {
	h := Handler{Exporter: k8s.NewExporter(k8s.NewIndex())}
	router := gin.New()
	router.GET("/metrics", h.Metrics)
	out := httptest.NewRecorder()
	router.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if out.Code != http.StatusOK || !strings.Contains(out.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("unexpected response %d %s", out.Code, out.Body.String())
	}
}

func TestAgentRegistrationIsAutomaticAndManualUpdateIsExplicit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &logAgentTestStore{listErr: errors.New("test endpoint list unavailable")}
	h := Handler{Policy: k8s.Policy{Enabled: true, LocalCollectionEnabled: true, Environment: "sit", LogAgentToken: "agent-token", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {}}}, LogStore: store}
	r := gin.New()
	r.POST("/register", h.RegisterLogAgent)
	r.PUT("/update", h.UpdateLogAgent)
	body := `{"clusterId":"cluster-a","node":"node-a","address":"192.0.2.10","port":9105}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer agent-token")
	req.Header.Set("Content-Type", "application/json")
	out := httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if len(store.manual) != 1 || store.manual[0] {
		t.Fatalf("agent registration was treated as manual: %+v status=%d", store.manual, out.Code)
	}
	req = httptest.NewRequest(http.MethodPut, "/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	out = httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if len(store.manual) != 2 || !store.manual[1] {
		t.Fatalf("manual update was not marked manual: %+v status=%d", store.manual, out.Code)
	}
	req = httptest.NewRequest(http.MethodPut, "/update", strings.NewReader(body[:len(body)-1]+`,"manualOverride":false}`))
	req.Header.Set("Content-Type", "application/json")
	out = httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if len(store.manual) != 2 || store.released != 1 {
		t.Fatalf("automatic override release was not accepted: released=%d status=%d", store.released, out.Code)
	}
}

func TestAgentHeartbeatUpdatesFileDiscoveryWithoutPrometheusReload(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	raw, _ := json.Marshal(map[string]interface{}{"prometheus": map[string]interface{}{"sdConfigPath": dir, "configReload": "invalid-reload-url"}})
	if err := os.WriteFile(configPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	models.InitConfig(configPath)
	// Set fields directly to avoid depending on JSON spelling of legacy settings.
	models.Config().Prometheus.SdConfigPath = dir
	models.Config().Prometheus.ConfigReload = "invalid-reload-url"
	h := Handler{LogStore: &logAgentTestStore{}}
	for i := 0; i < 2; i++ {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/register", nil)
		if err := h.syncLogAgentTargets(c, "cluster-a"); err != nil {
			t.Fatalf("file discovery required a reload: %v", err)
		}
	}
	body, err := os.ReadFile(filepath.Join(dir, "k8s_log_agents.json"))
	if err != nil || string(body) != "[]" {
		t.Fatalf("targets not written: %s %v", body, err)
	}
}

func initAgentTargetConfig(t *testing.T, directory string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	models.InitConfig(path)
	models.Config().Prometheus.SdConfigPath = directory
	models.Config().Prometheus.ConfigReload = "invalid-reload-url"
}

func TestDisabledCollectionPersistsAgentChangesWithoutLiveEndpointQuery(t *testing.T) {
	for _, policy := range []k8s.Policy{{Enabled: true}, {LocalCollectionEnabled: true}} {
		dir := t.TempDir()
		initAgentTargetConfig(t, dir)
		store := &logAgentTestStore{listErr: errors.New("live endpoint query must be skipped")}
		policy.LogAgentToken = "agent-token"
		h := Handler{Policy: policy, LogStore: store}
		router := gin.New()
		router.POST("/register", h.RegisterLogAgent)
		router.PUT("/update", h.UpdateLogAgent)
		router.POST("/release", h.ReleaseLogAgent)
		router.DELETE("/delete", h.DeleteLogAgent)
		for _, change := range []struct{ method, path, body string }{
			{"POST", "/register", `{"clusterId":"cluster-a","node":"node-a","address":"192.0.2.10","port":9105}`},
			{"PUT", "/update", `{"clusterId":"cluster-a","node":"node-a","address":"192.0.2.10","port":9105}`},
			{"PUT", "/update", `{"clusterId":"cluster-a","node":"node-a","address":"192.0.2.10","port":9105,"manualOverride":false}`},
			{"POST", "/release?clusterId=cluster-a&node=node-a", ""},
			{"DELETE", "/delete?clusterId=cluster-a&node=node-a", ""},
		} {
			if err := os.WriteFile(filepath.Join(dir, "k8s_log_agents.json"), []byte(`[{"targets":["stale.example.invalid:9105"]}]`), 0600); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(change.method, change.path, strings.NewReader(change.body))
			req.Header.Set("Authorization", "Bearer agent-token")
			out := httptest.NewRecorder()
			router.ServeHTTP(out, req)
			if out.Code != http.StatusOK {
				t.Fatalf("%s %s status=%d", change.method, change.path, out.Code)
			}
			raw, err := os.ReadFile(filepath.Join(dir, "k8s_log_agents.json"))
			if err != nil || string(raw) != "[]" {
				t.Fatalf("disabled operation left targets: %s %v", raw, err)
			}
		}
		if len(store.saved) != 1 || len(store.manual) != 2 || store.released != 2 || store.disabled != 1 || store.listCalls != 0 {
			t.Fatalf("registration/admin operations were skipped or queried live endpoints: %+v", store)
		}
	}
}

func TestAgentRegistrationRetriesTargetFailureAfterPersistence(t *testing.T) {
	for _, active := range []bool{false, true} {
		dir := t.TempDir()
		initAgentTargetConfig(t, dir)
		blocked := filepath.Join(dir, "k8s_log_agents.json")
		if err := os.Mkdir(blocked, 0700); err != nil {
			t.Fatal(err)
		}
		store := &logAgentTestStore{}
		h := Handler{Policy: k8s.Policy{Enabled: true, LocalCollectionEnabled: active, LogAgentToken: "agent-token"}, LogStore: store}
		router := gin.New()
		router.POST("/register", h.RegisterLogAgent)
		for attempt, want := range []int{http.StatusServiceUnavailable, http.StatusOK, http.StatusOK} {
			if attempt == 1 {
				if err := os.Remove(blocked); err != nil {
					t.Fatal(err)
				}
			}
			req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"clusterId":"cluster-a","node":"node-a","address":"192.0.2.10","port":9105}`))
			req.Header.Set("Authorization", "Bearer agent-token")
			out := httptest.NewRecorder()
			router.ServeHTTP(out, req)
			if out.Code != want || len(store.saved) != 1 {
				t.Fatalf("active=%v attempt=%d status=%d saved=%d", active, attempt, out.Code, len(store.saved))
			}
		}
		raw, err := os.ReadFile(blocked)
		want := "[]"
		if active {
			want = `[{"targets":["192.0.2.10:9105"]}]`
		}
		if err != nil || string(raw) != want {
			t.Fatalf("retry did not reconcile targets: %s %v", raw, err)
		}
	}
}
