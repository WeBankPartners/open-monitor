package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func ListLogMonitorTemplate(userRoles []string) (result *models.LogMonitorTemplateListResp, err error) {
	var rows []*models.LogMonitorTemplate
	err = x.SQL("select * from log_monitor_template order by update_time desc").Find(&rows)
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
		}
	}
	return
}

func ListLogMonitorTemplateOptions(userRoles []string) {

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
	err = x.SQL("select * from log_metric_template where log_monitor_template=?", logMonitorTemplateGuid).Find(&logMetricRows)
	if err != nil {
		err = fmt.Errorf("query log_metric_template table fail,%s ", err.Error())
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
	param.Guid = "lmt_" + guid.CreateGuid()
	nowTime := time.Now()
	var actions []*Action
	actions = append(actions, &Action{Sql: "insert into log_monitor_template(guid,name,log_type,json_regular,demo_log,calc_result,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		param.Guid, param.Name, param.LogType, param.JsonRegular, param.DemoLog, param.CalcResult, operator, operator, nowTime, nowTime,
	}})
	logParamGuidList := guid.CreateGuidList(len(param.ParamList))
	for i, logParamObj := range param.ParamList {
		actions = append(actions, &Action{Sql: "insert into log_param_template(guid,log_monitor_template,name,display_name,json_key,regular,demo_match_value,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lpt_" + logParamGuidList[i], param.Guid, logParamObj.Name, logParamObj.DisplayName, logParamObj.JsonKey, logParamObj.Regular, logParamObj.DemoMatchValue, operator, operator, nowTime, nowTime,
		}})
	}
	logMetricGuidList := guid.CreateGuidList(len(param.MetricList))
	for i, logMetricObj := range param.MetricList {
		actions = append(actions, &Action{Sql: "insert into log_metric_template(guid,log_monitor_template,log_param_name,metric,display_name,step,agg_type,tag_config,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			"lmet_" + logMetricGuidList[i], param.Guid, logMetricObj.LogParamName, logMetricObj.Metric, logMetricObj.DisplayName, logMetricObj.Step, logMetricObj.AggType, logMetricObj.TagConfig, operator, operator, nowTime, nowTime,
		}})
	}
	err = Transaction(actions)
	return
}

func UpdateLogMonitorTemplate(param *models.LogMonitorTemplateDto, operator string) (err error) {
	existLogMonitorObj, getExistDataErr := GetLogMonitorTemplate(param.Guid)
	if getExistDataErr != nil {
		err = fmt.Errorf("get exist log monitor data fail,%s ", getExistDataErr.Error())
		return
	}
	nowTime := time.Now()
	var actions []*Action
	actions = append(actions, &Action{Sql: "update log_monitor_template set name=?,json_regular=?,demo_log=?,calc_result=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
		param.Name, param.JsonRegular, param.DemoLog, param.CalcResult, operator, nowTime, param.Guid,
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
			actions = append(actions, &Action{Sql: "insert into log_metric_template(guid,log_monitor_template,log_param_name,metric,display_name,step,agg_type,tag_config,create_user,update_user,create_time,update_time) values (?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				"lmet_" + logMetricGuidList[i], param.Guid, logMetricObj.LogParamName, logMetricObj.Metric, logMetricObj.DisplayName, logMetricObj.Step, logMetricObj.AggType, logMetricObj.TagConfig, operator, operator, nowTime, nowTime,
			}})
		} else {
			actions = append(actions, &Action{Sql: "update log_metric_template set log_param_name=?,metric=?,display_name=?,step=?,agg_type=?,tag_config=?,update_user=?,update_time=? where guid=?", Param: []interface{}{
				logMetricObj.LogParamName, logMetricObj.Metric, logMetricObj.DisplayName, logMetricObj.Step, logMetricObj.AggType, logMetricObj.TagConfig, operator, nowTime, logMetricObj.Guid,
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
	err = Transaction(actions)
	return
}

func DeleteLogMonitorTemplate(logMonitorTemplateGuid string) (err error) {
	_, getErr := GetSimpleLogMonitorTemplate(logMonitorTemplateGuid)
	if getErr != nil {
		err = getErr
		return
	}
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from log_metric_template where log_monitor_template=?", Param: []interface{}{logMonitorTemplateGuid}})
	actions = append(actions, &Action{Sql: "delete from log_param_template where log_monitor_template=?", Param: []interface{}{logMonitorTemplateGuid}})
	actions = append(actions, &Action{Sql: "delete from log_monitor_template where guid=?", Param: []interface{}{logMonitorTemplateGuid}})
	err = Transaction(actions)
	return
}
