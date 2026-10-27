package k8smonitor

import (
	"context"
	"encoding/json"
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

type historicalObjectAlerts struct{ states []k8s.AlertStateView }

func (s historicalObjectAlerts) ListStates(context.Context, string, string) ([]k8s.AlertStateView, error) {
	return s.states, nil
}

func TestObjectsReportsModeWithoutRewritingHistoricalAlert(t *testing.T) {
	key := k8s.ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", PodName: "app-0", ContainerName: "main"}
	states := historicalObjectAlerts{states: []k8s.AlertStateView{{Key: k8s.AlertKey{ClusterID: key.ClusterID, Namespace: key.Namespace, PodName: key.PodName, Container: key.ContainerName, RuleID: "cpu"}, Status: "firing", PodUID: "historical-pod"}}}
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: key.ClusterID, Namespace: key.Namespace, Enabled: true}, State: "ACTIVE", Snapshot: &k8s.Snapshot{Containers: []k8s.Container{{Key: key, PodUID: "current-pod", Ready: true}}}}}}
	for _, local := range []bool{false, true} {
		h := Handler{Policy: k8s.Policy{Enabled: true, LocalCollectionEnabled: local}, Store: s, AlertStore: states}
		r := gin.New()
		r.GET("/objects", h.Objects)
		r.GET("/states", h.AlertStates)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, httptest.NewRequest("GET", "/objects?clusterId=cluster-a&namespace=sit-a", nil))
		var response struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(out.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"localCollectionEnabled", "alertEvaluationEnabled"} {
			if value, ok := response.Data[field].(bool); !ok || value != local {
				t.Fatalf("mode metadata missing: %s", out.Body.String())
			}
		}
		objects := response.Data["objects"].([]interface{})
		object := objects[0].(map[string]interface{})
		alert := object["alerts"].([]interface{})[0].(map[string]interface{})
		if alert["Status"] != "firing" || alert["PodUID"] != "historical-pod" || object["podUid"] != "current-pod" {
			t.Fatal("historical alert or discovered identity changed", object)
		}
		out = httptest.NewRecorder()
		r.ServeHTTP(out, httptest.NewRequest("GET", "/states?clusterId=cluster-a&namespace=sit-a", nil))
		var stateResponse struct {
			Data []k8s.AlertStateView `json:"data"`
		}
		if err := json.Unmarshal(out.Body.Bytes(), &stateResponse); err != nil || len(stateResponse.Data) != 1 || stateResponse.Data[0].Status != k8s.AlertFiring {
			t.Fatalf("raw alert array contract changed: %s %v", out.Body.String(), err)
		}
	}
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
