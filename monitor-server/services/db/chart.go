package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
)

func ChartList(id,groupId int) (result []*models.ChartTable,err error) {
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

func getMaxChartGroupId() (id int,err error) {
	query,queryErr := x.QueryString("select max(chart_group) as id from panel")
	if queryErr != nil {
		err = fmt.Errorf("Try to get max chart groupId fail,%s ", queryErr.Error())
		return
	}
	id,_ = strconv.Atoi(query[0]["id"])
	return
}

func ChartCreate(param []*models.ChartTable) error {
	var actions []*Action
	for _,chart := range param {
		if chart.AggType == "" {
			chart.AggType = "avg"
		}
		actions = append(actions, &Action{Sql: "insert into chart(group_id,metric,url,unit,title,agg_type,legend) value (?,?,'/dashboard/chart',?,?,?,'$metric')", Param: []interface{}{chart.GroupId,chart.Metric,chart.Unit,chart.Title,chart.AggType}})
	}
	return Transaction(actions)
}

func ChartUpdate(param []*models.ChartTable) error {
	var actions []*Action
	for _,chart := range param {
		if chart.AggType == "" {
			chart.AggType = "avg"
		}
		actions = append(actions, &Action{Sql: "update chart set metric=?,unit=?,title=?,agg_type=? where id=?", Param: []interface{}{chart.Metric,chart.Unit,chart.Title,chart.AggType,chart.Id}})
	}
	return Transaction(actions)
}

func ChartDelete(param []*models.ChartTable) error {
	var actions []*Action
	for _,chart := range param {
		actions = append(actions, &Action{Sql: "delete from chart where id=?", Param: []interface{}{chart.Id}})
	}
	return Transaction(actions)
}