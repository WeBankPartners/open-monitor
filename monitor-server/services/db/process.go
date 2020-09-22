package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
	"time"
	"encoding/json"
	"net/http"
	"strings"
	"io/ioutil"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

type processHttpDto struct {
	Process  []string  `json:"process"`
	Force    int       `json:"force"`
}

func UpdateNodeExporterProcessConfig(endpointId int) error {
	err,data := GetProcessList(endpointId)
	if err != nil {
		log.Logger.Error("Update node_exporter fail", log.Error(err))
		return err
	}
	endpointObj := m.EndpointTable{Id:endpointId}
	err = GetEndpoint(&endpointObj)
	if err != nil {
		log.Logger.Error("Update node_exporter fail, get endpoint msg fail", log.Error(err))
		return err
	}
	postParam := processHttpDto{Process:[]string{}, Force:0}
	for _,v := range data {
		postParam.Process = append(postParam.Process, v.Name)
	}
	postData,err := json.Marshal(postParam)
	if err != nil {
		log.Logger.Error("Update node_exporter fail, marshal post data fail", log.Error(err))
		return err
	}
	url := fmt.Sprintf("http://%s/process/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Update node_exporter fail, http post fail", log.Error(err))
		return err
	}
	responseBody,_ := ioutil.ReadAll(resp.Body)
	log.Logger.Info("curl "+url, log.String("response", string(responseBody)))
	resp.Body.Close()
	return nil
}

func CheckNodeExporterProcessConfig(endpointId int,processList []string) (err error,illegal bool,msg string) {
	endpointObj := m.EndpointTable{Id:endpointId}
	err = GetEndpoint(&endpointObj)
	if err != nil {
		log.Logger.Error("Check node_exporter fail, get endpoint msg fail", log.Error(err))
		return
	}
	postParam := processHttpDto{Process:processList, Force:1}
	postData,err := json.Marshal(postParam)
	if err != nil {
		log.Logger.Error("Check node_exporter fail, marshal post data fail", log.Error(err))
		return
	}
	url := fmt.Sprintf("http://%s/process/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Check node_exporter fail, http post fail", log.Error(err))
		return
	}
	responseBody,_ := ioutil.ReadAll(resp.Body)
	log.Logger.Info("curl "+url, log.String("response", string(responseBody)))
	resp.Body.Close()
	msg = string(responseBody)
	if resp.StatusCode > 300 {
		illegal = true
	}else{
		illegal = false
	}
	return
}

func GetProcessList(endpointId int) (err error, processList []*m.ProcessMonitorTable) {
	err = x.SQL("SELECT * FROM process_monitor WHERE endpoint_id=?", endpointId).Find(&processList)
	return err,processList
}

func UpdateProcess(param m.ProcessUpdateDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql:"DELETE FROM process_monitor WHERE endpoint_id=?", Param:[]interface{}{param.EndpointId}})
	for _,v := range param.ProcessList {
		var action Action
		params := make([]interface{}, 0)
		action.Sql = "INSERT INTO process_monitor(endpoint_id,NAME) VALUE (?,?)"
		params = append(params, param.EndpointId)
		params = append(params, v)
		action.Param = params
		actions = append(actions, &action)
	}
	return Transaction(actions)
}

func UpdateAliveCheckQueue(monitorIp string) error {
	_,err := x.Exec(fmt.Sprintf("INSERT INTO alive_check_queue(message) VALUE ('%s')", monitorIp))
	return err
}

func GetAliveCheckQueue(param string) (err error,result []*m.AliveCheckQueueTable) {
	lastMinDateString := time.Unix(time.Now().Unix()-60, 0).Format("2006-01-02 15:04:05")
	err = x.SQL(fmt.Sprintf("SELECT * FROM alive_check_queue WHERE message='%s' AND update_at>'%s' LIMIT 1", param, lastMinDateString)).Find(&result)
	if err != nil {
		return err,result
	}
	if len(result) == 0 {
		err = fmt.Errorf("get alive_check_queue table fail,nodata with message=%s and update_at>%s", param, lastMinDateString)
	}
	return err,result
}