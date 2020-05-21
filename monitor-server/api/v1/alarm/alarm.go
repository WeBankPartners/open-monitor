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
)

func AcceptAlertMsg(c *gin.Context)  {
	var param m.AlterManagerRespObj
	if err := c.ShouldBindJSON(&param); err==nil {
		if len(param.Alerts) == 0 {
			mid.LogInfo("accept alert is null")
			mid.ReturnSuccess(c, "Success")
		}
		var alarms []*m.AlarmTable
		for _,v := range param.Alerts {
			if v.Labels["instance"] == "127.0.0.1:8300" {
				continue
			}
			mid.LogInfo(fmt.Sprintf("accept alert msg : %v", v))
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
				if len(tmpSummaryMsg) != 3 {
					mid.LogInfo(fmt.Sprintf("summary illegal %s", v.Annotations["summary"]))
					continue
				}
				endpointObj := m.EndpointTable{Address: tmpSummaryMsg[0], AddressAgent: tmpSummaryMsg[0]}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id <= 0 || endpointObj.StopAlarm == 1 {
					continue
				}
				if endpointObj.ExportType == "telnet" || endpointObj.ExportType == "http" || endpointObj.ExportType == "ping" {
					continue
				}
				tmpAlarm.Endpoint = endpointObj.Guid
				tmpValue, _ = strconv.ParseFloat(tmpSummaryMsg[2], 64)
				tmpValue, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
				tmpAlarmQuery := m.AlarmTable{Endpoint: tmpAlarm.Endpoint, SMetric: tmpAlarm.SMetric}
				_, tmpAlarms = db.GetAlarms(tmpAlarmQuery)
			}else {
				// config strategy
				tmpAlarm.StrategyId, _ = strconv.Atoi(v.Labels["strategy_id"])
				if tmpAlarm.StrategyId <= 0 {
					mid.LogInfo(fmt.Sprintf("Alert's strategy id is null : %v ", v))
					continue
				}
				_, strategyObj := db.GetStrategy(m.StrategyTable{Id: tmpAlarm.StrategyId})
				if strategyObj.Id <= 0 {
					mid.LogInfo(fmt.Sprintf("Alert's strategy id can not found : %d ", tmpAlarm.StrategyId))
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
					mid.LogInfo(fmt.Sprintf("Can't find the endpoint %v", v))
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
				mid.LogInfo(fmt.Sprintf("Accept alert msg ,firing repeat,do nothing! Msg: %v", v))
				continue
			}
			if tmpOperation == "resolve" {
				tmpAlarm = m.AlarmTable{Id:tmpAlarms[0].Id, Endpoint:tmpAlarms[0].Endpoint, StrategyId:tmpAlarms[0].StrategyId, Status:"ok", EndValue:tmpValue, End:time.Now()}
			}else if tmpOperation == "add" {
				tmpAlarm.StartValue = tmpValue
				tmpAlarm.Start = time.Now()
			}
			mid.LogInfo(fmt.Sprintf("add alarm ,operation: %s ,value: %v", tmpOperation, tmpAlarm))
			alarms = append(alarms, &tmpAlarm)
		}
		err = db.UpdateAlarms(alarms)
		if err != nil {
			mid.ReturnError(c, "Failed to accept alert msg", err)
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
				notifyErr := db.NotifyCoreEvent(v.Endpoint, v.StrategyId)
				if notifyErr != nil {
					mid.LogError("notify core event fail", notifyErr)
				}
			}
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func GetHistoryAlarm(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnValidateFail(c, "Endpoint id validation failed")
		return
	}
	start := c.Query("start")
	end := c.Query("end")
	endpointObj := m.EndpointTable{Id:endpointId}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		mid.ReturnError(c, "Get historical alerts failed", fmt.Errorf("can't find endpoint with id: %d", endpointId))
		return
	}
	query := m.AlarmTable{Endpoint:endpointObj.Guid}
	if start != "" {
		startTime,err := time.Parse(m.DatetimeFormat, start)
		if err == nil {
			query.Start = startTime
		}else{
			mid.ReturnValidateFail(c, "Date and time format should be "+m.DatetimeFormat)
			return
		}
	}
	if end != "" {
		endTime,err := time.Parse(m.DatetimeFormat, end)
		if err == nil {
			query.End = endTime
		}else{
			mid.ReturnValidateFail(c, "Date and time format should be "+m.DatetimeFormat)
			return
		}
	}
	err,data := db.GetAlarms(query)
	if err != nil {
		mid.ReturnError(c, "Get historical alerts failed", err)
		return
	}
	mid.ReturnData(c, data)
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
		mid.ReturnError(c, "Get alerts failed", err)
		return
	}
	mid.ReturnData(c, data)
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
		mid.ReturnValidateFail(c, "Parameter \"id\" validation failed")
		return
	}
	if isCustom == "true" {
		err = db.CloseOpenAlarm(id)
	}else {
		err = db.CloseAlarm(id)
	}
	if err != nil {
		mid.ReturnError(c, "Close alert failed", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}

func OpenAlarmApi(c *gin.Context)  {
	var param m.OpenAlarmRequest
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.SaveOpenAlarm(param)
		if err != nil {
			mid.ReturnError(c, "Send alarm api fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
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
	if strings.Contains(value, "-") {
		tmpSplit := strings.Split(value, "-")
		id, _ = strconv.Atoi(tmpSplit[0])
		guid = value[len(tmpSplit[0])+1:]
	}else{
		id, _ = strconv.Atoi(value)
	}
	if id <= 0 {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Query criteria condition: %s -> filter validation failed", param.Criteria.Condition)
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

func TestNotifyAlarm(c *gin.Context)  {
	endpoint := c.Query("endpoint")
	strategyId,_ := strconv.Atoi(c.Query("id"))
	err := db.NotifyCoreEvent(endpoint, strategyId)
	if err != nil {
		mid.ReturnError(c, "", err)
	}else{
		mid.ReturnSuccess(c, "Success")
	}
}