package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"net/http"
	"strings"
	"time"
)

func ListDBKeywordConfig(listType, listGuid string) (result []*models.ListDbKeywordData, err error) {
	if listType == "endpoint" {
		result, err = GetDbKeywordByEndpoint(listGuid, false)
	} else {
		result, err = GetDbKeywordByServiceGroup(listGuid)
	}
	return
}

func GetDbKeywordByEndpoint(endpointGuid string, onlySource bool) (result []*models.ListDbKeywordData, err error) {
	result = []*models.ListDbKeywordData{}
	var logKeywordMonitorTable []*models.LogKeywordMonitorTable
	if onlySource {
		err = x.SQL("select distinct t2.service_group from db_keyword_endpoint_rel t1 left join db_keyword_monitor t2 on t1.db_keyword_monitor=t2.guid where t1.source_endpoint=?", endpointGuid).Find(&logKeywordMonitorTable)
	} else {
		err = x.SQL("select distinct t2.service_group from db_keyword_endpoint_rel t1 left join db_keyword_monitor t2 on t1.db_keyword_monitor=t2.guid where t1.source_endpoint=? or t1.target_endpoint=?", endpointGuid, endpointGuid).Find(&logKeywordMonitorTable)
	}
	if err != nil {
		return result, fmt.Errorf("Query table fail,%s ", err.Error())
	}
	for _, v := range logKeywordMonitorTable {
		tmpResult, tmpErr := GetDbKeywordByServiceGroup(v.ServiceGroup)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		result = append(result, tmpResult[0])
	}
	return
}

func GetDbKeywordByServiceGroup(serviceGroupGuid string) (result []*models.ListDbKeywordData, err error) {
	serviceGroupObj, getErr := getSimpleServiceGroup(serviceGroupGuid)
	if getErr != nil {
		return result, getErr
	}
	result = []*models.ListDbKeywordData{}
	var dbKeywordTable []*models.DbKeywordMonitor
	err = x.SQL("select * from db_keyword_monitor where service_group=?", serviceGroupGuid).Find(&dbKeywordTable)
	if err != nil {
		return result, fmt.Errorf("Query table fail,%s ", err.Error())
	}
	configList := []*models.DbKeywordConfigObj{}
	for _, v := range dbKeywordTable {
		configObj := models.DbKeywordConfigObj{DbKeywordMonitor: *v}
		if configObj.EndpointRel, err = ListDbKeywordEndpointRel(v.Guid); err != nil {
			return
		}
		if configObj.Notify, err = GetDbKeywordNotify(v.Guid); err != nil {
			return
		}
		configList = append(configList, &configObj)
	}
	result = append(result, &models.ListDbKeywordData{
		Guid:        serviceGroupObj.Guid,
		DisplayName: serviceGroupObj.DisplayName,
		Description: serviceGroupObj.Description,
		ServiceType: serviceGroupObj.ServiceType,
		UpdateTime:  serviceGroupObj.UpdateTime,
		Config:      configList,
	})
	return
}

func ListDbKeywordEndpointRel(dbKeywordMonitorGuid string) (result []*models.DbKeywordEndpointRel, err error) {
	err = x.SQL("select * from db_keyword_endpoint_rel where db_keyword_monitor=?", dbKeywordMonitorGuid).Find(&result)
	return
}

func GetDbKeywordNotify(dbKeywordMonitorGuid string) (result *models.NotifyObj, err error) {
	var notifyRows []*models.NotifyTable
	err = x.SQL("select * from notify where guid in (select notify from db_keyword_notify_rel where db_keyword_monitor=?)", dbKeywordMonitorGuid).Find(&notifyRows)
	if err != nil {
		return
	}
	if len(notifyRows) > 0 {
		result = buildNotifyObj(notifyRows[0])
	}
	return
}

func buildNotifyObj(notifyRow *models.NotifyTable) (notifyObj *models.NotifyObj) {
	notifyObj = &models.NotifyObj{
		Guid:             notifyRow.Guid,
		NotifyRoles:      getNotifyRoles(notifyRow.Guid),
		EndpointGroup:    notifyRow.EndpointGroup,
		ServiceGroup:     notifyRow.ServiceGroup,
		AlarmStrategy:    notifyRow.AlarmStrategy,
		AlarmAction:      notifyRow.AlarmAction,
		AlarmPriority:    notifyRow.AlarmPriority,
		NotifyNum:        notifyRow.NotifyNum,
		ProcCallbackName: notifyRow.ProcCallbackName,
		ProcCallbackKey:  notifyRow.ProcCallbackKey,
		CallbackUrl:      notifyRow.CallbackUrl,
		CallbackParam:    notifyRow.CallbackParam,
		ProcCallbackMode: notifyRow.ProcCallbackMode,
		Description:      notifyRow.Description,
	}
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

func StartDbKeywordMonitorCronJob() {
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go doDbKeywordMonitorJob()
	}
}

func doDbKeywordMonitorJob() {
	http.DefaultClient.CloseIdleConnections()
	dataMap, err := datasource.QueryLogKeywordData()
	if err != nil {
		log.Logger.Error("Check log keyword break with get prometheus data", log.Error(err))
		return
	}
	if len(dataMap) == 0 {
		return
	}
	var logKeywordConfigs []*models.LogKeywordCronJobQuery
	x.SQL("select t1.guid,t1.service_group,t1.log_path,t1.monitor_type,t2.keyword,t2.notify_enable,t2.priority,t2.content,t2.name,t3.source_endpoint,t3.target_endpoint,t4.agent_address from log_keyword_monitor t1 left join log_keyword_config t2 on t1.guid=t2.log_keyword_monitor left join log_keyword_endpoint_rel t3 on t1.guid=t3.log_keyword_monitor left join endpoint_new t4 on t3.source_endpoint=t4.guid where t3.source_endpoint is not null").Find(&logKeywordConfigs)
	if len(logKeywordConfigs) == 0 {
		log.Logger.Debug("Check log keyword break with empty config ")
		return
	}
	var alarmTable []*models.LogKeywordAlarmTable
	err = x.SQL("select * from log_keyword_alarm").Find(&alarmTable)
	if err != nil {
		log.Logger.Error("Check log keyword break with query exist closed alarm fail", log.Error(err))
		return
	}
	alarmMap := make(map[string]*models.LogKeywordAlarmTable)
	for _, v := range alarmTable {
		if _, b := alarmMap[v.Tags]; b {
			continue
		}
		alarmMap[v.Tags] = v
	}
	var addAlarmRows []*models.AlarmTable
	var newValue, oldValue float64
	notifyMap := make(map[string]string)
	nowTime := time.Now()
	for _, config := range logKeywordConfigs {
		key := fmt.Sprintf("e_guid:%s^t_guid:%s^file:%s^keyword:%s", config.SourceEndpoint, config.TargetEndpoint, config.LogPath, config.Keyword)
		newValue, oldValue = 0, 0
		if dataValue, b := dataMap[key]; b {
			newValue = dataValue
		} else {
			continue
		}
		if newValue == 0 {
			continue
		}
		addFlag := false
		if existAlarm, b := alarmMap[key]; b {
			if existAlarm.EndValue > 0 {
				oldValue = existAlarm.EndValue
			} else {
				oldValue = existAlarm.StartValue
			}
			if newValue == oldValue {
				continue
			}
			if existAlarm.Status == "firing" {
				existAlarm.Content = strings.Split(existAlarm.Content, "^^")[0] + "^^" + getLogKeywordLastRow(config.AgentAddress, config.LogPath, config.Keyword)
				addAlarmRows = append(addAlarmRows, &models.AlarmTable{Id: existAlarm.AlarmId, Status: existAlarm.Status, EndValue: newValue, Content: existAlarm.Content, End: nowTime})
			} else {
				addFlag = true
			}
		} else {
			addFlag = true
		}
		if addFlag {
			if config.NotifyEnable > 0 {
				notifyMap[key] = config.ServiceGroup
			}
			alarmContent := config.Content
			if alarmContent != "" {
				alarmContent = alarmContent + "<br/>"
			}
			addAlarmRows = append(addAlarmRows, &models.AlarmTable{StrategyId: 0, Endpoint: config.TargetEndpoint, Status: "firing", SMetric: "log_monitor", SExpr: "node_log_monitor_count_total", SCond: ">0", SLast: "10s", SPriority: config.Priority, Content: alarmContent + getLogKeywordLastRow(config.AgentAddress, config.LogPath, config.Keyword), Tags: key, StartValue: newValue, Start: nowTime, AlarmName: config.Name})
		}
	}
	if len(addAlarmRows) == 0 {
		return
	}
	for _, v := range addAlarmRows {
		if tmpErr := doLogKeywordDBAction(v); tmpErr != nil {
			log.Logger.Error("Update log keyword alarm table fail", log.String("tags", v.Tags), log.Error(tmpErr))
		} else {
			if v.Id <= 0 {
				if _, b := notifyMap[v.Tags]; !b {
					log.Logger.Warn("Log keyword monitor notify disable,ignore", log.String("tags", v.Tags))
					continue
				}
				tmpAlarmObj := getSimpleAlarmByLogKeywordTags(v.Tags)
				if tmpAlarmObj.Id <= 0 {
					log.Logger.Warn("Log keyword monitor notify fail,query alarm with tags fail", log.String("tags", v.Tags))
					continue
				}
				NotifyServiceGroup(notifyMap[v.Tags], &models.AlarmHandleObj{AlarmTable: tmpAlarmObj})
			}
		}
	}
}
