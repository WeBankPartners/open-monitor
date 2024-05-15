package monitor

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strings"
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

// GetCustomChart 查询图表详情
func GetCustomChart(c *gin.Context) {
	var chartDto *models.CustomChartDto
	var chart *models.CustomChart
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	var err error
	chartId := c.Query("chart_id")
	if strings.TrimSpace(chartId) == "" {
		middleware.ReturnParamEmptyError(c, "chart_id")
		return
	}
	if chart, err = db.GetCustomChartById(chartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if configMap, err = db.QueryAllChartSeriesConfig(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if tagMap, err = db.QueryAllChartSeriesTag(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if tagValueMap, err = db.QueryAllChartSeriesTagValue(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if chartDto, err = db.CreateCustomChartDto(&models.CustomChartExtend{CustomChart: chart}, configMap, tagMap, tagValueMap); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnData(c, chartDto)
}

// DeleteCustomChart 删除图表
func DeleteCustomChart(c *gin.Context) {
	var permissionMap map[string]string
	var err error
	var userRoles = middleware.GetOperateUserRoles(c)
	var hasDeletedPermission bool
	chartId := c.Query("chart_id")
	if strings.TrimSpace(chartId) == "" {
		middleware.ReturnParamEmptyError(c, "chart_id")
		return
	}
	if len(userRoles) == 0 {
		middleware.ReturnValidateError(c, "user roles is empty")
		return
	}
	// 判断是否拥有删除权限
	if permissionMap, err = db.QueryCustomChartPermissionByChart(chartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(permissionMap) == 0 {
		permissionMap = make(map[string]string)
	}
	for _, role := range userRoles {
		if v, ok := permissionMap[role]; ok && v == string(models.PermissionMgmt) {
			hasDeletedPermission = true
			break
		}
	}
	if !hasDeletedPermission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("not has deleted permission"))
		return
	}
	// 删除图表
	if err = db.DeleteCustomDashboardChart(chartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

// SharedCustomChart 分享图表
func SharedCustomChart(c *gin.Context) {
	var err error
	var param models.ChartSharedParam
	var actions []*db.Action
	var permissionList []*models.CustomChartPermission
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(param.ChartId) == "" {
		middleware.ReturnParamEmptyError(c, "chartId")
	}
	actions = append(actions, db.GetDeleteCustomChartPermissionSQL(param.ChartId)...)
	if len(param.UseRoles) > 0 || len(param.MgmtRoles) > 0 {
		if len(param.UseRoles) > 0 {
			for _, useRole := range param.UseRoles {
				permissionList = append(permissionList, &models.CustomChartPermission{
					Guid:           guid.CreateGuid(),
					DashboardChart: param.ChartId,
					RoleId:         useRole,
					Permission:     string(models.PermissionUse),
				})
			}
		}
		if len(param.MgmtRoles) > 0 {
			for _, mgmtRole := range param.MgmtRoles {
				permissionList = append(permissionList, &models.CustomChartPermission{
					Guid:           guid.CreateGuid(),
					DashboardChart: param.ChartId,
					RoleId:         mgmtRole,
					Permission:     string(models.PermissionMgmt),
				})
			}
		}
		actions = append(actions, db.GetInsertCustomChartPermissionSQL(permissionList)...)
	}
	actions = append(actions, db.GetUpdateCustomChartPublicSQL(param.ChartId)...)
	if err = db.Transaction(actions); err != nil {
		middleware.ReturnServerHandleError(c, err)
	}
	middleware.ReturnSuccess(c)
}
