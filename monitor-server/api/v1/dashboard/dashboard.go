package dashboard

import (
	"encoding/json"
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// @Summary 页面通用接口 : 视图
// @Description 获取主视图，有主机、网络等
// @Produce  json
// @Param type query string true "页面类型，主机页面type=host"
// @Success 200
// @Router /api/v1/dashboard/main [get]
func MainDashboard(c *gin.Context) {
	dType := c.Query("type")
	if dType == "" {
		mid.ReturnParamEmptyError(c, "type")
		return
	}
	err, dashboard := db.GetDashboard(dType)
	if err != nil {
		mid.ReturnQueryTableError(c, "dashboard", err)
		return
	}
	var dashboardDto m.Dashboard
	if dashboard.SearchEnable {
		err, search := db.GetSearch(dashboard.SearchId)
		if err == nil {
			search.Enable = true
			dashboardDto.Search = search
		}
	}
	if dashboard.ButtonEnable {
		err, button := db.GetButton(dashboard.ButtonGroup)
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
				defaultBV := dashboard.PanelsParam[strings.Index(dashboard.PanelsParam, "{")+1 : strings.Index(dashboard.PanelsParam, "}")]
				for _, v := range dashboardDto.Buttons {
					if v.Name == defaultBV {
						panels.Url = strings.Replace(panels.Url, fmt.Sprintf("{%s}", defaultBV), v.Options[0].OptionValue, -1)
					}
				}
			}
		}
		dashboardDto.Panels = panels
	}
	mid.ReturnSuccessData(c, dashboardDto)
}

// @Summary 页面通用接口 : 获取panels
// @Description 获取panels
// @Produce  json
// @Param group query int true "panels url 上自带该id"
// @Param endpoint query string true "需要在panels url上把{endpoint}替换"
// @Success 200
// @Router /api/v1/dashboard/panels [get]
func GetPanels(c *gin.Context) {
	group := c.Query("group")
	endpoint := c.Query("endpoint")
	if group == "" {
		mid.ReturnParamEmptyError(c, "group")
		return
	}
	groupId, err := strconv.Atoi(group)
	if err != nil {
		mid.ReturnParamTypeError(c, "group", "int")
		return
	}
	if groupId == 0 {
		mid.ReturnSuccessData(c, []string{})
		return
	}
	err, panels := db.GetPanels(groupId, endpoint)
	if err != nil {
		mid.ReturnQueryTableError(c, "panel", err)
		return
	}
	var panelsDto []*m.PanelModel
	for _, panel := range panels {
		//if panel.AutoDisplay > 0 && !endpointBusinessShow {
		//	continue
		//}
		if panel.AutoDisplay == 1 {
			continue
		}
		var panelDto m.PanelModel
		panelDto.Title = panel.Title
		if panel.ChartGroup < 0 {
			panelDto.Other = true
			panelsDto = append(panelsDto, &panelDto)
			continue
		}
		panelDto.Other = false
		err, charts := db.GetCharts(panel.ChartGroup, 0, 0)
		if err != nil {
			continue
		}
		tagsDto := m.TagsModel{Enable: false, Option: []*m.OptionModel{}}
		tagsValue := ""
		if panel.TagsEnable && endpoint != "" {
			var options []*m.OptionModel
			tagsDto.Enable = true
			tagsDto.Url = fmt.Sprintf(`%s?panel_id=%d&endpoint=%s&tag=`, panel.TagsUrl, panel.Id, endpoint)
			err, options = db.GetTags(endpoint, panel.TagsKey, charts[0].Metric)
			if err == nil {
				tagsDto.Option = options
				if len(options) > 0 {
					tagsValue = options[0].OptionText
				}
			}
		}
		panelDto.Tags = tagsDto
		var chartsDto []*m.ChartModel
		for _, chart := range charts {
			chartDto := m.ChartModel{Id: chart.Id, Col: chart.Col}
			chartDto.Url = `/dashboard/chart`
			chartDto.Endpoint = []string{endpoint}
			metricList := strings.Split(chart.Metric, "^")
			if panel.TagsEnable && tagsValue != "" {
				var newMetricList []string
				for _, m := range metricList {
					newMetric := m + `/` + panel.TagsKey + `=` + tagsValue
					newMetricList = append(newMetricList, newMetric)
				}
				chartDto.Metric = newMetricList
			} else {
				chartDto.Metric = metricList
			}
			chartsDto = append(chartsDto, &chartDto)
		}
		panelDto.Charts = chartsDto
		panelsDto = append(panelsDto, &panelDto)
	}
	businessPanel, matchErr := db.MatchServicePanel(endpoint)
	//err,businessPanel := fetchBusinessPanel(endpoint)
	if matchErr != nil {
		log.Error(nil, log.LOGGER_APP, "Fetch business panel fail", zap.Error(matchErr))
	}
	if len(businessPanel.Charts) > 0 {
		panelsDto = append(panelsDto, &businessPanel)
	}
	mid.ReturnSuccessData(c, panelsDto)
}

func fetchBusinessPanel(endpoint string) (err error, result m.PanelModel) {
	result.Tags = m.TagsModel{Enable: false, Option: []*m.OptionModel{}}
	var businessList m.BusinessUpdateDto
	realEndpoint := endpoint
	endpointObj := m.EndpointTable{Guid: endpoint}
	db.GetEndpoint(&endpointObj)
	if endpointObj.ExportType == "host" {
		err, businessList = db.GetBusinessListNew(endpointObj.Id, "")
	} else {
		err, businessList = db.GetBusinessListNew(0, endpoint)
		realEndpoint = db.GetBusinessRealEndpoint(endpoint)
	}
	if err != nil || len(businessList.PathList) == 0 {
		return err, result
	}
	chartTable, panelTable := db.GetBusinessPanelChart()
	if len(panelTable) == 0 || len(chartTable) == 0 {
		return err, result
	}
	result.Title = panelTable[0].Title
	//var promMetricKeys []string
	for _, path := range businessList.PathList {
		for _, rule := range path.Rules {
			for _, metricConfig := range rule.MetricConfig {
				//promMetricKeys = append(promMetricKeys, metricConfig.Metric)
				tmpChartObj := m.ChartModel{Id: chartTable[0].Id, Endpoint: []string{realEndpoint}, Url: chartTable[0].Url}
				tmpChartObj.Title = metricConfig.Title
				tmpChartObj.Metric = []string{fmt.Sprintf("%s/path=%s,key=%s", chartTable[0].Metric, path.Path, metricConfig.Metric)}
				result.Charts = append(result.Charts, &tmpChartObj)
			}
		}
		for _, custom := range path.CustomMetrics {
			tmpChartObj := m.ChartModel{Id: chartTable[0].Id, Endpoint: []string{realEndpoint}, Url: chartTable[0].Url}
			tmpChartObj.Title = custom.Metric
			tmpChartObj.Metric = []string{fmt.Sprintf("%s/path=%s,key=%s", chartTable[0].Metric, path.Path, custom.Metric)}
			result.Charts = append(result.Charts, &tmpChartObj)
		}
	}
	//_,extendMetric := db.GetBusinessPromMetric(promMetricKeys)
	//for _,v := range extendMetric {
	//	tmpChartObj := m.ChartModel{Id: chartTable[0].Id, Endpoint: []string{realEndpoint}, Url: chartTable[0].Url}
	//	tmpChartObj.Metric = []string{v.Metric}
	//	tmpChartObj.Title = v.Metric
	//	result.Charts = append(result.Charts, &tmpChartObj)
	//}
	return err, result
}

func UpdateChartsTitle(c *gin.Context) {
	var param m.UpdateChartTitleParam
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.ChartId > 0 {
			err = db.UpdateChartTitle(param)
		} else {
			err = db.UpdateServiceMetricTitle(param)
		}
		if err != nil {
			mid.ReturnUpdateTableError(c, "chart", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

// @Summary 页面通用接口 : 根据tag获取charts组
// @Description 根据tag获取charts组
// @Produce  json
// @Param panel_id query int true "url上自带该id"
// @Param endpoint query string true "url上自带该endpoint"
// @Param tag query string true "tag button里面的option_value"
// @Success 200
// @Router /api/v1/dashboard/tags [get]
func GetTags(c *gin.Context) {
	panelIdStr := c.Query("panel_id")
	endpoint := c.Query("endpoint")
	tag := c.Query("tag")
	if tag == "" {
		mid.ReturnParamEmptyError(c, "tag")
		return
	}
	panelId, err := strconv.Atoi(panelIdStr)
	if err != nil {
		mid.ReturnParamTypeError(c, "panel_id", "int")
		return
	}
	err, charts := db.GetCharts(0, 0, panelId)
	var chartsDto []*m.ChartModel
	for _, chart := range charts {
		chartDto := m.ChartModel{Id: chart.Id, Col: chart.Col}
		chartDto.Url = `/dashboard/chart`
		if endpoint != "" {
			chartDto.Endpoint = []string{endpoint}
		}
		metricList := strings.Split(chart.Metric, "^")
		var newMetricList []string
		for _, m := range metricList {
			newMetric := m + `/` + tag
			newMetricList = append(newMetricList, newMetric)
		}
		chartDto.Metric = newMetricList
		chartsDto = append(chartsDto, &chartDto)
	}
	mid.ReturnSuccessData(c, chartsDto)
}

func GetPieChart(c *gin.Context) {
	requestBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		mid.ReturnBodyError(c, err)
		return
	}
	var paramConfig []*m.PieChartConfigObj
	err = json.Unmarshal(requestBody, &paramConfig)
	if err != nil {
		mid.ReturnRequestJsonError(c, err)
		return
	}
	var queryResultList []*m.QueryMonitorData
	resultPieData := m.EChartPie{}
	for _, paramObj := range paramConfig {
		tmpQueryResult, tmpErr := getPieData(paramObj)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		queryResultList = append(queryResultList, tmpQueryResult...)
	}
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	for _, v := range queryResultList {
		resultPieData.Legend = append(resultPieData.Legend, v.PieData.Legend...)
		resultPieData.Data = append(resultPieData.Data, v.PieData.Data...)
	}
	mid.ReturnSuccessData(c, resultPieData)
}

func getPieData(paramConfig *m.PieChartConfigObj) (result []*m.QueryMonitorData, err error) {
	if paramConfig.MonitorType != "" {
		paramConfig.AppObjectEndpointType = paramConfig.MonitorType
	}
	if paramConfig.PieType != "" {
		paramConfig.PieMetricType = paramConfig.PieType
	}
	result = []*m.QueryMonitorData{}
	var tagNameList []string
	if paramConfig.CustomChartGuid != "" {
		customChartObj, getChartErr := db.GetCustomChartById(paramConfig.CustomChartGuid)
		if getChartErr != nil {
			err = getChartErr
			return
		}
		if customChartObj != nil {
			paramConfig.PieMetricType = customChartObj.PieType
		}
		chartSeries, getSeriesErr := db.GetCustomChartSeries(paramConfig.CustomChartGuid)
		if getSeriesErr != nil {
			err = getSeriesErr
			return
		}
		if len(chartSeries) == 0 {
			log.Warn(nil, log.LOGGER_APP, "Can not find chart series", zap.String("guid", paramConfig.CustomChartGuid))
			return
		}
		seriesObj := chartSeries[0]
		paramConfig.Metric = seriesObj.Metric
		paramConfig.AppObject = seriesObj.ServiceGroup
		paramConfig.Endpoint = seriesObj.Endpoint
		paramConfig.AppObjectEndpointType = seriesObj.MonitorType
		paramConfig.Tags = seriesObj.Tags
		paramConfig.PieDisplayTag = seriesObj.PieDisplayTag
	}
	log.Debug(nil, log.LOGGER_APP, "pie paramConfig", log.JsonObj("paramConfig", paramConfig))
	for _, v := range paramConfig.Tags {
		tagNameList = append(tagNameList, v.TagName)
	}
	if paramConfig.Metric == "" {
		err = fmt.Errorf("metric can not empty")
		return
	}
	if paramConfig.PieAggType == "" {
		paramConfig.PieAggType = "sum"
	}
	var endpointList []*m.EndpointNewTable
	if paramConfig.AppObject != "" && paramConfig.AppObjectEndpointType != "" {
		endpointList, err = db.GetRecursiveEndpointByTypeNew(paramConfig.AppObject, paramConfig.AppObjectEndpointType)
		if err != nil {
			err = fmt.Errorf("get service group endpoint fail,%s ", err.Error())
			return
		}

	} else if paramConfig.Endpoint != "" {
		endpointObj, tmpErr := db.GetEndpointNew(&m.EndpointNewTable{Guid: paramConfig.Endpoint})
		if tmpErr != nil {
			err = tmpErr
			return
		}
		paramConfig.AppObjectEndpointType = endpointObj.MonitorType
		endpointList = append(endpointList, &endpointObj)
	}
	if len(endpointList) == 0 {
		err = fmt.Errorf("endpoint can not empty ")
		return
	}
	var queryStart, queryEnd int64
	if paramConfig.PieAggType == "new" {
		queryEnd = time.Now().Unix()
		queryStart = queryEnd - 300
	} else {
		if paramConfig.Start > 0 && paramConfig.End > 0 {
			queryEnd = paramConfig.End
			queryStart = paramConfig.Start
		} else if paramConfig.TimeSecond < 0 {
			queryEnd = time.Now().Unix()
			queryStart = queryEnd + paramConfig.TimeSecond
		} else {
			queryEnd = time.Now().Unix()
			queryStart = queryEnd - 3600
		}
	}
	// fetch promQL
	if paramConfig.PromQl == "" {
		tmpPromQL, _ := db.GetPromQLByMetric(paramConfig.Metric, paramConfig.AppObjectEndpointType, paramConfig.AppObject)
		if tmpPromQL == "" {
			err = fmt.Errorf("metric:%s can not get any prom_ql ", paramConfig.Metric)
			return
		} else {
			paramConfig.PromQl = tmpPromQL
		}
	}
	promMap := make(map[string]bool)
	for _, endpoint := range endpointList {
		tmpPromQL := db.ReplacePromQlKeyword(paramConfig.PromQl, paramConfig.Metric, endpoint, paramConfig.Tags)
		if _, b := promMap[tmpPromQL]; b {
			continue
		}
		promMap[tmpPromQL] = true
		result = append(result, &m.QueryMonitorData{ChartType: "pie", Start: queryStart, End: queryEnd, PromQ: tmpPromQL, Legend: "", Metric: []string{paramConfig.Metric}, Endpoint: []string{endpoint.Guid}, Step: endpoint.Step, Cluster: endpoint.Cluster, PieMetricType: paramConfig.PieMetricType, PieAggType: paramConfig.PieAggType, Tags: tagNameList, PieDisplayTag: paramConfig.PieDisplayTag})
	}
	if paramConfig.PieMetricType == "value" {
		if len(result) == 0 {
			return
		}
		result[0].ChartType = "line"
		serialList := ds.PrometheusData(result[0])
		if len(serialList) > 0 {
			valueMap := make(map[float64]int)
			for _, v := range serialList[0].Data {
				if count, b := valueMap[v[1]]; b {
					valueMap[v[1]] = count + 1
				} else {
					valueMap[v[1]] = 1
				}
			}
			pieData := m.EChartPie{}
			for k, v := range valueMap {
				tmpName := fmt.Sprintf("%.3f", k)
				pieData.Legend = append(pieData.Legend, tmpName)
				pieData.Data = append(pieData.Data, &m.EChartPieObj{Name: tmpName, Value: float64(v)})
			}
			result[0].PieData = pieData
		}
	} else {
		for _, queryObj := range result {
			log.Info(nil, log.LOGGER_APP, "queryObj", log.JsonObj("data", queryObj))
			ds.PrometheusData(queryObj)
		}
	}
	return
}

// @Summary 主页面接口 : 模糊搜索
// @Description 模糊搜索
// @Produce  json
// @Param search query string true "放弃search_col,直接把用户输入拼到url后面请求"
// @Param limit query string false "数量限制"
// @Success 200
// @Router /api/v1/dashboard/search [get]
func MainSearch(c *gin.Context) {
	endpoint := c.Query("search")
	//limit := c.Query("limit")
	if endpoint == "" {
		mid.ReturnParamEmptyError(c, "search")
		return
	}
	tmpFlag := false
	if strings.Contains(endpoint, `:`) {
		endpoint = strings.Split(endpoint, `:`)[1]
		tmpFlag = true
	}
	err, result := db.SearchHost(endpoint)
	if err != nil {
		mid.ReturnQueryTableError(c, "endpoint", err)
		return
	}
	for _, v := range result {
		v.OptionTypeName = v.OptionType
	}
	sysResult := db.SearchRecursivePanel(endpoint)
	for _, v := range sysResult {
		result = append(result, v)
	}
	if tmpFlag {
		var tmpResult []*m.OptionModel
		for _, v := range result {
			if v.OptionText == c.Query("search") {
				tmpResult = append(tmpResult, v)
				break
			}
		}
		if len(tmpResult) > 0 {
			result = tmpResult
		}
	}
	var sortOptionList m.OptionModelSortList
	sortOptionList = append(sortOptionList, result...)
	sort.Sort(sortOptionList)
	mid.ReturnSuccessData(c, sortOptionList)
}

func GetPromMetric(c *gin.Context) {
	metricType := c.Query("type")
	err, data := db.GetPromMetricTable(metricType)
	if err != nil {
		mid.ReturnQueryTableError(c, "prom_metric", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdatePanelChartMetric(c *gin.Context) {
	var param []m.PromMetricUpdateParam
	if err := c.ShouldBindJSON(&param); err == nil {
		if len(param) == 0 {
			mid.ReturnParamEmptyError(c, "")
			return
		}
		err := db.UpdatePanelChartMetric(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "prom_metric", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdatePromMetric(c *gin.Context) {
	var param []*m.PromMetricTable
	if err := c.ShouldBindJSON(&param); err == nil {
		if len(param) == 0 {
			mid.ReturnParamEmptyError(c, "")
			return
		}
		err := db.UpdatePromMetric(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "prom_metric", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetEndpointMetric(c *gin.Context) {
	var param m.GetEndpointMetricParam
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	var data []*m.OptionModel
	if param.ServiceGroup != "" {
		err, data = db.GetServiceGroupPromMetric(param.ServiceGroup, param.Workspace, param.MonitorType)
	} else {
		err, data = db.GetEndpointMetric(param.Guid, param.MonitorType)
	}
	if err != nil {
		mid.ReturnHandleError(c, "Get endpoint metric failed", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func GetChartsByEndpoint(c *gin.Context) {
	// Validate ip and metric
	ip := c.Query("ip")
	metric := c.Query("metric")
	if ip == "" || metric == "" {
		mid.ReturnParamEmptyError(c, "ip or metric")
		return
	}
	endpointObj := m.EndpointTable{Ip: ip, ExportType: "host"}
	db.GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		mid.ReturnFetchDataError(c, "endpoint", "ip", ip)
		return
	}
	err, promQL := db.GetPromMetric([]string{endpointObj.Guid}, metric)
	if err != nil || promQL == "" {
		mid.ReturnFetchDataError(c, "prom_ql", "endpoint,metric", endpointObj.Guid+","+metric)
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
	start, err := strconv.ParseInt(paramStart, 10, 64)
	if err != nil {
		mid.ReturnParamTypeError(c, "start", "intString")
		return
	} else {
		if start < 0 {
			start = time.Now().Unix() + start
		}
		query.Start = start
	}
	query.End = time.Now().Unix()
	if paramEnd != "" {
		end, err := strconv.ParseInt(paramEnd, 10, 64)
		if err == nil && end <= query.End {
			query.End = end
		}
	}
	// Query data
	log.Debug(nil, log.LOGGER_APP, "Query param", zap.Strings("endpoint", query.Endpoint), zap.Strings("metric", query.Metric), zap.Int64("start", query.Start), zap.Int64("end", query.End), zap.String("promQl", query.PromQ))
	serials := ds.PrometheusData(&query)
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
	} else {
		eOption.Series = []*m.SerialModel{}
	}
	mid.ReturnSuccessData(c, eOption)
}

func GetMainPage(c *gin.Context) {
	var roleList []string
	var user string
	token := mid.GetCoreToken(c)
	roleList = token.Roles
	user = token.User
	if user == "" || len(roleList) == 0 {
		user = mid.GetOperateUser(c)
		_, userRoleList := db.GetUserRole(user)
		for _, v := range userRoleList {
			roleList = append(roleList, v.Name)
		}
	}
	err, result := db.GetMainCustomDashboard(roleList)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get main page failed", zap.Error(err))
	}
	mid.ReturnSuccessData(c, result)
}

func ListMainPageRole(c *gin.Context) {
	var roleList []string
	var user string
	token := mid.GetCoreToken(c)
	roleList = token.Roles
	user = token.User
	if user == "" || len(roleList) == 0 {
		user = mid.GetOperateUser(c)
		_, userRoleList := db.GetUserRole(user)
		for _, v := range userRoleList {
			roleList = append(roleList, v.Name)
		}
	}
	err, result := db.ListMainPageRole(roleList)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		if result == nil {
			result = []*m.MainPageRoleQuery{}
		}
		mid.ReturnSuccessData(c, result)
	}
}

func UpdateMainPage(c *gin.Context) {
	var param []m.MainPageRoleQuery
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if err = db.UpdateMainPageRole(param); err != nil {
		mid.ReturnServerHandleError(c, err)
		return
	}
	mid.ReturnSuccess(c)
}

func GetEndpointsByIp(c *gin.Context) {
	ipList := c.QueryArray("ip")
	if len(ipList) <= 0 {
		mid.ReturnParamEmptyError(c, "ip")
		return
	}
	err, endpoints := db.GetEndpointsByIp(ipList, "host")
	if err != nil {
		mid.ReturnQueryTableError(c, "endpoint", err)
		return
	}
	mid.ReturnSuccessData(c, endpoints)
}

func DisplayWatermark(c *gin.Context) {
	result := m.DisplayDemoFlagDto{Display: true}
	isDisplay := strings.ToLower(os.Getenv("DEMO_FLAG"))
	if isDisplay == "n" || isDisplay == "no" || isDisplay == "false" {
		result.Display = false
	}
	mid.ReturnSuccessData(c, result)
}

func GetDashboardPanelList(c *gin.Context) {
	endpointType := c.Query("type")
	metric := c.Query("metric")
	if endpointType == "" || metric == "" {
		mid.ReturnValidateError(c, "Param type and metric can not empty")
		return
	}
	result := db.GetDashboardPanelList(endpointType, metric)
	mid.ReturnSuccessData(c, result)
}
