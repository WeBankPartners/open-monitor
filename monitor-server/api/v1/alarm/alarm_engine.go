package alarm

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"strconv"
	"strings"
	"time"
)

func StartAlarmEngineCron() {
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go doMonitorEngineRuleJob()
	}
}

func doMonitorEngineRuleJob() {
	log.Logger.Debug("doAlarmEngineRuleJob")
	var err error
	var alarmStrategyMetricRows []*models.AlarmStrategyMetric
	var existAlarmRows []*models.AlarmTable
	defer func() {
		if err != nil {
			log.Logger.Warn("doAlarmEngineRuleJob fail", log.Error(err))
		} else {
			log.Logger.Debug("doAlarmEngineRuleJob done")
		}
	}()
	alarmStrategyMetricRows, err = db.GetMonitorEngineStrategy()
	if err != nil {
		return
	}
	existAlarmRows, err = db.GetMonitorEngineAlarmList()
	if err != nil {
		return
	}
	var alarmList []*models.AlarmHandleObj
	for _, row := range alarmStrategyMetricRows {
		condition, threshold, illegal := analyzeCondition(row.Condition)
		if illegal {
			log.Logger.Info("doAlarmEngineRuleJob condition illegal", log.String("alarmStrategyMetric", row.Guid), log.String("condition", row.Condition))
			continue
		}
		alarmObjList, tmpErr := buildMonitorEngineAlarm(row, condition, threshold, existAlarmRows)
		if tmpErr != nil {
			log.Logger.Warn("doAlarmEngineRuleJob buildMonitorEngineAlarm fail", log.Error(tmpErr))
		} else if len(alarmObjList) > 0 {
			alarmList = append(alarmList, alarmObjList...)
		}
	}
	if len(alarmList) == 0 {
		return
	}
	alarmList = db.UpdateAlarms(alarmList)
	for _, v := range alarmList {
		log.Logger.Debug("update alarm result", log.JsonObj("alarm", v))
		if v.AlarmConditionGuid != "" {
			continue
		}
		if v.NotifyEnable == 0 {
			continue
		}
		go db.NotifyStrategyAlarm(v)
	}
}

func analyzeCondition(conditionConfig string) (condition string, threshold float64, illegal bool) {
	if len(conditionConfig) < 2 {
		illegal = true
		return
	}
	if strings.Contains(conditionConfig, "=") {
		condition = conditionConfig[:2]
		conditionConfig = conditionConfig[2:]
	} else {
		condition = conditionConfig[:1]
		conditionConfig = conditionConfig[1:]
	}
	floatValue, err := strconv.ParseFloat(conditionConfig, 64)
	if err != nil {
		illegal = true
		return
	}
	threshold = floatValue
	return
}

func analyzeLast(lastConfig string) (timeSec int64) {
	lastConfigLen := len(lastConfig)
	if lastConfigLen < 2 {
		return
	}
	unit := lastConfig[lastConfigLen-1:]
	value := lastConfig[:lastConfigLen-1]
	valueInt, _ := strconv.ParseInt(value, 10, 64)
	if valueInt == 0 {
		return
	}
	if unit == "s" {
		timeSec = valueInt
	} else if unit == "m" {
		timeSec = valueInt * 60
	} else if unit == "h" {
		timeSec = valueInt * 3600
	}
	return
}

func compareFloatValue(inputValue, threshold float64, condition string) (match bool) {
	switch condition {
	case ">":
		if inputValue > threshold {
			match = true
		}
	case "<":
		if inputValue < threshold {
			match = true
		}
	case ">=":
		if inputValue >= threshold {
			match = true
		}
	case "<=":
		if inputValue <= threshold {
			match = true
		}
	case "==":
		if inputValue == threshold {
			match = true
		}
	case "!=":
		if inputValue != threshold {
			match = true
		}
	}
	return
}

func buildMonitorEngineAlarm(alarmStrategyMetric *models.AlarmStrategyMetric, condition string, threshold float64, existAlarmRows []*models.AlarmTable) (alarmObjList []*models.AlarmHandleObj, err error) {
	lastSec := analyzeLast(alarmStrategyMetric.Last)
	if lastSec == 0 {
		err = fmt.Errorf("lastConfig:%s illegal", alarmStrategyMetric.Last)
		return
	}
	endTime := time.Now().Unix()
	startTime := endTime - lastSec
	queryData, queryErr := datasource.QueryPrometheusRange(alarmStrategyMetric.MonitorEngineExpr, startTime, endTime, 10)
	if queryErr != nil {
		err = fmt.Errorf("query prometheus data fail,%s ", queryErr.Error())
		return
	}
	for _, queryObj := range queryData.Result {
		alarmObj := &models.AlarmHandleObj{}
		delete(queryObj.Metric, "__name__")
		tmpTags, getTagsErr := getNewAlarmTags(&models.AMRespAlert{Labels: queryObj.Metric})
		if getTagsErr != nil {
			log.Logger.Error("buildMonitorEngineAlarm get tags fail", log.JsonObj("labels", queryObj.Metric), log.Error(getTagsErr))
			continue
		}
		tmpExistAlarm := matchMonitorEngineExistAlarm(queryObj.Metric, existAlarmRows, tmpTags, alarmStrategyMetric)
		firingMatch := false
		firingNonMatch := false
		var startValue, endValue float64
		for _, v := range queryObj.Values {
			tmpValue, tmpParseErr := strconv.ParseFloat(v[1].(string), 64)
			if tmpParseErr == nil {
				if !compareFloatValue(tmpValue, threshold, condition) {
					firingNonMatch = true
					endValue = tmpValue
					break
				} else {
					firingMatch = true
					startValue = tmpValue
				}
			}
		}
		log.Logger.Debug("buildMonitorEngineAlarm condition", log.JsonObj("queryObj", queryObj), log.Bool("firingMatch", firingMatch), log.Bool("firingNonMatch", firingNonMatch), log.Float64("startValue", startValue), log.Float64("endValue", endValue))
		// 如果有一个不符合的值，则算是不满足
		if firingNonMatch {
			if tmpExistAlarm.Id > 0 {
				// 有正在发生的告警，需要恢复
				alarmObj.Id = tmpExistAlarm.Id
				alarmObj.AlarmStrategy = tmpExistAlarm.AlarmStrategy
				alarmObj.StrategyId = tmpExistAlarm.StrategyId
				alarmObj.Status = "ok"
				alarmObj.EndValue = endValue
				alarmObj.End = time.Now()
				//alarmObj.AlarmConditionGuid = alarmConditionGuid
			}
		} else {
			// 如果要满足，至少要有一个值满足
			if firingMatch {
				if tmpExistAlarm.Id <= 0 {
					// 没有正在发生的告警，需要新增
					strategyObj, _, tmpGetStrategyErr := db.GetAlarmStrategy(alarmStrategyMetric.AlarmStrategy, alarmStrategyMetric.CrcHash)
					if tmpGetStrategyErr != nil {
						log.Logger.Error("buildMonitorEngineAlarm get strategy object fail", log.String("alarmStrategy", alarmStrategyMetric.AlarmStrategy), log.Error(tmpGetStrategyErr))
						continue
					}
					// 如果不在告警窗口内跳过
					if !db.InActiveWindowList(strategyObj.ActiveWindow) {
						log.Logger.Warn("buildMonitorEngineAlarm alarm not in active window", log.String("alarmStrategy", alarmStrategyMetric.AlarmStrategy), log.String("activeWindow", strategyObj.ActiveWindow))
						continue
					}
					queryObj.Metric["strategy_guid"] = strategyObj.Guid
					endpointObj, tmpGetEndpointErr := getNewAlarmEndpoint(&models.AMRespAlert{Labels: queryObj.Metric}, &strategyObj)
					if tmpGetEndpointErr != nil {
						log.Logger.Error("buildMonitorEngineAlarm get endpoint fail", log.JsonObj("labels", queryObj.Metric), log.Error(tmpGetEndpointErr))
						continue
					}
					alarmObj.Endpoint = endpointObj.Guid
					alarmObj.Tags = tmpTags
					alarmObj.AlarmConditionCrcHash = alarmStrategyMetric.CrcHash
					alarmObj.Status = "firing"
					alarmObj.StartValue = startValue
					alarmObj.Start = time.Now()
					alarmObj.AlarmStrategy = strategyObj.Guid
					alarmObj.SMetric = strategyObj.MetricName
					alarmObj.SExpr = strategyObj.MetricExpr
					alarmObj.SCond = strategyObj.Condition
					alarmObj.SLast = strategyObj.Last
					alarmObj.SPriority = strategyObj.Priority
					// 自动化生成 告警阈值,包含 {code} 需要替换成真实的报警code
					if len(queryObj.Metric) > 0 {
						alarmObj.AlarmName = strings.ReplaceAll(strategyObj.Name, "{code}", queryObj.Metric["code"])
						alarmObj.Content = strings.ReplaceAll(strategyObj.Content, "{code}", queryObj.Metric["code"])
					} else {
						alarmObj.AlarmName = strategyObj.Name
						alarmObj.Content = strategyObj.Content
					}
					alarmObj.NotifyEnable = strategyObj.NotifyEnable
					alarmObj.NotifyDelay = strategyObj.NotifyDelaySecond
				}
			} else {
				// 又没有满足的值，又没不满足的值，可能是没值，不理
			}
		}
		if alarmObj.Status != "" {
			alarmObjList = append(alarmObjList, alarmObj)
		}
	}
	return
}

func matchMonitorEngineExistAlarm(metaMap map[string]string, existAlarmRows []*models.AlarmTable, tags string, alarmStrategyMetric *models.AlarmStrategyMetric) (existAlarm *models.AlarmTable) {
	existAlarm = &models.AlarmTable{}
	for _, v := range existAlarmRows {
		if v.AlarmStrategy == alarmStrategyMetric.AlarmStrategy && v.Tags == tags {
			existAlarm = v
			break
		}
	}
	return
}
