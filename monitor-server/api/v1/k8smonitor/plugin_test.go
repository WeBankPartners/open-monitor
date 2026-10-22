package k8smonitor

import (
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPluginEnvelopeAcceptsPlatformMetadataAndCorrelatesOutputs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &testStore{}
	h := Handler{Seed: "test-seed", Policy: k8s.Policy{Enabled: true, Environment: "sit", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {Namespaces: map[string]k8s.NamespaceScope{"sit-a": {}}}}}, Store: store}
	router := gin.New()
	router.POST("/plugin/:action", h.Plugin)
	body := `{"requestId":"request-a","operator":"operator-alias","serviceName":"monitor","servicePath":"/plugin/add","entityInstances":[],"inputs":[{"guid":"object-a","callbackParameter":"callback-a","clusterId":"cluster-a","namespace":"sit-a","version":"0","apiServer":"https://api.example.invalid","token":"test-token","ca":""},{"guid":"object-b","callbackParameter":"callback-b","clusterId":"cluster-a","namespace":"uat-a","version":"0","apiServer":"https://api.example.invalid","token":"test-token","ca":""}]}`
	out := httptest.NewRecorder()
	router.ServeHTTP(out, httptest.NewRequest(http.MethodPost, "/plugin/add", strings.NewReader(body)))
	var response struct {
		ResultCode string `json:"resultCode"`
		Results    struct {
			Outputs []struct{ Guid, CallbackParameter, ErrorCode string } `json:"outputs"`
		} `json:"results"`
	}
	if err := json.Unmarshal(out.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if out.Code != 200 || response.ResultCode != "1" || len(response.Results.Outputs) != 2 {
		t.Fatalf("bad plugin response: %s", out.Body.String())
	}
	if response.Results.Outputs[0].CallbackParameter != "callback-a" || response.Results.Outputs[0].Guid != "object-a" || response.Results.Outputs[0].ErrorCode != "0" || response.Results.Outputs[1].ErrorCode != "1" {
		t.Fatalf("lost row correlation: %s", out.Body.String())
	}
	if strings.Contains(out.Body.String(), "test-token") {
		t.Fatal("credential leaked")
	}
	if store.saved != 1 {
		t.Fatal("invalid scope persisted")
	}
}
