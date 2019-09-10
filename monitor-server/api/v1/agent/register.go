package agent

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/cron"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	"strings"
	"fmt"
)

const hostType  = "host"
const mysqlType  = "mysql"

func RegisterAgent(c *gin.Context)  {
	var param m.RegisterParam
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Type != hostType && param.Type != mysqlType {
			mid.ReturnError(c, "type " + param.Type + " is not supported yet", nil)
			return
		}
		step := 10
		var strList []string
		var endpoint m.EndpointTable
		if param.Type == hostType {
			err,strList = cron.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"node"})
			if err != nil {
				mid.ReturnError(c,"curl endpoint data fail ", err)
				return
			}
			var hostname,sysname,release,exportVersion string
			for _,v := range strList {
				if strings.HasPrefix(v, "node_uname_info{") {
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
				if strings.HasPrefix(v, "node_exporter_build_info{") {
					exportVersion = strings.Split(strings.Split(v, ",version=\"")[1], "\"")[0]
				}
			}
			endpoint.Guid = fmt.Sprintf("%s_%s_%s", hostname, param.ExporterIp, hostType)
			endpoint.Name = hostname
			endpoint.Ip = param.ExporterIp
			endpoint.ExportType = hostType
			endpoint.OsIp = fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort)
			endpoint.OsType = sysname
			endpoint.Step = step
			endpoint.EndpointVersion = release
			endpoint.ExportVersion = exportVersion
		}else if param.Type == mysqlType{
			if param.Instance == "" {
				mid.ReturnValidateFail(c, "mysql instance name is null")
				return
			}
			err,strList = cron.GetEndpointData(param.ExporterIp, param.ExporterPort, []string{"mysql", "mysqld"})
			if err != nil {
				mid.ReturnError(c,"curl endpoint data fail ", err)
				return
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
			endpoint.OsIp = fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort)
		}
		err = db.UpdateEndpoint(&endpoint)
		if err != nil {
			mid.ReturnError(c, "update endpoint error ", err)
			return
		}
		err = db.RegisterEndpointMetric(endpoint.Id, strList)
		if err != nil {
			mid.ReturnError(c, "update endpoint metric error ", err)
			return
		}
		err = cron.RegisteConsul(endpoint.Guid, param.ExporterIp, param.ExporterPort, []string{param.Type}, step)
		if err != nil {
			mid.ReturnError(c, "register consul fail ", err)
			return
		}
		mid.ReturnSuccess(c, fmt.Sprintf("register endpoint %s success", endpoint.Guid))
	}else{
		mid.ReturnValidateFail(c, "param validate fail")
	}
}

func DeregisterAgent(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "guid can't be null")
		return
	}
	err := db.DeleteEndpoint(guid)
	if err != nil {
		mid.ReturnError(c, fmt.Sprintf("delete endpint %s fail", guid), err)
		return
	}
	err = cron.DeregisteConsul(guid)
	if err != nil {
		mid.ReturnError(c, fmt.Sprintf("deregister consul %s fail ", guid), err)
		return
	}
	mid.ReturnSuccess(c, fmt.Sprintf("deregister %s success", guid))
}