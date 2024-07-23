package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func ListDBKeywordConfig(listType, listGuid string) (result *models.ListDbKeywordData, err error) {

	return
}

func GetDbKeywordByServiceGroup(serviceGroupGuid string) (result []*models.LogKeywordServiceGroupObj, err error) {
	serviceGroupObj, getErr := getSimpleServiceGroup(serviceGroupGuid)
	if getErr != nil {
		return result, getErr
	}
	result = []*models.LogKeywordServiceGroupObj{}
	var logKeywordTable []*models.LogKeywordMonitorTable
	err = x.SQL("select * from log_keyword_monitor where service_group=?", serviceGroupGuid).Find(&logKeywordTable)
	if err != nil {
		return result, fmt.Errorf("Query table fail,%s ", err.Error())
	}
	configList := []*models.LogKeywordMonitorObj{}
	for _, v := range logKeywordTable {
		configObj := models.LogKeywordMonitorObj{Guid: v.Guid, ServiceGroup: serviceGroupGuid, LogPath: v.LogPath, MonitorType: v.MonitorType}
		configObj.KeywordList = ListLogKeyword(v.Guid)
		configObj.EndpointRel = ListLogKeywordEndpointRel(v.Guid)
		configList = append(configList, &configObj)
	}
	result = append(result, &models.LogKeywordServiceGroupObj{ServiceGroupTable: serviceGroupObj, Config: configList})
	return
}

func CreateDBKeywordConfig(param *models.DbKeywordConfigObj, operator string) (err error) {
	actions := getCreateDbKeywordConfigActions(param, operator, time.Now())
	err = Transaction(actions)
	return
}

func getCreateDbKeywordConfigActions(input *models.DbKeywordConfigObj, operator string, nowTime time.Time) (actions []*Action) {
	input.Guid = "db_km_" + guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into db_keyword_monitor(guid,service_group,name,query_sql,priority,content,notify_enable,active_window,step,monitor_type,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		input.Guid, input.ServiceGroup, input.Name, input.QuerySql, input.Priority, input.Content, input.NotifyEnable, input.ActiveWindow, input.Step, input.MonitorType, operator, operator, nowTime, nowTime,
	}})
	endpointRelGuidList := guid.CreateGuidList(len(input.EndpointRel))
	for i, v := range input.EndpointRel {
		actions = append(actions, &Action{Sql: "insert into db_keyword_endpoint_rel(guid,db_keyword_monitor,source_endpoint,target_endpoint) values (?,?,?,?)", Param: []interface{}{
			endpointRelGuidList[i], input.Guid, v.SourceEndpoint, v.TargetEndpoint,
		}})
	}
	if input.Notify != nil {
		input.Notify.EndpointGroup = ""
		input.Notify.ServiceGroup = ""
		input.Notify.AlarmStrategy = ""
		notifyList := []*models.NotifyObj{input.Notify}
		actions = append(actions, getNotifyListUpdateAction(notifyList)...)
		actions = append(actions, &Action{Sql: "insert into db_keyword_notify_rel(guid,db_keyword_monitor,notify) values (?,?,?)", Param: []interface{}{
			"dkm_notify_" + guid.CreateGuid(), input.Guid, input.Notify.Guid,
		}})
	}
	return
}

func UpdateDBKeywordConfig(param *models.DbKeywordConfigObj, operator string) (err error) {
	actions := getUpdateDbKeywordConfigActions(param, operator, time.Now())
	err = Transaction(actions)
	return
}

func getUpdateDbKeywordConfigActions(input *models.DbKeywordConfigObj, operator string, nowTime time.Time) (actions []*Action) {
	actions = append(actions, &Action{Sql: "update db_keyword_monitor set name=?,query_sql=?,priority=?,content=?,notify_enable=?,active_window=?,step=?,monitor_type=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
		input.Name, input.QuerySql, input.Priority, input.Content, input.NotifyEnable, input.ActiveWindow, input.Step, input.MonitorType, operator, nowTime, input.Guid,
	}})
	actions = append(actions, &Action{Sql: "delete from db_keyword_endpoint_rel where db_keyword_monitor=?", Param: []interface{}{input.Guid}})
	endpointRelGuidList := guid.CreateGuidList(len(input.EndpointRel))
	for i, v := range input.EndpointRel {
		actions = append(actions, &Action{Sql: "insert into db_keyword_endpoint_rel(guid,db_keyword_monitor,source_endpoint,target_endpoint) values (?,?,?,?)", Param: []interface{}{
			endpointRelGuidList[i], input.Guid, v.SourceEndpoint, v.TargetEndpoint,
		}})
	}
	if input.Notify != nil {
		notifyGuid := input.Notify.Guid
		input.Notify.EndpointGroup = ""
		input.Notify.ServiceGroup = ""
		input.Notify.AlarmStrategy = ""
		notifyList := []*models.NotifyObj{input.Notify}
		actions = append(actions, getNotifyListUpdateAction(notifyList)...)
		if notifyGuid == "" {
			actions = append(actions, &Action{Sql: "delete from db_keyword_notify_rel where db_keyword_monitor=?", Param: []interface{}{input.Guid}})
			actions = append(actions, &Action{Sql: "insert into db_keyword_notify_rel(guid,db_keyword_monitor,notify) values (?,?,?)", Param: []interface{}{
				"dkm_notify_" + guid.CreateGuid(), input.Guid, input.Notify.Guid,
			}})
		}
	}
	return
}

func DeleteDBKeywordConfig(guid string) (err error) {
	actions, buildActionsErr := getDeleteDbKeywordConfigActions(guid)
	if buildActionsErr != nil {
		return buildActionsErr
	}
	if len(actions) > 0 {
		err = Transaction(actions)
	}
	return
}

func getDeleteDbKeywordConfigActions(dbKeywordConfigGuid string) (actions []*Action, err error) {
	dbKeywordMonitorObj, getErr := getSimpleDbKeywordConfig(dbKeywordConfigGuid, false)
	if getErr != nil {
		err = getErr
		return
	}
	if dbKeywordMonitorObj == nil {
		return
	}
	var dbKeywordNotifyRefRows []*models.DbKeywordNotifyRel
	err = x.SQL("select * from db_keyword_notify_rel where db_keyword_monitor=?", dbKeywordConfigGuid).Find(&dbKeywordNotifyRefRows)
	if err != nil {
		return
	}
	for _, row := range dbKeywordNotifyRefRows {
		actions = append(actions, &Action{Sql: "delete from db_keyword_notify_rel where guid=?", Param: []interface{}{row.Guid}})
		actions = append(actions, getNotifyDeleteAction(row.Notify)...)
	}
	actions = append(actions, &Action{Sql: "delete from db_keyword_endpoint_rel where db_keyword_monitor=?", Param: []interface{}{dbKeywordConfigGuid}})
	actions = append(actions, &Action{Sql: "delete from db_keyword_monitor where guid=?", Param: []interface{}{dbKeywordConfigGuid}})
	return
}

func getSimpleDbKeywordConfig(dbKeywordConfigGuid string, emptyCheck bool) (result *models.DbKeywordMonitor, err error) {
	var queryRows []*models.DbKeywordMonitor
	err = x.SQL("select * from db_keyword_monitor where guid=?", dbKeywordConfigGuid).Find(&queryRows)
	if err != nil {
		return
	}
	if len(queryRows) == 0 && emptyCheck {
		err = fmt.Errorf("can not find db_keyword_monitor with guid:%s ", dbKeywordConfigGuid)
		return
	}
	if len(queryRows) > 0 {
		result = queryRows[0]
	}
	return
}
