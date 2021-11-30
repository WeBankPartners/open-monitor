package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func GetLogMetricByServiceGroup(serviceGroup string) (result models.LogMetricQueryObj,err error) {
	serviceGroupObj,getErr := getSimpleServiceGroup(serviceGroup)
	if getErr != nil {
		return result,getErr
	}
	result.ServiceGroupTable = serviceGroupObj
	var logMetricMonitorTable []*models.LogMetricMonitorTable
	err = x.SQL("select * from log_metric_monitor where service_group=?", serviceGroup).Find(&logMetricMonitorTable)
	if err != nil {
		return
	}
	for _,logMetricMonitor := range logMetricMonitorTable {
		tmpConfig := models.LogMetricMonitorObj{Guid: logMetricMonitor.Guid,ServiceGroup: logMetricMonitor.ServiceGroup,LogPath: logMetricMonitor.LogPath,MetricType: logMetricMonitor.MetricType,MonitorType: logMetricMonitor.MonitorType}
		tmpConfig.EndpointRel = ListLogMetricEndpointRel(logMetricMonitor.Guid)
		tmpConfig.JsonConfigList = ListLogMetricJson(logMetricMonitor.Guid)
		tmpConfig.MetricConfigList = ListLogMetricConfig("", logMetricMonitor.Guid)
		result.Config = append(result.Config, &tmpConfig)
	}
	return 
}

func GetLogMetricByEndpoint(endpoint string) (result []*models.LogMetricQueryObj,err error) {
	result = []*models.LogMetricQueryObj{}
	var endpointServiceRelTable []*models.EndpointServiceRelTable
	err = x.SQL("select * from endpoint_service_rel where endpoint=?", endpoint).Find(&endpointServiceRelTable)
	for _,v := range endpointServiceRelTable {
		tmpObj,tmpErr := GetLogMetricByServiceGroup(v.ServiceGroup)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		result = append(result, &tmpObj)
	}
	return
}

func ListLogMetricEndpointRel(logMetricMonitor string) (result []*models.LogMetricEndpointRelTable) {
	result = []*models.LogMetricEndpointRelTable{}
	x.SQL("select * from log_metric_endpoint_rel where log_metric_monitor=?", logMetricMonitor).Find(&result)
	return result
}

func ListLogMetricJson(logMetricMonitor string) (result []*models.LogMetricJsonObj) {
	result = []*models.LogMetricJsonObj{}
	var logMetricJsonTable []*models.LogMetricJsonTable
	x.SQL("select * from log_metric_json where log_metric_monitor=?", logMetricMonitor).Find(&logMetricJsonTable)
	for _,v := range logMetricJsonTable {
		result = append(result, &models.LogMetricJsonObj{Guid: v.Guid,LogMetricMonitor: v.LogMetricMonitor,JsonRegular: v.JsonRegular,Tags: v.Tags,MetricList: ListLogMetricConfig(v.Guid, "")})
	}
	return result
}

func ListLogMetricConfig(logMetricJson,logMetricMonitor string) (result []*models.LogMetricConfigObj) {
	result = []*models.LogMetricConfigObj{}
	var logMetricConfigTable []*models.LogMetricConfigTable
	if logMetricJson != "" {
		x.SQL("select * from log_metric_config where log_metric_json=?", logMetricJson).Find(&logMetricConfigTable)
	}else{
		x.SQL("select * from log_metric_config where log_metric_monitor=?", logMetricMonitor).Find(&logMetricConfigTable)
	}
	for _,v := range logMetricConfigTable {
		result = append(result, &models.LogMetricConfigObj{Guid: v.Guid,LogMetricMonitor: v.LogMetricMonitor,LogMetricJson: v.LogMetricJson,Metric: v.Metric,DisplayName: v.DisplayName,JsonKey: v.JsonKey,Regular: v.Regular,AggType: v.AggType,StringMap: ListLogMetricStringMap(v.Guid)})
	}
	return result
}

func ListLogMetricStringMap(logMetricConfig string) (result []*models.LogMetricStringMapTable) {
	result = []*models.LogMetricStringMapTable{}
	x.SQL("select * from log_metric_string_map where log_metric_config=?", logMetricConfig).Find(&result)
	return result
}

func CreateLogMetricMonitor(param *models.LogMetricMonitorCreateDto) error {
	if len(param.LogPath) == 0 {
		return nil
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	logMonitorGuidList := guid.CreateGuidList(len(param.LogPath))
	for i,v := range param.LogPath {
		actions = append(actions, &Action{Sql: "insert into log_metric_monitor(guid,service_group,log_path,metric_type,monitor_type,update_time) value (?,?,?,?,?,?)", Param: []interface{}{logMonitorGuidList[i], param.ServiceGroup, v, param.MetricType, param.MonitorType, nowTime}})
		relGuidList := guid.CreateGuidList(len(param.EndpointRel))
		for ii, vv := range param.EndpointRel {
			actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{relGuidList[ii], logMonitorGuidList[i], vv.SourceEndpoint, vv.TargetEndpoint}})
		}
	}
	return Transaction(actions)
}

func UpdateLogMetricMonitor(param *models.LogMetricMonitorObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	actions = append(actions, &Action{Sql: "update log_metric_monitor set log_path=?,monitor_type=?,update_time=? where guid=?",Param: []interface{}{param.LogPath,param.MonitorType,nowTime,param.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=?",Param: []interface{}{param.Guid}})
	guidList := guid.CreateGuidList(len(param.EndpointRel))
	for i,v := range param.EndpointRel {
		actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)",Param: []interface{}{guidList[i],param.Guid,v.SourceEndpoint,v.TargetEndpoint}})
	}
	return Transaction(actions)
}

func DeleteLogMetricMonitor(logMetricMonitorGuid string) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=?",Param: []interface{}{logMetricMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config in (select guid from log_metric_config where log_metric_monitor=? or log_metric_json in (select guid from log_metric_json where log_metric_monitor=?))",Param: []interface{}{logMetricMonitorGuid,logMetricMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_config where log_metric_monitor=? or log_metric_json in (select guid from log_metric_json where log_metric_monitor=?)",Param: []interface{}{logMetricMonitorGuid,logMetricMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_json where log_metric_monitor=?",Param: []interface{}{logMetricMonitorGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_monitor where guid=?",Param: []interface{}{logMetricMonitorGuid}})
	return Transaction(actions)
}

func GetLogMetricJson(logMetricJsonGuid string) (result models.LogMetricJsonObj,err error) {
	var logMetricJsonTable []*models.LogMetricJsonTable
	err = x.SQL("select * from log_metric_json where guid=?", logMetricJsonGuid).Find(&logMetricJsonTable)
	if err != nil {
		return result,fmt.Errorf("Query log_metric_json table fail,%s ", err.Error())
	}
	if len(logMetricJsonTable) == 0 {
		return result,fmt.Errorf("Can not find log_metric_json with guid:%s ", logMetricJsonGuid)
	}
	result = models.LogMetricJsonObj{Guid: logMetricJsonTable[0].Guid,LogMetricMonitor: logMetricJsonTable[0].LogMetricMonitor,JsonRegular: logMetricJsonTable[0].JsonRegular,Tags: logMetricJsonTable[0].Tags}
	result.MetricList = ListLogMetricConfig(logMetricJsonGuid, "")
	return
}

func CreateLogMetricJson(param *models.LogMetricJsonObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	param.Guid = guid.CreateGuid()
	actions = append(actions,&Action{Sql: "insert into log_metric_json(guid,log_metric_monitor,json_regular,tags,update_time) value (?,?,?,?,?)",Param: []interface{}{param.Guid,param.LogMetricMonitor,param.JsonRegular,param.Tags,nowTime}})
	guidList := guid.CreateGuidList(len(param.MetricList))
	for i,v := range param.MetricList {
		v.LogMetricJson = param.Guid
		v.Guid = guidList[i]
		tmpActions := getCreateLogMetricConfigAction(v, nowTime)
		actions = append(actions, tmpActions...)
	}
	return Transaction(actions)
}

func UpdateLogMetricJson(param *models.LogMetricJsonObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	actions = append(actions, &Action{Sql: "update log_metric_json set json_regular=?,tags=?,update_time=? where guid=?",Param: []interface{}{param.JsonRegular,param.Tags,nowTime,param.Guid}})
	var logMetricConfigTable []*models.LogMetricConfigTable
	x.SQL("select * from log_metric_config where log_metric_json=?", param.Guid).Find(&logMetricConfigTable)
	for _,v := range param.MetricList {
		if v.Guid == "" {
			actions = append(actions, getCreateLogMetricConfigAction(v, nowTime)...)
			continue
		}
		actions = append(actions, getUpdateLogMetricConfigAction(v, nowTime)...)
	}
	for _,v := range logMetricConfigTable {
		existFlag := false
		for _,vv := range param.MetricList {
			if v.Guid == vv.Guid {
				existFlag = true
				break
			}
		}
		if !existFlag {
			actions = append(actions, getDeleteLogMetricConfigAction(v.Guid)...)
		}
	}
	return Transaction(actions)
}

func DeleteLogMetricJson(logMetricJsonGuid string) error {
	var actions []*Action
	var logMetricConfigTable []*models.LogMetricConfigTable
	x.SQL("select * from log_metric_config where log_metric_json=?", logMetricJsonGuid).Find(&logMetricConfigTable)
	for _,v := range logMetricConfigTable {
		actions = append(actions, getDeleteLogMetricConfigAction(v.Guid)...)
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_json where guid=?",Param: []interface{}{logMetricJsonGuid}})
	return Transaction(actions)
}

func GetLogMetricConfig(logMetricConfigGuid string) (result models.LogMetricConfigObj,err error) {
	var logMetricConfigTable []*models.LogMetricConfigTable
	err = x.SQL("select * from log_metric_config where guid=?",logMetricConfigGuid).Find(&logMetricConfigTable)
	if err != nil {
		return result,fmt.Errorf("Query table log_metric_config fail,%s ", err.Error())
	}
	if len(logMetricConfigTable) == 0 {
		return result,fmt.Errorf("Can not find log_metric_config with guid:%s ", logMetricConfigGuid)
	}
	result = models.LogMetricConfigObj{Guid: logMetricConfigGuid,LogMetricMonitor: logMetricConfigTable[0].LogMetricMonitor,LogMetricJson: logMetricConfigTable[0].LogMetricJson,Metric: logMetricConfigTable[0].Metric,DisplayName: logMetricConfigTable[0].DisplayName,JsonKey: logMetricConfigTable[0].JsonKey,Regular: logMetricConfigTable[0].Regular,AggType: logMetricConfigTable[0].AggType}
	result.StringMap = ListLogMetricStringMap(logMetricConfigGuid)
	return
}

func CreateLogMetricConfig(param *models.LogMetricConfigObj) error {
	actions := getCreateLogMetricConfigAction(param, time.Now().Format(models.DatetimeFormat))
	return Transaction(actions)
}

func UpdateLogMetricConfig(param *models.LogMetricConfigObj) error {
	actions := getUpdateLogMetricConfigAction(param, time.Now().Format(models.DatetimeFormat))
	return Transaction(actions)
}

func DeleteLogMetricConfig(logMetricConfigGuid string) error {
	actions := getDeleteLogMetricConfigAction(logMetricConfigGuid)
	return Transaction(actions)
}

func getCreateLogMetricConfigAction(param *models.LogMetricConfigObj, nowTime string) []*Action {
	var actions []*Action
	if param.Guid == "" {
		param.Guid = guid.CreateGuid()
	}
	actions = append(actions, &Action{Sql: "insert into log_metric_config(guid,log_metric_monitor,log_metric_json,metric,display_name,json_key,regular,agg_type,update_time) value (?,?,?,?,?,?,?,?,?)",Param: []interface{}{param.Guid,param.LogMetricMonitor,param.LogMetricJson,param.Metric,param.DisplayName,param.JsonKey,param.Regular,param.AggType,nowTime}})
	guidList := guid.CreateGuidList(len(param.StringMap))
	for i,v := range param.StringMap {
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_config,source_value,regulative,target_value,update_time) value (?,?,?,?,?,?)",Param: []interface{}{guidList[i],v.LogMetricConfig,v.SourceValue,v.Regulative,v.TargetValue,nowTime}})
	}
	return actions
}

func getUpdateLogMetricConfigAction(param *models.LogMetricConfigObj, nowTime string) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "update log_metric_config set metric=?,display_name=?,json_key=?,regular=?,agg_type=?,update_time=? where guid=?",Param: []interface{}{param.Metric,param.DisplayName,param.JsonKey,param.Regular,param.AggType,nowTime}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config=?",Param: []interface{}{param.Guid}})
	guidList := guid.CreateGuidList(len(param.StringMap))
	for i,v := range param.StringMap {
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_config,source_value,regulative,target_value,update_time) value (?,?,?,?,?,?)",Param: []interface{}{guidList[i],v.LogMetricConfig,v.SourceValue,v.Regulative,v.TargetValue,nowTime}})
	}
	return actions
}

func getDeleteLogMetricConfigAction(logMetricConfigGuid string) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config=?",Param: []interface{}{logMetricConfigGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_config where guid=?",Param: []interface{}{logMetricConfigGuid}})
	return actions
}