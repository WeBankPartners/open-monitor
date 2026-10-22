package k8smonitor

import (
	"context"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

type objectAlerts struct{}

func (objectAlerts) ListStates(context.Context, string, string) ([]k8s.AlertStateView, error) {
	return []k8s.AlertStateView{}, nil
}
func TestObjectsReturnsPendingInventoryWithoutCredentials(t *testing.T) {
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true, Credentials: k8s.Credentials{EncryptedToken: "must-not-leak"}}, State: "PENDING"}}}
	h := Handler{Policy: k8s.Policy{Environment: "sit"}, Store: s, AlertStore: objectAlerts{}}
	r := gin.New()
	r.GET("/objects", h.Objects)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("GET", "/objects?clusterId=cluster-a&namespace=sit-a", nil))
	if !strings.Contains(out.Body.String(), `"state":"PENDING"`) || !strings.Contains(out.Body.String(), `"observedAt":null`) || strings.Contains(out.Body.String(), "must-not-leak") {
		t.Fatal(out.Body.String())
	}
	s.rows[0].Enabled = false
	out = httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("GET", "/objects?clusterId=cluster-a&namespace=sit-a", nil))
	if strings.Contains(out.Body.String(), `"state":"PENDING"`) {
		t.Fatal("disabled scope exposed")
	}
}
