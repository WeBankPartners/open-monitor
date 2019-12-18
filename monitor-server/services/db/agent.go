package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"time"
)

func UpdateEndpoint(endpoint *m.EndpointTable) error {
	host := m.EndpointTable{Guid:endpoint.Guid}
	GetEndpoint(&host)
	if host.Id == 0 {
		insert := fmt.Sprintf("INSERT INTO endpoint(guid,name,ip,endpoint_version,export_type,export_version,step,address,os_type,create_at) VALUE ('%s','%s','%s','%s','%s','%s','%d','%s','%s','%s')",
			endpoint.Guid,endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,time.Now().Format(m.DatetimeFormat))
		_,err := x.Exec(insert)
		if err != nil {
			mid.LogError("insert endpoint fail ", err)
			return err
		}
		host := m.EndpointTable{Guid:endpoint.Guid}
		GetEndpoint(&host)
		endpoint.Id = host.Id
	}else{
		update := fmt.Sprintf("UPDATE endpoint SET name='%s',ip='%s',endpoint_version='%s',export_type='%s',export_version='%s',step=%d,address='%s',os_type='%s' WHERE id=%d",
			endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,endpoint.Id)
		_,err := x.Exec(update)
		if err != nil {
			mid.LogError("update endpoint fail ", err)
			return err
		}
		endpoint.Id = host.Id
	}
	return nil
}

func DeleteEndpoint(guid string) error {
	_,err := x.Exec(fmt.Sprintf("DELETE FROM endpoint_metric WHERE endpoint_id IN (SELECT id FROM endpoint WHERE guid='%s')", guid))
	_,err = x.Exec(fmt.Sprintf("DELETE FROM endpoint WHERE guid='%s'", guid))
	if err != nil {
		mid.LogError("delete endpoint fail", err)
		return err
	}
	return nil
}

func UpdateEndpointAlarmFlag(isStop bool,exportType,instance,ip,port string) error {
	var endpoints []*m.EndpointTable
	if exportType == "host" {
		x.SQL("SELECT id FROM endpoint WHERE export_type=? AND ip=?", exportType, ip).Find(&endpoints)
	}else {
		if exportType == "tomcat" {
			x.SQL("SELECT id FROM endpoint WHERE export_type=? AND address=? AND name=?", exportType, fmt.Sprintf("%s:%s", ip, port), instance).Find(&endpoints)
		} else {
			x.SQL("SELECT id FROM endpoint WHERE export_type=? AND ip=? AND name=?", exportType, ip, instance).Find(&endpoints)
		}
	}
	mid.LogInfo(fmt.Sprintf("update endpoint alarm flag : query endpoints -> %v ", endpoints))
	if len(endpoints) > 0 {
		stopAlarm := "0"
		if isStop {
			stopAlarm = "1"
		}
		_,err := x.Exec(fmt.Sprintf("UPDATE endpoint SET stop_alarm=%s WHERE id=%d", stopAlarm, endpoints[0].Id))
		return err
	}else{
		return fmt.Errorf("Can not find this monitor object with %s %s %s %s \n", exportType,instance,ip,port)
	}
}