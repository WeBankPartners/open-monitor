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
)

const hostType  = "host"
const mysqlType  = "mysql"
const redisType = "redis"
const tomcatType = "tomcat"
var agentManagerUrl string

func RegisterAgent(c *gin.Context)  {
	var param m.RegisterParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = RegisterJob(param)
		if err != nil {
			mid.ReturnError(c, "Register failed", err)
			return
		}
		mid.ReturnSuccess(c, "Register successfully")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func RegisterJob(param m.RegisterParam) error {
	var err error
	if param.Type != hostType && param.Type != mysqlType && param.Type != redisType && param.Type != tomcatType {
		return fmt.Errorf("Type " + param.Type + " is not supported yet")
	}
	step := 10
	var strList []string
	var endpoint m.EndpointTable
	var tmpAgentIp,tmpAgentPort string
	if agentManagerUrl == "" {
		for _, v := range m.Config().Dependence {
			if v.Name == "agent_manager" {
				agentManagerUrl = v.Server
				break
			}
		}
	}
	if param.Type == hostType {
		err,strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"node"}, []string{})
		if err != nil {
			mid.LogError("Get endpoint data failed ", err)
			return err
		}
		if len(strList) == 0 {
			return fmt.Errorf("Can't get anything from this address:port/metric ")
		}
		var hostname,sysname,release,exportVersion string
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
		endpoint.Guid = fmt.Sprintf("%s_%s_%s", hostname, param.ExporterIp, hostType)
		endpoint.Name = hostname
		endpoint.Ip = param.ExporterIp
		endpoint.ExportType = hostType
		endpoint.Address = fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort)
		endpoint.OsType = sysname
		endpoint.Step = step
		endpoint.EndpointVersion = release
		endpoint.ExportVersion = exportVersion
	}else if param.Type == mysqlType{
		if param.Instance == "" {
			return fmt.Errorf("Mysql instance name can not be empty")
		}
		var binPath,address string
		if agentManagerUrl != "" {
			if param.User == "" || param.Password == "" {
				for _,v := range m.Config().Agent {
					if v.AgentType == mysqlType {
						param.User = v.User
						param.Password = v.Password
						binPath = v.AgentBin
						break
					}
				}
			}
			if param.User == "" || param.Password == "" {
				return fmt.Errorf("mysql monitor must have user and password to connect")
			}
			if binPath == "" {
				for _,v := range m.Config().Agent {
					if v.AgentType == mysqlType {
						binPath = v.AgentBin
						break
					}
				}
			}
			address,err = prom.DeployAgent(mysqlType,param.Instance,binPath,param.ExporterIp,param.ExporterPort,param.User,param.Password,agentManagerUrl)
			if err != nil {
				return err
			}
		}
		if address == "" {
			err, strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"mysql", "mysqld"}, []string{})
		}else{
			if strings.Contains(address, ":") {
				tmpAddressList := strings.Split(address, ":")
				tmpAgentIp = tmpAddressList[0]
				tmpAgentPort = tmpAddressList[1]
				err, strList = prom.GetEndpointData(tmpAddressList[0], tmpAddressList[1], []string{"mysql", "mysqld"}, []string{})
			}else{
				mid.LogInfo(fmt.Sprintf("address : %s is bad", address))
				return fmt.Errorf("address : %s is bad", address)
			}
		}
		if err != nil {
			mid.LogError("curl endpoint data fail ", err)
			return err
		}
		if len(strList) == 0 {
			return fmt.Errorf("Can't get anything from this address:port/metric ")
		}
		var mysqlVersion,exportVersion string
		for _,v := range strList {
			if strings.HasPrefix(v, "mysql_version_info{") {
				mysqlVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
			}
			if strings.HasPrefix(v, "mysqld_exporter_build_info{") {
				exportVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
			}
		}
		endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Instance, param.ExporterIp, mysqlType)
		endpoint.Name = param.Instance
		endpoint.Ip = param.ExporterIp
		endpoint.EndpointVersion = mysqlVersion
		endpoint.ExportType = mysqlType
		endpoint.ExportVersion = exportVersion
		endpoint.Step = step
		endpoint.Address = fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort)
		endpoint.AddressAgent = address
	}else if param.Type == redisType {
		if param.Instance == "" {
			return fmt.Errorf("Redis instance name can not be empty")
		}
		var binPath,address string
		if agentManagerUrl != "" {
			if param.User == "" || param.Password == "" {
				for _,v := range m.Config().Agent {
					if v.AgentType == redisType {
						param.User = v.User
						if param.Password == "" {
							param.Password = v.Password
						}
						binPath = v.AgentBin
						break
					}
				}
			}
			if param.User == "" || param.Password == "" {
				return fmt.Errorf("mysql monitor must have user and password to connect")
			}
			if binPath == "" {
				for _,v := range m.Config().Agent {
					if v.AgentType == redisType {
						binPath = v.AgentBin
						break
					}
				}
			}
			address,err = prom.DeployAgent(redisType,param.Instance,binPath,param.ExporterIp,param.ExporterPort,param.User,param.Password,agentManagerUrl)
			if err != nil {
				return err
			}
		}
		if address == "" {
			err, strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"redis"}, []string{"redis_version", ",version"})
		}else{
			if strings.Contains(address, ":") {
				tmpAddressList := strings.Split(address, ":")
				tmpAgentIp = tmpAddressList[0]
				tmpAgentPort = tmpAddressList[1]
				err, strList = prom.GetEndpointData(tmpAddressList[0], tmpAddressList[1], []string{"redis"}, []string{"redis_version", ",version"})
			}else{
				mid.LogInfo(fmt.Sprintf("address : %s is bad", address))
				return fmt.Errorf("address : %s is bad", address)
			}
		}
		if err != nil {
			mid.LogError("curl endpoint data fail ", err)
			return err
		}
		if len(strList) == 0 {
			return fmt.Errorf("Can't get anything from this address:port/metric ")
		}
		var redisVersion,exportVersion string
		for _,v := range strList {
			if strings.Contains(v, "redis_version") {
				mid.LogInfo(fmt.Sprintf("redis str list : %s", v))
				redisVersion = strings.Split(strings.Split(v, ",redis_version=\"")[1], "\"")[0]
			}
			if strings.Contains(v, ",version") {
				exportVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
			}
		}
		endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Instance, param.ExporterIp, redisType)
		endpoint.Name = param.Instance
		endpoint.Ip = param.ExporterIp
		endpoint.EndpointVersion = redisVersion
		endpoint.ExportType = redisType
		endpoint.ExportVersion = exportVersion
		endpoint.Step = step
		endpoint.Address = fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort)
		endpoint.AddressAgent = address
	}else if param.Type == tomcatType {
		if param.Instance == "" {
			return fmt.Errorf("Tomcat instance name can not be empty")
		}
		var binPath,address string
		if agentManagerUrl != "" {
			if param.User == "" || param.Password == "" {
				for _,v := range m.Config().Agent {
					if v.AgentType == tomcatType {
						param.User = v.User
						param.Password = v.Password
						binPath = v.AgentBin
						break
					}
				}
			}
			if param.User == "" || param.Password == "" {
				return fmt.Errorf("mysql monitor must have user and password to connect")
			}
			if binPath == "" {
				for _,v := range m.Config().Agent {
					if v.AgentType == tomcatType {
						binPath = v.AgentBin
						break
					}
				}
			}
			address,err = prom.DeployAgent(tomcatType,param.Instance,binPath,param.ExporterIp,param.ExporterPort,param.User,param.Password,agentManagerUrl)
			if err != nil {
				return err
			}
		}
		if address == "" {
			err, strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"Catalina", "catalina", "jvm", "java"}, []string{"version"})
		}else{
			if strings.Contains(address, ":") {
				tmpAddressList := strings.Split(address, ":")
				tmpAgentIp = tmpAddressList[0]
				tmpAgentPort = tmpAddressList[1]
				err, strList = prom.GetEndpointData(tmpAddressList[0], tmpAddressList[1], []string{"Catalina", "catalina", "jvm", "java"}, []string{"version"})
			}else{
				mid.LogInfo(fmt.Sprintf("address : %s is bad", address))
				return fmt.Errorf("address : %s is bad", address)
			}
		}
		if err != nil {
			mid.LogError("curl endpoint data fail ", err)
			return err
		}
		if len(strList) == 0 {
			return fmt.Errorf("Can't get anything from this address:port/metric ")
		}
		var jvmVersion,exportVersion string
		for _,v := range strList {
			if strings.Contains(v, "jvm_info") {
				jvmVersion = strings.Split(strings.Split(v, "version=\"")[1], "\"")[0]
			}
			if strings.Contains(v, "jmx_exporter_build_info") {
				exportVersion = strings.Split(strings.Split(v, "version=\"")[1], "\"")[0]
			}
		}
		endpoint.Guid = fmt.Sprintf("%s_%s_%s", param.Instance, param.ExporterIp, tomcatType)
		endpoint.Name = param.Instance
		endpoint.Ip = param.ExporterIp
		endpoint.EndpointVersion = jvmVersion
		endpoint.ExportType = tomcatType
		endpoint.ExportVersion = exportVersion
		endpoint.Step = step
		endpoint.Address = fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort)
		endpoint.AddressAgent = address
	}
	err = db.UpdateEndpoint(&endpoint)
	if err != nil {
		mid.LogError( "Update endpoint failed ", err)
		return err
	}
	err = db.RegisterEndpointMetric(endpoint.Id, strList)
	if err != nil {
		mid.LogError( "Update endpoint metric failed ", err)
		return err
	}
	if tmpAgentIp != "" && tmpAgentPort != "" {
		param.ExporterIp = tmpAgentIp
		param.ExporterPort = tmpAgentPort
	}
	err = prom.RegisteConsul(endpoint.Guid, param.ExporterIp, param.ExporterPort, []string{param.Type}, step)
	if err != nil {
		mid.LogError( "Register consul failed ", err)
		return err
	}
	return nil
}

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
	err := db.DeleteEndpoint(guid)
	if err != nil {
		mid.LogError(fmt.Sprintf("Delete endpint %s failed", guid), err)
		return err
	}
	err = prom.DeregisteConsul(guid)
	if err != nil {
		mid.LogError(fmt.Sprintf("Deregister consul %s failed ", guid), err)
		return err
	}
	return err
}

type transGatewayRequestDto struct {
	Name  string  `json:"name"`
	HostIp  string  `json:"host_ip"`
	Address  string  `json:"address"`
}

func CustomRegister(c *gin.Context)  {
	var param transGatewayRequestDto
	if err:=c.ShouldBindJSON(&param); err==nil {
		var endpointObj m.EndpointTable
		endpointObj.Guid = fmt.Sprintf("%s_%s_custom", param.Name, param.HostIp)
		endpointObj.Address = param.Address
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