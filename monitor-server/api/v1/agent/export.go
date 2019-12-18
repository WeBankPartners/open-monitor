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
		// update table and register to consul
		for _,v := range param.Inputs {
			var param m.RegisterParam
			if agentType == "host" {
				param = m.RegisterParam{Type: agentType, ExporterIp: v.HostIp, ExporterPort: agentPort}
			}else{
				if agentType == "tomcat" && v.Port != "" {
						param = m.RegisterParam{Type: agentType, ExporterIp: v.HostIp, ExporterPort: v.Port, Instance: v.Instance}
				}else {
					param = m.RegisterParam{Type: agentType, ExporterIp: v.HostIp, ExporterPort: agentPort, Instance: v.Instance}
				}
			}
			if action == "register" {
				err = RegisterJob(param)
			}else{
				var endpointObj m.EndpointTable
				if agentType == "host" {
					endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: agentType}
				}else{
					endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: agentType, Name: v.Instance}
				}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id > 0 {
					err = DeregisterJob(endpointObj.Guid)
				}
			}
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
		// update group and sync prometheus config
		var groupTplMap = make(map[string]int)
		for _,v := range param.Inputs {
			if v.Group == "" {
				continue
			}
			_,grpObj := db.GetSingleGrp(0, v.Group)
			if grpObj.Id > 0 {
				_,tplObj := db.GetTpl(0,grpObj.Id,0)
				groupTplMap[grpObj.Name] = tplObj.Id
				var endpointObj m.EndpointTable
				if agentType == "host" {
					endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: agentType}
				}else{
					endpointObj = m.EndpointTable{Ip: v.HostIp, ExportType: agentType, Name: v.Instance}
				}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id > 0 {
					err,_ := db.UpdateGrpEndpoint(m.GrpEndpointParamNew{Grp:grpObj.Id, Endpoints:[]int{endpointObj.Id}, Operation:"add"})
					if err != nil {
						mid.LogError("register interface update group_endpoint fail ", err)
					}
				}
			}
		}
		for k,v := range groupTplMap {
			err := alarm.SaveConfigFile(v)
			if err != nil {
				mid.LogError(fmt.Sprintf("register interface update prometheus config fail , group : %s  error ", k), err)
			}
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
