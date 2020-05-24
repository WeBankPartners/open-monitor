package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func GetBusinessList(endpointId int,ownerEndpoint string) (err error, pathList []*m.BusinessMonitorTable) {
	if endpointId > 0 {
		err = x.SQL("SELECT * FROM business_monitor WHERE endpoint_id=?", endpointId).Find(&pathList)
	}else if ownerEndpoint != "" {
		err = x.SQL("SELECT * FROM business_monitor WHERE owner_endpoint=?", ownerEndpoint).Find(&pathList)
	}
	return err,pathList
}

func UpdateBusiness(param m.BusinessUpdateDto) error {
	var actions []*Action
	var tmpOwnerList []string
	for _,v := range param.PathList {
		exist := false
		for _,vv := range tmpOwnerList {
			if v.OwnerEndpoint == vv {
				exist = true
				break
			}
		}
		if !exist {
			tmpOwnerList = append(tmpOwnerList, v.OwnerEndpoint)
		}
	}
	for _,v := range tmpOwnerList {
		actions = append(actions, &Action{Sql:"DELETE FROM business_monitor WHERE endpoint_id=? and owner_endpoint=?", Param:[]interface{}{param.EndpointId, v}})
	}
	for _,v := range param.PathList {
		var action Action
		params := make([]interface{}, 0)
		action.Sql = "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)"
		params = append(params, param.EndpointId)
		params = append(params, v.Path)
		params = append(params, v.OwnerEndpoint)
		action.Param = params
		actions = append(actions, &action)
	}
	return Transaction(actions)
}

func CheckEndpointBusiness(endpoint string) bool {
	result := true
	var businessMonitorTables []*m.BusinessMonitorTable
	x.SQL("SELECT t2.* FROM endpoint t1 JOIN business_monitor t2 ON t1.id=t2.`endpoint_id` WHERE t1.`guid`=?", endpoint).Find(&businessMonitorTables)
	for _,v :=range businessMonitorTables {
		if v.OwnerEndpoint != "" {
			result = false
			break
		}
	}
	return result
}

func GetBusinessPanelChart() (charts []*m.ChartTable,panels []*m.PanelTable) {
	x.SQL("SELECT t2.* FROM dashboard t1 LEFT JOIN panel t2 ON t1.panels_group=t2.group_id WHERE t1.dashboard_type='host' AND t2.auto_display=1").Find(&panels)
	x.SQL("SELECT t3.* FROM dashboard t1 LEFT JOIN panel t2 ON t1.panels_group=t2.group_id LEFT JOIN chart t3 ON t2.chart_group=t3.group_id WHERE t1.dashboard_type='host' AND t2.auto_display=1").Find(&charts)
	return charts,panels
}