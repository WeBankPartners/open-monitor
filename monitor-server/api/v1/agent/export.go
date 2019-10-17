package agent

import (
	"github.com/gin-gonic/gin"
	"strings"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"net/http"
	"fmt"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
)

type resultObj struct {
	ResultCode  int  `json:"result_code"`
	ResultMessage  string  `json:"result_message"`
}

func StartHostAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	osType := strings.ToLower(c.Query("osType"))
	if !isLinuxType(osType) {
		mid.ReturnValidateFail(c, "os_type is not a linux type")
		return
	}
	param := m.RegisterParam{Type:hostType, ExporterIp:hostIp, ExporterPort:"9100"}
	err := RegisterJob(param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("register %s:%s fail,error %v",hostType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StopHostAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	endpointObj := m.EndpointTable{Ip:hostIp, ExportType:hostType}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,can't find this host",hostType, hostIp)})
		return
	}
	err := DeregisterJob(endpointObj.Guid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,error %v",hostType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StartMysqlAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	instance := c.Query("instance")
	if instance == "" {
		mid.ReturnValidateFail(c, "instance is null")
		return
	}
	param := m.RegisterParam{Type:mysqlType, ExporterIp:hostIp, ExporterPort:"9104", Instance:instance}
	err := RegisterJob(param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("register %s:%s fail,error %v",mysqlType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StopMysqlAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	instance := c.Query("instance")
	if instance == "" {
		mid.ReturnValidateFail(c, "instance is null")
		return
	}
	endpointObj := m.EndpointTable{Ip:hostIp, ExportType:mysqlType, Name:instance}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,can't find this host",mysqlType, hostIp)})
		return
	}
	err := DeregisterJob(endpointObj.Guid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,error %v",mysqlType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StartRedisAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	instance := c.Query("instance")
	if instance == "" {
		mid.ReturnValidateFail(c, "instance is null")
		return
	}
	param := m.RegisterParam{Type:redisType, ExporterIp:hostIp, ExporterPort:"9121", Instance:instance}
	err := RegisterJob(param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("register %s:%s fail,error %v",redisType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StopRedisAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	instance := c.Query("instance")
	if instance == "" {
		mid.ReturnValidateFail(c, "instance is null")
		return
	}
	endpointObj := m.EndpointTable{Ip:hostIp, ExportType:redisType, Name:instance}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,can't find this host",redisType, hostIp)})
		return
	}
	err := DeregisterJob(endpointObj.Guid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,error %v",redisType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StartTomcatAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	instance := c.Query("instance")
	if instance == "" {
		mid.ReturnValidateFail(c, "instance is null")
		return
	}
	param := m.RegisterParam{Type:tomcatType, ExporterIp:hostIp, ExporterPort:"9151", Instance:instance}
	err := RegisterJob(param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("register %s:%s fail,error %v",tomcatType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func StopTomcatAgent(c *gin.Context)  {
	hostIp := c.Query("host_ip")
	instance := c.Query("instance")
	if instance == "" {
		mid.ReturnValidateFail(c, "instance is null")
		return
	}
	endpointObj := m.EndpointTable{Ip:hostIp, ExportType:tomcatType, Name:instance}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,can't find this host",tomcatType, hostIp)})
		return
	}
	err := DeregisterJob(endpointObj.Guid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resultObj{ResultCode:2, ResultMessage:fmt.Sprintf("deregister %s:%s fail,error %v",tomcatType, hostIp, err)})
		return
	}
	mid.ReturnData(c, resultObj{ResultCode:1, ResultMessage:"success"})
}

func isLinuxType(osType string) bool {
	linuxType := []string{"linux", "redhat", "centos", "ubuntu", "unix"}
	result := false
	for _,v := range linuxType {
		if strings.Contains(osType, v) {
			result = true
			break
		}
	}
	return result
}
