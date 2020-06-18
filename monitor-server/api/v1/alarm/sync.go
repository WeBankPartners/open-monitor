package alarm

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
)

func SyncInitConfigFile()  {
	tplTable := db.ListTpl()
	for _,v := range tplTable {
		err := SaveConfigFile(v.Id, true)
		if err != nil {
			mid.LogError(fmt.Sprintf("Sync init config fail fail with tpl id:%d", v.Id), err)
		}
	}
	mid.LogInfo("Sync init config file done")
}

func SyncInitSdFile()  {
	if !m.Config().SdFile.Enable {
		return
	}
	var data []*m.ServiceDiscoverFileObj
	var stepList []int
	endpointTable := db.ListEndpoint()
	for _,v := range endpointTable {
		tmpAddress := v.Address
		if v.AddressAgent != "" {
			tmpAddress = v.AddressAgent
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
			mid.LogError(fmt.Sprintf("Sync service discover file fail,step: %d ", v), err)
		}
	}
}