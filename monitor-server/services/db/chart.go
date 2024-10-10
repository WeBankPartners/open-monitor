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
	if serviceGroup != "" && monitorType == "process" {
		err = x.SQL("select * from metric where metric=? and monitor_type=? and service_group=?", metric, monitorType, serviceGroup).Find(&metricTable)
	} else if monitorType != "" {
		err = x.SQL("select * from metric where metric=? and monitor_type=? and service_group is null", metric, monitorType).Find(&metricTable)
	} else {
		err = x.SQL("select * from metric where metric=?", metric).Find(&metricTable)
	}
	if err != nil {
		err = fmt.Errorf("query metric table fail,%s ", err.Error())
		return
	}
	if len(metricTable) > 0 {
		result = metricTable[0].PromExpr
	} else {
		if monitorType == "process" {
			err = x.SQL("select * from metric where metric=? and monitor_type=? and service_group is null", metric, monitorType).Find(&metricTable)
			if err != nil {
				err = fmt.Errorf("query metric table fail,%s ", err.Error())
				return
			}
			if len(metricTable) > 0 {
				result = metricTable[0].PromExpr
			}
		}
	}
	return
}

func GetCustomChartSeries(customChartGuid string) (series []*models.CustomChartSeriesDto, err error) {
	var seriesRows []*models.CustomChartSeries
	err = x.SQL("select * from custom_chart_series where dashboard_chart=?", customChartGuid).Find(&seriesRows)
	if err != nil {
		err = fmt.Errorf("query custom_chart_series table fail,%s ", err.Error())
		return
	}
	var tagValueRows []*models.CustomChartTagValueRow
	err = x.SQL("select t2.dashboard_chart_config as `chart_guid`,t2.name,t1.value,t2.equal from custom_chart_series_tagvalue t1 left join custom_chart_series_tag t2 on t1.dashboard_chart_tag=t2.guid where t2.dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart=?)", customChartGuid).Find(&tagValueRows)
	if err != nil {
		err = fmt.Errorf("query custom_chart_tag_value fail,%s ", err.Error())
		return
	}
	var tagRows []*models.CustomChartSeriesTag
	err = x.SQL("select * from custom_chart_series_tag where dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart=?)", customChartGuid).Find(&tagRows)
	if err != nil {
		err = fmt.Errorf("query custom_chart_tag fail,%s ", err.Error())
		return
	}
	for _, row := range seriesRows {
		tmpSerieObj := models.CustomChartSeriesDto{
			Endpoint:      row.Endpoint,
			ServiceGroup:  row.ServiceGroup,
			EndpointName:  row.EndpointName,
			MonitorType:   row.MonitorType,
			ColorGroup:    row.ColorGroup,
			PieDisplayTag: row.PieDisplayTag,
			EndpointType:  row.EndpointType,
			MetricType:    row.MetricType,
			MetricGuid:    row.MetricGuid,
			Metric:        row.Metric,
			Tags:          []*models.TagDto{},
			ColorConfig:   nil,
		}
		//tagNameList := []string{}
		//for _, tagName := range tagRows {
		//	tagNameList = append(tagNameList, tagName.Name)
		//}
		tagValueMap := make(map[string][]string)
		for _, tvRow := range tagValueRows {
			if tvRow.ChartGuid == row.Guid {
				if existValueList, ok := tagValueMap[tvRow.Name]; ok {
					tagValueMap[tvRow.Name] = append(existValueList, tvRow.Value)
				} else {
					//tagNameList = append(tagNameList, tvRow.Name)
					tagValueMap[tvRow.Name] = []string{tvRow.Value}
				}
			}
		}
		for _, tagName := range tagRows {
			if *tagName.DashboardChartConfig != row.Guid {
				continue
			}
			tmpValueList := []string{}
			if existTagList, ok := tagValueMap[tagName.Name]; ok {
				tmpValueList = existTagList
			}
			tmpSerieObj.Tags = append(tmpSerieObj.Tags, &models.TagDto{TagName: tagName.Name, TagValue: tmpValueList, Equal: tagName.Equal})
		}
		series = append(series, &tmpSerieObj)
	}
	return
}
