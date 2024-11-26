package db

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

const (
	constOtherCode      = "{code}"
	constOther          = "other"
	constCode           = "code"
	constRetCode        = "retcode"
	ConstEqualIn        = "in"
	ConstEqualNotIn     = "notin"
	constReqCount       = "req_count"
	constReqSuccessRate = "req_suc_rate"
	constReqFailCount   = "req_fail_count"
	constReqFailRate    = "req_fail_rate"
	constReqSucCount    = "req_suc_rate"
	constConstTimeAvg   = "req_costtime_avg"
	constConstTimeMax   = "req_costtime_max"
	constSuccess        = "success"
)

func autoGenerateAlarmStrategy(alarmStrategyParam models.AutoAlarmStrategyParam) (actions []*Action, result []string, err error) {
	var subActions []*Action
	var serviceGroupTable models.ServiceGroupTable
	// 显示名
	var displayServiceGroup = alarmStrategyParam.ServiceGroup
	result = []string{}
	actions = []*Action{}
	// 自动创建告警
	if alarmStrategyParam.AutoCreateWarn {
		x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", alarmStrategyParam.ServiceGroup).Get(&serviceGroupTable)
		if serviceGroupTable.DisplayName != "" {
			displayServiceGroup = serviceGroupTable.DisplayName
		}
		codeList := getTargetCodeMap(alarmStrategyParam.CodeStringMap)
		autoAlarmMetricList := getAutoAlarmMetricList(alarmStrategyParam.MetricList, alarmStrategyParam.ServiceGroup, alarmStrategyParam.MetricPrefixCode)
		// 添加 other默认告警
		codeList = append(codeList, constOtherCode)
		for _, code := range codeList {
			for _, alarmMetric := range autoAlarmMetricList {
				if !alarmMetric.AutoWarn {
					continue
				}
				// 添加告警配置基础信息
				alarmStrategy := &models.GroupStrategyObj{NotifyList: make([]*models.NotifyObj, 0), Conditions: make([]*models.StrategyConditionObj, 0)}
				metricTags := make([]*models.MetricTag, 0)
				alarmStrategy.Name = fmt.Sprintf("%s-%s%s%s%s", code, generateMetricGuidDisplayName(alarmStrategyParam.MetricPrefixCode, alarmMetric.Metric, displayServiceGroup), alarmMetric.Operator, alarmMetric.Threshold, getAlarmMetricUnit(alarmMetric.Metric))
				if code == constOtherCode {
					alarmStrategy.Priority = "medium"
				} else {
					alarmStrategy.Priority = "high"
				}
				alarmStrategy.NotifyEnable = 1
				alarmStrategy.ActiveWindow = "00:00-23:59"
				if strings.TrimSpace(alarmStrategyParam.EndpointGroup) != "" {
					alarmStrategy.EndpointGroup = alarmStrategyParam.EndpointGroup
				}
				alarmStrategy.LogMetricGroup = &alarmStrategyParam.LogMetricGroupGuid
				alarmStrategy.Metric = alarmMetric.MetricId
				alarmStrategy.MetricName = alarmMetric.Metric
				alarmStrategy.Content = fmt.Sprintf("%s continuing for more than %s%s", alarmStrategy.Name, alarmMetric.Time, alarmMetric.TimeUnit)
				// 添加编排与通知
				alarmStrategy.NotifyList = append(alarmStrategy.NotifyList, &models.NotifyObj{AlarmAction: "firing", NotifyRoles: alarmStrategyParam.ServiceGroupsRoles})
				alarmStrategy.NotifyList = append(alarmStrategy.NotifyList, &models.NotifyObj{AlarmAction: "ok", NotifyRoles: alarmStrategyParam.ServiceGroupsRoles})
				result = append(result, alarmStrategy.Name)
				// 添加指标阈值
				for _, tag := range alarmMetric.TagConfig {
					// 标签为code,需要配置 equal和TagValue值
					if tag == constCode {
						// code为 other 配置not in,其他配置 in
						if code == constOtherCode {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  constCode,
								Equal:    ConstEqualNotIn,
								TagValue: codeList[:len(codeList)-1],
							})
						} else {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  constCode,
								Equal:    ConstEqualIn,
								TagValue: []string{code},
							})
						}
					} else {
						// 平均耗时&最大耗时,只会统计成功请求的平均耗时
						if tag == constRetCode && (strings.HasSuffix(alarmMetric.Metric, constConstTimeAvg) || strings.HasSuffix(alarmMetric.Metric, constConstTimeMax)) {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  constRetCode,
								Equal:    ConstEqualIn,
								TagValue: []string{getRetCodeSuccessCode(alarmStrategyParam.RetCodeStringMap)},
							})
						} else {
							metricTags = append(metricTags, &models.MetricTag{
								TagName: tag,
							})
						}
					}
				}

				alarmStrategy.Conditions = append(alarmStrategy.Conditions, &models.StrategyConditionObj{
					Metric:     alarmMetric.MetricId,
					MetricName: alarmMetric.Metric,
					Condition:  fmt.Sprintf("%s%s", alarmMetric.Operator, alarmMetric.Threshold),
					Last:       fmt.Sprintf("%s%s", alarmMetric.Time, alarmMetric.TimeUnit),
					Tags:       metricTags,
				})
				alarmStrategy.Condition = fmt.Sprintf("%s%s", alarmMetric.Operator, alarmMetric.Threshold)
				alarmStrategy.Last = fmt.Sprintf("%s%s", alarmMetric.Time, alarmMetric.TimeUnit)
				if subActions, err = getCreateAlarmStrategyActions(alarmStrategy, time.Now().Format(models.DatetimeFormat), alarmStrategyParam.Operator); err != nil {
					return
				}
				if len(subActions) > 0 {
					actions = append(actions, subActions...)
				}
			}
		}
	}
	return
}

func autoGenerateSimpleAlarmStrategy(alarmStrategyParam models.AutoSimpleAlarmStrategyParam) (actions []*Action, result []string, err error) {
	var subActions []*Action
	var serviceGroupTable models.ServiceGroupTable
	// 显示名
	var displayServiceGroup = alarmStrategyParam.ServiceGroup
	result = []string{}
	actions = []*Action{}
	// 自动创建告警
	if alarmStrategyParam.AutoCreateWarn {
		x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", alarmStrategyParam.ServiceGroup).Get(&serviceGroupTable)
		if serviceGroupTable.DisplayName != "" {
			displayServiceGroup = serviceGroupTable.DisplayName
		}
		autoAlarmMetricList := getAutoSimpleAlarmMetricList(alarmStrategyParam.MetricList, alarmStrategyParam.ServiceGroup, alarmStrategyParam.MetricPrefixCode)
		for _, alarmMetric := range autoAlarmMetricList {
			if !alarmMetric.AutoWarn {
				continue
			}
			// 添加告警配置基础信息
			alarmStrategy := &models.GroupStrategyObj{NotifyList: make([]*models.NotifyObj, 0), Conditions: make([]*models.StrategyConditionObj, 0)}
			metricTags := make([]*models.MetricTag, 0)
			alarmStrategy.Name = fmt.Sprintf("%s%s%s%s", generateMetricGuidDisplayName(alarmStrategyParam.MetricPrefixCode, alarmMetric.Metric, displayServiceGroup), alarmMetric.Operator, alarmMetric.Threshold, getAlarmMetricUnit(alarmMetric.Metric))
			alarmStrategy.Priority = "medium"
			alarmStrategy.NotifyEnable = 1
			alarmStrategy.ActiveWindow = "00:00-23:59"
			if strings.TrimSpace(alarmStrategyParam.EndpointGroup) != "" {
				alarmStrategy.EndpointGroup = alarmStrategyParam.EndpointGroup
			}
			alarmStrategy.LogMetricGroup = &alarmStrategyParam.LogMetricGroupGuid
			alarmStrategy.Metric = alarmMetric.MetricId
			alarmStrategy.MetricName = alarmMetric.Metric
			alarmStrategy.Content = fmt.Sprintf("%s continuing for more than %s%s", alarmStrategy.Name, alarmMetric.Time, alarmMetric.TimeUnit)
			// 添加编排与通知
			alarmStrategy.NotifyList = append(alarmStrategy.NotifyList, &models.NotifyObj{AlarmAction: "firing", NotifyRoles: alarmStrategyParam.ServiceGroupsRoles})
			alarmStrategy.NotifyList = append(alarmStrategy.NotifyList, &models.NotifyObj{AlarmAction: "ok", NotifyRoles: alarmStrategyParam.ServiceGroupsRoles})
			result = append(result, alarmStrategy.Name)
			// 添加指标阈值
			for _, tag := range alarmMetric.TagConfig {
				metricTags = append(metricTags, &models.MetricTag{
					TagName: tag,
					Equal:   ConstEqualIn,
				})
			}

			alarmStrategy.Conditions = append(alarmStrategy.Conditions, &models.StrategyConditionObj{
				Metric:     alarmMetric.MetricId,
				MetricName: alarmMetric.Metric,
				Condition:  fmt.Sprintf("%s%s", alarmMetric.Operator, alarmMetric.Threshold),
				Last:       fmt.Sprintf("%s%s", alarmMetric.Time, alarmMetric.TimeUnit),
				Tags:       metricTags,
				LogType:    alarmStrategyParam.LogType,
			})
			alarmStrategy.Condition = fmt.Sprintf("%s%s", alarmMetric.Operator, alarmMetric.Threshold)
			alarmStrategy.Last = fmt.Sprintf("%s%s", alarmMetric.Time, alarmMetric.TimeUnit)
			if subActions, err = getCreateAlarmStrategyActions(alarmStrategy, time.Now().Format(models.DatetimeFormat), alarmStrategyParam.Operator); err != nil {
				return
			}
			if len(subActions) > 0 {
				actions = append(actions, subActions...)
			}
		}
	}
	return
}

func getRetCodeSuccessCode(stringMap []*models.LogMetricStringMapTable) string {
	if len(stringMap) == 0 {
		return ""
	}
	for _, table := range stringMap {
		if table.ValueType == constSuccess {
			return table.TargetValue
		}
	}
	return ""
}

func autoGenerateCustomDashboard(dashboardParam models.AutoCreateDashboardParam) (actions []*Action, customDashboard string, newDashboardId int64, err error) {
	var subDashboardActions, subChart1Actions, subChart2Actions, subChart3Actions []*Action
	var metricMap = getMetricMap(dashboardParam.MetricList, dashboardParam.MetricPrefixCode, dashboardParam.ServiceGroup)
	var reqCountMetric, failCountMetric, sucRateMetric, costTimeAvgMetric *models.LogMetricTemplate
	var sucCode = getRetCodeSuccessCode(dashboardParam.RetCodeStringMap)
	var serviceGroupTable = models.ServiceGroupTable{}
	var displayServiceGroup = dashboardParam.ServiceGroup
	var serviceGroupName string
	actions = []*Action{}
	now := time.Now()
	if serviceGroupObj, getErr := getSimpleServiceGroup(dashboardParam.ServiceGroup); getErr != nil {
		err = getErr
		return
	} else {
		serviceGroupName = serviceGroupObj.DisplayName
	}
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", dashboardParam.ServiceGroup).Get(&serviceGroupTable)
	if serviceGroupTable.DisplayName != "" {
		displayServiceGroup = serviceGroupTable.DisplayName
	}
	codeList := getTargetCodeMap(dashboardParam.CodeStringMap)
	// 添加 other默认告警
	codeList = append(codeList, constOther)
	if dashboardParam.AutoCreateDashboard {
		// 1. 先创建看板
		dashboard := &models.CustomDashboardTable{
			Name:           fmt.Sprintf("%s_%s", dashboardParam.ServiceGroup, dashboardParam.MetricPrefixCode),
			CreateUser:     dashboardParam.Operator,
			UpdateUser:     dashboardParam.Operator,
			CreateAt:       now,
			UpdateAt:       now,
			RefreshWeek:    10,
			TimeRange:      -1800,
			PanelGroups:    strings.Join(codeList, ","),
			LogMetricGroup: &dashboardParam.LogMetricGroupGuid,
		}
		// 看板名称使用显示名
		if displayServiceGroup != "" {
			dashboard.Name = fmt.Sprintf("%s_%s", displayServiceGroup, dashboardParam.MetricPrefixCode)
		}
		customDashboard = dashboard.Name
		if len(dashboardParam.ServiceGroupsRoles) == 0 {
			err = fmt.Errorf("config role empty")
			return
		}
		if subDashboardActions, newDashboardId, err = getAddCustomDashboardActions(dashboard, dashboardParam.ServiceGroupsRoles[:1], dashboardParam.ServiceGroupsRoles); err != nil {
			return
		}
		if len(subDashboardActions) > 0 {
			actions = append(actions, subDashboardActions...)
		}
		// 2. 新增图表
		for index, code := range codeList {
			// 请求量+失败量 柱状图
			if reqCountMetric = getMetricByKey(metricMap, dashboardParam.MetricPrefixCode+"_"+constReqCount); reqCountMetric == nil {
				continue
			}
			if failCountMetric = getMetricByKey(metricMap, dashboardParam.MetricPrefixCode+"_"+constReqFailCount); failCountMetric == nil {
				continue
			}
			chartParam1 := &models.CustomChartDto{
				Public:             true,
				SourceDashboard:    int(newDashboardId),
				Name:               fmt.Sprintf("%s-%s/%s", code, reqCountMetric.Metric, displayServiceGroup),
				ChartTemplate:      "one",
				ChartType:          "bar",
				LineType:           "bar",
				Aggregate:          "sum",
				AggStep:            60,
				ChartSeries:        []*models.CustomChartSeriesDto{},
				DisplayConfig:      calcDisplayConfig(index * 3),
				GroupDisplayConfig: calcDisplayConfig(0),
				Group:              code,
				LogMetricGroup:     &dashboardParam.LogMetricGroupGuid,
			}
			// 请求量标签线条
			chartParam1.ChartSeries = append(chartParam1.ChartSeries, generateChartSeries(dashboardParam.ServiceGroup, dashboardParam.MonitorType, code, serviceGroupName, codeList, reqCountMetric))
			// 失败量标签线条
			chartParam1.ChartSeries = append(chartParam1.ChartSeries, generateChartSeries(dashboardParam.ServiceGroup, dashboardParam.MonitorType, code, serviceGroupName, codeList, failCountMetric))
			subChart1Actions = handleAutoCreateChart(chartParam1, newDashboardId, dashboardParam.ServiceGroupsRoles, dashboardParam.ServiceGroupsRoles[0], dashboardParam.Operator)
			if len(subChart1Actions) > 0 {
				actions = append(actions, subChart1Actions...)
			}
			// 成功率
			if sucRateMetric = getMetricByKey(metricMap, dashboardParam.MetricPrefixCode+"_"+constReqSucCount); sucRateMetric == nil {
				continue
			}
			chartParam2 := &models.CustomChartDto{
				Public:             true,
				SourceDashboard:    int(newDashboardId),
				Name:               fmt.Sprintf("%s-%s/%s", code, sucRateMetric.Metric, displayServiceGroup),
				ChartTemplate:      "one",
				ChartType:          "line",
				LineType:           "line",
				Aggregate:          "none",
				AggStep:            60,
				Unit:               "%",
				ChartSeries:        []*models.CustomChartSeriesDto{},
				DisplayConfig:      calcDisplayConfig(index*3 + 1),
				GroupDisplayConfig: calcDisplayConfig(1),
				Group:              code,
				LogMetricGroup:     &dashboardParam.LogMetricGroupGuid,
			}
			// 请求量标签线条
			chartParam2.ChartSeries = append(chartParam2.ChartSeries, generateChartSeries(dashboardParam.ServiceGroup, dashboardParam.MonitorType, code, serviceGroupName, codeList, sucRateMetric))
			subChart2Actions = handleAutoCreateChart(chartParam2, newDashboardId, dashboardParam.ServiceGroupsRoles, dashboardParam.ServiceGroupsRoles[0], dashboardParam.Operator)
			if len(subChart2Actions) > 0 {
				actions = append(actions, subChart2Actions...)
			}
			// 平均耗时
			if costTimeAvgMetric = getMetricByKey(metricMap, dashboardParam.MetricPrefixCode+"_"+constConstTimeAvg); costTimeAvgMetric == nil {
				continue
			}
			chartParam3 := &models.CustomChartDto{
				Public:             true,
				SourceDashboard:    int(newDashboardId),
				Name:               fmt.Sprintf("%s-%s/%s", code, costTimeAvgMetric.Metric, displayServiceGroup),
				ChartTemplate:      "one",
				ChartType:          "line",
				LineType:           "line",
				Aggregate:          "none",
				AggStep:            60,
				Unit:               "ms",
				ChartSeries:        []*models.CustomChartSeriesDto{},
				DisplayConfig:      calcDisplayConfig(index*3 + 2),
				GroupDisplayConfig: calcDisplayConfig(1),
				Group:              code,
				LogMetricGroup:     &dashboardParam.LogMetricGroupGuid,
			}
			chartSeries := generateChartSeries(dashboardParam.ServiceGroup, dashboardParam.MonitorType, code, serviceGroupName, codeList, costTimeAvgMetric)
			// 耗时率 只计算成功请求的耗时率
			if len(chartSeries.Tags) > 0 {
				var hasRetCode bool
				for _, tag := range chartSeries.Tags {
					if tag.TagName == constRetCode {
						tag.TagValue = []string{sucCode}
						tag.Equal = ConstEqualIn
						hasRetCode = true
					}
				}
				// other code 重新添加 retcode = success
				if !hasRetCode && code == constOther {
					chartSeries.Tags = append(chartSeries.Tags, &models.TagDto{
						TagName:  constRetCode,
						TagValue: []string{sucCode},
						Equal:    ConstEqualIn,
					})
				}
			}
			// 请求量标签线条
			chartParam3.ChartSeries = append(chartParam3.ChartSeries, chartSeries)
			subChart3Actions = handleAutoCreateChart(chartParam3, newDashboardId, dashboardParam.ServiceGroupsRoles, dashboardParam.ServiceGroupsRoles[0], dashboardParam.Operator)
			if len(subChart3Actions) > 0 {
				actions = append(actions, subChart3Actions...)
			}
		}
	}
	return
}

func autoGenerateSimpleCustomDashboard(dashboardParam models.AutoSimpleCreateDashboardParam) (actions []*Action, customDashboard string, newDashboardId int64, err error) {
	var subDashboardActions, subChartActions []*Action
	var serviceGroupTable = models.ServiceGroupTable{}
	var displayServiceGroup = dashboardParam.ServiceGroup
	actions = []*Action{}
	now := time.Now()
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", dashboardParam.ServiceGroup).Get(&serviceGroupTable)
	if serviceGroupTable.DisplayName != "" {
		displayServiceGroup = serviceGroupTable.DisplayName
	}
	if dashboardParam.AutoCreateDashboard {
		// 1. 先创建看板
		dashboard := &models.CustomDashboardTable{
			Name:           fmt.Sprintf("%s_%s", dashboardParam.ServiceGroup, dashboardParam.MetricPrefixCode),
			CreateUser:     dashboardParam.Operator,
			UpdateUser:     dashboardParam.Operator,
			CreateAt:       now,
			UpdateAt:       now,
			RefreshWeek:    10,
			TimeRange:      -1800,
			LogMetricGroup: &dashboardParam.LogMetricGroupGuid,
		}
		// 看板名称使用显示名
		if displayServiceGroup != "" {
			dashboard.Name = fmt.Sprintf("%s_%s", displayServiceGroup, dashboardParam.MetricPrefixCode)
		}
		if len(dashboardParam.ServiceGroupsRoles) == 0 {
			err = fmt.Errorf("config role empty")
			return
		}
		if subDashboardActions, newDashboardId, err = getAddCustomDashboardActions(dashboard, dashboardParam.ServiceGroupsRoles[:1], dashboardParam.ServiceGroupsRoles); err != nil {
			return
		}
		if len(subDashboardActions) > 0 {
			actions = append(actions, subDashboardActions...)
		}
		// 2. 新增图表
		for index, metric := range dashboardParam.MetricList {
			chartParam := &models.CustomChartDto{
				Public:          true,
				SourceDashboard: int(newDashboardId),
				Name:            fmt.Sprintf("%s_%s/%s", dashboardParam.MetricPrefixCode, metric.Metric, displayServiceGroup),
				ChartTemplate:   "one",
				ChartType:       "line",
				LineType:        "line",
				Aggregate:       "none",
				AggStep:         60,
				ChartSeries:     []*models.CustomChartSeriesDto{},
				DisplayConfig:   calcDisplayConfig(index),
				LogMetricGroup:  &dashboardParam.LogMetricGroupGuid,
			}
			//标签线条
			chartParam.ChartSeries = append(chartParam.ChartSeries, generateSimpleChartSeries(dashboardParam.ServiceGroup, dashboardParam.MonitorType, dashboardParam.MetricPrefixCode, metric))
			subChartActions = handleAutoCreateChart(chartParam, newDashboardId, dashboardParam.ServiceGroupsRoles, dashboardParam.ServiceGroupsRoles[0], dashboardParam.Operator)
			if len(subChartActions) > 0 {
				actions = append(actions, subChartActions...)
			}
		}
	}
	return
}

func generateMetricGuid(metric, serviceGroup string) string {
	return fmt.Sprintf("%s__%s", metric, serviceGroup)
}

func generateMetricGuidDisplayName(metricPrefixCode, metric, displayServiceGroup string) string {
	if metricPrefixCode != "" {
		metric = metricPrefixCode + "_" + metric
	}
	return fmt.Sprintf("%s__%s", metric, displayServiceGroup)
}

func getServiceGroupRoles(serviceGroup string) []string {
	var optionModels []*models.OptionModel
	var roles []string
	optionModels, _ = GetOrgRoleNew(serviceGroup)
	if len(optionModels) == 0 {
		return roles
	}
	for _, model := range optionModels {
		roles = append(roles, model.OptionName)
	}
	return roles
}

func getMetricMap(list []*models.LogMetricTemplate, metricPrefixCode, serviceGroup string) map[string]*models.LogMetricTemplate {
	var metricGuid string
	var hashMap = make(map[string]*models.LogMetricTemplate)
	for _, template := range list {
		metricGuid = template.Metric
		if metricPrefixCode != "" {
			metricGuid = metricPrefixCode + "_" + metricGuid
			template.Metric = metricPrefixCode + "_" + template.Metric
		}
		metricGuid = generateMetricGuid(metricGuid, serviceGroup)
		template.Guid = metricGuid

		hashMap[template.Guid] = template
	}
	return hashMap
}

func getMetricByKey(metricMap map[string]*models.LogMetricTemplate, subKey string) *models.LogMetricTemplate {
	if len(metricMap) == 0 {
		return nil
	}
	log.Logger.Debug("getMetricByKey", log.JsonObj("metricMap", metricMap))
	subKey = subKey + "__"
	for key, template := range metricMap {
		if strings.HasPrefix(key, subKey) {
			return template
		}
	}
	return nil
}

// getAutoAlarmMetricList 获取自动告警的指标列表
func getAutoAlarmMetricList(list []*models.LogMetricTemplate, serviceGroup, metricPrefixCode string) []*models.LogMetricThreshold {
	var metricThresholdList []*models.LogMetricThreshold
	var metric string
	if len(list) == 0 {
		return metricThresholdList
	}
	for _, logMetricTemplate := range list {
		metric = logMetricTemplate.Metric
		if metricPrefixCode != "" {
			metric = metricPrefixCode + "_" + metric
		}
		if logMetricTemplate.AutoAlarm && logMetricTemplate.RangeConfig != "" {
			temp := &models.ThresholdConfig{}
			json.Unmarshal([]byte(logMetricTemplate.RangeConfig), temp)
			// 此处添加数据校验,强制校验阈值数据,防止Prometheus解析数据失败挂掉
			if strings.TrimSpace(temp.Operator) == "" || strings.TrimSpace(temp.Threshold) == "" || strings.TrimSpace(temp.Time) == "" || strings.TrimSpace(temp.TimeUnit) == "" {
				log.Logger.Warn("getAutoAlarmMetricList strategy format invalid", log.JsonObj("strategy", temp))
				continue
			}
			metricThresholdList = append(metricThresholdList, &models.LogMetricThreshold{
				MetricId:        generateMetricGuid(metric, serviceGroup),
				Metric:          logMetricTemplate.Metric,
				DisplayName:     logMetricTemplate.DisplayName,
				ThresholdConfig: temp,
				TagConfig:       logMetricTemplate.TagConfigList,
				AutoWarn:        logMetricTemplate.AutoAlarm,
			})
		}
	}
	return metricThresholdList
}

func getAutoSimpleAlarmMetricList(list []*models.LogMetricConfigDto, serviceGroup, metricPrefixCode string) []*models.LogMetricThreshold {
	var metricThresholdList []*models.LogMetricThreshold
	var metric string
	if len(list) == 0 {
		return metricThresholdList
	}
	for _, logMetricTemplate := range list {
		metric = logMetricTemplate.Metric
		if metricPrefixCode != "" {
			metric = metricPrefixCode + "_" + metric
		}
		if logMetricTemplate.AutoAlarm && logMetricTemplate.RangeConfig != "" {
			temp := &models.ThresholdConfig{}
			byteArr, _ := json.Marshal(logMetricTemplate.RangeConfig)
			json.Unmarshal(byteArr, temp)
			metricThresholdList = append(metricThresholdList, &models.LogMetricThreshold{
				MetricId:        generateMetricGuid(metric, serviceGroup),
				Metric:          logMetricTemplate.Metric,
				DisplayName:     logMetricTemplate.DisplayName,
				ThresholdConfig: temp,
				TagConfig:       logMetricTemplate.TagConfigList,
				AutoWarn:        logMetricTemplate.AutoAlarm,
			})
		}
	}
	return metricThresholdList
}

// getTargetCodeMap 获取配置的目标code集合
func getTargetCodeMap(codeList []*models.LogMetricStringMapTable) []string {
	var list []string
	if len(codeList) == 0 {
		return []string{}
	}
	for _, table := range codeList {
		list = append(list, table.TargetValue)
	}
	return list
}

func generateChartSeries(serviceGroup, monitorType, code, serviceGroupName string, codeList []string, metric *models.LogMetricTemplate) *models.CustomChartSeriesDto {
	var serviceGroupTable = &models.ServiceGroupTable{}
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", serviceGroup).Get(serviceGroupTable)
	dto := &models.CustomChartSeriesDto{
		Endpoint:     serviceGroup,
		ServiceGroup: serviceGroup,
		EndpointName: serviceGroup,
		MonitorType:  monitorType,
		ColorGroup:   metric.ColorGroup,
		MetricType:   "business",
		MetricGuid:   metric.Guid,
		Metric:       metric.Metric,
	}
	if serviceGroupTable != nil {
		dto.EndpointName = serviceGroupTable.DisplayName
		dto.EndpointType = serviceGroupTable.ServiceType
	}
	if code == "other" && len(codeList) > 0 {
		dto.Tags = []*models.TagDto{
			{
				TagName:  constCode,
				Equal:    ConstEqualNotIn,
				TagValue: codeList[:len(codeList)-1],
			},
		}
	} else {
		dto.Tags = []*models.TagDto{
			{
				TagName:  constCode,
				Equal:    ConstEqualIn,
				TagValue: []string{code},
			},
			{
				TagName: constRetCode,
			},
		}
		dto.ColorConfig = []*models.ColorConfigDto{
			{
				SeriesName: fmt.Sprintf("%s:%s{code=%s}", metric.Metric, serviceGroupName, code),
				Color:      metric.ColorGroup,
			},
		}
	}
	return dto
}

func generateSimpleChartSeries(serviceGroup, monitorType, metricPrefixCode string, metric *models.LogMetricConfigDto) *models.CustomChartSeriesDto {
	var serviceGroupTable = &models.ServiceGroupTable{}
	x.SQL("SELECT guid,display_name,service_type FROM service_group where guid=?", serviceGroup).Get(serviceGroupTable)
	metricGuid := metric.Metric
	if metricPrefixCode != "" {
		metricGuid = metricPrefixCode + "_" + metricGuid
		metric.Metric = metricPrefixCode + "_" + metric.Metric
	}
	dto := &models.CustomChartSeriesDto{
		Endpoint:     serviceGroup,
		ServiceGroup: serviceGroup,
		EndpointName: serviceGroup,
		MonitorType:  monitorType,
		ColorGroup:   metric.ColorGroup,
		MetricType:   "business",
		MetricGuid:   generateMetricGuid(metricGuid, serviceGroup),
		Metric:       metric.Metric,
		Tags:         make([]*models.TagDto, 0),
	}
	if serviceGroupTable != nil {
		dto.EndpointName = serviceGroupTable.DisplayName
		dto.EndpointType = serviceGroupTable.ServiceType
	}
	if len(metric.TagConfigList) > 0 {
		dto.Tags = append(dto.Tags, &models.TagDto{TagName: constCode, Equal: ConstEqualIn})
	}
	return dto
}

func calcDisplayConfig(index int) models.DisplayConfig { // index是item在数组中的索引，item是数组中的其中一个
	item := models.DisplayConfig{}
	item.W = 4
	item.H = 7
	item.X = float64((index % 3) * 4)
	item.Y = math.Floor(float64(index/3)) * 7
	return item
}

func getAlarmMetricUnit(metric string) string {
	if metric == constReqSuccessRate || metric == constReqFailRate {
		return "%"
	}
	if metric == constConstTimeAvg || metric == constConstTimeMax {
		return "ms"
	}
	return ""
}
