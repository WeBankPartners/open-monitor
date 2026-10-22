package k8s

import "testing"

func TestRebindLogAgentConfigRefreshesPodUIDAndDropsMissing(t *testing.T) {
	c := LogAgentConfig{Version: 2, Targets: []LogTarget{{ID: "old", Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "main", PodUID: "pod-old", WorkloadUID: "work-old", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}, {ID: "gone", Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "gone", PodName: "gone-0", Container: "main", PodUID: "pod-gone", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}
	s := Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", ContainerName: "main"}, WorkloadUID: "work-new", PodUID: "pod-new", Node: "node-a"}}}
	out, changed, err := RebindLogAgentConfig(c, s, "node-a")
	if err != nil || !changed || out.Version != 3 || len(out.Targets) != 1 || out.Targets[0].PodUID != "pod-new" || out.Targets[0].WorkloadUID != "work-new" {
		t.Fatalf("unexpected result: %#v changed=%v err=%v", out, changed, err)
	}
}

func TestRebindLogAgentConfigsMovesTargetToNewNode(t *testing.T) {
	configs := map[string]LogAgentConfig{
		"node-a": {Version: 2, Targets: []LogTarget{{ID: "target", Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "main", PodUID: "pod-old", WorkloadUID: "work", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}},
	}
	snapshot := Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", ContainerName: "main"}, WorkloadUID: "work", PodUID: "pod-new", Node: "node-b"}}}
	out, changed, err := RebindLogAgentConfigs(configs, snapshot)
	if err != nil || !changed {
		t.Fatalf("unexpected result changed=%v err=%v", changed, err)
	}
	if len(out["node-a"].Targets) != 0 || out["node-a"].Version != 3 {
		t.Fatalf("source node was not updated: %#v", out["node-a"])
	}
	if len(out["node-b"].Targets) != 1 || out["node-b"].Targets[0].PodUID != "pod-new" || out["node-b"].Version != 1 {
		t.Fatalf("destination node was not created: %#v", out["node-b"])
	}
	second, changed, err := RebindLogAgentConfigs(out, snapshot)
	if err != nil || changed {
		t.Fatalf("rebind was not idempotent: changed=%v err=%v", changed, err)
	}
	if len(second["node-b"].Targets) != 1 || second["node-b"].Targets[0].PodUID != "pod-new" {
		t.Fatalf("idempotent result changed: %#v", second["node-b"])
	}
}

func TestRebindLogAgentConfigsForNamespacePreservesOtherNamespaces(t *testing.T) {
	configs := map[string]LogAgentConfig{"node-a": {Version: 2, Targets: []LogTarget{
		{ID: "sit", Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "main", PodUID: "old", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "r", Version: 1, Mode: "literal", Pattern: "ERROR"}}},
		{ID: "uat", Namespace: "uat", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", Container: "main", PodUID: "uat", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "r", Version: 1, Mode: "literal", Pattern: "ERROR"}}},
	}}}
	out, changed, err := RebindLogAgentConfigsForNamespace(configs, Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit", WorkloadType: "StatefulSet", WorkloadName: "app", PodName: "app-0", ContainerName: "main"}, PodUID: "new", Node: "node-a"}}}, "sit")
	if err != nil || !changed || len(out["node-a"].Targets) != 2 {
		t.Fatalf("unexpected result: %#v changed=%v err=%v", out, changed, err)
	}
	if out["node-a"].Targets[0].Namespace != "uat" || out["node-a"].Targets[1].PodUID != "new" {
		t.Fatalf("namespace isolation failed: %#v", out["node-a"])
	}
}
