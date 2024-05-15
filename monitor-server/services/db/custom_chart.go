package db

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func GetCustomChartById(id string) (chart *models.CustomChart, err error) {
	_, err = x.SQL("select * from custom_chart where id = ?", id).Get(chart)
	return
}

func QueryAllPublicCustomChartList(roles []string) (list []*models.CustomChart, err error) {
	roleFilterSql, roleFilterParam := createListParams(roles, "")
	var params []interface{}
	var ids []string
	var sql = "select dashboard_chart from custom_chart_permission where role_id  in (" + roleFilterSql + ") and permission = 'USE'"
	params = append(params, roleFilterParam...)
	if err = x.SQL(sql).Find(&ids); err != nil {
		return
	}
	if len(ids) > 0 {
		idFilterSql, idFilterParam := createListParams(ids, "")
		err = x.SQL("select * from custom_chart where public = 1 and guid in ( "+idFilterSql+")", idFilterParam).Find(&list)
	}
	return
}

func QueryCustomChartListByDashboard(customDashboard int) (list []*models.CustomChartExtend, err error) {
	err = x.SQL("select c.*,r.`group`,r.display_config from custom_dashboard_chart_rel  r join custom_chart  c "+
		"on r.dashboard_chart = c.guid where r.custom_dashboard = ?", customDashboard).Find(&list)
	return
}

func QueryCustomChartSeriesByChart(chartId string) (list []*models.CustomChartSeries, err error) {
	err = x.SQL("select * from custom_chart_series where dashboard_chart  = ?", chartId).Find(&list)
	return
}

func QueryCustomChartPermissionByChart(chartId string) (hashMap map[string]string, err error) {
	var list []*models.CustomChartPermission
	hashMap = make(map[string]string)
	err = x.SQL("select * from custom_chart_permission where dashboard_chart = ?", chartId).Find(&list)
	if len(list) > 0 {
		for _, roleRel := range list {
			hashMap[roleRel.RoleId] = roleRel.Permission
		}
	}
	return
}

func QueryAllChartSeriesConfig() (configMap map[string][]*models.CustomChartSeriesConfig, err error) {
	var list []*models.CustomChartSeriesConfig
	configMap = make(map[string][]*models.CustomChartSeriesConfig)
	if err = x.SQL("select * from custom_chart_series_config").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, config := range list {
			if arr, ok := configMap[*config.DashboardChartConfig]; ok {
				arr = append(arr, config)
			} else {
				configMap[*config.DashboardChartConfig] = []*models.CustomChartSeriesConfig{}
				configMap[*config.DashboardChartConfig] = append(configMap[*config.DashboardChartConfig], config)
			}
		}
	}
	return
}

func QueryAllChartSeriesTag() (tagMap map[string][]*models.CustomChartSeriesTag, err error) {
	var list []*models.CustomChartSeriesTag
	tagMap = make(map[string][]*models.CustomChartSeriesTag)
	if err = x.SQL("select * from custom_chart_series_tag").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, config := range list {
			if arr, ok := tagMap[*config.DashboardChartConfig]; ok {
				arr = append(arr, config)
			} else {
				tagMap[*config.DashboardChartConfig] = []*models.CustomChartSeriesTag{}
				tagMap[*config.DashboardChartConfig] = append(tagMap[*config.DashboardChartConfig], config)
			}
		}
	}
	return
}

func QueryAllChartSeriesTagValue() (tagValueMap map[string][]*models.CustomChartSeriesTagValue, err error) {
	var list []*models.CustomChartSeriesTagValue
	tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	if err = x.SQL("select * from custom_chart_series_tagvalue").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, tagValue := range list {
			if arr, ok := tagValueMap[*tagValue.DashboardChartTag]; ok {
				arr = append(arr, tagValue)
			} else {
				tagValueMap[*tagValue.DashboardChartTag] = []*models.CustomChartSeriesTagValue{}
				tagValueMap[*tagValue.DashboardChartTag] = append(tagValueMap[*tagValue.DashboardChartTag], tagValue)
			}
		}
	}
	return
}

func QueryCustomDashboardChartRelListByDashboard(dashboardId int) (list []*models.CustomDashboardChartRel, err error) {
	err = x.SQL("select * from custom_dashboard_chart_rel where custom_dashboard = ?", dashboardId).Find(&list)
	return
}

func GetAddCustomDashboardChartRelSQL(chartRelList []*models.CustomDashboardChartRel) []*Action {
	var actions []*Action
	if len(chartRelList) > 0 {
		for _, rel := range chartRelList {
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel values(?,?,?,?,?,?,?,?,?)", Param: []interface{}{rel.Guid,
				rel.CustomDashboard, rel.DashboardChart, rel.Group, rel.DisplayConfig, rel.CreateUser, rel.UpdateUser, rel.CreateTime, rel.UpdateTime}})
		}
	}
	return actions
}

func GetUpdateCustomDashboardChartRelSQL(chartRelList []*models.CustomDashboardChartRel) []*Action {
	var actions []*Action
	if len(chartRelList) > 0 {
		for _, rel := range chartRelList {
			actions = append(actions, &Action{Sql: "update custom_dashboard_chart_rel set `group` = ?,display_config = ?,update_user = ?," +
				"update_time = ? where id =?", Param: []interface{}{rel.Group, rel.DisplayConfig, rel.UpdateUser, rel.UpdateTime, rel.Guid}})
		}
	}
	return actions
}

func GetDeleteCustomDashboardChartRelSQL(ids []string) []*Action {
	var actions []*Action
	if len(ids) > 0 {
		for _, id := range ids {
			actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where id = ?", Param: []interface{}{id}})
		}
	}
	return actions
}

func GetDeleteCustomChartPermissionSQL(chartId string) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from custom_chart_permission where dashboard_chart = ?", Param: []interface{}{chartId}})
	return actions
}

func GetInsertCustomChartPermissionSQL(permissionList []*models.CustomChartPermission) []*Action {
	var actions []*Action
	if len(permissionList) > 0 {
		for _, permission := range permissionList {
			actions = append(actions, &Action{Sql: "insert into custom_chart_permission values (?,?,?,?)",
				Param: []interface{}{permission.Guid, permission.DashboardChart, permission.RoleId, permission.Permission}})
		}
	}
	return actions
}

func GetUpdateCustomChartPublicSQL(chartId string) []*Action {
	var actions []*Action
	now := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "update  custom_chart set public = 1,update_time = ? where guid = ?", Param: []interface{}{now, chartId}})
	return actions
}

func DeleteCustomDashboardChart(chartId string) (err error) {
	var actions []*Action
	var chartSeriesIds, seriesConfIds, seriesTagIds, seriesTagValueIds []string
	if err = x.SQL("select guid from custom_chart_series where dashboard_chart = ?", chartId).Find(&chartSeriesIds); err != nil {
		return
	}
	if len(chartSeriesIds) > 0 {
		chartSeriesSQL, chartSeriesParams := createListParams(chartSeriesIds, "")
		if err = x.SQL("select guid from custom_chart_series_config where dashboard_chart_config in ("+chartSeriesSQL+")",
			chartSeriesParams).Find(&seriesConfIds); err != nil {
			return
		}
		if err = x.SQL("select guid from custom_chart_series_tag where dashboard_chart_config in ("+chartSeriesSQL+")",
			chartSeriesParams).Find(&seriesTagIds); err != nil {
			return
		}
		if len(seriesTagIds) > 0 {
			seriesTagSQL, seriesTagParams := createListParams(seriesTagIds, "")
			if err = x.SQL("select id from custom_chart_series_tagvalue where dashboard_chart_tag in ("+seriesTagSQL+")",
				seriesTagParams).Find(&seriesTagValueIds); err != nil {
				return
			}
		}
	}

	if len(seriesConfIds) > 0 {
		for _, confId := range seriesConfIds {
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_config where guid = ?", Param: []interface{}{confId}})
		}
	}
	if len(seriesTagValueIds) > 0 {
		for _, tagValueId := range seriesTagValueIds {
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_tagvalue where id = ?", Param: []interface{}{tagValueId}})
		}
	}
	if len(seriesTagIds) > 0 {
		for _, tagId := range seriesTagIds {
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag where id = ?", Param: []interface{}{tagId}})
		}
	}

	actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where dashboard_chart = ?", Param: []interface{}{chartId}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series where dashboard_chart = ?", Param: []interface{}{chartId}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_permission where dashboard_chart = ?", Param: []interface{}{chartId}})
	actions = append(actions, &Action{Sql: "delete from custom_chart WHERE id=?", Param: []interface{}{chartId}})
	return Transaction(actions)
}

func CreateCustomChartDto(chartExtend *models.CustomChartExtend, configMap map[string][]*models.CustomChartSeriesConfig, tagMap map[string][]*models.CustomChartSeriesTag, tagValueMap map[string][]*models.CustomChartSeriesTagValue) (chart *models.CustomChartDto, err error) {
	var list []*models.CustomChartSeries
	var seriesConfigList []*models.CustomChartSeriesConfig
	var chartSeriesTagList []*models.CustomChartSeriesTag
	var chartSeriesTagValueList []*models.CustomChartSeriesTagValue
	chart = &models.CustomChartDto{
		Id:              chartExtend.CustomChart.Guid,
		Public:          intToBool(chartExtend.CustomChart.Public),
		SourceDashboard: chartExtend.CustomChart.SourceDashboard,
		Name:            chartExtend.CustomChart.Name,
		Unit:            chartExtend.CustomChart.Unit,
		ChartType:       chartExtend.CustomChart.ChartType,
		LineType:        chartExtend.CustomChart.LineType,
		Aggregate:       chartExtend.CustomChart.Aggregate,
		AggStep:         chartExtend.CustomChart.AggStep,
		Query:           nil,
		DisplayConfig:   chartExtend.DisplayConfig,
		Group:           chartExtend.Group,
	}
	chart.Query = []*models.CustomChartSeriesDto{}
	if list, err = QueryCustomChartSeriesByChart(chartExtend.CustomChart.Guid); err != nil {
		return
	}
	if len(list) > 0 {
		for _, series := range list {
			seriesConfigList = []*models.CustomChartSeriesConfig{}
			chartSeriesTagList = []*models.CustomChartSeriesTag{}
			customChartSeriesDto := &models.CustomChartSeriesDto{
				Endpoint:     series.Endpoint,
				ServiceGroup: series.ServiceGroup,
				EndpointName: series.EndpointName,
				MonitorType:  series.MonitorType,
				ColorGroup:   series.ColorGroup,
				Metric:       series.Metric,
				Tags:         make([]*models.TagDto, 0),
				ColorConfig:  make([]*models.ColorConfigDto, 0),
			}
			if v, ok := configMap[series.Guid]; ok {
				seriesConfigList = v
			}
			if v, ok := tagMap[series.Guid]; ok {
				chartSeriesTagList = v
			}
			if len(chartSeriesTagList) > 0 {
				for _, tag := range chartSeriesTagList {
					chartSeriesTagValueList = []*models.CustomChartSeriesTagValue{}
					if v, ok := tagValueMap[tag.Guid]; ok {
						chartSeriesTagValueList = v
					}
					customChartSeriesDto.Tags = append(customChartSeriesDto.Tags, &models.TagDto{
						TagName:  tag.Name,
						TagValue: getChartSeriesTagValues(chartSeriesTagValueList),
					})
				}
			}
			if len(seriesConfigList) > 0 {
				for _, config := range seriesConfigList {
					customChartSeriesDto.ColorConfig = append(customChartSeriesDto.ColorConfig, &models.ColorConfigDto{
						Metric: series.Metric,
						Color:  config.Color,
					})
				}
			}
			chart.Query = append(chart.Query, customChartSeriesDto)
		}
	}
	return
}

func intToBool(num int) bool {
	return num != 0
}

func getChartSeriesTagValues(chartSeriesTagValueList []*models.CustomChartSeriesTagValue) []string {
	var result []string
	if len(chartSeriesTagValueList) > 0 {
		for _, tagValue := range chartSeriesTagValueList {
			result = append(result, tagValue.Value)
		}
	}
	return result
}
