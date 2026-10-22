package k8s

import (
	"context"
	"testing"
)

func TestMetricsIntervalUsesSafeDefaultAndBounds(t *testing.T) {
	if got := (Policy{}).MetricsInterval(); got.Seconds() != DefaultMetricsIntervalSeconds {
		t.Fatalf("default interval=%s", got)
	}
	if got := (Policy{MetricsIntervalSeconds: 60}).MetricsInterval(); got.Seconds() != 60 {
		t.Fatalf("configured interval=%s", got)
	}
	if got := (Policy{MetricsIntervalSeconds: 10}).MetricsInterval(); got.Seconds() != DefaultMetricsIntervalSeconds {
		t.Fatalf("invalid interval=%s", got)
	}
	if got := (Policy{MetricsIntervalSeconds: 301}).MetricsInterval(); got.Seconds() != DefaultMetricsIntervalSeconds {
		t.Fatalf("out of range interval=%s", got)
	}
}

func TestRegistrationOnlyAcceptsConfiguredScope(t *testing.T) {
	policy := Policy{Environment: "sit", Clusters: map[string]ClusterConfig{"cluster-a": {Namespaces: map[string]NamespaceScope{"sit-a": {}}}}}
	for _, input := range []Registration{{ClusterID: "cluster-a", Namespace: "uat-a", Version: 1}, {ClusterID: "other", Namespace: "sit-a", Version: 1}, {ClusterID: "cluster-a", Namespace: "../sit-a", Version: 1}} {
		if policy.Validate(input) == nil {
			t.Errorf("accepted unconfigured scope: %+v", input)
		}
	}
	if err := policy.Validate(Registration{ClusterID: "cluster-a", Namespace: "sit-a", Version: 1}); err != nil {
		t.Fatal(err)
	}
}
func TestRegistrationUsesConflictVersionAndNoCredentialResponse(t *testing.T) {
	repository := &memoryRegistry{}
	service := Registry{Policy: Policy{Environment: "sit", Clusters: map[string]ClusterConfig{"cluster-a": {Namespaces: map[string]NamespaceScope{"sit-a": {}}}}}, Repository: repository}
	input := Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true, Version: 0}
	if err := service.Save(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if repository.row.Version != 1 {
		t.Fatal("initial version not assigned")
	}
	if err := service.Save(context.Background(), input); err == nil {
		t.Fatal("stale registration overwrote state")
	}
	input.Version = 1
	input.Enabled = false
	if err := service.Save(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if repository.row.Enabled || repository.row.Version != 2 {
		t.Fatal("disable not versioned")
	}
}

type memoryRegistry struct{ row Registration }

func (m *memoryRegistry) Save(ctx context.Context, r Registration) error {
	if r.Version != m.row.Version {
		return ErrConflict
	}
	r.Version++
	m.row = r
	return nil
}
