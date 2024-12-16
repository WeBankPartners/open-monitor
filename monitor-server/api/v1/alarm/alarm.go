package alarm

import (
	"encoding/json"
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func AcceptAlert(c *gin.Context) {
	var param m.AlterManagerRespObj
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if len(param.Alerts) == 0 {
		mid.ReturnSuccess(c)
		return
	}
	log.Logger.Debug("accept", log.JsonObj("body", param))
	nowTime := time.Now()
	var alarms []*m.AlarmHandleObj
	for _, v := range param.Alerts {
		tmpAlarm, tmpErr := buildNewAlarm(&v, nowTime)
		if tmpErr != nil {
			log.Logger.Warn("Accept alert handle fail", log.Error(tmpErr))
			continue
		}
		log.Logger.Debug("build alarm result", log.JsonObj("alarm", tmpAlarm))
		alarms = append(alarms, &tmpAlarm)
	}
	alarms = db.UpdateAlarms(alarms)
	var treeventSendObj m.EventTreeventNotifyDto
	for _, v := range alarms {
		log.Logger.Debug("update alarm result", log.JsonObj("alarm", v))
		if v.AlarmConditionGuid != "" {
			continue
		}
		treeventSendObj.Data = append(treeventSendObj.Data, &m.EventTreeventNodeParam{EventId: fmt.Sprintf("%d", v.Id), Status: v.Status, Endpoint: v.Endpoint, StartUnix: v.Start.Unix(), Message: fmt.Sprintf("%s \n %s \n %.3f %s", v.Endpoint, v.SMetric, v.StartValue, v.SCond)})
		if v.NotifyEnable == 0 {
			continue
		}
		//go db.NotifyAlarm(v)
		go db.NotifyStrategyAlarm(v)
	}
	if m.NotifyTreeventEnable {
		go db.NotifyTreevent(treeventSendObj)
	}
	mid.ReturnSuccess(c)
}

func buildNewAlarm(param *m.AMRespAlert, nowTime time.Time) (alarm m.AlarmHandleObj, err error) {
	alarm = m.AlarmHandleObj{}
	alarm.Tags, err = getNewAlarmTags(param)
	if err != nil {
		return
	}
	summaryList := strings.Split(param.Annotations["summary"], "__")
	if len(summaryList) <= 3 {
		return alarm, fmt.Errorf("summary:%s illegal ", param.Annotations["summary"])
	}
	var strategyGuid, conditionCrc, alarmConditionGuid string
	var strategyObj m.AlarmStrategyMetricObj
	var multipleConditionFlag bool
	var strategyConditions []*m.AlarmStrategyMetricWithExpr
	strategyGuid = param.Labels["strategy_guid"]
	conditionCrc = param.Labels["condition_crc"]
	log.Logger.Debug("start build alarm data", log.String("strategyGuid", strategyGuid), log.String("conditionCrc", conditionCrc))
	alarm.AlarmConditionCrcHash = conditionCrc
	if strategyGuid != "" {
		strategyObj, strategyConditions, err = db.GetAlarmStrategy(strategyGuid, conditionCrc)
		if err != nil {
			return alarm, fmt.Errorf("Try to get alarm strategy with strategy_guid:%s fail,%s ", strategyGuid, err.Error())
		}
		log.Logger.Debug("getNewAlarmWithStrategyGuid", log.JsonObj("strategyObj", strategyObj), log.JsonObj("conditions", strategyConditions))
		if len(strategyConditions) > 1 {
			multipleConditionFlag = true
			alarm.MultipleConditionFlag = true
		}
	}
	var endpointObj m.EndpointNewTable
	endpointObj, err = getNewAlarmEndpoint(param, &strategyObj)
	if err != nil {
		return
	}
	alarm.Endpoint = endpointObj.Guid
	if len(param.Labels) > 0 {
		alarm.AlarmName = strings.ReplaceAll(strategyObj.Name, "{code}", param.Labels["code"])
	} else {
		alarm.AlarmName = strategyObj.Name
	}
	var alertValue float64
	alertValue, _ = strconv.ParseFloat(summaryList[len(summaryList)-1], 64)
	alertValue, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", alertValue), 64)
	if param.Labels["strategy_id"] == "" && strategyGuid == "" {
		return alarm, fmt.Errorf("labels strategy_id and strategy_guid is empty ")
	}
	existAlarm := m.AlarmTable{}
	log.Logger.Debug("accept strategy_guid", log.String("strategy_guid", strategyGuid))
	if strategyGuid != "" {
		existAlarm, alarmConditionGuid, err = getNewAlarmWithStrategyGuid(&alarm, param, &endpointObj, &strategyObj, multipleConditionFlag)
	} else if param.Labels["strategy_id"] == "up" {
		if endpointObj.MonitorType != "host" {
			return alarm, fmt.Errorf("Up alarm break,endpoint:%s export type illegal ", endpointObj.Guid)
		}
		existAlarm, err = getNewAlarmWithUpCase(&alarm, param)
	} else {
		existAlarm, err = getNewAlarmWithStrategyId(&alarm, param, &endpointObj)
	}
	if err != nil {
		return
	}
	log.Logger.Debug("exist alarm", log.JsonObj("existAlarm", existAlarm), log.String("alarmConditionGuid", alarmConditionGuid), log.JsonObj("alarm", alarm))
	alarm.Status = param.Status
	operation := "add"
	if alarm.Status == "firing" {
		if existAlarm.Status == "firing" {
			operation = "same"
		} else {
			operation = "add"
		}
	} else if alarm.Status == "resolved" {
		if existAlarm.Status == "firing" {
			operation = "resolve"
		} else {
			operation = "same"
		}
	} else {
		return alarm, fmt.Errorf("Accept alert status:%s illegal! ", alarm.Status)
	}
	//if existAlarm.Status != "" {
	//	if existAlarm.Status == "firing" {
	//		if alarm.Status == "firing" {
	//			operation = "same"
	//		} else {
	//			operation = "resolve"
	//		}
	//	} else if existAlarm.Status == "ok" {
	//		if alarm.Status == "resolved" {
	//			operation = "same"
	//		}
	//	} else if existAlarm.Status == "closed" {
	//		if alarm.Status == "resolved" {
	//			operation = "same"
	//		}
	//	}
	//}
	if operation == "same" {
		return alarm, fmt.Errorf("Accept alert msg ,firing repeat,do nothing! ")
	}
	if operation == "add" && param.Status == "resolved" {
		return alarm, fmt.Errorf("Accept alert msg ,cat not add resolved,do nothing! ")
	}
	if operation == "resolve" {
		alarm.Id = existAlarm.Id
		alarm.AlarmStrategy = existAlarm.AlarmStrategy
		alarm.StrategyId = existAlarm.StrategyId
		alarm.Status = "ok"
		alarm.EndValue = alertValue
		alarm.End = nowTime
		alarm.AlarmConditionGuid = alarmConditionGuid
	} else if operation == "add" {
		if !db.InActiveWindowList(strategyObj.ActiveWindow) {
			return alarm, fmt.Errorf("Alarm:%s not in active window:%s ", strategyObj.Guid, strategyObj.ActiveWindow)
		}
		alarm.StartValue = alertValue
		alarm.Start = nowTime
	}
	return
}

func checkIsInActiveWindow(input string) bool {
	if input == "" {
		return true
	}
	timeSplit := strings.Split(input, "-")
	if len(timeSplit) != 2 {
		log.Logger.Error("Active window illegal", log.String("input", input))
		return false
	}
	nowTime := time.Now()
	dayPrefix := nowTime.Format("2006-01-02")
	st, sErr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s:00", dayPrefix, timeSplit[0]), time.Local)
	if sErr != nil {
		log.Logger.Error("Active window start illegal", log.String("start", timeSplit[0]))
		return false
	}
	endString := timeSplit[1] + ":00"
	if strings.HasSuffix(timeSplit[1], "59") {
		endString = timeSplit[1] + ":59"
	}
	et, eErr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", dayPrefix, endString), time.Local)
	if eErr != nil {
		log.Logger.Error("Active window end illegal", log.String("end", timeSplit[1]))
		return false
	}
	if nowTime.Unix() >= st.Unix() && nowTime.Unix() <= et.Unix() {
		return true
	}
	return false
}

func getNewAlarmEndpoint(param *m.AMRespAlert, strategyObj *m.AlarmStrategyMetricObj) (result m.EndpointNewTable, err error) {
	result = m.EndpointNewTable{}
	if param.Labels["process_guid"] != "" {
		result.Guid = param.Labels["process_guid"]
	} else if param.Labels["t_endpoint"] != "" {
		result.Guid = param.Labels["t_endpoint"]
	} else if param.Labels["guid"] != "" {
		result.Guid = param.Labels["guid"]
	} else if param.Labels["e_guid"] != "" {
		result.Guid = param.Labels["e_guid"]
	} else if param.Labels["instance"] != "" && param.Labels["instance"] != "127.0.0.1:8181" {
		result.AgentAddress = param.Labels["instance"]
		//if result.AgentAddress == "127.0.0.1:8181" {
		//	if param.Labels["service_group"] != "" {
		//		result = m.EndpointNewTable{Guid: "sg__" + param.Labels["service_group"]}
		//		return
		//	}
		//}
		//if strings.Contains(result.AgentAddress, "9100") {
		//	result.MonitorType = "host"
		//}
	} else if param.Labels["strategy_guid"] != "" {
		endpointGroupObj, tmpErr := db.GetSimpleEndpointGroup(strategyObj.EndpointGroup)
		if tmpErr != nil {
			return result, fmt.Errorf("alert labels have no endpoint_group:%s message ", strategyObj.EndpointGroup)
		}
		if endpointGroupObj.ServiceGroup != "" {
			result = m.EndpointNewTable{Guid: "sg__" + endpointGroupObj.ServiceGroup}
		} else {
			result = m.EndpointNewTable{Guid: "eg__" + endpointGroupObj.Guid}
		}
		return
	} else {
		return result, fmt.Errorf("alert labels have no endpoint message ")
	}
	result, err = db.GetEndpointNew(&result)
	if err != nil {
		return
	}
	if result.AlarmEnable == 0 {
		err = fmt.Errorf("Endpoint %s alarm is disable ", result.Guid)
		return
	}
	if !db.CheckEndpointActiveAlert(result.Guid) {
		err = fmt.Errorf("Endpoint %s in alert maintain window ", result.Guid)
	}
	return
}

func getNewAlarmTags(param *m.AMRespAlert) (tagString string, err error) {
	var sortTagList m.DefaultSortList
	for labelKey, labelValue := range param.Labels {
		sortTagList = append(sortTagList, &m.DefaultSortObj{Key: labelKey, Value: labelValue})
	}
	sort.Sort(sortTagList)
	var guidTagString, eGuidTagString string
	for _, label := range sortTagList {
		if label.Key == "strategy_id" || label.Key == "job" || label.Key == "instance" || label.Key == "alertname" || label.Key == "strategy_guid" {
			continue
		}
		if label.Key == "guid" {
			guidTagString = label.Value
		}
		if label.Key == "e_guid" {
			eGuidTagString = label.Value
		}
		tmpLabelValue := label.Value
		tagString += fmt.Sprintf("%s:%s^", label.Key, tmpLabelValue)
	}
	if guidTagString != "" && eGuidTagString != "" {
		if guidTagString != eGuidTagString {
			log.Logger.Warn("EGuid diff with guid,ignore", log.String("guid", guidTagString), log.String("e_guid", eGuidTagString))
			err = fmt.Errorf("EGuid diff with guid,ignore ")
			return
		}
	}
	if tagString != "" {
		tagString = tagString[:len(tagString)-1]
	}
	return
}

func getNewAlarmWithStrategyGuid(alarm *m.AlarmHandleObj, param *m.AMRespAlert, endpointObj *m.EndpointNewTable, strategyObj *m.AlarmStrategyMetricObj, multipleConditionFlag bool) (existAlarm m.AlarmTable, alarmConditionGuid string, err error) {
	existAlarm = m.AlarmTable{}
	log.Logger.Debug("getNewAlarmWithStrategyGuid", log.JsonObj("strategyObj", strategyObj))
	alarm.AlarmStrategy = strategyObj.Guid
	alarm.SMetric = strategyObj.MetricName
	alarm.SExpr = strategyObj.MetricExpr
	alarm.SCond = strategyObj.Condition
	alarm.SLast = strategyObj.Last
	alarm.SPriority = strategyObj.Priority
	if len(param.Labels) > 0 {
		alarm.Content = strings.ReplaceAll(param.Annotations["description"], "{code}", param.Labels["code"])
	} else {
		alarm.Content = param.Annotations["description"]
	}
	alarm.NotifyEnable = strategyObj.NotifyEnable
	alarm.NotifyDelay = strategyObj.NotifyDelaySecond
	if strings.Contains(alarm.SMetric, "ping_alive") || strings.Contains(alarm.SMetric, "telnet_alive") || strings.Contains(alarm.SMetric, "http_alive") {
		if endpointObj.Ip == m.Config().Peer.InstanceHostIp {
			err = fmt.Errorf("Ignore check alive alarm,is self instance host,ip:%s ", endpointObj.Ip)
			return
		}
	}
	if multipleConditionFlag {
		existAlarm, alarmConditionGuid, err = db.GetExistAlarmCondition(alarm.Endpoint, strategyObj.Guid, strategyObj.ConditionCrc, alarm.Tags)
		return
	}
	existAlarmQuery := m.AlarmTable{Endpoint: alarm.Endpoint, Tags: alarm.Tags, AlarmStrategy: alarm.AlarmStrategy, Status: "firing"}
	existAlarm, _ = db.GetAlarmObj(&existAlarmQuery)
	return
}

func getNewAlarmWithStrategyId(alarm *m.AlarmHandleObj, param *m.AMRespAlert, endpointObj *m.EndpointNewTable) (existAlarm m.AlarmTable, err error) {
	existAlarm = m.AlarmTable{}
	alarm.StrategyId, _ = strconv.Atoi(param.Labels["strategy_id"])
	if alarm.StrategyId <= 0 {
		err = fmt.Errorf("Alert's strategy id is null ")
		return
	}
	strategyList, getStrategyErr := db.GetStrategyList(alarm.StrategyId, 0)
	if getStrategyErr != nil {
		err = fmt.Errorf("Alert's strategy:%d fetch error:%s ", alarm.StrategyId, getStrategyErr.Error())
		return
	}
	if len(strategyList) == 0 {
		err = fmt.Errorf("Alert's strategy:%d can not find ", alarm.StrategyId)
		return
	}
	strategyObj := strategyList[0]
	alarm.SMetric = strategyObj.Metric
	alarm.SExpr = strategyObj.Expr
	alarm.SCond = strategyObj.Cond
	alarm.SLast = strategyObj.Last
	alarm.SPriority = strategyObj.Priority
	alarm.Content = param.Annotations["description"]
	alarm.NotifyEnable = strategyObj.NotifyEnable
	alarm.NotifyDelay = strategyObj.NotifyDelay
	if strings.Contains(alarm.SMetric, "ping_alive") || strings.Contains(alarm.SMetric, "telnet_alive") || strings.Contains(alarm.SMetric, "http_alive") {
		if endpointObj.Ip == m.Config().Peer.InstanceHostIp {
			err = fmt.Errorf("Ignore check alive alarm,is self instance host,ip:%s ", endpointObj.Ip)
			return
		}
	}
	existAlarmQuery := m.AlarmTable{Endpoint: alarm.Endpoint, StrategyId: alarm.StrategyId, Tags: alarm.Tags, Status: "firing"}
	existAlarm, _ = db.GetAlarmObj(&existAlarmQuery)
	return
}

func getNewAlarmWithUpCase(alarm *m.AlarmHandleObj, param *m.AMRespAlert) (existAlarm m.AlarmTable, err error) {
	alarm.SMetric = "up"
	alarm.SExpr = "up"
	alarm.SCond = "<1"
	alarm.SLast = "30s"
	alarm.SPriority = "high"
	alarm.Content = param.Annotations["description"]
	existAlarmQuery := m.AlarmTable{Endpoint: alarm.Endpoint, SMetric: alarm.SMetric, Status: "firing"}
	existAlarm, _ = db.GetAlarmObj(&existAlarmQuery)
	return
}

func GetHistoryAlarm(c *gin.Context) {
	idParam := c.Query("id")
	page := c.Query("page")
	pageSize := c.Query("pageSize")
	serviceGroup := c.Query("serviceGroup")
	var err error
	var endpoint string
	var startTime, endTime time.Time
	var pageInfo m.PageInfo
	var returnData m.AlarmHistoryReturnData
	start := c.Query("start")
	end := c.Query("end")
	pageInt, _ := strconv.Atoi(page)
	if pageInt == 0 {
		pageInt = 1
	}
	pageSizeInt, _ := strconv.Atoi(pageSize)
	if pageSizeInt == 0 {
		pageSizeInt = 10
	}
	if start != "" {
		tmpStartTime, err := time.Parse(m.DatetimeFormat, start)
		if err == nil {
			startTime = tmpStartTime
		} else {
			mid.ReturnParamTypeError(c, "start", m.DatetimeFormat)
			return
		}
	}
	if end != "" {
		tmpEndTime, err := time.Parse(m.DatetimeFormat, end)
		if err == nil {
			endTime = tmpEndTime
		} else {
			mid.ReturnParamTypeError(c, "end", m.DatetimeFormat)
			return
		}
	}
	// 层级对象处理
	if strings.TrimSpace(serviceGroup) != "" && idParam == "-1" {
		endpoint = "sg__" + serviceGroup
	} else {
		endpointId, _ := strconv.Atoi(idParam)
		if endpointId > 0 {
			endpointObj := m.EndpointTable{Id: endpointId}
			err = db.GetEndpoint(&endpointObj)
			if err != nil || endpointObj.Guid == "" {
				mid.ReturnValidateError(c, fmt.Sprintf("Endpoint id:%d fetch data fail", endpointId))
				return
			}
			endpoint = endpointObj.Guid
		} else if idParam != "" {
			endpointObj := m.EndpointTable{Guid: idParam}
			err = db.GetEndpoint(&endpointObj)
			if err != nil || endpointObj.Guid == "" {
				mid.ReturnValidateError(c, fmt.Sprintf("Endpoint guid:%d fetch data fail", idParam))
				return
			}
			endpoint = endpointObj.Guid
		}
	}
	param := m.EndpointAlarmParam{
		Endpoint:  endpoint,
		StartTime: startTime,
		EndTime:   endTime,
		Page:      pageInt,
		PageSize:  pageSizeInt,
	}
	tmpData, totalRows, tmpErr := db.GetEndpointHistoryAlarm(param)
	if tmpErr != nil {
		err = tmpErr
	}
	if len(tmpData) > 0 {
		returnData = m.AlarmHistoryReturnData{Endpoint: serviceGroup, ProblemList: tmpData}
		pageInfo.StartIndex = (param.Page - 1) * param.PageSize
		pageInfo.PageSize = pageSizeInt
		pageInfo.TotalRows = totalRows
	}
	if err != nil {
		mid.ReturnHandleError(c, fmt.Sprintf("Get history data fail,%s ", err.Error()), err)
		return
	}
	mid.ReturnPageData(c, pageInfo, returnData)
}

func GetProblemAlarmOptions(c *gin.Context) {
	var err error
	var param m.AlarmOptionsParam
	if err = c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	var data = &m.ProblemAlarmOptions{
		EndpointList:  []m.AlarmEndpoint{},
		MetricList:    []string{},
		AlarmNameList: []string{},
	}
	// 查询全量
	if strings.TrimSpace(param.AlarmName) != "" {
		if data.AlarmNameList, err = db.GetAlarmNameList(param.Status, param.AlarmName); err != nil {
			mid.ReturnServerHandleError(c, err)
			return
		}
		mid.ReturnSuccessData(c, data)
		return
	}
	if strings.TrimSpace(param.Endpoint) != "" {
		if data.EndpointList, err = db.QueryEndpointList(param.Endpoint); err != nil {
			mid.ReturnServerHandleError(c, err)
			return
		}
		mid.ReturnSuccessData(c, data)
		return
	}
	if strings.TrimSpace(param.Metric) != "" {
		if data.MetricList, err = db.QueryMetricNameList(param.Metric); err != nil {
			mid.ReturnServerHandleError(c, err)
			return
		}
		mid.ReturnSuccessData(c, data)
		return
	}
	if data.AlarmNameList, err = db.GetAlarmNameList(param.Status, param.AlarmName); err != nil {
		mid.ReturnServerHandleError(c, err)
	}
	if data.EndpointList, err = db.QueryEndpointList(param.Endpoint); err != nil {
		mid.ReturnServerHandleError(c, err)
	}
	if data.MetricList, err = db.QueryMetricNameList(param.Metric); err != nil {
		mid.ReturnServerHandleError(c, err)
	}
	mid.ReturnSuccessData(c, data)
}

func GetProblemAlarm(c *gin.Context) {
	filters := c.QueryArray("filter[]")
	query := m.AlarmTable{Status: "firing"}
	for _, v := range filters {
		if strings.Contains(v, "=") {
			tmpSplit := strings.Split(v, "=")
			if tmpSplit[0] == "endpoint" {
				query.Endpoint = strings.Replace(tmpSplit[1], "\"", "", -1)
			}
			if tmpSplit[0] == "metric" {
				query.SMetric = strings.Replace(tmpSplit[1], "\"", "", -1)
			}
			if tmpSplit[0] == "priority" {
				query.SPriority = strings.Replace(tmpSplit[1], "\"", "", -1)
			}
		}
	}
	err, data := db.GetAlarms(m.QueryAlarmCondition{AlarmTable: query, ExtOpenAlarm: true, UserRoles: mid.GetOperateUserRoles(c)})
	if err != nil {
		mid.ReturnQueryTableError(c, "alarm", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func QueryProblemAlarm(c *gin.Context) {
	var param m.QueryProblemAlarmDto
	if err := c.ShouldBindJSON(&param); err == nil {
		query := m.AlarmTable{Status: "firing", Endpoint: param.Endpoint, SMetric: param.Metric, SPriority: param.Priority}
		err, data := db.GetAlarms(m.QueryAlarmCondition{AlarmTable: query, ExtOpenAlarm: true, UserRoles: mid.GetOperateUserRoles(c)})
		if err != nil {
			mid.ReturnQueryTableError(c, "alarm", err)
			return
		}
		var highCount, mediumCount, lowCount int
		metricMap := make(map[string]int)
		for _, v := range data {
			if v.SPriority == "high" {
				highCount += 1
			}
			if v.SPriority == "medium" {
				mediumCount += 1
			}
			if v.SPriority == "low" {
				lowCount += 1
			}
			tmpMetricLevel := fmt.Sprintf("%s^%s", v.SMetric, v.SPriority)
			if _, b := metricMap[tmpMetricLevel]; b {
				metricMap[tmpMetricLevel] += 1
			} else {
				metricMap[tmpMetricLevel] = 1
			}
		}
		if len(data) == 0 {
			data = []*m.AlarmProblemQuery{}
		}
		var resultCount m.AlarmProblemCountList
		for k, v := range metricMap {
			tmpSplit := strings.Split(k, "^")
			resultCount = append(resultCount, &m.AlarmProblemCountObj{Name: tmpSplit[0], Type: tmpSplit[1], Value: v, FilterType: "metric"})
		}
		sort.Sort(resultCount)
		result := m.AlarmProblemQueryResult{Data: data, High: highCount, Mid: mediumCount, Low: lowCount, Count: resultCount}
		mid.ReturnSuccessData(c, result)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func QueryProblemAlarmByPage(c *gin.Context) {
	var param m.QueryProblemAlarmPageDto
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if param.Page == nil {
		param.Page = &m.PageInfo{StartIndex: 0, PageSize: 0}
	}
	if len(param.Endpoint) > 3 || len(param.AlarmName) > 3 || len(param.Metric) > 3 {
		mid.ReturnValidateError(c, "query data too large")
		return
	}
	query := m.AlarmTable{Status: "firing", Endpoint: "", SMetric: "", SPriority: "", AlarmName: ""}
	var endpointList []string
	var err error
	if param.CustomDashboardId > 0 {
		endpointList, err = db.GetCustomDashboardEndpointList(param.CustomDashboardId)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		if len(endpointList) == 0 {
			mid.ReturnSuccessData(c, m.AlarmProblemQueryResult{Data: []*m.AlarmProblemQuery{}, Count: []*m.AlarmProblemCountObj{}, Page: &m.PageInfo{}})
			return
		}
	}
	if len(param.Endpoint) > 0 {
		endpointList = append(endpointList, param.Endpoint...)
	}
	err, data := db.GetAlarms(m.QueryAlarmCondition{
		AlarmTable:          query,
		ExtOpenAlarm:        true,
		EndpointFilterList:  endpointList,
		MetricFilterList:    param.Metric,
		AlarmNameFilterList: param.AlarmName,
		PriorityList:        param.Priority,
		UserRoles:           mid.GetOperateUserRoles(c),
		Token:               c.GetHeader("Authorization"),
		Query:               param.Query,
	})
	if err != nil {
		mid.ReturnQueryTableError(c, "alarm", err)
		return
	}
	param.Page.TotalRows = len(data)
	var highCount, mediumCount, lowCount int
	metricMap := make(map[string]int)
	for _, v := range data {
		if v.SPriority == "high" {
			highCount += 1
		}
		if v.SPriority == "medium" {
			mediumCount += 1
		}
		if v.SPriority == "low" {
			lowCount += 1
		}
		if len(v.AlarmMetricList) > 0 {
			for _, subMetric := range v.AlarmMetricList {
				tmpMetricLevel := fmt.Sprintf("%s^%s", subMetric, v.SPriority)
				if _, b := metricMap[tmpMetricLevel]; b {
					metricMap[tmpMetricLevel] += 1
				} else {
					metricMap[tmpMetricLevel] = 1
				}
			}
		} else {
			tmpMetricLevel := fmt.Sprintf("%s^%s", v.SMetric, v.SPriority)
			if _, b := metricMap[tmpMetricLevel]; b {
				metricMap[tmpMetricLevel] += 1
			} else {
				metricMap[tmpMetricLevel] = 1
			}
		}
		if v.AlarmName == "" && v.Title != "" {
			v.AlarmName = v.Title
		}
		if v.Endpoint == "custom_alarm" {
			v.EndpointGuid = "custom_alarm"
		}
	}
	if len(data) == 0 {
		data = []*m.AlarmProblemQuery{}
	}
	var resultCount m.AlarmProblemCountList
	for k, v := range metricMap {
		tmpSplit := strings.Split(k, "^")
		resultCount = append(resultCount, &m.AlarmProblemCountObj{Name: tmpSplit[0], Type: tmpSplit[1], Value: v, FilterType: "metric"})
	}
	sort.Sort(resultCount)
	// page
	if param.Page.PageSize > 0 {
		si := (param.Page.StartIndex - 1) * param.Page.PageSize
		ei := param.Page.StartIndex*param.Page.PageSize - 1
		var pageResult []*m.AlarmProblemQuery
		for i, v := range data {
			if i >= si && i <= ei {
				pageResult = append(pageResult, v)
			}
		}
		data = pageResult
	}
	result := m.AlarmProblemQueryResult{Data: data, High: highCount, Mid: mediumCount, Low: lowCount, Count: resultCount, Page: param.Page}
	mid.ReturnSuccessData(c, result)
}

// @Summary 手动关闭告警接口
// @Produce  json
// @Param id query int true "告警id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/problem/close [get]
func CloseAlarm(c *gin.Context) {
	var param m.AlarmCloseParam
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if len(param.Metric) == 0 && param.Id == 0 && len(param.Priority) == 0 && len(param.Endpoint) == 0 && len(param.AlarmName) == 0 {
		mid.ReturnValidateError(c, "param can not empty")
		return
	}
	var actions []*db.Action
	actions, err = db.CloseAlarm(param)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	customActions, getCustomErr := db.CloseOpenAlarm(param)
	if getCustomErr != nil {
		mid.ReturnHandleError(c, getCustomErr.Error(), getCustomErr)
		return
	}
	actions = append(actions, customActions...)
	if len(actions) > 0 {
		err = db.Transaction(actions)
	}
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccess(c)
}

func UpdateAlarmCustomMessage(c *gin.Context) {
	var param m.UpdateAlarmCustomMessageDto
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateAlarmCustomMessage(param)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccess(c)
}

func OpenAlarmApi(c *gin.Context) {
	var param m.OpenAlarmRequest
	contentType := c.Request.Header.Get("Content-Type")
	if strings.Contains(contentType, "x-www-form-urlencoded") {
		var requestObj m.OpenAlarmObj
		requestObj.AlertInfo = c.PostForm("alert_info")
		requestObj.AlertIp = c.PostForm("alert_ip")
		if requestObj.AlertIp == "" {
			requestObj.AlertIp = c.ClientIP()
		}
		requestObj.AlertLevel = c.PostForm("alert_level")
		requestObj.AlertObj = c.PostForm("alert_obj")
		requestObj.AlertTitle = c.PostForm("alert_title")
		requestObj.RemarkInfo = c.PostForm("remark_info")
		requestObj.SubSystemId = c.PostForm("sub_system_id")
		requestObj.UseUmgPolicy = c.PostForm("use_umg_policy")
		requestObj.AlertWay = c.PostForm("alert_way")
		requestObj.AlertReciver = c.PostForm("alert_reciver")
		param.AlertList = []m.OpenAlarmObj{requestObj}
		err := db.SaveOpenAlarm(param)
		if err != nil {
			c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode: -1, ResultMsg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode: 0, ResultMsg: "success"})
	} else {
		if err := c.ShouldBindJSON(&param); err == nil {
			if len(param.AlertList) == 0 {
				c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode: -1, ResultMsg: "alertList is empty"})
				return
			}
			for _, v := range param.AlertList {
				if v.AlertIp == "" {
					v.AlertIp = c.ClientIP()
				}
			}
			err = db.SaveOpenAlarm(param)
			if err != nil {
				c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode: -1, ResultMsg: err.Error()})
			} else {
				c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode: 0, ResultMsg: "success"})
			}
		} else {
			c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode: -1, ResultMsg: err.Error()})
		}
	}
}

func QueryEntityAlarm(c *gin.Context) {
	var param m.EntityQueryParam
	var result m.AlarmEntity
	result.Data = []*m.AlarmEntityObj{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body.Close()
	err := json.Unmarshal(data, &param)
	if err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %v", err)
		mid.ReturnData(c, result)
		return
	}
	var id int
	var notifyGuid, alarmStatus string
	value := param.Criteria.Condition
	if strings.Contains(value, "monitor-check") {
		alarmObj := db.GetCheckProgressContent(value)
		alarmObj.Id = value
		alarmObj.DisplayName = alarmObj.Subject
		result.Data = append(result.Data, &alarmObj)
	} else {
		if strings.Contains(value, "-") {
			// id-firing-notifyGuid
			tmpSplit := strings.Split(value, "-")
			id, _ = strconv.Atoi(tmpSplit[0])
			if len(tmpSplit) > 1 {
				alarmStatus = tmpSplit[1]
			}
			if len(tmpSplit) > 2 {
				notifyGuid = tmpSplit[2]
			}
		} else {
			id, _ = strconv.Atoi(value)
		}
		var alarmObj m.AlarmEntityObj
		var err error
		if id <= 0 {
			log.Logger.Warn("Can not find alarm with empty id,get last one firing alarm", log.String("request param", string(data)))
			alarmObj, err = db.GetAlarmEvent("alarm", "", 0, "firing")
		} else {
			alarmObj, err = db.GetAlarmEvent("alarm", notifyGuid, id, alarmStatus)
		}
		if err != nil {
			result.Status = "ERROR"
			result.Message = fmt.Sprintf("error: %v", err)
			mid.ReturnData(c, result)
			return
		}
		alarmObj.Id = value
		alarmObj.DisplayName = alarmObj.Subject
		result.Data = append(result.Data, &alarmObj)
	}
	result.Status = "OK"
	result.Message = "Success"
	mid.ReturnData(c, result)
}

func TestNotifyAlarm(c *gin.Context) {
	endpoint := c.Query("endpoint")
	strategyId, _ := strconv.Atoi(c.Query("id"))
	err := db.NotifyCoreEvent(endpoint, strategyId, 0, 0)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		mid.ReturnSuccess(c)
	}
}

func GetCustomDashboardAlarm(c *gin.Context) {
	var param m.QueryProblemAlarmDto
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if param.Page == nil {
		param.Page = &m.PageInfo{StartIndex: 0, PageSize: 0}
	}
	customDashboardId, _ := strconv.Atoi(c.Param("customDashboardId"))
	if customDashboardId <= 0 {
		mid.ReturnParamEmptyError(c, "id")
		return
	}
	err, result := db.GetCustomDashboardAlarms(customDashboardId, param.Page)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		mid.ReturnSuccessData(c, result)
	}
}

func QueryHistoryAlarm(c *gin.Context) {
	var param m.QueryHistoryAlarmParam
	if err := c.ShouldBindJSON(&param); err == nil {
		param.Filter = "start"
		err, result := db.QueryHistoryAlarm(param)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		} else {
			mid.ReturnSuccessData(c, result)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetAlertWindowList(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		mid.ReturnParamEmptyError(c, "endpoint")
		return
	}
	data, err := db.GetAlertWindowList(endpoint)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateAlertWindow(c *gin.Context) {
	var param m.AlertWindowParam
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateAlertWindowList(param.Endpoint, mid.GetOperateUser(c), param.Data)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

// @Summary 手动触发告警回调事件接口
// @Produce  json
// @Param id query int true "告警id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/problem/notify [post]
func NotifyAlarm(c *gin.Context) {
	var param m.AlarmCloseParam
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if param.Id == 0 {
		mid.ReturnValidateError(c, "param can not empty")
		return
	}
	for _, v := range param.Metric {
		if strings.ToLower(v) == "custom" {
		}
		param.Custom = true
		break
	}
	err = db.ManualNotifyAlarm(param.Id, mid.GetOperateUser(c))
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccess(c)
}

func QueryEntityAlarmEvent(c *gin.Context) {
	var param m.EntityQueryParam
	var result m.AlarmEventEntity
	result.Data = []*m.AlarmEventEntityObj{}
	data, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body.Close()
	err := json.Unmarshal(data, &param)
	if err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %v", err)
		mid.ReturnData(c, result)
		return
	}
	var id int
	var operator, value string
	if param.Criteria.AttrName == "id" {
		value = param.Criteria.Condition
	} else {
		for _, v := range param.AdditionalFilters {
			if v.AttrName == "id" {
				value = v.Condition
			}
		}
	}
	if value == "" {
		result.Status = "OK"
		result.Message = "Success"
		result.Data = []*m.AlarmEventEntityObj{}
		mid.ReturnData(c, result)
		return
	}
	if strings.Contains(value, "-") {
		// id-firing-notifyGuid-operator
		tmpSplit := strings.Split(value, "-")
		id, _ = strconv.Atoi(tmpSplit[0])
		//if len(tmpSplit) > 1 {
		//	alarmStatus = tmpSplit[1]
		//}
		//if len(tmpSplit) > 2 {
		//	notifyGuid = tmpSplit[2]
		//}
		if len(tmpSplit) > 3 {
			operator = tmpSplit[3]
		}
	} else {
		id, _ = strconv.Atoi(value)
	}
	if id <= 0 {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("condition alarm id parse int fail, id:%s", value)
		mid.ReturnData(c, result)
		return
	}
	resultData, getErr := db.GetAlarmEventEntityData(id)
	if getErr != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("get alarm_event entity data fail: %s", getErr.Error())
	} else {
		result.Status = "OK"
		result.Message = "Success"
		resultData.Id = value
		resultData.Handler = operator
		result.Data = append(result.Data, resultData)
	}
	mid.ReturnData(c, result)
}

func UpdateEntityAlarm(c *gin.Context) {
	var param []map[string]interface{}
	var result m.AlarmEventUpdateResponse
	data, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body.Close()
	err := json.Unmarshal(data, &param)
	if err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %v", err)
		mid.ReturnData(c, result)
		return
	}

}
