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
	param := make([]interface{}, 0)
	sql := "UPDATE user SET "
	if user.Passwd != "" {
		sql += "passwd=?,"
		param = append(param, user.Passwd)
	}
	if user.DisplayName != "" {
		sql += "display_name=?,"
		param = append(param, user.DisplayName)
	}
	if user.Email != "" {
		sql += "email=?,"
		param = append(param, user.Email)
	}
	if user.Phone != "" {
		sql += "phone=?,"
		param = append(param, user.Phone)
	}
	updateSql := sql[:len(sql)-1] + " WHERE name=?"
	param = append(param, user.Name)
	newParam := make([]interface{}, 0)
	newParam = append(newParam, updateSql)
	for _,v := range param {
		newParam = append(newParam, v)
	}
	_,err := x.Exec(newParam...)
	if err != nil {
		mid.LogError("update user error ",err)
	}
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

func ListUser(search string,role,page,size int) (err error,data m.TableData) {
	var users []*m.UserQuery
	var count []int
	var whereSql string
	if role > 0 {
		whereSql = fmt.Sprintf(" AND t1.id IN (SELECT user_id FROM rel_role_user WHERE role_id=%d) ", role)
	}
	if search != "" {
		whereSql = " AND t1.name LIKE '%"+search+"%' OR display_name LIKE '%"+search+"%'"
	}
	sql := `SELECT t5.* FROM (
	SELECT t4.id,t4.name,t4.display_name,t4.email,t4.phone,t4.created,GROUP_CONCAT(role) role FROM (
	SELECT t1.id,t1.name,t1.display_name,t1.email,t1.phone,t1.created,CONCAT(t3.name,':',t3.display_name) role FROM user t1
	LEFT JOIN rel_role_user t2 ON t1.id=t2.user_id
	LEFT JOIN role t3 ON t2.role_id=t3.id
	WHERE 1=1 ` + whereSql + `
	) t4 GROUP BY t4.id
	) t5`
	err = x.SQL(sql+fmt.Sprintf(" ORDER BY t5.id LIMIT %d,%d", (page-1)*size, size)).Find(&users)
	x.SQL(sql).Find(&count)
	if len(users) > 0 {
		for _,v := range users {
			v.CreatedString = v.Created.Format(m.DatetimeFormat)
		}
		data.Data = users
	}else{
		data.Data = []*m.UserQuery{}
	}
	data.Size = size
	data.Page = page
	if len(count) > 0 {
		data.Num = count[0]
	}else{
		data.Num = len(users)
	}
	return err,data
}

func ListRole(search string,page,size int) (err error,data m.TableData) {
	var roles []*m.RoleTable
	var count []int
	var whereSql string
	if search != "" {
		whereSql = "where name LIKE '%"+search+"%' OR display_name LIKE '%"+search+"%'"
	}
	err = x.SQL("SELECT * FROM role "+whereSql+fmt.Sprintf(" ORDER BY id LIMIT %d,%d", (page-1)*size, size)).Find(&roles)
	x.SQL("SELECT count(1) num FROM role " + whereSql).Find(&count)
	if len(roles) > 0 {
		data.Data = roles
	}else{
		data.Data = []*m.RoleTable{}
	}
	data.Size = size
	data.Page = page
	if len(count) > 0 {
		data.Num = count[0]
	}else{
		data.Num = len(roles)
	}
	return err,data
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
		role.Email = param.Email
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
		role.Email = param.Email
		force = true
	}
	if param.Operation == "delete" {
		if param.RoleId <= 0 {
			return fmt.Errorf("role id is null")
		}
		role.Id = param.RoleId
	}
	action := Classify(role, param.Operation, "role", force)
	mid.LogInfo(fmt.Sprintf("action sql : %s  param : %v ", action.Sql, action.Param))
	return Transaction([]*Action{&action})
}