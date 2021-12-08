package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/WeBankPartners/go-common-lib/smtp"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func QueryAlarmStrategyByGroup(endpointGroup string) (result []*models.EndpointStrategyObj, err error) {
	result = []*models.EndpointStrategyObj{}
	var strategy []*models.GroupStrategyObj
	var alarmStrategyTable []*models.AlarmStrategyMetricObj
	err = x.SQL("select t1.*,t2.metric as 'metric_name' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where t1.endpoint_group=?", endpointGroup).Find(&alarmStrategyTable)
	if err != nil {
		return
	}
	for _, v := range alarmStrategyTable {
		tmpStrategyObj := models.GroupStrategyObj{Guid: v.Guid, EndpointGroup: v.EndpointGroup, Metric: v.Metric, MetricName: v.MetricName, Condition: v.Condition, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelaySecond: v.NotifyDelaySecond}
		tmpStrategyObj.NotifyList = getNotifyList(v.Guid, "", "")
		strategy = append(strategy, &tmpStrategyObj)
	}
	resultObj := models.EndpointStrategyObj{EndpointGroup: endpointGroup, Strategy: strategy}
	notify,tmpErr := GetGroupEndpointNotify(endpointGroup)
	if tmpErr != nil {
		return result,tmpErr
	}
	resultObj.NotifyList = notify
	result = append(result, &resultObj)
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
		tmpEndpointStrategyList, tmpErr := QueryAlarmStrategyByGroup(v.Guid)
		if tmpErr != nil || len(tmpEndpointStrategyList) == 0 {
			err = tmpErr
			break
		}
		tmpEndpointStrategyList[0].ServiceGroup = v.ServiceGroup
		result = append(result, tmpEndpointStrategyList[0])
	}
	return
}

func GetAlarmStrategy(strategyGuid string) (result models.AlarmStrategyMetricObj,err error) {
	var strategyTable []*models.AlarmStrategyMetricObj
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where t1.guid=?", strategyGuid).Find(&strategyTable)
	if err != nil {
		return result,fmt.Errorf("Query alarm_strategy fail,%s ", err.Error())
	}
	if len(strategyTable) == 0 {
		return result,fmt.Errorf("Can not find alarm_strategy with guid:%s ", strategyGuid)
	}
	result = *strategyTable[0]
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

func DeleteAlarmStrategy(strategyGuid string) (endpointGroup string,err error) {
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
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where guid=?", Param: []interface{}{strategyGuid}})
	err = Transaction(actions)
	return
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

func SyncPrometheusRuleFile(endpointGroup string,fromPeer bool) error {
	if endpointGroup == "" {
		return fmt.Errorf("Sync prometheus rule fail,group is empty ")
	}
	var err error
	ruleFileName := "g_" + endpointGroup
	var endpointList []*models.EndpointNewTable
	err = x.SQL("select * from endpoint_new where guid in (select endpoint from endpoint_group_rel where endpoint_group=? union select endpoint from endpoint_service_rel where service_group in (select service_group from endpoint_group where guid=?))", endpointGroup, endpointGroup).Find(&endpointList)
	if err != nil {
		return err
	}
	// 获取strategy
	strategyList,getStrategyErr := getAlarmStrategyWithExpr(endpointGroup)
	if getStrategyErr != nil {
		return getStrategyErr
	}
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
	}else{
		clusterEndpointMap["default"] = []*models.EndpointNewTable{}
	}
	for _,cluster := range clusterList {
		guidExpr,addressExpr,ipExpr := buildRuleReplaceExprNew(clusterEndpointMap[cluster])
		ruleFileConfig := buildRuleFileContentNew(ruleFileName,guidExpr,addressExpr,ipExpr,copyStrategyListNew(strategyList))
		if cluster == "default" || cluster == "" {
			prom.SyncLocalRuleConfig(models.RuleLocalConfigJob{FromPeer: fromPeer,EndpointGroup: endpointGroup,Name: ruleFileConfig.Name,Rules: ruleFileConfig.Rules})
		}else{
			tmpErr := SyncRemoteRuleConfigFile(cluster, models.RFClusterRequestObj{Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
			if tmpErr != nil {
				err = fmt.Errorf("Update remote cluster:%s rule file fail,%s ", cluster, tmpErr.Error())
				log.Logger.Error("Update remote cluster rule file fail", log.String("cluster",cluster), log.Error(tmpErr))
			}
		}
	}
	return err
}

func getAlarmStrategyWithExpr(endpointGroup string) (result []*models.AlarmStrategyMetricObj,err error) {
	result = []*models.AlarmStrategyMetricObj{}
	err = x.SQL("select t1.*,t2.metric as 'metric_name',t2.prom_expr as 'metric_expr',t2.monitor_type as 'metric_type' from alarm_strategy t1 left join metric t2 on t1.metric=t2.guid where endpoint_group=?",endpointGroup).Find(&result)
	return
}

func buildRuleReplaceExprNew(endpointList []*models.EndpointNewTable) (guidExpr,addressExpr,ipExpr string) {
	for _,endpoint := range endpointList {
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

func buildRuleFileContentNew(ruleFileName,guidExpr,addressExpr,ipExpr string,strategyList []*models.AlarmStrategyMetricObj) models.RFGroup {
	result := models.RFGroup{Name: ruleFileName}
	if len(strategyList) == 0 {
		return result
	}
	for _,strategy := range strategyList {
		tmpRfu := models.RFRule{}
		tmpRfu.Alert = fmt.Sprintf("%s_%s", strategy.Metric, strategy.Guid)
		if !strings.Contains(strategy.Condition, " ") && strategy.Condition != "" {
			if strings.Contains(strategy.Condition, "=") {
				strategy.Condition = strategy.Condition[:2] + " " + strategy.Condition[2:]
			}else{
				strategy.Condition = strategy.Condition[:1] + " " + strategy.Condition[1:]
			}
		}
		if strings.Contains(strategy.MetricExpr, "$address") {
			if strings.Contains(addressExpr, "|") {
				strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$address\"", "=~\""+addressExpr+"\"", -1)
			}else{
				strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$address\"", "=\""+addressExpr+"\"", -1)
			}
		}
		if strings.Contains(strategy.MetricExpr, "$guid") {
			if strings.Contains(guidExpr, "|") {
				strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$guid\"", "=~\""+guidExpr+"\"", -1)
			}else{
				strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$guid\"", "=\""+guidExpr+"\"", -1)
			}
		}
		if strings.Contains(strategy.MetricExpr, "$ip") {
			if strings.Contains(ipExpr, "|") {
				tmpStr := strings.Split(strategy.MetricExpr, "$ip")[1]
				tmpStr = tmpStr[:strings.Index(tmpStr,"\"")]
				newList := []string{}
				for _,v := range strings.Split(ipExpr, "|") {
					newList = append(newList, v+tmpStr)
				}
				strategy.MetricExpr = strings.Replace(strategy.MetricExpr, "=\"$ip"+tmpStr+"\"", "=~\""+strings.Join(newList, "|")+"\"", -1)
			}else{
				strategy.MetricExpr = strings.ReplaceAll(strategy.MetricExpr, "$ip", ipExpr)
			}
		}
		tmpRfu.Expr = fmt.Sprintf("%s %s", strategy.MetricExpr, strategy.Condition)
		tmpRfu.For = strategy.Last
		tmpRfu.Labels = make(map[string]string)
		tmpRfu.Labels["strategy_guid"] = strategy.Guid
		tmpRfu.Annotations = models.RFAnnotation{Summary:fmt.Sprintf("{{$labels.instance}}__%s__%s__{{$value}}", strategy.Priority, strategy.Metric), Description:strategy.Content}
		result.Rules = append(result.Rules, &tmpRfu)
	}
	return result
}

func copyStrategyListNew(inputs []*models.AlarmStrategyMetricObj) (result []*models.AlarmStrategyMetricObj) {
	result = []*models.AlarmStrategyMetricObj{}
	for _,strategy := range inputs {
		tmpStrategy := models.AlarmStrategyMetricObj{Guid: strategy.Guid,Metric:strategy.Metric,Condition: strategy.Condition,Last: strategy.Last,Priority: strategy.Priority,Content: strategy.Content,NotifyEnable: strategy.NotifyEnable,NotifyDelaySecond: strategy.NotifyDelaySecond,MetricName: strategy.MetricName,MetricExpr: strategy.MetricExpr,MetricType: strategy.MetricType}
		result = append(result, &tmpStrategy)
	}
	return result
}

func GetAlarmObj(query *models.AlarmTable) (result models.AlarmTable,err error) {
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
	err = x.SQL(baseSql, queryParams...).Find(&alarmList)
	if len(alarmList) > 0 {
		result = *alarmList[0]
	}
	return
}

func NotifyStrategyAlarm(alarmObj *models.AlarmHandleObj)  {
	if alarmObj.AlarmStrategy == "" {
		log.Logger.Error("Notify strategy alarm fail,alarmStrategy is empty", log.JsonObj("alarm", alarmObj))
		return
	}
	var notifyTable []*models.NotifyTable
	err := x.SQL("select * from notify where alarm_strategy=?", alarmObj.AlarmStrategy).Find(&notifyTable)
	if err != nil {
		log.Logger.Error("Query notify table fail", log.Error(err))
		return
	}
	if len(notifyTable) == 0 {
		x.SQL("select * from notify where endpoint_group in (select endpoint_group from alarm_strategy where guid=?) or service_group in (select service_group from endpoint_service_rel where endpoint=?)", alarmObj.AlarmStrategy, alarmObj.Endpoint).Find(&notifyTable)
	}
	if len(notifyTable) == 0 {
		return
	}
	for _,v := range notifyTable {
		err = notifyAction(v,alarmObj)
		if err != nil {
			log.Logger.Error("Notify mail fail", log.String("notifyGuid", v.Guid), log.Error(err))
		}
	}
}

func notifyAction(notify *models.NotifyTable,alarmObj *models.AlarmHandleObj) error {
	if notify.ProcCallbackKey == "" {
		return notifyMailAction(notify,alarmObj)
	}
	var err error
	for i:=0;i<3;i++ {
		err = notifyEventAction(notify,alarmObj)
		if err == nil {
			break
		}else{
			log.Logger.Error("Notify event fail", log.String("notifyGuid", notify.Guid), log.Int("try",i), log.Error(err))
		}
	}
	if err != nil {
		return notifyMailAction(notify,alarmObj)
	}
	return nil
}

func notifyEventAction(notify *models.NotifyTable,alarmObj *models.AlarmHandleObj) error {
	if notify.ProcCallbackKey == "" {
		return fmt.Errorf("Notify:%s procCallbackKey is empty ", notify.Guid)
	}
	var requestParam models.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("%d-%s-%d-%s", alarmObj.Id, alarmObj.Status, time.Now().Unix(), notify.Guid)
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "SYS_MONITOR"
	requestParam.OperationKey = notify.ProcCallbackKey
	requestParam.OperationData = fmt.Sprintf("%d-%s", alarmObj.Id, notify.Guid)
	requestParam.OperationUser = ""
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
	return nil
}

func getNotifyEventMessage(notifyGuid string,alarm models.AlarmTable) (result models.AlarmEntityObj) {
	result = models.AlarmEntityObj{}
	result.Subject,result.Content = getNotifyMessage(&models.AlarmHandleObj{AlarmTable:alarm})
	var roles []*models.RoleNewTable
	x.SQL("select guid,email,phone from `role_new` where guid in (select `role` from notify_role_rel where notify=?)", notifyGuid).Find(&roles)
	var email,phone,role []string
	emailExistMap := make(map[string]int)
	phoneExistMap := make(map[string]int)
	for _,v := range roles {
		if v.Email != "" {
			if _,b:=emailExistMap[v.Email];!b {
				email = append(email, v.Email)
				emailExistMap[v.Email] = 1
			}
		}
		if v.Phone != "" {
			if _,b:=phoneExistMap[v.Phone];!b {
				phone = append(phone, v.Phone)
				phoneExistMap[v.Phone] = 1
			}
		}
		role = append(role, v.Guid)
	}
	result.To = strings.Join(email, ",")
	result.ToMail = result.To
	result.ToPhone = strings.Join(phone, ",")
	result.ToRole = strings.Join(role, ",")
	result.SmsContent = getSmsAlarmContent(&alarm)
	return result
}

func notifyMailAction(notify *models.NotifyTable,alarmObj *models.AlarmHandleObj) error {
	var roles []*models.RoleNewTable
	x.SQL("select distinct email from `role_new` where guid in (select `role` from notify_role_rel where notify=?)", notify.Guid).Find(&roles)
	toAddress := []string{}
	for _,v := range roles {
		if v.Email != "" {
			toAddress = append(toAddress, v.Email)
		}
	}
	if len(toAddress) == 0 {
		return nil
	}
	mailConfig,err := GetSysAlertMailConfig()
	if err != nil {
		return err
	}
	mailSender := smtp.MailSender{SenderName: mailConfig.SenderName,SenderMail: mailConfig.SenderMail,AuthServer: mailConfig.AuthServer,AuthPassword: mailConfig.AuthPassword}
	if mailConfig.SSL == "Y" {
		mailSender.SSL = true
	}
	err = mailSender.Init()
	if err != nil {
		return err
	}
	subject,content := getNotifyMessage(alarmObj)
	return mailSender.Send(subject, content, toAddress)
}

func getNotifyMessage(alarmObj *models.AlarmHandleObj) (subject,content string) {
	subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", alarmObj.Status, alarmObj.SPriority, alarmObj.Endpoint, alarmObj.SMetric)
	content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s",alarmObj.Endpoint,alarmObj.Status,alarmObj.SMetric,alarmObj.StartValue,alarmObj.SCond,alarmObj.SLast,alarmObj.SPriority,alarmObj.Content,time.Now().Format(models.DatetimeFormat))
	return
}