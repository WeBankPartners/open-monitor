package k8s

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
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
			if len(r.args[i]) != 7 || r.args[i][3] != "log_matches" || r.args[i][4] != ">" || r.args[i][5] != float64(0) || r.args[i][6] != 0 {
				t.Fatal(r.args[i])
			}
		}
	}
	if !found {
		t.Fatal("missing default log alert")
	}
}

func TestDefaultCPUThrottledFiltersNoiseAndWaitsForSustainedThrottling(t *testing.T) {
	r := &defaultRuleRecorder{}
	if err := (Store{}).ensureDefaultAlertRulesTx(context.Background(), r, "cluster-a", "sit-a"); err != nil {
		t.Fatal(err)
	}
	var rule AlertRule
	for _, args := range r.args {
		if args[0] == "default-cpu-throttled" {
			if len(args) != 7 {
				t.Fatalf("default CPU rule does not provide a configurable pending duration: %v", args)
			}
			rule = AlertRule{ID: args[0].(string), Metric: args[3].(string), Operator: args[4].(string), Threshold: args[5].(float64), ForSeconds: args[6].(int), RecoverAfter: 3}
		}
	}
	if rule.ID == "" || rule.Metric != "cpu_throttled" || rule.Threshold != 0.1 || rule.ForSeconds != 300 {
		t.Fatalf("unexpected default CPU rule: %+v", rule)
	}
	state := AlertState{}
	for _, tc := range []struct {
		seconds int64
		value   float64
		status  AlertStatus
		event   AlertEvent
	}{
		{0, 0.05, AlertNormal, AlertEventNone},
		{30, 0.1, AlertNormal, AlertEventNone},
		{60, 0.2, AlertPending, AlertEventNone},
		{330, 0.2, AlertPending, AlertEventNone},
		{360, 0.2, AlertFiring, AlertEventFiring},
		{390, 0.05, AlertFiring, AlertEventNone},
		{420, 0.05, AlertFiring, AlertEventNone},
		{450, 0.05, AlertRecovered, AlertEventRecovered},
		{480, 0.2, AlertPending, AlertEventNone},
		{510, 0.05, AlertNormal, AlertEventNone},
	} {
		transition := Evaluate(state, Observation{Valid: true, Violated: compare(rule.Operator, tc.value, rule.Threshold)}, rule, time.Unix(tc.seconds, 0))
		if transition.State.Status != tc.status || transition.Event != tc.event {
			t.Fatalf("at %ds with value %g: %+v", tc.seconds, tc.value, transition)
		}
		state = transition.State
	}
}
