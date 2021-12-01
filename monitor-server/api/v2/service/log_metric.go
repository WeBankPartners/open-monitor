package service

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/node_exporter"
	"github.com/gin-gonic/gin"
)

func ListLogMetricMonitor(c *gin.Context)  {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	if queryType == "endpoint" {
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

func GetLogMetricMonitor(c *gin.Context)  {
	logMonitorMonitorGuid := c.Param("logMonitorGuid")
	result,err := db.GetLogMetricMonitor(logMonitorMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else {
		middleware.ReturnSuccessData(c, result)
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
		err = syncNodeExporterConfig(param.ServiceGroup, "")
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
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
		err = syncNodeExporterConfig("", param.Guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricMonitor(c *gin.Context)  {
	logMonitorGuid := c.Param("logMonitorGuid")
	serviceGroup,err := db.DeleteLogMetricMonitor(logMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(),err)
	}else{
		if serviceGroup != "" {
			err = syncNodeExporterConfig(serviceGroup, "")
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		}else{
			middleware.ReturnSuccess(c)
		}
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
		err = syncNodeExporterConfig("", param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
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
		err = syncNodeExporterConfig("", param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricJson(c *gin.Context)  {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	logMetricMonitor,err := db.DeleteLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		if logMetricMonitor != "" {
			err = syncNodeExporterConfig("", logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		}else{
			middleware.ReturnSuccess(c)
		}
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
		err = syncNodeExporterConfig("", param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
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
		err = syncNodeExporterConfig("", param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricConfig(c *gin.Context)  {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	logMetricMonitor,err := db.DeleteLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else {
		if logMetricMonitor != "" {
			err = syncNodeExporterConfig("", logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		}else{
			middleware.ReturnSuccess(c)
		}
	}
}

func syncNodeExporterConfig(serviceGroup,logMetricMonitor string) error {
	if serviceGroup == "" {
		serviceGroup = db.GetServiceGroupByLogMetricMonitor(logMetricMonitor)
		if serviceGroup == "" {
			return fmt.Errorf("Sync node exporter log metric fail,serviceGroup and logMetricMonitor can not empty ")
		}
	}
	return node_exporter.SyncLogMetricConfig(serviceGroup)
}