package dashboard_new

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
)

func MetricList(c *gin.Context)  {
	var id int
	if c.Query("id") != "" {
		id,_ = strconv.Atoi(c.Query("id"))
	}
	endpointType := c.Query("endpointType")
	result,err := db.MetricList(id,endpointType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccessData(c, result)
	}
}

func MetricCreate(c *gin.Context)  {
	var param []*models.PromMetricTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.MetricCreate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func MetricUpdate(c *gin.Context)  {
	var param []*models.PromMetricTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.MetricUpdate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func MetricDelete(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		middleware.ReturnValidateError(c, "Param id is illegal")
		return
	}
	err := db.MetricDelete(id)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}
