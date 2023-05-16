package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
)

func ChartList(id, groupId int) (result []*models.ChartTable, err error) {
	result = []*models.ChartTable{}
	params := []interface{}{}
	baseSql := "select * from chart where 1=1 "
	if id > 0 {
		baseSql += " and id=? "
		params = append(params, id)
	}
	if groupId > 0 {
		baseSql += " and group_id=? "
		params = append(params, groupId)
	}
	err = x.SQL(baseSql, params...).Find(&result)
	return
}

func getMaxChartGroupId() (id int, err error) {
	query, queryErr := x.QueryString("select max(chart_group) as id from panel")
	if queryErr != nil {
		err = fmt.Errorf("Try to get max chart groupId fail,%s ", queryErr.Error())
		return
	}
	id, _ = strconv.Atoi(query[0]["id"])
	return
}

func ChartCreate(param []*models.ChartTable) error {
	var actions []*Action
	for _, chart := range param {
		if chart.AggType == "" {
			chart.AggType = "avg"
		}
		actions = append(actions, &Action{Sql: "insert into chart(group_id,metric,url,unit,title,agg_type,legend) value (?,?,'/dashboard/chart',?,?,?,?)", Param: []interface{}{chart.GroupId, chart.Metric, chart.Unit, chart.Title, chart.AggType, chart.Legend}})
	}
	return Transaction(actions)
}

func ChartUpdate(param []*models.ChartTable) error {
	var actions []*Action
	for _, chart := range param {
		if chart.AggType == "" {
			chart.AggType = "avg"
		}
		actions = append(actions, &Action{Sql: "update chart set metric=?,unit=?,title=?,agg_type=?,legend=? where id=?", Param: []interface{}{chart.Metric, chart.Unit, chart.Title, chart.AggType, chart.Legend, chart.Id}})
	}
	return Transaction(actions)
}

func ChartDelete(ids []string) error {
	var actions []*Action
	for _, id := range ids {
		idInt, tmpErr := strconv.Atoi(id)
		if tmpErr != nil {
			log.Logger.Error("Try to trans id to int fail", log.Error(tmpErr))
			continue
		}
		actions = append(actions, &Action{Sql: "delete from chart where id=?", Param: []interface{}{idInt}})
	}
	return Transaction(actions)
}

func GetPromQLByMetric(metric, monitorType, serviceGroup string) (result string, err error) {
	var metricTable []*models.MetricTable
	if serviceGroup != "" {
		err = x.SQL("select * from metric where metric=? and monitor_type=? and service_group=?", metric, monitorType, serviceGroup).Find(&metricTable)
	} else {
		err = x.SQL("select * from metric where metric=? and monitor_type=?", metric, monitorType).Find(&metricTable)
	}
	if len(metricTable) > 0 {
		result = metricTable[0].PromExpr
	}
	return
}
