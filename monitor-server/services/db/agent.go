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
		insert := fmt.Sprintf("INSERT INTO endpoint(guid,name,ip,endpoint_version,export_type,export_version,step,address,os_type,create_at,address_agent) VALUE ('%s','%s','%s','%s','%s','%s','%d','%s','%s','%s','%s')",
			endpoint.Guid,endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,time.Now().Format(m.DatetimeFormat),endpoint.AddressAgent)
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
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM endpoint_metric WHERE endpoint_id IN (SELECT id FROM endpoint WHERE guid=?)", Param:[]interface{}{guid}})
	actions = append(actions, &Action{Sql:"DELETE FROM endpoint WHERE guid=?", Param:[]interface{}{guid}})
	var alarms []*m.AlarmTable
	x.SQL("SELECT id FROM alarm WHERE endpoint=? AND status='firing'", guid).Find(&alarms)
	if len(alarms) > 0 {
		var ids string
		for _,v := range alarms {
			ids += fmt.Sprintf("%d,", v.Id)
		}
		ids = ids[:len(ids)-1]
		actions = append(actions, &Action{Sql:fmt.Sprintf("UPDATE alarm SET STATUS='closed' WHERE id IN (%s)", ids), Param:[]interface{}{}})
	}
	err := Transaction(actions)
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

func UpdateRecursivePanel(param m.PanelRecursiveTable) error {
	var prt []*m.PanelRecursiveTable
	err := x.SQL("SELECT * FROM panel_recursive WHERE guid=?", param.Guid).Find(&prt)
	if err != nil {
		return err
	}
	if len(prt) > 0 {
		_,err = x.Exec("UPDATE panel_recursive SET display_name=?,children=?,endpoint=?,endpoint_type=? WHERE guid=?", param.DisplayName,param.Children,param.Endpoint,param.EndpointType,param.Guid)
	}else{
		_,err = x.Exec("INSERT INTO panel_recursive(guid,display_name,children,endpoint,endpoint_type) VALUE (?,?,?,?,?)", param.Guid,param.DisplayName,param.Children,param.Endpoint,param.EndpointType)
	}
	return err
}