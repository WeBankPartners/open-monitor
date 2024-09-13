package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
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
	result = &models.LogMonitorTemplateDto{LogMonitorTemplate: *logMonitorTemplateRow, CalcResultObj: &models.CheckRegExpResult{}, ParamList: []*models.LogParamTemplate{}, MetricList: []*models.LogMetricTemplate{}}
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
	for _, row := range logParamRows {
		result.ParamList = append(result.ParamList, row)
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
	nowTime := time.Now()
	actions = append(actions, &Action{Sql: "insert into log_monitor_template(guid,name,log_type,json_regular,demo_log,calc_result,create_user,update_user,create_time,update_time,success_code) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		param.Guid, param.Name, param.LogType, param.JsonRegular, param.DemoLog, param.CalcResult, operator, operator, nowTime, nowTime, param.SuccessCode,
	}})
	logParamGuidList := guid.CreateGuidList(len(param.ParamList))
	for i, logParamObj := range param.ParamList {
		actions = append(actions, &Action{Sql: "insert into log_param_template(guid,log_monitor_template,name,display_name,json_key,regular,demo_match_value,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lpt_" + logParamGuidList[i], param.Guid, logParamObj.Name, logParamObj.DisplayName, logParamObj.JsonKey, logParamObj.Regular, logParamObj.DemoMatchValue, operator, operator, nowTime, nowTime,
		}})
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
	existLogMonitorObj, getExistDataErr := GetLogMonitorTemplate(param.Guid)
	if getExistDataErr != nil {
		err = fmt.Errorf("get exist log monitor data fail,%s ", getExistDataErr.Error())
		return
	}
	nowTime := time.Now()
	actions = append(actions, &Action{Sql: "update log_monitor_template set name=?,json_regular=?,demo_log=?,calc_result=?,update_user=?,update_time=?,success_code=? where guid=?", Param: []interface{}{
		param.Name, param.JsonRegular, param.DemoLog, param.CalcResult, operator, nowTime, param.SuccessCode, param.Guid,
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
		log.Logger.Error("query log metric template affect endpoints fail", log.String("logMonitorTemplate", param.Guid), log.Error(queryEndpointErr))
	} else {
		for _, v := range endpointRelRows {
			affectEndpoints = append(affectEndpoints, v.SourceEndpoint)
		}
	}
	return
}

func DeleteLogMonitorTemplate(logMonitorTemplateGuid string, errMsgObj *models.ErrorMessageObj) (err error) {
	var actions []*Action
	if actions, err = getDeleteLogMonitorTemplateActions(logMonitorTemplateGuid, errMsgObj); err != nil {
		return
	}
	err = Transaction(actions)
	return
}

func getDeleteLogMonitorTemplateActions(logMonitorTemplateGuid string, errMsgObj *models.ErrorMessageObj) (actions []*Action, err error) {
	var logMetricGroupNameList []string
	if _, err = GetSimpleLogMonitorTemplate(logMonitorTemplateGuid); err != nil {
		return
	}
	if logMetricGroupNameList, err = GetLogMetricGroupNameByLogMonitorTemplate(logMonitorTemplateGuid); err != nil {
		return
	}
	if len(logMetricGroupNameList) > 0 {
		err = fmt.Errorf(errMsgObj.LogMonitorTemplateDeleteError, strings.Join(logMetricGroupNameList, ","))
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

func ImportLogMonitorTemplate(params []*models.LogMonitorTemplateDto, operator string) (affectEndpoints []string, err error) {
	var existTemplateRows []*models.LogMonitorTemplate
	err = x.SQL("select guid,name from log_monitor_template").Find(&existTemplateRows)
	if err != nil {
		err = fmt.Errorf("query log monitor template table fail,%s ", err.Error())
		return
	}
	var actions []*Action
	for _, inputParam := range params {
		inputParam.Guid = ""
		inputParam.Name = fmt.Sprintf("%s(1)", inputParam.Name)
		if existLogMonitorTemplate, getErr := GetLogMonitorTemplateByName("", inputParam.Name); getErr != nil {
			err = getErr
			return
		} else if existLogMonitorTemplate != nil {
			err = fmt.Errorf("log monitor template name:%s duplicate", inputParam.Name)
			return
		}
		calcResultBytes, _ := json.Marshal(inputParam.CalcResultObj)
		inputParam.CalcResult = string(calcResultBytes)
		tmpActions := getCreateLogMonitorTemplateActions(inputParam, operator)
		actions = append(actions, tmpActions...)
		//addFlag := true
		//for _, existRow := range existTemplateRows {
		//	if existRow.Guid == inputParam.Guid {
		//		addFlag = false
		//		break
		//	}
		//}
		//if addFlag {
		//	tmpActions := getCreateLogMonitorTemplateActions(inputParam, operator)
		//	actions = append(actions, tmpActions...)
		//} else {
		//	tmpActions, tmpAffect, tmpErr := getUpdateLogMonitorTemplateActions(inputParam, operator)
		//	if tmpErr != nil {
		//		err = tmpErr
		//		return
		//	}
		//	actions = append(actions, tmpActions...)
		//	affectEndpoints = append(affectEndpoints, tmpAffect...)
		//}
	}
	err = Transaction(actions)
	if err == nil {
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
