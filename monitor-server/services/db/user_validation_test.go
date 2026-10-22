package db

import (
	"strings"
	"testing"
)

func TestCheckRoleIllegalIdentifiesRejectedRole(t *testing.T) {
	err := CheckRoleIllegal([]string{"missing-role"}, map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "missing-role") {
		t.Fatalf("missing rejected role: %v", err)
	}
}
