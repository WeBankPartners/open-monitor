package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func GetLogKeywordByServiceGroup(serviceGroupGuid string) (result []*models.LogKeywordServiceGroupObj,err error) {
	serviceGroupObj,getErr := getSimpleServiceGroup(serviceGroupGuid)
	if getErr != nil {
		return result,getErr
	}
	result = []*models.LogKeywordServiceGroupObj{}
	var logKeywordTable []*models.LogKeywordMonitorTable
	err = x.SQL("select * from log_keyword_monitor where service_group=?", serviceGroupGuid).Find(&logKeywordTable)
	if err != nil {
		return result,fmt.Errorf("Query table fail,%s ", err.Error())
	}
	configList := []*models.LogKeywordMonitorObj{}
	for _,v := range logKeywordTable {
		configObj := models.LogKeywordMonitorObj{Guid: v.Guid,ServiceGroup: serviceGroupGuid,LogPath: v.LogPath,MonitorType: v.MonitorType}
		configObj.KeywordList = ListLogKeyword(v.Guid)
		configObj.EndpointRel = ListLogKeywordEndpointRel(v.Guid)
		configList = append(configList, &configObj)
	}
	result = append(result, &models.LogKeywordServiceGroupObj{ServiceGroupTable:serviceGroupObj,Config: configList})
	return
}

func GetLogKeywordByEndpoint(endpointGuid string) (result []*models.LogKeywordServiceGroupObj,err error) {
	result = []*models.LogKeywordServiceGroupObj{}
	var logKeywordMonitorTable []*models.LogKeywordMonitorTable
	err = x.SQL("select distinct t2.service_group from log_keyword_endpoint_rel t1 left join log_keyword_monitor t2 on t1.log_keyword_monitor=t2.guid where t1.source_endpoint=? or t1.target_endpoint=?", endpointGuid, endpointGuid).Find(&logKeywordMonitorTable)
	if err != nil {
		return result,fmt.Errorf("Query table fail,%s ", err.Error())
	}
	for _,v := range logKeywordMonitorTable {
		tmpResult,tmpErr := GetLogKeywordByServiceGroup(v.ServiceGroup)
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
	for i,path := range param.LogPath {
		actions = append(actions, &Action{Sql: "insert into log_keyword_monitor(guid,service_group,log_path,monitor_type,update_time) value (?,?,?,?,?)",Param: []interface{}{logKeywordGuidList[i],param.ServiceGroup,path,param.MonitorType,nowTime}})
		endpointRelActions,tmpErr := getLogKeywordEndpointRelCreateAction(param.EndpointRel, logKeywordGuidList[i])
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

func getLogKeywordEndpointRelCreateAction(param []*models.LogKeywordEndpointRelTable,logKeywordMonitor string) (actions []*Action,err error) {
	endpointRelGuidList := guid.CreateGuidList(len(param))
	for ii,endpointRel := range param {
		if endpointRel.SourceEndpoint == "" || endpointRel.TargetEndpoint == "" {
			err = fmt.Errorf("endpointRel source and target can not empty ")
			break
		}
		actions = append(actions, &Action{Sql: "insert into log_keyword_endpoint_rel(guid,log_keyword_monitor,source_endpoint,target_endpoint) value (?,?,?,?)",Param: []interface{}{endpointRelGuidList[ii],logKeywordMonitor,endpointRel.SourceEndpoint,endpointRel.TargetEndpoint}})
	}
	return
}

func UpdateLogKeywordMonitor(param *models.LogKeywordMonitorObj) (err error) {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "update log_keyword_monitor set log_path=?,monitor_type=?,update_time=? where guid=?",Param: []interface{}{param.LogPath,param.MonitorType,nowTime,param.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=?",Param: []interface{}{param.Guid}})
	endpointRelActions,tmpErr := getLogKeywordEndpointRelCreateAction(param.EndpointRel, param.Guid)
	if tmpErr != nil {
		err = tmpErr
		return
	}
	actions = append(actions, endpointRelActions...)
	return Transaction(actions)
}

func DeleteLogKeywordMonitor(logKeywordMonitorGuid string) (err error) {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=?",Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_config where log_keyword_monitor=?",Param: []interface{}{logKeywordMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_monitor where guid=?",Param: []interface{}{logKeywordMonitorGuid}})
	return Transaction(actions)
}

func ListLogKeyword(logKeywordMonitor string) (result []*models.LogKeywordConfigTable) {
	result = []*models.LogKeywordConfigTable{}
	x.SQL("select * from log_keyword_config where log_keyword_monitor=?", logKeywordMonitor).Find(&result)
	return
}

func CreateLogKeyword(param *models.LogKeywordConfigTable) (err error) {
	_,err = x.Exec("insert into log_keyword_config(guid,log_keyword_monitor,keyword,regulative,notify_enable,priority,update_time) value (?,?,?,?,?,?,?)",guid.CreateGuid(),param.LogKeywordMonitor,param.Keyword,param.Regulative,param.NotifyEnable,param.Priority,time.Now().Format(models.DatetimeFormat))
	return
}

func UpdateLogKeyword(param *models.LogKeywordConfigTable) (err error) {
	_,err = x.Exec("update log_keyword_config set keyword=?,regulative=?,notify_enable=?,priority=?,update_time=? where guid=?",param.Keyword,param.Regulative,param.NotifyEnable,param.Priority,time.Now().Format(models.DatetimeFormat),param.Guid)
	return
}

func DeleteLogKeyword(logKeywordConfigGuid string) (err error) {
	_,err = x.Exec("delete from log_keyword_config where guid=?", logKeywordConfigGuid)
	return
}

func ListLogKeywordEndpointRel(logKeywordMonitor string) (result []*models.LogKeywordEndpointRelTable) {
	result = []*models.LogKeywordEndpointRelTable{}
	x.SQL("select * from log_keyword_endpoint_rel where log_keyword_monitor=?", logKeywordMonitor).Find(&result)
	return
}