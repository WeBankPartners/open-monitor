package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/agent"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func ListEndpoint(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	monitorType := c.Query("monitorType")
	param := models.QueryRequestParam{}
	if size > 0 {
		param.Paging = true
		param.Pageable = &models.PageInfo{PageSize: size, StartIndex: page - 1}
	}
	if monitorType != "" {
		param.Filters = []*models.QueryRequestFilterObj{{Name: "monitor_type", Operator: "eq", Value: monitorType}}
	}
	pageInfo, rowData, err := db.ListEndpoint(&param)
	returnData := models.TableData{Data: rowData, Page: page, Size: size, Num: pageInfo.TotalRows}
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, returnData)
	}
}

func GetEndpoint(c *gin.Context)  {
	guid := c.Param("guid")
	endpointObj,err := db.GetEndpointNew(&models.EndpointNewTable{Guid: guid})
	if err != nil {
		middleware.ReturnValidateError(c, fmt.Sprintf("Endpoint:%s query error:%s ", guid, err.Error()))
		return
	}
	result := models.RegisterParamNew{Guid: guid,Type: endpointObj.MonitorType,Name: endpointObj.Name,Ip: endpointObj.Ip}
	result.Cluster = endpointObj.Cluster
	result.Step = endpointObj.Step
	if strings.Contains(endpointObj.EndpointAddress, ":") {
		result.Port = strings.Split(endpointObj.EndpointAddress, ":")[1]
	}
	if endpointObj.ExtendParam != "" {
		if db.CheckEndpointInAgentManager(guid) {
			result.AgentManager = true
		}
		var extendObj models.EndpointExtendParamObj
		err = json.Unmarshal([]byte(endpointObj.ExtendParam), &extendObj)
		if err == nil {
			result.User = extendObj.User
			result.Password = extendObj.Password
			result.ProcessName = extendObj.ProcessName
			if endpointObj.MonitorType == "process" {
				result.Tags = extendObj.ProcessTags
			}
			result.ExportAddress = extendObj.ExportAddress
			result.Url = extendObj.HttpUrl
			result.Method = extendObj.HttpMethod
			result.ProxyExporter = extendObj.ProxyExporter
		}
	}
	middleware.ReturnSuccessData(c, result)
}

func AddEndpoint(c *gin.Context)  {

}

func UpdateEndpoint(c *gin.Context)  {
	var param models.RegisterParamNew
	var err error
	if err=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.Guid == "" {
		middleware.ReturnValidateError(c, "Param guid can not empty")
		return
	}
	defer func(resultErr error) {
		if resultErr != nil {
			middleware.ReturnHandleError(c, resultErr.Error(), resultErr)
		}else{
			middleware.ReturnSuccess(c)
		}
	}(err)
	guid := param.Guid
	endpointObj,queryErr := db.GetEndpointNew(&models.EndpointNewTable{Guid: guid})
	if queryErr != nil {
		err = queryErr
		return
	}
	var newEndpoint models.EndpointNewTable
	switch param.Type {
	case "host":
		newEndpoint,err = hostEndpointUpdate(&param,&endpointObj)
	case "mysql":
		newEndpoint,err = agentManagerEndpointUpdate(&param,&endpointObj)
	case "redis":
		newEndpoint,err = agentManagerEndpointUpdate(&param,&endpointObj)
	case "java":
		newEndpoint,err = agentManagerEndpointUpdate(&param,&endpointObj)
	case "nginx":
		newEndpoint,err = agentManagerEndpointUpdate(&param,&endpointObj)
	case "ping":
		newEndpoint,err = pingEndpointUpdate(&param,&endpointObj)
	case "telnet":
		newEndpoint,err = telnetEndpointUpdate(&param,&endpointObj)
	case "http":
		newEndpoint,err = httpEndpointUpdate(&param,&endpointObj)
	case "windows":
		newEndpoint,err = windowsEndpointUpdate(&param,&endpointObj)
	case "snmp":
		newEndpoint,err = snmpEndpointUpdate(&param,&endpointObj)
	case "process":
		newEndpoint,err = processEndpointUpdate(&param,&endpointObj)
	default:
		newEndpoint,err = otherEndpointUpdate(&param, &endpointObj)
	}
	if err != nil {
		return
	}
	if newEndpoint.Guid == "" {
		if endpointObj.Step == param.Step {
			// no change
			return
		}else{
			newEndpoint = models.EndpointNewTable{Guid: endpointObj.Guid,AgentAddress: endpointObj.AgentAddress,EndpointAddress: endpointObj.EndpointAddress,Step: param.Step,ExtendParam: endpointObj.ExtendParam}
		}
	}
	log.Logger.Info("new endpoint", log.JsonObj("endpoint", newEndpoint))
	// update endpoint table
	err = db.UpdateEndpointData(&newEndpoint)
	if err != nil {
		return
	}
	// update sd file if step change
	if endpointObj.Step != param.Step || endpointObj.AgentAddress != newEndpoint.AgentAddress {
		stepList := []int{endpointObj.Step}
		if endpointObj.Step != param.Step {
			stepList = append(stepList, param.Step)
		}
		err = db.SyncSdEndpointNew(stepList, endpointObj.Cluster, false)
		if err != nil {
			return
		}
	}
}

func DeleteEndpoint(c *gin.Context)  {

}

func hostEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	if strings.Contains(endpoint.AgentAddress, ":") {
		if param.Port == endpoint.AgentAddress[strings.LastIndex(endpoint.AgentAddress,":")+1:] {
			return
		}else{
			newAddress := fmt.Sprintf("%s:%s", param.Ip, param.Port)
			newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, AgentAddress: newAddress,EndpointAddress: newAddress}
		}
	}
	return
}

func agentManagerEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	if param.AgentManager {
		var extParamObj models.EndpointExtendParamObj
		err = json.Unmarshal([]byte(endpoint.ExtendParam), &extParamObj)
		if err != nil {
			return newEndpoint,fmt.Errorf("json unmarhsal extendParam fail,%s ", err.Error())
		}
		if param.Port == extParamObj.Port && param.User == extParamObj.User && param.Password == extParamObj.Password {
			return
		}else{
			err = prom.StopAgent(endpoint.MonitorType, endpoint.Name, endpoint.Ip, agent.AgentManagerServer)
			if err != nil {
				return newEndpoint,fmt.Errorf("stop agent manager instance fail,%s ", err.Error())
			}
			agentConfig := getAgentMangerInstanceConfig(endpoint.MonitorType)
			address, deployErr := prom.DeployAgent(param.Type, param.Name, agentConfig.AgentBin, param.Ip, param.Port, param.User, param.Password, agent.AgentManagerServer, agentConfig.ConfigFile)
			if deployErr != nil {
				return newEndpoint,fmt.Errorf("deploy agent manager instance fail,%s ", deployErr.Error())
			}
			newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, EndpointAddress: fmt.Sprintf("%s:%s",param.Ip,param.Port),AgentAddress: address}
			newParamObj := models.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port, User: param.User, Password: param.Password, BinPath: agentConfig.AgentBin, ConfigPath: agentConfig.ConfigFile}
			b,_ := json.Marshal(newParamObj)
			newEndpoint.ExtendParam = string(b)
			err = db.UpdateAgentManager(&models.AgentManagerTable{EndpointGuid: endpoint.Guid,User: param.User,Password: param.Password,InstanceAddress: newEndpoint.EndpointAddress,AgentAddress: address})
			return
		}
	}else{
		if strings.Contains(endpoint.AgentAddress, ":") {
			if param.Port == endpoint.AgentAddress[strings.LastIndex(endpoint.AgentAddress,":")+1:] {
				return
			}
		}else{
			newEndpoint = models.EndpointNewTable{Guid: param.Guid, AgentAddress: fmt.Sprintf("%s:%s", param.Ip, param.Port)}
		}
	}
	return
}

func processEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	var extParamObj models.EndpointExtendParamObj
	err = json.Unmarshal([]byte(endpoint.ExtendParam), &extParamObj)
	if err != nil {
		return newEndpoint,fmt.Errorf("json unmarhsal extendParam fail,%s ", err.Error())
	}
	if param.ProcessName == extParamObj.ProcessName && param.Tags == extParamObj.ProcessTags {
		return
	}
	newExtParamObj := models.EndpointExtendParamObj{Enable: true, ProcessName: param.ProcessName, ProcessTags: param.Tags}
	b, _ := json.Marshal(newExtParamObj)
	newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, EndpointAddress: endpoint.EndpointAddress,AgentAddress: endpoint.AgentAddress,ExtendParam: string(b)}
	err = db.SyncNodeExporterProcessConfig(endpoint.Ip, []*models.EndpointNewTable{&newEndpoint}, true)
	return
}

func windowsEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func pingEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	var extParamObj models.EndpointExtendParamObj
	err = json.Unmarshal([]byte(endpoint.ExtendParam), &extParamObj)
	if err != nil {
		return newEndpoint,fmt.Errorf("json unmarhsal extendParam fail,%s ", err.Error())
	}
	if param.ProxyExporter == extParamObj.ProxyExporter {
		return
	}
	newExtParamObj := models.EndpointExtendParamObj{Enable: true, ProxyExporter: param.ProxyExporter}
	b, _ := json.Marshal(newExtParamObj)
	newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, EndpointAddress: endpoint.EndpointAddress,AgentAddress: endpoint.AgentAddress,ExtendParam: string(b)}
	return
}

func telnetEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	if strings.Contains(endpoint.EndpointAddress, ":") {
		if param.Port == endpoint.EndpointAddress[strings.LastIndex(endpoint.EndpointAddress,":")+1:] {
			return
		}else{
			newAddress := fmt.Sprintf("%s:%s", param.Ip, param.Port)
			newExtParamObj := models.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port}
			b, _ := json.Marshal(newExtParamObj)
			newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, AgentAddress: newAddress, EndpointAddress: newAddress, ExtendParam: string(b)}
		}
	}
	return
}

func httpEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	var extParamObj models.EndpointExtendParamObj
	err = json.Unmarshal([]byte(endpoint.ExtendParam), &extParamObj)
	if err != nil {
		return newEndpoint,fmt.Errorf("json unmarhsal extendParam fail,%s ", err.Error())
	}
	if extParamObj.HttpUrl == param.Url && extParamObj.HttpMethod == param.Method {
		return
	}
	newExtParamObj := models.EndpointExtendParamObj{Enable: true, HttpUrl: param.Url, HttpMethod: param.Method}
	b, _ := json.Marshal(newExtParamObj)
	newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, EndpointAddress: endpoint.EndpointAddress,AgentAddress: endpoint.AgentAddress,ExtendParam: string(b)}
	return
}

func snmpEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func otherEndpointUpdate(param *models.RegisterParamNew,endpoint *models.EndpointNewTable) (newEndpoint models.EndpointNewTable,err error) {
	if strings.Contains(endpoint.AgentAddress, ":") {
		if param.Port == endpoint.AgentAddress[strings.LastIndex(endpoint.AgentAddress,":")+1:] {
			return
		}else{
			newAddress := fmt.Sprintf("%s:%s", param.Ip, param.Port)
			newEndpoint = models.EndpointNewTable{Guid: endpoint.Guid, AgentAddress: newAddress,EndpointAddress: newAddress}
		}
	}
	return
}

func getAgentMangerInstanceConfig(monitorType string) (result *models.AgentConfig) {
	for _, v := range models.Config().Agent {
		if v.AgentType == monitorType {
			result = v
			break
		}
	}
	return
}