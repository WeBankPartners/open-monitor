package alarm

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"strconv"
	"fmt"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	"strings"
	"time"
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
			mid.LogInfo(fmt.Sprintf("accept alert msg : %v", v))
			tmpAlarm := m.AlarmTable{Status:v.Status}
			tmpAlarm.StrategyId,_ = strconv.Atoi(v.Labels["strategy_id"])
			if tmpAlarm.StrategyId <= 0 {
				mid.LogInfo(fmt.Sprintf("Alerts strategy id is null : %v ", v))
				continue
			}
			_,strategyObj := db.GetStrategy(m.StrategyTable{Id:tmpAlarm.StrategyId})
			if strategyObj.Id <= 0 {
				mid.LogInfo(fmt.Sprintf("Alerts strategy id can't fetch in table : %d ", tmpAlarm.StrategyId))
				continue
			}
			tmpAlarm.SMetric = strategyObj.Metric
			tmpAlarm.SExpr = strategyObj.Expr
			tmpAlarm.SCond = strategyObj.Cond
			tmpAlarm.SLast = strategyObj.Last
			tmpAlarm.SPriority = strategyObj.Priority
			tmpAlarm.Content = v.Annotations["description"]
			tmpSummaryMsg := strings.Split(v.Annotations["summary"], "__")
			var tmpValue float64
			if len(tmpSummaryMsg) == 4 {
				endpointObj := m.EndpointTable{Address:tmpSummaryMsg[0]}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id > 0 {
					tmpAlarm.Endpoint = endpointObj.Guid
				}
				tmpValue,_ = strconv.ParseFloat(tmpSummaryMsg[3], 10)
			}
			if tmpAlarm.Endpoint == "" {
				mid.LogInfo(fmt.Sprintf("Can't find the endpoint %v", v))
				continue
			}
			tmpAlarmQuery := m.AlarmTable{Endpoint:tmpAlarm.Endpoint, StrategyId:tmpAlarm.StrategyId}
			_,tmpAlarms := db.GetAlarms(tmpAlarmQuery)
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
				}
			}
			if tmpOperation == "same" {
				mid.LogInfo(fmt.Sprintf("Accept alert msg ,firing repeat,do nothing! Msg: %v", v))
				continue
			}
			if tmpOperation == "resolve" {
				tmpAlarm = m.AlarmTable{Id:tmpAlarms[0].Id, Status:"ok", EndValue:tmpValue, End:time.Now()}
			}else if tmpOperation == "add" {
				tmpAlarm.StartValue = tmpValue
				tmpAlarm.Start = time.Now()
			}
			mid.LogInfo(fmt.Sprintf("add alarm ,operation: %s ,value: %v", tmpOperation, tmpAlarm))
			alarms = append(alarms, &tmpAlarm)
		}
		err = db.UpdateAlarms(alarms)
		if err != nil {
			mid.ReturnError(c, "Accept alert msg fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, "Param validate fail")
	}
}

func GetHistoryAlarm(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnValidateFail(c, "endpoint id validate fail")
		return
	}
	start := c.Query("start")
	end := c.Query("end")
	endpointObj := m.EndpointTable{Id:endpointId}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		mid.ReturnError(c, "get history fail", fmt.Errorf("can't find endpoint with id: %d", endpointId))
		return
	}
	query := m.AlarmTable{Endpoint:endpointObj.Guid}
	if start != "" {
		startTime,err := time.Parse(m.DatetimeFormat, start)
		if err == nil {
			query.Start = startTime
		}else{
			mid.ReturnValidateFail(c, "param start should like "+m.DatetimeFormat)
			return
		}
	}
	if end != "" {
		endTime,err := time.Parse(m.DatetimeFormat, end)
		if err == nil {
			query.End = endTime
		}else{
			mid.ReturnValidateFail(c, "param end should like "+m.DatetimeFormat)
			return
		}
	}
	err,data := db.GetAlarms(query)
	if err != nil {
		mid.ReturnError(c, "Get history fail", err)
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
		mid.ReturnError(c, "get problem alarm fail", err)
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
	if err != nil || id <= 0 {
		mid.ReturnValidateFail(c, "Param id validate fail")
		return
	}
	err = db.CloseAlarm(id)
	if err != nil {
		mid.ReturnError(c, "close alarm fail", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}