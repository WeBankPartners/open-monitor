package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RemoteWriteConfigList() (result []*models.RemoteWriteConfigTable, err error) {
	result = []*models.RemoteWriteConfigTable{}
	err = x.SQL("select * from remote_write_config").Find(&result)
	if err != nil {
		err = fmt.Errorf("query remote wirte config fail,%s ", err.Error())
		return
	}
	for _, row := range result {
		row.CreateTime = row.CreateAt.Format(models.DatetimeFormat)
		row.UpdateTime = row.UpdateAt.Format(models.DatetimeFormat)
	}
	return
}

func RemoteWriteConfigCreate(input models.RemoteWriteConfigTable, operator string) error {
	_, err := x.Exec("insert into remote_write_config(id,address,create_at,create_user) value (?,?,?,?)", input.Id, input.Address, time.Now(), operator)
	if err != nil {
		return fmt.Errorf("Insert database fail,%s ", err.Error())
	}
	CallSyncWritePrometheusConfig()
	return nil
}

func RemoteWriteConfigUpdate(input models.RemoteWriteConfigTable, operator string) error {
	_, err := x.Exec("update remote_write_config set address=?,update_user=? where id=?", input.Address, operator, input.Id)
	if err != nil {
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	CallSyncWritePrometheusConfig()
	return nil
}

func RemoteWriteConfigDelete(id string) error {
	_, err := x.Exec("delete from remote_write_config where id=?", id)
	if err != nil {
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	CallSyncWritePrometheusConfig()
	return nil
}

func CallSyncWritePrometheusConfig() {
	go func() {
		err := SyncRemoteWritePrometheusConfig()
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "SyncRemoteWritePrometheusConfig", zap.Error(err))
		} else {
			log.Info(nil, log.LOGGER_APP, "SyncRemoteWritePrometheusConfig done")
		}
	}()
}

// Database rows remain intact; repeated addresses must not produce repeated writes.
func renderRemoteWriteConfig(config, template string, rows []*models.RemoteWriteConfigTable) (string, int, error) {
	const startMarker, endMarker = "#Remote_write_start", "#Remote_write_end"
	start, end := strings.Index(config, startMarker), strings.Index(config, endMarker)
	if strings.Count(config, startMarker) != 1 || strings.Count(config, endMarker) != 1 || end < start {
		return "", 0, fmt.Errorf("invalid managed remote write markers")
	}
	seen := make(map[string]bool, len(rows))
	var entries strings.Builder
	duplicates := 0
	for _, row := range rows {
		address := strings.TrimSpace(row.Address)
		if address == "" {
			continue
		}
		if seen[address] {
			duplicates++
			continue
		}
		seen[address] = true
		// The packaged template uses a single-quoted YAML scalar.
		entries.WriteString(strings.ReplaceAll(template, "{{remote_write_url}}", strings.ReplaceAll(address, "'", "''")))
		entries.WriteByte('\n')
	}
	return config[:start+len(startMarker)] + "\n" + entries.String() + config[end:], duplicates, nil
}

func writeRemoteWriteConfig(name string, content []byte) error {
	f, err := os.CreateTemp(filepath.Dir(name), ".remote-write-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(content); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), name)
}

func SyncRemoteWritePrometheusConfig() error {
	prometheusConfigMutex.Lock()
	defer prometheusConfigMutex.Unlock()
	var remoteWriteConfigRows []*models.RemoteWriteConfigTable
	err := x.SQL("select * from remote_write_config where address<>''").Find(&remoteWriteConfigRows)
	if err != nil {
		return fmt.Errorf("Try to get remote write table fail,%s ", err.Error())
	}
	backupConfigName, backupErr := backupPrometheusConfig()
	if backupErr != nil {
		return backupErr
	}
	promBytes, err := ioutil.ReadFile("/app/monitor/prometheus/prometheus.yml")
	if err != nil {
		return fmt.Errorf("Read prometheus tpl file fail,%s ", err.Error())
	}
	tplBytes, err := ioutil.ReadFile("/app/monitor/prometheus/remote_write_prometheus.tpl")
	if err != nil {
		return fmt.Errorf("Read remote write prometheus tpl file fail,%s ", err.Error())
	}
	promString, duplicates, err := renderRemoteWriteConfig(string(promBytes), string(tplBytes), remoteWriteConfigRows)
	if err != nil {
		return err
	}
	if duplicates > 0 {
		log.Warn(nil, log.LOGGER_APP, "Duplicate remote write addresses omitted from generated config", zap.Int("duplicate_count", duplicates))
	}
	err = writeRemoteWriteConfig("/app/monitor/prometheus/prometheus.yml", []byte(promString))
	if err != nil {
		// Atomic replacement has not changed the original on failure.
		return fmt.Errorf("write remote write config to prometheus: %w", err)
	}
	err = prom.ReloadConfig()
	if err != nil {
		err = fmt.Errorf("Reload prometheus config fail,%s ", err.Error())
		recoverPrometheusConfig(backupConfigName)
		prom.ReloadConfig()
		return err
	}
	return nil
}
