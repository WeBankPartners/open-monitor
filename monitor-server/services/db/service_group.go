package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
)

var (
	globalServiceGroupMap = make(map[string]*models.ServiceGroupLinkNode)
)

func InitServiceGroup() {
	var serviceGroupTable []*models.ServiceGroupTable
	err := x.SQL("select guid,parent,display_name,service_type from service_group").Find(&serviceGroupTable)
	if err != nil {
		log.Logger.Error("Init service group fail", log.Error(err))
		return
	}
	if len(serviceGroupTable) == 0 {
		return
	}
	buildGlobalServiceGroupLink(serviceGroupTable)
}

func buildGlobalServiceGroupLink(serviceGroupTable []*models.ServiceGroupTable) {
	globalServiceGroupMap = make(map[string]*models.ServiceGroupLinkNode)
	for _, v := range serviceGroupTable {
		globalServiceGroupMap[v.Guid] = &models.ServiceGroupLinkNode{Guid: v.Guid}
	}
	for _, v := range serviceGroupTable {
		if v.Parent != "" {
			if _, b := globalServiceGroupMap[v.Parent]; b {
				globalServiceGroupMap[v.Guid].Parent = globalServiceGroupMap[v.Parent]
				globalServiceGroupMap[v.Parent].Children = append(globalServiceGroupMap[v.Parent].Children, globalServiceGroupMap[v.Guid])
			}
		}
	}
	displayGlobalServiceGroup()
}

func fetchGlobalServiceGroupChildGuidList(rootKey string) (result []string, err error) {
	if v, b := globalServiceGroupMap[rootKey]; b {
		result = v.FetchChildGuid()
	} else {
		err = fmt.Errorf("Can not find service group with guid:%s ", rootKey)
	}
	return
}

func addGlobalServiceGroupNode(param models.ServiceGroupTable) {
	displayGlobalServiceGroup()
	if _, b := globalServiceGroupMap[param.Guid]; !b {
		globalServiceGroupMap[param.Guid] = &models.ServiceGroupLinkNode{Guid: param.Guid}
		if param.Parent != "" {
			if _, bb := globalServiceGroupMap[param.Parent]; bb {
				globalServiceGroupMap[param.Guid] = &models.ServiceGroupLinkNode{Guid: param.Guid, Parent: globalServiceGroupMap[param.Parent]}
				globalServiceGroupMap[param.Parent].Children = append(globalServiceGroupMap[param.Parent].Children, globalServiceGroupMap[param.Guid])
			}
		}
		var serviceGroupTable []*models.ServiceGroupTable
		x.SQL("select guid,parent from service_group where parent=?", param.Guid).Find(&serviceGroupTable)
		if len(serviceGroupTable) > 0 {
			for _, v := range serviceGroupTable {
				if childNode, bb := globalServiceGroupMap[v.Guid]; bb {
					childNode.Parent = globalServiceGroupMap[param.Guid]
					globalServiceGroupMap[param.Guid].Children = append(globalServiceGroupMap[param.Guid].Children, childNode)
				}
			}
		}
	}
	displayGlobalServiceGroup()
}

func deleteGlobalServiceGroupNode(guid string) {
	log.Logger.Info("start deleteGlobalServiceGroupNode", log.String("guid", guid))
	if v, b := globalServiceGroupMap[guid]; b {
		if v.Parent != nil {
			newChildList := []*models.ServiceGroupLinkNode{}
			for _, child := range v.Parent.Children {
				if child.Guid != guid {
					newChildList = append(newChildList, child)
				}
			}
			v.Parent.Children = newChildList
		}
		for _, key := range v.FetchChildGuid() {
			delete(globalServiceGroupMap, key)
		}
	}
}

func displayGlobalServiceGroup()  {
	for k,v := range globalServiceGroupMap {
		if v.Parent != nil {
			log.Logger.Info("globalServiceGroupMap", log.String("k", k), log.String("parent", v.Parent.Guid))
		}else {
			log.Logger.Info("globalServiceGroupMap", log.String("k", k))
		}
	}
}

func ListServiceGroupOptions(searchText string) (result []*models.OptionModel, err error) {
	result = []*models.OptionModel{}
	if searchText == "." {
		searchText = ""
	}
	searchText = "%" + searchText + "%"
	var serviceGroupTable []*models.ServiceGroupTable
	err = x.SQL("select guid,service_type from service_group where guid like ?", searchText).Find(&serviceGroupTable)
	if err != nil {
		return
	}
	for _, v := range serviceGroupTable {
		result = append(result, &models.OptionModel{OptionValue: v.Guid, OptionText: v.Guid, OptionType: v.ServiceType, OptionTypeName: v.ServiceType})
	}
	return
}

func ListServiceGroup() (result []*models.ServiceGroupTable, err error) {
	result = []*models.ServiceGroupTable{}
	err = x.SQL("select * from service_group").Find(&result)
	return
}

func GetServiceGroupEndpointList(searchType string) (result []*models.ServiceGroupEndpointListObj, err error) {
	result = []*models.ServiceGroupEndpointListObj{}
	if searchType == "endpoint" {
		var endpointTable []*models.EndpointNewTable
		err = x.SQL("select guid from endpoint_new").Find(&endpointTable)
		for _, v := range endpointTable {
			result = append(result, &models.ServiceGroupEndpointListObj{Guid: v.Guid, DisplayName: v.Guid})
		}
	} else {
		var serviceGroupTable []*models.ServiceGroupTable
		err = x.SQL("select guid,display_name from service_group").Find(&serviceGroupTable)
		for _, v := range serviceGroupTable {
			result = append(result, &models.ServiceGroupEndpointListObj{Guid: v.Guid, DisplayName: v.DisplayName})
		}
	}
	return
}

func CreateServiceGroup(param *models.ServiceGroupTable) {

}

func getCreateServiceGroupAction(param *models.ServiceGroupTable) (actions []*Action) {
	if param.Parent == "" {
		actions = append(actions, &Action{Sql: "insert into service_group(guid,display_name,description,service_type,update_time) value (?,?,?,?,?)", Param: []interface{}{param.Guid, param.DisplayName, "", param.ServiceType, param.UpdateTime}})
	} else {
		actions = append(actions, &Action{Sql: "insert into service_group(guid,display_name,description,parent,service_type,update_time) value (?,?,?,?,?,?)", Param: []interface{}{param.Guid, param.DisplayName, "", param.Parent, param.ServiceType, param.UpdateTime}})
	}
	return actions
}

func UpdateServiceGroup(param *models.ServiceGroupTable) {

}

func getUpdateServiceEndpointAction(serviceGroupGuid, nowTime string, endpoint []string) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from endpoint_service_rel where service_group=?", Param: []interface{}{serviceGroupGuid}})
	if len(endpoint) == 0 {
		return actions
	}
	var endpointGroup []*models.EndpointGroupTable
	x.SQL("select guid,monitor_type from endpoint_group where service_group=?", serviceGroupGuid).Find(&endpointGroup)
	tmpMonitorTypeMap := make(map[string]int)
	for _, v := range endpointGroup {
		tmpMonitorTypeMap[v.MonitorType] = 1
	}
	for _, v := range endpoint {
		if !strings.Contains(v, "_") {
			continue
		}
		tmpMonitorType := v[strings.LastIndex(v, "_")+1:]
		actions = append(actions, &Action{Sql: "insert into endpoint_service_rel(guid,endpoint,service_group) value (?,?,?)", Param: []interface{}{guid.CreateGuid(), v, serviceGroupGuid}})
		if _, b := tmpMonitorTypeMap[tmpMonitorType]; !b {
			endpointGroupGuid := fmt.Sprintf("service_%s_%s", serviceGroupGuid, tmpMonitorType)
			actions = append(actions, &Action{Sql: "insert into endpoint_group(guid,display_name,monitor_type,service_group,update_time) value (?,?,?,?,?)", Param: []interface{}{endpointGroupGuid, endpointGroupGuid, tmpMonitorType, serviceGroupGuid, nowTime}})
			tmpMonitorTypeMap[tmpMonitorType] = 1
		}
	}
	return actions
}

func DeleteServiceGroup(serviceGroupGuid string) {

}

func getDeleteServiceGroupAction(serviceGroupGuid string) (actions []*Action) {
	guidList := []string{serviceGroupGuid}
	if sNode, b := globalServiceGroupMap[serviceGroupGuid]; b {
		guidList = sNode.FetchChildGuid()
	}
	var endpointGroup []*models.EndpointGroupTable
	x.SQL(fmt.Sprintf("select guid from endpoint_group where service_group in ('%s')", strings.Join(guidList, "','"))).Find(&endpointGroup)
	for _, v := range endpointGroup {
		actions = append(actions, getDeleteEndpointGroupAction(v.Guid)...)
	}
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from endpoint_service_rel where service_group in ('%s')", strings.Join(guidList, "','"))})
	actions = append(actions, &Action{Sql: fmt.Sprintf("DELETE FROM service_group WHERE guid in ('%s')", strings.Join(guidList, "','"))})
	return actions
}

func ListServiceGroupEndpoint(serviceGroup, monitorType string) (result []*models.ServiceGroupEndpointListObj, err error) {
	var guidList []string
	guidList, err = fetchGlobalServiceGroupChildGuidList(serviceGroup)
	if err != nil {
		return
	}
	result = []*models.ServiceGroupEndpointListObj{}
	var endpointServiceRel []*models.EndpointServiceRelTable
	err = x.SQL("select distinct t1.endpoint from endpoint_service_rel t1 left join endpoint_new t2 on t1.endpoint=t2.guid where t1.service_group in ('"+strings.Join(guidList, "','")+"') and t2.monitor_type=?", monitorType).Find(&endpointServiceRel)
	for _, v := range endpointServiceRel {
		result = append(result, &models.ServiceGroupEndpointListObj{Guid: v.Endpoint, DisplayName: v.Endpoint})
	}
	return
}

func getSimpleServiceGroup(serviceGroupGuid string) (result models.ServiceGroupTable, err error) {
	var serviceGroupTable []*models.ServiceGroupTable
	err = x.SQL("select * from service_group where guid=?", serviceGroupGuid).Find(&serviceGroupTable)
	if err != nil {
		return result, fmt.Errorf("Query service_group table fail,%s ", err.Error())
	}
	if len(serviceGroupTable) == 0 {
		return result, fmt.Errorf("Can not find service_group with guid:%s ", serviceGroupGuid)
	}
	result = *serviceGroupTable[0]
	return
}

func MatchServicePanel(endpointGuid string) (result models.PanelModel, err error) {
	result = models.PanelModel{Title: "service", Tags: models.TagsModel{Enable: false, Option: []*models.OptionModel{}}}
	var logMetricEndpointRel []*models.LogMetricEndpointRelTable
	err = x.SQL("select * from log_metric_endpoint_rel where target_endpoint=?", endpointGuid).Find(&logMetricEndpointRel)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_endpoint_rel fail,%s ", err.Error())
	}
	if len(logMetricEndpointRel) > 0 {
		logMetricMonitorList := []string{}
		for _, v := range logMetricEndpointRel {
			logMetricMonitorList = append(logMetricMonitorList, v.LogMetricMonitor)
		}
		var logMetricTable []*models.LogMetricConfigTable
		x.SQL("select metric,display_name,agg_type from log_metric_config where log_metric_monitor in ('" + strings.Join(logMetricMonitorList, "','") + "') or log_metric_json in (select guid from log_metric_json where log_metric_monitor in ('" + strings.Join(logMetricMonitorList, "','") + "'))").Find(&logMetricTable)
		for _, v := range logMetricTable {
			result.Charts = append(result.Charts, &models.ChartModel{Id: 0, Title: v.DisplayName, Endpoint: []string{endpointGuid}, Metric: []string{fmt.Sprintf("%s/key=%s,t_endpoint=%s,agg=%s", models.LogMetricName, v.Metric, endpointGuid, v.AggType)}})
		}
	}
	var dbMetricMonitor []*models.DbMetricMonitorTable
	x.SQL("select * from db_metric_monitor where guid in (select db_metric_monitor from db_metric_endpoint_rel where target_endpoint=?)", endpointGuid).Find(&dbMetricMonitor)
	for _, v := range dbMetricMonitor {
		result.Charts = append(result.Charts, &models.ChartModel{Id: 0, Title: v.DisplayName, Endpoint: []string{endpointGuid}, Metric: []string{fmt.Sprintf("%s/key=%s,t_endpoint=%s", models.DBMonitorMetricName, v.Metric, endpointGuid)}})
	}
	return
}

func UpdateServiceConfigWithEndpoint(serviceGroup string) {
	var endpointServiceRel []*models.EndpointServiceRelTable
	x.SQL("select * from endpoint_service_rel where service_group=?", serviceGroup).Find(&endpointServiceRel)
	var endpointList []string
	endpointTypeMap := make(map[string][]string)
	for _, v := range endpointServiceRel {
		if !strings.Contains(v.Endpoint, "_") {
			continue
		}
		tmpEndpointType := v.Endpoint[strings.LastIndex(v.Endpoint, "_")+1:]
		if vv, b := endpointTypeMap[tmpEndpointType]; b {
			vv = append(vv, v.Endpoint)
		} else {
			endpointTypeMap[tmpEndpointType] = []string{v.Endpoint}
		}
		endpointList = append(endpointList, v.Endpoint)
	}
	log.Logger.Info("UpdateServiceConfigWithEndpoint", log.String("serviceGroup", serviceGroup), log.StringList("endpointList", endpointList))
	err := UpdateLogMetricConfigByServiceGroup(serviceGroup, endpointTypeMap)
	if err != nil {
		log.Logger.Error("UpdateLogMetricConfigByServiceGroup fail", log.Error(err))
	}
	err = UpdateDbMetricConfigByServiceGroup(serviceGroup, endpointTypeMap)
	if err != nil {
		log.Logger.Error("UpdateDbMetricConfigByServiceGroup fail", log.Error(err))
	}
}

func UpdateLogMetricConfigByServiceGroup(serviceGroup string, endpointTypeMap map[string][]string) (err error) {
	var logMetricTable []*models.LogMetricMonitorTable
	x.SQL("select * from log_metric_monitor where service_group=?", serviceGroup).Find(&logMetricTable)
	if len(logMetricTable) == 0 {
		return
	}
	var hostEndpoint []string
	if hostListValue, b := endpointTypeMap["host"]; b {
		hostEndpoint = hostListValue
	}
	hostEndpointIpMap := make(map[string]string)
	if len(hostEndpoint) > 0 {
		var endpointTable []*models.EndpointNewTable
		x.SQL("select guid,ip from endpoint_new where guid in ('" + strings.Join(hostEndpoint, "','") + "')").Find(&endpointTable)
		for _, v := range endpointTable {
			hostEndpointIpMap[v.Guid] = v.Ip
		}
	}
	for _, v := range logMetricTable {
		UpdateLogMetricConfigAction(v, endpointTypeMap, hostEndpoint, hostEndpointIpMap)
	}
	return
}

func UpdateLogMetricConfigAction(logMonitor *models.LogMetricMonitorTable, endpointTypeMap map[string][]string, hostEndpoint []string, hostEndpointIpMap map[string]string) {
	var updateHostEndpointList []string
	var actions []*Action
	var logMetricRelTable []*models.LogMetricEndpointRelTable
	x.SQL("select * from log_metric_endpoint_rel where log_metric_monitor=?", logMonitor.Guid).Find(&logMetricRelTable)
	if len(logMetricRelTable) == 0 && len(hostEndpoint) == 0 {
		return
	}
	targetTypeMap := make(map[string]int)
	if targetTypeList, b := endpointTypeMap[logMonitor.MonitorType]; b {
		for _, target := range targetTypeList {
			targetTypeMap[target] = 1
		}
	}
	sourceTargetMap := make(map[string]string)
	for _, vv := range logMetricRelTable {
		sourceTargetMap[vv.SourceEndpoint] = vv.TargetEndpoint
	}
	for _, host := range hostEndpoint {
		if target, b := sourceTargetMap[host]; b {
			// target remove
			if _, bb := targetTypeMap[target]; !bb {
				actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=? and source_endpoint=?", Param: []interface{}{logMonitor.Guid, host}})
				updateHostEndpointList = append(updateHostEndpointList, host)
			}
		} else {
			for target, _ := range targetTypeMap {
				// match new target
				if strings.Contains(target, fmt.Sprintf("_%s_", hostEndpointIpMap[host])) {
					actions = append(actions, &Action{Sql: "insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), logMonitor.Guid, host, target}})
					updateHostEndpointList = append(updateHostEndpointList, host)
				}
			}
		}
	}
	for source, _ := range sourceTargetMap {
		existFlag := false
		for _, host := range hostEndpoint {
			if host == source {
				existFlag = true
				break
			}
		}
		// source remove
		if !existFlag {
			actions = append(actions, &Action{Sql: "delete from log_metric_endpoint_rel where log_metric_monitor=? and source_endpoint=?", Param: []interface{}{logMonitor.Guid, source}})
			updateHostEndpointList = append(updateHostEndpointList, source)
		}
	}
	if len(actions) > 0 {
		err := Transaction(actions)
		if err == nil {
			err = UpdateNodeExportConfig(updateHostEndpointList)
			if err != nil {
				log.Logger.Error("UpdateNodeExportConfig fail", log.String("logMetricMonitor", logMonitor.Guid), log.Error(err))
			}
		} else {
			log.Logger.Error("UpdateLogMetricConfigAction exec sql fail", log.String("logMetricMonitor", logMonitor.Guid), log.Error(err))
		}
	}
}

func UpdateDbMetricConfigByServiceGroup(serviceGroup string, endpointTypeMap map[string][]string) (err error) {
	var dbMetricTable []*models.DbMetricMonitorTable
	x.SQL("select * from db_metric_monitor where service_group=? and monitor_type='mysql'", serviceGroup).Find(&dbMetricTable)
	if len(dbMetricTable) == 0 {
		return
	}
	mysqlEndpointMap := make(map[string]int)
	if mysqlListValue, b := endpointTypeMap["mysql"]; b {
		for _, v := range mysqlListValue {
			mysqlEndpointMap[v] = 1
		}
	}
	for _, v := range dbMetricTable {
		tmpActions := []*Action{}
		var dbMetricRelTable []*models.DbMetricEndpointRelTable
		x.SQL("select * from db_metric_endpoint_rel where db_metric_monitor=?", v.Guid).Find(&dbMetricRelTable)
		for _, vv := range dbMetricRelTable {
			if _, b := mysqlEndpointMap[vv.SourceEndpoint]; !b {
				tmpActions = append(tmpActions, &Action{Sql: "delete from db_metric_endpoint_rel where db_metric_monitor=? and source_endpoint=?", Param: []interface{}{v.Guid, vv.SourceEndpoint}})
			}
		}
		for k, _ := range mysqlEndpointMap {
			existFlag := false
			for _, vv := range dbMetricRelTable {
				if vv.SourceEndpoint == k {
					existFlag = true
					break
				}
			}
			if !existFlag {
				tmpActions = append(tmpActions, &Action{Sql: "insert into db_metric_endpoint_rel(guid,db_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), v.Guid, k, k}})
			}
		}
		if len(tmpActions) == 0 {
			continue
		}
		err = Transaction(tmpActions)
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	err = SyncDbMetric()
	if err != nil {
		log.Logger.Error("UpdateDbMetricConfigByServiceGroup fail", log.String("serviceGroup", serviceGroup))
	}
	return
}

func DeleteServiceConfig(serviceGroup string) {
	var logMetricTable []*models.LogMetricMonitorTable
	x.SQL("select guid from log_metric_monitor where service_group=?", serviceGroup).Find(&logMetricTable)
	for _, v := range logMetricTable {
		tmpErr := DeleteLogMetricMonitor(v.Guid)
		if tmpErr != nil {
			log.Logger.Error("Try to DeleteLogMetricMonitor fail", log.Error(tmpErr))
		}
	}
	var dbMetricTable []*models.DbMetricMonitorTable
	x.SQL("select guid from db_metric_monitor where service_group=?", serviceGroup).Find(&dbMetricTable)
	for _, v := range dbMetricTable {
		tmpErr := DeleteDbMetric(v.Guid)
		if tmpErr != nil {
			log.Logger.Error("Try to DeleteDbMetric fail", log.Error(tmpErr))
		}
	}
	if len(dbMetricTable) > 0 {
		err := SyncDbMetric()
		if err != nil {
			log.Logger.Error("Try to SyncDbMetric fail", log.Error(err))
		}
	}
}
