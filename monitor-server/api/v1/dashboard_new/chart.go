package dashboard_new

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
)

func ChartList(c *gin.Context)  {
	var id,groupId int
	if c.Query("id") != "" {
		id,_ = strconv.Atoi(c.Query("id"))
	}
	if c.Query("groupId") != "" {
		groupId,_ = strconv.Atoi(c.Query("groupId"))
	}
	result,err := db.ChartList(id,groupId)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccessData(c, result)
	}
}

func ChartCreate(c *gin.Context)  {
	var param []*models.ChartTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.ChartCreate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func ChartUpdate(c *gin.Context)  {
	var param []*models.ChartTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.ChartUpdate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func ChartDelete(c *gin.Context)  {
	var param []*models.ChartTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.ChartDelete(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}
