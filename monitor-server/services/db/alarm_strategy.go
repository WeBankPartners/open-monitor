package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/go-common-lib/smtp"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"golang.org/x/net/context/ctxhttp"
	"hash/crc64"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func QueryAlarmStrategyByGroup(endpointGroup string) (result []*models.EndpointStrategyObj, err error) {
	result = []*models.EndpointStrategyObj{}
	strategy := []*models.GroupStrategyObj{}
	var alarmStrategyTable []*models.AlarmStrategyMetricObj
	err = x.SQL("select t1.*,t2.metric as 'metric_name' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where t1.endpoint_group=? order by t1.update_time desc", endpointGroup).Find(&alarmStrategyTable)
	if err != nil {
		return
	}
	for _, v := range alarmStrategyTable {
		tmpStrategyObj := models.GroupStrategyObj{Guid: v.Guid, Name: v.Name, EndpointGroup: v.EndpointGroup, Metric: v.Metric, MetricName: v.MetricName, Condition: v.Condition, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelaySecond: v.NotifyDelaySecond, ActiveWindow: v.ActiveWindow}
		tmpStrategyObj.UpdateTime = v.UpdateTime
		tmpStrategyObj.UpdateUser = v.UpdateUser
		tmpStrategyObj.NotifyList = getNotifyList(v.Guid, "", "")
		if tmpStrategyConditions, tmpErr := getStrategyConditions(v.Guid); tmpErr != nil {
			err = tmpErr
			return
		} else {
			if len(tmpStrategyConditions) == 0 {
				tmpStrategyConditions = append(tmpStrategyConditions, &models.StrategyConditionObj{
					Metric:    v.Metric,
					Condition: v.Condition,
					Last:      v.Last,
					Tags:      []*models.MetricTag{},
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

func QueryAlarmStrategyByEndpoint(endpoint string) (result []*models.EndpointStrategyObj, err error) {
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
		tmpEndpointStrategyList, tmpErr := QueryAlarmStrategyByGroup(v.Guid)
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

func QueryAlarmStrategyByServiceGroup(serviceGroup string) (result []*models.EndpointStrategyObj, err error) {
	result = []*models.EndpointStrategyObj{}
	var endpointGroupTable []*models.EndpointGroupTable
	err = x.SQL("select guid,monitor_type,service_group from endpoint_group where service_group=?", serviceGroup).Find(&endpointGroupTable)
	if err != nil {
		return
	}
	for _, v := range endpointGroupTable {
		tmpEndpointStrategyList, tmpErr := QueryAlarmStrategyByGroup(v.Guid)
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
	nowTime := time.Now().Format(models.DatetimeFormat)
	actions, err := getCreateAlarmStrategyActions(param, nowTime, operator)
	if err != nil {
		return err
	}
	return Transaction(actions)
}

func getCreateAlarmStrategyActions(param *models.GroupStrategyObj, nowTime, operator string) (actions []*Action, err error) {
	param.Guid = "strategy_" + guid.CreateGuid()
	insertAction := Action{Sql: "insert into alarm_strategy(guid,name,endpoint_group,metric,`condition`,`last`,priority,content,notify_enable,notify_delay_second,active_window,update_time,update_user) value (?,?,?,?,?,?,?,?,?,?,?,?,?)"}
	insertAction.Param = []interface{}{param.Guid, param.Name, param.EndpointGroup, param.Metric, param.Condition, param.Last, param.Priority, param.Content, param.NotifyEnable, param.NotifyDelaySecond, param.ActiveWindow, nowTime, operator}
	actions = append(actions, &insertAction)
	if len(param.NotifyList) > 0 {
		for _, v := range param.NotifyList {
			v.AlarmStrategy = param.Guid
		}
		actions = append(actions, getNotifyListInsertAction(param.NotifyList)...)
	}
	if len(param.Conditions) > 0 {
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
	var actions []*Action
	updateAction := Action{Sql: "update alarm_strategy set name=?,priority=?,content=?,notify_enable=?,notify_delay_second=?,active_window=?,update_time=?,update_user=? where guid=?"}
	updateAction.Param = []interface{}{param.Name, param.Priority, param.Content, param.NotifyEnable, param.NotifyDelaySecond, param.ActiveWindow, nowTime, operator, param.Guid}
	actions = append(actions, &updateAction)
	for _, v := range param.NotifyList {
		v.AlarmStrategy = param.Guid
	}
	//actions = append(actions, getNotifyListDeleteAction(param.Guid, "", "")...)
	//actions = append(actions, getNotifyListInsertAction(param.NotifyList)...)
	actions = append(actions, getNotifyListUpdateAction(param.NotifyList)...)
	updateConditionActions, err := getStrategyConditionUpdateAction(param.Guid, param.Conditions)
	if err != nil {
		return err
	}
	actions = append(actions, updateConditionActions...)
	return Transaction(actions)
}

func DeleteAlarmStrategy(strategyGuid string) (endpointGroup string, err error) {
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
	var actions []*Action
	actions = append(actions, getNotifyListDeleteAction(strategyGuid, "", "")...)
	actions = append(actions, getStrategyConditionDeleteAction(strategyGuid)...)
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where guid=?", Param: []interface{}{strategyGuid}})
	err = Transaction(actions)
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
	} else {
		return actions
	}
	notifyGuidList := guid.CreateGuidList(len(notifyList))
	for i, v := range notifyList {
		if v.NotifyNum == 0 {
			v.NotifyNum = 1
		}
		tmpAction := Action{Sql: fmt.Sprintf("insert into notify(guid,%s,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param,proc_callback_mode,description) value (?,'%s',?,?,?,?,?,?,?,?,?)", refColumn, refValue)}
		tmpAction.Param = []interface{}{"notify_" + notifyGuidList[i], v.AlarmAction, v.AlarmPriority, v.NotifyNum, v.ProcCallbackName, v.ProcCallbackKey, v.CallbackUrl, v.CallbackParam, v.ProcCallbackMode, v.Description}
		actions = append(actions, &tmpAction)
		if len(v.NotifyRoles) > 0 {
			tmpNotifyRoleGuidList := guid.CreateGuidList(len(v.NotifyRoles))
			for ii, vv := range v.NotifyRoles {
				actions = append(actions, &Action{Sql: "insert into notify_role_rel(guid,notify,`role`) value (?,?,?)", Param: []interface{}{tmpNotifyRoleGuidList[ii], "notify_" + notifyGuidList[i], vv}})
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
	} else {
		return actions
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
			tmpAction := Action{Sql: fmt.Sprintf("insert into notify(guid,%s,alarm_action,alarm_priority,notify_num,proc_callback_name,proc_callback_key,callback_url,callback_param,proc_callback_mode,description) value (?,'%s',?,?,?,?,?,?,?,?,?)", refColumn, refValue)}
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
				tmpTag := models.MetricTag{TagName: tagRow.Name, TagValue: []string{}}
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
	for i, metricRow := range conditions {
		tmpMetricRowString, _ := json.Marshal(metricRow)
		tmpCrcHash := fmt.Sprintf("%d", crc64.Checksum(tmpMetricRowString, crc64.MakeTable(crc64.ECMA)))
		if _, existFlag := existCrcMap[tmpCrcHash]; existFlag {
			err = fmt.Errorf("metric condition is duplicated")
			return
		}
		existCrcMap[tmpCrcHash] = 1
		actions = append(actions, &Action{Sql: "insert into alarm_strategy_metric(guid,alarm_strategy,metric,`condition`,`last`,create_time,crc_hash) values (?,?,?,?,?,?,?)", Param: []interface{}{
			metricGuidList[i], alarmStrategyGuid, metricRow.Metric, metricRow.Condition, metricRow.Last, nowTime, tmpCrcHash,
		}})
		if len(metricRow.Tags) > 0 {
			tagGuidList := guid.CreateGuidList(len(metricRow.Tags))
			for tagIndex, tagRow := range metricRow.Tags {
				actions = append(actions, &Action{Sql: "insert into alarm_strategy_tag(guid,alarm_strategy_metric,name) values (?,?,?)", Param: []interface{}{tagGuidList[tagIndex], metricGuidList[i], tagRow.TagName}})
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

func SyncPrometheusRuleFile(endpointGroup string, fromPeer bool) error {
	if endpointGroup == "" {
		return fmt.Errorf("Sync prometheus rule fail,group is empty ")
	}
	endpointGroupObj, err := GetSimpleEndpointGroup(endpointGroup)
	if err != nil {
		return fmt.Errorf("Sync prometheus rule fail,%s ", err.Error())
	}
	log.Logger.Info("SyncPrometheusRuleFile", log.String("endpointGroup", endpointGroup))
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
	strategyList, getStrategyErr := getAlarmStrategyWithExprNew(endpointGroup)
	if getStrategyErr != nil {
		return getStrategyErr
	}
	log.Logger.Debug("SyncPrometheusRuleFile alarm strategy data", log.JsonObj("strategyList", strategyList))
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
			prom.SyncLocalRuleConfig(models.RuleLocalConfigJob{FromPeer: fromPeer, EndpointGroup: endpointGroup, Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
		} else {
			tmpErr := SyncRemoteRuleConfigFile(cluster, models.RFClusterRequestObj{Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
			if tmpErr != nil {
				err = fmt.Errorf("Update remote cluster:%s rule file fail,%s ", cluster, tmpErr.Error())
				log.Logger.Error("Update remote cluster rule file fail", log.String("cluster", cluster), log.Error(tmpErr))
			}
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
				log.Logger.Error("Remove remote cluster rule file fail", log.String("cluster", cluster.Id), log.Error(tmpErr))
			}
		}
	}
}

func getAlarmStrategyWithExpr(endpointGroup string) (result []*models.AlarmStrategyMetricObj, err error) {
	result = []*models.AlarmStrategyMetricObj{}
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where endpoint_group=?", endpointGroup).Find(&result)
	return
}

func getAlarmStrategyWithExprNew(endpointGroup string) (result []*models.AlarmStrategyMetricObj, err error) {
	var strategyRows []*models.AlarmStrategyMetricObj
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where endpoint_group=?", endpointGroup).Find(&strategyRows)
	var strategyMetricRows []*models.AlarmStrategyMetricWithExpr
	err = x.SQL("select t1.guid,t1.alarm_strategy,t1.metric,t1.`condition`,t1.`last`,t1.crc_hash,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy_metric t1 left join metric t2 on t1.metric=t2.guid where t1.alarm_strategy in (select guid from alarm_strategy where endpoint_group=?)", endpointGroup).Find(&strategyMetricRows)
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
			tmpStrategyObj.Metric = metricRow.Metric
			tmpStrategyObj.Condition = metricRow.Condition
			tmpStrategyObj.Last = metricRow.Last
			tmpStrategyObj.MetricName = metricRow.MetricName
			tmpStrategyObj.MetricExpr = metricRow.MetricExpr
			tmpStrategyObj.MetricType = metricRow.MetricType
			tmpStrategyObj.ConditionCrc = metricRow.CrcHash
			tmpStrategyObj.Tags = []*models.MetricTag{}
			for _, tagRow := range tagRows {
				if tagRow.AlarmStrategyMetric == metricRow.Guid {
					tmpMetricTag := models.MetricTag{TagName: tagRow.Name}
					for _, tagValueRow := range tagValueRows {
						if tagValueRow.AlarmStrategyTag == tagRow.Guid {
							tmpMetricTag.TagValue = append(tmpMetricTag.TagValue, tagValueRow.Value)
						}
					}
					tmpStrategyObj.Tags = append(tmpStrategyObj.Tags, &tmpMetricTag)
				}
			}
			result = append(result, &tmpStrategyObj)
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
			for _, tagObj := range strategy.Tags {
				tagSourceString := "$t_" + tagObj.TagName
				if strings.Contains(strategy.MetricExpr, tagSourceString) {
					if len(tagObj.TagValue) == 0 {
						strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\""+tagSourceString+"\"", "=~\".*\"", -1)
					} else {
						strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\""+tagSourceString+"\"", "=~\""+strings.Join(tagObj.TagValue, "|")+"\"", -1)
					}
				}
			}
		}
		if strings.Contains(strategy.MetricExpr, "@") {
			strategy.MetricExpr = strings.ReplaceAll(strategy.MetricExpr, "@", "")
		}
		if strategy.MetricExpr == "" {
			log.Logger.Warn("metric expr empty", log.String("alertId", tmpRfu.Alert))
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

func copyStrategyListNew(inputs []*models.AlarmStrategyMetricObj) (result []*models.AlarmStrategyMetricObj) {
	result = []*models.AlarmStrategyMetricObj{}
	for _, strategy := range inputs {
		tmpStrategy := models.AlarmStrategyMetricObj{Guid: strategy.Guid, Metric: strategy.Metric, Condition: strategy.Condition, Last: strategy.Last, Priority: strategy.Priority, Content: strategy.Content, NotifyEnable: strategy.NotifyEnable, NotifyDelaySecond: strategy.NotifyDelaySecond, MetricName: strategy.MetricName, MetricExpr: strategy.MetricExpr, MetricType: strategy.MetricType, ConditionCrc: strategy.ConditionCrc, Tags: strategy.Tags}
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
	baseSql += " order by id asc"
	err = x.SQL(baseSql, queryParams...).Find(&alarmList)
	if len(alarmList) > 0 {
		result = *alarmList[len(alarmList)-1]
	}
	return
}

func NotifyServiceGroup(serviceGroup string, alarmObj *models.AlarmHandleObj) {
	var notifyList []*models.NotifyTable
	err := x.SQL("select * from notify where service_group=?", serviceGroup).Find(&notifyList)
	if err != nil {
		log.Logger.Warn("Notify serviceGroup fail,query notify data error", log.Error(err))
	}
	if len(notifyList) == 0 {
		notifyList = []*models.NotifyTable{&models.NotifyTable{Guid: "defaultNotify", AlarmAction: alarmObj.Status, NotifyNum: 1}}
	}
	for _, v := range notifyList {
		notifyAction(v, alarmObj)
	}
}

func NotifyStrategyAlarm(alarmObj *models.AlarmHandleObj) {
	if alarmObj.AlarmStrategy == "" {
		log.Logger.Error("Notify strategy alarm fail,alarmStrategy is empty", log.JsonObj("alarm", alarmObj))
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
					log.Logger.Info("Notify firing alarm break in delay time ", log.Int("alarmId", alarmObj.Id))
					return
				}
			}
		} else if alarmObj.Status == "ok" {
			var nowAlarms []*models.AlarmTable
			x.SQL("select id,`start` from alarm where id=?", alarmObj.Id).Find(&nowAlarms)
			if len(nowAlarms) > 0 {
				if (time.Now().Unix() - nowAlarms[0].Start.Unix()) < int64(alarmObj.NotifyDelay) {
					log.Logger.Info("Notify ok alarm break in delay time ", log.Int("alarmId", alarmObj.Id))
					return
				}
			}
		}
	}
	// 1.先去单条阈值配置里找通知配置(单条阈值配置里的通知配置)，优先找这颗粒度最小的配置
	var notifyTable, activeNotifyRows []*models.NotifyTable
	err := x.SQL("select * from notify where alarm_action=? and alarm_strategy=?", alarmObj.Status, alarmObj.AlarmStrategy).Find(&notifyTable)
	if err != nil {
		log.Logger.Error("Query notify table fail", log.Error(err))
		return
	}
	for _, v := range notifyTable {
		notifyRoles := getNotifyRoles(v.Guid)
		if len(notifyRoles) == 0 && v.ProcCallbackKey == "" {
			continue
		}
		activeNotifyRows = append(activeNotifyRows, v)
	}
	notifyTable = activeNotifyRows
	// 2.如果没有再去找策略所属endpoint_group组的策略(就是界面上阈值配置给某类对象组某种对象配的接收人设置)
	if len(notifyTable) == 0 {
		var affectServiceGroupList []string
		var serviceGroup []*models.EndpointServiceRelTable
		x.SQL("select distinct service_group from endpoint_service_rel where endpoint=?", alarmObj.Endpoint).Find(&serviceGroup)
		for _, v := range serviceGroup {
			tmpGuidList, _ := fetchGlobalServiceGroupParentGuidList(v.ServiceGroup)
			for _, vv := range tmpGuidList {
				affectServiceGroupList = append(affectServiceGroupList, vv)
			}
		}
		x.SQL("select * from notify where alarm_action=? and endpoint_group in (select endpoint_group from alarm_strategy where guid=?) or service_group in ('"+strings.Join(affectServiceGroupList, "','")+"')", alarmObj.Status, alarmObj.AlarmStrategy).Find(&notifyTable)
	}
	// 3.如果都没有，则构造一条通知配置defaultNotify，尝试使用全局接收人接收通知
	if len(notifyTable) == 0 {
		log.Logger.Info("can not find notify config,use default notify", log.Int("alarmId", alarmObj.Id), log.String("strategy", alarmObj.AlarmStrategy))
		notifyTable = []*models.NotifyTable{&models.NotifyTable{Guid: "defaultNotify", AlarmAction: alarmObj.Status, NotifyNum: 1}}
	} else if len(notifyTable) > 1 {
		// 按触发的编排信息去重
		//var newNotifyTable []*models.NotifyTable
		//existMap := make(map[string]int)
		//for _, v := range notifyTable {
		//	tmpKey := fmt.Sprintf("%s_%s_%s_%s", v.ProcCallbackName, v.ProcCallbackKey, v.CallbackUrl, v.CallbackParam)
		//	if _, b := existMap[tmpKey]; b {
		//		continue
		//	}
		//	newNotifyTable = append(newNotifyTable, v)
		//	existMap[tmpKey] = 1
		//}
		//if len(newNotifyTable) > 0 {
		//	notifyTable = newNotifyTable
		//}
	}
	if alarmObj.Status == "firing" {
		notifyObj := &models.NotifyTable{}
		for _, v := range notifyTable {
			if v.EndpointGroup != "" {
				notifyObj = v
				break
			}
		}
		if notifyObj.Guid == "" {
			notifyObj = notifyTable[0]
		}
		if notifyObj.ProcCallbackMode == models.AlarmNotifyManualMode && notifyObj.ProcCallbackKey != "" {
			if _, execErr := x.Exec("update alarm set notify_id=? where id=?", notifyObj.Guid, alarmObj.Id); execErr != nil {
				log.Logger.Error("update alarm table notify id fail", log.Int("alarmId", alarmObj.Id), log.Error(execErr))
			}
		}
	}
	for _, v := range notifyTable {
		notifyAction(v, alarmObj)
	}
}

func notifyAction(notify *models.NotifyTable, alarmObj *models.AlarmHandleObj) {
	log.Logger.Info("Start notify action", log.String("procCallKey", notify.ProcCallbackKey), log.String("notify", notify.Guid), log.Int("alarm", alarmObj.Id))
	// alarmMailEnable==Y
	var err, mailErr error
	if models.AlarmMailEnable {
		mailErr = notifyMailAction(notify, alarmObj)
		if mailErr != nil {
			log.Logger.Error("Notify mail fail", log.String("notifyGuid", notify.Guid), log.Error(mailErr))
		}
	}
	if notify.ProcCallbackMode != models.AlarmNotifyAutoMode {
		log.Logger.Info("notify proc callback mode is not auto,done", log.Int("alarmId", alarmObj.Id), log.String("notifyId", notify.Guid), log.String("mode", notify.ProcCallbackMode))
		return
	}
	if alarmObj.SPriority == "" {
		tmpAlarmRows, _ := x.QueryString("select s_priority from alarm where id=?", alarmObj.Id)
		if len(tmpAlarmRows) > 0 {
			alarmObj.SPriority = tmpAlarmRows[0]["s_priority"]
		}
	}
	for i := 0; i < 3; i++ {
		err = notifyEventAction(notify, alarmObj, true, "system")
		if err == nil {
			break
		} else {
			log.Logger.Error("Notify event fail", log.String("notifyGuid", notify.Guid), log.Int("try", i), log.Error(err))
		}
	}
	if err != nil {
		if models.AlarmMailEnable && mailErr == nil {
			log.Logger.Info("Event three times fail,but already send mail success ")
			return
		}
		// 如果上面编排触发失败又没发自身通知邮件，尝试发邮件通知
		err = notifyMailAction(notify, alarmObj)
		if err != nil {
			log.Logger.Error("Event three times fail,and notify mail fail", log.String("notifyGuid", notify.Guid), log.Error(err))
		} else {
			log.Logger.Info("Event three times fail,send mail success ")
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

func notifyEventAction(notify *models.NotifyTable, alarmObj *models.AlarmHandleObj, compareLevel bool, operator string) error {
	if compareLevel && !compareNotifyEventLevel(alarmObj.SPriority) {
		log.Logger.Info("notify event disable", log.String("level", alarmObj.SPriority), log.String("minLevel", models.Config().MonitorAlarmCallbackLevelMin))
		err := notifyMailAction(notify, alarmObj)
		return err
	}
	if notify.ProcCallbackKey == "" {
		if alarmObj.Status == "firing" {
			if models.FiringCallback != "" && models.FiringCallback != "default_firing_callback" {
				notify.ProcCallbackKey = models.FiringCallback
			}
		} else {
			if models.RecoverCallback != "" && models.RecoverCallback != "default_recover_callback" {
				notify.ProcCallbackKey = models.RecoverCallback
			}
		}
		if notify.ProcCallbackKey == "" {
			return fmt.Errorf("Notify:%s procCallbackKey is empty ", notify.Guid)
		}
	}
	var requestParam models.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("%d-%s-%d-%s", alarmObj.Id, alarmObj.Status, time.Now().Unix(), notify.Guid)
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "SYS_MONITOR"
	requestParam.OperationKey = notify.ProcCallbackKey
	requestParam.OperationData = fmt.Sprintf("%d-%s-%s-%s", alarmObj.Id, alarmObj.Status, notify.Guid, operator)
	requestParam.OperationUser = operator
	log.Logger.Info(fmt.Sprintf("new notify request data --> eventSeqNo:%s operationKey:%s operationData:%s", requestParam.EventSeqNo, requestParam.OperationKey, requestParam.OperationData))
	b, _ := json.Marshal(requestParam)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", models.CoreUrl), strings.NewReader(string(b)))
	request.Header.Set("Authorization", models.GetCoreToken())
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Logger.Error("Notify core event new request fail", log.Error(err))
		return err
	}
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Logger.Error("Notify core event ctxhttp request fail", log.Error(err))
		return err
	}
	resultBody, _ := ioutil.ReadAll(res.Body)
	var resultObj models.CoreNotifyResult
	err = json.Unmarshal(resultBody, &resultObj)
	res.Body.Close()
	if err != nil {
		log.Logger.Error("Notify core event unmarshal json body fail", log.Error(err))
		return err
	}
	log.Logger.Info("Notify core result", log.String("body", string(resultBody)))
	return nil
}

func getNotifyEventMessage(notifyGuid string, alarm models.AlarmTable) (result models.AlarmEntityObj) {
	notifyObj, err := getSimpleNotify(notifyGuid)
	if err != nil {
		log.Logger.Warn("getNotifyEventMessage fail", log.Error(err))
	} else {
		notifyObj = models.NotifyTable{}
	}
	result = models.AlarmEntityObj{}
	result.Subject, result.Content = getNotifyMessage(&models.AlarmHandleObj{AlarmTable: alarm})
	var roles []*models.RoleNewTable
	if notifyObj.ServiceGroup != "" {
		x.SQL("select guid,email from role_new where guid in (select `role` from service_group_role_rel where service_group=?)", notifyObj.ServiceGroup).Find(&roles)
	} else {
		x.SQL("select guid,email,phone from `role_new` where guid in (select `role` from notify_role_rel where notify=?)", notifyGuid).Find(&roles)
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
	if notify.ServiceGroup != "" {
		x.SQL("select guid,email from role_new where guid in (select `role` from service_group_role_rel where service_group=?)", notify.ServiceGroup).Find(&roles)
	} else {
		x.SQL("select guid,email from `role_new` where guid in (select `role` from notify_role_rel where notify=?)", notify.Guid).Find(&roles)
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
		log.Logger.Warn("notifyMailAction toAddress empty", log.String("notify", notify.Guid), log.StringList("roleList", roleList))
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
	mailSender := smtp.MailSender{SenderName: mailConfig.SenderName, SenderMail: mailConfig.SenderMail, AuthServer: mailConfig.AuthServer, AuthPassword: mailConfig.AuthPassword}
	if mailConfig.SSL == "Y" {
		mailSender.SSL = true
	}
	err = mailSender.Init()
	if err != nil {
		return err
	}
	subject, content := getNotifyMessage(alarmObj)
	return mailSender.Send(subject, content, toAddress)
}

func getNotifyMessage(alarmObj *models.AlarmHandleObj) (subject, content string) {
	subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", alarmObj.Status, alarmObj.SPriority, alarmObj.Endpoint, alarmObj.SMetric)
	content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s", alarmObj.Endpoint, alarmObj.Status, alarmObj.SMetric, alarmObj.StartValue, alarmObj.SCond, alarmObj.SLast, alarmObj.SPriority, alarmObj.Content, time.Now().Format(models.DatetimeFormat))
	return
}

func getRoleMail(roleList []string) (mailList []string) {
	if len(roleList) == 0 {
		return
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/roles/retrieve", models.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Logger.Error("Get core role key new request fail", log.Error(err))
		return
	}
	request.Header.Set("Authorization", models.GetCoreToken())
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Logger.Error("Get core role key ctxhttp request fail", log.Error(err))
		return
	}
	b, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var result models.CoreRoleDto
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Logger.Error("Get core role key json unmarshal result", log.Error(err))
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

func ImportAlarmStrategy(queryType, inputGuid string, param []*models.EndpointStrategyObj, operator string) (err error, metricNotFound []string) {
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
		tmpActions, tmpErr, tmpMetricNotFound := getAlarmStrategyImportActions(inputGuid, "", endpointGroupTable[0].MonitorType, nowTime, operator, param[0], metricMap)
		if tmpErr != nil {
			metricNotFound = tmpMetricNotFound
			err = tmpErr
			return
		}
		actions = append(actions, tmpActions...)
	} else if queryType == "service" {
		var endpointGroupTable []*models.EndpointGroupTable
		err = x.SQL("select guid,monitor_type,service_group from endpoint_group where service_group=?", inputGuid).Find(&endpointGroupTable)
		if err != nil {
			err = fmt.Errorf("query endpoint group table fail,%s ", err.Error())
			return
		}
		for _, v := range param {
			tmpEndpointGroupExistFlag := false
			tmpMonitorType := ""
			for _, vv := range endpointGroupTable {
				if v.EndpointGroup == vv.Guid {
					tmpEndpointGroupExistFlag = true
					tmpMonitorType = vv.MonitorType
					break
				}
			}
			if !tmpEndpointGroupExistFlag {
				continue
			}
			endpointGroupList = append(endpointGroupList, v.EndpointGroup)
			tmpActions, tmpErr, tmpMetricNotFound := getAlarmStrategyImportActions(v.EndpointGroup, inputGuid, tmpMonitorType, nowTime, operator, v, metricMap)
			if tmpErr != nil {
				metricNotFound = tmpMetricNotFound
				err = fmt.Errorf("handle endpointGroup:%s fail,%s ", v.EndpointGroup, tmpErr.Error())
				break
			}
			actions = append(actions, tmpActions...)
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

func getAlarmStrategyImportActions(endpointGroup, serviceGroup, monitorType, nowTime, operator string, param *models.EndpointStrategyObj, metricMap map[string]*models.MetricTable) (actions []*Action, err error, metricNotFound []string) {
	var existStrategyTable []*models.AlarmStrategyTable
	err = x.SQL("select guid,metric from alarm_strategy where endpoint_group=?", endpointGroup).Find(&existStrategyTable)
	if err != nil {
		err = fmt.Errorf("query alarm strategy table fail,%s ", err.Error())
		return
	}
	//existStrategyMap := make(map[string]int)
	for _, v := range existStrategyTable {
		actions = append(actions, getNotifyListDeleteAction(v.Guid, "", "")...)
		actions = append(actions, getStrategyConditionDeleteAction(v.Guid)...)
		actions = append(actions, &Action{Sql: "delete from alarm_strategy where guid=?", Param: []interface{}{v.Guid}})
		//existStrategyMap[v.Guid] = 1
	}
	for _, strategy := range param.Strategy {
		// 检测策略上的指标在不在
		if len(strategy.Conditions) > 0 {
			for _, v := range strategy.Conditions {
				if fMetric, b := metricMap[v.MetricName]; b {
					v.Metric = fMetric.Guid
				} else {
					metricNotFound = append(metricNotFound, v.MetricName)
					err = fmt.Errorf("Metric:%s not found ", v.MetricName)
					break
				}
			}
		} else {
			if fMetric, b := metricMap[strategy.MetricName]; b {
				strategy.Metric = fMetric.Guid
			} else {
				metricNotFound = append(metricNotFound, strategy.MetricName)
				err = fmt.Errorf("Metric:%s not found ", strategy.MetricName)
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
		//if fMetric, b := metricMap[strategy.Metric]; b {
		//	if fMetric.MonitorType != monitorType {
		//		err = fmt.Errorf("Metric:%s is in type:%s ", strategy.Metric, fMetric.MonitorType)
		//		break
		//	}
		//	if serviceGroup != "" {
		//		if fMetric.ServiceGroup != serviceGroup {
		//			err = fmt.Errorf("Metric:%s is in serviceGroup:%s ", strategy.Metric, fMetric.ServiceGroup)
		//			break
		//		}
		//	}
		//} else {
		//	err = fmt.Errorf("Metric:%s is not exist ", strategy.Metric)
		//	break
		//}
		//if _, b := existStrategyMap[strategy.Guid]; b {
		//	updateAction := Action{Sql: "update alarm_strategy set metric=?,`condition`=?,`last`=?,priority=?,content=?,notify_enable=?,notify_delay_second=?,update_time=? where guid=?"}
		//	updateAction.Param = []interface{}{strategy.Metric, strategy.Condition, strategy.Last, strategy.Priority, strategy.Content, strategy.NotifyEnable, strategy.NotifyDelaySecond, nowTime, strategy.Guid}
		//	actions = append(actions, &updateAction)
		//} else {
		//	insertAction := Action{Sql: "insert into alarm_strategy(guid,endpoint_group,metric,`condition`,`last`,priority,content,notify_enable,notify_delay_second,update_time) value (?,?,?,?,?,?,?,?,?,?)"}
		//	insertAction.Param = []interface{}{strategy.Guid, strategy.EndpointGroup, strategy.Metric, strategy.Condition, strategy.Last, strategy.Priority, strategy.Content, strategy.NotifyEnable, strategy.NotifyDelaySecond, nowTime}
		//	actions = append(actions, &insertAction)
		//}
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
