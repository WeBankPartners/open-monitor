package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetDbMetricByServiceGroup(serviceGroup, metricKey string) (result []*models.DbMetricMonitorObj, err error) {
	result = []*models.DbMetricMonitorObj{}
	var dbMetricTable []*models.DbMetricMonitorTable
	if metricKey != "" {
		err = x.SQL("select * from db_metric_monitor where service_group=? and metric like '%"+metricKey+"%' order by update_time desc", serviceGroup).Find(&dbMetricTable)
	} else {
		err = x.SQL("select * from db_metric_monitor where service_group=? order by update_time desc", serviceGroup).Find(&dbMetricTable)
	}
	if err != nil {
		return result, fmt.Errorf("Query db_metric_monitor table fail,%s ", err.Error())
	}
	for _, v := range dbMetricTable {
		result = append(result, &models.DbMetricMonitorObj{Guid: v.Guid, ServiceGroup: v.ServiceGroup, MetricSql: v.MetricSql,
			Metric: v.Metric, DisplayName: v.DisplayName, Step: v.Step, MonitorType: v.MonitorType,
			EndpointRel: getDbMetricEndpointRel(v.Guid), UpdateUser: v.UpdateUser, UpdateTime: v.UpdateTime,
		})
	}
	return
}

func QueryDbMetricWithServiceGroup(serviceGroup, metricKey string) (result *models.DbMetricQueryObj, err error) {
	serviceGroupObj, getServiceGroupErr := getSimpleServiceGroup(serviceGroup)
	if getServiceGroupErr != nil {
		err = getServiceGroupErr
		return
	}
	result = &models.DbMetricQueryObj{ServiceGroupTable: serviceGroupObj}
	result.Config, err = GetDbMetricByServiceGroup(serviceGroup, metricKey)
	return
}

func GetDbMetricByEndpoint(endpointGuid, metricKey string) (result []*models.DbMetricQueryObj, err error) {
	result = []*models.DbMetricQueryObj{}
	var serviceGroupTable []*models.ServiceGroupTable
	err = x.SQL("select distinct t3.* from db_metric_endpoint_rel t1 left join db_metric_monitor t2 on t1.db_metric_monitor=t2.guid left join service_group t3 on t2.service_group=t3.guid where t1.source_endpoint=? or t1.target_endpoint=?", endpointGuid, endpointGuid).Find(&serviceGroupTable)
	if err != nil {
		return result, fmt.Errorf("Query database fail,%s ", err.Error())
	}
	for _, v := range serviceGroupTable {
		tmpResult, tmpErr := GetDbMetricByServiceGroup(v.Guid, metricKey)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		resultObj := models.DbMetricQueryObj{ServiceGroupTable: *v, Config: tmpResult}
		result = append(result, &resultObj)
	}
	return
}

func GetDbMetric(dbMetricGuid string) (result models.DbMetricMonitorObj, err error) {
	var dbMetricTable []*models.DbMetricMonitorTable
	err = x.SQL("select * from db_metric_monitor where guid=?", dbMetricGuid).Find(&dbMetricTable)
	if err != nil {
		return result, fmt.Errorf("Query db_metric_monitor table fail,%s ", err.Error())
	}
	if len(dbMetricTable) == 0 {
		return result, fmt.Errorf("Can not find db_metric_monitor with guid:%s ", dbMetricGuid)
	}
	result = models.DbMetricMonitorObj{Guid: dbMetricTable[0].Guid, ServiceGroup: dbMetricTable[0].ServiceGroup, MetricSql: dbMetricTable[0].MetricSql, Metric: dbMetricTable[0].Metric, DisplayName: dbMetricTable[0].DisplayName, Step: dbMetricTable[0].Step, MonitorType: dbMetricTable[0].MonitorType}
	result.EndpointRel = getDbMetricEndpointRel(dbMetricGuid)
	return
}

func CreateDbMetric(param *models.DbMetricMonitorObj, operator string) error {
	if param.Step < 10 {
		param.Step = 10
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	actions := getCreateDBMetricActions(param, operator, nowTime)
	return Transaction(actions)
}

func getCreateDBMetricActions(param *models.DbMetricMonitorObj, operator, nowTime string) (actions []*Action) {
	param.Guid = "dbm_" + guid.CreateGuid()
	insertAction := Action{Sql: "insert into db_metric_monitor(guid,service_group,metric_sql,metric,display_name,step,monitor_type,update_time,update_user) value (?,?,?,?,?,?,?,?,?)"}
	insertAction.Param = []interface{}{param.Guid, param.ServiceGroup, param.MetricSql, param.Metric, param.DisplayName, param.Step, param.MonitorType, nowTime, operator}
	actions = append(actions, &insertAction)
	actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user,db_metric_monitor) value (?,?,?,?,?,?,?,?,?,?,?)",
		Param: []interface{}{fmt.Sprintf("%s__%s", param.Metric, param.ServiceGroup), param.Metric, param.MonitorType, getDbMetricExpr(param.Metric, param.ServiceGroup), param.ServiceGroup,
			models.MetricWorkspaceService, nowTime, nowTime, operator, operator, param.Guid}})
	guidList := guid.CreateGuidList(len(param.EndpointRel))
	for i, v := range param.EndpointRel {
		if v.TargetEndpoint == "" {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into db_metric_endpoint_rel(guid,db_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guidList[i], param.Guid, v.SourceEndpoint, v.TargetEndpoint}})
	}
	return
}

func getDbMetricExpr(metric, serviceGroup string) (result string) {
	result = fmt.Sprintf("%s{key=\"%s\",service_group=\"%s\"}", models.DBMonitorMetricName, metric, serviceGroup)
	return result
}

func UpdateDbMetric(param *models.DbMetricMonitorObj, operator string) error {
	if param.Step < 10 {
		param.Step = 10
	}
	var dbMetricTable []*models.DbMetricMonitorTable
	x.SQL("select * from db_metric_monitor where guid=?", param.Guid).Find(&dbMetricTable)
	if len(dbMetricTable) == 0 {
		return fmt.Errorf("Can not find db_metric_monitor with guid:%s ", param.Guid)
	}
	var affectEndpointGroup []string
	var actions []*Action
	updateAction := Action{Sql: "update db_metric_monitor set metric_sql=?,metric=?,display_name=?,step=?,monitor_type=?,update_time=?,update_user=? where guid=?"}
	updateAction.Param = []interface{}{param.MetricSql, param.Metric, param.DisplayName, param.Step, param.MonitorType, time.Now().Format(models.DatetimeFormat), operator, param.Guid}
	actions = append(actions, &updateAction)
	if dbMetricTable[0].Metric != param.Metric {
		oldMetricGuid := fmt.Sprintf("%s__%s", dbMetricTable[0].Metric, dbMetricTable[0].ServiceGroup)
		newMetricGuid := fmt.Sprintf("%s__%s", param.Metric, dbMetricTable[0].ServiceGroup)
		actions = append(actions, &Action{Sql: "update metric set guid=?,metric=?,monitor_type=?,prom_expr=?,update_user=?, update_time = ? where guid=?",
			Param: []interface{}{newMetricGuid, param.Metric, param.MonitorType, getDbMetricExpr(param.Metric, dbMetricTable[0].ServiceGroup), operator, time.Now().Format(models.DatetimeFormat), oldMetricGuid}})
		var alarmStrategyTable []*models.AlarmStrategyTable
		x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", oldMetricGuid).Find(&alarmStrategyTable)
		if len(alarmStrategyTable) > 0 {
			for _, v := range alarmStrategyTable {
				affectEndpointGroup = append(affectEndpointGroup, v.EndpointGroup)
			}
			actions = append(actions, &Action{Sql: "update alarm_strategy set metric=? where metric=?", Param: []interface{}{newMetricGuid, oldMetricGuid}})
		}
	}
	actions = append(actions, &Action{Sql: "delete from db_metric_endpoint_rel where db_metric_monitor=?", Param: []interface{}{param.Guid}})
	guidList := guid.CreateGuidList(len(param.EndpointRel))
	for i, v := range param.EndpointRel {
		if v.TargetEndpoint == "" {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into db_metric_endpoint_rel(guid,db_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)", Param: []interface{}{guidList[i], param.Guid, v.SourceEndpoint, v.TargetEndpoint}})
	}
	err := Transaction(actions)
	if err == nil && len(affectEndpointGroup) > 0 {
		for _, v := range affectEndpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return err
}

func DeleteDbMetric(dbMetricGuid string) (err error) {
	var actions, endpointGroup = GetDeleteDbMetricActions(dbMetricGuid)
	if err = Transaction(actions); err != nil {
		return
	}
	if len(endpointGroup) > 0 {
		for _, v := range endpointGroup {
			SyncPrometheusRuleFile(v, false)
		}
	}
	return
}

func GetDeleteDbMetricActions(dbMetricGuid string) (actions []*Action, endpointGroup []string) {
	actions = []*Action{}
	endpointGroup = []string{}
	var dbMetricTable []*models.DbMetricMonitorTable
	var alarmStrategyTable []*models.AlarmStrategyTable
	x.SQL("select * from db_metric_monitor where guid=?", dbMetricGuid).Find(&dbMetricTable)
	if len(dbMetricTable) == 0 {
		return
	}
	alarmMetricGuid := fmt.Sprintf("%s__%s", dbMetricTable[0].Metric, dbMetricTable[0].ServiceGroup)
	x.SQL("select guid,endpoint_group from alarm_strategy where metric=?", alarmMetricGuid).Find(&alarmStrategyTable)

	actions = append(actions, &Action{Sql: "delete from db_metric_endpoint_rel where db_metric_monitor=?", Param: []interface{}{dbMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where metric=?", Param: []interface{}{alarmMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{alarmMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from db_metric_monitor where guid=?", Param: []interface{}{dbMetricGuid}})
	return
}

func getDbMetricEndpointRel(dbMetricMonitorGuid string) (result []*models.DbMetricEndpointRelTable) {
	result = []*models.DbMetricEndpointRelTable{}
	x.SQL("select * from db_metric_endpoint_rel where db_metric_monitor=?", dbMetricMonitorGuid).Find(&result)
	return result
}

func SyncDbMetric(initFlag bool) error {
	var dbExportAddress string
	for _, v := range models.Config().Dependence {
		if v.Name == "db_data_exporter" {
			dbExportAddress = v.Server
			break
		}
	}
	if dbExportAddress == "" {
		return fmt.Errorf("Can not find db_data_exporter address ")
	}
	var dbMonitorQuery []*models.DbMetricMonitorQueryObj
	err := x.SQL("select distinct t1.*,t2.source_endpoint,t2.target_endpoint from db_metric_monitor t1 left join db_metric_endpoint_rel t2 on t1.guid=t2.db_metric_monitor").Find(&dbMonitorQuery)
	if err != nil {
		return fmt.Errorf("Query db_metric_monitor fail,%s ", err.Error())
	}
	var dbKeywordQuery []*models.DbKeywordMonitorQueryObj
	err = x.SQL("select distinct t1.guid,t1.service_group,t1.name,t1.query_sql,t1.step,t1.monitor_type,t2.source_endpoint,t2.target_endpoint from db_keyword_monitor t1 left join db_keyword_endpoint_rel t2 on t1.guid=t2.db_keyword_monitor").Find(&dbKeywordQuery)
	if err != nil {
		return fmt.Errorf("Query db_keyword_monitor fail,%s ", err.Error())
	}
	endpointGuidList := []string{}
	endpointExtMap := make(map[string]*models.EndpointExtendParamObj)
	for _, v := range dbMonitorQuery {
		endpointGuidList = append(endpointGuidList, v.SourceEndpoint)
		endpointExtMap[v.SourceEndpoint] = &models.EndpointExtendParamObj{}
	}
	for _, row := range dbKeywordQuery {
		endpointGuidList = append(endpointGuidList, row.SourceEndpoint)
		endpointExtMap[row.SourceEndpoint] = &models.EndpointExtendParamObj{}
	}
	var endpointTable []*models.EndpointNewTable
	x.SQL("select guid,endpoint_address,extend_param from endpoint_new where monitor_type='mysql' and guid in ('" + strings.Join(endpointGuidList, "','") + "')").Find(&endpointTable)
	for _, v := range endpointTable {
		if v.ExtendParam == "" {
			continue
		}
		tmpExtObj := models.EndpointExtendParamObj{}
		tmpErr := json.Unmarshal([]byte(v.ExtendParam), &tmpExtObj)
		if tmpErr != nil {
			continue
		}
		endpointExtMap[v.Guid] = &tmpExtObj
	}
	var postData []*models.DbMonitorTaskObj
	for _, v := range dbMonitorQuery {
		if extConfig, b := endpointExtMap[v.SourceEndpoint]; b {
			taskObj := models.DbMonitorTaskObj{DbType: "mysql", Name: v.Metric, Step: v.Step, Sql: v.MetricSql, Server: extConfig.Ip, Port: extConfig.Port, User: extConfig.User, Password: extConfig.Password, Endpoint: v.SourceEndpoint, ServiceGroup: v.ServiceGroup}
			if v.TargetEndpoint != "" {
				taskObj.Endpoint = v.TargetEndpoint
			}
			postData = append(postData, &taskObj)
		}
	}
	for _, v := range dbKeywordQuery {
		if extConfig, b := endpointExtMap[v.SourceEndpoint]; b {
			taskObj := models.DbMonitorTaskObj{DbType: "mysql", Name: "db_keyword_value", Step: v.Step, Sql: v.QuerySql, Server: extConfig.Ip, Port: extConfig.Port, User: extConfig.User, Password: extConfig.Password, Endpoint: v.SourceEndpoint, ServiceGroup: v.ServiceGroup, KeywordGuid: v.Guid}
			if v.TargetEndpoint != "" {
				taskObj.Endpoint = v.TargetEndpoint
			}
			postData = append(postData, &taskObj)
		}
	}
	if initFlag {
		var alarmTable []*models.DbKeywordAlarm
		err = x.SQL("select * from db_keyword_alarm").Find(&alarmTable)
		if err != nil {
			log.Logger.Error("init db keyword warning with query exist alarm fail", log.Error(err))
		} else {
			keywordCountMap := make(map[string]float64)
			for _, row := range alarmTable {
				keywordCountKey := fmt.Sprintf("%s^^^%s", row.DbKeywordMonitor, row.Endpoint)
				if existCount, ok := keywordCountMap[keywordCountKey]; ok {
					if existCount < row.EndValue {
						keywordCountMap[keywordCountKey] = row.EndValue
					}
				} else {
					keywordCountMap[keywordCountKey] = row.EndValue
				}
			}
			for _, v := range postData {
				if findCount, ok := keywordCountMap[fmt.Sprintf("%s^^^%s", v.KeywordGuid, v.Endpoint)]; ok {
					v.KeywordCount = int64(findCount)
				}
			}
		}
	}
	postDataByte, _ := json.Marshal(postData)
	log.Logger.Info("Sync db metric", log.String("postData", string(postDataByte)))
	resp, err := http.Post(fmt.Sprintf("%s/db/config", dbExportAddress), "application/json", strings.NewReader(string(postDataByte)))
	if err != nil {
		return fmt.Errorf("Http request to %s/db/config fail,%s ", dbExportAddress, err.Error())
	}
	bodyByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode > 300 {
		return fmt.Errorf("%s", string(bodyByte))
	}
	return nil
}
