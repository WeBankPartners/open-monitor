package agent

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/prom"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	"strings"
	"fmt"
)

const hostType  = "host"
const mysqlType  = "mysql"
const redisType = "redis"
const tomcatType = "tomcat"

func RegisterAgent(c *gin.Context)  {
	var param m.RegisterParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = RegisterJob(param)
		if err != nil {
			mid.ReturnError(c, "register error", err)
			return
		}
		mid.ReturnSuccess(c, "register success")
	}else{
		mid.ReturnValidateFail(c, "param validate fail")
	}
}

func RegisterJob(param m.RegisterParam) error {
	var err error
	if param.Type != hostType && param.Type != mysqlType && param.Type != redisType && param.Type != tomcatType {
		return fmt.Errorf("type " + param.Type + " is not supported yet")
	}
	step := 10
	var strList []string
	var endpoint m.EndpointTable
	if param.Type == hostType {
		err,strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"node"}, []string{})
		if err != nil {
			mid.LogError("curl endpoint data fail ", err)
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
			return fmt.Errorf("mysql instance name is null")
		}
		err,strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"mysql", "mysqld"}, []string{})
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
	}else if param.Type == redisType {
		if param.Instance == "" {
			return fmt.Errorf("redis instance name is null")
		}
		err,strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"redis"}, []string{"redis_version",",version"})
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
	}else if param.Type == tomcatType {
		if param.Instance == "" {
			return fmt.Errorf("tomcat instance name is null")
		}
		err,strList = prom.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"tomcat", "jvm", "jmx"}, []string{"version"})
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
	}
	err = db.UpdateEndpoint(&endpoint)
	if err != nil {
		mid.LogError( "update endpoint error ", err)
		return err
	}
	err = db.RegisterEndpointMetric(endpoint.Id, strList)
	if err != nil {
		mid.LogError( "update endpoint metric error ", err)
		return err
	}
	err = prom.RegisteConsul(endpoint.Guid, param.ExporterIp, param.ExporterPort, []string{param.Type}, step)
	if err != nil {
		mid.LogError( "register consul fail ", err)
		return err
	}
	return nil
}

func DeregisterAgent(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "guid can't be null")
		return
	}
	err := DeregisterJob(guid)
	if err != nil {
		mid.ReturnError(c, fmt.Sprintf("delete endpint %s fail", guid),err)
		return
	}
	mid.ReturnSuccess(c, fmt.Sprintf("deregister %s success", guid))
}

func DeregisterJob(guid string) error {
	err := db.DeleteEndpoint(guid)
	if err != nil {
		mid.LogError(fmt.Sprintf("delete endpint %s fail", guid), err)
		return err
	}
	err = prom.DeregisteConsul(guid)
	if err != nil {
		mid.LogError(fmt.Sprintf("deregister consul %s fail ", guid), err)
		return err
	}
	return err
}