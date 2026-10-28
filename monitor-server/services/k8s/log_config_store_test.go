package k8s

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// The driver exercises Store's transaction and rollback paths without a server.
// MySQL locking itself is covered separately by the opt-in integration test.
type logStoreDriver struct{}
type logStoreConn struct {
	test     *logStoreScript
	rowsOpen bool
}
type logStoreScript struct {
	t                     *testing.T
	steps                 []logStoreStep
	committed, rolledBack bool
}
type logStoreStep struct {
	match   string
	columns []string
	rows    [][]driver.Value
	err     error
	args    []driver.Value
	check   func([]driver.NamedValue)
}
type logStoreRows struct {
	columns []string
	rows    [][]driver.Value
	onClose func()
}

var logStoreScripts sync.Map

func init() { sql.Register("k8s-log-config-test", logStoreDriver{}) }
func (logStoreDriver) Open(name string) (driver.Conn, error) {
	value, _ := logStoreScripts.Load(name)
	return &logStoreConn{test: value.(*logStoreScript)}, nil
}
func (c *logStoreConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c *logStoreConn) Close() error              { return nil }
func (c *logStoreConn) Begin() (driver.Tx, error) { return c, nil }
func (c *logStoreConn) Commit() error             { c.test.committed = true; return nil }
func (c *logStoreConn) Rollback() error           { c.test.rolledBack = true; return nil }
func (c *logStoreConn) step(query string, args []driver.NamedValue) logStoreStep {
	c.test.t.Helper()
	if len(c.test.steps) == 0 {
		c.test.t.Fatalf("unexpected SQL: %s", query)
	}
	step := c.test.steps[0]
	c.test.steps = c.test.steps[1:]
	if !strings.Contains(query, step.match) {
		c.test.t.Fatalf("SQL %s does not contain %s", query, step.match)
	}
	if step.args != nil {
		if len(args) != len(step.args) {
			c.test.t.Fatalf("argument count: %d", len(args))
		}
		for i, arg := range args {
			if !reflect.DeepEqual(arg.Value, step.args[i]) {
				c.test.t.Fatalf("argument %d mismatch", i)
			}
		}
	}
	if step.check != nil {
		step.check(args)
	}
	return step
}
func (c *logStoreConn) QueryContext(_ context.Context, q string, a []driver.NamedValue) (driver.Rows, error) {
	s := c.step(q, a)
	c.rowsOpen = s.err == nil
	return &logStoreRows{columns: s.columns, rows: s.rows, onClose: func() { c.rowsOpen = false }}, s.err
}
func (c *logStoreConn) ExecContext(_ context.Context, q string, a []driver.NamedValue) (driver.Result, error) {
	if c.rowsOpen {
		return nil, errors.New("result set must be closed before executing another statement")
	}
	s := c.step(q, a)
	return driver.RowsAffected(1), s.err
}
func (r *logStoreRows) Columns() []string { return r.columns }
func (r *logStoreRows) Close() error {
	if r.onClose != nil {
		r.onClose()
	}
	return nil
}
func (r *logStoreRows) Next(out []driver.Value) error {
	if len(r.rows) == 0 {
		return io.EOF
	}
	copy(out, r.rows[0])
	r.rows = r.rows[1:]
	return nil
}
func newLogStoreTest(t *testing.T, steps []logStoreStep) (Store, *logStoreScript) {
	t.Helper()
	script := &logStoreScript{t: t, steps: steps}
	logStoreScripts.Store(t.Name(), script)
	db, err := sql.Open("k8s-log-config-test", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		logStoreScripts.Delete(t.Name())
		if len(script.steps) != 0 {
			t.Errorf("%d SQL operations not executed", len(script.steps))
		}
	})
	return Store{DB: db}, script
}
func logStoreReadSteps(t *testing.T, enabled bool, namespaces []string, config LogAgentConfig) []logStoreStep {
	raw, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	scopes := [][]driver.Value{}
	for _, ns := range namespaces {
		scopes = append(scopes, []driver.Value{ns})
	}
	return []logStoreStep{
		{match: "FROM k8s_metric_source", columns: []string{"enabled"}, rows: [][]driver.Value{{enabled}}},
		{match: "FROM k8s_namespace_source", columns: []string{"namespace"}, rows: scopes},
		{match: "FROM k8s_log_agent_config", columns: []string{"node_name", "version", "config"}, rows: [][]driver.Value{{"node-a", int64(config.Version), raw}}},
	}
}
func TestLogConfigStoreRejectsStaleWrite(t *testing.T) {
	existing := LogAgentConfig{Version: 5}
	steps := logStoreReadSteps(t, true, []string{"sit-a"}, existing)
	store, script := newLogStoreTest(t, steps)
	incoming := LogAgentConfig{Version: 5, Targets: []LogTarget{{ID: "target", Namespace: "sit-a", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"}}}
	if err := store.SaveLogAgentConfig(context.Background(), "cluster-a", "node-a", incoming); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale write accepted: %v", err)
	}
	if script.committed || !script.rolledBack {
		t.Fatal("conflicting transaction not rolled back")
	}
}
func TestLogConfigStoreRevokesDisabledScopeOnRead(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "namespace-disabled", false: "cluster-disabled"}[enabled], func(t *testing.T) {
			original := LogAgentConfig{Version: 5, Targets: []LogTarget{{ID: "target", Namespace: "sit-a", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"}}}
			steps := logStoreReadSteps(t, enabled, nil, original)
			steps = append(steps, logStoreStep{match: "INSERT INTO k8s_log_agent_config"})
			store, script := newLogStoreTest(t, steps)
			got, err := store.GetLogAgentConfig(context.Background(), "cluster-a", "node-a")
			if err != nil || got.Version != 6 || len(got.Targets) != 0 {
				t.Fatalf("scope not revoked: %+v %v", got, err)
			}
			if !script.committed {
				t.Fatal("revocation was not persisted")
			}
		})
	}
}

func TestLogConfigTransactionPreservesOtherNamespaceAndRollsBackWriteFailure(t *testing.T) {
	target := LogTarget{ID: "other", Namespace: "sit-b", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"}
	original := LogAgentConfig{Version: 7, Targets: []LogTarget{target}}
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			steps := logStoreReadSteps(t, true, []string{"sit-a", "sit-b"}, original)
			writeErr := error(nil)
			if fail {
				writeErr = errors.New("write unavailable")
			}
			steps = append(steps, logStoreStep{match: "INSERT INTO k8s_log_agent_config", err: writeErr})
			store, script := newLogStoreTest(t, steps)
			var result map[string]LogAgentConfig
			err := store.UpdateLogAgentConfigs(context.Background(), "cluster-a", func(configs map[string]LogAgentConfig) (map[string]LogAgentConfig, error) {
				c := configs["node-a"]
				added := target
				added.ID = "new"
				added.Namespace = "sit-a"
				c.Targets = append(c.Targets, added)
				configs["node-a"] = c
				result = configs
				return configs, nil
			})
			if !errors.Is(err, writeErr) {
				t.Fatalf("lost original error: %v", err)
			}
			if fail {
				if script.committed || !script.rolledBack {
					t.Fatal("failed write committed")
				}
				return
			}
			if !script.committed || result["node-a"].Version != 8 || len(result["node-a"].Targets) != 2 || result["node-a"].Targets[0].Namespace != "sit-b" {
				t.Fatalf("merge lost another namespace: %+v", result)
			}
		})
	}
}

func TestLogConfigReadDoesNotWriteUnchangedConfiguration(t *testing.T) {
	config := LogAgentConfig{Version: 9, Targets: []LogTarget{{ID: "target", Namespace: "sit-a", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"}}}
	store, script := newLogStoreTest(t, logStoreReadSteps(t, true, []string{"sit-a"}, config))
	got, err := store.GetLogAgentConfig(context.Background(), "cluster-a", "node-a")
	if err != nil || got.Version != 9 || !script.committed {
		t.Fatalf("unchanged configuration: %+v %v", got, err)
	}
}

func TestLogConfigStoreBindsJSONAsText(t *testing.T) {
	target := LogTarget{ID: "target", Namespace: "sit-a", PodUID: "pod", Directory: "logs", Glob: "*.log", Start: "end"}
	steps := []logStoreStep{
		{match: "FROM k8s_metric_source", columns: []string{"enabled"}, rows: [][]driver.Value{{true}}},
		{match: "FROM k8s_namespace_source", columns: []string{"namespace"}, rows: [][]driver.Value{{"sit-a"}}},
		{match: "FROM k8s_log_agent_config", columns: []string{"node_name", "version", "config"}},
		{match: "INSERT INTO k8s_log_agent_config", check: func(args []driver.NamedValue) {
			if len(args) != 4 {
				t.Fatalf("argument count: %d", len(args))
			}
			raw, ok := args[3].Value.(string)
			if !ok {
				t.Fatalf("JSON config argument has type %T, want string", args[3].Value)
			}
			var got LogAgentConfig
			if err := json.Unmarshal([]byte(raw), &got); err != nil || got.Version != 1 || len(got.Targets) != 1 {
				t.Fatalf("invalid JSON config %q: %+v %v", raw, got, err)
			}
		}},
	}
	store, _ := newLogStoreTest(t, steps)
	if err := store.SaveLogAgentConfig(context.Background(), "cluster-a", "node-a", LogAgentConfig{Version: 1, Targets: []LogTarget{target}}); err != nil {
		t.Fatalf("save config: %v", err)
	}
}

func TestLogAgentRegistrationInitializesIdleConfigAtomically(t *testing.T) {
	endpoint := LogAgentEndpoint{ClusterID: "cluster-a", Node: "node-a", Address: net.IPv6loopback.String(), Port: 9105}
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			writeErr := error(nil)
			if fail {
				writeErr = errors.New("configuration storage unavailable")
			}
			steps := []logStoreStep{
				{match: "FROM k8s_metric_source WHERE cluster_id=? FOR UPDATE", columns: []string{"enabled"}, rows: [][]driver.Value{{true}}, args: []driver.Value{"cluster-a"}},
				{match: "INSERT INTO k8s_log_agent_endpoint"},
				{match: "INSERT INTO k8s_log_agent_config", err: writeErr, check: func(args []driver.NamedValue) {
					if len(args) != 3 || args[1].Value != "cluster-a" || args[2].Value != "node-a" {
						t.Fatalf("configuration scope: %+v", args)
					}
					raw, ok := args[0].Value.(string)
					var config LogAgentConfig
					if !ok || json.Unmarshal([]byte(raw), &config) != nil || config.Version != 1 || config.Targets == nil || len(config.Targets) != 0 {
						t.Fatalf("invalid idle configuration: %+v", args[0].Value)
					}
				}},
			}
			if !fail {
				steps = append(steps, logStoreReadSteps(t, true, nil, LogAgentConfig{Version: 1, Targets: []LogTarget{}})...)
			}
			store, script := newLogStoreTest(t, steps)
			if err := store.UpsertLogAgentEndpoint(context.Background(), endpoint); !errors.Is(err, writeErr) {
				t.Fatalf("registration error: %v", err)
			}
			if fail {
				if script.committed || !script.rolledBack {
					t.Fatal("registration with failed configuration initialization was not rolled back")
				}
				return
			}
			if !script.committed {
				t.Fatal("registration and idle configuration were not committed")
			}
			config, err := store.GetLogAgentConfig(context.Background(), endpoint.ClusterID, endpoint.Node)
			if err != nil || config.Version != 1 || len(config.Targets) != 0 {
				t.Fatalf("idle configuration unavailable: %+v %v", config, err)
			}
		})
	}
}

func TestLogAgentRegistrationDoesNotInitializeDisabledOrUnknownCluster(t *testing.T) {
	for _, exists := range []bool{true, false} {
		t.Run(map[bool]string{true: "disabled", false: "unknown"}[exists], func(t *testing.T) {
			steps := []logStoreStep{{match: "FROM k8s_metric_source", columns: []string{"enabled"}}}
			if exists {
				steps[0].rows = [][]driver.Value{{false}}
				steps = append(steps, logStoreStep{match: "INSERT INTO k8s_log_agent_endpoint"})
			}
			store, script := newLogStoreTest(t, steps)
			err := store.UpsertLogAgentEndpoint(context.Background(), LogAgentEndpoint{ClusterID: "cluster-a", Node: "node-a", Address: net.IPv6loopback.String(), Port: 9105})
			if exists && (err != nil || !script.committed) {
				t.Fatalf("disabled cluster heartbeat: %v", err)
			}
			if !exists && (!errors.Is(err, sql.ErrNoRows) || script.committed || !script.rolledBack) {
				t.Fatalf("unknown cluster registration: %v", err)
			}
		})
	}
}
