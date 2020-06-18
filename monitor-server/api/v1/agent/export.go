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
	"strconv"
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
	Inputs  []endpointRequestObj  `json:"inputs"`
}

type endpointRequestObj struct {
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
	AppLogPaths   string  `json:"app_log_paths"`
	Step  string  `json:"step"`
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
		tmpStep := 10
		if v.Step != "" {
			tmpStep,_ = strconv.Atoi(v.Step)
			if tmpStep <= 0 {
				tmpStep = 10
			}
		}
		if tmpAgentType == "host" {
			param = m.RegisterParamNew{Type: tmpAgentType, Ip: v.HostIp, Port: "9100", AddDefaultGroup:true, AgentManager:false, FetchMetric:true, DefaultGroupName:v.Group, Step:tmpStep}
		} else {
			param = m.RegisterParamNew{Type: tmpAgentType, Ip: v.InstanceIp, Port: v.Port, Name: v.Instance, User: v.User, Password: v.Password, AgentManager:true, AddDefaultGroup:true, FetchMetric:true, DefaultGroupName:v.Group, Step:tmpStep}
		}
		if action == "register" {
			validateMessage,inputErr = AgentRegister(param)
			if validateMessage != "" {
				validateMessage = fmt.Sprintf("input index:%d validate fail -> ", i) + validateMessage
			}
			if validateMessage == "" && inputErr == nil && v.AppLogPaths != "" {
				inputErr = autoAddAppPathConfig(param, v.AppLogPaths)
			}
		} else {
			var endpointObj m.EndpointTable
			if tmpAgentType == "host" {
				endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: tmpAgentType}
			} else {
				endpointObj = m.EndpointTable{Ip: v.InstanceIp, ExportType: tmpAgentType, Name: v.Instance}
			}
			db.GetEndpoint(&endpointObj)
			mid.LogInfo(fmt.Sprintf("Export deregister endpoint id:%d guid:%s ", endpointObj.Id, endpointObj.Guid))
			if endpointObj.Id > 0 {
				inputErr = DeregisterJob(endpointObj.Guid)
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

func autoAddAppPathConfig(param m.RegisterParamNew, paths string) error {
	tmpPathList := strings.Split(trimListString(paths), ",")
	if len(tmpPathList) == 0 {
		return nil
	}
	hostEndpoint := m.EndpointTable{Ip:param.Ip, ExportType:"host"}
	db.GetEndpoint(&hostEndpoint)
	if hostEndpoint.Id <= 0 {
		return fmt.Errorf("Host endpoint with ip:%s can not find,please register this host first ", param.Ip)
	}
	var businessTables []*m.BusinessMonitorTable
	for _,v := range tmpPathList {
		businessTables = append(businessTables, &m.BusinessMonitorTable{EndpointId:hostEndpoint.Id, Path:v, OwnerEndpoint:fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)})
	}
	err := db.UpdateBusiness(m.BusinessUpdateDto{EndpointId:hostEndpoint.Id, PathList:businessTables})
	if err != nil {
		mid.LogError("Update endpoint business table error ", err)
		return err
	}
	err = alarm.UpdateNodeExporterBusinessConfig(hostEndpoint.Id)
	if err != nil {
		mid.LogError("Update business config error ", err)
	}
	return err
}