package k8smonitor

import (
	"context"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testStore struct{ saved int }

func (s *testStore) Save(context.Context, k8s.Registration) error { s.saved++; return nil }
func (s *testStore) List(context.Context) ([]k8s.NamespaceState, error) {
	return []k8s.NamespaceState{}, nil
}
func TestListFiltersDeploymentScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := Handler{Policy: k8s.Policy{Enabled: true, Environment: "sit", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {Namespaces: map[string]k8s.NamespaceScope{"sit-a": {}}}}}, Store: &inventoryStore{}}
	router := gin.New()
	router.POST("/list", handler.List)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/list", nil))
	var envelope struct {
		Status string               `json:"status"`
		Data   []k8s.NamespaceState `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil || envelope.Status != "OK" || len(envelope.Data) != 1 {
		t.Fatalf("invalid UI envelope: %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "uat-a") || !strings.Contains(recorder.Body.String(), "sit-a") {
		t.Fatalf("scope filter failed: %s", recorder.Body.String())
	}
}

type inventoryStore struct{ testStore }

func (*inventoryStore) List(context.Context) ([]k8s.NamespaceState, error) {
	return []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a"}}, {Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "uat-a"}}}, nil
}
