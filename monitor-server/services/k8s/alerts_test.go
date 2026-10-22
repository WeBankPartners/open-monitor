package k8s

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestAlertRuleRejectsAmbiguousMetricsAndInvalidFields(t *testing.T) {
	for _, field := range []string{"phase", "quality", "threshold", "name"} {
		t.Run(field, func(t *testing.T) {
			r := AlertRule{ID: "test", Metric: "cpu", Operator: ">"}
			switch field {
			case "phase", "quality":
				r.Metric = field
			case "threshold":
				r.Threshold = math.Inf(1)
			case "name":
				r.PodName = strings.Repeat("a", 254)
			}
			if ValidateAlertRule(r) == nil {
				t.Fatalf("invalid rule accepted: %+v", r)
			}
		})
	}
}

func TestAlertStatePendingThenFiring(t *testing.T) {
	now := time.Unix(100, 0)
	rule := AlertRule{ForSeconds: 30, RecoverAfter: 2}
	state := AlertState{Status: AlertNormal}
	state = Evaluate(state, Observation{Valid: true, Violated: true}, rule, now).State
	if state.Status != AlertPending {
		t.Fatalf("first violation status=%s", state.Status)
	}
	transition := Evaluate(state, Observation{Valid: true, Violated: true}, rule, now.Add(30*time.Second))
	if transition.State.Status != AlertFiring || transition.Event != AlertEventFiring {
		t.Fatalf("pending did not fire: %+v", transition)
	}
}

func TestAlertStateFiringRecoversAfterConsecutiveNormalObservations(t *testing.T) {
	now := time.Unix(100, 0)
	rule := AlertRule{ForSeconds: 0, RecoverAfter: 2}
	state := Evaluate(AlertState{}, Observation{Valid: true, Violated: true}, rule, now).State
	if state.Status != AlertFiring {
		t.Fatalf("expected firing, got %s", state.Status)
	}
	state = Evaluate(state, Observation{Valid: true}, rule, now.Add(time.Second)).State
	if state.Status != AlertFiring || state.RecoveryCount != 1 {
		t.Fatalf("first normal observation: %+v", state)
	}
	transition := Evaluate(state, Observation{Valid: true}, rule, now.Add(2*time.Second))
	if transition.State.Status != AlertRecovered || transition.Event != AlertEventRecovered {
		t.Fatalf("second normal observation did not recover: %+v", transition)
	}
}

func TestAlertStateUnknownDoesNotRecoverFiring(t *testing.T) {
	rule := AlertRule{ForSeconds: 0, RecoverAfter: 2}
	state := Evaluate(AlertState{}, Observation{Valid: true, Violated: true}, rule, time.Unix(100, 0)).State
	transition := Evaluate(state, Observation{Valid: false}, rule, time.Unix(101, 0))
	if transition.State.Status != AlertUnknown || transition.Event != AlertEventNone {
		t.Fatalf("unknown changed firing incorrectly: %+v", transition)
	}
	if transition.State.RecoveryCount != 0 {
		t.Fatalf("unknown changed recovery count: %+v", transition.State)
	}
	recovered := Evaluate(transition.State, Observation{Valid: true}, rule, time.Unix(102, 0))
	if recovered.State.Status != AlertFiring || recovered.State.RecoveryCount != 0 {
		t.Fatalf("unknown caused false recovery: %+v", recovered.State)
	}
}

func TestAlertStateUsesLogicalObjectAcrossPodUIDReplacement(t *testing.T) {
	rule := AlertRule{ForSeconds: 0, RecoverAfter: 1}
	state := Evaluate(AlertState{}, Observation{Valid: true, Violated: true, PodUID: "pod-old"}, rule, time.Unix(100, 0)).State
	transition := Evaluate(state, Observation{Valid: true, Violated: true, PodUID: "pod-new"}, rule, time.Unix(101, 0))
	if transition.State.Status != AlertFiring || transition.State.PodUID != "pod-new" {
		t.Fatalf("pod replacement was not accepted as same logical alert: %+v", transition.State)
	}
}

func TestLogAlertMissingDataDoesNotRecoverAndNewPodKeepsLogicalIncident(t *testing.T) {
	rule := AlertRule{ID: "default-log-error", Metric: "log_matches", Operator: ">", Threshold: 0, RecoverAfter: 3}
	now := time.Now()
	fired := Evaluate(AlertState{}, Observation{Valid: true, Violated: true, PodUID: "old"}, rule, now)
	if fired.Event != AlertEventFiring {
		t.Fatal(fired)
	}
	unknown := Evaluate(fired.State, Observation{Valid: false, PodUID: "new"}, rule, now)
	if unknown.State.Status != AlertUnknown || unknown.Event != AlertEventNone {
		t.Fatal(unknown)
	}
	current := Evaluate(unknown.State, Observation{Valid: true, PodUID: "new"}, rule, now)
	if current.Event != AlertEventNone || current.State.Status != AlertFiring {
		t.Fatal(current)
	}
	for i := 0; i < 3; i++ {
		current = Evaluate(current.State, Observation{Valid: true, PodUID: "new"}, rule, now.Add(time.Duration(i+1)*time.Minute))
	}
	if current.Event != AlertEventRecovered || current.State.PodUID != "new" {
		t.Fatal(current)
	}
}
