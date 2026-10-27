package models

import (
	"bytes"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/toolkits/file"
	"log"
	"os"
	"strings"
	"testing"
)

func TestConfigFileFormat(t *testing.T) {
	defaultFile := "../conf/default.json"
	configContent, err := file.ToTrimString(defaultFile)
	if err != nil {
		t.Errorf("read config file: %s fail: %v ", defaultFile, err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		t.Errorf("parse config file: %s fail: %v ", defaultFile, err)
	}
}

func TestKubernetesLocalCollectionEnvironment(t *testing.T) {
	for _, tc := range []struct {
		value            string
		enabled, warning bool
	}{
		{"", false, false}, {"N", false, false}, {"false", false, false},
		{" Y ", true, false}, {"TrUe", true, false}, {" true\n", true, false},
		{"yes", false, true}, {"1", false, true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("MONITOR_K8S_LOCAL_COLLECTION_ENABLED", tc.value)
			var output bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&output)
			defer log.SetOutput(previous)
			p := k8s.Policy{Enabled: true, LocalCollectionEnabled: true}
			configureKubernetesLocalCollection(&p)
			if p.LocalCollectionActive() != tc.enabled {
				t.Fatal("unexpected environment override")
			}
			if strings.Contains(output.String(), "invalid MONITOR_K8S_LOCAL_COLLECTION_ENABLED") != tc.warning {
				t.Fatalf("unexpected warning: %s", output.String())
			}
		})
	}
	t.Setenv("MONITOR_K8S_LOCAL_COLLECTION_ENABLED", "")
	if err := os.Unsetenv("MONITOR_K8S_LOCAL_COLLECTION_ENABLED"); err != nil {
		t.Fatal(err)
	}
	p := k8s.Policy{Enabled: true, LocalCollectionEnabled: true}
	configureKubernetesLocalCollection(&p)
	if p.LocalCollectionActive() {
		t.Fatal("unset must disable collection")
	}
}
