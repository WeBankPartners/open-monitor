package agent

import (
	"github.com/gin-gonic/gin"
	"strings"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
)

type resultObj struct {
	ResultCode  string  `json:"resultCode"`
	ResultMessage  string  `json:"resultMessage"`
	Results  resultOutput  `json:"results"`
}

type resultOutput struct {
	Outputs  []resultOutputObj  `json:"outputs"`
}

type resultOutputObj struct {
	CallbackParameter  string  `json:"callbackParameter"`
	Guid  string  `json:"guid"`
	ErrorCode  string  `json:"errorCode"`
	ErrorMessage  string  `json:"errorMessage"`
}

type requestObj struct {
	RequestId  string  	`json:"requestId"`
	Inputs  []hostRequestObj  `json:"inputs"`
}

type hostRequestObj struct {
	CallbackParameter  string  `json:"callbackParameter"`
	HostIp  string  `json:"host_ip"`
	InstanceIp  string  `json:"instance_ip"`
	Group  string  `json:"group"`
	Port  string  `json:"port"`
	Instance  string  `json:"instance"`
	User  string  `json:"user"`
	Password  string  `json:"password"`
	JavaType  string  `json:"java_type"`
	PasswordGuid  string  `json:"password_guid"`
	PasswordSeed  string  `json:"password_seed"`
}

func ExportAgentNew(c *gin.Context)  {
	agentType := c.Param("name")
	action := "register"
	if strings.Contains(c.Request.URL.String(), "deregister") {
		action = "deregister"
	}
	var resultCode,resultMessage string
	resultCode = "0"
	resultData := resultOutput{}
	defer func() {
		b,_ := json.Marshal(resultData)
		mid.LogInfo(fmt.Sprintf("plugin result -> code: %s ,message: %s ,data: %s", resultCode, resultMessage, string(b)))
		if strings.Contains(resultMessage, "validate") {
			c.JSON(http.StatusBadRequest, resultObj{ResultCode:resultCode, ResultMessage:resultMessage})
		}else{
			c.JSON(http.StatusOK, resultObj{ResultCode:resultCode, ResultMessage:resultMessage, Results:resultData})
		}
	}()
	data,_ := ioutil.ReadAll(c.Request.Body)
	mid.LogInfo(fmt.Sprintf("plugin request -> agent %s, type %s, param : %v", action, agentType, string(data)))
	var param requestObj
	err := json.Unmarshal(data, &param)
	if err != nil {
		resultCode = "1"
		resultMessage = fmt.Sprintf("Param validate fail : %v", err)
		return
	}
	if len(param.Inputs) == 0 {
		resultCode = "0"
		resultMessage = "inputs length is zero,do nothing"
		return
	}
	for i,v := range param.Inputs {
		tmpAgentType := agentType
		if v.JavaType == "tomcat" || v.Group == "default_tomcat_group" {
			tmpAgentType = "tomcat"
		}
		if v.Password != "" {
			tmpPassword, tmpErr := mid.AesDePassword(v.PasswordGuid, v.PasswordSeed, v.Password)
			if tmpErr == nil {
				v.Password = tmpPassword
			}
		}
		var param m.RegisterParamNew
		var validateMessage string
		var inputErr error
		if tmpAgentType == "host" {
			param = m.RegisterParamNew{Type: tmpAgentType, Ip: v.HostIp, Port: "9100", AddDefaultGroup:true, AgentManager:false, FetchMetric:true, DefaultGroupName:v.Group}
		} else {
			param = m.RegisterParamNew{Type: tmpAgentType, Ip: v.InstanceIp, Port: v.Port, Name: v.Instance, User: v.User, Password: v.Password, AgentManager:true, AddDefaultGroup:true, FetchMetric:true, DefaultGroupName:v.Group}
		}
		if action == "register" {
			validateMessage,inputErr = AgentRegister(param)
			if validateMessage != "" {
				validateMessage = fmt.Sprintf("input index:%d validate fail -> ", i) + validateMessage
			}
		} else {
			var endpointObj m.EndpointTable
			if tmpAgentType == "host" {
				endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: tmpAgentType}
			} else {
				endpointObj = m.EndpointTable{Ip: v.InstanceIp, ExportType: tmpAgentType, Name: v.Instance}
			}
			db.GetEndpoint(&endpointObj)
			if endpointObj.AddressAgent != "" {
				agentManagerUrl := ""
				for _, v := range m.Config().Dependence {
					if v.Name == "agent_manager" {
						agentManagerUrl = v.Server
						break
					}
				}
				if agentManagerUrl != "" {
					inputErr = prom.StopAgent(endpointObj.ExportType, endpointObj.Name, endpointObj.Ip, agentManagerUrl)
				}
			}
			if endpointObj.Id > 0 && inputErr == nil {
				inputErr = DeregisterJob(endpointObj.Guid, endpointObj.Step)
			}
		}
		if validateMessage != "" || inputErr != nil {
			errorMessage := validateMessage
			if inputErr != nil {
				errorMessage = fmt.Sprintf("input index:%d %s [agentType:%s, name:%s, hostIp:%s, instanceIp:%s] error -> %v ", i, action, agentType, v.Instance, v.HostIp, v.InstanceIp, inputErr)
			}
			resultData.Outputs = append(resultData.Outputs, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:errorMessage})
			resultCode = "1"
		}else{
			resultData.Outputs = append(resultData.Outputs, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
		}
	}
}

func ExportAgent(c *gin.Context)  {
	agentType := c.Param("name")
	action := "register"
	if strings.Contains(c.Request.URL.String(), "deregister") {
		action = "deregister"
	}
	var agentPort string
	var result resultObj
	//illegal := true
	for _,v := range m.Config().Agent {
		tmpAgentType := agentType
		if tmpAgentType == "java" {
			tmpAgentType = "tomcat"
		}
		if v.AgentType == tmpAgentType {
			//illegal = false
			agentPort = v.Port
			break
		}
	}
	if action != "register" && action != "deregister" {
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("No such action like %s", action)}
		mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
		c.JSON(http.StatusBadRequest, result)
		return
	}
	data,_ := ioutil.ReadAll(c.Request.Body)
	mid.LogInfo(fmt.Sprintf("param : %v", string(data)))
	var param requestObj
	err := json.Unmarshal(data, &param)
	if err == nil {
		if len(param.Inputs) == 0 {
			result = resultObj{ResultCode:"0", ResultMessage:"inputs length is zero,do nothing"}
			mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
			c.JSON(http.StatusOK, result)
			return
		}
		var tmpResult []resultOutputObj
		successFlag := "0"
		errMessage := "Done"
		// update table and register to consul
		for _,v := range param.Inputs {
			tmpAgentType := agentType
			if v.JavaType == "tomcat" || v.Group == "default_tomcat_group" {
				tmpAgentType = "tomcat"
			}
			if v.Password != "" {
				tmpPassword,tmpErr := mid.AesDePassword(v.PasswordGuid,v.PasswordSeed,v.Password)
				if tmpErr == nil {
					v.Password = tmpPassword
				}
			}
			var param m.RegisterParam
			if tmpAgentType == "host" {
				param = m.RegisterParam{Type: tmpAgentType, ExporterIp: v.HostIp, ExporterPort: agentPort}
			}else{
				param = m.RegisterParam{Type: tmpAgentType, ExporterIp: v.InstanceIp, ExporterPort: v.Port, Instance: v.Instance, User:v.User, Password:v.Password}
			}
			if action == "register" {
				err = RegisterJob(param)
			}else{
				var endpointObj m.EndpointTable
				if tmpAgentType == "host" {
					endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: tmpAgentType}
				}else{
					endpointObj = m.EndpointTable{Ip: v.InstanceIp, ExportType: tmpAgentType, Name: v.Instance}
				}
				db.GetEndpoint(&endpointObj)
				if endpointObj.AddressAgent != "" {
					agentManagerUrl := ""
					for _, v := range m.Config().Dependence {
						if v.Name == "agent_manager" {
							agentManagerUrl = v.Server
							break
						}
					}
					if agentManagerUrl != "" {
						err = prom.StopAgent(endpointObj.ExportType, endpointObj.Name, endpointObj.Ip, agentManagerUrl)
					}
				}
				if endpointObj.Id > 0 {
					err = DeregisterJob(endpointObj.Guid)
				}
			}
			var msg string
			if err != nil {
				msg = fmt.Sprintf("%s %s:%s %s fail,error %v",action, tmpAgentType, v.HostIp, v.Instance, err)
				errMessage = msg
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:msg})
				successFlag = "1"
			}else{
				msg = fmt.Sprintf("%s %s:%s %s succeed", action, tmpAgentType, v.HostIp, v.Instance)
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
			}
			mid.LogInfo(msg)
		}
		// update group and sync prometheus config
		if action == "register" {
			var groupTplMap= make(map[string]int)
			for _, v := range param.Inputs {
				if v.Group == "" {
					continue
				}
				_, grpObj := db.GetSingleGrp(0, v.Group)
				if grpObj.Id > 0 {
					_, tplObj := db.GetTpl(0, grpObj.Id, 0)
					groupTplMap[grpObj.Name] = tplObj.Id
					var endpointObj m.EndpointTable
					if agentType == "host" {
						endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: agentType}
					} else {
						endpointObj = m.EndpointTable{Ip: v.InstanceIp, ExportType: agentType, Name: v.Instance}
					}
					db.GetEndpoint(&endpointObj)
					if endpointObj.Id > 0 {
						err, _ := db.UpdateGrpEndpoint(m.GrpEndpointParamNew{Grp: grpObj.Id, Endpoints: []int{endpointObj.Id}, Operation: "add"})
						if err != nil {
							mid.LogError("register interface update group_endpoint fail ", err)
						}
						if agentType == "telnet" {
							var eto []*m.EndpointTelnetObj
							eto = append(eto, &m.EndpointTelnetObj{Port:v.Port, Note:""})
							db.UpdateEndpointTelnet(m.UpdateEndpointTelnetParam{Guid:endpointObj.Guid, Config:eto})
						}
					}
				}
			}
			for k, v := range groupTplMap {
				err := alarm.SaveConfigFile(v, false)
				if err != nil {
					mid.LogError(fmt.Sprintf("register interface update prometheus config fail , group : %s  error ", k), err)
				}
			}
		}
		result = resultObj{ResultCode: successFlag, ResultMessage: errMessage, Results: resultOutput{Outputs: tmpResult}}
		resultString,_ := json.Marshal(result)
		mid.LogInfo(string(resultString))
		mid.ReturnData(c, result)
	}else{
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("Param validate fail : %v", err)}
		mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
		c.JSON(http.StatusBadRequest, result)
	}
}

func GetSystemDashboardUrl(c *gin.Context)  {
	name := c.Query("system_name")
	ips := c.Query("ips")
	urlParms := url.Values{}
	urlParms.Set("systemName", name)
	urlParms.Set("ips", ips)
	urlPath := fmt.Sprintf("http://%s/wecube-monitor/#/systemMonitoring?%s", c.Request.Host, urlParms.Encode())
	mid.LogInfo(fmt.Sprintf("url : %s", urlPath))
	mid.ReturnData(c, resultObj{ResultCode:"0", ResultMessage:urlPath})
}

func AlarmControl(c *gin.Context)  {
	agentType := c.Param("name")
	if agentType == "java" {
		agentType = "tomcat"
	}
	isStop := false
	action := "start"
	if strings.Contains(c.Request.URL.String(), "stop") {
		isStop = true
		action = "stop"
	}
	var result resultObj
	var agentPort string
	for _,v := range m.Config().Agent {
		if v.AgentType == agentType {
			agentPort = v.Port
			break
		}
	}
	//if illegal {
	//	result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("No such monitor type like %s", agentType)}
	//	mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
	//	c.JSON(http.StatusBadRequest, result)
	//	return
	//}
	data,_ := ioutil.ReadAll(c.Request.Body)
	mid.LogInfo(fmt.Sprintf("param : %v", string(data)))
	var param requestObj
	err := json.Unmarshal(data, &param)
	if err == nil {
		if len(param.Inputs) == 0 {
			result = resultObj{ResultCode:"0", ResultMessage:"inputs length is zero,do nothing"}
			mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
			c.JSON(http.StatusOK, result)
			return
		}
		var tmpResult []resultOutputObj
		for _,v := range param.Inputs {
			if agentType == "tomcat" && v.Port != "" {
				agentPort = v.Port
			}
			tmpIp := v.HostIp
			if agentType != "host" {
				tmpIp = v.InstanceIp
			}
			err := db.UpdateEndpointAlarmFlag(isStop,agentType,v.Instance,tmpIp,agentPort)
			var msg string
			if err != nil {
				msg = fmt.Sprintf("%s %s:%s %s fail,error %v",action, agentType, v.HostIp, v.Instance, err)
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:msg})
			}else{
				msg = fmt.Sprintf("%s %s:%s %s succeed", action, agentType, v.HostIp, v.Instance)
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
			}
			mid.LogInfo(msg)
		}
		result = resultObj{ResultCode:"0", ResultMessage:"Done", Results:resultOutput{Outputs:tmpResult}}
		resultString,_ := json.Marshal(result)
		mid.LogInfo(string(resultString))
		mid.ReturnData(c, result)
	}else{
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("Param validate fail : %v", err)}
		mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
		c.JSON(http.StatusBadRequest, result)
	}
}

func ExportPingSource(c *gin.Context)  {
	ips := db.GetPingExporterSource()
	mid.ReturnData(c, m.PingExporterSourceDto{Config:ips})
}

func UpdateEndpointTelnet(c *gin.Context)  {
	var param m.UpdateEndpointTelnetParam
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateEndpointTelnet(param)
		if err != nil {
			mid.ReturnError(c, "Update endpoint telnet config fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validate fail %v", err))
	}
}

func GetEndpointTelnet(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "Guid can not empty")
		return
	}
	result,err := db.GetEndpointTelnet(guid)
	if err != nil {
		mid.ReturnError(c, "Get endpoint telnet config fail", err)
	}else{
		mid.ReturnData(c, result)
	}
}