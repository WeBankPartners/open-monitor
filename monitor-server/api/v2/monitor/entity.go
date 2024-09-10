package monitor

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func QueryEntityEndpoint(c *gin.Context) {
	var param models.EntityQueryParam
	result := models.EndpointEntityResp{}
	if err := c.ShouldBindJSON(&param); err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %s", err.Error())
		middleware.ReturnData(c, result)
		return
	}
	queryResult, queryErr := db.QueryEntityEndpoint(&param)
	if queryErr != nil {
		result.Status = "ERROR"
		result.Message = queryErr.Error()
	} else {
		result.Status = "OK"
		result.Message = "Success"
		result.Data = queryResult
	}
	middleware.ReturnData(c, result)
}

func QueryEntityServiceGroup(c *gin.Context) {
	var param models.EntityQueryParam
	result := models.ServiceGroupEntityResp{}
	if err := c.ShouldBindJSON(&param); err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %s", err.Error())
		middleware.ReturnData(c, result)
		return
	}
	queryResult, queryErr := db.QueryEntityServiceGroup(&param)
	if queryErr != nil {
		result.Status = "ERROR"
		result.Message = queryErr.Error()
	} else {
		result.Status = "OK"
		result.Message = "Success"
		result.Data = queryResult
	}
	middleware.ReturnData(c, result)
}

func QueryEntityEndpointGroup(c *gin.Context) {
	var param models.EntityQueryParam
	result := models.EndpointGroupEntityResp{}
	if err := c.ShouldBindJSON(&param); err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %s", err.Error())
		middleware.ReturnData(c, result)
		return
	}
	queryResult, queryErr := db.QueryEntityEndpointGroup(&param)
	if queryErr != nil {
		result.Status = "ERROR"
		result.Message = queryErr.Error()
	} else {
		result.Status = "OK"
		result.Message = "Success"
		result.Data = queryResult
	}
	middleware.ReturnData(c, result)
}

func QueryEntityMonitorType(c *gin.Context) {
	var param models.EntityQueryParam
	result := models.MonitorTypeEntityResp{}
	if err := c.ShouldBindJSON(&param); err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %s", err.Error())
		middleware.ReturnData(c, result)
		return
	}
	queryResult, queryErr := db.QueryEntityMonitorType(&param)
	if queryErr != nil {
		result.Status = "ERROR"
		result.Message = queryErr.Error()
	} else {
		result.Status = "OK"
		result.Message = "Success"
		result.Data = queryResult
	}
	middleware.ReturnData(c, result)
}
