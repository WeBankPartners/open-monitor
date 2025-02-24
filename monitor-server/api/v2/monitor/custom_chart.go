package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/dashboard_new"
	"github.com/WeBankPartners/open-monitor/monitor-server/common"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetSharedChartList 获取可分享的图表列表
func GetSharedChartList(c *gin.Context) {
	var sharedResultMap = make(map[string][]*models.ChartSharedDto)
	var chartList, newChartList []*models.CustomChart
	var customChartList []*models.CustomChartExtend
	var param models.SharedChartListParam
	var customDashboard *models.CustomDashboardObj
	var err error
	var exist bool
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if chartList, err = db.QueryAllPublicCustomChartList(param.DashboardId, param.ChartName, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(chartList) > 0 {
		// 去掉看板里面 已有重复的图表
		if param.CurDashboardId != 0 {
			if customChartList, err = db.QueryCustomChartListByDashboard(param.CurDashboardId); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if len(customChartList) > 0 {
				for _, chart := range chartList {
					exist = false
					for _, customChart := range customChartList {
						if chart.Guid == customChart.Guid {
							exist = true
							break
						}
					}
					if !exist {
						newChartList = append(newChartList, chart)
					}
				}
			} else {
				newChartList = chartList
			}
		} else {
			newChartList = chartList
		}
		if len(newChartList) > 0 {
			for _, chart := range newChartList {
				sharedDto := &models.ChartSharedDto{
					Id:              chart.Guid,
					SourceDashboard: chart.SourceDashboard,
					Name:            chart.Name,
					UpdateTime:      chart.UpdateTime,
				}
				if strings.TrimSpace(chart.ChartType) == "" {
					continue
				}
				if _, ok := sharedResultMap[chart.ChartType]; !ok {
					sharedResultMap[chart.ChartType] = []*models.ChartSharedDto{}
				}
				sharedResultMap[chart.ChartType] = append(sharedResultMap[chart.ChartType], sharedDto)
			}
		}
	}
	// 每种类型中最多展示20条数据
	for key, valueList := range sharedResultMap {
		sort.Sort(models.ChartSharedDtoSort(valueList))
		valueList = valueList[:min(20, len(valueList))]
		for _, chart := range valueList {
			if customDashboard, err = db.GetCustomDashboard(chart.SourceDashboard); err != nil {
				continue
			}
			if customDashboard != nil {
				chart.DashboardName = customDashboard.Name
			}
		}
		sharedResultMap[key] = valueList
	}
	middleware.ReturnSuccessData(c, sharedResultMap)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
func AddCustomChart(c *gin.Context) {
	var err error
	var param models.AddCustomChartParam
	var id string
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.DashboardId == 0 {
		middleware.ReturnParamEmptyError(c, "dashboardId")
		return
	}
	if id, err = db.AddCustomChart(param, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, id)
}

func CopyCustomChart(c *gin.Context) {
	var err error
	var param models.CopyCustomChartParam
	var customDashboard *models.CustomDashboardTable
	var chart *models.CustomChart
	var displayConfig []byte
	var newChartId string
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
	defer common.ExceptionStack(func(e interface{}, err interface{}) {
		retErr := fmt.Errorf("%v", err)
		middleware.ReturnServerHandleError(c, retErr)
		log.Error(nil, log.LOGGER_APP, e.(string))
	})
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
		if err = db.UpdateCustomDashboardTime(param.DashboardId, user); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
		newChartId = param.OriginChartId
		middleware.ReturnSuccessData(c, newChartId)
		return
	}
	// 复制图表,copy 图表的所有数据并且与看板关联
	if newChartId, err = db.CopyCustomChart(param.DashboardId, middleware.GetOperateUser(c), param.Group, chart, param.DisplayConfig); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, newChartId)
}

// UpdateCustomChart 更新自定义图表,先删除图表配置再新增
func UpdateCustomChart(c *gin.Context) {
	var chartDto models.CustomChartDto
	var chart *models.CustomChart
	var err error
	var permission bool
	if err = c.ShouldBindJSON(&chartDto); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(chartDto.Id) == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	// 判断是否拥有删除权限
	if permission, err = CheckHasChartManagePermission(chartDto.Id, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("no edit permission"))
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
	if err = db.UpdateCustomChart(chartDto, middleware.GetOperateUser(c), chart.SourceDashboard); err != nil {
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
	var metricComparisonMap = make(map[string]string)
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
	if metricComparisonMap, err = db.GetAllMetricComparison(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	chartParam := models.CreateCustomChartParam{
		ChartExtend:         models.ConvertCustomChartToExtend(chart),
		ConfigMap:           configMap,
		TagMap:              tagMap,
		TagValueMap:         tagValueMap,
		MetricComparisonMap: metricComparisonMap,
		ChartSeries:         []*models.CustomChartSeries{},
	}
	if chartDto, err = db.CreateCustomChartDto(chartParam); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, chartDto)
}

// UpdateCustomChartName 更新图表名称
func UpdateCustomChartName(c *gin.Context) {
	var chartNameParam models.UpdateCustomChartNameParam
	var permission bool
	var chart *models.CustomChart
	var err error
	if err = c.ShouldBindJSON(&chartNameParam); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(chartNameParam.ChartId) == "" || strings.TrimSpace(chartNameParam.Name) == "" {
		middleware.ReturnParamEmptyError(c, "chartId or name")
		return
	}
	if chart, err = db.GetCustomChartById(chartNameParam.ChartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if chart == nil {
		middleware.ReturnValidateError(c, "chartId is invalid")
		return
	}
	// 判断是否拥有删除权限
	if permission, err = CheckHasChartManagePermission(chartNameParam.ChartId, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("no update permission"))
		return
	}
	if err = db.UpdateCustomChartName(chartNameParam.ChartId, chartNameParam.Name, middleware.GetOperateUser(c), chart.SourceDashboard); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func QueryCustomChartNameExist(c *gin.Context) {
	var err error
	var list []*models.CustomChart
	var chart *models.CustomChart
	var customChartExtendList []*models.CustomChartExtend
	chartId := c.Query("chart_id")
	name := c.Query("name")
	// public 表示是否是存入图表库操作,是存入图表库则需要查询图表库里面是否有重名
	public, _ := strconv.Atoi(c.Query("public"))
	if strings.TrimSpace(chartId) == "" || strings.TrimSpace(name) == "" {
		middleware.ReturnParamEmptyError(c, "chart_id or name")
		return
	}
	if chart, err = db.GetCustomChartById(chartId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if chart == nil {
		middleware.ReturnValidateError(c, "chart_id is invalid")
		return
	}
	if chart.Public == 0 && public == 0 {
		// 没有存入到图表库的图表,在源看板中图表不能重复
		if chart.SourceDashboard != 0 {
			if customChartExtendList, err = db.QueryCustomChartListByDashboard(chart.SourceDashboard); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if len(customChartExtendList) > 0 {
				for _, extend := range customChartExtendList {
					if extend.Guid != chart.Guid && extend.Name == name {
						middleware.ReturnSuccessData(c, true)
						return
					}
				}
			}
		}
		middleware.ReturnSuccessData(c, false)
		return
	}
	if list, err = db.QueryCustomChartNameExist(name); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(list) > 0 {
		for _, chart := range list {
			if chart.Guid != chartId {
				middleware.ReturnSuccessData(c, true)
				return
			}
		}
	}
	middleware.ReturnSuccessData(c, false)
}

// DeleteCustomChart 删除图表
func DeleteCustomChart(c *gin.Context) {
	var err error
	var permission bool
	chartId := c.Query("chart_id")
	if strings.TrimSpace(chartId) == "" {
		middleware.ReturnParamEmptyError(c, "chart_id")
		return
	}
	// 判断是否拥有删除权限
	if permission, err = CheckHasChartManagePermission(chartId, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("no delete permission"))
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
	var permission bool
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(param.ChartId) == "" {
		middleware.ReturnParamEmptyError(c, "chartId")
		return
	}
	if permission, err = CheckHasChartManagePermission(param.ChartId, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("no edit permission"))
		return
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

func GetSharedChartPermissionBatch(c *gin.Context) {
	var err error
	var param models.ChartPermissionBatchParam
	var list []*models.CustomChartPermission
	var roles []string
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if list, err = db.QueryChartPermissionByCustomChartList(param.Ids); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	for _, permission := range list {
		roles = append(roles, permission.RoleId)
	}
	middleware.ReturnSuccessData(c, roles)
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
	if pageInfo, customChartList, err = db.QueryCustomChartList(param, middleware.GetOperateUser(c), middleware.GetOperateUserRoles(c)); err != nil {
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
						if userRoleMap[roleRel.RoleId] {
							permission = string(models.PermissionMgmt)
						}
					} else if roleRel.Permission == string(models.PermissionUse) {
						useRoles = append(useRoles, roleRel.RoleId)
						if v, ok := displayNameRoleMap[roleRel.RoleId]; ok {
							displayUseRoles = append(displayUseRoles, v)
						}
					}
				}
			}
			if len(chartRelList) > 0 {
				for _, rel := range chartRelList {
					if rel.CustomDashboard != nil {
						useDashboard = append(useDashboard, customDashboardMap[*rel.CustomDashboard])
					}
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
				LogMetricGroup:   chart.LogMetricGroup,
			}
			dataList = append(dataList, resultDto)
		}
	}
	middleware.ReturnPageData(c, pageInfo, dataList)
}

func CheckHasChartManagePermission(chartId string, userRoles []string) (permission bool, err error) {
	var permissionMap map[string]string
	var chart *models.CustomChart
	if len(userRoles) == 0 {
		return
	}
	if chart, err = db.GetCustomChartById(chartId); err != nil {
		return
	}
	if chart == nil {
		return
	}
	// 私有图表,看是否拥有源看板的管理权限
	if chart.Public == 0 && chart.SourceDashboard != 0 {
		if permissionMap, err = db.QueryCustomDashboardManagePermissionByDashboard(chart.SourceDashboard); err != nil {
			return
		}
		for _, role := range userRoles {
			if _, ok := permissionMap[role]; ok {
				permission = true
				return
			}
		}
	}
	// 公开图表,看是否有图表的管理权限
	if permissionMap, err = db.QueryCustomChartManagePermissionByChart(chartId); err != nil {
		return
	}
	for _, role := range userRoles {
		if _, ok := permissionMap[role]; ok {
			permission = true
			return
		}
	}
	return
}

func GetChartSeriesColor(c *gin.Context) {
	var param models.GetChartSeriesColorParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	result := []*models.ColorConfigDto{}
	// 增加实时查chart合并series
	queryChartParam := models.ChartQueryParam{Start: time.Now().Unix() - 1800, End: time.Now().Unix(), Aggregate: "none", Step: 10, Data: []*models.ChartQueryConfigObj{{
		Endpoint:     param.Endpoint,
		Metric:       param.Metric,
		AppObject:    param.ServiceGroup,
		MonitorType:  param.MonitorType,
		EndpointType: param.MonitorType,
		Tags:         param.Tags,
	}}}
	var querySeriesResult = models.EChartOption{Legend: []string{}, Series: []*models.SerialModel{}}
	querySeriesConfigList, buildQueryConfigErr := dashboard_new.GetChartConfigByCustom(&queryChartParam)
	if buildQueryConfigErr != nil {
		middleware.ReturnServerHandleError(c, buildQueryConfigErr)
		return
	}
	if len(param.Tags) > 0 && len(querySeriesConfigList) > 0 {
		var tagValue []string
		var equal string
		for _, tag := range param.Tags {
			if tag.TagName == "calc_type" {
				tagValue = tag.TagValue
				if tag.Equal == db.ConstEqualIn {
					equal = "="
				} else {
					equal = "!="
				}
				break
			}
		}
		if len(tagValue) > 0 {
			for _, data := range querySeriesConfigList {
				var promQ = data.PromQ
				if promQ == "" {
					continue
				}
				if strings.Contains(data.PromQ, "{") {
					for i, tag := range tagValue {
						if i == 0 {
							data.PromQ = data.PromQ[:len(data.PromQ)-1] + ",calc_type" + equal + "'" + tag + "'}"
						} else {
							data.PromQ = data.PromQ + " or " + promQ[:len(promQ)-1] + ",calc_type" + equal + "'" + tag + "'}"
						}
					}
				} else {
					for i, tag := range tagValue {
						if i == 0 {
							data.PromQ = data.PromQ + "{calc_type" + equal + "'" + tag + "'}"
						} else {
							data.PromQ = data.PromQ + " or " + promQ + "{calc_type" + equal + "'" + tag + "'}"
						}
					}
				}
			}
		}
	}
	log.Debug(nil, log.LOGGER_APP, "GetChartSeriesColor config", log.JsonObj("querySeriesConfigList", querySeriesConfigList))
	err := dashboard_new.GetChartQueryData(querySeriesConfigList, &queryChartParam, &querySeriesResult)
	if err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	for _, v := range querySeriesResult.Legend {
		result = append(result, &models.ColorConfigDto{SeriesName: v, New: true})
	}
	log.Debug(nil, log.LOGGER_APP, "GetChartSeriesColor result", log.JsonObj("result", result))
	existSeriesMap := make(map[string]string)
	// 查已保存的series颜色配置
	if param.ChartSeriesGuid != "" {
		configSeriesRows, getErr := db.GetChartSeriesConfig(param.ChartSeriesGuid)
		if getErr != nil {
			middleware.ReturnServerHandleError(c, getErr)
			return
		}
		for _, row := range configSeriesRows {
			existSeriesMap[row.SeriesName] = row.Color
		}
		for _, v := range result {
			if color, ok := existSeriesMap[v.SeriesName]; ok {
				v.New = false
				v.Color = color
			}
		}
	}
	middleware.ReturnSuccessData(c, result)
}
