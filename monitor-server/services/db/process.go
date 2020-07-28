package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
	"time"
)

func GetProcessList(endpointId int) (err error, processList []*m.ProcessMonitorTable) {
	err = x.SQL("SELECT * FROM process_monitor WHERE endpoint_id=?", endpointId).Find(&processList)
	return err,processList
}

func UpdateProcess(param m.ProcessUpdateDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM process_monitor WHERE endpoint_id=?", Param:[]interface{}{param.EndpointId}})
	for _,v := range param.ProcessList {
		var action Action
		params := make([]interface{}, 0)
		action.Sql = "INSERT INTO process_monitor(endpoint_id,NAME) VALUE (?,?)"
		params = append(params, param.EndpointId)
		params = append(params, v)
		action.Param = params
		actions = append(actions, &action)
	}
	return Transaction(actions)
}

func UpdateAliveCheckQueue(monitorIp string) error {
	_,err := x.Exec(fmt.Sprintf("INSERT INTO alive_check_queue(message) VALUE ('%s')", monitorIp))
	return err
}

func GetAliveCheckQueue(param string) (err error,result []*m.AliveCheckQueueTable) {
	lastMinDateString := time.Unix(time.Now().Unix()-60, 0).Format("2006-01-02 15:04:05")
	err = x.SQL(fmt.Sprintf("SELECT * FROM alive_check_queue WHERE message='%s' AND update_at>'%s' LIMIT 1", param, lastMinDateString)).Find(&result)
	if err != nil {
		return err,result
	}
	if len(result) == 0 {
		err = fmt.Errorf("get alive_check_queue table fail,nodata with message=%s and update_at>%s", param, lastMinDateString)
	}
	return err,result
}