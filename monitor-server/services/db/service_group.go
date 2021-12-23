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
		models.GlobalSGDisplayNameMap[v.Guid] = v.DisplayName
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

func fetchGlobalServiceGroupParentGuidList(childKey string) (result []string, err error) {
	if v, b := globalServiceGroupMap[childKey]; b {
		result = v.FetchParentGuid()
	} else {
		err = fmt.Errorf("Can not find service group with guid:%s ", childKey)
	}
	return
}

func addGlobalServiceGroupNode(param models.ServiceGroupTable) {
	displayGlobalServiceGroup()
	if _, b := globalServiceGroupMap[param.Guid]; !b {
		globalServiceGroupMap[param.Guid] = &models.ServiceGroupLinkNode{Guid: param.Guid}
		models.GlobalSGDisplayNameMap[param.Guid] = param.DisplayName
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
	displayGlobalServiceGroup()
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
	displayGlobalServiceGroup()
}

func displayGlobalServiceGroup() {
	for k, v := range globalServiceGroupMap {
		if v.Parent != nil {
			log.Logger.Info("globalServiceGroupMap", log.String("k", k), log.String("parent", v.Parent.Guid))
		} else {
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
	err = x.SQL("select * from service_group where guid like ?", searchText).Find(&serviceGroupTable)
	if err != nil {
		return
	}
	for _, v := range serviceGroupTable {
		result = append(result, &models.OptionModel{OptionValue: v.Guid, OptionText: v.DisplayName, OptionType: v.ServiceType, OptionTypeName: v.ServiceType})
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
		err = x.SQL("select guid,monitor_type from endpoint_new").Find(&endpointTable)
		for _, v := range endpointTable {
			result = append(result, &models.ServiceGroupEndpointListObj{Guid: v.Guid, DisplayName: v.Guid, Type: v.MonitorType})
		}
	} else {
		var serviceGroupTable []*models.ServiceGroupTable
		err = x.SQL("select guid,display_name,service_type from service_group").Find(&serviceGroupTable)
		for _, v := range serviceGroupTable {
			result = append(result, &models.ServiceGroupEndpointListObj{Guid: v.Guid, DisplayName: v.DisplayName, Type: v.ServiceType})
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
	for _, v := range endpoint {
		if !strings.Contains(v, "_") {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into endpoint_service_rel(guid,endpoint,service_group) value (?,?,?)", Param: []interface{}{guid.CreateGuid(), v, serviceGroupGuid}})
	}
	guidList, _ := fetchGlobalServiceGroupParentGuidList(serviceGroupGuid)
	for _, v := range guidList {
		actions = append(actions, getCreateEndpointGroupByServiceAction(v, nowTime, endpoint)...)
	}
	return actions
}

func getCreateEndpointGroupByServiceAction(serviceGroupGuid, nowTime string, endpoint []string) (actions []*Action) {
	actions = []*Action{}
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

func GetDeleteServiceGroupAffectList(serviceGroup string) (result []string, err error) {
	guidList, _ := fetchGlobalServiceGroupChildGuidList(serviceGroup)
	for _, sg := range guidList {
		logMetricConfig, tmpErr := GetLogMetricByServiceGroup(sg)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		for _, logMetricMonitor := range logMetricConfig.Config {
			for _, logMetricJson := range logMetricMonitor.JsonConfigList {
				for _, logMetricConfig := range logMetricJson.MetricList {
					result = append(result, fmt.Sprintf("logMetric  path:%s metric:%s", logMetricMonitor.LogPath, logMetricConfig.Metric))
				}
			}
			for _, logMetricConfig := range logMetricMonitor.MetricConfigList {
				result = append(result, fmt.Sprintf("logMetric  path:%s metric:%s", logMetricMonitor.LogPath, logMetricConfig.Metric))
			}
		}
		dbMetricConfig, tmpErr := GetDbMetricByServiceGroup(sg)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		for _, dbMetric := range dbMetricConfig {
			result = append(result, fmt.Sprintf("dbMetric metric:%s", dbMetric.Metric))
		}
		keyWordConfigList, tmpErr := GetLogKeywordByServiceGroup(sg)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		if len(keyWordConfigList) == 0 {
			continue
		}
		for _, keywordConfig := range keyWordConfigList[0].Config {
			for _, keyword := range keywordConfig.KeywordList {
				result = append(result, fmt.Sprintf("logKeywrod path:%s keyword:%s", keywordConfig.LogPath, keyword.Keyword))
			}
		}
	}
	return
}

func getDeleteServiceGroupAction(serviceGroupGuid string) (actions []*Action) {
	guidList := []string{serviceGroupGuid}
	if sNode, b := globalServiceGroupMap[serviceGroupGuid]; b {
		guidList = sNode.FetchChildGuid()
	}
	guidFilterString := strings.Join(guidList, "','")
	var endpointGroup []*models.EndpointGroupTable
	x.SQL(fmt.Sprintf("select guid from endpoint_group where service_group in ('%s')", guidFilterString)).Find(&endpointGroup)
	for _, v := range endpointGroup {
		actions = append(actions, getDeleteEndpointGroupAction(v.Guid)...)
	}
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from endpoint_service_rel where service_group in ('%s')", guidFilterString)})
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from service_group_role_rel where service_group in ('%s')", guidFilterString)})
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from notify_role_rel where notify in (select guid from notify where service_group in ('%s'))", guidFilterString)})
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from notify where service_group in ('%s')", guidFilterString)})
	actions = append(actions, &Action{Sql: fmt.Sprintf("DELETE FROM service_group WHERE guid in ('%s')", guidFilterString)})
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
	err = x.SQL("select distinct log_metric_monitor from log_metric_endpoint_rel where target_endpoint=?", endpointGuid).Find(&logMetricEndpointRel)
	if err != nil {
		return result, fmt.Errorf("Query table log_metric_endpoint_rel fail,%s ", err.Error())
	}
	if len(logMetricEndpointRel) > 0 {
		for _, endpointRel := range logMetricEndpointRel {
			logMetricMonitor := endpointRel.LogMetricMonitor
			serviceGroup, _ := getLogMetricServiceGroup(logMetricMonitor)
			for _, jsonConfig := range ListLogMetricJson(logMetricMonitor) {
				for _, metricConfig := range jsonConfig.MetricList {
					result.Charts = append(result.Charts, &models.ChartModel{Id: 0, Title: metricConfig.DisplayName, Endpoint: []string{endpointGuid}, Metric: []string{fmt.Sprintf("%s/key=%s,t_endpoint=%s,agg=%s,service_group=%s", models.LogMetricName, metricConfig.Metric, endpointGuid, metricConfig.AggType, serviceGroup)}})
				}
			}
			for _, metricConfig := range ListLogMetricConfig("", logMetricMonitor) {
				result.Charts = append(result.Charts, &models.ChartModel{Id: 0, Title: metricConfig.DisplayName, Endpoint: []string{endpointGuid}, Metric: []string{fmt.Sprintf("%s/key=%s,t_endpoint=%s,agg=%s,service_group=%s", models.LogMetricName, metricConfig.Metric, endpointGuid, metricConfig.AggType, serviceGroup)}})
			}
		}
	}
	var dbMetricMonitor []*models.DbMetricMonitorTable
	x.SQL("select * from db_metric_monitor where guid in (select db_metric_monitor from db_metric_endpoint_rel where target_endpoint=?)", endpointGuid).Find(&dbMetricMonitor)
	for _, v := range dbMetricMonitor {
		result.Charts = append(result.Charts, &models.ChartModel{Id: 0, Title: v.DisplayName, Endpoint: []string{endpointGuid}, Metric: []string{fmt.Sprintf("%s/key=%s,t_endpoint=%s,service_group=%s", models.DBMonitorMetricName, v.Metric, endpointGuid, v.ServiceGroup)}})
	}
	return
}

func UpdateServiceConfigWithParent(serviceGroup string) {
	guidList, _ := fetchGlobalServiceGroupParentGuidList(serviceGroup)
	for _, v := range guidList {
		UpdateServiceConfigWithEndpoint(v)
	}
}

func getServiceGroupEndpointWithType(serviceGroup,monitorType string) (result []*models.EndpointNewTable) {
	result = []*models.EndpointNewTable{}
	x.SQL("select guid,name,ip,monitor_type from endpoint_new where monitor_type=? and guid in (select endpoint from endpoint_service_rel where service_group=?)", monitorType, serviceGroup).Find(&result)
	return result
}

func getServiceGroupEndpointWithChild(serviceGroup string) map[string][]string {
	serviceGroupList := []string{serviceGroup}
	fetchServiceGroupList, err := fetchGlobalServiceGroupChildGuidList(serviceGroup)
	if err == nil {
		serviceGroupList = fetchServiceGroupList
	}
	var endpointServiceRel []*models.EndpointServiceRelTable
	x.SQL("select * from endpoint_service_rel where service_group in ('" + strings.Join(serviceGroupList, "','") + "')").Find(&endpointServiceRel)
	endpointExistMap := make(map[string]int)
	endpointTypeMap := make(map[string][]string)
	for _, v := range endpointServiceRel {
		if _, b := endpointExistMap[v.Endpoint]; b {
			continue
		}
		endpointExistMap[v.Endpoint] = 1
		if !strings.Contains(v.Endpoint, "_") {
			continue
		}
		tmpEndpointType := v.Endpoint[strings.LastIndex(v.Endpoint, "_")+1:]
		if _, b := endpointTypeMap[tmpEndpointType]; b {
			endpointTypeMap[tmpEndpointType] = append(endpointTypeMap[tmpEndpointType], v.Endpoint)
		} else {
			endpointTypeMap[tmpEndpointType] = []string{v.Endpoint}
		}
	}
	return endpointTypeMap
}

func UpdateServiceConfigWithEndpoint(serviceGroup string) {
	var err error
	endpointTypeMap := getServiceGroupEndpointWithChild(serviceGroup)
	log.Logger.Info("UpdateServiceConfigWithEndpoint", log.String("serviceGroup", serviceGroup))
	err = UpdateLogMetricConfigByServiceGroup(serviceGroup, endpointTypeMap)
	if err != nil {
		log.Logger.Error("UpdateLogMetricConfigByServiceGroup fail", log.Error(err))
	}
	err = UpdateDbMetricConfigByServiceGroup(serviceGroup, endpointTypeMap)
	if err != nil {
		log.Logger.Error("UpdateDbMetricConfigByServiceGroup fail", log.Error(err))
	}
	err = UpdateLogKeywordConfigByServiceGroup(serviceGroup, endpointTypeMap)
	if err != nil {
		log.Logger.Error("UpdateLogKeywordConfigByServiceGroup fail", log.Error(err))
	}
}

func UpdateLogKeywordConfigByServiceGroup(serviceGroup string, endpointTypeMap map[string][]string) (err error) {
	var logKeywordTable []*models.LogKeywordMonitorTable
	x.SQL("select * from log_keyword_monitor where service_group=?", serviceGroup).Find(&logKeywordTable)
	if len(logKeywordTable) == 0 {
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
	for _, v := range logKeywordTable {
		UpdateLogKeywordConfigAction(v, endpointTypeMap, hostEndpoint, hostEndpointIpMap)
	}
	return
}

func UpdateLogKeywordConfigAction(logKeyword *models.LogKeywordMonitorTable, endpointTypeMap map[string][]string, hostEndpoint []string, hostEndpointIpMap map[string]string) {
	log.Logger.Info("UpdateLogKeywordConfigAction", log.String("guid", logKeyword.Guid), log.String("monitorType", logKeyword.MonitorType), log.StringList("hostEndpoint", hostEndpoint))
	var updateHostEndpointList []string
	var actions []*Action
	var logKeywordRelTable []*models.LogMetricEndpointRelTable
	x.SQL("select * from log_keyword_endpoint_rel where log_keyword_monitor=?", logKeyword.Guid).Find(&logKeywordRelTable)
	if len(logKeywordRelTable) == 0 && len(hostEndpoint) == 0 {
		return
	}
	targetTypeMap := make(map[string]int)
	if targetTypeList, b := endpointTypeMap[logKeyword.MonitorType]; b {
		for _, target := range targetTypeList {
			targetTypeMap[target] = 1
		}
	}
	sourceTargetMap := make(map[string]string)
	for _, vv := range logKeywordRelTable {
		sourceTargetMap[vv.SourceEndpoint] = vv.TargetEndpoint
	}
	for _, host := range hostEndpoint {
		if target, b := sourceTargetMap[host]; b {
			// target remove
			if _, bb := targetTypeMap[target]; !bb {
				actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=? and source_endpoint=?", Param: []interface{}{logKeyword.Guid, host}})
				updateHostEndpointList = append(updateHostEndpointList, host)
			}
		} else {
			for target, _ := range targetTypeMap {
				// match new target
				if strings.Contains(target, fmt.Sprintf("_%s_", hostEndpointIpMap[host])) {
					actions = append(actions, &Action{Sql: "insert into log_keyword_endpoint_rel(guid,log_keyword_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guid.CreateGuid(), logKeyword.Guid, host, target}})
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
			actions = append(actions, &Action{Sql: "delete from log_keyword_endpoint_rel where log_keyword_monitor=? and source_endpoint=?", Param: []interface{}{logKeyword.Guid, source}})
			updateHostEndpointList = append(updateHostEndpointList, source)
		}
	}
	if len(actions) > 0 {
		err := Transaction(actions)
		if err == nil {
			err = SyncLogKeywordExporterConfig(updateHostEndpointList)
			if err != nil {
				log.Logger.Error("SyncLogKeywordExporterConfig fail", log.String("logKeywordMonitor", logKeyword.Guid), log.Error(err))
			}
		} else {
			log.Logger.Error("UpdateLogKeywordConfigAction exec sql fail", log.String("logKeywordMonitor", logKeyword.Guid), log.Error(err))
		}
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
	log.Logger.Info("UpdateLogMetricConfigAction", log.String("guid", logMonitor.Guid), log.String("monitorType", logMonitor.MonitorType), log.StringList("hostEndpoint", hostEndpoint))
	for k, v := range endpointTypeMap {
		log.Logger.Debug("endpointTypeMap", log.String("k", k), log.StringList("v", v))
	}
	for k, v := range hostEndpointIpMap {
		log.Logger.Debug("hostEndpointIpMap", log.String("k", k), log.String("v", v))
	}
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
		log.Logger.Info("sourceTargetMap", log.String("source", vv.SourceEndpoint), log.String("target", vv.TargetEndpoint))
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
			err = SyncLogMetricExporterConfig(updateHostEndpointList)
			if err != nil {
				log.Logger.Error("SyncLogMetricExporterConfig fail", log.String("logMetricMonitor", logMonitor.Guid), log.Error(err))
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

func DeleteServiceWithChildConfig(serviceGroup string) {
	var parentGuidList []string
	guidList, _ := fetchGlobalServiceGroupParentGuidList(serviceGroup)
	for _, v := range guidList {
		if v == serviceGroup {
			continue
		}
		parentGuidList = append(parentGuidList, v)
		UpdateServiceConfigWithEndpoint(v)
	}
	if len(parentGuidList) > 0 {
		var endpointGroup []*models.EndpointGroupTable
		x.SQL("select guid from endpoint_group where service_group in ('" + strings.Join(parentGuidList, "','") + "')").Find(&endpointGroup)
		for _, v := range endpointGroup {
			tmpErr := SyncPrometheusRuleFile(v.Guid, false)
			if tmpErr != nil {
				log.Logger.Error("DeleteServiceWithChildConfig SyncPrometheusRuleFile fail", log.Error(tmpErr))
			}
		}
	}
	guidList, _ = fetchGlobalServiceGroupChildGuidList(serviceGroup)
	for _, v := range guidList {
		DeleteServiceConfig(v)
	}
}

func DeleteServiceConfig(serviceGroup string) {
	// Remove logMetric config
	var logMetricTable []*models.LogMetricMonitorTable
	x.SQL("select guid from log_metric_monitor where service_group=?", serviceGroup).Find(&logMetricTable)
	for _, v := range logMetricTable {
		tmpErr := DeleteLogMetricMonitor(v.Guid)
		if tmpErr != nil {
			log.Logger.Error("Try to DeleteLogMetricMonitor fail", log.Error(tmpErr))
		}
	}
	// Remove dbMetric config
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
	// Remove logKeyword config
	var logKeywordTable []*models.LogKeywordMonitorTable
	x.SQL("select guid from log_keyword_monitor where service_group=?", serviceGroup).Find(&logKeywordTable)
	for _, v := range logKeywordTable {
		tmpErr := DeleteLogKeywordMonitor(v.Guid)
		if tmpErr != nil {
			log.Logger.Error("Try to DeleteLogKeywordMonitor fail", log.Error(tmpErr))
		}
	}
}

func getUpdateServiceGroupNotifyActions(serviceGroup, firingCallback, recoverCallback string, roleList []string) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify in (select guid from notify where service_group=?)", Param: []interface{}{serviceGroup}})
	actions = append(actions, &Action{Sql: "delete from notify where service_group=?", Param: []interface{}{serviceGroup}})
	if firingCallback != "" {
		firingActionGuid := guid.CreateGuid()
		actions = append(actions, &Action{Sql: "insert into notify(guid,service_group,alarm_action,proc_callback_key) value (?,?,?,?)", Param: []interface{}{firingActionGuid, serviceGroup, "firing", firingCallback}})
	}
	if recoverCallback != "" {
		recoverActionGuid := guid.CreateGuid()
		actions = append(actions, &Action{Sql: "insert into notify(guid,service_group,alarm_action,proc_callback_key) value (?,?,?,?)", Param: []interface{}{recoverActionGuid, serviceGroup, "ok", recoverCallback}})
	}
	actions = append(actions, &Action{Sql: "delete from service_group_role_rel where service_group=?", Param: []interface{}{serviceGroup}})
	for _, role := range roleList {
		if role == "" {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into service_group_role_rel(guid,service_group,`role`) value (?,?,?)", Param: []interface{}{guid.CreateGuid(), serviceGroup, role}})
	}
	return actions
}

func getUpdateServiceGroupNotifyRoles(serviceGroup string, roleList []string) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from service_group_role_rel where service_group=?", Param: []interface{}{serviceGroup}})
	tmpGuidList := guid.CreateGuidList(len(roleList))
	for i, role := range roleList {
		actions = append(actions, &Action{Sql: "insert into service_group_role_rel(guid,service_group,`role`) value (?,?,?)", Param: []interface{}{tmpGuidList[i], serviceGroup, role}})
	}
	var notifyTable []*models.NotifyTable
	x.SQL("select guid from notify where service_group=?", serviceGroup).Find(&notifyTable)
	if len(notifyTable) == 0 {
		actions = append(actions, &Action{Sql: "insert into notify(guid,service_group,alarm_action) value (?,?,?)", Param: []interface{}{guid.CreateGuid(), serviceGroup, "firing"}})
		actions = append(actions, &Action{Sql: "insert into notify(guid,service_group,alarm_action) value (?,?,?)", Param: []interface{}{guid.CreateGuid(), serviceGroup, "ok"}})
	}
	return actions
}

func CheckMetricIsServiceMetric(metric, serviceGroup string) bool {
	serviceGroupGuidList, _ := fetchGlobalServiceGroupChildGuidList(serviceGroup)
	queryRows, _ := x.QueryString("select guid from metric where metric=? and service_group in ('"+strings.Join(serviceGroupGuidList, "','")+"')", metric)
	if len(queryRows) > 0 {
		return true
	}
	return false
}
