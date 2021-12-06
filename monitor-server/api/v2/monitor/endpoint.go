package monitor

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func ListEndpoint(c *gin.Context) {
	var param models.QueryRequestParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	pageInfo, rowData, err := db.ListEndpoint(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnPageData(c, pageInfo, rowData)
	}
}

func ListMetric(c *gin.Context)  {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	result,err := db.MetricListNew(guid, monitorType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccessData(c, result)
	}
}