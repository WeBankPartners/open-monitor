package db

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/common/smtp"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
	"golang.org/x/net/context/ctxhttp"
)

// 系统内置 指标阈值-组 ,& old_1~old_20
var systemAlarmStrategyIds = []string{"new_host_ping_loss", "new_ping_ping_loss", "old_process_group"}

func QueryAlarmStrategyByGroup(endpointGroup, alarmName, show, operator string) (result []*models.EndpointStrategyObj, err error) {
	result = []*models.EndpointStrategyObj{}
	var strategy []*models.GroupStrategyObj
	var alarmStrategyTable []*models.AlarmStrategyMetricObj
	var baseSql = "select t1.*,t2.metric as 'metric_name' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where t1.endpoint_group=? "
	var params []interface{}
	params = append(params, endpointGroup)
	if strings.TrimSpace(alarmName) != "" {
		baseSql = baseSql + " and t1.name like '%" + alarmName + "%'"
	}
	if show == "me" {
		baseSql = baseSql + " and t1.log_metric_group is null"
	}
	baseSql = baseSql + " order by t1.update_time desc"
	err = x.SQL(baseSql, params...).Find(&alarmStrategyTable)
	if err != nil {
		return
	}
	for _, v := range alarmStrategyTable {
		tmpStrategyObj := models.GroupStrategyObj{Guid: v.Guid, Name: v.Name, EndpointGroup: v.EndpointGroup, Metric: v.Metric, MetricName: v.MetricName, Condition: v.Condition, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelaySecond: v.NotifyDelaySecond, ActiveWindow: v.ActiveWindow}
		tmpStrategyObj.ActiveWindowList = strings.Split(tmpStrategyObj.ActiveWindow, ",")
		tmpStrategyObj.UpdateTime = v.UpdateTime
		tmpStrategyObj.UpdateUser = v.UpdateUser
		if strings.TrimSpace(v.LogMetricGroup) != "" {
			tmpStrategyObj.LogMetricGroup = &v.LogMetricGroup
		}
		tmpStrategyObj.NotifyList = getNotifyList(v.Guid, "", "")
		if tmpStrategyConditions, tmpErr := getStrategyConditions(v.Guid); tmpErr != nil {
			err = tmpErr
			return
		} else {
			if len(tmpStrategyConditions) == 0 {
				tmpStrategyConditions = append(tmpStrategyConditions, &models.StrategyConditionObj{
					Metric:     v.Metric,
					MetricName: v.MetricName,
					Condition:  v.Condition,
					Last:       v.Last,
					Tags:       []*models.MetricTag{},
				})
			}
			tmpStrategyObj.Conditions = tmpStrategyConditions
		}
		strategy = append(strategy, &tmpStrategyObj)
	}
	resultObj := models.EndpointStrategyObj{EndpointGroup: endpointGroup, Strategy: strategy, DisplayName: endpointGroup}
	notify, tmpErr := GetGroupEndpointNotify(endpointGroup)
	if tmpErr != nil {
		return result, tmpErr
	}
	resultObj.NotifyList = notify
	result = append(result, &resultObj)
	return
}

func QueryAlarmStrategyByEndpoint(endpoint, alarmName, show, operator string) (result []*models.EndpointStrategyObj, err error) {
	endpointObj, getErr := GetEndpointNew(&models.EndpointNewTable{Guid: endpoint})
	if getErr != nil {
		return result, getErr
	}
	result = []*models.EndpointStrategyObj{}
	var endpointGroupTable []*models.EndpointGroupTable
	err = x.SQL("select guid,service_group from endpoint_group where monitor_type=? and (guid in (select endpoint_group from endpoint_group_rel where endpoint=?) or service_group in (select service_group from endpoint_service_rel where endpoint=?))", endpointObj.MonitorType, endpoint, endpoint).Find(&endpointGroupTable)
	if err != nil {
		return
	}
	for _, v := range endpointGroupTable {
		tmpEndpointStrategyList, tmpErr := QueryAlarmStrategyByGroup(v.Guid, alarmName, show, operator)
		if tmpErr != nil || len(tmpEndpointStrategyList) == 0 {
			err = tmpErr
			break
		}
		if v.ServiceGroup != "" {
			tmpEndpointStrategyList[0].ServiceGroup = v.ServiceGroup
			tmpEndpointStrategyList[0].DisplayName = models.GlobalSGDisplayNameMap[v.ServiceGroup]
		}
		result = append(result, tmpEndpointStrategyList[0])
	}
	return
}

func QueryAlarmStrategyByServiceGroup(serviceGroup, alarmName, show, operator string) (result []*models.EndpointStrategyObj, err error) {
	result = []*models.EndpointStrategyObj{}
	var endpointGroupTable []*models.EndpointGroupTable
	err = x.SQL("select guid,monitor_type,service_group from endpoint_group where service_group=?", serviceGroup).Find(&endpointGroupTable)
	if err != nil {
		return
	}
	for _, v := range endpointGroupTable {
		tmpEndpointStrategyList, tmpErr := QueryAlarmStrategyByGroup(v.Guid, alarmName, show, operator)
		if tmpErr != nil || len(tmpEndpointStrategyList) == 0 {
			err = tmpErr
			break
		}
		tmpEndpointStrategyList[0].ServiceGroup = v.ServiceGroup
		tmpEndpointStrategyList[0].MonitorType = v.MonitorType
		tmpEndpointStrategyList[0].DisplayName = v.MonitorType
		result = append(result, tmpEndpointStrategyList[0])
	}
	return
}

func GetAlarmStrategy(strategyGuid, conditionCrc string) (result models.AlarmStrategyMetricObj, conditions []*models.AlarmStrategyMetricWithExpr, err error) {
	var strategyTable []*models.AlarmStrategyMetricObj
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where t1.guid=?", strategyGuid).Find(&strategyTable)
	if err != nil {
		err = fmt.Errorf("Query alarm_strategy fail,%s ", err.Error())
		return
	}
	if len(strategyTable) == 0 {
		err = fmt.Errorf("Can not find alarm_strategy with guid:%s ", strategyGuid)
		return
	}
	result = *strategyTable[0]
	conditions = []*models.AlarmStrategyMetricWithExpr{}
	err = x.SQL("select t1.guid,t1.alarm_strategy,t1.metric,t1.`condition`,t1.`last`,t1.crc_hash,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy_metric t1 left join metric t2 on t1.metric=t2.guid where t1.alarm_strategy=?", strategyGuid).Find(&conditions)
	if err != nil {
		err = fmt.Errorf("Query alarm strategy metric table fail,%s ", err.Error())
		return
	}
	if conditionCrc != "" && len(conditions) > 0 {
		for _, conditionRow := range conditions {
			if conditionRow.CrcHash == conditionCrc {
				result.ConditionCrc = conditionRow.CrcHash
				result.Metric = conditionRow.Metric
				result.MetricName = conditionRow.MetricName
				result.MetricExpr = conditionRow.MetricExpr
				result.Condition = conditionRow.Condition
				result.Last = conditionRow.Last
				break
			}
		}
		if result.ConditionCrc == "" {
			err = fmt.Errorf("can not find condition crc:%s in alarmStrategy:%s ", conditionCrc, strategyGuid)
		}
	}
	return
}

func CreateAlarmStrategy(param *models.GroupStrategyObj, operator string) error {
	var err error
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	if actions, err = getCreateAlarmStrategyActions(param, nowTime, operator); err != nil {
		return err
	}
	return Transaction(actions)
}

func getCreateAlarmStrategyActions(param *models.GroupStrategyObj, nowTime, operator string) (actions []*Action, err error) {
	param.Guid = "strategy_" + guid.CreateGuid()
	var insertAction = Action{Sql: "insert into alarm_strategy(guid,name,endpoint_group,metric,`condition`,`last`,priority,content,notify_enable,notify_delay_second,active_window,update_time,create_user,update_user,log_metric_group) value (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"}
	insertAction.Param = []interface{}{param.Guid, param.Name, param.EndpointGroup, param.Metric, param.Condition, param.Last, param.Priority, param.Content, param.NotifyEnable, param.NotifyDelaySecond, param.ActiveWindow, nowTime, operator, operator, param.LogMetricGroup}
	actions = append(actions, &insertAction)
	if len(param.NotifyList) > 0 {
		for _, v := range param.NotifyList {
			v.AlarmStrategy = param.Guid
		}
		actions = append(actions, getNotifyListInsertAction(param.NotifyList)...)
	}
	if len(param.Conditions) > 0 {
		for _, condition := range param.Conditions {
			// 创建时候 metric_name web传递可能为空,从metric截取下
			if strings.TrimSpace(condition.MetricName) == "" && len(condition.Metric) > 0 {
				condition.MetricName = strings.Split(condition.Metric, "__")[0]
			}
			if condition.LogType == "" {
				if logType, err2 := GetLogTypeByMetric(condition.MetricName); err != nil {
					log.Error(nil, log.LOGGER_APP, "GetLogTypeByMetric err", zap.Error(err2))
				} else {
					condition.LogType = logType
				}
			}
		}
		insertConditionActions, buildActionErr := getStrategyConditionInsertAction(param.Guid, param.Conditions)
		if buildActionErr != nil {
			err = buildActionErr
			return
		}
		actions = append(actions, insertConditionActions...)
	}
	return
}

func ValidateAlarmStrategyName(param *models.GroupStrategyObj) (err error) {
	queryResult, queryErr := x.QueryString("select guid from alarm_strategy where endpoint_group=? and name=? and guid!=?", param.EndpointGroup, param.Name, param.Guid)
	if queryErr != nil {
		err = fmt.Errorf("query alarm strategy table fail,%s ", queryErr.Error())
		return
	}
	if len(queryResult) > 0 {
		err = fmt.Errorf("alarm strategy name:%s duplicate", param.Name)
	}
	return
}

func UpdateAlarmStrategy(param *models.GroupStrategyObj, operator string) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	var updateConditionActions, actions []*Action
	var err error
	updateAction := Action{Sql: "update alarm_strategy set name=?,priority=?,content=?,notify_enable=?,notify_delay_second=?,active_window=?,update_time=?,update_user=? where guid=?"}
	updateAction.Param = []interface{}{param.Name, param.Priority, param.Content, param.NotifyEnable, param.NotifyDelaySecond, param.ActiveWindow, nowTime, operator, param.Guid}
	actions = append(actions, &updateAction)
	for _, v := range param.NotifyList {
		v.AlarmStrategy = param.Guid
	}
	actions = append(actions, getNotifyListUpdateAction(param.NotifyList)...)
	updateConditionActions, err = getStrategyConditionUpdateAction(param.Guid, param.Conditions)
	if err != nil {
		return err
	}
	actions = append(actions, updateConditionActions...)
	return Transaction(actions)
}

func DeleteAlarmStrategy(strategyGuid string) (endpointGroup string, err error) {
	var delAlarmStrategyActions []*Action
	if delAlarmStrategyActions, endpointGroup, err = GetDeleteAlarmStrategyActions(strategyGuid); err != nil {
		return
	}
	err = Transaction(delAlarmStrategyActions)
	return
}

func GetDeleteAlarmStrategyActions(strategyGuid string) (delAlarmStrategyActions []*Action, endpointGroup string, err error) {
	delAlarmStrategyActions = []*Action{}
	var strategyTable []*models.AlarmStrategyTable
	err = x.SQL("select * from alarm_strategy where guid=?", strategyGuid).Find(&strategyTable)
	if err != nil {
		return
	}
	if len(strategyTable) == 0 {
		err = fmt.Errorf("Can not find strategy with guid:%s ", strategyGuid)
		return
	}
	endpointGroup = strategyTable[0].EndpointGroup
	delAlarmStrategyActions = append(delAlarmStrategyActions, getNotifyListDeleteAction(strategyGuid, "", "")...)
	delAlarmStrategyActions = append(delAlarmStrategyActions, getStrategyConditionDeleteAction(strategyGuid)...)
	delAlarmStrategyActions = append(delAlarmStrategyActions, &Action{Sql: "delete from alarm_strategy where guid=?", Param: []interface{}{strategyGuid}})
	return
}

func getNotifyList(alarmStrategy, endpointGroup, serviceGroup string) (result []*models.NotifyObj) {
	result = []*models.NotifyObj{}
	var notifyTable []*models.NotifyTable
	var refColumn, refValue string
	firingNotify, okNotify := &models.NotifyTable{AlarmAction: "firing", NotifyNum: 1}, &models.NotifyTable{AlarmAction: "ok", NotifyNum: 1}
	if alarmStrategy != "" {
		refColumn, refValue = "alarm_strategy", alarmStrategy
		firingNotify.AlarmStrategy = alarmStrategy
		okNotify.AlarmStrategy = alarmStrategy
	} else if endpointGroup != "" {
		refColumn, refValue = "endpoint_group", endpointGroup
	} else if serviceGroup != "" {
		refColumn, refValue = "service_group", serviceGroup
	}
	x.SQL(fmt.Sprintf("select * from notify where %s=?", refColumn), refValue).Find(&notifyTable)
	for _, v := range notifyTable {
		if v.AlarmAction == "firing" {
			firingNotify = v
		} else if v.AlarmAction == "ok" {
			okNotify = v
		}
	}
	result = append(result, &models.NotifyObj{Guid: firingNotify.Guid, NotifyRoles: getNotifyRoles(firingNotify.Guid), EndpointGroup: firingNotify.EndpointGroup, ServiceGroup: firingNotify.ServiceGroup, AlarmStrategy: firingNotify.AlarmStrategy, AlarmAction: firingNotify.AlarmAction, AlarmPriority: firingNotify.AlarmPriority, NotifyNum: firingNotify.NotifyNum, ProcCallbackName: firingNotify.ProcCallbackName, ProcCallbackKey: firingNotify.ProcCallbackKey, CallbackUrl: firingNotify.CallbackUrl, CallbackParam: firingNotify.CallbackParam, ProcCallbackMode: firingNotify.ProcCallbackMode, Description: firingNotify.Description})
	result = append(result, &models.NotifyObj{Guid: okNotify.Guid, NotifyRoles: getNotifyRoles(okNotify.Guid), EndpointGroup: okNotify.EndpointGroup, ServiceGroup: okNotify.ServiceGroup, AlarmStrategy: okNotify.AlarmStrategy, AlarmAction: okNotify.AlarmAction, AlarmPriority: okNotify.AlarmPriority, NotifyNum: okNotify.NotifyNum, ProcCallbackName: okNotify.ProcCallbackName, ProcCallbackKey: okNotify.ProcCallbackKey, CallbackUrl: okNotify.CallbackUrl, CallbackParam: okNotify.CallbackParam, ProcCallbackMode: okNotify.ProcCallbackMode, Description: okNotify.Description})
	return result
}

func getSimpleNotify(notifyGuid string) (result models.NotifyTable, err error) {
	var notifyTable []*models.NotifyTable
	err = x.SQL("select * from notify where guid=?", notifyGuid).Find(&notifyTable)
	if err != nil {
		return result, fmt.Errorf("Query notify table tail,%s ", err.Error())
	}
	if len(notifyTable) == 0 {
		return result, fmt.Errorf("Can not find notify with guid:%s ", notifyGuid)
	}
	result = *notifyTable[0]
	return
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
	}
	notifyGuidList := guid.CreateGuidList(len(notifyList))
	for i, v := range notifyList {
		if v.NotifyNum == 0 {
			v.NotifyNum = 1
		}
		v.Guid = "notify_" + notifyGuidList[i]
		tmpAction := Action{}
		if refColumn != "" {
			tmpAction = Action{Sql: fmt.Sprintf("insert into notify(guid,%s,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param,proc_callback_mode,description) value (?,'%s',?,?,?,?,?,?,?,?,?)", refColumn, refValue)}
		} else {
			tmpAction = Action{Sql: "insert into notify(guid,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param,proc_callback_mode,description) value (?,?,?,?,?,?,?,?,?,?)"}
		}
		tmpAction.Param = []interface{}{v.Guid, v.AlarmAction, v.AlarmPriority, v.NotifyNum, v.ProcCallbackName, v.ProcCallbackKey, v.CallbackUrl, v.CallbackParam, v.ProcCallbackMode, v.Description}
		actions = append(actions, &tmpAction)
		if len(v.NotifyRoles) > 0 {
			tmpNotifyRoleGuidList := guid.CreateGuidList(len(v.NotifyRoles))
			for ii, vv := range v.NotifyRoles {
				actions = append(actions, &Action{Sql: "insert into notify_role_rel(guid,notify,`role`) value (?,?,?)", Param: []interface{}{tmpNotifyRoleGuidList[ii], v.Guid, vv}})
			}
		}
	}
	return actions
}

func getNotifyListUpdateAction(notifyList []*models.NotifyObj) (actions []*Action) {
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
	}
	notifyGuidList := guid.CreateGuidList(len(notifyList))
	for i, v := range notifyList {
		if v.NotifyNum == 0 {
			v.NotifyNum = 1
		}
		if v.Guid != "" {
			tmpAction := Action{Sql: fmt.Sprintf("update notify set alarm_action=?,notify_num=?,proc_callback_name=?,proc_callback_key=?,callback_url=?,callback_param=?,proc_callback_mode=?,description=? where guid=?")}
			tmpAction.Param = []interface{}{v.AlarmAction, v.NotifyNum, v.ProcCallbackName, v.ProcCallbackKey, v.CallbackUrl, v.CallbackParam, v.ProcCallbackMode, v.Description, v.Guid}
			actions = append(actions, &tmpAction)
			actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify=?", Param: []interface{}{v.Guid}})
		} else {
			v.Guid = "notify_" + notifyGuidList[i]
			tmpAction := Action{}
			if refColumn != "" {
				tmpAction = Action{Sql: fmt.Sprintf("insert into notify(guid,%s,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param,proc_callback_mode,description) value (?,'%s',?,?,?,?,?,?,?,?,?)", refColumn, refValue)}
			} else {
				tmpAction = Action{Sql: "insert into notify(guid,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param,proc_callback_mode,description) value (?,?,?,?,?,?,?,?,?,?)"}
			}
			tmpAction.Param = []interface{}{v.Guid, v.AlarmAction, v.AlarmPriority, v.NotifyNum, v.ProcCallbackName, v.ProcCallbackKey, v.CallbackUrl, v.CallbackParam, v.ProcCallbackMode, v.Description}
			actions = append(actions, &tmpAction)
		}
		if len(v.NotifyRoles) > 0 {
			tmpNotifyRoleGuidList := guid.CreateGuidList(len(v.NotifyRoles))
			for ii, vv := range v.NotifyRoles {
				actions = append(actions, &Action{Sql: "insert into notify_role_rel(guid,notify,`role`) value (?,?,?)", Param: []interface{}{tmpNotifyRoleGuidList[ii], v.Guid, vv}})
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

func getNotifyDeleteAction(notifyGuid string) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify=?", Param: []interface{}{notifyGuid}})
	actions = append(actions, &Action{Sql: "delete from notify where guid=?", Param: []interface{}{notifyGuid}})
	return actions
}

func getStrategyConditions(alarmStrategyGuid string) (conditions []*models.StrategyConditionObj, err error) {
	conditions = []*models.StrategyConditionObj{}
	var strategyMetricRows []*models.AlarmStrategyMetricQueryRow
	err = x.SQL("select t1.guid,t1.alarm_strategy,t1.metric,t1.`condition`,t1.`last`,t2.metric as `metric_name` from alarm_strategy_metric t1 left join metric t2 on t1.metric=t2.guid where t1.alarm_strategy=?", alarmStrategyGuid).Find(&strategyMetricRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy metric with strategyGuid:%s fail,%s ", alarmStrategyGuid, err.Error())
		return
	}
	if len(strategyMetricRows) == 0 {
		return
	}
	var strategyTagRows []*models.AlarmStrategyTag
	err = x.SQL("select * from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?)", alarmStrategyGuid).Find(&strategyTagRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy tag with strategyGuid:%s fail,%s ", alarmStrategyGuid, err.Error())
		return
	}
	var tagValueRows []*models.AlarmStrategyTagValue
	err = x.SQL("select * from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?))", alarmStrategyGuid).Find(&tagValueRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy tag value with strategyGuid:%s fail,%s ", alarmStrategyGuid, err.Error())
		return
	}
	tagValueMap := make(map[string][]string)
	for _, v := range tagValueRows {
		if existList, ok := tagValueMap[v.AlarmStrategyTag]; ok {
			tagValueMap[v.AlarmStrategyTag] = append(existList, v.Value)
		} else {
			tagValueMap[v.AlarmStrategyTag] = []string{v.Value}
		}
	}
	for _, metricRow := range strategyMetricRows {
		conditionRow := models.StrategyConditionObj{Metric: metricRow.Metric, Condition: metricRow.Condition, Last: metricRow.Last, Tags: []*models.MetricTag{}, MetricName: metricRow.MetricName}
		for _, tagRow := range strategyTagRows {
			if tagRow.AlarmStrategyMetric == metricRow.Guid {
				tmpTag := models.MetricTag{TagName: tagRow.Name, TagValue: []string{}, Equal: tagRow.Equal}
				if existList, ok := tagValueMap[tagRow.Guid]; ok {
					tmpTag.TagValue = existList
				}
				conditionRow.Tags = append(conditionRow.Tags, &tmpTag)
			}
		}
		conditions = append(conditions, &conditionRow)
	}
	return
}

func getStrategyConditionInsertAction(alarmStrategyGuid string, conditions []*models.StrategyConditionObj) (actions []*Action, err error) {
	nowTime := time.Now()
	metricGuidList := guid.CreateGuidList(len(conditions))
	existCrcMap := make(map[string]int)
	monitorEngineMetricMap := make(map[string]int)
	if monitorEngineMetricMap, err = GetMonitorEngineMetricMap(); err != nil {
		return
	}
	for i, metricRow := range conditions {
		tmpMetricRowString, _ := json.Marshal(metricRow)
		tmpCrcHash := fmt.Sprintf("%d", crc64.Checksum(tmpMetricRowString, crc64.MakeTable(crc64.ECMA)))
		if _, existFlag := existCrcMap[tmpCrcHash]; existFlag {
			err = fmt.Errorf("metric condition is duplicated")
			return
		}
		existCrcMap[tmpCrcHash] = 1
		monitorEngineFlag := 0
		if _, ok := monitorEngineMetricMap[metricRow.Metric]; ok {
			monitorEngineFlag = 1
		}
		actions = append(actions, &Action{Sql: "insert into alarm_strategy_metric(guid,alarm_strategy,metric,`condition`,`last`,create_time,crc_hash,monitor_engine,log_type) values (?,?,?,?,?,?,?,?,?)", Param: []interface{}{
			metricGuidList[i], alarmStrategyGuid, metricRow.Metric, metricRow.Condition, metricRow.Last, nowTime, tmpCrcHash, monitorEngineFlag, metricRow.LogType,
		}})
		if len(metricRow.Tags) > 0 {
			tagGuidList := guid.CreateGuidList(len(metricRow.Tags))
			for tagIndex, tagRow := range metricRow.Tags {
				actions = append(actions, &Action{Sql: "insert into alarm_strategy_tag(guid,alarm_strategy_metric,name,equal) values (?,?,?,?)", Param: []interface{}{tagGuidList[tagIndex], metricGuidList[i], tagRow.TagName, tagRow.Equal}})
				for _, tagValue := range tagRow.TagValue {
					actions = append(actions, &Action{Sql: "insert into alarm_strategy_tag_value(alarm_strategy_tag,value) values (?,?)", Param: []interface{}{tagGuidList[tagIndex], tagValue}})
				}
			}
		}
	}
	return
}

func getStrategyConditionUpdateAction(alarmStrategyGuid string, conditions []*models.StrategyConditionObj) (actions []*Action, err error) {
	actions = append(actions, getStrategyConditionDeleteAction(alarmStrategyGuid)...)
	for _, condition := range conditions {
		if condition.LogType == "" {
			if logType, err2 := GetLogTypeByMetric(condition.MetricName); err != nil {
				log.Error(nil, log.LOGGER_APP, "GetLogTypeByMetric err", zap.Error(err2))
			} else {
				condition.LogType = logType
			}
		}
	}
	insertConditionActions, getErr := getStrategyConditionInsertAction(alarmStrategyGuid, conditions)
	if getErr != nil {
		err = getErr
		return
	}
	actions = append(actions, insertConditionActions...)
	return
}

func getStrategyConditionDeleteAction(alarmStrategyGuid string) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?))", Param: []interface{}{alarmStrategyGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy=?)", Param: []interface{}{alarmStrategyGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy_metric where alarm_strategy=?", Param: []interface{}{alarmStrategyGuid}})
	return
}

func SyncPrometheusRuleFile(endpointGroup string, withoutReloadConfig bool) error {
	if endpointGroup == "" {
		return fmt.Errorf("Sync prometheus rule fail,group is empty ")
	}
	endpointGroupObj, err := GetSimpleEndpointGroup(endpointGroup)
	if err != nil {
		return fmt.Errorf("Sync prometheus rule fail,%s ", err.Error())
	}
	log.Info(nil, log.LOGGER_APP, "SyncPrometheusRuleFile", zap.String("endpointGroup", endpointGroup))
	ruleFileName := "g_" + endpointGroup
	var endpointList []*models.EndpointNewTable
	if endpointGroupObj.ServiceGroup == "" {
		err = x.SQL("select * from endpoint_new where monitor_type=? and guid in (select endpoint from endpoint_group_rel where endpoint_group=?)", endpointGroupObj.MonitorType, endpointGroup).Find(&endpointList)
	} else {
		serviceGroupGuidList, _ := fetchGlobalServiceGroupChildGuidList(endpointGroupObj.ServiceGroup)
		err = x.SQL("select * from endpoint_new where monitor_type=? and guid in (select endpoint from endpoint_service_rel where service_group in ('"+strings.Join(serviceGroupGuidList, "','")+"'))", endpointGroupObj.MonitorType).Find(&endpointList)
	}
	if err != nil {
		return err
	}
	// 获取strategy
	strategyList, monitorEngineStrategyList, getStrategyErr := getAlarmStrategyWithExprNew(endpointGroup)
	if getStrategyErr != nil {
		return getStrategyErr
	}
	log.Debug(nil, log.LOGGER_APP, "SyncPrometheusRuleFile alarm strategy data", log.JsonObj("strategyList", strategyList))
	// 区分cluster，分别下发
	var clusterList []string
	var clusterEndpointMap = make(map[string][]*models.EndpointNewTable)
	if len(endpointList) > 0 {
		for _, endpoint := range endpointList {
			if _, b := clusterEndpointMap[endpoint.Cluster]; !b {
				clusterList = append(clusterList, endpoint.Cluster)
				clusterEndpointMap[endpoint.Cluster] = []*models.EndpointNewTable{endpoint}
			} else {
				clusterEndpointMap[endpoint.Cluster] = append(clusterEndpointMap[endpoint.Cluster], endpoint)
			}
		}
	} else {
		var clusterTable []*models.ClusterTable
		x.SQL("select id from cluster").Find(&clusterTable)
		for _, tmpCluster := range clusterTable {
			clusterList = append(clusterList, tmpCluster.Id)
			clusterEndpointMap[tmpCluster.Id] = []*models.EndpointNewTable{}
		}
	}
	for _, cluster := range clusterList {
		guidExpr, addressExpr, ipExpr := buildRuleReplaceExprNew(clusterEndpointMap[cluster])
		ruleFileConfig := buildRuleFileContentNew(ruleFileName, guidExpr, addressExpr, ipExpr, copyStrategyListNew(strategyList))
		if cluster == "default" || cluster == "" {
			prom.SyncLocalRuleConfig(models.RuleLocalConfigJob{WithoutReloadConfig: withoutReloadConfig, EndpointGroup: endpointGroup, Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
		} else {
			tmpErr := SyncRemoteRuleConfigFile(cluster, models.RFClusterRequestObj{Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
			if tmpErr != nil {
				err = fmt.Errorf("Update remote cluster:%s rule file fail,%s ", cluster, tmpErr.Error())
				log.Error(nil, log.LOGGER_APP, "Update remote cluster rule file fail", zap.String("cluster", cluster), zap.Error(tmpErr))
			}
		}
		for _, monitorEngineStrategy := range monitorEngineStrategyList {
			buildStrategyAlarmRuleExpr(guidExpr, addressExpr, ipExpr, monitorEngineStrategy)
			UpdateAlarmStrategyMetricExpr(monitorEngineStrategy)
		}
	}
	return err
}

func RemovePrometheusRuleFile(endpointGroup string, fromPeer bool) {
	ruleFileName := "g_" + endpointGroup
	var clusterTable []*models.ClusterTable
	x.SQL("select id from cluster").Find(&clusterTable)
	for _, cluster := range clusterTable {
		if cluster.Id == "default" || cluster.Id == "" {
			prom.SyncLocalRuleConfig(models.RuleLocalConfigJob{FromPeer: fromPeer, EndpointGroup: endpointGroup, Name: ruleFileName, Rules: []*models.RFRule{}})
		} else {
			tmpErr := SyncRemoteRuleConfigFile(cluster.Id, models.RFClusterRequestObj{Name: ruleFileName, Rules: []*models.RFRule{}})
			if tmpErr != nil {
				log.Error(nil, log.LOGGER_APP, "Remove remote cluster rule file fail", zap.String("cluster", cluster.Id), zap.Error(tmpErr))
			}
		}
	}
}

func getAlarmStrategyWithExpr(endpointGroup string) (result []*models.AlarmStrategyMetricObj, err error) {
	result = []*models.AlarmStrategyMetricObj{}
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where endpoint_group=?", endpointGroup).Find(&result)
	return
}

func getAlarmStrategyWithExprNew(endpointGroup string) (result, monitorEngineStrategyList []*models.AlarmStrategyMetricObj, err error) {
	var strategyRows []*models.AlarmStrategyMetricObj
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where t1.endpoint_group=?", endpointGroup).Find(&strategyRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy table fail with endpointGroup:%s ,err:%s ", endpointGroup, err.Error())
		return
	}
	var strategyMetricRows []*models.AlarmStrategyMetricWithExpr
	err = x.SQL("select t1.guid,t1.alarm_strategy,t1.metric,t1.`condition`,t1.`last`,t1.crc_hash,t1.log_type,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type',t1.monitor_engine from alarm_strategy_metric t1 left join metric t2 on t1.metric=t2.guid where t1.alarm_strategy in (select guid from alarm_strategy where endpoint_group=?)", endpointGroup).Find(&strategyMetricRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy metric with endpointGroup:%s fail,%s ", endpointGroup, err.Error())
		return
	}
	if len(strategyMetricRows) == 0 {
		result = strategyRows
		return
	}
	strategyRowMap := make(map[string]*models.AlarmStrategyMetricObj)
	for _, strategyRow := range strategyRows {
		strategyRowMap[strategyRow.Guid] = strategyRow
		withConditionFlag := false
		for _, metricRow := range strategyMetricRows {
			if metricRow.AlarmStrategy == strategyRow.Guid {
				withConditionFlag = true
				break
			}
		}
		if !withConditionFlag {
			result = append(result, strategyRow)
		}
	}
	var metricGuidList []string
	for _, row := range strategyMetricRows {
		metricGuidList = append(metricGuidList, row.Guid)
	}
	filterSql, filterParams := createListParams(metricGuidList, "")
	var tagRows []*models.AlarmStrategyTag
	err = x.SQL("select * from alarm_strategy_tag where alarm_strategy_metric in ("+filterSql+")", filterParams...).Find(&tagRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy tag fail,%s ", err.Error())
		return
	}
	var tagValueRows []*models.AlarmStrategyTagValue
	err = x.SQL("select * from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from alarm_strategy_tag where alarm_strategy_metric in ("+filterSql+"))", filterParams...).Find(&tagValueRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy tag value fail,%s ", err.Error())
		return
	}
	for _, metricRow := range strategyMetricRows {
		if strategyRow, ok := strategyRowMap[metricRow.AlarmStrategy]; ok {
			tmpStrategyObj := models.AlarmStrategyMetricObj{
				Guid:              strategyRow.Guid,
				Name:              strategyRow.Name,
				EndpointGroup:     strategyRow.EndpointGroup,
				Priority:          strategyRow.Priority,
				Content:           strategyRow.Content,
				NotifyEnable:      strategyRow.NotifyEnable,
				NotifyDelaySecond: strategyRow.NotifyDelaySecond,
				ActiveWindow:      strategyRow.ActiveWindow,
			}
			tmpStrategyObj.AlarmStrategyMetricGuid = metricRow.Guid
			tmpStrategyObj.Metric = metricRow.Metric
			tmpStrategyObj.Condition = metricRow.Condition
			tmpStrategyObj.Last = metricRow.Last
			tmpStrategyObj.MetricName = metricRow.MetricName
			tmpStrategyObj.MetricExpr = metricRow.MetricExpr
			tmpStrategyObj.MetricType = metricRow.MetricType
			tmpStrategyObj.ConditionCrc = metricRow.CrcHash
			tmpStrategyObj.LogType = metricRow.LogType
			tmpStrategyObj.Tags = []*models.MetricTag{}
			for _, tagRow := range tagRows {
				if tagRow.AlarmStrategyMetric == metricRow.Guid {
					tmpMetricTag := models.MetricTag{TagName: tagRow.Name, Equal: tagRow.Equal}
					for _, tagValueRow := range tagValueRows {
						if tagValueRow.AlarmStrategyTag == tagRow.Guid {
							tmpMetricTag.TagValue = append(tmpMetricTag.TagValue, tagValueRow.Value)
						}
					}
					tmpStrategyObj.Tags = append(tmpStrategyObj.Tags, &tmpMetricTag)
				}
			}
			if metricRow.MonitorEngine > 0 {
				monitorEngineStrategyList = append(monitorEngineStrategyList, &tmpStrategyObj)
			} else {
				result = append(result, &tmpStrategyObj)
			}
		}
	}
	return
}

func buildRuleReplaceExprNew(endpointList []*models.EndpointNewTable) (guidExpr, addressExpr, ipExpr string) {
	for _, endpoint := range endpointList {
		addressExpr += endpoint.AgentAddress + "|"
		guidExpr += endpoint.Guid + "|"
		ipExpr += endpoint.Ip + "|"
	}
	if addressExpr != "" {
		addressExpr = addressExpr[:len(addressExpr)-1]
	}
	if guidExpr != "" {
		guidExpr = guidExpr[:len(guidExpr)-1]
	}
	if ipExpr != "" {
		ipExpr = ipExpr[:len(ipExpr)-1]
	}
	return
}

func buildRuleFileContentNew(ruleFileName, guidExpr, addressExpr, ipExpr string, strategyList []*models.AlarmStrategyMetricObj) models.RFGroup {
	result := models.RFGroup{Name: ruleFileName}
	if len(strategyList) == 0 {
		return result
	}
	for _, strategy := range strategyList {
		tmpRfu := models.RFRule{}
		if strategy.ConditionCrc != "" {
			tmpRfu.Alert = fmt.Sprintf("%s_%s", strategy.ConditionCrc, strategy.Guid)
		} else {
			tmpRfu.Alert = fmt.Sprintf("%s_%s", strategy.Metric, strategy.Guid)
		}
		if !strings.Contains(strategy.Condition, " ") && strategy.Condition != "" {
			if strings.Contains(strategy.Condition, "=") {
				strategy.Condition = strategy.Condition[:2] + " " + strategy.Condition[2:]
			} else {
				strategy.Condition = strategy.Condition[:1] + " " + strategy.Condition[1:]
			}
		}
		buildStrategyAlarmRuleExpr(guidExpr, addressExpr, ipExpr, strategy)
		if strategy.MetricExpr == "" {
			log.Warn(nil, log.LOGGER_APP, "metric expr empty", zap.String("alertId", tmpRfu.Alert))
			continue
		}
		tmpRfu.Expr = fmt.Sprintf("(%s) %s", strategy.MetricExpr, strategy.Condition)
		tmpRfu.For = strategy.Last
		tmpRfu.Labels = make(map[string]string)
		tmpRfu.Labels["strategy_guid"] = strategy.Guid
		tmpRfu.Labels["condition_crc"] = strategy.ConditionCrc
		tmpRfu.Annotations = models.RFAnnotation{Summary: fmt.Sprintf("{{$labels.instance}}__%s__%s__{{$value}}", strategy.Priority, strategy.Metric), Description: strategy.Content}
		result.Rules = append(result.Rules, &tmpRfu)
	}
	return result
}

func buildStrategyAlarmRuleExpr(guidExpr, addressExpr, ipExpr string, strategy *models.AlarmStrategyMetricObj) {
	if strings.Contains(strategy.MetricExpr, "$address") {
		if strings.Contains(addressExpr, "|") {
			strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$address\"", "=~\""+addressExpr+"\"", -1)
		} else {
			strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$address\"", "=\""+addressExpr+"\"", -1)
		}
	}
	if strings.Contains(strategy.MetricExpr, "$guid") {
		if strings.Contains(guidExpr, "|") {
			strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$guid\"", "=~\""+guidExpr+"\"", -1)
		} else {
			strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$guid\"", "=\""+guidExpr+"\"", -1)
		}
	}
	if strings.Contains(strategy.MetricExpr, "$ip") {
		if strings.Contains(ipExpr, "|") {
			tmpStr := strings.Split(strategy.MetricExpr, "$ip")[1]
			tmpStr = tmpStr[:strings.Index(tmpStr, "\"")]
			newList := []string{}
			for _, v := range strings.Split(ipExpr, "|") {
				newList = append(newList, v+tmpStr)
			}
			strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$ip"+tmpStr+"\"", "=~\""+strings.Join(newList, "|")+"\"", -1)
		} else {
			strategy.MetricExpr = strings.ReplaceAll(strategy.MetricExpr, "$ip", ipExpr)
		}
	}
	if len(strategy.Tags) > 0 {
		if strategy.LogType == models.LogMonitorCustomType {
			originPromQl := strategy.MetricExpr
			// 原表达式: node_log_metric_monitor_value{key="key",agg="max",service_group="log_sys",tags="$t_tags"}
			// 正则字符串匹配 替换成 node_log_metric_monitor_value{key="key",agg="max",service_group="log_sys",tags!~".*test_service_code=addUser.*"}
			for i, tagObj := range strategy.Tags {
				if i == 0 {
					strategy.MetricExpr = getTagPromQl(convertMetricTag2Dto(tagObj), originPromQl)
				} else {
					strategy.MetricExpr = "(" + strategy.MetricExpr + ") and (" + getTagPromQl(convertMetricTag2Dto(tagObj), originPromQl) + ")"
				}
			}
		} else {
			for _, tagObj := range strategy.Tags {
				tagSourceString := "$t_" + tagObj.TagName
				if strings.Contains(strategy.MetricExpr, tagSourceString) {
					if len(tagObj.TagValue) == 0 {
						strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\""+tagSourceString+"\"", "=~\".*\"", -1)
					} else {
						tmpEqual := "=~"
						if tagObj.Equal == "notin" {
							tmpEqual = "!~"
						}
						strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\""+tagSourceString+"\"", tmpEqual+"\""+strings.Join(tagObj.TagValue, "|")+"\"", -1)
					}
				}
			}
		}
	}
	if strings.Contains(strategy.MetricExpr, "@") {
		strategy.MetricExpr = strings.ReplaceAll(strategy.MetricExpr, "@", "")
	}
}

func convertMetricTag2Dto(tag *models.MetricTag) *models.TagDto {
	if tag == nil {
		return &models.TagDto{}
	}
	return &models.TagDto{
		TagName:  tag.TagName,
		Equal:    tag.Equal,
		TagValue: tag.TagValue,
	}
}

func copyStrategyListNew(inputs []*models.AlarmStrategyMetricObj) (result []*models.AlarmStrategyMetricObj) {
	result = []*models.AlarmStrategyMetricObj{}
	for _, strategy := range inputs {
		tmpStrategy := models.AlarmStrategyMetricObj{Guid: strategy.Guid, Metric: strategy.Metric, Condition: strategy.Condition,
			Last: strategy.Last, Priority: strategy.Priority, Content: strategy.Content, NotifyEnable: strategy.NotifyEnable,
			NotifyDelaySecond: strategy.NotifyDelaySecond, MetricName: strategy.MetricName, MetricExpr: strategy.MetricExpr,
			MetricType: strategy.MetricType, ConditionCrc: strategy.ConditionCrc, Tags: strategy.Tags, LogType: strategy.LogType}
		result = append(result, &tmpStrategy)
	}
	return result
}

func GetAlarmObj(query *models.AlarmTable) (result models.AlarmTable, err error) {
	result = models.AlarmTable{}
	var alarmList []*models.AlarmTable
	baseSql := "select * from alarm where 1=1 "
	queryParams := []interface{}{}
	if query.Id > 0 {
		baseSql += " and id=? "
		queryParams = append(queryParams, query.Id)
	}
	if query.Endpoint != "" {
		baseSql += " and endpoint=? "
		queryParams = append(queryParams, query.Endpoint)
	}
	if query.Tags != "" {
		baseSql += " and tags=? "
		queryParams = append(queryParams, query.Tags)
	}
	if query.StrategyId > 0 {
		baseSql += " and strategy_id=? "
		queryParams = append(queryParams, query.StrategyId)
	}
	if query.AlarmStrategy != "" {
		baseSql += " and alarm_strategy=? "
		queryParams = append(queryParams, query.AlarmStrategy)
	}
	if query.SMetric != "" {
		baseSql += " and s_metric=? "
		queryParams = append(queryParams, query.SMetric)
	}
	if query.Status != "" {
		baseSql += " and status=? "
		queryParams = append(queryParams, query.Status)
	}
	err = x.SQL(baseSql, queryParams...).Find(&alarmList)
	if len(alarmList) > 0 {
		result = *alarmList[len(alarmList)-1]
	}
	return
}

func NotifyStrategyAlarm(alarmObj *models.AlarmHandleObj) {
	if alarmObj.AlarmStrategy == "" {
		log.Error(nil, log.LOGGER_APP, "Notify strategy alarm fail,alarmStrategy is empty", log.JsonObj("alarm", alarmObj))
		return
	}
	// 延迟发送通知，在延迟时间内如果告警恢复，则不发送通知，避免那种频繁告警恢复的场景
	if alarmObj.NotifyDelay > 0 {
		if alarmObj.Status == "firing" {
			time.Sleep(time.Duration(alarmObj.NotifyDelay) * time.Second)
			var nowAlarms []*models.AlarmTable
			x.SQL("select id,status from alarm where id=?", alarmObj.Id).Find(&nowAlarms)
			if len(nowAlarms) > 0 {
				if nowAlarms[0].Status == "ok" {
					log.Info(nil, log.LOGGER_APP, "Notify firing alarm break in delay time ", zap.Int("alarmId", alarmObj.Id))
					return
				}
			}
		} else if alarmObj.Status == "ok" {
			var nowAlarms []*models.AlarmTable
			x.SQL("select id,`start` from alarm where id=?", alarmObj.Id).Find(&nowAlarms)
			if len(nowAlarms) > 0 {
				if (time.Now().Unix() - nowAlarms[0].Start.Unix()) < int64(alarmObj.NotifyDelay) {
					log.Info(nil, log.LOGGER_APP, "Notify ok alarm break in delay time ", zap.Int("alarmId", alarmObj.Id))
					return
				}
			}
		}
	}
	// 1.先去单条阈值配置里找通知配置(单条阈值配置里的通知配置)，优先找这颗粒度最小的配置
	notifyObject := &models.NotifyTable{}
	var notifyQueryRows []*models.NotifyTable
	err := x.SQL("select * from notify where alarm_action=? and alarm_strategy=?", alarmObj.Status, alarmObj.AlarmStrategy).Find(&notifyQueryRows)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query notify table fail", zap.Error(err))
		return
	}
	for _, v := range notifyQueryRows {
		notifyRoles := getNotifyRoles(v.Guid)
		if len(notifyRoles) == 0 && v.ProcCallbackKey == "" {
			continue
		}
		notifyObject = v
		break
	}
	// 2.如果没有再去找策略所属endpoint_group组的策略(就是界面上阈值配置给某类对象组某种对象配的接收人设置)
	if notifyObject.Guid == "" {
		var affectServiceGroupList []string
		var serviceGroup []*models.EndpointServiceRelTable
		queryErr := x.SQL("select distinct service_group from endpoint_service_rel where endpoint=?", alarmObj.Endpoint).Find(&serviceGroup)
		if queryErr != nil {
			log.Error(nil, log.LOGGER_APP, "NotifyStrategyAlarm query endpoint service rel fail", zap.Int("alarmId", alarmObj.Id), zap.Error(queryErr))
		}
		for _, v := range serviceGroup {
			tmpGuidList, _ := fetchGlobalServiceGroupParentGuidList(v.ServiceGroup)
			for _, vv := range tmpGuidList {
				affectServiceGroupList = append(affectServiceGroupList, vv)
			}
		}
		var tmpNotifyQueryRows []*models.NotifyTable
		queryErr = x.SQL("select * from notify where alarm_action=? and endpoint_group in (select endpoint_group from alarm_strategy where guid=?)", alarmObj.Status, alarmObj.AlarmStrategy).Find(&tmpNotifyQueryRows)
		if queryErr != nil {
			log.Error(nil, log.LOGGER_APP, "NotifyStrategyAlarm query alarm notify fail", zap.Int("alarmId", alarmObj.Id), zap.String("alarmStrategy", alarmObj.AlarmStrategy), zap.Error(queryErr))
		} else {
			if len(tmpNotifyQueryRows) > 0 {
				notifyObject = tmpNotifyQueryRows[0]
				notifyObject.AffectServiceGroup = affectServiceGroupList
			}
		}
	}
	// 3.如果都没有，则构造一条通知配置defaultNotify，尝试使用全局接收人接收通知
	if notifyObject.Guid == "" {
		log.Info(nil, log.LOGGER_APP, "can not find notify config,use default notify", zap.Int("alarmId", alarmObj.Id), zap.String("strategy", alarmObj.AlarmStrategy))
		notifyObject = &models.NotifyTable{Guid: "defaultNotify", AlarmAction: alarmObj.Status, NotifyNum: 1}
	}
	if alarmObj.Status == "firing" {
		if notifyObject.ProcCallbackMode == models.AlarmNotifyManualMode && notifyObject.ProcCallbackKey != "" {
			if _, execErr := x.Exec("update alarm set notify_id=? where id=?", notifyObject.Guid, alarmObj.Id); execErr != nil {
				log.Error(nil, log.LOGGER_APP, "update alarm table notify id fail", zap.Int("alarmId", alarmObj.Id), zap.Error(execErr))
			}
		}
	}
	notifyAction(notifyObject, alarmObj)
}

func notifyAction(notify *models.NotifyTable, alarmObj *models.AlarmHandleObj) {
	log.Info(nil, log.LOGGER_APP, "Start notify action", zap.String("procCallKey", notify.ProcCallbackKey), zap.String("notify", notify.Guid), zap.Int("alarm", alarmObj.Id))
	// alarmMailEnable==Y
	var err, mailErr error
	if models.AlarmMailEnable {
		mailErr = notifyMailAction(notify, alarmObj)
		if mailErr != nil {
			log.Error(nil, log.LOGGER_APP, "Notify mail fail", zap.String("notifyGuid", notify.Guid), zap.Error(mailErr))
		}
	}
	if notify.ProcCallbackMode != models.AlarmNotifyAutoMode {
		log.Info(nil, log.LOGGER_APP, "notify proc callback mode is not auto,done", zap.Int("alarmId", alarmObj.Id), zap.String("notifyId", notify.Guid), zap.String("mode", notify.ProcCallbackMode))
		return
	}
	if alarmObj.SPriority == "" {
		tmpAlarmRows, _ := x.QueryString("select s_priority from alarm where id=?", alarmObj.Id)
		if len(tmpAlarmRows) > 0 {
			alarmObj.SPriority = tmpAlarmRows[0]["s_priority"]
		}
	}
	withoutRetry := false
	for i := 0; i < 3; i++ {
		withoutRetry, err = notifyEventAction(notify, alarmObj, true, "system")
		if err == nil {
			break
		} else {
			log.Error(nil, log.LOGGER_APP, "Notify event fail", zap.String("notifyGuid", notify.Guid), zap.Int("try", i), zap.Error(err))
		}
		if withoutRetry {
			break
		}
	}
}

func compareNotifyEventLevel(level string) bool {
	result := true
	if level == "medium" {
		if models.Config().MonitorAlarmCallbackLevelMin == "high" {
			result = false
		}
	} else if level == "low" {
		if models.Config().MonitorAlarmCallbackLevelMin == "high" || models.Config().MonitorAlarmCallbackLevelMin == "medium" {
			result = false
		}
	}
	return result
}

func notifyEventAction(notify *models.NotifyTable, alarmObj *models.AlarmHandleObj, compareLevel bool, operator string) (withoutRetry bool, err error) {
	if compareLevel && !compareNotifyEventLevel(alarmObj.SPriority) {
		log.Info(nil, log.LOGGER_APP, "notify event disable", zap.String("level", alarmObj.SPriority), zap.String("minLevel", models.Config().MonitorAlarmCallbackLevelMin))
		// err = notifyMailAction(notify, alarmObj)
		return
	}
	if notify.ProcCallbackKey == "" {
		if alarmObj.Status == "firing" {
			if models.FiringCallback != "" && models.FiringCallback != models.DefaultFiringCallback {
				notify.ProcCallbackKey = models.FiringCallback
			}
		} else {
			if models.RecoverCallback != "" && models.RecoverCallback != models.DefaultRecoverCallback {
				notify.ProcCallbackKey = models.RecoverCallback
			}
		}
		if notify.ProcCallbackKey == "" {
			err = fmt.Errorf("Notify:%s procCallbackKey is empty ", notify.Guid)
			withoutRetry = true
			return
		}
	}
	var requestParam models.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("%d-%s-%d-%s", alarmObj.Id, alarmObj.Status, time.Now().Unix(), notify.Guid)
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "SYS_MONITOR"
	requestParam.OperationKey = notify.ProcCallbackKey
	requestParam.OperationData = fmt.Sprintf("%d-%s-%s-%s", alarmObj.Id, alarmObj.Status, notify.Guid, operator)
	requestParam.OperationUser = operator
	log.Info(nil, log.LOGGER_APP, fmt.Sprintf("new notify request data --> eventSeqNo:%s operationKey:%s operationData:%s", requestParam.EventSeqNo, requestParam.OperationKey, requestParam.OperationData))
	b, _ := json.Marshal(requestParam)
	request, reqErr := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", models.CoreUrl), strings.NewReader(string(b)))
	request.Header.Set("Authorization", models.GetCoreToken())
	request.Header.Set("Content-Type", "application/json")
	if reqErr != nil {
		err = fmt.Errorf("Notify core event new request fail, %s ", reqErr.Error())
		return
	}
	res, doHttpErr := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if doHttpErr != nil {
		err = fmt.Errorf("Notify core event ctxhttp request fail,%s ", doHttpErr.Error())
		return
	}
	resultBody, _ := ioutil.ReadAll(res.Body)
	var resultObj models.CoreNotifyResult
	err = json.Unmarshal(resultBody, &resultObj)
	res.Body.Close()
	if err != nil {
		err = fmt.Errorf("Notify core event unmarshal json body fail,%s ", err.Error())
		return
	}
	log.Info(nil, log.LOGGER_APP, "Notify core result", zap.String("body", string(resultBody)))
	return
}

func getNotifyEventMessage(notifyGuid string, alarm models.AlarmTable) (result models.AlarmEntityObj) {
	notifyObj, err := getSimpleNotify(notifyGuid)
	if err != nil {
		log.Warn(nil, log.LOGGER_APP, "getNotifyEventMessage fail", zap.Error(err))
	} else {
		notifyObj = models.NotifyTable{}
	}
	result = models.AlarmEntityObj{}
	alarmDetailList := []*models.AlarmDetailData{}
	if strings.HasPrefix(alarm.EndpointTags, "ac_") {
		alarmDetailList, err = GetAlarmDetailList(alarm.Id)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "get alarm detail list fail", zap.Error(err))
		}
	} else {
		alarmDetailList = append(alarmDetailList, &models.AlarmDetailData{Metric: alarm.SMetric, Cond: alarm.SCond, Last: alarm.SLast, Start: alarm.Start, StartValue: alarm.StartValue, End: alarm.End, EndValue: alarm.EndValue, Tags: alarm.Tags})
	}
	alarmObj := models.AlarmHandleObj{AlarmTable: alarm}
	alarmObj.AlarmDetail = buildAlarmDetailData(alarmDetailList, "\r\n")
	result.Subject, result.Content = getNotifyMessage(&alarmObj)
	var roles []*models.RoleNewTable
	if notifyObj.ServiceGroup != "" {
		x.SQL("select guid,email from role_new where disable=0 and guid in (select `role` from service_group_role_rel where service_group=?)", notifyObj.ServiceGroup).Find(&roles)
	} else {
		x.SQL("select guid,email,phone from `role_new` where disable=0 and guid in (select `role` from notify_role_rel where notify=?)", notifyGuid).Find(&roles)
	}
	var email, phone, role []string
	emailExistMap := make(map[string]int)
	phoneExistMap := make(map[string]int)
	for _, v := range roles {
		if v.Email != "" {
			if _, b := emailExistMap[v.Email]; !b {
				email = append(email, v.Email)
				emailExistMap[v.Email] = 1
			}
		}
		if v.Phone != "" {
			if _, b := phoneExistMap[v.Phone]; !b {
				phone = append(phone, v.Phone)
				phoneExistMap[v.Phone] = 1
			}
		}
		role = append(role, v.Guid)
	}
	tmpEmailList := getRoleMail(role)
	if len(tmpEmailList) > 0 {
		email = tmpEmailList
	}
	if len(email) == 0 {
		email = models.DefaultMailReceiver
	}
	result.To = strings.Join(email, ",")
	result.ToMail = result.To
	result.ToPhone = strings.Join(phone, ",")
	result.ToRole = strings.Join(role, ",")
	result.SmsContent = getSmsAlarmContent(&alarm)
	return result
}

func notifyMailAction(notify *models.NotifyTable, alarmObj *models.AlarmHandleObj) error {
	var roles []*models.RoleNewTable
	var toAddress, roleList, tmpToAddress []string
	var queryRoleErr error
	if notify.ServiceGroup != "" {
		queryRoleErr = x.SQL("select guid,email from role_new where disable=0 and guid in (select `role` from service_group_role_rel where service_group=?)", notify.ServiceGroup).Find(&roles)
	} else {
		if len(notify.AffectServiceGroup) > 0 {
			queryRoleErr = x.SQL("select guid,email from `role_new` where disable=0 and guid in (select `role` from notify_role_rel where notify=? union select `role` from service_group_role_rel where service_group in ('"+strings.Join(notify.AffectServiceGroup, "','")+"'))", notify.Guid).Find(&roles)
		} else {
			queryRoleErr = x.SQL("select guid,email from `role_new` where disable=0 and guid in (select `role` from notify_role_rel where notify=?)", notify.Guid).Find(&roles)
		}
	}
	if queryRoleErr != nil {
		log.Error(nil, log.LOGGER_APP, "notifyMailAction query role fail", zap.Int("alarmId", alarmObj.Id), zap.Error(queryRoleErr))
	}
	// 先拿自己角色表的邮箱，独立运行的情况下有用
	for _, v := range roles {
		if v.Email != "" {
			toAddress = append(toAddress, v.Email)
		}
		roleList = append(roleList, v.Guid)
	}
	// 尝试去拿平台角色的邮箱
	if models.CoreUrl != "" {
		tmpToAddress = getRoleMail(roleList)
		if len(tmpToAddress) > 0 {
			toAddress = tmpToAddress
		}
	}
	// 如果都没有，那就尝试用全局默认接收人
	if len(toAddress) == 0 {
		toAddress = models.DefaultMailReceiver
	}
	if len(toAddress) == 0 {
		log.Warn(nil, log.LOGGER_APP, "notifyMailAction toAddress empty", zap.String("notify", notify.Guid), zap.Strings("roleList", roleList))
		return nil
	}
	for _, v := range toAddress {
		for _, vv := range strings.Split(v, ",") {
			if vv != "" {
				tmpToAddress = append(tmpToAddress, vv)
			}
		}
	}
	toAddress = tmpToAddress
	mailConfig, err := GetSysAlertMailConfig()
	if err != nil {
		return err
	}
	mailSender := smtp.MailSender{SenderName: mailConfig.SenderName, SenderMail: mailConfig.SenderMail, AuthServer: mailConfig.AuthServer, AuthPassword: mailConfig.AuthPassword, AuthUser: mailConfig.AuthUser}
	mailConfig.SSL = strings.ToLower(mailConfig.SSL)
	if mailConfig.SSL == "y" {
		mailSender.SSL = true
	} else if mailConfig.SSL == "starttls" {
		mailSender.SSL = true
		mailSender.ByStartTLS = true
	}
	err = mailSender.Init()
	if err != nil {
		return err
	}
	alarmDetailList := []*models.AlarmDetailData{}
	if strings.HasPrefix(alarmObj.EndpointTags, "ac_") {
		alarmDetailList, err = GetAlarmDetailList(alarmObj.Id)
		if err != nil {
			return err
		}
	} else {
		alarmDetailList = append(alarmDetailList, &models.AlarmDetailData{Metric: alarmObj.SMetric, Cond: alarmObj.SCond, Last: alarmObj.SLast, Start: alarmObj.Start, StartValue: alarmObj.StartValue, End: alarmObj.End, EndValue: alarmObj.EndValue, Tags: alarmObj.Tags})
	}
	alarmObj.AlarmDetail = buildAlarmDetailData(alarmDetailList, "\r\n")
	subject, content := getNotifyMessage(alarmObj)
	return mailSender.Send(subject, content, toAddress)
}

func getNotifyMessage(alarmObj *models.AlarmHandleObj) (subject, content string) {
	subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", alarmObj.Status, alarmObj.SPriority, alarmObj.Endpoint, alarmObj.SMetric)
	if strings.HasPrefix(alarmObj.EndpointTags, "ac_") {
		content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s\r\nDetail:\r\n%s", alarmObj.Endpoint, alarmObj.Status, alarmObj.SMetric, alarmObj.SPriority, alarmObj.Content, time.Now().Format(models.DatetimeFormat), alarmObj.AlarmDetail)
	} else {
		content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s\r\nDetail:\r\n%s", alarmObj.Endpoint, alarmObj.Status, alarmObj.SMetric, alarmObj.StartValue, alarmObj.SCond, alarmObj.SLast, alarmObj.SPriority, alarmObj.Content, time.Now().Format(models.DatetimeFormat), alarmObj.AlarmDetail)
	}
	return
}

func getRoleMail(roleList []string) (mailList []string) {
	if len(roleList) == 0 {
		return
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/roles/retrieve", models.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core role key new request fail", zap.Error(err))
		return
	}
	request.Header.Set("Authorization", models.GetCoreToken())
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core role key ctxhttp request fail", zap.Error(err))
		return
	}
	b, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var result models.CoreRoleDto
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core role key json unmarshal result", zap.Error(err))
		return
	}
	existMap := make(map[string]int)
	for _, v := range roleList {
		tmpMail := ""
		for _, vv := range result.Data {
			if vv.Name == v {
				tmpMail = vv.Email
				break
			}
		}
		if tmpMail != "" {
			if _, b := existMap[tmpMail]; !b {
				mailList = append(mailList, tmpMail)
				existMap[tmpMail] = 1
			}
		}
	}
	return
}

func ImportAlarmStrategy(queryType, inputGuid string, param []*models.EndpointStrategyObj, operator, importRule string) (err error, metricNotFound, nameDuplicate []string) {
	if len(param) == 0 {
		err = fmt.Errorf("import content empty ")
		return
	}
	var actions []*Action
	var metricTable []*models.MetricTable
	err = x.SQL("select guid,metric,monitor_type,service_group from metric where service_group is null or service_group=?", inputGuid).Find(&metricTable)
	if err != nil {
		err = fmt.Errorf("query metric table fail,%s ", err.Error())
		return
	}
	var endpointGroupList []string
	metricMap := make(map[string]*models.MetricTable)
	for _, v := range metricTable {
		metricMap[v.Metric] = v
	}
	log.Info(nil, log.LOGGER_APP, "ImportAlarmStrategy", zap.String("inputGuid", inputGuid), log.JsonObj("metricMap", metricMap))
	nowTime := time.Now().Format(models.DatetimeFormat)
	if queryType == "group" {
		var endpointGroupTable []*models.EndpointGroupTable
		err = x.SQL("select guid,monitor_type,service_group from endpoint_group where guid=?", inputGuid).Find(&endpointGroupTable)
		if err != nil {
			err = fmt.Errorf("query endpoint group table fail,%s ", err.Error())
			return
		}
		if len(endpointGroupTable) == 0 {
			err = fmt.Errorf("can not find endpoint group with guid:%s ", inputGuid)
			return
		}
		endpointGroupList = append(endpointGroupList, inputGuid)
		tmpActions, tmpErr, tmpMetricNotFound, tmpNameDuplicate := getAlarmStrategyImportActions(inputGuid, nowTime, operator, importRule, param[0], metricMap)
		if tmpErr != nil {
			metricNotFound = tmpMetricNotFound
			nameDuplicate = tmpNameDuplicate
			err = tmpErr
			return
		}
		actions = append(actions, tmpActions...)
		actions = append(actions, getStrategyNotifyImportActions(inputGuid, param[0].NotifyList)...)
	} else if queryType == "service" {
		var endpointGroupTable []*models.EndpointGroupTable
		err = x.SQL("select guid,monitor_type,service_group from endpoint_group where service_group=?", inputGuid).Find(&endpointGroupTable)
		if err != nil {
			err = fmt.Errorf("query endpoint group table fail,%s ", err.Error())
			return
		}
		for _, v := range param {
			tmpMatchEndpointGroup := ""
			for _, vv := range endpointGroupTable {
				if vv.MonitorType == v.MonitorType {
					tmpMatchEndpointGroup = vv.Guid
					break
				}
			}
			if tmpMatchEndpointGroup == "" {
				err = fmt.Errorf("ServiceGroup:%s can not find monitorType:%s ", inputGuid, v.MonitorType)
				return
			}
			endpointGroupList = append(endpointGroupList, tmpMatchEndpointGroup)
			tmpActions, tmpErr, tmpMetricNotFound, tmpNameDuplicate := getAlarmStrategyImportActions(tmpMatchEndpointGroup, nowTime, operator, importRule, v, metricMap)
			if tmpErr != nil {
				metricNotFound = tmpMetricNotFound
				nameDuplicate = tmpNameDuplicate
				err = fmt.Errorf("handle endpointGroup:%s fail,%s ", v.EndpointGroup, tmpErr.Error())
				break
			}
			actions = append(actions, tmpActions...)
			actions = append(actions, getStrategyNotifyImportActions(tmpMatchEndpointGroup, v.NotifyList)...)
		}
		if err != nil {
			return
		}
	}
	if len(actions) == 0 {
		err = fmt.Errorf("no alarm strategy match in exist data,do nothing ")
		return
	}
	err = Transaction(actions)
	if err == nil {
		for _, v := range endpointGroupList {
			err = SyncPrometheusRuleFile(v, false)
			if err != nil {
				break
			}
		}
	}
	return
}

func getAlarmStrategyImportActions(endpointGroup, nowTime, operator, importRule string, param *models.EndpointStrategyObj, metricMap map[string]*models.MetricTable) (actions []*Action, err error, metricNotFound, nameDuplicate []string) {
	var existStrategyTable []*models.AlarmStrategyTable
	var systemAlarmStrategyMap = convertString2Map(systemAlarmStrategyIds)
	var list []*models.GroupStrategyObj
	for i := 1; i <= 24; i++ {
		systemAlarmStrategyMap[fmt.Sprintf("old_%d", i)] = true
	}
	// 过滤掉系统自带数据,根据ID过滤,ID是固定的
	for _, obj := range param.Strategy {
		if systemAlarmStrategyMap[obj.Guid] {
			continue
		}
		list = append(list, obj)
	}
	param.Strategy = list

	err = x.SQL("select guid,name,metric from alarm_strategy where endpoint_group=?", endpointGroup).Find(&existStrategyTable)
	if err != nil {
		err = fmt.Errorf("query alarm strategy table fail,%s ", err.Error())
		return
	}
	existNameMap := make(map[string]string)
	for _, v := range existStrategyTable {
		existNameMap[v.Name] = v.Guid
	}
	for _, strategy := range param.Strategy {
		strategy.EndpointGroup = endpointGroup
		if guid, ok := existNameMap[strategy.Name]; ok {
			// 覆盖模式
			if importRule == string(models.ImportRuleCover) {
				var delAlarmStrategyActions []*Action
				if delAlarmStrategyActions, _, err = GetDeleteAlarmStrategyActions(guid); err != nil {
					return
				}
				actions = append(actions, delAlarmStrategyActions...)
			} else {
				strategy.Name = strategy.Name + "_1"
				if _, doubleCheck := existNameMap[strategy.Name]; doubleCheck {
					err = fmt.Errorf("name: %s duplicate", strategy.Name)
					nameDuplicate = append(nameDuplicate, strategy.Name)
					return
				}
			}
		}
		// 检测策略上的指标在不在
		if len(strategy.Conditions) > 0 {
			for _, v := range strategy.Conditions {
				tmpMetricName := v.MetricName
				if tmpMetricName == "" {
					if metricSplitIndex := strings.LastIndex(v.Metric, "__"); metricSplitIndex > 0 {
						tmpMetricName = v.Metric[:metricSplitIndex]
					}
				}
				if fMetric, b := metricMap[tmpMetricName]; b {
					v.Metric = fMetric.Guid
				} else {
					metricNotFound = append(metricNotFound, tmpMetricName)
					err = fmt.Errorf("Metric:%s not found ", tmpMetricName)
					break
				}
				if logType, err2 := GetLogTypeByMetric(tmpMetricName); err2 != nil {
					log.Error(nil, log.LOGGER_APP, "GetLogTypeByMetric err", zap.Error(err2))
				} else {
					v.LogType = logType
				}
			}
		} else {
			tmpMetricName := strategy.MetricName
			if tmpMetricName == "" {
				if metricSplitIndex := strings.LastIndex(strategy.Metric, "__"); metricSplitIndex > 0 {
					tmpMetricName = strategy.Metric[:metricSplitIndex]
				}
			}
			if fMetric, b := metricMap[tmpMetricName]; b {
				strategy.Metric = fMetric.Guid
			} else {
				metricNotFound = append(metricNotFound, tmpMetricName)
				err = fmt.Errorf("Metric:%s not found ", tmpMetricName)
			}
		}
		if err != nil {
			return
		}

		newAction, buildErr := getCreateAlarmStrategyActions(strategy, nowTime, operator)
		if buildErr != nil {
			err = buildErr
			break
		}
		actions = append(actions, newAction...)
	}
	return
}

func GetExistAlarmCondition(endpoint, alarmStrategyGuid, crcHash, tags string) (existAlarm models.AlarmTable, alarmConditionGuid string, err error) {
	var alarmConditionRows []*models.AlarmCondition
	err = x.SQL("select * from alarm_condition where endpoint=? and alarm_strategy=? and crc_hash=? and tags=? order by `start` desc limit 1", endpoint, alarmStrategyGuid, crcHash, tags).Find(&alarmConditionRows)
	if err != nil {
		err = fmt.Errorf("query alarm condition table fail,%s ", err.Error())
		return
	}
	existAlarm = models.AlarmTable{}
	if len(alarmConditionRows) > 0 {
		alarmConditionRow := alarmConditionRows[0]
		alarmConditionGuid = alarmConditionRow.Guid
		existAlarm.Status = alarmConditionRow.Status
		existAlarm.AlarmStrategy = alarmStrategyGuid
	}
	return
}

func GetSimpleAlarmStrategy(alarmStrategyGuid string) (result *models.AlarmStrategyTable, err error) {
	var alarmStrategyRows []*models.AlarmStrategyTable
	err = x.SQL("select * from alarm_strategy where guid=?", alarmStrategyGuid).Find(&alarmStrategyRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy table fail,%s ", alarmStrategyGuid)
		return
	}
	if len(alarmStrategyRows) == 0 {
		err = fmt.Errorf("can not find alarm strategy with guid:%s ", alarmStrategyGuid)
		return
	}
	result = alarmStrategyRows[0]
	return
}

func getStrategyNotifyImportActions(endpointGroup string, notifyList []*models.NotifyObj) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from notify where endpoint_group=?", Param: []interface{}{endpointGroup}})
	for _, v := range notifyList {
		v.AlarmStrategy = ""
		v.EndpointGroup = endpointGroup
	}
	actions = append(actions, getNotifyListInsertAction(notifyList)...)
	return
}

func GetMailSender() (mailSender *smtp.MailSender, err error) {
	mailConfig, getConfigErr := GetSysAlertMailConfig()
	if getConfigErr != nil {
		return
	}
	mailSender = &smtp.MailSender{SenderName: mailConfig.SenderName, SenderMail: mailConfig.SenderMail, AuthServer: mailConfig.AuthServer, AuthPassword: mailConfig.AuthPassword}
	mailConfig.SSL = strings.ToLower(mailConfig.SSL)
	if mailConfig.SSL == "y" {
		mailSender.SSL = true
	} else if mailConfig.SSL == "starttls" {
		mailSender.SSL = true
		mailSender.ByStartTLS = true
	}
	err = mailSender.Init()
	if err != nil {
		err = fmt.Errorf("mail init fail,%s ", err.Error())
	}
	return
}

func GetMonitorEngineStrategy() (alarmStrategyMetricRows []*models.AlarmStrategyMetric, err error) {
	err = x.SQL("select * from alarm_strategy_metric where monitor_engine=1").Find(&alarmStrategyMetricRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy metric fail,%s ", err.Error())
	}
	return
}

func GetMonitorEngineMetricMap() (metricMap map[string]int, err error) {
	metricMap = make(map[string]int)
	queryRows, queryErr := x.QueryString("select guid from metric where db_metric_monitor<>'' union select metric_id as `guid` from metric_comparison")
	if queryErr != nil {
		err = fmt.Errorf("query dbMetric and comparison metric fail,%s ", queryErr.Error())
		return
	}
	for _, row := range queryRows {
		metricMap[row["guid"]] = 1
	}
	return
}

func UpdateAlarmStrategyMetricExpr(alarmStrategyMetricObj *models.AlarmStrategyMetricObj) {
	_, err := x.Exec("update alarm_strategy_metric set monitor_engine_expr=? where guid=?", alarmStrategyMetricObj.MetricExpr, alarmStrategyMetricObj.AlarmStrategyMetricGuid)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "UpdateAlarmStrategyMetricExpr fail", zap.String("alarmStrategyMetric", alarmStrategyMetricObj.Guid), zap.Error(err))
	}
}

func GetQuery() string {
	var value string
	query := "SELECT VARIABLE_VALUE AS Value FROM information_schema.GLOBAL_STATUS WHERE VARIABLE_NAME = 'Com_stmt_prepare';"
	_, err := x.SQL(query).Get(&value)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "doSQLQueryJob fail", zap.Error(err))
		return value
	}
	return value
}
func GetMonitorEngineAlarmList() (alarmList []*models.AlarmTable, err error) {
	err = x.SQL("select id,endpoint,status,s_metric,tags,alarm_strategy from alarm where status='firing' and alarm_strategy in (select alarm_strategy from alarm_strategy_metric where monitor_engine=1) order by id desc").Find(&alarmList)
	if err != nil {
		err = fmt.Errorf("get monitor engine alarm firing list fail,%s ", err.Error())
	}
	return
}

func GetAlarmStrategyNotifyWorkflowList() (result []*models.WorkflowDto, err error) {
	result = []*models.WorkflowDto{}
	var tempList []*models.WorkflowDto
	err = x.SQL("select distinct  proc_callback_name as name,proc_callback_key as 'key' from notify where  proc_callback_key!='' and proc_callback_name!=''").Find(&tempList)
	for _, dto := range tempList {
		if strings.TrimSpace(dto.Name) == "" || strings.TrimSpace(dto.Key) == "" {
			continue
		}
		result = append(result, &models.WorkflowDto{Key: dto.Key, Name: dto.Name})
	}
	return
}
