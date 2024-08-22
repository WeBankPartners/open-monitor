package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func autoGenerateAlarmStrategy(param *models.LogMetricGroupWithTemplate, metricList []*models.LogMetricTemplate, serviceGroup, operator string) (subCreateAlarmStrategyActions []*Action, err error) {
	subCreateAlarmStrategyActions = []*Action{}
	// 自动创建告警
	if param.AutoCreateWarn {
		var endpointGroup string
		var serviceGroupsRoles []string
		codeList := getTargetCodeMap(param.CodeStringMap)
		autoAlarmMetricList := getAutoAlarmMetricList(metricList, serviceGroup)
		if param.LogMetricMonitorGuid != "" {
			var logMetricMonitor = &models.LogMetricMonitorTable{}
			var endpointGroupIds []string
			if _, err = x.SQL("select service_group,monitor_type from log_metric_monitor where guid=?", param.LogMetricMonitorGuid).Get(logMetricMonitor); err != nil {
				return
			}
			if logMetricMonitor != nil {
				serviceGroupsRoles = getServiceGroupRoles(logMetricMonitor.ServiceGroup)
				if err = x.SQL("select guid from endpoint_group where service_group=? and monitor_type=?", logMetricMonitor.ServiceGroup, logMetricMonitor.MonitorType).Find(&endpointGroupIds); err != nil {
					return
				}
				if len(endpointGroupIds) > 0 {
					endpointGroup = endpointGroupIds[0]
				}
			}
		}
		// 添加 other默认告警
		codeList = append(codeList, "{code}")
		for _, code := range codeList {
			for _, alarmMetric := range autoAlarmMetricList {
				// 添加告警配置基础信息
				alarmStrategyParam := &models.GroupStrategyObj{NotifyList: make([]*models.NotifyObj, 0), Conditions: make([]*models.StrategyConditionObj, 0)}
				metricTags := make([]*models.MetricTag, 0)
				alarmStrategyParam.Name = fmt.Sprintf("%s%s%s %d %s", code, alarmMetric.DisplayName, translateSymbol(alarmMetric.Operator), alarmMetric.Threshold, alarmMetric.TimeUnit)
				alarmStrategyParam.Priority = "medium"
				alarmStrategyParam.NotifyEnable = 1
				alarmStrategyParam.ActiveWindow = "00:00-23:59"
				alarmStrategyParam.EndpointGroup = endpointGroup
				alarmStrategyParam.LogMetricGroup = param.LogMetricGroupGuid
				alarmStrategyParam.Content = fmt.Sprintf("%s continuing for more than %d %s", alarmStrategyParam.Name, alarmMetric.Time, alarmMetric.TimeUnit)
				// 添加编排与通知
				alarmStrategyParam.NotifyList = append(alarmStrategyParam.NotifyList, &models.NotifyObj{AlarmAction: "firing", NotifyRoles: serviceGroupsRoles})
				alarmStrategyParam.NotifyList = append(alarmStrategyParam.NotifyList, &models.NotifyObj{AlarmAction: "ok", NotifyRoles: serviceGroupsRoles})
				// 添加指标阈值
				for _, tag := range alarmMetric.TagConfig {
					// 标签为code,需要配置 equal和TagValue值
					if tag == "code" {
						// code为 other 配置not in,其他配置 in
						if code == "{code}" {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  "code",
								Equal:    "notin",
								TagValue: codeList[:len(codeList)-1],
							})
						} else {
							metricTags = append(metricTags, &models.MetricTag{
								TagName:  "code",
								Equal:    "in",
								TagValue: []string{code},
							})
						}
					} else {
						metricTags = append(metricTags, &models.MetricTag{
							TagName: tag,
						})
					}
				}

				alarmStrategyParam.Conditions = append(alarmStrategyParam.Conditions, &models.StrategyConditionObj{
					Metric:     alarmMetric.MetricId,
					MetricName: alarmMetric.Metric,
					Condition:  fmt.Sprintf("%s%d", alarmMetric.Operator, alarmMetric.Threshold),
					Last:       fmt.Sprintf("%d%s", alarmMetric.Time, alarmMetric.TimeUnit),
					Tags:       metricTags,
				})
				if subCreateAlarmStrategyActions, err = getCreateAlarmStrategyActions(alarmStrategyParam, time.Now().Format(models.DatetimeFormat), operator); err != nil {
					return
				}
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

// getAutoAlarmMetricList 获取自动告警的指标列表
func getAutoAlarmMetricList(list []*models.LogMetricTemplate, serviceGroup string) []*models.LogMetricThreshold {
	var metricThresholdList []*models.LogMetricThreshold
	if len(list) == 0 {
		return metricThresholdList
	}
	for _, logMetricTemplate := range list {
		if logMetricTemplate.AutoAlarm && logMetricTemplate.RangeConfig != "" {
			temp := &models.ThresholdConfig{}
			json.Unmarshal([]byte(logMetricTemplate.RangeConfig), temp)
			metricThresholdList = append(metricThresholdList, &models.LogMetricThreshold{
				MetricId:        generateMetricGuid(logMetricTemplate.Metric, serviceGroup),
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
