package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func GetTypeConfigList(name string) (list []*models.TypeConfig, err error) {
	var endpointList []*models.EndpointNewTable
	if name == "" {
		err = x.SQL("select * from monitor_type").Find(&list)
	} else {
		sql := fmt.Sprintf("select * from monitor_type where display_name like '%%s%s'", name)
		err = x.SQL(sql).Find(&list)
	}
	if err != nil {
		return
	}
	if len(list) > 0 {
		if err = x.SQL("select * from endpoint_new ").Find(&endpointList); err != nil {
			return
		}
		if len(endpointList) > 0 {
			for _, typeConf := range list {
				for _, endpoint := range endpointList {
					if endpoint.MonitorType == typeConf.DisplayName && endpoint.MonitorType != "" {
						typeConf.ObjectCount++
					}
				}
			}
		}
	}
	return
}

func AddTypeConfig(param models.TypeConfig) (err error) {
	_, err = x.Exec("insert into monitor_type(guid,display_name,system_type,create_user,create_time) values(?,?,?,?,?)",
		param.Guid, param.DisplayName, param.SystemType, param.CreateUser, time.Now().Format(models.DatetimeFormat))
	return
}

func DeleteTypeConfig(id string) (err error) {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from endpoint_service_rel where endpoint in (select guid from endpoint_new where monitor_type=?)", Param: []interface{}{id}})
	actions = append(actions, &Action{Sql: "delete from endpoint_new where monitor_type=?", Param: []interface{}{id}})
	actions = append(actions, &Action{Sql: "delete from monitor_type where guid=?", Param: []interface{}{id}})
	return Transaction(actions)
}

func QueryTypeConfigByName(name string) (typeConfigList []*models.TypeConfig, err error) {
	err = x.SQL("select * from monitor_type where display_name=?", name).Find(&typeConfigList)
	return
}

func GetTypeConfig(id string) (typeConfig *models.TypeConfig, err error) {
	typeConfig = &models.TypeConfig{}
	_, err = x.SQL("select * from monitor_type where guid=?", id).Get(typeConfig)
	return
}
