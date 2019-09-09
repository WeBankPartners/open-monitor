package agent

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	u "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware/util"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/cron"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	"strings"
	"fmt"
)

func RegisteAgent(c *gin.Context)  {
	var param m.RegisterParam
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Type == "host" {
			err,strList := cron.GetEndpointData(param.ExporterIp, param.ExporterPort, "node_")
			if err != nil {
				mid.LogError("curl endpoint data fail ", err)
				u.ReturnError(c,"curl endpoint data fail ", err)
				return
			}
			var hostname,sysname string
			for _,v := range strList {
				if strings.HasPrefix(v, "node_uname_info{") {
					if strings.Contains(v, "nodename") {
						hostname = strings.Split(strings.Split(v, "nodename=\"")[1], "\"")[0]
					}
					if strings.Contains(v, "sysname") {
						sysname = strings.Split(strings.Split(v, "sysname=\"")[1], "\"")[0]
					}
					break
				}
			}
			endpoint := m.EndpointTable{Guid:param.ExporterIp, Name:hostname, Ip:param.ExporterIp, ExportType:"node", OsIp:fmt.Sprintf("%s:%s", param.ExporterIp, param.ExporterPort), OsType:sysname}
			err = db.UpdateEndpoint(&endpoint)
			if err != nil {
				u.ReturnError(c, "update endpoint error ", err)
				return
			}
			err = db.RegisterEndpoint(endpoint.Id, strList)
			if err != nil {
				u.ReturnError(c, "update endpoint metric error ", err)
				return
			}
			u.ReturnSuccess(c, "register endpoint " + endpoint.Guid + " success")
		}else{
			u.ReturnError(c, "other type is not supported yet", nil)
		}
	}else{
		u.ReturnValidateFail(c, "param validate fail")
	}
}
