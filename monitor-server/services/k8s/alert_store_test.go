package k8s

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"
)

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
