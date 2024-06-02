package db

import (
	"database/sql"
	"encoding/json"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
	"strings"
	"time"
)

func QueryCustomDashboardList(condition models.CustomDashboardQueryParam, roles []string) (pageInfo models.PageInfo, list []*models.CustomDashboardTable, err error) {
	var params []interface{}
	var ids []string
	var sql = "select id,name, create_user,update_user,create_at,update_at from custom_dashboard where 1=1 "
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
		sql = sql + " and update_user = ?"
		params = append(params, condition.UpdateUser)
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
	_, err = x.SQL("select * from custom_dashboard where id = ?", id).Get(customDashboard)
	return
}

func AddCustomDashboard(customDashboard *models.CustomDashboardTable, mgmtRoles, useRoles []string) (err error) {
	var actions []*Action
	var result sql.Result
	var insertId int64
	result, err = x.Exec("insert into custom_dashboard(name,create_user,update_user,create_at,update_at) values(?,?,?,?,?)", customDashboard.Name, customDashboard.CreateUser, customDashboard.UpdateUser, customDashboard.CreateAt.Format(models.DatetimeFormat),
		customDashboard.UpdateAt.Format(models.DatetimeFormat))
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
	return Transaction(actions)
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
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from main_dashboard where custom_dashboard = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_role_rel where custom_dashboard_id = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where custom_dashboard = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard WHERE id=?", Param: []interface{}{dashboard}})
	return Transaction(actions)
}

func getQueryIdsByPermission(condition models.CustomDashboardQueryParam, roles []string) (strArr []string, err error) {
	var ids []int
	var sql = "select custom_dashboard_id from custom_dashboard_role_rel "
	var params []interface{}
	strArr = []string{}
	if len(roles) == 0 {
		return
	}
	roleFilterSql, roleFilterParam := createListParams(roles, "")
	sql = sql + " where role_id  in (" + roleFilterSql + ")"
	params = append(params, roleFilterParam...)

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
					actions = append(actions, &Action{Sql: "insert into custom_chart_series(guid,dashboard_chart,endpoint,service_group,endpoint_name,monitor_type,metric,color_group,pie_display_tag,endpoint_type,metric_type,metric_guid) values(?,?,?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{
						seriesId, newChartId, series.Endpoint, series.AppObject, series.EndpointName, series.EndpointType, series.Metric, series.DefaultColor, "", "", "", ""}})
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
