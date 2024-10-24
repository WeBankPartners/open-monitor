package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"sort"
	"strconv"
	"strings"
	"time"
)

func QueryCustomDashboardList(condition models.CustomDashboardQueryParam, operator string, roles []string) (pageInfo models.PageInfo, list []*models.CustomDashboardTable, err error) {
	var params []interface{}
	var ids []string
	var sql = "select id,name, create_user,update_user,create_at,update_at,log_metric_group from custom_dashboard where 1=1 "
	if ids, err = getQueryIdsByPermission(condition, roles); err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	if condition.Id != 0 {
		sql = sql + " and id = ?"
		params = append(params, condition.Id)
	}
	if condition.Name != "" {
		sql = sql + " and name like '%" + condition.Name + "%'"
	}
	if condition.UpdateUser != "" {
		sql = sql + " and update_user like '%" + condition.UpdateUser + "%'"
	}
	if condition.Show == "me" {
		sql = sql + " and log_metric_group is  null"
	}
	sql = sql + " and id in (" + strings.Join(ids, ",") + ")"
	sql = sql + " order by update_at desc "
	pageInfo.StartIndex = condition.StartIndex
	pageInfo.PageSize = condition.PageSize
	pageInfo.TotalRows = queryCount(sql, params...)
	sql = sql + " limit ?,? "
	params = append(params, condition.StartIndex, condition.PageSize)
	err = x.SQL(sql, params...).Find(&list)
	return
}

func QueryAllCustomDashboard() (list []*models.SimpleCustomDashboardDto, err error) {
	err = x.SQL("select id,name from custom_dashboard").Find(&list)
	return
}

func QueryCustomDashboardRoleRelMap() (dashboardRelMap map[string][]int, err error) {
	var list []*models.CustomDashBoardRoleRel
	dashboardRelMap = make(map[string][]int)
	if err = x.SQL("select role_id,custom_dashboard_id from custom_dashboard_role_rel where permission = ?", models.PermissionUse).Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, dashboardRoleRel := range list {
			if _, ok := dashboardRelMap[dashboardRoleRel.RoleId]; !ok {
				dashboardRelMap[dashboardRoleRel.RoleId] = []int{}
			}
			dashboardRelMap[dashboardRoleRel.RoleId] = append(dashboardRelMap[dashboardRoleRel.RoleId], dashboardRoleRel.CustomDashboard)
		}
	}
	return
}

func QueryAllCustomDashboardNameMap() (resultMap map[int]string, err error) {
	var list []*models.SimpleCustomDashboardDto
	resultMap = make(map[int]string)
	if err = x.SQL("select id,name from custom_dashboard").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, dto := range list {
			resultMap[dto.Id] = dto.Name
		}
	}
	return
}

func QueryCustomDashboardRoleRelByCustomDashboard(dashboardId int) (list []*models.CustomDashBoardRoleRel, err error) {
	err = x.SQL("select * from custom_dashboard_role_rel where custom_dashboard_id = ?", dashboardId).Find(&list)
	return
}

func QueryMainDashboardByCustomDashboard(dashboardId int) (list []*models.MainDashboard, err error) {
	err = x.SQL("select * from main_dashboard where custom_dashboard = ?", dashboardId).Find(&list)
	return
}

// QueryAllRoleDisplayNameMap 查询全量角色显示名map
func QueryAllRoleDisplayNameMap() (roleMap map[string]string, err error) {
	var list []*models.RoleTable
	roleMap = make(map[string]string)
	if err = x.SQL("select * from role").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, role := range list {
			roleMap[role.Name] = role.DisplayName
		}
	}
	return
}

func GetCustomDashboardById(id int) (customDashboard *models.CustomDashboardTable, err error) {
	customDashboard = &models.CustomDashboardTable{}
	_, err = x.SQL("select id,name,create_user,update_user,panel_groups,time_range,refresh_week,log_metric_group from custom_dashboard where id = ?", id).Get(customDashboard)
	return
}

func QueryCustomDashboardListByName(name string) (customDashboardList []*models.CustomDashboardTable, err error) {
	err = x.SQL("select id from custom_dashboard where name = ?", name).Find(&customDashboardList)
	return
}

func GetDashboardPermissionMap(dashboard int, permission string) (permissionMap map[string]bool, err error) {
	var list []*models.CustomDashBoardRoleRel
	permissionMap = make(map[string]bool)
	err = x.SQL("select role_id from custom_dashboard_role_rel where custom_dashboard_id = ? and permission = ?", dashboard, permission).Find(&list)
	if len(list) > 0 {
		for _, perm := range list {
			permissionMap[perm.RoleId] = true
		}
	}
	return
}

func AddCustomDashboard(customDashboard *models.CustomDashboardTable, mgmtRoles, useRoles []string) (insertId int64, err error) {
	var actions []*Action
	actions, insertId, err = getAddCustomDashboardActions(customDashboard, mgmtRoles, useRoles)
	err = Transaction(actions)
	return
}

func getAddCustomDashboardActions(customDashboard *models.CustomDashboardTable, mgmtRoles, useRoles []string) (actions []*Action, insertId int64, err error) {
	var result sql.Result
	actions = []*Action{}
	result, err = x.Exec("insert into custom_dashboard(name,create_user,update_user,create_at,update_at,log_metric_group,time_range,refresh_week,panel_groups) values(?,?,?,?,?,?,?,?,?)", customDashboard.Name, customDashboard.CreateUser, customDashboard.UpdateUser, customDashboard.CreateAt.Format(models.DatetimeFormat),
		customDashboard.UpdateAt.Format(models.DatetimeFormat), customDashboard.LogMetricGroup, customDashboard.TimeRange, customDashboard.RefreshWeek, customDashboard.PanelGroups)
	if err != nil {
		return
	}
	if insertId, err = result.LastInsertId(); err != nil {
		return
	}
	if len(mgmtRoles) > 0 {
		for _, role := range mgmtRoles {
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_role_rel (custom_dashboard_id,permission,role_id)values(?,?,?)",
				Param: []interface{}{insertId, models.PermissionMgmt, role}})
		}
	}
	if len(useRoles) > 0 {
		for _, role := range useRoles {
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_role_rel (custom_dashboard_id,permission,role_id)values(?,?,?)",
				Param: []interface{}{insertId, models.PermissionUse, role}})
		}
	}
	return
}

func AddCustomDashboardChartRel(rel *models.CustomDashboardChartRel) (err error) {
	_, err = x.Exec("insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,"+
		"create_time,update_time) values(?,?,?,?,?,?,?,?,?)", rel.Guid, rel.CustomDashboard, rel.DashboardChart, rel.Group, rel.DisplayConfig, rel.CreateUser,
		rel.UpdateUser, rel.CreateTime, rel.UpdateTime)
	return
}

func QueryCustomDashboardPermissionByDashboard(dashboard int) (list []*models.CustomDashBoardRoleRel, err error) {
	err = x.SQL("select * from custom_dashboard_role_rel where custom_dashboard_id = ?", dashboard).Find(&list)
	return
}

func QueryCustomDashboardManagePermissionByDashboard(dashboard int) (hashMap map[string]string, err error) {
	var list []*models.CustomDashBoardRoleRel
	hashMap = make(map[string]string)
	err = x.SQL("select * from custom_dashboard_role_rel where custom_dashboard_id = ? and permission = ?", dashboard, models.PermissionMgmt).Find(&list)
	if len(list) > 0 {
		for _, roleRel := range list {
			hashMap[roleRel.RoleId] = roleRel.Permission
		}
	}
	return
}

func DeleteCustomDashboardById(dashboard int) (err error) {
	var actions = GetDeleteCustomDashboardByIdActions(dashboard)
	return Transaction(actions)
}

func GetDeleteCustomDashboardByIdActions(dashboard int) (actions []*Action) {
	actions = []*Action{}
	actions = append(actions, &Action{Sql: "delete from main_dashboard where custom_dashboard = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_role_rel where custom_dashboard_id = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where custom_dashboard = ?", Param: []interface{}{dashboard}})
	// 删除以该看板为源看板,并且还没有公开的图表
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_config  where dashboard_chart_config  in(select guid from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0))", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_tagvalue where dashboard_chart_tag in(select guid from custom_chart_series_tag  where dashboard_chart_config  in(select guid from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0)))", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag  where dashboard_chart_config  in(select guid from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0))", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0)", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_permission where dashboard_chart in(select guid from custom_chart where source_dashboard =? and public = 0)", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_chart where source_dashboard = ? and public = 0", Param: []interface{}{dashboard}})

	actions = append(actions, &Action{Sql: "delete from custom_dashboard WHERE id=?", Param: []interface{}{dashboard}})
	return
}

func UpdateCustomDashboardTimeActions(dashboard int, operator string) []*Action {
	var actions []*Action
	actions = append(actions, &Action{Sql: "update custom_dashboard set update_at=?,update_user=? where id=?", Param: []interface{}{time.Now().Format(models.DatetimeFormat), operator, dashboard}})
	return actions
}

func UpdateCustomDashboardTime(dashboard int, operator string) (err error) {
	_, err = x.Exec("update custom_dashboard set update_at=?,update_user=? where id=?", time.Now().Format(models.DatetimeFormat), operator, dashboard)
	return
}

func getQueryIdsByPermission(condition models.CustomDashboardQueryParam, roles []string) (strArr []string, err error) {
	var ids []int
	var sql = "select custom_dashboard_id from custom_dashboard_role_rel where 1=1 "
	var params []interface{}
	strArr = []string{}
	if len(roles) == 0 {
		return
	}
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
			sql = sql + " and custom_dashboard_id in (select custom_dashboard_id from custom_dashboard_role_rel where role_id in (" + roleFilterSql + ") and  and permission = ?)"
			params = append(append(params, roleFilterParam...), models.PermissionMgmt)
		} else {
			sql = sql + " and custom_dashboard_id in (select custom_dashboard_id from custom_dashboard_role_rel where role_id in (" + roleFilterSql + "))"
			params = append(params, roleFilterParam...)
		}
	}
	if err = x.SQL(sql, params...).Find(&ids); err != nil {
		return
	}
	if len(ids) > 0 {
		strArr = TransformInToStrArray(ids)
	}
	return
}

func TransformInToStrArray(ids []int) []string {
	var strMap = make(map[string]bool)
	var stringArray []string
	for _, v := range ids {
		strMap[strconv.Itoa(v)] = true
	}
	for key, _ := range strMap {
		stringArray = append(stringArray, key)
	}
	return stringArray
}

func TransformArrayToMap(strArr []string) map[string]bool {
	var hashMap = make(map[string]bool)
	if len(strArr) > 0 {
		for _, str := range strArr {
			hashMap[str] = true
		}
	}
	return hashMap
}

func TransformMapToArray(hashMap map[string]bool) []string {
	var res []string
	if len(hashMap) > 0 {
		for key, _ := range hashMap {
			if strings.TrimSpace(key) != "" {
				res = append(res, key)
			}
		}
	}
	sort.Strings(res)
	return res
}

func SyncData() (err error) {
	var dashboardList []*models.CustomDashboardTable
	var historyChartList []*models.HistoryChart
	var dashboardChartRelList []*models.CustomDashboardChartRel
	if err = x.SQL("select * from custom_dashboard").Find(&dashboardList); err != nil {
		return
	}
	for _, dashboard := range dashboardList {
		var actions []*Action
		// cfg 为空,直接跳过
		if strings.TrimSpace(dashboard.Cfg) == "" {
			continue
		}
		if err = json.Unmarshal([]byte(dashboard.Cfg), &historyChartList); err != nil {
			return
		}
		// 没数据直接跳过
		if len(historyChartList) == 0 {
			continue
		}
		dashboardChartRelList = []*models.CustomDashboardChartRel{}
		if err = x.SQL("select * from custom_dashboard_chart_rel where custom_dashboard = ?", dashboard.Id).Find(&dashboardChartRelList); err != nil {
			return
		}
		// 已经有看板和图表的关联关系,说明数据已经生成,本次不处理
		if len(dashboardChartRelList) > 0 {
			continue
		}
		now := time.Now().Format(models.DatetimeFormat)
		for _, chart := range historyChartList {
			newChartId := chart.ViewConfig.ID
			group := ""
			displayConfig := ""
			if chart.ViewConfig != nil {
				group = chart.ViewConfig.Group
				config := models.NewDisplayConfig{
					X: chart.ViewConfig.X,
					Y: chart.ViewConfig.Y,
					W: chart.ViewConfig.W,
					H: chart.ViewConfig.H,
				}
				byteArr, _ := json.Marshal(config)
				displayConfig = string(byteArr)
			}
			// 新增图表表
			actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit," +
				"create_user,update_user,create_time,update_time,chart_template) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				newChartId, dashboard.Id, 0, chart.PanalTitle, chart.ChartType, convertLineTypeIntToString(chart.LineType), chart.Aggregate,
				chart.AggStep, chart.PanalUnit, dashboard.CreateUser, dashboard.CreateUser, now, now, ""}})
			// 新增看板图表关系表
			actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time) values(?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				guid.CreateGuid(), dashboard.Id, newChartId, group, displayConfig, dashboard.CreateUser, dashboard.CreateUser, now, now}})
			if len(chart.Query) > 0 {
				for _, series := range chart.Query {
					seriesId := guid.CreateGuid()
					monitorType := ""
					if strings.TrimSpace(series.Endpoint) != "" {
						x.SQL("select monitor_type from endpoint_new where guid=?", series.Endpoint).Get(&monitorType)
					}
					if monitorType == "" {
						// 数据兜底
						monitorType = series.EndpointType
					}
					if series.EndpointName == "" {
						x.SQL("select display_name from service_group where guid=?", series.AppObject).Get(&series.EndpointName)
					}
					actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
						seriesId, newChartId, series.Endpoint, series.AppObject, series.EndpointName, monitorType, series.Metric, series.DefaultColor, "", series.EndpointType, "", ""}})
					if len(series.MetricToColor) > 0 {
						for _, colorConfig := range series.MetricToColor {
							tags := ""
							if strings.Contains(colorConfig.Metric, "{") {
								start := strings.LastIndex(colorConfig.Metric, "{")
								tags = colorConfig.Metric[start+1 : len(colorConfig.Metric)-1]
							}
							actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{
								guid.CreateGuid(), seriesId, tags, colorConfig.Color, colorConfig.Metric,
							}})
						}
					}
				}
			}
		}
		if err = Transaction(actions); err != nil {
			return
		}
	}
	return
}

func CopyCustomDashboard(param models.CopyCustomDashboardParam, customDashboard *models.CustomDashboardTable, operator string, errMsgObj *models.ErrorMessageObj) (err error) {
	var result sql.Result
	var newDashboardId int64
	var actions, subDashboardPermActions, subDashboardChartActions []*Action
	var customChartExtendList []*models.CustomChartExtend
	var exportDto = &models.CustomDashboardExportDto{Charts: make([]*models.CustomChartDto, 0)}
	var chart *models.CustomChartDto
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	now := time.Now()
	// 新增看板
	customDashboard.Name = customDashboard.Name + "(1)"
	if CountCustomDashboardByName(customDashboard.Name) > 0 {
		err = fmt.Errorf("%s", errMsgObj.DashboardNameRepeatError)
		return
	}
	result, err = x.Exec("insert into custom_dashboard(name,panel_groups,create_user,update_user,create_at,update_at,time_range,refresh_week) values(?,?,?,?,?,?,?,?)",
		customDashboard.Name, customDashboard.PanelGroups, operator, operator, now, now, customDashboard.TimeRange, customDashboard.RefreshWeek)
	if err != nil {
		return
	}
	if newDashboardId, err = result.LastInsertId(); err != nil {
		return
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
	// 插入看板权限表
	subDashboardPermActions = GetInsertCustomDashboardRoleRelSQL(int(newDashboardId), []string{param.MgmtRole}, param.UseRoles)
	if len(subDashboardPermActions) > 0 {
		actions = append(actions, subDashboardPermActions...)
	}
	if customChartExtendList, err = QueryCustomChartListByDashboard(customDashboard.Id); err != nil {
		return
	}
	if len(customChartExtendList) > 0 {
		for _, chartExtend := range customChartExtendList {
			if chart, err = CreateCustomChartDto(chartExtend, configMap, tagMap, tagValueMap); err != nil {
				return
			}
			if chart != nil {
				exportDto.Charts = append(exportDto.Charts, chart)
			}
		}
	}
	if subDashboardChartActions, _, err = handleDashboardChart(exportDto, newDashboardId, operator, now.Format(models.DatetimeFormat), param.MgmtRole, param.UseRoles); err != nil {
		return
	}
	if len(subDashboardChartActions) > 0 {
		actions = append(actions, subDashboardChartActions...)
	}
	err = Transaction(actions)
	return
}

func CountCustomDashboardByName(name string) int {
	var count int
	x.SQL("select count(1) from custom_dashboard where name=?", name).Get(&count)
	return count
}

func ImportCustomDashboard(param *models.CustomDashboardExportDto, operator, rule, mgmtRole string, useRoles []string, errMsgObj *models.ErrorMessageObj) (customDashboard *models.CustomDashboardTable, importRes *models.CustomDashboardImportRes, err error) {
	var customDashboardList []*models.CustomDashboardTable
	var actions, subDashboardPermActions, subDashboardChartActions []*Action
	var result sql.Result
	var newDashboardId int64
	var logMetricGroup = sql.NullString{String: param.LogMetricGroup, Valid: param.LogMetricGroup != ""}
	now := time.Now().Format(models.DatetimeFormat)
	importRes = &models.CustomDashboardImportRes{ChartMap: make(map[string][]string)}
	if customDashboardList, err = QueryCustomDashboardListByName(param.Name); err != nil {
		return
	}
	if len(customDashboardList) > 0 {
		// 根据rule 判断导入模式, 同名覆盖 or 同名新增
		if rule == string(models.ImportRuleCover) {
			historyDashboard := customDashboardList[0]
			// 覆盖模式:删除以该看板为源看板,并且还没有公开的图表,删除看板的图表关联关系
			actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where custom_dashboard = ?", Param: []interface{}{historyDashboard.Id}})
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_config  where dashboard_chart_config  in(select guid from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0))", Param: []interface{}{historyDashboard.Id}})
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_tagvalue where dashboard_chart_tag in(select guid from custom_chart_series_tag  where dashboard_chart_config  in(select guid from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0)))", Param: []interface{}{historyDashboard.Id}})
			actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag  where dashboard_chart_config  in(select guid from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0))", Param: []interface{}{historyDashboard.Id}})
			actions = append(actions, &Action{Sql: "delete from custom_chart_series  where dashboard_chart  in(select guid from custom_chart where source_dashboard =? and public = 0)", Param: []interface{}{historyDashboard.Id}})
			actions = append(actions, &Action{Sql: "delete from custom_chart_permission where dashboard_chart in(select guid from custom_chart where source_dashboard =? and public = 0)", Param: []interface{}{historyDashboard.Id}})
			actions = append(actions, &Action{Sql: "delete from custom_chart where source_dashboard = ? and public = 0", Param: []interface{}{historyDashboard.Id}})
			// 更新看板操作人和时间
			actions = append(actions, &Action{Sql: "update custom_dashboard set update_at=?,update_user=? where id=?", Param: []interface{}{now, operator, historyDashboard.Id}})
			if subDashboardChartActions, importRes, err = handleDashboardChart(param, int64(historyDashboard.Id), operator, now, mgmtRole, useRoles); err != nil {
				return
			}
			if len(subDashboardChartActions) > 0 {
				actions = append(actions, subDashboardChartActions...)
			}
			err = Transaction(actions)
			return
		}
		// 同名新增
		param.Name = param.Name + "(1)"
		tempList, _ := QueryCustomDashboardListByName(param.Name)
		if len(tempList) > 0 {
			err = fmt.Errorf(errMsgObj.ImportDashboardNameExistError, param.Name)
			return
		}
	}
	result, err = x.Exec("insert into custom_dashboard(name,panel_groups,create_user,update_user,create_at,update_at,time_range,refresh_week,log_metric_group) values(?,?,?,?,?,?,?,?,?)",
		param.Name, param.PanelGroups, operator, operator, now, now, param.TimeRange, param.RefreshWeek, logMetricGroup)
	if err != nil {
		return
	}
	if newDashboardId, err = result.LastInsertId(); err != nil {
		return
	}
	// 插入看板权限表
	subDashboardPermActions = GetInsertCustomDashboardRoleRelSQL(int(newDashboardId), []string{mgmtRole}, useRoles)
	if len(subDashboardPermActions) > 0 {
		actions = append(actions, subDashboardPermActions...)
	}
	if subDashboardChartActions, importRes, err = handleDashboardChart(param, newDashboardId, operator, now, mgmtRole, useRoles); err != nil {
		return
	}
	if len(subDashboardChartActions) > 0 {
		actions = append(actions, subDashboardChartActions...)
	}
	err = Transaction(actions)
	return
}

func handleDashboardChart(param *models.CustomDashboardExportDto, newDashboardId int64, operator, now, mgmtRole string, useRoles []string) (actions []*Action, importRes *models.CustomDashboardImportRes, err error) {
	var permissionList []*models.CustomChartPermission
	var list []*models.CustomChart
	var logMetricGroup string
	actions = []*Action{}
	importRes = &models.CustomDashboardImportRes{ChartMap: make(map[string][]string)}
	for _, chart := range param.Charts {
		list = []*models.CustomChart{}
		logMetricGroup = ""
		permissionList = []*models.CustomChartPermission{}
		newChartId := guid.CreateGuid()
		// 如果图表公共,则去图表库中根据名称查询是否已有该图表,有的话添加看板的关联关系即可
		if chart.Public {
			if list, err = QueryCustomChartByName(chart.Name); err != nil {
				return
			}
			if len(list) > 0 {
				// 新增看板图表关系表
				actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
					guid.CreateGuid(), newDashboardId, list[0].Guid, chart.Group, chart.DisplayConfig, operator, operator, now, now, chart.GroupDisplayConfig}})
				continue
			}
		}
		// 新增图表和图表配置
		if chart.LogMetricGroup != nil {
			logMetricGroup = *chart.LogMetricGroup
		}
		actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit," +
			"create_user,update_user,create_time,update_time,chart_template,pie_type,log_metric_group) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			newChartId, newDashboardId, chart.Public, chart.Name, chart.ChartType, chart.LineType, chart.Aggregate,
			chart.AggStep, chart.Unit, operator, operator, now, now, chart.ChartTemplate, chart.PieType, logMetricGroup}})
		// 新增看板图表关系表
		actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			guid.CreateGuid(), newDashboardId, newChartId, chart.Group, chart.DisplayConfig, operator, operator, now, now, chart.GroupDisplayConfig}})
		if chart.Public {
			// 新增图表权限
			for _, useRole := range useRoles {
				permissionList = append(permissionList, &models.CustomChartPermission{
					Guid:           guid.CreateGuid(),
					DashboardChart: newChartId,
					RoleId:         useRole,
					Permission:     string(models.PermissionUse),
				})
			}
			permissionList = append(permissionList, &models.CustomChartPermission{
				Guid:           guid.CreateGuid(),
				DashboardChart: newChartId,
				RoleId:         mgmtRole,
				Permission:     string(models.PermissionMgmt),
			})
			actions = append(actions, GetInsertCustomChartPermissionSQL(permissionList)...)
		}
		if len(chart.ChartSeries) > 0 {
			var exist bool
			for _, series := range chart.ChartSeries {
				// 查询每个指标是否存在,不存在需要记录下来
				exist = true
				if series.MetricGuid != "" {
					metricGuid := ""
					if _, err = x.SQL("select guid from metric where guid=?", series.MetricGuid).Get(&metricGuid); err != nil {
						return
					}
					if metricGuid == "" {
						exist = false
					}
				} else {
					// 指标guid为空,则通过metric+MonitorType+ServiceGroup三个都要匹配上则认为指标存在
					var metricList []*models.MetricTable
					if err = x.SQL("select * from metric where metric =? and monitor_type =?", series.Metric, series.MonitorType).Find(&metricList); err != nil {
						return
					}
					if len(metricList) == 0 {
						exist = false
					} else {
						exist = false
						for _, metricTable := range metricList {
							if metricTable.ServiceGroup == series.ServiceGroup {
								exist = true
								break
							}
						}
					}
				}
				if strings.TrimSpace(series.Endpoint) != "" && series.ServiceGroup == "" {
					// 监控对象不存在,记录下来
					var endpointObj models.EndpointTable
					if _, err = x.SQL("SELECT * FROM endpoint_new WHERE guid=?", series.Endpoint).Get(&endpointObj); err != nil {
						return
					}
					if endpointObj.Name == "" {
						exist = false
					}
				}
				// 指标不存在,统计不存在指标返回
				if !exist {
					if len(importRes.ChartMap[chart.Name]) == 0 {
						importRes.ChartMap[chart.Name] = []string{}
					}
					importRes.ChartMap[chart.Name] = append(importRes.ChartMap[chart.Name], series.Metric)
					continue
				}
				seriesId := guid.CreateGuid()
				actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
					seriesId, newChartId, series.Endpoint, series.ServiceGroup, series.EndpointName, series.MonitorType, series.Metric, series.ColorGroup, series.PieDisplayTag, series.EndpointType, series.MetricType, series.MetricGuid}})
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
				if len(series.Tags) > 0 {
					for _, tag := range series.Tags {
						tagId := guid.CreateGuid()
						if tag.Equal == "" {
							tag.Equal = "in"
						}
						actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{
							tagId, seriesId, tag.TagName, tag.Equal}})
						if len(tag.TagValue) > 0 {
							for _, tagValue := range tag.TagValue {
								actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tagId, tagValue}})
							}
						}
					}
				}
			}
		}
	}
	return
}

func deleteCustomDashboard(customDashboardId int64) (err error) {
	_, err = x.Exec("delete from custom_dashboard where id=?", customDashboardId)
	return
}

func handleAutoCreateChart(chart *models.CustomChartDto, newDashboardId int64, useRoles []string, mgmtRole, operator string) (actions []*Action) {
	var permissionList []*models.CustomChartPermission
	var displayConfig, groupDisplayConfig []byte
	newChartId := guid.CreateGuid()
	now := time.Now()
	displayConfig, _ = json.Marshal(chart.DisplayConfig)
	groupDisplayConfig, _ = json.Marshal(chart.GroupDisplayConfig)
	// 新增图表和图表配置
	actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit," +
		"create_user,update_user,create_time,update_time,chart_template,pie_type,log_metric_group) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		newChartId, newDashboardId, chart.Public, chart.Name, chart.ChartType, chart.LineType, chart.Aggregate,
		chart.AggStep, chart.Unit, operator, operator, now, now, chart.ChartTemplate, chart.PieType, chart.LogMetricGroup}})
	// 新增看板图表关系表
	actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		guid.CreateGuid(), newDashboardId, newChartId, chart.Group, displayConfig, operator, operator, now, now, groupDisplayConfig}})
	if chart.Public {
		// 新增图表权限
		for _, useRole := range useRoles {
			permissionList = append(permissionList, &models.CustomChartPermission{
				Guid:           guid.CreateGuid(),
				DashboardChart: newChartId,
				RoleId:         useRole,
				Permission:     string(models.PermissionUse),
			})
		}
		permissionList = append(permissionList, &models.CustomChartPermission{
			Guid:           guid.CreateGuid(),
			DashboardChart: newChartId,
			RoleId:         mgmtRole,
			Permission:     string(models.PermissionMgmt),
		})
		actions = append(actions, GetInsertCustomChartPermissionSQL(permissionList)...)
	}
	if len(chart.ChartSeries) > 0 {
		for _, series := range chart.ChartSeries {
			seriesId := guid.CreateGuid()
			actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
				seriesId, newChartId, series.Endpoint, series.ServiceGroup, series.EndpointName, series.MonitorType, series.Metric, series.ColorGroup, series.PieDisplayTag, series.EndpointType, series.MetricType, series.MetricGuid}})
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
			if len(series.Tags) > 0 {
				for _, tag := range series.Tags {
					tagId := guid.CreateGuid()
					if tag.Equal == "" {
						tag.Equal = "in"
					}
					actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{
						tagId, seriesId, tag.TagName, tag.Equal}})
					if len(tag.TagValue) > 0 {
						for _, tagValue := range tag.TagValue {
							actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tagId, tagValue}})
						}
					}
				}
			}
		}
	}
	return
}

func convertLineTypeIntToString(lineType int) string {
	switch lineType {
	case 1:
		return "line"
	case 0:
		return "area"
	default:
		return "bar"
	}
	return ""
}
