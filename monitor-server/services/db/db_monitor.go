package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
)

func ListDbMonitor(endpointId int) (result []*m.DbMonitorTable, err error) {
	err = x.SQL("SELECT * FROM db_monitor WHERE endpoint_guid IN (SELECT id FROM endpoint WHERE id=?)", endpointId).Find(&result)
	return result,err
}

func AddDbMonitor(param m.DbMonitorUpdateDto) error {
	endpointObj := m.EndpointTable{Id:param.EndpointId}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("Can not find endpoint with id=%d ", param.EndpointId)
	}
	_,err := x.Exec("INSERT INTO db_monitor(`endpoint_guid`,`name`,`sql`) VALUE (?,?,?)", endpointObj.Guid, param.Name, param.Sql)
	return err
}

func UpdateDbMonitor(param m.DbMonitorUpdateDto) error {
	_,err := x.Exec("UPDATE db_monitor SET name=?,sql=? WHERE id=?", param.Name, param.Sql, param.Id)
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
		return fmt.Errorf("%s", resp.Body)
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
		postData = append(postData, &m.DbMonitorTaskObj{DbType:"mysql",Name:v.Name,Endpoint:v.EndpointGuid,Sql:v.Sql,User:v.User,Password:v.Password,Server:tmpAddress[0],Port:tmpAddress[1]})
	}
	postDataByte,_ := json.Marshal(postData)
	resp,err := http.Post(fmt.Sprintf("%s/db/config", dbExportAddress), "application/json", strings.NewReader(string(postDataByte)))
	if err != nil {
		return fmt.Errorf("Http request to %s/db/config fail,%s ", dbExportAddress, err.Error())
	}
	if resp.StatusCode > 300 {
		return fmt.Errorf("%s", resp.Body)
	}
	return nil
}