package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func QueryAlarmStrategyByGroup(endpointGroup string) (result []*models.GroupStrategyObj, err error) {
	result = []*models.GroupStrategyObj{}
	var alarmStrategyTable []*models.AlarmStrategyTable
	err = x.SQL("select * from alarm_strategy where endpoint_group=?", endpointGroup).Find(&alarmStrategyTable)
	if err != nil {
		return
	}
	for _, v := range alarmStrategyTable {
		tmpStrategyObj := models.GroupStrategyObj{Guid: v.Guid, EndpointGroup: v.EndpointGroup, Metric: v.Metric, Condition: v.Condition, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelaySecond: v.NotifyDelaySecond}
		tmpStrategyObj.NotifyList = getNotifyList(v.Guid, "", "")
		result = append(result, &tmpStrategyObj)
	}
	return
}

func QueryAlarmStrategyByEndpoint(endpoint string) (result []*models.EndpointStrategyObj, err error) {
	result = []*models.EndpointStrategyObj{}
	var endpointGroupTable []*models.EndpointGroupTable
	err = x.SQL("select guid,service_group from endpoint_group where guid in (select endpoint_group from endpoint_group_rel where endpoint=?) or service_group in (select service_group from endpoint_service_rel where endpoint=?)", endpoint, endpoint).Find(&endpointGroupTable)
	if err != nil {
		return
	}
	for _, v := range endpointGroupTable {
		groupStrategy, tmpErr := QueryAlarmStrategyByGroup(v.Guid)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		tmpEndpointStrategyObj := models.EndpointStrategyObj{EndpointGroup: v.Guid, ServiceGroup: v.ServiceGroup, Strategy: groupStrategy}
		result = append(result, &tmpEndpointStrategyObj)
	}
	return
}

func CreateAlarmStrategy(param *models.GroupStrategyObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	param.Guid = guid.CreateGuid()
	var actions []*Action
	insertAction := Action{Sql: "insert into alarm_strategy(guid,endpoint_group,metric,`condition`,`last`,priority,content,notify_enable,notify_delay_second,update_time) value (?,?,?,?,?,?,?,?,?,?)"}
	insertAction.Param = []interface{}{param.Guid, param.EndpointGroup, param.Metric, param.Condition, param.Last, param.Priority, param.Content, param.NotifyEnable, param.NotifyDelaySecond, nowTime}
	actions = append(actions, &insertAction)
	if len(param.NotifyList) > 0 {
		for _, v := range param.NotifyList {
			v.AlarmStrategy = param.Guid
		}
		actions = append(actions, getNotifyListInsertAction(param.NotifyList)...)
	}
	return Transaction(actions)
}

func UpdateAlarmStrategy(param *models.GroupStrategyObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	updateAction := Action{Sql: "update alarm_strategy set metric=?,`condition`=?,`last`=?,priority=?,content=?,notify_enable=?,notify_delay_second=?,update_time=? where guid=?"}
	updateAction.Param = []interface{}{param.Metric, param.Condition, param.Last, param.Priority, param.Content, param.NotifyEnable, param.NotifyDelaySecond, nowTime, param.Guid}
	actions = append(actions, &updateAction)
	for _, v := range param.NotifyList {
		v.AlarmStrategy = param.Guid
	}
	actions = append(actions, getNotifyListDeleteAction(param.Guid, "", "")...)
	actions = append(actions, getNotifyListInsertAction(param.NotifyList)...)
	return Transaction(actions)
}

func DeleteAlarmStrategy(strategyGuid string) error {
	var actions []*Action
	actions = append(actions, getNotifyListDeleteAction(strategyGuid, "", "")...)
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where guid=?", Param: []interface{}{strategyGuid}})
	return Transaction(actions)
}

func getNotifyList(alarmStrategy, endpointGroup, serviceGroup string) (result []*models.NotifyObj) {
	result = []*models.NotifyObj{}
	var notifyTable []*models.NotifyTable
	var refColumn, refValue string
	if alarmStrategy != "" {
		refColumn, refValue = "alarm_strategy", alarmStrategy
	} else if endpointGroup != "" {
		refColumn, refValue = "endpoint_group", endpointGroup
	} else if serviceGroup != "" {
		refColumn, refValue = "service_group", serviceGroup
	}
	x.SQL(fmt.Sprintf("select * from notify where %s=?", refColumn), refValue).Find(&notifyTable)
	for _, v := range notifyTable {
		tmpNotifyObj := models.NotifyObj{Guid: v.Guid, EndpointGroup: v.EndpointGroup, ServiceGroup: v.ServiceGroup, AlarmStrategy: v.AlarmStrategy, AlarmAction: v.AlarmAction, AlarmPriority: v.AlarmPriority, NotifyNum: v.NotifyNum, ProcCallbackName: v.ProcCallbackName, ProcCallbackKey: v.ProcCallbackKey, CallbackUrl: v.CallbackUrl, CallbackParam: v.CallbackParam}
		tmpNotifyObj.NotifyRoles = getNotifyRoles(v.Guid)
		result = append(result, &tmpNotifyObj)
	}
	return result
}

func getNotifyRoles(notifyId string) []string {
	roles := []string{}
	var notifyRoleRel []*models.NotifyRoleRelTable
	x.SQL("select `role` from notify_role_rel where notify=?", notifyId).Find(&notifyRoleRel)
	for _, v := range notifyRoleRel {
		roles = append(roles, v.Role)
	}
	return roles
}

func getNotifyListInsertAction(notifyList []*models.NotifyObj) (actions []*Action) {
	actions = []*Action{}
	if len(notifyList) == 0 {
		return actions
	}
	var refColumn, refValue string
	if notifyList[0].AlarmStrategy != "" {
		refColumn, refValue = "alarm_strategy", notifyList[0].AlarmStrategy
	} else if notifyList[0].EndpointGroup != "" {
		refColumn, refValue = "endpoint_group", notifyList[0].EndpointGroup
	} else if notifyList[0].ServiceGroup != "" {
		refColumn, refValue = "service_group", notifyList[0].ServiceGroup
	} else {
		return actions
	}
	notifyGuidList := guid.CreateGuidList(len(notifyList))
	for i, v := range notifyList {
		if v.NotifyNum == 0 {
			v.NotifyNum = 1
		}
		tmpAction := Action{Sql: fmt.Sprintf("insert into notify(guid,%s,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param) value (?,'%s',?,?,?,?,?,?,?)", refColumn, refValue)}
		tmpAction.Param = []interface{}{notifyGuidList[i], v.AlarmAction, v.AlarmPriority, v.NotifyNum, v.ProcCallbackName, v.ProcCallbackKey, v.CallbackUrl, v.CallbackParam}
		actions = append(actions, &tmpAction)
		if len(v.NotifyRoles) > 0 {
			tmpNotifyRoleGuidList := guid.CreateGuidList(len(v.NotifyRoles))
			for ii, vv := range v.NotifyRoles {
				actions = append(actions, &Action{Sql: "insert into notify_role_rel(guid,notify,`role`) value (?,?,?)", Param: []interface{}{tmpNotifyRoleGuidList[ii], notifyGuidList[i], vv}})
			}
		}
	}
	return actions
}

func getNotifyListDeleteAction(alarmStrategy, endpointGroup, serviceGroup string) (actions []*Action) {
	actions = []*Action{}
	var refColumn string
	var actionParam []interface{}
	if alarmStrategy != "" {
		refColumn = "alarm_strategy"
		actionParam = []interface{}{alarmStrategy}
	} else if endpointGroup != "" {
		refColumn = "endpoint_group"
		actionParam = []interface{}{endpointGroup}
	} else if serviceGroup != "" {
		refColumn = "service_group"
		actionParam = []interface{}{serviceGroup}
	}
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from notify_role_rel where notify in (select guid from notify where %s=?)", refColumn), Param: actionParam})
	actions = append(actions, &Action{Sql: fmt.Sprintf("delete from notify where %s=?", refColumn), Param: actionParam})
	return actions
}

func SyncPrometheusRuleFile(endpointGroup string) {

}

func RemovePrometheusRuleFile(endpointGroup string) {

}
