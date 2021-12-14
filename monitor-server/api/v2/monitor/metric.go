package monitor

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func ListMetric(c *gin.Context) {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	result, err := db.MetricListNew(guid, monitorType, serviceGroup)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func GetSysMetricTemplate(c *gin.Context) {
	result, err := db.GetSysMetricTemplateConfig()
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}
