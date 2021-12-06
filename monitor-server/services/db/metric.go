package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
)

func MetricList(id int, endpointType string) (result []*models.PromMetricTable, err error) {
	params := []interface{}{}
	baseSql := "select * from prom_metric where 1=1 "
	if id > 0 {
		baseSql += " and id=? "
		params = append(params, id)
	}
	if endpointType != "" {
		baseSql += " and metric_type=? "
		params = append(params, endpointType)
	}
	result = []*models.PromMetricTable{}
	err = x.SQL(baseSql, params...).Find(&result)
	if err != nil {
		return
	}
	// append service metric
	var logMetricTable []*models.LogMetricConfigTable
	x.SQL("select guid,metric,display_name,agg_type from log_metric_config where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?) or log_metric_json in (select guid from log_metric_json where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?))", endpointType, endpointType).Find(&logMetricTable)
	for _, v := range logMetricTable {
		result = append(result, &models.PromMetricTable{Id: 0, MetricType: endpointType, Metric: v.Metric, PromQl: fmt.Sprintf("%s{key=\"%s\",agg=\"%s\",t_endpoint=\"$guid\"}", models.LogMetricName, v.Metric, v.AggType)})
	}
	var dbMetricTable []*models.DbMetricMonitorTable
	x.SQL("select guid,metric,display_name from db_metric_monitor where monitor_type=?", endpointType).Find(&dbMetricTable)
	for _, v := range dbMetricTable {
		result = append(result, &models.PromMetricTable{Id: 0, MetricType: endpointType, Metric: v.Metric, PromQl: fmt.Sprintf("%s{key=\"%s\",t_endpoint=\"$guid\"}", models.DBMonitorMetricName, v.Metric)})
	}
	return
}

func MetricCreate(param []*models.PromMetricTable) error {
	var actions []*Action
	for _, metric := range param {
		actions = append(actions, &Action{Sql: "insert into prom_metric(metric,metric_type,prom_ql) value (?,?,?)", Param: []interface{}{metric.Metric, metric.MetricType, metric.PromQl}})
	}
	return Transaction(actions)
}

func MetricUpdate(param []*models.PromMetricTable) error {
	var actions []*Action
	for _, metric := range param {
		actions = append(actions, &Action{Sql: "update prom_metric set metric=?,prom_ql=? where id=?", Param: []interface{}{metric.Metric, metric.PromQl, metric.Id}})
	}
	return Transaction(actions)
}

func MetricDelete(id int) error {
	metricQuery, err := MetricList(id, "")
	if err != nil {
		return fmt.Errorf("Try to query prom metric table fail,%s ", err.Error())
	}
	if len(metricQuery) == 0 {
		return nil
	}
	metric := metricQuery[0].Metric
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from prom_metric where id=?", Param: []interface{}{id}})
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

func MetricListNew(guid,monitorType string) (result []*models.MetricTable,err error) {
	params := []interface{}{}
	baseSql := "select * from metric where 1=1 "
	if guid != "" {
		baseSql += " and guid=? "
		params = append(params, guid)
	}
	if monitorType != "" {
		baseSql += " and monitor_type=? "
		params = append(params, monitorType)
	}
	result = []*models.MetricTable{}
	err = x.SQL(baseSql, params...).Find(&result)
	if err != nil {
		return
	}
	// append service metric
	var logMetricTable []*models.LogMetricConfigTable
	x.SQL("select guid,metric,display_name,agg_type from log_metric_config where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?) or log_metric_json in (select guid from log_metric_json where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?))", monitorType, monitorType).Find(&logMetricTable)
	for _, v := range logMetricTable {
		result = append(result, &models.MetricTable{Guid: v.Guid, MonitorType: monitorType, Metric: v.Metric, PromExpr: fmt.Sprintf("%s{key=\"%s\",agg=\"%s\",t_endpoint=\"$guid\"}", models.LogMetricName, v.Metric, v.AggType)})
	}
	var dbMetricTable []*models.DbMetricMonitorTable
	x.SQL("select guid,metric,display_name from db_metric_monitor where monitor_type=?", monitorType).Find(&dbMetricTable)
	for _, v := range dbMetricTable {
		result = append(result, &models.MetricTable{Guid: v.Guid, MonitorType: monitorType, Metric: v.Metric, PromExpr: fmt.Sprintf("%s{key=\"%s\",t_endpoint=\"$guid\"}", models.DBMonitorMetricName, v.Metric)})
	}
	return
}