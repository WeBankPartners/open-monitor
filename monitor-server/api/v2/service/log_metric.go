package service

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func ListLogMetricMonitor(c *gin.Context) {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	if queryType == "endpoint" {
		result, err := db.GetLogMetricByEndpoint(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.GetLogMetricByServiceGroup(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	}
}

func GetLogMetricMonitor(c *gin.Context) {
	logMonitorMonitorGuid := c.Param("logMonitorGuid")
	result, err := db.GetLogMetricMonitor(logMonitorMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricMonitor(c *gin.Context) {
	var param models.LogMetricMonitorCreateDto
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogMetricMonitor(c *gin.Context) {
	var param models.LogMetricMonitorObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result, err := db.GetLogMetricMonitor(param.Guid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	hostEndpointList := []string{}
	if param.LogPath != result.LogPath {
		for _, v := range param.EndpointRel {
			hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
		}
	} else {
		for _, v := range result.EndpointRel {
			existFlag, changeFlag := false, false
			for _, vv := range param.EndpointRel {
				if vv.SourceEndpoint == v.SourceEndpoint {
					existFlag = true
					if vv.TargetEndpoint != v.TargetEndpoint {
						changeFlag = true
					}
					break
				}
			}
			if !existFlag {
				// remove endpoint
				hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
			} else {
				if changeFlag {
					// update endpoint rel
					hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
				}
			}
		}
		for _, v := range param.EndpointRel {
			existFlag := false
			for _, vv := range result.EndpointRel {
				if vv.SourceEndpoint == v.SourceEndpoint {
					existFlag = true
					break
				}
			}
			if !existFlag {
				// add endpoint
				hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
			}
		}
	}
	err = db.UpdateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncNodeExporterConfig(hostEndpointList)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricMonitor(c *gin.Context) {
	logMonitorGuid := c.Param("logMonitorGuid")
	result, err := db.GetLogMetricMonitor(logMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	hostEndpointList := []string{}
	for _, v := range result.EndpointRel {
		if v.TargetEndpoint != "" {
			hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
		}
	}
	serviceGroup, err := db.DeleteLogMetricMonitor(logMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if serviceGroup != "" {
			err = syncNodeExporterConfig(hostEndpointList)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func GetLogMetricJson(c *gin.Context) {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	result, err := db.GetLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricJson(c *gin.Context) {
	var param models.LogMetricJsonObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogMetricJson(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateLogMetricJson(c *gin.Context) {
	var param models.LogMetricJsonObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateLogMetricJson(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricJson(c *gin.Context) {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	logMetricMonitor, err := db.DeleteLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if logMetricMonitor != "" {
			err = syncLogMetricMonitorConfig(logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func GetLogMetricConfig(c *gin.Context) {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	result, err := db.GetLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricConfig(c *gin.Context) {
	var param models.LogMetricConfigObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogMetricConfig(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateLogMetricConfig(c *gin.Context) {
	var param models.LogMetricConfigObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateLogMetricConfig(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricConfig(c *gin.Context) {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	logMetricMonitor, err := db.DeleteLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if logMetricMonitor != "" {
			err = syncLogMetricMonitorConfig(logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func syncLogMetricMonitorConfig(logMetricMonitor string) error {
	endpointList := []string{}
	endpointRel := db.ListLogMetricEndpointRel("", logMetricMonitor)
	for _, v := range endpointRel {
		if v.TargetEndpoint != "" {
			endpointList = append(endpointList, v.SourceEndpoint)
		}
	}
	err := syncNodeExporterConfig(endpointList)
	return err
}

func syncNodeExporterConfig(endpointList []string) error {
	err := db.UpdateNodeExportConfig(endpointList)
	return err
}

func CheckRegExpMatch(c *gin.Context) {
	var param models.CheckRegExpParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result := db.CheckRegExpMatch(param)
	middleware.ReturnSuccessData(c, result)
}

func GetServiceGroupEndpointRel(c *gin.Context)  {
	serviceGroup := c.Query("serviceGroup")
	sourceType := c.Query("sourceType")
	targetType := c.Query("targetType")
	result,err := db.GetServiceGroupEndpointRel(serviceGroup,sourceType,targetType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccessData(c, result)
	}
}