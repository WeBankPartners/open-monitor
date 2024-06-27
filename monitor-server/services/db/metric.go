package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"regexp"
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

func MetricComparisonListNew(guid, monitorType, serviceGroup, onlyService string) (result []*models.MetricComparisonExtend, err error) {
	var params []interface{}
	baseSql := "select m.*,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.metric_id as metric_id from metric m join metric_comparison mc on mc.metric_id = m.guid "
	if guid != "" {
		baseSql += " and m.guid=? "
		params = append(params, guid)
	} else {
		if serviceGroup != "" {
			if monitorType == "" {
				return result, fmt.Errorf("serviceGroup is disable when monitorType is null ")
			}
			if onlyService == "Y" {
				baseSql = "select m.*,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.metric_id as metric_id from metric m join metric_comparison mc on mc.metric_id = m.guid amd m.monitor_type=? and m.service_group=?"
				params = []interface{}{monitorType, serviceGroup}
			} else {
				baseSql = "select m.*,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.metric_id as metric_id from metric m join metric_comparison mc on mc.metric_id = m.guid  and m.monitor_type=? and (m.service_group is null or m.service_group=?)"
				params = []interface{}{monitorType, serviceGroup}
			}
		} else {
			baseSql = "select m.*,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.metric_id as metric_id from metric m join metric_comparison mc on mc.metric_id = m.guid  and  m.monitor_type=? and m.service_group is null"
			params = []interface{}{monitorType}
		}
	}
	result = []*models.MetricComparisonExtend{}
	baseSql = baseSql + " order by m.update_time desc"
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
			// 业务配置类型 兜底
			if strings.TrimSpace(metric.LogMetricConfig) != "" || strings.TrimSpace(metric.LogMetricTemplate) != "" {
				metric.MetricType = string(models.MetricTypeBusiness)
			} else {
				metric.MetricType = string(models.MetricTypeCustom)
			}
		}
	}
	return
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
				baseSql = "select * from metric m where monitor_type=? and service_group=? and not exists (select guid from metric_comparison mc where mc.origin_metric_id = m.guid)"
				params = []interface{}{monitorType, serviceGroup}
			} else {
				baseSql = "select * from metric m where monitor_type=? and (service_group is null or service_group=?) and not exists (select guid from metric_comparison mc where mc.origin_metric_id = m.guid)"
				params = []interface{}{monitorType, serviceGroup}
			}
		} else {
			baseSql = "select * from metric m where monitor_type=? and service_group is null and not exists (select guid from metric_comparison mc where mc.origin_metric_id = m.guid)"
			params = []interface{}{monitorType}
		}
	}
	result = []*models.MetricTable{}
	baseSql = baseSql + " order by update_time desc"
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
			// 业务配置类型 兜底
			if strings.TrimSpace(metric.LogMetricConfig) != "" || strings.TrimSpace(metric.LogMetricTemplate) != "" {
				metric.MetricType = string(models.MetricTypeBusiness)
			} else {
				metric.MetricType = string(models.MetricTypeCustom)
			}
		}
	}
	return
}

func MetricImport(serviceGroup, operator string, inputMetrics []*models.MetricTable) ([]string, error) {
	var failList []string
	var err error
	existMetrics, getExistErr := MetricListNew("", inputMetrics[0].MonitorType, serviceGroup, "Y")
	if getExistErr != nil {
		return failList, fmt.Errorf("get serviceGroup:%s exist metric list fail,%s ", serviceGroup, getExistErr.Error())
	}
	var alarmStrategyRows []*models.AlarmStrategyTable
	if err = x.SQL("select endpoint_group,metric from alarm_strategy").Find(&alarmStrategyRows); err != nil {
		return failList, fmt.Errorf("query alarm strategy fail,%s ", err.Error())
	}
	oldServerGroup := inputMetrics[0].ServiceGroup
	strategyMap := make(map[string]string)
	for _, v := range alarmStrategyRows {
		strategyMap[v.Metric] = v.EndpointGroup
	}
	var actions []*Action
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
		if matchMetric.Guid != "" {
			// 指标重复后,指标名_1,如果还是重复,记录在失败列表
			inputMetric.Metric = inputMetric.Metric + "_1"
			inputMetric.Guid = fmt.Sprintf("%s__%s", inputMetric.Metric, serviceGroup)
			matchMetric = &models.MetricTable{}
			for _, existMetric := range existMetrics {
				if existMetric.Guid == inputMetric.Guid {
					matchMetric = existMetric
					break
				}
			}
		}
		inputMetric.PromExpr = strings.ReplaceAll(inputMetric.PromExpr, oldServerGroup, serviceGroup)
		if matchMetric.Guid != "" {
			failList = append(failList, inputMetric.Metric)
		} else {
			var tempMetric string
			x.SQL("select metric from metric where guid = ?", inputMetric.Guid).Get(&tempMetric)
			if tempMetric != "" {
				failList = append(failList, tempMetric)
			} else {
				actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?)",
					Param: []interface{}{inputMetric.Guid, inputMetric.Metric, inputMetric.MonitorType, inputMetric.PromExpr, serviceGroup, inputMetric.Workspace, nowTime, nowTime, operator, operator}})
			}
		}
	}
	log.Logger.Info("import metric", log.Int("actionLen", len(actions)))
	if len(actions) > 0 {
		if err = Transaction(actions); err != nil {
			return failList, fmt.Errorf("import metric fail with exec database,%s ", err.Error())
		}
	}
	return failList, nil
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
				var tmpTagList []string
				if err = json.Unmarshal([]byte(logMetricConfigRows[0].TagConfig), &tmpTagList); err != nil {
					err = fmt.Errorf("json unmarshal tag config: %s fail,%s ", tagConfigString, err.Error())
					return
				}
				if len(tmpTagList) > 0 {
					tags = []string{"tags"}
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
					return
				}
			}
		}
		return
	}
	tagParamList := getPromTagParamList(metricRow.PromExpr)
	for _, v := range tagParamList {
		if strings.HasPrefix(v, "$t_") {
			tags = append(tags, v[3:])
		}
	}
	return
}

func getPromTagParamList(promQl string) (tagList []string) {
	if strings.Contains(promQl, "$") {
		re, _ := regexp.Compile("=\"[\\$]+[^\"]+\"")
		fetchTag := re.FindAll([]byte(promQl), -1)
		for _, vv := range fetchTag {
			tmpV := string(vv)
			if len(tmpV) < 3 {
				continue
			}
			tmpV = tmpV[2 : len(tmpV)-1]
			tagList = append(tagList, tmpV)
		}
	}
	return
}

func GetMetric(id string) (metric *models.MetricTable, err error) {
	metric = &models.MetricTable{}
	_, err = x.SQL("select * from metric where guid=?", id).Get(metric)
	return
}

func AddComparisonMetric(param models.MetricComparisonDto, metric *models.MetricTable, operator string) (err error) {
	var actions []*Action
	newMetricId := getComparisonMetricId(metric.Guid, param.ComparisonType, param.CalcMethod, param.CalcPeriod)
	now := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user) values (?,?,?,?,?,?,?,?,?,?)",
		Param: []interface{}{newMetricId, metric.Metric, metric.MonitorType, newMetricId, metric.ServiceGroup, metric.Workspace, now, now, operator, operator}})
	actions = append(actions, &Action{Sql: "insert into metric_comparison(guid,comparison_type,calc_type,calc_method,calc_period,metric_id,origin_metric_id,create_user,create_time) values(?,?,?,?,?,?,?,?,?)",
		Param: []interface{}{guid.CreateGuid(), param.ComparisonType, param.CalcType, param.CalcMethod, param.CalcPeriod, newMetricId, metric.Guid, operator, now}})
	return Transaction(actions)
}

func GetComparisonMetricDtoList() (list []*models.MetricComparisonDto, err error) {
	err = x.SQL("select m.metric,m.prom_expr as origin_prom_expr,mc.guid as prom_expr,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period from metric m join metric_comparison mc on m.guid = mc.origin_metric_id").Find(&list)
	return
}

func getComparisonMetricId(originMetricId, comparisonType, calcMethod, calcPeriod string) string {
	if comparisonType == "" {
		return ""
	}
	return originMetricId + "_" + comparisonType[0:1] + "_" + calcMethod + "_" + calcPeriod
}
