package monitor

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
)

// QueryCustomDashboardList 查询自定义看板列表
func QueryCustomDashboardList(c *gin.Context) {
	var param models.CustomDashboardQueryParam
	var err error
	var pageInfo models.PageInfo
	var rowsData []*models.CustomDashboardResultDto
	var list []*models.CustomDashboardTable
	var roleRelList []*models.CustomDashBoardRoleRel
	var mainDashBoardList []*models.MainDashboard
	var mgmtRoles, useRoles, mainPages []string
	var displayNameRoleMap map[string]string
	var userRoleMap map[string]bool
	var permission string
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if pageInfo, list, err = db.QueryCustomDashboardList(param, middleware.GetOperateUserRoles(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if displayNameRoleMap, err = db.QueryAllRoleDisplayNameMap(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	userRoleMap = db.TransformArrayToMap(middleware.GetOperateUserRoles(c))
	if len(list) > 0 {
		for _, dashboard := range list {
			mgmtRoles = []string{}
			useRoles = []string{}
			mainPages = []string{}
			permission = string(models.PermissionUse)
			if roleRelList, err = db.QueryCustomDashboardRoleRelByCustomDashboard(dashboard.Id); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if mainDashBoardList, err = db.QueryMainDashboardByCustomDashboard(dashboard.Id); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if len(roleRelList) > 0 {
				for _, roleRel := range roleRelList {
					if roleRel.Permission == string(models.PermissionMgmt) {
						if v, ok := displayNameRoleMap[roleRel.RoleId]; ok {
							mgmtRoles = append(mgmtRoles, v)
						}
					} else if roleRel.Permission == string(models.PermissionUse) {
						if v, ok := displayNameRoleMap[roleRel.RoleId]; ok {
							useRoles = append(useRoles, v)
						}
					}
					if userRoleMap[roleRel.RoleId] {
						permission = string(models.PermissionMgmt)
					}
				}
			}
			if len(mainDashBoardList) > 0 {
				for _, mainDashBoard := range mainDashBoardList {
					if v, ok := displayNameRoleMap[mainDashBoard.RoleId]; ok {
						mainPages = append(mainPages, v)
					}
				}
			}
			result := &models.CustomDashboardResultDto{
				Id:         dashboard.Id,
				Name:       dashboard.Name,
				MgmtRoles:  mgmtRoles,
				UseRoles:   useRoles,
				Permission: permission,
				CreateUser: dashboard.CreateUser,
				UpdateUser: dashboard.UpdateUser,
				MainPage:   mainPages,
			}
			rowsData = append(rowsData, result)
		}
	}
	middleware.ReturnPageData(c, pageInfo, rowsData)
}

func GetCustomDashboard(c *gin.Context) {
	var err error
	var customDashboard *models.CustomDashboardTable
	var customDashboardDto *models.CustomDashboardDto
	var customChartExtendList []*models.CustomChartExtend
	var groupMap = make(map[string]bool)
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	id, _ := strconv.Atoi(c.Query("id"))
	if id == 0 {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	// 获取自定义看板
	if customDashboard, err = db.GetCustomDashboardById(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboard == nil {
		middleware.ReturnValidateError(c, "id is invalid")
		return
	}
	if customChartExtendList, err = db.QueryCustomChartListByDashboard(customDashboard.Id); err != nil {
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
	if len(customChartExtendList) > 0 {
		customDashboardDto.Charts = []*models.CustomChartDto{}
		for _, chartExtend := range customChartExtendList {
			if chartExtend.CustomChart == nil {
				continue
			}
			groupMap[chartExtend.Group] = true
			chart, err2 := createCustomChartDto(chartExtend, configMap, tagMap, tagValueMap)
			if err2 != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if chart != nil {
				customDashboardDto.Charts = append(customDashboardDto.Charts, chart)
			}
		}
		customDashboardDto.PanelGroupList = db.TransformMapToArray(groupMap)
	}
	middleware.ReturnData(c, customDashboardDto)
}

func intToBool(num int) bool {
	return num != 0
}

func createCustomChartDto(chartExtend *models.CustomChartExtend, configMap map[string][]*models.CustomChartSeriesConfig, tagMap map[string][]*models.CustomChartSeriesTag, tagValueMap map[string][]*models.CustomChartSeriesTagValue) (chart *models.CustomChartDto, err error) {
	var list []*models.CustomChartSeries
	var seriesConfigList []*models.CustomChartSeriesConfig
	var chartSeriesTagList []*models.CustomChartSeriesTag
	var chartSeriesTagValueList []*models.CustomChartSeriesTagValue
	chart = &models.CustomChartDto{
		Id:              chartExtend.CustomChart.Guid,
		Public:          intToBool(chartExtend.CustomChart.Public),
		SourceDashboard: chartExtend.CustomChart.SourceDashboard,
		Name:            chartExtend.CustomChart.Name,
		Unit:            chartExtend.CustomChart.Unit,
		ChartType:       chartExtend.CustomChart.ChartType,
		LineType:        chartExtend.CustomChart.LineType,
		Aggregate:       chartExtend.CustomChart.Aggregate,
		AggStep:         chartExtend.CustomChart.AggStep,
		Query:           nil,
		DisplayConfig:   chartExtend.DisplayConfig,
		Group:           chartExtend.Group,
	}
	chart.Query = []*models.CustomChartSeriesDto{}
	if list, err = db.QueryCustomChartSeriesByChart(chartExtend.CustomChart.Guid); err != nil {
		return
	}
	if len(list) > 0 {
		for _, series := range list {
			seriesConfigList = []*models.CustomChartSeriesConfig{}
			chartSeriesTagList = []*models.CustomChartSeriesTag{}
			customChartSeriesDto := &models.CustomChartSeriesDto{
				Endpoint:     series.Endpoint,
				ServiceGroup: series.ServiceGroup,
				EndpointName: series.EndpointName,
				MonitorType:  series.MonitorType,
				ColorGroup:   series.ColorGroup,
			}
			customChartSeriesDto.Metrics = []*models.MetricDto{}
			if v, ok := configMap[series.Guid]; ok {
				seriesConfigList = v
			}
			if v, ok := tagMap[series.Guid]; ok {
				chartSeriesTagList = v
			}
			customChartSeriesDto.Metrics = append(customChartSeriesDto.Metrics, &models.MetricDto{
				Metric:      series.Metric,
				Tags:        make([]*models.TagDto, 0),
				ColorConfig: make([]*models.ColorConfigDto, 0),
			})
			if len(chartSeriesTagList) > 0 {
				for _, tag := range chartSeriesTagList {
					chartSeriesTagValueList = []*models.CustomChartSeriesTagValue{}
					if v, ok := tagValueMap[tag.Guid]; ok {
						chartSeriesTagValueList = v
					}
					customChartSeriesDto.Metrics[0].Tags = append(customChartSeriesDto.Metrics[0].Tags, &models.TagDto{
						TagName:  tag.Name,
						TagValue: getChartSeriesTagValues(chartSeriesTagValueList),
					})
				}
			}
			if len(seriesConfigList) > 0 {
				for _, config := range seriesConfigList {
					customChartSeriesDto.Metrics[0].ColorConfig = append(customChartSeriesDto.Metrics[0].ColorConfig, &models.ColorConfigDto{
						Metric: series.Metric,
						Color:  config.Color,
					})
				}
			}
			chart.Query = append(chart.Query, customChartSeriesDto)
		}
	}
	return
}

func getChartSeriesTagValues(chartSeriesTagValueList []*models.CustomChartSeriesTagValue) []string {
	var result []string
	if len(chartSeriesTagValueList) > 0 {
		for _, tagValue := range chartSeriesTagValueList {
			result = append(result, tagValue.Value)
		}
	}
	return result
}
