package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestServerRejectsInvalidFileExclusions(t *testing.T) {
	for _, pattern := range []string{"", "[", "../*.log", "nested/*.log", `nested\*.log`} {
		t.Run(pattern, func(t *testing.T) {
			exclusions, _ := json.Marshal([]string{pattern})
			var cfg LogAgentConfig
			if err := json.Unmarshal([]byte(fmt.Sprintf(`{"version":1,"targets":[{"id":"app","namespace":"test","podUid":"pod-a","directory":"logs","glob":"*.log","excludeGlobs":%s,"start":"end"}]}`, exclusions)), &cfg); err != nil {
				t.Fatal(err)
			}
			if ValidateLogAgentConfig(cfg) == nil {
				t.Fatal("invalid file exclusion accepted")
			}
		})
	}
}

func TestFileExclusionChangesPublishNewConfigWithoutResettingRuleVersion(t *testing.T) {
	var before, after LogAgentConfig
	if err := json.Unmarshal([]byte(`{"version":4,"targets":[{"id":"app","namespace":"test","podUid":"pod-a","directory":"logs","glob":"*.log","start":"end","rules":[{"id":"error","version":3,"mode":"literal","pattern":"ERROR"}]}]}`), &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"version":5,"targets":[{"id":"app","namespace":"test","podUid":"pod-a","directory":"logs","glob":"*.log","excludeGlobs":["*.log.*.log"],"start":"end","rules":[{"id":"error","version":3,"mode":"literal","pattern":"ERROR"}]}]}`), &after); err != nil {
		t.Fatal(err)
	}
	if sameLogTargets(before.Targets, after.Targets) {
		t.Fatal("file exclusion change would be discarded")
	}
	versioned := versionLogRules(before, after)
	if versioned.Targets[0].Rules[0].Version != 3 || !sameLogTargets(after.Targets, versioned.Targets) {
		t.Fatal("file filtering must preserve the rule counter version and be idempotent")
	}
}

func TestDiscoveredFileDefaultsExcludeArchivesAndKeepActiveLogs(t *testing.T) {
	// Exercise the defaults attached by discovery, rather than a separate copy.
	client := cmdbRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"code":0,"statusCode":"OK","data":{"pageInfo":{"totalRows":1},"contents":[{"guid":"instance-a","deploy_path":"/var/log/app"}]}}`))}, nil
	})
	target, err := (CMDBLogTargetResolver{Client: CMDBClient{BaseURL: "https://cmdb.example.invalid", HTTP: client}}).Resolve(context.Background(), Container{InstanceID: "instance-a"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(target)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		ExcludeGlobs []string `json:"excludeGlobs"`
	}
	if err = json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		skip bool
	}{
		{"app.log.2026-01-01_08-1.log", true},
		{"app.log.2026-01-01_08-12.log", true},
		{"app.log-archive-20260101-08-1.log", true},
		{"app.log_2026_01_01_08-1.log", true},
		{"app.log.2026.01.01-08.log", true},
		{"app.log", false},
		{"rmb-client.log", false},
		{"app-worker-1.log", false},
		{"app.log.worker.log", false},
		{"app.log.backup.log", false},
		{"app.log.12345678.log", false},
		{"app-2026-01-01.log", false},
	} {
		skipped := false
		for _, pattern := range wire.ExcludeGlobs {
			match, err := filepath.Match(pattern, tc.name)
			if err != nil {
				t.Fatal(err)
			}
			skipped = skipped || match
		}
		if skipped != tc.skip {
			t.Errorf("file=%s excluded=%t want=%t", tc.name, skipped, tc.skip)
		}
	}
}
