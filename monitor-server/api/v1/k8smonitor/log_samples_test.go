package k8smonitor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

func TestLogSamplesValidateScopeAndKeepQueryErrorsVisible(t *testing.T) {
	now := time.Now().Unix()
	s := &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "ns-a", Enabled: true}}}}
	calls, fail := 0, false
	h := Handler{Store: s, QueryInstant: func(ctx context.Context, query, at string) (map[string]interface{}, error) {
		calls++
		if !strings.Contains(query, `pod="app-0",container="main"`) || at != strconv.FormatInt(now, 10) {
			t.Fatalf("incorrect sample query: %s at %s", query, at)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("query has no deadline")
		}
		if fail {
			return nil, errors.New("datasource unavailable")
		}
		return map[string]interface{}{"status": "success", "data": map[string]interface{}{"resultType": "vector", "result": []interface{}{}}}, nil
	}}
	r := gin.New()
	r.POST("/samples", h.LogSamples)
	input := map[string]interface{}{"clusterId": "cluster-a", "namespace": "ns-a", "podName": "app-0", "container": "main", "start": now - 900, "end": now}
	request := func() string {
		body, _ := json.Marshal(input)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, httptest.NewRequest("POST", "/samples", strings.NewReader(string(body))))
		return out.Body.String()
	}
	if body := request(); !strings.Contains(body, `"samples":[]`) || !strings.Contains(body, `"status":"OK"`) {
		t.Fatal(body)
	}
	fail = true
	if body := request(); !strings.Contains(body, `"status":"ERROR"`) {
		t.Fatal(body)
	}
	input["podName"] = `x"} or up{`
	request()
	input["podName"] = "app-0"
	input["end"] = now + 3600
	request()
	input["end"] = now
	input["start"] = now - 31*86400
	request()
	input["start"] = now - 900
	s.rows[0].Enabled = false
	request()
	if calls != 2 {
		t.Fatal("invalid or disabled scope reached Prometheus", calls)
	}
}

func TestLogSamplesPropagateRequestCancellation(t *testing.T) {
	now := time.Now().Unix()
	called := false
	h := Handler{Store: &sourceMemory{rows: []k8s.NamespaceState{{Registration: k8s.Registration{ClusterID: "cluster-a", Namespace: "ns-a", Enabled: true}}}}, QueryInstant: func(ctx context.Context, _, _ string) (map[string]interface{}, error) {
		called = true
		if ctx.Err() != context.Canceled {
			t.Fatal("request cancellation was lost")
		}
		return nil, ctx.Err()
	}}
	r := gin.New()
	r.POST("/samples", h.LogSamples)
	body, _ := json.Marshal(map[string]interface{}{"clusterId": "cluster-a", "namespace": "ns-a", "podName": "app-0", "container": "main", "start": now - 900, "end": now})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := httptest.NewRecorder()
	r.ServeHTTP(out, httptest.NewRequest("POST", "/samples", strings.NewReader(string(body))).WithContext(ctx))
	if !called || !strings.Contains(out.Body.String(), `"status":"ERROR"`) {
		t.Fatal(out.Body.String())
	}
}
