package k8s

import (
	"context"
	"database/sql/driver"
	"github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestNamespaceCreateRetryReleasesTransactionBeforeReadingCurrentState(t *testing.T) {
	r := Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: false}
	store, script := newLogStoreTest(t, []logStoreStep{
		{match: "INSERT INTO k8s_namespace_source", err: &mysql.MySQLError{Number: 1062}},
		{match: "SELECT enabled,version,credential_guid,token", columns: []string{"enabled", "version", "credential_guid", "token"}, rows: [][]driver.Value{{false, int64(1), "", ""}}},
	})
	store.DB.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := store.Save(ctx, r); err != nil || !script.rolledBack {
		t.Fatalf("retry blocked by transaction: %v", err)
	}
}

func TestRegistrationRetryRequiresSameConfigurationAndAppliedVersion(t *testing.T) {
	credentials, err := PrepareCredentials("object-a", "test-seed", "https://api.example.invalid", "", "test-token")
	if err != nil {
		t.Fatal(err)
	}
	input := Registration{ClusterID: "cluster-a", Namespace: "sit-a", Enabled: true, Credentials: credentials}
	current := input
	current.Version = 1
	if !sameAppliedRegistration(current, input, "test-seed") {
		t.Fatal("identical retry rejected")
	}
	current.Version = 2
	if sameAppliedRegistration(current, input, "test-seed") {
		t.Fatal("stale retry accepted")
	}
	current.Version = 1
	current.Credentials.CA = "different-ca"
	if sameAppliedRegistration(current, input, "test-seed") {
		t.Fatal("different configuration accepted")
	}
	current.Credentials = credentials
	if sameAppliedRegistration(current, input, "wrong-seed") {
		t.Fatal("invalid credential accepted")
	}
	input.Enabled = false
	input.Version = 1
	current.Version = 2
	current.Enabled = false
	if !sameAppliedRegistration(current, input, "test-seed") {
		t.Fatal("delete retry rejected")
	}
	current.Enabled = true
	if sameAppliedRegistration(current, input, "test-seed") {
		t.Fatal("different enabled state accepted")
	}
}
