package k8s

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type cancellationDBScript struct {
	t                                     *testing.T
	query, update                         string
	queryArgs, updateArgs                 []driver.NamedValue
	rows                                  int64
	ids                                   []string
	rowsOpen                              bool
	queryErr, execErr, commitErr, rowsErr error
	begins, commits, rollbacks            int
}
type cancellationDBDriver struct{}
type cancellationDBConn struct{ script *cancellationDBScript }
type cancellationDBResult struct{ script *cancellationDBScript }

var cancellationDBScripts sync.Map

func init() { sql.Register("k8s-alert-cancellation-test", cancellationDBDriver{}) }
func (cancellationDBDriver) Open(name string) (driver.Conn, error) {
	script, _ := cancellationDBScripts.Load(name)
	return &cancellationDBConn{script: script.(*cancellationDBScript)}, nil
}
func (c *cancellationDBConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c *cancellationDBConn) Close() error { return nil }
func (c *cancellationDBConn) Begin() (driver.Tx, error) {
	return nil, errors.New("explicit isolation required")
}
func (c *cancellationDBConn) BeginTx(_ context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if opts.Isolation != driver.IsolationLevel(sql.LevelReadCommitted) || opts.ReadOnly {
		c.script.t.Fatalf("unexpected cancellation transaction options: %+v", opts)
	}
	c.script.begins++
	return c, nil
}
func (c *cancellationDBConn) Commit() error   { c.script.commits++; return c.script.commitErr }
func (c *cancellationDBConn) Rollback() error { c.script.rollbacks++; return nil }
func (c *cancellationDBConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.script.rowsOpen {
		return nil, errors.New("rows must be closed before update")
	}
	c.script.update, c.script.updateArgs = query, args
	return cancellationDBResult{script: c.script}, c.script.execErr
}
func (c *cancellationDBConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.script.query, c.script.queryArgs = query, args
	var rows [][]driver.Value
	for _, id := range c.script.ids {
		rows = append(rows, []driver.Value{id})
	}
	c.script.rowsOpen = true
	return &logStoreRows{columns: []string{"event_id"}, rows: rows, onClose: func() { c.script.rowsOpen = false }}, c.script.queryErr
}
func (r cancellationDBResult) LastInsertId() (int64, error) {
	return 0, errors.New("unexpected insert")
}
func (r cancellationDBResult) RowsAffected() (int64, error) { return r.script.rows, r.script.rowsErr }

func newCancellationDBTest(t *testing.T) (AlertStore, *cancellationDBScript) {
	t.Helper()
	script := &cancellationDBScript{t: t, rows: 1, ids: []string{"event-a"}}
	cancellationDBScripts.Store(t.Name(), script)
	database, err := sql.Open("k8s-alert-cancellation-test", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close(); cancellationDBScripts.Delete(t.Name()) })
	return AlertStore{DB: database}, script
}

func TestCancelPendingOutboxUsesBoundedIndexedStatusConditionalUpdate(t *testing.T) {
	for _, limit := range []int{1, 500, 501} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			store, script := newCancellationDBTest(t)
			count, err := store.CancelPendingOutbox(context.Background(), limit)
			if err != nil || count != 1 || script.begins != 1 || script.commits != 1 || script.rollbacks != 0 {
				t.Fatalf("batch was not committed: count=%d err=%v script=%+v", count, err, script)
			}
			if !strings.Contains(script.query, "FORCE INDEX (idx_k8s_alert_outbox_claim)") || !strings.Contains(script.query, "WHERE status=?\nORDER BY available_at,lease_until,event_id LIMIT ? FOR UPDATE") || strings.Contains(script.query, "SKIP LOCKED") {
				t.Fatalf("cancellation is not bounded by the pending index: %s", script.query)
			}
			if !strings.Contains(script.update, "SET status=?,last_error=?,lease_token=NULL,lease_until=NULL WHERE status=? AND event_id IN (?)") {
				t.Fatalf("unexpected cancellation mutations: %s", script.update)
			}
			if !strings.Contains(script.update, "FORCE INDEX (PRIMARY)") {
				t.Fatalf("update does not use bounded primary keys: %s", script.update)
			}
			bound := int64(limit)
			if bound > 500 {
				bound = 500
			}
			var values []driver.Value
			for _, arg := range script.queryArgs {
				values = append(values, arg.Value)
			}
			if !reflect.DeepEqual(values, []driver.Value{"pending", bound}) {
				t.Fatalf("unexpected selection bindings: %#v", values)
			}
			values = nil
			for _, arg := range script.updateArgs {
				values = append(values, arg.Value)
			}
			if !reflect.DeepEqual(values, []driver.Value{"cancelled", "local collection disabled", "pending", "event-a"}) {
				t.Fatalf("unexpected update bindings: %#v", values)
			}
		})
	}
}

func TestCancelPendingOutboxEmptyBatchCommitsWithoutUpdate(t *testing.T) {
	store, script := newCancellationDBTest(t)
	script.ids = nil
	count, err := store.CancelPendingOutbox(context.Background(), 500)
	if err != nil || count != 0 || script.commits != 1 || script.update != "" {
		t.Fatalf("empty batch mutated outbox or failed completion: count=%d err=%v script=%+v", count, err, script)
	}
}

func TestCancelPendingOutboxPropagatesBatchFailure(t *testing.T) {
	for _, phase := range []string{"select", "update", "rows", "commit"} {
		t.Run(phase, func(t *testing.T) {
			store, script := newCancellationDBTest(t)
			upstream := errors.New("database batch failure")
			switch phase {
			case "select":
				script.queryErr = upstream
			case "update":
				script.execErr = upstream
			case "rows":
				script.rowsErr = upstream
			case "commit":
				script.commitErr = upstream
			}
			count, err := store.CancelPendingOutbox(context.Background(), 500)
			if count != 0 || !errors.Is(err, upstream) {
				t.Fatalf("batch failure was lost: count=%d err=%v", count, err)
			}
			if phase != "commit" && (script.commits != 0 || script.rollbacks != 1) {
				t.Fatalf("failed batch did not roll back: %+v", script)
			}
		})
	}
}

func TestCancelPendingOutboxRejectsInvalidArguments(t *testing.T) {
	if _, err := (AlertStore{}).CancelPendingOutbox(context.Background(), 500); err == nil {
		t.Fatal("missing database accepted")
	}
	store, script := newCancellationDBTest(t)
	for _, limit := range []int{0, -1} {
		if _, err := store.CancelPendingOutbox(context.Background(), limit); err == nil {
			t.Fatalf("invalid batch limit accepted: %d", limit)
		}
	}
	if script.begins != 0 {
		t.Fatal("invalid batch opened a transaction")
	}
}
