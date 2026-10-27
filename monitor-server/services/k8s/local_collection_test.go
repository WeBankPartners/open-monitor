package k8s

import (
	"encoding/json"
	"testing"
)

func TestLocalCollectionRequiresBothSwitches(t *testing.T) {
	for _, tc := range []struct{ master, local, want bool }{
		{false, false, false}, {false, true, false},
		{true, false, false}, {true, true, true},
	} {
		p := Policy{Enabled: tc.master, LocalCollectionEnabled: tc.local}
		if p.LocalCollectionActive() != tc.want {
			t.Fatalf("master=%v local=%v: unexpected mode", tc.master, tc.local)
		}
	}
}

func TestLocalCollectionCannotBeEnabledThroughJSON(t *testing.T) {
	var p Policy
	if err := json.Unmarshal([]byte(`{"enabled":true,"localCollectionEnabled":true,"LocalCollectionEnabled":true}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.LocalCollectionActive() {
		t.Fatal("JSON enabled environment-only collection switch")
	}
}
