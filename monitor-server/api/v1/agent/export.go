package agent

import (
	"github.com/gin-gonic/gin"
	"strings"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"io/ioutil"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
	"strconv"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
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
	MonitorKey  string  `json:"monitor_key"`
	ErrorCode  string  `json:"errorCode"`
	ErrorMessage  string  `json:"errorMessage"`
	ErrorDetail  string  `json:"errorDetail,omitempty"`
}

type requestObj struct {
	RequestId  string  	`json:"requestId"`
	Inputs  []endpointRequestObj  `json:"inputs"`
}

type endpointRequestObj struct {
	Guid  string  `json:"guid"`
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
	Url   string  `json:"url"`
	Method  string  `json:"method"`
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
		log.Logger.Info("Plugin result", log.JsonObj("result", resultData))
		//if strings.Contains(resultMessage, "validate") {
		//	c.JSON(http.StatusBadRequest, resultObj{ResultCode:resultCode, ResultMessage:resultMessage})
		//}else{
		//	c.JSON(http.StatusOK, resultObj{ResultCode:resultCode, ResultMessage:resultMessage, Results:resultData})
		//}
		c.JSON(http.StatusOK, resultObj{ResultCode:resultCode, ResultMessage:resultMessage, Results:resultData})
	}()
	data,_ := ioutil.ReadAll(c.Request.Body)
	log.Logger.Debug("Plugin request", log.String("action", action), log.String("agentType", agentType), log.String("param", string(data)))
	var param requestObj
	err := json.Unmarshal(data, &param)
	if err != nil {
		resultCode = "1"
		resultMessage = mid.GetMessageMap(c).RequestJsonUnmarshalError
		return
	}
	if len(param.Inputs) == 0 {
		resultCode = "0"
		resultMessage = fmt.Sprintf(mid.GetMessageMap(c).ParamEmptyError, "inputs")
		return
	}
	for _,v := range param.Inputs {
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
		var validateMessage,endpointGuid string
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
			param.Url = v.Url
			param.Method = v.Method
		}
		if action == "register" {
			validateMessage,endpointGuid,inputErr = AgentRegister(param)
			if validateMessage != "" {
				validateMessage = fmt.Sprintf(mid.GetMessageMap(c).ParamValidateError, validateMessage)
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
			if endpointObj.Id > 0 {
				log.Logger.Debug("Export deregister endpoint", log.Int("id", endpointObj.Id), log.String("guid", endpointObj.Guid))
				inputErr = DeregisterJob(endpointObj.Guid)
				endpointGuid = endpointObj.Guid
			}
		}
		if validateMessage != "" || inputErr != nil {
			errorMessage := validateMessage
			if errorMessage == "" {
				errorMessage = fmt.Sprintf(mid.GetMessageMap(c).HandleError, inputErr.Error())
			}
			resultData.Outputs = append(resultData.Outputs, resultOutputObj{CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: errorMessage, Guid: v.Guid, MonitorKey: endpointGuid})
			resultCode = "1"
		}else{
			resultData.Outputs = append(resultData.Outputs, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:"", Guid:v.Guid, MonitorKey:endpointGuid})
		}
	}
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

	data,_ := ioutil.ReadAll(c.Request.Body)
	log.Logger.Info("", log.String("param", string(data)))
	var param requestObj
	err := json.Unmarshal(data, &param)
	if err == nil {
		if len(param.Inputs) == 0 {
			result = resultObj{ResultCode:"0", ResultMessage:fmt.Sprintf(mid.GetMessageMap(c).ParamEmptyError, "inputs")}
			log.Logger.Warn(result.ResultMessage)
			c.JSON(http.StatusOK, result)
			return
		}
		var tmpResult []resultOutputObj
		var resultMessage string
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
				resultMessage = fmt.Sprintf(mid.GetMessageMap(c).HandleError, msg)
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:fmt.Sprintf(mid.GetMessageMap(c).HandleError, msg)})
			}else{
				msg = fmt.Sprintf("%s %s:%s %s succeed", action, agentType, v.HostIp, v.Instance)
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
			}
			log.Logger.Info(msg)
		}
		result = resultObj{ResultCode:"0", ResultMessage:resultMessage, Results:resultOutput{Outputs:tmpResult}}
		log.Logger.Info("result", log.JsonObj("result", result))
		mid.ReturnData(c, result)
	}else{
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf(mid.GetMessageMap(c).ParamValidateError, err.Error())}
		log.Logger.Error("Param validate fail", log.Error(err))
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
			mid.ReturnUpdateTableError(c, "endpoint_telnet", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetEndpointTelnet(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	result,err := db.GetEndpointTelnet(guid)
	if err != nil {
		mid.ReturnQueryTableError(c, "endpoint_telnet", err)
	}else{
		mid.ReturnSuccessData(c, result)
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
		log.Logger.Error("Update endpoint business table error", log.Error(err))
		return err
	}
	err = alarm.UpdateNodeExporterBusinessConfig(hostEndpoint.Id)
	if err != nil {
		log.Logger.Error("Update business config error", log.Error(err))
	}
	return err
}