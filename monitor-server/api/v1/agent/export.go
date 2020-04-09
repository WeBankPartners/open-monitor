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
}

func ExportAgent(c *gin.Context)  {
	agentType := c.Param("name")
	action := "register"
	if strings.Contains(c.Request.URL.String(), "deregister") {
		action = "deregister"
	}
	var agentPort string
	var result resultObj
	illegal := true
	for _,v := range m.Config().Agent {
		tmpAgentType := agentType
		if tmpAgentType == "java" {
			tmpAgentType = "tomcat"
		}
		if v.AgentType == tmpAgentType {
			illegal = false
			agentPort = v.Port
			break
		}
	}
	if agentType == "other" {
		illegal = false
	}
	if illegal {
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("No such monitor type like %s", agentType)}
		mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
		c.JSON(http.StatusBadRequest, result)
		return
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
		if action == "register" && agentType != "other" {
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
	illegal := true
	for _,v := range m.Config().Agent {
		if v.AgentType == agentType {
			illegal = false
			agentPort = v.Port
			break
		}
	}
	if illegal {
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("No such monitor type like %s", agentType)}
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
		for _,v := range param.Inputs {
			if agentType == "tomcat" && v.Port != "" {
				agentPort = v.Port
			}
			err := db.UpdateEndpointAlarmFlag(isStop,agentType,v.Instance,v.HostIp,agentPort)
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