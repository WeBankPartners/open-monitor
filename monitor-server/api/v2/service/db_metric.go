package service

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strings"
)

func ListDbMetricMonitor(c *gin.Context) {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	metricKey := c.Query("metricKey")
	if queryType == "endpoint" {
		result, err := db.GetDbMetricByEndpoint(guid, metricKey)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.QueryDbMetricWithServiceGroup(guid, metricKey)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, []*models.DbMetricQueryObj{result})
		}
	}
}

func GetDbMetricMonitor(c *gin.Context) {
	dbMonitorMonitorGuid := c.Param("dbMonitorGuid")
	result, err := db.GetDbMetric(dbMonitorMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateDbMetricMonitor(c *gin.Context) {
	var param models.DbMetricMonitorObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if middleware.IsIllegalLogParamNameOrMetric(param.Metric) {
		middleware.ReturnValidateError(c, "metric param invalid")
		return
	}
	param.MetricSql = strings.TrimSpace(param.MetricSql)
	param.MetricSql = strings.ReplaceAll(param.MetricSql, "\n", " ")
	err := db.CreateDbMetric(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncDbMetric(false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateDbMetricMonitor(c *gin.Context) {
	var param models.DbMetricMonitorObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	param.MetricSql = strings.TrimSpace(param.MetricSql)
	param.MetricSql = strings.ReplaceAll(param.MetricSql, "\n", " ")
	err := db.UpdateDbMetric(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncDbMetric(false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteDbMetricMonitor(c *gin.Context) {
	dbMonitorMonitorGuid := c.Param("dbMonitorGuid")
	err := db.DeleteDbMetric(dbMonitorMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncDbMetric(false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}
