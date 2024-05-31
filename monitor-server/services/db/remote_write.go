package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
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

	return nil
}
