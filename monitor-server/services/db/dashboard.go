package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"sort"
	"strings"
	"time"
)

func GetDashboard(dType string) (error, m.DashboardTable) {
	var dashboards []*m.DashboardTable
	err := x.SQL("select * from dashboard where dashboard_type=?", dType).Find(&dashboards)
	if len(dashboards) > 0 {
		return nil, *dashboards[0]
	}else{
		if err == nil {
			err = fmt.Errorf("No rows fetch ")
		}
		log.Logger.Error("Query dashboard fail", log.Error(err))
		return err, m.DashboardTable{}
	}
}

func GetSearch(id int) (error, m.SearchModel) {
	var search []*m.SearchModel
	sql := `select * from search where id=?`
	err := x.SQL(sql, id).Find(&search)
	if err!=nil {
		log.Logger.Error("Query search fail", log.Error(err))
	}
	if len(search) > 0 {
		return err, *search[0]
	}else{
		return fmt.Errorf("no rows fetch"), *new(m.SearchModel)
	}
}

func GetButton(bGroup int) (error, []*m.ButtonModel)  {
	var buttons []*m.ButtonModel
	sql := `select * from button where group_id=?`
	err := x.SQL(sql, bGroup).Find(&buttons)
	if err!=nil {
		log.Logger.Error("Query button fail", log.Error(err))
	}
	if len(buttons) > 0 {
		for _,v := range buttons {
			var options []*m.OptionModel
			cSql := `select * from option where group_id=?`
			cErr := x.SQL(cSql, v.OptionGroup).Find(&options)
			if cErr == nil {
				v.Options = options
			}
		}
		return err, buttons
	}else{
		return fmt.Errorf("no rows fetch"), buttons
	}
}

func GetPanels(pGroup int) (error, []*m.PanelTable) {
	var panels []*m.PanelTable
	sql := `select * from panel where group_id=?`
	err := x.SQL(sql, pGroup).Find(&panels)
	if err!=nil {
		log.Logger.Error("Query panels fail", log.Error(err))
	}
	return err, panels
}

func GetCharts(cGroup int, chartId int, panelId int) (error, []*m.ChartTable) {
	var charts []*m.ChartTable
	sql := ``
	var err error
	if cGroup > 0{
		sql = `select * from chart where group_id=?`
		err = x.SQL(sql, cGroup).Find(&charts)
	}else if chartId > 0{
		sql = `select * from chart where id=?`
		err = x.SQL(sql, chartId).Find(&charts)
	}else if panelId > 0{
		sql = `SELECT t1.* FROM chart t1 INNER JOIN panel t2 ON t1.group_id=t2.chart_group WHERE t2.id=?`
		err = x.SQL(sql, panelId).Find(&charts)
	}
	if err!=nil {
		log.Logger.Error("Query charts fail", log.Error(err))
	}
	if len(charts) > 0 {
		return err,charts
	}
	return fmt.Errorf("Get charts error "), charts
}

func GetPromMetric(endpoint []string,metric string) (error, string) {
	promQL := metric
	tmpMetric := metric
	var tmpTag string
	if strings.Contains(tmpMetric, "/") {
		tmpMetric = metric[:strings.Index(metric, "/")]
		tmpTag = metric[strings.Index(metric, "/")+1:]
	}
	var query []*m.PromMetricTable
	err := x.SQL("SELECT prom_ql FROM prom_metric WHERE metric=?", tmpMetric).Find(&query)
	if err!=nil {
		log.Logger.Error("Query prom_metric fail", log.Error(err))
	}
	if len(query) > 0 {
		host := m.EndpointTable{Guid:endpoint[0]}
		GetEndpoint(&host)
		if err!=nil || host.Id==0 {
			log.Logger.Error("Find endpoint fail", log.String("endpoint",endpoint[0]),log.Error(err))
			return err,promQL
		}
		reg := query[0].PromQl
		if strings.Contains(reg, `$ip`) {
			reg = strings.Replace(reg, "$ip", host.Ip, -1)
		}
		if strings.Contains(reg, `$address`) {
			if tmpTag != "" {
				tagAppendString := ""
				for _,tagString := range strings.Split(tmpTag, ",") {
					tagKeyValue := strings.Split(tagString, "=")
					if len(tagKeyValue) == 2 {
						tagAppendString += fmt.Sprintf(",%s=\"%s\"", tagKeyValue[0], tagKeyValue[1])
					}
				}
				reg = strings.ReplaceAll(reg, "\"$address\"", fmt.Sprintf("\"$address\"%s", tagAppendString))
			}
			if host.AddressAgent != "" {
				reg = strings.Replace(reg, "$address", host.AddressAgent, -1)
			}else {
				reg = strings.Replace(reg, "$address", host.Address, -1)
			}
		}
		if strings.Contains(reg, `$guid`) {
			reg = strings.Replace(reg, "$guid", host.Guid, -1)
		}
		if strings.Contains(reg, "$pod") {
			reg = strings.Replace(reg, "$pod", host.Name, -1)
		}
		if strings.Contains(reg, "$k8s_cluster") {
			reg = strings.Replace(reg, "$k8s_cluster", host.OsType, -1)
		}
		promQL = reg
	}
	return err,promQL
}

func GetDbPromMetric(endpoint,metric,legend string) (error,string) {
	promQL := "db_monitor_count"
	var query []*m.PromMetricTable
	err := x.SQL("SELECT prom_ql FROM prom_metric WHERE metric='db_monitor_count'").Find(&query)
	if err!=nil {
		log.Logger.Error("Query prom_metric fail", log.Error(err))
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
	log.Logger.Debug("db prom metric", log.String("promQl", promQL))
	return err,promQL
}

func SearchHost(endpoint string) (error, []*m.OptionModel) {
	options := []*m.OptionModel{}
	var hosts []*m.EndpointTable
	endpoint = `%` + endpoint + `%`
	err := x.SQL("SELECT * FROM endpoint WHERE (ip LIKE ? OR name LIKE ?) and export_type<>'custom' order by export_type,ip limit 100", endpoint, endpoint).Find(&hosts)
	if err != nil {
		log.Logger.Error("Search host fail", log.Error(err))
		return err,options
	}
	for _,host := range hosts {
		if host.ExportType == "node" {
			host.ExportType = "host"
		}
		//options = append(options, &m.OptionModel{OptionText: fmt.Sprintf("%s:%s", host.Name, host.Ip), OptionValue: fmt.Sprintf("%s:%s", host.Guid, host.ExportType), Id:host.Id})
		options = append(options, &m.OptionModel{OptionText: fmt.Sprintf("%s:%s", host.Name, host.Ip), OptionValue: host.Guid, OptionType:host.ExportType, Id:host.Id})
	}
	return err,options
}

func GetEndpoint(query *m.EndpointTable) error {
	var endpointObj []*m.EndpointTable
	var err error
	if query.Id > 0 {
		err = x.SQL("SELECT * FROM endpoint WHERE id=?", query.Id).Find(&endpointObj)
	}else if query.Guid != ""{
		err = x.SQL("SELECT * FROM endpoint WHERE guid=?", query.Guid).Find(&endpointObj)
	}else if query.Address != "" {
		if query.AddressAgent != "" {
			err = x.SQL("SELECT * FROM endpoint WHERE address=? or address_agent=?", query.Address, query.AddressAgent).Find(&endpointObj)
		}else {
			err = x.SQL("SELECT * FROM endpoint WHERE address=?", query.Address).Find(&endpointObj)
		}
	}
	if query.Ip != "" && query.ExportType != "" {
		if query.Name == "" {
			err = x.SQL("SELECT * FROM endpoint WHERE ip=? and export_type=?", query.Ip, query.ExportType).Find(&endpointObj)
		}else{
			err = x.SQL("SELECT * FROM endpoint WHERE ip=? and export_type=? and name=?", query.Ip, query.ExportType, query.Name).Find(&endpointObj)
		}
	}
	if err != nil {
		log.Logger.Error("Get tags fail", log.Error(err))
		return err
	}
	if len(endpointObj) <= 0 {
		return fmt.Errorf("no data")
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
	return nil
}

func ListEndpoint() []*m.EndpointTable {
	var result []*m.EndpointTable
	x.SQL("SELECT * FROM endpoint").Find(&result)
	return result
}

func GetTags(endpoint string, key string, metric string) (error, []*m.OptionModel) {
	var options []*m.OptionModel
	var endpointObj []*m.EndpointTable
	err := x.SQL("SELECT id FROM endpoint WHERE guid=?", endpoint).Find(&endpointObj)
	if err != nil || len(endpointObj) <= 0{
		log.Logger.Error("Get tags fail, can't find endpoint", log.Error(err))
		return err,options
	}
	var promMetricObj []*m.PromMetricTable
	err = x.SQL("SELECT prom_main FROM prom_metric WHERE metric=?", metric).Find(&promMetricObj)
	if err != nil || len(promMetricObj) <=0{
		log.Logger.Error("Get tags fail,can't find prom_metric", log.Error(err))
		return err,options
	}
	var endpointMetricObj []*m.EndpointMetricTable
	err = x.SQL("SELECT metric FROM endpoint_metric WHERE endpoint_id=? AND metric LIKE ?", endpointObj[0].Id, `%`+promMetricObj[0].PromMain+`%`).Find(&endpointMetricObj)
	if err != nil || len(endpointMetricObj) <=0{
		log.Logger.Error("Get tags fail,can't find metric", log.Error(err))
		return err,options
	}
	key = key + "="
	for _,c := range endpointMetricObj {
		if !strings.Contains(c.Metric, key) {
			continue
		}
		metricStr := strings.Replace(c.Metric, `'`, `"`, -1)
		tmpStr := strings.Split(metricStr, key+`"`)[1]
		tagStr := strings.Split(tmpStr, `"`)[0]
		inBlacklist := false
		for _,v := range m.Config().TagBlacklist {
			if strings.Contains(tagStr, v) {
				inBlacklist = true
				break
			}
		}
		if inBlacklist {
			continue
		}
		options = append(options, &m.OptionModel{OptionText:tagStr, OptionValue:key+tagStr})
	}
	return err,options
}

func RegisterEndpointMetric(endpointId int,endpointMetrics []string) error {
	if endpointId <= 0 {
		return fmt.Errorf("endpoint id is 0")
	}
	maxNum  := 50
	var actions []*Action
	actions = append(actions, &Action{Sql:"delete FROM endpoint_metric WHERE endpoint_id=?", Param:[]interface{}{endpointId}})
	i := 0
	var insertSql string
	params := make([]interface{}, 0)
	for _,v := range endpointMetrics {
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

func GetPromMetricTable(metricType string) (err error,result []*m.PromMetricUpdateParam) {
	if metricType != "" {
		err = x.SQL("SELECT t1.*,t2.group_id as panel_id FROM prom_metric t1 left join chart t2 on t1.metric=t2.metric WHERE t1.metric_type=? order by t1.id", metricType).Find(&result)
	}else{
		err = x.SQL("SELECT t1.*,t2.group_id as panel_id FROM prom_metric t1 left join chart t2 on t1.metric=t2.metric order by t1.id").Find(&result)
	}
	if err != nil {
		log.Logger.Error("Get prom metric table fail", log.Error(err))
	}
	return err,result
}

func UpdatePromMetric(data []m.PromMetricUpdateParam) error {
	if len(data) == 0 {
		return fmt.Errorf("data is null")
	}
	var insertData,updateData []m.PromMetricTable
	var chartTable []*m.ChartTable
	x.SQL("select * from chart").Find(&chartTable)
	var updateChartAction []*Action
	for _,v := range data {
		if v.Id > 0 {
			updateData = append(updateData, m.PromMetricTable{Id: v.Id,Metric: v.Metric,MetricType: v.MetricType,PromQl: v.PromQl,PromMain: v.PromMain})
		}else{
			insertData = append(insertData, m.PromMetricTable{Metric: v.Metric,MetricType: v.MetricType,PromQl: v.PromQl,PromMain: v.PromMain})
		}
		newChartGroupId := v.PanelId
		if newChartGroupId > 0 {
			var existChartObj,newUpdateChartObj m.ChartTable
			for _,chart := range chartTable {
				for _,tmpMetric := range strings.Split(chart.Metric, "^") {
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
							for _,tmpExistMetricObj := range strings.Split(existChartObj.Metric, "^") {
								if tmpExistMetricObj != v.Metric {
									newOldMetricList = append(newOldMetricList, tmpExistMetricObj)
								}
							}
							updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newOldMetricList, "^"),existChartObj.Id}})
						}else{
							updateChartAction = append(updateChartAction, &Action{Sql: "delete from chart where id=?", Param: []interface{}{existChartObj.Id}})
						}
						updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=?,title=?,unit=? where id=?", Param: []interface{}{fmt.Sprintf("%s^%s",newUpdateChartObj.Metric,v.Metric),v.Chart.Title,v.Chart.Unit,newUpdateChartObj.Id}})
					}
				}else{
					if existChartObj.GroupId != newChartGroupId {
						if strings.Contains(existChartObj.Metric, "^") {
							var newOldMetricList []string
							for _,tmpExistMetricObj := range strings.Split(existChartObj.Metric, "^") {
								if tmpExistMetricObj != v.Metric {
									newOldMetricList = append(newOldMetricList, tmpExistMetricObj)
								}
							}
							updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=? where id=?", Param: []interface{}{strings.Join(newOldMetricList, "^"),existChartObj.Id}})
							updateChartAction = append(updateChartAction, &Action{Sql: "insert into chart(group_id,metric,url,title,unit,legend) value (?,?,'/dashboard/chart',?,?,'$metric')",Param: []interface{}{newChartGroupId,v.Metric,v.Chart.Title,v.Chart.Unit}})
						}else{
							updateChartAction = append(updateChartAction, &Action{Sql: "update chart set group_id=? where id=?", Param: []interface{}{newChartGroupId,existChartObj.Id}})
						}
					}
				}
			}else{
				if newUpdateChartObj.Id > 0 {
					updateChartAction = append(updateChartAction, &Action{Sql: "update chart set metric=?,title=?,unit=? where id=?", Param: []interface{}{fmt.Sprintf("%s^%s",newUpdateChartObj.Metric,v.Metric),v.Chart.Title,v.Chart.Unit,newUpdateChartObj.Id}})
				}else{
					updateChartAction = append(updateChartAction, &Action{Sql: "insert into chart(group_id,metric,url,title,unit,legend) value (?,?,'/dashboard/chart',?,?,'$metric')",Param: []interface{}{newChartGroupId,v.Metric,v.Chart.Title,v.Chart.Unit}})
				}
			}
		}
	}
	var actions []*Action
	for _,v := range insertData {
		action := Classify(v, "insert", "prom_metric", false)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	for _,v := range updateData {
		action := Classify(v, "update", "prom_metric", false)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	if len(actions) > 0 {
		err := Transaction(actions)
		if err != nil {
			return err
		}
		if len(updateChartAction) > 0 {
			for _,v := range updateChartAction {
				log.Logger.Info("Update chart action", log.String("sql", v.Sql), log.String("param", fmt.Sprintf("%v", v.Param)))
			}
			err = Transaction(updateChartAction)
			if err != nil {
				log.Logger.Error("Update chart config fail", log.Error(err))
			}
		}
	}
	return nil
}

func GetEndpointMetric(id int) (err error,result []*m.OptionModel) {
	endpointObj := m.EndpointTable{Id:id}
	GetEndpoint(&endpointObj)
	if endpointObj.Guid == "" {
		return fmt.Errorf("endpoint id: %d can not find ", id),result
	}
	if endpointObj.ExportType == "ping" || endpointObj.ExportType == "telnet" || endpointObj.ExportType == "http" {
		return nil,result
	}
	var ip,port string
	if endpointObj.AddressAgent != "" {
		ip = endpointObj.AddressAgent[:strings.Index(endpointObj.AddressAgent, ":")]
		port = endpointObj.AddressAgent[strings.Index(endpointObj.AddressAgent, ":")+1:]
	}else{
		ip = endpointObj.Address[:strings.Index(endpointObj.Address, ":")]
		port = endpointObj.Address[strings.Index(endpointObj.Address, ":")+1:]
	}
	if ip == "" || port == "" {
		return fmt.Errorf("endpoint: %s address illegal ", endpointObj.Guid),result
	}
	err, strList := prom.GetEndpointData(ip, port, []string{}, []string{})
	if err != nil {
		return err,result
	}
	for _,v := range strList {
		if strings.HasPrefix(v, "go_") || v == "" {
			continue
		}
		if v[len(v)-1:] == "}" {
			result = append(result, &m.OptionModel{OptionText: v, OptionValue: fmt.Sprintf("%s,instance=\"$address\"}", v[:len(v)-1])})
		}else {
			result = append(result, &m.OptionModel{OptionText: v, OptionValue: fmt.Sprintf("%s{instance=\"$address\"}", v)})
		}
	}
	return nil,result
}

func GetMainCustomDashboard(roleList []string) (err error,result []*m.CustomDashboardTable) {
	sql := "SELECT t2.* FROM role t1 LEFT JOIN custom_dashboard t2 ON t1.main_dashboard=t2.id WHERE t1.name IN ('"+strings.Join(roleList, "','")+"') and t1.main_dashboard>0"
	log.Logger.Debug("Get main dashboard", log.String("sql", sql))
	err = x.SQL(sql).Find(&result)
	if len(result) == 0 {
		result = []*m.CustomDashboardTable{}
	}
	return err,result
}

func GetEndpointsByIp(ipList []string, exportType string) (err error,endpoints []m.EndpointTable) {
	sql := fmt.Sprintf("SELECT * FROM endpoint WHERE export_type='%s' AND ip IN ('%s')", exportType, strings.Join(ipList, "','"))
	err = x.SQL(sql).Find(&endpoints)
	return err,endpoints
}

func UpdateChartTitle(param m.UpdateChartTitleParam) error {
	var chartTables []*m.ChartTable
	x.SQL("SELECT id,group_id,metric,title FROM chart").Find(&chartTables)
	if len(chartTables) == 0 {
		return fmt.Errorf("Chart table can not find any data ")
	}
	var chartExist,titleAuto bool
	var groupId int
	for _,v := range chartTables {
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
		for _,v := range chartTables {
			if v.GroupId == groupId && v.Metric == tmpMetric {
				tmpExistId = v.Id
				break
			}
		}
		if tmpExistId > 0 {
			_,err = x.Exec("UPDATE chart SET title=? WHERE id=?", param.Name, tmpExistId)
		}else{
			_,err = x.Exec("INSERT INTO chart(group_id,metric,title,legend) VALUE (?,?,?,?)", groupId, tmpMetric, param.Name, "$metric")
		}
	}else{
		_,err = x.Exec("UPDATE chart SET title=? WHERE id=?", param.Name, param.ChartId)
	}
	return err
}

func GetChartTitle(metric string,id int) string {
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

func GetArchiveData(query *m.QueryMonitorData,agg string) (err error,step int,result []*m.SerialModel) {
	if !ArchiveEnable {
		err = fmt.Errorf("please make sure archive mysql connect done")
		log.Logger.Error("", log.Error(err))
		return err,step,result
	}
	checkArchiveDatabase()
	if query.Start == 0 || query.End == 0 || (query.Start>=query.End) {
		err = fmt.Errorf("get archive data query start and end validate fail,start:%d end:%d ", query.Start, query.End)
		log.Logger.Error("", log.Error(err))
		return err,step,result
	}
	log.Logger.Info("Start to get archive data", log.StringList("endpoint",query.Endpoint),log.StringList("metric",query.Metric),log.Int64("start",query.Start),log.Int64("end",query.End))
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
	for _,endpoint := range query.Endpoint {
		for _,metric := range query.Metric {
			var tmpTag string
			tmpMetric := metric
			if strings.Contains(tmpMetric, "/") {
				tmpTag = tmpMetric[strings.Index(tmpMetric, "/")+1:]+"\""
				tmpTag = strings.Replace(tmpTag, "=", "=\"", -1)
				tmpMetric = metric[:strings.Index(metric, "/")]
			}
			tmpQueryResult := queryArchiveTables(endpoint, tmpMetric, tmpTag, agg, dateStringList, query, tagLength)
			result = append(result, tmpQueryResult...)
		}
	}
	return err,step,result
}

func getDateStringList(start,end int64) []string {
	var dateList []string
	cursorTime := start
	for {
		if cursorTime > end {
			break
		}
		dateList = append(dateList, time.Unix(cursorTime, 0).Format("2006_01_02"))
		cursorTime += 86400
	}
	t,_ := time.Parse("2006_01_02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, time.Unix(cursorTime, 0).Format("2006_01_02")))
	if end > t.Unix()+60 {
		dateList = append(dateList, time.Unix(cursorTime, 0).Format("2006_01_02"))
	}
	return dateList
}

func queryArchiveTables(endpoint,metric,tag,agg string,dateList []string,query *m.QueryMonitorData,tagLength int) []*m.SerialModel {
	var result []*m.SerialModel
	query.Endpoint = []string{endpoint}
	query.Metric = []string{metric}
	resultMap := make(map[string]m.DataSort)
	recordTagMap := make(map[string]map[string]string)
	recordNameMap := make(map[string]string)
	for i,v := range dateList {
		var tmpStart,tmpEnd int64
		if i == 0 {
			tmpStart = query.Start
		}else{
			tmpT,err := time.Parse("2006_01_02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, v))
			if err == nil {
				tmpStart = tmpT.Unix()
			}else{
				continue
			}
		}
		if i == len(v)-1 {
			tmpEnd = query.End
		}else{
			tmpT,err := time.Parse("2006_01_02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, v))
			if err == nil {
				tmpEnd = tmpT.Unix()+86400
			}else{
				continue
			}
		}
		var tableData []*m.ArchiveQueryTable
		err := archiveMysql.SQL(fmt.Sprintf("SELECT `endpoint`,metric,tags,unix_time,`avg` AS `value`  FROM archive_%s WHERE `endpoint`='%s' AND metric='%s' AND unix_time>=%d AND unix_time<=%d", v,endpoint,metric,tmpStart,tmpEnd)).Find(&tableData)
		if err != nil {
			if strings.Contains(err.Error(), "doesn't exist") {
				log.Logger.Warn(fmt.Sprintf("Query archive table:archive_%s error,table doesn't exist", v))
			}else {
				log.Logger.Error(fmt.Sprintf("query archive table:archive_%s error", v), log.Error(err))
			}
			continue
		}
		if len(tableData) == 0 {
			log.Logger.Info(fmt.Sprintf("query archive table:archive_%s empty", v))
			continue
		}
		if tagLength <= 1 && i == 0 {
			tmpTagString := tableData[0].Tags
			for _,vv := range tableData {
				if vv.Tags != tmpTagString {
					tagLength = 2
					break
				}
			}
		}
		for _,rowData := range tableData {
			if tag != "" {
				if !strings.Contains(rowData.Tags, tag) {
					continue
				}
			}
			if _,b := recordNameMap[rowData.Tags];!b {
				if _,b := recordTagMap[rowData.Tags];!b {
					recordTagMap[rowData.Tags] = getKVMapFromArchiveTags(rowData.Tags)
				}
				recordNameMap[rowData.Tags] = datasource.GetSerialName(query, recordTagMap[rowData.Tags], tagLength)
			}
			tmpRowName := recordNameMap[rowData.Tags]
			if _,b := resultMap[tmpRowName];!b {
				resultMap[tmpRowName] = m.DataSort{[]float64{float64(rowData.UnixTime)*1000,rowData.Value}}
			}else{
				resultMap[tmpRowName] = append(resultMap[tmpRowName], []float64{float64(rowData.UnixTime)*1000, rowData.Value})
			}
		}
	}
	for k,v := range resultMap {
		sort.Sort(v)
		result = append(result, &m.SerialModel{Name:k, Data:v})
	}
	return result
}

func getKVMapFromArchiveTags(tag string) map[string]string {
	tMap := make(map[string]string)
	for _,v := range strings.Split(tag, ",") {
		kv := strings.Split(v, "=\"")
		tMap[kv[0]] = kv[1][:len(kv[1])-1]
	}
	return tMap
}

func GetAutoDisplay(businessMonitorMap map[int][]string,tagKey string,charts []*m.ChartTable) (result []*m.ChartModel,fetch bool) {
	result = []*m.ChartModel{}
	if len(charts) == 0 {
		return result,false
	}
	if tagKey == "" {
		return result,false
	}
	for endpointId,paths := range businessMonitorMap {
		if len(paths) == 0 {
			continue
		}
		endpointObj := m.EndpointTable{Id:endpointId}
		GetEndpoint(&endpointObj)
		if endpointObj.Guid == "" {
			continue
		}
		_,promQl := GetPromMetric([]string{endpointObj.Guid}, charts[0].Metric)
		if promQl == "" {
			continue
		}
		tmpLegend := charts[0].Legend
		if paths[0] != "" {
			tmpLegend = "$custom_all"
		}
		sm := datasource.PrometheusData(&m.QueryMonitorData{Start:time.Now().Unix()-300, End:time.Now().Unix(), PromQ:promQl, Legend:tmpLegend, Metric:[]string{charts[0].Metric}, Endpoint:[]string{endpointObj.Guid}})
		for _,v := range sm {
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
					tmpName = tmpName[strings.Index(tmpName,":")+1:]
				}
				if path != "" && strings.Contains(tmpName, tagKey+"=") {
					tmpName = strings.Split(tmpName, tagKey+"=")[1]
					if strings.Contains(tmpName, ",") {
						tmpName = strings.Split(tmpName, ",")[0]
					}else{
						tmpName = strings.Split(tmpName, "}")[0]
					}
				}
				chartDto.Metric = []string{fmt.Sprintf("%s/%s=%s", charts[0].Metric, tagKey, tmpName)}
				result = append(result, &chartDto)
			}
		}
	}

	return result,true
}

func GetDashboardPanelList(endpointType,searchMetric string) m.PanelResult {
	returnObj := m.PanelResult{}
	var result []*m.PanelResultObj
	var panelChartQuery []*m.PanelChartQueryObj
	err := x.SQL("select t2.id,t2.tags_key,t2.title,t3.group_id,t3.metric,t3.title as chart_title,t3.unit as chart_unit from dashboard t1 left join panel t2 on t1.panels_group=t2.group_id left join chart t3 on t2.chart_group=t3.group_id where t1.dashboard_type=?", endpointType).Find(&panelChartQuery)
	if err != nil {
		log.Logger.Error("Get dashboard panel chart list error", log.String("type", endpointType), log.Error(err))
	}
	if len(panelChartQuery) == 0 {
		return returnObj
	}
	tmpPanelGroupId := panelChartQuery[0].GroupId
	tmpChartList := []*m.PanelResultChartObj{}
	for i,v := range panelChartQuery {
		if v.Title == "Business" {
			continue
		}
		if tmpPanelGroupId != v.GroupId {
			result = append(result, &m.PanelResultObj{GroupId: tmpPanelGroupId,PanelTitle: panelChartQuery[i-1].Title,TagsKey: panelChartQuery[i-1].TagsKey,Charts: tmpChartList})
			tmpPanelGroupId = v.GroupId
			tmpChartList = []*m.PanelResultChartObj{}
		}
		tmpChartObj := m.PanelResultChartObj{Metric: v.Metric, Title: v.ChartTitle, Unit: v.ChartUnit}
		metricContainFlag := false
		for _,tmpMetricObj := range strings.Split(v.Metric, "^") {
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
		result = append(result, &m.PanelResultObj{GroupId: tmpPanelGroupId,PanelTitle: panelChartQuery[len(panelChartQuery)-1].Title,TagsKey: panelChartQuery[len(panelChartQuery)-1].TagsKey,Charts: tmpChartList})
	}
	returnObj.PanelList = result
	return returnObj
}