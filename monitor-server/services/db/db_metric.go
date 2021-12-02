package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func GetDbMetricByServiceGroup(serviceGroup string) (result []*models.DbMetricMonitorObj,err error) {
	result = []*models.DbMetricMonitorObj{}
	var dbMetricTable []*models.DbMetricMonitorTable
	err = x.SQL("select * from db_metric_monitor where service_group=?", serviceGroup).Find(&dbMetricTable)
	if err != nil {
		return result,fmt.Errorf("Query db_metric_monitor table fail,%s ", err.Error())
	}
	for _,v := range dbMetricTable {
		result = append(result, &models.DbMetricMonitorObj{Guid: v.Guid,ServiceGroup: v.ServiceGroup,MetricSql: v.MetricSql,Metric: v.Metric,DisplayName: v.DisplayName,Step: v.Step,MonitorType: v.MonitorType,EndpointRel: getDbMetricEndpointRel(v.Guid)})
	}
	return
}

func GetDbMetricByEndpoint(endpointGuid string) (result []*models.DbMetricMonitorObj,err error) {
	result = []*models.DbMetricMonitorObj{}
	var serviceGroupTable []*models.ServiceGroupTable
	err = x.SQL("select distinct t3.* from db_metric_endpoint_rel t1 left join db_metric_monitor t2 on t1.db_metric_monitor=t2.guid left join service_group t3 on t2.service_group=t3.guid where t1.source_endpoint=?",endpointGuid).Find(&serviceGroupTable)
	if err != nil {
		return result,fmt.Errorf("Query database fail,%s ", err.Error())
	}
	for _,v := range serviceGroupTable {
		tmpResult,tmpErr := GetDbMetricByServiceGroup(v.Guid)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		for _,vv := range tmpResult {
			vv.ServiceGroupName = v.DisplayName
		}
		result = append(result, tmpResult...)
	}
	return
}

func GetDbMetric(dbMetricGuid string) (result models.DbMetricMonitorObj,err error) {
	var dbMetricTable []*models.DbMetricMonitorTable
	err = x.SQL("select * from db_metric_monitor where guid=?", dbMetricGuid).Find(&dbMetricTable)
	if err != nil {
		return result,fmt.Errorf("Query db_metric_monitor table fail,%s ", err.Error())
	}
	if len(dbMetricTable) == 0 {
		return result,fmt.Errorf("Can not find db_metric_monitor with guid:%s ", dbMetricGuid)
	}
	result = models.DbMetricMonitorObj{Guid: dbMetricTable[0].Guid,ServiceGroup: dbMetricTable[0].ServiceGroup,MetricSql: dbMetricTable[0].MetricSql,Metric: dbMetricTable[0].Metric,DisplayName: dbMetricTable[0].DisplayName,Step: dbMetricTable[0].Step,MonitorType: dbMetricTable[0].MonitorType}
	result.EndpointRel = getDbMetricEndpointRel(dbMetricGuid)
	return
}

func CreateDbMetric(param *models.DbMetricMonitorObj) error {
	nowTime := time.Now().Format(models.DatetimeFormat)
	param.Guid = guid.CreateGuid()
	var actions []*Action
	insertAction := Action{Sql: "insert into db_metric_monitor(guid,service_group,metric_sql,metric,display_name,step,monitor_type,update_time) value (?,?,?,?,?,?,?,?)"}
	insertAction.Param = []interface{}{param.Guid,param.ServiceGroup,param.MetricSql,param.Metric,param.DisplayName,param.Step,param.MonitorType,nowTime}
	actions = append(actions, &insertAction)
	guidList := guid.CreateGuidList(len(param.EndpointRel))
	for i,v := range param.EndpointRel {
		if v.TargetEndpoint == "" {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into db_metric_endpoint_rel(guid,db_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)",Param: []interface{}{guidList[i],param.Guid,v.SourceEndpoint,v.TargetEndpoint}})
	}
	return Transaction(actions)
}

func UpdateDbMetric(param *models.DbMetricMonitorObj) error {
	var actions []*Action
	updateAction := Action{Sql: "update db_metric_monitor set metric_sql=?,metric=?,display_name=?,step=?,monitor_type=?,update_time=? where guid=?"}
	updateAction.Param = []interface{}{param.MetricSql,param.Metric,param.DisplayName,param.Step,param.MonitorType,time.Now().Format(models.DatetimeFormat),param.Guid}
	actions = append(actions, &updateAction)
	actions = append(actions, &Action{Sql: "delete from db_metric_endpoint_rel where db_metric_monitor=?",Param: []interface{}{param.Guid}})
	guidList := guid.CreateGuidList(len(param.EndpointRel))
	for i,v := range param.EndpointRel {
		if v.TargetEndpoint == "" {
			continue
		}
		actions = append(actions, &Action{Sql: "insert into db_metric_endpoint_rel(guid,db_metric_monitor,source_endpoint,target_endpoint) value (?,?,?,?)",Param: []interface{}{guidList[i],param.Guid,v.SourceEndpoint,v.TargetEndpoint}})
	}
	return Transaction(actions)
}

func DeleteDbMetric(dbMetricGuid string) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from db_metric_endpoint_rel where db_metric_monitor=?",Param: []interface{}{dbMetricGuid}})
	actions = append(actions, &Action{Sql: "delete from db_metric_monitor where guid=?",Param: []interface{}{dbMetricGuid}})
	return Transaction(actions)
}

func getDbMetricEndpointRel(dbMetricMonitorGuid string) (result []*models.DbMetricEndpointRelTable) {
	result = []*models.DbMetricEndpointRelTable{}
	x.SQL("select * from db_metric_endpoint_rel where db_metric_monitor=?", dbMetricMonitorGuid).Find(&result)
	return result
}
