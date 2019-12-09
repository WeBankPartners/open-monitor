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
)

type resultObj struct {
	ResultCode  string  `json:"result_code"`
	ResultMessage  string  `json:"result_message"`
}

type requestObj struct {
	RequestId  string  	`json:"requestId"`
	Inputs  []hostRequestObj  `json:"inputs"`
}

type hostRequestObj struct {
	CallbackParameter  string  `json:"callbackParameter"`
	HostIp  string  `json:"host_ip"`
	Instance  string  `json:"instance"`
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
			result = resultObj{ResultCode:"1", ResultMessage:"Param validate fail : inputs length is zero"}
			mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
			c.JSON(http.StatusBadRequest, result)
			return
		}
		var ipList,instanceList []string
		var tmpHostIp,tmpInstance string
		if strings.Contains(param.Inputs[0].HostIp, "[") || strings.Contains(param.Inputs[0].HostIp, ",") {
			tmpHostIp = strings.ReplaceAll(param.Inputs[0].HostIp, "[", "")
			tmpHostIp = strings.ReplaceAll(tmpHostIp, "]", "")
			ipList = strings.Split(tmpHostIp, ",")
			if agentType != "host" {
				tmpInstance = strings.ReplaceAll(param.Inputs[0].Instance, "]", "")
				tmpInstance = strings.ReplaceAll(tmpInstance, "]", "")
				instanceList = strings.Split(tmpInstance, ",")
			}
		}else{
			ipList = append(ipList, param.Inputs[0].HostIp)
			if agentType != "host" {
				instanceList = append(instanceList, param.Inputs[0].Instance)
			}
		}
		for i,hostIp := range ipList {
			var param m.RegisterParam
			if agentType == "host" {
				param = m.RegisterParam{Type: agentType, ExporterIp: hostIp, ExporterPort: agentPort}
			}else{
				if len(instanceList) > i {
					param = m.RegisterParam{Type: agentType, ExporterIp: hostIp, ExporterPort:agentPort, Instance:instanceList[i]}
				}else{
					param = m.RegisterParam{Type: agentType, ExporterIp: hostIp, ExporterPort:agentPort, Instance:instanceList[0]}
				}
			}
			if action == "register" {
				err = RegisterJob(param)
			}else{
				var endpointObj m.EndpointTable
				if agentType == "host" {
					endpointObj = m.EndpointTable{Ip: param.ExporterIp, ExportType: param.Type}
				}else{
					endpointObj = m.EndpointTable{Ip: param.ExporterIp, ExportType: param.Type, Name: param.Instance}
				}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id > 0 {
					err = DeregisterJob(endpointObj.Guid)
				}
			}
			if err != nil {
				result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("%s %s:%s fail,error %v",action, agentType, hostIp, err)}
				mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
				c.JSON(http.StatusInternalServerError, result)
				return
			}
		}
		result = resultObj{ResultCode:"0", ResultMessage:"Success"}
		mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
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

func isLinuxType(osType string) bool {
	linuxType := []string{"linux", "redhat", "centos", "ubuntu", "unix"}
	result := false
	for _,v := range linuxType {
		if strings.Contains(osType, v) {
			result = true
			break
		}
	}
	return result
}
