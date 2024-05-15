package db

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
	"strings"
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

func QueryCustomDashboardRoleRelByCustomDashboard(dashboardId int) (list []*models.CustomDashBoardRoleRel, err error) {
	err = x.SQL("select * from custom_dashboard_role_rel where custom_dashboard = ?", dashboardId).Find(&list)
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

func AddCustomDashboard(customDashboard *models.CustomDashboardTable) (err error) {
	_, err = x.Exec("insert into custom_dashboard (name,create_user,update_user,create_at,update_at) values(?,?,?,?,?)", customDashboard.Name,
		customDashboard.CreateUser, customDashboard.UpdateUser, customDashboard.CreateAt.Format(models.DatetimeFormat), customDashboard.UpdateAt.Format(models.DatetimeFormat))
	return
}

func QueryCustomDashboardRoleRefListByDashboard(dashboard int) (hashMap map[string]string, err error) {
	var list []*models.CustomDashBoardRoleRel
	hashMap = make(map[string]string)
	err = x.SQL("select * from custom_dashboard_role_rel where custom_dashboard = ?", dashboard).Find(&list)
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
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_role_rel where custom_dashboard = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard_chart_rel where custom_dashboard = ?", Param: []interface{}{dashboard}})
	actions = append(actions, &Action{Sql: "delete from custom_dashboard WHERE id=?", Param: []interface{}{dashboard}})
	return Transaction(actions)
}

func getQueryIdsByPermission(condition models.CustomDashboardQueryParam, roles []string) (strArr []string, err error) {
	var ids []int
	var sql = "select custom_dashboard from custom_dashboard_role_rel "
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
		sql = sql + " and role_id  in (" + useRoleFilterSql + ")"
		params = append(params, useRoleFilterParam...)
	}

	if len(condition.MgmtRoles) > 0 {
		mgmtRoleFilterSql, mgmtRoleFilterParam := createListParams(condition.MgmtRoles, "")
		sql = sql + " and role_id  in (" + mgmtRoleFilterSql + ")"
		params = append(params, mgmtRoleFilterParam...)
	}
	if condition.Permission == string(models.PermissionMgmt) {
		sql = sql + " and permission = ? "
		params = append(params, models.PermissionMgmt)
	}
	if err = x.SQL(sql).Find(&ids); err != nil {
		return
	}
	if len(ids) > 0 {
		strArr = TransformInToStrArray(ids)
	}
	return
}

func TransformInToStrArray(ids []int) []string {
	stringArray := make([]string, len(ids))
	for i, v := range ids {
		stringArray[i] = strconv.Itoa(v) // 将整数转换为字符串
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
		for _, str := range res {
			hashMap[str] = true
		}
	}
	return res
}
