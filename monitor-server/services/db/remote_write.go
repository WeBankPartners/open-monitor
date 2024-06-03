package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"io/ioutil"
	"strings"
	"time"
)

func RemoteWriteConfigList() (result []*models.RemoteWriteConfigTable, err error) {
	result = []*models.RemoteWriteConfigTable{}
	err = x.SQL("select * from remote_write_config").Find(&result)
	return
}

func RemoteWriteConfigCreate(input models.RemoteWriteConfigTable) error {
	_, err := x.Exec("insert into remote_write_config(id,address,create_at) value (?,?,?)", input.Id, input.Address, time.Now())
	if err != nil {
		return fmt.Errorf("Insert database fail,%s ", err.Error())
	}
	err = SyncRemoteWritePrometheusConfig()
	if err != nil {
		return fmt.Errorf("Sync prometheus config fail,%s ", err.Error())
	}
	return nil
}

func RemoteWriteConfigUpdate(input models.RemoteWriteConfigTable) error {
	_, err := x.Exec("update remote_write_config set address=? where id=?", input.Address, input.Id)
	if err != nil {
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	err = SyncRemoteWritePrometheusConfig()
	if err != nil {
		return fmt.Errorf("Sync prometheus config fail,%s ", err.Error())
	}
	return nil
}

func RemoteWriteConfigDelete(id string) error {
	_, err := x.Exec("delete from remote_write_config where id=?", id)
	if err != nil {
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	err = SyncRemoteWritePrometheusConfig()
	if err != nil {
		return fmt.Errorf("Sync prometheus config fail,%s ", err.Error())
	}
	return nil
}

func SyncRemoteWritePrometheusConfig() error {
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
	promString := string(promBytes)
	startIndex := strings.Index(promString, "#Remote_write_start")
	endIndex := strings.Index(promString, "#Remote_write_end")
	tplBytes, err := ioutil.ReadFile("/app/monitor/prometheus/remote_write_prometheus.tpl")
	if err != nil {
		return fmt.Errorf("Read remote write prometheus tpl file fail,%s ", err.Error())
	}
	remoteWriteConfigString := ""
	tplString := string(tplBytes)
	for _, row := range remoteWriteConfigRows {
		tmpConfigString := tplString + "\n"
		tmpConfigString = strings.ReplaceAll(tmpConfigString, "{{remote_write_url}}", row.Address)
		remoteWriteConfigString += tmpConfigString + "\n"
	}
	promString = promString[:startIndex+11] + "\n" + remoteWriteConfigString + promString[endIndex:]
	err = ioutil.WriteFile("/app/monitor/prometheus/prometheus.yml", []byte(promString), 0644)
	if err != nil {
		err = fmt.Errorf("Write remote write config to prometheus fail,%s ", err.Error())
		recoverPrometheusConfig(backupConfigName)
		return err
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
