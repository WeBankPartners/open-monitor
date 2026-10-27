package k8s

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Requires an exclusively owned, migrated MySQL 8 database. The global
// cancellation operation deliberately rejects a nonempty outbox fixture.
func TestMySQLDisabledAlertCancellation(t *testing.T) {
	dsn := os.Getenv("MONITOR_K8S_TEST_DSN")
	if dsn == "" {
		t.Skip("MONITOR_K8S_TEST_DSN requires an exclusively owned migrated MySQL 8 database with an empty outbox")
	}
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal("open MySQL test database failed")
	}
	defer database.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var version string
	if err = database.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		t.Fatal("query MySQL test version failed")
	}
	if !strings.HasPrefix(version, "8.") {
		t.Fatalf("MySQL 8 is required, got %s", version)
	}
	var existing int
	if err = database.QueryRowContext(ctx, "SELECT COUNT(*) FROM k8s_alert_outbox").Scan(&existing); err != nil {
		t.Fatal(err)
	}
	if existing != 0 {
		t.Fatal("cancellation integration test requires an empty outbox in an exclusively owned database")
	}
	cluster := "cancel-test-" + uuid.NewString()
	defer database.Exec("DELETE FROM k8s_alert_outbox WHERE cluster_id=?", cluster)
	defer database.Exec("DELETE FROM k8s_alert_state WHERE cluster_id=?", cluster)
	store := AlertStore{DB: database}
	now := time.Now().UTC().Truncate(time.Second)
	if _, err = store.EvaluateAndPersist(ctx, AlertEvaluation{Key: AlertKey{ClusterID: cluster, Namespace: "scope", RuleID: "cpu"}, Rule: AlertRule{ForSeconds: 60}, Observation: Observation{Valid: true, Violated: true, PodUID: "pod"}}, now); err != nil {
		t.Fatal(err)
	}
	stateBefore := cancellationTableSnapshot(t, database, "k8s_alert_state", cluster)
	const historical, pending = 20000, 1003
	for start := 0; start < historical+pending; start += 500 {
		end := start + 500
		if end > historical+pending {
			end = historical + pending
		}
		var values []string
		var args []any
		for i := start; i < end; i++ {
			status, lastError := []string{"sent", "failed", "cancelled"}[i%3], "history retained"
			if i >= historical {
				status, lastError = "pending", "retry retained"
			}
			available := now.Add(time.Duration(i%2) * 365 * 24 * time.Hour)
			var leaseToken, leaseUntil any
			if i%3 != 0 {
				leaseToken, leaseUntil = "old-lease", now.Add(time.Duration(i%3-1)*24*time.Hour)
			}
			values = append(values, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
			args = append(args, fmt.Sprintf("cancel-event-%023d", i), fmt.Sprintf("cancel-key-%d", i), "cpu", cluster, "scope", []string{"firing", "recovered", "custom"}[i%3], `{"preserve":"payload"}`, status, 7, available, leaseToken, leaseUntil, lastError, now, now)
		}
		query := "INSERT INTO k8s_alert_outbox(event_id,idempotency_key,rule_id,cluster_id,namespace,event_type,payload,status,attempts,available_at,lease_token,lease_until,last_error,created_at,sent_at) VALUES "
		// Each fixture batch and each production cancellation batch has its own
		// bounded transaction; no transaction spans the historical data load.
		if _, err = database.ExecContext(ctx, query+strings.Join(values, ","), args...); err != nil {
			t.Fatal(err)
		}
	}
	before := cancellationTableSnapshot(t, database, "k8s_alert_outbox", cluster)
	verifyCancellationIndexAndLockBound(t, database)
	if count, batchErr := store.CancelPendingOutbox(ctx, 10000); batchErr != nil || count != 500 {
		t.Fatalf("partial batch: count=%d err=%v", count, batchErr)
	}
	// A process can stop after this commit. Two same-mode startup workers must
	// finish the remainder without treating another worker's locks as absence.
	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { <-start; results <- RunDisabledAlertMaintenance(ctx, Policy{}, store, nil) }()
	}
	close(start)
	for i := 0; i < 2; i++ {
		if err = <-results; err != nil {
			t.Fatal(err)
		}
	}
	after := cancellationTableSnapshot(t, database, "k8s_alert_outbox", cluster)
	if len(after) != historical+pending {
		t.Fatalf("outbox rows changed: %d", len(after))
	}
	for id, original := range before {
		if original["status"] == "pending" {
			original["status"], original["last_error"] = "cancelled", "local collection disabled"
			original["lease_token"], original["lease_until"] = "<NULL>", "<NULL>"
		}
		if !reflect.DeepEqual(original, after[id]) {
			t.Fatalf("event fields were not preserved for %s", id)
		}
	}
	if !reflect.DeepEqual(stateBefore, cancellationTableSnapshot(t, database, "k8s_alert_state", cluster)) {
		t.Fatal("cancellation changed alert state or duration/version fields")
	}
	if count, batchErr := store.CancelPendingOutbox(ctx, 500); batchErr != nil || count != 0 {
		t.Fatalf("repeat cancellation: count=%d err=%v", count, batchErr)
	}
	if !reflect.DeepEqual(after, cancellationTableSnapshot(t, database, "k8s_alert_outbox", cluster)) {
		t.Fatal("repeat cancellation changed terminal events")
	}
}

func cancellationTableSnapshot(t *testing.T, database *sql.DB, table, cluster string) map[string]map[string]string {
	t.Helper()
	// Table identifiers are fixed test constants, never external input.
	if table != "k8s_alert_outbox" && table != "k8s_alert_state" {
		t.Fatal("invalid snapshot table")
	}
	rows, err := database.Query("SELECT * FROM "+table+" WHERE cluster_id=?", cluster)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]map[string]string{}
	for rows.Next() {
		values, dest := make([]sql.RawBytes, len(columns)), make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}
		if err = rows.Scan(dest...); err != nil {
			t.Fatal(err)
		}
		record := map[string]string{}
		for i, column := range columns {
			if values[i] == nil {
				record[column] = "<NULL>"
			} else {
				record[column] = string(values[i])
			}
		}
		id := record["event_id"]
		if table == "k8s_alert_state" {
			id = record["rule_id"]
		}
		result[id] = record
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func verifyCancellationIndexAndLockBound(t *testing.T, database *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var planJSON string
	if err := database.QueryRowContext(ctx, "EXPLAIN FORMAT=JSON "+pendingOutboxCancellationSelect, "pending", 500).Scan(&planJSON); err != nil {
		t.Fatal(err)
	}
	t.Logf("MySQL cancellation EXPLAIN: %s", planJSON)
	var plan any
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		t.Fatal(err)
	}
	found := false
	var visit func(any)
	visit = func(node any) {
		switch item := node.(type) {
		case map[string]any:
			if item["using_filesort"] == true {
				t.Fatal("cancellation selection sorts and can scan/lock beyond the batch")
			}
			if item["table_name"] == "k8s_alert_outbox" {
				found = true
				if item["key"] != "idx_k8s_alert_outbox_claim" || (item["access_type"] != "ref" && item["access_type"] != "range") {
					t.Fatalf("unbounded cancellation access path: %s", planJSON)
				}
			}
			for _, child := range item {
				visit(child)
			}
		case []any:
			for _, child := range item {
				visit(child)
			}
		}
	}
	visit(plan)
	if !found {
		t.Fatal("EXPLAIN did not include cancellation table")
	}
	conn, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	readNext := func() int64 {
		var name string
		var count int64
		if queryErr := conn.QueryRowContext(ctx, "SHOW SESSION STATUS LIKE 'Handler_read_next'").Scan(&name, &count); queryErr != nil {
			t.Fatal(queryErr)
		}
		return count
	}
	startReads := readNext()
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, pendingOutboxCancellationSelect, "pending", 500)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	if delta := readNext() - startReads; count != 500 || delta > 500 {
		t.Fatalf("selection scanned beyond batch: rows=%d index_next_reads=%d", count, delta)
	}
	var nextID string
	if err = tx.QueryRowContext(ctx, "SELECT event_id FROM k8s_alert_outbox FORCE INDEX (idx_k8s_alert_outbox_claim) WHERE status=? ORDER BY available_at,lease_until,event_id LIMIT 1 OFFSET 500", "pending").Scan(&nextID); err != nil {
		t.Fatal(err)
	}
	// A pending row after the limit and a historical terminal row must both
	// remain independently lockable while the first 500 rows are locked.
	for _, id := range []string{nextID, fmt.Sprintf("cancel-event-%023d", 0)} {
		probeCtx, probeCancel := context.WithTimeout(ctx, time.Second)
		probe, probeErr := database.BeginTx(probeCtx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if probeErr != nil {
			probeCancel()
			t.Fatal(probeErr)
		}
		var locked string
		probeErr = probe.QueryRowContext(probeCtx, "SELECT event_id FROM k8s_alert_outbox WHERE event_id=? FOR UPDATE", id).Scan(&locked)
		probe.Rollback()
		probeCancel()
		if probeErr != nil {
			t.Fatalf("cancellation scan locked outside its batch: %v", probeErr)
		}
	}
}
