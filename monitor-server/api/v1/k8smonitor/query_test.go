package k8smonitor

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

func TestHistoryFiltersAtPrometheusAndRejectsDisabledScopes(t *testing.T) {
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true}}}}
	calls := 0
	h := Handler{Store: s, Policy: k8s.Policy{Environment: "sit"}, QueryRange: func(ctx context.Context, query string, start, end, step int64) (*models.PrometheusData, error) {
		calls++
		if !strings.Contains(query, `pod="app-0",container="main"`) || ctx.Err() != nil || step != 60 {
			t.Fatalf("incorrect history query: %s %v", query, ctx.Err())
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("query has no deadline")
		}
		return &models.PrometheusData{ResultType: "matrix"}, nil
	}}
	r := gin.New()
	r.POST("/history", h.History)
	input := map[string]interface{}{"clusterId": "cluster-a", "namespace": "sit-a", "metric": "cpu", "podName": "app-0", "container": "main", "start": time.Now().Unix() - 900, "end": time.Now().Unix(), "step": 60}
	request := func() string {
		body, _ := json.Marshal(input)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, httptest.NewRequest("POST", "/history", strings.NewReader(string(body))))
		return out.Body.String()
	}
	if body := request(); !strings.Contains(body, `"resultType":"matrix"`) || calls != 1 {
		t.Fatal(body, calls)
	}
	input["podName"] = `x"} or up{`
	request()
	input["podName"] = "app-0"
	s.rows[0].Enabled = false
	request()
	if calls != 1 {
		t.Fatal("invalid scope reached Prometheus")
	}
}

func TestMetricCatalogReportsConfiguredScrapeInterval(t *testing.T) {
	for _, seconds := range []int{30, 60, 300} {
		r := gin.New()
		r.GET("/catalog", (Handler{Policy: k8s.Policy{MetricsIntervalSeconds: seconds}}).MetricCatalog)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, httptest.NewRequest("GET", "/catalog", nil))
		var response struct {
			Data struct {
				Interval int                    `json:"scrapeIntervalSeconds"`
				Window   int                    `json:"countWindowSeconds"`
				Metrics  []k8s.MetricDefinition `json:"metrics"`
			} `json:"data"`
		}
		if err := json.Unmarshal(out.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		window := 4 * seconds
		if window < 300 {
			window = 300
		}
		if response.Data.Interval != seconds || response.Data.Window != window || len(response.Data.Metrics) == 0 {
			t.Fatal(out.Body.String())
		}
	}
}

func TestHistoryPropagatesRequestCancellation(t *testing.T) {
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true}}}}
	called := false
	h := Handler{Store: s, QueryRange: func(ctx context.Context, _ string, _, _, _ int64) (*models.PrometheusData, error) {
		called = true
		if ctx.Err() != context.Canceled {
			t.Fatal("request cancellation was lost")
		}
		return nil, ctx.Err()
	}}
	r := gin.New()
	r.POST("/history", h.History)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	body, _ := json.Marshal(map[string]interface{}{"clusterId": "cluster-a", "namespace": "sit-a", "metric": "cpu", "start": time.Now().Unix() - 900, "end": time.Now().Unix(), "step": 30})
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("POST", "/history", strings.NewReader(string(body))).WithContext(ctx))
	if !called || strings.Contains(out.Body.String(), `"status":"OK"`) {
		t.Fatal(out.Body.String())
	}
}
