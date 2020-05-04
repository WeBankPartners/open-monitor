package agent

import (
	"fmt"
	"strings"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
)

const prometheusStep = 10

var agentManagerServer string

type returnData struct {
	endpoint  m.EndpointTable
	metricList  []string
	defaultGroup  string
	validateMessage  string
	err  error
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
		case "jmx": rData = jmxRegister(param)
		default: rData.validateMessage = fmt.Sprintf("Type : %s not supported yet ", param.Type)
	}
	if rData.validateMessage != "" || rData.err != nil {
		return rData.validateMessage,rData.err
	}
	err = db.UpdateEndpoint(&rData.endpoint)
	if err != nil {
		return validateMessage,err
	}
	if param.FetchMetric {
		err = db.RegisterEndpointMetric(rData.endpoint.Id, rData.metricList)
		if err != nil {
			return validateMessage,err
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
	if param.AddDefaultGroup {
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
	if param.FetchMetric {
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
	}
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
		address,err = prom.DeployAgent(param.Type,param.Name,binPath,param.Ip,param.Port,param.User,param.Password,agentManagerUrl)
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
		address,err = prom.DeployAgent(param.Type,param.Name,binPath,param.Ip,param.Port,param.User,param.Password,agentManagerUrl)
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
	return result
}

func jmxRegister(param m.RegisterParamNew) returnData {
	var result returnData
	var err error
	if param.Name == "" || param.Ip == "" || param.Port == "" {
		result.validateMessage = "Jmx instance name and ip and post can not empty "
		return result
	}
	var binPath,address string
	if param.AgentManager {
		if param.User == "" || param.Password == "" {
			result.validateMessage = "Jmx user and password can not empty"
			return result
		}
		for _,v := range m.Config().Agent {
			if v.AgentType == param.Type {
				binPath = v.AgentBin
				break
			}
		}
		if binPath == "" {
			result.err = fmt.Errorf("Jmx agnet bin can not found in config ")
			return result
		}
		address,err = prom.DeployAgent(param.Type,param.Name,binPath,param.Ip,param.Port,param.User,param.Password,agentManagerUrl)
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
	return result
}

func pingRegister()  {

}

func telnetRegister()  {

}

func httpRegister()  {

}

func windowsRegister()  {

}

func nginxRegister()  {

}

func otherExporterRegister()  {

}