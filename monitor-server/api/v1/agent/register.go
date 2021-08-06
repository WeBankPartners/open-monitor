package agent

import (
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

const defaultStep = 10
const longStep = 60

var agentManagerServer string

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
}

func RegisterAgentNew(c *gin.Context) {
	var param m.RegisterParamNew
	if err := c.ShouldBindJSON(&param); err == nil {
		validateMessage, _, err := AgentRegister(param)
		if validateMessage != "" {
			mid.ReturnValidateError(c, validateMessage)
			return
		}
		if err != nil {
			log.Logger.Error("Register agent fail", log.Error(err))
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func InitAgentManager() {
	for _, v := range m.Config().Dependence {
		if v.Name == "agent_manager" {
			agentManagerServer = v.Server
			break
		}
	}
	if agentManagerServer != "" {
		param, err := db.GetAgentManager("")
		if err != nil {
			log.Logger.Error("Get agent manager table fail", log.Error(err))
			return
		}
		go prom.InitAgentManager(param, agentManagerServer)
		go prom.StartSyncAgentManagerJob(param, agentManagerServer)
	}
}

func AgentRegister(param m.RegisterParamNew) (validateMessage, guid string, err error) {
	if agentManagerServer == "" && param.AgentManager {
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
	default:
		rData = otherExporterRegister(param)
	}
	guid = rData.endpoint.Guid
	rData.endpoint.Cluster = param.Cluster
	if rData.validateMessage != "" || rData.err != nil {
		return rData.validateMessage, guid, rData.err
	}
	stepList, err = db.UpdateEndpoint(&rData.endpoint)
	if err != nil {
		return validateMessage, guid, err
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
			err, grpObj := db.GetSingleGrp(0, rData.defaultGroup)
			if err != nil || grpObj.Id <= 0 {
				return validateMessage, guid, fmt.Errorf("Add group %s fail,id:%d err:%v ", rData.defaultGroup, grpObj.Id, err)
			}
			err, _ = db.UpdateGrpEndpoint(m.GrpEndpointParamNew{Grp: grpObj.Id, Endpoints: []int{rData.endpoint.Id}, Operation: "add"})
			if err != nil {
				return validateMessage, guid, err
			}
			_, tplObj := db.GetTpl(0, grpObj.Id, 0)
			if tplObj.Id > 0 {
				//err := alarm.SaveConfigFile(tplObj.Id, false)
				err := db.SyncRuleConfigFile(tplObj.Id, []string{}, false)
				if err != nil {
					log.Logger.Error("Register interface update prometheus config fail", log.String("group", rData.defaultGroup), log.Error(err))
					return validateMessage, guid, err
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
			log.Logger.Error("Update agent manager table fail", log.Error(err))
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
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Mysql instance name and ip and post can not empty "
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
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, agentManagerServer, configFile)
		if err != nil {
			result.err = err
			return result
		}
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
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Redis instance name and ip and post can not empty "
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
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, agentManagerServer, "")
		if err != nil {
			result.err = err
			return result
		}
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
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Java instance name and ip and post can not empty "
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
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, agentManagerServer, configFile)
		if err != nil {
			result.err = err
			return result
		}
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
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Nginx instance name and ip and post can not empty "
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
		address, err = prom.DeployAgent(param.Type, param.Name, binPath, param.Ip, param.Port, param.User, param.Password, agentManagerServer, "")
		if err != nil {
			result.err = err
			return result
		}
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
	if param.Name == "" || param.Ip == "" {
		result.validateMessage = "Ping instance name and ip can not empty "
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
	}
	return result
}

func telnetRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if param.Name == "" || param.Ip == "" {
		result.validateMessage = "Telnet instance name and ip can not empty "
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
	return result
}

func httpRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if param.Name == "" || param.Ip == "" || param.Url == "" || param.Method == "" {
		result.validateMessage = "Http check name/ip/url/method can not empty "
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
	if param.Ip == "" {
		result.validateMessage = "Snmp target ip can not empty"
		return result
	}
	if param.Name == "" {
		result.validateMessage = "Snmp target name can not empty"
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
	return result
}

func otherExporterRegister(param m.RegisterParamNew) returnData {
	var result returnData
	result.endpoint.Step = defaultStep
	if param.Name == "" || param.Ip == "" {
		result.validateMessage = "Default endpoint name and ip can not empty "
		return result
	}
	if param.FetchMetric {
		if param.Port == "" {
			result.validateMessage = "Default endpoint port can not empty if you want to get exporter metric "
			return result
		}
		startTime := time.Now().Unix()
		err, strList := db.QueryExporterMetric(m.QueryPrometheusMetricParam{Ip: param.Ip, Port: param.Port, Cluster: param.Cluster, Prefix: []string{}, Keyword: []string{}})
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
		result.metricList = strList
	}
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Name, param.Ip, param.Type)
	result.endpoint.Name = param.Name
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.fetchMetric = true
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
	log.Logger.Debug("Calc step", log.Int("step", step))
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
