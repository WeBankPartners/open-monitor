package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
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
		SELECT DISTINCT t1.id,t1.name,t1.panels_group,t1.cfg,t1.main,t1.create_user,t1.update_user,t1.create_at,t1.update_at,t2.permission,t1.panel_groups FROM custom_dashboard t1 LEFT JOIN custom_dashboard_role_rel t2 ON t1.id=t2.custom_dashboard_id LEFT JOIN role t3 ON t2.role_id=t3.id WHERE t1.create_user<>'` + user + `' and t3.name IN ('` + roleString + `')
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
	err = x.SQL("SELECT DISTINCT t1.id,t1.name,t1.display_name,t2.permission FROM role t1 LEFT JOIN custom_dashboard_role_rel t2 ON t1.id=t2.role_id WHERE t2.custom_dashboard_id=?", id).Find(&roleTables)
	for _, v := range roleTables {
		result = append(result, &m.CustomDashboardRoleObj{RoleId: v.Id, Permission: v.Permission})
	}
	return err, result
}

func SaveCustomeDashboardRole(param m.CustomDashboardRoleDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM custom_dashboard_role_rel WHERE custom_dashboard_id=?", Param: []interface{}{param.DashboardId}})
	for _, v := range param.PermissionList {
		if v.Permission != "use" && v.Permission != "mgmt" {
			continue
		}
		actions = append(actions, &Action{Sql: "INSERT INTO custom_dashboard_role_rel(role_id,custom_dashboard_id,permission) VALUE (?,?,?)", Param: []interface{}{v.RoleId, param.DashboardId, v.Permission}})
	}
	return Transaction(actions)
}

func GetCustomDashboardAlarms(id int, page *m.PageInfo) (err error, result m.AlarmProblemQueryResult) {
	result = m.AlarmProblemQueryResult{High: 0, Mid: 0, Low: 0, Data: []*m.AlarmProblemQuery{}}
	//customQuery := &m.CustomDashboardObj{}
	var customChartSeriesRows []*m.CustomChartSeries
	err = x.SQL("select dashboard_chart,endpoint,service_group,monitor_type,metric from custom_chart_series where dashboard_chart in (select dashboard_chart from custom_dashboard_chart_rel where custom_dashboard=?)", id).Find(&customChartSeriesRows)
	if err != nil {
		err = fmt.Errorf("query chart series fail,%s ", err.Error())
		return
	}
	var endpointList []string
	for _, row := range customChartSeriesRows {
		if row.ServiceGroup != "" {
			endpointList = append(endpointList, "sg__"+row.ServiceGroup)
			serviceGuidList, _ := fetchGlobalServiceGroupChildGuidList(row.ServiceGroup)
			serviceGroupEndpoint := getServiceGroupEndpointWithType(row.MonitorType, serviceGuidList)
			for _, sgEndpoint := range serviceGroupEndpoint {
				endpointList = append(endpointList, sgEndpoint.Guid)
			}
		} else {
			endpointList = append(endpointList, row.Endpoint)
		}
	}
	//customQuery, err = GetCustomDashboard(id)
	//if err != nil || customQuery.Cfg == "" {
	//	return err, result
	//}
	//var customConfig []*m.CustomDashboardConfigObj
	//err = json.Unmarshal([]byte(customQuery.Cfg), &customConfig)
	//if err != nil {
	//	return fmt.Errorf("json unmarshal dashboard config fail,%s", err.Error()), result
	//}
	//var endpointList []string
	//for _, v := range customConfig {
	//	for _, vv := range v.Query {
	//		if vv.AppObject != "" {
	//			endpointList = append(endpointList, "sg__"+vv.AppObject)
	//			serviceGuidList, _ := fetchGlobalServiceGroupChildGuidList(vv.AppObject)
	//			serviceGroupEndpoint := getServiceGroupEndpointWithType(vv.EndpointType, serviceGuidList)
	//			for _, sgEndpoint := range serviceGroupEndpoint {
	//				endpointList = append(endpointList, sgEndpoint.Guid)
	//			}
	//		} else {
	//			endpointList = append(endpointList, vv.Endpoint)
	//		}
	//	}
	//}
	if len(endpointList) > 0 {
		sql := "SELECT * FROM alarm WHERE status='firing' AND endpoint IN ('" + strings.Join(endpointList, "','") + "') ORDER BY id DESC"
		err, result = QueryAlarmBySql(sql, []interface{}{}, m.CustomAlarmQueryParam{Enable: false}, page)
	}
	return err, result
}

func ListMainPageRole(roleList []string) (err error, result []*m.MainPageRoleQuery) {
	var displayNameRoleMap map[string]string
	var mainDashboardList []*m.MainDashboard
	var mainPageId int
	var customDashboardNameMap = make(map[int]string)
	var dashboardRelMap = make(map[string][]int)
	var sql = "select custom_dashboard_id from custom_dashboard_role_rel "
	var ids []int
	result = []*m.MainPageRoleQuery{}
	if len(roleList) == 0 {
		return
	}
	if customDashboardNameMap, err = QueryAllCustomDashboardNameMap(); err != nil {
		return
	}
	if displayNameRoleMap, err = QueryAllRoleDisplayNameMap(); err != nil {
		return
	}
	if err = x.SQL("select * from main_dashboard").Find(&mainDashboardList); err != nil {
		return
	}
	if dashboardRelMap, err = QueryCustomDashboardRoleRelMap(); err != nil {
		return
	}
	roleFilterSql, roleFilterParam := createListParams(roleList, "")
	sql = sql + " where role_id  in (" + roleFilterSql + ")"
	if err = x.SQL(sql, roleFilterParam...).Find(&ids); err != nil {
		return
	}
	for _, role := range roleList {
		if _, ok := displayNameRoleMap[role]; !ok {
			continue
		}
		var dashboardIds []int
		mainPageId = 0
		for _, dashboard := range mainDashboardList {
			if dashboard.RoleId == role {
				mainPageId = *dashboard.CustomDashboard
				break
			}
		}
		mainPageRole := &m.MainPageRoleQuery{
			RoleName:        role,
			DisplayRoleName: displayNameRoleMap[role],
			MainPageId:      mainPageId,
			MainPageName:    customDashboardNameMap[mainPageId],
			Options:         []*m.OptionModel{},
		}
		if dashboardIds = dashboardRelMap[role]; len(dashboardIds) > 0 {
			idMap := make(map[int]bool)
			for _, id := range ids {
				for _, dashboardId := range dashboardIds {
					if id == dashboardId {
						if _, ok := idMap[id]; !ok && customDashboardNameMap[dashboardId] != "" {
							idMap[id] = true
							mainPageRole.Options = append(mainPageRole.Options, &m.OptionModel{
								Id:          dashboardId,
								OptionValue: strconv.Itoa(dashboardId),
								OptionText:  customDashboardNameMap[dashboardId],
							})
						}
						break
					}
				}
			}
		}
		result = append(result, mainPageRole)
	}
	return err, result
}

func UpdateMainPageRole(param []m.MainPageRoleQuery) error {
	var actions []*Action
	if len(param) > 0 {
		actions = append(actions, &Action{Sql: "delete from main_dashboard"})
	}
	var idList = guid.CreateGuidList(len(param))
	for i, v := range param {
		if v.MainPageId > 0 {
			actions = append(actions, &Action{Sql: "insert into main_dashboard(guid,role_id,custom_dashboard) values(?,?,?)", Param: []interface{}{idList[i], v.RoleName, v.MainPageId}})
		}
	}
	return Transaction(actions)
}
