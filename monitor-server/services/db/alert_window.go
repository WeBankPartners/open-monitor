package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strings"
	"time"
)

func GetAlertWindowList(endpoint string) (result []*m.AlertWindowObj, err error) {
	var tableData []*m.AlertWindowTable
	err = x.SQL("select id,endpoint,`start`,`end`,`weekday` from alert_window where endpoint=?", endpoint).Find(&tableData)
	if err != nil {
		err = fmt.Errorf("Query alert window table fail,%s ", err.Error())
		log.Error(nil, log.LOGGER_APP, err.Error())
		return result, err
	}
	for _, v := range tableData {
		result = append(result, &m.AlertWindowObj{Start: v.Start, End: v.End, Weekday: v.Weekday, TimeList: []string{v.Start, v.End}})
	}
	return result, err
}

func UpdateAlertWindowList(endpoint, updateUser string, data []*m.AlertWindowObj) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from alert_window where endpoint=?", Param: []interface{}{endpoint}})
	if len(data) > 0 {
		var tmpParams []interface{}
		sql := "insert into alert_window(endpoint,start,end,weekday,update_user) values "
		for i, v := range data {
			if len(v.TimeList) == 2 {
				v.Start = v.TimeList[0]
				v.End = v.TimeList[1]
			}
			sql += "(?,?,?,?,?)"
			tmpParams = append(tmpParams, endpoint, v.Start, v.End, v.Weekday, updateUser)
			if i < len(data)-1 {
				sql += ","
			}
		}
		actions = append(actions, &Action{Sql: sql, Param: tmpParams})
	}
	err := Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update alert window table fail,%s ", err.Error())
		log.Error(nil, log.LOGGER_APP, err.Error())
	}
	return err
}

func CheckEndpointActiveAlert(endpoint string) bool {
	var tableData []*m.AlertWindowTable
	x.SQL("select id,endpoint,`start`,`end`,`weekday` from alert_window where endpoint=?", endpoint).Find(&tableData)
	if len(tableData) == 0 {
		return true
	}
	activeFlag := true
	nTime := time.Now()
	for _, v := range tableData {
		if strings.Contains(v.Weekday, "All") || strings.Contains(v.Weekday, time.Now().Weekday().String()) {
			startTime, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s %s:00 "+m.DefaultLocalTimeZone, nTime.Format("2006-01-02"), v.Start))
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "Check endpoint is in active alert window error", zap.String("start", v.Start), zap.Error(err))
				continue
			}
			endTime, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s %s:00 "+m.DefaultLocalTimeZone, nTime.Format("2006-01-02"), v.End))
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "Check endpoint is in active alert window error", zap.String("end", v.End), zap.Error(err))
				continue
			}
			if (nTime.Unix() >= startTime.Unix()) && (nTime.Unix() <= endTime.Unix()) {
				activeFlag = false
				break
			}
		}
	}
	return activeFlag
}
