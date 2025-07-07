package db

import (
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"go.uber.org/zap"

	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetLogKeywordByServiceGroup(serviceGroupGuid, alarmName string) (result []*models.LogKeywordServiceGroupObj, err error) {
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
	var configList []*models.LogKeywordMonitorObj
	for _, v := range logKeywordTable {
		configObj := models.LogKeywordMonitorObj{Guid: v.Guid, ServiceGroup: serviceGroupGuid, LogPath: v.LogPath, MonitorType: v.MonitorType}
		if configObj.KeywordList, err = ListLogKeyword(v.Guid, alarmName); err != nil {
			return
		}
		configObj.EndpointRel = ListLogKeywordEndpointRel(v.Guid)
		if configObj.Notify, err = GetLogKeywordNotify(v.Guid); err != nil {
			return
		}
		configList = append(configList, &configObj)
	}

	// 导出Db keyword
	var dbKeywordTable []*models.DbKeywordMonitor
	var dbConfigList []*models.DbKeywordConfigObj
	if err = x.SQL("select * from db_keyword_monitor where service_group=? order by  update_time desc", serviceGroupGuid).Find(&dbKeywordTable); err != nil {
		return
	}
	for _, v := range dbKeywordTable {
		configObj := models.DbKeywordConfigObj{DbKeywordMonitor: *v}
		if configObj.EndpointRel, err = ListDbKeywordEndpointRel(v.Guid); err != nil {
			return
		}
		if configObj.Notify, _, err = GetDbKeywordNotify(v.Guid); err != nil {
			return
		}
		configObj.ActiveWindowList = strings.Split(configObj.ActiveWindow, ",")
		dbConfigList = append(dbConfigList, &configObj)
	}

	result = append(result, &models.LogKeywordServiceGroupObj{ServiceGroupTable: serviceGroupObj, Config: configList, DbConfig: dbConfigList})
	return
}

func GetLogKeywordByEndpoint(endpointGuid, alarmName string, onlySource bool) (result []*models.LogKeywordServiceGroupObj, err error) {
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
		tmpResult, tmpErr := GetLogKeywordByServiceGroup(v.ServiceGroup, alarmName)
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
	actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify in (select notify from log_keyword_notify_rel where log_keyword_monitor=?)", Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from notify where guid in (select notify from log_keyword_notify_rel where log_keyword_monitor=?)", Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_notify_rel where log_keyword_monitor=?", Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_monitor where guid=?", Param: []interface{}{logKeywordMonitorGuid}})
	return actions
}

func ListLogKeyword(logKeywordMonitor, alarmName string) (result []*models.LogKeywordConfigTable, err error) {
	result = []*models.LogKeywordConfigTable{}
	if strings.TrimSpace(alarmName) == "" {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? order by update_time desc", logKeywordMonitor).Find(&result)
	} else {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and name like '%"+alarmName+"%' order by update_time desc", logKeywordMonitor).Find(&result)
	}
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

func GetLogKeywordConfigUniqueData(guid, name, keyword, logKeywordMonitorGuid string) (sameNameList, sameKeywordList []*models.LogKeywordConfigTable, err error) {
	sameNameList = []*models.LogKeywordConfigTable{}
	if guid == "" {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and name=?", logKeywordMonitorGuid, name).Find(&sameNameList)
	} else {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and name=? and guid<>?", logKeywordMonitorGuid, name, guid).Find(&sameNameList)
	}
	if err != nil {
		return
	}
	// 由于db设置不区分大小写,结果集过滤大小写
	var filteredSameNameList []*models.LogKeywordConfigTable
	for _, config := range sameNameList {
		if config.Name == name {
			filteredSameNameList = append(filteredSameNameList, config)
		}
	}
	sameNameList = filteredSameNameList
	sameKeywordList = []*models.LogKeywordConfigTable{}
	if guid == "" {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and keyword=?", logKeywordMonitorGuid, keyword).Find(&sameKeywordList)
	} else {
		err = x.SQL("select * from log_keyword_config where log_keyword_monitor=? and keyword=? and guid<>?", logKeywordMonitorGuid, keyword, guid).Find(&sameKeywordList)
	}
	// 过滤大小写
	var filteredSameKeywordList []*models.LogKeywordConfigTable
	for _, config := range sameKeywordList {
		if config.Keyword == keyword {
			filteredSameKeywordList = append(filteredSameKeywordList, config)
		}
	}
	sameKeywordList = filteredSameKeywordList
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
		closeAlarmActions, tmpErr := getLogKeywordCloseAlarmActions(param.Guid)
		if tmpErr != nil {
			err = fmt.Errorf("try to get close alarm actions fail,%s ", tmpErr.Error())
			return
		}
		if len(closeAlarmActions) > 0 {
			actions = append(actions, closeAlarmActions...)
		}
	}
	err = Transaction(actions)
	return
}

func getLogKeywordCloseAlarmActions(logKeywordConfigGuid string) (actions []*Action, err error) {
	var logKeywordAlarmRows []*models.LogKeywordAlarmTable
	err = x.SQL("select id,alarm_id from log_keyword_alarm where log_keyword_config=? and status='firing'", logKeywordConfigGuid).Find(&logKeywordAlarmRows)
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
	return
}

func DeleteLogKeyword(logKeywordConfigGuid string) (err error) {
	var actions []*Action
	closeAlarmActions, tmpErr := getLogKeywordCloseAlarmActions(logKeywordConfigGuid)
	if tmpErr != nil {
		err = fmt.Errorf("try to get close alarm actions fail,%s ", tmpErr.Error())
		return
	}
	actions = append(actions, closeAlarmActions...)
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
	dataMap, previousResult, err := datasource.QueryLogKeywordData("log")
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Check log keyword break with get prometheus data", zap.Error(err))
		return
	}
	if len(dataMap) == 0 {
		log.Debug(nil, log.LOGGER_APP, "doLogKeywordMonitorJob break with dataMap empty")
		return
	}

	// 1. 分批查 log_keyword_monitor
	var (
		logKeywordConfigs []*models.LogKeywordCronJobQuery
		batchSize         = 500
		lastGuid          = ""
	)
	for {
		var monitors []*models.LogKeywordMonitorTable
		monitorSql := "select guid,service_group,log_path,monitor_type from log_keyword_monitor"
		var monitorArgs []interface{}
		if lastGuid != "" {
			monitorSql += " where guid > ?"
			monitorArgs = append(monitorArgs, lastGuid)
		}
		monitorSql += " order by guid limit ?"
		monitorArgs = append(monitorArgs, batchSize)
		err = x.SQL(monitorSql, monitorArgs...).Find(&monitors)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Query log_keyword_monitor fail", zap.Error(err))
			return
		}
		if len(monitors) == 0 {
			break
		}
		monitorGuids := make([]string, 0, len(monitors))
		for _, m := range monitors {
			monitorGuids = append(monitorGuids, m.Guid)
		}
		// 2. 查 log_keyword_config
		var configs []*models.LogKeywordConfigTable
		err = x.In("log_keyword_monitor", monitorGuids).Find(&configs)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Query log_keyword_config fail", zap.Error(err))
			return
		}
		// 3. 查 log_keyword_endpoint_rel
		var endpointRels []*models.LogKeywordEndpointRelTable
		err = x.In("log_keyword_monitor", monitorGuids).Find(&endpointRels)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Query log_keyword_endpoint_rel fail", zap.Error(err))
			return
		}
		sourceEndpointGuids := make([]string, 0)
		for _, rel := range endpointRels {
			if rel.SourceEndpoint != "" {
				sourceEndpointGuids = append(sourceEndpointGuids, rel.SourceEndpoint)
			}
		}
		// 4. 查 endpoint_new
		var endpoints []*models.EndpointNewTable
		if len(sourceEndpointGuids) > 0 {
			err = x.In("guid", sourceEndpointGuids).Find(&endpoints)
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "Query endpoint_new fail", zap.Error(err))
				return
			}
		}
		// 5. 组装数据
		configMap := make(map[string][]*models.LogKeywordConfigTable)
		for _, c := range configs {
			configMap[c.LogKeywordMonitor] = append(configMap[c.LogKeywordMonitor], c)
		}
		relMap := make(map[string][]*models.LogKeywordEndpointRelTable)
		for _, r := range endpointRels {
			relMap[r.LogKeywordMonitor] = append(relMap[r.LogKeywordMonitor], r)
		}
		endpointMap := make(map[string]*models.EndpointNewTable)
		for _, e := range endpoints {
			endpointMap[e.Guid] = e
		}
		for _, m := range monitors {
			for _, c := range configMap[m.Guid] {
				for _, r := range relMap[m.Guid] {
					if r.SourceEndpoint == "" {
						continue
					}
					endpoint := endpointMap[r.SourceEndpoint]
					var agentAddress string
					if endpoint != nil {
						agentAddress = endpoint.AgentAddress
					}
					logKeywordConfigs = append(logKeywordConfigs, &models.LogKeywordCronJobQuery{
						Guid:                 m.Guid,
						ServiceGroup:         m.ServiceGroup,
						LogPath:              m.LogPath,
						MonitorType:          m.MonitorType,
						Keyword:              c.Keyword,
						NotifyEnable:         c.NotifyEnable,
						Priority:             c.Priority,
						Content:              c.Content,
						Name:                 c.Name,
						LogKeywordConfigGuid: c.Guid,
						ActiveWindow:         c.ActiveWindow,
						SourceEndpoint:       r.SourceEndpoint,
						TargetEndpoint:       r.TargetEndpoint,
						AgentAddress:         agentAddress,
					})
				}
			}
		}
		// 分批处理，记录最后一条 guid
		lastGuid = monitors[len(monitors)-1].Guid
		if len(monitors) < batchSize {
			break
		}
	}
	if len(logKeywordConfigs) == 0 {
		log.Debug(nil, log.LOGGER_APP, "Check log keyword break with empty config ")
		return
	}
	var alarmTable []*models.LogKeywordAlarmTable
	err = x.SQL("select * from log_keyword_alarm").Find(&alarmTable)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Check log keyword break with query exist closed alarm fail", zap.Error(err))
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
			notifyConfigMap[config.LogKeywordConfigGuid] = 1
		}
		key := fmt.Sprintf("e_guid:%s^t_guid:%s^file:%s^keyword:%s", config.SourceEndpoint, config.TargetEndpoint, config.LogPath, config.Keyword)
		newValue, oldValue = 0, 0
		if dataValue, b := dataMap[key]; b {
			newValue = dataValue
		} else {
			log.Debug(nil, log.LOGGER_APP, "doLogKeywordMonitorJob ignore logKeywordConfig", zap.String("key", key))
			continue
		}
		if newValue == 0 {
			log.Debug(nil, log.LOGGER_APP, "doLogKeywordMonitorJob ignore logKeywordConfig with empty value", zap.String("key", key))
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
				if v, existKey := previousResult[key]; existKey && newValue > v {
					// 说明数据被重置过了 也需要告警
					log.Info(nil, log.LOGGER_APP, "doLogKeywordMonitorJob Counter reset", zap.String("key", key))
				} else {
					continue
				}
			}
			if existAlarm.Status == "firing" || !InActiveWindowList(config.ActiveWindow) {
				existAlarm.Content = strings.Split(existAlarm.Content, "^^")[0] + "^^" + getLogKeywordLastRow(config.AgentAddress, config.LogPath, config.Keyword)
				addAlarmRows = append(addAlarmRows, &models.AlarmTable{Id: existAlarm.AlarmId, Status: existAlarm.Status, EndValue: newValue, Content: existAlarm.Content, End: nowTime})
			} else {
				addFlag = true
			}
		} else {
			if InActiveWindowList(config.ActiveWindow) {
				addFlag = true
			}
		}
		if addFlag {
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
			log.Error(nil, log.LOGGER_APP, "Update log keyword alarm table fail", zap.String("tags", v.Tags), zap.Error(tmpErr))
		} else {
			if v.Id <= 0 {
				if _, b := notifyConfigMap[v.AlarmStrategy]; !b {
					log.Warn(nil, log.LOGGER_APP, "Log keyword monitor notify disable,ignore", zap.String("logKeywordConfig", v.AlarmStrategy))
					continue
				}
				tmpAlarmObj := getSimpleAlarmByLogKeywordTags(v.Tags)
				if tmpAlarmObj.Id <= 0 {
					log.Warn(nil, log.LOGGER_APP, "Log keyword monitor notify fail,query alarm with tags fail", zap.String("tags", v.Tags))
					continue
				}
				tmpNotifyRow, getNotifyErr := getLogKeywordAlarmNotify(v.AlarmStrategy)
				if getNotifyErr != nil {
					log.Error(nil, log.LOGGER_APP, "doLogKeywordMonitorJob get alarm notify fail", zap.String("logKeywordConfigGuid", v.AlarmStrategy), zap.Int("alarmId", tmpAlarmObj.Id), zap.Error(getNotifyErr))
					continue
				}
				if tmpNotifyRow.ProcCallbackMode == models.AlarmNotifyManualMode && tmpNotifyRow.ProcCallbackKey != "" {
					if _, execErr := x.Exec("update alarm set notify_id=? where id=?", tmpNotifyRow.Guid, tmpAlarmObj.Id); execErr != nil {
						log.Error(nil, log.LOGGER_APP, "update alarm table notify id fail", zap.Int("alarmId", tmpAlarmObj.Id), zap.Error(execErr))
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
		log.Error(nil, log.LOGGER_APP, "Get log keyword rows fail,new request error", zap.Error(err))
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		log.Error(nil, log.LOGGER_APP, "Get log keyword rows fail,response error", zap.Error(respErr))
		return result
	}
	var responseData models.LogKeywordRowsHttpResult
	respBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(respBytes, &responseData)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get log keyword rows fail,response data json unmarshal error", zap.Error(err))
		return result
	}
	if responseData.Status != "ok" {
		log.Error(nil, log.LOGGER_APP, "Get log keyword rows fail,response status error", zap.String("status", responseData.Status), zap.String("message", responseData.Message))
		return result
	}
	for _, v := range responseData.Data {
		result = v.Content
	}
	return result
}

func ImportLogAndDbKeyword(param *models.LogKeywordServiceGroupObj, operator string) (err error) {
	existSGs, getExistDataErr := GetLogKeywordByServiceGroup(param.Guid, "")
	if getExistDataErr != nil {
		return fmt.Errorf("get exist log keyword data fail,%s ", getExistDataErr.Error())
	}
	if len(existSGs) == 0 {
		return fmt.Errorf("get empty log keyword data,please check service group")
	}
	var actions, subDeleteDbKeywordConfigActions []*Action
	var affectHostList []string
	existSGConfig := existSGs[0]
	for _, existKeywordConfig := range existSGConfig.Config {
		actions = append(actions, getDeleteLogKeywordMonitorAction(existKeywordConfig.Guid)...)
		for _, v := range existKeywordConfig.EndpointRel {
			affectHostList = append(affectHostList, v.SourceEndpoint)
		}
	}
	for _, existDbKeywordConfig := range existSGConfig.DbConfig {
		if subDeleteDbKeywordConfigActions, err = getDeleteDbKeywordConfigActions(existDbKeywordConfig.Guid); err != nil {
			return
		}
		actions = append(actions, subDeleteDbKeywordConfigActions...)
	}

	nowTime := time.Now().Format(models.DatetimeFormat)
	for _, inputKeywordConfig := range param.Config {
		actions = append(actions, &Action{Sql: "insert into log_keyword_monitor(guid,service_group,log_path,monitor_type,update_time,update_user) value (?,?,?,?,?,?)",
			Param: []interface{}{inputKeywordConfig.Guid, inputKeywordConfig.ServiceGroup, inputKeywordConfig.LogPath, inputKeywordConfig.MonitorType, nowTime, operator}})
		if inputKeywordConfig.Notify != nil {
			inputKeywordConfig.Notify.EndpointGroup = ""
			inputKeywordConfig.Notify.ServiceGroup = ""
			inputKeywordConfig.Notify.AlarmStrategy = ""
			notifyList := []*models.NotifyObj{inputKeywordConfig.Notify}
			actions = append(actions, getNotifyListInsertAction(notifyList)...)
			actions = append(actions, &Action{Sql: "insert into log_keyword_notify_rel(guid,log_keyword_monitor,notify) values (?,?,?)", Param: []interface{}{
				"lk_notify_" + guid.CreateGuid(), inputKeywordConfig.Guid, inputKeywordConfig.Notify.Guid,
			}})
		}
		for _, keywordObj := range inputKeywordConfig.KeywordList {
			actions = append(actions, &Action{Sql: "insert into log_keyword_config(guid,log_keyword_monitor,keyword,regulative,notify_enable,priority," +
				"update_time,name,content,active_window,create_time,update_user) value (?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{keywordObj.Guid,
				keywordObj.LogKeywordMonitor, keywordObj.Keyword, keywordObj.Regulative, keywordObj.NotifyEnable, keywordObj.Priority,
				nowTime, keywordObj.Name, keywordObj.Content, keywordObj.ActiveWindow, nowTime, operator}})
			if keywordObj.Notify != nil {
				keywordObj.Notify.EndpointGroup = ""
				keywordObj.Notify.ServiceGroup = ""
				keywordObj.Notify.AlarmStrategy = ""
				notifyList := []*models.NotifyObj{keywordObj.Notify}
				actions = append(actions, getNotifyListInsertAction(notifyList)...)
				actions = append(actions, &Action{Sql: "insert into log_keyword_notify_rel(guid,log_keyword_config,notify) values (?,?,?)", Param: []interface{}{
					"lk_notify_" + guid.CreateGuid(), keywordObj.Guid, keywordObj.Notify.Guid,
				}})
			}

		}
	}
	for _, inputDbKeywordConfig := range param.DbConfig {
		actions = append(actions, getCreateDbKeywordConfigActions(inputDbKeywordConfig, operator, time.Now())...)
	}
	err = Transaction(actions)
	if len(affectHostList) > 0 && err == nil {
		if syncErr := SyncLogKeywordExporterConfig(affectHostList); syncErr != nil {
			log.Error(nil, log.LOGGER_APP, "import log keyword fail with sync host keyword config", zap.Error(syncErr), zap.Strings("hosts", affectHostList))
		}
	}
	if len(param.DbConfig) > 0 {
		if syncDbErr := SyncDbMetric(false); syncDbErr != nil {
			log.Error(nil, log.LOGGER_APP, "import db keyword fail with sync config", zap.Error(syncDbErr))
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
			affectRowNum, affectRowErr := execResult.RowsAffected()
			if affectRowErr != nil {
				err = fmt.Errorf("update log keyword alarm table tail,get affect row result error:%s ", affectRowErr.Error())
				return
			} else if affectRowNum <= 0 {
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
