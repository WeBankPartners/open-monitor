package k8s

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
)

func TestPublishBindsSnapshotJSONAsText(t *testing.T) {
	snapshot := &Snapshot{WorkloadVersion: "v1", Workloads: []Workload{}}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	store, _ := newLogStoreTest(t, []logStoreStep{{
		match: "SET snapshot=?",
		args:  []driver.Value{string(raw), "cluster-a", "namespace-a", int64(3), "lease-ticket"},
	}})
	registration := Registration{ClusterID: "cluster-a", Namespace: "namespace-a", Enabled: true, Version: 3}
	if err := store.Publish(context.Background(), registration, "lease-ticket", snapshot, nil); err != nil {
		t.Fatalf("publish snapshot: %v", err)
	}
}
