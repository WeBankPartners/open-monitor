package db

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var systemMetricList = []string{"cpu_detail_percent__host", "cpu_used_percent__host", "db_count_change__mysql", "db_monitor_count__mysql",
	"disk_iops__host", "disk_read_bytes__host", "disk_write_bytes__host", "file_handler_free_percent__host", "gc_marksweep_time__java",
	"heap_mem_used_percent__java", "http_status__http", "jvm_gc_time__java", "jvm_memory_heap_max__java", "jvm_memory_heap_used__java",
	"jvm_thread_count__java", "load_1min__host", "mem_total__host", "mem_used_percent__host", "mem_used__host", "mysql_alive__mysql",
	"mysql_buffer_status__mysql", "mysql_connect_used_percent__mysql", "mysql_requests__mysql", "mysql_threads_connected__mysql",
	"mysql_threads_max__mysql", "net_if_in_bytes__host", "net_if_out_bytes__host", "nginx_connect_active__nginx", "nginx_handle_request__nginx",
	"ping_alive__host", "ping_alive__ping", "ping_loss__host", "ping_loss__ping", "ping_time__host", "pod_cpu_used_percent__pod",
	"pod_mem_used_percent__pod", "process_alive_count__host", "process_alive_count__process", "process_cpu_used_percent__host",
	"process_cpu_used_percent__process", "process_mem_byte__host", "process_mem_byte__process", "redis_alive__redis", "redis_client_used_percent__redis",
	"redis_cmd_num__redis", "redis_db_keys__redis", "redis_expire_key__redis", "redis_mem_used__redis", "telnet_alive__host", "telnet_alive__telnet",
	"tomcat_connection__java", "tomcat_request__java", "volume_used_percent__host"}

func MetricList(id string, endpointType, serviceGroup string) (result []*models.PromMetricTable, err error) {
	params := []interface{}{}
	baseSql := "select guid as id,metric,monitor_type as metric_type,prom_expr as prom_ql from metric where 1=1 "
	if id != "" {
		baseSql += " and guid=? "
		params = append(params, id)
	}
	if endpointType != "" {
		baseSql += " and monitor_type=? "
		params = append(params, endpointType)
	}
	if serviceGroup != "" {
		baseSql += " and service_group=? "
		params = append(params, serviceGroup)
	}
	result = []*models.PromMetricTable{}
	err = x.SQL(baseSql, params...).Find(&result)
	return
}

func MetricCreate(param []*models.MetricTable, operator string, errMsgObj *models.ErrorMessageObj) error {
	var actions []*Action
	var metricTemp *models.MetricTable
	var metricList []*models.MetricTable
	var guid string
	var err error
	nowTime := time.Now().Format(models.DatetimeFormat)
	for _, metric := range param {
		if metric.ServiceGroup != "" {
			guid = fmt.Sprintf("%s__%s", metric.Metric, metric.MonitorType)
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{guid, metric.Metric, metric.MonitorType, metric.PromExpr, metric.ServiceGroup, metric.Workspace, nowTime, nowTime, operator, operator}})
		} else if metric.EndpointGroup != "" {
			var monitorType string
			x.SQL("select monitor_type from endpoint_group where guid=?", metric.EndpointGroup).Get(&monitorType)
			guid = fmt.Sprintf("%s__%s", metric.Metric, monitorType)
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,update_time,create_time,create_user,update_user,endpoint_group) value (?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{guid, metric.Metric, metric.MonitorType, metric.PromExpr, nowTime, nowTime, operator, operator, metric.EndpointGroup}})
		} else {
			guid = fmt.Sprintf("%s__%s", metric.Metric, metric.MonitorType)
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,update_time,create_time,create_user,update_user) value (?,?,?,?,?,?,?,?)",
				Param: []interface{}{guid, metric.Metric, metric.MonitorType, metric.PromExpr, nowTime, nowTime, operator, operator}})
		}
		if metricTemp, err = GetMetric(guid); err != nil {
			return err
		}
		if metricTemp != nil && metricTemp.Guid != "" {
			return fmt.Errorf(errMsgObj.AddMetricRepeatError)
		}
		// 同一个层级对象或者对象组里面指标名称重复,也需要校验
		if metricList, err = GetMetricByName(metric.Metric); err != nil {
			return err
		}
		if len(metricList) > 0 {
			for _, m := range metricList {
				if m.ServiceGroup != "" && m.ServiceGroup == metric.ServiceGroup {
					return fmt.Errorf(errMsgObj.AddMetricRepeatError)
				}
				if m.EndpointGroup != "" && m.EndpointGroup == metric.EndpointGroup {
					return fmt.Errorf(errMsgObj.AddMetricRepeatError)
				}
			}
		}
	}
	return Transaction(actions)
}

func MetricUpdate(param []*models.MetricTable, operator string) (err error) {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	var metricGuidList []string
	for _, metric := range param {
		if metric.Guid == "" {
			err = fmt.Errorf("Guid can not empty ")
			break
		}
		if metric.ServiceGroup != "" {
			actions = append(actions, &Action{Sql: "update metric set prom_expr=?,service_group=?,workspace=?,update_user=?,update_time=? where guid=?", Param: []interface{}{metric.PromExpr, metric.ServiceGroup, metric.Workspace, operator, nowTime, metric.Guid}})
			actions = append(actions, &Action{Sql: "update metric set service_group=?,workspace=?,update_user=?,update_time=? where guid in (select metric_id from metric_comparison where origin_metric_id=?)",
				Param: []interface{}{metric.ServiceGroup, metric.Workspace, operator, nowTime, metric.Guid}})
		} else {
			actions = append(actions, &Action{Sql: "update metric set prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{metric.PromExpr, operator, nowTime, metric.Guid}})
			actions = append(actions, &Action{Sql: "update metric set update_user=?,update_time=? where guid in (select metric_id from metric_comparison where origin_metric_id=?)",
				Param: []interface{}{operator, nowTime, metric.Guid}})
		}
		metricGuidList = append(metricGuidList, metric.Guid)
	}
	if err != nil {
		return err
	}
	err = Transaction(actions)
	if err != nil {
		return err
	}
	var alarmStrategyTable []*models.AlarmStrategyTable
	err = x.SQL("select distinct endpoint_group from alarm_strategy where metric in ('" + strings.Join(metricGuidList, "','") + "')").Find(&alarmStrategyTable)
	if len(alarmStrategyTable) > 0 {
		for _, v := range alarmStrategyTable {
			err = SyncPrometheusRuleFile(v.EndpointGroup, false)
			if err != nil {
				break
			}
		}
	}
	return err
}

func getMetricUpdateAction(oldGuid, operator string, newMetricObj *models.MetricTable) (actions []*Action) {
	actions = []*Action{}
	if newMetricObj.Guid != oldGuid {
		actions = append(actions, &Action{Sql: "update metric set guid=?,metric=?,monitor_type=?,prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{newMetricObj.Guid, newMetricObj.Metric, newMetricObj.MonitorType, newMetricObj.PromExpr, operator, newMetricObj.UpdateTime, oldGuid}})
		actions = append(actions, &Action{Sql: "update alarm_strategy set metric=? where metric=?", Param: []interface{}{newMetricObj.Guid, oldGuid}})
	} else {
		actions = append(actions, &Action{Sql: "update metric set metric=?,monitor_type=?,prom_expr=?,update_user=?,update_time=? where guid=?", Param: []interface{}{newMetricObj.Metric, newMetricObj.MonitorType, newMetricObj.PromExpr, operator, newMetricObj.UpdateTime, oldGuid}})
	}
	return actions
}

func MetricDeleteNew(id string) (withComparison bool, err error) {
	metricQuery, getErr := MetricList(id, "", "")
	if getErr != nil {
		err = fmt.Errorf("Try to query prom metric table fail,%s ", getErr.Error())
		return
	}
	if len(metricQuery) == 0 {
		return
	}
	metricRow := metricQuery[0]
	var actions []*Action
	var affectEndpointGroup []string
	// 删除同环比 指标
	tmpActions, tmpEndpointGroup := getMetricComparisonDeleteAction(id)
	if len(tmpActions) > 0 {
		withComparison = true
	}
	actions = append(actions, tmpActions...)
	affectEndpointGroup = append(tmpEndpointGroup, tmpEndpointGroup...)
	tmpActions, tmpEndpointGroup = getDeleteMetricActions(id)
	actions = append(actions, tmpActions...)
	affectEndpointGroup = append(tmpEndpointGroup, tmpEndpointGroup...)
	tmpActions = getDeleteEndpointDashboardChartMetricAction(metricRow.Metric, metricRow.MetricType)
	actions = append(actions, tmpActions...)
	err = Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update database fail,%s ", err.Error())
	} else {
		for _, v := range affectEndpointGroup {
			if tmpErr := SyncPrometheusRuleFile(v, false); tmpErr != nil {
				log.Logger.Error("sync prometheus rule file fail", log.Error(tmpErr))
			}
		}
	}
	return
}

func MetricDelete(id string) error {
	metricQuery, err := MetricList(id, "", "")
	if err != nil {
		return fmt.Errorf("Try to query prom metric table fail,%s ", err.Error())
	}
	if len(metricQuery) == 0 {
		return nil
	}
	metric := metricQuery[0].Metric
	var actions, subActions []*Action
	// 删除同环比 指标
	var metricIds []string
	if err = x.SQL("select metric_id from metric_comparison where origin_metric_id = ?", id).Find(&metricIds); err != nil {
		return fmt.Errorf("query metric_comparison fail,%s", err.Error())
	}

	actions = append(actions, &Action{Sql: "delete from metric_comparison where  metric_id = ?", Param: []interface{}{id}})
	actions = append(actions, &Action{Sql: "delete from metric_comparison where  origin_metric_id = ?", Param: []interface{}{id}})
	actions = append(actions, &Action{Sql: "delete from metric where guid in ('" + strings.Join(metricIds, "','") + "') "})
	actions = append(actions, &Action{Sql: "delete from metric where guid=?", Param: []interface{}{id}})
	// 删除看板里面引入的当前指标
	if subActions, err = DeleteCustomChartSeriesByMetricIdSQL(id); err != nil {
		return err
	}
	if len(subActions) > 0 {
		actions = append(actions, subActions...)
	}
	var charts []*models.ChartTable
	err = x.SQL("select id,metric from chart where metric like ? and group_id in (select chart_group from panel where group_id in (select panels_group from dashboard where dashboard_type=?))", "%"+metric+"%", metricQuery[0].MetricType).Find(&charts)
	if err != nil {
		return fmt.Errorf("Try to get charts data fail,%s ", err.Error())
	}
	if len(charts) > 0 {
		for _, chart := range charts {
			newChartMetricList := []string{}
			for _, v := range strings.Split(chart.Metric, "^") {
				if v == metric || v == "" {
					continue
				}
				newChartMetricList = append(newChartMetricList, v)
			}
			if len(newChartMetricList) == 0 {
				actions = append(actions, &Action{Sql: "delete from chart where id=?", Param: []interface{}{chart.Id}})
			} else {
				actions = append(actions, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newChartMetricList, "^"), chart.Id}})
			}
		}
	}
	err = Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update database fail,%s ", err.Error())
	}
	return err
}

func MetricComparisonListNew(guid, monitorType, serviceGroup, onlyService, endpointGroup, endpoint, metric string) (result []*models.MetricComparisonExtend, err error) {
	var params []interface{}
	baseSql := "select m.*,mc.guid as metric_comparison_id,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.origin_metric_id as metric_id,mc.origin_metric from metric m join metric_comparison mc on mc.metric_id = m.guid "
	if guid != "" {
		baseSql += " and m.guid=? "
		params = append(params, guid)
	} else {
		if serviceGroup != "" {
			if monitorType == "" {
				return result, fmt.Errorf("serviceGroup is disable when monitorType is null ")
			}
			if onlyService == "Y" {
				baseSql = "select m.*,mc.guid as metric_comparison_id,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.origin_metric_id as metric_id,mc.origin_metric from metric m join metric_comparison mc on mc.metric_id = m.guid and m.monitor_type=? and m.service_group=?"
				params = []interface{}{monitorType, serviceGroup}
			} else {
				baseSql = "select m.*,mc.guid as metric_comparison_id,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.origin_metric_id as metric_id,mc.origin_metric from metric m join metric_comparison mc on mc.metric_id = m.guid  and m.monitor_type=? and (m.service_group is null or m.service_group=?)"
				params = []interface{}{monitorType, serviceGroup}
			}
		} else if endpointGroup != "" {
			baseSql = "select m.*,mc.guid as metric_comparison_id,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.origin_metric_id as metric_id,mc.origin_metric from metric m join metric_comparison mc on mc.metric_id = m.guid  and m.service_group is null and endpoint_group = ?"
			params = []interface{}{endpointGroup}
		} else if endpoint != "" {
			baseSql = "select m.*,mc.guid as metric_comparison_id,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.origin_metric_id as metric_id,mc.origin_metric  from ("
			baseSql = baseSql + "select * from metric where service_group in (select service_group from endpoint_service_rel where endpoint=?)"
			baseSql = baseSql + " union "
			baseSql = baseSql + " select * from metric where endpoint_group in (select endpoint_group from endpoint_group_rel where endpoint=?) "
			baseSql = baseSql + " union "
			baseSql = baseSql + " select * from metric where monitor_type in (select monitor_type from endpoint_new where guid=?) and service_group is null and endpoint_group  is null "
			baseSql = baseSql + ") m join metric_comparison mc on mc.metric_id = m.guid"
			params = []interface{}{endpoint, endpoint, endpoint}
		} else {
			baseSql = "select m.*,mc.guid as metric_comparison_id,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period,mc.origin_metric_id as metric_id,mc.origin_metric from metric m join metric_comparison mc on mc.metric_id = m.guid  and  m.monitor_type=? and m.service_group is null and m.endpoint_group is null"
			params = []interface{}{monitorType}
		}
		if strings.TrimSpace(metric) != "" {
			baseSql = baseSql + " and m.metric like '%" + metric + "%'"
		}
	}
	result = []*models.MetricComparisonExtend{}
	baseSql = baseSql + " order by m.update_time desc"
	err = x.SQL(baseSql, params...).Find(&result)
	if err != nil {
		return
	}
	for _, metric := range result {
		if strings.TrimSpace(metric.ServiceGroup) == "" && strings.TrimSpace(metric.EndpointGroup) == "" {
			metric.MetricType = string(models.MetricTypeCommon)
		} else if strings.TrimSpace(metric.LogMetricGroup) != "" {
			metric.MetricType = string(models.MetricTypeBusiness)
		} else {
			// 业务配置类型 兜底
			if strings.TrimSpace(metric.LogMetricConfig) != "" || strings.TrimSpace(metric.LogMetricTemplate) != "" {
				metric.MetricType = string(models.MetricTypeBusiness)
			} else {
				metric.MetricType = string(models.MetricTypeCustom)
			}
		}
		if strings.TrimSpace(metric.OriginCalcType) != "" {
			metric.CalcType = strings.Split(metric.OriginCalcType, ",")
		} else {
			metric.CalcType = []string{}
		}
		if endpoint != "" {
			if strings.TrimSpace(metric.ServiceGroup) != "" {
				metric.GroupType = "level"
				metric.GroupName = metric.ServiceGroup
			} else if strings.TrimSpace(metric.EndpointGroup) != "" {
				metric.GroupType = "object"
				metric.GroupName = metric.EndpointGroup
			} else {
				metric.GroupType = "system"
				metric.GroupName = metric.MonitorType
			}
		}
		if strings.TrimSpace(metric.LogMetricGroup) != "" {
			var name string
			if _, err = x.SQL("select name from log_metric_group where guid = ?", metric.LogMetricGroup).Get(&name); err != nil {
				return
			}
			metric.LogMetricGroupName = name
		}
	}
	return
}

func MetricListNew(guid, monitorType, serviceGroup, onlyService, endpointGroup, endpoint, query, metric string) (result []*models.MetricTable, err error) {
	var params []interface{}
	baseSql := "select * from metric where 1=1 "
	if guid != "" {
		baseSql += " and guid=? "
		params = append(params, guid)
	} else {
		if serviceGroup != "" {
			if onlyService == "Y" {
				baseSql = "select * from metric m where service_group=?"
				params = []interface{}{serviceGroup}
			} else {
				baseSql = "select * from metric m where monitor_type=? and (service_group is null or service_group=?)"
				params = []interface{}{monitorType, serviceGroup}
			}
		} else if endpointGroup != "" {
			baseSql = "select * from metric m where service_group is null and endpoint_group = ? "
			params = []interface{}{endpointGroup}
		} else if endpoint != "" {
			baseSql = "select * from ("
			baseSql = baseSql + "select * from metric where service_group in (select service_group from endpoint_service_rel where endpoint=?)"
			baseSql = baseSql + " union "
			baseSql = baseSql + " select * from metric where endpoint_group in (select endpoint_group from endpoint_group_rel where endpoint=?) "
			baseSql = baseSql + " union "
			baseSql = baseSql + " select * from metric where monitor_type in (select monitor_type from endpoint_new where guid=?) and service_group is null and endpoint_group  is null "
			baseSql = baseSql + ") m where 1=1"
			params = []interface{}{endpoint, endpoint, endpoint}
		} else {
			baseSql = "select * from metric m where monitor_type=? and service_group is null and endpoint_group is null"
			params = []interface{}{monitorType}
		}
		if query == "comparison" {
			baseSql = baseSql + " and exists (select guid from metric_comparison mc where mc.metric_id = m.guid)"
		} else if query != "all" {
			baseSql = baseSql + " and not exists (select guid from metric_comparison mc where mc.metric_id = m.guid)"
		}
	}
	if strings.TrimSpace(metric) != "" {
		baseSql = baseSql + " and metric like '%" + metric + "%'"
	}
	result = []*models.MetricTable{}
	baseSql = baseSql + " order by update_time desc"
	err = x.SQL(baseSql, params...).Find(&result)
	if err != nil {
		return
	}
	for _, metric := range result {
		if strings.TrimSpace(metric.ServiceGroup) == "" && strings.TrimSpace(metric.EndpointGroup) == "" {
			metric.MetricType = string(models.MetricTypeCommon)
		} else if strings.TrimSpace(metric.LogMetricGroup) != "" {
			metric.MetricType = string(models.MetricTypeBusiness)
		} else {
			// 业务配置类型 兜底
			if strings.TrimSpace(metric.LogMetricConfig) != "" || strings.TrimSpace(metric.LogMetricTemplate) != "" {
				metric.MetricType = string(models.MetricTypeBusiness)
			} else {
				metric.MetricType = string(models.MetricTypeCustom)
			}
		}
		if endpoint != "" {
			if strings.TrimSpace(metric.ServiceGroup) != "" {
				metric.GroupType = "level"
				metric.GroupName = metric.ServiceGroup
			} else if strings.TrimSpace(metric.EndpointGroup) != "" {
				metric.GroupType = "object"
				metric.GroupName = metric.EndpointGroup
			} else {
				metric.GroupType = "system"
				metric.GroupName = metric.MonitorType
			}
		}
		if strings.TrimSpace(metric.LogMetricGroup) != "" {
			var name string
			if _, err = x.SQL("select name from log_metric_group where guid = ?", metric.LogMetricGroup).Get(&name); err != nil {
				return
			}
			metric.LogMetricGroupName = name
		}
	}
	return
}

// MetricComparisonImport  同环比指标导入
func MetricComparisonImport(operator string, inputMetrics []*models.MetricComparisonExtend) (failList []string, err error) {
	failList = []string{}
	var actions []*Action
	for _, metric := range inputMetrics {
		// 1.先查询原始指标是否存在,不存在该指标记录为失败
		targetMetric := &models.MetricTable{}
		if _, err = x.SQL("select * from metric where guid =?", metric.MetricId).Get(targetMetric); err != nil {
			return
		}
		if targetMetric == nil || targetMetric.Metric == "" {
			failList = append(failList, metric.Metric)
			continue
		}
		// 2. 查询同环比指标是否存在
		guid := ""
		if _, err = x.SQL("select guid from metric where guid=?", metric.Guid).Get(&guid); err != nil {
			return
		}
		if guid != "" {
			failList = append(failList, metric.Metric)
			continue
		}
		param := convertMetric2ComparisonParam(metric)
		newMetricId := GetComparisonMetricId(metric.Guid, param.ComparisonType, param.CalcMethod, param.CalcPeriod)
		promExpr := NewPromExpr(newMetricId)
		if err = datasource.CheckPrometheusQL(promExpr); err != nil {
			failList = append(failList, metric.Metric)
			continue
		}
		// 新增同环比
		if subActions := GetAddComparisonMetricActions(param, targetMetric, operator); len(subActions) > 0 {
			actions = append(actions, subActions...)
		}
	}
	if len(actions) > 0 {
		err = Transaction(actions)
	}
	return
}

func MetricImport(monitorType, serviceGroup, endPointGroup, operator string, inputMetrics []*models.MetricTable) ([]string, error) {
	var failList []string
	var err error
	if monitorType == "" {
		monitorType = inputMetrics[0].MonitorType
	}
	existMetrics, getExistErr := MetricListNew("", monitorType, serviceGroup, "Y", endPointGroup, "", "all", "")
	if getExistErr != nil {
		return failList, fmt.Errorf("get serviceGroup:%s exist metric list fail,%s ", serviceGroup, getExistErr.Error())
	}
	var alarmStrategyRows []*models.AlarmStrategyTable
	if err = x.SQL("select endpoint_group,metric from alarm_strategy").Find(&alarmStrategyRows); err != nil {
		return failList, fmt.Errorf("query alarm strategy fail,%s ", err.Error())
	}
	oldServerGroup := inputMetrics[0].ServiceGroup
	strategyMap := make(map[string]string)
	for _, v := range alarmStrategyRows {
		strategyMap[v.Metric] = v.EndpointGroup
	}
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	systemMetricMap := convertString2Map(systemMetricList)
	for _, inputMetric := range inputMetrics {
		// 命中系统自带指标直接跳过
		if systemMetricMap[inputMetric.Guid] {
			continue
		}
		if serviceGroup != "" {
			inputMetric.Guid = fmt.Sprintf("%s__%s", inputMetric.Metric, serviceGroup)
		} else {
			inputMetric.Guid = fmt.Sprintf("%s__%s", inputMetric.Metric, monitorType)
		}
		matchMetric := &models.MetricTable{}
		for _, existMetric := range existMetrics {
			if existMetric.Guid == inputMetric.Guid {
				matchMetric = existMetric
				break
			}
		}
		if matchMetric.Guid != "" {
			// 指标重复后,指标名_1,如果还是重复,记录在失败列表
			inputMetric.Metric = inputMetric.Metric + "_1"
			if serviceGroup != "" {
				inputMetric.Guid = fmt.Sprintf("%s__%s", inputMetric.Metric, serviceGroup)
			} else {
				inputMetric.Guid = fmt.Sprintf("%s__%s", inputMetric.Metric, monitorType)
			}
			matchMetric = &models.MetricTable{}
			for _, existMetric := range existMetrics {
				if existMetric.Guid == inputMetric.Guid {
					matchMetric = existMetric
					break
				}
			}
		}
		inputMetric.PromExpr = strings.ReplaceAll(inputMetric.PromExpr, oldServerGroup, serviceGroup)
		if matchMetric.Guid != "" {
			failList = append(failList, inputMetric.Metric)
		} else {
			var tempMetric string
			var newServiceGroup, newEndpointGroup, dbMetricMonitor sql.NullString
			x.SQL("select metric from metric where guid = ?", inputMetric.Guid).Get(&tempMetric)
			if tempMetric != "" {
				failList = append(failList, tempMetric)
			} else {
				newServiceGroup = sql.NullString{String: serviceGroup}
				newEndpointGroup = sql.NullString{String: endPointGroup}
				dbMetricMonitor = sql.NullString{String: inputMetric.DbMetricMonitor}
				actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user,endpoint_group,db_metric_monitor) value (?,?,?,?,?,?,?,?,?,?,?,?)",
					Param: []interface{}{inputMetric.Guid, inputMetric.Metric, monitorType, inputMetric.PromExpr, newServiceGroup, inputMetric.Workspace, nowTime, nowTime, operator, operator, newEndpointGroup, dbMetricMonitor}})
			}
		}
	}
	log.Logger.Info("import metric", log.Int("actionLen", len(actions)))
	if len(actions) > 0 {
		if err = Transaction(actions); err != nil {
			return failList, fmt.Errorf("import metric fail with exec database,%s ", err.Error())
		}
	}
	return failList, nil
}

func GetSimpleMetric(metricId string) (metricRow *models.MetricTable, err error) {
	var metricRows []*models.MetricTable
	err = x.SQL("select * from metric where guid=?", metricId).Find(&metricRows)
	if err != nil {
		err = fmt.Errorf("query metric table with guid:%s fail,%s ", metricId, err.Error())
		return
	}
	if len(metricRows) == 0 {
		err = fmt.Errorf("can not find metric with guid:%s ", metricId)
	} else {
		metricRow = metricRows[0]
	}
	return
}

func GetOriginMetricByComparisonId(metricId string) (metricRow *models.MetricTable, err error) {
	var metricList []*models.MetricTable
	err = x.SQL("select * from metric where guid in (select origin_metric_id from metric_comparison where metric_id = ?)", metricId).Find(&metricList)
	if len(metricList) > 0 {
		metricRow = metricList[0]
	}
	return
}

func GetMetricTags(metricRow *models.MetricTable) (tags []string, tagConfigValue map[string][]string, err error) {
	if metricRow == nil {
		return
	}
	tagConfigValue = make(map[string][]string)
	if metricRow.LogMetricGroup != "" {
		var stringMapRows []*models.LogMetricStringMapTable
		err = x.SQL("select target_value,log_param_name,value_type from log_metric_string_map where log_metric_group=?", metricRow.LogMetricGroup).Find(&stringMapRows)
		if err != nil {
			err = fmt.Errorf("query log metric string map table fail,%s ", err.Error())
			return
		}
		var isSuccessMetric, isFailMetric bool
		if metricRow.LogMetricTemplate != "" {
			var logMetricTemplateRows []*models.LogMetricTemplate
			err = x.SQL("select metric from log_metric_template where guid=?", metricRow.LogMetricTemplate).Find(&logMetricTemplateRows)
			if err != nil {
				err = fmt.Errorf("query log metric template table fail,%s ", err.Error())
				return
			}
			if len(logMetricTemplateRows) > 0 {
				if logMetricTemplateRows[0].Metric == "req_suc_count" {
					isSuccessMetric = true
				} else if logMetricTemplateRows[0].Metric == "req_fail_count" || logMetricTemplateRows[0].Metric == "req_fail_count_detail" {
					isFailMetric = true
				}
			}
		}
		for _, row := range stringMapRows {
			if row.ValueType != "" {
				if isSuccessMetric && row.ValueType == "fail" {
					continue
				}
				if isFailMetric && row.ValueType == "success" {
					continue
				}
			}
			if v, ok := tagConfigValue[row.LogParamName]; ok {
				tagConfigValue[row.LogParamName] = append(v, row.TargetValue)
			} else {
				tagConfigValue[row.LogParamName] = []string{row.TargetValue}
			}
		}
	}
	if metricRow.LogMetricConfig != "" {
		var logMetricConfigRows []*models.LogMetricConfigTable
		err = x.SQL("select metric,tag_config from log_metric_config where guid=?", metricRow.LogMetricConfig).Find(&logMetricConfigRows)
		if err != nil {
			err = fmt.Errorf("query log metric config with guid:%s fail,%s ", metricRow.LogMetricConfig, err.Error())
		}
		if len(logMetricConfigRows) > 0 {
			tagConfigString := logMetricConfigRows[0].TagConfig
			if tagConfigString != "" && tagConfigString != "null" && tagConfigString != "[]" {
				var tmpTagList []string
				if err = json.Unmarshal([]byte(logMetricConfigRows[0].TagConfig), &tmpTagList); err != nil {
					err = fmt.Errorf("json unmarshal tag config: %s fail,%s ", tagConfigString, err.Error())
					return
				}
				if len(tmpTagList) > 0 {
					tags = tmpTagList
				}
			}
		}
		return
	}
	if metricRow.LogMetricTemplate != "" {
		var logMetricTemplateRows []*models.LogMetricTemplate
		err = x.SQL("select metric,tag_config from log_metric_template where guid=?", metricRow.LogMetricTemplate).Find(&logMetricTemplateRows)
		if err != nil {
			err = fmt.Errorf("query log metric template with guid:%s fail,%s ", metricRow.LogMetricTemplate, err.Error())
		}
		if len(logMetricTemplateRows) > 0 {
			tagConfigString := logMetricTemplateRows[0].TagConfig
			if tagConfigString != "" && tagConfigString != "null" && tagConfigString != "[]" {
				if err = json.Unmarshal([]byte(logMetricTemplateRows[0].TagConfig), &tags); err != nil {
					err = fmt.Errorf("json unmarshal tag config: %s fail,%s ", tagConfigString, err.Error())
					return
				}
			}
		}
		return
	}
	tagParamList := getPromTagParamList(metricRow.PromExpr)
	for _, v := range tagParamList {
		if strings.HasPrefix(v, "$t_") {
			tags = append(tags, v[3:])
		}
	}
	return
}

func getPromTagParamList(promQl string) (tagList []string) {
	if strings.Contains(promQl, "$") {
		re, _ := regexp.Compile("=\"[\\$]+[^\"]+\"")
		fetchTag := re.FindAll([]byte(promQl), -1)
		for _, vv := range fetchTag {
			tmpV := string(vv)
			if len(tmpV) < 3 {
				continue
			}
			tmpV = tmpV[2 : len(tmpV)-1]
			tagList = append(tagList, tmpV)
		}
	}
	return
}

func GetMetric(id string) (metric *models.MetricTable, err error) {
	metric = &models.MetricTable{}
	_, err = x.SQL("select * from metric where guid=?", id).Get(metric)
	return
}

func GetMetricByName(name string) (metricList []*models.MetricTable, err error) {
	metricList = []*models.MetricTable{}
	err = x.SQL("select * from metric where metric =?", name).Find(&metricList)
	return
}

func QueryMetricNameList(metric string) (list []string, err error) {
	if metric != "" {
		err = x.SQL("select distinct metric from metric  where metric like '%" + metric + "%'order by update_time desc limit 20").Find(&list)
	} else {
		err = x.SQL("select distinct metric from metric order by update_time desc limit 20").Find(&list)
	}
	if strings.Contains("log_monitor", metric) && notContains(list, "log_monitor") {
		list = append(list, "log_monitor")
	}
	if strings.Contains("db_keyword_monitor", metric) && notContains(list, "db_keyword_monitor") {
		list = append(list, "db_keyword_monitor")
	}
	return
}

func notContains(list []string, str string) bool {
	if len(list) == 0 {
		return true
	}
	for _, s := range list {
		if s == str {
			return false
		}
	}
	return true
}

func GetAddComparisonMetricActions(param models.MetricComparisonParam, metric *models.MetricTable, operator string) (actions []*Action) {
	actions = []*Action{}
	var calcType, metricName, promExpr string
	if len(param.CalcType) > 0 {
		calcType = strings.Join(param.CalcType, ",")
	}
	newMetricId := GetComparisonMetricId(metric.Guid, param.ComparisonType, param.CalcMethod, param.CalcPeriod)
	now := time.Now().Format(models.DatetimeFormat)
	metricName = getComparisonMetric(metric.Metric, param.ComparisonType, param.CalcMethod, param.CalcPeriod)
	promExpr = NewPromExpr(metricName)
	if metric.ServiceGroup != "" && strings.Contains(metric.PromExpr, "service_group=\""+metric.ServiceGroup+"\"") {
		promExpr = promExpr + "{service_group=\"" + metric.ServiceGroup + "\"}"
	}
	if strings.Contains(metric.PromExpr, "instance=\"$address\"") {
		if strings.Contains(promExpr, "{") {
			promExpr = promExpr[:len(promExpr)-1] + ",e_guid=\"$guid\"}"
		} else {
			promExpr = promExpr + "{e_guid=\"$guid\"}"
		}
	}
	tagParamList := getPromTagParamList(metric.PromExpr)
	if len(tagParamList) > 0 {
		for _, v := range tagParamList {
			if strings.HasPrefix(v, "$t_") {
				if strings.Contains(promExpr, "{") {
					promExpr = promExpr[:len(promExpr)-1] + "," + v[3:] + "=\"" + v + "\"}"
				} else {
					promExpr = promExpr + "{" + v[3:] + "=\"" + v + "\"}"
				}
			}
		}
	}
	if strings.Contains(promExpr, "{") {
		promExpr = promExpr[:len(promExpr)-1] + ",calc_type=\"$t_calc_type\"}"
	} else {
		promExpr = promExpr + "{calc_type=\"$t_calc_type\"}"
	}

	if metric.ServiceGroup == "" {
		if metric.EndpointGroup == "" {
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,workspace,update_time,create_time,create_user,update_user,log_metric_config,log_metric_template,log_metric_group) values (?,?,?,?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{newMetricId, metricName, metric.MonitorType, promExpr, metric.Workspace, now, now, operator, operator, metric.LogMetricConfig, metric.LogMetricTemplate, metric.LogMetricGroup}})
		} else {
			actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,workspace,update_time,create_time,create_user,update_user,endpoint_group,log_metric_config,log_metric_template,log_metric_group) values (?,?,?,?,?,?,?,?,?,?,?,?,?)",
				Param: []interface{}{newMetricId, metricName, metric.MonitorType, promExpr, metric.Workspace, now, now, operator, operator, metric.EndpointGroup, metric.LogMetricConfig, metric.LogMetricTemplate, metric.LogMetricGroup}})
		}
	} else {
		actions = append(actions, &Action{Sql: "insert into metric(guid,metric,monitor_type,prom_expr,service_group,workspace,update_time,create_time,create_user,update_user,log_metric_config,log_metric_template,log_metric_group) values (?,?,?,?,?,?,?,?,?,?,?,?,?)",
			Param: []interface{}{newMetricId, metricName, metric.MonitorType, promExpr, metric.ServiceGroup, metric.Workspace, now, now, operator, operator, metric.LogMetricConfig, metric.LogMetricTemplate, metric.LogMetricGroup}})
	}
	actions = append(actions, &Action{Sql: "insert into metric_comparison(guid,comparison_type,calc_type,calc_method,calc_period,metric_id,origin_metric_id,origin_metric,origin_prom_expr,create_user,create_time) values(?,?,?,?,?,?,?,?,?,?,?)",
		Param: []interface{}{guid.CreateGuid(), param.ComparisonType, calcType, param.CalcMethod, param.CalcPeriod, newMetricId, metric.Guid, metric.Metric, metric.PromExpr, operator, now}})
	return
}

// NewPromExpr 需要将 . 替换成 _
func NewPromExpr(newMetricId string) string {
	if newMetricId == "" {
		return ""
	}
	return strings.ReplaceAll(newMetricId, ".", "_")
}

func AddComparisonMetric(param models.MetricComparisonParam, metric *models.MetricTable, operator string) error {
	var actions []*Action
	actions = GetAddComparisonMetricActions(param, metric, operator)
	return Transaction(actions)
}

func GetComparisonMetricByMetricId(metricId string) (comparison models.MetricComparison, err error) {
	_, err = x.SQL("select * from metric_comparison where metric_id=?", metricId).Get(&comparison)
	return
}

func UpdateComparisonMetric(metricComparisonId string, calcTypeList []string) (err error) {
	var calcType string
	if len(calcTypeList) > 0 {
		calcType = strings.Join(calcTypeList, ",")
	}
	_, err = x.Exec("update metric_comparison set calc_type=? where guid = ?", calcType, metricComparisonId)
	return
}

func DeleteComparisonMetric(id string) (err error) {
	var actions, subActions []*Action
	actions = append(actions, &Action{"delete from metric_comparison where metric_id = ?", []interface{}{id}})
	actions = append(actions, &Action{"delete from metric where guid = ?", []interface{}{id}})
	if subActions, err = DeleteCustomChartSeriesByMetricIdSQL(id); err != nil {
		return
	}
	if len(subActions) > 0 {
		actions = append(actions, subActions...)
	}
	return Transaction(actions)
}

func GetComparisonMetricDtoList() (list []*models.MetricComparisonDto, err error) {
	err = x.SQL("select mc.origin_metric,mc.origin_prom_expr as origin_prom_expr,m.metric,m.prom_expr,mc.comparison_type,mc.calc_type,mc.calc_method,mc.calc_period from metric m join metric_comparison mc on m.guid = mc.metric_id").Find(&list)
	return
}

func GetComparisonMetricId(originMetricId, comparisonType, calcMethod string, calcPeriod int) string {
	if comparisonType == "" {
		return ""
	}
	return originMetricId + "_" + comparisonType[0:1] + "_" + calcMethod + "_" + fmt.Sprintf("%d", calcPeriod)
}

func getComparisonMetric(metric, comparisonType, calcMethod string, calcPeriod int) string {
	if comparisonType == "" {
		return ""
	}
	metric = strings.ReplaceAll(metric, ".", "_")
	return metric + "__" + comparisonType[0:1] + "_" + calcMethod + "_" + fmt.Sprintf("%d", calcPeriod)
}

func convertMetric2ComparisonParam(comparison *models.MetricComparisonExtend) models.MetricComparisonParam {
	if comparison == nil {
		return models.MetricComparisonParam{}
	}
	return models.MetricComparisonParam{
		Metric:         comparison.Metric,
		MonitorType:    comparison.MonitorType,
		ComparisonType: comparison.ComparisonType,
		PromExpr:       comparison.PromExpr,
		CalcType:       comparison.CalcType,
		CalcMethod:     comparison.CalcMethod,
		CalcPeriod:     comparison.CalcPeriod,
	}
}

// SyncMetricComparison  服务启动休眠1min向exporter同步同环比数据(数据兜底用)
func SyncMetricComparison() {
	time.Sleep(1 * time.Minute)
	SyncMetricComparisonData()
}

// SyncMetricComparisonData 同步同环比指标数据
func SyncMetricComparisonData() (err error) {
	var list []*models.MetricComparisonDto
	var resByteArr []byte
	var response models.Response
	if list, err = GetComparisonMetricDtoList(); err != nil {
		return
	}
	if len(list) > 0 {
		param, _ := json.Marshal(list)
		if resByteArr, err = HttpPost("http://127.0.0.1:8181/receive", "", param); err != nil {
			return
		}
		if err = json.Unmarshal(resByteArr, &response); err != nil {
			return
		}
		if response.Status != "OK" {
			err = fmt.Errorf(response.Message)
		}
	}
	return
}

// HttpPost Post请求
func HttpPost(url, token string, postBytes []byte) (byteArr []byte, err error) {
	req, reqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(postBytes))
	if reqErr != nil {
		err = fmt.Errorf("new http reqeust fail,%s ", reqErr.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = fmt.Errorf("do http reqeust fail,%s ", respErr.Error())
		return
	}
	byteArr, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}

func getDeleteEndpointDashboardChartMetricAction(metric, monitorType string) (actions []*Action) {
	var charts []*models.ChartTable
	err := x.SQL("select id,metric from chart where metric like ? and group_id in (select chart_group from panel where group_id in (select panels_group from dashboard where dashboard_type=?))", "%"+metric+"%", monitorType).Find(&charts)
	if err != nil {
		log.Logger.Error("getDeleteEndpointDashboardChartMetricAction,try to get charts data fail", log.String("metric", metric), log.Error(err))
		return
	}
	if len(charts) > 0 {
		for _, chart := range charts {
			newChartMetricList := []string{}
			for _, v := range strings.Split(chart.Metric, "^") {
				if v == metric || v == "" {
					continue
				}
				newChartMetricList = append(newChartMetricList, v)
			}
			if len(newChartMetricList) == 0 {
				actions = append(actions, &Action{Sql: "delete from chart where id=?", Param: []interface{}{chart.Id}})
			} else {
				actions = append(actions, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newChartMetricList, "^"), chart.Id}})
			}
		}
	}
	return
}

func getMetricComparisonDeleteAction(sourceMetricGuid string) (actions []*Action, affectEndpointGroup []string) {
	// 删除同环比 指标
	var metricRows []*models.MetricComparison
	if err := x.SQL("select metric_id from metric_comparison where origin_metric_id = ?", sourceMetricGuid).Find(&metricRows); err != nil {
		log.Logger.Error("getMetricComparisonDeleteAction,query metric_comparison fail", log.Error(err))
		return
	}
	for _, row := range metricRows {
		actions = append(actions, &Action{Sql: "delete from metric_comparison where  metric_id = ?", Param: []interface{}{row.MetricId}})
		tmpActions, tmpEndpointGroup := getDeleteMetricActions(row.MetricId)
		actions = append(actions, tmpActions...)
		affectEndpointGroup = append(affectEndpointGroup, tmpEndpointGroup...)
	}
	return
}
