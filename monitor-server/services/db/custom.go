package db

import (
	"encoding/json"
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
	"strings"
)

func ListCustomDashboard(user string, coreToken m.CoreJwtToken) (err error, result []*m.CustomDashboardQuery) {
	var sql string
	roleList := coreToken.Roles
	if user == "" {
		user = coreToken.User
	} else {
		if len(roleList) == 0 {
			_, userRoleList := GetUserRole(user)
			for _, v := range userRoleList {
				roleList = append(roleList, v.Name)
			}
		}
	}
	roleString := strings.Join(roleList, "','")
	sql = `SELECT * FROM (
		SELECT DISTINCT t1.id,t1.name,t1.panels_group,t1.cfg,t1.main,t1.create_user,t1.update_user,t1.create_at,t1.update_at,t2.permission,t1.panel_groups FROM custom_dashboard t1 LEFT JOIN rel_role_custom_dashboard t2 ON t1.id=t2.custom_dashboard_id LEFT JOIN role t3 ON t2.role_id=t3.id WHERE t1.create_user<>'` + user + `' and t3.name IN ('` + roleString + `')
		UNION
		SELECT id,name,panels_group,cfg,main,create_user,update_user,create_at,update_at,'mgmt',panel_groups FROM custom_dashboard WHERE create_user='` + user + `'
		) t ORDER BY t.name`
	err = x.SQL(sql).Find(&result)
	if err != nil {
		return err, result
	}
	result = distinctCustomDashboard(result)
	var roleTables []*m.RoleTable
	x.SQL("SELECT * FROM role WHERE name IN ('" + roleString + "')").Find(&roleTables)
	for _, v := range result {
		v.Main = 0
		v.MainPage = []string{}
		for _, vv := range roleTables {
			if v.Id == vv.MainDashboard {
				v.MainPage = append(v.MainPage, vv.Name)
				v.Main = 1
			}
		}
		if v.PanelGroups != "" {
			v.PanelGroupList = strings.Split(v.PanelGroups, ",")
		} else {
			v.PanelGroupList = []string{}
		}
	}
	return err, result
}

func distinctCustomDashboard(input []*m.CustomDashboardQuery) (output []*m.CustomDashboardQuery) {
	permissionMap := make(map[int]string)
	for _, v := range input {
		if vv, b := permissionMap[v.Id]; b {
			if vv == "mgmt" {
				continue
			}
			if v.Permission == "mgmt" {
				permissionMap[v.Id] = "mgmt"
			}
		} else {
			permissionMap[v.Id] = v.Permission
		}
	}
	for _, v := range input {
		if permissionMap[v.Id] != v.Permission {
			continue
		}
		output = append(output, v)
	}
	return
}

func GetCustomDashboard(id int) (result *m.CustomDashboardObj, err error) {
	if id == 0 {
		err = fmt.Errorf("custom dashboard id:%d illegal", id)
		return
	}
	var customRows []*m.CustomDashboardTable
	err = x.SQL("SELECT * FROM custom_dashboard WHERE id=?", id).Find(&customRows)
	if err != nil {
		err = fmt.Errorf("query custom dashboard table fail,%s ", err.Error())
		return
	}
	if len(customRows) == 0 {
		err = fmt.Errorf("can not find custom dashboard with id:%d", id)
		return
	}
	result = &m.CustomDashboardObj{CustomDashboardTable: *customRows[0]}
	if result.PanelGroups != "" {
		result.PanelGroupList = strings.Split(result.PanelGroups, ",")
	} else {
		result.PanelGroupList = []string{}
	}
	return
}

func SaveCustomDashboard(query *m.CustomDashboardObj) error {
	param := make([]interface{}, 0)
	if query.Id > 0 {
		param = append(param, fmt.Sprintf("UPDATE custom_dashboard SET name=?,cfg=?,update_user=?,panel_groups=? WHERE id=?"))
		param = append(param, query.Name)
		param = append(param, query.Cfg)
		param = append(param, query.UpdateUser)
		param = append(param, query.PanelGroups)
		param = append(param, query.Id)
	} else {
		param = append(param, fmt.Sprintf("INSERT INTO custom_dashboard(name,cfg,create_user,create_at,update_user,panel_groups) VALUE (?,?,?,NOW(),?,?)"))
		param = append(param, query.Name)
		param = append(param, query.Cfg)
		param = append(param, query.UpdateUser)
		param = append(param, query.UpdateUser)
		param = append(param, query.PanelGroups)
	}
	_, err := x.Exec(param...)
	return err
}

func DeleteCustomDashboard(query *m.CustomDashboardTable) error {
	_, err := x.Exec("DELETE FROM custom_dashboard WHERE id=?", query.Id)
	return err
}

func GetCustomDashboardRole(id int) (err error, result []*m.CustomDashboardRoleObj) {
	var roleTables []*m.CustomerDashboardRoleQuery
	err = x.SQL("SELECT DISTINCT t1.id,t1.name,t1.display_name,t2.permission FROM role t1 LEFT JOIN rel_role_custom_dashboard t2 ON t1.id=t2.role_id WHERE t2.custom_dashboard_id=?", id).Find(&roleTables)
	for _, v := range roleTables {
		result = append(result, &m.CustomDashboardRoleObj{RoleId: v.Id, Permission: v.Permission})
	}
	return err, result
}

func SaveCustomeDashboardRole(param m.CustomDashboardRoleDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM rel_role_custom_dashboard WHERE custom_dashboard_id=?", Param: []interface{}{param.DashboardId}})
	for _, v := range param.PermissionList {
		if v.Permission != "use" && v.Permission != "mgmt" {
			continue
		}
		actions = append(actions, &Action{Sql: "INSERT INTO rel_role_custom_dashboard(role_id,custom_dashboard_id,permission) VALUE (?,?,?)", Param: []interface{}{v.RoleId, param.DashboardId, v.Permission}})
	}
	return Transaction(actions)
}

func GetCustomDashboardAlarms(id int) (err error, result m.AlarmProblemQueryResult) {
	result = m.AlarmProblemQueryResult{High: 0, Mid: 0, Low: 0, Data: []*m.AlarmProblemQuery{}}
	customQuery := &m.CustomDashboardObj{}
	customQuery, err = GetCustomDashboard(id)
	if err != nil || customQuery.Cfg == "" {
		return err, result
	}
	var customConfig []*m.CustomDashboardConfigObj
	err = json.Unmarshal([]byte(customQuery.Cfg), &customConfig)
	if err != nil {
		return fmt.Errorf("json unmarshal dashboard config fail,%s", err.Error()), result
	}
	var endpointList []string
	for _, v := range customConfig {
		for _, vv := range v.Query {
			if vv.AppObject != "" {
				endpointList = append(endpointList, "sg__"+vv.AppObject)
				serviceGuidList, _ := fetchGlobalServiceGroupChildGuidList(vv.AppObject)
				serviceGroupEndpoint := getServiceGroupEndpointWithType(vv.EndpointType, serviceGuidList)
				for _, sgEndpoint := range serviceGroupEndpoint {
					endpointList = append(endpointList, sgEndpoint.Guid)
				}
			} else {
				endpointList = append(endpointList, vv.Endpoint)
			}
		}
	}
	if len(endpointList) > 0 {
		sql := "SELECT * FROM alarm WHERE status='firing' AND endpoint IN ('" + strings.Join(endpointList, "','") + "') ORDER BY id DESC"
		err, result = QueryAlarmBySql(sql, []interface{}{}, m.CustomAlarmQueryParam{Enable: false}, &m.PageInfo{})
	}
	return err, result
}

func ListMainPageRole(user string, roleList []string) (err error, result []*m.MainPageRoleQuery) {
	var customDashboards []*m.CustomDashboardQuery
	roleString := strings.Join(roleList, "','")
	sql := `SELECT * FROM (
		SELECT DISTINCT t1.id,t1.name,t1.panels_group,t1.cfg,t1.main,t1.create_user,t1.update_user,t1.create_at,t1.update_at,t2.permission FROM custom_dashboard t1 LEFT JOIN rel_role_custom_dashboard t2 ON t1.id=t2.custom_dashboard_id LEFT JOIN role t3 ON t2.role_id=t3.id WHERE t1.create_user<>'` + user + `' and t3.name IN ('` + roleString + `')
		UNION
		SELECT id,name,panels_group,cfg,main,create_user,update_user,create_at,update_at,'mgmt' FROM custom_dashboard WHERE create_user='` + user + `'
		) t ORDER BY t.name`
	err = x.SQL(sql).Find(&customDashboards)
	if err != nil {
		return err, result
	}
	customDashboards = distinctCustomDashboard(customDashboards)
	var options []*m.OptionModel
	options = append(options, &m.OptionModel{Id: 0, OptionValue: "0", OptionText: "null"})
	for _, v := range customDashboards {
		options = append(options, &m.OptionModel{Id: v.Id, OptionValue: strconv.Itoa(v.Id), OptionText: v.Name})
	}
	var roleTables []*m.RoleTable
	x.SQL("SELECT * FROM role WHERE name IN ('" + roleString + "')").Find(&roleTables)
	for _, v := range roleTables {
		var tmpMainName string
		for _, vv := range options {
			if v.MainDashboard == vv.Id {
				tmpMainName = vv.OptionText
				break
			}
		}
		result = append(result, &m.MainPageRoleQuery{RoleName: v.Name, MainPageId: v.MainDashboard, MainPageName: tmpMainName, Options: options})
	}
	return err, result
}

func UpdateMainPageRole(param []m.MainPageRoleQuery) error {
	var actions []*Action
	for _, v := range param {
		var tmpAction Action
		var tmpParam []interface{}
		tmpAction.Sql = "UPDATE role SET main_dashboard=? WHERE name=?"
		tmpParam = append(tmpParam, v.MainPageId)
		tmpParam = append(tmpParam, v.RoleName)
		tmpAction.Param = tmpParam
		actions = append(actions, &tmpAction)
	}
	return Transaction(actions)
}
