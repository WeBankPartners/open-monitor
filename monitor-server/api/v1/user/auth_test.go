package user

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
)

func TestAuthRequiredDoesNotTreatQueryParametersAsPublicPaths(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"http":{"session":{"enable":"false"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	models.InitConfig(path)
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
	for _, path := range []string{"/monitor/api/v1/query", "/monitor/api/v1/export/ping/source", "/monitor/api/v1/agent/export/custom"} {
		reached := false
		r := gin.New()
		r.GET(path, AuthRequired(), func(c *gin.Context) { reached = true })
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
		if !reached {
			t.Fatalf("existing public path was blocked: %s", path)
		}
	}
}
