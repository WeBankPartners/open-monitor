package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
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
			if _, ok := sharedResultMap[chart.ChartType]; !ok {
				sharedResultMap[chart.ChartType] = []*models.ChartSharedDto{}
			}
			sharedResultMap[chart.ChartType] = append(sharedResultMap[chart.ChartType], sharedDto)
		}
	}
	middleware.ReturnSuccessData(c, sharedResultMap)
}

func AddCustomChart(c *gin.Context) {
	var err error
	var param models.AddCustomChartParam
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.DashboardId == 0 {
		middleware.ReturnParamEmptyError(c, "dashboardId")
		return
	}
	if err = db.AddCustomChart(param, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func CopyCustomChart(c *gin.Context) {
	var err error
	var param models.CopyCustomChartParam
	var customDashboard *models.CustomDashboardTable
	var chart *models.CustomChart
	var displayConfig []byte
	user := middleware.GetOperateUser(c)
	now := time.Now().Format(models.DatetimeFormat)
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.DashboardId == 0 {
		middleware.ReturnParamEmptyError(c, "dashboardId")
		return
	}
	if strings.TrimSpace(param.OriginChartId) == "" {
		middleware.ReturnParamEmptyError(c, "originChartId")
		return
	}
	if customDashboard, err = db.GetCustomDashboardById(param.DashboardId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboard == nil || customDashboard.Id == 0 {
		middleware.ReturnValidateError(c, "dashboardId is invalid")
		return
	}
	if chart, err = db.GetCustomChartById(param.OriginChartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if chart == nil {
		middleware.ReturnValidateError(c, "originChartId is invalid")
		return
	}
	if param.Ref {
		// 将已有图表加入到看板中
		displayConfig, _ = json.Marshal(param.DisplayConfig)
		rel := &models.CustomDashboardChartRel{
			Guid:            guid.CreateGuid(),
			CustomDashboard: &param.DashboardId,
			DashboardChart:  &param.OriginChartId,
			Group:           param.Group,
			DisplayConfig:   string(displayConfig),
			CreateUser:      user,
			UpdateUser:      user,
			CreateTime:      now,
			UpdateTime:      now,
		}
		if err = db.AddCustomDashboardChartRel(rel); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
		middleware.ReturnSuccess(c)
		return
	}
	// 复制图表,copy 图表的所有数据并且与看板关联
	if err = db.CopyCustomChart(param.DashboardId, param.OriginChartId, param.DisplayConfig); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

// UpdateCustomChart 更新自定义图表,先删除图表配置再新增
func UpdateCustomChart(c *gin.Context) {
	var chartDto models.CustomChartDto
	var chart *models.CustomChart
	var err error
	if err = c.ShouldBindJSON(&chartDto); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(chartDto.Id) == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	if chart, err = db.GetCustomChartById(chartDto.Id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if chart == nil {
		middleware.ReturnValidateError(c, "id is invalid")
		return
	}
	if err = db.UpdateCustomChart(chartDto, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
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
	if chartDto, err = db.CreateCustomChartDto(models.ConvertCustomChartToExtend(chart), configMap, tagMap, tagValueMap); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, chartDto)
}

// UpdateCustomChartName 更新图表名称
func UpdateCustomChartName(c *gin.Context) {
	var chartNameParam models.UpdateCustomChartNameParam
	var err error
	if err = c.ShouldBindJSON(&chartNameParam); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(chartNameParam.ChartId) == "" || strings.TrimSpace(chartNameParam.Name) == "" {
		middleware.ReturnParamEmptyError(c, "chartId or name")
		return
	}
	if err = db.UpdateCustomChartName(chartNameParam.ChartId, chartNameParam.Name, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
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
	if permissionMap, err = db.QueryCustomChartManagePermissionByChart(chartId); err != nil {
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
		return
	}
	middleware.ReturnSuccess(c)
}

// GetSharedChartPermission 查询分享图表的权限
func GetSharedChartPermission(c *gin.Context) {
	var list []*models.CustomChartPermission
	var err error
	result := &models.SharedChartPermissionDto{UseRoles: []string{}, MgmtRoles: []string{}}
	chartId := c.Query("chart_id")
	if strings.TrimSpace(chartId) == "" {
		middleware.ReturnParamEmptyError(c, "chart_id")
		return
	}
	if list, err = db.QueryChartPermissionByCustomChart(chartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(list) > 0 {
		for _, permission := range list {
			if permission.Permission == string(models.PermissionUse) {
				result.UseRoles = append(result.UseRoles, permission.RoleId)
			} else if permission.Permission == string(models.PermissionMgmt) {
				result.MgmtRoles = append(result.MgmtRoles, permission.RoleId)
			}
		}
	}
	middleware.ReturnSuccessData(c, result)
}

// QueryCustomChart 查询图表管理列表
func QueryCustomChart(c *gin.Context) {
	var param models.QueryChartParam
	var pageInfo models.PageInfo
	var customChartList []*models.CustomChart
	var dataList []*models.QueryChartResultDto
	var roleRelList []*models.CustomChartPermission
	var customDashboardList []*models.SimpleCustomDashboardDto
	var chartRelList []*models.CustomDashboardChartRel
	var customDashboardMap = make(map[int]string)
	var mgmtRoles, displayMgmtRoles, useRoles, displayUseRoles, useDashboard []string
	var displayNameRoleMap map[string]string
	var userRoleMap map[string]bool
	var permission string
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.PageSize == 0 {
		param.PageSize = 10
	}
	if pageInfo, customChartList, err = db.QueryCustomChartList(param, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if displayNameRoleMap, err = db.QueryAllRoleDisplayNameMap(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboardList, err = db.QueryAllCustomDashboard(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(customDashboardList) > 0 {
		for _, dto := range customDashboardList {
			customDashboardMap[dto.Id] = dto.Name
		}
	}

	userRoleMap = db.TransformArrayToMap(middleware.GetOperateUserRoles(c))
	if len(customChartList) > 0 {
		for _, chart := range customChartList {
			mgmtRoles = []string{}
			displayMgmtRoles = []string{}
			useRoles = []string{}
			displayUseRoles = []string{}
			useDashboard = []string{}
			permission = string(models.PermissionUse)
			if roleRelList, err = db.QueryChartPermissionByCustomChart(chart.Guid); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if chartRelList, err = db.QueryCustomDashboardChartRelListByChart(chart.Guid); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if len(roleRelList) > 0 {
				for _, roleRel := range roleRelList {
					if roleRel.Permission == string(models.PermissionMgmt) {
						mgmtRoles = append(mgmtRoles, roleRel.RoleId)
						if v, ok := displayNameRoleMap[roleRel.RoleId]; ok {
							displayMgmtRoles = append(displayMgmtRoles, v)
						}
					} else if roleRel.Permission == string(models.PermissionUse) {
						useRoles = append(useRoles, roleRel.RoleId)
						if v, ok := displayNameRoleMap[roleRel.RoleId]; ok {
							displayUseRoles = append(displayUseRoles, v)
						}
					}
					if userRoleMap[roleRel.RoleId] {
						permission = string(models.PermissionMgmt)
					}
				}
			}
			if len(chartRelList) > 0 {
				for _, rel := range chartRelList {
					useDashboard = append(useDashboard, customDashboardMap[*rel.CustomDashboard])
				}
			}
			resultDto := &models.QueryChartResultDto{
				ChartId:          chart.Guid,
				ChartName:        chart.Name,
				ChartType:        chart.ChartType,
				SourceDashboard:  customDashboardMap[chart.SourceDashboard],
				UseDashboard:     useDashboard,
				MgmtRoles:        mgmtRoles,
				DisplayMgmtRoles: displayMgmtRoles,
				UseRoles:         useRoles,
				DisplayUseRoles:  displayUseRoles,
				UpdateUser:       chart.UpdateUser,
				CreatedTime:      chart.CreateTime,
				UpdatedTime:      chart.UpdateTime,
				Permission:       permission,
			}
			dataList = append(dataList, resultDto)
		}
	}
	middleware.ReturnPageData(c, pageInfo, dataList)
}
