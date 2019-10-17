package db

import (
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"fmt"
)

func ListCustomDashboard() (err error,result []*m.CustomDashboardTable) {
	err = x.SQL("SELECT * FROM custom_dashboard").Find(&result)
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
