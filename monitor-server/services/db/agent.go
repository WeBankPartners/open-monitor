package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"time"
	"strings"
)

func UpdateEndpoint(endpoint *m.EndpointTable) error {
	host := m.EndpointTable{Guid:endpoint.Guid}
	GetEndpoint(&host)
	if host.Id == 0 {
		insert := fmt.Sprintf("INSERT INTO endpoint(guid,name,ip,endpoint_version,export_type,export_version,step,address,os_type,create_at,address_agent) VALUE ('%s','%s','%s','%s','%s','%s','%d','%s','%s','%s','%s')",
			endpoint.Guid,endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,time.Now().Format(m.DatetimeFormat),endpoint.AddressAgent)
		_,err := x.Exec(insert)
		if err != nil {
			mid.LogError("insert endpoint fail ", err)
			return err
		}
		host := m.EndpointTable{Guid:endpoint.Guid}
		GetEndpoint(&host)
		endpoint.Id = host.Id
	}else{
		update := fmt.Sprintf("UPDATE endpoint SET name='%s',ip='%s',endpoint_version='%s',export_type='%s',export_version='%s',step=%d,address='%s',os_type='%s',address_agent='%s' WHERE id=%d",
			endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,endpoint.AddressAgent,endpoint.Id)
		_,err := x.Exec(update)
		if err != nil {
			mid.LogError("update endpoint fail ", err)
			return err
		}
		endpoint.Id = host.Id
	}
	return nil
}

func AddCustomMetric(param m.TransGatewayMetricDto) error {
	distinctMetricMap := make(map[string]bool)
	var err error
	for _,v := range param.Params {
		var endpointObjs []*m.EndpointTable
		x.SQL("SELECT id FROM endpoint WHERE export_type='custom' AND name=?", v.Name).Find(&endpointObjs)
		if len(endpointObjs) == 0 {
			continue
		}
		tmpEndpointId := endpointObjs[0].Id
		var endpointMetricObjs []*m.EndpointMetricTable
		x.SQL("SELECT * FROM endpoint_metric WHERE endpoint_id=?", tmpEndpointId).Find(&endpointMetricObjs)
		var tmpMetricList []string
		for _,vv := range v.Metrics {
			distinctMetricMap[vv] = true
			existFlag := false
			for _,vvv := range endpointMetricObjs {
				if vvv.Metric == vv {
					existFlag = true
					break
				}
			}
			if !existFlag {
				tmpMetricList = append(tmpMetricList, vv)
			}
		}
		if len(tmpMetricList) == 0 {
			continue
		}
		var insertSql string
		for _,vv := range tmpMetricList {
			insertSql = insertSql + fmt.Sprintf("(%d,'%s'),", tmpEndpointId, vv)
		}
		insertSql = insertSql[:len(insertSql)-1]
		_,err = x.Exec("INSERT INTO endpoint_metric(endpoint_id,metric) VALUES " + insertSql)
		if err != nil {
			mid.LogError("update custom endpoint_metric fail", err)
		}
	}
	if err != nil {
		return err
	}
	if len(distinctMetricMap) == 0 {
		return nil
	}
	var promMetricObjs []*m.PromMetricTable
	var whereSql,insertSql string
	for k,_ := range distinctMetricMap {
		whereSql = whereSql + fmt.Sprintf("'%s',", k)
	}
	whereSql = whereSql[:len(whereSql)-1]
	x.SQL("SELECT * FROM prom_metric WHERE metric_type='custom' AND metric IN (" + whereSql + ")").Find(&promMetricObjs)
	for k,_ := range distinctMetricMap {
		existFlag := false
		for _,v := range promMetricObjs {
			if k == v.Metric {
				existFlag = true
			}
		}
		if !existFlag {
			insertSql = insertSql + fmt.Sprintf("('%s','custom','%s{instance=\"$address\"}'),", k, k)
		}
	}
	if insertSql != "" {
		insertSql = insertSql[:len(insertSql)-1]
		_,err = x.Exec("INSERT INTO prom_metric(metric,metric_type,prom_ql) VALUES " + insertSql)
		if err != nil {
			mid.LogError("update custom prom_metric fail", err)
			return err
		}
	}
	return nil
}

func DeleteEndpoint(guid string) error {
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM endpoint_metric WHERE endpoint_id IN (SELECT id FROM endpoint WHERE guid=?)", Param:[]interface{}{guid}})
	actions = append(actions, &Action{Sql:"DELETE FROM endpoint WHERE guid=?", Param:[]interface{}{guid}})
	var alarms []*m.AlarmTable
	x.SQL("SELECT id FROM alarm WHERE endpoint=? AND status='firing'", guid).Find(&alarms)
	if len(alarms) > 0 {
		var ids string
		for _,v := range alarms {
			ids += fmt.Sprintf("%d,", v.Id)
		}
		ids = ids[:len(ids)-1]
		actions = append(actions, &Action{Sql:fmt.Sprintf("UPDATE alarm SET STATUS='closed' WHERE id IN (%s)", ids), Param:[]interface{}{}})
	}
	err := Transaction(actions)
	if err != nil {
		mid.LogError("delete endpoint fail", err)
		return err
	}
	return nil
}

func UpdateEndpointAlarmFlag(isStop bool,exportType,instance,ip,port string) error {
	var endpoints []*m.EndpointTable
	if exportType == "host" {
		x.SQL("SELECT id FROM endpoint WHERE export_type=? AND ip=?", exportType, ip).Find(&endpoints)
	}else {
		if exportType == "tomcat" {
			x.SQL("SELECT id FROM endpoint WHERE export_type=? AND address=? AND name=?", exportType, fmt.Sprintf("%s:%s", ip, port), instance).Find(&endpoints)
		} else {
			x.SQL("SELECT id FROM endpoint WHERE export_type=? AND ip=? AND name=?", exportType, ip, instance).Find(&endpoints)
		}
	}
	mid.LogInfo(fmt.Sprintf("update endpoint alarm flag : query endpoints -> %v ", endpoints))
	if len(endpoints) > 0 {
		stopAlarm := "0"
		if isStop {
			stopAlarm = "1"
		}
		_,err := x.Exec(fmt.Sprintf("UPDATE endpoint SET stop_alarm=%s WHERE id=%d", stopAlarm, endpoints[0].Id))
		return err
	}else{
		return fmt.Errorf("Can not find this monitor object with %s %s %s %s \n", exportType,instance,ip,port)
	}
}

func UpdateRecursivePanel(param m.PanelRecursiveTable) error {
	var prt []*m.PanelRecursiveTable
	err := x.SQL("SELECT * FROM panel_recursive WHERE guid=?", param.Guid).Find(&prt)
	if err != nil {
		return err
	}
	if len(prt) > 0 {
		tmpParent := unionList(param.Parent, prt[0].Parent, "^")
		//tmpEndpoint := unionList(param.Endpoint, prt[0].Endpoint, "^")
		//tmpEmail := unionList(param.Email, prt[0].Email, ",")
		//tmpPhone := unionList(param.Phone, prt[0].Phone, ",")
		//tmpRole := unionList(param.Role, prt[0].Role, ",")
		_,err = x.Exec("UPDATE panel_recursive SET display_name=?,parent=?,endpoint=?,email=?,phone=?,role=?,firing_callback_key=?,recover_callback_key=?,obj_type=? WHERE guid=?", param.DisplayName,tmpParent,param.Endpoint,param.Email,param.Phone,param.Role,param.FiringCallbackKey,param.RecoverCallbackKey,param.ObjType,param.Guid)
	}else{
		_,err = x.Exec("INSERT INTO panel_recursive(guid,display_name,parent,endpoint,email,phone,role,firing_callback_key,recover_callback_key,obj_type) VALUE (?,?,?,?,?,?,?,?,?,?)", param.Guid,param.DisplayName,param.Parent,param.Endpoint,param.Email,param.Phone,param.Role,param.FiringCallbackKey,param.RecoverCallbackKey,param.ObjType)
	}
	return err
}

func DeleteRecursivePanel(guid string) error {
	_,err := x.Exec("DELETE FROM panel_recursive WHERE guid=?", guid)
	return err
}

func unionList(param,exist,split string) string {
	paramList := strings.Split(param, split)
	existList := strings.Split(exist, split)
	for _,v := range paramList {
		tmpExist := false
		for _,vv := range existList {
			if vv == v {
				tmpExist = true
				break
			}
		}
		if !tmpExist {
			existList = append(existList, v)
		}
	}
	return strings.Join(existList, split)
}

func SearchRecursivePanel(search string) []*m.OptionModel {
	options := []*m.OptionModel{}
	var prt []*m.PanelRecursiveTable
	if search == "." {
		search = ""
	}
	sql := `SELECT * FROM panel_recursive WHERE display_name LIKE  '%` + search + `%' limit 10`
	x.SQL(sql).Find(&prt)
	for _,v := range prt {
		//options = append(options, &m.OptionModel{Id:-1, OptionValue:fmt.Sprintf("%s:sys", v.Guid), OptionText:v.DisplayName})
		options = append(options, &m.OptionModel{Id:-1, OptionValue:v.Guid, OptionText:v.DisplayName, OptionType:"sys", OptionTypeName:v.ObjType})
	}
	return options
}

func GetRecursivePanel(guid string) (err error, result m.RecursivePanelObj) {
	var prt []*m.PanelRecursiveTable
	err = x.SQL("SELECT guid,display_name,CONCAT(parent,'^') parent,endpoint FROM panel_recursive").Find(&prt)
	if err != nil {
		return err,result
	}
	for i,v := range prt {
		mid.LogInfo(fmt.Sprintf("prt data %d -> %v", i ,v))
	}
	result = recursiveData(guid, prt, len(prt), 1)
	return nil,result
}

func recursiveData(guid string, prt []*m.PanelRecursiveTable, length,depth int) m.RecursivePanelObj {
	mid.LogInfo(fmt.Sprintf("recursive : guid->%s length->%d depth->%d", guid, length, depth))
	var obj m.RecursivePanelObj
	if length < depth {
		return obj
	}
	tmp := guid + "^"
	for _,v := range prt {
		if v.Guid == guid {
			obj.DisplayName = v.DisplayName
			if v.Endpoint != "" {
				endpointList := strings.Split(v.Endpoint, "^")
				tmpMap := make(map[string][]string)
				for _,vv := range endpointList {
					tmpEndpointObj := m.EndpointTable{Guid:vv}
					GetEndpoint(&tmpEndpointObj)
					if tmpEndpointObj.ExportType == "" {
						continue
					}
					tmpType := tmpEndpointObj.ExportType
					if _,b := tmpMap[tmpType]; b {
						tmpMap[tmpType] = append(tmpMap[tmpType], vv)
					}else{
						tmpMap[tmpType] = []string{vv}
					}
				}
				for mk,mv := range tmpMap {
					chartTables := getChartsByEndpointType(mk)
					for _,cv := range chartTables {
						obj.Charts = append(obj.Charts, &m.ChartModel{Id:cv.Id, Endpoint:mv, Metric:strings.Split(cv.Metric, "^"), Aggregate:cv.AggType})
					}
				}
				break
			}
			continue
		}
		if strings.Contains(v.Parent, tmp) {
			tmpObj := recursiveData(v.Guid, prt, length, depth+1)
			obj.Children = append(obj.Children, &tmpObj)
		}
	}
	return obj
}

func getChartsByEndpointType(endpointType string) []*m.ChartTable {
	var result []*m.ChartTable
	x.SQL("SELECT t3.id,t3.group_id,t3.metric,t3.unit,t3.title,t3.agg_type FROM dashboard t1 LEFT JOIN panel t2 ON t1.panels_group=t2.group_id LEFT JOIN chart t3 ON t2.chart_group=t3.group_id WHERE t1.dashboard_type=? ORDER BY t3.group_id,t3.id", endpointType).Find(&result)
	if len(result) == 0 {
		return result
	}
	tmpGroupId := result[0].GroupId
	var ct []*m.ChartTable
	for _,v := range result {
		if v.GroupId == tmpGroupId {
			ct = append(ct, v)
		}
	}
	return ct
}