package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

// GetAllCustomDashboardList 获取所有看板(包括源看板、应用看板)
func GetAllCustomDashboardList(c *gin.Context) {
	var list []*models.SimpleCustomDashboardDto
	var err error
	if list, err = db.QueryAllCustomDashboard(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, list)
}

// QueryCustomDashboardList 查询自定义看板列表
func QueryCustomDashboardList(c *gin.Context) {
	var param models.CustomDashboardQueryParam
	var err error
	var pageInfo models.PageInfo
	var rowsData []*models.CustomDashboardResultDto
	var list []*models.CustomDashboardTable
	var roleRelList []*models.CustomDashBoardRoleRel
	var mainDashBoardList []*models.MainDashboard
	var mgmtRoles, displayMgmtRoles, useRoles, displayUseRoles, mainPages []string
	var displayNameRoleMap map[string]string
	var userRoleMap map[string]bool
	var permission string
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.PageSize == 0 {
		param.PageSize = 10
	}
	if pageInfo, list, err = db.QueryCustomDashboardList(param, middleware.GetOperateUser(c), middleware.GetOperateUserRoles(c)); err != nil {
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
			displayMgmtRoles = []string{}
			displayUseRoles = []string{}
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
			if len(mainDashBoardList) > 0 {
				for _, mainDashBoard := range mainDashBoardList {
					if v, ok := displayNameRoleMap[mainDashBoard.RoleId]; ok {
						mainPages = append(mainPages, v)
					}
				}
			}
			result := &models.CustomDashboardResultDto{
				Id:               dashboard.Id,
				Name:             dashboard.Name,
				MgmtRoles:        mgmtRoles,
				DisplayMgmtRoles: displayMgmtRoles,
				UseRoles:         useRoles,
				DisplayUseRoles:  displayUseRoles,
				Permission:       permission,
				CreateUser:       dashboard.CreateUser,
				UpdateUser:       dashboard.UpdateUser,
				UpdateTime:       dashboard.UpdateAt.Format(models.DatetimeFormat),
				MainPage:         mainPages,
			}
			if dashboard.LogMetricGroup != nil {
				result.LogMetricGroup = *dashboard.LogMetricGroup
			}
			rowsData = append(rowsData, result)
		}
	}
	middleware.ReturnPageData(c, pageInfo, rowsData)
}

func GetCustomDashboard(c *gin.Context) {
	var err error
	var customDashboard *models.CustomDashboardTable
	var customDashboardDto = &models.CustomDashboardDto{UseRoles: []string{}, MgmtRoles: []string{}}
	var customChartExtendList []*models.CustomChartExtend
	var groupMap = make(map[string]bool)
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	var boardRoleRelList []*models.CustomDashBoardRoleRel
	var metricComparisonMap = make(map[string]string)
	var chartSeriesMap = make(map[string][]*models.CustomChartSeries)
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
	if customDashboard == nil || customDashboard.Id == 0 {
		middleware.ReturnValidateError(c, "id is invalid")
		return
	}
	customDashboardDto.Name = customDashboard.Name
	customDashboardDto.TimeRange = customDashboard.TimeRange
	customDashboardDto.RefreshWeek = customDashboard.RefreshWeek
	if customDashboard.LogMetricGroup != nil {
		customDashboardDto.LogMetricGroup = *customDashboard.LogMetricGroup
	}
	if customChartExtendList, err = db.QueryCustomChartListByDashboard(customDashboard.Id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if metricComparisonMap, err = db.GetAllMetricComparison(); err != nil {
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
		var chartSeries []*models.CustomChartSeries
		// 图表大于等于10时候 查询所有图表数据
		if len(customChartExtendList) >= 10 {
			if chartSeriesMap, err = db.QueryAllChartSeries(); err != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
		}
		customDashboardDto.Charts = []*models.CustomChartDto{}
		for _, chartExtend := range customChartExtendList {
			groupMap[chartExtend.Group] = true
			if len(chartSeriesMap) > 0 {
				chartSeries = chartSeriesMap[chartExtend.Guid]
			}
			chartParam := models.CreateCustomChartParam{
				ChartExtend:         chartExtend,
				ConfigMap:           configMap,
				TagMap:              tagMap,
				TagValueMap:         tagValueMap,
				MetricComparisonMap: metricComparisonMap,
				ChartSeries:         chartSeries,
			}
			chart, err2 := db.CreateCustomChartDto(chartParam)
			if err2 != nil {
				middleware.ReturnServerHandleError(c, err)
				return
			}
			if chart != nil {
				customDashboardDto.Charts = append(customDashboardDto.Charts, chart)
			}
		}
	}
	if strings.TrimSpace(customDashboard.PanelGroups) == "" {
		customDashboardDto.PanelGroupList = db.TransformMapToArray(groupMap)
	} else {
		customDashboardDto.PanelGroupList = strings.Split(customDashboard.PanelGroups, ",")
	}
	if boardRoleRelList, err = db.QueryCustomDashboardPermissionByDashboard(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(boardRoleRelList) > 0 {
		for _, rel := range boardRoleRelList {
			if rel.Permission == string(models.PermissionUse) {
				customDashboardDto.UseRoles = append(customDashboardDto.UseRoles, rel.RoleId)
			} else if rel.Permission == string(models.PermissionMgmt) {
				customDashboardDto.MgmtRoles = append(customDashboardDto.MgmtRoles, rel.RoleId)
			}
		}
	}
	middleware.ReturnSuccessData(c, customDashboardDto)
}

// AddCustomDashboard 新增自定义看板
func AddCustomDashboard(c *gin.Context) {
	var err error
	var param models.AddCustomDashboardParam
	var customDashboardList []*models.CustomDashboardTable
	var dashboardId int64
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(param.Name) == "" {
		middleware.ReturnParamEmptyError(c, "name")
		return
	}
	if len(param.MgmtRoles) != 1 {
		middleware.ReturnValidateError(c, "mgmtRoles error")
		return
	}
	if len(param.UseRoles) == 0 {
		middleware.ReturnParamEmptyError(c, "useRoles")
		return
	}
	// 查询名称是否重复
	if customDashboardList, err = db.QueryCustomDashboardListByName(param.Name); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(customDashboardList) > 0 {
		middleware.ReturnDashboardNameRepeatError(c)
		return
	}
	now := time.Now()
	user := middleware.GetOperateUser(c)
	dashboard := &models.CustomDashboardTable{
		Name:        param.Name,
		CreateUser:  user,
		UpdateUser:  user,
		CreateAt:    now,
		UpdateAt:    now,
		RefreshWeek: 60,
		TimeRange:   -1800,
	}
	if dashboardId, err = db.AddCustomDashboard(dashboard, param.MgmtRoles, param.UseRoles); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	dashboard.Id = int(dashboardId)
	middleware.ReturnSuccessData(c, dashboard)
}

// DeleteCustomDashboard 删除自定义看板
func DeleteCustomDashboard(c *gin.Context) {
	var err error
	var id int
	var permission bool
	id, err = strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		middleware.ReturnParamTypeError(c, "id", "int")
		return
	}
	if permission, err = CheckHasDashboardManagePermission(id, middleware.GetOperateUserRoles(c), middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("no delete permission"))
	}
	if err = db.DeleteCustomDashboardById(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func CopyCustomDashboard(c *gin.Context) {
	var customDashboard *models.CustomDashboardTable
	var param models.CopyCustomDashboardParam
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.DashboardId == 0 {
		middleware.ReturnParamEmptyError(c, "dashboardId")
		return
	}
	if strings.TrimSpace(param.MgmtRole) == "" {
		middleware.ReturnValidateError(c, "mgmtRoles empty!")
		return
	}
	if len(param.UseRoles) == 0 {
		middleware.ReturnParamEmptyError(c, "useRoles empty!")
		return
	}
	if customDashboard, err = db.GetCustomDashboardById(param.DashboardId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboard == nil {
		middleware.ReturnValidateError(c, "invalid id")
		return
	}
	if err = db.CopyCustomDashboard(param, customDashboard, middleware.GetOperateUser(c), middleware.GetMessageMap(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

// UpdateCustomDashboard 修改自定义看板
func UpdateCustomDashboard(c *gin.Context) {
	var err error
	var param models.UpdateCustomDashboardParam
	var hasChartRelList, insertChartRelList, updateChartRelList []*models.CustomDashboardChartRel
	var deleteChartRelIds []string
	var insert, delete, permission bool
	var actions []*db.Action
	var nameMap = make(map[string]bool)
	var panelGroups string
	var customDashboardList []*models.CustomDashboardTable
	user := middleware.GetOperateUser(c)
	now := time.Now().Format(models.DatetimeFormat)
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.Id == 0 {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	if strings.TrimSpace(param.Name) == "" {
		middleware.ReturnParamEmptyError(c, "name")
		return
	}
	// 查询名称是否重复
	if customDashboardList, err = db.QueryCustomDashboardListByName(param.Name); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(customDashboardList) > 0 {
		for _, customDashboard := range customDashboardList {
			if customDashboard.Id != param.Id {
				middleware.ReturnDashboardNameRepeatError(c)
				return
			}
		}
	}
	if permission, err = CheckHasDashboardManagePermission(param.Id, middleware.GetOperateUserRoles(c), middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("no edit permission"))
		return
	}
	if len(param.Charts) > 0 {
		for _, chart := range param.Charts {
			if nameMap[chart.Name] {
				middleware.ReturnValidateError(c, fmt.Sprintf("chart name:%s repeat", chart.Name))
				return
			}
			nameMap[chart.Name] = true
		}
	}
	if hasChartRelList, err = db.QueryCustomDashboardChartRelListByDashboard(param.Id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(param.Charts) == 0 {
		param.Charts = []*models.CustomChartDto{}
	}
	if len(hasChartRelList) == 0 {
		hasChartRelList = []*models.CustomDashboardChartRel{}
	}

	for _, chart := range param.Charts {
		insert = true
		for _, chartRel := range hasChartRelList {
			if chartRel.DashboardChart != nil && *chartRel.DashboardChart == chart.Id {
				displayConfig, _ := json.Marshal(chart.DisplayConfig)
				groupDisplayConfig, _ := json.Marshal(chart.GroupDisplayConfig)
				updateChartRelList = append(updateChartRelList, &models.CustomDashboardChartRel{
					Guid:               chartRel.Guid,
					Group:              chart.Group,
					DisplayConfig:      string(displayConfig),
					GroupDisplayConfig: string(groupDisplayConfig),
					UpdateUser:         user,
					UpdateTime:         now,
				})
				insert = false
				break
			}
		}
		if insert {
			displayConfig, _ := json.Marshal(chart.DisplayConfig)
			groupDisplayConfig, _ := json.Marshal(chart.GroupDisplayConfig)
			insertChartRelList = append(insertChartRelList, &models.CustomDashboardChartRel{
				Guid:               guid.CreateGuid(),
				CustomDashboard:    &param.Id,
				DashboardChart:     &chart.Id,
				Group:              chart.Group,
				DisplayConfig:      string(displayConfig),
				GroupDisplayConfig: string(groupDisplayConfig),
				CreateUser:         user,
				UpdateUser:         user,
				CreateTime:         now,
				UpdateTime:         now,
			})
		}
	}

	for _, chartRel := range hasChartRelList {
		delete = true
		for _, chart := range param.Charts {
			if chart.Id == *chartRel.DashboardChart {
				delete = false
				break
			}
		}
		if delete {
			deleteChartRelIds = append(deleteChartRelIds, chartRel.Guid)
		}
	}
	if len(insertChartRelList) > 0 {
		actions = append(actions, db.GetAddCustomDashboardChartRelSQL(insertChartRelList)...)
	}
	if len(updateChartRelList) > 0 {
		actions = append(actions, db.GetUpdateCustomDashboardChartRelSQL(updateChartRelList)...)
	}
	if len(deleteChartRelIds) > 0 {
		actions = append(actions, db.GetDeleteCustomDashboardChartRelSQL(deleteChartRelIds)...)
	}
	if len(param.PanelGroups) > 0 {
		panelGroups = strings.Join(param.PanelGroups, ",")
	}
	actions = append(actions, db.GetUpdateCustomDashboardSQL(param.Name, panelGroups, middleware.GetOperateUser(c), param.TimeRange, param.RefreshWeek, param.Id)...)
	if err = db.Transaction(actions); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

// UpdateCustomDashboardPermission 修改自定义面板权限
func UpdateCustomDashboardPermission(c *gin.Context) {
	var err error
	var param models.UpdateCustomDashboardPermissionParam
	var actions, deleteActions, subActions, updateActions []*db.Action
	var permission bool
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(param.MgmtRoles) != 1 {
		middleware.ReturnValidateError(c, "mgmtRoles error")
		return
	}
	if len(param.UseRoles) == 0 {
		middleware.ReturnParamEmptyError(c, "useRoles is empty")
		return
	}
	if permission, err = CheckHasDashboardManagePermission(param.Id, middleware.GetOperateUserRoles(c), middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if !permission {
		middleware.ReturnServerHandleError(c, fmt.Errorf("not has edit permission"))
	}
	deleteActions = db.GetDeleteCustomDashboardRoleRelSQL(param.Id)
	if len(deleteActions) > 0 {
		actions = append(actions, deleteActions...)
	}
	subActions = db.GetInsertCustomDashboardRoleRelSQL(param.Id, param.MgmtRoles, param.UseRoles)
	if len(subActions) > 0 {
		actions = append(actions, subActions...)
	}
	updateActions = db.UpdateCustomDashboardTimeActions(param.Id, middleware.GetOperateUser(c))
	actions = append(actions, updateActions...)
	if err = db.Transaction(actions); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func CheckHasDashboardManagePermission(dashboard int, userRoles []string, user string) (permission bool, err error) {
	var permissionMap map[string]string
	var customDashboard *models.CustomDashboardTable
	if len(userRoles) == 0 {
		return
	}
	// 判断是否拥有删除权限
	if permissionMap, err = db.QueryCustomDashboardManagePermissionByDashboard(dashboard); err != nil {
		return
	}
	if len(permissionMap) == 0 {
		permissionMap = make(map[string]string)
	}
	for _, role := range userRoles {
		if v, ok := permissionMap[role]; ok && v == string(models.PermissionMgmt) {
			permission = true
			break
		}
	}
	if !permission && user != "" {
		if customDashboard, err = db.GetCustomDashboardById(dashboard); err != nil {
			return
		}
		if customDashboard != nil && user == customDashboard.CreateUser {
			permission = true
			return
		}
	}
	return
}

func SyncData(c *gin.Context) {
	err := db.SyncData()
	if err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

// ExportCustomDashboard 看板导出
func ExportCustomDashboard(c *gin.Context) {
	var err error
	var param models.CustomDashboardExportParam
	var result *models.CustomDashboardExportDto
	var customDashboard *models.CustomDashboardTable
	var customChartExtendList []*models.CustomChartExtend
	var configMap = make(map[string][]*models.CustomChartSeriesConfig)
	var tagMap = make(map[string][]*models.CustomChartSeriesTag)
	var tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	var exportChartIdMap = make(map[string]bool)
	var dashboardPermissionList []*models.CustomDashBoardRoleRel
	var metricComparisonMap = make(map[string]string)
	var useRoles []string
	var mgmtRole string
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if param.Id == 0 {
		middleware.ReturnParamEmptyError(c, "param id")
		return
	}
	// 获取自定义看板
	if customDashboard, err = db.GetCustomDashboardById(param.Id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboard == nil || customDashboard.Id == 0 {
		middleware.ReturnValidateError(c, "id is invalid")
		return
	}
	if len(param.ChartIds) > 0 {
		for _, id := range param.ChartIds {
			exportChartIdMap[id] = true
		}
	}
	if dashboardPermissionList, err = db.QueryCustomDashboardRoleRelByCustomDashboard(customDashboard.Id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	for _, role := range dashboardPermissionList {
		if role.Permission == string(models.PermissionMgmt) {
			mgmtRole = role.RoleId
		} else if role.Permission == string(models.PermissionUse) {
			useRoles = append(useRoles, role.RoleId)
		}
	}
	result = &models.CustomDashboardExportDto{
		Id:          customDashboard.Id,
		Name:        customDashboard.Name,
		PanelGroups: customDashboard.PanelGroups,
		TimeRange:   customDashboard.TimeRange,
		RefreshWeek: customDashboard.RefreshWeek,
		Charts:      []*models.CustomChartDto{},
		UseRoles:    useRoles,
		MgmtRole:    mgmtRole,
	}
	if customDashboard.LogMetricGroup != nil {
		result.LogMetricGroup = *customDashboard.LogMetricGroup
	}
	if customChartExtendList, err = db.QueryCustomChartListByDashboard(customDashboard.Id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if metricComparisonMap, err = db.GetAllMetricComparison(); err != nil {
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
		for _, chartExtend := range customChartExtendList {
			chartParam := models.CreateCustomChartParam{
				ChartExtend:         chartExtend,
				ConfigMap:           configMap,
				TagMap:              tagMap,
				TagValueMap:         tagValueMap,
				MetricComparisonMap: metricComparisonMap,
				ChartSeries:         []*models.CustomChartSeries{},
			}
			// 只导出指定图表数据
			if exportChartIdMap[chartExtend.Guid] {
				chart, err2 := db.CreateCustomChartDto(chartParam)
				if err2 != nil {
					middleware.ReturnServerHandleError(c, err)
					return
				}
				if chart != nil {
					result.Charts = append(result.Charts, chart)
				}
			}
		}
	}
	b, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		middleware.ReturnHandleError(c, "export custom dashboard fail, json marshal object error", marshalErr)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%d_%s.json", result.Id, time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func ImportCustomDashboard(c *gin.Context) {
	var param *models.CustomDashboardExportDto
	var customDashboard *models.CustomDashboardTable
	var importRes *models.CustomDashboardImportRes
	var customDashboardList []*models.CustomDashboardTable
	var permissionMap map[string]bool
	var hasPerm bool
	rule, _ := c.GetPostForm("rule")
	useRoleStr, _ := c.GetPostForm("useRoles")
	mgmtRole, _ := c.GetPostForm("mgmtRoles")
	file, err := c.FormFile("file")
	if rule == "" || len(useRoleStr) == 0 || mgmtRole == "" {
		middleware.ReturnParamEmptyError(c, "rule or permission")
		return
	}
	useRoles := strings.Split(useRoleStr, ",")
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	f, err := file.Open()
	if err != nil {
		middleware.ReturnHandleError(c, "file open error ", err)
		return
	}
	b, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		middleware.ReturnHandleError(c, "read content fail error ", err)
		return
	}
	err = json.Unmarshal(b, &param)
	if err != nil {
		middleware.ReturnHandleError(c, "json unmarshal fail error ", err)
		return
	}
	if param == nil || strings.TrimSpace(param.Name) == "" {
		middleware.ReturnParamEmptyError(c, "import data is empty")
		return
	}
	if len(param.Charts) == 0 {
		middleware.ReturnParamEmptyError(c, "import dashboard chart is empty")
		return
	}
	// 判断操作人是否有覆盖看板权限
	if customDashboardList, err = db.QueryCustomDashboardListByName(param.Name); err != nil {
		return
	}
	if rule == string(models.ImportRuleCover) && len(customDashboardList) > 0 {
		if permissionMap, err = db.GetDashboardPermissionMap(customDashboardList[0].Id, string(models.PermissionMgmt)); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
		if len(permissionMap) == 0 {
			permissionMap = make(map[string]bool)
		}
		for _, userRole := range middleware.GetOperateUserRoles(c) {
			if permissionMap[userRole] {
				hasPerm = true
			}
		}
		if !hasPerm {
			middleware.ReturnServerHandleError(c, fmt.Errorf("dashboard %s no edit permission", param.Name))
			return
		}
	}
	errMsgObj := middleware.GetMessageMap(c)
	if customDashboard, importRes, err = db.ImportCustomDashboard(param, middleware.GetOperateUser(c), rule, mgmtRole, useRoles, errMsgObj); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboard != nil && customDashboard.Id != 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).DashboardIdExistError))
		return
	}
	if importRes != nil && len(importRes.ChartMap) > 0 {
		middleware.ReturnSuccessData(c, importRes.ChartMap)
		return
	}
	middleware.ReturnSuccess(c)
}

func TransImportCustomDashboard(c *gin.Context) {
	var param *models.CustomDashboardExportDto
	var customDashboard *models.CustomDashboardTable
	file, err := c.FormFile("file")
	f, err := file.Open()
	if err != nil {
		middleware.ReturnHandleError(c, "file open error ", err)
		return
	}
	b, err := io.ReadAll(f)
	defer f.Close()
	if err != nil {
		middleware.ReturnHandleError(c, "read content fail error ", err)
		return
	}
	err = json.Unmarshal(b, &param)
	if param == nil || strings.TrimSpace(param.Name) == "" {
		middleware.ReturnParamEmptyError(c, "import data is empty")
		return
	}
	if len(param.Charts) == 0 {
		middleware.ReturnParamEmptyError(c, "import dashboard chart is empty")
		return
	}
	if customDashboard, _, err = db.ImportCustomDashboard(param, middleware.GetOperateUser(c), "cover", param.MgmtRole, param.UseRoles, middleware.GetMessageMap(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if customDashboard != nil && customDashboard.Id != 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).DashboardIdExistError))
		return
	}
	middleware.ReturnSuccess(c)
}

func BatchGetDashboard(c *gin.Context) {
	var param models.IdsParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result, err := db.BatchGetCustomDashboardByIds(param.Ids)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}
