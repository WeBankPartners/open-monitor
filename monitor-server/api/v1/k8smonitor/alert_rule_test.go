package k8smonitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

type ruleStoreFake struct {
	rows    []k8s.AlertRuleRecord
	created int
	updated int
	deleted int
}

type enabledScopeStore struct{}

func (enabledScopeStore) Save(context.Context, k8s.Registration) error { return nil }
func (enabledScopeStore) List(context.Context) ([]k8s.NamespaceState, error) {
	return []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true}}}, nil
}

func (s *ruleStoreFake) ListAlertRuleRecords(context.Context, string, string) ([]k8s.AlertRuleRecord, error) {
	return s.rows, nil
}
func (s *ruleStoreFake) CreateAlertRule(_ context.Context, row k8s.AlertRuleRecord) error {
	s.created++
	s.rows = append(s.rows, row)
	return nil
}
func (s *ruleStoreFake) UpdateAlertRule(_ context.Context, _ k8s.AlertRuleRecord, _ uint64) error {
	s.updated++
	return nil
}
func (s *ruleStoreFake) DeleteAlertRule(_ context.Context, _, _, _ string, _ uint64) error {
	s.deleted++
	return nil
}

func TestCreateAlertRuleRequiresEnabledRegisteredNamespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := enabledScopeStore{}
	rules := &ruleStoreFake{}
	h := Handler{Policy: k8s.Policy{Enabled: true, Environment: "sit", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {Namespaces: map[string]k8s.NamespaceScope{"sit-a": {}}}}}, Store: store, Rules: rules}
	r := gin.New()
	r.POST("/rules", h.CreateAlertRule)
	req := httptest.NewRequest(http.MethodPost, "/rules", strings.NewReader(`{"clusterId":"cluster-a","namespace":"sit-a","ruleId":"cpu-high","metric":"cpu","operator":">","threshold":0.8,"enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rules.created != 1 {
		t.Fatalf("expected rule create, status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateAlertRuleRejectsUnregisteredNamespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Handler{Policy: k8s.Policy{Enabled: true, Environment: "sit", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {Namespaces: map[string]k8s.NamespaceScope{"sit-a": {}}}}}, Store: enabledScopeStore{}, Rules: &ruleStoreFake{}}
	r := gin.New()
	r.POST("/rules", h.CreateAlertRule)
	req := httptest.NewRequest(http.MethodPost, "/rules", strings.NewReader(`{"clusterId":"cluster-a","namespace":"uat-a","ruleId":"cpu-high","metric":"cpu","operator":">","threshold":0.8,"enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"status":"OK"`) {
		t.Fatalf("unregistered namespace accepted: %s", rec.Body.String())
	}
}
