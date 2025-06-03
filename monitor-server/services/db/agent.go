package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/cipher"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strings"
	"time"
)

func UpdateEndpoint(endpoint *m.EndpointTable, extendParam, operator string) (stepList []int, err error) {
	stepList = append(stepList, endpoint.Step)
	if endpoint.Cluster == "" {
		endpoint.Cluster = "default"
	}
	host := m.EndpointTable{Guid: endpoint.Guid}
	GetEndpoint(&host)
	nowTime := time.Now().Format(m.DatetimeFormat)
	var actions []*Action
	if host.Id == 0 {
		existMonitorTypeList, _ := GetSimpleEndpointTypeList()
		addMonitorTypeFlag := true
		for _, v := range existMonitorTypeList {
			if endpoint.ExportType == v {
				addMonitorTypeFlag = false
				break
			}
		}
		if addMonitorTypeFlag {
			actions = append(actions, &Action{Sql: "INSERT INTO monitor_type (guid,display_name) VALUES (?,?)", Param: []interface{}{endpoint.ExportType, endpoint.ExportType}})
		}
		insert := fmt.Sprintf("INSERT INTO endpoint(guid,name,ip,endpoint_version,export_type,export_version,step,address,os_type,create_at,address_agent,cluster,tags) VALUE ('%s','%s','%s','%s','%s','%s','%d','%s','%s','%s','%s','%s','%s')",
			endpoint.Guid, endpoint.Name, endpoint.Ip, endpoint.EndpointVersion, endpoint.ExportType, endpoint.ExportVersion, endpoint.Step, endpoint.Address, endpoint.OsType, nowTime, endpoint.AddressAgent, endpoint.Cluster, endpoint.Tags)
		actions = append(actions, &Action{Sql: insert})
		// V2
		tmpAgentAddress := endpoint.Address
		if endpoint.AddressAgent != "" {
			tmpAgentAddress = endpoint.AddressAgent
		}
		actions = append(actions, &Action{Sql: "insert into endpoint_new(guid,name,ip,monitor_type,agent_version,agent_address,step,endpoint_version,endpoint_address,cluster,extend_param,update_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
			Param: []interface{}{endpoint.Guid, endpoint.Name, endpoint.Ip, endpoint.ExportType, endpoint.ExportVersion, tmpAgentAddress, endpoint.Step, endpoint.EndpointVersion, endpoint.Address, endpoint.Cluster, extendParam, nowTime, operator, operator}})
		err = Transaction(actions)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Insert endpoint fail", zap.Error(err))
			return
		}
		host := m.EndpointTable{Guid: endpoint.Guid}
		GetEndpoint(&host)
		endpoint.Id = host.Id
	} else {
		if host.Step != endpoint.Step {
			stepList = append(stepList, host.Step)
		}
		update := fmt.Sprintf("UPDATE endpoint SET name='%s',ip='%s',endpoint_version='%s',export_type='%s',export_version='%s',step=%d,address='%s',os_type='%s',address_agent='%s',cluster='%s',tags='%s' WHERE id=%d",
			endpoint.Name, endpoint.Ip, endpoint.EndpointVersion, endpoint.ExportType, endpoint.ExportVersion, endpoint.Step, endpoint.Address, endpoint.OsType, endpoint.AddressAgent, endpoint.Cluster, endpoint.Tags, host.Id)
		actions = append(actions, &Action{Sql: update})
		tmpAgentAddress := endpoint.Address
		if endpoint.AddressAgent != "" {
			tmpAgentAddress = endpoint.AddressAgent
		}
		actions = append(actions, &Action{Sql: "update endpoint_new set agent_address=?,step=?,endpoint_version=?,endpoint_address=?,extend_param=?,update_time=? where guid=?", Param: []interface{}{tmpAgentAddress, endpoint.Step, endpoint.EndpointVersion, endpoint.Address, extendParam, nowTime, endpoint.Guid}})
		err = Transaction(actions)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update endpoint fail", zap.Error(err))
			return
		}
		endpoint.Id = host.Id
	}
	return
}

func AddCustomMetric(param m.TransGatewayMetricDto) error {
	distinctMetricMap := make(map[string]bool)
	var err error
	for _, v := range param.Params {
		var endpointObjs []*m.EndpointTable
		x.SQL("SELECT id FROM endpoint WHERE export_type='custom' AND name=?", v.Name).Find(&endpointObjs)
		if len(endpointObjs) == 0 {
			continue
		}
		tmpEndpointId := endpointObjs[0].Id
		var endpointMetricObjs []*m.EndpointMetricTable
		x.SQL("SELECT * FROM endpoint_metric WHERE endpoint_id=?", tmpEndpointId).Find(&endpointMetricObjs)
		var tmpMetricList []string
		for _, vv := range v.Metrics {
			distinctMetricMap[vv] = true
			existFlag := false
			for _, vvv := range endpointMetricObjs {
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
		for _, vv := range tmpMetricList {
			insertSql = insertSql + fmt.Sprintf("(%d,'%s'),", tmpEndpointId, vv)
		}
		insertSql = insertSql[:len(insertSql)-1]
		_, err = x.Exec("INSERT INTO endpoint_metric(endpoint_id,metric) VALUES " + insertSql)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update custom endpoint_metric fail", zap.Error(err))
		}
	}
	if err != nil {
		return err
	}
	if len(distinctMetricMap) == 0 {
		return nil
	}
	var promMetricObjs []*m.PromMetricTable
	var whereSql, insertSql string
	for k, _ := range distinctMetricMap {
		whereSql = whereSql + fmt.Sprintf("'%s',", k)
	}
	whereSql = whereSql[:len(whereSql)-1]
	x.SQL("SELECT * FROM prom_metric WHERE metric_type='custom' AND metric IN (" + whereSql + ")").Find(&promMetricObjs)
	for k, _ := range distinctMetricMap {
		existFlag := false
		for _, v := range promMetricObjs {
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
		_, err = x.Exec("INSERT INTO prom_metric(metric,metric_type,prom_ql) VALUES " + insertSql)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update custom prom_metric fail", zap.Error(err))
			return err
		}
	}
	return nil
}

func DeleteEndpoint(guid, operator string) error {
	var actions []*Action
	nowTime := time.Now().Format(m.DatetimeFormat)
	actions = append(actions, &Action{Sql: "DELETE FROM endpoint_metric WHERE endpoint_id IN (SELECT id FROM endpoint WHERE guid=?)", Param: []interface{}{guid}})
	actions = append(actions, &Action{Sql: "DELETE FROM endpoint WHERE guid=?", Param: []interface{}{guid}})
	var alarms []*m.AlarmTable
	x.SQL("SELECT id FROM alarm WHERE endpoint=? AND status='firing'", guid).Find(&alarms)
	if len(alarms) > 0 {
		var ids string
		for _, v := range alarms {
			ids += fmt.Sprintf("%d,", v.Id)
		}
		ids = ids[:len(ids)-1]
		actions = append(actions, &Action{Sql: fmt.Sprintf("UPDATE alarm SET STATUS='closed' WHERE id IN (%s)", ids), Param: []interface{}{}})
	}
	var endpointGroup []*m.EndpointGroupRelTable
	x.SQL("select * from endpoint_group_rel where endpoint=?", guid).Find(&endpointGroup)
	var serviceGroup []*m.EndpointServiceRelTable
	x.SQL("select * from endpoint_service_rel where endpoint=?", guid).Find(&serviceGroup)
	actions = append(actions, &Action{Sql: "delete from endpoint_group_rel where endpoint=?", Param: []interface{}{guid}})
	for _, v := range endpointGroup {
		actions = append(actions, &Action{Sql: "update endpoint_group set update_time=?,update_user=? where guid=?", Param: []interface{}{nowTime, operator, v.EndpointGroup}})
	}
	actions = append(actions, &Action{Sql: "delete from endpoint_service_rel where endpoint=?", Param: []interface{}{guid}})
	for _, v := range serviceGroup {
		actions = append(actions, &Action{Sql: "update service_group set update_time=?,update_user=? where guid=?", Param: []interface{}{nowTime, operator, v.ServiceGroup}})
	}
	actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where source_endpoint=? or target_endpoint=?", Param: []interface{}{guid, guid}})
	actions = append(actions, &Action{Sql: "delete from db_metric_endpoint_rel where source_endpoint=? or target_endpoint=?", Param: []interface{}{guid, guid}})
	actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where source_endpoint=? or target_endpoint=?", Param: []interface{}{guid, guid}})
	actions = append(actions, &Action{Sql: "delete from db_keyword_endpoint_rel where source_endpoint=? or target_endpoint=?", Param: []interface{}{guid, guid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_tagvalue where dashboard_chart_tag in (select guid from custom_chart_series_tag where dashboard_chart_config in (select guid from custom_chart_series where endpoint=?))", Param: []interface{}{guid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_tag where dashboard_chart_config in (select guid from custom_chart_series where endpoint=?)", Param: []interface{}{guid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series_config where dashboard_chart_config in (select guid from custom_chart_series where endpoint=?)", Param: []interface{}{guid}})
	actions = append(actions, &Action{Sql: "delete from custom_chart_series where endpoint=?", Param: []interface{}{guid}})
	actions = append(actions, &Action{Sql: "delete from endpoint_new where guid=?", Param: []interface{}{guid}})
	err := Transaction(actions)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Delete endpoint fail", zap.Error(err))
		return err
	} else {
		for _, v := range endpointGroup {
			SyncPrometheusRuleFile(v.EndpointGroup, false)
		}
		for _, v := range serviceGroup {
			UpdateServiceConfigWithParent(v.ServiceGroup)
		}
	}
	return nil
}

func UpdateEndpointAlarmFlag(isStop bool, exportType, instance, ip, port, pod, k8sCluster string) error {
	var endpoints []*m.EndpointNewTable
	if exportType == "host" {
		x.SQL("SELECT guid FROM endpoint_new WHERE monitor_type=? AND ip=?", exportType, ip).Find(&endpoints)
	} else if exportType == "pod" {
		if k8sCluster != "" {
			x.SQL("select * from endpoint where name=? and os_type=? and export_type='pod'", pod, k8sCluster).Find(&endpoints)
		} else {
			x.SQL("select * from endpoint where name=? and export_type='pod'", pod).Find(&endpoints)
		}
	} else {
		if port != "" {
			x.SQL("SELECT guid FROM endpoint_new WHERE monitor_type=? AND endpoint_address=? AND name=?", exportType, fmt.Sprintf("%s:%s", ip, port), instance).Find(&endpoints)
		} else {
			x.SQL("SELECT guid FROM endpoint_new WHERE monitor_type=? AND ip=? AND name=?", exportType, ip, instance).Find(&endpoints)
		}
	}
	if len(endpoints) > 0 {
		var actions []*Action
		if isStop {
			actions = append(actions, &Action{Sql: "UPDATE endpoint SET stop_alarm=1 WHERE guid=?", Param: []interface{}{endpoints[0].Guid}})
			actions = append(actions, &Action{Sql: "UPDATE endpoint_new SET alarm_enable=0 WHERE guid=?", Param: []interface{}{endpoints[0].Guid}})
		} else {
			actions = append(actions, &Action{Sql: "UPDATE endpoint SET stop_alarm=0 WHERE guid=?", Param: []interface{}{endpoints[0].Guid}})
			actions = append(actions, &Action{Sql: "UPDATE endpoint_new SET alarm_enable=1 WHERE guid=?", Param: []interface{}{endpoints[0].Guid}})
		}
		return Transaction(actions)
	} else {
		return fmt.Errorf("Can not find this monitor object with %s %s %s %s \n", exportType, instance, ip, port)
	}
}

func UpdateRecursivePanel(param m.PanelRecursiveTable, operator string) error {
	var prt []*m.PanelRecursiveTable
	err := x.SQL("SELECT * FROM panel_recursive WHERE guid=?", param.Guid).Find(&prt)
	if err != nil {
		return err
	}
	nowTime := time.Now().Format(m.DatetimeFormat)
	var actions []*Action
	if len(prt) > 0 {
		tmpParent := unionList(param.Parent, prt[0].Parent, "^")
		tmpEndpoint := unionList(param.Endpoint, prt[0].Endpoint, "^")
		//_, err = x.Exec("UPDATE panel_recursive SET display_name=?,parent=?,endpoint=?,email=?,phone=?,role=?,firing_callback_key=?,recover_callback_key=?,obj_type=? WHERE guid=?", param.DisplayName, tmpParent, tmpEndpoint, param.Email, param.Phone, param.Role, param.FiringCallbackKey, param.RecoverCallbackKey, param.ObjType, param.Guid)
		actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET display_name=?,parent=?,endpoint=?,email=?,phone=?,role=?,firing_callback_key=?,recover_callback_key=?,obj_type=? WHERE guid=?", Param: []interface{}{param.DisplayName, tmpParent, tmpEndpoint, param.Email, param.Phone, param.Role, param.FiringCallbackKey, param.RecoverCallbackKey, param.ObjType, param.Guid}})
		actions = append(actions, &Action{Sql: "update service_group set display_name=?,service_type=?,update_time=?,update_user=? where guid=?", Param: []interface{}{param.DisplayName, param.ObjType, nowTime, operator, param.Guid}})
		endpointList := strings.Split(tmpEndpoint, "^")
		actions = append(actions, getUpdateServiceEndpointAction(param.Guid, nowTime, operator, endpointList)...)
		actions = append(actions, getUpdateServiceGroupNotifyActions(param.Guid, param.FiringCallbackKey, param.RecoverCallbackKey, strings.Split(param.Role, ","))...)
		err = Transaction(actions)
		if err == nil {
			var endpointGroup []*m.EndpointGroupTable
			parentGuidList, _ := fetchGlobalServiceGroupParentGuidList(param.Guid)
			x.SQL("select guid from endpoint_group where service_group in ('" + strings.Join(parentGuidList, "','") + "')").Find(&endpointGroup)
			for _, v := range endpointGroup {
				err = SyncPrometheusRuleFile(v.Guid, false)
				if err != nil {
					log.Error(nil, log.LOGGER_APP, "UpdateRecursivePanel warn,syncPrometheusRule fail", zap.Error(err))
				}
			}
			if err == nil {
				UpdateServiceConfigWithParent(param.Guid)
			}
		}
	} else {
		//_, err = x.Exec("INSERT INTO panel_recursive(guid,display_name,parent,endpoint,email,phone,role,firing_callback_key,recover_callback_key,obj_type) VALUE (?,?,?,?,?,?,?,?,?,?)", param.Guid, param.DisplayName, param.Parent, param.Endpoint, param.Email, param.Phone, param.Role, param.FiringCallbackKey, param.RecoverCallbackKey, param.ObjType)
		actions = append(actions, &Action{Sql: "INSERT INTO panel_recursive(guid,display_name,parent,endpoint,email,phone,role,firing_callback_key,recover_callback_key,obj_type) VALUE (?,?,?,?,?,?,?,?,?,?)", Param: []interface{}{param.Guid, param.DisplayName, param.Parent, param.Endpoint, param.Email, param.Phone, param.Role, param.FiringCallbackKey, param.RecoverCallbackKey, param.ObjType}})
		actions = append(actions, getCreateServiceGroupAction(&m.ServiceGroupTable{Guid: param.Guid, DisplayName: param.DisplayName, Description: "", Parent: param.Parent, ServiceType: param.ObjType, UpdateTime: nowTime}, operator)...)
		actions = append(actions, getUpdateServiceEndpointAction(param.Guid, nowTime, operator, strings.Split(param.Endpoint, "^"))...)
		actions = append(actions, getUpdateServiceGroupNotifyActions(param.Guid, param.FiringCallbackKey, param.RecoverCallbackKey, strings.Split(param.Role, ","))...)
		err = Transaction(actions)
		if err == nil {
			addGlobalServiceGroupNode(m.ServiceGroupTable{Guid: param.Guid, Parent: param.Parent, DisplayName: param.DisplayName})
		}
	}
	return err
}

func UpdateRecursiveEndpoint(guid, operator string, endpoint []string) error {
	var prt []*m.PanelRecursiveTable
	err := x.SQL("SELECT * FROM panel_recursive WHERE guid=?", guid).Find(&prt)
	if err != nil {
		return err
	}
	if len(prt) == 0 {
		return fmt.Errorf("update recursive endpoint error,no recored find with guid:%s", guid)
	}
	tmpEndpoint := strings.Split(prt[0].Endpoint, "^")
	if len(tmpEndpoint) == 0 {
		return nil
	}
	var newEndpoint []string
	for _, v := range tmpEndpoint {
		tmpFlag := false
		for _, vv := range endpoint {
			if vv == v {
				tmpFlag = true
				break
			}
		}
		if !tmpFlag {
			newEndpoint = append(newEndpoint, v)
		}
	}
	nowTime := time.Now().Format(m.DatetimeFormat)
	var actions []*Action
	actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET endpoint=? WHERE guid=?", Param: []interface{}{strings.Join(newEndpoint, "^"), guid}})
	actions = append(actions, getUpdateServiceEndpointAction(guid, nowTime, operator, newEndpoint)...)
	err = Transaction(actions)
	return err
}

func DeleteRecursivePanel(guid string) (err error) {
	var actions []*Action
	var tableData []*m.PanelRecursiveTable
	x.SQL("SELECT guid,display_name,parent FROM panel_recursive").Find(&tableData)
	if len(tableData) == 0 {
		return nil
	}
	guidList := getNodeFromParent(tableData, []string{guid}, guid)
	tmpMap := make(map[string]bool)
	for _, v := range guidList {
		tmpMap[v] = true
	}
	guidList = []string{}
	for k, _ := range tmpMap {
		if k != "" {
			guidList = append(guidList, k)
		}
	}
	actions = append(actions, &Action{Sql: "DELETE FROM panel_recursive WHERE guid in ('" + strings.Join(guidList, "','") + "')", Param: []interface{}{}})
	actions = append(actions, getDeleteServiceGroupAction(guid, guidList)...)
	err = Transaction(actions)
	if err == nil {
		DeleteServiceWithChildConfig(guid)
		deleteGlobalServiceGroupNode(guid)
	}
	return err
}

func unionList(param, exist, split string) string {
	paramList := strings.Split(param, split)
	existList := strings.Split(exist, split)
	for _, v := range paramList {
		tmpExist := false
		for _, vv := range existList {
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

// 查询所有类型
func SearchRecursivePanelAll(search string) []*m.OptionModel {
	var options []*m.OptionModel
	var prt []*m.PanelRecursiveTable
	if search == "." {
		search = ""
	}
	search = `%` + search + `%`
	x.SQL(`SELECT * FROM panel_recursive WHERE display_name LIKE  ? limit 100`, search).Find(&prt)
	for _, v := range prt {
		options = append(options, &m.OptionModel{Id: -1, OptionValue: v.Guid, OptionText: v.DisplayName, OptionType: "sys", OptionTypeName: v.ObjType, AppObject: v.Guid})
	}
	return options
}

// 按类型过滤
func SearchRecursivePanelByType(search, optionTypeName string) []*m.OptionModel {
	var options []*m.OptionModel
	var prt []*m.PanelRecursiveTable
	if search == "." {
		search = ""
	}
	search = `%` + search + `%`
	x.SQL(`SELECT * FROM panel_recursive WHERE display_name LIKE ? AND obj_type=? limit 100`, search, optionTypeName).Find(&prt)
	for _, v := range prt {
		options = append(options, &m.OptionModel{Id: -1, OptionValue: v.Guid, OptionText: v.DisplayName, OptionType: "sys", OptionTypeName: v.ObjType, AppObject: v.Guid})
	}
	return options
}

func ListRecursiveEndpointType(guid string) (result []string, err error) {
	var endpointRows []*m.EndpointNewTable
	err = x.SQL("select monitor_type from endpoint_new where guid=?", guid).Find(&endpointRows)
	if err != nil {
		err = fmt.Errorf("query endpoint table fail,%s ", err.Error())
		return
	}
	if len(endpointRows) > 0 {
		result = []string{endpointRows[0].MonitorType}
		return
	}
	result = []string{}
	resultMap := make(map[string]int)
	for _, v := range getRecursiveEndpointList(guid) {
		tmpType := v[strings.LastIndex(v, "_")+1:]
		if tmpType == "" {
			continue
		}
		if _, b := resultMap[tmpType]; !b {
			result = append(result, tmpType)
			resultMap[tmpType] = 1
		}
	}
	if _, ok := resultMap["process"]; ok {
		newResult := []string{"process"}
		for _, v := range result {
			if v != "process" {
				newResult = append(newResult, v)
			}
		}
		result = newResult
	}
	return
}

func GetRecursiveEndpointByType(guid, endpointType string) (result []*m.EndpointTable, err error) {
	result = []*m.EndpointTable{}
	guidList := []string{}
	for _, v := range getRecursiveEndpointList(guid) {
		if strings.HasSuffix(v, fmt.Sprintf("_%s", endpointType)) {
			guidList = append(guidList, v)
		}
	}
	err = x.SQL("select * from endpoint where guid in ('" + strings.Join(guidList, "','") + "')").Find(&result)
	return
}

func GetRecursiveEndpointByTypeNew(guid, endpointType string) (result []*m.EndpointNewTable, err error) {
	guidList, _ := fetchGlobalServiceGroupChildGuidList(guid)
	if len(guidList) == 0 {
		return
	}
	err = x.SQL("select * from endpoint_new where monitor_type=? and guid in (select endpoint from endpoint_service_rel where service_group in ('"+strings.Join(guidList, "','")+"'))", endpointType).Find(&result)
	return
}

func getRecursiveEndpointList(guid string) []string {
	result := []string{}
	resultMap := make(map[string]int)
	var prt []*m.PanelRecursiveTable
	x.SQL("select guid,endpoint from panel_recursive where guid=? or parent=?", guid, guid).Find(&prt)
	for _, v := range prt {
		if v.Guid == guid {
			for _, vv := range strings.Split(v.Endpoint, "^") {
				if vv == "" {
					continue
				}
				if _, b := resultMap[vv]; !b {
					result = append(result, vv)
					resultMap[vv] = 1
				}
			}
			continue
		}
		for _, vv := range getRecursiveEndpointList(v.Guid) {
			if vv == "" {
				continue
			}
			if _, b := resultMap[vv]; !b {
				result = append(result, vv)
				resultMap[vv] = 1
			}
		}
	}
	return result
}

func GetRecursivePanel(guid string) (err error, result m.RecursivePanelObj) {
	var prt []*m.PanelRecursiveTable
	err = x.SQL("SELECT guid,display_name,CONCAT(parent,'^') parent,endpoint FROM panel_recursive").Find(&prt)
	if err != nil {
		return err, result
	}
	result = recursiveData(guid, prt, len(prt), 1)
	return nil, result
}

func recursiveData(guid string, prt []*m.PanelRecursiveTable, length, depth int) m.RecursivePanelObj {
	var obj m.RecursivePanelObj
	if length < depth {
		return obj
	}
	tmp := guid + "^"
	for _, v := range prt {
		if v.Guid == guid {
			obj.DisplayName = v.DisplayName
			if v.Endpoint != "" {
				endpointList := strings.Split(v.Endpoint, "^")
				tmpMap := make(map[string][]string)
				for _, vv := range endpointList {
					tmpEndpointObj := m.EndpointTable{Guid: vv}
					GetEndpoint(&tmpEndpointObj)
					if tmpEndpointObj.ExportType == "" {
						continue
					}
					tmpType := tmpEndpointObj.ExportType
					if _, b := tmpMap[tmpType]; b {
						tmpMap[tmpType] = append(tmpMap[tmpType], vv)
					} else {
						tmpMap[tmpType] = []string{vv}
					}
				}
				for mk, mv := range tmpMap {
					chartTables := getChartsByEndpointType(mk)
					for _, cv := range chartTables {
						obj.Charts = append(obj.Charts, &m.ChartModel{Id: cv.Id, Endpoint: mv, Metric: strings.Split(cv.Metric, "^"), Aggregate: cv.AggType, MonitorType: mk})
					}
					for _, extendChart := range getServiceGroupCharts(mv, mk, v.Guid) {
						obj.Charts = append(obj.Charts, extendChart)
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

func getServiceGroupCharts(endpoints []string, monitorType, serviceGroup string) (result []*m.ChartModel) {
	result = []*m.ChartModel{}
	serviceGroupList, _ := fetchGlobalServiceGroupParentGuidList(serviceGroup)
	var chartTable []*m.ChartTable
	x.SQL("select * from chart where group_id in (select chart_group from panel where group_id in (select panels_group from dashboard where dashboard_type=?) and service_group in ('"+strings.Join(serviceGroupList, "','")+"'))", monitorType).Find(&chartTable)
	if len(chartTable) == 0 {
		return
	}
	for _, chart := range chartTable {
		result = append(result, &m.ChartModel{Id: chart.Id, Col: chart.Col, Title: chart.Title, Endpoint: endpoints, Metric: []string{chart.Metric}, MonitorType: monitorType, Aggregate: chart.AggType})
	}
	return result
}

func getExtendPanelCharts(endpoints []string, exportType, guid string) []*m.ChartModel {
	log.Debug(nil, log.LOGGER_APP, "getExtendPanel", zap.String("exportType", exportType), zap.Strings("endpoints", endpoints), zap.String("guid", guid))
	var result []*m.ChartModel
	if exportType == "java" {
		for _, endpoint := range endpoints {
			_, businessMonitor := GetBusinessList(0, endpoint)
			if len(businessMonitor) > 0 {
				businessMonitorMap := make(map[int][]string)
				for _, v := range businessMonitor {
					if _, b := businessMonitorMap[v.EndpointId]; b {
						exist := false
						for _, vv := range businessMonitorMap[v.EndpointId] {
							if vv == v.Path {
								exist = true
								break
							}
						}
						if !exist {
							businessMonitorMap[v.EndpointId] = append(businessMonitorMap[v.EndpointId], v.Path)
						}
					} else {
						businessMonitorMap[v.EndpointId] = []string{v.Path}
					}
				}
				businessCharts, businessPanels := GetBusinessPanelChart()
				if len(businessCharts) > 0 {
					chartsDto, _ := GetAutoDisplay(businessMonitorMap, businessPanels[0].TagsKey, businessCharts)
					for _, tmpChartModel := range chartsDto {
						result = append(result, tmpChartModel)
					}
				}
			}
		}
		// Agg same metric
		var newResult []*m.ChartModel
		for _, v := range result {
			metricIndex := -1
			for i, vv := range newResult {
				if v.Metric[0] == vv.Metric[0] {
					metricIndex = i
					break
				}
			}
			if metricIndex >= 0 {
				newResult[metricIndex].Endpoint = append(newResult[metricIndex].Endpoint, v.Endpoint...)
			} else {
				newResult = append(newResult, v)
			}
		}
		result = newResult
	}
	if exportType == "mysql" {
		dbMonitorList, _ := GetDbMonitorByPanel(guid)
		if len(dbMonitorList) > 0 {
			dbMonitorChart, _ := GetDbMonitorChart()
			if len(dbMonitorChart) > 0 {
				var tmpMetrics []string
				for _, v := range dbMonitorList {
					tmpMetrics = append(tmpMetrics, v.Name)
				}
				result = append(result, &m.ChartModel{Id: dbMonitorChart[0].Id, Endpoint: endpoints, Aggregate: dbMonitorChart[0].AggType, Metric: tmpMetrics})
			}
		}
	}
	return result
}

func getChartsByEndpointType(endpointType string) []*m.ChartTable {
	var result []*m.ChartTable
	x.SQL("SELECT t3.id,t3.group_id,t3.metric,t3.unit,t3.title,t3.agg_type FROM dashboard t1 LEFT JOIN panel t2 ON t1.panels_group=t2.group_id LEFT JOIN chart t3 ON t2.chart_group=t3.group_id WHERE t1.dashboard_type=? ORDER BY t3.group_id,t3.id", endpointType).Find(&result)
	if len(result) == 0 {
		return result
	}
	tmpGroupId := result[0].GroupId
	var ct []*m.ChartTable
	for _, v := range result {
		if v.GroupId == tmpGroupId {
			ct = append(ct, v)
		}
	}
	return ct
}

func UpdateEndpointTelnet(param m.UpdateEndpointTelnetParam) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM endpoint_telnet WHERE endpoint_guid=?", Param: []interface{}{param.Guid}})
	for _, v := range param.Config {
		actions = append(actions, &Action{Sql: "INSERT INTO endpoint_telnet(`endpoint_guid`,`port`,`note`) VALUE (?,?,?)", Param: []interface{}{param.Guid, v.Port, v.Note}})
	}
	err := Transaction(actions)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update endpoint table fail", zap.Error(err))
	}
	return err
}

func GetEndpointTelnet(guid string) (result []*m.EndpointTelnetTable, err error) {
	err = x.SQL("SELECT port,note FROM endpoint_telnet WHERE endpoint_guid=?", guid).Find(&result)
	return result, err
}

func UpdateEndpointHttp(param []*m.EndpointHttpTable) error {
	if len(param) == 0 {
		return nil
	}
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM endpoint_http WHERE endpoint_guid=?", Param: []interface{}{param[0].EndpointGuid}})
	for _, v := range param {
		actions = append(actions, &Action{Sql: "INSERT INTO endpoint_http(`endpoint_guid`,`method`,`url`) VALUE (?,?,?)", Param: []interface{}{v.EndpointGuid, v.Method, v.Url}})
	}
	err := Transaction(actions)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update endpoint http fail", zap.Error(err))
	}
	return err
}

func GetPingExporterSource() []*m.PingExportSourceObj {
	result := []*m.PingExportSourceObj{}
	var endpointTable []*m.EndpointTable
	x.SQL("select guid,ip from endpoint where address_agent='' and guid in (select endpoint from endpoint_group_rel where endpoint_group in (select endpoint_group from alarm_strategy where metric like 'ping_alive%'))").Find(&endpointTable)
	for _, v := range endpointTable {
		result = append(result, &m.PingExportSourceObj{Ip: v.Ip, Guid: v.Guid})
	}
	var telnetQuery []*m.TelnetSourceQuery
	x.SQL("SELECT t2.guid,t1.port,t2.ip FROM endpoint_telnet t1 JOIN endpoint t2 ON t1.endpoint_guid=t2.guid WHERE t2.address_agent=''").Find(&telnetQuery)
	if len(telnetQuery) > 0 {
		for _, v := range telnetQuery {
			if v.Ip != "" && v.Port > 0 {
				result = append(result, &m.PingExportSourceObj{Ip: fmt.Sprintf("%s:%d", v.Ip, v.Port), Guid: v.Guid})
			}
		}
	}
	var endpointHttpTable []*m.EndpointHttpTable
	x.SQL("SELECT t1.id,t1.endpoint_guid,t1.`method`,t1.url FROM endpoint_http t1 join endpoint t2 on t1.endpoint_guid=t2.guid where t2.address_agent=''").Find(&endpointHttpTable)
	if len(endpointHttpTable) > 0 {
		for _, v := range endpointHttpTable {
			tmpUrl := fmt.Sprintf("%s_%s", strings.ToUpper(v.Method), v.Url)
			result = append(result, &m.PingExportSourceObj{Ip: tmpUrl, Guid: v.EndpointGuid})
		}
	}
	return result
}

func UpdateAgentManagerTable(endpoint m.EndpointTable, user, password, configFile, binPath string, isAdd bool) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: fmt.Sprintf("DELETE FROM agent_manager WHERE endpoint_guid='%s'", endpoint.Guid)})
	if password != "" {
		encodePwd, encodeErr := cipher.AesEnPasswordByGuid(endpoint.Guid, m.Config().EncryptSeed, password, "")
		if encodeErr != nil {
			return encodeErr
		}
		password = encodePwd
	}
	if isAdd {
		var agentRemotePort string
		if splitIndex := strings.Index(endpoint.AddressAgent, ":"); splitIndex >= 0 {
			agentRemotePort = endpoint.AddressAgent[splitIndex+1:]
		}
		actions = append(actions, &Action{Sql: fmt.Sprintf("INSERT INTO agent_manager(endpoint_guid,name,user,password,instance_address,agent_address,config_file,bin_path,agent_remote_port) VALUE ('%s','%s','%s','%s','%s','%s','%s','%s','%s')", endpoint.Guid, endpoint.Name, user, password, endpoint.Address, endpoint.AddressAgent, configFile, binPath, agentRemotePort)})
	}
	return Transaction(actions)
}

func GetAgentManager(guid string) (result []*m.AgentManagerTable, err error) {
	if guid != "" {
		err = x.SQL("SELECT * FROM agent_manager where endpoint_guid=?", guid).Find(&result)
	} else {
		err = x.SQL("SELECT * FROM agent_manager").Find(&result)
	}
	for _, row := range result {
		if row.Password != "" {
			row.Password, _ = cipher.AesDePasswordByGuid(row.EndpointGuid, m.Config().EncryptSeed, row.Password)
		}
	}
	return result, err
}
