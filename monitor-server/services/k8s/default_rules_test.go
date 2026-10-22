package k8s

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type defaultRuleRecorder struct {
	queries []string
	args    [][]interface{}
}

func (r *defaultRuleRecorder) ExecContext(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
	r.queries = append(r.queries, q)
	r.args = append(r.args, a)
	return nil, nil
}
func TestDefaultRulesIncludeLogsAndPreserveExistingRules(t *testing.T) {
	r := &defaultRuleRecorder{}
	if err := (Store{}).ensureDefaultAlertRulesTx(context.Background(), r, "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	found := false
	for i, q := range r.queries {
		if !strings.HasSuffix(q, "ON DUPLICATE KEY UPDATE rule_id=rule_id") {
			t.Fatal("defaults overwrite operator settings")
		}
		if r.args[i][0] == "default-log-error" {
			found = true
			if r.args[i][3] != "log_matches" || r.args[i][4] != ">" || r.args[i][5] != float64(0) {
				t.Fatal(r.args[i])
			}
		}
	}
	if !found {
		t.Fatal("missing default log alert")
	}
}
