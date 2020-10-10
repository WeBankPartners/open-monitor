package alarm

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"strconv"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"strings"
	"time"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"io/ioutil"
	"encoding/json"
	"sort"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

func AcceptAlertMsg(c *gin.Context)  {
	var param m.AlterManagerRespObj
	if err := c.ShouldBindJSON(&param); err==nil {
		if len(param.Alerts) == 0 {
			log.Logger.Warn("Accept alert is null")
			mid.ReturnSuccess(c)
		}
		log.Logger.Debug("accept", log.JsonObj("body", param))
		var alarms []*m.AlarmTable
		for _,v := range param.Alerts {
			if v.Labels["instance"] == "127.0.0.1:8300" {
				continue
			}
			log.Logger.Debug("Accept alert msg", log.JsonObj("alert", v))
			var tmpValue float64
			var tmpAlarms m.AlarmProblemList
			var tmpTags  string
			var sortTagList m.DefaultSortList
			tmpAlarm := m.AlarmTable{Status: v.Status}
			for labelKey,labelValue := range v.Labels {
				sortTagList = append(sortTagList, &m.DefaultSortObj{Key:labelKey, Value:labelValue})
			}
			sort.Sort(sortTagList)
			for _,label := range sortTagList {
				if label.Key == "strategy_id" || label.Key == "job" || label.Key == "instance" || label.Key == "alertname" {
					continue
				}
				tmpTags += fmt.Sprintf("%s:%s^", label.Key, label.Value)
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
				if endpointObj.ExportType == "telnet" || endpointObj.ExportType == "http" || endpointObj.ExportType == "ping" {
					log.Logger.Debug("Up alarm break,endpoint export type illegal", log.String("exportType", endpointObj.ExportType))
					continue
				}
				tmpAlarm.Endpoint = endpointObj.Guid
				tmpValue, _ = strconv.ParseFloat(tmpSummaryMsg[3], 64)
				tmpValue, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
				tmpAlarmQuery := m.AlarmTable{Endpoint: tmpAlarm.Endpoint, SMetric: tmpAlarm.SMetric}
				_, tmpAlarms = db.GetAlarms(tmpAlarmQuery)
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
				tmpAlarmQuery := m.AlarmTable{Endpoint: tmpAlarm.Endpoint, StrategyId: tmpAlarm.StrategyId, Tags:tmpAlarm.Tags, SCond:tmpAlarm.SCond, SLast:tmpAlarm.SLast}
				_, tmpAlarms = db.GetAlarms(tmpAlarmQuery)
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
			if tmpOperation == "resolve" {
				tmpAlarm = m.AlarmTable{Id:tmpAlarms[0].Id, Endpoint:tmpAlarms[0].Endpoint, StrategyId:tmpAlarms[0].StrategyId, Status:"ok", EndValue:tmpValue, End:time.Now()}
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
		if m.Config().Alert.Enable {
			for _,v := range alarms {
				var sao m.SendAlertObj
				accept := db.GetMailByStrategy(v.StrategyId)
				if len(accept) == 0 {
					continue
				}
				sao.Accept = accept
				sao.Subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", v.Status, v.SPriority, v.Endpoint, v.SMetric)
				sao.Content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s",v.Endpoint,v.Status,v.SMetric,v.StartValue,v.SCond,v.SLast,v.SPriority,v.Content,v.Start.Format(m.DatetimeFormat))
				other.SendSmtpMail(sao)
			}
		}
		if m.CoreUrl != "" {
			for _, v := range alarms {
				notifyErr := db.NotifyCoreEvent(v.Endpoint, v.StrategyId, 0, 0)
				if notifyErr != nil {
					log.Logger.Error("notify core event fail", log.Error(notifyErr))
				}
			}
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
	err,data := db.GetAlarms(query)
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
	err,data := db.GetAlarms(query)
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
		err,data := db.GetAlarms(query)
		if err != nil {
			mid.ReturnQueryTableError(c, "alarm", err)
			return
		}
		var highCount,mediumCount,lowCount int
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
		}
		result := m.AlarmProblemQueryResult{Data:data,High:highCount,Mid:mediumCount,Low:lowCount}
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
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.SaveOpenAlarm(param)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
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
	var guid string
	value := idSplit[1]
	if strings.Contains(value, "-") {
		tmpSplit := strings.Split(value, "-")
		id, _ = strconv.Atoi(tmpSplit[0])
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
	alarmObj,err := db.GetAlarmEvent("alarm", guid, id)
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
	var guid string
	value := param.Criteria.Condition
	if strings.Contains(value, "monitor-check") {
		alarmObj := db.GetCheckProgressContent(value)
		result.Data = append(result.Data, &alarmObj)
	}else {
		if strings.Contains(value, "-") {
			tmpSplit := strings.Split(value, "-")
			id, _ = strconv.Atoi(tmpSplit[0])
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
		alarmObj, err := db.GetAlarmEvent("alarm", guid, id)
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