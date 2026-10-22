package k8s

import "testing"

func TestValidateLogAgentConfigRejectsUnsafeRuleAndTarget(t *testing.T) {
	bad := []LogAgentConfig{{Version: 1, Targets: []LogTarget{{ID: "x", Namespace: "sit", PodUID: "p", Directory: "../x", Glob: "*.log", Start: "end"}}}, {Version: 1, Targets: []LogTarget{{ID: "x", Namespace: "sit", PodUID: "p", Directory: "x", Glob: "*.log", Start: "end", Rules: []LogRule{{ID: "r", Version: 1, Mode: "regex", Pattern: "["}}}}}}
	for _, c := range bad {
		if ValidateLogAgentConfig(c) == nil {
			t.Fatal("invalid config accepted")
		}
	}
}
func TestEncodeLogAgentConfig(t *testing.T) {
	b, err := EncodeLogAgentConfig(LogAgentConfig{Version: 1, Targets: []LogTarget{{ID: "x", Namespace: "sit", PodUID: "p", Directory: "x", Glob: "*.log", Start: "end"}}})
	if err != nil || len(b) == 0 {
		t.Fatal(err)
	}
}
