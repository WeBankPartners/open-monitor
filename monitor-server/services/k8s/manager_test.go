package k8s

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakePersistence struct {
	claimed   bool
	published bool
	scanErr   error
	configs   map[string]LogAgentConfig
	saved     int
}

func (f *fakePersistence) Claim(context.Context, Registration, time.Duration) (string, error) {
	if f.claimed {
		return "", ErrConflict
	}
	f.claimed = true
	return "ticket", nil
}
func (f *fakePersistence) Publish(_ context.Context, _ Registration, _ string, _ *Snapshot, err error) error {
	f.published = true
	f.scanErr = err
	return nil
}
func (f *fakePersistence) GetLogAgentConfig(_ context.Context, _, node string) (LogAgentConfig, error) {
	return f.configs[node], nil
}
func (f *fakePersistence) SaveLogAgentConfig(_ context.Context, _, node string, config LogAgentConfig) error {
	if f.configs == nil {
		f.configs = map[string]LogAgentConfig{}
	}
	f.configs[node] = config
	f.saved++
	return nil
}
func (f *fakePersistence) ListLogAgentNodes(context.Context, string) ([]string, error) {
	if len(f.configs) == 0 {
		return []string{"node-a"}, nil
	}
	nodes := make([]string, 0, len(f.configs))
	for node := range f.configs {
		nodes = append(nodes, node)
	}
	return nodes, nil
}
func TestPersistentScanPublishesFailureWithoutSuccessSnapshot(t *testing.T) {
	persistence := &fakePersistence{}
	source := persistentSource{Persistence: persistence, Registration: Registration{ClusterID: "cluster-a", Namespace: "sit-a"}, DiscoverySource: &failingSource{}}
	snapshot, err := source.Snapshot(context.Background(), "cluster-a")
	if err == nil || snapshot != nil || !persistence.published || persistence.scanErr == nil {
		t.Fatal("scan failure not persisted correctly")
	}
	if _, err = source.Snapshot(context.Background(), "cluster-a"); !errors.Is(err, ErrConflict) {
		t.Fatal("claim conflict ignored")
	}
}

type snapshotSource struct{ snapshot *Snapshot }

func (s *snapshotSource) Snapshot(context.Context, string) (*Snapshot, error) { return s.snapshot, nil }
func (*snapshotSource) Watch(context.Context, string, string, func()) error   { return nil }

func TestPersistentScanRebindsPublishedLogAgentTargets(t *testing.T) {
	persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 4, Targets: []LogTarget{{ID: "target", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", WorkloadUID: "old-workload", PodUID: "old-pod", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}}}}
	source := persistentSource{Persistence: persistence, Registration: Registration{ClusterID: "cluster-a", Namespace: "sit-a"}, DiscoverySource: &snapshotSource{snapshot: &Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, WorkloadUID: "new-workload", PodUID: "new-pod", Node: "node-a"}}}}}
	if _, err := source.Snapshot(context.Background(), "cluster-a"); err != nil {
		t.Fatal(err)
	}
	if persistence.saved != 1 || persistence.configs["node-a"].Version != 5 || persistence.configs["node-a"].Targets[0].PodUID != "new-pod" {
		t.Fatalf("log target was not rebound: %+v", persistence.configs["node-a"])
	}
}

type fixedLogResolver struct{ target LogTarget }

func (r fixedLogResolver) Resolve(context.Context, Container) (LogTarget, error) {
	return r.target, nil
}

func TestPersistentScanIncrementsVersionForAutomaticTargetChange(t *testing.T) {
	old := LogTarget{ID: "cmdb-app", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", PodUID: "old-pod", WorkloadUID: "old-workload", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}
	persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 4, Targets: []LogTarget{old}}}}
	ConfigureLogTargetResolver(fixedLogResolver{target: LogTarget{ID: "cmdb-app", Namespace: old.Namespace, WorkloadType: old.WorkloadType, WorkloadName: old.WorkloadName, PodName: old.PodName, Container: old.Container, PodUID: "new-pod", WorkloadUID: "new-workload", Directory: old.Directory, Glob: old.Glob, Start: old.Start}})
	t.Cleanup(func() { ConfigureLogTargetResolver(nil) })
	source := persistentSource{Persistence: persistence, Registration: Registration{ClusterID: "cluster-a", Namespace: "sit-a"}, DiscoverySource: &snapshotSource{snapshot: &Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: old.Namespace, WorkloadType: old.WorkloadType, WorkloadName: old.WorkloadName, PodName: old.PodName, ContainerName: old.Container}, WorkloadUID: "new-workload", PodUID: "new-pod", Node: "node-a", InstanceID: "instance"}}}}}
	if _, err := source.Snapshot(context.Background(), "cluster-a"); err != nil {
		t.Fatal(err)
	}
	if persistence.configs["node-a"].Version != 5 {
		t.Fatalf("automatic change did not advance version: %+v", persistence.configs["node-a"])
	}
}

func TestPersistentScanReportsLogTargetResolutionFailure(t *testing.T) {
	var reported Container
	var reportedErr error
	ConfigureLogTargetResolverReport(func(container Container, err error) { reported, reportedErr = container, err })
	t.Cleanup(func() {
		ConfigureLogTargetResolver(nil)
		ConfigureLogTargetResolverReport(nil)
	})
	ConfigureLogTargetResolver(failingLogResolver{})
	persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 1}}}
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", PodName: "app-0", ContainerName: "main"}, Node: "node-a", InstanceID: "instance-a"}}}
	if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
		t.Fatal(err)
	}
	if reported.InstanceID != "instance-a" || reportedErr == nil {
		t.Fatalf("resolution failure was not reported: %+v %v", reported, reportedErr)
	}
}

type failingLogResolver struct{}

func (failingLogResolver) Resolve(context.Context, Container) (LogTarget, error) {
	return LogTarget{}, errors.New("CMDB app instance not found")
}

type failingSource struct{}

func (*failingSource) Snapshot(context.Context, string) (*Snapshot, error) {
	return nil, errors.New("unavailable")
}
func (*failingSource) Watch(context.Context, string, string, func()) error { return nil }

func TestAutomaticLogReconcileIsIdempotentAndRevokesOldNode(t *testing.T) {
	target := LogTarget{ID: "cmdb-app", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"}
	persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 4, Targets: []LogTarget{target}}}}
	ConfigureLogTargetResolver(fixedLogResolver{target: target})
	t.Cleanup(func() { ConfigureLogTargetResolver(nil) })
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, Node: "node-b", PodUID: "pod", InstanceID: "instance"}}}
	if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
		t.Fatal(err)
	}
	if persistence.configs["node-a"].Version != 5 || len(persistence.configs["node-a"].Targets) != 0 || len(persistence.configs["node-b"].Targets) != 2 {
		t.Fatalf("migration failed: %+v", persistence.configs)
	}
	saved := persistence.saved
	if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
		t.Fatal(err)
	}
	if persistence.saved != saved {
		t.Fatalf("unchanged scan wrote configurations: %d -> %d", saved, persistence.saved)
	}
}

func TestAutomaticLogReconcilePreservesExistingIDsAcrossPrefixChange(t *testing.T) {
	for _, id := range []string{"cmdb-app", "k8s-log-app"} {
		t.Run(id, func(t *testing.T) {
			target := LogTarget{ID: id, Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", WorkloadUID: "work-a", PodName: "orders-0", Container: "app", PodUID: "pod-a", InstanceID: "instance-a", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 3, Mode: "literal", Pattern: "ERROR", Exclude: []LogExclusion{{Mode: "literal", Pattern: "EXPECTED_ERROR"}}}}}
			metricTarget := target
			metricTarget.ID = id + "-metric"
			metricTarget.Glob, metricTarget.Format, metricTarget.Rules = "*_metric.log", "otlp_metric", nil
			persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 4, Targets: []LogTarget{target, metricTarget}}}}
			generated := target
			generated.ID = "k8s-log-app"
			generated.Rules = []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}
			ConfigureLogTargetResolver(fixedLogResolver{target: generated})
			t.Cleanup(func() { ConfigureLogTargetResolver(nil) })
			snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, Node: "node-a", PodUID: "pod-a", WorkloadUID: "work-a", InstanceID: "instance-a"}}}
			if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
				t.Fatal(err)
			}
			if persistence.saved != 0 || persistence.configs["node-a"].Version != 4 || !sameLogTargets(persistence.configs["node-a"].Targets, []LogTarget{target, metricTarget}) {
				t.Fatalf("prefix change rewrote existing configuration: %+v", persistence.configs)
			}

			// A rebuilt Pod on another node keeps both target IDs and custom rules.
			snapshot.Containers[0].Node, snapshot.Containers[0].PodUID = "node-b", "pod-b"
			generated.PodUID = "pod-b"
			ConfigureLogTargetResolver(fixedLogResolver{target: generated})
			if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
				t.Fatal(err)
			}
			target.PodUID, metricTarget.PodUID = "pod-b", "pod-b"
			if len(persistence.configs["node-a"].Targets) != 0 || persistence.configs["node-a"].Version != 5 || persistence.configs["node-b"].Version != 1 || !sameLogTargets(persistence.configs["node-b"].Targets, []LogTarget{target, metricTarget}) {
				t.Fatalf("migration changed target identity or rules: %+v", persistence.configs)
			}
			saved := persistence.saved
			if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
				t.Fatal(err)
			}
			if persistence.saved != saved {
				t.Fatal("unchanged discovery rewrote migrated configuration")
			}
		})
	}
}

func TestLogTargetComparisonIncludesRules(t *testing.T) {
	a := []LogTarget{{ID: "target", Rules: []LogRule{{ID: "rule", Version: 1, Mode: "literal", Pattern: "ERROR"}}}}
	b := []LogTarget{{ID: "target", Rules: []LogRule{{ID: "rule", Version: 2, Mode: "literal", Pattern: "FATAL"}}}}
	if sameLogTargets(a, b) {
		t.Fatal("rule change was ignored")
	}
}

type atomicLogPersistence struct {
	*fakePersistence
	updates int
}

func (s *atomicLogPersistence) UpdateLogAgentConfigs(ctx context.Context, cluster string, update func(map[string]LogAgentConfig) (map[string]LogAgentConfig, error)) error {
	s.updates++
	next, err := update(cloneLogAgentConfigs(s.configs))
	if err == nil {
		s.configs = next
	}
	return err
}
func TestReconcileUsesAtomicClusterMerge(t *testing.T) {
	store := &atomicLogPersistence{fakePersistence: &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 1}}}}
	if err := reconcileLogAgents(context.Background(), store, "cluster-a", "sit-a", &Snapshot{}); err != nil {
		t.Fatal(err)
	}
	if store.updates != 1 {
		t.Fatal("reconciliation bypassed atomic cluster update")
	}
}

func TestDiscoveryDesiredScopesRequireEnabledCluster(t *testing.T) {
	namespace := NamespaceState{Registration: Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true, Version: 2}}
	cluster := NamespaceState{Registration: Registration{ClusterID: "cluster-a", Enabled: true, Version: 4}}
	for _, tc := range []struct {
		name     string
		clusters []NamespaceState
		want     int
	}{{"enabled", []NamespaceState{cluster}, 1}, {"missing", nil, 0}, {"disabled", []NamespaceState{{Registration: Registration{ClusterID: "cluster-a", Enabled: false, Version: 4}}}, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			desired := desiredDiscoveryScopes(Policy{}, []NamespaceState{namespace}, tc.clusters)
			if len(desired) != tc.want {
				t.Fatalf("desired scopes: %+v", desired)
			}
		})
	}
	desired := desiredDiscoveryScopes(Policy{}, []NamespaceState{namespace}, []NamespaceState{cluster})
	first := desired[scope{"cluster-a", "sit-a"}]
	cluster.Version++
	desired = desiredDiscoveryScopes(Policy{}, []NamespaceState{namespace}, []NamespaceState{cluster})
	if first.sameVersion(desired[scope{"cluster-a", "sit-a"}]) {
		t.Fatal("cluster version change did not invalidate worker")
	}
}

func TestLogTargetOrderDoesNotChangeVersion(t *testing.T) {
	a := LogTarget{ID: "a", Namespace: "a"}
	b := LogTarget{ID: "b", Namespace: "b"}
	before := map[string]LogAgentConfig{"node": {Version: 4, Targets: []LogTarget{a, b}}}
	after, changed := applyLogAgentVersions(before, map[string]LogAgentConfig{"node": {Targets: []LogTarget{b, a}}})
	if changed || after["node"].Version != 4 {
		t.Fatal("order-only change advanced version")
	}
}

func TestAutoDiscoveryCreatesSeparateMetricTargetAndPreservesBothOnRescan(t *testing.T) {
	target := LogTarget{ID: "k8s-log-app", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", Container: "app", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "error", Version: 1, Mode: "literal", Pattern: "ERROR"}}}
	persistence := &fakePersistence{configs: map[string]LogAgentConfig{"node-a": {Version: 1}}}
	ConfigureLogTargetResolver(fixedLogResolver{target: target})
	t.Cleanup(func() { ConfigureLogTargetResolver(nil) })
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, Node: "node-a", PodUID: "pod", InstanceID: "instance"}}}
	for i := 0; i < 2; i++ {
		if err := reconcileLogAgents(context.Background(), persistence, "cluster-a", "sit-a", snapshot); err != nil {
			t.Fatal(err)
		}
		config := persistence.configs["node-a"]
		if len(config.Targets) != 2 || config.Version != 2 {
			t.Fatalf("targets/version: %+v", config)
		}
		ordinaryFound := false
		for _, target := range config.Targets {
			if target.Format == "otlp_metric" && (target.Glob != "*_metric.log" || len(target.Rules) != 0) {
				t.Fatalf("invalid metric target: %+v", target)
			}
			if target.Format == "" {
				ordinaryFound = true
				if len(target.Rules) != 1 || target.Rules[0].Pattern != "ERROR" {
					t.Fatalf("default error rule was not preserved: %+v", target)
				}
			}
		}
		if !ordinaryFound {
			t.Fatal("ordinary default target was not created")
		}
	}
}
