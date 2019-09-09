package dashboard

import (
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	u "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware/util"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	ds "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/datasource"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/cron"
)

// @Summary 页面通用接口 : 视图
// @Description 获取主视图，有主机、网络等
// @Produce  json
// @Param type query string true "页面类型，主机页面type=host"
// @Success 200 {string} json "{'search':{'id': 0,'enable': false,'name': '','search_url': '','search_col': '','refresh_panels': false,'refresh_message': false},'buttons':[{'id': 1,'group_id': 1,'name': 'time','b_type': 'select','b_text': '时间段','refresh_panels': false,'refresh_charts': true,'option_group': 1,'option':[{'option_value': '-3600', 'option_text': '1小时'},{'option_value': '-10800', 'option_text': '3小时'},{'option_value': '-21600', 'option_text': '6小时'},{'option_value': '-43200', 'option_text': '12小时'}]}],'message':{'enable': true,'url': '/dashboard/message?group=1&endpoint={endpoint}'},'panels':{'enable': true,'type': 'tabs','url': '/dashboard/panels?group=1'}}"
// @Router /api/v1/dashboard/main [get]
func MainDashboard(c *gin.Context)  {
	dType := c.Query("type")
	if dType == "" {
		u.Return(c, u.RespJson{Msg:"param error", Code:http.StatusBadRequest})
		return
	}
	err,dashboard := db.GetDashboard(dType)
	if err != nil {
		u.ReturnError(c, "query dashboard fail", err)
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
	u.ReturnData(c, dashboardDto)
}

// @Summary 页面通用接口 : 获取panels
// @Description 获取panels
// @Produce  json
// @Param group query int true "panels url 上自带该id"
// @Param endpoint query string true "需要在panels url上把{endpoint}替换"
// @Success 200 {string} json "[{'title': 'panel_title','tags':{'enable': false,'url': '','option':[]},'charts':[{'id': 1,'col': 6,'endpoint':['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0'],'metric':['cpu.busy'],'url': '/dashboard/chart'}]},{'title': 'disk','tags':{'enable': true,'url': '/dashboard/tags?panel_id=1&endpoint=88B525B4-43E8-4A7A-8E11-0E664B5CB8D0&tag=','option':[{'option_value': 'device=vda','option_text': 'vda'},{'option_value': 'device=vdb','option_text': 'vdb'}]},'charts':[{'id': 2,'col': 6,'endpoint':['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0'],'metric':['disk.io.util/device=vda'],'url': '/dashboard/chart'},{'id': 3,'col': 6,'endpoint':['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0'],'metric':['disk.io.await/device=vda'],'url': '/dashboard/chart'}]}]"
// @Router /api/v1/dashboard/panels [get]
func GetPanels(c *gin.Context)  {
	group := c.Query("group")
	endpoint := c.Query("endpoint")
	if group == "" {
		u.Return(c, u.RespJson{Msg:"param error", Code:http.StatusBadRequest})
		return
	}
	groupId,err := strconv.Atoi(group)
	if err != nil {
		u.Return(c, u.RespJson{Msg:"param group is not number error", Code:http.StatusBadRequest})
		return
	}
	err,panels := db.GetPanels(groupId, endpoint)
	if err != nil {
		u.ReturnError(c, "get panels fail", err)
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
		var chartEndpoints []string
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
			if panel.ExIsPhy {
				chartDto.Endpoint = chartEndpoints
			}else{
				chartDto.Endpoint = []string{endpoint}
			}
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
	u.ReturnData(c, panelsDto)
}

// @Summary 页面通用接口 : 根据tag获取charts组
// @Description 根据tag获取charts组
// @Produce  json
// @Param panel_id query int true "url上自带该id"
// @Param endpoint query string true "url上自带该endpoint"
// @Param tag query string true "tag button里面的option_value"
// @Success 200 {string} json "[{'id': 2,'col': 6,'endpoint':['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0'],'metric':['disk.io.util/device=vdb'],'url': '/dashboard/chart'},{'id': 3,'col': 6,'endpoint':['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0'],'metric':['disk.io.await/device=vdb'],'url': '/dashboard/chart'}]"
// @Router /api/v1/dashboard/tags [get]
func GetTags(c *gin.Context)  {
	panelIdStr := c.Query("panel_id")
	endpoint := c.Query("endpoint")
	tag := c.Query("tag")
	if tag == "" {
		u.Return(c, u.RespJson{Msg:"param error", Code:http.StatusBadRequest})
		return
	}
	panelId,err := strconv.Atoi(panelIdStr)
	if err != nil {
		u.Return(c, u.RespJson{Msg:"param panel_id is not number error", Code:http.StatusBadRequest})
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
	u.ReturnData(c, chartsDto)
}

// @Summary 页面通用接口 : 获取chart数据
// @Description 获取chart数据
// @Produce  json
// @Param id query int true "panel里的chart id"
// @Param endpoint query []string true "endpoint数组, ['88B525B4-43E8-4A7A-8E11-0E664B5CB8D0']"
// @Param metric query []string true "metric数组, ['cpu.busy']"
// @Param start query string true "开始时间"
// @Param end query string false "结束时间"
// @Param aggregate query string false "聚合类型 枚举 min max avg p95 none"
// @Success 200 {string} json "{'title':'chart1','legend':['a','b'],'xaxis':{},'yaxis':{'unit':'%'},'series':[{'type':'line','name':'a','data':[[1550221207000,100],[1550221217000,200],[1550221227000,150],[1550221237000,100],[1550221247000,120],[1550221257000,210],[1550221267000,130],[1550221277000,180]]},{'type':'line','name':'b','data':[[1550221207000,110],[1550221217000,210],[1550221227000,130],[1550221237000,120],[1550221247000,100],[1550221257000,200],[1550221267000,170],[1550221277000,120]]}]}"
// @Router /api/v1/dashboard/chart [post]
func GetChart(c *gin.Context)  {
	var param m.GetChartDto
	if err := c.ShouldBindJSON(&param); err==nil {
		if !mid.ValidatePost(c, param) {return}
		//isDefault := true
		var chart m.ChartTable
		if param.Id <= 0 {
			u.ReturnError(c, "chart id validate error", nil)
			return
		}
		//var echartOption m.EChartOption
		err, charts := db.GetCharts(0, param.Id, 0)
		if err != nil {
			u.ReturnError(c, "get chart config error", err)
			return
		}
		if len(charts) > 0 {
			chart = *charts[0]
		}
		var eOption m.EChartOption
		var endpoints []string
		var metrics []string
		var legends []string
		//var endpointIp map[string]string
		//var endpointShow bool
		eOption.Title = chart.Title
		eOption.Id = param.Id
		// 参数校验
		if chart.Endpoint == "" {
			endpoints = param.Endpoint
		}else{
			endpoints = strings.Split(chart.Endpoint, "^")
		}
		if chart.Metric == "" || strings.Contains(param.Metric[0], "/") {
			metrics = param.Metric
		}else{
			metrics = strings.Split(chart.Metric, "^")
		}
		var query m.QueryMonitorData
		if param.Time!="" && param.Start=="" {
			param.Start = param.Time
		}
		start,err := strconv.ParseInt(param.Start, 10, 64)
		if err==nil {
			if start<0 {
				start = time.Now().Unix() + start
			}
			query.Start = start
		}
		if param.End!=""{
			end,err := strconv.ParseInt(param.End, 10, 64)
			if err==nil {
				if end > time.Now().Unix() {
					end = time.Now().Unix()
				}
				query.End = end
			}
		}else{
			query.End = time.Now().Unix()
		}
		query.Endpoint = endpoints
		query.Metric = metrics
		err,query.PromQ = db.GetPromMetric(endpoints, metrics)
		if err!=nil {
			u.ReturnError(c, "query promQL fail", err)
			return
		}
		mid.LogInfo(fmt.Sprintf("endpoint : %v  metric : %v  start:%d  end:%d  promql:%s", query.Endpoint, query.Metric, query.Start, query.End, query.PromQ))
		// 取数据库中的要计算rate的指标，把相应指标的rate置为true
		query.ComputeRate = chart.Rate
		// 取数据
		serials := ds.PrometheusData(query)
		//tmpTitle := ""
		for _,s := range serials{
			s.Name = fmt.Sprintf("%s:%s", s.Name, metrics[0])
			legends = append(legends, s.Name)
		}
		//eOption.Title = tmpTitle
		if len(legends) > 0 {
			eOption.Legend = legends
		}else{
			eOption.Legend = []string{}
		}
		eOption.Xaxis = make(map[string]interface{})
		eOption.Yaxis = m.YaxisModel{Unit: chart.Unit}
		if len(serials) > 0 {
			eOption.Series = serials
		}else{
			eOption.Series = []*m.SerialModel{}
		}
		u.ReturnData(c, eOption)
	}else{
		u.Return(c, u.RespJson{Msg:"param validate fail", Code:http.StatusBadRequest})
	}
}

// @Summary 主页面接口 : 模糊搜索
// @Description 模糊搜索
// @Produce  json
// @Param search query string true "放弃search_col,直接把用户输入拼到url后面请求"
// @Param limit query string false "数量限制"
// @Success 200 {string} json "[{'option_value': 'E7196678-F696-4AE3-8D5A-83CFC80B0801','option_text': 'cnsz92vl00311:100.69.12.10'},{'option_value': '5DE6A2DE-B506-4385-82EB-B2533D751E7F','option_text': 'cnsz92vl00312.cmftdc.cn:100.69.12.11'}]"  说明: 放弃search_col，统一以option的格式来返回
// @Router /api/v1/dashboard/search [get]
func MainSearch(c *gin.Context)  {
	endpoint := c.Query("search")
	//limit := c.Query("limit")
	if endpoint == ""{
		u.Return(c, u.RespJson{Msg:"param error", Code:http.StatusBadRequest})
		return
	}
	if strings.Contains(endpoint, `:`) {
		endpoint = strings.Split(endpoint, `:`)[1]
	}
	err,result := db.SearchHost(endpoint)
	if err != nil {
		u.ReturnError(c, "search host fail", err)
		return
	}
	u.ReturnData(c, result)
}

func RegisterEndpoint(c *gin.Context)  {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		u.ReturnValidateFail(c, "endpoint is null")
		return
	}
	err := cron.SyncEndpointMetric(endpoint)
	if err != nil {
		u.ReturnError(c, "register endpoint fail", err)
		return
	}
	u.ReturnSuccess(c, fmt.Sprintf("register endpoint %s success", endpoint))
}