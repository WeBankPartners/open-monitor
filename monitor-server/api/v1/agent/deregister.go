package agent

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"strings"
	"fmt"
	"strconv"
	"time"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
)

const hostType  = "host"
const mysqlType  = "mysql"
const redisType = "redis"
const tomcatType = "tomcat"
const javaType = "java"
const otherType = "other"

func DeregisterAgent(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "Guid can not be empty")
		return
	}
	err := DeregisterJob(guid)
	if err != nil {
		mid.ReturnError(c, fmt.Sprintf("Delete endpint %s failed", guid),err)
		return
	}
	mid.ReturnSuccess(c, fmt.Sprintf("Deregister %s successfully", guid))
}

func DeregisterJob(guid string) error {
	var err error
	endpointObj := m.EndpointTable{Guid:guid}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		return fmt.Errorf("Guid:%s can not find in table ", guid)
	}
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
			if err != nil {
				return fmt.Errorf("deregister endpoint:%s stop agent error: %v ", guid, err)
			}
		}
	}
	mid.LogInfo(fmt.Sprintf("start delete endpoint:%s ", guid))
	err = db.DeleteEndpoint(guid)
	if err != nil {
		mid.LogError(fmt.Sprintf("Delete endpint %s failed", guid), err)
		return err
	}

	mid.LogInfo(fmt.Sprintf("delete endpoint:%s step:%d", guid, endpointObj.Step))
	prom.DeleteSdEndpoint(guid)
	err = prom.SyncSdConfigFile(endpointObj.Step)
	if err != nil {
		mid.LogError("sync service discover file error: ", err)
		return err
	}

	db.UpdateAgentManagerTable(m.EndpointTable{Guid:guid}, "", "", "", "", false)
	return err
}

var TransGateWayAddress string

func CustomRegister(c *gin.Context)  {
	var param m.TransGatewayRequestDto
	if err:=c.ShouldBindJSON(&param); err==nil {
		if TransGateWayAddress == "" {
			query := m.QueryMonitorData{Start:time.Now().Unix()-60, End:time.Now().Unix(), Endpoint:[]string{"endpoint"}, Metric:[]string{"metric"}, PromQ:"up{job=\"transgateway\"}", Legend:"$custom_all"}
			sm := datasource.PrometheusData(&query)
			mid.LogInfo(fmt.Sprintf("sm length : %d ", len(sm)))
			if len(sm) > 0 {
				mid.LogInfo(fmt.Sprintf("sm0 -> %s  %s  %v", sm[0].Name, sm[0].Type, sm[0].Data))
				TransGateWayAddress = strings.Split(strings.Split(sm[0].Name, "instance=")[1], ",job")[0]
				mid.LogInfo(fmt.Sprintf("TransGateWayAddress : %s", TransGateWayAddress))
			}
		}
		var endpointObj m.EndpointTable
		endpointObj.Guid = fmt.Sprintf("%s_%s_custom", param.Name, param.HostIp)
		endpointObj.Address = TransGateWayAddress
		endpointObj.Name = param.Name
		endpointObj.Ip = param.HostIp
		endpointObj.ExportType = "custom"
		endpointObj.Step = 10
		err := db.UpdateEndpoint(&endpointObj)
		if err != nil {
			mid.ReturnError(c, fmt.Sprintf("Update endpoint %s_%s_custom fail", param.Name, param.HostIp), err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validate fail %v", err))
	}
}

func CustomMetricPush(c *gin.Context)  {
	var param m.TransGatewayMetricDto
	if err:=c.ShouldBindJSON(&param); err==nil {
		err = db.AddCustomMetric(param)
		if err != nil {
			mid.LogError("Add custom metric fail", err)
			mid.ReturnError(c, "Add custom metric fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validate fail %v", err))
	}
}

func ReloadEndpointMetric(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		mid.ReturnValidateFail(c, "Param id validate fail")
		return
	}
	endpointObj := m.EndpointTable{Id:id}
	db.GetEndpoint(&endpointObj)
	var address string
	if endpointObj.Address == "" {
		if endpointObj.AddressAgent == "" {
			mid.ReturnError(c, fmt.Sprintf("Endpoint id %d have no address", id), nil)
			return
		}
		address = endpointObj.AddressAgent
	}else{
		address = endpointObj.Address
	}
	tmpExporterIp := strings.Split(address, ":")[0]
	tmpExporterPort := strings.Split(address, ":")[1]
	var strList []string
	if endpointObj.ExportType == hostType {
		_, strList = prom.GetEndpointData(tmpExporterIp, tmpExporterPort, []string{"node"}, []string{})
	}else if endpointObj.ExportType == mysqlType {
		_, strList = prom.GetEndpointData(tmpExporterIp, tmpExporterPort, []string{"mysql", "mysqld"}, []string{})
	}else if endpointObj.ExportType == redisType {
		_, strList = prom.GetEndpointData(tmpExporterIp, tmpExporterPort, []string{"redis"}, []string{"redis_version", ",version"})
	}else if endpointObj.ExportType == tomcatType {
		_, strList = prom.GetEndpointData(tmpExporterIp, tmpExporterPort, []string{"Catalina", "catalina", "jvm", "java"}, []string{"version"})
	}else{
		_, strList = prom.GetEndpointData(tmpExporterIp, tmpExporterPort, []string{}, []string{""})
	}
	err := db.RegisterEndpointMetric(id, strList)
	if err != nil {
		mid.ReturnError(c, "Update endpoint metric db fail", err)
	}else{
		mid.ReturnSuccess(c, "Success")
	}
}