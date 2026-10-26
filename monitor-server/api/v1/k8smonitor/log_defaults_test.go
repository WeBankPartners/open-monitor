package k8smonitor

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

func TestLogSourceOffersCurrentDefaultsWithoutChangingSavedRules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, missing := range []bool{false, true} {
		store := &logAgentTestStore{config: k8s.LogAgentConfig{Version: 5, Targets: []k8s.LogTarget{{ID: "target-a", Rules: []k8s.LogRule{{ID: "default-error", Version: 3, Mode: "literal", Pattern: "custom"}}}}}}
		if missing {
			store.err = sql.ErrNoRows
		}
		h := Handler{Policy: k8s.Policy{Enabled: true, Clusters: map[string]k8s.ClusterConfig{"cluster-a": {}}}, LogStore: store}
		router := gin.New()
		router.GET("/source", h.GetLogAgentSource)
		out := httptest.NewRecorder()
		router.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/source?clusterId=cluster-a&node=node-a", nil))
		if out.Code != http.StatusOK {
			t.Fatalf("status=%d", out.Code)
		}
		var body struct {
			Data struct {
				Version      uint64          `json:"version"`
				Targets      []k8s.LogTarget `json:"targets"`
				DefaultRules []k8s.LogRule   `json:"defaultRules"`
			} `json:"data"`
		}
		if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Data.DefaultRules) != 1 || body.Data.DefaultRules[0].ID != "default-error" || body.Data.DefaultRules[0].Mode != "regex" {
			t.Fatalf("current defaults missing: %s", out.Body.String())
		}
		pattern, err := regexp.Compile(body.Data.DefaultRules[0].Pattern)
		if err != nil || !pattern.MatchString("2026-01-01 00:00:00:123 ERROR worker failed") || pattern.MatchString("2026-01-01 00:00:00:123 INFO body error") {
			t.Fatalf("incorrect default matching behavior: %v", err)
		}
		if missing {
			if body.Data.Version != 0 || len(body.Data.Targets) != 0 {
				t.Fatal("missing configuration must remain empty")
			}
		} else if body.Data.Version != 5 || !reflect.DeepEqual(body.Data.Targets, store.config.Targets) {
			t.Fatal("reading defaults changed the saved configuration")
		}
	}
}
