package service

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func GetLogMetricMonitor(c *gin.Context)  {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	if queryType == "host" {
		result,err := db.GetLogMetricByEndpoint(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else{
			middleware.ReturnSuccessData(c, result)
		}
	}else{
		result,err := db.GetLogMetricByServiceGroup(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else{
			middleware.ReturnSuccessData(c, result)
		}
	}
}

func CreateLogMetricMonitor(c *gin.Context)  {
	var param models.LogMetricMonitorCreateDto
	if err:=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogMetricMonitor(c *gin.Context)  {
	var param models.LogMetricMonitorObj
	if err:=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func DeleteLogMetricMonitor(c *gin.Context)  {
	logMonitorGuid := c.Param("logMonitorGuid")
	err := db.DeleteLogMetricMonitor(logMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func GetLogMetricJson(c *gin.Context)  {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	result,err := db.GetLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricJson(c *gin.Context)  {
	var param models.LogMetricJsonObj
	if err:=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogMetricJson(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		result,_ := db.GetLogMetricJson(param.Guid)
		middleware.ReturnSuccessData(c, result)
	}
}

func UpdateLogMetricJson(c *gin.Context)  {
	var param models.LogMetricJsonObj
	if err:=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateLogMetricJson(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		result,_ := db.GetLogMetricJson(param.Guid)
		middleware.ReturnSuccessData(c, result)
	}
}

func DeleteLogMetricJson(c *gin.Context)  {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	err := db.DeleteLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func GetLogMetricConfig(c *gin.Context)  {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	result,err := db.GetLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricConfig(c *gin.Context)  {
	var param models.LogMetricConfigObj
	if err:=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogMetricConfig(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		result,_ := db.GetLogMetricConfig(param.Guid)
		middleware.ReturnSuccessData(c, result)
	}
}

func UpdateLogMetricConfig(c *gin.Context)  {
	var param models.LogMetricConfigObj
	if err:=c.ShouldBindJSON(&param);err!=nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateLogMetricConfig(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		result,_ := db.GetLogMetricConfig(param.Guid)
		middleware.ReturnSuccessData(c, result)
	}
}

func DeleteLogMetricConfig(c *gin.Context)  {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	err := db.DeleteLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else {
		middleware.ReturnSuccess(c)
	}
}