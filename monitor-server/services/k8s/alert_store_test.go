package k8s

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAlertOutboxBindsJSONAsText(t *testing.T) {
	store, _ := newLogStoreTest(t, []logStoreStep{
		{match: "FROM k8s_alert_state", err: sql.ErrNoRows},
		{match: "INSERT INTO k8s_alert_state"},
		{match: "INSERT INTO k8s_alert_outbox", check: func(args []driver.NamedValue) {
			if len(args) != 9 {
				t.Fatalf("argument count: %d", len(args))
			}
			raw, ok := args[6].Value.(string)
			if !ok {
				t.Fatalf("JSON payload argument has type %T, want string", args[6].Value)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload["key"] == nil {
				t.Fatalf("invalid alert payload %q: %v", raw, err)
			}
		}},
	})
	value := 1.0
	input := AlertEvaluation{
		Key:         AlertKey{RuleID: "cpu", ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "main"},
		Rule:        AlertRule{ID: "cpu", NotifyStrategyID: 1},
		Observation: Observation{Valid: true, Violated: true, PodUID: "pod-a"},
		Value:       &value,
	}
	if _, err := (AlertStore{DB: store.DB}).EvaluateAndPersist(context.Background(), input, time.Unix(1, 0)); err != nil {
		t.Fatalf("persist alert: %v", err)
	}
}

func TestImmediateAlertAndRecoveryPersistTheSameStartTime(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 123456789, time.UTC)
	var storedStart time.Time
	var payloadStates []AlertState
	recoveryRows := [][]driver.Value{{string(AlertFiring), "", "pod-a", nil, nil, int64(0), int64(1)}}
	capturePayload := func(args []driver.NamedValue) {
		var payload struct{ State AlertState }
		if err := json.Unmarshal([]byte(args[6].Value.(string)), &payload); err != nil {
			t.Fatal(err)
		}
		payloadStates = append(payloadStates, payload.State)
	}
	store, _ := newLogStoreTest(t, []logStoreStep{
		{match: "FROM k8s_alert_state", err: sql.ErrNoRows},
		{match: "INSERT INTO k8s_alert_state", check: func(args []driver.NamedValue) {
			// DATETIME(6) cannot retain nanoseconds. Simulate the database
			// round trip before evaluating a later recovery of this period.
			storedStart = args[12].Value.(time.Time).Round(time.Microsecond)
			recoveryRows[0][3] = storedStart
			recoveryRows[0][4] = args[11].Value.(time.Time).Round(time.Microsecond)
		}},
		{match: "INSERT INTO k8s_alert_outbox", check: capturePayload},
		{match: "FROM k8s_alert_state", columns: []string{"status", "unknown_from", "pod_uid", "started_at", "observed_at", "recovery_count", "state_version"}, rows: recoveryRows},
		{match: "UPDATE k8s_alert_state SET"},
		{match: "INSERT INTO k8s_alert_outbox", check: capturePayload},
	})
	input := AlertEvaluation{
		Key:         AlertKey{RuleID: "cpu", ClusterID: "cluster-a", Namespace: "sit-a", PodName: "app-0", Container: "main"},
		Rule:        AlertRule{ID: "cpu", ForSeconds: 0, RecoverAfter: 1, NotifyStrategyID: 1},
		Observation: Observation{Valid: true, Violated: true, PodUID: "pod-a"},
	}
	alerts := AlertStore{DB: store.DB}
	firing, err := alerts.EvaluateAndPersist(context.Background(), input, now)
	if err != nil || firing.Event != AlertEventFiring {
		t.Fatalf("immediate firing was not persisted: %+v %v", firing, err)
	}
	input.Observation.Violated = false
	recovery, err := alerts.EvaluateAndPersist(context.Background(), input, now.Add(time.Minute))
	if err != nil || recovery.Event != AlertEventRecovered {
		t.Fatalf("recovery was not persisted: %+v %v", recovery, err)
	}
	if len(payloadStates) != 2 || !payloadStates[0].StartedAt.Equal(payloadStates[1].StartedAt) || !payloadStates[0].StartedAt.Equal(storedStart) {
		t.Fatalf("firing, persisted state and recovery identify different periods: %+v", payloadStates)
	}
	for _, state := range payloadStates {
		if state.StartedAt.Nanosecond()%1000 != 0 || state.ObservedAt.Nanosecond()%1000 != 0 {
			t.Fatalf("outbox timestamps exceed database precision: %+v", state)
		}
	}
}

func TestAlertObjectEndpointDistinguishesClustersAndContainers(t *testing.T) {
	key := AlertKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "app", RuleID: "cpu"}
	first := key.Endpoint()
	other := key
	other.ClusterID = "cluster-b"
	if first == other.Endpoint() {
		t.Fatal("cluster omitted from notification identity")
	}
	other = key
	other.Container = "sidecar"
	if first == other.Endpoint() {
		t.Fatal("container omitted from notification identity")
	}
	other = key
	other.RuleID = "memory"
	if first != other.Endpoint() {
		t.Fatal("same object has different endpoints for different rules")
	}
	key.WorkloadName = strings.Repeat("a", 253)
	if len(key.Endpoint()) > 255 {
		t.Fatal("notification endpoint exceeds database field")
	}
}

func TestClaimOutboxClosesRowsBeforeLeasingEvents(t *testing.T) {
	store, script := newLogStoreTest(t, []logStoreStep{
		{match: "FROM k8s_alert_outbox", columns: []string{"event_id", "idempotency_key", "rule_id", "cluster_id", "namespace", "event_type", "payload", "attempts"}, rows: [][]driver.Value{
			{"event-a", "key-a", "cpu", "cluster-a", "sit-a", "firing", []byte(`{"key":{"RuleID":"cpu","ClusterID":"cluster-a","Namespace":"sit-a","PodName":"app-0"}}`), int64(0)},
			{"event-b", "key-b", "cpu", "cluster-a", "sit-a", "recovered", []byte(`{"key":{"RuleID":"cpu","ClusterID":"cluster-a","Namespace":"sit-a","PodName":"app-1"}}`), int64(1)},
		}},
		{match: "UPDATE k8s_alert_outbox SET lease_token"},
		{match: "UPDATE k8s_alert_outbox SET lease_token"},
	})
	events, err := (AlertStore{DB: store.DB}).ClaimOutbox(context.Background(), "worker", 32, time.Minute, time.Now())
	if err != nil || len(events) != 2 || !script.committed {
		t.Fatalf("outbox claim failed: events=%+v err=%v committed=%v", events, err, script.committed)
	}
	if events[0].Attempts != 1 || events[1].Attempts != 2 || events[0].LeaseToken == "" {
		t.Fatalf("invalid leased events: %+v", events)
	}
}

func TestAlertIdentityAndOutboxKeysAreStable(t *testing.T) {
	key := AlertKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", RuleID: "cpu-high"}
	first := OutboxIdempotencyKey(key, 4, AlertEventFiring)
	second := OutboxIdempotencyKey(key, 4, AlertEventFiring)
	if len(first) != 64 || first != second {
		t.Fatalf("outbox key is not stable: %q %q", first, second)
	}
	for _, char := range first {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Fatalf("outbox key is not lowercase hex: %q", first)
		}
	}
	if first == OutboxIdempotencyKey(key, 5, AlertEventFiring) || first == OutboxIdempotencyKey(key, 4, AlertEventRecovered) {
		t.Fatal("outbox key does not distinguish state version or event")
	}
	if first == OutboxIdempotencyKey(AlertKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", RuleID: "cpu-high\x00x"}, 4, AlertEventFiring) {
		t.Fatal("outbox key does not distinguish embedded control characters")
	}
}

func TestDecodeOutboxKeyRestoresFullNotificationIdentity(t *testing.T) {
	key := AlertKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", RuleID: "cpu-high"}
	payload := map[string]any{"key": key}
	var decoded AlertKey
	if err := decodeOutboxKey(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != key {
		t.Fatalf("decoded key mismatch: got=%+v want=%+v", decoded, key)
	}
}

func TestDecodeOutboxKeyRejectsMissingIdentity(t *testing.T) {
	var key AlertKey
	if err := decodeOutboxKey(map[string]any{}, &key); err == nil {
		t.Fatal("missing outbox key accepted")
	}
}
