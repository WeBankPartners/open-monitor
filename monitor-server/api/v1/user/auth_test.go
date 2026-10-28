package user

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
)

func initAuthTestConfig(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"http":{"session":{"enable":"false"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	models.InitConfig(path)
}

func TestAuthRequiredDoesNotTreatQueryParametersAsPublicPaths(t *testing.T) {
	initAuthTestConfig(t)
	for _, suffix := range []string{"", "?x=/query", "?x=/export/ping/source", "?x=/agent/export/custom"} {
		reached := false
		r := gin.New()
		r.POST("/monitor/api/v1/k8s/log-samples", AuthRequired(), func(c *gin.Context) { reached = true })
		out := httptest.NewRecorder()
		r.ServeHTTP(out, httptest.NewRequest("POST", "/monitor/api/v1/k8s/log-samples"+suffix, nil))
		if reached {
			t.Fatalf("unauthenticated sample request reached the handler: %s", suffix)
		}
	}
}

func TestAuthRequiredPreservesExistingPublicPaths(t *testing.T) {
	initAuthTestConfig(t)
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/monitor/api/v1/agent/export/ping/source"},
		{http.MethodPost, "/monitor/api/v1/agent/export/custom/endpoint/add"},
		{http.MethodPost, "/monitor/api/v1/alarm/problem/query"},
		{http.MethodGet, "/monitor/api/v2/alarm/endpoint_group/query"},
		{http.MethodPost, "/monitor/api/v2/alarm/strategy/query"},
		{http.MethodGet, "/monitor/api/v2/monitor/endpoint/query"},
		{http.MethodGet, "/monitor/api/v2/config/type/query"},
		{http.MethodPost, "/monitor/entities/alarm/query"},
		{http.MethodPost, "/monitor/entities/alarm_event/query"},
		{http.MethodPost, "/monitor/entities/endpoint/query"},
		{http.MethodPost, "/monitor/entities/service_group/query"},
		{http.MethodPost, "/monitor/entities/endpoint_group/query"},
		{http.MethodPost, "/monitor/entities/monitor_type/query"},
		{http.MethodPost, "/monitor/entities/log_monitor_template/query"},
	} {
		reached := false
		r := gin.New()
		r.Handle(route.method, route.path, AuthRequired(), func(c *gin.Context) { reached = true })
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(route.method, route.path, nil))
		if !reached {
			t.Errorf("existing public route was blocked: %s %s", route.method, route.path)
		}
	}
}

func TestAuthRequiredRejectsPublicPathSubstringsAndWrongMethods(t *testing.T) {
	initAuthTestConfig(t)
	for _, route := range []struct{ method, path string }{
		{http.MethodPut, "/monitor/api/v1/k8s/alert-rules/query"},
		{http.MethodDelete, "/monitor/api/v1/k8s/alert-rules/query-error"},
		{http.MethodGet, "/monitor/api/v2/service/log_metric/list/query/instance-a"},
		{http.MethodGet, "/monitor/api/v2/config/type/query-extra"},
		{http.MethodGet, "/monitor/api/v1/agent/export/ping/source-extra"},
		{http.MethodPost, "/monitor/api/v1/agent/export/custom/endpoint/add-extra"},
		{http.MethodPut, "/monitor/api/v1/agent/export/ping/source"},
		{http.MethodGet, "/monitor/api/v1/agent/export/custom/endpoint/add"},
		{http.MethodPost, "/monitor/api/v2/config/type/query"},
		{http.MethodGet, "/monitor/entities/alarm/query"},
		{http.MethodGet, "/monitor/api/v1/query"},
		{http.MethodGet, "/monitor/api/v1/export/ping/source"},
		{http.MethodGet, "/monitor/api/v1/agent/export/custom"},
	} {
		reached := false
		r := gin.New()
		r.Handle(route.method, route.path, AuthRequired(), func(c *gin.Context) { reached = true })
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(route.method, route.path, nil))
		if reached {
			t.Errorf("unauthenticated request reached protected handler: %s %s", route.method, route.path)
		}
	}
}
