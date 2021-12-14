package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
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
	guid := c.Param("guid")
	endpointObj,queryErr := db.GetEndpointNew(&models.EndpointNewTable{Guid: guid})
	if queryErr != nil {
		err = queryErr
		return
	}
	if endpointObj.Step != param.Step {
		err = db.SyncSdEndpointNew([]int{endpointObj.Step,param.Step}, endpointObj.Cluster, false)
		if err != nil {
			return
		}
	}
	var newEndpoint models.EndpointNewTable
	switch param.Type {
	case "host":
		newEndpoint,err = hostEndpointUpdate(param)
	case "mysql":
		newEndpoint,err = mysqlEndpointUpdate(param)
	case "redis":
		newEndpoint,err = redisEndpointUpdate(param)
	case "java":
		newEndpoint,err = javaEndpointUpdate(param)
	case "nginx":
		newEndpoint,err = nginxEndpointUpdate(param)
	case "ping":
		newEndpoint,err = pingEndpointUpdate(param)
	case "telnet":
		newEndpoint,err = telnetEndpointUpdate(param)
	case "http":
		newEndpoint,err = httpEndpointUpdate(param)
	case "windows":
		newEndpoint,err = windowsEndpointUpdate(param)
	case "snmp":
		newEndpoint,err = snmpEndpointUpdate(param)
	case "process":
		newEndpoint,err = processEndpointUpdate(param)
	default:
		newEndpoint = models.EndpointNewTable{Guid: param.Guid}
	}
	if err != nil {
		return
	}
	log.Logger.Info("new endpoint", log.JsonObj("endpoint", newEndpoint))
}

func DeleteEndpoint(c *gin.Context)  {

}

func hostEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func mysqlEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func redisEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func javaEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func nginxEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func processEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func windowsEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func pingEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func telnetEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func httpEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}

func snmpEndpointUpdate(models.RegisterParamNew) (newEndpoint models.EndpointNewTable,err error) {
	return
}