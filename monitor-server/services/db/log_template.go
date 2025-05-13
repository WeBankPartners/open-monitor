package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strings"
	"time"
)

func ListLogMonitorTemplateOptions() (result []string, err error) {
	err = x.SQL("select distinct update_user from log_monitor_template").Find(&result)
	return
}

func ListLogMonitorTemplate(param *models.LogMonitorTemplateListParam, userRoles []string) (result *models.LogMonitorTemplateListResp, err error) {
	var rows []*models.LogMonitorTemplate
	var filterSql string
	var filterParams []interface{}
	if param.Name != "" {
		filterSql += " and name like ?"
		filterParams = append(filterParams, fmt.Sprintf("%%%s%%", param.Name))
	}
	if param.UpdateUser != "" {
		filterSql += " and update_user like ?"
		filterParams = append(filterParams, fmt.Sprintf("%%%s%%", param.UpdateUser))
	}
	if len(filterParams) > 0 {
		err = x.SQL("select * from log_monitor_template where 1=1 "+filterSql+" order by update_time desc", filterParams...).Find(&rows)
	} else {
		err = x.SQL("select * from log_monitor_template order by update_time desc").Find(&rows)
	}
	if err != nil {
		err = fmt.Errorf("query log_monitor_template table fail,%s ", err.Error())
		return
	}
	result = &models.LogMonitorTemplateListResp{JsonList: []*models.LogMonitorTemplate{}, RegularList: []*models.LogMonitorTemplate{}}
	for _, row := range rows {
		row.CreateTimeString = row.CreateTime.Format(models.DatetimeFormat)
		row.UpdateTimeString = row.UpdateTime.Format(models.DatetimeFormat)
		if row.LogType == models.LogMonitorJsonType {
			result.JsonList = append(result.JsonList, row)
		} else if row.LogType == models.LogMonitorRegularType {
			result.RegularList = append(result.RegularList, row)
		} else if row.LogType == models.LogMonitorCustomType {
			result.CustomList = append(result.CustomList, row)
		}
	}
	return
}

func GetLogMonitorTemplate(logMonitorTemplateGuid string) (result *models.LogMonitorTemplateDto, err error) {
	logMonitorTemplateRow, getErr := GetSimpleLogMonitorTemplate(logMonitorTemplateGuid)
	if getErr != nil {
		err = getErr
		return
	}
	logMonitorTemplateRow.CreateTimeString = logMonitorTemplateRow.CreateTime.Format(models.DatetimeFormat)
	logMonitorTemplateRow.UpdateTimeString = logMonitorTemplateRow.UpdateTime.Format(models.DatetimeFormat)
	result = &models.LogMonitorTemplateDto{LogMonitorTemplate: *logMonitorTemplateRow, CalcResultObj: &models.CheckRegExpResult{},
		ParamList: []*models.LogParamTemplateObj{}, MetricList: []*models.LogMetricTemplate{},
		AutoCreateWarn: logMonitorTemplateRow.AutoAlarm == 1, AutoCreateDashboard: logMonitorTemplateRow.AutoDashboard == 1}
	result.LogMonitorTemplateVersion = logMonitorTemplateRow.UpdateTime.Format(models.DatetimeDigitFormat)
	if result.CalcResult != "" {
		if err = json.Unmarshal([]byte(result.CalcResult), result.CalcResultObj); err != nil {
			err = fmt.Errorf("json unmarhsal calc result:%s fail:%s ", result.CalcResult, err.Error())
			return
		}
	}
	var logParamRows []*models.LogParamTemplate
	err = x.SQL("select * from log_param_template where log_monitor_template=?", logMonitorTemplateGuid).Find(&logParamRows)
	if err != nil {
		err = fmt.Errorf("query log_param_template table fail,%s ", err.Error())
		return
	}
	var logMetricStringMaps []*models.LogMetricStringMapTable
	var stringCodeMap = make(map[string][]*models.LogMetricStringMapTable)
	err = x.SQL("select * from log_metric_string_map where log_monitor_template=?", logMonitorTemplateGuid).Find(&logMetricStringMaps)
	if err != nil {
		err = fmt.Errorf("query log_metric_string_map table fail,%s ", err.Error())
		return
	}
	for _, logMetricString := range logMetricStringMaps {
		if _, ok := stringCodeMap[logMetricString.LogParamName]; ok {
			stringCodeMap[logMetricString.LogParamName] = append(stringCodeMap[logMetricString.LogParamName], logMetricString)
		} else {
			stringCodeMap[logMetricString.LogParamName] = []*models.LogMetricStringMapTable{logMetricString}
		}
	}
	for _, row := range logParamRows {
		if row == nil {
			continue
		}
		logParamTemplateObj := &models.LogParamTemplateObj{
			LogParamTemplate: *row,
		}
		if v, ok := stringCodeMap[row.Name]; ok {
			logParamTemplateObj.StringMap = v
		}
		result.ParamList = append(result.ParamList, logParamTemplateObj)
	}
	var logMetricRows []*models.LogMetricTemplate
	if logMetricRows, err = getLogMetricTemplateWithMonitor(logMonitorTemplateGuid); err != nil {
		return
	}
	for _, row := range logMetricRows {
		row.TagConfigList = []string{}
		if row.TagConfig != "" {
			if err = json.Unmarshal([]byte(row.TagConfig), &row.TagConfigList); err != nil {
				err = fmt.Errorf("json unmarhsal metric tag config:%s fail,%s ", row.TagConfig, err.Error())
				return
			}
		}
		result.MetricList = append(result.MetricList, row)
	}
	return
}

func getLogMetricTemplateWithMonitor(logMonitorTemplateGuid string) (logMetricRows []*models.LogMetricTemplate, err error) {
	err = x.SQL("select * from log_metric_template where log_monitor_template=?", logMonitorTemplateGuid).Find(&logMetricRows)
	if err != nil {
		err = fmt.Errorf("query log_metric_template table fail,%s ", err.Error())
	}
	return
}

func GetSimpleLogMonitorTemplate(guid string) (logMonitorTemplateRow *models.LogMonitorTemplate, err error) {
	var rows []*models.LogMonitorTemplate
	err = x.SQL("select * from log_monitor_template where guid=?", guid).Find(&rows)
	if err != nil {
		err = fmt.Errorf("query log_monitor_template table fail,%s ", err.Error())
		return
	}
	if len(rows) == 0 {
		err = fmt.Errorf("can not find log_monitor_template with guid:%s ", guid)
		return
	}
	logMonitorTemplateRow = rows[0]
	return
}

func CreateLogMonitorTemplate(param *models.LogMonitorTemplateDto, operator string) (err error) {
	param.Guid = ""
	actions := getCreateLogMonitorTemplateActions(param, operator)
	err = Transaction(actions)
	return
}

func getCreateLogMonitorTemplateActions(param *models.LogMonitorTemplateDto, operator string) (actions []*Action) {
	if param.Guid == "" {
		param.Guid = "lmt_" + guid.CreateGuid()
	}
	if param.AutoCreateWarn {
		param.AutoAlarm = 1
	}
	if param.AutoCreateDashboard {
		param.AutoDashboard = 1
	}
	nowTime := time.Now()
	actions = append(actions, &Action{Sql: "insert into log_monitor_template(guid,name,log_type,json_regular,demo_log,calc_result,create_user,update_user,create_time,update_time,success_code,auto_alarm,auto_dashboard) values (?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		param.Guid, param.Name, param.LogType, param.JsonRegular, param.DemoLog, param.CalcResult, operator, operator, nowTime, nowTime, param.SuccessCode, param.AutoDashboard, param.AutoAlarm, param.AutoDashboard,
	}})
	logParamGuidList := guid.CreateGuidList(len(param.ParamList))
	for i, logParamObj := range param.ParamList {
		actions = append(actions, &Action{Sql: "insert into log_param_template(guid,log_monitor_template,name,display_name,json_key,regular,demo_match_value,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lpt_" + logParamGuidList[i], param.Guid, logParamObj.Name, logParamObj.DisplayName, logParamObj.JsonKey, logParamObj.Regular, logParamObj.DemoMatchValue, operator, operator, nowTime, nowTime,
		}})
		tmpStringMapGuidList := guid.CreateGuidList(len(logParamObj.StringMap))
		for stringMapIndex, stringMapObj := range logParamObj.StringMap {
			actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_monitor_template,log_param_name,value_type,source_value,regulative,target_value,update_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmsm_" + tmpStringMapGuidList[stringMapIndex], param.Guid, logParamObj.Name, stringMapObj.ValueType, stringMapObj.SourceValue, stringMapObj.Regulative, stringMapObj.TargetValue, nowTime.Format(models.DatetimeFormat),
			}})
		}
	}
	logMetricGuidList := guid.CreateGuidList(len(param.MetricList))
	for i, logMetricObj := range param.MetricList {
		if logMetricObj.TagConfig == "" {
			tmpTagConfigBytes, _ := json.Marshal(logMetricObj.TagConfigList)
			logMetricObj.TagConfig = string(tmpTagConfigBytes)
		}
		actions = append(actions, &Action{Sql: "insert into log_metric_template(guid,log_monitor_template,log_param_name,metric,display_name,step,agg_type,tag_config,create_user,update_user,create_time,update_time,color_group,auto_alarm,range_config) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lmet_" + logMetricGuidList[i], param.Guid, logMetricObj.LogParamName, logMetricObj.Metric, logMetricObj.DisplayName, logMetricObj.Step, logMetricObj.AggType, logMetricObj.TagConfig, operator, operator, nowTime, nowTime, logMetricObj.ColorGroup, logMetricObj.AutoAlarm, logMetricObj.RangeConfig,
		}})
	}
	return
}

func UpdateLogMonitorTemplate(param *models.LogMonitorTemplateDto, operator string) (affectEndpoints []string, err error) {
	var actions []*Action
	actions, affectEndpoints, err = getUpdateLogMonitorTemplateActions(param, operator)
	if err != nil {
		return
	}
	err = Transaction(actions)
	return
}

func getUpdateLogMonitorTemplateActions(param *models.LogMonitorTemplateDto, operator string) (actions []*Action, affectEndpoints []string, err error) {
	if param.AutoCreateWarn {
		param.AutoAlarm = 1
	}
	if param.AutoCreateDashboard {
		param.AutoDashboard = 1
	}
	existLogMonitorObj, getExistDataErr := GetLogMonitorTemplate(param.Guid)
	if getExistDataErr != nil {
		err = fmt.Errorf("get exist log monitor data fail,%s ", getExistDataErr.Error())
		return
	}
	nowTime := time.Now()
	actions = append(actions, &Action{Sql: "update log_monitor_template set name=?,json_regular=?,demo_log=?,calc_result=?,update_user=?,update_time=?,success_code=?,auto_alarm=?,auto_dashboard=? where guid=?", Param: []interface{}{
		param.Name, param.JsonRegular, param.DemoLog, param.CalcResult, operator, nowTime, param.SuccessCode, param.AutoAlarm, param.AutoDashboard, param.Guid,
	}})
	logParamGuidList := guid.CreateGuidList(len(param.ParamList))
	for i, logParamObj := range param.ParamList {
		if logParamObj.Guid == "" {
			actions = append(actions, &Action{Sql: "insert into log_param_template(guid,log_monitor_template,name,display_name,json_key,regular,demo_match_value,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lpt_" + logParamGuidList[i], param.Guid, logParamObj.Name, logParamObj.DisplayName, logParamObj.JsonKey, logParamObj.Regular, logParamObj.DemoMatchValue, operator, operator, nowTime, nowTime,
			}})
		} else {
			actions = append(actions, &Action{Sql: "update log_param_template set name=?,display_name=?,json_key=?,regular=?,demo_match_value=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
				logParamObj.Name, logParamObj.DisplayName, logParamObj.JsonKey, logParamObj.Regular, logParamObj.DemoMatchValue, operator, nowTime, logParamObj.Guid,
			}})
		}
		tmpStringMapGuidList := guid.CreateGuidList(len(logParamObj.StringMap))
		for stringMapIndex, stringMapObj := range logParamObj.StringMap {
			actions = append(actions, &Action{Sql: "insert into log_metric_string_map(guid,log_monitor_template,log_param_name,value_type,source_value,regulative,target_value,update_time) values (?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmsm_" + tmpStringMapGuidList[stringMapIndex], param.Guid, logParamObj.Name, stringMapObj.ValueType, stringMapObj.SourceValue, stringMapObj.Regulative, stringMapObj.TargetValue, nowTime.Format(models.DatetimeFormat),
			}})
		}
	}
	for _, existParamObj := range existLogMonitorObj.ParamList {
		deleteFlag := true
		for _, inputLogParamObj := range param.ParamList {
			if existParamObj.Guid == inputLogParamObj.Guid {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			actions = append(actions, &Action{Sql: "delete from log_param_template where guid=?", Param: []interface{}{existParamObj.Guid}})
		}
		for _, stringMapObj := range existParamObj.StringMap {
			actions = append(actions, &Action{Sql: "delete from log_metric_string_map where guid=?", Param: []interface{}{stringMapObj.Guid}})
		}
	}
	logMetricGuidList := guid.CreateGuidList(len(param.MetricList))
	for i, logMetricObj := range param.MetricList {
		if logMetricObj.Guid == "" {
			actions = append(actions, &Action{Sql: "insert into log_metric_template(guid,log_monitor_template,log_param_name,metric,display_name,step,agg_type,tag_config,create_user,update_user,create_time,update_time,color_group,auto_alarm,range_config) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmet_" + logMetricGuidList[i], param.Guid, logMetricObj.LogParamName, logMetricObj.Metric, logMetricObj.DisplayName, logMetricObj.Step, logMetricObj.AggType, logMetricObj.TagConfig, operator, operator, nowTime, nowTime, logMetricObj.ColorGroup, logMetricObj.AutoAlarm, logMetricObj.RangeConfig,
			}})
		} else {
			actions = append(actions, &Action{Sql: "update log_metric_template set log_param_name=?,metric=?,display_name=?,step=?,agg_type=?,tag_config=?,update_user=?,update_time=?,color_group=?,auto_alarm=?,range_config=? where guid=?", Param: []interface{}{
				logMetricObj.LogParamName, logMetricObj.Metric, logMetricObj.DisplayName, logMetricObj.Step, logMetricObj.AggType, logMetricObj.TagConfig, operator, nowTime, logMetricObj.ColorGroup, logMetricObj.AutoAlarm, logMetricObj.RangeConfig, logMetricObj.Guid,
			}})
		}
	}
	for _, existMetricObj := range existLogMonitorObj.MetricList {
		deleteFlag := true
		for _, inputLogMetricObj := range param.MetricList {
			if existMetricObj.Guid == inputLogMetricObj.Guid {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			actions = append(actions, &Action{Sql: "delete from log_metric_template where guid=?", Param: []interface{}{existMetricObj.Guid}})
		}
	}
	var endpointRelRows []*models.LogMetricEndpointRelTable
	queryEndpointErr := x.SQL("select source_endpoint from log_metric_endpoint_rel where log_metric_monitor in (select log_metric_monitor from log_metric_group where log_monitor_template=?)", param.Guid).Find(&endpointRelRows)
	if queryEndpointErr != nil {
		log.Error(nil, log.LOGGER_APP, "query log metric template affect endpoints fail", zap.String("logMonitorTemplate", param.Guid), zap.Error(queryEndpointErr))
	} else {
		for _, v := range endpointRelRows {
			affectEndpoints = append(affectEndpoints, v.SourceEndpoint)
		}
	}
	return
}

func DeleteLogMonitorTemplate(logMonitorTemplateGuid string, errMsgObj *models.ErrorTemplate) (err error) {
	var actions []*Action
	if actions, err = getDeleteLogMonitorTemplateActions(logMonitorTemplateGuid, errMsgObj); err != nil {
		return
	}
	err = Transaction(actions)
	return
}

func getDeleteLogMonitorTemplateActions(logMonitorTemplateGuid string, errMsgObj *models.ErrorTemplate) (actions []*Action, err error) {
	var logMetricGroupNameList []string
	if _, err = GetSimpleLogMonitorTemplate(logMonitorTemplateGuid); err != nil {
		return
	}
	if logMetricGroupNameList, err = GetLogMetricGroupNameByLogMonitorTemplate(logMonitorTemplateGuid); err != nil {
		return
	}
	if len(logMetricGroupNameList) > 0 {
		err = errMsgObj.LogMonitorTemplateDeleteError.WithParam(strings.Join(logMetricGroupNameList, ","))
		return
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_template where log_monitor_template=?", Param: []interface{}{logMonitorTemplateGuid}})
	actions = append(actions, &Action{Sql: "delete from log_param_template where log_monitor_template=?", Param: []interface{}{logMonitorTemplateGuid}})
	actions = append(actions, &Action{Sql: "delete from log_monitor_template where guid=?", Param: []interface{}{logMonitorTemplateGuid}})
	return
}

func GetLogMonitorTemplateServiceGroup(logMonitorTemplateGuid string) (result []*models.ServiceGroupTable, err error) {
	err = x.SQL("select * from service_group where guid in (select service_group from log_metric_monitor where guid in (select log_metric_monitor from log_metric_group where log_monitor_template=?))", logMonitorTemplateGuid).Find(&result)
	if err != nil {
		err = fmt.Errorf("query service_group table fail,%s ", err.Error())
	}
	return
}

func GetLogMetricGroupNameByLogMonitorTemplate(logMonitorTemplateGuid string) (result []string, err error) {
	err = x.SQL(" select name from log_metric_group where log_monitor_template=?", logMonitorTemplateGuid).Find(&result)
	return
}

func GetLogMonitorTemplateByName(guid, name string) (logMonitorTemplate *models.LogMonitorTemplate, err error) {
	var logMonitorTemplateRows []*models.LogMonitorTemplate
	err = x.SQL("select * from log_monitor_template where name=? and guid!=?", name, guid).Find(&logMonitorTemplateRows)
	if err != nil {
		err = fmt.Errorf("query log monitor template table fail,%s ", err.Error())
		return
	}
	if len(logMonitorTemplateRows) > 0 {
		logMonitorTemplate = logMonitorTemplateRows[0]
	}
	return
}

func GetLogMonitorTemplateById(guid string) (logMonitorTemplate *models.LogMonitorTemplate, err error) {
	var logMonitorTemplateRows []*models.LogMonitorTemplate
	err = x.SQL("select * from log_monitor_template where guid=?", guid).Find(&logMonitorTemplateRows)
	if err != nil {
		err = fmt.Errorf("query log monitor template table fail,%s ", err.Error())
		return
	}
	if len(logMonitorTemplateRows) > 0 {
		logMonitorTemplate = logMonitorTemplateRows[0]
	}
	return
}

func ImportLogMonitorTemplate(params []*models.LogMonitorTemplateDto, operator string) (affectEndpoints []string, err error) {
	var actions []*Action
	var existLogMonitorTemplate *models.LogMonitorTemplate
	for _, inputParam := range params {
		if existLogMonitorTemplate, err = GetLogMonitorTemplateById(inputParam.Guid); err != nil {
			return
		}
		if existLogMonitorTemplate != nil {
			inputParam.Guid = ""
		}
		// 看下ID 能否复用
		if existLogMonitorTemplate, err = GetLogMonitorTemplateById(inputParam.Guid); err != nil {
			return
		}
		if existLogMonitorTemplate != nil {
			inputParam.Guid = ""
		}
		// 看下名称能否复用
		if existLogMonitorTemplate, err = GetLogMonitorTemplateByName("", inputParam.Name); err != nil {
			return
		}
		if existLogMonitorTemplate != nil {
			// 名称已有,追加名称
			inputParam.Name = fmt.Sprintf("%s(1)", inputParam.Name)
			if existLogMonitorTemplate, err = GetLogMonitorTemplateByName("", inputParam.Name); err != nil {
				return
			}
			if existLogMonitorTemplate != nil {
				err = fmt.Errorf("log monitor template name:%s duplicate", inputParam.Name)
				return
			}
		}
		calcResultBytes, _ := json.Marshal(inputParam.CalcResultObj)
		inputParam.CalcResult = string(calcResultBytes)
		tmpActions := getCreateLogMonitorTemplateActions(inputParam, operator)
		actions = append(actions, tmpActions...)
	}
	if err = Transaction(actions); err != nil {
		affectEndpoints = distinctStringList(affectEndpoints)
	}
	return
}

func distinctStringList(input []string) (output []string) {
	dMap := make(map[string]int)
	for _, v := range input {
		if _, ok := dMap[v]; ok {
			continue
		}
		output = append(output, v)
		dMap[v] = 1
	}
	return
}

func GetLogTemplateGuidByName(name string) (logTemplateGuid string, err error) {
	queryResult, queryErr := x.QueryString("select guid from log_monitor_template where name=?", name)
	if queryErr != nil {
		err = fmt.Errorf("query log template by name:%s fail,%s ", name, err.Error())
		return
	}
	if len(queryResult) == 0 {
		err = fmt.Errorf("can not find template with name:%s ", name)
		return
	}
	logTemplateGuid = queryResult[0]["guid"]
	return
}

func GetLogMetricGroupById(id string) (result *models.LogMetricGroup, err error) {
	var list []*models.LogMetricGroup
	if err = x.SQL("select * from log_metric_group where guid=?", id).Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		result = list[0]
	}
	return
}

func BatchGetLogTemplateByGuid(ids []string) (list []*models.LogMonitorTemplate, err error) {
	err = x.SQL(fmt.Sprintf("select name,log_type,update_user,update_time from log_monitor_template where  guid in ('%s')", strings.Join(ids, "','"))).Find(&list)
	for _, logMonitorTemplate := range list {
		logMonitorTemplate.UpdateTimeString = logMonitorTemplate.UpdateTime.Format(models.DatetimeFormat)
	}
	return
}
