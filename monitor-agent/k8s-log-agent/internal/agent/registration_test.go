package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterSendsNumericPortAndAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" || r.URL.Path != "/monitor/internal/k8s/log-agent/register" {
			t.Fatalf("unexpected request")
		}
		var body struct {
			Address string `json:"address"`
			Port    int    `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Address != "10.0.0.2" || body.Port != 9105 {
			t.Fatalf("invalid registration payload: %+v", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := RuntimeConfig{MonitorURL: server.URL + "/monitor", ClusterID: "sit", NodeName: "node-a", AuthToken: "token", AdvertiseAddress: "10.0.0.2", HTTPAddr: ":9105"}
	if err := Register(server.Client(), c); err != nil {
		t.Fatal(err)
	}
}
