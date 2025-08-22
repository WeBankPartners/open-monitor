package service

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func GetServiceGroupEndpointList(c *gin.Context) {
	searchType := c.Param("searchType")
	query := c.Query("query")
	monitorType := c.Query("monitorType")
	search := c.Query("search")

	// 如果 query 等于 "Y" 且 searchType 为 "endpoint"，支持模糊搜索和 monitor_type 过滤
	if query == "Y" && searchType == "endpoint" {
		result, err := db.GetServiceGroupEndpointListWithFilter(search, monitorType)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
		middleware.ReturnSuccessData(c, result)
		return
	}
	// 原逻辑保持不变
	result, err := db.GetServiceGroupEndpointList(searchType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func ListServiceGroupEndpoint(c *gin.Context) {
	serviceGroup := c.Param("serviceGroup")
	monitorType := c.Param("monitorType")
	result, err := db.ListServiceGroupEndpoint(serviceGroup, monitorType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}
