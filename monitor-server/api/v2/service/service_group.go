package service

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func GetServiceGroupEndpointList(c *gin.Context)  {
	searchType := c.Param("searchType")
	result,err := db.GetServiceGroupEndpointList(searchType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnData(c, result)
	}
}

func ListServiceGroupEndpoint(c *gin.Context)  {
	serviceGroup := c.Param("serviceGroup")
	monitorType := c.Param("monitorType")
	result,err := db.ListServiceGroupEndpoint(serviceGroup,monitorType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnData(c, result)
	}
}