package db

import (
	"encoding/json"
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

type processHttpDto struct {
	Process []string `json:"process"`
	Check   int      `json:"check"`
}

func UpdateNodeExporterProcessConfig(endpointId int) error {
	err, data := GetProcessList(endpointId)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail", zap.Error(err))
		return err
	}
	endpointObj := m.EndpointTable{Id: endpointId}
	err = GetEndpoint(&endpointObj)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, get endpoint msg fail", zap.Error(err))
		return err
	}
	postParam := processHttpDto{Process: []string{}, Check: 0}
	for _, v := range data {
		postParam.Process = append(postParam.Process, fmt.Sprintf("%s^%s", v.ProcessName, v.Tags))
	}
	postData, err := json.Marshal(postParam)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, marshal post data fail", zap.Error(err))
		return err
	}
	url := fmt.Sprintf("http://%s/process/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, http post fail", zap.Error(err))
		return err
	}
	responseBody, _ := ioutil.ReadAll(resp.Body)
	log.Info(nil, log.LOGGER_APP, "curl "+url, zap.String("response", string(responseBody)))
	resp.Body.Close()
	return nil
}

func SyncNodeExporterProcessConfig(hostIp string, newEndpoints []*m.EndpointNewTable, updateFlag bool) (err error) {
	var endpointTable []*m.EndpointNewTable
	//if updateFlag {
	//	updateGuidList := []string{}
	//	for _,v := range newEndpoints {
	//		updateGuidList = append(updateGuidList, v.Guid)
	//	}
	//	err = x.SQL("select * from endpoint_new where monitor_type='process' and ip=? and guid not in ('"+strings.Join(updateGuidList,"','")+"')", hostIp).Find(&endpointTable)
	//}else {
	//	err = x.SQL("select * from endpoint_new where monitor_type='process' and ip=?", hostIp).Find(&endpointTable)
	//}
	updateGuidList := []string{}
	for _, v := range newEndpoints {
		updateGuidList = append(updateGuidList, v.Guid)
	}
	err = x.SQL("select * from endpoint_new where monitor_type='process' and ip=? and guid not in ('"+strings.Join(updateGuidList, "','")+"')", hostIp).Find(&endpointTable)
	if err != nil {
		return fmt.Errorf("Query table endpoint_new fail,%s ", err.Error())
	}
	if len(newEndpoints) > 0 {
		endpointTable = append(endpointTable, newEndpoints...)
	}
	var nodeExportAddress string
	if len(endpointTable) > 0 {
		nodeExportAddress = endpointTable[0].AgentAddress
	} else {
		endpointObj := m.EndpointTable{Ip: hostIp, ExportType: "host"}
		GetEndpoint(&endpointObj)
		nodeExportAddress = endpointObj.Address
	}
	syncParam := m.SyncProcessDto{Check: 0, Process: []*m.SyncProcessObj{}}
	for _, v := range endpointTable {
		if v.ExtendParam == "" {
			continue
		}
		tmpExtendObj := m.EndpointExtendParamObj{}
		tmpErr := json.Unmarshal([]byte(v.ExtendParam), &tmpExtendObj)
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "Sync process config,extendParam illegal", zap.String("processEndpoint", v.Guid), zap.String("extendParam", v.ExtendParam), zap.Error(tmpErr))
			continue
		}
		syncParam.Process = append(syncParam.Process, &m.SyncProcessObj{ProcessGuid: v.Guid, ProcessName: tmpExtendObj.ProcessName, ProcessTags: tmpExtendObj.ProcessTags})
	}
	postData, _ := json.Marshal(syncParam)
	log.Info(nil, log.LOGGER_APP, "sync new process config", zap.String("postData", string(postData)))
	url := fmt.Sprintf("http://%s/process/config", nodeExportAddress)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, http post fail", zap.Error(err))
		return err
	}
	responseBody, _ := ioutil.ReadAll(resp.Body)
	log.Info(nil, log.LOGGER_APP, "curl "+url, zap.String("response", string(responseBody)))
	resp.Body.Close()
	return nil
}

func CheckNodeExporterProcessConfig(endpointId int, processList []m.ProcessMonitorTable) (err error, illegal bool, msg string) {
	endpointObj := m.EndpointTable{Id: endpointId}
	err = GetEndpoint(&endpointObj)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Check node_exporter fail, get endpoint msg fail", zap.Error(err))
		return
	}
	var processNameList []string
	for _, v := range processList {
		processNameList = append(processNameList, fmt.Sprintf("%s^%s", v.ProcessName, v.Tags))
	}
	postParam := processHttpDto{Process: processNameList, Check: 1}
	postData, err := json.Marshal(postParam)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Check node_exporter fail, marshal post data fail", zap.Error(err))
		return
	}
	url := fmt.Sprintf("http://%s/process/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Check node_exporter fail, http post fail", zap.Error(err))
		return
	}
	responseBody, _ := ioutil.ReadAll(resp.Body)
	log.Info(nil, log.LOGGER_APP, "curl "+url, zap.String("response", string(responseBody)))
	resp.Body.Close()
	msg = string(responseBody)
	if resp.StatusCode > 300 {
		illegal = true
	} else {
		illegal = false
	}
	return
}

func GetProcessList(endpointId int) (err error, processList []*m.ProcessMonitorTable) {
	err = x.SQL("SELECT * FROM process_monitor WHERE endpoint_id=?", endpointId).Find(&processList)
	return err, processList
}

func UpdateProcess(param m.ProcessUpdateDtoNew, operation string) error {
	var actions []*Action
	if operation == "update" || operation == "add" {
		if operation == "update" {
			actions = append(actions, &Action{Sql: "DELETE FROM process_monitor WHERE endpoint_id=?", Param: []interface{}{param.EndpointId}})
		}
		existMap := make(map[string]int)
		if operation == "add" {
			_, nowProcessList := GetProcessList(param.EndpointId)
			for _, v := range nowProcessList {
				existMap[fmt.Sprintf("%s^%s", v.ProcessName, v.Tags)] = 1
			}
		}
		for _, v := range param.ProcessList {
			var action Action
			params := make([]interface{}, 0)
			existFlag := false
			if operation == "add" {
				if _, b := existMap[fmt.Sprintf("%s^%s", v.ProcessName, v.Tags)]; b {
					if v.DisplayName != "" {
						action.Sql = "UPDATE process_monitor SET display_name=? WHERE endpoint_id=? AND process_name=? AND tags=?"
						params = append(params, v.DisplayName)
						params = append(params, param.EndpointId)
						params = append(params, v.ProcessName)
						params = append(params, v.Tags)
					}
					existFlag = true
				}
			}
			if !existFlag {
				action.Sql = "INSERT INTO process_monitor(endpoint_id,process_name,display_name,tags) VALUE (?,?,?,?)"
				params = append(params, param.EndpointId)
				params = append(params, v.ProcessName)
				params = append(params, v.DisplayName)
				params = append(params, v.Tags)
			}
			action.Param = params
			if action.Sql != "" {
				actions = append(actions, &action)
			}
		}
	}
	if operation == "delete" {
		for _, v := range param.ProcessList {
			actions = append(actions, &Action{Sql: "DELETE FROM process_monitor WHERE endpoint_id=? and process_name=? and tags=?", Param: []interface{}{param.EndpointId, v.ProcessName, v.Tags}})
		}
	}
	if len(actions) > 0 {
		return Transaction(actions)
	} else {
		return nil
	}
}

func UpdateAliveCheckQueue(monitorIp string) error {
	_, err := x.Exec(fmt.Sprintf("INSERT INTO alive_check_queue(message) VALUE ('%s')", monitorIp))
	return err
}

func GetAliveCheckQueue(param string) (err error, result []*m.AliveCheckQueueTable) {
	lastMinDateString := time.Unix(time.Now().Unix()-60, 0).Format("2006-01-02 15:04:05")
	err = x.SQL(fmt.Sprintf("SELECT * FROM alive_check_queue WHERE message='%s' AND update_at>'%s' LIMIT 1", param, lastMinDateString)).Find(&result)
	if err != nil {
		return err, result
	}
	if len(result) == 0 {
		err = fmt.Errorf("get alive_check_queue table fail,nodata with message=%s and update_at>%s", param, lastMinDateString)
	}
	return err, result
}

func GetProcessDisplayMap(endpoint string) map[string]string {
	result := make(map[string]string)
	var processData []m.ProcessMonitorTable
	err := x.SQL("SELECT * FROM process_monitor WHERE endpoint_id IN (SELECT id FROM endpoint WHERE guid=?)", endpoint).Find(&processData)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "get process monitor data with endpoint fail", zap.String("endpoint", endpoint), zap.Error(err))
		return result
	}
	for _, v := range processData {
		for _, vv := range strings.Split(v.ProcessName, ",") {
			tmpName := fmt.Sprintf("%s(%s)", vv, v.Tags)
			if v.DisplayName != "" {
				result[tmpName] = v.DisplayName
			} else {
				result[tmpName] = tmpName
			}
		}
	}
	return result
}
