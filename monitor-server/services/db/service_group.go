package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
)

var (
	globalServiceGroupMap  = make(map[string]*models.ServiceGroupLinkNode)
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
			globalServiceGroupMap[v.Guid].Parent = globalServiceGroupMap[v.Parent]
			globalServiceGroupMap[v.Parent].Children = append(globalServiceGroupMap[v.Parent].Children, globalServiceGroupMap[v.Guid])
		}
	}
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
	if _, b := globalServiceGroupMap[param.Guid]; !b {
		globalServiceGroupMap[param.Guid] = &models.ServiceGroupLinkNode{Guid: param.Guid}
		if param.Parent != "" {
			if _,bb := globalServiceGroupMap[param.Parent]; bb {
				globalServiceGroupMap[param.Guid] = &models.ServiceGroupLinkNode{Guid: param.Guid, Parent: globalServiceGroupMap[param.Parent]}
				globalServiceGroupMap[param.Parent].Children = append(globalServiceGroupMap[param.Parent].Children, globalServiceGroupMap[param.Guid])
			}
		}
	}
}

func deleteGlobalServiceGroupNode(guid string) {
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

func CreateServiceGroup(param models.ServiceGroupTable) {
	if param.Parent != "" {

	}
}

func UpdateServiceGroup() {

}

func DeleteServiceGroup() {

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

func AppendServiceConfigWithEndpoint(serviceGroup,newEndpoint string,endpointList []string)  {
	endpointObj,_ := GetEndpointNew(&models.EndpointNewTable{Guid: newEndpoint})
	if endpointObj.MonitorType == "" {
		return
	}
	var logMetricTable []*models.LogMetricMonitorTable
	x.SQL("select * from log_metric_monitor where service_group=? and monitor_type=?", serviceGroup, endpointObj.MonitorType).Find(&logMetricTable)
	for _,v := range logMetricTable {
		var logMetricRelTable []*models.LogMetricEndpointRelTable
		x.SQL("select * from log_metric_endpoint_rel where log_metric_monitor=?", v.Guid).Find(&logMetricRelTable)
		existFlag := false
		for _,vv := range logMetricRelTable {
			if vv.TargetEndpoint == endpointObj.Guid {
				existFlag = true
			}
		}
		if existFlag {
			continue
		}
		sourceEndpoint := ""
		for _,vv := range endpointList {
			if strings.Contains(vv, fmt.Sprintf("_%s_", endpointObj.Ip)) {
				sourceEndpoint = vv
			}
		}
		_,tmpErr := x.Exec("insert into log_metric_endpoint_rel(guid,log_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", guid.CreateGuid(), v.Guid, sourceEndpoint, endpointObj.Guid)
		if tmpErr == nil {
			tmpErr = UpdateNodeExportConfig([]string{sourceEndpoint})
			if tmpErr != nil {
				log.Logger.Error("AppendServiceConfigWithEndpoint log metric fail", log.String("newEndpoint", newEndpoint), log.String("sourceEndpoint", sourceEndpoint), log.String("log_metric_monitor",v.Guid))
			}
		}
	}
	if endpointObj.MonitorType != "mysql" {
		return
	}
	var dbMetricTable []*models.DbMetricMonitorTable
	x.SQL("select * from db_metric_monitor where service_group=? and monitor_type=?", serviceGroup, endpointObj.MonitorType).Find(&dbMetricTable)
	for _,v := range dbMetricTable {
		var dbMetricRelTable []*models.DbMetricEndpointRelTable
		x.SQL("select * from db_metric_endpoint_rel where db_metric_monitor=?", v.Guid).Find(&dbMetricRelTable)
		existFlag := false
		for _,vv := range dbMetricRelTable {
			if vv.SourceEndpoint == endpointObj.Guid {
				existFlag = true
			}
		}
		if existFlag {
			continue
		}
		_,tmpErr := x.Exec("insert into db_metric_endpoint_rel(guid,db_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", guid.CreateGuid(), v.Guid, endpointObj.Guid, endpointObj.Guid)
		if tmpErr == nil {
			tmpErr = SyncDbMetric()
			if tmpErr != nil {
				log.Logger.Error("AppendServiceConfigWithEndpoint db metric fail", log.String("newEndpoint", newEndpoint), log.String("log_metric_monitor",v.Guid))
			}
		}
	}
}