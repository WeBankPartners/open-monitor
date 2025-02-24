package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
)

func ListDbMonitor(endpointId int) (result []*m.DbMonitorListObj, err error) {
	var tableData []*m.DbMonitorTable
	err = x.SQL("SELECT * FROM db_monitor WHERE endpoint_guid IN (SELECT guid FROM endpoint WHERE id=?) ORDER BY sys_panel DESC", endpointId).Find(&tableData)
	if len(tableData) == 0 {
		return result, err
	}
	recursiveDatas := SearchPanelByName("", "")
	recursiveMap := make(map[string]string)
	for _, v := range recursiveDatas {
		recursiveMap[v.OptionValue] = v.OptionText
	}
	tmpSysPanel := tableData[0].SysPanel
	var tmpRowData []*m.DbMonitorTable
	for _, v := range tableData {
		if v.SysPanel != tmpSysPanel {
			result = append(result, &m.DbMonitorListObj{SysPanel: recursiveMap[tmpSysPanel], SysPanelValue: tmpSysPanel, Data: tmpRowData})
			tmpRowData = []*m.DbMonitorTable{}
			tmpSysPanel = v.SysPanel
		}
		tmpRowData = append(tmpRowData, v)
	}
	if len(tmpRowData) > 0 {
		result = append(result, &m.DbMonitorListObj{SysPanel: recursiveMap[tableData[len(tableData)-1].SysPanel], SysPanelValue: tableData[len(tableData)-1].SysPanel, Data: tmpRowData})
	}
	return result, err
}

func AddDbMonitor(param m.DbMonitorUpdateDto) error {
	endpointObj := m.EndpointTable{Id: param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Can not find endpoint with id=%d ", param.EndpointId)
	}
	_, err := x.Exec("INSERT INTO db_monitor(`endpoint_guid`,`name`,`sql`,`sys_panel`) VALUE (?,?,?,?)", endpointObj.Guid, param.Name, param.Sql, param.SysPanel)
	return err
}

func UpdateDbMonitor(param m.DbMonitorUpdateDto) error {
	_, err := x.Exec("UPDATE db_monitor SET `name`=?,`sql`=?,`sys_panel`=? WHERE id=?", param.Name, param.Sql, param.SysPanel, param.Id)
	return err
}

func DeleteDbMonitor(id int) error {
	_, err := x.Exec("DELETE FROM db_monitor WHERE id=?", id)
	return err
}

func CheckDbMonitor(param m.DbMonitorUpdateDto) error {
	var dbExportAddress string
	for _, v := range m.Config().Dependence {
		if v.Name == "db_data_exporter" {
			dbExportAddress = v.Server
			break
		}
	}
	if dbExportAddress == "" {
		return fmt.Errorf("Can not find db_data_exporter address ")
	}
	endpointObj := m.EndpointTable{Id: param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Can not find endpoint with id=%d ", param.EndpointId)
	}
	agentManagerTable, err := GetAgentManager(endpointObj.Guid)
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
	postDataByte, _ := json.Marshal(postData)
	resp, err := http.Post(fmt.Sprintf("%s/db/check", dbExportAddress), "application/json", strings.NewReader(string(postDataByte)))
	if err != nil {
		return fmt.Errorf("Http request to %s/db/check fail,%s ", dbExportAddress, err.Error())
	}
	bodyByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode > 300 {
		return fmt.Errorf("%s", string(bodyByte))
	}
	return nil
}

func SendConfigToDbManager() error {
	var dbExportAddress string
	for _, v := range m.Config().Dependence {
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
	for _, v := range queryData {
		tmpAddress := strings.Split(v.InstanceAddress, ":")
		if len(tmpAddress) < 2 {
			continue
		}
		postData = append(postData, &m.DbMonitorTaskObj{DbType: "mysql", Name: v.Name, Endpoint: v.EndpointGuid, Sql: v.Sql, User: v.User, Password: v.Password, Server: tmpAddress[0], Port: tmpAddress[1]})
	}
	postDataByte, _ := json.Marshal(postData)
	resp, err := http.Post(fmt.Sprintf("%s/db/config", dbExportAddress), "application/json", strings.NewReader(string(postDataByte)))
	if err != nil {
		return fmt.Errorf("Http request to %s/db/config fail,%s ", dbExportAddress, err.Error())
	}
	bodyByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode > 300 {
		return fmt.Errorf("%s", string(bodyByte))
	}
	return nil
}

func GetDbMonitorByPanel(guid string) (result []*m.DbMonitorTable, err error) {
	err = x.SQL("SELECT * FROM db_monitor WHERE sys_panel=?", guid).Find(&result)
	return result, err
}

func GetDbMonitorChart() (result []*m.ChartTable, err error) {
	err = x.SQL("SELECT * FROM chart WHERE metric='db_monitor_count'").Find(&result)
	return result, err
}

func UpdateDbMonitorSysName(param m.DbMonitorSysNameDto) error {
	endpointObj := m.EndpointTable{Id: param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Endpoint id %d can not find any endpoint ", param.EndpointId)
	}
	_, err := x.Exec("UPDATE db_monitor SET sys_panel=? WHERE endpoint_guid=? AND sys_panel=?", param.NewName, endpointObj.Guid, param.OldName)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "UpdateDbMonitorSysName fail", zap.Error(err))
	}
	return err
}

type logKeywordHttpRuleObj struct {
	RegularEnable bool    `json:"regular_enable"`
	Keyword       string  `json:"keyword"`
	Count         float64 `json:"count"`
}

type logKeywordHttpDto struct {
	Path     string                   `json:"path"`
	Keywords []*logKeywordHttpRuleObj `json:"keywords"`
}

func SendLogConfig(endpointId, grpId, tplId int) error {
	log.Info(nil, log.LOGGER_APP, "SendLogConfig", zap.Int("endpointId", endpointId), zap.Int("grpId", grpId))
	var endpoints []*m.EndpointTable
	var err error
	if grpId > 0 {
		err, endpoints = GetEndpointsByGrp(grpId)
		if err != nil {
			return err
		}
	}
	if endpointId > 0 {
		endpointQuery := m.EndpointTable{Id: endpointId}
		err = GetEndpoint(&endpointQuery)
		if err != nil {
			return err
		}
		endpoints = append(endpoints, &endpointQuery)
	}
	log.Info(nil, log.LOGGER_APP, "SendLogConfig", log.JsonObj("endpoints", endpoints))
	var tmpPath string
	for _, v := range endpoints {
		err, logMonitors := GetLogMonitorByEndpointNew(v.Id)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Send log config with endpoint failed", zap.String("endpoint", v.Guid), zap.Error(err))
			continue
		}
		if len(logMonitors) == 0 {
			continue
		}
		postParam := []*logKeywordHttpDto{}
		keywordList := []*logKeywordHttpRuleObj{}
		tmpPath = logMonitors[0].Path
		for _, v := range logMonitors {
			if v.Path != tmpPath {
				postParam = append(postParam, &logKeywordHttpDto{Path: tmpPath, Keywords: keywordList})
				tmpPath = v.Path
				keywordList = []*logKeywordHttpRuleObj{}
			}
			keywordList = append(keywordList, &logKeywordHttpRuleObj{Keyword: v.Keyword})
		}
		postParam = append(postParam, &logKeywordHttpDto{Path: logMonitors[len(logMonitors)-1].Path, Keywords: keywordList})
		postData, err := json.Marshal(postParam)
		if err == nil {
			log.Info(nil, log.LOGGER_APP, "Sync log keyword config", zap.String("endpoint", v.Address), zap.String("param", string(postData)))
			url := fmt.Sprintf("http://%s/log/config", v.Address)
			resp, respErr := http.Post(url, "application/json", strings.NewReader(string(postData)))
			if respErr != nil {
				log.Error(nil, log.LOGGER_APP, "curl "+url+" error ", zap.Error(respErr))
			} else {
				responseBody, _ := ioutil.ReadAll(resp.Body)
				log.Info(nil, log.LOGGER_APP, "curl "+url, zap.String("response", string(responseBody)))
				resp.Body.Close()
			}
		}
	}
	return nil
}
