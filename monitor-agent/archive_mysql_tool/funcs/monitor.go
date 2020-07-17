package funcs

import (
	"fmt"
	"strings"
)

var MonitorObjList []*MonitorArchiveObj

func InitMonitorMetricMap() error  {
	var monitorEndpointTableData []*MonitorEndpointTable
	err := monitorMysqlEngine.SQL("SELECT guid,export_type,address,address_agent FROM endpoint").Find(&monitorEndpointTableData)
	if err != nil {
		return fmt.Errorf("get monitor endpoint table data fail,%v \n", err)
	}
	var monitorMetricTableData []*MonitorPromMetricTable
	err = monitorMysqlEngine.SQL("SELECT metric,prom_ql,metric_type FROM prom_metric").Find(&monitorMetricTableData)
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
	MonitorObjList = []*MonitorArchiveObj{}
	for _,v := range monitorEndpointTableData {
		var tmpMonitorMetricTable []*MonitorPromMetricTable
		for _,vv := range monitorMetricMap[v.ExportType] {
			tmpPromQl := vv.PromQl
			if strings.Contains(tmpPromQl, "$address") {
				if v.AddressAgent != "" {
					tmpPromQl = strings.Replace(tmpPromQl, "$address", v.AddressAgent, -1)
				}else{
					tmpPromQl = strings.Replace(tmpPromQl, "$address", v.Address, -1)
				}
			}
			if strings.Contains(tmpPromQl, "$guid") {
				tmpPromQl = strings.Replace(tmpPromQl, "$guid", v.Guid, -1)
			}
			tmpMonitorMetricTable = append(tmpMonitorMetricTable, &MonitorPromMetricTable{Metric:vv.Metric, PromQl:tmpPromQl})
		}
		MonitorObjList = append(MonitorObjList, &MonitorArchiveObj{Endpoint:v.Guid, Metrics:tmpMonitorMetricTable})
	}
	return nil
}