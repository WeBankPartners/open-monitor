package agent

import (
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

const hostType = "host"
const mysqlType = "mysql"
const redisType = "redis"
const tomcatType = "tomcat"
const javaType = "java"
const otherType = "other"

func DeregisterAgent(c *gin.Context) {
	var param m.EndpointTable
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	endpointObj := m.EndpointTable{Guid: param.Guid}
	err := db.GetEndpoint(&endpointObj)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	if endpointObj.ExportType == "host" {
		processList, getErr := db.GetProcessByHostEndpoint(endpointObj.Ip)
		if getErr != nil {
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		if len(processList) > 0 {
			mid.ReturnServerHandleError(c, m.GetMessageMap(c).EndpointHostDeleteError)
			return
		}
	}
	err = DeregisterJob(endpointObj, mid.GetOperateUser(c))
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccess(c)
}

func DeregisterJob(endpointObj m.EndpointTable, operator string) error {
	var err error
	guid := endpointObj.Guid
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
	// Remove from group
	affectTplList, deleteErr := db.DeleteEndpointFromGroup(endpointObj.Id)
	if deleteErr != nil {
		return deleteErr
	}

	log.Debug(nil, log.LOGGER_APP, "Start delete endpoint", zap.String("guid", guid))
	err = db.DeleteEndpoint(guid, operator)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Delete endpoint failed", zap.Error(err))
		return err
	}
	// Update sd file
	err = db.SyncSdEndpointNew([]int{endpointObj.Step}, endpointObj.Cluster, false)
	if err != nil {
		return fmt.Errorf("Sync sd config fail,%s ", err.Error())
	}
	// Update rule file
	tplObj, _ := db.GetTemplateObject(0, 0, endpointObj.Id)
	if tplObj.Id > 0 {
		affectTplList = append(affectTplList, tplObj.Id)
	}
	for _, tplId := range affectTplList {
		tmpErr := db.SyncRuleConfigFile(tplId, []string{endpointObj.Guid}, false)
		if tmpErr != nil {
			err = fmt.Errorf("Sync rule config fail,%s ", tmpErr.Error())
			break
		}
	}
	if err != nil {
		return err
	}
	// If deregistering a process monitor endpoint, sync node_exporter process config for the host
	// to ensure removed processes are cleared from agent-side cache/config immediately.
	if endpointObj.ExportType == "process" {
		if syncErr := db.SyncNodeExporterProcessConfig(endpointObj.Ip, []*m.EndpointNewTable{}, false); syncErr != nil {
			log.Error(nil, log.LOGGER_APP, "Sync process config after deregister fail", zap.Error(syncErr), zap.String("ip", endpointObj.Ip))
		}
	}
	// 删除pod 和集群关系
	if endpointObj.ExportType == "pod" {
		err = db.DeleteKubernetesEndpointRelByEndpointId(endpointObj.Guid)
	}
	if endpointObj.ExportType == "snmp" {
		if err = db.SnmpEndpointDelete(endpointObj.Guid); err != nil {
			return err
		}
	}
	if endpointObj.AddressAgent != "" {
		if err = db.UpdateAgentManagerTable(m.EndpointTable{Guid: guid}, "", "", "", "", false); err != nil {
			return err
		}
	}
	return err
}

var TransGateWayAddress string

func CustomRegister(c *gin.Context) {
	var param m.TransGatewayRequestDto
	if err := c.ShouldBindJSON(&param); err == nil {
		if TransGateWayAddress == "" {
			query := m.QueryMonitorData{Start: time.Now().Unix() - 60, End: time.Now().Unix(), Endpoint: []string{"endpoint"}, Metric: []string{"metric"}, PromQ: "up{job=\"transgateway\"}", Legend: "$custom_all"}
			sm := datasource.PrometheusData(&query)
			log.Debug(nil, log.LOGGER_APP, "", zap.Int("sm length", len(sm)))
			if len(sm) > 0 {
				log.Debug(nil, log.LOGGER_APP, "", zap.String("sm0", fmt.Sprintf(" -> %s  %s  %v", sm[0].Name, sm[0].Type, sm[0].Data)))
				TransGateWayAddress = strings.Split(strings.Split(sm[0].Name, "instance=")[1], ",job")[0]
				TransGateWayAddress = strings.ReplaceAll(TransGateWayAddress, "{", "")
				TransGateWayAddress = strings.ReplaceAll(TransGateWayAddress, "}", "")
				log.Debug(nil, log.LOGGER_APP, "", zap.String("TransGateWayAddress", TransGateWayAddress))
			}
		}
		var endpointObj m.EndpointTable
		endpointObj.Guid = fmt.Sprintf("%s_%s_custom", param.Name, param.HostIp)
		endpointObj.Address = TransGateWayAddress
		endpointObj.Name = param.Name
		endpointObj.Ip = param.HostIp
		endpointObj.ExportType = "custom"
		endpointObj.Step = 10
		_, err := db.UpdateEndpoint(&endpointObj, "", mid.GetOperateUser(c))
		if err != nil {
			mid.ReturnUpdateTableError(c, "endpoint", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func CustomMetricPush(c *gin.Context) {
	var param m.TransGatewayMetricDto
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.AddCustomMetric(param)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func ReloadEndpointMetric(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	endpointObj := m.EndpointTable{Id: id}
	db.GetEndpoint(&endpointObj)
	var address string
	if endpointObj.Address == "" {
		if endpointObj.AddressAgent == "" {
			mid.ReturnHandleError(c, fmt.Sprintf("Endpoint id %d have no address", id), nil)
			return
		}
		address = endpointObj.AddressAgent
	} else {
		address = endpointObj.Address
	}
	tmpExporterIp := strings.Split(address, ":")[0]
	tmpExporterPort := strings.Split(address, ":")[1]
	var strList []string
	getEndpointParam := m.QueryPrometheusMetricParam{Ip: tmpExporterIp, Port: tmpExporterPort, Cluster: endpointObj.Cluster, Prefix: []string{}, Keyword: []string{}}
	if endpointObj.ExportType == hostType {
		getEndpointParam.Prefix = []string{"node"}
		_, strList = db.QueryExporterMetric(getEndpointParam)
	} else if endpointObj.ExportType == mysqlType {
		getEndpointParam.Prefix = []string{"mysql", "mysqld"}
		_, strList = db.QueryExporterMetric(getEndpointParam)
	} else if endpointObj.ExportType == redisType {
		getEndpointParam.Prefix = []string{"redis"}
		getEndpointParam.Keyword = []string{"redis_version", ",version"}
		_, strList = db.QueryExporterMetric(getEndpointParam)
	} else if endpointObj.ExportType == tomcatType {
		getEndpointParam.Prefix = []string{"Catalina", "catalina", "jvm", "java"}
		getEndpointParam.Keyword = []string{"version"}
		_, strList = db.QueryExporterMetric(getEndpointParam)
	} else {
		_, strList = db.QueryExporterMetric(getEndpointParam)
	}
	err := db.RegisterEndpointMetric(id, strList)
	if err != nil {
		mid.ReturnHandleError(c, "Update endpoint metric db fail", err)
	} else {
		mid.ReturnSuccess(c)
	}
}
