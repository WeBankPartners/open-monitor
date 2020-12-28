package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
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

func GetBusinessListNew(endpointId int,ownerEndpoint string) (err error,result m.BusinessUpdateDto) {
	result.EndpointId = endpointId
	result.PathList = []*m.BusinessUpdatePathObj{}
	var businessMonitorTable []*m.BusinessMonitorTable
	if endpointId > 0 {
		err = x.SQL("SELECT * FROM business_monitor WHERE endpoint_id=?", endpointId).Find(&businessMonitorTable)
	}else if ownerEndpoint != "" {
		err = x.SQL("SELECT * FROM business_monitor WHERE owner_endpoint=?", ownerEndpoint).Find(&businessMonitorTable)
	}
	if err != nil {
		return err,result
	}
	var tmpErr error
	for _,v := range businessMonitorTable {
		var businessMonitorConfigTable []*m.BusinessMonitorCfgTable
		tmpBup := m.BusinessUpdatePathObj{Id: v.Id,Path: v.Path,OwnerEndpoint: v.OwnerEndpoint,Rules: []*m.BusinessMonitorCfgObj{}}
		x.SQL("select * from business_monitor_cfg where business_monitor_id=?", v.Id).Find(&businessMonitorConfigTable)
		for _,vv := range businessMonitorConfigTable {
			tmpBmc := m.BusinessMonitorCfgObj{Id: vv.Id,Regular: vv.Regular,Tags: vv.Tags}
			tmpStringMap := []*m.BusinessStringMapObj{}
			tmpErr = json.Unmarshal([]byte(vv.StringMap), &tmpStringMap)
			if tmpErr != nil {
				log.Logger.Error("json unmarshal string map obj fail", log.String("string_map", vv.StringMap), log.Error(tmpErr))
			}else{
				tmpBmc.StringMap = tmpStringMap
			}
			tmpMetricConfig := []*m.BusinessMetricObj{}
			tmpErr = json.Unmarshal([]byte(vv.MetricConfig), &tmpMetricConfig)
			if tmpErr != nil {
				log.Logger.Error("json unmarshal metric config obj fail", log.String("metric_config", vv.MetricConfig), log.Error(tmpErr))
			}else{
				tmpBmc.MetricConfig = tmpMetricConfig
			}
			tmpBup.Rules = append(tmpBup.Rules, &tmpBmc)
		}
		result.PathList = append(result.PathList, &tmpBup)
	}
	return nil,result
}

func GetBusinessRealEndpoint(endpoint string) string {
	var endpointTable []*m.EndpointTable
	x.SQL("select t2.guid from business_monitor t1 left join endpoint t2 on t1.endpoint_id=t2.id where t1.owner_endpoint=?", endpoint).Find(&endpointTable)
	if len(endpointTable) > 0 {
		return endpointTable[0].Guid
	}else{
		return endpoint
	}
}

func UpdateBusiness(param m.BusinessUpdateDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM business_monitor WHERE endpoint_id=?", Param:[]interface{}{param.EndpointId}})
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

func AddBusinessTable(param m.BusinessUpdateDto) error {
	var actions []*Action
	var err error
	var businessMonitorTable []*m.BusinessMonitorTable
	x.SQL("select * from business_monitor where endpoint_id=?", param.EndpointId).Find(&businessMonitorTable)
	for _,v := range param.PathList {
		for _,vv := range businessMonitorTable {
			if v.Path == vv.Path {
				err = fmt.Errorf("Path %s already exists ", v.Path)
				break
			}
		}
		actions = append(actions, &Action{Sql: "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)", Param: []interface{}{param.EndpointId,v.Path,v.OwnerEndpoint}})
	}
	if err != nil {
		return err
	}
	return Transaction(actions)
}

func UpdateBusinessNew(param m.BusinessUpdateDto) error {
	var actions []*Action
	var businessMonitorTable []*m.BusinessMonitorTable
	x.SQL("select * from business_monitor where endpoint_id=?", param.EndpointId).Find(&businessMonitorTable)
	if len(param.PathList) == 0 {
		if len(businessMonitorTable) == 0 {
			return nil
		}
		actions = append(actions, &Action{Sql: "delete from business_monitor where endpoint_id=?", Param: []interface{}{param.EndpointId}})
		for _,v := range businessMonitorTable {
			actions = append(actions, &Action{Sql: "delete from business_monitor_cfg where business_monitor_id=?", Param: []interface{}{v.Id}})
		}
		return Transaction(actions)
	}
	for _,v := range param.PathList {
		if v.Id == 0 {
			actions = append(actions, &Action{Sql: "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)", Param: []interface{}{param.EndpointId,v.Path,v.OwnerEndpoint}})
			continue
		}
		for _,vv := range businessMonitorTable {
			if v.Id == vv.Id {
				actions = append(actions, &Action{Sql: "update business_monitor set path=?,owner_endpoint=? where id=?", Param: []interface{}{v.Path,v.OwnerEndpoint,v.Id}})
				break
			}
		}
	}
	for _,v := range businessMonitorTable {
		delFlag := true
		for _,vv := range param.PathList {
			if vv.Id == v.Id {
				delFlag = false
				break
			}
		}
		if delFlag {
			actions = append(actions, &Action{Sql: "delete from business_monitor where id=?", Param: []interface{}{v.Id}})
		}
	}
	err := Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update business_monitor table fail,%s ", err.Error())
		return err
	}
	actions = []*Action{}
	businessMonitorTable = []*m.BusinessMonitorTable{}
	x.SQL("select * from business_monitor where endpoint_id=?", param.EndpointId).Find(&businessMonitorTable)
	for _,v := range param.PathList {
		tmpId := v.Id
		for _,vv := range businessMonitorTable {
			if v.Path == vv.Path {
				tmpId = vv.Id
				break
			}
		}
		actions = append(actions, &Action{Sql: "delete from business_monitor_cfg where business_monitor_id=?", Param: []interface{}{tmpId}})
		for _,rule := range v.Rules {
			stringMapBytes,_ := json.Marshal(rule.StringMap)
			metricConfigBytes,_ := json.Marshal(rule.MetricConfig)
			actions = append(actions, &Action{Sql: "insert into business_monitor_cfg(business_monitor_id,regular,tags,string_map,metric_config) value (?,?,?,?,?)", Param: []interface{}{tmpId,rule.Regular,rule.Tags,string(stringMapBytes),string(metricConfigBytes)}})
		}
	}
	return Transaction(actions)
}

func UpdateAppendBusiness(param m.BusinessUpdateDto) error {
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
		actions = append(actions, &Action{Sql: "DELETE FROM business_monitor WHERE endpoint_id=? AND owner_endpoint=?", Param: []interface{}{param.EndpointId, v}})
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

func GetBusinessPromMetric(keys []string) (err error,result []*m.PromMetricTable) {
	if len(keys) == 0 {
		return err,result
	}
	sql := "SELECT * FROM monitor.prom_metric where "
	for i,v := range keys {
		if v == "" {
			continue
		}
		sql += " (prom_ql like '%key=\"" + v + "\"%') "
		if i < len(keys)-1 {
			sql += " OR "
		}
	}
	err = x.SQL(sql).Find(&result)
	return err,result
}