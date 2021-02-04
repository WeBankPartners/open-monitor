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

func AcceptAlertMsg(c *gin.Context)  {
	var param m.AlterManagerRespObj
	if err := c.ShouldBindJSON(&param); err==nil {
		if len(param.Alerts) == 0 {
			log.Logger.Warn("Accept alert is null")
			mid.ReturnSuccess(c)
		}
		log.Logger.Debug("accept", log.JsonObj("body", param))
		var alarms []*m.AlarmHandleObj
		for _,v := range param.Alerts {
			if v.Labels["instance"] == "127.0.0.1:8300" {
				continue
			}
			var tmpValue float64
			var tmpAlarms m.AlarmProblemList
			var tmpTags  string
			var sortTagList m.DefaultSortList
			tmpAlarm := m.AlarmHandleObj{}
			tmpAlarm.Status = v.Status
			for labelKey,labelValue := range v.Labels {
				sortTagList = append(sortTagList, &m.DefaultSortObj{Key:labelKey, Value:labelValue})
			}
			sort.Sort(sortTagList)
			var guidTagString,eGuidTagString string
			for _,label := range sortTagList {
				if label.Key == "strategy_id" || label.Key == "job" || label.Key == "instance" || label.Key == "alertname" {
					continue
				}
				if label.Key == "guid" {
					guidTagString = label.Value
				}
				if label.Key == "e_guid" {
					eGuidTagString = label.Value
				}
				tmpLabelValue := label.Value
				tmpTags += fmt.Sprintf("%s:%s^", label.Key, tmpLabelValue)
			}
			if guidTagString != "" && eGuidTagString != "" {
				if guidTagString != eGuidTagString {
					log.Logger.Warn("EGuid diff with guid,ignore", log.String("guid", guidTagString), log.String("e_guid", eGuidTagString))
					continue
				}
			}
			if tmpTags != "" {
				tmpTags = tmpTags[:len(tmpTags)-1]
			}
			tmpAlarm.Tags = tmpTags
			if v.Labels["strategy_id"] == "up" {
				// base strategy
				tmpAlarm.SMetric = "up"
				tmpAlarm.SExpr = "up"
				tmpAlarm.SCond = "<1"
				tmpAlarm.SLast = "30s"
				tmpAlarm.SPriority = "high"
				tmpAlarm.Content = v.Annotations["description"]
				tmpSummaryMsg := strings.Split(v.Annotations["summary"], "__")
				if len(tmpSummaryMsg) != 4 {
					log.Logger.Warn("Summary illegal", log.String("summary", v.Annotations["summary"]))
					continue
				}
				endpointObj := m.EndpointTable{Address: tmpSummaryMsg[0], AddressAgent: tmpSummaryMsg[0]}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id <= 0 || endpointObj.StopAlarm == 1 {
					log.Logger.Debug("Up alarm break,endpoint not exists or stop alarm", log.String("endpoint", endpointObj.Guid))
					continue
				}
				if !db.CheckEndpointActiveAlert(endpointObj.Guid) {
					log.Logger.Info("Check endpoint is in maintain window,continue", log.String("endpoint", endpointObj.Guid))
					continue
				}
				if endpointObj.ExportType != "host" {
					log.Logger.Debug("Up alarm break,endpoint export type illegal", log.String("exportType", endpointObj.ExportType))
					continue
				}
				tmpAlarm.Endpoint = endpointObj.Guid
				tmpValue, _ = strconv.ParseFloat(tmpSummaryMsg[3], 64)
				tmpValue, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
				tmpAlarmQuery := m.AlarmTable{Endpoint: tmpAlarm.Endpoint, SMetric: tmpAlarm.SMetric}
				_, tmpAlarms = db.GetAlarms(tmpAlarmQuery, 1, false, false)
			}else {
				// config strategy
				tmpAlarm.StrategyId, _ = strconv.Atoi(v.Labels["strategy_id"])
				if tmpAlarm.StrategyId <= 0 {
					log.Logger.Warn("Alert's strategy id is null")
					continue
				}
				_, strategyObj := db.GetStrategy(m.StrategyTable{Id: tmpAlarm.StrategyId})
				if strategyObj.Id <= 0 {
					log.Logger.Warn("Alert's strategy id can not found", log.Int("id", tmpAlarm.StrategyId))
					continue
				}
				tmpAlarm.SMetric = strategyObj.Metric
				tmpAlarm.SExpr = strategyObj.Expr
				tmpAlarm.SCond = strategyObj.Cond
				tmpAlarm.SLast = strategyObj.Last
				tmpAlarm.SPriority = strategyObj.Priority
				tmpAlarm.Content = v.Annotations["description"]
				tmpAlarm.NotifyEnable = strategyObj.NotifyEnable
				tmpAlarm.NotifyDelay = strategyObj.NotifyDelay
				tmpSummaryMsg := strings.Split(v.Annotations["summary"], "__")
				var tmpEndpointIp string
				if len(tmpSummaryMsg) == 4 {
					var endpointObj m.EndpointTable
					if v.Labels["guid"] != "" {
						endpointObj = m.EndpointTable{Guid:v.Labels["guid"]}
					}else {
						endpointObj = m.EndpointTable{Address: tmpSummaryMsg[0], AddressAgent: tmpSummaryMsg[0]}
					}
					db.GetEndpoint(&endpointObj)
					if endpointObj.Id > 0 {
						tmpAlarm.Endpoint = endpointObj.Guid
						tmpEndpointIp = endpointObj.Ip
						if endpointObj.StopAlarm == 1 {
							continue
						}
						if !db.CheckEndpointActiveAlert(endpointObj.Guid) {
							log.Logger.Info("Check endpoint is in maintain window,continue", log.String("endpoint", endpointObj.Guid))
							continue
						}
					}
					tmpValue, _ = strconv.ParseFloat(tmpSummaryMsg[3], 64)
					tmpValue, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
				}
				if tmpAlarm.Endpoint == "" {
					log.Logger.Warn("Can't find the endpoint")
					continue
				}
				if strings.Contains(tmpAlarm.SMetric, "ping_alive") || strings.Contains(tmpAlarm.SMetric, "telnet_alive") || strings.Contains(tmpAlarm.SMetric, "http_alive") {
					if len(m.Config().Cluster.ServerList) > 0 {
						if m.Config().Cluster.ServerList[0] == tmpEndpointIp {
							continue
						}
					}
					if tmpValue != 0 && tmpValue != 1 && tmpValue != 2 {
						continue
					}
				}
				tmpAlarmQuery := m.AlarmTable{Endpoint: tmpAlarm.Endpoint, StrategyId: tmpAlarm.StrategyId, Tags:tmpAlarm.Tags}
				_, tmpAlarms = db.GetAlarms(tmpAlarmQuery, 1, false, false)
			}
			tmpOperation := "add"
			if len(tmpAlarms) > 0 {
				if tmpAlarms[0].Status == "firing" {
					if v.Status == "firing" {
						tmpOperation = "same"
					}else{
						tmpOperation = "resolve"
					}
				}else if tmpAlarms[0].Status == "ok" {
					if v.Status == "resolved" {
						tmpOperation = "same"
					}
				}else if tmpAlarms[0].Status == "closed" {
					if v.Status == "resolved" {
						tmpOperation = "same"
					}
				}
			}
			if tmpOperation == "same" {
				log.Logger.Debug("Accept alert msg ,firing repeat,do nothing!")
				continue
			}
			if tmpOperation == "add" && v.Status == "resolved" {
				log.Logger.Debug("Accept alert msg ,cat not add resolved,do nothing!")
				continue
			}
			if tmpOperation == "resolve" {
				tmpAlarm.Id = tmpAlarms[0].Id
				tmpAlarm.Endpoint = tmpAlarms[0].Endpoint
				tmpAlarm.StrategyId = tmpAlarms[0].StrategyId
				tmpAlarm.Status = "ok"
				tmpAlarm.EndValue = tmpValue
				tmpAlarm.End = time.Now()
			}else if tmpOperation == "add" {
				tmpAlarm.StartValue = tmpValue
				tmpAlarm.Start = time.Now()
			}
			log.Logger.Debug("Add alarm", log.String("operation", tmpOperation), log.JsonObj("alarm", tmpAlarm))
			alarms = append(alarms, &tmpAlarm)
		}
		err = db.UpdateAlarms(alarms)
		if err != nil {
			mid.ReturnUpdateTableError(c, "alarm", err)
			return
		}
		for _,v := range alarms {
			if v.NotifyEnable == 0 {
				continue
			}
			go db.NotifyAlarm(v)
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetHistoryAlarm(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	start := c.Query("start")
	end := c.Query("end")
	endpointObj := m.EndpointTable{Id:endpointId}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		mid.ReturnFetchDataError(c, "endpoint", "id", strconv.Itoa(endpointId))
		return
	}
	query := m.AlarmTable{Endpoint:endpointObj.Guid}
	if start != "" {
		startTime,err := time.Parse(m.DatetimeFormat, start)
		if err == nil {
			query.Start = startTime
		}else{
			mid.ReturnParamTypeError(c, "start", m.DatetimeFormat)
			return
		}
	}
	if end != "" {
		endTime,err := time.Parse(m.DatetimeFormat, end)
		if err == nil {
			query.End = endTime
		}else{
			mid.ReturnParamTypeError(c, "end", m.DatetimeFormat)
			return
		}
	}
	err,data := db.GetAlarms(query, 0, true, false)
	if err != nil {
		mid.ReturnQueryTableError(c, "alarm", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func GetProblemAlarm(c *gin.Context)  {
	filters := c.QueryArray("filter[]")
	query := m.AlarmTable{Status:"firing"}
	for _,v := range filters {
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
	err,data := db.GetAlarms(query, 0, true, true)
	if err != nil {
		mid.ReturnQueryTableError(c, "alarm", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func QueryProblemAlarm(c *gin.Context)  {
	var param m.QueryProblemAlarmDto
	if err := c.ShouldBindJSON(&param);err == nil {
		query := m.AlarmTable{Status:"firing", Endpoint:param.Endpoint, SMetric:param.Metric, SPriority:param.Priority}
		err,data := db.GetAlarms(query, 0, true, true)
		if err != nil {
			mid.ReturnQueryTableError(c, "alarm", err)
			return
		}
		var highCount,mediumCount,lowCount int
		metricMap := make(map[string]int)
		for _,v := range data {
			if v.SPriority == "high" {
				highCount += 1
			}
			if v.SPriority == "medium" {
				mediumCount += 1
			}
			if v.SPriority == "low" {
				lowCount += 1
			}
			if _,b:=metricMap[v.SMetric];b {
				metricMap[v.SMetric] += 1
			}else{
				metricMap[v.SMetric] = 1
			}
		}
		if len(data) == 0 {
			data = []*m.AlarmProblemQuery{}
		}
		result := m.AlarmProblemQueryResult{Data:data,High:highCount,Mid:mediumCount,Low:lowCount,MetricMap: metricMap}
		mid.ReturnSuccessData(c, result)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

// @Summary 手动关闭告警接口
// @Produce  json
// @Param id query int true "告警id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/problem/close [get]
func CloseALarm(c *gin.Context)  {
	id,err := strconv.Atoi(c.Query("id"))
	isCustom := strings.ToLower(c.Query("custom"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	if isCustom == "true" {
		err = db.CloseOpenAlarm(id)
	}else {
		err = db.CloseAlarm(id)
	}
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccess(c)
}

func OpenAlarmApi(c *gin.Context)  {
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
			c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode:-1, ResultMsg:err.Error()})
			return
		}
		c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode:0, ResultMsg:"success"})
	}else {
		if err := c.ShouldBindJSON(&param); err == nil {
			for _,v := range param.AlertList {
				if v.AlertIp == "" {
					v.AlertIp = c.ClientIP()
				}
			}
			err = db.SaveOpenAlarm(param)
			if err != nil {
				c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode:-1, ResultMsg:err.Error()})
			} else {
				c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode:0, ResultMsg:"success"})
			}
		} else {
			c.JSON(http.StatusOK, m.OpenAlarmResponse{ResultCode:-1, ResultMsg:err.Error()})
		}
	}
}

func GetEntityAlarm(c *gin.Context)  {
	var result m.AlarmEntity
	result.Data = []*m.AlarmEntityObj{}
	idSplit := strings.Split(c.Query("filter"), ",")
	if len(idSplit) < 2 {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Parameter %s -> filter validation failed", c.Query("filter"))
		mid.ReturnData(c, result)
		return
	}
	var id int
	var guid,alarmStatus string
	value := idSplit[1]
	if strings.Contains(value, "-") {
		tmpSplit := strings.Split(value, "-")
		id, _ = strconv.Atoi(tmpSplit[0])
		if len(tmpSplit) > 1 {
			alarmStatus = tmpSplit[1]
		}
		guid = value[len(tmpSplit[0])+1:]
	}else{
		id, _ = strconv.Atoi(value)
	}
	if id <= 0 {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Parameter %s -> filter validation failed", c.Query("filter"))
		mid.ReturnData(c, result)
		return
	}
	alarmObj,err := db.GetAlarmEvent("alarm", guid, id, alarmStatus)
	if err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("error: %v", err)
		mid.ReturnData(c, result)
		return
	}
	result.Data = append(result.Data, &alarmObj)
	result.Status = "OK"
	result.Message = "Success"
	mid.ReturnData(c, result)
}

func QueryEntityAlarm(c *gin.Context)  {
	var param m.EntityQueryParam
	var result m.AlarmEntity
	result.Data = []*m.AlarmEntityObj{}
	data,_ := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	err := json.Unmarshal(data, &param)
	if err != nil {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Request body json unmarshal failed: %v", err)
		mid.ReturnData(c, result)
		return
	}
	var id int
	var guid,alarmStatus string
	value := param.Criteria.Condition
	if strings.Contains(value, "monitor-check") {
		alarmObj := db.GetCheckProgressContent(value)
		result.Data = append(result.Data, &alarmObj)
	}else {
		if strings.Contains(value, "-") {
			tmpSplit := strings.Split(value, "-")
			id, _ = strconv.Atoi(tmpSplit[0])
			if len(tmpSplit) > 1 {
				alarmStatus = tmpSplit[1]
			}
			guid = value[len(tmpSplit[0])+1:]
		} else {
			id, _ = strconv.Atoi(value)
		}
		if id <= 0 {
			result.Status = "ERROR"
			result.Message = fmt.Sprintf("Query criteria condition: %s -> filter validation failed", param.Criteria.Condition)
			mid.ReturnData(c, result)
			return
		}
		alarmObj, err := db.GetAlarmEvent("alarm", guid, id, alarmStatus)
		if err != nil {
			result.Status = "ERROR"
			result.Message = fmt.Sprintf("error: %v", err)
			mid.ReturnData(c, result)
			return
		}
		result.Data = append(result.Data, &alarmObj)
	}
	result.Status = "OK"
	result.Message = "Success"
	mid.ReturnData(c, result)
}

func TestNotifyAlarm(c *gin.Context)  {
	endpoint := c.Query("endpoint")
	strategyId,_ := strconv.Atoi(c.Query("id"))
	err := db.NotifyCoreEvent(endpoint, strategyId, 0, 0)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	}else{
		mid.ReturnSuccess(c)
	}
}

func GetCustomDashboardAlarm(c *gin.Context)  {
	customDashboardId,_ := strconv.Atoi(c.Query("id"))
	if customDashboardId <= 0 {
		mid.ReturnParamEmptyError(c, "id")
		return
	}
	err,result := db.GetCustomDashboardAlarms(customDashboardId)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	}else{
		mid.ReturnSuccessData(c, result)
	}
}

func QueryHistoryAlarm(c *gin.Context)  {
	var param m.QueryHistoryAlarmParam
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Filter != "all" && param.Filter != "start" && param.Filter != "end" {
			mid.ReturnValidateError(c, "filter must in [all,start,end]")
			return
		}
		err,result := db.QueryHistoryAlarm(param)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		}else{
			mid.ReturnSuccessData(c, result)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetAlertWindowList(c *gin.Context)  {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		mid.ReturnParamEmptyError(c, "endpoint")
		return
	}
	data,err := db.GetAlertWindowList(endpoint)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	}else{
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateAlertWindow(c *gin.Context)  {
	var param m.AlertWindowParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateAlertWindowList(param.Endpoint, mid.GetOperateUser(c), param.Data)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}