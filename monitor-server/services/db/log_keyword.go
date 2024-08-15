package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetLogKeywordByServiceGroup(serviceGroupGuid string) (result []*models.LogKeywordServiceGroupObj, err error) {
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
		if configObj.KeywordList, err = ListLogKeyword(v.Guid); err != nil {
			return
		}
		configObj.EndpointRel = ListLogKeywordEndpointRel(v.Guid)
		if configObj.Notify, err = GetLogKeywordNotify(v.Guid); err != nil {
			return
		}
		configList = append(configList, &configObj)
	}
	result = append(result, &models.LogKeywordServiceGroupObj{ServiceGroupTable: serviceGroupObj, Config: configList})
	return
}

func GetLogKeywordByEndpoint(endpointGuid string, onlySource bool) (result []*models.LogKeywordServiceGroupObj, err error) {
	result = []*models.LogKeywordServiceGroupObj{}
	var logKeywordMonitorTable []*models.LogKeywordMonitorTable
	if onlySource {
		err = x.SQL("select distinct t2.service_group from log_keyword_endpoint_rel t1 left join log_keyword_monitor t2 on t1.log_keyword_monitor=t2.guid where t1.source_endpoint=?", endpointGuid).Find(&logKeywordMonitorTable)
	} else {
		err = x.SQL("select distinct t2.service_group from log_keyword_endpoint_rel t1 left join log_keyword_monitor t2 on t1.log_keyword_monitor=t2.guid where t1.source_endpoint=? or t1.target_endpoint=?", endpointGuid, endpointGuid).Find(&logKeywordMonitorTable)
	}
	if err != nil {
		return result, fmt.Errorf("Query table fail,%s ", err.Error())
	}
	for _, v := range logKeywordMonitorTable {
		tmpResult, tmpErr := GetLogKeywordByServiceGroup(v.ServiceGroup)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		result = append(result, tmpResult[0])
	}
	return
}

func CreateLogKeywordMonitor(param *models.LogKeywordMonitorCreateObj) (err error) {
	var existLogKeywordMonitorRows []*models.LogKeywordMonitorTable
	err = x.SQL("select log_path from log_keyword_monitor where service_group=?", param.ServiceGroup).Find(&existLogKeywordMonitorRows)
	if err != nil {
		err = fmt.Errorf("query log keyword monitor fail,%s ", err.Error())
		return
	}
	existLogPathMap := make(map[string]int)
	for _, row := range existLogKeywordMonitorRows {
		existLogPathMap[row.LogPath] = 1
	}
	param.Guid = guid.CreateGuid()
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	logKeywordGuidList := guid.CreateGuidList(len(param.LogPath))
	for i, path := range param.LogPath {
		if _, existFlag := existLogPathMap[path]; existFlag {
			err = fmt.Errorf("Path:%s Already exists ", path)
			return
		} else {
			existLogPathMap[path] = 1
		}
		actions = append(actions, &Action{Sql: "insert into log_keyword_monitor(guid,service_group,log_path,monitor_type,update_time) value (?,?,?,?,?)", Param: []interface{}{logKeywordGuidList[i], param.ServiceGroup, path, param.MonitorType, nowTime}})
		endpointRelActions, tmpErr := getLogKeywordEndpointRelCreateAction(param.EndpointRel, logKeywordGuidList[i])
		if tmpErr != nil {
			err = tmpErr
			break
		}
		actions = append(actions, endpointRelActions...)
	}
	if err != nil {
		return err
	}
	return Transaction(actions)
}

func getLogKeywordEndpointRelCreateAction(param []*models.LogKeywordEndpointRelTable, logKeywordMonitor string) (actions []*Action, err error) {
	endpointRelGuidList := guid.CreateGuidList(len(param))
	for ii, endpointRel := range param {
		if endpointRel.SourceEndpoint == "" || endpointRel.TargetEndpoint == "" {
			err = fmt.Errorf("endpointRel source and target can not empty ")
			break
		}
		actions = append(actions, &Action{Sql: "insert into log_keyword_endpoint_rel(guid,log_keyword_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{endpointRelGuidList[ii], logKeywordMonitor, endpointRel.SourceEndpoint, endpointRel.TargetEndpoint}})
	}
	return
}

func UpdateLogKeywordMonitor(param *models.LogKeywordMonitorObj) (err error) {
	var existLogKeywordMonitorRows []*models.LogKeywordMonitorTable
	err = x.SQL("select guid,log_path from log_keyword_monitor where service_group=?", param.ServiceGroup).Find(&existLogKeywordMonitorRows)
	if err != nil {
		err = fmt.Errorf("query log keyword monitor fail,%s ", err.Error())
		return
	}
	for _, row := range existLogKeywordMonitorRows {
		if row.LogPath == param.LogPath && param.Guid != row.Guid {
			err = fmt.Errorf("Path:%s Already exists ", param.LogPath)
			return
		}
	}
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "update log_keyword_monitor set log_path=?,monitor_type=?,update_time=? where guid=?", Param: []interface{}{param.LogPath, param.MonitorType, nowTime, param.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=?", Param: []interface{}{param.Guid}})
	endpointRelActions, tmpErr := getLogKeywordEndpointRelCreateAction(param.EndpointRel, param.Guid)
	if tmpErr != nil {
		err = tmpErr
		return
	}
	actions = append(actions, endpointRelActions...)
	return Transaction(actions)
}

func DeleteLogKeywordMonitor(logKeywordMonitorGuid string) (err error) {
	var logKeywordMonitorTable []*models.LogKeywordMonitorTable
	err = x.SQL("select * from log_keyword_monitor where guid=?", logKeywordMonitorGuid).Find(&logKeywordMonitorTable)
	if len(logKeywordMonitorTable) == 0 {
		return
	}
	var hostEndpointList []string
	for _, v := range ListLogKeywordEndpointRel(logKeywordMonitorGuid) {
		hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
	}
	err = Transaction(getDeleteLogKeywordMonitorAction(logKeywordMonitorGuid))
	if err != nil {
		return err
	}
	err = SyncLogKeywordExporterConfig(hostEndpointList)
	return
}

func getDeleteLogKeywordMonitorAction(logKeywordMonitorGuid string) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=?", Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_config where log_keyword_monitor=?", Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_monitor where guid=?", Param: []interface{}{logKeywordMonitorGuid}})
	return actions
}

func ListLogKeyword(logKeywordMonitor string) (result []*models.LogKeywordConfigTable, err error) {
	result = []*models.LogKeywordConfigTable{}
	err = x.SQL("select * from log_keyword_config where log_keyword_monitor=?", logKeywordMonitor).Find(&result)
	if err != nil {
		return
	}
	var logKeywordNotifyRelRows []*models.LogKeywordNotifyRel
	err = x.SQL("select * from log_keyword_notify_rel where log_keyword_config in (select guid from log_keyword_config where log_keyword_monitor=?)", logKeywordMonitor).Find(&logKeywordNotifyRelRows)
	if err != nil {
		return
	}
	if len(logKeywordNotifyRelRows) == 0 {
		return
	}
	var notifyGuidList []string
	notifyMap := make(map[string]*models.NotifyTable)
	for _, row := range logKeywordNotifyRelRows {
		notifyGuidList = append(notifyGuidList, row.Notify)
		notifyMap[row.LogKeywordConfig] = &models.NotifyTable{Guid: row.Notify}
	}
	filterSql, filterParam := createListParams(notifyGuidList, "")
	var notifyRows []*models.NotifyTable
	err = x.SQL("select * from notify where guid in ("+filterSql+")", filterParam...).Find(&notifyRows)
	if err != nil {
		return
	}
	for _, v := range notifyMap {
		for _, row := range notifyRows {
			if v.Guid == row.Guid {
				v.NotifyNum = row.NotifyNum
				v.Description = row.Description
				v.ProcCallbackKey = row.ProcCallbackKey
				v.ProcCallbackName = row.ProcCallbackName
				v.ProcCallbackMode = row.ProcCallbackMode
				v.AlarmAction = row.AlarmAction
				v.AlarmPriority = row.AlarmPriority
				v.CallbackUrl = row.CallbackUrl
				v.CallbackParam = row.CallbackParam
				break
			}
		}
	}
	for _, row := range result {
		if v, ok := notifyMap[row.Guid]; ok {
			row.Notify = buildNotifyObj(v)
		}
		row.ActiveWindowList = strings.Split(row.ActiveWindow, ",")
	}
	return
}

func GetDbKeywordMonitorByName(guid, name, serviceGroup string) (list []*models.DbKeywordMonitor, err error) {
	list = []*models.DbKeywordMonitor{}
	if guid == "" {
		err = x.SQL("select * from db_keyword_monitor where name = ? and service_group=?", name, serviceGroup).Find(&list)
	} else {
		err = x.SQL("select * from db_keyword_monitor where name = ? and service_group=? and guid <> ?", name, serviceGroup, guid).Find(&list)
	}
	return
}

func GetLogKeywordConfigByName(guid, name, logKeywordMonitorGuid string) (list []*models.LogKeywordConfigTable, err error) {
	list = []*models.LogKeywordConfigTable{}
	if guid == "" {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and name=?", logKeywordMonitorGuid, name).Find(&list)
	} else {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and name=? and guid<>?", logKeywordMonitorGuid, name, guid).Find(&list)
	}
	return
}

func CreateLogKeyword(param *models.LogKeywordConfigTable, operator string) (err error) {
	var actions []*Action
	param.Guid = "lk_config_" + guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into log_keyword_config(guid,log_keyword_monitor,keyword,regulative,notify_enable,priority,update_time,content,name,active_window,update_user) value (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		param.Guid, param.LogKeywordMonitor, param.Keyword, param.Regulative, param.NotifyEnable, param.Priority, time.Now().Format(models.DatetimeFormat), param.Content, param.Name, param.ActiveWindow, operator}})
	if param.Notify != nil {
		actions = append(actions, getNotifyListInsertAction([]*models.NotifyObj{param.Notify})...)
		actions = append(actions, &Action{Sql: "insert into log_keyword_notify_rel(guid,log_keyword_config,notify) values (?,?,?)", Param: []interface{}{
			guid.CreateGuid(), param.Guid, param.Notify.Guid,
		}})
	}
	err = Transaction(actions)
	return
}

func GetSimpleLogKeywordConfig(logKeywordConfigGuid string) (result *models.LogKeywordConfigTable, err error) {
	var logKeywordConfigRows []*models.LogKeywordConfigTable
	err = x.SQL("select * from log_keyword_config where guid=?", logKeywordConfigGuid).Find(&logKeywordConfigRows)
	if err != nil {
		err = fmt.Errorf("query log keyword config table fail,%s ", err.Error())
		return
	}
	if len(logKeywordConfigRows) == 0 {
		err = fmt.Errorf("can not find log keyword config with guid:%s ", logKeywordConfigGuid)
		return
	}
	result = logKeywordConfigRows[0]
	return
}

func UpdateLogKeyword(param, existData *models.LogKeywordConfigTable, operator string) (err error) {
	var actions []*Action
	actions = append(actions, &Action{Sql: "update log_keyword_config set keyword=?,regulative=?,notify_enable=?,priority=?,update_time=?,content=?,name=?,active_window=?,update_user=? where guid=?", Param: []interface{}{
		param.Keyword, param.Regulative, param.NotifyEnable, param.Priority, time.Now().Format(models.DatetimeFormat), param.Content, param.Name, param.ActiveWindow, operator, param.Guid}})
	if param.Notify != nil {
		actions = append(actions, getNotifyListUpdateAction([]*models.NotifyObj{param.Notify})...)
		actions = append(actions, &Action{Sql: "delete from log_keyword_notify_rel where log_keyword_config=?", Param: []interface{}{param.Guid}})
		actions = append(actions, &Action{Sql: "insert into log_keyword_notify_rel(guid,log_keyword_config,notify) values (?,?,?)", Param: []interface{}{
			guid.CreateGuid(), param.Guid, param.Notify.Guid,
		}})
	}
	if existData.Name != param.Name || existData.Keyword != param.Keyword || existData.Priority != param.Priority {
		// 关键信息改了，把已有告警关闭
		var logKeywordAlarmRows []*models.LogKeywordAlarmTable
		err = x.SQL("select id,alarm_id from log_keyword_alarm where log_keyword_config=? and status='firing'", param.Guid).Find(&logKeywordAlarmRows)
		if err != nil {
			err = fmt.Errorf("query log keyword alarm table fail,%s ", err.Error())
			return
		}
		for _, row := range logKeywordAlarmRows {
			if row.AlarmId > 0 {
				closeAlarmActions, tmpErr := CloseAlarm(models.AlarmCloseParam{Id: row.AlarmId})
				if tmpErr != nil {
					err = fmt.Errorf("try to get close alarm actions fail,%s ", tmpErr.Error())
					return
				}
				if len(closeAlarmActions) > 0 {
					actions = append(actions, closeAlarmActions...)
				}
			}
		}
	}
	err = Transaction(actions)
	return
}

func DeleteLogKeyword(logKeywordConfigGuid string) (err error) {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify in (select notify from log_keyword_notify_rel where log_keyword_config=?)", Param: []interface{}{logKeywordConfigGuid}})
	actions = append(actions, &Action{Sql: "delete from notify where guid in (select notify from log_keyword_notify_rel where log_keyword_config=?)", Param: []interface{}{logKeywordConfigGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_notify_rel where log_keyword_config=?", Param: []interface{}{logKeywordConfigGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_config where guid=?", Param: []interface{}{logKeywordConfigGuid}})
	err = Transaction(actions)
	return
}

func ListLogKeywordEndpointRel(logKeywordMonitor string) (result []*models.LogKeywordEndpointRelTable) {
	result = []*models.LogKeywordEndpointRelTable{}
	x.SQL("select * from log_keyword_endpoint_rel where log_keyword_monitor=?", logKeywordMonitor).Find(&result)
	return
}

func StartLogKeywordMonitorCronJob() {
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go doLogKeywordMonitorJob()
	}
}

func doLogKeywordMonitorJob() {
	http.DefaultClient.CloseIdleConnections()
	dataMap, err := datasource.QueryLogKeywordData("log")
	if err != nil {
		log.Logger.Error("Check log keyword break with get prometheus data", log.Error(err))
		return
	}
	if len(dataMap) == 0 {
		return
	}
	var logKeywordConfigs []*models.LogKeywordCronJobQuery
	x.SQL("select t1.guid,t1.service_group,t1.log_path,t1.monitor_type,t2.keyword,t2.notify_enable,t2.priority,t2.content,t2.name,t2.guid as log_keyword_config_guid,t3.source_endpoint,t3.target_endpoint,t4.agent_address from log_keyword_monitor t1 left join log_keyword_config t2 on t1.guid=t2.log_keyword_monitor left join log_keyword_endpoint_rel t3 on t1.guid=t3.log_keyword_monitor left join endpoint_new t4 on t3.source_endpoint=t4.guid where t3.source_endpoint is not null").Find(&logKeywordConfigs)
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
	//notifyMap := make(map[string]string)
	nowTime := time.Now()
	notifyConfigMap := make(map[string]int)
	for _, config := range logKeywordConfigs {
		if config.LogKeywordConfigGuid == "" {
			continue
		}
		if config.NotifyEnable > 0 {
			notifyConfigMap[config.Guid] = 1
		}
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
			//if config.NotifyEnable > 0 {
			//	notifyMap[key] = config.ServiceGroup
			//}
			alarmContent := config.Content
			alarmContent = alarmContent + "<br/>"
			addAlarmRows = append(addAlarmRows, &models.AlarmTable{StrategyId: 0, Endpoint: config.TargetEndpoint, Status: "firing", SMetric: "log_monitor", SExpr: "node_log_monitor_count_total", SCond: ">0", SLast: "10s", SPriority: config.Priority, Content: alarmContent + getLogKeywordLastRow(config.AgentAddress, config.LogPath, config.Keyword), Tags: key, StartValue: newValue, Start: nowTime, AlarmName: config.Name, AlarmStrategy: config.LogKeywordConfigGuid})
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
				if _, b := notifyConfigMap[v.AlarmStrategy]; !b {
					log.Logger.Warn("Log keyword monitor notify disable,ignore", log.String("logKeywordConfig", v.AlarmStrategy))
					continue
				}
				tmpAlarmObj := getSimpleAlarmByLogKeywordTags(v.Tags)
				if tmpAlarmObj.Id <= 0 {
					log.Logger.Warn("Log keyword monitor notify fail,query alarm with tags fail", log.String("tags", v.Tags))
					continue
				}
				tmpNotifyRow, getNotifyErr := getLogKeywordAlarmNotify(v.AlarmStrategy)
				if getNotifyErr != nil {
					log.Logger.Error("doLogKeywordMonitorJob get alarm notify fail", log.String("logKeywordConfigGuid", v.AlarmStrategy), log.Int("alarmId", tmpAlarmObj.Id), log.Error(getNotifyErr))
					continue
				}
				if tmpNotifyRow.ProcCallbackMode == models.AlarmNotifyManualMode && tmpNotifyRow.ProcCallbackKey != "" {
					if _, execErr := x.Exec("update alarm set notify_id=? where id=?", tmpNotifyRow.Guid, tmpAlarmObj.Id); execErr != nil {
						log.Logger.Error("update alarm table notify id fail", log.Int("alarmId", tmpAlarmObj.Id), log.Error(execErr))
					}
				}
				notifyAction(tmpNotifyRow, &models.AlarmHandleObj{AlarmTable: tmpAlarmObj})
			}
		}
	}
}

func calcAlarmUniqueFlag(input *models.AlarmTable) {
	input.EndpointTags = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d_%s_%s_%d_%s", input.StrategyId, input.Endpoint, input.SMetric, input.Start.Unix(), input.Tags))))
}

func getSimpleAlarmByLogKeywordTags(tags string) (result models.AlarmTable) {
	var alarmTable []*models.AlarmTable
	x.SQL("select * from alarm where tags=? and status='firing' order by id desc limit 1", tags).Find(&alarmTable)
	if len(alarmTable) > 0 {
		result = *alarmTable[0]
	}
	return
}

func getLogKeywordLastRow(address, path, keyword string) string {
	var result string
	if address == "" || path == "" || keyword == "" {
		return result
	}
	param := models.LogKeywordRowsHttpDto{Path: path, Keyword: keyword}
	postData, _ := json.Marshal(param)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/log_keyword/rows", address), strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Get log keyword rows fail,new request error", log.Error(err))
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		log.Logger.Error("Get log keyword rows fail,response error", log.Error(respErr))
		return result
	}
	var responseData models.LogKeywordRowsHttpResult
	respBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(respBytes, &responseData)
	if err != nil {
		log.Logger.Error("Get log keyword rows fail,response data json unmarshal error", log.Error(err))
		return result
	}
	if responseData.Status != "ok" {
		log.Logger.Error("Get log keyword rows fail,response status error", log.String("status", responseData.Status), log.String("message", responseData.Message))
		return result
	}
	for _, v := range responseData.Data {
		result = v.Content
	}
	return result
}

func ImportLogKeyword(param *models.LogKeywordServiceGroupObj, operator string) (err error) {
	existSGs, getExistDataErr := GetLogKeywordByServiceGroup(param.Guid)
	if getExistDataErr != nil {
		return fmt.Errorf("get exist log keyword data fail,%s ", getExistDataErr.Error())
	}
	if len(existSGs) == 0 {
		return fmt.Errorf("get empty log keyword data,please check service group")
	}
	var actions []*Action
	var affectHostList []string
	existSGConfig := existSGs[0]
	for _, existKeywordConfig := range existSGConfig.Config {
		actions = append(actions, getDeleteLogKeywordMonitorAction(existKeywordConfig.Guid)...)
		for _, v := range existKeywordConfig.EndpointRel {
			affectHostList = append(affectHostList, v.SourceEndpoint)
		}
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	for _, inputKeywordConfig := range param.Config {
		actions = append(actions, &Action{Sql: "insert into log_keyword_monitor(guid,service_group,log_path,monitor_type,update_time,update_user) value (?,?,?,?,?,?)", Param: []interface{}{inputKeywordConfig.Guid, inputKeywordConfig.ServiceGroup, inputKeywordConfig.LogPath, inputKeywordConfig.MonitorType, nowTime, operator}})
		for _, keywordObj := range inputKeywordConfig.KeywordList {
			actions = append(actions, &Action{Sql: "insert into log_keyword_config(guid,log_keyword_monitor,keyword,regulative,notify_enable,priority,update_time) value (?,?,?,?,?,?,?)", Param: []interface{}{keywordObj.Guid, keywordObj.LogKeywordMonitor, keywordObj.Keyword, keywordObj.Regulative, keywordObj.NotifyEnable, keywordObj.Priority, nowTime}})
		}
	}
	err = Transaction(actions)
	if len(affectHostList) > 0 && err == nil {
		if syncErr := SyncLogKeywordExporterConfig(affectHostList); syncErr != nil {
			log.Logger.Error("import log keyword fail with sync host keyword config", log.Error(syncErr), log.StringList("hosts", affectHostList))
		}
	}
	return
}

func doLogKeywordDBAction(alarmObj *models.AlarmTable) (err error) {
	session := x.NewSession()
	if err = session.Begin(); err != nil {
		return
	}
	defer func() {
		if err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
		session.Close()
	}()
	if alarmObj.Id > 0 {
		_, err = session.Exec("UPDATE alarm SET content=?,end_value=?,end=? WHERE id=?", alarmObj.Content, alarmObj.EndValue, alarmObj.End.Format(models.DatetimeFormat), alarmObj.Id)
		if err != nil {
			return
		}
		var execResult sql.Result
		var execErr error
		if alarmObj.SMetric == "db_keyword_monitor" {
			execResult, execErr = session.Exec("UPDATE db_keyword_alarm SET content=?,end_value=?,updated_time=? WHERE alarm_id=?", alarmObj.Content, alarmObj.EndValue, alarmObj.End.Format(models.DatetimeFormat), alarmObj.Id)
		} else {
			execResult, execErr = session.Exec("UPDATE log_keyword_alarm SET content=?,end_value=?,updated_time=? WHERE alarm_id=?", alarmObj.Content, alarmObj.EndValue, alarmObj.End.Format(models.DatetimeFormat), alarmObj.Id)
		}
		if execErr != nil {
			err = execErr
		} else {
			if affectRowNum, _ := execResult.RowsAffected(); affectRowNum <= 0 {
				err = fmt.Errorf("update log keyword alarm table fail,affect row 0 with alarm_id=%d ", alarmObj.Id)
			}
		}
		return
	} else {
		calcAlarmUniqueFlag(alarmObj)
		execResult, execErr := session.Exec("INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start,tags,alarm_strategy,endpoint_tags,alarm_name) VALUE (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
			alarmObj.StrategyId, alarmObj.Endpoint, alarmObj.Status, alarmObj.SMetric, alarmObj.SExpr, alarmObj.SCond, alarmObj.SLast, alarmObj.SPriority, alarmObj.Content, alarmObj.StartValue, alarmObj.Start.Format(models.DatetimeFormat), alarmObj.Tags, alarmObj.AlarmStrategy, alarmObj.EndpointTags, alarmObj.AlarmName)
		if execErr != nil {
			err = execErr
			return
		}
		lastInsertId, _ := execResult.LastInsertId()
		if lastInsertId <= 0 {
			err = fmt.Errorf("insert alarm table get 0 alarm id,tags:%s ", alarmObj.Tags)
			return
		}
		if alarmObj.SMetric == "db_keyword_monitor" {
			session.Exec("delete from db_keyword_alarm where tags=?", alarmObj.Tags)
			_, err = session.Exec("insert into db_keyword_alarm(alarm_id,endpoint,status,db_keyword_monitor,content,tags,start_value,updated_time) values (?,?,?,?,?,?,?,?)",
				lastInsertId, alarmObj.Endpoint, alarmObj.Status, alarmObj.AlarmStrategy, alarmObj.Content, alarmObj.Tags, alarmObj.StartValue, alarmObj.Start.Format(models.DatetimeFormat))
		} else {
			session.Exec("delete from log_keyword_alarm where tags=?", alarmObj.Tags)
			_, err = session.Exec("insert into log_keyword_alarm(alarm_id,endpoint,status,content,tags,start_value,updated_time,log_keyword_config) values (?,?,?,?,?,?,?,?)",
				lastInsertId, alarmObj.Endpoint, alarmObj.Status, alarmObj.Content, alarmObj.Tags, alarmObj.StartValue, alarmObj.Start.Format(models.DatetimeFormat), alarmObj.AlarmStrategy)
		}
	}
	return
}

func UpdateLogKeywordNotify(param *models.LogKeywordNotifyParam) (err error) {
	if param.Notify == nil {
		return
	}
	var actions []*Action
	actions = getNotifyListUpdateAction([]*models.NotifyObj{param.Notify})
	actions = append(actions, &Action{Sql: "delete from log_keyword_notify_rel where log_keyword_monitor=?", Param: []interface{}{param.LogKeywordMonitor}})
	actions = append(actions, &Action{Sql: "insert into log_keyword_notify_rel(guid,log_keyword_monitor,notify) values (?,?,?)", Param: []interface{}{
		"lk_notify_" + guid.CreateGuid(), param.LogKeywordMonitor, param.Notify.Guid,
	}})
	err = Transaction(actions)
	return
}

func GetLogKeywordNotify(logKeywordMonitorGuid string) (result *models.NotifyObj, err error) {
	var notifyRows []*models.NotifyTable
	err = x.SQL("select * from notify where guid in (select notify from log_keyword_notify_rel where log_keyword_monitor=?)", logKeywordMonitorGuid).Find(&notifyRows)
	if err != nil {
		return
	}
	if len(notifyRows) > 0 {
		result = buildNotifyObj(notifyRows[0])
	} else {
		result = &models.NotifyObj{}
	}
	return
}

func getLogKeywordAlarmNotify(logKeywordConfigGuid string) (notifyRow *models.NotifyTable, err error) {
	var logKeywordNotifyRelRows []*models.LogKeywordNotifyRel
	err = x.SQL("select * from log_keyword_notify_rel where log_keyword_config=? or log_keyword_monitor in (select log_keyword_monitor from log_keyword_config where guid=?)", logKeywordConfigGuid, logKeywordConfigGuid).Find(&logKeywordNotifyRelRows)
	if err != nil {
		err = fmt.Errorf("query log keyword notify rel table fail,%s ", err.Error())
		return
	}
	notifyRow = &models.NotifyTable{}
	if len(logKeywordNotifyRelRows) == 0 {
		return
	}
	notifyGuidList := []string{}
	for _, row := range logKeywordNotifyRelRows {
		notifyGuidList = append(notifyGuidList, row.Notify)
	}
	if len(notifyGuidList) == 0 {
		return
	}
	filterSql, filterParam := createListParams(notifyGuidList, "")
	var notifyRows []*models.NotifyTable
	err = x.SQL("select * from notify where guid in ("+filterSql+")", filterParam...).Find(&notifyRows)
	if err != nil {
		err = fmt.Errorf("query notify row fail,%s ", err.Error())
		return
	}
	for _, relRow := range logKeywordNotifyRelRows {
		matchNotify := &models.NotifyTable{}
		for _, row := range notifyRows {
			if relRow.Notify == row.Guid {
				matchNotify = row
				break
			}
		}
		if matchNotify.Guid != "" {
			tmpNotifyObj := buildNotifyObj(matchNotify)
			if len(tmpNotifyObj.NotifyRoles) == 0 && matchNotify.ProcCallbackKey == "" {
				continue
			}
		} else {
			continue
		}
		if relRow.LogKeywordConfig != "" {
			notifyRow = matchNotify
			break
		} else {
			notifyRow = matchNotify
		}
	}
	return
}
