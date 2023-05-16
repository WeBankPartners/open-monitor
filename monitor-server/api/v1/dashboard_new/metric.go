package dashboard_new

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func MetricList(c *gin.Context) {
	id := c.Query("id")
	endpointType := c.Query("endpointType")
	serviceGroup := c.Query("serviceGroup")
	result, err := db.MetricList(id, endpointType, serviceGroup)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func MetricCreate(c *gin.Context) {
	var param []*models.MetricTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	for _, v := range param {
		if middleware.IsIllegalMetricName(v.Metric) {
			err = fmt.Errorf("metric name illegal")
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err = db.MetricCreate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func MetricUpdate(c *gin.Context) {
	var param []*models.MetricTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	for _, v := range param {
		if middleware.IsIllegalMetricName(v.Metric) {
			err = fmt.Errorf("metric name illegal")
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err = db.MetricUpdate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func MetricDelete(c *gin.Context) {
	id := c.Query("id")
	err := db.MetricDelete(id)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}
