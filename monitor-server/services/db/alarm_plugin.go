package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
	"strings"
)

func PluginCloseAlarmAction(input *models.PluginCloseAlarmRequestObj) (result *models.PluginCloseAlarmOutputObj, err error)  {
	log.Logger.Info("PluginCloseAlarmAction", log.JsonObj("input", input))
	result = &models.PluginCloseAlarmOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", AlarmId: input.AlarmId}
	// alarmId -> id-firing-notifyGuid
	alarmSplit := strings.Split(input.AlarmId, "-")
	alarmId,_ := strconv.Atoi(alarmSplit[0])
	if alarmId <= 0 {
		err = fmt.Errorf("AlarmId:%s illegal ", input.AlarmId)
		return
	}
	if len(alarmSplit) == 3 {
		if alarmSplit[2] == "custom_alarm_guid" {
			queryRows,_ := x.QueryString("SELECT id FROM alarm_custom WHERE id=?", alarmId)
			if len(queryRows) == 0 {
				err = fmt.Errorf("Can not find custom alarm with id:%s ", input.AlarmId)
				return
			}
			_, err = x.Exec("UPDATE alarm_custom SET closed=1,custom_message=?,closed_at=NOW() WHERE id=?", input.Message, alarmId)
			return
		}
	}
	queryRows,_ := x.QueryString("SELECT id FROM alarm WHERE id=?", alarmId)
	if len(queryRows) == 0 {
		err = fmt.Errorf("Can not find alarm with id:%s ", input.AlarmId)
		return
	}
	_, err = x.Exec("UPDATE alarm SET status='closed',close_msg=?,custom_message=?,close_user='system',end=NOW() WHERE id=?", input.Message, input.Message, alarmId)
	return
}
