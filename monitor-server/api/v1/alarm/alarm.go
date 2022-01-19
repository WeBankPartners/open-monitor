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
		alarms = append(alarms, &tmpAlarm)
	}
	alarms = db.UpdateAlarms(alarms)
	var treeventSendObj m.EventTreeventNotifyDto
	for _, v := range alarms {
		treeventSendObj.Data = append(treeventSendObj.Data, &m.EventTreeventNodeParam{EventId: fmt.Sprintf("%d", v.Id), Status: v.Status, Endpoint: v.Endpoint, StartUnix: v.Start.Unix(), Message: fmt.Sprintf("%s \n %s \n %.3f %s", v.Endpoint, v.SMetric, v.StartValue, v.SCond)})
		if v.NotifyEnable == 0 {
			continue
		}
		//go db.NotifyAlarm(v)
		go db.NotifyStrategyAlarm(v)
	}
	go db.NotifyTreevent(treeventSendObj)
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
	var strategyObj m.AlarmStrategyMetricObj
	if param.Labels["strategy_guid"] != "" {
		strategyObj, err = db.GetAlarmStrategy(param.Labels["strategy_guid"])
		if err != nil {
			return alarm, fmt.Errorf("Try to get alarm strategy with strategy_guid:%s fail,%s ", param.Labels["strategy_guid"], err.Error())
		}
		log.Logger.Debug("getNewAlarmWithStrategyGuid", log.String("query guid", strategyObj.Guid))
	}
	var endpointObj m.EndpointNewTable
	endpointObj, err = getNewAlarmEndpoint(param, &strategyObj)
	if err != nil {
		return
	}
	alarm.Endpoint = endpointObj.Guid
	var alertValue float64
	alertValue, _ = strconv.ParseFloat(summaryList[len(summaryList)-1], 64)
	alertValue, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", alertValue), 64)
	if param.Labels["strategy_id"] == "" && param.Labels["strategy_guid"] == "" {
		return alarm, fmt.Errorf("labels strategy_id and strategy_guid is empty ")
	}
	existAlarm := m.AlarmTable{}
	log.Logger.Debug("accept strategy_guid", log.String("strategy_guid", param.Labels["strategy_guid"]))
	if param.Labels["strategy_guid"] != "" {
		existAlarm, err = getNewAlarmWithStrategyGuid(&alarm, param, &endpointObj, &strategyObj)
	} else if param.Labels["strategy_id"] == "up" {
		if endpointObj.MonitorType != "host" {
			return alarm, fmt.Errorf("Up alarm break,endpoint:%s export type illegal ", endpointObj.Guid)
		}
		existAlarm, err = getNewAlarmWithUpCase(&alarm, param)
	} else {
		existAlarm, err = getNewAlarmWithStrategyId(&alarm, param, &endpointObj)
	}
	log.Logger.Debug("exist alarm", log.JsonObj("existAlarm", existAlarm))
	alarm.Status = param.Status
	operation := "add"
	if existAlarm.Status != "" {
		if existAlarm.Status == "firing" {
			if alarm.Status == "firing" {
				operation = "same"
			} else {
				operation = "resolve"
			}
		} else if existAlarm.Status == "ok" {
			if alarm.Status == "resolved" {
				operation = "same"
			}
		} else if existAlarm.Status == "closed" {
			if alarm.Status == "resolved" {
				operation = "same"
			}
		}
	}
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
	} else if operation == "add" {
		alarm.StartValue = alertValue
		alarm.Start = nowTime
	}
	return
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
	} else if param.Labels["instance"] != "" {
		result.AgentAddress = param.Labels["instance"]
		if strings.Contains(result.AgentAddress, "9100") {
			result.MonitorType = "host"
		}
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
		err = fmt.Errorf("Endpoint %s in alert window ", result.Guid)
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

func getNewAlarmWithStrategyGuid(alarm *m.AlarmHandleObj, param *m.AMRespAlert, endpointObj *m.EndpointNewTable, strategyObj *m.AlarmStrategyMetricObj) (existAlarm m.AlarmTable, err error) {
	existAlarm = m.AlarmTable{}
	log.Logger.Info("getNewAlarmWithStrategyGuid", log.String("query guid", strategyObj.Guid))
	alarm.AlarmStrategy = strategyObj.Guid
	alarm.SMetric = strategyObj.MetricName
	alarm.SExpr = strategyObj.MetricExpr
	alarm.SCond = strategyObj.Condition
	alarm.SLast = strategyObj.Last
	alarm.SPriority = strategyObj.Priority
	alarm.Content = param.Annotations["description"]
	alarm.NotifyEnable = strategyObj.NotifyEnable
	alarm.NotifyDelay = strategyObj.NotifyDelaySecond
	if strings.Contains(alarm.SMetric, "ping_alive") || strings.Contains(alarm.SMetric, "telnet_alive") || strings.Contains(alarm.SMetric, "http_alive") {
		if endpointObj.Ip == m.Config().Peer.InstanceHostIp {
			err = fmt.Errorf("Ignore check alive alarm,is self instance host,ip:%s ", endpointObj.Ip)
			return
		}
	}
	existAlarmQuery := m.AlarmTable{Endpoint: alarm.Endpoint, Tags: alarm.Tags, AlarmStrategy: alarm.AlarmStrategy}
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
	existAlarmQuery := m.AlarmTable{Endpoint: alarm.Endpoint, StrategyId: alarm.StrategyId, Tags: alarm.Tags}
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
	existAlarmQuery := m.AlarmTable{Endpoint: alarm.Endpoint, SMetric: alarm.SMetric}
	existAlarm, _ = db.GetAlarmObj(&existAlarmQuery)
	return
}

func GetHistoryAlarm(c *gin.Context) {
	endpointId, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	var ids []string
	var startTime, endTime time.Time
	if endpointId < 0 {
		guid := c.Query("guid")
		if guid == "" {
			mid.ReturnValidateError(c, "Param guid can not empty when id<0 ")
			return
		}
		err, recursiveObj := db.GetRecursivePanel(guid)
		if err != nil {
			mid.ReturnHandleError(c, fmt.Sprintf("Get recursive panel data fail %s", err.Error()), err)
			return
		}
		ids = recursiveHistoryEndpoint(&recursiveObj)
	} else {
		endpointObj := m.EndpointTable{Id: endpointId}
		err = db.GetEndpoint(&endpointObj)
		if err != nil || endpointObj.Guid == "" {
			mid.ReturnValidateError(c, fmt.Sprintf("Endpoint id:%d fetch data fail", endpointId))
			return
		}
		ids = append(ids, endpointObj.Guid)
	}
	start := c.Query("start")
	end := c.Query("end")
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
	var returnData []*m.AlarmHistoryReturnData
	for _, endpointGuid := range ids {
		tmpErr, tmpData := getEndpointHistoryAlarm(endpointGuid, startTime, endTime)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		returnData = append(returnData, &m.AlarmHistoryReturnData{Endpoint: endpointGuid, ProblemList: tmpData})
	}
	if err != nil {
		mid.ReturnHandleError(c, fmt.Sprintf("Get history data fail,%s ", err.Error()), err)
		return
	}
	mid.ReturnSuccessData(c, returnData)
}

func getEndpointHistoryAlarm(endpointGuid string, startTime, endTime time.Time) (err error, data m.AlarmProblemList) {
	endpointObj := m.EndpointTable{Guid: endpointGuid}
	err = db.GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("EndpointGuid:%s fetch endpoint fail, %s ", endpointGuid, err.Error()), data
	}
	query := m.AlarmTable{Endpoint: endpointObj.Guid, Start: startTime, End: endTime}
	err, data = db.GetAlarms(query, 0, true, false)
	return err, data
}

func recursiveHistoryEndpoint(input *m.RecursivePanelObj) []string {
	endpoints := []string{}
	if len(input.Children) > 0 {
		for _, v := range input.Children {
			for _, vv := range recursiveHistoryEndpoint(v) {
				existFlag := false
				for _, vvv := range endpoints {
					if vvv == vv {
						existFlag = true
						break
					}
				}
				if !existFlag {
					endpoints = append(endpoints, vv)
				}
			}
		}
	}
	for _, v := range input.Charts {
		for _, vv := range v.Endpoint {
			existFlag := false
			for _, vvv := range endpoints {
				if vvv == vv {
					existFlag = true
					break
				}
			}
			if !existFlag {
				endpoints = append(endpoints, vv)
			}
		}
	}
	return endpoints
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
	err, data := db.GetAlarms(query, 0, true, true)
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
		err, data := db.GetAlarms(query, 0, true, true)
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

// @Summary 手动关闭告警接口
// @Produce  json
// @Param id query int true "告警id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/problem/close [get]
func CloseAlarm(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	isCustom := strings.ToLower(c.Query("custom"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	if isCustom == "true" {
		err = db.CloseOpenAlarm(id)
	} else {
		err = db.CloseAlarm(id)
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
	customDashboardId, _ := strconv.Atoi(c.Query("id"))
	if customDashboardId <= 0 {
		mid.ReturnParamEmptyError(c, "id")
		return
	}
	err, result := db.GetCustomDashboardAlarms(customDashboardId)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		mid.ReturnSuccessData(c, result)
	}
}

func QueryHistoryAlarm(c *gin.Context) {
	var param m.QueryHistoryAlarmParam
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.Filter != "all" && param.Filter != "start" && param.Filter != "end" {
			mid.ReturnValidateError(c, "filter must in [all,start,end]")
			return
		}
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
