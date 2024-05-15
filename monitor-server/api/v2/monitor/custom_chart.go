package monitor

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

// GetSharedChartList 获取可分享的图表列表
func GetSharedChartList(c *gin.Context) {
	var sharedResultMap = make(map[string][]*models.ChartSharedDto)
	var chartList []*models.CustomChart
	var err error
	if chartList, err = db.QueryAllPublicCustomChartList(middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(chartList) > 0 {
		for _, chart := range chartList {
			sharedDto := &models.ChartSharedDto{
				Id:              chart.Guid,
				SourceDashboard: chart.SourceDashboard,
				Name:            chart.Name,
			}
			if _, ok := sharedResultMap[chart.LineType]; !ok {
				sharedResultMap[chart.LineType] = []*models.ChartSharedDto{}
			}
			sharedResultMap[chart.LineType] = append(sharedResultMap[chart.LineType], sharedDto)
		}
	}
	middleware.ReturnData(c, sharedResultMap)
}
