package dashboard

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func GetEndpointTypeList(c *gin.Context) {
	result, err := db.GetSimpleEndpointTypeList()
	if err != nil {
		middleware.ReturnQueryTableError(c, "endpoint", err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func GetEndpointTypeNewList(c *gin.Context) {
	result, err := db.GetEndpointTypeList()
	if err != nil {
		middleware.ReturnQueryTableError(c, "endpoint", err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func GetEndpointList(c *gin.Context) {
	endpointType := c.Query("type")
	if endpointType == "" {
		middleware.ReturnParamEmptyError(c, "type")
		return
	}
	serviceGroup := c.Query("serviceGroup")
	endpointGroup := c.Query("endpointGroup")
	workspace := c.Query("workspace")
	result, err := db.GetEndpointByType(endpointType, serviceGroup, endpointGroup, workspace)
	if err != nil {
		middleware.ReturnQueryTableError(c, "endpoint", err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}
