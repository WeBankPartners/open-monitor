package k8s

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"os"
	"testing"
	"time"
)

// This test needs a disposable MySQL 8 database with the migration applied.
func TestMySQLRegistrationAndFencedPublish(t *testing.T) {
	dsn := os.Getenv("MONITOR_K8S_TEST_DSN")
	if dsn == "" {
		t.Skip("MONITOR_K8S_TEST_DSN requires a disposable migrated MySQL 8 database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := Store{DB: db}
	ctx := context.Background()
	r := Registration{ClusterID: "integration-test", Namespace: "test-scope", Enabled: true}
	defer db.Exec("DELETE FROM k8s_namespace_source WHERE cluster_id=? AND namespace=?", r.ClusterID, r.Namespace)
	if err = store.Save(ctx, r); err != nil {
		t.Fatal(err)
	}
	r.Version = 1
	if err = store.Save(ctx, Registration{ClusterID: r.ClusterID, Namespace: r.Namespace}); err != ErrConflict {
		t.Fatalf("expected conflict: %v", err)
	}
	ticket, err := store.Claim(ctx, r, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Claim(ctx, r, time.Minute); err != ErrConflict {
		t.Fatalf("concurrent claim accepted: %v", err)
	}
	r.Enabled = false
	if err = store.Save(ctx, r); err != nil {
		t.Fatal(err)
	}
	if err = store.Publish(ctx, r, ticket, &Snapshot{}, nil); err != ErrConflict {
		t.Fatalf("disabled namespace accepted old scan: %v", err)
	}
	rows, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row.ClusterID == r.ClusterID && row.Namespace == r.Namespace {
			found = true
			if row.Enabled || row.Version != 2 || row.Snapshot != nil {
				t.Fatalf("unexpected state: %+v", row)
			}
		}
	}
	if !found {
		t.Fatal("registration missing")
	}
}

func TestMySQLConcurrentLogConfigMerges(t *testing.T) {
	dsn := os.Getenv("MONITOR_K8S_TEST_DSN")
	if dsn == "" {
		t.Skip("MONITOR_K8S_TEST_DSN requires a disposable migrated MySQL 8 database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := Store{DB: db}
	ctx := context.Background()
	cluster := "log-config-integration-test"
	defer db.Exec("DELETE FROM k8s_metric_source WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_namespace_source WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_alert_rule WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_log_agent_config WHERE cluster_id=?", cluster)
	if err = (ClusterStore{Store: store}).Save(ctx, Registration{ClusterID: cluster, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	for _, ns := range []string{"scope-a", "scope-b"} {
		if err = store.Save(ctx, Registration{ClusterID: cluster, Namespace: ns, Enabled: true}); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, ns := range []string{"scope-a", "scope-b"} {
		go func(namespace string) {
			<-start
			results <- store.UpdateLogAgentConfigs(ctx, cluster, func(configs map[string]LogAgentConfig) (map[string]LogAgentConfig, error) {
				c := configs["node-a"]
				c.Targets = append(c.Targets, LogTarget{ID: namespace, Namespace: namespace, PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"})
				configs["node-a"] = c
				return configs, nil
			})
		}(ns)
	}
	close(start)
	for i := 0; i < 2; i++ {
		if err = <-results; err != nil {
			t.Fatal(err)
		}
	}
	config, err := store.GetLogAgentConfig(ctx, cluster, "node-a")
	if err != nil || config.Version != 2 || len(config.Targets) != 2 {
		t.Fatalf("concurrent merge lost targets: %+v %v", config, err)
	}
}

func TestMySQLLogAgentIdleRegistrationAndConcurrentPublish(t *testing.T) {
	dsn := os.Getenv("MONITOR_K8S_TEST_DSN")
	if dsn == "" {
		t.Skip("MONITOR_K8S_TEST_DSN requires a disposable migrated MySQL 8 database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store := Store{DB: db}
	cluster := "log-agent-idle-integration-test"
	defer db.Exec("DELETE FROM k8s_metric_source WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_namespace_source WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_alert_rule WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_log_agent_config WHERE cluster_id=?", cluster)
	defer db.Exec("DELETE FROM k8s_log_agent_endpoint WHERE cluster_id=?", cluster)
	if err = (ClusterStore{Store: store}).Save(ctx, Registration{ClusterID: cluster, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err = store.Save(ctx, Registration{ClusterID: cluster, Namespace: "scope-a", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	endpoint := LogAgentEndpoint{ClusterID: cluster, Node: "node-a", Address: net.IPv6loopback.String(), Port: 9105}
	heartbeat := func() error { return store.UpsertLogAgentEndpoint(ctx, endpoint) }
	target := LogTarget{ID: "target-a", Namespace: "scope-a", PodUID: "pod-a", Directory: "logs", Glob: "*.log", Start: "end"}
	for _, phase := range []struct {
		name    string
		write   func() error
		version uint64
		targets []LogTarget
	}{
		{name: "concurrent-registration", write: heartbeat, version: 1},
		{name: "concurrent-publish", write: func() error {
			return store.SaveLogAgentConfig(ctx, cluster, endpoint.Node, LogAgentConfig{Version: 2, Targets: []LogTarget{target}})
		}, version: 2, targets: []LogTarget{target}},
	} {
		start, results := make(chan struct{}), make(chan error, 2)
		for _, write := range []func() error{heartbeat, phase.write} {
			go func(write func() error) { <-start; results <- write() }(write)
		}
		close(start)
		for i := 0; i < 2; i++ {
			if err = <-results; err != nil {
				t.Fatalf("%s: %v", phase.name, err)
			}
		}
		config, readErr := store.GetLogAgentConfig(ctx, cluster, endpoint.Node)
		if readErr != nil || config.Version != phase.version || !sameLogTargets(config.Targets, phase.targets) {
			t.Fatalf("%s lost configuration: %+v %v", phase.name, config, readErr)
		}
	}
	if err = heartbeat(); err != nil {
		t.Fatal(err)
	}
	config, err := store.GetLogAgentConfig(ctx, cluster, endpoint.Node)
	if err != nil || config.Version != 2 || !sameLogTargets(config.Targets, []LogTarget{target}) {
		t.Fatalf("repeated heartbeat changed published configuration: %+v %v", config, err)
	}
	endpoint.Node = "disabled-node"
	if err = store.SetLogAgentEndpoint(ctx, endpoint, true); err != nil {
		t.Fatal(err)
	}
	if err = store.DisableLogAgentEndpoint(ctx, cluster, endpoint.Node); err != nil {
		t.Fatal(err)
	}
	if err = heartbeat(); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetLogAgentConfig(ctx, cluster, endpoint.Node); err != sql.ErrNoRows {
		t.Fatalf("disabled node received an initial configuration: %v", err)
	}
	endpoints, err := store.ListLogAgentEndpoints(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range endpoints {
		if row.Node == endpoint.Node && row.Enabled {
			t.Fatal("heartbeat re-enabled a disabled node")
		}
	}
}
