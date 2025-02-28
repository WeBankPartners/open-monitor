package db

import (
	"crypto/sha256"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
)

func doInsertOrUpdateAlarm(alarmObj *models.AlarmHandleObj) (err error) {
	session := x.NewSession()
	session.Begin()
	defer func() {
		if err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
		session.Close()
	}()
	if alarmObj.Id > 0 {
		if alarmObj.Status != "firing" {
			updateExecResult, updateExecErr := session.Exec("UPDATE alarm SET status=?,end_value=?,end=? WHERE id=? AND status='firing'",
				alarmObj.Status, alarmObj.EndValue, alarmObj.End.Format(models.DatetimeFormat), alarmObj.Id)
			if updateExecErr != nil {
				err = fmt.Errorf("update alarm data fail,%s ", updateExecErr.Error())
				return
			}
			if rowAffectNum, _ := updateExecResult.RowsAffected(); rowAffectNum > 0 {
				_, deleteAlarmFiringErr := session.Exec("delete from alarm_firing where alarm_id=?", alarmObj.Id)
				if deleteAlarmFiringErr != nil {
					err = fmt.Errorf("remove alarm firing data fail,%s ", deleteAlarmFiringErr.Error())
					return
				}
			}
		} else {
			log.Warn(nil, log.LOGGER_APP, "doInsertOrUpdateAlarm ignore with illegal status", zap.Int("alarmId", alarmObj.Id), zap.String("status", alarmObj.Status))
		}
		return
	}
	firingUniqueHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s_%s", alarmObj.Endpoint, alarmObj.SMetric, alarmObj.Tags, alarmObj.AlarmStrategy))))
	insertAlarmFiringResult, insertFiringErr := session.Exec("insert into alarm_firing(endpoint,metric,tags,alarm_name,alarm_strategy,expr,cond,`last`,priority,content,start_value,`start`,unique_hash) values (?,?,?,?,?,?,?,?,?,?,?,?,?)",
		alarmObj.Endpoint, alarmObj.SMetric, alarmObj.Tags, alarmObj.AlarmName, alarmObj.AlarmStrategy, alarmObj.SExpr, alarmObj.SCond, alarmObj.SLast, alarmObj.SPriority, alarmObj.Content, alarmObj.StartValue, alarmObj.Start.Format(models.DatetimeFormat), firingUniqueHash)
	if insertFiringErr != nil {
		err = fmt.Errorf("insert alarm firing data fail,%s ", insertFiringErr.Error())
		return
	}
	alarmFiringId, _ := insertAlarmFiringResult.LastInsertId()
	if alarmFiringId <= 0 {
		err = fmt.Errorf("insert alarm firing get new id fail")
		return
	}
	alarmUniqueHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s_%s_%d", alarmObj.Endpoint, alarmObj.SMetric, alarmObj.Tags, alarmObj.AlarmStrategy, alarmObj.Start.Unix()))))
	insertAlarmResult, insertErr := session.Exec("INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start,tags,endpoint_tags,alarm_strategy,alarm_name) VALUE (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		alarmObj.StrategyId, alarmObj.Endpoint, alarmObj.Status, alarmObj.SMetric, alarmObj.SExpr, alarmObj.SCond, alarmObj.SLast, alarmObj.SPriority, alarmObj.Content, alarmObj.StartValue, alarmObj.Start.Format(models.DatetimeFormat), alarmObj.Tags, alarmUniqueHash, alarmObj.AlarmStrategy, alarmObj.AlarmName)
	if insertErr != nil {
		err = fmt.Errorf("insert alarm data fail,%s ", insertErr.Error())
		return
	}
	alarmId, _ := insertAlarmResult.LastInsertId()
	if alarmId <= 0 {
		err = fmt.Errorf("insert alarm get new id fail")
		return
	}
	alarmObj.Id = int(alarmId)
	updateAlarmIdResult, updateErr := session.Exec("update alarm_firing set alarm_id=? where id=?", alarmId, alarmFiringId)
	if updateErr != nil {
		err = fmt.Errorf("update alarm firing with alarmId fail,%s ", updateErr.Error())
		return
	}
	if rowAffectNum, _ := updateAlarmIdResult.RowsAffected(); rowAffectNum <= 0 {
		err = fmt.Errorf("update alarm firing fail with no row affect")
		return
	}
	return
}
