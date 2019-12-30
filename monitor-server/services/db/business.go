package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func GetBusinessList(endpointId int) (err error, pathList []*m.BusinessMonitorTable) {
	err = x.SQL("SELECT * FROM business_monitor WHERE endpoint_id=?", endpointId).Find(&pathList)
	return err,pathList
}

func UpdateBusiness(param m.BusinessUpdateDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM business_monitor WHERE endpoint_id=?", Param:[]interface{}{param.EndpointId}})
	for _,v := range param.PathList {
		var action Action
		params := make([]interface{}, 0)
		action.Sql = "INSERT INTO business_monitor(endpoint_id,path) VALUE (?,?)"
		params = append(params, param.EndpointId)
		params = append(params, v)
		action.Param = params
		actions = append(actions, &action)
	}
	return Transaction(actions)
}
