package db

import (
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"fmt"
	"strings"
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

func GetPanels(pGroup int,endpoint string) (error, []*m.PanelTable) {
	var panels []*m.PanelTable
	sql := `select * from panel where group_id=?`
	err := x.SQL(sql, pGroup).Find(&panels)
	if err!=nil {
		mid.LogError("query panels fail", err)
	}
	if len(panels) > 0 {
		var dashboards []*m.DashboardTable
		sqlSec := `select dashboard_type from dashboard where panels_group=?`
		err := x.SQL(sqlSec, pGroup).Find(&dashboards)
		return err,panels
	}
	return fmt.Errorf("get panels error"), panels
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

func GetPromMetric(endpoint,metric []string) (error, string) {
	promQL := ""
	tmpMetric := metric[0]
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
		reg := query[0].PromQl
		if strings.Contains(reg, `$endpoint`) {
			reg = strings.Replace(reg, "$endpoint", endpoint[0], -1)
		}
		if tmpTag != "" {
			tmpList := strings.Split(tmpTag, "=")
			if strings.Contains(reg, `$`+tmpList[0]) {
				reg = strings.Replace(reg, `$`+tmpList[0], tmpList[1], -1)
			}
		}
		promQL = reg
	}else{
		promQL = metric[0]
	}
	return err,promQL
}

func SearchHost(endpoint string) (error, []*m.OptionModel) {
	options := []*m.OptionModel{}
	var hosts []*m.EndpointTable
	endpoint = `%` + endpoint + `%`
	err := x.SQL("SELECT * FROM endpoint WHERE ip LIKE ? OR NAME LIKE ? order by ip limit 10", endpoint, endpoint).Find(&hosts)
	if err != nil {
		mid.LogError("search host fail", err)
		return err,options
	}
	for _,host := range hosts {
		if host.ExportType == "node" {
			host.ExportType = "host"
		}
		options = append(options, &m.OptionModel{OptionText:fmt.Sprintf("%s:%s", host.Name, host.Ip), OptionValue:fmt.Sprintf("%s:%s", host.Guid, host.ExportType)})
	}
	return err,options
}

func GetEndpoint(endpoint string) (error, m.EndpointTable) {
	var endpointObj []*m.EndpointTable
	err := x.SQL("SELECT * FROM endpoint WHERE guid=?", endpoint).Find(&endpointObj)
	if err != nil {
		mid.LogError("get tags fail ", err)
		return err,m.EndpointTable{Id:0}
	}
	if len(endpointObj) <= 0 {
		return nil,m.EndpointTable{Id:0}
	}
	return nil,*endpointObj[0]
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
		options = append(options, &m.OptionModel{OptionText:tagStr, OptionValue:key+tagStr})
	}
	return err,options
}

func RegisterEndpointMetric(endpointId int,endpointMetrics []string) error {
	maxNum  := 100
	var sqls []string
	delSql := fmt.Sprintf("delete FROM endpoint_metric WHERE endpoint_id=%d", endpointId)
	sqls = append(sqls, delSql)
	i := 0
	var insertSql string
	for _,v := range endpointMetrics {
		i = i + 1
		if i == 1 {
			insertSql = "INSERT INTO endpoint_metric(endpoint_id,metric) VALUES "
		}
		insertSql = insertSql + fmt.Sprintf("(%d,'%s'),", endpointId, v)
		if i == maxNum {
			tmpSql := insertSql[:len(insertSql)-1]
			sqls = append(sqls, tmpSql)
			i = 0
		}
	}
	if i > 0 {
		sqls = append(sqls, insertSql[:len(insertSql)-1])
	}
	err := ExecuteTransactionSql(sqls)
	return err
}