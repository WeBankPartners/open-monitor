package db

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

const (
	constOtherCode    = "{code}"
	constOther        = "other"
	constCode         = "code"
	constRetCode      = "retcode"
	constEqualIn      = "in"
	constEqualNotIn   = "notin"
	constConstTimeAvg = "req_costtime_avg"
	constReqCount     = "req_count"
	constReqFailCount = "req_fail_count"
	constReqSucCount  = "req_suc_rate"
	constSuccess      = "success"
)

func autoGenerateAlarmStrategy(param *models.LogMetricGroupWithTemplate, metricList []*models.LogMetricTemplate, serviceGroupsRoles []string, serviceGroup, endpointGroup, operator string) (actions []*Action, result []string, err error) {
	var subActions []*Action
	result = []string{}
	actions = []*Action{}
	// 自动创建告警
	if param.AutoCreateWarn {
		codeList := getTargetCodeMap(param.CodeStringMap)
		autoAlarmMetricList := getAutoAlarmMetricList(metricList, serviceGroup, param.MetricPrefixCode)
		// 添加 other默认告警
		codeList = append(codeList, constOtherCode)
		for _, code := range codeList {
			for _, alarmMetric := range autoAlarmMetricList {
				// 添加告警配置基础信息
				alarmStrategyParam := &models.GroupStrategyObj{NotifyList: make([]*models.NotifyObj, 0), Conditions: make([]*models.StrategyConditionObj, 0)}
				metricTags := make([]*models.MetricTag, 0)
				alarmStrategyParam.Name = fmt.Sprintf("%s_%s%s %d %s", code, alarmMetric.Metric, translateSymbol(alarmMetric.Operator), alarmMetric.Threshold, alarmMetric.TimeUnit)
				alarmStrategyParam.Priority = "medium"
				alarmStrategyParam.NotifyEnable = 1
				alarmStrategyParam.ActiveWindow = "00:00-23:59"
				if strings.TrimSpace(endpointGroup) != "" {
					alarmStrategyParam.EndpointGroup = endpointGroup
				}
				alarmStrategyParam.LogMetricGroup = &param.LogMetricGroupGuid
				alarmStrategyParam.Metric = alarmMetric.MetricId
				alarmStrategyParam.MetricName = alarmMetric.Metric
				alarmStrategyParam.Content = fmt.Sprintf("%s continuing for more than %d %s", alarmStrategyParam.Name, alarmMetric.Time, alarmMetric.TimeUnit)
				// 添加编排与通知
				alarmStrategyParam.NotifyList = append(alarmStrategyParam.NotifyList, &models.NotifyObj{AlarmAction: "firing", NotifyRoles: serviceGroupsRoles})
				alarmStrategyParam.NotifyList = append(alarmStrategyParam.NotifyList, &models.NotifyObj{AlarmAction: "ok", NotifyRoles: serviceGroupsRoles})
				result = append(result, alarmStrategyParam.Name)
				// 添加指标阈值
				for _, tag := range alarmMetric.TagConfig {
					// 标签为code,需要配置 equal和TagValue值
					if tag == code {
						// code为 other 配置not in,其他配置 in
						if code == constOtherCode {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  constCode,
								Equal:    constEqualNotIn,
								TagValue: codeList[:len(codeList)-1],
							})
						} else {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  constCode,
								Equal:    constEqualIn,
								TagValue: []string{code},
							})
						}
					} else {
						// 平均耗时,只会统计成功请求的平均耗时
						if code == constRetCode && strings.HasSuffix(alarmMetric.Metric, constConstTimeAvg) {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  constRetCode,
								Equal:    constEqualIn,
								TagValue: []string{getRetCodeSuccessCode(param.CodeStringMap)},
							})
						} else {
							metricTags = append(metricTags, &models.MetricTag{
								TagName: tag,
							})
						}
					}
				}

				alarmStrategyParam.Conditions = append(alarmStrategyParam.Conditions, &models.StrategyConditionObj{
					Metric:     alarmMetric.MetricId,
					MetricName: alarmMetric.Metric,
					Condition:  fmt.Sprintf("%s%d", alarmMetric.Operator, alarmMetric.Threshold),
					Last:       fmt.Sprintf("%d%s", alarmMetric.Time, alarmMetric.TimeUnit),
					Tags:       metricTags,
				})
				if subActions, err = getCreateAlarmStrategyActions(alarmStrategyParam, time.Now().Format(models.DatetimeFormat), operator); err != nil {
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

func getRetCodeSuccessCode(stringMap []*models.LogMetricStringMapTable) string {
	if len(stringMap) == 0 {
		return ""
	}
	for _, table := range stringMap {
		if table.ValueType == constSuccess && table.LogParamName == constRetCode {
			return table.TargetValue
		}
	}
	return ""
}

func autoGenerateCustomDashboard(param *models.LogMetricGroupWithTemplate, metricList []*models.LogMetricTemplate, serviceGroupsRoles []string, serviceGroup, operator string) (actions []*Action, customDashboard string, err error) {
	var subDashboardActions, subChart1Actions, subChart2Actions, subChart3Actions []*Action
	var newDashboardId int64
	actions = []*Action{}
	now := time.Now()
	var metricMap = getMetricMap(metricList, param.MetricPrefixCode, serviceGroup)
	var reqCountMetric, failCountMetric, sucRateMetric, costTimeAvgMetric *models.LogMetricTemplate
	if param.AutoCreateDashboard {
		// 1. 先创建看板
		dashboard := &models.CustomDashboardTable{
			Name:           fmt.Sprintf("%s_%s", serviceGroup, now.Format(models.DatetimeDigitFormat)),
			CreateUser:     operator,
			UpdateUser:     operator,
			CreateAt:       now,
			UpdateAt:       now,
			RefreshWeek:    60,
			TimeRange:      -1800,
			LogMetricGroup: &param.LogMetricGroupGuid,
		}
		customDashboard = dashboard.Name
		if len(serviceGroupsRoles) == 0 {
			err = fmt.Errorf("config role empty")
			return
		}
		if subDashboardActions, newDashboardId, err = getAddCustomDashboardActions(dashboard, serviceGroupsRoles[:1], serviceGroupsRoles); err != nil {
			return
		}
		if len(subDashboardActions) > 0 {
			actions = append(actions, subDashboardActions...)
		}
		// 2. 新增图表
		codeList := getTargetCodeMap(param.CodeStringMap)
		// 添加 other默认告警
		codeList = append(codeList, constOther)
		for index, code := range codeList {
			// 请求量+失败量 柱状图
			if reqCountMetric = getMetricByKey(metricMap, param.MetricPrefixCode+"_"+constReqCount); reqCountMetric == nil {
				continue
			}
			if failCountMetric = getMetricByKey(metricMap, param.MetricPrefixCode+"_"+constReqFailCount); failCountMetric == nil {
				continue
			}
			chartParam1 := &models.CustomChartDto{
				Public:             true,
				SourceDashboard:    int(newDashboardId),
				Name:               fmt.Sprintf("%s-%s/%s", code, reqCountMetric.Metric, serviceGroup),
				ChartTemplate:      "one",
				ChartType:          "bar",
				LineType:           "bar",
				Aggregate:          "none",
				AggStep:            60,
				ChartSeries:        []*models.CustomChartSeriesDto{},
				DisplayConfig:      calcDisplayConfig(index * 3),
				GroupDisplayConfig: calcDisplayConfig(0),
				Group:              code,
				LogMetricGroup:     &param.LogMetricGroupGuid,
			}
			// 请求量标签线条
			chartParam1.ChartSeries = append(chartParam1.ChartSeries, generateChartSeries(serviceGroup, param.MonitorType, code, codeList, reqCountMetric))
			// 失败量标签线条
			chartParam1.ChartSeries = append(chartParam1.ChartSeries, generateChartSeries(serviceGroup, param.MonitorType, code, codeList, failCountMetric))
			subChart1Actions = handleAutoCreateChart(chartParam1, newDashboardId, serviceGroupsRoles, serviceGroupsRoles[0], operator)
			if len(subChart1Actions) > 0 {
				actions = append(actions, subChart1Actions...)
			}
			// 成功率
			if sucRateMetric = getMetricByKey(metricMap, param.MetricPrefixCode+"_"+constReqSucCount); sucRateMetric == nil {
				continue
			}
			chartParam2 := &models.CustomChartDto{
				Public:             true,
				SourceDashboard:    int(newDashboardId),
				Name:               fmt.Sprintf("%s-%s/%s", code, sucRateMetric.Metric, serviceGroup),
				ChartTemplate:      "one",
				ChartType:          "line",
				LineType:           "line",
				Aggregate:          "none",
				AggStep:            60,
				ChartSeries:        []*models.CustomChartSeriesDto{},
				DisplayConfig:      calcDisplayConfig(index*3 + 1),
				GroupDisplayConfig: calcDisplayConfig(1),
				Group:              code,
				LogMetricGroup:     &param.LogMetricGroupGuid,
			}
			// 请求量标签线条
			sucCode := getRetCodeSuccessCode(param.CodeStringMap)
			chartSeries := generateChartSeries(serviceGroup, param.MonitorType, code, codeList, sucRateMetric)
			// 耗时率 只计算成功请求的耗时率
			if len(chartSeries.Tags) > 0 {
				for _, tag := range chartSeries.Tags {
					if tag.TagName == constRetCode {
						tag.TagValue = []string{sucCode}
						tag.Equal = constEqualIn
					}
				}
			}
			chartParam2.ChartSeries = append(chartParam2.ChartSeries)
			subChart2Actions = handleAutoCreateChart(chartParam2, newDashboardId, serviceGroupsRoles, serviceGroupsRoles[0], operator)
			if len(subChart2Actions) > 0 {
				actions = append(actions, subChart2Actions...)
			}
			// 耗时
			if costTimeAvgMetric = getMetricByKey(metricMap, param.MetricPrefixCode+"_"+constConstTimeAvg); costTimeAvgMetric == nil {
				continue
			}
			chartParam3 := &models.CustomChartDto{
				Public:             true,
				SourceDashboard:    int(newDashboardId),
				Name:               fmt.Sprintf("%s-%s/%s", code, costTimeAvgMetric.Metric, serviceGroup),
				ChartTemplate:      "one",
				ChartType:          "line",
				LineType:           "line",
				Aggregate:          "none",
				AggStep:            60,
				ChartSeries:        []*models.CustomChartSeriesDto{},
				DisplayConfig:      calcDisplayConfig(index*3 + 2),
				GroupDisplayConfig: calcDisplayConfig(1),
				Group:              code,
				LogMetricGroup:     &param.LogMetricGroupGuid,
			}
			// 请求量标签线条
			chartParam3.ChartSeries = append(chartParam3.ChartSeries, generateChartSeries(serviceGroup, param.MonitorType, code, codeList, costTimeAvgMetric))
			subChart3Actions = handleAutoCreateChart(chartParam3, newDashboardId, serviceGroupsRoles, serviceGroupsRoles[0], operator)
			if len(subChart3Actions) > 0 {
				actions = append(actions, subChart3Actions...)
			}
		}
	}
	return
}

func generateMetricGuid(metric, serviceGroup string) string {
	return fmt.Sprintf("%s__%s", metric, serviceGroup)
}

func getServiceGroupRoles(serviceGroup string) []string {
	var optionModels []*models.OptionModel
	var roles []string
	optionModels, _ = GetOrgRole(serviceGroup)
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
			metricThresholdList = append(metricThresholdList, &models.LogMetricThreshold{
				MetricId:        generateMetricGuid(metric, serviceGroup),
				Metric:          logMetricTemplate.Metric,
				DisplayName:     logMetricTemplate.DisplayName,
				ThresholdConfig: temp,
				TagConfig:       logMetricTemplate.TagConfigList,
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

func generateChartSeries(serviceGroup, monitorType, code string, codeList []string, metric *models.LogMetricTemplate) *models.CustomChartSeriesDto {
	var endpointType string
	x.SQL("SELECT obj_type FROM panel_recursive where guid=?", serviceGroup).Get(&endpointType)
	dto := &models.CustomChartSeriesDto{
		Endpoint:     serviceGroup,
		ServiceGroup: serviceGroup,
		EndpointName: serviceGroup,
		MonitorType:  monitorType,
		ColorGroup:   metric.ColorGroup,
		EndpointType: endpointType,
		MetricType:   "business",
		MetricGuid:   metric.Guid,
		Metric:       metric.Metric,
	}
	if code == "other" && len(codeList) > 0 {
		dto.Tags = []*models.TagDto{
			{
				TagName:  constCode,
				Equal:    constEqualNotIn,
				TagValue: codeList[:len(codeList)-1],
			},
		}
	} else {
		dto.Tags = []*models.TagDto{
			{
				TagName:  constCode,
				Equal:    constEqualIn,
				TagValue: []string{code},
			},
			{
				TagName: constRetCode,
			},
		}
		dto.ColorConfig = []*models.ColorConfigDto{
			{
				SeriesName: fmt.Sprintf("%s:%s{code=%s}", metric.Metric, serviceGroup, code),
				Color:      metric.ColorGroup,
			},
		}
	}
	return dto
}

// translateSymbol 字符翻译
func translateSymbol(operator string) string {
	switch operator {
	case ">":
		return "greater than"
	case ">=":
		return "greater than or equal"
	case "<":
		return "less than"
	case "<=":
		return "less than or equal"
	}
	return ""
}

func calcDisplayConfig(index int) models.DisplayConfig { // index是item在数组中的索引，item是数组中的其中一个
	item := models.DisplayConfig{}
	item.W = 4
	item.H = 7
	item.X = float64((index % 3) * 4)
	item.Y = math.Floor(float64(index/3)) * 7
	return item
}
