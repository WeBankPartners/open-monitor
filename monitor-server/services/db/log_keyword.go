package db

import (
	"crypto/sha256"
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
		configObj.KeywordList = ListLogKeyword(v.Guid)
		configObj.EndpointRel = ListLogKeywordEndpointRel(v.Guid)
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
	param.Guid = guid.CreateGuid()
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	logKeywordGuidList := guid.CreateGuidList(len(param.LogPath))
	for i, path := range param.LogPath {
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

func ListLogKeyword(logKeywordMonitor string) (result []*models.LogKeywordConfigTable) {
	result = []*models.LogKeywordConfigTable{}
	x.SQL("select * from log_keyword_config where log_keyword_monitor=?", logKeywordMonitor).Find(&result)
	return
}

func CreateLogKeyword(param *models.LogKeywordConfigTable) (err error) {
	_, err = x.Exec("insert into log_keyword_config(guid,log_keyword_monitor,keyword,regulative,notify_enable,priority,update_time,content) value (?,?,?,?,?,?,?,?)", guid.CreateGuid(), param.LogKeywordMonitor, param.Keyword, param.Regulative, param.NotifyEnable, param.Priority, time.Now().Format(models.DatetimeFormat), param.Content)
	return
}

func UpdateLogKeyword(param *models.LogKeywordConfigTable) (err error) {
	_, err = x.Exec("update log_keyword_config set keyword=?,regulative=?,notify_enable=?,priority=?,update_time=?,content=? where guid=?", param.Keyword, param.Regulative, param.NotifyEnable, param.Priority, time.Now().Format(models.DatetimeFormat), param.Content, param.Guid)
	return
}

func DeleteLogKeyword(logKeywordConfigGuid string) (err error) {
	_, err = x.Exec("delete from log_keyword_config where guid=?", logKeywordConfigGuid)
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
	dataMap, err := datasource.QueryLogKeywordData()
	if err != nil {
		log.Logger.Error("Check log keyword break with get prometheus data", log.Error(err))
		return
	}
	if len(dataMap) == 0 {
		return
	}
	var logKeywordConfigs []*models.LogKeywordCronJobQuery
	x.SQL("select t1.guid,t1.service_group,t1.log_path,t1.monitor_type,t2.keyword,t2.notify_enable,t2.priority,t2.content,t3.source_endpoint,t3.target_endpoint,t4.agent_address from log_keyword_monitor t1 left join log_keyword_config t2 on t1.guid=t2.log_keyword_monitor left join log_keyword_endpoint_rel t3 on t1.guid=t3.log_keyword_monitor left join endpoint_new t4 on t3.source_endpoint=t4.guid where t3.source_endpoint is not null").Find(&logKeywordConfigs)
	if len(logKeywordConfigs) == 0 {
		log.Logger.Debug("Check log keyword break with empty config ")
		return
	}
	var firingAlarmTable, closeAlarmTable []*models.AlarmTable
	err = x.SQL("SELECT * FROM alarm WHERE s_metric='log_monitor' and status='firing'").Find(&firingAlarmTable)
	if err != nil {
		log.Logger.Error("Check log keyword break with query exist firing alarm fail", log.Error(err))
		return
	}
	err = x.SQL("select id,endpoint,tags,start_value,end_value,`start` from alarm where id in (select max(id) as id from alarm where s_metric='log_monitor' and status='closed' group by tags)").Find(&closeAlarmTable)
	if err != nil {
		log.Logger.Error("Check log keyword break with query exist closed alarm fail", log.Error(err))
		return
	}
	alarmMap := make(map[string]*models.AlarmTable)
	for _, v := range firingAlarmTable {
		if _, b := alarmMap[v.Tags]; b {
			continue
		}
		alarmMap[v.Tags] = v
	}
	for _, v := range closeAlarmTable {
		if firingExistAlarm, b := alarmMap[v.Tags]; b {
			if firingExistAlarm.Start.Unix() < v.Start.Unix() {
				alarmMap[v.Tags] = v
			}
		} else {
			alarmMap[v.Tags] = v
		}
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
				addAlarmRows = append(addAlarmRows, &models.AlarmTable{Id: existAlarm.Id, Status: existAlarm.Status, EndValue: newValue, Content: existAlarm.Content, End: nowTime})
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
			addAlarmRows = append(addAlarmRows, &models.AlarmTable{StrategyId: 0, Endpoint: config.TargetEndpoint, Status: "firing", SMetric: "log_monitor", SExpr: "node_log_monitor_count_total", SCond: ">0", SLast: "10s", SPriority: config.Priority, Content: alarmContent + getLogKeywordLastRow(config.AgentAddress, config.LogPath, config.Keyword), Tags: key, StartValue: newValue, Start: nowTime})
		}
	}
	if len(addAlarmRows) == 0 {
		return
	}
	var actions []*Action
	for _, v := range addAlarmRows {
		tmpAction := Action{}
		if v.Id > 0 {
			tmpAction.Sql = "UPDATE alarm SET content=?,end_value=?,end=? WHERE id=?"
			tmpAction.Param = []interface{}{v.Content, v.EndValue, v.End.Format(models.DatetimeFormat), v.Id}
		} else {
			calcAlarmUniqueFlag(v)
			tmpAction.Sql = "INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start,tags,alarm_strategy,endpoint_tags) VALUE (?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
			tmpAction.Param = []interface{}{v.StrategyId, v.Endpoint, v.Status, v.SMetric, v.SExpr, v.SCond, v.SLast, v.SPriority, v.Content, v.StartValue, v.Start.Format(models.DatetimeFormat), v.Tags, "log_keyword_strategy", v.EndpointTags}
		}
		actions = append(actions, &tmpAction)
	}
	err = Transaction(actions)
	if err != nil {
		log.Logger.Error("Update log keyword alarm table fail", log.Error(err))
		return
	}
	for _, v := range addAlarmRows {
		if v.Id > 0 {
			continue
		}
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

func ImportLogKeyword(param *models.LogKeywordServiceGroupObj) (err error) {
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
		actions = append(actions, &Action{Sql: "insert into log_keyword_monitor(guid,service_group,log_path,monitor_type,update_time) value (?,?,?,?,?)", Param: []interface{}{inputKeywordConfig.Guid, inputKeywordConfig.ServiceGroup, inputKeywordConfig.LogPath, inputKeywordConfig.MonitorType, nowTime}})
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
