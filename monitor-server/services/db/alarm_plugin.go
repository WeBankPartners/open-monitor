package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
	"strings"
)

func PluginCloseAlarmAction(input *models.PluginCloseAlarmRequestObj) (result *models.PluginCloseAlarmOutputObj, err error) {
	log.Info(nil, log.LOGGER_APP, "PluginCloseAlarmAction", log.JsonObj("input", input))
	result = &models.PluginCloseAlarmOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", AlarmId: input.AlarmId, Guid: input.Guid}
	// alarmId -> id-firing-notifyGuid
	alarmSplit := strings.Split(input.AlarmId, "-")
	alarmId, _ := strconv.Atoi(alarmSplit[0])
	if alarmId <= 0 {
		err = fmt.Errorf("AlarmId:%s illegal ", input.AlarmId)
		return
	}
	if len(alarmSplit) == 3 {
		if alarmSplit[2] == "custom_alarm_guid" {
			// 直接使用事务处理：写入历史表并删除原记录
			var actions []*Action
			// 先查询记录并写入 history_alarm_custom 表
			actions = append(actions, &Action{Sql: "INSERT INTO history_alarm_custom(alert_info,alert_ip,alert_level,alert_obj,alert_title,alert_reciver,remark_info,sub_system_id," +
				"use_umg_policy,alert_way,custom_message,alarm_total,create_at,close_user) " +
				"SELECT alert_info,alert_ip,alert_level,alert_obj,alert_title,alert_reciver,remark_info,sub_system_id," +
				"use_umg_policy,alert_way,?,alarm_total,create_at,'system' FROM alarm_custom WHERE id=?", Param: []interface{}{input.Message, alarmId}})

			// 删除 alarm_custom 记录
			actions = append(actions, &Action{Sql: "DELETE FROM alarm_custom WHERE id=?", Param: []interface{}{alarmId}})

			// 执行事务
			err = Transaction(actions)
			return
		}
	}
	queryRows, _ := x.QueryString("SELECT id,status,endpoint_tags FROM alarm WHERE id=?", alarmId)
	if len(queryRows) == 0 {
		err = fmt.Errorf("Can not find alarm with id:%s ", input.AlarmId)
		return
	}
	if queryRows[0]["status"] == "closed" {
		return
	}
	var actions []*Action
	actions = append(actions, &Action{Sql: "UPDATE alarm SET status='closed',close_msg=?,custom_message=?,close_user='system',end=NOW() WHERE id=?", Param: []interface{}{input.Message, input.Message, alarmId}})
	//_, err = x.Exec("UPDATE alarm SET status='closed',close_msg=?,custom_message=?,close_user='system',end=NOW() WHERE id=?", input.Message, input.Message, alarmId)
	endpointTags := queryRows[0]["endpoint_tags"]
	if strings.HasPrefix(endpointTags, "ac_") {
		actions = append(actions, &Action{Sql: "UPDATE alarm_condition SET STATUS='closed',end=NOW() WHERE guid in (select alarm_condition from alarm_condition_rel where alarm=?)", Param: []interface{}{alarmId}})
	}
	err = Transaction(actions)
	return
}
