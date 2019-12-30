package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
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