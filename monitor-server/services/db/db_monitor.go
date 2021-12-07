package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

func ListDbMonitor(endpointId int) (result []*m.DbMonitorListObj, err error) {
	var tableData []*m.DbMonitorTable
	err = x.SQL("SELECT * FROM db_monitor WHERE endpoint_guid IN (SELECT guid FROM endpoint WHERE id=?) ORDER BY sys_panel DESC", endpointId).Find(&tableData)
	if len(tableData) == 0 {
		return result,err
	}
	recursiveDatas := SearchPanelByName("", "")
	recursiveMap := make(map[string]string)
	for _,v := range recursiveDatas {
		recursiveMap[v.OptionValue] = v.OptionText
	}
	tmpSysPanel := tableData[0].SysPanel
	var tmpRowData []*m.DbMonitorTable
	for _,v := range tableData {
		if v.SysPanel != tmpSysPanel {
			result = append(result, &m.DbMonitorListObj{SysPanel:recursiveMap[tmpSysPanel], SysPanelValue:tmpSysPanel, Data:tmpRowData})
			tmpRowData = []*m.DbMonitorTable{}
			tmpSysPanel = v.SysPanel
		}
		tmpRowData = append(tmpRowData, v)
	}
	if len(tmpRowData) > 0 {
		result = append(result, &m.DbMonitorListObj{SysPanel:recursiveMap[tableData[len(tableData)-1].SysPanel], SysPanelValue:tableData[len(tableData)-1].SysPanel, Data:tmpRowData})
	}
	return result,err
}

func AddDbMonitor(param m.DbMonitorUpdateDto) error {
	endpointObj := m.EndpointTable{Id:param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Can not find endpoint with id=%d ", param.EndpointId)
	}
	_,err := x.Exec("INSERT INTO db_monitor(`endpoint_guid`,`name`,`sql`,`sys_panel`) VALUE (?,?,?,?)", endpointObj.Guid, param.Name, param.Sql, param.SysPanel)
	return err
}

func UpdateDbMonitor(param m.DbMonitorUpdateDto) error {
	_,err := x.Exec("UPDATE db_monitor SET `name`=?,`sql`=?,`sys_panel`=? WHERE id=?", param.Name, param.Sql, param.SysPanel, param.Id)
	return err
}

func DeleteDbMonitor(id int) error {
	_,err := x.Exec("DELETE FROM db_monitor WHERE id=?", id)
	return err
}

func CheckDbMonitor(param m.DbMonitorUpdateDto) error {
	var dbExportAddress string
	for _,v := range m.Config().Dependence {
		if v.Name == "db_data_exporter" {
			dbExportAddress = v.Server
			break
		}
	}
	if dbExportAddress == "" {
		return fmt.Errorf("Can not find db_data_exporter address ")
	}
	endpointObj := m.EndpointTable{Id:param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Can not find endpoint with id=%d ", param.EndpointId)
	}
	agentManagerTable,err := GetAgentManager(endpointObj.Guid)
	if err != nil {
		return fmt.Errorf("Get agent manager config with endpoint_guid=%s fail,%s ", endpointObj.Guid, err.Error())
	}
	if len(agentManagerTable) == 0 {
		return fmt.Errorf("Can not find agent manager config with endpoint_guid=%s ", endpointObj.Guid)
	}
	var postData m.DbMonitorTaskObj
	postData.DbType = "mysql"
	postData.Endpoint = endpointObj.Guid
	postData.Name = param.Name
	postData.Sql = param.Sql
	instanceAddress := strings.Split(agentManagerTable[0].InstanceAddress, ":")
	postData.Server = instanceAddress[0]
	postData.Port = instanceAddress[1]
	postData.User = agentManagerTable[0].User
	postData.Password = agentManagerTable[0].Password
	postDataByte,_ := json.Marshal(postData)
	resp,err := http.Post(fmt.Sprintf("%s/db/check", dbExportAddress), "application/json", strings.NewReader(string(postDataByte)))
	if err != nil {
		return fmt.Errorf("Http request to %s/db/check fail,%s ", dbExportAddress, err.Error())
	}
	if resp.StatusCode > 300 {
		bodyByte,_ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("%s", string(bodyByte))
	}
	return nil
}

func SendConfigToDbManager() error {
	var dbExportAddress string
	for _,v := range m.Config().Dependence {
		if v.Name == "db_data_exporter" {
			dbExportAddress = v.Server
			break
		}
	}
	if dbExportAddress == "" {
		return fmt.Errorf("Can not find db_data_exporter address ")
	}
	var queryData []*m.DbMonitorConfigQuery
	err := x.SQL("SELECT t1.endpoint_guid,t1.name,t1.sql,t2.user,t2.password,t2.instance_address FROM db_monitor t1 LEFT JOIN agent_manager t2 ON t1.endpoint_guid=t2.endpoint_guid").Find(&queryData)
	if err != nil {
		return fmt.Errorf("Query db monitor table data fail,%s ", err.Error())
	}
	if len(queryData) == 0 {
		return nil
	}
	var postData []*m.DbMonitorTaskObj
	for _,v := range queryData {
		tmpAddress := strings.Split(v.InstanceAddress, ":")
		if len(tmpAddress) < 2 {
			continue
		}
		postData = append(postData, &m.DbMonitorTaskObj{DbType:"mysql",Name:v.Name,Endpoint:v.EndpointGuid,Sql:v.Sql,User:v.User,Password:v.Password,Server:tmpAddress[0],Port:tmpAddress[1]})
	}
	postDataByte,_ := json.Marshal(postData)
	resp,err := http.Post(fmt.Sprintf("%s/db/config", dbExportAddress), "application/json", strings.NewReader(string(postDataByte)))
	if err != nil {
		return fmt.Errorf("Http request to %s/db/config fail,%s ", dbExportAddress, err.Error())
	}
	if resp.StatusCode > 300 {
		bodyByte,_ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("%s", string(bodyByte))
	}
	return nil
}

func GetDbMonitorByPanel(guid string) (result []*m.DbMonitorTable, err error) {
	err = x.SQL("SELECT * FROM db_monitor WHERE sys_panel=?", guid).Find(&result)
	return result,err
}

func GetDbMonitorChart() (result []*m.ChartTable, err error) {
	err = x.SQL("SELECT * FROM chart WHERE metric='db_monitor_count'").Find(&result)
	return result,err
}

func UpdateDbMonitorSysName(param m.DbMonitorSysNameDto) error {
	endpointObj := m.EndpointTable{Id:param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Endpoint id %d can not find any endpoint ", param.EndpointId)
	}
	_,err := x.Exec("UPDATE db_monitor SET sys_panel=? WHERE endpoint_guid=? AND sys_panel=?", param.NewName, endpointObj.Guid, param.OldName)
	if err != nil {
		log.Logger.Error("UpdateDbMonitorSysName fail", log.Error(err))
	}
	return err
}

type logHttpDto struct {
	Path  string  `json:"path"`
	Keywords  []string  `json:"keywords"`
}

func SendLogConfig(endpointId,grpId,tplId int) error {
	var endpoints []*m.EndpointTable
	var err error
	if grpId > 0 {
		err,endpoints = GetEndpointsByGrp(grpId)
		if err != nil {
			return err
		}
	}
	if endpointId > 0 {
		endpointQuery := m.EndpointTable{Id:endpointId}
		err = GetEndpoint(&endpointQuery)
		if err != nil {
			return err
		}
		endpoints = append(endpoints, &endpointQuery)
	}
	var postParam []logHttpDto
	var tmpList []string
	var tmpPath string
	for _,v := range endpoints {
		err,logMonitors := GetLogMonitorByEndpointNew(v.Id)
		if err != nil {
			log.Logger.Error("Send log config with endpoint failed", log.String("endpoint", v.Guid), log.Error(err))
			continue
		}
		if len(logMonitors) == 0 {
			continue
		}
		postParam = []logHttpDto{}
		tmpList = []string{}
		tmpPath = logMonitors[0].Path
		for _,v := range logMonitors {
			if v.Path != tmpPath {
				postParam = append(postParam, logHttpDto{Path:tmpPath, Keywords:tmpList})
				tmpPath = v.Path
				tmpList = []string{}
			}
			tmpList = append(tmpList, v.Keyword)
		}
		postParam = append(postParam, logHttpDto{Path:logMonitors[len(logMonitors)-1].Path, Keywords:tmpList})
		postData,err := json.Marshal(postParam)
		if err == nil {
			url := fmt.Sprintf("http://%s/log/config", v.Address)
			resp,err := http.Post(url, "application/json", strings.NewReader(string(postData)))
			if err != nil {
				log.Logger.Error("curl "+url+" error ", log.Error(err))
			}else{
				responseBody,_ := ioutil.ReadAll(resp.Body)
				log.Logger.Info("curl " + url, log.String("response", string(responseBody)))
				resp.Body.Close()
			}
		}
	}
	return nil
}