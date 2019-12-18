package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
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