package agent

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
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
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
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
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	err := DeregisterJob(guid)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccess(c)
}

func DeregisterJob(guid string) error {
	var err error
	endpointObj := m.EndpointTable{Guid:guid}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		return fmt.Errorf("Guid:%s can not find in table ", guid)
	}
	pingExporterFlag := false
	if endpointObj.ExportType == "ping" || endpointObj.ExportType == "telnet" || endpointObj.ExportType == "http" {
		pingExporterFlag = true
	}
	if endpointObj.AddressAgent != "" && pingExporterFlag == false {
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
	log.Logger.Debug("Start delete endpoint", log.String("guid", guid))
	err = db.DeleteEndpoint(guid)
	if err != nil {
		log.Logger.Error("Delete endpoint failed", log.Error(err))
		return err
	}
	prom.DeleteSdEndpoint(guid)
	err = prom.SyncSdConfigFile(endpointObj.Step)
	if err != nil {
		log.Logger.Error("Sync service discover file error", log.Error(err))
		return err
	}
	go other.SyncConfig(0, m.SyncSdConfigDto{Guid:guid, Step:endpointObj.Step, IsRegister:false})
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
			log.Logger.Debug("", log.Int("sm length", len(sm)))
			if len(sm) > 0 {
				log.Logger.Debug("", log.String("sm0", fmt.Sprintf(" -> %s  %s  %v", sm[0].Name, sm[0].Type, sm[0].Data)))
				TransGateWayAddress = strings.Split(strings.Split(sm[0].Name, "instance=")[1], ",job")[0]
				log.Logger.Debug("", log.String("TransGateWayAddress", TransGateWayAddress))
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
			mid.ReturnUpdateTableError(c, "endpoint", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, fmt.Sprintf(mid.GetMessageMap(c).ParamValidateError, err.Error()))
	}
}

func CustomMetricPush(c *gin.Context)  {
	var param m.TransGatewayMetricDto
	if err:=c.ShouldBindJSON(&param); err==nil {
		err = db.AddCustomMetric(param)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func ReloadEndpointMetric(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	endpointObj := m.EndpointTable{Id:id}
	db.GetEndpoint(&endpointObj)
	var address string
	if endpointObj.Address == "" {
		if endpointObj.AddressAgent == "" {
			mid.ReturnHandleError(c, fmt.Sprintf("Endpoint id %d have no address", id), nil)
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
		mid.ReturnHandleError(c, "Update endpoint metric db fail", err)
	}else{
		mid.ReturnSuccess(c)
	}
}