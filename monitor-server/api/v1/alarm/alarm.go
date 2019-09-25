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
			tmpAlarm := m.AlarmTable{Status:v.Status}
			tmpAlarm.StrategyId,_ = strconv.Atoi(v.Labels["strategy_id"])
			if tmpAlarm.StrategyId <= 0 {
				mid.LogInfo(fmt.Sprintf("Alerts strategy id is null : %v ", v))
				continue
			}
			_,strategyObj := db.GetStrategy(tmpAlarm.StrategyId)
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
				if tmpAlarms[0].Status != "ok" {
					if v.Status == "firing" {
						tmpOperation = "same"
					}else{
						tmpOperation = "resolve"
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

func GetHistoryAlarm()  {
	
}

func GetProblemAlarm()  {
	
}
