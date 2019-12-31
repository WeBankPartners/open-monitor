package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"strings"
	"time"
)

func AddUser(user m.UserTable, creator string) error {
	_,err := x.Exec("INSERT INTO user(name,passwd,display_name,email,phone,creator,created) VALUE (?,?,?,?,?,?,NOW())", user.Name,user.Passwd,user.DisplayName,user.Email,user.Phone,creator)
	if err != nil {
		mid.LogError(fmt.Sprintf("add user %s fail", user.Name), err)
	}
	return err
}

func GetUser(username string) (err error,user m.UserTable) {
	var users []*m.UserTable
	err = x.SQL("SELECT * FROM user WHERE name=?", username).Find(&users)
	if len(users) == 0 {
		return err,m.UserTable{}
	}
	return nil,*users[0]
}

func UpdateUser(user m.UserTable) error {
	sql := "UPDATE user SET "
	if user.Passwd != "" {
		sql += fmt.Sprintf("passwd='%s',", user.Passwd)
	}
	if user.DisplayName != "" {
		sql += fmt.Sprintf("display_name='%s',", user.DisplayName)
	}
	if user.Email != "" {
		sql += fmt.Sprintf("email='%s',", user.Email)
	}
	if user.Phone != "" {
		sql += fmt.Sprintf("phone='%s'", user.Phone)
	}
	updateSql := sql[:len(sql)-1] + fmt.Sprintf(" WHERE name='%s'", user.Name)
	_,err := x.Exec(updateSql)
	return err
}

func SearchUserRole(search string,searchType string) (err error,options []*m.OptionModel) {
	likeString := "%" + search + "%"
	var result []*m.RoleTable
	err = x.SQL(fmt.Sprintf("SELECT id,name,display_name FROM %s WHERE name LIKE '%s' OR display_name LIKE '%s' ORDER BY id LIMIT 15", searchType, likeString, likeString)).Find(&result)
	if err != nil {
		return err,options
	}
	tmpActive := false
	if searchType == "role" {
		tmpActive = true
	}
	for _,v := range result {
		tmpText := v.Name
		if v.DisplayName != "" {
			tmpText = tmpText + "(" + v.DisplayName + ")"
		}
		options = append(options, &m.OptionModel{Id:v.Id, OptionText:tmpText, OptionValue:fmt.Sprintf("%s_%d", searchType, v.Id), Active:tmpActive})
	}
	return nil,options
}

func GetMailByStrategy(strategyId int) []string {
	result := []string{}
	var tpls []*m.TplTable
	x.SQL("SELECT DISTINCT t2.action_user,t2.action_role FROM strategy t1 LEFT JOIN tpl t2 ON t1.tpl_id=t2.id WHERE t1.id=?", strategyId).Find(&tpls)
	if len(tpls) == 0 {
		mid.LogInfo(fmt.Sprintf("can not find tpl with strategy %d",strategyId))
		return result
	}
	userIds := tpls[0].ActionUser
	if tpls[0].ActionRole != "" {
		var tmpRel []*m.RelRoleUserTable
		x.SQL(fmt.Sprintf("SELECT user_id FROM rel_role_user WHERE role_id IN (%s)", tpls[0].ActionRole)).Find(&tmpRel)
		for _,v := range tmpRel {
			userIds = userIds + fmt.Sprintf(",%d", v.UserId)
		}
		if strings.HasPrefix(userIds, ",") {
			userIds = userIds[1:]
		}
	}
	if userIds != "" {
		var users []*m.UserTable
		x.SQL(fmt.Sprintf("SELECT DISTINCT email FROM user WHERE id IN (%s)", userIds)).Find(&users)
		for _,v := range users {
			result = append(result, v.Email)
		}
	}
	return result
}

func ListUser(search string,role,page,size int) (err error,users []*m.UserTable) {
	var whereSql string
	if role > 0 {
		whereSql = fmt.Sprintf(" WHERE id IN (SELECT user_id FROM rel_role_user WHERE role_id=%d) ", role)
	}
	if search != "" {
		whereSql = " WHERE name LIKE '%"+search+"%' OR display_name LIKE '%"+search+"%'"
	}
	err = x.SQL("SELECT id,name,display_name,email,phone FROM user "+whereSql+fmt.Sprintf(" ORDER BY id LIMIT %d,%d", (page-1)*size, size)).Find(&users)
	return err,users
}

func ListRole(search string,page,size int) (err error,roles []*m.RoleTable) {
	var whereSql string
	if search != "" {
		whereSql = "name LIKE '%"+search+"%' OR display_name LIKE '%"+search+"%'"
	}
	err = x.SQL("SELECT * FROM role "+whereSql+fmt.Sprintf(" ORDER BY id LIMIT %d,%d", (page-1)*size, size)).Find(&roles)
	return err,roles
}

func UpdateRoleUser(param m.UpdateRoleUserDto) error {
	var roleUserTable []*m.RelRoleUserTable
	err := x.SQL("SELECT user_id FROM rel_role_user WHERE role_id=?", param.RoleId).Find(&roleUserTable)
	if err != nil {
		return err
	}
	isSame := true
	if len(roleUserTable) != len(param.UserId) {
		isSame = false
	}else{
		for _,v := range roleUserTable {
			tmp := false
			for _,vv := range param.UserId {
				if v.UserId == vv {
					tmp = true
					break
				}
			}
			if !tmp {
				isSame = false
				break
			}
		}
	}
	if isSame {
		return nil
	}
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM rel_role_user WHERE role_id=?", Param:[]interface{}{param.RoleId}})
	for _,v := range param.UserId {
		actions = append(actions, &Action{Sql:"INSERT INTO rel_role_user(role_id,user_id) VALUE (?,?)", Param:[]interface{}{param.RoleId, v}})
	}
	err = Transaction(actions)
	return err
}

func UpdateRole(param m.UpdateRoleDto) error {
	var role m.RoleTable
	force := false
	if param.Operation == "add" {
		if param.Name == "" {
			return fmt.Errorf("role name is null")
		}
		role.Name = param.Name
		role.DisplayName = param.DisplayName
		role.Creator = param.Operator
		role.Created = time.Now()
		param.Operation = "insert"
	}
	if param.Operation == "update" {
		if param.RoleId <= 0 {
			return fmt.Errorf("role id is null")
		}
		if param.Name == "" {
			return fmt.Errorf("role name is null")
		}
		role.Id = param.RoleId
		role.Name = param.Name
		role.DisplayName = param.DisplayName
		force = true
	}
	if param.Operation == "delete" {
		if param.RoleId <= 0 {
			return fmt.Errorf("role id is null")
		}
		role.Id = param.RoleId
	}
	action := Classify(role, param.Operation, "role", force)
	return Transaction([]*Action{&action})
}