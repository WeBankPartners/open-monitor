package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"strings"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
)

func GetDashboard(dType string) (error, m.DashboardTable) {
	var dashboards []*m.DashboardTable
	sql := `select * from dashboard where dashboard_type=?`
	err := x.SQL(sql, dType).Find(&dashboards)
	if err!=nil {
		mid.LogError("query dashboard fail", err)
	}
	if len(dashboards) > 0 {
		return err, *dashboards[0]
	}else{
		return fmt.Errorf("no rows fetch"), *new(m.DashboardTable)
	}
}

func GetSearch(id int) (error, m.SearchModel) {
	var search []*m.SearchModel
	sql := `select * from search where id=?`
	err := x.SQL(sql, id).Find(&search)
	if err!=nil {
		mid.LogError("query search fail", err)
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
		mid.LogError("query button fail", err)
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
		mid.LogError("query panels fail", err)
	}
	//if len(panels) > 0 {
	//	var dashboards []*m.DashboardTable
	//	sqlSec := `select dashboard_type from dashboard where panels_group=?`
	//	err := x.SQL(sqlSec, pGroup).Find(&dashboards)
	//	return err,panels
	//}
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
		mid.LogError("query charts fail", err)
	}
	if len(charts) > 0 {
		return err,charts
	}
	return fmt.Errorf("get charts error"), charts
}

func GetPromMetric(endpoint []string,metric string) (error, string) {
	promQL := metric
	tmpMetric := metric
	var tmpTag string
	if strings.Contains(tmpMetric, "/") {
		tmpList := strings.Split(tmpMetric, "/")
		tmpMetric = tmpList[0]
		tmpTag = tmpList[1]
	}
	var query []*m.PromMetricTable
	err := x.SQL("SELECT prom_ql FROM prom_metric WHERE metric=?", tmpMetric).Find(&query)
	if err!=nil {
		mid.LogError("query prom_metric fail", err)
	}
	if len(query) > 0 {
		host := m.EndpointTable{Guid:endpoint[0]}
		GetEndpoint(&host)
		if err!=nil || host.Id==0 {
			mid.LogError("can't find endpoint "+endpoint[0], err)
			return err,promQL
		}
		reg := query[0].PromQl
		if strings.Contains(reg, `$ip`) {
			reg = strings.Replace(reg, "$ip", host.Ip, -1)
		}
		if strings.Contains(reg, `$address`) {
			if strings.Contains(tmpTag, "=") {
				tmpList := strings.Split(tmpTag, "=")
				reg = strings.Replace(reg, "$address\"", fmt.Sprintf("$address\",%s=\"%s\"", tmpList[0], tmpList[1]), -1)
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
		promQL = reg
	}
	return err,promQL
}

func SearchHost(endpoint string) (error, []*m.OptionModel) {
	options := []*m.OptionModel{}
	var hosts []*m.EndpointTable
	endpoint = `%` + endpoint + `%`
	err := x.SQL("SELECT * FROM endpoint WHERE (ip LIKE ? OR name LIKE ?) and export_type<>'custom' order by export_type,ip limit 10", endpoint, endpoint).Find(&hosts)
	if err != nil {
		mid.LogError("search host fail", err)
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
		mid.LogError("get tags fail ", err)
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
		mid.LogError("get tags fail, can't find endpoint ", err)
		return err,options
	}
	var promMetricObj []*m.PromMetricTable
	err = x.SQL("SELECT prom_main FROM prom_metric WHERE metric=?", metric).Find(&promMetricObj)
	if err != nil || len(promMetricObj) <=0{
		mid.LogError("get tags fail,can't find prom_metric ", err)
		return err,options
	}
	var endpointMetricObj []*m.EndpointMetricTable
	err = x.SQL("SELECT metric FROM endpoint_metric WHERE endpoint_id=? AND metric LIKE ?", endpointObj[0].Id, `%`+promMetricObj[0].PromMain+`%`).Find(&endpointMetricObj)
	if err != nil || len(endpointMetricObj) <=0{
		mid.LogError("get tags fail,can't find metric ", err)
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

func GetPromMetricTable(metricType string) (err error,result []*m.PromMetricTable) {
	if metricType != "" {
		err = x.SQL("SELECT * FROM prom_metric WHERE metric_type=?", metricType).Find(&result)
	}else{
		err = x.SQL("SELECT * FROM prom_metric").Find(&result)
	}
	if err != nil {
		mid.LogError("get prom metric table fail", err)
	}
	return err,result
}

func UpdatePromMetric(data []m.PromMetricTable) error {
	if len(data) == 0 {
		return fmt.Errorf("data is null")
	}
	var insertData,updateData []m.PromMetricTable
	for _,v := range data {
		if v.Id > 0 {
			updateData = append(updateData, v)
		}else{
			insertData = append(insertData, v)
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
		return Transaction(actions)
	}
	return nil
}

func GetEndpointMetric(id int) (err error,result []*m.OptionModel) {
	var endpointMetrics []*m.EndpointMetricTable
	err = x.SQL("SELECT id,endpoint_id,metric FROM endpoint_metric WHERE endpoint_id=?", id).Find(&endpointMetrics)
	if err != nil {
		mid.LogError("get endpoint metric fail", err)
	}
	metricMap := make(map[string]string)
	for _,v := range endpointMetrics {
		tmpMetric := v.Metric
		if strings.Contains(v.Metric, "{") {
			tmpMetric = strings.Split(v.Metric, "{")[0]
		}
		metricMap[tmpMetric] = fmt.Sprintf("%s{instance=\"$address\"}", tmpMetric)
	}
	for k,v := range metricMap {
		result = append(result, &m.OptionModel{OptionText:k, OptionValue:v})
	}
	return err,result
}

func GetEndpointMetricNew(id int) (err error,result []*m.OptionModel) {
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

func GetMainCustomDashboard() (error, m.CustomDashboardTable) {
	var result []*m.CustomDashboardTable
	err := x.SQL("SELECT * FROM custom_dashboard WHERE main=1").Find(&result)
	if len(result) > 0 {
		return nil,*result[0]
	}else{
		return err,m.CustomDashboardTable{Id:0}
	}
}

func SetMainCustomDashboard(id int) error {
	var actions []*Action
	err,cdt := GetMainCustomDashboard()
	if cdt.Id > 0 {
		if cdt.Id == id {
			return nil
		}
		actions = append(actions, &Action{Sql:"UPDATE custom_dashboard SET main=0 WHERE id=?", Param:[]interface{}{cdt.Id}})
	}
	actions = append(actions, &Action{Sql:"UPDATE custom_dashboard SET main=1 WHERE id=?", Param:[]interface{}{id}})
	err = Transaction(actions)
	return err
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
		return fmt.Errorf("chart table can not find any data")
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
		return fmt.Errorf("chart id %d can not find any record", param.ChartId)
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