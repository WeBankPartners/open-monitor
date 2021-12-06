package monitor

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
)

func ListEndpoint(c *gin.Context) {
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	monitorType := c.Query("monitorType")
	param := models.QueryRequestParam{}
	if size > 0 {
		param.Paging = true
		param.Pageable = &models.PageInfo{PageSize: size,StartIndex: page-1}
	}
	if monitorType != "" {
		param.Filters = []*models.QueryRequestFilterObj{{Name: "monitor_type",Operator: "eq",Value: monitorType}}
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