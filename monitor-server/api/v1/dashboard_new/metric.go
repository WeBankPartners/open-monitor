package dashboard_new

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
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
		if middleware.IsIllegalNameNew(v.Metric) {
			err = fmt.Errorf("metric name illegal")
			break
		}
		if err = datasource.CheckPrometheusQL(v.PromExpr); err != nil {
			err = fmt.Errorf("表达式语法校验失败,%s", err.Error())
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err = db.MetricCreate(param, middleware.GetOperateUser(c), middleware.GetMessageMap(c))
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
		if middleware.IsIllegalName(v.Metric) {
			err = fmt.Errorf("metric name illegal")
			break
		}
		if err = datasource.CheckPrometheusQL(v.PromExpr); err != nil {
			err = fmt.Errorf("表达式语法校验失败,%s", err.Error())
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err = db.MetricUpdate(param, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if err = db.SyncMetricComparisonData(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func MetricDelete(c *gin.Context) {
	id := c.Query("id")
	withComparison, err := db.MetricDeleteNew(id)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if withComparison {
			db.SyncMetricComparisonData()
		}
		middleware.ReturnSuccess(c)
	}
}
