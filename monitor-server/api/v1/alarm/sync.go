package alarm

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"strings"
	"golang.org/x/net/context/ctxhttp"
	"context"
	"io/ioutil"
	"encoding/json"
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

func SyncInitConsul()  {
	var consulUrl string
	for _,v := range m.Config().Dependence {
		if v.Name == "consul" {
			consulUrl = v.Server
			break
		}
	}
	if consulUrl == "" {
		mid.LogInfo("Sync init consul endpoint fail, consul url is empty")
		return
	}
	req,_ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/internal/ui/services", consulUrl), strings.NewReader(""))
	resp,err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		mid.LogError("Sync init consul endpoint fail, request consul error", err)
		return
	}
	var responseData []*m.ConsulServicesDto
	b,_ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, &responseData)
	if err != nil {
		mid.LogError("Sync init consul endpoint fail, response json format error", err)
		return
	}
	resp.Body.Close()
	endpointTable := db.ListEndpoint()
	var insertData,deleteData []*m.EndpointTable
	for _,v := range endpointTable {
		if v.ExportType == "custom" {
			continue
		}
		existFlag := false
		for _,vv := range responseData {
			if vv.Name == v.Guid {
				existFlag = true
				break
			}
		}
		if !existFlag {
			insertData = append(insertData, v)
		}
	}
	for _,v := range responseData {
		existFlag := false
		for _,vv := range endpointTable {
			if vv.Guid == v.Name {
				existFlag = true
				break
			}
		}
		if !existFlag {
			deleteData = append(deleteData, &m.EndpointTable{Guid:v.Name})
		}
	}
	for _,v := range insertData {
		tmpAddress := v.Address
		if v.AddressAgent != "" {
			tmpAddress = v.AddressAgent
		}
		tmpSplit := strings.Split(tmpAddress, ":")
		if len(tmpSplit) != 2 {
			continue
		}
		err := prom.RegisteConsul(v.Guid, tmpSplit[0], tmpSplit[1], []string{v.ExportType}, 10, true)
		if err != nil {
			mid.LogError(fmt.Sprintf("init register guid:%s ip:%s port:%s error", v.Guid, tmpSplit[0], tmpSplit[1]), err)
		}
	}
	for _,v := range deleteData {
		err := prom.DeregisteConsul(v.Guid, true)
		if err != nil {
			mid.LogError(fmt.Sprintf("init deregister guid:%s error", v.Guid), err)
		}
	}
	mid.LogInfo("Sync init consul endpoint done")
}
