package k8smonitor

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginEnvelopeAcceptsPlatformMetadataAndCorrelatesOutputs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &testStore{}
	h := Handler{Seed: "test-seed", Policy: k8s.Policy{Enabled: true, Environment: "sit", Clusters: map[string]k8s.ClusterConfig{"cluster-a": {Namespaces: map[string]k8s.NamespaceScope{"sit-a": {}}}}}, Store: store}
	router := gin.New()
	router.POST("/plugin/:action", h.Plugin)
	body := `{"requestId":"request-a","operator":"operator-alias","serviceName":"monitor","servicePath":"/plugin/put","entityInstances":[],"inputs":[{"guid":"object-a","callbackParameter":"callback-a","clusterId":"cluster-a","namespace":"sit-a","token":"test-token"},{"guid":"object-b","callbackParameter":"callback-b","clusterId":"cluster-a","namespace":"uat-a","token":"test-token"}]}`
	out := httptest.NewRecorder()
	router.ServeHTTP(out, httptest.NewRequest(http.MethodPost, "/plugin/put", strings.NewReader(body)))
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

type pluginMemory struct {
	rows    []k8s.NamespaceState
	writes  int
	listErr error
}

func (s *pluginMemory) List(context.Context) ([]k8s.NamespaceState, error) { return s.rows, s.listErr }
func (s *pluginMemory) Save(_ context.Context, r k8s.Registration) error {
	for i := range s.rows {
		if s.rows[i].ClusterID == r.ClusterID && s.rows[i].Namespace == r.Namespace {
			if s.rows[i].Version != r.Version {
				return k8s.ErrConflict
			}
			r.Version++
			s.rows[i].Registration = r
			s.writes++
			return nil
		}
	}
	r.Version = 1
	s.rows = append(s.rows, k8s.NamespaceState{Registration: r})
	s.writes++
	return nil
}

func invokeSourcePlugin(t *testing.T, h Handler, action, body string) pluginResult {
	t.Helper()
	router := gin.New()
	router.POST("/plugin/:action", h.Plugin)
	out := httptest.NewRecorder()
	router.ServeHTTP(out, httptest.NewRequest(http.MethodPost, "/plugin/"+action, strings.NewReader(body)))
	var result pluginResult
	if err := json.Unmarshal(out.Body.Bytes(), &result); err != nil || out.Code != http.StatusOK {
		t.Fatalf("invalid plugin response: status=%d body=%s", out.Code, out.Body.String())
	}
	var envelope struct {
		Results struct {
			Outputs []map[string]json.RawMessage `json:"outputs"`
		} `json:"results"`
	}
	if err := json.Unmarshal(out.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	for _, output := range envelope.Results.Outputs {
		for _, field := range []string{"guid", "callbackParameter", "errorCode", "errorMessage"} {
			if _, ok := output[field]; !ok {
				t.Fatalf("plugin output omitted %q: %s", field, out.Body.String())
			}
		}
	}
	return result
}

func TestNamespacePluginPutAndDeleteWithoutExternalVersion(t *testing.T) {
	s := &pluginMemory{}
	h := Handler{Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: s}
	body := `{"requestId":"request-a","inputs":[{"guid":"object-a","clusterId":"cluster-a","namespace":"sit-a","callbackParameter":"callback-a"}]}`
	for i := 0; i < 2; i++ {
		result := invokeSourcePlugin(t, h, "put", body)
		if result.ResultCode != "0" || len(result.Results.Outputs) != 1 || result.Results.Outputs[0].CallbackParameter != "callback-a" {
			t.Fatalf("put failed: %+v", result)
		}
	}
	if len(s.rows) != 1 || s.rows[0].Credentials.EncryptedToken != "" || s.writes != 1 {
		t.Fatalf("inherited token or repeated put changed registration: %+v", s)
	}
	for i := 0; i < 2; i++ {
		if result := invokeSourcePlugin(t, h, "delete", body); result.ResultCode != "0" {
			t.Fatalf("delete failed: %+v", result)
		}
	}
	if s.rows[0].Enabled || s.writes != 2 {
		t.Fatal("delete was not idempotent")
	}
	if result := invokeSourcePlugin(t, h, "put", body); result.ResultCode != "0" || !s.rows[0].Enabled {
		t.Fatal("put did not re-enable the namespace")
	}
}

func TestNamespacePluginEmptyTokenClearsOverride(t *testing.T) {
	s := &pluginMemory{}
	h := Handler{Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: s}
	for _, token := range []string{"test-token", ""} {
		body := `{"requestId":"request-a","inputs":[{"guid":"object-a","clusterId":"cluster-a","namespace":"sit-a","token":"` + token + `"}]}`
		if result := invokeSourcePlugin(t, h, "put", body); result.ResultCode != "0" {
			t.Fatalf("put failed: %+v", result)
		}
		if token != "" {
			plain, err := s.rows[0].Credentials.PlainToken(h.Seed)
			if err != nil || plain != token {
				t.Fatal("namespace override not stored")
			}
		}
	}
	if s.rows[0].Credentials.EncryptedToken != "" {
		t.Fatal("empty namespace token must clear its override")
	}
}

func TestClusterPluginAllowsEmptyCAAndPreservesOmittedToken(t *testing.T) {
	s := &pluginMemory{}
	h := Handler{ClusterMode: true, Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: s}
	for _, row := range []string{
		`{"guid":"object-a","clusterId":"cluster-a","apiServer":"https://api.example.invalid","token":"test-token"}`,
		`{"guid":"object-a","clusterId":"cluster-a","apiServer":"https://api.example.invalid"}`,
		`{"guid":"object-a","clusterId":"cluster-a","apiServer":"https://new-api.example.invalid"}`,
	} {
		if result := invokeSourcePlugin(t, h, "put", `{"requestId":"request-a","inputs":[`+row+`]}`); result.ResultCode != "0" {
			t.Fatalf("cluster put failed: %+v", result)
		}
	}
	if len(s.rows) != 1 || s.writes != 2 || s.rows[0].Credentials.CA != "" || s.rows[0].Credentials.APIServer != "https://new-api.example.invalid" {
		t.Fatalf("unexpected cluster state: %+v", s)
	}
	plain, err := s.rows[0].Credentials.PlainToken(h.Seed)
	if err != nil || plain != "test-token" {
		t.Fatal("existing cluster token was not retained")
	}
}

func TestSourcePluginErrorsDoNotAcknowledgePersistence(t *testing.T) {
	body := `{"requestId":"request-a","inputs":[{"guid":"object-a","clusterId":"cluster-a","namespace":"sit-a"}]}`
	for _, store := range []Store{nil, &pluginMemory{listErr: errors.New("database unavailable")}} {
		h := Handler{Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: store}
		result := invokeSourcePlugin(t, h, "put", body)
		if result.ResultCode != "1" || len(result.Results.Outputs) != 1 || result.Results.Outputs[0].ErrorCode != "1" {
			t.Fatalf("failure acknowledged: %+v", result)
		}
	}
}

func TestRegisteredSourcePluginsUseBatchRoutesAndOptionalInputs(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "build", "register.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Plugins []struct {
			Name       string `xml:"name,attr"`
			Interfaces []struct {
				Action string `xml:"action,attr"`
				Path   string `xml:"path,attr"`
				Inputs []struct {
					Name     string `xml:",chardata"`
					Required string `xml:"required,attr"`
				} `xml:"inputParameters>parameter"`
				Outputs []struct {
					Name string `xml:",chardata"`
				} `xml:"outputParameters>parameter"`
			} `xml:"interface"`
		} `xml:"plugins>plugin"`
	}
	if err = xml.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, p := range pkg.Plugins {
		kind := ""
		if p.Name == "kubernetes_namespace" {
			kind = "namespace"
		}
		if p.Name == "kubernetes_metric_cluster" {
			kind = "cluster"
		}
		if kind == "" {
			continue
		}
		found++
		if len(p.Interfaces) != 2 {
			t.Fatalf("%s must expose put/delete", p.Name)
		}
		for _, in := range p.Interfaces {
			if (in.Action != "put" && in.Action != "delete") || in.Path != "/monitor/api/v1/k8s/"+kind+"/plugin/"+in.Action {
				t.Fatalf("wrong plugin route: %+v", in)
			}
			for _, param := range in.Inputs {
				name := strings.TrimSpace(param.Name)
				if name == "version" || (kind == "namespace" && (name == "ca" || name == "apiServer")) {
					t.Fatalf("unexpected external input: %s", name)
				}
				if (name == "ca" || (name == "token" && kind == "namespace")) && param.Required == "Y" {
					t.Fatalf("%s must be optional", name)
				}
			}
			for _, param := range in.Outputs {
				if strings.TrimSpace(param.Name) == "version" {
					t.Fatalf("version is internal and must not be an output: %+v", in)
				}
			}
		}
	}
	if found != 2 {
		t.Fatalf("expected two source plugins, found %d", found)
	}
}
