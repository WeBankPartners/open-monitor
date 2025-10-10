package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/go-common-lib/pcre"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/dlclark/regexp2"
	"go.uber.org/zap"
)

func GetLogMetricByServiceGroup(serviceGroup, metricKey string) (result models.LogMetricQueryObj, err error) {
	serviceGroupObj, getErr := getSimpleServiceGroup(serviceGroup)
	if getErr != nil {
		return result, getErr
	}
	result.ServiceGroupTable = serviceGroupObj
	result.Config = []*models.LogMetricMonitorObj{}
	var logMetricMonitorTable []*models.LogMetricMonitorTable
	err = x.SQL("select * from log_metric_monitor where service_group=?", serviceGroup).Find(&logMetricMonitorTable)
	if err != nil {
		return
	}
	for _, logMetricMonitor := range logMetricMonitorTable {
		tmpConfig := models.LogMetricMonitorObj{Guid: logMetricMonitor.Guid, ServiceGroup: logMetricMonitor.ServiceGroup, LogPath: logMetricMonitor.LogPath, MetricType: logMetricMonitor.MetricType, MonitorType: logMetricMonitor.MonitorType}
		tmpConfig.EndpointRel = ListLogMetricEndpointRel(logMetricMonitor.Guid)
		tmpConfig.JsonConfigList = ListLogMetricJson(logMetricMonitor.Guid)
		tmpConfig.MetricConfigList = ListLogMetricConfig("", logMetricMonitor.Guid)
		tmpConfig.MetricGroups = ListLogMetricGroups(logMetricMonitor.Guid, metricKey)
		for _, logJsonObj := range tmpConfig.JsonConfigList {
			for _, logMetricObj := range logJsonObj.MetricList {
				logMetricObj.ServiceGroup = serviceGroup
				logMetricObj.MonitorType = logMetricMonitor.MonitorType
			}
		}
		for _, logMetricObj := range tmpConfig.MetricConfigList {
			logMetricObj.ServiceGroup = serviceGroup
			logMetricObj.MonitorType = logMetricMonitor.MonitorType
		}
		for _, logMetricGroupObj := range tmpConfig.MetricGroups {
			var logMonitorTemplateDto *models.LogMonitorTemplateDto
			logMetricGroupObj.ServiceGroup = serviceGroup
			logMetricGroupObj.MonitorType = logMetricMonitor.MonitorType
			if strings.TrimSpace(logMetricGroupObj.LogMonitorTemplate) != "" {
				if logMonitorTemplateDto, err = GetLogMonitorTemplate(logMetricGroupObj.LogMonitorTemplate); err != nil {
					return
				}
				logMetricGroupObj.LogMonitorTemplateDto = logMonitorTemplateDto
			}
		}
		result.Config = append(result.Config, &tmpConfig)
	}
	result.DBConfig, err = GetDbMetricByServiceGroup(serviceGroup, metricKey)
	if err != nil {
		err = fmt.Errorf("query db metric config fail,%s ", err.Error())
	}
	return
}

func GetLogMetricByEndpoint(endpoint, metricKey string, onlySource bool) (result []*models.LogMetricQueryObj, err error) {
	result = []*models.LogMetricQueryObj{}
	var endpointServiceRelTable []*models.EndpointServiceRelTable
	if onlySource {
		err = x.SQL("select distinct t2.service_group from log_metric_endpoint_rel t1 left join log_metric_monitor t2 on t1.log_metric_monitor=t2.guid where t1.source_endpoint=?", endpoint).Find(&endpointServiceRelTable)
	} else {
		err = x.SQL("select distinct t2.service_group from log_metric_endpoint_rel t1 left join log_metric_monitor t2 on t1.log_metric_monitor=t2.guid where t1.source_endpoint=? or t1.target_endpoint=?", endpoint, endpoint).Find(&endpointServiceRelTable)
	}
	for _, v := range endpointServiceRelTable {
		tmpObj, tmpErr := GetLogMetricByServiceGroup(v.ServiceGroup, metricKey)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		for _, logMetricConfig := range tmpObj.Config {
			for _, logMetricGroup := range logMetricConfig.MetricGroups {
				for _, tmpLogMetric := range logMetricGroup.MetricList {
					if tmpLogMetric.Metric == "req_suc_count" {
						if len(tmpLogMetric.TagConfigList) == 1 {
							tmpLogMetric.TagConfigList = append(tmpLogMetric.TagConfigList, "retcode")
						}
					}
				}
			}
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

func GetServiceGroupEndpointRel(serviceGroup, sourceType, targetType string) (result []*models.LogMetricEndpointRelTable, err error) {
	result = []*models.LogMetricEndpointRelTable{}
	var guidList []string
	guidList, err = fetchGlobalServiceGroupChildGuidList(serviceGroup)
	if err != nil {
		return
	}
	var endpointTable []*models.EndpointNewTable
	err = x.SQL("select guid,monitor_type,ip from endpoint_new where guid in (select endpoint from endpoint_service_rel where service_group in ('" + strings.Join(guidList, "','") + "'))").Find(&endpointTable)
	if err != nil {
		return
	}
	sourceMap := make(map[string]string)
	targetMap := make(map[string]string)
	var tmpResult []*models.LogMetricEndpointRelTable
	for _, v := range endpointTable {
		if v.MonitorType == sourceType {
			sourceMap[v.Guid] = v.Ip
			tmpResult = append(tmpResult, &models.LogMetricEndpointRelTable{SourceEndpoint: v.Guid})
		}
		if v.MonitorType == targetType {
			targetMap[v.Ip] = v.Guid
		}
	}
	for _, v := range tmpResult {
		if targetGuid, b := targetMap[sourceMap[v.SourceEndpoint]]; b {
			v.TargetEndpoint = targetGuid
			result = append(result, v)
		}
	}
	return
}

func ListLogMetricJson(logMetricMonitor string) (result []*models.LogMetricJsonObj) {
	result = []*models.LogMetricJsonObj{}
	var logMetricJsonTable []*models.LogMetricJsonTable
	x.SQL("select * from log_metric_json where log_metric_monitor=?", logMetricMonitor).Find(&logMetricJsonTable)
	for _, v := range logMetricJsonTable {
		result = append(result, &models.LogMetricJsonObj{Guid: v.Guid, LogMetricMonitor: v.LogMetricMonitor, JsonRegular: v.JsonRegular, Tags: v.Tags, MetricList: ListLogMetricConfig(v.Guid, "")})
	}
	return result
}

func ListLogMetricConfig(logMetricJson, logMetricMonitor string) (result []*models.LogMetricConfigObj) {
	result = []*models.LogMetricConfigObj{}
	var logMetricConfigTable []*models.LogMetricConfigTable
	if logMetricJson != "" {
		x.SQL("select * from log_metric_config where log_metric_json=? and update_user='old_data'", logMetricJson).Find(&logMetricConfigTable)
	} else {
		x.SQL("select * from log_metric_config where log_metric_monitor=? and update_user='old_data'", logMetricMonitor).Find(&logMetricConfigTable)
	}
	for _, v := range logMetricConfigTable {
		tmpTagConfig := []*models.LogMetricConfigTag{}
		if v.TagConfig != "" {
			if tmpErr := json.Unmarshal([]byte(v.TagConfig), &tmpTagConfig); tmpErr != nil {
				log.Warn(nil, log.LOGGER_APP, "query log metric config warning with json unmarshal error", zap.String("tagConfig", v.TagConfig), zap.Error(tmpErr))
			}
		}
		tmpJsonTagList := []string{}
		for _, tagConfigItem := range tmpTagConfig {
			tmpJsonTagList = append(tmpJsonTagList, tagConfigItem.Key)
		}
		result = append(result, &models.LogMetricConfigObj{Guid: v.Guid, LogMetricMonitor: v.LogMetricMonitor, LogMetricJson: v.LogMetricJson, Metric: v.Metric, DisplayName: v.DisplayName, JsonKey: v.JsonKey, Regular: v.Regular, AggType: v.AggType, Step: v.Step, StringMap: ListLogMetricStringMap(v.Guid), TagConfig: tmpTagConfig, JsonTagList: tmpJsonTagList})
	}
	return result
}

func ListLogMetricStringMap(logMetricConfig string) (result []*models.LogMetricStringMapTable) {
	result = []*models.LogMetricStringMapTable{}
	x.SQL("select * from log_metric_string_map where log_metric_config=?", logMetricConfig).Find(&result)
	return result
}

func GetLogMetricMonitorByCond(logPath []string, guid, serviceGroup string) (list []*models.LogMetricMonitorTable, err error) {
	list = []*models.LogMetricMonitorTable{}
	for _, path := range logPath {
		var subList []*models.LogMetricMonitorTable
		if guid == "" {
			err = x.SQL("select * from log_metric_monitor where log_path=? and service_group=?", path, serviceGroup).Find(&subList)
		} else {
			err = x.SQL("select * from log_metric_monitor where log_path=? and service_group=? and guid <>?", path, serviceGroup, guid).Find(&subList)
		}
		if len(subList) > 0 {
			list = append(list, subList...)
		}
	}
	return
}

func CreateLogMetricMonitor(param *models.LogMetricMonitorCreateDto) error {
	if len(param.LogPath) == 0 {
		return nil
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	logMonitorGuidList := guid.CreateGuidList(len(param.LogPath))
	for i, v := range param.LogPath {
		tmpLogPath := strings.TrimSpace(v)
		actions = append(actions, &Action{Sql: "insert into log_metric_monitor(guid,service_group,log_path,metric_type,monitor_type,update_time) value (?,?,?,?,?,?)", Param: []interface{}{logMonitorGuidList[i], param.ServiceGroup, tmpLogPath, param.MetricType, param.MonitorType, nowTime}})
		relGuidList := guid.CreateGuidList(len(param.EndpointRel))
		for ii, vv := range param.EndpointRel {
			if vv.TargetEndpoint == "" {
				continue
			}
			actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{relGuidList[ii], logMonitorGuidList[i], vv.SourceEndpoint, vv.TargetEndpoint}})
		}
	}
	return Transaction(actions)
}

func GetLogMetricMonitor(logMetricMonitorGuid string) (result models.LogMetricMonitorObj, err error) {
	var logMetricMonitorTable []*models.LogMetricMonitorTable
	err = x.SQL("select * from log_metric_monitor where guid=?", logMetricMonitorGuid).Find(&logMetricMonitorTable)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_monitor fail,%s ", err.Error())
	}
	if len(logMetricMonitorTable) == 0 {
		return result, fmt.Errorf("Can not find log_metric_monitor with guid:%s ", logMetricMonitorGuid)
	}
	result = models.LogMetricMonitorObj{Guid: logMetricMonitorTable[0].Guid, ServiceGroup: logMetricMonitorTable[0].ServiceGroup, LogPath: logMetricMonitorTable[0].LogPath, MetricType: logMetricMonitorTable[0].MetricType, MonitorType: logMetricMonitorTable[0].MonitorType}
	result.EndpointRel = ListLogMetricEndpointRel(logMetricMonitorTable[0].Guid)
	result.JsonConfigList = ListLogMetricJson(logMetricMonitorTable[0].Guid)
	result.MetricConfigList = ListLogMetricConfig("", logMetricMonitorTable[0].Guid)
	result.MetricGroups = ListLogMetricGroups(logMetricMonitorTable[0].Guid, "")
	return result, nil
}

func UpdateLogMetricMonitor(param *models.LogMetricMonitorObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	actions = append(actions, &Action{Sql: "update log_metric_monitor set log_path=?,monitor_type=?,update_time=? where guid=?", Param: []interface{}{param.LogPath, param.MonitorType, nowTime, param.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=?", Param: []interface{}{param.Guid}})
	guidList := guid.CreateGuidList(len(param.EndpointRel))
	for i, v := range param.EndpointRel {
		if v.TargetEndpoint == "" {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guidList[i], param.Guid, v.SourceEndpoint, v.TargetEndpoint}})
	}
	return Transaction(actions)
}

func DeleteLogMetricMonitor(logMetricMonitorGuid string) (err error) {
	var logMetricMonitorTable []*models.LogMetricMonitorTable
	err = x.SQL("select * from log_metric_monitor where guid=?", logMetricMonitorGuid).Find(&logMetricMonitorTable)
	if len(logMetricMonitorTable) == 0 {
		return
	}
	actions, affectHost, affectEndpointGroup := getDeleteLogMetricMonitor(logMetricMonitorGuid)
	err = Transaction(actions)
	if err != nil {
		return err
	}
	if len(affectHost) > 0 {
		err = SyncLogMetricExporterConfig(affectHost)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "SyncLogMetricExporterConfig fail", zap.Error(err))
		}
	}
	if len(affectEndpointGroup) > 0 {
		for _, v := range affectEndpointGroup {
			err = SyncPrometheusRuleFile(v, false)
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "SyncPrometheusRuleFile fail", zap.Error(err))
			}
		}
	}
	return nil
}

func getDeleteLogMetricMonitor(logMetricMonitorGuid string) (actions []*Action, affectHost, affectEndpointGroup []string) {
	endpointRel := ListLogMetricEndpointRel(logMetricMonitorGuid)
	jsonConfigList := ListLogMetricJson(logMetricMonitorGuid)
	metricConfigList := ListLogMetricConfig("", logMetricMonitorGuid)
	logMetricGroupList := ListLogMetricGroups(logMetricMonitorGuid, "")
	for _, v := range endpointRel {
		affectHost = append(affectHost, v.SourceEndpoint)
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=?", Param: []interface{}{logMetricMonitorGuid}})
	for _, v := range jsonConfigList {
		for _, vv := range v.MetricList {
			deleteActions, tmpEndpointGroup := getDeleteLogMetricConfigAction(vv.Guid, logMetricMonitorGuid)
			actions = append(actions, deleteActions...)
			affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
		}
		actions = append(actions, &Action{Sql: "delete from log_metric_json where guid=?", Param: []interface{}{v.Guid}})
	}
	for _, v := range metricConfigList {
		deleteActions, tmpEndpointGroup := getDeleteLogMetricConfigAction(v.Guid, logMetricMonitorGuid)
		actions = append(actions, deleteActions...)
		affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
	}
	for _, v := range logMetricGroupList {
		deleteActions, tmpEndpointGroup, _, tmpErr := getDeleteLogMetricGroupActions(v.Guid)
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "try to get delete logMetricGroupAction fail", zap.String("logMetricGroupGuid", v.Guid), zap.Error(tmpErr))
			continue
		} else {
			actions = append(actions, deleteActions...)
			affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
		}
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_monitor where guid=?", Param: []interface{}{logMetricMonitorGuid}})
	return
}

func GetLogMetricJson(logMetricJsonGuid string) (result models.LogMetricJsonObj, err error) {
	var logMetricJsonTable []*models.LogMetricJsonTable
	err = x.SQL("select * from log_metric_json where guid=?", logMetricJsonGuid).Find(&logMetricJsonTable)
	if err != nil {
		return result, fmt.Errorf("Query log_metric_json table fail,%s ", err.Error())
	}
	if len(logMetricJsonTable) == 0 {
		return result, fmt.Errorf("Can not find log_metric_json with guid:%s ", logMetricJsonGuid)
	}
	result = models.LogMetricJsonObj{Guid: logMetricJsonTable[0].Guid, Name: logMetricJsonTable[0].Name, LogMetricMonitor: logMetricJsonTable[0].LogMetricMonitor, JsonRegular: logMetricJsonTable[0].JsonRegular, Tags: logMetricJsonTable[0].Tags, DemoLog: logMetricJsonTable[0].DemoLog, TrialCalculationResult: []string{}}
	json.Unmarshal([]byte(logMetricJsonTable[0].CalcResult), &result.TrialCalculationResult)
	result.MetricList = ListLogMetricConfig(logMetricJsonGuid, "")
	return
}

func CreateLogMetricJson(param *models.LogMetricJsonObj, operator string) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	param.Guid = guid.CreateGuid()
	calcResultBytes, _ := json.Marshal(param.TrialCalculationResult)
	actions = append(actions, &Action{Sql: "insert into log_metric_json(guid,name,log_metric_monitor,json_regular,tags,demo_log,calc_result,update_time) value (?,?,?,?,?,?,?,?)", Param: []interface{}{param.Guid, param.Name, param.LogMetricMonitor, param.JsonRegular, param.Tags, param.DemoLog, string(calcResultBytes), nowTime}})
	guidList := guid.CreateGuidList(len(param.MetricList))
	for i, v := range param.MetricList {
		v.LogMetricJson = param.Guid
		v.LogMetricMonitor = param.LogMetricMonitor
		v.Guid = guidList[i]
		tmpActions := getCreateLogMetricConfigAction(v, nowTime, operator)
		actions = append(actions, tmpActions...)
	}
	return Transaction(actions)
}

func UpdateLogMetricJson(param *models.LogMetricJsonObj, operator string) error {
	if param.LogMetricMonitor == "" {
		logMetricMonitorGuid, err := getLogMetricJsonMonitor(param.Guid)
		if err != nil {
			return err
		}
		param.LogMetricMonitor = logMetricMonitorGuid
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	calcResultBytes, _ := json.Marshal(param.TrialCalculationResult)
	actions = append(actions, &Action{Sql: "update log_metric_json set name=?,json_regular=?,tags=?,demo_log=?,calc_result=?,update_time=? where guid=?", Param: []interface{}{param.Name, param.JsonRegular, param.Tags, param.DemoLog, string(calcResultBytes), nowTime, param.Guid}})
	var logMetricConfigTable []*models.LogMetricConfigTable
	x.SQL("select * from log_metric_config where log_metric_json=?", param.Guid).Find(&logMetricConfigTable)
	var affectEndpointGroup []string
	for _, v := range param.MetricList {
		v.LogMetricJson = param.Guid
		v.LogMetricMonitor = param.LogMetricMonitor
		if v.Guid == "" {
			actions = append(actions, getCreateLogMetricConfigAction(v, nowTime, operator)...)
			continue
		}
		tmpUpdateActions, tmpEndpointGroup := getUpdateLogMetricConfigAction(v, operator, nowTime)
		actions = append(actions, tmpUpdateActions...)
		affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
	}
	for _, v := range logMetricConfigTable {
		existFlag := false
		for _, vv := range param.MetricList {
			if v.Guid == vv.Guid {
				existFlag = true
				break
			}
		}
		if !existFlag {
			deleteActions, tmpEndpointGroup := getDeleteLogMetricConfigAction(v.Guid, param.LogMetricMonitor)
			actions = append(actions, deleteActions...)
			affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
		}
	}
	err := Transaction(actions)
	if err == nil && len(affectEndpointGroup) > 0 {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return err
}

func DeleteLogMetricJson(logMetricJsonGuid string) (logMetricMonitorGuid string, err error) {
	logMetricMonitorGuid, err = getLogMetricJsonMonitor(logMetricJsonGuid)
	var actions []*Action
	var logMetricConfigTable []*models.LogMetricConfigTable
	var affectEndpointGroup []string
	x.SQL("select * from log_metric_config where log_metric_json=?", logMetricJsonGuid).Find(&logMetricConfigTable)
	for _, v := range logMetricConfigTable {
		deleteActions, tmpEndpointGroup := getDeleteLogMetricConfigAction(v.Guid, logMetricMonitorGuid)
		actions = append(actions, deleteActions...)
		affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_json where guid=?", Param: []interface{}{logMetricJsonGuid}})
	err = Transaction(actions)
	if err == nil && len(affectEndpointGroup) > 0 {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return
}

func getLogMetricJsonMonitor(logMetricJsonGuid string) (logMetricMonitorGuid string, err error) {
	var logMetricJsonTable []*models.LogMetricJsonTable
	err = x.SQL("select * from log_metric_json where guid=?", logMetricJsonGuid).Find(&logMetricJsonTable)
	if err != nil {
		return logMetricMonitorGuid, fmt.Errorf("Query log_metric_json fail,%s ", err.Error())
	}
	if len(logMetricJsonTable) == 0 {
		return logMetricMonitorGuid, fmt.Errorf("Can not find log_metric_json with guid:%s ", logMetricJsonGuid)
	}
	logMetricMonitorGuid = logMetricJsonTable[0].LogMetricMonitor
	return logMetricMonitorGuid, nil
}

func GetLogMetricConfig(logMetricConfigGuid string) (result models.LogMetricConfigObj, err error) {
	var logMetricConfigTable []*models.LogMetricConfigTable
	err = x.SQL("select * from log_metric_config where guid=?", logMetricConfigGuid).Find(&logMetricConfigTable)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_config fail,%s ", err.Error())
	}
	if len(logMetricConfigTable) == 0 {
		return result, fmt.Errorf("Can not find log_metric_config with guid:%s ", logMetricConfigGuid)
	}
	result = models.LogMetricConfigObj{Guid: logMetricConfigGuid, LogMetricMonitor: logMetricConfigTable[0].LogMetricMonitor, LogMetricJson: logMetricConfigTable[0].LogMetricJson, Metric: logMetricConfigTable[0].Metric, DisplayName: logMetricConfigTable[0].DisplayName, JsonKey: logMetricConfigTable[0].JsonKey, Regular: logMetricConfigTable[0].Regular, AggType: logMetricConfigTable[0].AggType, Step: logMetricConfigTable[0].Step}
	result.StringMap = ListLogMetricStringMap(logMetricConfigGuid)
	return
}

func CreateLogMetricConfig(param *models.LogMetricConfigObj, operator string) error {
	param.Guid = guid.CreateGuid()
	actions := getCreateLogMetricConfigAction(param, time.Now().Format(models.DatetimeFormat), operator)
	return Transaction(actions)
}

func UpdateLogMetricConfig(param *models.LogMetricConfigObj, operator string) error {
	logMetricMonitorGuid, err := getLogMetricConfigMonitor(param.Guid)
	if err != nil {
		return fmt.Errorf("Query table log_metric_config fail,%s ", err.Error())
	}
	param.LogMetricMonitor = logMetricMonitorGuid
	actions, affectEndpointGroup := getUpdateLogMetricConfigAction(param, operator, time.Now().Format(models.DatetimeFormat))
	err = Transaction(actions)
	if err == nil {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return err
}

func DeleteLogMetricConfig(logMetricConfigGuid string) (logMetricMonitorGuid string, err error) {
	logMetricMonitorGuid, err = getLogMetricConfigMonitor(logMetricConfigGuid)
	actions, affectEndpointGroup := getDeleteLogMetricConfigAction(logMetricConfigGuid, logMetricMonitorGuid)
	err = Transaction(actions)
	if err == nil && len(affectEndpointGroup) > 0 {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return
}

func getLogMetricConfigMonitor(logMetricConfigGuid string) (logMetricMonitorGuid string, err error) {
	var logMetricConfigTable []*models.LogMetricConfigTable
	err = x.SQL("select guid,log_metric_monitor from log_metric_config where guid=?", logMetricConfigGuid).Find(&logMetricConfigTable)
	if err != nil {
		return logMetricMonitorGuid, fmt.Errorf("Query log_metric_config fail,%s ", err.Error())
	}
	if len(logMetricConfigTable) == 0 {
		return logMetricMonitorGuid, fmt.Errorf("Can not find log_metric_config with guid:%s ", logMetricConfigGuid)
	}
	logMetricMonitorGuid = logMetricConfigTable[0].LogMetricMonitor
	return logMetricMonitorGuid, err
}

func getCreateLogMetricConfigAction(param *models.LogMetricConfigObj, nowTime, operator string) []*Action {
	var actions []*Action
	//if param.Guid == "" {
	//	param.Guid = guid.CreateGuid()
	//}
	param.Guid = "lmc_" + guid.CreateGuid()
	tagString := ""
	for _, jsonTagItem := range param.JsonTagList {
		param.TagConfig = append(param.TagConfig, &models.LogMetricConfigTag{Key: jsonTagItem})
	}
	if len(param.TagConfig) > 0 {
		tagBytes, _ := json.Marshal(param.TagConfig)
		tagString = string(tagBytes)
	}
	param.Step = 10
	if param.LogMetricJson != "" {
		actions = append(actions, &Action{Sql: "insert into log_metric_config(guid,log_metric_json,metric,display_name,json_key,regular,agg_type,step,update_time,tag_config) value (?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{param.Guid, param.LogMetricJson, param.Metric, param.DisplayName, param.JsonKey, param.Regular, param.AggType, param.Step, nowTime, tagString}})
	} else {
		actions = append(actions, &Action{Sql: "insert into log_metric_config(guid,log_metric_monitor,metric,display_name,json_key,regular,agg_type,step,update_time,tag_config) value (?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{param.Guid, param.LogMetricMonitor, param.Metric, param.DisplayName, param.JsonKey, param.Regular, param.AggType, param.Step, nowTime, tagString}})
	}
	if param.ServiceGroup == "" || param.MonitorType == "" {
		param.ServiceGroup, param.MonitorType = GetLogMetricServiceGroup(param.LogMetricMonitor)
	}
	actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,log_metric_config,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?,?)",
		Param: []interface{}{fmt.Sprintf("%s__%s", param.Metric, param.ServiceGroup), param.Metric, param.MonitorType, getLogMetricExprByAggType(param.Metric, param.AggType, param.ServiceGroup, []string{}), param.ServiceGroup,
			models.MetricWorkspaceService, nowTime, param.Guid, nowTime, operator, operator}})
	guidList := guid.CreateGuidList(len(param.StringMap))
	for i, v := range param.StringMap {
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_config,source_value,regulative,target_value,update_time) value (?,?,?,?,?,?)", Param: []interface{}{guidList[i], param.Guid, v.SourceValue, v.Regulative, v.TargetValue, nowTime}})
	}
	return actions
}

func getLogMetricExprByAggType(metric, aggType, serviceGroup string, tagList []string) (result string) {
	var tagString, tagFilterString string
	if len(tagList) > 0 {
		tagString = "," + strings.Join(tagList, ",")
		for _, v := range tagList {
			tagFilterString += fmt.Sprintf(",%s=\"$t_%s\"", v, v)
		}
	}
	switch aggType {
	case "sum":
		result = fmt.Sprintf("sum(%s{key=\"%s\",agg=\"%s\",service_group=\"%s\"%s}) by (key,agg,service_group%s)", models.LogMetricName, metric, aggType, serviceGroup, tagFilterString, tagString)
	case "count":
		result = fmt.Sprintf("sum(%s{key=\"%s\",agg=\"%s\",service_group=\"%s\"%s}) by (key,agg,service_group%s)", models.LogMetricName, metric, aggType, serviceGroup, tagFilterString, tagString)
	case "max":
		result = fmt.Sprintf("max(%s{key=\"%s\",agg=\"%s\",service_group=\"%s\"%s}) by (key,agg,service_group%s)", models.LogMetricName, metric, aggType, serviceGroup, tagFilterString, tagString)
	case "min":
		result = fmt.Sprintf("min(%s{key=\"%s\",agg=\"%s\",service_group=\"%s\"%s}) by (key,agg,service_group%s)", models.LogMetricName, metric, aggType, serviceGroup, tagFilterString, tagString)
	case "avg":
		result = fmt.Sprintf("sum(%s{key=\"%s\",agg=\"sum\",service_group=\"%s\"%s}) by (key,service_group%s)/sum(%s{key=\"%s\",agg=\"count\",service_group=\"%s\"%s}) by (key,service_group%s) > 0 or (0*sum(%s{key=\"%s\",agg=\"sum\",service_group=\"%s\"%s}) by (key,service_group%s))", models.LogMetricName, metric, serviceGroup, tagFilterString, tagString, models.LogMetricName, metric, serviceGroup, tagFilterString, tagString, models.LogMetricName, metric, serviceGroup, tagFilterString, tagString)
	default:
		result = fmt.Sprintf("%s{key=\"%s\",agg=\"%s\",service_group=\"%s\"%s}", models.LogMetricName, metric, aggType, serviceGroup, tagFilterString)
	}
	return result
}

func getLogMetricRatePromExpr(metric, metricPrefix, aggType, serviceGroup, sucRetCode string) (result string) {
	aggType = "count"
	if metricPrefix != "" {
		metricPrefix = metricPrefix + "_"
	}
	if metric == "req_suc_count" {
		result = fmt.Sprintf("sum(%s{key=\"%sreq_suc_count\",agg=\"%s\",service_group=\"%s\",retcode=\"%s\",code=\"$t_code\"}) by (key,agg,service_group,code,retcode)", models.LogMetricName, metricPrefix, aggType, serviceGroup, sucRetCode)
		return
	}
	if metric == "req_fail_count_detail" {
		result = fmt.Sprintf("sum(%s{key=\"%sreq_suc_count\",agg=\"%s\",service_group=\"%s\",retcode!=\"%s\",retcode=\"$t_retcode\",code=\"$t_code\"}) by (key,agg,service_group,code,retcode)", models.LogMetricName, metricPrefix, aggType, serviceGroup, sucRetCode)
		return
	}
	if metric == "req_fail_count" {
		// 失败数：直接统计 retcode!="success" 的请求数. 失败数=总数-成功数,当接口一开始只有失败请求时候,图表数据出不来
		result = fmt.Sprintf("sum(%s{key=\"%sreq_suc_count\",agg=\"%s\",service_group=\"%s\",code=\"$t_code\",retcode!=\"%s\"}) by (key,agg,service_group,code,retcode)", models.LogMetricName, metricPrefix, aggType, serviceGroup, sucRetCode)
		return
	}
	if metric == "req_suc_rate" {
		// 成功率 = 成功数/总数，没有请求时返回NaN
		result = fmt.Sprintf("100 * (sum(%s{key=\"%sreq_suc_count\",agg=\"%s\",service_group=\"%s\",retcode=\"%s\",code=\"$t_code\"}) by (service_group,code) / sum(%s{key=\"%sreq_count\",agg=\"%s\",service_group=\"%s\",code=\"$t_code\"}) by (service_group,code))",
			models.LogMetricName, metricPrefix, aggType, serviceGroup, sucRetCode, models.LogMetricName, metricPrefix, aggType, serviceGroup)
	}
	if metric == "req_fail_rate" {
		result = fmt.Sprintf("100*((sum(%s{key=\"%sreq_suc_count\",agg=\"%s\",service_group=\"%s\",retcode!=\"%s\",code=\"$t_code\"}) by (service_group,code))/(sum(%s{key=\"%sreq_count\",agg=\"%s\",service_group=\"%s\",code=\"$t_code\"}) by (service_group,code) > 0 or (sum(%s{key=\"%sreq_count\",agg=\"%s\",service_group=\"%s\",code=\"$t_code\"}) by (service_group,code) + 1)))",
			models.LogMetricName, metricPrefix, aggType, serviceGroup, sucRetCode, models.LogMetricName, metricPrefix, aggType, serviceGroup, models.LogMetricName, metricPrefix, aggType, serviceGroup)

	}
	return
}

func GetLogMetricServiceGroup(logMetricMonitor string) (serviceGroup, monitorType string) {
	var logMetricMonitorTable []*models.LogMetricMonitorTable
	x.SQL("select guid,service_group,monitor_type from log_metric_monitor where guid=?", logMetricMonitor).Find(&logMetricMonitorTable)
	if len(logMetricMonitorTable) > 0 {
		serviceGroup = logMetricMonitorTable[0].ServiceGroup
		monitorType = logMetricMonitorTable[0].MonitorType
	}
	return
}

func getUpdateLogMetricConfigAction(param *models.LogMetricConfigObj, operator, nowTime string) (actions []*Action, affectEndpointGroup []string) {
	param.Step = 10
	var logMetricConfigTable []*models.LogMetricConfigTable
	x.SQL("select * from log_metric_config where guid=?", param.Guid).Find(&logMetricConfigTable)
	if len(logMetricConfigTable) > 0 {
		if logMetricConfigTable[0].Metric != param.Metric || logMetricConfigTable[0].AggType != param.AggType {
			serviceGroup, _ := GetLogMetricServiceGroup(param.LogMetricMonitor)
			oldMetricGuid := fmt.Sprintf("%s__%s", logMetricConfigTable[0].Metric, serviceGroup)
			newMetricGuid := fmt.Sprintf("%s__%s", param.Metric, serviceGroup)
			actions = append(actions, &Action{Sql: "update metric set guid=?,metric=?,prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{newMetricGuid, param.Metric, getLogMetricExprByAggType(param.Metric, param.AggType, serviceGroup, []string{}), operator, nowTime, oldMetricGuid}})
			var alarmStrategyTable []*models.AlarmStrategyTable
			x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", oldMetricGuid).Find(&alarmStrategyTable)
			if len(alarmStrategyTable) > 0 {
				for _, v := range alarmStrategyTable {
					affectEndpointGroup = append(affectEndpointGroup, v.EndpointGroup)
				}
				actions = append(actions, &Action{Sql: "update alarm_strategy set metric=? where metric=?", Param: []interface{}{newMetricGuid, oldMetricGuid}})
			}
		}
	}
	tagString := ""
	for _, jsonTagItem := range param.JsonTagList {
		param.TagConfig = append(param.TagConfig, &models.LogMetricConfigTag{Key: jsonTagItem})
	}
	if len(param.TagConfig) > 0 {
		tagBytes, _ := json.Marshal(param.TagConfig)
		tagString = string(tagBytes)
	}
	actions = append(actions, &Action{Sql: "update log_metric_config set metric=?,display_name=?,json_key=?,regular=?,agg_type=?,step=?,update_time=?,tag_config=? where guid=?", Param: []interface{}{param.Metric, param.DisplayName, param.JsonKey, param.Regular, param.AggType, param.Step, nowTime, tagString, param.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config=?", Param: []interface{}{param.Guid}})
	guidList := guid.CreateGuidList(len(param.StringMap))
	for i, v := range param.StringMap {
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_config,source_value,regulative,target_value,update_time) value (?,?,?,?,?,?)", Param: []interface{}{guidList[i], param.Guid, v.SourceValue, v.Regulative, v.TargetValue, nowTime}})
	}
	return
}

func getDeleteLogMetricConfigAction(logMetricConfigGuid, logMetricMonitorGuid string) (actions []*Action, endpointGroup []string) {
	lmObj, err := getSimpleLogMetricConfig(logMetricConfigGuid)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "getDeleteLogMetricConfigAction", zap.Error(err))
		return
	}
	serviceGroup, _ := GetLogMetricServiceGroup(logMetricMonitorGuid)
	alarmMetricGuid := fmt.Sprintf("%s__%s", lmObj.Metric, serviceGroup)
	var alarmStrategyTable []*models.AlarmStrategyTable
	x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", alarmMetricGuid).Find(&alarmStrategyTable)
	for _, v := range alarmStrategyTable {
		endpointGroup = append(endpointGroup, v.EndpointGroup)
	}
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where metric=?", Param: []interface{}{alarmMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{alarmMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config=?", Param: []interface{}{logMetricConfigGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_config where guid=?", Param: []interface{}{logMetricConfigGuid}})
	return
}

func getSimpleLogMetricConfig(logMetricConfigGuid string) (result models.LogMetricConfigTable, err error) {
	var queryTable []*models.LogMetricConfigTable
	err = x.SQL("select * from log_metric_config where guid=?", logMetricConfigGuid).Find(&queryTable)
	if err != nil {
		return result, err
	}
	if len(queryTable) == 0 {
		return result, fmt.Errorf("Can not find logMetricConfig with guid:%s ", logMetricConfigGuid)
	}
	result = *queryTable[0]
	return
}

func GetServiceGroupByLogMetricMonitor(logMetricMonitorGuid string) string {
	if logMetricMonitorGuid == "" {
		return ""
	}
	var logMetricMonitorTable []*models.LogMetricMonitorTable
	x.SQL("select * from log_metric_monitor where guid=?", logMetricMonitorGuid).Find(&logMetricMonitorTable)
	if len(logMetricMonitorTable) > 0 {
		return logMetricMonitorTable[0].ServiceGroup
	}
	return ""
}

func CheckRegExpMatchPCRE(param models.CheckRegExpParam) (message, matchString string) {
	re, tmpErr := pcre.Compile(param.RegString, 0)
	if tmpErr != nil {
		return fmt.Sprintf("reg compile fail,%s ", tmpErr.Message), matchString
	}
	matchString = pcreMatchSubString(&re, param.TestContext)
	if matchString == "" {
		return fmt.Sprintf("can not match any data"), matchString
	}
	return fmt.Sprintf("success match:%s", matchString), matchString
}

func CheckRegExpMatch(param models.CheckRegExpParam) (message string) {
	re, tmpErr := regexp2.Compile(param.RegString, 0)
	if tmpErr != nil {
		return fmt.Sprintf("reg compile fail,%s ", tmpErr.Error())
	}
	matchString := regexp2FindStringMatch(re, param.TestContext)
	if matchString == "" {
		return fmt.Sprintf("can not match any data")
	}
	return fmt.Sprintf("success match:%s", matchString)
}

func pcreMatchSubString(re *pcre.Regexp, lineText string) (matchString string) {
	if re == nil {
		return
	}
	lineText = strings.TrimSpace(lineText)
	mat := re.MatcherString(lineText, 0)
	if mat != nil {
		for i := 0; i <= mat.Groups(); i++ {
			groupString := mat.GroupString(i)
			if (i == 0 && groupString == lineText) || groupString == "" {
				continue
			}
			matchString = groupString
			break
		}
	}
	return
}

func regexp2FindStringMatch(re *regexp2.Regexp, lineText string) (matchString string) {
	if re == nil {
		return
	}
	mat, err := re.FindStringMatch(lineText)
	if err != nil || mat == nil {
		return
	}
	for i, v := range mat.Groups() {
		groupString := v.String()
		if (i == 0 && groupString == lineText) || groupString == "" {
			continue
		}
		matchString = groupString
		break
	}
	return
}

func ImportLogMetric(param *models.LogMetricQueryObj, operator string, roles []string, errMsgObj *models.ErrorTemplate) (err error) {
	var actions []*Action
	var dashboardIdList []int64
	var existLogMetricMonitorMap = make(map[string]*models.LogMetricMonitorObj)
	existData, queryErr := GetLogMetricByServiceGroup(param.Guid, "")
	if queryErr != nil {
		return fmt.Errorf("get exist log metric data fail,%s ", queryErr.Error())
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	affectHostMap := make(map[string]int)
	affectEndpointGroupMap := make(map[string]int)
	serviceGroupMetricMap, getErr := getServiceGroupMetricMap(param.Guid)
	if getErr != nil {
		err = getErr
		return
	}
	// delete action
	if len(param.Config) > 0 {
		for _, existLogMonitor := range existData.Config {
			existLogMetricMonitorMap[existLogMonitor.LogPath] = existLogMonitor
			for _, v := range existLogMonitor.EndpointRel {
				affectHostMap[v.SourceEndpoint] = 1
			}
			tmpDeleteActions, affectHost, affectEndpointGroup := getDeleteLogMetricMonitor(existLogMonitor.Guid)
			actions = append(actions, tmpDeleteActions...)
			for _, v := range affectHost {
				affectHostMap[v] = 1
			}
			for _, v := range affectEndpointGroup {
				affectEndpointGroupMap[v] = 1
			}
		}
	}
	for _, inputLogMonitor := range param.Config {
		existObj := &models.LogMetricMonitorObj{}
		// 业务配置映射关系需要保留
		if v, ok := existLogMetricMonitorMap[inputLogMonitor.LogPath]; ok {
			existObj.EndpointRel = v.EndpointRel
		}
		if existObj.Guid != "" {
			if existObj.LogPath != inputLogMonitor.LogPath || existObj.MonitorType != inputLogMonitor.MonitorType {
				actions = append(actions, &Action{Sql: "update log_metric_monitor set log_path=?,monitor_type=? where guid=?", Param: []interface{}{inputLogMonitor.LogPath, inputLogMonitor.MonitorType, inputLogMonitor.Guid}})
			}
		} else {
			inputLogMonitor.Guid = "lmm_" + guid.CreateGuid()
			actions = append(actions, &Action{Sql: "insert into log_metric_monitor(guid,service_group,log_path,metric_type,monitor_type,update_time) value (?,?,?,?,?,?)", Param: []interface{}{inputLogMonitor.Guid, param.Guid, inputLogMonitor.LogPath, inputLogMonitor.MetricType, inputLogMonitor.MonitorType, nowTime}})
			if len(existObj.EndpointRel) > 0 {
				for _, endpointRel := range existObj.EndpointRel {
					actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), inputLogMonitor.Guid, endpointRel.SourceEndpoint, endpointRel.TargetEndpoint}})
				}
			} else {
				// 如果 不存在对象映射关系,查询所有可以添加的映射关系
				var endpointRelList []*models.LogMetricEndpointRelTable
				if endpointRelList, err = GetServiceGroupEndpointRel(param.ServiceGroupTable.Guid, "host", inputLogMonitor.MonitorType); err != nil {
					log.Error(nil, log.LOGGER_APP, "GetServiceGroupEndpointRel err", zap.Error(err))
				}
				for _, endpointRel := range endpointRelList {
					actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), inputLogMonitor.Guid, endpointRel.SourceEndpoint, endpointRel.TargetEndpoint}})
				}
			}
		}
		tmpActions, subDashboardIdList, tmpAffectHosts, tmpAffectEndpointGroup, tmpErr := getUpdateLogMetricMonitorByImport(existObj, inputLogMonitor, nowTime, operator, serviceGroupMetricMap, errMsgObj, roles)
		dashboardIdList = append(dashboardIdList, subDashboardIdList...)
		if tmpErr != nil {
			err = tmpErr
			deleteCustomDashboardList(dashboardIdList)
			return
		}
		actions = append(actions, tmpActions...)
		for _, v := range tmpAffectHosts {
			affectHostMap[v] = 1
		}
		for _, v := range tmpAffectEndpointGroup {
			affectEndpointGroupMap[v] = 1
		}
	}

	for _, dbConfig := range param.DBConfig {
		tmpDBMetricGuid := fmt.Sprintf("%s__%s", dbConfig.Metric, dbConfig.ServiceGroup)
		if tmpMetric, existFlag := serviceGroupMetricMap[tmpDBMetricGuid]; existFlag {
			err = fmt.Errorf("Metric: %s duplicate ", tmpMetric)
			deleteCustomDashboardList(dashboardIdList)
			return
		}
		// 查询所有可以添加的映射关系,重写导入EndpointRel关系
		var endpointRelList []*models.LogMetricEndpointRelTable
		if endpointRelList, err = GetServiceGroupEndpointRel(param.ServiceGroupTable.Guid, "mysql", dbConfig.MonitorType); err != nil {
			log.Error(nil, log.LOGGER_APP, "GetServiceGroupEndpointRel err", zap.Error(err))
			deleteCustomDashboardList(dashboardIdList)
		}
		dbConfig.EndpointRel = []*models.DbMetricEndpointRelTable{}
		for _, endpointRel := range endpointRelList {
			dbConfig.EndpointRel = append(dbConfig.EndpointRel, &models.DbMetricEndpointRelTable{SourceEndpoint: endpointRel.SourceEndpoint, TargetEndpoint: endpointRel.TargetEndpoint})
		}
		tmpActions, tmpAffectHosts, tmpAffectEndpointGroup, tmpErr := getCreateDBMetricMonitorByImport(dbConfig, nowTime, operator)
		if tmpErr != nil {
			err = tmpErr
			deleteCustomDashboardList(dashboardIdList)
			return
		}
		actions = append(actions, tmpActions...)
		for _, v := range tmpAffectHosts {
			affectHostMap[v] = 1
		}
		for _, v := range tmpAffectEndpointGroup {
			affectEndpointGroupMap[v] = 1
		}
	}
	var affectHostList, affectEndpointGroupList []string
	for k, _ := range affectHostMap {
		affectHostList = append(affectHostList, k)
	}
	for k, _ := range affectEndpointGroupMap {
		affectEndpointGroupList = append(affectEndpointGroupList, k)
	}
	log.Info(nil, log.LOGGER_APP, "importActions", zap.Int("length", len(actions)), zap.Strings("affectHostList", affectHostList), zap.Strings("affectEndpointGroupList", affectEndpointGroupList))
	err = Transaction(actions)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "import log monitor exec database fail", zap.Error(err))
		deleteCustomDashboardList(dashboardIdList)
		return
	}
	if tmpErr := SyncLogMetricExporterConfig(affectHostList); tmpErr != nil {
		log.Error(nil, log.LOGGER_APP, "sync log metric to affect host fail", zap.Error(tmpErr))
	}
	for _, v := range affectEndpointGroupList {
		if tmpErr := SyncPrometheusRuleFile(v, false); tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "sync prometheus rule file fail", zap.Error(tmpErr))
		}
	}
	return
}

func getUpdateLogMetricMonitorByImport(existObj, inputObj *models.LogMetricMonitorObj, nowTime, operator string, existMetricMap map[string]string, errMsgObj *models.ErrorTemplate, roles []string) (actions []*Action, dashboardId []int64, affectHost []string, affectEndpointGroup []string, err error) {
	dashboardId = []int64{}
	if existObj.Guid != "" {
		// compare log json monitor
		for _, inputJsonObj := range inputObj.JsonConfigList {
			matchExistJsonObj := &models.LogMetricJsonObj{}
			for _, existJsonObj := range existObj.JsonConfigList {
				if existJsonObj.Guid == inputJsonObj.Guid {
					matchExistJsonObj = existJsonObj
					break
				}
			}
			if matchExistJsonObj.Guid != "" {
				actions = append(actions, &Action{Sql: "update log_metric_json set json_regular=?,tags=?,update_time=? where guid=?", Param: []interface{}{inputJsonObj.JsonRegular, inputJsonObj.Tags, nowTime, inputJsonObj.Guid}})
				tmpActions, tmpAffectEndpointGroup := getCompareLogMetricConfigByImport(inputJsonObj.MetricList, matchExistJsonObj.MetricList, nowTime, operator)
				actions = append(actions, tmpActions...)
				affectEndpointGroup = append(affectEndpointGroup, tmpAffectEndpointGroup...)
			} else {
				actions = append(actions, &Action{Sql: "insert into log_metric_json(guid,log_metric_monitor,json_regular,tags,update_time) value (?,?,?,?,?)", Param: []interface{}{inputJsonObj.Guid, inputJsonObj.LogMetricMonitor, inputJsonObj.JsonRegular, inputJsonObj.Tags, nowTime}})
				for _, logMetricConfig := range inputJsonObj.MetricList {
					tmpActions := getCreateLogMetricConfigAction(logMetricConfig, nowTime, operator)
					actions = append(actions, tmpActions...)
				}
			}
		}
		for _, existJsonObj := range existObj.JsonConfigList {
			deleteFlag := true
			for _, inputJsonObj := range inputObj.JsonConfigList {
				if inputJsonObj.Guid == existJsonObj.Guid {
					deleteFlag = false
					break
				}
			}
			if deleteFlag {
				for _, v := range existJsonObj.MetricList {
					deleteActions, tmpEndpointGroup := getDeleteLogMetricConfigByImport(v)
					actions = append(actions, deleteActions...)
					affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
				}
				actions = append(actions, &Action{Sql: "delete from log_metric_json where guid=?", Param: []interface{}{existJsonObj.Guid}})
			}
		}
		// compare log metric config
		tmpActions, tmpAffectEndpointGroup := getCompareLogMetricConfigByImport(inputObj.MetricConfigList, existObj.MetricConfigList, nowTime, operator)
		actions = append(actions, tmpActions...)
		affectEndpointGroup = append(affectEndpointGroup, tmpAffectEndpointGroup...)
		// compare metric group config
		for _, inputMetricGroup := range inputObj.MetricGroups {
			matchMetricGroupObj := &models.LogMetricGroupObj{}
			for _, existMetricGroup := range existObj.MetricGroups {
				if existMetricGroup.Guid == inputMetricGroup.Guid {
					matchMetricGroupObj = existMetricGroup
					break
				}
			}
			if matchMetricGroupObj.Guid != "" {
				tmpMetricGroupActions, tmpAffect, tmpErr := getUpdateLogMetricGroupByImport(inputMetricGroup, operator, roles)
				if tmpErr != nil {
					err = tmpErr
					return
				}
				actions = append(actions, tmpMetricGroupActions...)
				affectEndpointGroup = append(affectEndpointGroup, tmpAffect...)
			} else {
				tmpMetricGroupActions, newDashboardId, tmpErr := getCreateLogMetricGroupByImport(inputMetricGroup, operator, existMetricMap, errMsgObj, roles)
				dashboardId = append(dashboardId, newDashboardId)
				if tmpErr != nil {
					err = tmpErr
					return
				}
				actions = append(actions, tmpMetricGroupActions...)
			}
		}
		for _, existMetricGroup := range existObj.MetricGroups {
			deleteFlag := true
			for _, inputMetricGroup := range inputObj.MetricGroups {
				if inputMetricGroup.Guid == existMetricGroup.Guid {
					deleteFlag = false
					break
				}
			}
			if deleteFlag {
				tmpMetricGroupActions, tmpAffect, _, tmpErr := getDeleteLogMetricGroupActions(existMetricGroup.Guid)
				if tmpErr != nil {
					err = tmpErr
					return
				}
				actions = append(actions, tmpMetricGroupActions...)
				affectEndpointGroup = append(affectEndpointGroup, tmpAffect...)
			}
		}
	} else {
		// create
		for _, inputJsonObj := range inputObj.JsonConfigList {
			inputJsonObj.Guid = guid.CreateGuid()
			inputJsonObj.LogMetricMonitor = inputObj.Guid
			actions = append(actions, &Action{Sql: "insert into log_metric_json(guid,log_metric_monitor,json_regular,tags,update_time) value (?,?,?,?,?)", Param: []interface{}{inputJsonObj.Guid, inputJsonObj.LogMetricMonitor, inputJsonObj.JsonRegular, inputJsonObj.Tags, nowTime}})
			for _, logMetricConfig := range inputJsonObj.MetricList {
				logMetricConfig.LogMetricJson = inputJsonObj.Guid
				logMetricConfig.LogMetricMonitor = inputObj.Guid
				tmpActions := getCreateLogMetricConfigAction(logMetricConfig, nowTime, operator)
				actions = append(actions, tmpActions...)
			}
		}
		for _, logMetricConfig := range inputObj.MetricConfigList {
			logMetricConfig.LogMetricMonitor = inputObj.Guid
			tmpActions := getCreateLogMetricConfigAction(logMetricConfig, nowTime, operator)
			actions = append(actions, tmpActions...)
		}
		for _, metricGroup := range inputObj.MetricGroups {
			var tempLogMetricGroup *models.LogMetricGroup
			// log_metric_group guid 先查询存在则新增
			if metricGroup.Guid == "" {
				metricGroup.Guid = "lmg_" + guid.CreateGuid()
			} else {
				if tempLogMetricGroup, err = GetLogMetricGroupById(metricGroup.Guid); err != nil {
					return
				}
				if tempLogMetricGroup != nil {
					metricGroup.Guid = "lmg_" + guid.CreateGuid()
				}
			}
			metricGroup.LogMetricMonitor = inputObj.Guid
			metricGroup.ServiceGroup = inputObj.ServiceGroup
			metricGroup.MonitorType = inputObj.MonitorType
			tmpActions, newDashboardId, tmpErr := getCreateLogMetricGroupByImport(metricGroup, operator, existMetricMap, errMsgObj, roles)
			dashboardId = append(dashboardId, newDashboardId)
			if tmpErr != nil {
				err = tmpErr
				return
			}
			actions = append(actions, tmpActions...)
		}
	}
	return
}

func getCreateDBMetricMonitorByImport(inputObj *models.DbMetricMonitorObj, nowTime, operator string) (actions []*Action, affectHost []string, affectEndpointGroup []string, err error) {
	actions = getCreateDBMetricActions(inputObj, operator, nowTime)
	return
}

func getCreateLogMetricGroupByImport(metricGroup *models.LogMetricGroupObj, operator string, existMetricMap map[string]string, errMsgObj *models.ErrorTemplate, roles []string) (actions []*Action, newDashboardId int64, err error) {
	var tmpActions []*Action
	if metricGroup.LogMonitorTemplate != "" && (metricGroup.LogType == models.LogMonitorRegularType || metricGroup.LogType == models.LogMonitorJsonType) {
		metricGroup.LogMonitorTemplate, err = GetLogTemplateGuidByName(metricGroup.LogMonitorTemplateName)
		if err != nil {
			return
		}
		tmpCreateParam := models.LogMetricGroupWithTemplate{
			LogMetricGroupGuid:     metricGroup.Guid,
			Name:                   metricGroup.Name,
			LogMetricMonitorGuid:   metricGroup.LogMetricMonitor,
			LogMonitorTemplateGuid: metricGroup.LogMonitorTemplate,
			MetricPrefixCode:       metricGroup.MetricPrefixCode,
			ServiceGroup:           metricGroup.ServiceGroup,
			MonitorType:            metricGroup.MonitorType,
			LogMonitorTemplate:     metricGroup.LogMonitorTemplateDto,
			AutoCreateWarn:         metricGroup.AutoCreateWarn,
			AutoCreateDashboard:    metricGroup.AutoCreateDashboard,
		}
		for _, mgParamObj := range metricGroup.ParamList {
			if mgParamObj.Name == "code" {
				tmpCreateParam.CodeStringMap = mgParamObj.StringMap
			} else if mgParamObj.Name == "retcode" {
				tmpCreateParam.RetCodeStringMap = mgParamObj.StringMap
			}
		}
		if tmpActions, _, newDashboardId, err = getCreateLogMetricGroupActions(&tmpCreateParam, operator, roles, existMetricMap, errMsgObj, true); err != nil {
			return
		}
		actions = append(actions, tmpActions...)
	} else {
		if tmpActions, _, newDashboardId, err = getCreateLogMetricCustomGroupActions(metricGroup, operator, existMetricMap, roles, errMsgObj, true); err != nil {
			return
		}
		actions = append(actions, tmpActions...)
	}
	return
}

func getUpdateLogMetricGroupByImport(metricGroup *models.LogMetricGroupObj, operator string, roles []string) (actions []*Action, affectEndpointGroups []string, err error) {
	if metricGroup.LogMonitorTemplate != "" {
		tmpCreateParam := models.LogMetricGroupWithTemplate{
			LogMetricGroupGuid:     metricGroup.Guid,
			Name:                   metricGroup.Name,
			LogMetricMonitorGuid:   metricGroup.LogMetricMonitor,
			LogMonitorTemplateGuid: metricGroup.LogMonitorTemplate,
			MetricPrefixCode:       metricGroup.MetricPrefixCode,
			AutoCreateDashboard:    metricGroup.AutoCreateDashboard,
			AutoCreateWarn:         metricGroup.AutoCreateWarn,
		}
		for _, mgParamObj := range metricGroup.ParamList {
			if mgParamObj.Name == "code" {
				tmpCreateParam.CodeStringMap = mgParamObj.StringMap
			} else if mgParamObj.Name == "retcode" {
				tmpCreateParam.RetCodeStringMap = mgParamObj.StringMap
			}
		}
		// directly perform update with dashboard sync
		if err = UpdateLogMetricGroupWithDashboardAndAlarm(&tmpCreateParam, operator, roles); err != nil {
			return
		}
		// no direct actions because we executed update; return empty actions
		return []*Action{}, []string{}, nil
	} else {
		actions, affectEndpointGroups, err = getUpdateLogMetricCustomGroupActions(metricGroup, operator)
	}
	return
}

func getCompareLogMetricConfigByImport(inputLogMetricList, existLogMetricList []*models.LogMetricConfigObj, nowTime, operator string) (actions []*Action, affectEndpointGroup []string) {
	for _, inputLogMetricObj := range inputLogMetricList {
		matchExistMetricObj := &models.LogMetricConfigObj{}
		for _, existLogMetricObj := range existLogMetricList {
			if inputLogMetricObj.Guid == existLogMetricObj.Guid {
				matchExistMetricObj = existLogMetricObj
				break
			}
		}
		if matchExistMetricObj.Guid != "" {
			tmpActions, tmpAffectEndpointGroup := getUpdateLogMetricConfigByImport(inputLogMetricObj, matchExistMetricObj, nowTime, operator)
			actions = append(actions, tmpActions...)
			affectEndpointGroup = append(affectEndpointGroup, tmpAffectEndpointGroup...)
		} else {
			tmpActions := getCreateLogMetricConfigAction(inputLogMetricObj, nowTime, operator)
			actions = append(actions, tmpActions...)
		}
	}
	for _, existLogMetricObj := range existLogMetricList {
		deleteFlag := true
		for _, inputLogMetricObj := range inputLogMetricList {
			if inputLogMetricObj.Guid == existLogMetricObj.Guid {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			deleteActions, tmpEndpointGroup := getDeleteLogMetricConfigByImport(existLogMetricObj)
			actions = append(actions, deleteActions...)
			affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
		}
	}
	return
}

func getUpdateLogMetricConfigByImport(inputLogMetric, existLogMetric *models.LogMetricConfigObj, nowTime, operator string) (actions []*Action, affectEndpointGroup []string) {
	if existLogMetric.Metric != inputLogMetric.Metric || existLogMetric.AggType != inputLogMetric.AggType {
		oldMetricGuid := fmt.Sprintf("%s__%s", existLogMetric.Metric, inputLogMetric.ServiceGroup)
		newMetricGuid := fmt.Sprintf("%s__%s", inputLogMetric.Metric, inputLogMetric.ServiceGroup)
		actions = append(actions, &Action{Sql: "update metric set guid=?,metric=?,prom_expr=?,update_user=?,update_time=? where guid=?",
			Param: []interface{}{newMetricGuid, inputLogMetric.Metric, getLogMetricExprByAggType(inputLogMetric.Metric, inputLogMetric.AggType, inputLogMetric.ServiceGroup, []string{}), operator, nowTime, oldMetricGuid}})
		var alarmStrategyTable []*models.AlarmStrategyTable
		x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", oldMetricGuid).Find(&alarmStrategyTable)
		if len(alarmStrategyTable) > 0 {
			for _, v := range alarmStrategyTable {
				affectEndpointGroup = append(affectEndpointGroup, v.EndpointGroup)
			}
			actions = append(actions, &Action{Sql: "update alarm_strategy set metric=? where metric=?", Param: []interface{}{newMetricGuid, oldMetricGuid}})
		}
	}
	actions = append(actions, &Action{Sql: "update log_metric_config set metric=?,display_name=?,json_key=?,regular=?,agg_type=?,step=?,update_time=? where guid=?", Param: []interface{}{inputLogMetric.Metric, inputLogMetric.DisplayName, inputLogMetric.JsonKey, inputLogMetric.Regular, inputLogMetric.AggType, inputLogMetric.Step, nowTime, inputLogMetric.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config=?", Param: []interface{}{inputLogMetric.Guid}})
	guidList := guid.CreateGuidList(len(inputLogMetric.StringMap))
	for i, v := range inputLogMetric.StringMap {
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_config,source_value,regulative,target_value,update_time) value (?,?,?,?,?,?)", Param: []interface{}{guidList[i], inputLogMetric.Guid, v.SourceValue, v.Regulative, v.TargetValue, nowTime}})
	}
	return
}

func getDeleteLogMetricConfigByImport(existLogMetric *models.LogMetricConfigObj) (actions []*Action, endpointGroup []string) {
	alarmMetricGuid := fmt.Sprintf("%s__%s", existLogMetric.Metric, existLogMetric.ServiceGroup)
	var alarmStrategyTable []*models.AlarmStrategyTable
	x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", alarmMetricGuid).Find(&alarmStrategyTable)
	for _, v := range alarmStrategyTable {
		endpointGroup = append(endpointGroup, v.EndpointGroup)
	}
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where metric=?", Param: []interface{}{alarmMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{alarmMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_config=?", Param: []interface{}{existLogMetric.Guid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_config where guid=?", Param: []interface{}{existLogMetric.Guid}})
	return
}

func ImportLogMetricExcel(logMonitorGuid, operator string, param []*models.LogMetricConfigObj) (err error) {
	var actions []*Action
	var affectEndpointGroupList, affectHostList []string
	for _, v := range ListLogMetricEndpointRel(logMonitorGuid) {
		affectHostList = append(affectHostList, v.SourceEndpoint)
	}
	for _, existLogConfig := range ListLogMetricConfig("", logMonitorGuid) {
		tmpActions, tmpAffectEndpointGroup := getDeleteLogMetricConfigAction(existLogConfig.Guid, logMonitorGuid)
		actions = append(actions, tmpActions...)
		affectEndpointGroupList = append(affectEndpointGroupList, tmpAffectEndpointGroup...)
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	for _, inputLogConfig := range param {
		inputLogConfig.LogMetricMonitor = logMonitorGuid
		actions = append(actions, getCreateLogMetricConfigAction(inputLogConfig, nowTime, operator)...)
	}
	err = Transaction(actions)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "import log metric from excel exec database fail", zap.Error(err))
		return
	}
	if tmpErr := SyncLogMetricExporterConfig(affectHostList); tmpErr != nil {
		log.Error(nil, log.LOGGER_APP, "sync log metric to affect host fail", zap.Error(tmpErr))
	}
	for _, v := range affectEndpointGroupList {
		if tmpErr := SyncPrometheusRuleFile(v, false); tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "sync prometheus rule file fail", zap.Error(tmpErr))
		}
	}
	return
}

func GetSimpleLogMetricGroup(logMetricGroupGuid string) (result *models.LogMetricGroup, err error) {
	var logMetricGroupRows []*models.LogMetricGroup
	err = x.SQL("select * from log_metric_group where guid=?", logMetricGroupGuid).Find(&logMetricGroupRows)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_group fail,%s ", err.Error())
	}
	if len(logMetricGroupRows) == 0 {
		return result, fmt.Errorf("Can not find log_metric_group with guid:%s ", logMetricGroupGuid)
	}
	result = logMetricGroupRows[0]
	return
}

func GetLogMetricGroup(logMetricGroupGuid string) (result *models.LogMetricGroupWithTemplate, err error) {
	var logMonitorTemplate = &models.LogMonitorTemplateDto{}
	metricGroupObj, getGroupErr := GetSimpleLogMetricGroup(logMetricGroupGuid)
	if getGroupErr != nil {
		err = getGroupErr
		return
	}
	if metricGroupObj.TemplateSnapshot != "" {
		if err = json.Unmarshal([]byte(metricGroupObj.TemplateSnapshot), logMonitorTemplate); err != nil {
			return
		}
	} else {
		// 历史数据 模版查询兜底
		if logMonitorTemplate, err = GetLogMonitorTemplate(metricGroupObj.LogMonitorTemplate); err != nil {
			return
		}
		metricGroupObj.RefTemplateVersion = logMonitorTemplate.UpdateTime.Format(models.DatetimeDigitFormat)
	}
	var logMetricStringMapRows []*models.LogMetricStringMapTable
	err = x.SQL("select * from log_metric_string_map where log_metric_group=?", logMetricGroupGuid).Find(&logMetricStringMapRows)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_string_map fail,%s ", err.Error())
	}
	result = &models.LogMetricGroupWithTemplate{Name: metricGroupObj.Name, MetricPrefixCode: metricGroupObj.MetricPrefixCode,
		LogMetricGroupGuid: logMetricGroupGuid, LogMetricMonitorGuid: metricGroupObj.LogMetricMonitor,
		LogMonitorTemplateGuid: metricGroupObj.LogMonitorTemplate, CodeStringMap: []*models.LogMetricStringMapTable{},
		LogMonitorTemplateVersion: metricGroupObj.RefTemplateVersion,
		LogMonitorTemplate:        logMonitorTemplate,
		AutoCreateDashboard:       metricGroupObj.AutoDashboard == 1,
		AutoCreateWarn:            metricGroupObj.AutoAlarm == 1,
		RetCodeStringMap:          []*models.LogMetricStringMapTable{},
		Status:                    metricGroupObj.Status,
	}
	for _, row := range logMetricStringMapRows {
		if row.LogParamName == "code" {
			result.CodeStringMap = append(result.CodeStringMap, row)
		} else if row.LogParamName == "retcode" {
			result.RetCodeStringMap = append(result.RetCodeStringMap, row)
		}
	}
	return
}

func CreateLogMetricGroup(param *models.LogMetricGroupWithTemplate, operator string, roles []string, errMsgObj *models.ErrorTemplate) (result *models.CreateLogMetricGroupDto, err error) {
	param.LogMetricGroupGuid = ""
	var actions []*Action
	var newDashboardId int64
	actions, result, newDashboardId, err = getCreateLogMetricGroupActions(param, operator, roles, make(map[string]string), errMsgObj, false)
	if err != nil {
		deleteCustomDashboard(newDashboardId)
		return
	}
	// 看板名称重复校验(业务配置导入时候不需要校验,导入重复数据会先删除)
	if param.AutoCreateDashboard && result != nil {
		var customDashboardList []*models.CustomDashboardTable
		customDashboardName := result.CustomDashboard
		// 查询看板 名称是否已存在,去掉看板名称校验
		if customDashboardList, err = QueryCustomDashboardListByName(customDashboardName); err != nil {
			return
		}
		// 新增看板已经入库,这里数量大于1表示有重复名称
		if len(customDashboardList) > 1 {
			err = fmt.Errorf(errMsgObj.ImportDashboardNameExistError.Error(), customDashboardName)
			return
		}
	}
	if err = Transaction(actions); err != nil {
		deleteCustomDashboard(newDashboardId)
	}
	return
}

func getCreateLogMetricGroupActions(param *models.LogMetricGroupWithTemplate, operator string, roles []string, existMetricMap map[string]string, errMsgObj *models.ErrorTemplate, doImport bool) (actions []*Action, result *models.CreateLogMetricGroupDto, newDashboardId int64, err error) {
	var templateSnapshot []byte
	var refTemplateVersion, endpointGroup string
	var subCreateAlarmStrategyActions, subCreateDashboardActions []*Action
	var serviceGroupsRoles, alarmStrategyList []string
	var autoAlarm, autoDashboard int
	if param.LogMetricGroupGuid == "" {
		param.LogMetricGroupGuid = "lmg_" + guid.CreateGuid()
	}
	if param.AutoCreateWarn {
		autoAlarm = 1
	}
	if param.AutoCreateDashboard {
		autoDashboard = 1
	}
	if param.Status == "" {
		param.Status = "enable"
	}
	result = &models.CreateLogMetricGroupDto{AlarmList: make([]string, 0)}
	logMonitorTemplateObj, getErr := GetLogMonitorTemplate(param.LogMonitorTemplateGuid)
	if getErr != nil {
		err = getErr
		return
	}
	if logMonitorTemplateObj == nil {
		// 业务配置导入会带业务日志模版,如果模版未提前创建好,需要自动支持先导入模版
		if param.LogMonitorTemplate != nil {
			if tmpActions := getCreateLogMonitorTemplateActions(param.LogMonitorTemplate, operator); len(tmpActions) > 0 {
				actions = append(actions, tmpActions...)
			}
		} else {
			err = fmt.Errorf("LogMonitorTemplateGuid is valid")
			return
		}
	}
	refTemplateVersion = logMonitorTemplateObj.UpdateTime.Format(models.DatetimeDigitFormat)
	if templateSnapshot, err = json.Marshal(logMonitorTemplateObj); err != nil {
		return
	}

	nowTime := time.Now()
	if param.Name == "" {
		param.Name = logMonitorTemplateObj.Name
	}
	// 如果输入的映射为空，尝试拿模版的成功映射，为了插件服务自动化生成配置
	if len(param.RetCodeStringMap) == 0 && logMonitorTemplateObj.SuccessCode != "" {
		var templateRetCodeMap models.LogMetricStringMapTable
		if unmarshalErr := json.Unmarshal([]byte(logMonitorTemplateObj.SuccessCode), &templateRetCodeMap); unmarshalErr == nil {
			templateRetCodeMap.ValueType = "success"
			param.RetCodeStringMap = append(param.RetCodeStringMap, &templateRetCodeMap)
		} else {
			log.Warn(nil, log.LOGGER_APP, "json unmarshal log template success code fail", zap.String("successCode", logMonitorTemplateObj.SuccessCode), zap.Error(unmarshalErr))
		}
	}
	actions = append(actions, &Action{Sql: "insert into log_metric_group(guid,name,metric_prefix_code,log_type,log_metric_monitor,log_monitor_template,create_user," +
		"create_time,update_user,update_time,template_snapshot,ref_template_version,auto_alarm,auto_dashboard,status) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		param.LogMetricGroupGuid, param.Name, param.MetricPrefixCode, logMonitorTemplateObj.LogType, param.LogMetricMonitorGuid, param.LogMonitorTemplateGuid, operator,
		nowTime, operator, nowTime, templateSnapshot, refTemplateVersion, autoAlarm, autoDashboard, param.Status,
	}})
	sucRetCode, createMapActions := getCreateLogMetricGroupMapAction(param, nowTime)
	actions = append(actions, createMapActions...)
	// 自动添加增加 metric
	serviceGroup, monitorType := param.ServiceGroup, param.MonitorType
	if serviceGroup == "" {
		serviceGroup, monitorType = GetLogMetricServiceGroup(param.LogMetricMonitorGuid)
		param.ServiceGroup = serviceGroup
		param.MonitorType = monitorType
	}
	for _, v := range logMonitorTemplateObj.MetricList {
		promExpr := ""
		tmpMetricWithPrefix := v.Metric
		if param.MetricPrefixCode != "" {
			tmpMetricWithPrefix = param.MetricPrefixCode + "_" + v.Metric
		}
		if v.Metric == "req_suc_count" || v.Metric == "req_fail_count" || v.Metric == "req_fail_count_detail" || v.Metric == "req_suc_rate" || v.Metric == "req_fail_rate" {
			promExpr = getLogMetricRatePromExpr(v.Metric, param.MetricPrefixCode, v.AggType, serviceGroup, sucRetCode)
		} else {
			promExpr = getLogMetricExprByAggType(tmpMetricWithPrefix, v.AggType, serviceGroup, v.TagConfigList)
		}
		tmpMetricGuid := generateMetricGuid(tmpMetricWithPrefix, serviceGroup)
		if duplicateMetric, ok := existMetricMap[tmpMetricGuid]; ok && !doImport {
			err = fmt.Errorf("Metric: %s duplicate ", duplicateMetric)
			return
		}
		actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,log_metric_template,log_metric_group,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?,?,?)",
			Param: []interface{}{tmpMetricGuid, tmpMetricWithPrefix, monitorType, promExpr, serviceGroup, models.MetricWorkspaceService, nowTime, v.Guid, param.LogMetricGroupGuid, nowTime, operator, operator}})
	}
	if serviceGroup != "" && monitorType != "" {
		var endpointGroupIds []string
		serviceGroupsRoles = getServiceGroupRoles(serviceGroup)
		if err = x.SQL("select guid from endpoint_group where service_group=? and monitor_type=?", serviceGroup, monitorType).Find(&endpointGroupIds); err != nil {
			return
		}
		if len(endpointGroupIds) > 0 {
			endpointGroup = endpointGroupIds[0]
		}
	}
	if len(serviceGroupsRoles) == 0 && len(roles) > 0 {
		serviceGroupsRoles = roles[:1]
	}
	alarmStrategyParam := models.AutoAlarmStrategyParam{LogMetricGroupWithTemplate: param, MetricList: logMonitorTemplateObj.MetricList, ServiceGroupsRoles: serviceGroupsRoles,
		ServiceGroup: serviceGroup, EndpointGroup: endpointGroup, Operator: operator, ErrMsgObj: errMsgObj}
	if subCreateAlarmStrategyActions, alarmStrategyList, err = autoGenerateAlarmStrategy(alarmStrategyParam); err != nil {
		return
	}
	if len(subCreateAlarmStrategyActions) > 0 {
		actions = append(actions, subCreateAlarmStrategyActions...)
		result.AlarmList = alarmStrategyList
	}
	var dashboardParam = models.AutoCreateDashboardParam{LogMetricGroupWithTemplate: param, MetricList: logMonitorTemplateObj.MetricList, ServiceGroupsRoles: serviceGroupsRoles,
		ServiceGroup: serviceGroup, Operator: operator, ErrMsgObj: errMsgObj}
	if subCreateDashboardActions, result.CustomDashboard, newDashboardId, err = autoGenerateCustomDashboard(dashboardParam); err != nil {
		return
	}
	if len(subCreateDashboardActions) > 0 {
		actions = append(actions, subCreateDashboardActions...)
	}
	return
}

func getCreateLogMetricGroupMapAction(param *models.LogMetricGroupWithTemplate, nowTime time.Time) (sucRetCode string, actions []*Action) {
	codeGuidList := guid.CreateGuidList(len(param.CodeStringMap))
	for i, v := range param.CodeStringMap {
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_group,log_param_name,value_type,source_value,regulative,target_value,update_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lmsm_" + codeGuidList[i], param.LogMetricGroupGuid, "code", v.ValueType, v.SourceValue, v.Regulative, v.TargetValue, nowTime.Format(models.DatetimeFormat),
		}})
	}
	retCodeGuidList := guid.CreateGuidList(len(param.RetCodeStringMap))
	for i, v := range param.RetCodeStringMap {
		if v.ValueType == "success" {
			sucRetCode = v.TargetValue
		}
		actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_group,log_param_name,value_type,source_value,regulative,target_value,update_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lmsm_" + retCodeGuidList[i], param.LogMetricGroupGuid, "retcode", v.ValueType, v.SourceValue, v.Regulative, v.TargetValue, nowTime.Format(models.DatetimeFormat),
		}})
	}
	return
}

func UpdateLogMetricGroup(param *models.LogMetricGroupWithTemplate, operator string) (err error) {
	var actions []*Action
	actions, err = getUpdateLogMetricGroupActions(param, operator)
	if err != nil {
		return
	}
	err = Transaction(actions)
	return
}

func UpdateLogMetricGroupStatus(param models.UpdateLogMetricGroupStatus, operator string) (err error) {
	_, err = x.Exec("update log_metric_group set update_user=?,update_time=?,status=? where guid=?", operator, time.Now(), param.Status, param.LogMetricGroupGuid)
	return
}

func BatchQueryDisabledLogMetricGroupStatus(ids []string) (disableIdsMap map[string]bool, err error) {
	var list []string
	if len(ids) == 0 {
		err = errors.New("no IDs provided")
		return
	}
	disableIdsMap = make(map[string]bool)
	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}
	sql := fmt.Sprintf("select guid from  log_metric_group  WHERE guid IN (%s) and status = 'disabled'", strings.Join(placeholders, ", "))
	// 创建一个足够大的 args 切片
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	if err = x.SQL(sql, args...).Find(&list); err != nil {
		return
	}
	for _, guid := range list {
		disableIdsMap[guid] = true
	}
	return
}

func BatchDisableLogMetricGroupStatus(ids []string) (err error) {
	if len(ids) == 0 {
		return errors.New("no IDs provided")
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}
	sql := fmt.Sprintf("UPDATE log_metric_group SET update_time=?, status=? WHERE guid IN (%s)", strings.Join(placeholders, ", "))
	// 创建一个足够大的 args 切片
	args := make([]interface{}, len(ids)+3)
	args[0] = sql
	args[1] = time.Now()
	args[2] = "disabled"
	for i, id := range ids {
		args[i+3] = id
	}
	_, err = x.Exec(args...)
	return
}

func getUpdateLogMetricGroupActions(param *models.LogMetricGroupWithTemplate, operator string) (actions []*Action, err error) {
	nowTime := time.Now()
	actions = append(actions, &Action{Sql: "update log_metric_group set name=?,update_user=?,update_time=?,status=? where guid=?", Param: []interface{}{
		param.Name, operator, nowTime, param.Status, param.LogMetricGroupGuid,
	}})
	var logMetricStringMapRows []*models.LogMetricStringMapTable
	err = x.SQL("select source_value,target_value,value_type from log_metric_string_map where log_metric_group=?", param.LogMetricGroupGuid).Find(&logMetricStringMapRows)
	if err != nil {
		err = fmt.Errorf("Query table log_metric_string_map fail,%s ", err.Error())
		return
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_group=?", Param: []interface{}{param.LogMetricGroupGuid}})
	newSucRetCode, updateMapActions := getCreateLogMetricGroupMapAction(param, nowTime)
	var oldSucRetCode string
	for _, v := range logMetricStringMapRows {
		if v.ValueType == "success" {
			oldSucRetCode = v.TargetValue
			break
		}
	}
	if newSucRetCode != oldSucRetCode {
		// update metric
		logMetricGroupObj, getErr := GetSimpleLogMetricGroup(param.LogMetricGroupGuid)
		if getErr != nil {
			err = getErr
			return
		}
		logMonitorTemplateObj, getTemplateErr := GetLogMonitorTemplate(param.LogMonitorTemplateGuid)
		if getTemplateErr != nil {
			err = getTemplateErr
			return
		}
		serviceGroup, _ := GetLogMetricServiceGroup(logMetricGroupObj.LogMetricMonitor)
		for _, v := range logMonitorTemplateObj.MetricList {
			tmpMetricWithPrefix := v.Metric
			if logMetricGroupObj.MetricPrefixCode != "" {
				tmpMetricWithPrefix = logMetricGroupObj.MetricPrefixCode + "_" + v.Metric
			}
			promExpr := ""
			if v.Metric == "req_suc_count" || v.Metric == "req_fail_count" || v.Metric == "req_fail_count_detail" || v.Metric == "req_suc_rate" || v.Metric == "req_fail_rate" {
				promExpr = getLogMetricRatePromExpr(v.Metric, logMetricGroupObj.MetricPrefixCode, v.AggType, serviceGroup, newSucRetCode)
			} else {
				promExpr = getLogMetricExprByAggType(tmpMetricWithPrefix, v.AggType, serviceGroup, v.TagConfigList)
			}
			actions = append(actions, &Action{Sql: "update metric set prom_expr=?,update_time=?,update_user=? where guid=?", Param: []interface{}{promExpr, nowTime, operator, fmt.Sprintf("%s__%s", tmpMetricWithPrefix, serviceGroup)}})
		}
	}
	actions = append(actions, updateMapActions...)
	// Ensure 'other' panel group is always placed at the end for all related dashboards
	if reorderActions, reorderErr := ensureOtherPanelGroupLastActions(param.LogMetricGroupGuid, operator); reorderErr == nil {
		actions = append(actions, reorderActions...)
	} else {
		// Ignore reorder error, do not block main update path
		log.Warn(nil, log.LOGGER_APP, "Skip ensure other panel group last due to error", zap.Error(reorderErr))
	}
	return
}

// ensureOtherPanelGroupLastActions builds update actions to force 'other' to be at the end of panel_groups
func ensureOtherPanelGroupLastActions(logMetricGroupGuid, operator string) (actions []*Action, err error) {
	const otherCode = "other"
	type boardRow struct {
		Id          int    `json:"id"`
		PanelGroups string `json:"panel_groups"`
	}
	var boards []boardRow
	if err = x.SQL("select id, panel_groups from custom_dashboard where log_metric_group=?", logMetricGroupGuid).Find(&boards); err != nil {
		return
	}
	now := time.Now().Format(models.DatetimeFormat)
	for _, b := range boards {
		pg := strings.TrimSpace(b.PanelGroups)
		if pg == "" {
			continue
		}
		var reordered []string
		hasOther := false
		for _, g := range strings.Split(pg, ",") {
			g = strings.TrimSpace(g)
			if g == "" {
				continue
			}
			if strings.EqualFold(g, otherCode) {
				hasOther = true
				continue
			}
			reordered = append(reordered, g)
		}
		if hasOther {
			reordered = append(reordered, otherCode)
		}
		newPG := strings.Join(reordered, ",")
		if newPG != pg {
			actions = append(actions, &Action{Sql: "update custom_dashboard set panel_groups=?, update_user=?, update_at=? where id= ?", Param: []interface{}{newPG, operator, now, b.Id}})
		}
	}
	return
}

func DeleteLogMetricGroup(logMetricGroupGuid string) (logMetricMonitorGuid string, err error) {
	var actions []*Action
	var affectEndpointGroup []string
	actions, affectEndpointGroup, logMetricMonitorGuid, err = getDeleteLogMetricGroupActions(logMetricGroupGuid)
	err = Transaction(actions)
	if err == nil && len(affectEndpointGroup) > 0 {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return
}

func getDeleteLogMetricGroupActions(logMetricGroupGuid string) (actions []*Action, affectEndpointGroup []string, logMetricMonitorGuid string, err error) {
	var delAlarmStrategyActions, delDashboardActions, delChartActions []*Action
	metricGroupObj, getGroupErr := GetSimpleLogMetricGroup(logMetricGroupGuid)
	if getGroupErr != nil {
		err = getGroupErr
		return
	}
	logMetricMonitorGuid = metricGroupObj.LogMetricMonitor
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_group=?", Param: []interface{}{logMetricGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_param where log_metric_group=?", Param: []interface{}{logMetricGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from log_metric_config where log_metric_group=?", Param: []interface{}{logMetricGroupGuid}})

	// 删除阈值
	var strategyGuids []string
	if err = x.SQL("select guid from alarm_strategy where log_metric_group=?", logMetricGroupGuid).Find(&strategyGuids); err != nil {
		return
	}
	if len(strategyGuids) > 0 {
		for _, strategyGuid := range strategyGuids {
			if delAlarmStrategyActions, _, err = GetDeleteAlarmStrategyActions(strategyGuid); err != nil {
				return
			}
			if len(delAlarmStrategyActions) > 0 {
				actions = append(actions, delAlarmStrategyActions...)
			}
		}
	}
	// 删除看板&图表
	var dashboardIds []int
	var chartIds []string
	if err = x.SQL("select id from custom_dashboard where log_metric_group=?", logMetricGroupGuid).Find(&dashboardIds); err != nil {
		return
	}
	if err = x.SQL("select guid from custom_chart where log_metric_group=?", logMetricGroupGuid).Find(&chartIds); err != nil {
		return
	}
	if len(dashboardIds) > 0 {
		for _, boardId := range dashboardIds {
			if delDashboardActions = GetDeleteCustomDashboardByIdActions(boardId); len(delDashboardActions) > 0 {
				actions = append(actions, delDashboardActions...)
			}
		}
	}
	if len(chartIds) > 0 {
		for _, chartId := range chartIds {
			if delChartActions, err = GetDeleteCustomDashboardChart(chartId); err != nil {
				return
			}
			if len(delChartActions) > 0 {
				actions = append(actions, delChartActions...)
			}
		}
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_group where guid=?", Param: []interface{}{logMetricGroupGuid}})

	// 查找关联的指标并删除
	serviceGroup, _ := GetLogMetricServiceGroup(metricGroupObj.LogMetricMonitor)
	var existMetricRows []*models.LogMetricConfigDto
	if metricGroupObj.LogMonitorTemplate != "" {
		logMonitorTemplateObj, getTemplateErr := GetLogMonitorTemplate(metricGroupObj.LogMonitorTemplate)
		if getTemplateErr != nil {
			err = getTemplateErr
			return
		}
		for _, v := range logMonitorTemplateObj.MetricList {
			existMetricRows = append(existMetricRows, v.TransToLogMetric())
		}
	} else {
		logMetricGroupObj, getLogGroupErr := GetLogMetricCustomGroup(logMetricGroupGuid)
		if getLogGroupErr != nil {
			err = getLogGroupErr
			return
		}
		existMetricRows = logMetricGroupObj.MetricList
	}
	for _, existMetric := range existMetricRows {
		if metricGroupObj.MetricPrefixCode != "" {
			existMetric.Metric = metricGroupObj.MetricPrefixCode + "_" + existMetric.Metric
		}
		deleteMetricActions, endpointGroups := getDeleteLogMetricActions(existMetric.Metric, serviceGroup)
		actions = append(actions, deleteMetricActions...)
		affectEndpointGroup = append(affectEndpointGroup, endpointGroups...)
	}
	return
}

func getDeleteLogMetricActions(metric, serviceGroup string) (actions []*Action, affectEndpointGroup []string) {
	alarmMetricGuid := fmt.Sprintf("%s__%s", metric, serviceGroup)
	tmpActions, tmpEndpointGroups := getMetricComparisonDeleteAction(alarmMetricGuid)
	actions = append(actions, tmpActions...)
	affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroups...)
	tmpActions, tmpEndpointGroups = getDeleteMetricActions(alarmMetricGuid)
	actions = append(actions, tmpActions...)
	affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroups...)
	return
}

func getDeleteMetricActions(metricGuid string) (actions []*Action, affectEndpointGroup []string) {
	var alarmStrategyTable []*models.AlarmStrategyTable
	x.SQL("select t1.guid,t1.endpoint_group,t2.guid as `condition` from alarm_strategy t1 left join alarm_strategy_metric t2 on t1.guid=t2.alarm_strategy where (t1.metric=? or t2.metric=?)", metricGuid, metricGuid).Find(&alarmStrategyTable)
	for _, v := range alarmStrategyTable {
		affectEndpointGroup = append(affectEndpointGroup, v.EndpointGroup)
	}
	for _, row := range alarmStrategyTable {
		if row.Condition == "" {
			actions = append(actions, &Action{Sql: "delete from alarm_strategy where guid=?", Param: []interface{}{row.Guid}})
		} else {
			actions = append(actions, &Action{Sql: "delete from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric=?)", Param: []interface{}{row.Condition}})
			actions = append(actions, &Action{Sql: "delete from alarm_strategy_tag where alarm_strategy_metric=?", Param: []interface{}{row.Condition}})
			actions = append(actions, &Action{Sql: "delete from alarm_strategy_metric where guid=?", Param: []interface{}{row.Condition}})
		}
	}
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_tagvalue where dashboard_chart_tag in (select guid from custom_chart_series_tag where dashboard_chart_config in (select guid from custom_chart_series where metric_guid=?))", Param: []interface{}{metricGuid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag where dashboard_chart_config in (select guid from custom_chart_series where metric_guid=?)", Param: []interface{}{metricGuid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_config where dashboard_chart_config in (select guid from custom_chart_series where metric_guid=?)", Param: []interface{}{metricGuid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series where metric_guid=?", Param: []interface{}{metricGuid}})
	actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{metricGuid}})
	return
}

func ListLogMetricGroups(logMetricMonitor, metricKey string) (result []*models.LogMetricGroupObj) {
	result = []*models.LogMetricGroupObj{}
	var logMetricGroupTable []*models.LogMetricGroup
	x.SQL("select * from log_metric_group where log_metric_monitor=? order by update_time desc", logMetricMonitor).Find(&logMetricGroupTable)
	for _, v := range logMetricGroupTable {
		v.CreateTimeString = v.CreateTime.Format(models.DatetimeFormat)
		v.UpdateTimeString = v.UpdateTime.Format(models.DatetimeFormat)
		logMetricGroupData := &models.LogMetricGroupObj{LogMetricGroup: *v, AutoCreateDashboard: v.AutoDashboard == 1, AutoCreateWarn: v.AutoAlarm == 1}
		if v.LogMonitorTemplate != "" && v.LogType != "custom" {
			tmpTemplateObj, tmpGetTemplateErr := GetLogMonitorTemplate(v.LogMonitorTemplate)
			if tmpGetTemplateErr != nil {
				log.Error(nil, log.LOGGER_APP, "ListLogMetricGroups fail get template data ", zap.String("templateGuid", v.LogMonitorTemplate), zap.Error(tmpGetTemplateErr))
			} else {
				logMetricGroupData.JsonRegular = tmpTemplateObj.JsonRegular
				logMetricStringMapData, getStringMapErr := getLogMetricGroupMapData(v.Guid)
				if getStringMapErr != nil {
					log.Error(nil, log.LOGGER_APP, "ListLogMetricGroups getLogMetricGroupMapData fail ", zap.String("logMetricGroupGuid", v.Guid), zap.Error(getStringMapErr))
				}
				logMetricGroupData.LogMonitorTemplateName = tmpTemplateObj.Name
				// 读取 模版的 ParamList
				for _, tplParam := range tmpTemplateObj.ParamList {
					tmpLogMetricParamObj := tplParam.TransToLogParam()
					tmpLogMetricParamObj.StringMap = logMetricStringMapData[tmpLogMetricParamObj.Name]
					logMetricGroupData.ParamList = append(logMetricGroupData.ParamList, tmpLogMetricParamObj)
				}
				for _, tplMetric := range tmpTemplateObj.MetricList {
					output := tplMetric.TransToLogMetric()
					if strings.TrimSpace(v.MetricPrefixCode) == "" {
						output.FullMetric = output.Metric
					} else {
						output.FullMetric = fmt.Sprintf("%s_%s", v.MetricPrefixCode, output.Metric)
					}
					logMetricGroupData.MetricList = append(logMetricGroupData.MetricList, output)
				}
			}
		} else {
			customGroupData, getCustomErr := GetLogMetricCustomGroup(v.Guid)
			if getCustomErr != nil {
				log.Error(nil, log.LOGGER_APP, "ListLogMetricGroups fail get custom metric group data ", zap.String("logMetricGroupGuid", v.Guid), zap.Error(getCustomErr))
			} else {
				logMetricGroupData = customGroupData
				logMetricGroupData.AutoCreateDashboard = v.AutoDashboard == 1
				logMetricGroupData.AutoCreateWarn = v.AutoAlarm == 1
			}
			if v.LogMonitorTemplate != "" {
				tmpTemplateObj, tmpGetTemplateErr := GetLogMonitorTemplate(v.LogMonitorTemplate)
				if tmpGetTemplateErr != nil {
					log.Error(nil, log.LOGGER_APP, "ListLogMetricGroups fail get template data ", zap.String("templateGuid", v.LogMonitorTemplate), zap.Error(tmpGetTemplateErr))
				} else {
					logMetricGroupData.LogMonitorTemplateName = tmpTemplateObj.Name
					logMetricGroupData.LogMonitorTemplateGuid = v.LogMonitorTemplate
				}
			}
		}
		if metricKey != "" && len(logMetricGroupData.MetricList) > 0 {
			for _, metric := range logMetricGroupData.MetricList {
				if strings.Contains(metric.FullMetric, metricKey) {
					result = append(result, logMetricGroupData)
					break
				}
			}
		} else {
			result = append(result, logMetricGroupData)
		}
	}
	return result
}

func GetLogMetricCustomGroup(logMetricGroupGuid string) (result *models.LogMetricGroupObj, err error) {
	metricGroupObj, getGroupErr := GetSimpleLogMetricGroup(logMetricGroupGuid)
	if getGroupErr != nil {
		err = getGroupErr
		return
	}
	result = &models.LogMetricGroupObj{LogMetricGroup: *metricGroupObj, ParamList: []*models.LogMetricParamObj{}, MetricList: []*models.LogMetricConfigDto{}}
	result.CreateTimeString = result.CreateTime.Format(models.DatetimeFormat)
	result.UpdateTimeString = result.UpdateTime.Format(models.DatetimeFormat)
	result.AutoAlarm = metricGroupObj.AutoAlarm
	result.AutoDashboard = metricGroupObj.AutoDashboard
	result.AutoCreateWarn = metricGroupObj.AutoAlarm == 1
	result.AutoCreateDashboard = metricGroupObj.AutoDashboard == 1
	logMetricStringMapData, getStringMapErr := getLogMetricGroupMapData(logMetricGroupGuid)
	if getStringMapErr != nil {
		err = getStringMapErr
		return
	}
	var logMetricParamRows []*models.LogMetricParam
	err = x.SQL("select * from log_metric_param where log_metric_group=?", logMetricGroupGuid).Find(&logMetricParamRows)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_param fail,%s ", err.Error())
	}
	for _, row := range logMetricParamRows {
		tmpParamObj := models.LogMetricParamObj{LogMetricParam: *row, StringMap: []*models.LogMetricStringMapTable{}}
		if stringMapData, ok := logMetricStringMapData[row.Name]; ok {
			tmpParamObj.StringMap = stringMapData
		}
		result.ParamList = append(result.ParamList, &tmpParamObj)
	}
	var logMetricConfigRows []*models.LogMetricConfigTable
	logMetricConfigRows, err = getLogMetricConfigByMetricGroup(logMetricGroupGuid)
	if err != nil {
		return
	}
	for _, row := range logMetricConfigRows {
		json.Unmarshal([]byte(row.TagConfig), &row.TagConfigList)
		row.FullMetric = row.Metric
		// 只有当指标名称确实以指标前缀代码开头时，才去掉前缀
		if strings.TrimSpace(metricGroupObj.MetricPrefixCode) != "" &&
			strings.HasPrefix(row.Metric, metricGroupObj.MetricPrefixCode+"_") {
			row.Metric = row.Metric[len(metricGroupObj.MetricPrefixCode)+1:]
		}
		result.MetricList = append(result.MetricList, convertLogMetricConfigTable2Dto(row))
	}
	return
}

func getLogMetricConfigByMetricGroup(logMetricGroupGuid string) (logMetricConfigRows []*models.LogMetricConfigTable, err error) {
	err = x.SQL("select * from log_metric_config where log_metric_group=?", logMetricGroupGuid).Find(&logMetricConfigRows)
	if err != nil {
		err = fmt.Errorf("Query table log_metric_param fail,%s ", err.Error())
	}
	return
}

func getLogMetricGroupMapData(logMetricGroupGuid string) (result map[string][]*models.LogMetricStringMapTable, err error) {
	result = make(map[string][]*models.LogMetricStringMapTable)
	var logMetricStringMapRows []*models.LogMetricStringMapTable
	err = x.SQL("select * from log_metric_string_map where log_metric_group=?", logMetricGroupGuid).Find(&logMetricStringMapRows)
	if err != nil {
		err = fmt.Errorf("Query table log_metric_string_map fail,%s ", err.Error())
		return
	}
	for _, stringMapRow := range logMetricStringMapRows {
		if existList, ok := result[stringMapRow.LogParamName]; ok {
			result[stringMapRow.LogParamName] = append(existList, stringMapRow)
		} else {
			result[stringMapRow.LogParamName] = []*models.LogMetricStringMapTable{stringMapRow}
		}
	}
	return
}

func CreateLogMetricCustomGroup(param *models.LogMetricGroupObj, operator string, roles []string, errMsgObj *models.ErrorTemplate) (result *models.CreateLogMetricGroupDto, err error) {
	param.Guid = ""
	var actions []*Action
	var newDashboardId int64
	if actions, result, newDashboardId, err = getCreateLogMetricCustomGroupActions(param, operator, make(map[string]string), roles, errMsgObj, false); err != nil {
		deleteCustomDashboard(newDashboardId)
		return
	}
	if err = Transaction(actions); err != nil {
		deleteCustomDashboard(newDashboardId)
	}
	return
}

func getCreateLogMetricCustomGroupActions(param *models.LogMetricGroupObj, operator string, existMetricMap map[string]string, roles []string, errMsgObj *models.ErrorTemplate, doImport bool) (actions []*Action, result *models.CreateLogMetricGroupDto, newDashboardId int64, err error) {
	var endpointGroup, refTemplateVersion string
	var subCreateAlarmStrategyActions, subCreateDashboardActions []*Action
	var serviceGroupsRoles, alarmStrategyList []string
	var logMonitorTemplate *models.LogMonitorTemplateDto
	var templateSnapshot []byte
	var autoAlarm, autoDashboard int
	param.LogType = "custom"
	if param.Guid == "" {
		param.Guid = "lmg_" + guid.CreateGuid()
	}
	if param.AutoCreateWarn {
		autoAlarm = 1
	}
	if param.AutoCreateDashboard {
		autoDashboard = 1
	}
	nowTime := time.Now()
	result = &models.CreateLogMetricGroupDto{AlarmList: []string{}}
	if strings.TrimSpace(param.LogMonitorTemplateGuid) == "" && strings.TrimSpace(param.LogMonitorTemplate) != "" {
		param.LogMonitorTemplateGuid = param.LogMonitorTemplate
	}
	if strings.TrimSpace(param.LogMonitorTemplateGuid) != "" {
		if logMonitorTemplate, err = GetLogMonitorTemplate(param.LogMonitorTemplateGuid); err != nil {
			return
		}
		if logMonitorTemplate == nil {
			err = fmt.Errorf("logMonitorTemplate id:%s is valid", param.LogMonitorTemplateGuid)
			return
		}
		if templateSnapshot, err = json.Marshal(logMonitorTemplate); err != nil {
			return
		}
		refTemplateVersion = logMonitorTemplate.UpdateTime.Format(models.DatetimeDigitFormat)
		actions = append(actions, &Action{Sql: "insert into log_metric_group(guid,name,log_type,log_metric_monitor,log_monitor_template,demo_log,calc_result,create_user," +
			"create_time,update_user,update_time,metric_prefix_code,template_snapshot,ref_template_version,auto_alarm,auto_dashboard) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			param.Guid, param.Name, param.LogType, param.LogMetricMonitor, param.LogMonitorTemplateGuid, param.DemoLog, param.CalcResult, operator, nowTime, operator,
			nowTime, param.MetricPrefixCode, templateSnapshot, refTemplateVersion, autoAlarm, autoDashboard,
		}})
	} else {
		actions = append(actions, &Action{Sql: "insert into log_metric_group(guid,name,log_type,log_metric_monitor,demo_log,calc_result,create_user,create_time,update_user," +
			"update_time,metric_prefix_code,auto_alarm,auto_dashboard) values (?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			param.Guid, param.Name, param.LogType, param.LogMetricMonitor, param.DemoLog, param.CalcResult, operator, nowTime, operator, nowTime, param.MetricPrefixCode, autoAlarm, autoDashboard,
		}})
	}
	paramGuidList := guid.CreateGuidList(len(param.ParamList))
	for i, v := range param.ParamList {
		actions = append(actions, &Action{Sql: "insert into log_metric_param(guid,name,display_name,log_metric_group,regular,demo_match_value,create_user,create_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lmp_" + paramGuidList[i], v.Name, v.DisplayName, param.Guid, v.Regular, v.DemoMatchValue, operator, nowTime,
		}})
		tmpStringMapGuidList := guid.CreateGuidList(len(v.StringMap))
		for stringMapIndex, stringMapObj := range v.StringMap {
			actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_group,log_param_name,value_type,source_value,regulative,target_value,update_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmsm_" + tmpStringMapGuidList[stringMapIndex], param.Guid, v.Name, stringMapObj.ValueType, stringMapObj.SourceValue, stringMapObj.Regulative, stringMapObj.TargetValue, nowTime.Format(models.DatetimeFormat),
			}})
		}
	}
	metricGuidList := guid.CreateGuidList(len(param.MetricList))
	serviceGroup, monitorType := param.ServiceGroup, param.MonitorType
	if param.ServiceGroup == "" {
		serviceGroup, monitorType = GetLogMetricServiceGroup(param.LogMetricMonitor)
	}
	for i, v := range param.MetricList {
		v.Step = 10
		tmpTagListBytes, _ := json.Marshal(v.TagConfigList)
		tmpMetricConfigGuid := "lmc_" + metricGuidList[i]
		rangeConf, _ := json.Marshal(v.RangeConfig)
		tmpMetricWithPrefix := v.Metric
		if param.MetricPrefixCode != "" {
			tmpMetricWithPrefix = param.MetricPrefixCode + "_" + v.Metric
		}
		actions = append(actions, &Action{Sql: "insert into log_metric_config(guid,log_metric_monitor,log_metric_group,log_param_name,metric,display_name," +
			"regular,step,agg_type,tag_config,create_user,create_time,auto_alarm,range_config,color_group) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			tmpMetricConfigGuid, param.LogMetricMonitor, param.Guid, v.LogParamName, tmpMetricWithPrefix, v.DisplayName, v.Regular, v.Step, v.AggType, string(tmpTagListBytes),
			operator, nowTime, v.AutoAlarm, string(rangeConf), v.ColorGroup,
		}})
		// 自动添加增加 metric
		var tmpMetricTags []string
		if len(v.TagConfigList) > 0 {
			tmpMetricTags = []string{"tags"}
		}
		tmpMetricGuid := generateMetricGuid(tmpMetricWithPrefix, serviceGroup)
		if duplicateMetric, ok := existMetricMap[tmpMetricGuid]; ok && !doImport {
			err = fmt.Errorf("Metric: %s duplicate ", duplicateMetric)
			return
		}
		actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,log_metric_config,log_metric_group,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?,?,?)",
			Param: []interface{}{tmpMetricGuid, tmpMetricWithPrefix, monitorType, getLogMetricExprByAggType(tmpMetricWithPrefix, v.AggType, serviceGroup, tmpMetricTags), serviceGroup, models.MetricWorkspaceService, nowTime, tmpMetricConfigGuid, param.Guid, nowTime, operator, operator}})
	}
	if param.LogMetricMonitor != "" {
		var logMetricMonitor = &models.LogMetricMonitorTable{}
		var endpointGroupIds []string
		if _, err = x.SQL("select service_group,monitor_type from log_metric_monitor where guid=?", param.LogMetricMonitor).Get(logMetricMonitor); err != nil {
			return
		}
		if logMetricMonitor != nil {
			serviceGroupsRoles = getServiceGroupRoles(logMetricMonitor.ServiceGroup)
			if err = x.SQL("select guid from endpoint_group where service_group=? and monitor_type=?", logMetricMonitor.ServiceGroup, logMetricMonitor.MonitorType).Find(&endpointGroupIds); err != nil {
				return
			}
			if len(endpointGroupIds) > 0 {
				endpointGroup = endpointGroupIds[0]
			}
		}
	}
	// endpointGroup为空
	if strings.TrimSpace(endpointGroup) == "" {
		var endpointGroupIds []string
		serviceGroupsRoles = getServiceGroupRoles(param.ServiceGroup)
		if err = x.SQL("select guid from endpoint_group where service_group=? and monitor_type=?", param.ServiceGroup, param.MonitorType).Find(&endpointGroupIds); err != nil {
			return
		}
		if len(endpointGroupIds) > 0 {
			endpointGroup = endpointGroupIds[0]
		}
	}
	if len(serviceGroupsRoles) == 0 && len(roles) > 0 {
		serviceGroupsRoles = roles[:1]
	}
	alarmStrategyParam := models.AutoSimpleAlarmStrategyParam{MetricList: param.MetricList, ServiceGroupsRoles: serviceGroupsRoles, LogMetricGroupGuid: param.LogMetricGroup.Guid,
		ServiceGroup: serviceGroup, EndpointGroup: endpointGroup, Operator: operator, ErrMsgObj: errMsgObj, AutoCreateWarn: param.AutoCreateWarn,
		MetricPrefixCode: param.LogMetricGroup.MetricPrefixCode, LogType: models.LogMonitorCustomType}
	if subCreateAlarmStrategyActions, alarmStrategyList, err = autoGenerateSimpleAlarmStrategy(alarmStrategyParam); err != nil {
		return
	}
	if len(subCreateAlarmStrategyActions) > 0 {
		actions = append(actions, subCreateAlarmStrategyActions...)
		result.AlarmList = alarmStrategyList
	}
	var dashboardParam = models.AutoSimpleCreateDashboardParam{MetricList: param.MetricList, ServiceGroupsRoles: serviceGroupsRoles,
		ServiceGroup: serviceGroup, Operator: operator, ErrMsgObj: errMsgObj, AutoCreateDashboard: param.AutoCreateDashboard,
		LogMetricGroupGuid: param.LogMetricGroup.Guid, MetricPrefixCode: param.LogMetricGroup.MetricPrefixCode, MonitorType: param.MonitorType}
	if subCreateDashboardActions, result.CustomDashboard, newDashboardId, err = autoGenerateSimpleCustomDashboard(dashboardParam); err != nil {
		return
	}
	if len(subCreateDashboardActions) > 0 {
		actions = append(actions, subCreateDashboardActions...)
	}
	return
}

func ValidateLogMetricGroupName(guid, name, logMetricMonitor string) (err error) {
	queryResult, queryErr := x.QueryString("select guid from log_metric_group where log_metric_monitor=? and name=? and guid!=?", logMetricMonitor, name, guid)
	if queryErr != nil {
		err = fmt.Errorf("query log metric group table fail,%s ", queryErr.Error())
		return
	}
	if len(queryResult) > 0 {
		err = fmt.Errorf("log metric group name:%s duplicate", name)
	}
	return
}

func UpdateLogMetricCustomGroup(param *models.LogMetricGroupObj, operator string) (err error) {
	var actions []*Action
	var affectEndpointGroup []string
	actions, affectEndpointGroup, err = getUpdateLogMetricCustomGroupActions(param, operator)
	if err != nil {
		return
	}

	/* 自定义告警更新更新对应看板逻辑 开启前需要测试验证下
	// Add dashboard sync actions if auto_dashboard is enabled
	if param.AutoCreateDashboard {
		dashboardSyncActions, syncErr := getCustomLogMetricGroupDashboardSyncActions(param, operator)
		if syncErr != nil {
			log.Error(nil, log.LOGGER_APP, "Failed to generate dashboard sync actions for custom log metric group", zap.Error(syncErr), zap.String("guid", param.Guid))
			return syncErr
		}
		actions = append(actions, dashboardSyncActions...)
	}
	*/
	err = Transaction(actions)
	if err == nil {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return
}

func getUpdateLogMetricCustomGroupActions(param *models.LogMetricGroupObj, operator string) (actions []*Action, affectEndpointGroup []string, err error) {
	existLogGroupData, getExistErr := GetLogMetricCustomGroup(param.Guid)
	if getExistErr != nil {
		err = getExistErr
		return
	}
	nowTime := time.Now()
	actions = append(actions, &Action{Sql: "update log_metric_group set name=?,demo_log=?,calc_result=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
		param.Name, param.DemoLog, param.CalcResult, operator, nowTime, param.Guid,
	}})
	actions = append(actions, &Action{Sql: "delete from log_metric_string_map where log_metric_group=?", Param: []interface{}{param.Guid}})
	paramGuidList := guid.CreateGuidList(len(param.ParamList))
	for i, inputParamObj := range param.ParamList {
		if inputParamObj.Guid == "" {
			actions = append(actions, &Action{Sql: "insert into log_metric_param(guid,name,display_name,log_metric_group,regular,demo_match_value,create_user,create_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmp_" + paramGuidList[i], inputParamObj.Name, inputParamObj.DisplayName, param.Guid, inputParamObj.Regular, inputParamObj.DemoMatchValue, operator, nowTime,
			}})
		} else {
			actions = append(actions, &Action{Sql: "update log_metric_param set name=?,display_name=?,regular=?,demo_match_value=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
				inputParamObj.Name, inputParamObj.DisplayName, inputParamObj.Regular, inputParamObj.DemoMatchValue, operator, nowTime, inputParamObj.Guid,
			}})
		}
		tmpStringMapGuidList := guid.CreateGuidList(len(inputParamObj.StringMap))
		for stringMapIndex, stringMapObj := range inputParamObj.StringMap {
			actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_metric_group,log_param_name,value_type,source_value,regulative,target_value,update_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmsm_" + tmpStringMapGuidList[stringMapIndex], param.Guid, inputParamObj.Name, stringMapObj.ValueType, stringMapObj.SourceValue, stringMapObj.Regulative, stringMapObj.TargetValue, nowTime.Format(models.DatetimeFormat),
			}})
		}
	}
	for _, existParamObj := range existLogGroupData.ParamList {
		deleteFlag := true
		for _, inputParamObj := range param.ParamList {
			if inputParamObj.Guid == existParamObj.Guid {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			actions = append(actions, &Action{Sql: "delete from log_metric_param where guid=?", Param: []interface{}{existParamObj.Guid}})
		}
	}
	serviceGroup, monitorType := GetLogMetricServiceGroup(param.LogMetricMonitor)
	metricGuidList := guid.CreateGuidList(len(param.MetricList))
	existMetricDataMap := make(map[string]*models.LogMetricConfigDto)
	for _, existMetricObj := range existLogGroupData.MetricList {
		existMetricDataMap[existMetricObj.Guid] = existMetricObj
		deleteFlag := true
		for _, inputMetricObj := range param.MetricList {
			if inputMetricObj.Guid == existMetricObj.Guid {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			actions = append(actions, &Action{Sql: "delete from log_metric_config where guid=?", Param: []interface{}{existMetricObj.Guid}})
		}
	}
	for i, inputMetricObj := range param.MetricList {
		tmpTagListBytes, _ := json.Marshal(inputMetricObj.TagConfigList)
		inputMetricObj.TagConfig = string(tmpTagListBytes)
		if inputMetricObj.Guid == "" {
			tmpMetricConfigGuid := "lmc_" + metricGuidList[i]
			actions = append(actions, &Action{Sql: "insert into log_metric_config(guid,log_metric_monitor,log_metric_group,log_param_name,metric,display_name,regular,step,agg_type,tag_config,color_group,create_user,create_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				tmpMetricConfigGuid, existLogGroupData.LogMetricMonitor, param.Guid, inputMetricObj.LogParamName, inputMetricObj.Metric, inputMetricObj.DisplayName, inputMetricObj.Regular, inputMetricObj.Step, inputMetricObj.AggType, string(tmpTagListBytes), inputMetricObj.ColorGroup, operator, nowTime,
			}})
			tmpTagList := []string{}
			if len(inputMetricObj.TagConfigList) > 0 {
				tmpTagList = []string{"tags"}
			}
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,log_metric_config,log_metric_group,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{fmt.Sprintf("%s__%s", inputMetricObj.Metric, serviceGroup), inputMetricObj.Metric, monitorType, getLogMetricExprByAggType(inputMetricObj.Metric, inputMetricObj.AggType, serviceGroup,
					tmpTagList), serviceGroup, models.MetricWorkspaceService, nowTime, tmpMetricConfigGuid, param.Guid, nowTime, operator, operator}})
		} else {
			actions = append(actions, &Action{Sql: "update log_metric_config set log_param_name=?,metric=?,display_name=?,regular=?,step=?,agg_type=?,tag_config=?,color_group=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
				inputMetricObj.LogParamName, inputMetricObj.Metric, inputMetricObj.DisplayName, inputMetricObj.Regular, inputMetricObj.Step, inputMetricObj.AggType, string(tmpTagListBytes), inputMetricObj.ColorGroup, operator, nowTime, inputMetricObj.Guid,
			}})
			if existMetricObj, ok := existMetricDataMap[inputMetricObj.Guid]; ok {
				oldMetricGuid := fmt.Sprintf("%s__%s", existMetricObj.Metric, serviceGroup)
				newMetricGuid := fmt.Sprintf("%s__%s", inputMetricObj.Metric, serviceGroup)
				tmpTagList := []string{}
				if len(inputMetricObj.TagConfigList) > 0 {
					tmpTagList = []string{"tags"}
				}
				actions = append(actions, &Action{Sql: "update metric set guid=?,metric=?,prom_expr=?,update_user=?,update_time=?,log_metric_group=? where guid=?",
					Param: []interface{}{newMetricGuid, inputMetricObj.Metric, getLogMetricExprByAggType(inputMetricObj.Metric, inputMetricObj.AggType, serviceGroup, tmpTagList), operator, nowTime, param.Guid, oldMetricGuid}})
				var alarmStrategyTable []*models.AlarmStrategyTable
				x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", oldMetricGuid).Find(&alarmStrategyTable)
				if len(alarmStrategyTable) > 0 {
					for _, v := range alarmStrategyTable {
						affectEndpointGroup = append(affectEndpointGroup, v.EndpointGroup)
					}
					actions = append(actions, &Action{Sql: "update alarm_strategy set metric=? where metric=?", Param: []interface{}{newMetricGuid, oldMetricGuid}})
				}
			}
		}
	}
	log.Info(nil, log.LOGGER_APP, "getUpdateLogMetricCustomGroupActions generated actions", zap.Int("actionCount", len(actions)))
	return
}

// getCustomLogMetricGroupDashboardSyncActions generates actions to sync dashboard when custom log metric group is updated
func getCustomLogMetricGroupDashboardSyncActions(param *models.LogMetricGroupObj, operator string) (actions []*Action, err error) {
	// Get existing data for comparison
	existData, getErr := GetLogMetricCustomGroup(param.Guid)
	if getErr != nil {
		err = getErr
		return
	}

	// Calculate metric changes by comparing existing vs new metric list
	var metricsAdded []models.LogMetricConfigDto
	var metricsDeleted []models.LogMetricConfigDto
	var metricsModified []models.LogMetricConfigDto

	// Build maps for comparison
	existMetricMap := make(map[string]*models.LogMetricConfigDto) // metric name -> metric
	newMetricMap := make(map[string]*models.LogMetricConfigDto)   // metric name -> metric

	for _, metric := range existData.MetricList {
		existMetricMap[metric.Metric] = metric
	}

	for _, metric := range param.MetricList {
		newMetricMap[metric.Metric] = metric
	}

	// Find added and modified metrics
	for metricName, newMetric := range newMetricMap {
		if existMetric, exists := existMetricMap[metricName]; exists {
			// Check if metric was modified (display name, agg type, log_param_name etc.)
			if existMetric.AggType != newMetric.AggType || existMetric.LogParamName != newMetric.LogParamName ||
				!compareTagConfigList(existMetric.TagConfigList, newMetric.TagConfigList) {
				metricsModified = append(metricsModified, *newMetric)
			}
		} else {
			// New metric
			metricsAdded = append(metricsAdded, *newMetric)
		}
	}

	// Find deleted metrics
	for metricName, existMetric := range existMetricMap {
		if _, exists := newMetricMap[metricName]; !exists {
			metricsDeleted = append(metricsDeleted, *existMetric)
		}
	}

	// Sync dashboard if there are changes
	if len(metricsAdded) > 0 || len(metricsDeleted) > 0 || len(metricsModified) > 0 {
		syncActions, syncErr := syncCustomDashboardForMetricChanges(param.Guid, metricsAdded, metricsDeleted, metricsModified, param.MetricPrefixCode, operator)
		if syncErr != nil {
			err = syncErr
			return
		}
		actions = append(actions, syncActions...)

		// 1) 处理 tag_config 的增删（按每个被修改的 metric）
		// 规则：新增 key -> 为该组相关图表下的每个 series 新增一条 tag 记录（equal 默认 in）；
		//      删除 key -> 删除对应 tag 记录及其 tag_value 记录
		// 先获取当前组对应的 dashboardId
		var dashboardId int
		_, _ = x.SQL("select id from custom_dashboard where log_metric_group=?", param.Guid).Get(&dashboardId)

		// 计算展示用的 ServiceGroup 名称
		displayServiceGroupName := existData.ServiceGroup
		var sg models.ServiceGroupTable
		if existData.ServiceGroup != "" {
			x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", existData.ServiceGroup).Get(&sg)
			if sg.DisplayName != "" {
				displayServiceGroupName = sg.DisplayName
			}
		}
		for _, newMetric := range metricsModified {
			oldMetric := existMetricMap[newMetric.Metric]
			if oldMetric == nil {
				continue
			}
			// 计算该 metric 的图表名称
			chartNameForMetric := fmt.Sprintf("%s_%s/%s", param.MetricPrefixCode, newMetric.Metric, displayServiceGroupName)
			// 计算差集
			oldSet := make(map[string]bool)
			newSet := make(map[string]bool)
			for _, k := range oldMetric.TagConfigList {
				oldSet[k] = true
			}
			for _, k := range newMetric.TagConfigList {
				newSet[k] = true
			}
			// 删除的 keys
			for k := range oldSet {
				if !newSet[k] {
					// 删除 tag_value
					actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag_value where dashboard_chart_tag in (select guid from custom_chart_series_tag where name=? and dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=? and name=?)))",
						Param: []interface{}{k, param.Guid, chartNameForMetric}})
					// 删除 tag
					actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag where name=? and dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=? and name=?))",
						Param: []interface{}{k, param.Guid, chartNameForMetric}})
				}
			}
			// 新增的 keys
			for k := range newSet {
				if !oldSet[k] {
					// 为所有相关 series 新增 tag（equal 默认 in）
					actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) select ?, s.guid, ?, ? from custom_chart_series s where s.dashboard_chart in (select guid from custom_chart where log_metric_group=? and name=?)",
						Param: []interface{}{guid.CreateGuid(), k, "in", param.Guid, chartNameForMetric}})
				}
			}
		}

		// 2) 处理 param_list.string_map 值变更对 series_config 的回溯替换
		// 规则：仅当该 param 被 metric_list[*].log_param_name 使用时，针对 old->new 的 target_value 逐一 REPLACE
		// 构建：paramName -> (source_value -> target_value)
		oldParamMap := make(map[string]map[string]string)
		newParamMap := make(map[string]map[string]string)
		for _, p := range existData.ParamList {
			m := make(map[string]string)
			for _, sm := range p.StringMap {
				m[sm.SourceValue] = sm.TargetValue
			}
			oldParamMap[p.Name] = m
		}
		for _, p := range param.ParamList {
			m := make(map[string]string)
			for _, sm := range p.StringMap {
				m[sm.SourceValue] = sm.TargetValue
			}
			newParamMap[p.Name] = m
		}
		// 计算被使用的 param 名称集合
		usedParam := make(map[string]bool)
		for _, m := range param.MetricList {
			if m.LogParamName != "" {
				usedParam[m.LogParamName] = true
			}
		}
		// 遍历变化
		for pname := range usedParam {
			oldKV := oldParamMap[pname]
			newKV := newParamMap[pname]
			if oldKV == nil || newKV == nil {
				continue
			}
			for src, oldTarget := range oldKV {
				if newTarget, ok := newKV[src]; ok && newTarget != oldTarget {
					// 统一替换 tags 与 series_name 中的值片段： pname=oldTarget -> pname=newTarget
					actions = append(actions, &Action{Sql: "update custom_chart_series_config set tags=REPLACE(tags, CONCAT(?, '=' , ?), CONCAT(?, '=' , ?)), series_name=REPLACE(series_name, CONCAT(?, '=' , ?), CONCAT(?, '=' , ?)) where dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=?))",
						Param: []interface{}{pname, oldTarget, pname, newTarget, pname, oldTarget, pname, newTarget, param.Guid}})
				}
			}
		}
	}

	return
}

// compareTagConfigList compares two tag config lists for equality
func compareTagConfigList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// syncCustomDashboardForMetricChanges handles dashboard synchronization for custom log metric group metric changes
func syncCustomDashboardForMetricChanges(logMetricGroupGuid string, metricsAdded, metricsDeleted, metricsModified []models.LogMetricConfigDto,
	metricPrefixCode, operator string) (actions []*Action, err error) {

	// Get existing dashboard
	var dashboardId int
	if _, err = x.SQL("select id from custom_dashboard where log_metric_group=?", logMetricGroupGuid).Get(&dashboardId); err != nil {
		return
	}
	if dashboardId == 0 {
		// No dashboard exists, nothing to sync
		return
	}

	// Get service group info
	serviceGroup, monitorType := GetLogMetricServiceGroup(logMetricGroupGuid)

	// Get service group display name
	var serviceGroupTable models.ServiceGroupTable
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", serviceGroup).Get(&serviceGroupTable)
	displayServiceGroup := serviceGroup
	if serviceGroupTable.DisplayName != "" {
		displayServiceGroup = serviceGroupTable.DisplayName
	}

	now := time.Now()

	// Prepare dashboard roles for chart creation via handleAutoCreateChart
	var roleRelList []*models.CustomDashBoardRoleRel
	var useRoles []string
	var mgmtRole string
	if dashboardId > 0 {
		if roleRelList, err = QueryCustomDashboardRoleRelByCustomDashboard(dashboardId); err != nil {
			return
		}
		if len(roleRelList) > 0 {
			for _, r := range roleRelList {
				if r.Permission == string(models.PermissionMgmt) {
					mgmtRole = r.RoleId
				} else if r.Permission == string(models.PermissionUse) {
					useRoles = append(useRoles, r.RoleId)
				}
			}
		}
	}

	// Handle deleted metrics - remove corresponding charts
	for _, metric := range metricsDeleted {
		chartName := fmt.Sprintf("%s_%s/%s", metricPrefixCode, metric.Metric, displayServiceGroup)
		actions = append(actions, &Action{
			Sql:   "delete from custom_chart where log_metric_group=? and name=?",
			Param: []interface{}{logMetricGroupGuid, chartName},
		})
	}

	// Handle added metrics - create new charts
	for _, metric := range metricsAdded {
		chartName := fmt.Sprintf("%s_%s/%s", metricPrefixCode, metric.Metric, displayServiceGroup)

		// Check if chart already exists
		var existingChartId string
		x.SQL("select guid from custom_chart where log_metric_group=? and name=?", logMetricGroupGuid, chartName).Get(&existingChartId)
		if existingChartId != "" {
			continue // Chart already exists
		}

		// Get current chart count for display config calculation
		var chartCount int
		x.SQL("select count(*) from custom_chart where log_metric_group=?", logMetricGroupGuid).Get(&chartCount)

		// Create new chart params, align with autoGenerateSimpleCustomDashboard
		chartParam := &models.CustomChartDto{
			Public:          true,
			SourceDashboard: dashboardId,
			Name:            chartName,
			ChartTemplate:   "one",
			ChartType:       "line",
			LineType:        "line",
			Aggregate:       "none",
			AggStep:         60,
			ChartSeries:     []*models.CustomChartSeriesDto{},
			DisplayConfig:   calcDisplayConfigForCustom(chartCount),
			LogMetricGroup:  &logMetricGroupGuid,
		}

		// Generate chart series using the same logic as generateSimpleChartSeries
		chartParam.ChartSeries = append(chartParam.ChartSeries, generateCustomChartSeries(serviceGroup, monitorType, metricPrefixCode, &metric))

		// Create chart via handleAutoCreateChart to keep logic consistent with creation flow
		subChartActions := handleAutoCreateChart(chartParam, int64(dashboardId), useRoles, mgmtRole, operator)
		if len(subChartActions) > 0 {
			actions = append(actions, subChartActions...)
		}
	}

	// Handle modified metrics - update existing charts
	for _, metric := range metricsModified {
		chartName := fmt.Sprintf("%s_%s/%s", metricPrefixCode, metric.Metric, displayServiceGroup)

		// Update chart display name if it changed
		actions = append(actions, &Action{
			Sql:   "update custom_chart set name=?, update_user=?, update_time=? where log_metric_group=? and name=?",
			Param: []interface{}{chartName, operator, now, logMetricGroupGuid, chartName},
		})

		// Update chart series with new metric info
		fullMetricName := metric.Metric
		if metricPrefixCode != "" && !strings.HasPrefix(metric.Metric, metricPrefixCode+"_") {
			fullMetricName = metricPrefixCode + "_" + metric.Metric
		}

		// Update series metric and prom_expr
		promExpr := getLogMetricExprByAggType(fullMetricName, metric.AggType, serviceGroup, []string{"tags"})
		actions = append(actions, &Action{
			Sql:   "update custom_chart_series set metric=?, prom_expr=?, display_name=?, legend=?, update_user=?, update_time=? where custom_chart in (select guid from custom_chart where log_metric_group=? and name=?)",
			Param: []interface{}{fullMetricName, promExpr, metric.DisplayName, metric.DisplayName, operator, now, logMetricGroupGuid, chartName},
		})

		// If log_param_name changed, unify replace in series_config.tags/series_name and series_tag.name
		var oldLogParamName string
		x.SQL("select log_param_name from log_metric_config where log_metric_group=? and metric=?", logMetricGroupGuid, metric.Metric).Get(&oldLogParamName)
		if oldLogParamName != "" && oldLogParamName != metric.LogParamName {
			// Update custom_chart_series_config tags and series_name via dashboard_chart path
			actions = append(actions, &Action{Sql: "update custom_chart_series_config set tags=REPLACE(tags, CONCAT(?, '='), CONCAT(?, '=')), series_name=REPLACE(series_name, CONCAT('{', ?, '='), CONCAT('{', ?, '=')) where dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=? and name=?))",
				Param: []interface{}{oldLogParamName, metric.LogParamName, oldLogParamName, metric.LogParamName, dashboardId, chartName}})
			// Update custom_chart_series_tag name via dashboard_chart path
			actions = append(actions, &Action{Sql: "update custom_chart_series_tag set name=? where name=? and dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=? and name=?))",
				Param: []interface{}{metric.LogParamName, oldLogParamName, dashboardId, chartName}})
		}
	}

	// Recalculate and unify layout at the end when there are any changes
	if (len(metricsAdded) + len(metricsDeleted) + len(metricsModified)) > 0 {
		var chartIds []string
		if err = x.SQL("select guid from custom_chart where source_dashboard=? order by name", dashboardId).Find(&chartIds); err != nil {
			return
		}
		for idx, chartId := range chartIds {
			cfg := calcDisplayConfigForCustom(idx)
			cfgBytes, _ := json.Marshal(cfg)
			actions = append(actions, &Action{Sql: "update custom_dashboard_chart_rel set display_config=? where custom_dashboard=? and dashboard_chart=?", Param: []interface{}{string(cfgBytes), dashboardId, chartId}})
		}
	}

	return
}

// generateCustomChartSeries generates chart series for custom metrics (similar to generateSimpleChartSeries)
func generateCustomChartSeries(serviceGroup, monitorType, metricPrefixCode string, metric *models.LogMetricConfigDto) *models.CustomChartSeriesDto {
	var serviceGroupTable = &models.ServiceGroupTable{}
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", serviceGroup).Get(serviceGroupTable)

	metricGuid := metric.Metric
	fullMetricName := metric.Metric
	// Check if metric.Metric already contains metricPrefixCode, avoid duplicate addition
	if metricPrefixCode != "" && !strings.HasPrefix(metric.Metric, metricPrefixCode+"_") {
		metricGuid = metricPrefixCode + "_" + metricGuid
		fullMetricName = metricPrefixCode + "_" + metric.Metric
	}

	dto := &models.CustomChartSeriesDto{
		Endpoint:     serviceGroup,
		ServiceGroup: serviceGroup,
		EndpointName: serviceGroup,
		MonitorType:  monitorType,
		ColorGroup:   metric.ColorGroup,
		MetricType:   "business",
		MetricGuid:   generateMetricGuid(metricGuid, serviceGroup),
		Metric:       fullMetricName,
		Tags:         make([]*models.TagDto, 0),
	}

	if serviceGroupTable != nil {
		dto.EndpointName = serviceGroupTable.DisplayName
		dto.EndpointType = serviceGroupTable.ServiceType
	}

	if len(metric.TagConfigList) > 0 {
		dto.Tags = append(dto.Tags, &models.TagDto{TagName: "code", Equal: "in"})
	}

	return dto
}

// handleCustomChartCreation handles the creation of custom charts and series
// handleCustomChartCreation 已废弃：更新场景改为复用 handleAutoCreateChart，统一落库到 custom_dashboard_chart_rel

// calcDisplayConfigForCustom generates display config for custom charts (similar to calcDisplayConfig)
func calcDisplayConfigForCustom(index int) models.DisplayConfig {
	item := models.DisplayConfig{}
	item.W = 4
	item.H = 7
	item.X = float64((index % 3) * 4)
	item.Y = math.Floor(float64(index/3)) * 7
	return item
}

func GetServiceGroupMetricMap(serviceGroup string) (existMetricMap map[string]string, err error) {
	var metricRows []*models.MetricTable
	err = x.SQL("select metric,monitor_type,log_metric_group from metric where service_group=?", serviceGroup).Find(&metricRows)
	if err != nil {
		err = fmt.Errorf("query metric table fail,%s ", err.Error())
		return
	}
	existMetricMap = make(map[string]string)
	for _, v := range metricRows {
		existMetricMap[v.Metric] = v.LogMetricGroup
	}
	return
}

func GetLogMetricMonitorMetricPrefixMap(logMetricMonitor string) (existPrefixMap map[string]int, err error) {
	queryResult, queryErr := x.QueryString("select metric_prefix_code from log_metric_group where log_metric_monitor=?", logMetricMonitor)
	if queryErr != nil {
		err = fmt.Errorf("query log metric group table fail,%s ", queryErr.Error())
		return
	}
	existPrefixMap = make(map[string]int)
	for _, v := range queryResult {
		existPrefixMap[v["metric_prefix_code"]] = 1
	}
	return
}

func GetLogMetricByLogMonitor(logMetricMonitor string) (result []*models.LogMetricConfigTable, err error) {
	var logMetricGroupRows []*models.LogMetricGroup
	err = x.SQL("select * from log_metric_group where log_metric_monitor=?", logMetricMonitor).Find(&logMetricGroupRows)
	if err != nil {
		err = fmt.Errorf("query log_metric_group table fail,%s ", err.Error())
		return
	}
	for _, logMetricGroup := range logMetricGroupRows {
		if logMetricGroup.LogMonitorTemplate != "" {
			if logMetricTemplateRows, tmpErr := getLogMetricTemplateWithMonitor(logMetricGroup.LogMonitorTemplate); tmpErr != nil {
				err = tmpErr
				return
			} else {
				for _, logMetricTemplate := range logMetricTemplateRows {
					result = append(result, &models.LogMetricConfigTable{Metric: fmt.Sprintf("%s_%s", logMetricGroup.MetricPrefixCode, logMetricTemplate.Metric), AggType: logMetricTemplate.AggType, DisplayName: logMetricTemplate.DisplayName})
				}
			}
		} else {
			if logMetricConfigRows, tmpErr := getLogMetricConfigByMetricGroup(logMetricGroup.Guid); tmpErr != nil {
				err = tmpErr
				return
			} else {
				for _, logMetricConfig := range logMetricConfigRows {
					result = append(result, &models.LogMetricConfigTable{Metric: logMetricConfig.Metric, AggType: logMetricConfig.AggType, DisplayName: logMetricConfig.DisplayName})
				}
			}
		}
	}
	return
}

func getServiceGroupMetricMap(serviceGroup string) (metricGuidMap map[string]string, err error) {
	var metricRows []*models.MetricTable
	err = x.SQL("select guid,metric from metric where service_group=?", serviceGroup).Find(&metricRows)
	if err != nil {
		err = fmt.Errorf("query metric table fail,%s ", err.Error())
		return
	}
	metricGuidMap = make(map[string]string)
	for _, row := range metricRows {
		metricGuidMap[row.Guid] = row.Metric
	}
	return
}

func convertLogMetricConfigTable2Dto(config *models.LogMetricConfigTable) (dto *models.LogMetricConfigDto) {
	if config.TagConfig != "" && len(config.TagConfigList) == 0 {
		config.TagConfigList = strings.Split(config.TagConfig, ",")
	}
	var rangeConfig interface{}
	if config.RangeConfig != "" {
		json.Unmarshal([]byte(config.RangeConfig), &rangeConfig)
	}
	dto = &models.LogMetricConfigDto{
		Guid:             config.Guid,
		LogMetricMonitor: config.LogMetricMonitor,
		LogMetricGroup:   config.LogMetricGroup,
		LogParamName:     config.LogParamName,
		LogMetricJson:    config.LogMetricJson,
		Metric:           config.Metric,
		DisplayName:      config.DisplayName,
		JsonKey:          config.JsonKey,
		TagConfig:        config.TagConfig,
		Regular:          config.Regular,
		AggType:          config.AggType,
		Step:             config.Step,
		TagConfigList:    config.TagConfigList,
		CreateUser:       config.CreateUser,
		UpdateUser:       config.UpdateUser,
		CreateTime:       config.CreateTime.Format(models.DatetimeFormat),
		UpdateTime:       config.UpdateTime.Format(models.DatetimeFormat),
		AutoAlarm:        config.AutoAlarm,
		RangeConfig:      rangeConfig,
		ColorGroup:       config.ColorGroup,
		FullMetric:       config.FullMetric,
	}
	return
}

func GetLogMetricMonitorServiceGroups() []string {
	var serviceGroups []string
	x.SQL("select distinct service_group from log_metric_monitor").Find(&serviceGroups)
	return serviceGroups
}

func GetLogMetricGroupDto(logMetricGroup string) (dto *models.LogMetricGroupWarnDto, err error) {
	dto = &models.LogMetricGroupWarnDto{}
	_, err = x.SQL("select lmg.name as 'log_metric_group_name',lmg.log_metric_monitor as 'log_metric_monitor_guid',lmm.service_group from log_metric_group lmg join log_metric_monitor lmm  on lmg.log_metric_monitor= lmm.guid where lmg.guid =?", logMetricGroup).Get(dto)
	return
}

// UpdateLogMetricGroupWithDashboardAndAlarm wraps UpdateLogMetricGroup and, when auto_dashboard is enabled,
// computes code string_map diffs (add/rename/delete) and synchronizes dashboards accordingly.
// When auto_alarm is enabled, it will also (reserved) sync alarm strategies in future.
func UpdateLogMetricGroupWithDashboardAndAlarm(param *models.LogMetricGroupWithTemplate, operator string, roles []string) (err error) {
	// load old code string_map for diff
	var oldMaps []*models.LogMetricStringMapTable
	x.SQL("select guid,log_param_name,value_type,source_value,regulative,target_value from log_metric_string_map where log_metric_group=?", param.LogMetricGroupGuid).Find(&oldMaps)
	oldCode := make(map[string]*models.LogMetricStringMapTable)
	for _, v := range oldMaps {
		if v.LogParamName == "code" {
			oldCode[v.Guid] = v
		}
	}
	// check auto flags
	var autoDashboard, autoAlarm int
	x.SQL("select auto_dashboard,auto_alarm from log_metric_group where guid=?", param.LogMetricGroupGuid).Get(&autoDashboard, &autoAlarm)

	// 只有当auto_dashboard 才同步看板
	if autoDashboard == 1 {
		// diffs
		newByGuid := make(map[string]*models.LogMetricStringMapTable)
		for _, v := range param.CodeStringMap {
			newByGuid[v.Guid] = v
		}
		var adds []string
		renames := make(map[string]string)
		var deletes []string
		for guidKey, ov := range oldCode {
			if nv, ok := newByGuid[guidKey]; ok {
				if ov.TargetValue != nv.TargetValue {
					renames[ov.TargetValue] = nv.TargetValue
				}
			} else {
				found := false
				for _, nv := range param.CodeStringMap {
					if (nv.LogParamName == "code" || nv.LogParamName == "") && nv.ValueType == ov.ValueType && nv.SourceValue == ov.SourceValue && nv.Regulative == ov.Regulative {
						found = true
						if ov.TargetValue != nv.TargetValue {
							renames[ov.TargetValue] = nv.TargetValue
						}
						break
					}
				}
				if !found {
					deletes = append(deletes, ov.TargetValue)
				}
			}
		}
		oldTargets := make(map[string]bool)
		for _, ov := range oldCode {
			oldTargets[ov.TargetValue] = true
		}
		// Exclude pure renames from adds: collect all new target values coming from renames
		renamedNewTargets := make(map[string]bool)
		for _, newV := range renames {
			renamedNewTargets[newV] = true
		}
		for _, nv := range param.CodeStringMap {
			if nv.LogParamName != "code" {
				nv.LogParamName = "code"
			}
			if !oldTargets[nv.TargetValue] && !renamedNewTargets[nv.TargetValue] {
				adds = append(adds, nv.TargetValue)
			}
		}
		if len(renames) > 0 || len(deletes) > 0 || len(adds) > 0 {
			if autoDashboard == 1 {
				if err = SyncDashboardForCodeChanges(param.LogMetricGroupGuid, renames, deletes, adds, operator); err != nil {
					return err
				}
			}
			/*
				    // NY要求只更新看板,不更新阈值告警
					 if autoAlarm == 1 {
						if err = SyncAlarmStrategyForCodeChanges(param, renames, deletes, adds, operator, roles); err != nil {
							return err
						}
					}*/
		}
	}

	// 无论auto_dashboard和auto_alarm是否开启，都要执行核心更新操作
	if err = UpdateLogMetricGroup(param, operator); err != nil {
		return err
	}
	return nil
}

// SyncAlarmStrategyForCodeChanges synchronizes alarm strategies when codes are added, renamed, or deleted
func SyncAlarmStrategyForCodeChanges(param *models.LogMetricGroupWithTemplate, codeRenames map[string]string, codeDeletes []string, codesAdded []string, operator string, roles []string) (err error) {
	var actions []*Action

	// Get existing alarm strategies for this log_metric_group
	var existingStrategies []models.AlarmStrategyTable
	if err = x.SQL("select guid,name,metric from alarm_strategy where log_metric_group=?", param.LogMetricGroupGuid).Find(&existingStrategies); err != nil {
		return err
	}

	// Get log metric group info for context (query to ensure data consistency)
	var metricPrefixCode, logMetricMonitor, logMonitorTemplateGuid string
	x.SQL("select metric_prefix_code,log_metric_monitor,log_monitor_template from log_metric_group where guid=?", param.LogMetricGroupGuid).Get(&metricPrefixCode, &logMetricMonitor, &logMonitorTemplateGuid)
	var serviceGroup, monitorType string
	x.SQL("select service_group,monitor_type from log_metric_monitor where guid=?", logMetricMonitor).Get(&serviceGroup, &monitorType)

	// Get service group info and display name (consistent with autoGenerateAlarmStrategy)
	var serviceGroupTable models.ServiceGroupTable
	var displayServiceGroup = serviceGroup
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", serviceGroup).Get(&serviceGroupTable)
	if serviceGroupTable.DisplayName != "" {
		displayServiceGroup = serviceGroupTable.DisplayName
	}

	// Get service group roles for alarm notifications
	serviceGroupsRoles := getServiceGroupRoles(serviceGroup)

	// Handle roles fallback logic (same as autoGenerateAlarmStrategy)
	if len(serviceGroupsRoles) == 0 && len(roles) > 0 {
		serviceGroupsRoles = roles[:1]
	}

	log.Debug(nil, log.LOGGER_APP, "SyncAlarmStrategyForCodeChanges service group roles",
		zap.String("serviceGroup", serviceGroup),
		zap.Strings("roles", serviceGroupsRoles))

	// Get endpoint group for this service group
	var endpointGroup string
	var endpointGroupMonitorType string

	// Use the logic from autoGenerateAlarmStrategy to get endpoint group
	if serviceGroup != "" && monitorType != "" {
		var endpointGroupIds []string
		if err = x.SQL("select guid from endpoint_group where service_group=? and monitor_type=?", serviceGroup, monitorType).Find(&endpointGroupIds); err != nil {
			return err
		}
		if len(endpointGroupIds) > 0 {
			endpointGroup = endpointGroupIds[0]
		}
	}

	// Fallback: if no endpoint group found with service_group and monitor_type,
	// try to find one with service_group only, but this should be avoided
	if endpointGroup == "" {
		log.Warn(nil, log.LOGGER_APP, "SyncAlarmStrategyForCodeChanges no endpoint group found with service_group and monitor_type",
			zap.String("serviceGroup", serviceGroup),
			zap.String("monitorType", monitorType))
		// Don't use fallback as it may select wrong endpoint_group
		// Instead, we should create a proper endpoint_group or skip alarm creation
		return fmt.Errorf("no suitable endpoint_group found for service_group=%s and monitor_type=%s", serviceGroup, monitorType)
	}

	// Handle roles fallback logic (same as autoGenerateAlarmStrategy)
	if len(serviceGroupsRoles) == 0 && len(roles) > 0 {
		serviceGroupsRoles = roles[:1]
	}

	log.Debug(nil, log.LOGGER_APP, "SyncAlarmStrategyForCodeChanges service group roles",
		zap.String("serviceGroup", serviceGroup),
		zap.Strings("roles", serviceGroupsRoles))
	if endpointGroup == "" {
		// If no endpoint group exists for this service group, we need to create one or use a default
		// For now, we'll skip creating alarm strategies if no endpoint group exists
		log.Warn(nil, log.LOGGER_APP, "SyncAlarmStrategyForCodeChanges no endpoint group found for service group", zap.String("serviceGroup", serviceGroup))
		return nil
	}

	log.Debug(nil, log.LOGGER_APP, "SyncAlarmStrategyForCodeChanges found endpoint group",
		zap.String("serviceGroup", serviceGroup),
		zap.String("endpointGroup", endpointGroup),
		zap.String("monitorType", endpointGroupMonitorType))

	// Get retcode success value
	var retCodeStringMapRows []*models.LogMetricStringMapTable
	x.SQL("select * from log_metric_string_map where log_metric_group=? and log_param_name='retcode'", param.LogMetricGroupGuid).Find(&retCodeStringMapRows)
	sucCode := getRetCodeSuccessCode(retCodeStringMapRows)

	// Get log monitor template to get metric list
	logMonitorTemplateObj, getErr := GetLogMonitorTemplate(logMonitorTemplateGuid)
	if getErr != nil {
		return fmt.Errorf("failed to get log monitor template: %v", getErr)
	}

	// Get auto alarm metric list using template's metric list
	autoAlarmMetricList := getAutoAlarmMetricList(logMonitorTemplateObj.MetricList, serviceGroup, metricPrefixCode)

	// Handle code additions
	for _, code := range codesAdded {
		for _, alarmMetric := range autoAlarmMetricList {
			if !alarmMetric.AutoWarn {
				continue
			}

			// Create alarm strategy for the new code
			alarmStrategy := &models.GroupStrategyObj{
				NotifyList: make([]*models.NotifyObj, 0),
				Conditions: make([]*models.StrategyConditionObj, 0),
			}

			// Generate alarm strategy name following autoGenerateAlarmStrategy logic
			alarmStrategy.Name = fmt.Sprintf("%s-%s%s%s%s", code, generateMetricGuidDisplayName(metricPrefixCode, alarmMetric.Metric, displayServiceGroup), alarmMetric.Operator, alarmMetric.Threshold, getAlarmMetricUnit(alarmMetric.Metric))
			alarmStrategy.Priority = "high"
			alarmStrategy.NotifyEnable = 1
			alarmStrategy.ActiveWindow = "00:00-23:59"
			// Set EndpointGroup to the found endpoint group
			alarmStrategy.EndpointGroup = endpointGroup
			alarmStrategy.LogMetricGroup = &param.LogMetricGroupGuid
			alarmStrategy.Metric = alarmMetric.MetricId
			alarmStrategy.MetricName = alarmMetric.Metric
			alarmStrategy.Content = fmt.Sprintf("%s continuing for more than %s%s", alarmStrategy.Name, alarmMetric.Time, alarmMetric.TimeUnit)

			// Add notification configuration
			alarmStrategy.NotifyList = append(alarmStrategy.NotifyList, &models.NotifyObj{AlarmAction: "firing", NotifyRoles: serviceGroupsRoles})
			alarmStrategy.NotifyList = append(alarmStrategy.NotifyList, &models.NotifyObj{AlarmAction: "ok", NotifyRoles: serviceGroupsRoles})

			// Add metric tags
			metricTags := make([]*models.MetricTag, 0)
			for _, tag := range alarmMetric.TagConfig {
				if tag == constCode {
					metricTags = append(metricTags, &models.MetricTag{
						TagName:  constCode,
						Equal:    ConstEqualIn,
						TagValue: []string{code},
					})
				} else if tag == constRetCode && (strings.HasSuffix(alarmMetric.Metric, constConstTimeAvg) || strings.HasSuffix(alarmMetric.Metric, constConstTimeMax)) {
					metricTags = append(metricTags, &models.MetricTag{
						TagName:  constRetCode,
						Equal:    ConstEqualIn,
						TagValue: []string{sucCode},
					})
				} else {
					metricTags = append(metricTags, &models.MetricTag{
						TagName: tag,
					})
				}
			}

			alarmStrategy.Conditions = append(alarmStrategy.Conditions, &models.StrategyConditionObj{
				Metric:     alarmMetric.MetricId,
				MetricName: alarmMetric.Metric,
				Condition:  fmt.Sprintf("%s%s", alarmMetric.Operator, alarmMetric.Threshold),
				Last:       fmt.Sprintf("%s%s", alarmMetric.Time, alarmMetric.TimeUnit),
				Tags:       metricTags,
			})
			alarmStrategy.Condition = fmt.Sprintf("%s%s", alarmMetric.Operator, alarmMetric.Threshold)
			alarmStrategy.Last = fmt.Sprintf("%s%s", alarmMetric.Time, alarmMetric.TimeUnit)

			// Generate actions for creating alarm strategy
			now := time.Now().Format(models.DatetimeFormat)
			if subActions, createErr := getCreateAlarmStrategyActions(alarmStrategy, now, operator); createErr != nil {
				return createErr
			} else {
				actions = append(actions, subActions...)
			}
		}
	}

	// Handle code renames - update existing alarm strategies
	for oldCode, newCode := range codeRenames {
		// Update alarm strategy names and code tags
		for _, strategy := range existingStrategies {
			// Check if this strategy is for the old code (name starts with oldCode-)
			if strings.HasPrefix(strategy.Name, oldCode+"-") {
				newName := strings.Replace(strategy.Name, oldCode+"-", newCode+"-", 1)
				actions = append(actions, &Action{
					Sql:   "update alarm_strategy set name=? where guid=?",
					Param: []interface{}{newName, strategy.Guid},
				})

				// Update code tag values in alarm_strategy_tag_value
				actions = append(actions, &Action{
					Sql:   "update alarm_strategy_tag_value set value=? where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?) and name='code') and value=?",
					Param: []interface{}{newCode, strategy.Guid, oldCode},
				})
			}
		}
	}

	// Handle code deletions
	for _, delCode := range codeDeletes {
		// Delete alarm strategies for the deleted code
		for _, strategy := range existingStrategies {
			if strings.HasPrefix(strategy.Name, delCode+"-") {
				// Delete alarm strategy and all related records
				actions = append(actions, &Action{
					Sql:   "delete from alarm_strategy where guid=?",
					Param: []interface{}{strategy.Guid},
				})
			}
		}

		// Update other alarm strategies to remove the deleted code from NotIn list
		for _, strategy := range existingStrategies {
			if strings.HasPrefix(strategy.Name, constOtherCode+"-") {
				// Remove the deleted code from other strategies' NotIn list
				actions = append(actions, &Action{
					Sql:   "delete from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?) and name='code') and value=?",
					Param: []interface{}{strategy.Guid, delCode},
				})
			}
		}
	}

	// Handle other alarm strategies for code renames
	for oldCode, newCode := range codeRenames {
		for _, strategy := range existingStrategies {
			if strings.HasPrefix(strategy.Name, constOtherCode+"-") {
				// Update code in other strategies' NotIn list: remove old code, add new code
				actions = append(actions, &Action{
					Sql:   "delete from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?) and name='code') and value=?",
					Param: []interface{}{strategy.Guid, oldCode},
				})
				actions = append(actions, &Action{
					Sql:   "insert into alarm_strategy_tag_value(alarm_strategy_tag,value) select t.guid,? from alarm_strategy_tag t where t.alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?) and t.name='code' and not exists (select 1 from alarm_strategy_tag_value v where v.alarm_strategy_tag=t.guid and v.value=?)",
					Param: []interface{}{newCode, strategy.Guid, newCode},
				})
			}
		}
	}

	// Handle other alarm strategies for code additions
	if len(codesAdded) > 0 {
		for _, strategy := range existingStrategies {
			if strings.HasPrefix(strategy.Name, constOtherCode+"-") {
				// Add new codes to other strategies' NotIn list
				for _, code := range codesAdded {
					actions = append(actions, &Action{
						Sql:   "insert into alarm_strategy_tag_value(alarm_strategy_tag,value) select t.guid,? from alarm_strategy_tag t where t.alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?) and t.name='code' and not exists (select 1 from alarm_strategy_tag_value v where v.alarm_strategy_tag=t.guid and v.value=?)",
						Param: []interface{}{code, strategy.Guid, code},
					})
				}
			}
		}
	}

	if len(actions) > 0 {
		return Transaction(actions)
	}
	return nil
}
