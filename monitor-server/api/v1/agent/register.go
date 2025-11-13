package agent

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/cipher"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

const defaultStep = 10
const longStep = 60

var AgentManagerServer string

type returnData struct {
	endpoint        m.EndpointTable
	metricList      []string
	defaultGroup    string
	validateMessage string
	storeMetric     bool
	fetchMetric     bool
	addDefaultGroup bool
	agentManager    bool
	err             error
	extendParam     m.EndpointExtendParamObj
}

func RegisterAgentNew(c *gin.Context) {
	var param m.RegisterParamNew
	if bindErr := c.ShouldBindJSON(&param); bindErr == nil {
		if param.Password != "" {
			decodePassword, decodeErr := DecodeUIPassword(c, param.Password)
			if decodeErr != nil {
				log.Info(nil, log.LOGGER_APP, "try to decode ui password fail", zap.Error(decodeErr))
			} else {
				param.Password = decodePassword
			}
		}
		validateMessage, _, err := AgentRegister(param, mid.GetOperateUser(c))
		if validateMessage != "" {
			mid.ReturnValidateError(c, validateMessage)
			return
		}
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Register agent fail", zap.Error(err))
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, bindErr.Error())
	}
}

func InitAgentManager() {
	for _, v := range m.Config().Dependence {
		if v.Name == "agent_manager" {
			AgentManagerServer = v.Server
			break
		}
	}
	if AgentManagerServer != "" {
		param, err := db.GetAgentManager("")
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Get agent manager table fail", zap.Error(err))
		} else {
			prom.InitAgentManager(param, AgentManagerServer)
		}
		startSyncAgentManagerJob(AgentManagerServer)
	}
}

func startSyncAgentManagerJob(url string) {
	intervalSecond := 3600
	timeStartValue, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s 00:00:00", time.Now().Format("2006-01-02")), time.Local)
	time.Sleep(time.Duration(timeStartValue.Unix()+86400-time.Now().Unix()) * time.Second)
	t := time.NewTicker(time.Duration(intervalSecond) * time.Second).C
	for {
		param, tmpErr := db.GetAgentManager("")
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "Sync agent manager job fail with get config", zap.Error(tmpErr))
		} else {
			prom.DoSyncAgentManagerJob(param, url)
		}
		<-t
	}
}

func AgentRegister(param m.RegisterParamNew, operator string) (validateMessage, guid string, err error) {
	if AgentManagerServer == "" && param.AgentManager {
		return validateMessage, guid, fmt.Errorf("agent manager server not found,can not enable agent manager ")
	}
	if param.Type == "tomcat" {
		param.Type = "java"
	}
	var rData returnData
	var stepList []int
	switch param.Type {
	case "host":
		rData = hostRegister(param)
	case "mysql":
		rData = mysqlRegister(param)
	case "redis":
		rData = redisRegister(param)
	case "java":
		rData = javaRegister(param)
	case "nginx":
		rData = nginxRegister(param)
	case "ping":
		rData = pingRegister(param)
	case "telnet":
		rData = telnetRegister(param)
	case "http":
		rData = httpRegister(param)
	case "windows":
		rData = windowsRegister(param)
	case "snmp":
		rData = snmpExporterRegister(param)
	case "process":
		rData = processMonitorRegister(param)
	case "pod":
		rData = k8sPodRegister(param)
	default:
		rData = otherExporterRegister(param)
	}
	guid = rData.endpoint.Guid
	rData.endpoint.Cluster = param.Cluster
	rData.endpoint.Tags = param.Tags
	if rData.validateMessage != "" || rData.err != nil {
		return rData.validateMessage, guid, rData.err
	}
	extendString := ""
	if rData.extendParam.Enable {
		tmpExtendBytes, _ := json.Marshal(rData.extendParam)
		extendString = string(tmpExtendBytes)
	}
	stepList, err = db.UpdateEndpoint(&rData.endpoint, extendString, operator)
	if err != nil {
		return validateMessage, guid, err
	}
	// 设置pod类型与集群关联
	if param.Type == "pod" {
		// 查询集群名称是否存在
		var k8sCluster *m.KubernetesClusterTable
		if k8sCluster, err = db.GetKubernetesByName(param.KubernetesCluster); err != nil {
			return validateMessage, guid, err
		}
		if k8sCluster == nil {
			return validateMessage, guid, fmt.Errorf("agent register fail, cluster is not exist")
		}
		if err = db.AddKubernetesEndpointRel(k8sCluster.Id, guid, param.Name); err != nil {
			return
		}
	}
	if rData.fetchMetric {
		if rData.storeMetric {
			err = db.RegisterEndpointMetric(rData.endpoint.Id, rData.metricList)
			if err != nil {
				return validateMessage, guid, err
			}
		}
		err = db.SyncSdEndpointNew(stepList, rData.endpoint.Cluster, false)
		if err != nil {
			err = fmt.Errorf("Sync sd config file fail,%s ", err.Error())
			return
		}
	}
	if rData.addDefaultGroup {
		if param.DefaultGroupName != "" {
			rData.defaultGroup = param.DefaultGroupName
		}
		if rData.defaultGroup != "" {
			_, tmpErr := db.GetSimpleEndpointGroup(rData.defaultGroup)
			if tmpErr != nil {
				log.Error(nil, log.LOGGER_APP, "add default group fail", zap.String("group", rData.defaultGroup), zap.Error(err))
			} else {
				tmpErr = db.UpdateGroupEndpoint(&m.UpdateGroupEndpointParam{GroupGuid: rData.defaultGroup, EndpointGuidList: []string{rData.endpoint.Guid}}, operator, true)
				if tmpErr != nil {
					log.Error(nil, log.LOGGER_APP, "append default group endpoint fail", zap.String("group", rData.defaultGroup), zap.Error(err))
				} else {
					db.SyncPrometheusRuleFile(rData.defaultGroup, false)
				}
			}
		}
	}
	if rData.agentManager {
		var binPath, configFile string
		for _, v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				configFile = v.ConfigFile
			}
		}
		err = db.UpdateAgentManagerTable(rData.endpoint, param.User, param.Password, configFile, binPath, true)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update agent manager table fail", zap.Error(err))
		}
	}
	return validateMessage, guid, err
}

func hostRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Host ip and port can not empty"
		return result
	}
	var hostname, sysname, release, exportVersion string
	startTime := time.Now().Unix()
	err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: param.Ip, Port: param.Port, Cluster: param.Cluster, Prefix: []string{"node"}, Keyword: []string{}})
	if err != nil {
		result.err = err
		return result
	}
	if len(strList) == 0 {
		result.err = fmt.Errorf("Can't get anything from http://%s:%s/metrics ", param.Ip, param.Port)
		return result
	}
	result.endpoint.Step, err = calcStep(startTime, param.Step)
	if err != nil {
		result.err = err
		return result
	}
	for _, v := range strList {
		if strings.Contains(v, "node_uname_info{") {
			if strings.Contains(v, "nodename") {
				hostname = strings.Split(strings.Split(v, "nodename=\"")[1], "\"")[0]
			}
			if strings.Contains(v, "sysname") {
				sysname = strings.Split(strings.Split(v, "sysname=\"")[1], "\"")[0]
			}
			if strings.Contains(v, "release") {
				release = strings.Split(strings.Split(v, "release=\"")[1], "\"")[0]
			}
		}
		if strings.Contains(v, "node_exporter_build_info{") {
			exportVersion = strings.Split(strings.Split(v, "version=\"")[1], "\"")[0]
		}
	}
	result.metricList = strList
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", hostname, param.Ip, param.Type)
	result.endpoint.Name = hostname
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.OsType = sysname
	result.endpoint.EndpointVersion = release
	result.endpoint.ExportVersion = exportVersion
	result.defaultGroup = "default_host_group"
	result.addDefaultGroup = true
	result.storeMetric = true
	result.fetchMetric = true
	result.agentManager = false
	return result
}

func mysqlRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	var err error
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Mysql ip and post can not empty "
		return result
	}
	var binPath, address, configFile string
	if param.AgentManager {
		if param.User == "" || param.Password == "" {
			result.validateMessage = "Mysql user and password can not empty"
			return result
		}
		for _, v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				configFile = v.ConfigFile
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Mysql agnet bin can not found in config ")
			return result
		}
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, AgentManagerServer, configFile)
		if err != nil {
			result.err = err
			return result
		}
		result.extendParam = m.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port, User: param.User, Password: param.Password, BinPath: binPath, ConfigPath: configFile}
	}
	var mysqlVersion, exportVersion string
	if param.FetchMetric {
		tmpIp, tmpPort := param.Ip, param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: tmpIp, Port: tmpPort, Cluster: param.Cluster, Prefix: []string{"mysql", "mysqld"}, Keyword: []string{}})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) <= 30 {
			result.err = fmt.Errorf("Connect to instance get metric error, please check param ")
			return result
		}
		result.endpoint.Step, err = calcStep(startTime, param.Step)
		if err != nil {
			result.err = err
			return result
		}
		for _, v := range strList {
			if strings.HasPrefix(v, "mysql_version_info{") {
				mysqlVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
			}
			if strings.HasPrefix(v, "mysqld_exporter_build_info{") {
				exportVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
			}
		}
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.EndpointVersion = mysqlVersion
	result.endpoint.ExportType = param.Type
	result.endpoint.ExportVersion = exportVersion
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_mysql_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	result.agentManager = true
	return result
}

func redisRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	var err error
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Redis ip and post can not empty "
		return result
	}
	var binPath, address string
	if param.AgentManager {
		if param.Password == "" {
			result.validateMessage = "Redis password can not empty"
			return result
		}
		for _, v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Redis agnet bin can not found in config ")
			return result
		}
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, AgentManagerServer, "")
		if err != nil {
			result.err = err
			return result
		}
		result.extendParam = m.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port, User: param.User, Password: param.Password, BinPath: binPath}
	}
	var redisVersion, exportVersion string
	if param.FetchMetric {
		tmpIp, tmpPort := param.Ip, param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: tmpIp, Port: tmpPort, Cluster: param.Cluster, Prefix: []string{"redis"}, Keyword: []string{"redis_version", ",version"}})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) <= 30 {
			result.err = fmt.Errorf("Connect to instance get metric error, please check param ")
			return result
		}
		result.endpoint.Step, err = calcStep(startTime, param.Step)
		if err != nil {
			result.err = err
			return result
		}
		for _, v := range strList {
			if strings.Contains(v, "redis_version") {
				redisVersion = strings.Split(strings.Split(v, ",redis_version=\"")[1], "\"")[0]
			}
			if strings.Contains(v, ",version") {
				exportVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
			}
		}
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.EndpointVersion = redisVersion
	result.endpoint.ExportType = param.Type
	result.endpoint.ExportVersion = exportVersion
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_redis_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	result.agentManager = true
	return result
}

func javaRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	var err error
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Java ip and post can not empty "
		return result
	}
	var binPath, address, configFile string
	if param.AgentManager {
		for _, v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				configFile = v.ConfigFile
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Java agnet bin can not found in config ")
			return result
		}
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, AgentManagerServer, configFile)
		if err != nil {
			result.err = err
			return result
		}
		result.extendParam = m.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port, User: param.User, Password: param.Password, BinPath: binPath, ConfigPath: configFile}
	}
	var jvmVersion, exportVersion string
	if param.FetchMetric {
		tmpIp, tmpPort := param.Ip, param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: tmpIp, Port: tmpPort, Cluster: param.Cluster, Prefix: []string{"catalina", "jvm", "java", "tomcat", "process", "com"}, Keyword: []string{"version"}})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) <= 60 {
			result.err = fmt.Errorf("Connect to instance get metric error, please check param ")
			return result
		}
		result.endpoint.Step, err = calcStep(startTime, param.Step)
		if err != nil {
			result.err = err
			return result
		}
		for _, v := range strList {
			if strings.Contains(v, "jvm_info") {
				jvmVersion = strings.Split(strings.Split(v, "version=\"")[1], "\"")[0]
			}
			if strings.Contains(v, "jmx_exporter_build_info") {
				exportVersion = strings.Split(strings.Split(v, "version=\"")[1], "\"")[0]
			}
		}
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.EndpointVersion = jvmVersion
	result.endpoint.ExportType = param.Type
	result.endpoint.ExportVersion = exportVersion
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_java_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	result.agentManager = true
	return result
}

func nginxRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	var err error
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Nginx ip and post can not empty "
		return result
	}
	var binPath, address string
	if param.AgentManager {
		for _, v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Nginx agnet bin can not found in config ")
			return result
		}
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, AgentManagerServer, "")
		if err != nil {
			result.err = err
			return result
		}
		result.extendParam = m.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port, User: param.User, Password: param.Password, BinPath: binPath}
	}
	if param.FetchMetric {
		tmpIp, tmpPort := param.Ip, param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: tmpIp, Port: tmpPort, Cluster: param.Cluster, Prefix: []string{}, Keyword: []string{}})
		if err != nil {
			result.err = err
			return result
		}
		result.endpoint.Step, err = calcStep(startTime, param.Step)
		if err != nil {
			result.err = err
			return result
		}
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.EndpointVersion = ""
	result.endpoint.ExportType = param.Type
	result.endpoint.ExportVersion = ""
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_nginx_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	result.agentManager = true
	return result
}

func pingRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" {
		result.validateMessage = "Ping ip can not empty "
		return result
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	tmpPort, _ := strconv.Atoi(param.Port)
	if tmpPort <= 0 {
		tmpPort = 9191
	}
	result.endpoint.Address = fmt.Sprintf("%s:%d", param.Ip, tmpPort)
	result.endpoint.ExportType = param.Type
	result.endpoint.Step = defaultStep
	result.defaultGroup = "default_ping_group"
	result.addDefaultGroup = true
	result.agentManager = false
	if param.ExportAddress != "" {
		param.ExportAddress = formatExportAddress(param.ExportAddress)
		result.endpoint.AddressAgent = param.ExportAddress
		result.fetchMetric = true
		result.extendParam = m.EndpointExtendParamObj{Enable: true, ExportAddress: param.ExportAddress}
	}
	return result
}

func telnetRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" {
		result.validateMessage = "Telnet ip can not empty "
		return result
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.ExportType = param.Type
	result.endpoint.Step = defaultStep
	if param.ExportAddress != "" {
		param.ExportAddress = formatExportAddress(param.ExportAddress)
		result.endpoint.AddressAgent = param.ExportAddress
		result.fetchMetric = true
		result.extendParam = m.EndpointExtendParamObj{Enable: true, ExportAddress: param.ExportAddress}
	}
	result.defaultGroup = "default_telnet_group"
	result.addDefaultGroup = true
	result.agentManager = false
	// store to db -> endpoint_telnet
	var eto []*m.EndpointTelnetObj
	eto = append(eto, &m.EndpointTelnetObj{Port: param.Port, Note: ""})
	err := db.UpdateEndpointTelnet(m.UpdateEndpointTelnetParam{Guid: result.endpoint.Guid, Config: eto})
	if err != nil {
		result.err = err
	}
	result.extendParam = m.EndpointExtendParamObj{Enable: true, Ip: param.Ip, Port: param.Port}
	return result
}

func httpRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" || param.Url == "" || param.Method == "" {
		result.validateMessage = "Http check ip/url/method can not empty "
		return result
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.ExportType = param.Type
	result.endpoint.Step = defaultStep
	result.extendParam = m.EndpointExtendParamObj{Enable: true, HttpMethod: param.Method, HttpUrl: param.Url}
	if param.ExportAddress != "" {
		param.ExportAddress = formatExportAddress(param.ExportAddress)
		result.endpoint.AddressAgent = param.ExportAddress
		result.fetchMetric = true
		result.extendParam.ExportAddress = param.ExportAddress
	}
	result.defaultGroup = "default_http_group"
	result.addDefaultGroup = true
	result.agentManager = false
	var eho []*m.EndpointHttpTable
	eho = append(eho, &m.EndpointHttpTable{EndpointGuid: result.endpoint.Guid, Url: param.Url, Method: param.Method})
	err := db.UpdateEndpointHttp(eho)
	if err != nil {
		result.err = err
	}
	return result
}

func windowsRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Windows exporter ip and port can not empty"
		return result
	}
	var hostname, sysname, release string
	if param.FetchMetric {
		startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: param.Ip, Port: param.Port, Cluster: param.Cluster, Prefix: []string{"wmi"}, Keyword: []string{}})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) == 0 {
			result.err = fmt.Errorf("Can't get anything from http://%s:%d/metrics ", param.Ip, &param.Port)
			return result
		}
		result.endpoint.Step, err = calcStep(startTime, param.Step)
		if err != nil {
			result.err = err
			return result
		}
		for _, v := range strList {
			if strings.Contains(v, "wmi_cs_hostname{") {
				hostname = strings.Split(strings.Split(v, "hostname=\"")[1], "\"")[0]
			}
			if strings.Contains(v, "wmi_os_info") {
				sysname = strings.Split(strings.Split(v, "product=\"")[1], "\"")[0]
				release = strings.Split(strings.Split(v, "version=\"")[1], "\"")[0]
			}
		}
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", hostname, param.Ip, param.Type)
	result.endpoint.Name = hostname
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.OsType = sysname
	result.endpoint.Step = defaultStep
	result.endpoint.EndpointVersion = release
	result.defaultGroup = "default_windows_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	result.agentManager = false
	return result
}

func snmpExporterRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if param.Ip == "" {
		result.validateMessage = "Snmp target ip can not empty"
		return result
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = param.Ip
	result.defaultGroup = "default_snmp_group"
	result.addDefaultGroup = true
	result.storeMetric = false
	result.fetchMetric = false
	result.agentManager = false
	err := db.SnmpEndpointAdd(param.ProxyExporter, result.endpoint.Guid, param.Ip)
	if err != nil {
		result.err = err
	}
	result.extendParam = m.EndpointExtendParamObj{Enable: true, ProxyExporter: param.ProxyExporter}
	return result
}

func processMonitorRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	if param.Ip == "" {
		result.validateMessage = "Process host ip can not empty"
		return result
	}
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	endpointObj := m.EndpointTable{Ip: param.Ip, ExportType: "host"}
	db.GetEndpoint(&endpointObj)
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = endpointObj.Address
	result.defaultGroup = "default_process_group"
	result.addDefaultGroup = true
	result.storeMetric = false
	result.fetchMetric = false
	result.agentManager = false
	result.extendParam = m.EndpointExtendParamObj{Enable: true, ProcessName: param.ProcessName, ProcessTags: param.Tags}
	newEndpointObj := m.EndpointNewTable{Guid: result.endpoint.Guid, Name: result.endpoint.Name, Ip: result.endpoint.Ip, MonitorType: result.endpoint.ExportType, AgentAddress: result.endpoint.Address}
	b, _ := json.Marshal(result.extendParam)
	newEndpointObj.ExtendParam = string(b)
	err := db.SyncNodeExporterProcessConfig(param.Ip, []*m.EndpointNewTable{&newEndpointObj}, false)
	if err != nil {
		result.err = err
	}
	return result
}

func k8sPodRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = param.Step
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if strings.Contains(param.Type, "_") {
		result.validateMessage = "ExporterType illegal "
		return result
	}
	if param.Ip == "" {
		result.validateMessage = "Default ip can not empty "
		return result
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.extendParam.Enable = true
	result.extendParam.NodeIp = param.NodeIp
	result.agentManager = false
	return result
}

func otherExporterRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = param.Step
	if mid.IsIllegalName(param.Name) {
		result.validateMessage = "param instance name illegal"
		return result
	}
	if strings.Contains(param.Type, "_") {
		result.validateMessage = "ExporterType illegal "
		return result
	}
	if param.Ip == "" {
		result.validateMessage = "Default ip can not empty "
		return result
	}
	if param.FetchMetric {
		if param.Port == "" {
			result.validateMessage = "Default endpoint port can not empty if you want to get exporter metric "
			return result
		}
		//startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: param.Ip, Port: param.Port, Cluster: param.Cluster, Prefix: []string{}, Keyword: []string{}})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) == 0 {
			result.err = fmt.Errorf("Can't get anything from http://%s:%d/metrics ", param.Ip, &param.Port)
			return result
		}
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.fetchMetric = true
	if param.Type == "pod" {
		result.extendParam.Enable = true
		result.endpoint.Address = ""
		result.fetchMetric = false
	}
	result.agentManager = false
	return result
}

func calcStep(startTime int64, paramStep int) (step int, err error) {
	subTime := int(time.Now().Unix() - startTime)
	if subTime > defaultStep {
		if subTime > longStep {
			err = fmt.Errorf("get exporter data use too many time:%d seconds", subTime)
		} else {
			if paramStep > subTime {
				step = paramStep
			} else {
				step = longStep
			}
		}
	} else {
		if paramStep > 0 {
			step = paramStep
		} else {
			step = defaultStep
		}
	}
	log.Debug(nil, log.LOGGER_APP, "Calc step", zap.Int("step", step))
	return step, err
}

func formatExportAddress(input string) string {
	result := input
	if input == "" {
		return result
	}
	if strings.HasPrefix(result, "http://") {
		result = strings.ReplaceAll(result, "http://", "")
	}
	if strings.Contains(result, "/") {
		result = result[:strings.Index(result, "/")]
	}
	if !strings.Contains(result, ":") {
		result = ""
	}
	return result
}

func DecodeUIPassword(ctx context.Context, inputValue string) (output string, err error) {
	if inputValue == "" {
		return
	}
	seed := m.Config().EncryptSeed
	if pwdBytes, pwdErr := base64.StdEncoding.DecodeString(inputValue); pwdErr == nil {
		inputValue = hex.EncodeToString(pwdBytes)
	} else {
		err = fmt.Errorf("base64 decode input data fail,%s ", pwdErr.Error())
		return
	}
	output, err = decodeUIAesPassword(seed, inputValue)
	return
}

func decodeUIAesPassword(seed, password string) (decodePwd string, err error) {
	unixTime := time.Now().Unix() / 100
	decodePwd, err = cipher.AesDePasswordWithIV(seed, password, fmt.Sprintf("%d", unixTime*100000000))
	if err != nil {
		unixTime = unixTime - 1
		decodePwd, err = cipher.AesDePasswordWithIV(seed, password, fmt.Sprintf("%d", unixTime*100000000))
	}
	if err != nil {
		err = fmt.Errorf("aes decode with iv fail,%s ", err.Error())
	}
	return
}
