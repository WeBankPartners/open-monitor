package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"time"
)

func MetricList(id string, endpointType, serviceGroup string) (result []*models.PromMetricTable, err error) {
	params := []interface{}{}
	baseSql := "select guid as id,metric,monitor_type as metric_type,prom_expr as prom_ql from metric where 1=1 "
	if id != "" {
		baseSql += " and guid=? "
		params = append(params, id)
	}
	if endpointType != "" {
		baseSql += " and monitor_type=? "
		params = append(params, endpointType)
	}
	if serviceGroup != "" {
		baseSql += " and service_group=? "
		params = append(params, serviceGroup)
	}
	result = []*models.PromMetricTable{}
	err = x.SQL(baseSql, params...).Find(&result)
	//if err != nil {
	//	return
	//}
	// append service metric
	//var logMetricTable []*models.LogMetricConfigTable
	//x.SQL("select guid,metric,display_name,agg_type from log_metric_config where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?) or log_metric_json in (select guid from log_metric_json where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?))", endpointType, endpointType).Find(&logMetricTable)
	//for _, v := range logMetricTable {
	//	result = append(result, &models.PromMetricTable{Id: 0, MetricType: endpointType, Metric: v.Metric, PromQl: fmt.Sprintf("%s{key=\"%s\",agg=\"%s\",t_endpoint=\"$guid\"}", models.LogMetricName, v.Metric, v.AggType)})
	//}
	//var dbMetricTable []*models.DbMetricMonitorTable
	//x.SQL("select guid,metric,display_name from db_metric_monitor where monitor_type=?", endpointType).Find(&dbMetricTable)
	//for _, v := range dbMetricTable {
	//	result = append(result, &models.PromMetricTable{Id: 0, MetricType: endpointType, Metric: v.Metric, PromQl: fmt.Sprintf("%s{key=\"%s\",t_endpoint=\"$guid\"}", models.DBMonitorMetricName, v.Metric)})
	//}
	return
}

func MetricCreate(param []*models.MetricTable, operator string) error {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	for _, metric := range param {
		//actions = append(actions, &Action{Sql: "insert into prom_metric(metric,metric_type,prom_ql) value (?,?,?)", Param: []interface{}{metric.Metric, metric.MetricType, metric.PromQl}})
		if metric.ServiceGroup != "" {
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{fmt.Sprintf("%s__%s", metric.Metric, metric.MonitorType), metric.Metric, metric.MonitorType, metric.PromExpr, metric.ServiceGroup, metric.Workspace, nowTime, nowTime, operator, operator}})
		} else {
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,update_time,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?)",
				Param: []interface{}{fmt.Sprintf("%s__%s", metric.Metric, metric.MonitorType), metric.Metric, metric.MonitorType, metric.PromExpr, nowTime, nowTime, operator, operator}})
		}
	}
	return Transaction(actions)
}

func MetricUpdate(param []*models.MetricTable, operator string) (err error) {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	var metricGuidList []string
	for _, metric := range param {
		if metric.Guid == "" {
			err = fmt.Errorf("Guid can not empty ")
			break
		}
		if metric.ServiceGroup != "" {
			actions = append(actions, &Action{Sql: "update metric set prom_expr=?,service_group=?,workspace=?,update_user=?,update_time=? where guid=?", Param: []interface{}{metric.PromExpr, metric.ServiceGroup, metric.Workspace, operator, nowTime, metric.Guid}})
		} else {
			actions = append(actions, &Action{Sql: "update metric set prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{metric.PromExpr, operator, nowTime, metric.Guid}})
		}
		metricGuidList = append(metricGuidList, metric.Guid)
	}
	if err != nil {
		return err
	}
	err = Transaction(actions)
	if err != nil {
		return err
	}
	var alarmStrategyTable []*models.AlarmStrategyTable
	err = x.SQL("select distinct endpoint_group from alarm_strategy where metric in ('" + strings.Join(metricGuidList, "','") + "')").Find(&alarmStrategyTable)
	if len(alarmStrategyTable) > 0 {
		for _, v := range alarmStrategyTable {
			err = SyncPrometheusRuleFile(v.EndpointGroup, false)
			if err != nil {
				break
			}
		}
	}
	return err
}

func getMetricUpdateAction(oldGuid, operator string, newMetricObj *models.MetricTable) (actions []*Action) {
	actions = []*Action{}
	if newMetricObj.Guid != oldGuid {
		actions = append(actions, &Action{Sql: "update metric set guid=?,metric=?,monitor_type=?,prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{newMetricObj.Guid, newMetricObj.Metric, newMetricObj.MonitorType, newMetricObj.PromExpr, operator, newMetricObj.UpdateTime, oldGuid}})
		actions = append(actions, &Action{Sql: "update alarm_strategy set metric=? where metric=?", Param: []interface{}{newMetricObj.Guid, oldGuid}})
	} else {
		actions = append(actions, &Action{Sql: "update metric set metric=?,monitor_type=?,prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{newMetricObj.Metric, newMetricObj.MonitorType, newMetricObj.PromExpr, operator, newMetricObj.UpdateTime, oldGuid}})
	}
	return actions
}

func MetricDelete(id string) error {
	metricQuery, err := MetricList(id, "", "")
	if err != nil {
		return fmt.Errorf("Try to query prom metric table fail,%s ", err.Error())
	}
	if len(metricQuery) == 0 {
		return nil
	}
	metric := metricQuery[0].Metric
	var actions []*Action
	//actions = append(actions, &Action{Sql: "delete from prom_metric where id=?", Param: []interface{}{id}})
	actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{id}})
	var charts []*models.ChartTable
	err = x.SQL("select id,metric from chart where metric like ? and group_id in (select chart_group from panel where group_id in (select panels_group from dashboard where dashboard_type=?))", "%"+metric+"%", metricQuery[0].MetricType).Find(&charts)
	if err != nil {
		return fmt.Errorf("Try to get charts data fail,%s ", err.Error())
	}
	if len(charts) > 0 {
		for _, chart := range charts {
			newChartMetricList := []string{}
			for _, v := range strings.Split(chart.Metric, "^") {
				if v == metric || v == "" {
					continue
				}
				newChartMetricList = append(newChartMetricList, v)
			}
			if len(newChartMetricList) == 0 {
				actions = append(actions, &Action{Sql: "delete from chart where id=?", Param: []interface{}{chart.Id}})
			} else {
				actions = append(actions, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newChartMetricList, "^"), chart.Id}})
			}
		}
	}
	err = Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update database fail,%s ", err.Error())
	}
	return err
}

func MetricListNew(guid, monitorType, serviceGroup, onlyService string) (result []*models.MetricTable, err error) {
	params := []interface{}{}
	baseSql := "select * from metric where 1=1 "
	if guid != "" {
		baseSql += " and guid=? "
		params = append(params, guid)
	} else {
		if serviceGroup != "" {
			if monitorType == "" {
				return result, fmt.Errorf("serviceGroup is disable when monitorType is null ")
			}
			if onlyService == "Y" {
				baseSql = "select * from metric where monitor_type=? and service_group=?"
				params = []interface{}{monitorType, serviceGroup}
			} else {
				baseSql = "select * from metric where monitor_type=? and (service_group is null or service_group=?)"
				params = []interface{}{monitorType, serviceGroup}
			}
		} else {
			baseSql = "select * from metric where monitor_type=? and service_group is null"
			params = []interface{}{monitorType}
		}
	}
	result = []*models.MetricTable{}
	err = x.SQL(baseSql, params...).Find(&result)
	if err != nil {
		return
	}
	for _, metric := range result {
		if strings.TrimSpace(metric.ServiceGroup) == "" {
			metric.MetricType = string(models.MetricTypeCommon)
		} else if strings.TrimSpace(metric.LogMetricGroup) != "" {
			metric.MetricType = string(models.MetricTypeBusiness)
			if serviceGroup != "" {
				var name string
				if _, err = x.SQL("select name from log_metric_group where guid = ?", metric.LogMetricGroup).Get(&name); err != nil {
					return
				}
				metric.LogMetricGroupName = name
			}
		} else {
			metric.MetricType = string(models.MetricTypeCustom)
		}
	}
	return
}

func MetricImport(serviceGroup, operator string, inputMetrics []*models.MetricTable) (err error) {
	existMetrics, getExistErr := MetricListNew("", inputMetrics[0].MonitorType, serviceGroup, "Y")
	if getExistErr != nil {
		return fmt.Errorf("get serviceGroup:%s exist metric list fail,%s ", serviceGroup, getExistErr.Error())
	}
	var alarmStrategyRows []*models.AlarmStrategyTable
	if err = x.SQL("select endpoint_group,metric from alarm_strategy").Find(&alarmStrategyRows); err != nil {
		return fmt.Errorf("query alarm strategy fail,%s ", err.Error())
	}
	oldServerGroup := inputMetrics[0].ServiceGroup
	strategyMap := make(map[string]string)
	for _, v := range alarmStrategyRows {
		strategyMap[v.Metric] = v.EndpointGroup
	}
	var actions []*Action
	var affectEndpointGroupList []string
	nowTime := time.Now().Format(models.DatetimeFormat)
	for _, inputMetric := range inputMetrics {
		inputMetric.Guid = fmt.Sprintf("%s__%s", inputMetric.Metric, serviceGroup)
		matchMetric := &models.MetricTable{}
		for _, existMetric := range existMetrics {
			if existMetric.Guid == inputMetric.Guid {
				matchMetric = existMetric
				break
			}
		}
		inputMetric.PromExpr = strings.ReplaceAll(inputMetric.PromExpr, oldServerGroup, serviceGroup)
		if matchMetric.Guid != "" {
			if v, b := strategyMap[matchMetric.Guid]; b {
				affectEndpointGroupList = append(affectEndpointGroupList, v)
			}
			actions = append(actions, &Action{Sql: "update metric set prom_expr=?,workspace=?,update_user=?,update_time=? where guid=?", Param: []interface{}{inputMetric.PromExpr, inputMetric.Workspace, operator, nowTime, matchMetric.Guid}})
		} else {
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{inputMetric.Guid, inputMetric.Metric, inputMetric.MonitorType, inputMetric.PromExpr, serviceGroup, inputMetric.Workspace, nowTime, nowTime, operator, operator}})
		}
	}
	for _, existMetric := range existMetrics {
		deleteFlag := true
		for _, inputMetric := range inputMetrics {
			if inputMetric.Metric == existMetric.Metric {
				deleteFlag = false
				break
			}
		}
		if deleteFlag {
			if v, b := strategyMap[existMetric.Guid]; b {
				affectEndpointGroupList = append(affectEndpointGroupList, v)
				actions = append(actions, &Action{Sql: "delete from alarm_strategy where metric=?", Param: []interface{}{existMetric.Guid}})
			}
			actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{existMetric.Guid}})
		}
	}
	log.Logger.Info("import metric", log.Int("actionLen", len(actions)))
	if len(actions) > 0 {
		if err = Transaction(actions); err != nil {
			return fmt.Errorf("import metric fail with exec database,%s ", err.Error())
		}
	}
	if len(affectEndpointGroupList) > 0 {
		egMap := make(map[string]int)
		for _, v := range affectEndpointGroupList {
			if _, b := egMap[v]; b {
				continue
			}
			egMap[v] = 1
			if tmpErr := SyncPrometheusRuleFile(v, false); tmpErr != nil {
				log.Logger.Error("sync prometheus endpoint group fail", log.Error(tmpErr))
			}
		}
	}
	return
}

func GetSimpleMetric(metricId string) (metricRow *models.MetricTable, err error) {
	var metricRows []*models.MetricTable
	err = x.SQL("select * from metric where guid=?", metricId).Find(&metricRows)
	if err != nil {
		err = fmt.Errorf("query metric table with guid:%s fail,%s ", metricId, err.Error())
		return
	}
	if len(metricRows) == 0 {
		err = fmt.Errorf("can not find metric with guid:%s ", metricId)
	} else {
		metricRow = metricRows[0]
	}
	return
}

func GetMetricTags(metricRow *models.MetricTable) (tags []string, err error) {
	if metricRow == nil {
		return
	}
	if metricRow.LogMetricConfig != "" {
		var logMetricConfigRows []*models.LogMetricConfigTable
		err = x.SQL("select metric,tag_config from log_metric_config where guid=?", metricRow.LogMetricConfig).Find(&logMetricConfigRows)
		if err != nil {
			err = fmt.Errorf("query log metric config with guid:%s fail,%s ", metricRow.LogMetricConfig, err.Error())
		}
		if len(logMetricConfigRows) > 0 {
			tagConfigString := logMetricConfigRows[0].TagConfig
			if tagConfigString != "" && tagConfigString != "null" && tagConfigString != "[]" {
				if err = json.Unmarshal([]byte(logMetricConfigRows[0].TagConfig), &tags); err != nil {
					err = fmt.Errorf("json unmarshal tag config: %s fail,%s ", tagConfigString, err.Error())
				}
			}
		}
		return
	}
	if metricRow.LogMetricTemplate != "" {
		var logMetricTemplateRows []*models.LogMetricTemplate
		err = x.SQL("select metric,tag_config from log_metric_template where guid=?", metricRow.LogMetricTemplate).Find(&logMetricTemplateRows)
		if err != nil {
			err = fmt.Errorf("query log metric template with guid:%s fail,%s ", metricRow.LogMetricTemplate, err.Error())
		}
		if len(logMetricTemplateRows) > 0 {
			tagConfigString := logMetricTemplateRows[0].TagConfig
			if tagConfigString != "" && tagConfigString != "null" && tagConfigString != "[]" {
				if err = json.Unmarshal([]byte(logMetricTemplateRows[0].TagConfig), &tags); err != nil {
					err = fmt.Errorf("json unmarshal tag config: %s fail,%s ", tagConfigString, err.Error())
				}
			}
		}
		return
	}
	return
}
