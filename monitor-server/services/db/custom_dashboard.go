package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"sort"
	"strconv"
	"strings"
	"time"
)

// makeInClauseFromStrings builds a SQL IN clause placeholder string like "(?, ?, ?)" and the corresponding args.
// It assumes the caller has validated that values is non-empty.
func makeInClauseFromStrings(values []string) (string, []interface{}) {
	if len(values) == 0 {
		return "(?)", []interface{}{""}
	}
	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, v := range values {
		placeholders[i] = "?"
		args[i] = v
	}
	return "(" + strings.Join(placeholders, ",") + ")", args
}

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
	sql = sql + " order by name  COLLATE utf8_bin ASC "
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
		if err = x.SQL(sql, params...).Find(&ids); err != nil {
			return
		}
		if len(ids) > 0 {
			strArr = TransformInToStrArray(ids)
		}
	} else {
		var useIds, mgmtIds []int
		originSql := sql
		if len(condition.UseRoles) > 0 {
			var tempParams []interface{}
			useRoleFilterSql, useRoleFilterParam := createListParams(condition.UseRoles, "")
			sql = originSql + " and (role_id  in (" + useRoleFilterSql + ") and permission = ?)"
			tempParams = append(append(tempParams, useRoleFilterParam...), models.PermissionUse)
			if err = x.SQL(sql, tempParams...).Find(&useIds); err != nil {
				return
			}
		}
		if len(condition.MgmtRoles) > 0 {
			var tempParams []interface{}
			mgmtRoleFilterSql, mgmtRoleFilterParam := createListParams(condition.MgmtRoles, "")
			sql = originSql + " and (role_id  in (" + mgmtRoleFilterSql + ") and permission = ?)"
			tempParams = append(append(tempParams, mgmtRoleFilterParam...), models.PermissionMgmt)
			if err = x.SQL(sql, tempParams...).Find(&mgmtIds); err != nil {
				return
			}
		}
		roleFilterSql, roleFilterParam := createListParams(roles, "")
		if condition.Permission == string(models.PermissionMgmt) {
			sql = originSql + " and custom_dashboard_id in (select custom_dashboard_id from custom_dashboard_role_rel where role_id in (" + roleFilterSql + ") and permission = ?)"
			params = append(append(params, roleFilterParam...), models.PermissionMgmt)
		} else {
			sql = originSql + " and custom_dashboard_id in (select custom_dashboard_id from custom_dashboard_role_rel where role_id in (" + roleFilterSql + "))"
			params = append(params, roleFilterParam...)
		}
		if err = x.SQL(sql, params...).Find(&ids); err != nil {
			return
		}
		useIdsStr := convertIntArrToStr(useIds)
		mgmtIdsStr := convertIntArrToStr(mgmtIds)
		idsStr := convertIntArrToStr(ids)
		if len(condition.UseRoles) == 0 {
			strArr = mergeArray(mgmtIdsStr, idsStr)
		} else if len(condition.MgmtRoles) == 0 {
			strArr = mergeArray(useIdsStr, idsStr)
		} else {
			strArr = mergeArray(useIdsStr, mgmtIdsStr, idsStr)
		}
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

// filterRepeatStringIds 去重函数
func filterRepeatStringIds(ids []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// mergeArray 泛型函数，合并多个切片，并返回在所有输入切片中都出现的元素
func mergeArray(ids ...[]string) []string {
	// 创建一个map来记录每个元素出现的次数
	countMap := make(map[string]int)

	// 遍历所有输入的切片
	for _, idList := range ids {
		idList = filterRepeatStringIds(idList)
		for _, id := range idList {
			countMap[id]++
		}
	}

	// 创建一个切片来存储重复的元素
	var duplicates []string

	// 遍历map，找到出现次数等于 len(ids) 的元素
	for id, count := range countMap {
		if count == len(ids) {
			duplicates = append(duplicates, id)
		}
	}

	return duplicates
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

func CopyCustomDashboard(param models.CopyCustomDashboardParam, customDashboard *models.CustomDashboardTable, operator string, errMsgObj *models.ErrorTemplate) (err error) {
	var result sql.Result
	var newDashboardId int64
	var actions, subDashboardPermActions, subDashboardChartActions []*Action
	var customChartExtendList []*models.CustomChartExtend
	var exportDto = &models.CustomDashboardExportDto{Charts: make([]*models.CustomChartDto, 0)}
	var chart *models.CustomChartDto
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	var metricComparisonMap = make(map[string]string)
	var chartSeriesMap = make(map[string][]*models.CustomChartSeries)
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
	if metricComparisonMap, err = GetAllMetricComparison(); err != nil {
		return
	}
	if customChartExtendList, err = QueryCustomChartListByDashboard(customDashboard.Id); err != nil {
		return
	}
	if len(customChartExtendList) > 0 {
		var chartSeries []*models.CustomChartSeries
		// 图表大于等于10时候 查询所有图表数据
		if len(customChartExtendList) >= 10 {
			if chartSeriesMap, err = QueryAllChartSeries(); err != nil {
				return
			}
		}
		for _, chartExtend := range customChartExtendList {
			if len(chartSeriesMap) > 0 {
				chartSeries = chartSeriesMap[chartExtend.Guid]
			}
			chartParam := models.CreateCustomChartParam{
				ChartExtend:         chartExtend,
				ConfigMap:           configMap,
				TagMap:              tagMap,
				TagValueMap:         tagValueMap,
				MetricComparisonMap: metricComparisonMap,
				ChartSeries:         chartSeries,
			}
			if chart, err = CreateCustomChartDto(chartParam); err != nil {
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

func ImportCustomDashboard(param *models.CustomDashboardExportDto, operator, rule, mgmtRole string, useRoles []string, errMsgObj *models.ErrorTemplate) (customDashboard *models.CustomDashboardTable, importRes *models.CustomDashboardImportRes, err error) {
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
			err = errMsgObj.ImportDashboardNameExistError.WithParam(param.Name)
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

func deleteCustomDashboard(customDashboardId int64) {
	var err error
	if _, err = x.Exec("delete from custom_dashboard where id=?", customDashboardId); err != nil {
		log.Error(nil, log.LOGGER_APP, "deleteCustomDashboard fail", zap.Error(err))
	}
	return
}

func deleteCustomDashboardList(customDashboardIdList []int64) {
	var err error
	for _, id := range customDashboardIdList {
		if _, err = x.Exec("delete from custom_dashboard where id=?", id); err != nil {
			log.Error(nil, log.LOGGER_APP, "deleteCustomDashboard fail", zap.Error(err))
		}
	}
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

func BatchGetCustomDashboardByIds(ids []string) (list []*models.CustomDashboardTable, err error) {
	err = x.SQL(fmt.Sprintf("select id,name,update_user,update_at from custom_dashboard where  id in (%s)", strings.Join(ids, ","))).Find(&list)
	for _, dashboard := range list {
		dashboard.UpdateAtStr = dashboard.UpdateAt.Format(models.DatetimeFormat)
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

// SyncDashboardForCodeChanges updates dashboard charts under a log_metric_group when
// code tag values are renamed, deleted, or added. This function is safe to call when no dashboards exist.
func SyncDashboardForCodeChanges(logMetricGroupGuid string, codeRenames map[string]string, codeDeletes []string, codesAdded []string, operator, sucCode string) (err error) {
	var dashboardInfo models.CustomDashboardTable
	if _, err = x.SQL("select id,panel_groups from custom_dashboard where log_metric_group=? limit 1", logMetricGroupGuid).Get(&dashboardInfo); err != nil {
		return
	}
	dashboardId := dashboardInfo.Id

	var chartRows []models.CustomChart
	if err = x.SQL("select guid,name,public from custom_chart where log_metric_group=?", logMetricGroupGuid).Find(&chartRows); err != nil {
		return err
	}

	// build actions
	var actions []*Action

	// Handle renames
	for oldVal, newVal := range codeRenames {
		matchedAny := false
		for _, ch := range chartRows {
			idx := strings.LastIndex(ch.Name, "-")
			if idx > 0 {
				prefix := ch.Name[:idx]
				suffix := ch.Name[idx:]
				if strings.Contains(prefix, oldVal) {
					newPrefix := strings.Replace(prefix, oldVal, newVal, -1)
					newName := newPrefix + suffix
					actions = append(actions, &Action{Sql: "update custom_chart set name=? where guid=?", Param: []interface{}{newName, ch.Guid}})
					matchedAny = true
				}
			}
		}

		// 2) series config series_name by id
		var seriesCfgRows []models.CustomChartSeriesConfig
		if err = x.SQL("select guid,series_name from custom_chart_series_config where dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=?))", logMetricGroupGuid).Find(&seriesCfgRows); err != nil {
			return err
		}
		for _, cfg := range seriesCfgRows {
			if cfg.SeriesName == oldVal {
				actions = append(actions, &Action{Sql: "update custom_chart_series_config set series_name=? where guid=?", Param: []interface{}{newVal, cfg.Guid}})
				matchedAny = true
				continue
			}
			needle := "{code=" + oldVal + "}"
			if strings.Contains(cfg.SeriesName, needle) {
				newSeriesName := strings.Replace(cfg.SeriesName, needle, "{code="+newVal+"}", -1)
				actions = append(actions, &Action{Sql: "update custom_chart_series_config set series_name=? where guid=?", Param: []interface{}{newSeriesName, cfg.Guid}})
				matchedAny = true
			}
		}

		var codeTagRows []models.CustomChartSeriesTag
		if err = x.SQL("select guid from custom_chart_series_tag where name='code' and dashboard_chart_config in (select guid from custom_chart_series where dashboard_chart in (select guid from custom_chart where log_metric_group=?))", logMetricGroupGuid).Find(&codeTagRows); err != nil {
			return err
		}
		if len(codeTagRows) > 0 {
			var codeTagIds []string
			for _, tr := range codeTagRows {
				codeTagIds = append(codeTagIds, tr.Guid)
			}
			inPh, inArgs := makeInClauseFromStrings(codeTagIds)
			sql := "update custom_chart_series_tagvalue set value=? where dashboard_chart_tag in " + inPh + " and value=?"
			params := append([]interface{}{newVal}, inArgs...)
			params = append(params, oldVal)
			actions = append(actions, &Action{Sql: sql, Param: params})
			matchedAny = true
		}

		if !matchedAny {
			log.Warn(nil, log.LOGGER_APP, "no rename match in SyncDashboardForCodeChanges", zap.String("lmg", logMetricGroupGuid), zap.String("old", oldVal), zap.String("new", newVal))
		}
	}

	// Handle deletes: delete charts that exclusively represent this code
	for _, delCode := range codeDeletes {
		// 1) locate charts for this code, but exclude other-charts
		var delChatIds []string
		for _, ch := range chartRows {
			// Only delete dedicated charts for this code, not other-charts
			if !strings.HasPrefix(ch.Name, constOther+"-") && strings.Contains(ch.Name, delCode) {
				delChatIds = append(delChatIds, ch.Guid)
			}
		}
		if len(delChatIds) == 0 {
			continue
		}

		//  图表如果被别的看板引用,则不能删除，需要报错
		for _, chartId := range delChatIds {
			var refCount int
			x.SQL("select count(1) from custom_dashboard_chart_rel where dashboard_chart=? and custom_dashboard not in (?)", chartId, dashboardId).Get(&refCount)
			if refCount > 0 {
				return fmt.Errorf("can not delete public chart %s referenced by other dashboards for code %s", chartId, delCode)
			}
			deleteActions, err := GetDeleteCustomDashboardChart(chartId)
			if err != nil {
				return err
			}
			actions = append(actions, deleteActions...)
		}
	}

	// Adjust "other-" charts
	// 1) identify other charts in this group
	var otherChartIds []string
	for _, ch := range chartRows {
		if strings.HasPrefix(ch.Name, constOther+"-") {
			otherChartIds = append(otherChartIds, ch.Guid)
		}
	}
	log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges other-charts found", zap.String("lmg", logMetricGroupGuid), zap.Strings("otherChartIds", otherChartIds), zap.Strings("codeDeletes", codeDeletes), zap.Strings("codesAdded", codesAdded))
	if len(otherChartIds) > 0 {
		// 2) series under other charts
		var otherSeriesIds []string
		inPhCharts, inArgsCharts := makeInClauseFromStrings(otherChartIds)
		if err = x.SQL("select guid from custom_chart_series where dashboard_chart in "+inPhCharts, inArgsCharts...).Find(&otherSeriesIds); err != nil {
			return err
		}
		// 3) code tag ids under those series
		var otherCodeTagIds []string
		if len(otherSeriesIds) > 0 {
			inPhSeries, inArgsSeries := makeInClauseFromStrings(otherSeriesIds)
			if err = x.SQL("select guid from custom_chart_series_tag where name='code' and dashboard_chart_config in "+inPhSeries, inArgsSeries...).Find(&otherCodeTagIds); err != nil {
				return err
			}
		}
		if len(otherCodeTagIds) > 0 {
			inPhTags, inArgsTags := makeInClauseFromStrings(otherCodeTagIds)
			log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges updating other-charts NotIn list", zap.String("lmg", logMetricGroupGuid), zap.Strings("otherCodeTagIds", otherCodeTagIds), zap.Strings("codesAdded", codesAdded), zap.Strings("codeDeletes", codeDeletes))

			// remove newly added codes from NotIn list of other-charts
			for _, c := range codesAdded {
				params := append([]interface{}{c}, inArgsTags...)
				params = append(params, c)
				actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) select t.guid,? from custom_chart_series_tag t where t.guid in " + inPhTags + " and not exists (select 1 from custom_chart_series_tagvalue v where v.dashboard_chart_tag=t.guid and v.value=?)", Param: params})
				log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges adding code to other-charts NotIn", zap.String("code", c))
			}
			// add deleted codes back to NotIn of other-charts
			for _, c := range codeDeletes {
				actions = append(actions, &Action{Sql: "delete from custom_chart_series_tagvalue where dashboard_chart_tag in " + inPhTags + " and value=?", Param: append(inArgsTags, c)})
				log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges removing code from other-charts NotIn", zap.String("code", c))
			}
		} else {
			log.Warn(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges no other-charts code tags found", zap.String("lmg", logMetricGroupGuid))
		}
	}

	// 看板分组更新
	{
		panelGroupsStr := dashboardInfo.PanelGroups
		// parse current groups
		groupSet := make(map[string]bool)
		var groups []string
		if strings.TrimSpace(panelGroupsStr) != "" {
			for _, g := range strings.Split(panelGroupsStr, ",") {
				g = strings.TrimSpace(g)
				if g == "" {
					continue
				}
				if !groupSet[g] {
					groupSet[g] = true
					groups = append(groups, g)
				}
			}
		}
		// ensure other exists but handle at the end
		delete(groupSet, constOther)
		var filtered []string
		for _, g := range groups {
			if g != constOther {
				filtered = append(filtered, g)
			}
		}

		// apply renames - 处理分组名称的重命名
		for oldVal, newVal := range codeRenames {
			for i, g := range filtered {
				if g == oldVal {
					filtered[i] = newVal
					log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges updating panel group name",
						zap.String("lmg", logMetricGroupGuid),
						zap.String("oldGroup", oldVal),
						zap.String("newGroup", newVal))
					break
				}
			}
		}

		// apply deletes - 从 filtered 中删除被删除的分组
		for _, c := range codeDeletes {
			for i, g := range filtered {
				if g == c {
					// 从 filtered 中删除这个分组
					filtered = append(filtered[:i], filtered[i+1:]...)
					break
				}
			}
		}

		// apply adds - 添加新的分组
		for _, c := range codesAdded {
			// 检查是否已经存在
			exists := false
			for _, g := range filtered {
				if g == c {
					exists = true
					break
				}
			}
			if !exists {
				filtered = append(filtered, c)
			}
		}
		// append other at the end always
		filtered = append(filtered, constOther)
		newPanelGroups := strings.Join(filtered, ",")
		if newPanelGroups != panelGroupsStr {
			// update dashboard panel_groups and audit update user/time
			now := time.Now().Format(models.DatetimeFormat)
			actions = append(actions, &Action{Sql: "update custom_dashboard set panel_groups=?, update_user=?, update_at=? where id=?", Param: []interface{}{newPanelGroups, operator, now, dashboardId}})
			log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges updating panel groups",
				zap.String("lmg", logMetricGroupGuid),
				zap.String("oldGroups", panelGroupsStr),
				zap.String("newGroups", newPanelGroups))
		}
	}

	// Recalculate chart coordinates for all charts in this dashboard
	// Get all charts for this log_metric_group
	var allChartRows []models.CustomChart
	if err = x.SQL("select guid,name from custom_chart where log_metric_group=?", logMetricGroupGuid).Find(&allChartRows); err != nil {
		return err
	}

	// Group charts by their code and sort by metric type to ensure correct order
	codeToCharts := make(map[string][]models.CustomChart)
	for _, chart := range allChartRows {
		// Extract code from chart name (format: code-metric/serviceGroup)
		parts := strings.Split(chart.Name, "-")
		if len(parts) > 0 {
			code := parts[0]
			codeToCharts[code] = append(codeToCharts[code], chart)
		}
	}

	// Sort charts within each code group by metric type (req_count, req_suc_rate, req_costtime_avg)
	for code, charts := range codeToCharts {
		sort.Slice(charts, func(i, j int) bool {
			// Extract metric from chart name and sort by priority
			metricI := extractMetricFromChartName(charts[i].Name)
			metricJ := extractMetricFromChartName(charts[j].Name)

			// Define sort order: req_count (0), req_suc_rate (1), req_costtime_avg (2)
			orderI := getMetricSortOrder(metricI)
			orderJ := getMetricSortOrder(metricJ)
			return orderI < orderJ
		})
		codeToCharts[code] = charts
	}

	// Get current panel_groups to determine order
	var currentPanelGroups string
	x.SQL("select panel_groups from custom_dashboard where id=?", dashboardId).Get(&currentPanelGroups)

	// Parse panel_groups to get the order
	var orderedCodes []string
	if strings.TrimSpace(currentPanelGroups) != "" {
		for _, g := range strings.Split(currentPanelGroups, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				orderedCodes = append(orderedCodes, g)
			}
		}
	}

	// Handle additions: create new charts for added codes
	// 收集新增图表的 GUID 信息，用于后续坐标重新计算
	var newChartsByCode = make(map[string][]models.CustomChart)

	if len(codesAdded) > 0 {
		// Create charts for each added code using autoGenerateCustomDashboard logic
		// Find the position of each added code in the final panel_groups order
		for _, code := range codesAdded {
			// Find the index of this code in the orderedCodes
			codeIndex := -1
			for i, orderedCode := range orderedCodes {
				if orderedCode == code {
					codeIndex = i
					break
				}
			}
			if codeIndex == -1 {
				// This shouldn't happen, but fallback to end
				codeIndex = len(orderedCodes) - 1
			}

			chartActions, createdCharts, err := createChartForCode(logMetricGroupGuid, code, operator, dashboardId, codeIndex, sucCode)
			if err != nil {
				return err
			}
			actions = append(actions, chartActions...)

			// 直接使用 createChartForCode 返回的图表信息
			newChartsByCode[code] = createdCharts
		}
	}

	// 最后统一重新计算所有图表的坐标（包括新增的图表）
	// 这确保了所有图表（包括新增的图表）都有正确的坐标
	if len(codeRenames) > 0 || len(codeDeletes) > 0 || len(codesAdded) > 0 {
		log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges recalculating ALL chart coordinates after all changes",
			zap.String("lmg", logMetricGroupGuid),
			zap.Int("renameCount", len(codeRenames)),
			zap.Int("deleteCount", len(codeDeletes)),
			zap.Int("addCount", len(codesAdded)))

		// 构建最终的图表映射，包含所有图表（现有 + 新增）
		finalCodeToCharts := make(map[string][]models.CustomChart)

		// 1. 先添加现有的图表（排除被删除的）
		for code, charts := range codeToCharts {
			// 检查这个代码是否被删除了
			isDeleted := false
			for _, deletedCode := range codeDeletes {
				if code == deletedCode {
					isDeleted = true
					break
				}
			}
			if !isDeleted {
				finalCodeToCharts[code] = charts
			}
		}

		// 2. 添加新增的图表
		// 使用之前收集的新增图表信息
		for _, code := range codesAdded {
			if newCharts, exists := newChartsByCode[code]; exists {
				finalCodeToCharts[code] = newCharts
			}
		}

		// 按 metric type 排序每个代码组的图表
		for code, charts := range finalCodeToCharts {
			sort.Slice(charts, func(i, j int) bool {
				metricI := extractMetricFromChartName(charts[i].Name)
				metricJ := extractMetricFromChartName(charts[j].Name)
				orderI := getMetricSortOrder(metricI)
				orderJ := getMetricSortOrder(metricJ)
				return orderI < orderJ
			})
			finalCodeToCharts[code] = charts
		}

		// 按 orderedCodes 顺序重新计算所有图表的坐标
		chartIndex := 0
		for _, code := range orderedCodes {
			if charts, exists := finalCodeToCharts[code]; exists {
				for i, chart := range charts {
					displayConfig := calcDisplayConfig(chartIndex*3 + i)
					displayConfigBytes, _ := json.Marshal(displayConfig)

					var groupDisplayConfig models.DisplayConfig
					if i == 0 {
						groupDisplayConfig = calcDisplayConfig(0)
					} else {
						groupDisplayConfig = calcDisplayConfig(1)
					}
					groupDisplayConfigBytes, _ := json.Marshal(groupDisplayConfig)

					actions = append(actions, &Action{
						Sql:   "update custom_dashboard_chart_rel set display_config=?, group_display_config=? where dashboard_chart=?",
						Param: []interface{}{string(displayConfigBytes), string(groupDisplayConfigBytes), chart.Guid},
					})
				}
				chartIndex++
			}
		}
	}

	log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges recalculated chart coordinates",
		zap.String("lmg", logMetricGroupGuid),
		zap.Strings("orderedCodes", orderedCodes),
		zap.Int("totalCharts", len(allChartRows)))

	if len(actions) == 0 {
		log.Debug(nil, log.LOGGER_APP, "SyncDashboardForCodeChanges no actions to execute", zap.String("lmg", logMetricGroupGuid))
		return nil
	}

	return Transaction(actions)
}

// createChartForCode creates charts for a specific code using autoGenerateCustomDashboard logic
func createChartForCode(logMetricGroupGuid, code, operator string, dashboardId, chartIndex int, sucCode string) ([]*Action, []models.CustomChart, error) {
	var actions []*Action
	var createdCharts []models.CustomChart

	// Get group context: prefix, monitor, service group & display
	var metricPrefixCode, logMetricMonitor string
	x.SQL("select metric_prefix_code,log_metric_monitor from log_metric_group where guid=?", logMetricGroupGuid).Get(&metricPrefixCode, &logMetricMonitor)
	var serviceGroup, monitorType string
	x.SQL("select service_group,monitor_type from log_metric_monitor where guid=?", logMetricMonitor).Get(&serviceGroup, &monitorType)
	var serviceGroupDisplay string
	x.SQL("select display_name from service_group where guid=?", serviceGroup).Get(&serviceGroupDisplay)
	if serviceGroupDisplay == "" {
		serviceGroupDisplay = serviceGroup
	}

	now := time.Now().Format(models.DatetimeFormat)

	buildMetric := func(base string) string {
		if strings.TrimSpace(metricPrefixCode) == "" {
			return base
		}
		return metricPrefixCode + "_" + base
	}

	generateMetricGuid := func(metric, serviceGroup string) string {
		return fmt.Sprintf("%s__%s", metric, serviceGroup)
	}

	// Chart 1: requests (two series) - 请求量+失败量 柱状图
	chartId1 := guid.CreateGuid()
	chartName1 := fmt.Sprintf("%s-%s/%s", code, buildMetric(constReqCount), serviceGroupDisplay)
	actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit,create_user,update_user,create_time,update_time,chart_template,pie_type,log_metric_group) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		chartId1, dashboardId, 1, chartName1, "bar", "bar", "sum", 60, "", operator, operator, now, now, "one", "", logMetricGroupGuid}})

	// 收集创建的图表信息
	createdCharts = append(createdCharts, models.CustomChart{
		Guid: chartId1,
		Name: chartName1,
	})
	// Calculate display config for chart 1 (requests)
	displayConfig1 := calcDisplayConfig(chartIndex*3 + 0)
	displayConfig1Bytes, _ := json.Marshal(displayConfig1)
	groupDisplayConfig1 := calcDisplayConfig(0)
	groupDisplayConfig1Bytes, _ := json.Marshal(groupDisplayConfig1)

	actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		guid.CreateGuid(), dashboardId, chartId1, code, string(displayConfig1Bytes), operator, operator, now, now, string(groupDisplayConfig1Bytes)}})

	// Get service group info for proper endpoint and color configuration
	var serviceGroupTable models.ServiceGroupTable
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", serviceGroup).Get(&serviceGroupTable)

	// Generate series name and color config for req_count
	seriesName1 := fmt.Sprintf("%s:%s{code=%s}", buildMetric(constReqCount), serviceGroupDisplay, code)
	series1 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{series1, chartId1, serviceGroup, serviceGroup, serviceGroupDisplay, monitorType, buildMetric(constReqCount), "#1a94bc", "", serviceGroupTable.ServiceType, "business", generateMetricGuid(buildMetric(constReqCount), serviceGroup)}})

	// Add series config for req_count
	config1 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{config1, series1, fmt.Sprintf("code=%s", code), "#1a94bc", seriesName1}})

	tag1 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{tag1, series1, constCode, ConstEqualIn}})
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tag1, code}})

	// Generate series name and color config for req_fail_count
	seriesName2 := fmt.Sprintf("%s:%s{code=%s}", buildMetric(constReqFailCount), serviceGroupDisplay, code)
	series2 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{series2, chartId1, serviceGroup, serviceGroup, serviceGroupDisplay, monitorType, buildMetric(constReqFailCount), "#ff6b6b", "", serviceGroupTable.ServiceType, "business", generateMetricGuid(buildMetric(constReqFailCount), serviceGroup)}})

	// Add series config for req_fail_count
	config2 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{config2, series2, fmt.Sprintf("code=%s", code), "#ff6b6b", seriesName2}})

	tag2 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{tag2, series2, constCode, ConstEqualIn}})
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tag2, code}})

	// Chart 2: success rate - 成功率
	chartId2 := guid.CreateGuid()
	chartName2 := fmt.Sprintf("%s-%s/%s", code, buildMetric(constReqSucCount), serviceGroupDisplay)
	actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit,create_user,update_user,create_time,update_time,chart_template,pie_type,log_metric_group) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		chartId2, dashboardId, 1, chartName2, "line", "line", "none", 60, "%", operator, operator, now, now, "one", "", logMetricGroupGuid}})

	// 收集创建的图表信息
	createdCharts = append(createdCharts, models.CustomChart{
		Guid: chartId2,
		Name: chartName2,
	})
	// Calculate display config for chart 2 (success rate)
	displayConfig2 := calcDisplayConfig(chartIndex*3 + 1)
	displayConfig2Bytes, _ := json.Marshal(displayConfig2)
	groupDisplayConfig2 := calcDisplayConfig(1)
	groupDisplayConfig2Bytes, _ := json.Marshal(groupDisplayConfig2)

	actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		guid.CreateGuid(), dashboardId, chartId2, code, string(displayConfig2Bytes), operator, operator, now, now, string(groupDisplayConfig2Bytes)}})

	// Generate series name and color config for req_suc_count
	seriesName3 := fmt.Sprintf("%s:%s{code=%s}", buildMetric(constReqSucCount), serviceGroupDisplay, code)
	series3 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{series3, chartId2, serviceGroup, serviceGroup, serviceGroupDisplay, monitorType, buildMetric(constReqSucCount), "#52c41a", "", serviceGroupTable.ServiceType, "business", generateMetricGuid(buildMetric(constReqSucCount), serviceGroup)}})

	// Add series config for req_suc_count
	config3 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{config3, series3, fmt.Sprintf("code=%s", code), "#52c41a", seriesName3}})

	tag3 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{tag3, series3, constCode, ConstEqualIn}})
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tag3, code}})

	// Chart 3: avg costtime - 平均耗时
	chartId3 := guid.CreateGuid()
	chartName3 := fmt.Sprintf("%s-%s/%s", code, buildMetric(constConstTimeAvg), serviceGroupDisplay)
	actions = append(actions, &Action{Sql: "insert into custom_chart(guid,source_dashboard,public,name,chart_type,line_type,aggregate,agg_step,unit,create_user,update_user,create_time,update_time,chart_template,pie_type,log_metric_group) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		chartId3, dashboardId, 1, chartName3, "line", "line", "none", 60, "ms", operator, operator, now, now, "one", "", logMetricGroupGuid}})

	// 收集创建的图表信息
	createdCharts = append(createdCharts, models.CustomChart{
		Guid: chartId3,
		Name: chartName3,
	})
	// Calculate display config for chart 3 (avg cost time)
	displayConfig3 := calcDisplayConfig(chartIndex*3 + 2)
	displayConfig3Bytes, _ := json.Marshal(displayConfig3)
	groupDisplayConfig3 := calcDisplayConfig(1)
	groupDisplayConfig3Bytes, _ := json.Marshal(groupDisplayConfig3)

	actions = append(actions, &Action{Sql: "insert into custom_dashboard_chart_rel(guid,custom_dashboard,dashboard_chart,`group`,display_config,create_user,updated_user,create_time,update_time,group_display_config) values(?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
		guid.CreateGuid(), dashboardId, chartId3, code, string(displayConfig3Bytes), operator, operator, now, now, string(groupDisplayConfig3Bytes)}})

	// Generate series name and color config for req_costtime_avg
	seriesName4 := fmt.Sprintf("%s:%s{code=%s,retcode=%s}", buildMetric(constConstTimeAvg), serviceGroupDisplay, code, constSuccess)
	series4 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{series4, chartId3, serviceGroup, serviceGroup, serviceGroupDisplay, monitorType, buildMetric(constConstTimeAvg), "#fa8c16", "", serviceGroupTable.ServiceType, "business", generateMetricGuid(buildMetric(constConstTimeAvg), serviceGroup)}})

	// Add series config for req_costtime_avg
	config4 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_config(guid,dashboard_chart_config,tags,color,series_name) values(?,?,?,?,?)", Param: []interface{}{config4, series4, fmt.Sprintf("code=%s,retcode=%s", code, constSuccess), "#fa8c16", seriesName4}})

	tag4 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{tag4, series4, constCode, ConstEqualIn}})
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tag4, code}})

	// 为 retcode 标签创建新的 GUID
	tag5 := guid.CreateGuid()
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tag(guid,dashboard_chart_config,name,equal) values(?,?,?,?)", Param: []interface{}{tag5, series4, constRetCode, ConstEqualIn}})
	actions = append(actions, &Action{Sql: "insert into custom_chart_series_tagvalue(dashboard_chart_tag,value) values(?,?)", Param: []interface{}{tag5, sucCode}})

	return actions, createdCharts, nil
}

// SyncAlarmStrategyForCodeChanges function moved to log_metric.go

// extractMetricFromChartName extracts metric name from chart name
// Format: code-metricPrefixCode_metric/serviceGroup
// Returns the metric part (e.g., "req_count", "req_suc_rate", "req_costtime_avg")
func extractMetricFromChartName(chartName string) string {
	// Split by "-" to get code and metric part
	parts := strings.Split(chartName, "-")
	if len(parts) < 2 {
		return ""
	}

	// Get the metric part (everything after the first "-")
	metricPart := parts[1]

	// Split by "/" to remove service group
	metricParts := strings.Split(metricPart, "/")
	if len(metricParts) == 0 {
		return ""
	}

	// Extract metric from metricPrefixCode_metric format
	metricWithPrefix := metricParts[0]
	// Find the last "_" to separate prefix from metric
	lastUnderscore := strings.LastIndex(metricWithPrefix, "_")
	if lastUnderscore == -1 {
		return metricWithPrefix
	}

	return metricWithPrefix[lastUnderscore+1:]
}

// getMetricSortOrder returns sort order for metrics
// req_count=0, req_suc_rate=1, req_costtime_avg=2
func getMetricSortOrder(metric string) int {
	switch metric {
	case constReqCount:
		return 0
	case constReqSuccessRate:
		return 1
	case constConstTimeAvg:
		return 2
	default:
		return 999 // Unknown metrics go to the end
	}
}
