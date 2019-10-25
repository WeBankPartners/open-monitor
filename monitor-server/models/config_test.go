package models

import (
	"testing"
	"encoding/json"
	"github.com/toolkits/file"
)

func TestConfigFileFormat(t *testing.T)  {
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
