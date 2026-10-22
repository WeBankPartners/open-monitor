package k8s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientFetchCAdvisorUsesNodeProxyAndBearerToken(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("container_memory_working_set_bytes{} 1\n"))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, nil, "token-a", "sit-a")
	if err != nil {
		t.Fatal(err)
	}
	// Test transport trusts the test server certificate explicitly.
	client.http.Transport = server.Client().Transport
	payload, err := client.FetchCAdvisor(context.Background(), "node-a")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(payload, "container_memory_working_set_bytes") || gotPath != "/api/v1/nodes/node-a/proxy/metrics/cadvisor" || gotAuth != "Bearer token-a" {
		t.Fatalf("unexpected request path=%s auth=%s payload=%q", gotPath, gotAuth, payload)
	}
}

func TestClientFetchCAdvisorRejectsInvalidNodeName(t *testing.T) {
	client, err := NewClient("https://example.invalid", nil, "token-a", "sit-a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.FetchCAdvisor(context.Background(), "node/a"); err == nil {
		t.Fatal("invalid node accepted")
	}
}
