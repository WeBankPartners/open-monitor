package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"time"
)

func PluginUpdateServicePathAction(input *models.PluginUpdateServicePathRequestObj) (result *models.PluginUpdateServicePathOutputObj, err error) {
	log.Logger.Info("PluginUpdateServicePathAction", log.JsonObj("input", input))
	result = &models.PluginUpdateServicePathOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", Guid: input.Guid}
	monitorTypeQuery,_ := x.QueryString("select guid from monitor_type where guid=?", input.MonitorType)
	if len(monitorTypeQuery) == 0 {
		err = fmt.Errorf("MonitorType:%s illegal ", input.MonitorType)
		return
	}
	serviceGroupObj,getErr := getSimpleServiceGroup(input.SystemName)
	if getErr != nil {
		err = fmt.Errorf("System:%s can not find,%s ", input.SystemName, getErr.Error())
		return
	}
	pathList := models.TransPluginMultiStringParam(input.LogPathList)
	endpointTypeMap := getServiceGroupEndpointWithChild(input.SystemName)
	sourceTargetMap := make(map[string]string)
	var hostEndpoint,targetEndpoint []string
	if hostListValue, b := endpointTypeMap["host"]; b {
		hostEndpoint = hostListValue
	}
	if targetListValue, b := endpointTypeMap[input.MonitorType]; b {
		targetEndpoint = targetListValue
	}
	if len(hostEndpoint) > 0 && len(targetEndpoint) > 0 {
		var endpointTable []*models.EndpointNewTable
		x.SQL("select guid,ip from endpoint_new where guid in ('" + strings.Join(hostEndpoint, "','") + "')").Find(&endpointTable)
		for _,target := range targetEndpoint {
			matchHost := ""
			for _,host := range endpointTable {
				if strings.Contains(target, fmt.Sprintf("_%s_", host.Ip)) {
					matchHost = host.Guid
					break
				}
			}
			if matchHost != "" {
				sourceTargetMap[matchHost] = target
			}
		}
	}
	err = updateServiceLogMetricPath(pathList, serviceGroupObj.Guid, input.MonitorType, sourceTargetMap)
	if err != nil {
		err = fmt.Errorf("Update logMetric config fail,%s ", err.Error())
		return
	}
	err = updateServiceLogKeywordPath(pathList, serviceGroupObj.Guid, input.MonitorType, sourceTargetMap)
	if err != nil {
		err = fmt.Errorf("Update logKeyword config fail,%s ", err.Error())
	}
	return
}

func updateServiceLogMetricPath(pathList []string,serviceGroup,monitorType string,sourceTargetMap map[string]string) (err error) {
	var logMetricTable []*models.LogMetricMonitorTable
	err = x.SQL("select * from log_metric_monitor where service_group=?", serviceGroup).Find(&logMetricTable)
	if err != nil {
		return
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	var affectHostList []string
	existPathTypeMap := make(map[string]string)
	existPathGuidMap := make(map[string]string)
	for _,v := range logMetricTable {
		existPathTypeMap[v.LogPath] = v.MonitorType
		existPathGuidMap[v.LogPath] = v.Guid
		deleteFlag := true
		for _,path := range pathList {
			if v.LogPath == path {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			// delete path
			tmpAffectList,tmpActions := getLogMetricPathDeleteAction(v.Guid)
			affectHostList = append(affectHostList, tmpAffectList...)
			actions = append(actions, tmpActions...)
		}
	}
	for _,path := range pathList {
		if existMonitorType,b:=existPathTypeMap[path];b {
			if existMonitorType != monitorType {
				// change monitor type
				tmpAffectList,tmpActions := getLogMetricPathChangeTypeAction(existPathGuidMap[path],monitorType,nowTime,sourceTargetMap)
				affectHostList = append(affectHostList, tmpAffectList...)
				actions = append(actions, tmpActions...)
			}
			continue
		}
		// add path
		tmpAffectList,tmpActions := getLogMetricPathAddAction(path,serviceGroup,monitorType,nowTime,sourceTargetMap)
		affectHostList = append(affectHostList, tmpAffectList...)
		actions = append(actions, tmpActions...)
	}
	if len(actions) > 0 {
		err = Transaction(actions)
		if err != nil {
			return err
		}
	}
	if len(affectHostList) > 0 {
		err = SyncLogMetricExporterConfig(affectHostList)
	}
	return err
}

func getLogMetricPathDeleteAction(logMetricMonitor string) (hostList []string,actions []*Action) {
	logMetricEndpointRel := ListLogMetricEndpointRel(logMetricMonitor)
	for _,v := range logMetricEndpointRel {
		hostList = append(hostList, v.SourceEndpoint)
	}
	actions = getDeleteLogMetricMonitorAction(logMetricMonitor)
	return
}

func getLogMetricPathAddAction(path,serviceGroup,monitorType,nowTime string,sourceTargetMap map[string]string) (hostList []string,actions []*Action) {
	newLogMetricGuid := guid.CreateGuid()
	path = strings.TrimSpace(path)
	actions = append(actions, &Action{Sql: "insert into log_metric_monitor(guid,service_group,log_path,monitor_type,update_time) value (?,?,?,?,?)", Param: []interface{}{newLogMetricGuid, serviceGroup, path, monitorType, nowTime}})
	for k,v := range sourceTargetMap {
		hostList = append(hostList, k)
		actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), newLogMetricGuid, k, v}})
	}
	return
}

func getLogMetricPathChangeTypeAction(logMetricMonitor,monitorType,nowTime string,sourceTargetMap map[string]string) (hostList []string,actions []*Action) {
	logMetricEndpointRel := ListLogMetricEndpointRel(logMetricMonitor)
	for _,v := range logMetricEndpointRel {
		hostList = append(hostList, v.SourceEndpoint)
	}
	actions = append(actions, &Action{Sql: "update log_metric_monitor set monitor_type=?,update_time=? where guid=?",Param: []interface{}{monitorType,nowTime,logMetricMonitor}})
	actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=?",Param: []interface{}{logMetricMonitor}})
	for k,v := range sourceTargetMap {
		hostList = append(hostList, k)
		actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), logMetricMonitor, k, v}})
	}
	return
}

func updateServiceLogKeywordPath(pathList []string,serviceGroup,monitorType string,sourceTargetMap map[string]string) (err error)  {
	var logKeywordTable []*models.LogKeywordMonitorTable
	err = x.SQL("select * from log_keyword_monitor where service_group=?", serviceGroup).Find(&logKeywordTable)
	if err != nil {
		return
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	var affectHostList []string
	existPathTypeMap := make(map[string]string)
	existPathGuidMap := make(map[string]string)
	for _,v := range logKeywordTable {
		existPathTypeMap[v.LogPath] = v.MonitorType
		existPathGuidMap[v.LogPath] = v.Guid
		deleteFlag := true
		for _,path := range pathList {
			if v.LogPath == path {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			// delete path
			tmpAffectList,tmpActions := getLogKeywordPathDeleteAction(v.Guid)
			affectHostList = append(affectHostList, tmpAffectList...)
			actions = append(actions, tmpActions...)
		}
	}
	for _,path := range pathList {
		if existMonitorType,b:=existPathTypeMap[path];b {
			if existMonitorType != monitorType {
				// change monitor type
				tmpAffectList,tmpActions := getLogKeywordPathChangeTypeAction(existPathGuidMap[path],monitorType,nowTime,sourceTargetMap)
				affectHostList = append(affectHostList, tmpAffectList...)
				actions = append(actions, tmpActions...)
			}
			continue
		}
		// add path
		tmpAffectList,tmpActions := getLogKeywordPathAddAction(path,serviceGroup,monitorType,nowTime,sourceTargetMap)
		affectHostList = append(affectHostList, tmpAffectList...)
		actions = append(actions, tmpActions...)
	}
	if len(actions) > 0 {
		err = Transaction(actions)
		if err != nil {
			return err
		}
	}
	if len(affectHostList) > 0 {
		err = SyncLogKeywordExporterConfig(affectHostList)
	}
	return err
}

func getLogKeywordPathDeleteAction(logKeywordMonitor string) (hostList []string,actions []*Action) {
	logKeywordEndpointRel := ListLogKeywordEndpointRel(logKeywordMonitor)
	for _,v := range logKeywordEndpointRel {
		hostList = append(hostList, v.SourceEndpoint)
	}
	actions = getDeleteLogKeywordMonitorAction(logKeywordMonitor)
	return
}

func getLogKeywordPathAddAction(path,serviceGroup,monitorType,nowTime string,sourceTargetMap map[string]string) (hostList []string,actions []*Action) {
	newLogKeywordGuid := guid.CreateGuid()
	path = strings.TrimSpace(path)
	actions = append(actions, &Action{Sql: "insert into log_keyword_monitor(guid,service_group,log_path,monitor_type,update_time) value (?,?,?,?,?)", Param: []interface{}{newLogKeywordGuid, serviceGroup, path, monitorType, nowTime}})
	for k,v := range sourceTargetMap {
		hostList = append(hostList, k)
		actions = append(actions, &Action{Sql: "insert into log_keyword_endpoint_rel(guid,log_keyword_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), newLogKeywordGuid, k, v}})
	}
	return
}

func getLogKeywordPathChangeTypeAction(logKeywordMonitor,monitorType,nowTime string,sourceTargetMap map[string]string) (hostList []string,actions []*Action) {
	logKeywordEndpointRel := ListLogKeywordEndpointRel(logKeywordMonitor)
	for _,v := range logKeywordEndpointRel {
		hostList = append(hostList, v.SourceEndpoint)
	}
	actions = append(actions, &Action{Sql: "update log_keyword_monitor set monitor_type=?,update_time=? where guid=?",Param: []interface{}{monitorType,nowTime,logKeywordMonitor}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=?",Param: []interface{}{logKeywordMonitor}})
	for k,v := range sourceTargetMap {
		hostList = append(hostList, k)
		actions = append(actions, &Action{Sql: "insert into log_keyword_endpoint_rel(guid,log_keyword_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), logKeywordMonitor, k, v}})
	}
	return
}