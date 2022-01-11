package funcs

import (
	"fmt"
	"log"
	"strings"
)

var MonitorObjList []*MonitorArchiveObj

func InitMonitorMetricMap() error  {
	var monitorEndpointTableData []*MonitorEndpointTable
	err := monitorMysqlEngine.SQL("select guid,step,monitor_type as 'export_type',agent_address as 'address' from endpoint_new").Find(&monitorEndpointTableData)
	if err != nil {
		return fmt.Errorf("get monitor endpoint table data fail,%v \n", err)
	}
	var monitorMetricTableData []*MonitorPromMetricTable
	err = monitorMysqlEngine.SQL("select metric,prom_expr as 'prom_ql',monitor_type as 'metric_type' from metric where service_group is null").Find(&monitorMetricTableData)
	if err != nil {
		return fmt.Errorf("get monitor prom_metric table data fail,%v \n", err)
	}
	monitorMetricMap := make(map[string][]*MonitorPromMetricTable)
	for _,v := range monitorMetricTableData {
		if _,b := monitorMetricMap[v.MetricType];b {
			monitorMetricMap[v.MetricType] = append(monitorMetricMap[v.MetricType], v)
		}else{
			monitorMetricMap[v.MetricType] = []*MonitorPromMetricTable{v}
		}
	}
	extEndpointMetricMap,serviceMetricList := getMonitorServiceMetricMap()
	MonitorObjList = []*MonitorArchiveObj{}
	for _,v := range monitorEndpointTableData {
		var tmpMonitorMetricTable []*MonitorPromMetricTable
		for _,vv := range monitorMetricMap[v.ExportType] {
			tmpPromQl := vv.PromQl
			if strings.Contains(tmpPromQl, "$address") {
				tmpPromQl = strings.Replace(tmpPromQl, "$address", v.Address, -1)
			}
			if strings.Contains(tmpPromQl, "$guid") {
				tmpPromQl = strings.Replace(tmpPromQl, "$guid", v.Guid, -1)
			}
			tmpMonitorMetricTable = append(tmpMonitorMetricTable, &MonitorPromMetricTable{Metric:vv.Metric, PromQl:tmpPromQl})
		}
		if extMetricList,b:=extEndpointMetricMap[v.Guid];b {
			tmpMonitorMetricTable = append(tmpMonitorMetricTable, extMetricList...)
		}
		MonitorObjList = append(MonitorObjList, &MonitorArchiveObj{Endpoint:v.Guid, Metrics:tmpMonitorMetricTable})
	}
	if len(serviceMetricList) > 0 {
		MonitorObjList = append(MonitorObjList, serviceMetricList...)
	}
	return nil
}

func getMonitorServiceMetricMap() (endpointMetricMap map[string][]*MonitorPromMetricTable,serviceMetricList []*MonitorArchiveObj) {
	endpointMetricMap = make(map[string][]*MonitorPromMetricTable)
	var serviceMetricTable []*MonitorMetricTable
	err := monitorMysqlEngine.SQL("select guid,metric,monitor_type,prom_expr,service_group from metric where service_group is not null").Find(&serviceMetricTable)
	if err != nil {
		log.Printf("get service metric table fail,%s \n", err.Error())
		return
	}
	if len(serviceMetricTable) == 0 {
		return
	}
	serviceMetricMap := make(map[string][]*MonitorPromMetricTable)
	serviceEndpointTypeMap := make(map[string][]*MonitorEndpointTable)
	for _,v := range serviceMetricTable {
		if strings.Contains(v.PromExpr, "$guid") || strings.Contains(v.PromExpr, "$address") {
			tmpStKey := v.ServiceGroup+"__"+v.MonitorType
			if _,b:=serviceEndpointTypeMap[tmpStKey];!b {
				tmpEndpointTable := []*MonitorEndpointTable{}
				tmpErr := monitorMysqlEngine.SQL("select guid,step,monitor_type as 'export_type',agent_address as 'address' from endpoint_new where monitor_type=? and guid in (select endpoint from endpoint_service_rel where service_group=?)",v.MonitorType,v.ServiceGroup).Find(&tmpEndpointTable)
				if tmpErr != nil {
					log.Printf("Query service:%s type:%s endpoint fail,%s \n", v.ServiceGroup, v.MonitorType, tmpErr.Error())
					continue
				}
				serviceEndpointTypeMap[tmpStKey] = tmpEndpointTable
			}
			if len(serviceEndpointTypeMap[tmpStKey]) > 0 {
				for _,tmpEndpoint := range serviceEndpointTypeMap[tmpStKey] {
					tmpMetricObj := MonitorPromMetricTable{Metric: v.Metric,MetricType: v.MonitorType,PromQl: v.PromExpr}
					if strings.Contains(tmpMetricObj.PromQl, "$address") {
						tmpMetricObj.PromQl = strings.Replace(tmpMetricObj.PromQl, "$address", tmpEndpoint.Address, -1)
					}
					if strings.Contains(tmpMetricObj.PromQl, "$guid") {
						tmpMetricObj.PromQl = strings.Replace(tmpMetricObj.PromQl, "$guid", tmpEndpoint.Guid, -1)
					}
					if _,eExist:=endpointMetricMap[tmpEndpoint.Guid];eExist {
						endpointMetricMap[tmpEndpoint.Guid] = append(endpointMetricMap[tmpEndpoint.Guid], &tmpMetricObj)
					}else{
						endpointMetricMap[tmpEndpoint.Guid] = []*MonitorPromMetricTable{&tmpMetricObj}
					}
				}
			}
			continue
		}
		tmpServiceMetricObj := MonitorPromMetricTable{Metric: v.Metric,MetricType: v.MonitorType,PromQl: v.PromExpr}
		if _,b:=serviceMetricMap[v.ServiceGroup];b {
			serviceMetricMap[v.ServiceGroup] = append(serviceMetricMap[v.ServiceGroup], &tmpServiceMetricObj)
		}else{
			serviceMetricMap[v.ServiceGroup] = []*MonitorPromMetricTable{&tmpServiceMetricObj}
		}
	}

	return
}