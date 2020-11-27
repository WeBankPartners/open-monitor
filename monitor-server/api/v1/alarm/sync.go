package alarm

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

func SyncInitConfigFile()  {
	tplTable := db.ListTpl()
	for _,v := range tplTable {
		err := SaveConfigFile(v.Id, true)
		if err != nil {
			log.Logger.Error("Sync init config fail fail", log.Int("tplId", v.Id), log.Error(err))
		}
	}
	log.Logger.Info("Sync init config file done")
}

func SyncInitSdFile()  {
	if !m.Config().SdFile.Enable {
		return
	}
	var data []*m.ServiceDiscoverFileObj
	var stepList []int
	endpointTable := db.ListEndpoint()
	for _,v := range endpointTable {
		if v.ExportType == "ping" || v.ExportType == "telnet" || v.ExportType == "http" {
			if v.AddressAgent == "" {
				continue
			}
		}
		tmpAddress := v.Address
		if v.AddressAgent != "" {
			tmpAddress = v.AddressAgent
		}
		if tmpAddress == "" {
			continue
		}
		data = append(data, &m.ServiceDiscoverFileObj{Guid:v.Guid, Address:tmpAddress, Step:v.Step})
		exist := false
		for _,vv := range stepList {
			if v.Step == vv {
				exist = true
				break
			}
		}
		if !exist {
			stepList = append(stepList, v.Step)
		}
	}
	prom.InitSdConfig(data)
	for _,v := range stepList {
		err := prom.SyncSdConfigFile(v)
		if err != nil {
			log.Logger.Error("Sync service discover file fail", log.Int("step", v), log.Error(err))
		}
	}
}