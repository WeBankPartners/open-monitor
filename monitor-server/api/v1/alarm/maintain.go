package alarm

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"fmt"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
)

func UpdateEndpointMaintain(c *gin.Context) {
	var param m.MaintainDto
	if err := c.ShouldBindJSON(&param);err == nil {
		var endpointObj m.EndpointTable
		if param.Endpoint == "" {
			if param.Ip == "" || param.EndpointType == "" {
				mid.ReturnValidateFail(c, "IP or endpoint type can not be empty")
				return
			}
			endpointObj.Ip = param.Ip
			endpointObj.ExportType = param.EndpointType
			err = db.GetEndpoint(&endpointObj)
			if err != nil || endpointObj.Id <= 0 {
				mid.ReturnError(c, fmt.Sprintf("t endpoint failed with ip :%s type:%s", param.Ip, param.EndpointType), err)
				return
			}
		}else{
			endpointObj.Guid = param.Endpoint
			err = db.GetEndpoint(&endpointObj)
			if err != nil || endpointObj.Id <= 0 {
				mid.ReturnError(c, fmt.Sprintf("Get endpoint failed with guid:%s", param.Endpoint), err)
				return
			}
		}

	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}
