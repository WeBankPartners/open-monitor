package k8smonitor

import (
	"context"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

type sourceMemory struct {
	rows  []k8s.NamespaceState
	saved []k8s.Registration
}

func (s *sourceMemory) List(context.Context) ([]k8s.NamespaceState, error) { return s.rows, nil }
func (s *sourceMemory) Save(_ context.Context, r k8s.Registration) error {
	s.saved = append(s.saved, r)
	return nil
}
func TestWebSourceCreateWithoutExternalGuid(t *testing.T) {
	s := &sourceMemory{}
	h := Handler{Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: s}
	r := gin.New()
	r.PUT("/source", h.PutSource)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("PUT", "/source", strings.NewReader(`{"clusterId":"cluster-a","namespace":"sit-a","token":""}`)))
	if out.Code != 200 || len(s.saved) != 1 || !strings.Contains(out.Body.String(), `"status":"OK"`) {
		t.Fatalf("save failed: %d %s", out.Code, out.Body.String())
	}
	if s.saved[0].Credentials.EncryptedToken != "" {
		t.Fatal("inherited token must not be copied")
	}
}
func TestWebNamespaceEmptyTokenClearsOverride(t *testing.T) {
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a", Version: 2, Credentials: k8s.Credentials{Guid: "old", EncryptedToken: "old-token"}}}}}
	h := Handler{Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: s}
	r := gin.New()
	r.PUT("/source", h.PutSource)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("PUT", "/source", strings.NewReader(`{"clusterId":"cluster-a","namespace":"sit-a","token":""}`)))
	if len(s.saved) != 1 || s.saved[0].Credentials.EncryptedToken != "" {
		t.Fatalf("override not cleared: %s", out.Body.String())
	}
}
func TestWebSourceDeleteQueryIsIdempotent(t *testing.T) {
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: false, Version: 2}}}}
	h := Handler{Policy: k8s.Policy{Environment: "sit"}, Store: s}
	r := gin.New()
	r.DELETE("/source", h.DeleteSource)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("DELETE", "/source?clusterId=cluster-a&namespace=sit-a", nil))
	if len(s.saved) != 0 || !strings.Contains(out.Body.String(), `"status":"OK"`) {
		t.Fatalf("delete not idempotent: %s", out.Body.String())
	}
}

func TestClusterCreateWithoutTokenReturnsActionableError(t *testing.T) {
	s := &sourceMemory{}
	h := Handler{ClusterMode: true, Seed: "test-seed", Policy: k8s.Policy{Environment: "sit"}, Store: s}
	r := gin.New()
	r.PUT("/source", h.PutSource)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("PUT", "/source", strings.NewReader(`{"clusterId":"cluster-a","apiServer":"https://api.example.invalid","token":"","enabled":true}`)))
	if out.Code != 400 || !strings.Contains(out.Body.String(), `"error":"cluster token is required"`) {
		t.Fatalf("expected actionable missing-token error, got %d %s", out.Code, out.Body.String())
	}
	if len(s.saved) != 0 {
		t.Fatal("cluster without a token must not be saved")
	}
}

func TestClusterCreateWithoutEncryptionSeedReturnsServerError(t *testing.T) {
	s := &sourceMemory{}
	h := Handler{ClusterMode: true, Policy: k8s.Policy{Environment: "sit"}, Store: s}
	r := gin.New()
	r.PUT("/source", h.PutSource)
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("PUT", "/source", strings.NewReader(`{"clusterId":"cluster-a","apiServer":"https://api.example.invalid","token":"provided-token","enabled":true}`)))
	if out.Code != 500 || !strings.Contains(out.Body.String(), `"error":"credential encryption is not configured"`) {
		t.Fatalf("expected server configuration error, got %d %s", out.Code, out.Body.String())
	}
	if len(s.saved) != 0 {
		t.Fatal("cluster must not be saved without the encryption seed")
	}
}
