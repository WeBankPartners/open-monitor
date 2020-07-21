package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
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

func GetAliveCheckQueue() (err error,result []*m.AliveCheckQueueTable) {
	err = x.SQL("SELECT * FROM alive_check_queue LIMIT 1").Find(&result)
	if err != nil {
		return err,result
	}
	if len(result) > 0 {
		_,err = x.Exec(fmt.Sprintf("DELETE FROM alive_check_queue WHERE message='%s'", result[0].Message))
	}else{
		err = fmt.Errorf("alive check queue table is empty")
	}
	return err,result
}