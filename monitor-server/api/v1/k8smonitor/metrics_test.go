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
	config  k8s.LogAgentConfig
	err     error
	manual  []bool
	listErr error
}

func (s logAgentTestStore) SaveLogAgentConfig(context.Context, string, string, k8s.LogAgentConfig) error {
	return nil
}
func (s logAgentTestStore) UpsertLogAgentEndpoint(context.Context, k8s.LogAgentEndpoint) error {
	return nil
}

func (s *logAgentTestStore) SetLogAgentEndpoint(_ context.Context, _ k8s.LogAgentEndpoint, manual bool) error {
	s.manual = append(s.manual, manual)
	return nil
}
func (s logAgentTestStore) ListLiveLogAgentEndpoints(context.Context, string, time.Duration) ([]k8s.LogAgentEndpoint, error) {
	return nil, s.listErr
}

func (s logAgentTestStore) GetLogAgentConfig(context.Context, string, string) (k8s.LogAgentConfig, error) {
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
	h := Handler{Policy: k8s.Policy{Enabled: true, Environment: "sit", LogAgentToken: "agent-token", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {}}}, LogStore: store}
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
	if len(store.manual) != 3 || store.manual[2] {
		t.Fatalf("automatic override release was not accepted: %+v status=%d", store.manual, out.Code)
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
