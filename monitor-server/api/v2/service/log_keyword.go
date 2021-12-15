package service

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func ListLogKeywordMonitor(c *gin.Context)  {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	if queryType == "endpoint" {
		result, err := db.GetLogKeywordByEndpoint(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.GetLogKeywordByServiceGroup(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	}
}

func CreateLogKeywordMonitor(c *gin.Context)  {
	var param models.LogKeywordMonitorCreateObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogKeywordMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogKeywordMonitor(c *gin.Context)  {
	var param models.LogKeywordMonitorObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var endpointList []string
	for _,v := range db.ListLogKeywordEndpointRel(param.Guid) {
		endpointList = append(endpointList, v.SourceEndpoint)
	}
	for _,v := range param.EndpointRel {
		endpointList = append(endpointList, v.SourceEndpoint)
	}
	err := db.UpdateLogKeywordMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordNodeExporterConfig(endpointList)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogKeywordMonitor(c *gin.Context)  {
	logKeywordMonitorGuid := c.Param("logKeywordMonitorGuid")
	var endpointList []string
	for _,v := range db.ListLogKeywordEndpointRel(logKeywordMonitorGuid) {
		endpointList = append(endpointList, v.SourceEndpoint)
	}
	err := db.DeleteLogKeywordMonitor(logKeywordMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordNodeExporterConfig(endpointList)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func CreateLogKeyword(c *gin.Context)  {
	var param models.LogKeywordConfigTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateLogKeyword(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordMonitorConfig(param.LogKeywordMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateLogKeyword(c *gin.Context)  {
	var param models.LogKeywordConfigTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateLogKeyword(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordMonitorConfig(param.LogKeywordMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogKeyword(c *gin.Context)  {
	logKeywordGuid := c.Param("logKeywordGuid")
	err := db.DeleteLogKeyword(logKeywordGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordMonitorConfig(logKeywordGuid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		}else {
			middleware.ReturnSuccess(c)
		}
	}
}

func syncLogKeywordMonitorConfig(logKeywordMonitor string) error {
	endpointList := []string{}
	endpointRel := db.ListLogKeywordEndpointRel(logKeywordMonitor)
	for _, v := range endpointRel {
		if v.TargetEndpoint != "" {
			endpointList = append(endpointList, v.SourceEndpoint)
		}
	}
	err := syncLogKeywordNodeExporterConfig(endpointList)
	return err
}

func syncLogKeywordNodeExporterConfig(endpointList []string) error {
	err := db.SyncLogKeywordExporterConfig(endpointList)
	return err
}