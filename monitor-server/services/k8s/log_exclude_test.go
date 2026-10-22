package k8s

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"
)

func excludedLogConfig(t *testing.T, exclude string) LogAgentConfig {
	t.Helper()
	var c LogAgentConfig
	if err := json.Unmarshal([]byte(`{"version":1,"targets":[{"id":"target","namespace":"sit-a","podUid":"pod-a","directory":"logs","glob":"*.log","start":"end","rules":[{"id":"error","version":1,"mode":"literal","pattern":"ERROR","exclude":`+exclude+`}]}]}`), &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLogConfigValidatesAndTransmitsExclusions(t *testing.T) {
	c := excludedLogConfig(t, `[{"mode":"regex","pattern":"(?i)healthcheck"},{"mode":"literal","pattern":"EXPECTED_ERROR"}]`)
	raw, err := EncodeLogAgentConfig(c)
	if err != nil || !strings.Contains(string(raw), `"exclude"`) || !strings.Contains(string(raw), "EXPECTED_ERROR") {
		t.Fatalf("exclusions lost: %s %v", raw, err)
	}
	for _, exclude := range []string{`[{"mode":"regex","pattern":"["}]`, `[{"mode":"literal","pattern":""}]`, `[{"mode":"unknown","pattern":"error"}]`} {
		invalid := excludedLogConfig(t, exclude)
		if ValidateLogAgentConfig(invalid) == nil {
			t.Errorf("invalid exclusion accepted: %s", exclude)
		}
		if (Store{}).SaveLogAgentConfig(context.Background(), "cluster-a", "node-a", invalid) == nil {
			t.Fatal("invalid exclusion submitted to database")
		}
	}
}

func TestLogConfigRejectsDuplicateRuleIDsBeforePublishing(t *testing.T) {
	c := excludedLogConfig(t, `[]`)
	c.Targets[0].Rules = append(c.Targets[0].Rules, c.Targets[0].Rules[0])
	if ValidateLogAgentConfig(c) == nil {
		t.Fatal("duplicate rule IDs would be rejected only by the Agent")
	}
}

func TestLogConfigSaveVersionsExclusionChangesAndRetries(t *testing.T) {
	original := excludedLogConfig(t, `[]`)
	original.Version = 5
	original.Targets[0].Rules[0].Version = 3
	incoming := excludedLogConfig(t, `[{"mode":"literal","pattern":"EXPECTED_ERROR"}]`)
	incoming.Version = 6
	// The caller's stale rule version is ignored; only the server versions semantics.
	incoming.Targets[0].Rules[0].Version = 3
	expected := excludedLogConfig(t, `[{"mode":"literal","pattern":"EXPECTED_ERROR"}]`)
	expected.Version = 6
	expected.Targets[0].Rules[0].Version = 4
	raw, err := EncodeLogAgentConfig(expected)
	if err != nil {
		t.Fatal(err)
	}
	steps := logStoreReadSteps(t, true, []string{"sit-a"}, original)
	steps = append(steps, logStoreStep{match: "INSERT INTO k8s_log_agent_config", args: []driver.Value{"cluster-a", "node-a", int64(6), raw}})
	store, script := newLogStoreTest(t, steps)
	if err = store.SaveLogAgentConfig(context.Background(), "cluster-a", "node-a", incoming); err != nil {
		t.Fatal(err)
	}
	if !script.committed {
		t.Fatal("configuration not committed")
	}
	// Retrying the identical request must neither reset counters nor write again.
	store, _ = newLogStoreTest(t, logStoreReadSteps(t, true, []string{"sit-a"}, expected))
	if err = store.SaveLogAgentConfig(context.Background(), "cluster-a", "node-a", incoming); err != nil {
		t.Fatal(err)
	}
}

func TestLogRuleVersionIsOwnedByServer(t *testing.T) {
	for _, version := range []uint64{0, 99} {
		original := excludedLogConfig(t, `[]`)
		original.Version = 5
		original.Targets[0].Rules[0].Version = 3
		incoming := excludedLogConfig(t, `[]`)
		incoming.Version = 6
		incoming.Targets[0].Rules[0].Version = version
		store, _ := newLogStoreTest(t, logStoreReadSteps(t, true, []string{"sit-a"}, original))
		if err := store.SaveLogAgentConfig(context.Background(), "cluster-a", "node-a", incoming); err != nil {
			t.Fatalf("caller rule version must not affect unchanged semantics: %v", err)
		}
	}
}

func TestExclusionsSurviveDiscoveryAndNodeMove(t *testing.T) {
	config := excludedLogConfig(t, `[{"mode":"literal","pattern":"EXPECTED_ERROR"}]`)
	target := &config.Targets[0]
	target.ID = "cmdb-app"
	target.WorkloadType, target.WorkloadName, target.PodName, target.Container = "StatefulSet", "app", "app-0", "main"
	defaultTarget := *target
	defaultTarget.Rules = []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}
	persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": config}}
	ConfigureLogTargetResolver(fixedLogResolver{target: defaultTarget})
	t.Cleanup(func() { ConfigureLogTargetResolver(nil) })
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", ContainerName: "main"}, PodUID: "pod-new", Node: "node-b", InstanceID: "instance"}}}
	for i := 0; i < 2; i++ {
		if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
			t.Fatal(err)
		}
	}
	for _, target := range persistence.configs["node-b"].Targets {
		if target.Format != "" {
			continue
		}
		raw, err := json.Marshal(target.Rules)
		if err != nil || !strings.Contains(string(raw), "EXPECTED_ERROR") || target.PodUID != "pod-new" {
			t.Fatalf("manual exclusion lost on discovery: %s %v", raw, err)
		}
		return
	}
	t.Fatal("log target missing from destination node")
}
