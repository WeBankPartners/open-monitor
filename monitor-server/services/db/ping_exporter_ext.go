package db

import (
	"time"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"strings"
	"fmt"
	"net/http"
	"encoding/json"
)

func StartNotifyPingExport()  {
	t := time.NewTicker(time.Second*60).C
	for {
		<- t
		go notifyPingExport()
	}
}

func notifyPingExport()  {
	log.Logger.Debug("start to notify ping exporter")
	var endpointTable []*m.EndpointTable
	err := x.SQL("select * from endpoint where export_type in ('ping','telnet','http') and address_agent<>'' order by address_agent").Find(&endpointTable)
	if err != nil {
		log.Logger.Error("Notify ping export fail,query endpoint table fail", log.Error(err))
		return
	}
	if len(endpointTable) == 0 {
		log.Logger.Warn("Notify ping export done with empty data")
		return
	}
	var telnetGuidList,httpGuidList []string
	for _,v := range endpointTable {
		if v.ExportType == "telnet" {
			telnetGuidList = append(telnetGuidList, v.Guid)
		}
		if v.ExportType == "http" {
			httpGuidList = append(httpGuidList, v.Guid)
		}
	}
	var telnetTables []*m.EndpointTelnetTable
	var httpTables []*m.EndpointHttpTable
	if len(telnetGuidList) > 0 {
		x.SQL("select * from endpoint_telnet where endpoint_guid in ('"+strings.Join(telnetGuidList, "','")+"')").Find(&telnetTables)
	}
	if len(httpGuidList) > 0 {
		x.SQL("select * from endpoint_http where endpoint_guid in ('"+strings.Join(telnetGuidList, "','")+"')").Find(&httpTables)
	}
	var extendExporterMap = make(map[string][]*m.PingExportSourceObj)
	for _,v := range endpointTable {
		if !strings.Contains(v.AddressAgent, ":") {
			continue
		}
		var tmpPingExporterSourceObj m.PingExportSourceObj
		if v.ExportType == "ping" {
			tmpPingExporterSourceObj.Ip = v.Ip
			tmpPingExporterSourceObj.Guid = v.Guid
		}else if v.ExportType == "telnet" {
			for _,vv := range telnetTables {
				if vv.EndpointGuid == v.Guid {
					tmpPingExporterSourceObj.Ip = fmt.Sprintf("%s:%s", v.Ip, vv.Port)
					tmpPingExporterSourceObj.Guid = v.Guid
					break
				}
			}
		}else if v.ExportType == "http" {
			for _,vv := range httpTables {
				if vv.EndpointGuid == v.Guid {
					tmpPingExporterSourceObj.Ip = fmt.Sprintf("%s_%s",strings.ToUpper(vv.Method),vv.Url)
					tmpPingExporterSourceObj.Guid = v.Guid
					break
				}
			}
		}
		if _,b := extendExporterMap[v.AddressAgent];b {
			extendExporterMap[v.AddressAgent] = append(extendExporterMap[v.AddressAgent], &tmpPingExporterSourceObj)
		}else{
			extendExporterMap[v.AddressAgent] = []*m.PingExportSourceObj{&tmpPingExporterSourceObj}
		}
	}
	for k,v := range extendExporterMap {
		requestPingExporter(k, v)
	}

}

func requestPingExporter(address string,objList []*m.PingExportSourceObj)  {
	if address == "" || len(objList) == 0 {
		return
	}
	url := fmt.Sprintf("http://%s/config/ip", address)
	var param m.PingExporterSourceDto
	param.Config = objList
	paramBytes,_ := json.Marshal(param)
	log.Logger.Debug("request ping exporter", log.String("address", address), log.String("body", string(paramBytes)))
	resp,err := http.Post(url, "application/json", strings.NewReader(string(paramBytes)))
	if err != nil {
		log.Logger.Error("Request ping exporter fail", log.Error(err))
		return
	}
	if resp.StatusCode >= 300 {
		log.Logger.Error("Request ping exporter fail,status code error", log.Int("status", resp.StatusCode))
	}else{
		log.Logger.Info("Request ping exporter success", log.String("address", address))
	}
}