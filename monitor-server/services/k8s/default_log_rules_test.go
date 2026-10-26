package k8s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func TestDiscoveredDefaultLogRuleMatchesSeverityHeaders(t *testing.T) {
	client := cmdbRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"code":0,"statusCode":"OK","data":{"pageInfo":{"totalRows":1},"contents":[{"guid":"instance-a","deploy_path":"/var/log/app"}]}}`))}, nil
	})
	target, err := (CMDBLogTargetResolver{Client: CMDBClient{BaseURL: "https://cmdb.example.invalid", HTTP: client}}).Resolve(context.Background(), Container{InstanceID: "instance-a"})
	if err != nil || len(target.Rules) != 1 {
		t.Fatalf("default rule unavailable: %+v %v", target.Rules, err)
	}
	pattern, err := regexp.Compile(target.Rules[0].Pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{
		`[2026-01-01 00:00:00.123][worker][INFO] {"error":""}`,
		`[2026-01-01 00:00:00.123][worker][DEBUG] <Error>OK</Error>`,
		`[2026-01-01 00:00:00.123][worker][INFO] select ERROR from SAMPLE_TABLE`,
		`[2026-01-01 00:00:00.123][worker][INFO] body [ERROR]`,
		`2026-01-01 00:00:00:123 INFO worker body ERROR`,
		`2026-01-01 00:00:00:123 DEBUG worker {"error":""}`,
		`2026-01-01 00:00:00:123 ERROR_CODE worker`,
		`[ERROR_CODE] payload`,
		`Caused by: example.SampleError: error response`,
		`[2026-01-01 00:00:00.123]{"sql":"select ERROR from SAMPLE_TABLE"}`,
		`unstructured payload ERROR`,
	} {
		if pattern.MatchString(line) {
			t.Errorf("non-error header or continuation counted: %q", line)
		}
	}
	for _, format := range []string{
		"[2026-01-01 00:00:00.123][worker][%s][Logger] failed",
		"[2026-01-01 00:00:00.123][%s] failed",
		"[%s] failed",
		"2026-01-01 00:00:00:123 %s worker [Logger] failed",
		"2026-01-01 00:00:00.123 %s worker failed",
		"2026-01-01 00:00:00,123 %s worker failed",
		"2026-01-01 00:00:00 %s worker failed",
	} {
		for _, level := range []string{"ERROR", "FATAL", "PANIC", "error", "fatal", "panic"} {
			line := fmt.Sprintf(format, level)
			if !pattern.MatchString(line) {
				t.Errorf("severity header omitted: %q", line)
			}
		}
	}
	if !pattern.MatchString(`[2026-01-01 00:00:00.123][worker][ERROR] INFO appears in body`) {
		t.Fatal("ERROR header excluded because its body contains INFO")
	}
}
