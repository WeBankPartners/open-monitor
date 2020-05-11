package agent

import (
	"fmt"
	"strings"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

const prometheusStep = 10

var agentManagerServer string

type returnData struct {
	endpoint  m.EndpointTable
	metricList  []string
	defaultGroup  string
	validateMessage  string
	storeMetric bool
	fetchMetric bool
	addDefaultGroup bool
	err  error
}

func RegisterAgentNew(c *gin.Context)  {
	var param m.RegisterParamNew
	if err := c.ShouldBindJSON(&param); err==nil {
		validateMessage,err := AgentRegister(param)
		if validateMessage != "" {
			mid.ReturnValidateFail(c, validateMessage)
			return
		}
		if err != nil {
			mid.LogError("Register agent fail ", err)
			mid.ReturnError(c, "Register agent error ", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func InitAgentManager()  {
	for _, v := range m.Config().Dependence {
		if v.Name == "agent_manager" {
			agentManagerServer = v.Server
			break
		}
	}
}

func AgentRegister(param m.RegisterParamNew) (validateMessage string,err error) {
	if agentManagerServer == "" && param.AgentManager {
		return validateMessage,fmt.Errorf("agent manager server not found,can not enable agent manager ")
	}
	var rData returnData
	switch param.Type {
		case "host": rData = hostRegister(param)
		case "mysql": rData = mysqlRegister(param)
		case "redis": rData = redisRegister(param)
		case "java": rData = javaRegister(param)
		case "ping": rData = pingRegister(param)
		case "telnet": rData = telnetRegister(param)
		case "http": rData = httpRegister(param)
		case "windows": rData = windowsRegister(param)
		default: rData = otherExporterRegister(param)
	}
	if rData.validateMessage != "" || rData.err != nil {
		return rData.validateMessage,rData.err
	}
	err = db.UpdateEndpoint(&rData.endpoint)
	if err != nil {
		return validateMessage,err
	}
	if rData.fetchMetric {
		if rData.storeMetric {
			err = db.RegisterEndpointMetric(rData.endpoint.Id, rData.metricList)
			if err != nil {
				return validateMessage, err
			}
		}
		tmpIp,tmpPort := param.Ip,param.Port
		if strings.Contains(rData.endpoint.AddressAgent, ":") {
			tmpIp = rData.endpoint.AddressAgent[:strings.Index(rData.endpoint.AddressAgent, ":")]
			tmpPort = rData.endpoint.AddressAgent[strings.Index(rData.endpoint.AddressAgent, ":")+1:]
		}
		err = prom.RegisteConsul(rData.endpoint.Guid, tmpIp, tmpPort, []string{param.Type}, prometheusStep, false)
		if err != nil {
			return validateMessage,err
		}
	}
	if rData.addDefaultGroup {
		if param.DefaultGroupName != "" {
			rData.defaultGroup = param.DefaultGroupName
		}
		if rData.defaultGroup != "" {
			err, grpObj := db.GetSingleGrp(0, rData.defaultGroup)
			if err != nil || grpObj.Id <= 0 {
				return validateMessage, fmt.Errorf("Add group %s fail,id:%d err:%v ", rData.defaultGroup, grpObj.Id, err)
			}
			err, _ = db.UpdateGrpEndpoint(m.GrpEndpointParamNew{Grp: grpObj.Id, Endpoints: []int{rData.endpoint.Id}, Operation: "add"})
			if err != nil {
				return validateMessage,err
			}
		}
	}
	return validateMessage,err
}

func hostRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Host ip and port can not empty"
		return result
	}
	var hostname,sysname,release,exportVersion string
	err,strList := prom.GetEndpointData(param.Ip, param.Port, []string{"node"}, []string{})
	if err != nil {
		result.err = err
		return result
	}
	if len(strList) == 0 {
		result.err = fmt.Errorf("Can't get anything from http://%s:%d/metrics ", param.Ip, &param.Port)
		return result
	}
	for _,v := range strList {
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
			exportVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
		}
	}
	result.metricList = strList
	result.endpoint.Guid = fmt.Sprintf("%s_%s_%s", hostname, param.Ip, param.Type)
	result.endpoint.Name = hostname
	result.endpoint.Ip = param.Ip
	result.endpoint.ExportType = param.Type
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.OsType = sysname
	result.endpoint.Step = prometheusStep
	result.endpoint.EndpointVersion = release
	result.endpoint.ExportVersion = exportVersion
	result.defaultGroup = "default_host_group"
	result.addDefaultGroup = true
	result.storeMetric = true
	result.fetchMetric = true
	return result
}

func mysqlRegister(param m.RegisterParamNew) returnData {
	var result returnData
	var err error
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Mysql instance name and ip and post can not empty "
		return result
	}
	var binPath,address string
	if param.AgentManager {
		if param.User == "" || param.Password == "" {
			result.validateMessage = "Mysql user and password can not empty"
			return result
		}
		for _,v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Mysql agnet bin can not found in config ")
			return result
		}
		address,err = prom.DeployAgent(param.Type,param.Name,binPath,param.Ip,param.Port,param.User,param.Password,agentManagerServer)
		if err != nil {
			result.err = err
			return result
		}
	}
	var mysqlVersion,exportVersion string
	if param.FetchMetric {
		tmpIp,tmpPort := param.Ip,param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		err, strList := prom.GetEndpointData(tmpIp, tmpPort, []string{"mysql", "mysqld"}, []string{})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) <= 30 {
			result.err = fmt.Errorf("Connect to instance get metric error, please check param ")
			return result
		}
		for _,v := range strList {
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
	result.endpoint.Step = prometheusStep
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_mysql_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	return result
}

func redisRegister(param m.RegisterParamNew) returnData {
	var result returnData
	var err error
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Redis instance name and ip and post can not empty "
		return result
	}
	var binPath,address string
	if param.AgentManager {
		if param.Password == "" {
			result.validateMessage = "Redis password can not empty"
			return result
		}
		for _,v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Redis agnet bin can not found in config ")
			return result
		}
		address,err = prom.DeployAgent(param.Type,param.Name,binPath,param.Ip,param.Port,param.User,param.Password,agentManagerServer)
		if err != nil {
			result.err = err
			return result
		}
	}
	var redisVersion,exportVersion string
	if param.FetchMetric {
		tmpIp,tmpPort := param.Ip,param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		err, strList := prom.GetEndpointData(tmpIp, tmpPort, []string{"redis"}, []string{"redis_version", ",version"})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) <= 30 {
			result.err = fmt.Errorf("Connect to instance get metric error, please check param ")
			return result
		}
		for _,v := range strList {
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
	result.endpoint.Step = prometheusStep
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_redis_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	return result
}

func javaRegister(param m.RegisterParamNew) returnData {
	var result returnData
	var err error
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Java instance name and ip and post can not empty "
		return result
	}
	var binPath,address string
	if param.AgentManager {
		if param.User == "" || param.Password == "" {
			result.validateMessage = "Java user and password can not empty"
			return result
		}
		for _,v := range m.Config().Agent {
			if v.AgentType == "tomcat" {
				binPath = v.AgentBin
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Java agnet bin can not found in config ")
			return result
		}
		address,err = prom.DeployAgent(param.Type,param.Name,binPath,param.Ip,param.Port,param.User,param.Password,agentManagerServer)
		if err != nil {
			result.err = err
			return result
		}
	}
	var jvmVersion,exportVersion string
	if param.FetchMetric {
		tmpIp,tmpPort := param.Ip,param.Port
		if strings.Contains(address, ":") {
			tmpIp = address[:strings.Index(address, ":")]
			tmpPort = address[strings.Index(address, ":")+1:]
		}
		err, strList := prom.GetEndpointData(tmpIp, tmpPort, []string{"catalina", "jvm", "java", "tomcat", "process", "com"}, []string{"version"})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) <= 60 {
			result.err = fmt.Errorf("Connect to instance get metric error, please check param ")
			return result
		}
		for _,v := range strList {
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
	result.endpoint.Step = prometheusStep
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.AddressAgent = address
	result.defaultGroup = "default_java_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
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
	result.endpoint.Address = fmt.Sprintf("%s:%s", param.Ip, param.Port)
	result.endpoint.ExportType = param.Type
	result.endpoint.Step = prometheusStep
	result.defaultGroup = "default_ping_group"
	result.addDefaultGroup = true
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
	result.endpoint.Step = prometheusStep
	result.defaultGroup = "default_telnet_group"
	result.addDefaultGroup = true
	// store to db -> endpoint_telnet
	var eto []*m.EndpointTelnetObj
	eto = append(eto, &m.EndpointTelnetObj{Port:param.Port, Note:""})
	err := db.UpdateEndpointTelnet(m.UpdateEndpointTelnetParam{Guid:result.endpoint.Guid, Config:eto})
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
	result.endpoint.Step = prometheusStep
	result.defaultGroup = "default_http_group"
	result.addDefaultGroup = true
	var eho []*m.EndpointHttpTable
	eho = append(eho, &m.EndpointHttpTable{EndpointGuid:result.endpoint.Guid, Url:param.Url, Method:param.Method})
	err := db.UpdateEndpointHttp(eho)
	if err != nil {
		result.err = err
	}
	return result
}

func windowsRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if param.Ip == "" || param.Port == "" {
		result.validateMessage = "Windows exporter ip and port can not empty"
		return result
	}
	var hostname,sysname,release string
	if param.FetchMetric {
		err,strList := prom.GetEndpointData(param.Ip, param.Port, []string{"wmi"}, []string{})
		if err != nil {
			result.err = err
			return result
		}
		if len(strList) == 0 {
			result.err = fmt.Errorf("Can't get anything from http://%s:%d/metrics ", param.Ip, &param.Port)
			return result
		}
		for _,v := range strList {
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
	result.endpoint.Step = prometheusStep
	result.endpoint.EndpointVersion = release
	result.defaultGroup = "default_windows_group"
	result.addDefaultGroup = true
	result.fetchMetric = true
	return result
}

func nginxRegister(param m.RegisterParamNew)  {

}

func otherExporterRegister(param m.RegisterParamNew) returnData {
	var result returnData
	if param.Name == "" || param.Ip == "" {
		result.validateMessage = "Default endpoint name and ip can not empty "
		return result
	}
	if param.FetchMetric {
		if param.Port == "" {
			result.validateMessage = "Default endpoint port can not empty if you want to get exporter metric "
			return result
		}
		err,strList := prom.GetEndpointData(param.Ip, param.Port, []string{}, []string{})
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
	result.endpoint.Step = prometheusStep
	result.fetchMetric = true
	return result
}