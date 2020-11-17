package db

import (
	"time"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

func StartNotifyPingExport()  {
	t := time.NewTicker(time.Second*60).C
	for {
		<- t
		go notifyPingExport()
	}
}

func notifyPingExport()  {
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

	if len(telnetGuidList) > 0 {

	}
	if len(httpGuidList) > 0 {

	}

}
