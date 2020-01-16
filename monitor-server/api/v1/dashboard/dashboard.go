package dashboard

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"
	"encoding/json"
	"regexp"
	"io/ioutil"
)

// @Summary 页面通用接口 : 视图
// @Description 获取主视图，有主机、网络等
// @Produce  json
// @Param type query string true "页面类型，主机页面type=host"
// @Success 200
// @Router /api/v1/dashboard/main [get]
func MainDashboard(c *gin.Context)  {
	dType := c.Query("type")
	if dType == "" {
		mid.ReturnValidateFail(c, "Parameter error")
		return
	}
	err,dashboard := db.GetDashboard(dType)
	if err != nil {
		mid.ReturnError(c, "Query dashboard failed", err)
		return
	}
	var dashboardDto m.Dashboard
	if dashboard.SearchEnable {
		err,search := db.GetSearch(dashboard.SearchId)
		if err == nil {
			search.Enable = true
			dashboardDto.Search = search
		}
	}
	if dashboard.ButtonEnable {
		err,button := db.GetButton(dashboard.ButtonGroup)
		if err == nil {
			dashboardDto.Buttons = button
		}
	}
	if dashboard.MessageEnable {
		var message m.MessageModel
		message.Enable = true
		messageUrl := strings.Replace(dashboard.MessageUrl, "{group}", fmt.Sprintf("%d", dashboard.MessageGroup), -1)
		message.Url = messageUrl
		dashboardDto.Message = message
	}
	if dashboard.PanelsEnable {
		var panels m.PanelsModel
		panels.Enable = true
		panels.Type = dashboard.PanelsType
		panels.Url = fmt.Sprintf("/dashboard/panels?group=%d", dashboard.PanelsGroup)
		if dashboard.PanelsParam != "" {
			panels.Url = panels.Url + `&` + dashboard.PanelsParam
			if dashboard.SearchEnable == false && len(dashboardDto.Buttons) > 0 {
				defaultBV := dashboard.PanelsParam[strings.Index(dashboard.PanelsParam, "{")+1:strings.Index(dashboard.PanelsParam, "}")]
				for _,v := range dashboardDto.Buttons {
					if v.Name == defaultBV {
						panels.Url = strings.Replace(panels.Url, fmt.Sprintf("{%s}", defaultBV) ,v.Options[0].OptionValue, -1)
					}
				}
			}
		}
		dashboardDto.Panels = panels
	}
	mid.ReturnData(c, dashboardDto)
}

// @Summary 页面通用接口 : 获取panels
// @Description 获取panels
// @Produce  json
// @Param group query int true "panels url 上自带该id"
// @Param endpoint query string true "需要在panels url上把{endpoint}替换"
// @Success 200
// @Router /api/v1/dashboard/panels [get]
func GetPanels(c *gin.Context)  {
	group := c.Query("group")
	endpoint := c.Query("endpoint")
	if group == "" {
		mid.ReturnValidateFail(c, "Param error")
		return
	}
	groupId,err := strconv.Atoi(group)
	if err != nil {
		mid.ReturnValidateFail(c, "Parameter \"group\" is not number")
		return
	}
	err,panels := db.GetPanels(groupId, endpoint)
	if err != nil {
		mid.ReturnError(c, "Get panels failed", err)
		return
	}
	var panelsDto []*m.PanelModel
	for _,panel := range panels {
		var panelDto m.PanelModel
		panelDto.Title = panel.Title
		if panel.ChartGroup < 0 {
			panelDto.Other = true
			panelsDto = append(panelsDto, &panelDto)
			continue
		}
		panelDto.Other = false
		err,charts := db.GetCharts(panel.ChartGroup, 0, 0)
		if err!=nil {
			continue
		}
		tagsDto := m.TagsModel{Enable:false, Option:[]*m.OptionModel{}}
		tagsValue := ""
		if panel.TagsEnable && endpoint!="" {
			var options []*m.OptionModel
			tagsDto.Enable = true
			tagsDto.Url = fmt.Sprintf(`%s?panel_id=%d&endpoint=%s&tag=`, panel.TagsUrl, panel.Id, endpoint)
			err,options = db.GetTags(endpoint, panel.TagsKey, charts[0].Metric)
			if err==nil {
				tagsDto.Option = options
				if len(options) > 0 {
					tagsValue = options[0].OptionText
				}
			}
		}
		panelDto.Tags = tagsDto
		var chartsDto []*m.ChartModel
		for _,chart := range charts {
			chartDto := m.ChartModel{Id:chart.Id, Col:chart.Col}
			chartDto.Url = `/dashboard/chart`
			chartDto.Endpoint = []string{endpoint}
			metricList := strings.Split(chart.Metric, "^")
			if panel.TagsEnable && tagsValue != ""{
				var newMetricList []string
				for _,m := range metricList {
					newMetric := m+`/`+panel.TagsKey+`=`+tagsValue
					newMetricList = append(newMetricList, newMetric)
				}
				chartDto.Metric = newMetricList
			}else{
				chartDto.Metric = metricList
			}
			chartsDto = append(chartsDto, &chartDto)
		}
		panelDto.Charts = chartsDto
		panelsDto = append(panelsDto, &panelDto)
	}
	mid.ReturnData(c, panelsDto)
}

// @Summary 页面通用接口 : 根据tag获取charts组
// @Description 根据tag获取charts组
// @Produce  json
// @Param panel_id query int true "url上自带该id"
// @Param endpoint query string true "url上自带该endpoint"
// @Param tag query string true "tag button里面的option_value"
// @Success 200
// @Router /api/v1/dashboard/tags [get]
func GetTags(c *gin.Context)  {
	panelIdStr := c.Query("panel_id")
	endpoint := c.Query("endpoint")
	tag := c.Query("tag")
	if tag == "" {
		mid.Return(c, mid.RespJson{Msg:"param error", Code:http.StatusBadRequest})
		return
	}
	panelId,err := strconv.Atoi(panelIdStr)
	if err != nil {
		mid.Return(c, mid.RespJson{Msg:"Parameter \"panel_id\" is not number", Code:http.StatusBadRequest})
		return
	}
	err,charts := db.GetCharts(0, 0, panelId)
	var chartsDto []*m.ChartModel
	for _,chart := range charts {
		chartDto := m.ChartModel{Id:chart.Id, Col:chart.Col}
		chartDto.Url = `/dashboard/chart`
		if endpoint!="" {
			chartDto.Endpoint = []string{endpoint}
		}
		metricList := strings.Split(chart.Metric, "^")
		var newMetricList []string
		for _,m := range metricList {
			newMetric := m+`/`+tag
			newMetricList = append(newMetricList, newMetric)
		}
		chartDto.Metric = newMetricList
		chartsDto = append(chartsDto, &chartDto)
	}
	mid.ReturnData(c, chartsDto)
}

// @Summary 页面通用接口 : 获取chart数据
// @Description 获取chart数据
// @Produce  json
// @Param id query int true "panel里的chart id"
// @Param endpoint query []string true "endpoint数组, ['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0']"
// @Param metric query []string true "metric数组, ['cpmid.busy']"
// @Param start query string true "开始时间"
// @Param end query string false "结束时间"
// @Param aggregate query string false "聚合类型 枚举 min max avg p95 none"
// @Success 200
// @Router /api/v1/dashboard/chart [get]
func GetChartOld(c *gin.Context)  {
	paramId,err := strconv.Atoi(c.Query("id"))
	if err != nil || paramId <= 0 {
		mid.ReturnValidateFail(c, "Chart id validation failed")
		return
	}
	err, charts := db.GetCharts(0, paramId, 0)
	if err != nil || len(charts) <= 0 {
		mid.ReturnError(c, "Get chart config failed", err)
		return
	}
	chart := *charts[0]
	var eOption m.EChartOption
	var query m.QueryMonitorData
	eOption.Id = paramId
	eOption.Title = chart.Title
	if chart.Endpoint == "" {
		query.Endpoint = c.QueryArray("endpoint[]")
	}else{
		query.Endpoint = strings.Split(chart.Endpoint, "^")
	}
	if len(query.Endpoint) <= 0 {
		mid.ReturnValidateFail(c ,"Parameter \"endpoint\" validation failed")
		return
	}
	query.Metric = c.QueryArray("metric[]")
	if chart.Metric != "" && len(query.Metric) > 0 {
		if !strings.Contains(query.Metric[0], "/") {
			query.Metric = strings.Split(chart.Metric, "^")
		}
	}
	paramTime := c.Query("time")
	paramStart := c.Query("start")
	paramEnd := c.Query("end")
	if paramTime != "" && paramStart == "" {
		paramStart = paramTime
	}
	start,err := strconv.ParseInt(paramStart, 10, 64)
	if err != nil {
		mid.ReturnError(c, "Parameter \"start\" validation failed", err)
		return
	}else{
		if start < 0 {
			start = time.Now().Unix() + start
		}
		query.Start = start
	}
	query.End = time.Now().Unix()
	if paramEnd != "" {
		end,err := strconv.ParseInt(paramEnd, 10, 64)
		if err == nil && end <= query.End {
			query.End = end
		}
	}
	err,query.PromQ = db.GetPromMetric(query.Endpoint, query.Metric[0])
	if err!=nil {
		mid.ReturnError(c, "Query promQL failed", err)
		return
	}
	query.Legend = chart.Legend
	mid.LogInfo(fmt.Sprintf("endpoint : %v  metric : %v  start:%d  end:%d  promql:%s", query.Endpoint, query.Metric, query.Start, query.End, query.PromQ))
	serials := ds.PrometheusData(query)
	agg := db.CheckAggregate(query.Start, query.End, query.Endpoint[0], 0, len(serials))
	for _, s := range serials {
		if strings.Contains(s.Name, "$metric") {
			s.Name = strings.Replace(s.Name, "$metric", query.Metric[0], -1)
		}
		eOption.Legend = append(eOption.Legend, s.Name)
		if agg > 1 {
			aggType := chart.AggType
			if c.Query("agg") != "" {
				aggType = c.Query("agg")
			}
			if aggType != "none" && aggType != "" {
				s.Data = db.Aggregate(s.Data, agg, aggType)
			}
		}
	}
	eOption.Xaxis = make(map[string]interface{})
	eOption.Yaxis = m.YaxisModel{Unit: chart.Unit}
	if len(serials) > 0 {
		eOption.Series = serials
	}else{
		eOption.Series = []*m.SerialModel{}
	}
	mid.ReturnData(c, eOption)
}

func GetChart(c *gin.Context)  {
	requestBody,err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		mid.ReturnValidateFail(c, "Read request body data fail")
		return
	}
	var paramConfig []m.ChartConfigObj
	err = json.Unmarshal(requestBody, &paramConfig)
	if err != nil || len(paramConfig) == 0 {
		mid.ReturnValidateFail(c, "Illegal parameter")
		return
	}
	var eOption m.EChartOption
	var query m.QueryMonitorData
	// validate config time
	if paramConfig[0].Time != "" && paramConfig[0].Start == "" {
		paramConfig[0].Start = paramConfig[0].Time
	}
	start,err := strconv.ParseInt(paramConfig[0].Start, 10, 64)
	if err != nil {
		mid.ReturnError(c, "Param start validation failed", err)
		return
	}else{
		if start < 0 {
			start = time.Now().Unix() + start
		}
		query.Start = start
	}
	query.End = time.Now().Unix()
	if paramConfig[0].End != "" {
		end,err := strconv.ParseInt(paramConfig[0].End, 10, 64)
		if err == nil && end <= query.End {
			query.End = end
		}
	}
	// custom or from mysql
	var querys []m.QueryMonitorData
	step := 0
	var firstEndpoint,firstMetric,unit string
	if paramConfig[0].Id > 0 {
		// one endpoint -> metrics
		err, charts := db.GetCharts(0, paramConfig[0].Id, 0)
		if err != nil || len(charts) <= 0 {
			mid.ReturnError(c, "Get chart config failed", err)
			return
		}
		chart := *charts[0]
		eOption.Id = chart.Id
		eOption.Title = chart.Title
		unit = chart.Unit
		if paramConfig[0].Endpoint == "" {
			mid.ReturnValidateFail(c, "Endpoint can not be empty")
			return
		}
		firstEndpoint = paramConfig[0].Endpoint
		if strings.Contains(paramConfig[0].Metric, "/") {
			chart.Metric = paramConfig[0].Metric
		}
		for i,v := range strings.Split(chart.Metric, "^") {
			if i == 0 {
				firstMetric = v
			}
			err,tmpPromQl := db.GetPromMetric([]string{paramConfig[0].Endpoint}, v)
			if err != nil {
				mid.LogError("Get prometheus metric failed", err)
				continue
			}
			querys = append(querys, m.QueryMonitorData{Start:query.Start, End:query.End, PromQ:tmpPromQl, Legend:chart.Legend, Metric:[]string{v}})
		}
	}else{
		step = 10
		var customLegend,tmpEndpointParam,tmpMetricParam string
		var diffEndpoint,diffMetric bool
		for i,v := range paramConfig {
			if v.PromQl == "" {
				_,tmpPromQL := db.GetPromMetric([]string{v.Endpoint}, v.Metric)
				if tmpPromQL == "" {
					continue
				}else{
					paramConfig[i].PromQl = tmpPromQL
				}
			}
			if i == 0 {
				tmpEndpointParam = v.Endpoint
				tmpMetricParam = v.Metric
			}else{
				if tmpEndpointParam != v.Endpoint {
					diffEndpoint = true
				}
				if tmpMetricParam != v.Metric {
					diffMetric = true
				}
			}
		}
		if diffEndpoint && !diffMetric {
			customLegend = "$custom_endpoint"
		}
		if !diffEndpoint && diffMetric {
			customLegend = "$custom_metric"
		}
		if diffEndpoint == diffMetric {
			customLegend = "$custom"
		}
		for _,v := range paramConfig {
			if v.PromQl == "" {
				continue
			}
			if strings.Contains(v.PromQl, "$address") {
				if v.Endpoint == "" {
					continue
				}
				endpointObj := m.EndpointTable{Guid:v.Endpoint}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Address == "" {
					continue
				}
				if endpointObj.AddressAgent != "" {
					v.PromQl = strings.Replace(v.PromQl, "$address", endpointObj.AddressAgent, -1)
				}else {
					v.PromQl = strings.Replace(v.PromQl, "$address", endpointObj.Address, -1)
				}
			}
			if strings.Contains(v.PromQl, "$") {
				re, _ := regexp.Compile("=\"[\\$]+[^\"]+\"")
				fetchTag := re.FindAll([]byte(v.PromQl), -1)
				for _,vv := range fetchTag {
					v.PromQl = strings.Replace(v.PromQl, string(vv), "=~\".*\"", -1)
				}
			}
			querys = append(querys, m.QueryMonitorData{Start:query.Start, End:query.End, PromQ:v.PromQl, Legend:customLegend, Metric:[]string{v.Metric}, Endpoint:[]string{v.Endpoint}})
		}
	}
	if len(querys) == 0 {
		mid.ReturnError(c, "Query list is empty", nil)
		return
	}
	var serials []*m.SerialModel
	for _,v := range querys {
		mid.LogInfo(fmt.Sprintf("query : endpoint : %v  metric : %v  start:%d  end:%d  promql:%s", v.Endpoint, v.Metric, v.Start, v.End, v.PromQ))
		tmpSerials := ds.PrometheusData(v)
		for _,vv := range tmpSerials {
			serials = append(serials, vv)
		}
	}
	// agg
	agg := db.CheckAggregate(query.Start, query.End, firstEndpoint, step, len(serials))
	var firstSerialTime float64
	for i, s := range serials {
		if strings.Contains(s.Name, "$metric") {
			s.Name = strings.Replace(s.Name, "$metric", firstMetric, -1)
		}
		eOption.Legend = append(eOption.Legend, s.Name)
		if agg > 1 {
			aggType := paramConfig[0].Aggregate
			if aggType != "none" && aggType != "" {
				s.Data = db.Aggregate(s.Data, agg, aggType)
			}
		}
		if i > 0 {
			if s.Data[0][0] != firstSerialTime {
				tmpSub := firstSerialTime - s.Data[0][0]
				for i,v := range s.Data {
					s.Data[i][0] = v[0] + tmpSub
				}
			}
		}else{
			firstSerialTime = s.Data[0][0]
		}
	}
	eOption.Xaxis = make(map[string]interface{})
	eOption.Yaxis = m.YaxisModel{Unit: unit}
	if len(serials) > 0 {
		eOption.Series = serials
	}else{
		eOption.Series = []*m.SerialModel{}
	}
	mid.ReturnData(c, eOption)
}

// @Summary 主页面接口 : 模糊搜索
// @Description 模糊搜索
// @Produce  json
// @Param search query string true "放弃search_col,直接把用户输入拼到url后面请求"
// @Param limit query string false "数量限制"
// @Success 200
// @Router /api/v1/dashboard/search [get]
func MainSearch(c *gin.Context)  {
	endpoint := c.Query("search")
	//limit := c.Query("limit")
	if endpoint == ""{
		mid.Return(c, mid.RespJson{Msg:"Param error", Code:http.StatusBadRequest})
		return
	}
	//tmpFlag := true
	if strings.Contains(endpoint, `:`) {
		endpoint = strings.Split(endpoint, `:`)[1]
		//tmpFlag = false
	}
	err,result := db.SearchHost(endpoint)
	if err != nil {
		mid.ReturnError(c, "Search hosts failed", err)
		return
	}
	sysResult := db.SearchRecursivePanel(endpoint)
	for _,v := range sysResult {
		result = append(result, v)
	}
	mid.ReturnData(c, result)
}

func GetPromMetric(c *gin.Context)  {
	metricType := c.Query("type")
	//if metricType == "" {
	//	mid.ReturnValidateFail(c, "Type is null")
	//	return
	//}
	err,data := db.GetPromMetricTable(metricType)
	if err != nil {
		mid.ReturnError(c, "Get prometheus metric error", err)
		return
	}
	mid.ReturnData(c, data)
}

func UpdatePromMetric(c *gin.Context)  {
	var param []m.PromMetricTable
	if err := c.ShouldBindJSON(&param);err == nil {
		if len(param) == 0 {
			mid.ReturnValidateFail(c, "List is null")
			return
		}
		err := db.UpdatePromMetric(param)
		if err != nil {
			mid.ReturnError(c, "Update prometheus metric failed", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func GetEndpointMetric(c *gin.Context)  {
	id,_ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		mid.ReturnValidateFail(c, "Param id validate failed")
		return
	}
	err,data := db.GetEndpointMetric(id)
	if err != nil {
		mid.ReturnError(c, "Get endpoint metric failed", err)
		return
	}
	mid.ReturnData(c, data)
}

func GetChartsByEndpoint(c *gin.Context)  {
	// Validate ip and metric
	ip := c.Query("ip")
	metric := c.Query("metric")
	if ip == "" || metric == "" {
		mid.ReturnValidateFail(c, "Ip or metric can not be empty")
		return
	}
	endpointObj := m.EndpointTable{Ip:ip, ExportType:"host"}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		mid.ReturnValidateFail(c, fmt.Sprintf("Can not find the host register message with ip %s", ip))
		return
	}
	err,promQL := db.GetPromMetric([]string{endpointObj.Guid}, metric)
	if err != nil || promQL == "" {
		mid.ReturnError(c, fmt.Sprintf("Get promQL failed with endpoint : %s metric : %s", endpointObj.Guid, metric), err)
		return
	}
	var eOption m.EChartOption
	var query m.QueryMonitorData
	query.Endpoint = []string{endpointObj.Guid}
	query.Metric = []string{metric}
	query.PromQ = promQL
	query.Legend = "$metric"
	// Validate time start end
	paramTime := c.Query("time")
	paramStart := c.Query("start")
	paramEnd := c.Query("end")
	if paramTime != "" && paramStart == "" {
		paramStart = paramTime
	}
	start,err := strconv.ParseInt(paramStart, 10, 64)
	if err != nil {
		mid.ReturnError(c, "Param start validation failed", err)
		return
	}else{
		if start < 0 {
			start = time.Now().Unix() + start
		}
		query.Start = start
	}
	query.End = time.Now().Unix()
	if paramEnd != "" {
		end,err := strconv.ParseInt(paramEnd, 10, 64)
		if err == nil && end <= query.End {
			query.End = end
		}
	}
	// Query data
	mid.LogInfo(fmt.Sprintf("endpoint : %v  metric : %v  start:%d  end:%d  promql:%s", query.Endpoint, query.Metric, query.Start, query.End, query.PromQ))
	serials := ds.PrometheusData(query)
	for _, s := range serials {
		if strings.Contains(s.Name, "$metric") {
			s.Name = strings.Replace(s.Name, "$metric", metric, -1)
		}
		eOption.Legend = append(eOption.Legend, s.Name)
	}
	eOption.Xaxis = make(map[string]interface{})
	var unit string
	if strings.Contains(metric, "percent") {
		unit = "%"
	}
	eOption.Yaxis = m.YaxisModel{Unit: unit}
	if len(serials) > 0 {
		eOption.Series = serials
	}else{
		eOption.Series = []*m.SerialModel{}
	}
	mid.ReturnData(c, eOption)
}

func GetMainPage(c *gin.Context)  {
	err,result := db.GetMainCustomDashboard()
	if err != nil {
		mid.LogError("Get main page failed ", err)
	}
	mid.ReturnData(c, result)
}

func SetMainPage(c *gin.Context)  {
	id,err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnValidateFail(c, "Id validation failed")
		return
	}
	err = db.SetMainCustomDashboard(id)
	if err != nil {
		mid.ReturnError(c, "Set main dashboard failed", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}

func GetEndpointsByIp(c *gin.Context)  {
	ipList := c.QueryArray("ip")
	if len(ipList) <= 0 {
		mid.ReturnValidateFail(c, "Parameter validation failed")
		return
	}
	err,endpoints := db.GetEndpointsByIp(ipList, "host")
	if err != nil {
		mid.ReturnError(c, fmt.Sprintf("Get endpoints by ipList : %s fail ", ipList), err)
		return
	}
	mid.ReturnData(c, endpoints)
}