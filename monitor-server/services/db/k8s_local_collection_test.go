package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/k8smonitor"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

func TestKubernetesTargetReconciliationRequiresBothSwitches(t *testing.T) {
	for _, tc := range []struct {
		name           string
		master, local  bool
		monitor, agent string
	}{
		{"default", false, false, "[]", "[]"},
		{"master only", true, false, "[]", "[]"},
		{"local only", false, true, "[]", "[]"},
		{"active", true, true, `[{"labels":{"__metrics_path__":"/monitor/internal/k8s/metrics"},"targets":["localhost:8080"]}]`, `[{"targets":["agent.example.invalid:9105"]}]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, file := range []string{"k8s_monitor.json", "k8s_log_agents.json"} {
				if err := os.WriteFile(filepath.Join(dir, file), []byte(`[{"targets":["stale.example.invalid:9105"]}]`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			calls := 0
			list := func(context.Context, time.Duration) ([]k8s.LogAgentEndpoint, error) {
				calls++
				if !tc.master || !tc.local {
					t.Fatal("disabled collection queried live endpoints")
				}
				return []k8s.LogAgentEndpoint{{Address: "agent.example.invalid", Port: 9105}}, nil
			}
			for i := 0; i < 2; i++ {
				if err := syncKubernetesTargets(context.Background(), k8s.Policy{Enabled: tc.master, LocalCollectionEnabled: tc.local}, dir, "8080", list); err != nil {
					t.Fatal(err)
				}
				for file, want := range map[string]string{"k8s_monitor.json": tc.monitor, "k8s_log_agents.json": tc.agent} {
					raw, err := os.ReadFile(filepath.Join(dir, file))
					if err != nil || string(raw) != want {
						t.Fatalf("%s=%s, want %s: %v", file, raw, want, err)
					}
				}
			}
			if tc.master && tc.local && calls != 2 {
				t.Fatalf("active reconciliation skipped endpoint discovery: %d calls", calls)
			}
		})
	}
}

type concurrentAgentStore struct {
	writes, queries atomic.Int64
	active          bool
}

func (s *concurrentAgentStore) GetLogAgentConfig(context.Context, string, string) (k8s.LogAgentConfig, error) {
	return k8s.LogAgentConfig{}, nil
}
func (s *concurrentAgentStore) SaveLogAgentConfig(context.Context, string, string, k8s.LogAgentConfig) error {
	return nil
}
func (s *concurrentAgentStore) UpsertLogAgentEndpoint(context.Context, k8s.LogAgentEndpoint) error {
	return nil
}
func (s *concurrentAgentStore) SetLogAgentEndpoint(context.Context, k8s.LogAgentEndpoint, bool) error {
	s.writes.Add(1)
	return nil
}
func (s *concurrentAgentStore) ListLiveLogAgentEndpoints(context.Context, string, time.Duration) ([]k8s.LogAgentEndpoint, error) {
	query := s.queries.Add(1)
	if !s.active {
		return nil, errors.New("disabled registration queried endpoints")
	}
	return []k8s.LogAgentEndpoint{{Address: "agent.example.invalid", Port: 9105 + int(query%2)}}, nil
}

func TestConcurrentPeriodicSyncAndAgentRegistrationKeepCompleteJSON(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(fmt.Sprintf("active=%v", active), func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			dir := t.TempDir()
			configPath := filepath.Join(dir, "config.json")
			if err := os.WriteFile(configPath, []byte(`{}`), 0600); err != nil {
				t.Fatal(err)
			}
			models.InitConfig(configPath)
			models.Config().Prometheus.SdConfigPath = dir
			policy := k8s.Policy{Enabled: true, LocalCollectionEnabled: active, LogAgentToken: "agent-token"}
			store := &concurrentAgentStore{active: active}
			list := func(ctx context.Context, ttl time.Duration) ([]k8s.LogAgentEndpoint, error) {
				return store.ListLiveLogAgentEndpoints(ctx, "", ttl)
			}
			if err := syncKubernetesTargets(context.Background(), policy, dir, "8080", list); err != nil {
				t.Fatal(err)
			}
			handler := k8smonitor.Handler{Policy: policy, LogStore: store}
			router := gin.New()
			router.POST("/register", handler.RegisterLogAgent)
			const attempts = 50
			var writers sync.WaitGroup
			writers.Add(2)
			go func() {
				defer writers.Done()
				for i := 0; i < attempts; i++ {
					if err := syncKubernetesTargets(context.Background(), policy, dir, "8080", list); err != nil {
						t.Error(err)
					}
				}
			}()
			go func() {
				defer writers.Done()
				for i := 0; i < attempts; i++ {
					req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"clusterId":"cluster-a","node":"node-a","address":"192.0.2.10","port":9105}`))
					req.Header.Set("Authorization", "Bearer agent-token")
					out := httptest.NewRecorder()
					router.ServeHTTP(out, req)
					if out.Code != http.StatusOK {
						t.Errorf("concurrent registration returned %d", out.Code)
					}
				}
			}()
			done := make(chan struct{})
			go func() { writers.Wait(); close(done) }()
			check := func() {
				for _, name := range []string{"k8s_monitor.json", "k8s_log_agents.json"} {
					raw, err := os.ReadFile(filepath.Join(dir, name))
					var targets []interface{}
					if err != nil || json.Unmarshal(raw, &targets) != nil || (!active && string(raw) != "[]") || (active && len(targets) != 1) {
						t.Errorf("concurrent sync exposed incomplete or unexpected %s: %s %v", name, raw, err)
					}
				}
			}
			for {
				check()
				select {
				case <-done:
					check()
					wantQueries := int64(0)
					if active {
						wantQueries = 2*attempts + 1
					}
					if store.writes.Load() != attempts || store.queries.Load() != wantQueries {
						t.Fatalf("registration writes=%d live queries=%d", store.writes.Load(), store.queries.Load())
					}
					return
				default:
				}
			}
		})
	}
}

func TestKubernetesTargetReconciliationReturnsFailuresAndRetries(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "k8s_log_agents.json")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	list := func(context.Context, time.Duration) ([]k8s.LogAgentEndpoint, error) {
		t.Fatal("disabled retry queried endpoints")
		return nil, nil
	}
	if err := syncKubernetesTargets(context.Background(), k8s.Policy{}, dir, "8080", list); err == nil {
		t.Fatal("failed target write was reported as success")
	}
	if err := os.Remove(blocked); err != nil {
		t.Fatal(err)
	}
	if err := syncKubernetesTargets(context.Background(), k8s.Policy{}, dir, "8080", list); err != nil {
		t.Fatalf("failed reconciliation could not retry: %v", err)
	}
	wantErr := errors.New("live endpoint database unavailable")
	list = func(context.Context, time.Duration) ([]k8s.LogAgentEndpoint, error) { return nil, wantErr }
	if err := syncKubernetesTargets(context.Background(), k8s.Policy{Enabled: true, LocalCollectionEnabled: true}, dir, "8080", list); !errors.Is(err, wantErr) {
		t.Fatalf("endpoint failure lost: %v", err)
	}
}
