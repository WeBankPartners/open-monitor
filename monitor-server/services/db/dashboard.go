package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
	"regexp"
	"sort"
	"strings"
	"time"
)

func GetDashboard(dType string) (error, m.DashboardTable) {
	var dashboards []*m.DashboardTable
	err := x.SQL("select * from dashboard where dashboard_type=?", dType).Find(&dashboards)
	if len(dashboards) > 0 {
		return nil, *dashboards[0]
	} else {
		if err == nil {
			err = fmt.Errorf("No rows fetch ")
		}
		log.Error(nil, log.LOGGER_APP, "Query dashboard fail", zap.Error(err))
		return err, m.DashboardTable{}
	}
}

func GetSearch(id int) (error, m.SearchModel) {
	var search []*m.SearchModel
	sql := `select * from search where id=?`
	err := x.SQL(sql, id).Find(&search)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query search fail", zap.Error(err))
	}
	if len(search) > 0 {
		return err, *search[0]
	} else {
		return fmt.Errorf("no rows fetch"), *new(m.SearchModel)
	}
}

func GetButton(bGroup int) (error, []*m.ButtonModel) {
	var buttons []*m.ButtonModel
	sql := `select * from button where group_id=?`
	err := x.SQL(sql, bGroup).Find(&buttons)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query button fail", zap.Error(err))
	}
	if len(buttons) > 0 {
		for _, v := range buttons {
			var options []*m.OptionModel
			cSql := `select * from option where group_id=?`
			cErr := x.SQL(cSql, v.OptionGroup).Find(&options)
			if cErr == nil {
				v.Options = options
			}
		}
		return err, buttons
	} else {
		return fmt.Errorf("no rows fetch"), buttons
	}
}

func GetPanels(pGroup int, endpoint string) (error, []*m.PanelTable) {
	var serviceGroupList []string
	var endpointServiceRel []*m.EndpointServiceRelTable
	x.SQL("select distinct service_group from endpoint_service_rel where endpoint=?", endpoint).Find(&endpointServiceRel)
	if len(endpointServiceRel) > 0 {
		for _, v := range endpointServiceRel {
			tmpParentList, _ := fetchGlobalServiceGroupParentGuidList(v.ServiceGroup)
			serviceGroupList = append(serviceGroupList, tmpParentList...)
		}
	}
	var panels []*m.PanelTable
	sql := "select * from panel where group_id=? and (service_group is null or service_group in ('" + strings.Join(serviceGroupList, "','") + "'))"
	err := x.SQL(sql, pGroup).Find(&panels)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query panels fail", zap.Error(err))
	}
	return err, panels
}

func GetCharts(cGroup int, chartId int, panelId int) (error, []*m.ChartTable) {
	var charts []*m.ChartTable
	sql := ``
	var err error
	if cGroup > 0 {
		sql = `select * from chart where group_id=?`
		err = x.SQL(sql, cGroup).Find(&charts)
	} else if chartId > 0 {
		sql = `select * from chart where id=?`
		err = x.SQL(sql, chartId).Find(&charts)
	} else if panelId > 0 {
		sql = `SELECT t1.* FROM chart t1 INNER JOIN panel t2 ON t1.group_id=t2.chart_group WHERE t2.id=?`
		err = x.SQL(sql, panelId).Find(&charts)
	}
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query charts fail", zap.Error(err))
	}
	if len(charts) > 0 {
		return err, charts
	}
	return fmt.Errorf("Get charts error "), charts
}

func GetPromMetric(endpoint []string, metric string) (error, string) {
	promQL := metric
	tmpMetric := metric
	var tmpTag string
	if strings.Contains(tmpMetric, "/") {
		tmpMetric = metric[:strings.Index(metric, "/")]
		tmpTag = metric[strings.Index(metric, "/")+1:]
	}
	var query []*m.MetricTable
	err := x.SQL("SELECT * FROM metric WHERE metric=? and service_group is null", tmpMetric).Find(&query)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query metric fail", zap.Error(err))
	}
	if len(query) > 0 {
		host := m.EndpointTable{Guid: endpoint[0]}
		GetEndpoint(&host)
		if err != nil || host.Id == 0 {
			log.Error(nil, log.LOGGER_APP, "Find endpoint fail", zap.String("endpoint", endpoint[0]), zap.Error(err))
			return err, promQL
		}
		var reg string
		for _, v := range query {
			if v.MonitorType == host.ExportType {
				reg = v.PromExpr
				break
			}
		}
		if reg == "" {
			return fmt.Errorf("Can not match promQl with metric:%s and type:%s ", metric, host.ExportType), promQL
		}
		if strings.Contains(reg, `$ip`) {
			reg = strings.Replace(reg, "$ip", host.Ip, -1)
		}
		if strings.Contains(reg, `$address`) {
			if tmpTag != "" {
				tagAppendString := ""
				for _, tagString := range strings.Split(tmpTag, ",") {
					tagKeyValue := strings.Split(tagString, "=")
					if len(tagKeyValue) == 2 {
						tagAppendString += fmt.Sprintf(",%s=\"%s\"", tagKeyValue[0], tagKeyValue[1])
					}
				}
				reg = strings.ReplaceAll(reg, "\"$address\"", fmt.Sprintf("\"$address\"%s", tagAppendString))
			}
			if host.AddressAgent != "" {
				reg = strings.Replace(reg, "$address", host.AddressAgent, -1)
			} else {
				reg = strings.Replace(reg, "$address", host.Address, -1)
			}
		}
		if strings.Contains(reg, `$guid`) {
			reg = strings.Replace(reg, "$guid", host.Guid, -1)
		}
		if strings.Contains(reg, "$pod") {
			reg = strings.Replace(reg, "$pod", host.Name, -1)
		}
		if strings.Contains(reg, "$k8s_namespace") {
			reg = strings.Replace(reg, "$k8s_namespace", host.ExportVersion, -1)
		}
		if strings.Contains(reg, "$k8s_cluster") {
			reg = strings.Replace(reg, "$k8s_cluster", host.OsType, -1)
		}
		promQL = reg
	}
	return err, promQL
}

func ReplacePromQlKeyword(promQl, metric string, host *m.EndpointNewTable, tagList []*m.TagDto) string {
	var tmpTag, logType string
	var err error
	if strings.Contains(metric, "/") {
		tmpTag = metric[strings.Index(metric, "/")+1:]
	}
	if strings.Contains(promQl, `$ip`) {
		promQl = strings.Replace(promQl, "$ip", host.Ip, -1)
	}
	if strings.Contains(promQl, `$address`) {
		if tmpTag != "" {
			tagAppendString := ""
			for _, tagString := range strings.Split(tmpTag, ",") {
				tagKeyValue := strings.Split(tagString, "=")
				if len(tagKeyValue) == 2 {
					tagAppendString += fmt.Sprintf(",%s=\"%s\"", tagKeyValue[0], tagKeyValue[1])
				}
			}
			promQl = strings.ReplaceAll(promQl, "\"$address\"", fmt.Sprintf("\"$address\"%s", tagAppendString))
		}
		if host.AgentAddress != ".*" {
			promQl = strings.Replace(promQl, "$address", host.AgentAddress, -1)
		}
	}
	if strings.Contains(promQl, `$guid`) {
		if host.Guid == "" {
			promQl = strings.Replace(promQl, "=\"$guid\"", "=~\".*\"", -1)
		} else {
			promQl = strings.Replace(promQl, "$guid", host.Guid, -1)
		}
	}
	if strings.Contains(promQl, "$pod") {
		promQl = strings.Replace(promQl, "$pod", host.Name, -1)
	}
	// 看指标是否为 业务配置过来的,查询业务配置类型,自定义类型需要特殊处理 tags,tags="test_service_code=deleteUser,test_retcode=304"
	if logType, err = GetLogTypeByMetric(metric); err != nil {
		log.Error(nil, log.LOGGER_APP, "GetLogTypeByMetric err", zap.Error(err))
	}
	if len(tagList) > 0 {
		if logType == m.LogMonitorCustomType {
			originPromQl := promQl
			// 原表达式: node_log_metric_monitor_value{key="key",agg="max",service_group="log_sys",tags="$t_tags"}
			// 正则字符串匹配 替换成 node_log_metric_monitor_value{key="key",agg="max",service_group="log_sys",tags!~".*test_service_code=addUser.*"}
			for i, tagObj := range tagList {
				if i == 0 {
					promQl = getTagPromQl(tagObj, originPromQl)
				} else {
					promQl = "(" + promQl + ") and (" + getTagPromQl(tagObj, originPromQl) + ")"
				}
			}
		} else {
			for _, tagObj := range tagList {
				tagSourceString := "$t_" + tagObj.TagName
				if strings.Contains(promQl, tagSourceString) {
					if len(tagObj.TagValue) == 0 {
						promQl = strings.Replace(promQl, "=\""+tagSourceString+"\"", "=~\".*\"", -1)
					} else {
						tmpEqual := "=~"
						if tagObj.Equal == "notin" {
							tmpEqual = "!~"
						}
						promQl = strings.Replace(promQl, "=\""+tagSourceString+"\"", tmpEqual+"\""+strings.Join(tagObj.TagValue, "|")+"\"", -1)
					}
				}
			}
		}
	}
	if strings.Contains(promQl, "$") {
		re, _ := regexp.Compile("=\"[\\$]+[^\"]+\"")
		fetchTag := re.FindAll([]byte(promQl), -1)
		for _, vv := range fetchTag {
			promQl = strings.Replace(promQl, string(vv), "=~\".*\"", -1)
		}
	}
	return promQl
}

func getTagPromQl(tagObj *m.TagDto, originPromQl string) (promQl string) {
	if len(tagObj.TagValue) == 0 {
		promQl = strings.Replace(originPromQl, "=\"$t_tags\"", "=~\".*"+tagObj.TagName+".*\"", -1)
	} else {
		tmpEqual := "=~"
		if tagObj.Equal == "notin" {
			tmpEqual = "!~"
		}
		for i, tagValue := range tagObj.TagValue {
			tagObj.TagValue[i] = fmt.Sprintf(".*%s=%s.*", tagObj.TagName, tagValue)
		}
		promQl = strings.Replace(originPromQl, "=\"$t_tags\"", tmpEqual+"\""+strings.Join(tagObj.TagValue, "|")+"\"", -1)
	}
	return
}

func GetDbPromMetric(endpoint, metric, legend string) (error, string) {
	promQL := "db_monitor_count"
	var query []*m.PromMetricTable
	err := x.SQL("SELECT prom_ql FROM prom_metric WHERE metric='db_monitor_count'").Find(&query)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Query prom_metric fail", zap.Error(err))
	}
	if len(query) > 0 {
		promQL = query[0].PromQl
		if strings.Contains(promQL, "$guid") {
			promQL = strings.Replace(promQL, "$guid", endpoint, -1)
		}
		if metric != "" && legend != "" {
			promQL = promQL[:len(promQL)-1]
			legend = legend[1:]
			promQL = promQL + fmt.Sprintf(",%s=\"%s\"}", legend, metric)
		}
	}
	log.Debug(nil, log.LOGGER_APP, "db prom metric", zap.String("promQl", promQL))
	return err, promQL
}

// 查询所有类型
func SearchHost(endpoint string) (error, []*m.OptionModel) {
	options := []*m.OptionModel{}
	var hosts []*m.EndpointTable
	endpoint = `%` + endpoint + `%`
	err := x.SQL("SELECT * FROM endpoint WHERE guid LIKE ? order by export_type,ip LIMIT 100", endpoint).Find(&hosts)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Search host fail", zap.Error(err))
		return err, options
	}
	for _, host := range hosts {
		if host.ExportType == "node" {
			host.ExportType = "host"
		}
		options = append(options, &m.OptionModel{OptionText: host.Guid, OptionValue: host.Guid, OptionType: host.ExportType, Id: host.Id})
	}
	return err, options
}

// 按类型过滤
func SearchHostByType(endpoint, optionTypeName string) (error, []*m.OptionModel) {
	var options []*m.OptionModel
	var hosts []*m.EndpointTable
	endpoint = `%` + endpoint + `%`
	err := x.SQL("SELECT * FROM endpoint WHERE guid LIKE ? AND export_type=? order by export_type,ip LIMIT 100", endpoint, optionTypeName).Find(&hosts)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Search host fail", zap.Error(err))
		return err, options
	}
	for _, host := range hosts {
		if host.ExportType == "node" {
			host.ExportType = "host"
		}
		options = append(options, &m.OptionModel{OptionText: host.Guid, OptionValue: host.Guid, OptionType: host.ExportType, Id: host.Id})
	}
	return err, options
}

func GetEndpoint(query *m.EndpointTable) error {
	var endpointObj []*m.EndpointTable
	var err error
	if query.Id > 0 {
		err = x.SQL("SELECT * FROM endpoint WHERE id=?", query.Id).Find(&endpointObj)
	} else if query.Guid != "" {
		err = x.SQL("SELECT * FROM endpoint WHERE guid=?", query.Guid).Find(&endpointObj)
	} else if query.Address != "" {
		if query.AddressAgent != "" {
			err = x.SQL("SELECT * FROM endpoint WHERE address=? or address_agent=?", query.Address, query.AddressAgent).Find(&endpointObj)
		} else {
			err = x.SQL("SELECT * FROM endpoint WHERE address=?", query.Address).Find(&endpointObj)
		}
	}
	if query.Ip != "" && query.ExportType != "" {
		if query.Name == "" {
			err = x.SQL("SELECT * FROM endpoint WHERE ip=? and export_type=?", query.Ip, query.ExportType).Find(&endpointObj)
		} else {
			err = x.SQL("SELECT * FROM endpoint WHERE ip=? and export_type=? and name=?", query.Ip, query.ExportType, query.Name).Find(&endpointObj)
		}
	}
	if err != nil {
		return fmt.Errorf("Get endpoint table fail,%s ", err.Error())
	}
	if len(endpointObj) <= 0 {
		log.Warn(nil, log.LOGGER_APP, "Get endpoint fail", log.JsonObj("param", query))
		return fmt.Errorf("Get endpoint fail,can not fetch any data ")
	}
	query.Id = endpointObj[0].Id
	query.Guid = endpointObj[0].Guid
	query.Address = endpointObj[0].Address
	query.Name = endpointObj[0].Name
	query.Ip = endpointObj[0].Ip
	query.Step = endpointObj[0].Step
	query.OsType = endpointObj[0].OsType
	query.ExportVersion = endpointObj[0].ExportVersion
	query.ExportType = endpointObj[0].ExportType
	query.StopAlarm = endpointObj[0].StopAlarm
	query.AddressAgent = endpointObj[0].AddressAgent
	query.Cluster = endpointObj[0].Cluster
	query.Tags = endpointObj[0].Tags
	return nil
}

func GetTags(endpoint string, key string, metric string) (error, []*m.OptionModel) {
	var options []*m.OptionModel
	var endpointObj []*m.EndpointTable
	err := x.SQL("SELECT id FROM endpoint WHERE guid=?", endpoint).Find(&endpointObj)
	if err != nil || len(endpointObj) <= 0 {
		log.Error(nil, log.LOGGER_APP, "Get tags fail, can't find endpoint", zap.Error(err))
		return err, options
	}
	var promMetricObj []*m.PromMetricTable
	err = x.SQL("SELECT prom_main FROM prom_metric WHERE metric=?", metric).Find(&promMetricObj)
	if err != nil || len(promMetricObj) <= 0 {
		log.Error(nil, log.LOGGER_APP, "Get tags fail,can't find prom_metric", zap.Error(err))
		return err, options
	}
	var endpointMetricObj []*m.EndpointMetricTable
	err = x.SQL("SELECT metric FROM endpoint_metric WHERE endpoint_id=? AND metric LIKE ?", endpointObj[0].Id, `%`+promMetricObj[0].PromMain+`%`).Find(&endpointMetricObj)
	if err != nil || len(endpointMetricObj) <= 0 {
		log.Error(nil, log.LOGGER_APP, "Get tags fail,can't find metric", zap.Error(err))
		return err, options
	}
	key = key + "="
	for _, c := range endpointMetricObj {
		if !strings.Contains(c.Metric, key) {
			continue
		}
		metricStr := strings.Replace(c.Metric, `'`, `"`, -1)
		tmpStr := strings.Split(metricStr, key+`"`)[1]
		tagStr := strings.Split(tmpStr, `"`)[0]
		inBlacklist := false
		for _, v := range m.Config().TagBlacklist {
			if strings.Contains(tagStr, v) {
				inBlacklist = true
				break
			}
		}
		if inBlacklist {
			continue
		}
		options = append(options, &m.OptionModel{OptionText: tagStr, OptionValue: key + tagStr})
	}
	return err, options
}

func RegisterEndpointMetric(endpointId int, endpointMetrics []string) error {
	if endpointId <= 0 {
		return fmt.Errorf("endpoint id is 0")
	}
	maxNum := 50
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete FROM endpoint_metric WHERE endpoint_id=?", Param: []interface{}{endpointId}})
	i := 0
	var insertSql string
	params := make([]interface{}, 0)
	for _, v := range endpointMetrics {
		i = i + 1
		if i == 1 {
			insertSql = "INSERT INTO endpoint_metric(endpoint_id,metric) VALUES "
		}
		insertSql = insertSql + "(?,?),"
		params = append(params, endpointId)
		params = append(params, v)
		if i == maxNum {
			var action Action
			action.Sql = insertSql[:len(insertSql)-1]
			action.Param = params
			actions = append(actions, &action)
			i = 0
			params = make([]interface{}, 0)
		}
	}
	if i > 0 {
		var action Action
		action.Sql = insertSql[:len(insertSql)-1]
		action.Param = params
		actions = append(actions, &action)
	}
	err := Transaction(actions)
	return err
}

func GetPromMetricTable(metricType string) (err error, result []*m.PromMetricUpdateParam) {
	if metricType != "" {
		err = x.SQL("SELECT t1.*,t2.group_id as panel_id FROM prom_metric t1 left join chart t2 on t1.metric=t2.metric WHERE t1.metric_type=? order by t1.id", metricType).Find(&result)
	} else {
		result = []*m.PromMetricUpdateParam{}
	}
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get prom metric table fail", zap.Error(err))
	}
	// append service metric
	var logMetricTable []*m.LogMetricConfigTable
	x.SQL("select guid,metric,display_name,agg_type from log_metric_config where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?) or log_metric_json in (select guid from log_metric_json where log_metric_monitor in (select guid from log_metric_monitor where monitor_type=?))", metricType, metricType).Find(&logMetricTable)
	for _, v := range logMetricTable {
		result = append(result, &m.PromMetricUpdateParam{Id: 0, MetricType: metricType, Metric: v.Metric, PromQl: fmt.Sprintf("%s{key=\"%s\",agg=\"%s\",t_endpoint=\"$guid\"}", m.LogMetricName, v.Metric, v.AggType)})
	}
	var dbMetricTable []*m.DbMetricMonitorTable
	x.SQL("select guid,metric,display_name from db_metric_monitor where monitor_type=?", metricType).Find(&dbMetricTable)
	for _, v := range dbMetricTable {
		result = append(result, &m.PromMetricUpdateParam{Id: 0, MetricType: metricType, Metric: v.Metric, PromQl: fmt.Sprintf("%s{key=\"%s\",t_endpoint=\"$guid\"}", m.DBMonitorMetricName, v.Metric)})
	}
	return err, result
}

func UpdatePromMetric(data []*m.PromMetricTable) error {
	var actions []*Action
	for _, v := range data {
		if v.Id != "" {
			actions = append(actions, &Action{Sql: "update prom_metric set metric=?,prom_ql=? where id=?", Param: []interface{}{v.Metric, v.PromQl, v.Id}})
		} else {
			actions = append(actions, &Action{Sql: "insert into prom_metric(metric,metric_type,prom_ql) value (?,?,?)", Param: []interface{}{v.Metric, v.MetricType, v.PromQl}})
		}
	}
	return Transaction(actions)
}

func DeletePromMetric(metric string) (tplIds []int, err error) {
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from prom_metric where metric=?", Param: []interface{}{metric}})
	var charts []*m.ChartTable
	err = x.SQL("select id,metric from chart where metric like ?", "%"+metric+"%").Find(&charts)
	if err != nil {
		err = fmt.Errorf("Try to get charts data fail,%s ", err.Error())
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
	var strategys []*m.StrategyTable
	err = x.SQL("select id,tpl_id,metric from strategy where metric=?", metric).Find(&strategys)
	if err != nil {
		err = fmt.Errorf("Try to get strategy fail,%s ", err.Error())
		return
	}
	if len(strategys) > 0 {
		for _, strategy := range strategys {
			tplIds = append(tplIds, strategy.TplId)
			actions = append(actions, &Action{Sql: "delete from strategy where id=?", Param: []interface{}{strategy.Id}})
		}
	}
	err = Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update database fail,%s ", err.Error())
		return
	}
	return
}

func UpdatePanelChartMetric(data []m.PromMetricUpdateParam) error {
	var chartTable []*m.ChartTable
	x.SQL("select * from chart").Find(&chartTable)
	var updateChartAction []*Action
	for _, v := range data {
		newChartGroupId := v.PanelId
		if newChartGroupId > 0 {
			var existChartObj, newUpdateChartObj m.ChartTable
			for _, chart := range chartTable {
				for _, tmpMetric := range strings.Split(chart.Metric, "^") {
					if tmpMetric == v.Metric {
						existChartObj = *chart
						break
					}
				}
				if chart.GroupId == newChartGroupId && v.Chart.Metric == chart.Metric {
					newUpdateChartObj = *chart
				}
			}
			if existChartObj.Id > 0 {
				if newUpdateChartObj.Id > 0 {
					if existChartObj.Id != newUpdateChartObj.Id {
						if strings.Contains(existChartObj.Metric, "^") {
							var newOldMetricList []string
							for _, tmpExistMetricObj := range strings.Split(existChartObj.Metric, "^") {
								if tmpExistMetricObj != v.Metric {
									newOldMetricList = append(newOldMetricList, tmpExistMetricObj)
								}
							}
							updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newOldMetricList, "^"), existChartObj.Id}})
						} else {
							updateChartAction = append(updateChartAction, &Action{Sql: "delete from chart where id=?", Param: []interface{}{existChartObj.Id}})
						}
						updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=?,title=?,unit=? where id=?", Param: []interface{}{fmt.Sprintf("%s^%s", newUpdateChartObj.Metric, v.Metric), v.Chart.Title, v.Chart.Unit, newUpdateChartObj.Id}})
					}
				} else {
					if existChartObj.GroupId != newChartGroupId {
						if strings.Contains(existChartObj.Metric, "^") {
							var newOldMetricList []string
							for _, tmpExistMetricObj := range strings.Split(existChartObj.Metric, "^") {
								if tmpExistMetricObj != v.Metric {
									newOldMetricList = append(newOldMetricList, tmpExistMetricObj)
								}
							}
							updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newOldMetricList, "^"), existChartObj.Id}})
							updateChartAction = append(updateChartAction, &Action{Sql: "insert into chart(group_id,metric,url,title,unit,legend) value (?,?,'/dashboard/chart',?,?,'$metric')", Param: []interface{}{newChartGroupId, v.Metric, v.Chart.Title, v.Chart.Unit}})
						} else {
							updateChartAction = append(updateChartAction, &Action{Sql: "update chart set group_id=? where id=?", Param: []interface{}{newChartGroupId, existChartObj.Id}})
						}
					}
				}
			} else {
				if newUpdateChartObj.Id > 0 {
					updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=?,title=?,unit=? where id=?", Param: []interface{}{fmt.Sprintf("%s^%s", newUpdateChartObj.Metric, v.Metric), v.Chart.Title, v.Chart.Unit, newUpdateChartObj.Id}})
				} else {
					updateChartAction = append(updateChartAction, &Action{Sql: "insert into chart(group_id,metric,url,title,unit,legend) value (?,?,'/dashboard/chart',?,?,'$metric')", Param: []interface{}{newChartGroupId, v.Metric, v.Chart.Title, v.Chart.Unit}})
				}
			}
		}
	}
	if len(updateChartAction) == 0 {
		return nil
	}
	return Transaction(updateChartAction)
}

func GetServiceGroupPromMetric(serviceGroup, workspace, monitorType string) (err error, result []*m.OptionModel) {
	fetchServiceGroupList, fetchErr := fetchGlobalServiceGroupChildGuidList(serviceGroup)
	if fetchErr != nil {
		err = fetchErr
		return
	}
	result = []*m.OptionModel{}
	var metricList []string
	nowTime := time.Now().Unix()
	queryPromQl := fmt.Sprintf("{service_group=\"%s\"}", serviceGroup)
	if len(fetchServiceGroupList) > 0 {
		queryPromQl = fmt.Sprintf("{service_group=~\"%s\"}", strings.Join(fetchServiceGroupList, "|"))
	}
	metricList, err = datasource.QueryPromQLMetric(queryPromQl, GetClusterAddress(""), nowTime-120, nowTime)
	if err != nil {
		return
	}
	existMap := make(map[string]int)
	for _, v := range metricList {
		if strings.HasPrefix(v, "go_") || v == "" {
			continue
		}
		tmpPromExpr := v
		tmpTEndpoint := ""
		if strings.Contains(v, "t_endpoint") {
			tmpPromExpr, tmpTEndpoint = trimTEndpointTag(tmpPromExpr)
			if !strings.HasSuffix(tmpTEndpoint, monitorType) {
				continue
			}
			if workspace != m.MetricWorkspaceService {
				tmpPromExpr = tmpPromExpr[:len(tmpPromExpr)-1] + ",t_endpoint=\"$guid\"}"
			}
		}
		if _, b := existMap[tmpPromExpr]; b {
			continue
		}
		result = append(result, &m.OptionModel{OptionText: tmpPromExpr, OptionValue: tmpPromExpr})
		existMap[tmpPromExpr] = 1
	}
	return
}

func GetEndpointMetric(endpointGuid, monitorType string) (err error, result []*m.OptionModel) {
	result = []*m.OptionModel{}
	endpointObj := m.EndpointTable{Guid: endpointGuid}
	if endpointGuid == "" && monitorType != "" {
		var endpointList []*m.EndpointTable
		x.SQL("select * from endpoint where export_type=? limit 1", monitorType).Find(&endpointList)
		if len(endpointList) > 0 {
			endpointObj = *endpointList[0]
		}
	} else {
		GetEndpoint(&endpointObj)
	}
	if endpointObj.Guid == "" {
		return fmt.Errorf("endpoint guid: %s can not find ", endpointGuid), result
	}
	if endpointObj.ExportType == "ping" || endpointObj.ExportType == "telnet" || endpointObj.ExportType == "http" {
		return nil, result
	}
	var ip, port, exporterAddress string
	var strList []string
	if endpointObj.ExportType == "snmp" {
		snmpQueryList, err := x.QueryString("select address from snmp_exporter where id in (select snmp_exporter from snmp_endpoint_rel where endpoint_guid=?)", endpointObj.Guid)
		if err != nil {
			return fmt.Errorf("Try to get snmp address fail,%s ", err.Error()), result
		}
		if len(snmpQueryList) == 0 {
			return fmt.Errorf("Can not find snmp address record "), result
		}
		strList, err = prom.GetSnmpMetricList(snmpQueryList[0]["address"], endpointObj.Ip)
	} else {
		if endpointObj.AddressAgent != "" {
			exporterAddress = endpointObj.AddressAgent
		} else {
			exporterAddress = endpointObj.Address
		}
		if !strings.Contains(exporterAddress, ":") {
			return fmt.Errorf("Address %s is illegal ", exporterAddress), result
		}
		ip = exporterAddress[:strings.Index(exporterAddress, ":")]
		port = exporterAddress[strings.Index(exporterAddress, ":")+1:]
		if ip == "" || port == "" {
			log.Warn(nil, log.LOGGER_APP, "endpoint address illegal ", zap.String("endpoint", endpointObj.Guid))
			return nil, result
		}
		metricQueryParam := m.QueryPrometheusMetricParam{Ip: ip, Port: port, Cluster: endpointObj.Cluster, Prefix: []string{}, Keyword: []string{}, TargetGuid: endpointObj.Guid, IsConfigQuery: true, ServiceGroup: ""}
		err, strList = QueryExporterMetric(metricQueryParam)
	}
	if err != nil {
		return err, result
	}
	resultExistMap := make(map[string]int)
	for _, v := range strList {
		if strings.HasPrefix(v, "go_") || v == "" {
			continue
		}
		tmpText := v
		if strings.Contains(v, "{") {
			tmpText = v[:strings.Index(tmpText, "{")]
		}
		if _, b := resultExistMap[tmpText]; b {
			continue
		}
		result = append(result, &m.OptionModel{OptionText: tmpText, OptionValue: fmt.Sprintf("%s{instance=\"$address\"}", tmpText)})
		resultExistMap[tmpText] = 1
	}
	return nil, result
}

func trimTEndpointTag(input string) (output, tEndpoint string) {
	tIndex := strings.Index(input, "t_endpoint=\"")
	tailPart := input[tIndex+12:]
	tEndpoint = tailPart[:strings.Index(tailPart, "\"")]
	tailPart = tailPart[strings.Index(tailPart, "\"")+1:]
	input = input[:tIndex] + tailPart
	input = strings.ReplaceAll(input, ",,", ",")
	input = strings.ReplaceAll(input, ",}", "}")
	output = input
	return
}

func GetEndpointMetricByEndpointType(endpointType string) (err error, result []*m.OptionModel) {
	var endpointMetricTable []*m.EndpointMetricTable
	err = x.SQL("select distinct metric from endpoint_metric where endpoint_id in (select id from endpoint where export_type=?)", endpointType).Find(&endpointMetricTable)
	result = []*m.OptionModel{}
	for _, v := range endpointMetricTable {
		if v.Metric[len(v.Metric)-1:] == "}" {
			result = append(result, &m.OptionModel{OptionText: v.Metric, OptionValue: fmt.Sprintf("%s,instance=\"$address\"}", v.Metric[:len(v.Metric)-1])})
		} else {
			result = append(result, &m.OptionModel{OptionText: v.Metric, OptionValue: fmt.Sprintf("%s{instance=\"$address\"}", v.Metric)})
		}
	}
	return
}

func GetMainCustomDashboard(roleList []string) (err error, result []*m.CustomDashboardTable) {
	result = []*m.CustomDashboardTable{}
	var ids []int
	var idMap = make(map[int]bool)
	if len(roleList) == 0 {
		return
	}
	var customDashboardRows []*m.CustomDashboardTable
	err = x.SQL("select id,name from custom_dashboard").Find(&customDashboardRows)
	if err != nil {
		return
	}
	dashboardIdNameMap := make(map[int]string)
	for _, row := range customDashboardRows {
		dashboardIdNameMap[row.Id] = row.Name
	}
	roleFilterSql, roleFilterParam := createListParams(roleList, "")
	sql := "SELECT custom_dashboard FROM main_dashboard WHERE role_id in (" + roleFilterSql + ") order by guid asc"
	if err = x.SQL(sql, roleFilterParam...).Find(&ids); err != nil {
		return
	}
	if len(ids) > 0 {
		for _, id := range ids {
			if _, ok := idMap[id]; !ok {
				idMap[id] = true
				result = append(result, &m.CustomDashboardTable{Id: id, Name: dashboardIdNameMap[id]})
			}
		}
	}
	return
}

func GetEndpointsByIp(ipList []string, exportType string) (err error, endpoints []m.EndpointTable) {
	sql := fmt.Sprintf("SELECT * FROM endpoint WHERE export_type='%s' AND ip IN ('%s')", exportType, strings.Join(ipList, "','"))
	err = x.SQL(sql).Find(&endpoints)
	return err, endpoints
}

func UpdateChartTitle(param m.UpdateChartTitleParam) error {
	var chartTables []*m.ChartTable
	x.SQL("SELECT id,group_id,metric,title FROM chart").Find(&chartTables)
	if len(chartTables) == 0 {
		return fmt.Errorf("Chart table can not find any data ")
	}
	var chartExist, titleAuto bool
	var groupId int
	for _, v := range chartTables {
		if v.Id == param.ChartId {
			chartExist = true
			groupId = v.GroupId
			if v.Title == "${auto}" {
				titleAuto = true
			}
			break
		}
	}
	if !chartExist {
		return fmt.Errorf("Chart id %d can not find any record ", param.ChartId)
	}
	var err error
	if titleAuto {
		if !strings.Contains(param.Metric, "=") {
			return fmt.Errorf("Parameter metric is illegal ")
		}
		tmpMetric := strings.Split(param.Metric, "=")[1]
		tmpExistId := 0
		for _, v := range chartTables {
			if v.GroupId == groupId && v.Metric == tmpMetric {
				tmpExistId = v.Id
				break
			}
		}
		if tmpExistId > 0 {
			_, err = x.Exec("UPDATE chart SET title=? WHERE id=?", param.Name, tmpExistId)
		} else {
			_, err = x.Exec("INSERT INTO chart(group_id,metric,title,legend) VALUE (?,?,?,?)", groupId, tmpMetric, param.Name, "$metric")
		}
	} else {
		_, err = x.Exec("UPDATE chart SET title=? WHERE id=?", param.Name, param.ChartId)
	}
	return err
}

func UpdateServiceMetricTitle(param m.UpdateChartTitleParam) error {
	var err error
	var key, endpoint string
	var matchMetricGuidList []string
	if !strings.Contains(param.Metric, "/") {
		return fmt.Errorf("Try to match service metric fail,metric illegal ")
	}
	tags := param.Metric[strings.Index(param.Metric, "/")+1:]
	for _, v := range strings.Split(tags, ",") {
		if strings.HasPrefix(v, "key=") {
			key = v[4:]
			continue
		}
		if strings.HasPrefix(v, "t_endpoint=") {
			endpoint = v[11:]
			continue
		}
	}
	if strings.HasPrefix(param.Metric, m.LogMetricName) {
		var logMetricTable []*m.LogMetricConfigTable
		x.SQL("select guid,display_name from log_metric_config where metric=? and log_metric_monitor in (select log_metric_monitor from log_metric_endpoint_rel where target_endpoint=?) union select guid,display_name from log_metric_config where metric=? and log_metric_json in (select guid from log_metric_json where log_metric_monitor in (select log_metric_monitor from log_metric_endpoint_rel where target_endpoint=?))", key, endpoint, key, endpoint).Find(&logMetricTable)
		if len(logMetricTable) > 0 {
			for _, v := range logMetricTable {
				matchMetricGuidList = append(matchMetricGuidList, v.Guid)
			}
			_, err = x.Exec("update log_metric_config set display_name=? where guid in ('"+strings.Join(matchMetricGuidList, "','")+"')", param.Name)
		}
	} else if strings.HasPrefix(param.Metric, m.DBMonitorMetricName) {
		var dbMetricTable []*m.DbMetricMonitorTable
		x.SQL("select guid,display_name from db_metric_monitor where metric=? and guid in (select db_metric_monitor from db_metric_endpoint_rel where target_endpoint=?)", key, endpoint).Find(&dbMetricTable)
		if len(dbMetricTable) > 0 {
			for _, v := range dbMetricTable {
				matchMetricGuidList = append(matchMetricGuidList, v.Guid)
			}
			_, err = x.Exec("update db_metric_monitor set display_name=? where guid in ('"+strings.Join(matchMetricGuidList, "','")+"')", param.Name)
		}
	} else {
		err = fmt.Errorf("Can not match service metric ")
	}
	return err
}

func GetChartTitle(metric string, id int) string {
	if metric == "" {
		return ""
	}
	var chartTables []*m.ChartTable
	x.SQL("SELECT id,metric,title FROM chart WHERE metric=? AND group_id IN (SELECT group_id FROM chart WHERE id=?)", metric, id).Find(&chartTables)
	if len(chartTables) == 0 {
		return ""
	}
	return chartTables[0].Title
}

func GetArchiveData(query *m.QueryMonitorData, agg string) (err error, step int, result []*m.SerialModel) {
	if !ArchiveEnable {
		err = fmt.Errorf("please make sure archive mysql connect done")
		log.Error(nil, log.LOGGER_APP, "", zap.Error(err))
		return err, step, result
	}
	checkArchiveDatabase()
	if query.Start == 0 || query.End == 0 || (query.Start >= query.End) {
		err = fmt.Errorf("get archive data query start and end validate fail,start:%d end:%d ", query.Start, query.End)
		log.Error(nil, log.LOGGER_APP, "", zap.Error(err))
		return err, step, result
	}
	log.Debug(nil, log.LOGGER_APP, "Start to get archive data", zap.Strings("endpoint", query.Endpoint), zap.Strings("metric", query.Metric), zap.Int64("start", query.Start), zap.Int64("end", query.End))
	if agg == "" || agg == "none" {
		agg = "avg"
	}
	step = 60
	if query.Start < time.Now().Unix()-(m.Config().ArchiveMysql.FiveMinStartDay*86400) {
		step = 300
	}
	dateStringList := getDateStringList(query.Start, query.End)
	tagLength := 0
	if len(query.Endpoint) > 1 || len(query.Metric) > 1 {
		tagLength = 2
	}
	for _, endpoint := range query.Endpoint {
		for _, metric := range query.Metric {
			var tmpTag string
			tmpMetric := metric
			if strings.Contains(tmpMetric, "/") {
				tmpTag = tmpMetric[strings.Index(tmpMetric, "/")+1:] + "\""
				tmpTag = strings.Replace(tmpTag, "=", "=\"", -1)
				tmpMetric = metric[:strings.Index(metric, "/")]
			}
			tmpQueryResult := queryArchiveTables(endpoint, tmpMetric, tmpTag, agg, dateStringList, query, tagLength)
			result = append(result, tmpQueryResult...)
		}
	}
	return err, step, result
}

func getDateStringList(start, end int64) []string {
	var dateList []string
	cursorTime := start
	for {
		if cursorTime > end {
			break
		}
		dateList = append(dateList, time.Unix(cursorTime, 0).Format("2006_01_02"))
		cursorTime += 86400
	}
	t, _ := time.Parse("2006_01_02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, time.Unix(cursorTime, 0).Format("2006_01_02")))
	if end > t.Unix()+60 {
		dateList = append(dateList, time.Unix(cursorTime, 0).Format("2006_01_02"))
	}
	return dateList
}

func queryArchiveTables(endpoint, metric, tag, agg string, dateList []string, query *m.QueryMonitorData, tagLength int) []*m.SerialModel {
	var result []*m.SerialModel
	query.Endpoint = []string{endpoint}
	query.Metric = []string{metric}
	resultMap := make(map[string]m.DataSort)
	recordTagMap := make(map[string]map[string]string)
	recordNameMap := make(map[string]string)
	for i, v := range dateList {
		var tmpStart, tmpEnd int64
		if i == 0 {
			tmpStart = query.Start
		} else {
			tmpT, err := time.Parse("2006_01_02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, v))
			if err == nil {
				tmpStart = tmpT.Unix()
			} else {
				continue
			}
		}
		if i == len(v)-1 {
			tmpEnd = query.End
		} else {
			tmpT, err := time.Parse("2006_01_02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, v))
			if err == nil {
				tmpEnd = tmpT.Unix() + 86400
			} else {
				continue
			}
		}
		// 分页查询，避免一次性加载全部行
		sqlStr := fmt.Sprintf("SELECT `endpoint`,metric,tags,unix_time,`avg` AS `value`  FROM archive_%s WHERE `endpoint`='%s' AND metric='%s' AND unix_time>=%d AND unix_time<=%d", v, endpoint, metric, tmpStart, tmpEnd)
		startAt := time.Now()
		totalRows := 0
		page := 1
		pageSize := 1000
		// 仅分页拉取，组装在循环后统一处理
		var allRows []*m.ArchiveQueryTable
		for {
			pageStart := time.Now()
			var tableData []*m.ArchiveQueryTable
			offset := (page - 1) * pageSize
			pageSql := fmt.Sprintf("%s ORDER BY unix_time ASC LIMIT %d OFFSET %d", sqlStr, pageSize, offset)
			err := archiveMysql.SQL(pageSql).Find(&tableData)
			if err != nil {
				if strings.Contains(err.Error(), "doesn't exist") {
					log.Debug(nil, log.LOGGER_APP, fmt.Sprintf("Query archive table:archive_%s error,table doesn't exist", v))
				} else {
					log.Info(nil, log.LOGGER_APP, fmt.Sprintf("query archive table:archive_%s error", v), zap.Error(err))
				}
				break
			}
			if len(tableData) == 0 {
				break
			}
			allRows = append(allRows, tableData...)
			totalRows += len(tableData)
			log.Info(nil, log.LOGGER_APP, "archive page query",
				zap.String("table", fmt.Sprintf("archive_%s", v)),
				zap.Int("page", page),
				zap.Int("page_size", pageSize),
				zap.Int("rows", len(tableData)),
				zap.Float64("cost", time.Since(pageStart).Seconds()),
			)
			if len(tableData) < pageSize {
				break
			}
			page++
		}
		log.Info(nil, log.LOGGER_APP, "archive sql query",
			zap.String("table", fmt.Sprintf("archive_%s", v)),
			zap.Float64("cost", time.Since(startAt).Seconds()),
			zap.String("sql", sqlStr),
			zap.Int("rows", totalRows),
		)
		if totalRows == 0 {
			log.Info(nil, log.LOGGER_APP, fmt.Sprintf("query archive table:archive_%s empty", v))
			continue
		}
		// 在统一组装前判断是否需要扩展 tagLength
		if tagLength <= 1 && i == 0 {
			tmpTagString := allRows[0].Tags
			for _, vv := range allRows {
				if vv.Tags != tmpTagString {
					tagLength = 2
					break
				}
			}
		}
		// 统一处理累计的数据
		for _, rowData := range allRows {
			if !filterByTagValues(rowData.Tags, query.TagValues) {
				continue
			}
			if _, b := recordNameMap[rowData.Tags]; !b {
				if _, b := recordTagMap[rowData.Tags]; !b {
					recordTagMap[rowData.Tags] = getKVMapFromArchiveTags(rowData.Tags)
				}
				recordNameMap[rowData.Tags] = datasource.GetSerialName(query, recordTagMap[rowData.Tags], tagLength, query.CustomDashboard)
			}
			tmpRowName := recordNameMap[rowData.Tags]
			if _, b := resultMap[tmpRowName]; !b {
				resultMap[tmpRowName] = m.DataSort{[]float64{float64(rowData.UnixTime) * 1000, rowData.Value}}
			} else {
				resultMap[tmpRowName] = append(resultMap[tmpRowName], []float64{float64(rowData.UnixTime) * 1000, rowData.Value})
			}
		}
	}
	for k, v := range resultMap {
		sort.Sort(v)
		result = append(result, &m.SerialModel{Name: k, Data: v})
	}
	return result
}

func getKVMapFromArchiveTags(tag string) map[string]string {
	tMap := make(map[string]string)
	for _, v := range strings.Split(tag, ",") {
		kv := strings.Split(v, "=\"")
		tMap[kv[0]] = kv[1][:len(kv[1])-1]
	}
	return tMap
}

func GetAutoDisplay(businessMonitorMap map[int][]string, tagKey string, charts []*m.ChartTable) (result []*m.ChartModel, fetch bool) {
	result = []*m.ChartModel{}
	if len(charts) == 0 {
		return result, false
	}
	if tagKey == "" {
		return result, false
	}
	for endpointId, paths := range businessMonitorMap {
		if len(paths) == 0 {
			continue
		}
		endpointObj := m.EndpointTable{Id: endpointId}
		GetEndpoint(&endpointObj)
		if endpointObj.Guid == "" {
			continue
		}
		_, promQl := GetPromMetric([]string{endpointObj.Guid}, charts[0].Metric)
		if promQl == "" {
			continue
		}
		tmpLegend := charts[0].Legend
		if paths[0] != "" {
			tmpLegend = "$custom_all"
		}
		sm := datasource.PrometheusData(&m.QueryMonitorData{Start: time.Now().Unix() - 300, End: time.Now().Unix(), PromQ: promQl, Legend: tmpLegend, Metric: []string{charts[0].Metric}, Endpoint: []string{endpointObj.Guid}})
		for _, v := range sm {
			for _, path := range paths {
				if path != "" {
					if !strings.Contains(v.Name, path) {
						continue
					}
				}
				chartDto := m.ChartModel{Id: charts[0].Id, Col: charts[0].Col}
				chartDto.Url = `/dashboard/chart`
				chartDto.Endpoint = []string{endpointObj.Guid}
				tmpName := v.Name
				if strings.Contains(tmpName, ":") {
					tmpName = tmpName[strings.Index(tmpName, ":")+1:]
				}
				if path != "" && strings.Contains(tmpName, tagKey+"=") {
					tmpName = strings.Split(tmpName, tagKey+"=")[1]
					if strings.Contains(tmpName, ",") {
						tmpName = strings.Split(tmpName, ",")[0]
					} else {
						tmpName = strings.Split(tmpName, "}")[0]
					}
				}
				chartDto.Metric = []string{fmt.Sprintf("%s/%s=%s", charts[0].Metric, tagKey, tmpName)}
				result = append(result, &chartDto)
			}
		}
	}

	return result, true
}

func GetDashboardPanelList(endpointType, searchMetric string) m.PanelResult {
	returnObj := m.PanelResult{}
	var result []*m.PanelResultObj
	var panelChartQuery []*m.PanelChartQueryObj
	err := x.SQL("select t2.id,t2.tags_key,t2.title,t3.group_id,t3.metric,t3.title as chart_title,t3.unit as chart_unit from dashboard t1 left join panel t2 on t1.panels_group=t2.group_id left join chart t3 on t2.chart_group=t3.group_id where t1.dashboard_type=?", endpointType).Find(&panelChartQuery)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get dashboard panel chart list error", zap.String("type", endpointType), zap.Error(err))
	}
	if len(panelChartQuery) == 0 {
		return returnObj
	}
	tmpPanelGroupId := panelChartQuery[0].GroupId
	tmpChartList := []*m.PanelResultChartObj{}
	for i, v := range panelChartQuery {
		if tmpPanelGroupId != v.GroupId {
			if panelChartQuery[i-1].Title != "Business" {
				result = append(result, &m.PanelResultObj{GroupId: tmpPanelGroupId, PanelTitle: panelChartQuery[i-1].Title, TagsKey: panelChartQuery[i-1].TagsKey, Charts: tmpChartList})
			}
			tmpPanelGroupId = v.GroupId
			tmpChartList = []*m.PanelResultChartObj{}
		}
		tmpChartObj := m.PanelResultChartObj{Metric: v.Metric, Title: v.ChartTitle, Unit: v.ChartUnit}
		metricContainFlag := false
		for _, tmpMetricObj := range strings.Split(v.Metric, "^") {
			if tmpMetricObj == searchMetric {
				metricContainFlag = true
				returnObj.PanelGroupId = v.GroupId
				returnObj.ActiveChart = m.PanelResultChartObj{Metric: v.Metric, Title: v.ChartTitle, Unit: v.ChartUnit, Active: true}
				break
			}
		}
		tmpChartObj.Active = metricContainFlag
		tmpChartList = append(tmpChartList, &tmpChartObj)
	}
	if len(tmpChartList) > 0 {
		if panelChartQuery[len(panelChartQuery)-1].Title != "Business" {
			result = append(result, &m.PanelResultObj{GroupId: tmpPanelGroupId, PanelTitle: panelChartQuery[len(panelChartQuery)-1].Title, TagsKey: panelChartQuery[len(panelChartQuery)-1].TagsKey, Charts: tmpChartList})
		}
	}
	returnObj.PanelList = result
	return returnObj
}

// SearchHostDistinct 查询 endpoint 某个字段的去重集合，不加 limit，只查一个字段
func SearchHostDistinct(field string) ([]string, error) {
	if field == "" {
		return nil, fmt.Errorf("field is empty")
	}
	var result []string
	sql := fmt.Sprintf("SELECT DISTINCT %s FROM endpoint", field)
	err := x.SQL(sql).Find(&result)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "SearchHostDistinct fail", zap.Error(err))
		return nil, err
	}
	return result, nil
}

// SearchHostDistinctField 查询 endpoint 表 export_type 字段的去重集合，不加 limit
func SearchHostDistinctField() ([]string, error) {
	var result []string
	sql := "SELECT DISTINCT export_type FROM endpoint"
	err := x.SQL(sql).Find(&result)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "SearchHostDistinctField fail", zap.Error(err))
		return nil, err
	}
	return result, nil
}

// SearchRecursivePanelDistinctField 查询 panel_recursive 表 obj_type 字段的去重集合，不加 limit
func SearchRecursivePanelDistinctField() ([]string, error) {
	var result []string
	sql := "SELECT DISTINCT obj_type FROM panel_recursive"
	err := x.SQL(sql).Find(&result)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "SearchRecursivePanelDistinctField fail", zap.Error(err))
		return nil, err
	}
	return result, nil
}

// filterByTagValues 根据 TagValues 过滤数据
// archiveTags 格式: agg="count",code="monitor_agent_kubernetes_cluster_update",key="test20250901195_req_count",service_group="gateway_log",retcode="success"
// tagValues 包含 TagName 和 Equal (in/notin) 以及 TagValue 数组
func filterByTagValues(archiveTags string, tagValues []*m.TagDto) bool {
	if len(tagValues) == 0 {
		return true // 没有过滤条件，返回 true
	}

	// 解析 archiveTags 为 map
	archiveTagMap := getKVMapFromArchiveTags(archiveTags)

	// 检查每个 TagValue 条件
	for _, tagValue := range tagValues {
		if tagValue.TagName == "" || len(tagValue.TagValue) == 0 {
			continue // 跳过空的过滤条件
		}

		// 获取 archiveTags 中对应 TagName 的值
		archiveValue, exists := archiveTagMap[tagValue.TagName]
		if !exists {
			// 如果 archiveTags 中没有这个 TagName，根据 Equal 类型判断
			if tagValue.Equal == "in" {
				return false // in 类型要求必须存在，不存在则过滤掉
			}
			continue // notin 类型不存在则通过
		}

		// 检查 archiveValue 是否在 TagValue 列表中
		found := false
		for _, expectedValue := range tagValue.TagValue {
			if archiveValue == expectedValue {
				found = true
				break
			}
		}

		// 根据 Equal 类型判断是否通过
		if tagValue.Equal == "in" {
			if !found {
				return false // in 类型要求必须匹配其中一个值
			}
		} else if tagValue.Equal == "notin" {
			if found {
				return false // notin 类型要求不能匹配任何值
			}
		}
	}

	return true // 所有条件都通过
}
