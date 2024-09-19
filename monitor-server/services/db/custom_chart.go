package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"sort"
	"strings"
	"time"
)

func GetCustomChartById(id string) (chart *models.CustomChart, err error) {
	chart = &models.CustomChart{}
	_, err = x.SQL("select * from custom_chart where guid = ?", id).Get(chart)
	return
}

func QueryCustomChartByName(name string) (list []*models.CustomChart, err error) {
	err = x.SQL("select * from custom_chart where name = ? and public=1", name).Find(&list)
	return
}

func QueryAllPublicCustomChartList(dashboardId int, chartName string, roles []string) (list []*models.CustomChart, err error) {
	roleFilterSql, roleFilterParam := createListParams(roles, "")
	var params []interface{}
	var ids, newIds []string
	var sql = "select distinct dashboard_chart from custom_chart_permission where role_id  in (" + roleFilterSql + ") and permission = ?"
	params = append(append(params, roleFilterParam...), models.PermissionUse)
	if err = x.SQL(sql, params...).Find(&ids); err != nil {
		return
	}
	if dashboardId != 0 {
		var dashboardChartList []string
		if err = x.SQL("select dashboard_chart from  custom_dashboard_chart_rel where custom_dashboard=?", dashboardId).Find(&dashboardChartList); err != nil {
			return
		}
		// 取图表id交集
		for _, id := range ids {
			for _, chartId := range dashboardChartList {
				if id == chartId {
					newIds = append(newIds, id)
					break
				}
			}
		}
	} else {
		newIds = ids
	}
	if len(newIds) > 0 {
		idFilterSql, idFilterParam := createListParams(newIds, "")
		if chartName == "" {
			err = x.SQL("select * from custom_chart where public = 1 and guid in ( "+idFilterSql+")", idFilterParam...).Find(&list)
		} else {
			err = x.SQL("select * from custom_chart where public = 1 and name like '%"+chartName+"%' and guid in ( "+idFilterSql+")", idFilterParam...).Find(&list)
		}
	}
	return
}

func QueryCustomChartListByDashboard(customDashboard int) (list []*models.CustomChartExtend, err error) {
	err = x.SQL("select c.*,r.`group`,r.display_config,r.group_display_config from custom_dashboard_chart_rel  r join custom_chart  c "+
		"on r.dashboard_chart = c.guid where r.custom_dashboard = ?", customDashboard).Find(&list)
	return
}

func QueryCustomChartSeriesByChart(chartId string) (list []*models.CustomChartSeries, err error) {
	err = x.SQL("select * from custom_chart_series where dashboard_chart  = ?", chartId).Find(&list)
	return
}

func QueryCustomChartManagePermissionByChart(chartId string) (hashMap map[string]string, err error) {
	var list []*models.CustomChartPermission
	hashMap = make(map[string]string)
	err = x.SQL("select * from custom_chart_permission where dashboard_chart = ? and permission = ?", chartId, models.PermissionMgmt).Find(&list)
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
			if _, ok := configMap[*config.DashboardChartConfig]; !ok {
				configMap[*config.DashboardChartConfig] = []*models.CustomChartSeriesConfig{}
			}
			configMap[*config.DashboardChartConfig] = append(configMap[*config.DashboardChartConfig], config)
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
			if _, ok := tagMap[*config.DashboardChartConfig]; !ok {
				tagMap[*config.DashboardChartConfig] = []*models.CustomChartSeriesTag{}
			}
			tagMap[*config.DashboardChartConfig] = append(tagMap[*config.DashboardChartConfig], config)
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
			if _, ok := tagValueMap[*tagValue.DashboardChartTag]; !ok {
				tagValueMap[*tagValue.DashboardChartTag] = []*models.CustomChartSeriesTagValue{}
			}
			tagValueMap[*tagValue.DashboardChartTag] = append(tagValueMap[*tagValue.DashboardChartTag], tagValue)
		}
	}
	return
}

func QueryCustomDashboardChartRelListByDashboard(dashboardId int) (list []*models.CustomDashboardChartRel, err error) {
	err = x.SQL("select * from custom_dashboard_chart_rel where custom_dashboard = ?", dashboardId).Find(&list)
	return
}

func QueryCustomDashboardChartRelListByChart(chartId string) (list []*models.CustomDashboardChartRel, err error) {
	err = x.SQL("select * from custom_dashboard_chart_rel where dashboard_chart = ?", chartId).Find(&list)
	return
}

func GetAddCustomDashboardChartRelSQL(chartRelList []*models.CustomDashboardChartRel) []*Action {
	var actions []*Action
	if len(chartRelList) > 0 {
		for _, rel := range chartRelList {
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{rel.Guid,
				rel.CustomDashboard, rel.DashboardChart, rel.Group, rel.DisplayConfig, rel.CreateUser, rel.UpdateUser, rel.CreateTime, rel.UpdateTime, rel.GroupDisplayConfig}})
		}
	}
	return actions
}

func GetDeleteCustomDashboardRoleRelSQL(dashboardId int) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_role_rel where custom_dashboard_id = ?", Param: []interface{}{dashboardId}})
	return actions
}

func GetInsertCustomDashboardRoleRelSQL(dashboardId int, mgmtRoles, useRoles []string) []*Action {
	var actions []*Action
	if len(mgmtRoles) > 0 {
		for _, role := range mgmtRoles {
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_role_rel (custom_dashboard_id,permission,role_id)values(?,?,?)",
				Param: []interface{}{dashboardId, models.PermissionMgmt, role}})
		}
	}
	if len(useRoles) > 0 {
		for _, role := range useRoles {
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_role_rel (custom_dashboard_id,permission,role_id)values(?,?,?)",
				Param: []interface{}{dashboardId, models.PermissionUse, role}})
		}
	}
	return actions
}

func GetUpdateCustomDashboardChartRelSQL(chartRelList []*models.CustomDashboardChartRel) []*Action {
	var actions []*Action
	if len(chartRelList) > 0 {
		for _, rel := range chartRelList {
			actions = append(actions, &Action{Sql: "update custom_dashboard_chart_rel set `group` = ?,display_config = ?,group_display_config = ?,updated_user = ?," +
				"update_time = ? where guid =?", Param: []interface{}{rel.Group, rel.DisplayConfig, rel.GroupDisplayConfig, rel.UpdateUser, rel.UpdateTime, rel.Guid}})
		}
	}
	return actions
}

func GetUpdateCustomDashboardSQL(name, panelGroups, user string, timeRange, refreshWeek, id int) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "update custom_dashboard set name=?,update_user=?,update_at=?,panel_groups=?,time_range=?,refresh_week=? where id =?",
		Param: []interface{}{name, user, time.Now().Format(models.DatetimeFormat), panelGroups, timeRange, refreshWeek, id}})
	return actions
}

func GetDeleteCustomDashboardChartRelSQL(ids []string) []*Action {
	var actions []*Action
	if len(ids) > 0 {
		for _, id := range ids {
			actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where guid = ?", Param: []interface{}{id}})
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
			actions = append(actions, &Action{Sql: "insert into custom_chart_permission(guid,dashboard_chart,role_id,permission) values (?,?,?,?)",
				Param: []interface{}{permission.Guid, permission.DashboardChart, permission.RoleId, permission.Permission}})
		}
	}
	return actions
}

func QueryChartPermissionByCustomChart(customChart string) (list []*models.CustomChartPermission, err error) {
	err = x.SQL("select * from custom_chart_permission where dashboard_chart = ?", customChart).Find(&list)
	return
}

func QueryChartPermissionByCustomChartList(chartIds []string) (list []*models.CustomChartPermission, err error) {
	err = x.SQL("select * from custom_chart_permission where dashboard_chart in ('" + strings.Join(chartIds, "','") + "')").Find(&list)
	return
}

func GetUpdateCustomChartPublicSQL(chartId string) []*Action {
	var actions []*Action
	now := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "update  custom_chart set public = 1,update_time = ? where guid = ?", Param: []interface{}{now, chartId}})
	return actions
}

func DeleteCustomChartConfigSQL(chartId string) (actions []*Action, err error) {
	actions = []*Action{}
	var chartSeriesIds, seriesConfIds, seriesTagIds []string
	var seriesTagValueIds []int
	if err = x.SQL("select guid from custom_chart_series where dashboard_chart = ?", chartId).Find(&chartSeriesIds); err != nil {
		return
	}
	if len(chartSeriesIds) > 0 {
		chartSeriesSQL, chartSeriesParams := createListParams(chartSeriesIds, "")
		if err = x.SQL("select guid from custom_chart_series_config where dashboard_chart_config in ("+chartSeriesSQL+")",
			chartSeriesParams...).Find(&seriesConfIds); err != nil {
			return
		}
		if err = x.SQL("select guid from custom_chart_series_tag where dashboard_chart_config in ("+chartSeriesSQL+")",
			chartSeriesParams...).Find(&seriesTagIds); err != nil {
			return
		}
		if len(seriesTagIds) > 0 {
			seriesTagSQL, seriesTagParams := createListParams(seriesTagIds, "")
			if err = x.SQL("select id from custom_chart_series_tagvalue where dashboard_chart_tag in ("+seriesTagSQL+")",
				seriesTagParams...).Find(&seriesTagValueIds); err != nil {
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
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag where guid = ?", Param: []interface{}{tagId}})
		}
	}
	actions = append(actions, &Action{Sql: "delete from custom_chart_series where dashboard_chart = ?", Param: []interface{}{chartId}})
	return
}

func DeleteCustomChartSeriesByMetricIdSQL(metricId string) (actions []*Action, err error) {
	actions = []*Action{}
	var chartSeriesIds, seriesConfIds, seriesTagIds []string
	var seriesTagValueIds []int
	if err = x.SQL("select guid from custom_chart_series where metric_guid = ?", metricId).Find(&chartSeriesIds); err != nil {
		return
	}
	if len(chartSeriesIds) > 0 {
		chartSeriesSQL, chartSeriesParams := createListParams(chartSeriesIds, "")
		if err = x.SQL("select guid from custom_chart_series_config where dashboard_chart_config in ("+chartSeriesSQL+")",
			chartSeriesParams...).Find(&seriesConfIds); err != nil {
			return
		}
		if err = x.SQL("select guid from custom_chart_series_tag where dashboard_chart_config in ("+chartSeriesSQL+")",
			chartSeriesParams...).Find(&seriesTagIds); err != nil {
			return
		}
		if len(seriesTagIds) > 0 {
			seriesTagSQL, seriesTagParams := createListParams(seriesTagIds, "")
			if err = x.SQL("select id from custom_chart_series_tagvalue where dashboard_chart_tag in ("+seriesTagSQL+")",
				seriesTagParams...).Find(&seriesTagValueIds); err != nil {
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
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag where guid = ?", Param: []interface{}{tagId}})
		}
	}
	actions = append(actions, &Action{Sql: "delete from custom_chart_series where metric_guid = ?", Param: []interface{}{metricId}})
	return
}

func DeleteCustomDashboardChart(chartId string) (err error) {
	var actions []*Action
	actions, err = GetDeleteCustomDashboardChart(chartId)
	return Transaction(actions)
}

func GetDeleteCustomDashboardChart(chartId string) (actions []*Action, err error) {
	var subActions []*Action
	actions = []*Action{}
	if subActions, err = DeleteCustomChartConfigSQL(chartId); err != nil {
		return
	}
	if len(subActions) > 0 {
		actions = append(actions, subActions...)
	}
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where dashboard_chart = ?", Param: []interface{}{chartId}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_permission where dashboard_chart = ?", Param: []interface{}{chartId}})
	actions = append(actions, &Action{Sql: "delete from custom_chart WHERE guid = ?", Param: []interface{}{chartId}})
	return
}

func UpdateCustomChart(chartDto models.CustomChartDto, user string, sourceDashboard int) (err error) {
	var actions, subActions []*Action
	var seriesIdList []string
	now := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "update custom_chart set name =?,chart_type=?,line_type=?,pie_type=?,aggregate=?," +
		"agg_step=?,unit=?,update_user=?,update_time=?,chart_template = ? where guid=?", Param: []interface{}{chartDto.Name, chartDto.ChartType,
		chartDto.LineType, chartDto.PieType, chartDto.Aggregate, chartDto.AggStep, chartDto.Unit, user, now, chartDto.ChartTemplate, chartDto.Id}})
	// 更新源看板
	if sourceDashboard != 0 {
		actions = append(actions, &Action{Sql: "update custom_dashboard set update_user =?,update_at=? where id = ?", Param: []interface{}{user, now, sourceDashboard}})
	}
	// 先删除图表配置
	if subActions, err = DeleteCustomChartConfigSQL(chartDto.Id); err != nil {
		return
	}
	if len(subActions) > 0 {
		actions = append(actions, subActions...)
	}
	// 新增图表配置
	if len(chartDto.ChartSeries) > 0 {
		seriesIdList = guid.CreateGuidList(len(chartDto.ChartSeries))
		for i, series := range chartDto.ChartSeries {
			seriesId := seriesIdList[i]
			actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type," +
				"metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				seriesId, chartDto.Id, series.Endpoint, series.ServiceGroup, series.EndpointName, series.MonitorType, series.Metric, series.ColorGroup,
				series.PieDisplayTag, series.EndpointType, series.MetricType, series.MetricGuid}})
			if len(series.Tags) > 0 {
				for _, tag := range series.Tags {
					tagId := guid.CreateGuid()
					actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{
						tagId, seriesId, tag.TagName, tag.Equal}})
					if len(tag.TagValue) > 0 {
						for _, tagValue := range tag.TagValue {
							actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tagId, tagValue}})
						}
					}
				}
			}
			if len(series.ColorConfig) > 0 {
				for _, colorConfig := range series.ColorConfig {
					tags := ""
					if strings.Contains(colorConfig.SeriesName, "{") {
						start := strings.LastIndex(colorConfig.SeriesName, "{")
						tags = colorConfig.SeriesName[start+1 : len(colorConfig.SeriesName)-1]
					}
					actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{
						guid.CreateGuid(), seriesId, tags, colorConfig.Color, colorConfig.SeriesName,
					}})
				}
			}
		}
	}
	return Transaction(actions)
}

func AddCustomChart(param models.AddCustomChartParam, user string) (id string, err error) {
	var actions []*Action
	actions, id = getAddCustomChartActions(param, user)
	err = Transaction(actions)
	return
}

func getAddCustomChartActions(param models.AddCustomChartParam, user string) (actions []*Action, newChartId string) {
	actions = []*Action{}
	var displayConfig []byte
	newChartId = guid.CreateGuid()
	now := time.Now().Format(models.DatetimeFormat)
	chart := &models.CustomChart{
		Guid:            newChartId,
		SourceDashboard: param.DashboardId,
		Name:            param.Name,
		ChartTemplate:   param.ChartTemplate,
		ChartType:       param.ChartType,
		LineType:        param.LineType,
		PieType:         param.PieType,
		Aggregate:       param.Aggregate,
		AggStep:         param.AggStep,
		Unit:            param.Unit,
		CreateUser:      user,
		UpdateUser:      user,
		CreateTime:      now,
		UpdateTime:      now,
	}
	displayConfig, _ = json.Marshal(param.DisplayConfig)
	actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit,create_user,update_user,create_time,update_time,chart_template,pie_type) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		chart.Guid, chart.SourceDashboard, chart.Public, chart.Name, chart.ChartType, chart.LineType, chart.Aggregate,
		chart.AggStep, chart.Unit, chart.CreateUser, chart.UpdateUser, chart.CreateTime, chart.UpdateTime, chart.ChartTemplate, chart.PieType}})
	actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart, `group`,display_config,create_user,updated_user,create_time,update_time) values(?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		guid.CreateGuid(), param.DashboardId, chart.Guid, param.Group, string(displayConfig), user, user, now, now}})
	return
}
func QueryCustomChartList(condition models.QueryChartParam, operator string, roles []string) (pageInfo models.PageInfo, list []*models.CustomChart, err error) {
	var params []interface{}
	var ids []string
	var sql = "select * from custom_chart where 1=1 "
	if ids, err = getChartQueryIdsByPermission(condition, roles); err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	if condition.ChartId != "" {
		sql = sql + " and guid = ?"
		params = append(params, condition.ChartId)
	}
	if condition.ChartName != "" {
		sql = sql + " and name like '%" + condition.ChartName + "%'"
	}
	if condition.ChartType != "" {
		sql = sql + " and chart_type = ?"
		params = append(params, condition.ChartType)
	}
	if condition.SourceDashboard != 0 {
		sql = sql + " and source_dashboard = ?"
		params = append(params, condition.SourceDashboard)
	}
	if condition.UpdateUser != "" {
		sql = sql + " and update_user like '%" + condition.UpdateUser + "%'"
	}
	if condition.Show == "me" {
		sql = sql + " and log_metric_group is null"
	}
	if condition.UpdatedTimeStart != "" && condition.UpdatedTimeEnd != "" {
		sql = sql + " and update_time >= ? and update_time <= ?"
		params = append(params, condition.UpdatedTimeStart, condition.UpdatedTimeEnd)
	}
	idsFilterSql, idsFilterParam := createListParams(ids, "")
	sql = sql + " and guid  in (" + idsFilterSql + ")"
	params = append(params, idsFilterParam...)
	sql = sql + " order by update_time desc "
	pageInfo.StartIndex = condition.StartIndex
	pageInfo.PageSize = condition.PageSize
	pageInfo.TotalRows = queryCount(sql, params...)
	sql = sql + " limit ?,? "
	params = append(params, condition.StartIndex, condition.PageSize)
	err = x.SQL(sql, params...).Find(&list)
	return
}

// CopyCustomChart 复制图表
func CopyCustomChart(dashboardId int, user, group string, chart *models.CustomChart, displayConfig interface{}) (newChartId string, err error) {
	var chartSeriesList []*models.CustomChartSeries
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	var actions []*Action
	var chartName string
	newChartId = guid.CreateGuid()
	byteConf, _ := json.Marshal(displayConfig)
	now := time.Now().Format(models.DatetimeFormat)
	if err = x.SQL("select * from custom_chart_series where dashboard_chart = ?", chart.Guid).Find(&chartSeriesList); err != nil {
		return
	}
	if len(chartSeriesList) == 0 {
		chartSeriesList = []*models.CustomChartSeries{}
	}
	if configMap, err = QueryAllChartSeriesConfig(); err != nil {
		return
	}
	if tagMap, err = QueryAllChartSeriesTag(); err != nil {
		return
	}
	if tagValueMap, err = QueryAllChartSeriesTagValue(); err != nil {
		return
	}
	chartName = getNewChartName(chart.Name)
	actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit,create_user,update_user,create_time,update_time,chart_template,pie_type) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		newChartId, dashboardId, 0, chartName, chart.ChartType, chart.LineType, chart.Aggregate,
		chart.AggStep, chart.Unit, user, user, now, now, chart.ChartTemplate, chart.PieType}})
	for _, series := range chartSeriesList {
		seriesId := guid.CreateGuid()
		actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid)values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			seriesId, newChartId, series.Endpoint, series.ServiceGroup, series.EndpointName, series.MonitorType, series.Metric, series.ColorGroup,
			series.PieDisplayTag, series.EndpointType, series.MetricType, series.MetricGuid}})
		if confArr, ok := configMap[series.Guid]; ok {
			if len(confArr) > 0 {
				for _, config := range confArr {
					actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{
						guid.CreateGuid(), seriesId, config.Tags, config.Color, config.SeriesName,
					}})
				}
			}
		}
		if tagArr, ok := tagMap[series.Guid]; ok {
			if len(tagArr) > 0 {
				for _, tag := range tagArr {
					newTagId := guid.CreateGuid()
					actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{
						newTagId, seriesId, tag.Name, tag.Equal}})
					if tagValueArr, ok2 := tagValueMap[tag.Guid]; ok2 {
						if len(tagValueArr) > 0 {
							for _, tagValue := range tagValueArr {
								actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{newTagId, tagValue.Value}})
							}
						}
					}
				}
			}
		}
	}
	actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time) values(?,?,?,?,?,?,?,?,?)", Param: []interface{}{guid.CreateGuid(),
		dashboardId, newChartId, group, string(byteConf), user, user, now, now}})
	actions = append(actions, &Action{Sql: "update custom_dashboard set update_at=?,update_user=? where id=?", Param: []interface{}{now, user, dashboardId}})
	err = Transaction(actions)
	return
}

func getNewChartName(name string) string {
	var layout = "060102150405"
	if strings.TrimSpace(name) != "" {
		start := strings.LastIndex(name, "-")
		if start == -1 {
			// 没有匹配到 - ,直接拼接时间搓
			name = name + "-" + time.Now().Format(layout)
			return name
		}
		suffix := name[start+1:]
		if len(suffix) == 12 && start > 0 {
			name = name[:start] + "-" + time.Now().Format(layout)
		} else {
			name = name + "-" + time.Now().Format(layout)
		}
		return name
	}
	return time.Now().Format(layout)
}

func UpdateCustomChartName(chartId, name, user string, sourceDashboard int) (err error) {
	var actions []*Action
	now := time.Now().Format(models.DatetimeFormat)
	actions = append(actions, &Action{Sql: "update custom_chart set name = ?,update_user = ?,update_time=? where guid = ?", Param: []interface{}{name, user, now, chartId}})
	// 更新源看板
	if sourceDashboard != 0 {
		actions = append(actions, &Action{Sql: "update custom_dashboard set update_user =?,update_at=? where id = ?", Param: []interface{}{user, now, sourceDashboard}})
	}
	return Transaction(actions)
}

func QueryCustomChartNameExist(name string) (list []*models.CustomChart, err error) {
	err = x.SQL("select * from custom_chart where name = ? and public = 1", name).Find(&list)
	return
}

func CreateCustomChartDto(chartExtend *models.CustomChartExtend, configMap map[string][]*models.CustomChartSeriesConfig, tagMap map[string][]*models.CustomChartSeriesTag, tagValueMap map[string][]*models.CustomChartSeriesTagValue) (chart *models.CustomChartDto, err error) {
	var list []*models.CustomChartSeries
	var seriesConfigList []*models.CustomChartSeriesConfig
	var chartSeriesTagList []*models.CustomChartSeriesTag
	var chartSeriesTagValueList []*models.CustomChartSeriesTagValue
	chart = &models.CustomChartDto{
		Id:                 chartExtend.Guid,
		Public:             intToBool(chartExtend.Public),
		SourceDashboard:    chartExtend.SourceDashboard,
		Name:               chartExtend.Name,
		ChartTemplate:      chartExtend.ChartTemplate,
		Unit:               chartExtend.Unit,
		ChartType:          chartExtend.ChartType,
		LineType:           chartExtend.LineType,
		PieType:            chartExtend.PieType,
		Aggregate:          chartExtend.Aggregate,
		AggStep:            chartExtend.AggStep,
		DisplayConfig:      chartExtend.DisplayConfig,
		GroupDisplayConfig: chartExtend.GroupDisplayConfig,
		Group:              chartExtend.Group,
		LogMetricGroup:     &chartExtend.LogMetricGroup,
	}
	chart.ChartSeries = []*models.CustomChartSeriesDto{}
	if list, err = QueryCustomChartSeriesByChart(chartExtend.Guid); err != nil {
		return
	}
	if len(list) > 0 {
		for _, series := range list {
			seriesConfigList = []*models.CustomChartSeriesConfig{}
			chartSeriesTagList = []*models.CustomChartSeriesTag{}
			customChartSeriesDto := &models.CustomChartSeriesDto{
				Guid:          series.Guid,
				Endpoint:      series.Endpoint,
				ServiceGroup:  series.ServiceGroup,
				EndpointName:  series.EndpointName,
				MonitorType:   series.MonitorType,
				ColorGroup:    series.ColorGroup,
				PieDisplayTag: series.PieDisplayTag,
				EndpointType:  series.EndpointType,
				MetricType:    series.MetricType,
				MetricGuid:    series.MetricGuid,
				Metric:        series.Metric,
				Comparison:    false,
				Tags:          make([]*models.TagDto, 0),
				ColorConfig:   make([]*models.ColorConfigDto, 0),
			}
			// 判断是否是同环比
			var tempGuid string
			_, _ = x.SQL("select guid from metric_comparison where metric_id = ?", series.MetricGuid).Get(&tempGuid)
			if tempGuid != "" {
				customChartSeriesDto.Comparison = true
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
						Equal:    tag.Equal,
						TagValue: getChartSeriesTagValues(chartSeriesTagValueList),
					})
				}
			}
			if len(seriesConfigList) > 0 {
				for _, config := range seriesConfigList {
					customChartSeriesDto.ColorConfig = append(customChartSeriesDto.ColorConfig, &models.ColorConfigDto{
						SeriesName: config.SeriesName,
						Color:      config.Color,
					})
				}
			}
			chart.ChartSeries = append(chart.ChartSeries, customChartSeriesDto)
		}
		if len(chart.ChartSeries) > 0 {
			sort.Sort(models.CustomChartSeriesDtoSort(chart.ChartSeries))
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

func getChartQueryIdsByPermission(condition models.QueryChartParam, roles []string) (ids []string, err error) {
	var sql = "select dashboard_chart from custom_chart_permission where 1=1 "
	var params []interface{}
	if len(roles) == 0 {
		return
	}
	ids = []string{}
	if len(condition.UseRoles) == 0 && len(condition.MgmtRoles) == 0 {
		roleFilterSql, roleFilterParam := createListParams(roles, "")
		sql = sql + " and role_id  in (" + roleFilterSql + ")"
		params = append(params, roleFilterParam...)
		if condition.Permission == string(models.PermissionMgmt) {
			sql = sql + " and permission = ? "
			params = append(params, models.PermissionMgmt)
		}
	} else {
		if len(condition.UseRoles) > 0 {
			useRoleFilterSql, useRoleFilterParam := createListParams(condition.UseRoles, "")
			sql = sql + " and (role_id  in (" + useRoleFilterSql + ") and permission = ?)"
			params = append(append(params, useRoleFilterParam...), models.PermissionUse)
		}
		if len(condition.MgmtRoles) > 0 {
			mgmtRoleFilterSql, mgmtRoleFilterParam := createListParams(condition.MgmtRoles, "")
			sql = sql + " and (role_id  in (" + mgmtRoleFilterSql + ") and permission = ?)"
			params = append(append(params, mgmtRoleFilterParam...), models.PermissionMgmt)
		}
		roleFilterSql, roleFilterParam := createListParams(roles, "")
		if condition.Permission == string(models.PermissionMgmt) {
			sql = sql + " and dashboard_chart in (select dashboard_chart from custom_chart_permission where role_id in (" + roleFilterSql + ") and  and permission = ?)"
			params = append(append(params, roleFilterParam...), models.PermissionMgmt)
		} else {
			sql = sql + " and dashboard_chart in (select dashboard_chart from custom_chart_permission where role_id in (" + roleFilterSql + "))"
			params = append(params, roleFilterParam...)
		}
	}
	if err = x.SQL(sql, params...).Find(&ids); err != nil {
		return
	}
	// 应用看板,需要做ID交集
	if len(condition.UseDashboard) > 0 {
		var tempIds, newIds []string
		strArr := strings.Join(convertIntArrToStr(condition.UseDashboard), ",")
		if err = x.SQL("select dashboard_chart from custom_dashboard_chart_rel where custom_dashboard  in (" + strArr + ")").Find(&tempIds); err != nil {
			return
		}
		if len(tempIds) > 0 {
			for _, id := range ids {
				for _, tempId := range tempIds {
					if id == tempId {
						newIds = append(newIds, id)
						break
					}
				}
			}
			return filterRepeatIds(newIds), nil
		} else {
			return []string{}, nil
		}
	}
	return filterRepeatIds(ids), nil
}

func filterRepeatIds(ids []string) []string {
	var newIds []string
	if len(ids) == 0 {
		return newIds
	}
	var hashMap = make(map[string]bool)
	for _, id := range ids {
		hashMap[id] = true
	}
	for key, _ := range hashMap {
		newIds = append(newIds, key)
	}
	return newIds
}

func convertIntArrToStr(ids []int) []string {
	var arr []string
	if len(ids) == 0 {
		return arr
	}
	for _, id := range ids {
		arr = append(arr, fmt.Sprintf("%d", id))
	}
	return arr
}

func GetChartSeriesConfig(customChartSeriesGuid string) (result []*models.CustomChartSeriesConfig, err error) {
	if err = x.SQL("select * from custom_chart_series_config where dashboard_chart_config=?", customChartSeriesGuid).Find(&result); err != nil {
		err = fmt.Errorf("query custom chart series config table fail,%s ", err.Error())
	}
	return
}
