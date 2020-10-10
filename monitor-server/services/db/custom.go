package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
	"strings"
)

func ListCustomDashboard(user string,coreToken m.CoreJwtToken) (err error,result []*m.CustomDashboardTable) {
	var sql string
	roleList := coreToken.Roles
	if user == "" {
		user = coreToken.User
	}else {
		if len(roleList) == 0 {
			_, userRoleList := GetUserRole(user)
			for _, v := range userRoleList {
				roleList = append(roleList, v.Name)
			}
		}
	}
	roleString := strings.Join(roleList, "','")
	sql = `SELECT * FROM (
		SELECT DISTINCT t1.* FROM custom_dashboard t1 LEFT JOIN rel_role_custom_dashboard t2 ON t1.id=t2.custom_dashboard_id LEFT JOIN role t3 ON t2.role_id=t3.id WHERE t3.name IN ('` + roleString + `')
		UNION
		SELECT * FROM custom_dashboard WHERE create_user='` + user + `'
		) t ORDER BY t.id`
	err = x.SQL(sql).Find(&result)
	return err,result
}

func GetCustomDashboard(query *m.CustomDashboardTable) error {
	var err error
	var result []*m.CustomDashboardTable
	if query.Id > 0 {
		err = x.SQL("SELECT * FROM custom_dashboard WHERE id=?", query.Id).Find(&result)
		if len(result) > 0 {
			query.Id = result[0].Id
			query.Name = result[0].Name
			query.Cfg = result[0].Cfg
			query.CreateUser = result[0].CreateUser
			query.UpdateUser = result[0].UpdateUser
			query.CreateAt = result[0].CreateAt
			query.UpdateAt = result[0].UpdateAt
		}
	}
	return err
}

func SaveCustomDashboard(query *m.CustomDashboardTable) error {
	param := make([]interface{}, 0)
	if query.Id > 0 {
		param = append(param, fmt.Sprintf("UPDATE custom_dashboard SET name=?,cfg=?,update_user=? WHERE id=?"))
		param = append(param, query.Name)
		param = append(param, query.Cfg)
		param = append(param, query.UpdateUser)
		param = append(param, query.Id)
	}else{
		param = append(param, fmt.Sprintf("INSERT INTO custom_dashboard(name,cfg,create_user,create_at,update_user) VALUE (?,?,?,NOW(),?)"))
		param = append(param, query.Name)
		param = append(param, query.Cfg)
		param = append(param, query.UpdateUser)
		param = append(param, query.UpdateUser)
	}
	_,err := x.Exec(param...)
	return err
}

func DeleteCustomDashboard(query *m.CustomDashboardTable) error {
	_,err := x.Exec("DELETE FROM custom_dashboard WHERE id=?", query.Id)
	return err
}

func GetCustomDashboardRole(id int) (err error,result []*m.OptionModel) {
	var roleTables []*m.RoleTable
	err = x.SQL("SELECT DISTINCT t1.* FROM role t1 LEFT JOIN rel_role_custom_dashboard t2 ON t1.id=t2.role_id WHERE t2.custom_dashboard_id=?", id).Find(&roleTables)
	for _,v := range roleTables {
		tmpName := v.Name
		if v.DisplayName != "" {
			tmpName = v.DisplayName
		}
		result = append(result, &m.OptionModel{OptionText:tmpName, OptionValue:fmt.Sprintf("%d", v.Id), Id:v.Id})
	}
	return err,result
}

func SaveCustomeDashboardRole(param m.CustomDashboardRoleDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM rel_role_custom_dashboard WHERE custom_dashboard_id=?", Param:[]interface{}{param.DashboardId}})
	for _,v := range param.RoleId {
		actions = append(actions, &Action{Sql:"INSERT INTO rel_role_custom_dashboard(role_id,custom_dashboard_id) VALUE (?,?)", Param:[]interface{}{v, param.DashboardId}})
	}
	return Transaction(actions)
}