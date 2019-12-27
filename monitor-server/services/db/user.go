package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"strings"
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