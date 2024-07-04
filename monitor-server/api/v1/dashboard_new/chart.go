package dashboard_new

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

func ChartList(c *gin.Context) {
	var id, groupId int
	if c.Query("id") != "" {
		id, _ = strconv.Atoi(c.Query("id"))
	}
	if c.Query("groupId") != "" {
		groupId, _ = strconv.Atoi(c.Query("groupId"))
	}
	result, err := db.ChartList(id, groupId)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func ChartCreate(c *gin.Context) {
	var param []*models.ChartTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.ChartCreate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func ChartUpdate(c *gin.Context) {
	var param []*models.ChartTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.ChartUpdate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func ChartDelete(c *gin.Context) {
	ids := c.Query("ids")
	err := db.ChartDelete(strings.Split(ids, ","))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func GetChartData(c *gin.Context) {
	var param models.ChartQueryParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if len(param.Data) == 0 && param.ChartId == 0 && param.CustomChartGuid == "" {
		middleware.ReturnValidateError(c, "Param config can not empty ")
		return
	}
	if param.Start == 0 && param.End == 0 && param.TimeSecond < 0 {
		param.End = time.Now().Unix()
		param.Start = param.End + param.TimeSecond
	}
	param.Step = 10
	if param.Aggregate == "" {
		param.Aggregate = "avg"
	}
	for _, v := range param.Data {
		if v.MonitorType != "" {
			v.EndpointType = v.MonitorType
		}
	}
	var err error
	var queryList []*models.QueryMonitorData
	var result = models.EChartOption{Legend: []string{}, Series: []*models.SerialModel{}}
	if param.ChartId > 0 {
		// handle dashboard chart with config
		queryList, err = getChartConfigByChartId(&param, &result)
	} else if param.CustomChartGuid != "" {
		queryList, err = GetCustomChartConfig(&param, &result)
	} else {
		// handle custom chart
		queryList, err = GetChartConfigByCustom(&param)
	}
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if len(queryList) == 0 {
		middleware.ReturnSuccessData(c, result)
		return
	}
	log.Logger.Debug("chartData param", log.JsonObj("param", param))
	// query from prometheus
	err = GetChartQueryData(queryList, &param, &result)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func getChartConfigByChartId(param *models.ChartQueryParam, result *models.EChartOption) (queryList []*models.QueryMonitorData, err error) {
	chartList, queryChartErr := db.ChartList(param.ChartId, 0)
	if queryChartErr != nil {
		err = fmt.Errorf("Try to query chart table fail,%s ", queryChartErr.Error())
		return
	}
	if len(chartList) == 0 {
		err = fmt.Errorf("Can not find chart with id:%d ", param.ChartId)
		return
	}
	err = chartCompare(param)
	if err != nil {
		return
	}
	//param.Aggregate = chartList[0].AggType
	param.Unit = chartList[0].Unit
	result.Id = param.ChartId
	result.Title = chartList[0].Title
	queryList = []*models.QueryMonitorData{}
	existEndpointMap := make(map[string]int)
	for _, dataConfig := range param.Data {
		if _, b := existEndpointMap[dataConfig.Endpoint]; b {
			continue
		}
		existEndpointMap[dataConfig.Endpoint] = 1
		endpointObj := models.EndpointTable{Guid: dataConfig.Endpoint}
		db.GetEndpoint(&endpointObj)
		if endpointObj.Id <= 0 {
			err = fmt.Errorf("Param data endpoint can not found with guid:%s ", dataConfig.Endpoint)
			break
		}
		tmpMetric := chartList[0].Metric
		if strings.Contains(dataConfig.Metric, "/") || tmpMetric == "app.metric" {
			tmpMetric = dataConfig.Metric
		}
		for _, metric := range strings.Split(tmpMetric, "^") {
			tmpPromQl := ""
			if chartList[0].Metric == "db_monitor_count" && metric != "db_monitor_count" {
				err, tmpPromQl = db.GetDbPromMetric(dataConfig.Endpoint, dataConfig.Metric, chartList[0].Legend)
			} else {
				err, tmpPromQl = db.GetPromMetric([]string{dataConfig.Endpoint}, metric)
			}
			if err != nil {
				break
			}
			tmpLegend := chartList[0].Legend
			if len(param.Data) > 1 && strings.HasPrefix(tmpLegend, "$custom_") {
				tmpLegend = "$custom"
			}
			queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQl, Legend: tmpLegend, Metric: []string{metric}, Endpoint: []string{dataConfig.Endpoint}, CompareLegend: param.Compare.CompareFirstLegend, SameEndpoint: true, Step: param.Step, Cluster: endpointObj.Cluster})
			if param.Compare.CompareFirstLegend != "" {
				queryList = append(queryList, &models.QueryMonitorData{Start: param.Compare.CompareSecondStartTimestamp, End: param.Compare.CompareSecondEndTimestamp, PromQ: tmpPromQl, Legend: tmpLegend, Metric: []string{metric}, Endpoint: []string{dataConfig.Endpoint}, CompareLegend: param.Compare.CompareSecondLegend, SameEndpoint: true, Step: param.Step, Cluster: endpointObj.Cluster})
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func GetCustomChartConfig(param *models.ChartQueryParam, result *models.EChartOption) (queryList []*models.QueryMonitorData, err error) {
	chartObj, getChartErr := db.GetCustomChartById(param.CustomChartGuid)
	if getChartErr != nil {
		err = fmt.Errorf("get custom chart with guid:%s fail,%s ", param.CustomChartGuid, getChartErr.Error())
		return
	}
	chartSeries, getErr := db.GetCustomChartSeries(param.CustomChartGuid)
	if getErr != nil {
		err = getErr
		return
	}
	if len(chartSeries) == 0 {
		err = fmt.Errorf("Can not find chart serie with guid:%d ", param.CustomChartGuid)
		return
	}
	err = chartCompare(param)
	if err != nil {
		return
	}
	//param.Aggregate = chartList[0].AggType
	param.Unit = chartObj.Unit
	result.Title = chartObj.Name
	queryList = []*models.QueryMonitorData{}
	legend := "$custom"
	for _, dataConfig := range chartSeries {
		log.Logger.Debug("chart series display config", log.JsonObj("dataConfig", dataConfig))
		tmpPromQl := ""
		tmpPromQl, err = db.GetPromQLByMetric(dataConfig.Metric, dataConfig.MonitorType, dataConfig.ServiceGroup)
		if err != nil {
			break
		}
		isServiceMetric, tmpTags, tmpErr := db.CheckMetricIsServiceMetric(dataConfig.Metric, dataConfig.ServiceGroup)
		if tmpErr != nil {
			err = tmpErr
			return
		}
		if isServiceMetric {
			log.Logger.Debug("getChartConfigByCustom $app_metric")
			legend = "$app_metric"
			tmpPromQl = db.ReplacePromQlKeyword(tmpPromQl, dataConfig.Metric, &models.EndpointNewTable{}, dataConfig.Tags)
			queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQl, Legend: legend, Metric: []string{dataConfig.Metric}, Endpoint: []string{dataConfig.Endpoint}, CompareLegend: param.Compare.CompareFirstLegend, SameEndpoint: true, Step: param.Step, Cluster: "default", CustomDashboard: true, Tags: tmpTags})
		} else {
			endpointList := []*models.EndpointNewTable{}
			if dataConfig.ServiceGroup == "" {
				endpointObj, _ := db.GetEndpointNew(&models.EndpointNewTable{Guid: dataConfig.Endpoint})
				if endpointObj.MonitorType == "" {
					err = fmt.Errorf("Param data endpoint:%s can not find ", dataConfig.Endpoint)
					break
				}
				endpointList = append(endpointList, &endpointObj)
			} else {
				endpointList, err = db.GetRecursiveEndpointByTypeNew(dataConfig.ServiceGroup, dataConfig.MonitorType)
				if err != nil {
					err = fmt.Errorf("Try to get endpoints from serviceGroup:%s fail,%s ", dataConfig.ServiceGroup, err.Error())
					break
				}
				if len(endpointList) == 0 {
					continue
				}
			}
			for _, endpoint := range endpointList {
				tmpPromQL := db.ReplacePromQlKeyword(tmpPromQl, dataConfig.Metric, endpoint, dataConfig.Tags)
				queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQL, Legend: legend, Metric: []string{dataConfig.Metric}, Endpoint: []string{endpoint.Guid}, Step: endpoint.Step, Cluster: endpoint.Cluster, CustomDashboard: true})
			}
		}
	}
	return
}

func chartCompare(param *models.ChartQueryParam) error {
	var err error
	if param.Compare == nil {
		param.Compare = &models.ChartQueryCompareParam{CompareFirstLegend: ""}
		return err
	}
	if param.Compare.CompareFirstStart == "" {
		param.Compare.CompareFirstLegend = ""
		return err
	}
	firstStartTime, parseErr := time.Parse(models.DateFormatWithZone, fmt.Sprintf("%s 00:00:00 "+models.DefaultLocalTimeZone, param.Compare.CompareFirstStart))
	if parseErr != nil {
		return fmt.Errorf("Param compare first start:%s format fail:%s ", param.Compare.CompareFirstStart, parseErr.Error())
	}
	firstEndTime, parseErr := time.Parse(models.DateFormatWithZone, fmt.Sprintf("%s 23:59:59 "+models.DefaultLocalTimeZone, param.Compare.CompareFirstEnd))
	if parseErr != nil {
		return fmt.Errorf("Param compare first end:%s format fail:%s ", param.Compare.CompareFirstEnd, parseErr.Error())
	}
	secondStartTime, parseErr := time.Parse(models.DateFormatWithZone, fmt.Sprintf("%s 00:00:00 "+models.DefaultLocalTimeZone, param.Compare.CompareSecondStart))
	if parseErr != nil {
		return fmt.Errorf("Param compare second start:%s format fail:%s ", param.Compare.CompareSecondStart, parseErr.Error())
	}
	secondEndTime, parseErr := time.Parse(models.DateFormatWithZone, fmt.Sprintf("%s 23:59:59 "+models.DefaultLocalTimeZone, param.Compare.CompareSecondEnd))
	if parseErr != nil {
		return fmt.Errorf("Param compare second end:%s format fail:%s ", param.Compare.CompareSecondEnd, parseErr.Error())
	}
	param.Start = firstStartTime.Unix()
	param.End = firstEndTime.Unix()
	param.Compare.CompareSecondStartTimestamp = secondStartTime.Unix()
	param.Compare.CompareSecondEndTimestamp = secondEndTime.Unix()
	param.Compare.CompareFirstLegend = fmt.Sprintf("%s_%s", param.Compare.CompareFirstStart, param.Compare.CompareFirstEnd)
	param.Compare.CompareSubTime = param.Compare.CompareSecondStartTimestamp - param.Start
	param.Compare.CompareSecondLegend = fmt.Sprintf("%s_%s", param.Compare.CompareSecondStart, param.Compare.CompareSecondEnd)
	return nil
}

func GetChartConfigByCustom(param *models.ChartQueryParam) (queryList []*models.QueryMonitorData, err error) {
	param.Compare = &models.ChartQueryCompareParam{CompareFirstLegend: ""}
	queryList = []*models.QueryMonitorData{}
	var endpointList []*models.EndpointNewTable
	var serviceGroupTag string
	for _, dataConfig := range param.Data {
		endpointList = []*models.EndpointNewTable{}
		tmpMonitorType := dataConfig.EndpointType
		metricLegend := "$custom"
		customPromQL := dataConfig.PromQl
		if dataConfig.PromQl != "" {
			metricLegend = "$custom_with_tag"
		}
		serviceTags := []string{}
		// check endpoint if is service group
		if dataConfig.AppObject != "" {
			serviceGroupTag = fmt.Sprintf("service_group=\"%s\"", dataConfig.AppObject)
			endpointList, err = db.GetRecursiveEndpointByTypeNew(dataConfig.AppObject, dataConfig.EndpointType)
			if err != nil {
				err = fmt.Errorf("Try to get endpoints from object:%s fail,%s ", dataConfig.AppObject, err.Error())
				break
			}
			if len(endpointList) == 0 {
				continue
			}
			param.Data[0].Endpoint = endpointList[0].Guid
			log.Logger.Debug("getChartConfigByCustom", log.String("app", dataConfig.AppObject), log.String("metric", dataConfig.Metric))
			isServiceMetric, tmpTags, tmpErr := db.CheckMetricIsServiceMetric(dataConfig.Metric, dataConfig.AppObject)
			if tmpErr != nil {
				err = tmpErr
				return
			}
			if isServiceMetric {
				serviceTags = tmpTags
				log.Logger.Debug("getChartConfigByCustom $app_metric")
				metricLegend = "$app_metric"
			}
		} else {
			//endpointObj := models.EndpointTable{Guid: dataConfig.Endpoint}
			//db.GetEndpoint(&endpointObj)
			endpointObj, _ := db.GetEndpointNew(&models.EndpointNewTable{Guid: dataConfig.Endpoint})
			if endpointObj.MonitorType == "" {
				err = fmt.Errorf("Param data endpoint:%s can not find ", dataConfig.Endpoint)
				break
			}
			endpointList = append(endpointList, &endpointObj)
			tmpMonitorType = endpointObj.MonitorType
		}
		if dataConfig.Metric != "" {
			if strings.HasPrefix(dataConfig.Metric, models.LogMetricName) || strings.HasPrefix(dataConfig.Metric, models.DBMonitorMetricName) {
				metricLegend = "$app_metric"
				if dataConfig.PromQl == "" {
					tmpSplitIndex := strings.Index(dataConfig.Metric, "/")
					tmpTags := dataConfig.Metric[tmpSplitIndex+1:]
					tmpTags = strings.ReplaceAll(tmpTags, ",", "\",")
					tmpTags = strings.ReplaceAll(tmpTags, "=", "=\"")
					dataConfig.PromQl = fmt.Sprintf("%s{%s\"}", dataConfig.Metric[:tmpSplitIndex], tmpTags)
				}
			} else {
				tmpPromQL, _ := db.GetPromQLByMetric(dataConfig.Metric, tmpMonitorType, dataConfig.AppObject)
				if tmpPromQL == "" {
					if dataConfig.PromQl == "" {
						continue
					}
				} else {
					dataConfig.PromQl = tmpPromQL
				}
			}
		}
		//queryAppendFlag := false
		if len(endpointList) > 0 && metricLegend == "$app_metric" {
			tmpPromQL := db.ReplacePromQlKeyword(dataConfig.PromQl, dataConfig.Metric, endpointList[0], dataConfig.Tags)
			queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQL, Legend: metricLegend, Metric: []string{dataConfig.Metric}, Endpoint: []string{endpointList[0].Guid}, Step: endpointList[0].Step, Cluster: endpointList[0].Cluster, CustomDashboard: true, Tags: serviceTags})
			continue
			//log.Logger.Debug("check prom is same", log.String("tmpPromQl", tmpPromQL), log.String("dataProm", dataConfig.PromQl))
			//if tmpPromQL == dataConfig.PromQl {
			//	queryAppendFlag = true
			//	log.Logger.Debug("prom is same")
			//	queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQL, Legend: metricLegend, Metric: []string{dataConfig.Metric}, Endpoint: []string{endpointList[0].Guid}, Step: endpointList[0].Step, Cluster: endpointList[0].Cluster, CustomDashboard: true})
			//}
		}
		for _, endpoint := range endpointList {
			tmpPromQL := dataConfig.PromQl
			if customPromQL != "" && serviceGroupTag != "" && strings.Contains(tmpPromQL, serviceGroupTag) {
				tmpPromQL = strings.ReplaceAll(tmpPromQL, serviceGroupTag, serviceGroupTag+",instance=\"$address\"")
				if strings.Contains(tmpPromQL, "service_group,") {
					tmpPromQL = strings.ReplaceAll(tmpPromQL, "service_group,", "service_group,instance,")
				}
				if strings.Contains(tmpPromQL, "service_group)") {
					tmpPromQL = strings.ReplaceAll(tmpPromQL, "service_group)", "service_group,instance)")
				}
				log.Logger.Debug("build custom chart query", log.String("tmpPromQL", tmpPromQL))
			}
			tmpPromQL = db.ReplacePromQlKeyword(tmpPromQL, dataConfig.Metric, endpoint, dataConfig.Tags)
			queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQL, Legend: metricLegend, Metric: []string{dataConfig.Metric}, Endpoint: []string{endpoint.Guid}, Step: endpoint.Step, Cluster: endpoint.Cluster, CustomDashboard: true})
		}
	}
	return
}

func GetChartComparisonQueryData(queryList []*models.QueryMonitorData, param models.ComparisonChartQueryParam) (result *models.EChartOption, err error) {
	var serials []*models.SerialModel
	var difference int64
	result = &models.EChartOption{Legend: []string{}, Series: []*models.SerialModel{}}
	calcTypeMap := convertArray2Map(param.CalcType)
	for _, query := range queryList {
		if query.Cluster != "" && query.Cluster != "default" {
			query.Cluster = db.GetClusterAddress(query.Cluster)
		}
		tmpSerials := ds.PrometheusData(query)
		switch param.ComparisonType {
		case "day":
			difference = 86400
		case "week":
			difference = 86400 * 7
		case "month":
			// 预览数据,一个月当作30d处理
			difference = 86400 * 30
		}
		query.Start = query.Start - difference
		query.End = query.End - difference
		tmpSerials2 := ds.PrometheusData(query)
		var comparisonSerialList []*models.SerialModel
		// 计算同环比数据
		if len(tmpSerials2) > 0 && len(tmpSerials) > 0 {
			for i, model := range tmpSerials2 {
				serialModel := &models.SerialModel{
					Type: "bar",
					Name: getNewName(model.Name, "diff"),
					Data: make([][]float64, len(model.Data)),
				}
				serialModel2 := &models.SerialModel{
					Type:       "line",
					YAxisIndex: 1,
					Name:       getNewName(model.Name, "diff_percent"),
					Data:       make([][]float64, len(model.Data)),
				}
				if i < len(tmpSerials) && len(model.Data) > 0 && len(tmpSerials[i].Data) > 0 {
					for k, arr := range model.Data {
						if len(arr) == 2 && k < len(tmpSerials[i].Data) && len(tmpSerials[i].Data[k]) == 2 && arr[0]+float64(difference)*1000 == tmpSerials[i].Data[k][0] {
							if calcTypeMap["diff"] {
								serialModel.Data[k] = append(serialModel.Data[k], []float64{tmpSerials[i].Data[k][0], RoundToOneDecimal(arr[1] - tmpSerials[i].Data[k][1])}...)
							}
							if calcTypeMap["diff_percent"] {
								val := float64(0)
								if tmpSerials[i].Data[k][1] != 0 {
									val = ((arr[1] - tmpSerials[i].Data[k][1]) * 1.0 / tmpSerials[i].Data[k][1]) * 100
								}
								val = RoundToOneDecimal(val)
								serialModel2.Data[k] = append(serialModel2.Data[k], []float64{tmpSerials[i].Data[k][0], val}...)
							}
						}
					}
				}
				if calcTypeMap["diff"] {
					comparisonSerialList = append(comparisonSerialList, serialModel)
				}
				if calcTypeMap["diff_percent"] {
					comparisonSerialList = append(comparisonSerialList, serialModel2)
				}
			}
		}
		if len(comparisonSerialList) > 0 {
			serials = append(serials, comparisonSerialList...)
		}
	}
	processDisplayMap := make(map[string]string)
	for i, s := range serials {
		if strings.Contains(s.Name, "$metric") {
			queryIndex := i
			if i >= len(queryList) {
				queryIndex = len(queryList) - 1
			}
			s.Name = strings.Replace(s.Name, "$metric", queryList[queryIndex].Metric[0], -1)
		}
		if processName, b := processDisplayMap[s.Name]; b {
			s.Name = processName
		}
		result.Legend = append(result.Legend, s.Name)
		if result.Title == "${auto}" {
			result.Title = s.Name[:strings.Index(s.Name, "{")]
		}
		_, tmpJsonMarshalErr := json.Marshal(s)
		if tmpJsonMarshalErr == nil {
			result.Series = append(result.Series, s)
		}
	}
	result.Xaxis = make(map[string]interface{})
	result.Yaxis = models.YaxisModel{Unit: ""}
	return
}

func RoundToOneDecimal(value float64) float64 {
	v := strconv.FormatFloat(value, 'f', 1, 64)
	fmt.Println(v)
	floatValue, _ := strconv.ParseFloat(v, 64)
	return floatValue
}

func getNewName(name, calcType string) string {
	var newName string
	if strings.TrimSpace(name) != "" {
		start2 := strings.Index(name, "}")
		if start2 == -1 {
			newName = name + fmt.Sprintf("{calc_type=%s}", calcType)
		} else {
			newName = name[0:start2-1] + fmt.Sprintf(",calc_type=%s}", calcType)
		}
	}
	return newName
}

func convertArray2Map(arr []string) map[string]bool {
	hashMap := make(map[string]bool)
	if len(arr) == 0 {
		return hashMap
	}
	for _, s := range arr {
		hashMap[s] = true
	}
	return hashMap
}

func GetChartQueryData(queryList []*models.QueryMonitorData, param *models.ChartQueryParam, result *models.EChartOption) error {
	serials := []*models.SerialModel{}
	var err error
	archiveQueryFlag := false
	if param.Start < (time.Now().Unix()-models.Config().ArchiveMysql.LocalStorageMaxDay*86400) && db.ArchiveEnable {
		archiveQueryFlag = true
	}
	for _, query := range queryList {
		log.Logger.Debug("Query param", log.JsonObj("param", query))
		if query.Cluster != "" && query.Cluster != "default" {
			query.Cluster = db.GetClusterAddress(query.Cluster)
		}
		if archiveQueryFlag {
			// Query db archive data
			tmpErr, tmpStep, tmpSerials := db.GetArchiveData(query, param.Aggregate)
			if tmpErr != nil {
				err = tmpErr
				break
			}
			serials = append(serials, tmpSerials...)
			param.Step = tmpStep
			continue
		}
		tmpSerials := ds.PrometheusData(query)
		// 如果归档数据可用，尝试从归档数据中补全数据
		if db.ArchiveEnable {
			if len(tmpSerials) > 0 {
				if len(tmpSerials[0].Data) > 0 {
					tmpSerialDataStart := int64(tmpSerials[0].Data[0][0]) / 1000
					// 如果查出来的数据时间和查询时间对不上，说明缺了一些数据，尝试从归档数据中去查
					if tmpSerialDataStart > (param.Start + 120) {
						_, _, tmpArchiveSerials := db.GetArchiveData(&models.QueryMonitorData{Start: query.Start, End: tmpSerialDataStart, Endpoint: query.Endpoint, Metric: query.Metric, Legend: query.Legend, CompareLegend: query.CompareLegend, SameEndpoint: query.SameEndpoint}, param.Aggregate)
						for _, tmpSerial := range tmpArchiveSerials {
							if len(tmpSerial.Data) > 0 {
								param.Aggregate = "none"
								for si, vv := range tmpSerials {
									if tmpSerial.Name == vv.Name {
										tmpSerials[si].Data = append(tmpSerial.Data, vv.Data...)
									}
								}
							}
						}
					}
				}
			} else {
				tmpErr, tmpStep, tmpSerials := db.GetArchiveData(&models.QueryMonitorData{Start: query.Start, End: query.End, Endpoint: query.Endpoint, Metric: query.Metric, Legend: query.Legend, CompareLegend: query.CompareLegend, SameEndpoint: query.SameEndpoint}, param.Aggregate)
				if tmpErr != nil {
					err = tmpErr
					break
				}
				serials = append(serials, tmpSerials...)
				param.Step = tmpStep
				continue
			}
		}
		serials = append(serials, tmpSerials...)
	}
	// handle serials data
	//agg := 0
	//if param.Aggregate != "none" {
	//	agg = db.CheckAggregate(param.Start, param.End, "", param.Step, len(serials))
	//}
	processDisplayMap := make(map[string]string)
	if len(param.Data) > 0 {
		if strings.HasPrefix(param.Data[0].Metric, "process_") {
			processDisplayMap = db.GetProcessDisplayMap(param.Data[0].Endpoint)
		}
	}
	if len(serials) > 0 && param.Aggregate != "none" && param.AggStep <= 0 {
		param.AggStep = int64(db.CheckAggregate(param.Start, param.End, "", 10, len(serials)))
	}
	for i, s := range serials {
		if strings.Contains(s.Name, "$metric") {
			queryIndex := i
			if i >= len(queryList) {
				queryIndex = len(queryList) - 1
			}
			s.Name = strings.Replace(s.Name, "$metric", queryList[queryIndex].Metric[0], -1)
		}
		if processName, b := processDisplayMap[s.Name]; b {
			s.Name = processName
		}
		result.Legend = append(result.Legend, s.Name)
		if result.Title == "${auto}" {
			result.Title = s.Name[:strings.Index(s.Name, "{")]
		}
		if param.Aggregate != "none" && param.AggStep > 10 {
			log.Logger.Debug("AggregateNew", log.Int64("aggStep", param.AggStep), log.String("agg", param.Aggregate))
			s.Data = models.Aggregate(s.Data, param.AggStep, param.Aggregate)
		}
		//if agg > 1 && len(s.Data) > 300 {
		//	s.Data = db.Aggregate(s.Data, agg, param.Aggregate)
		//}
		if param.Compare.CompareSubTime > 0 {
			if strings.Contains(s.Name, param.Compare.CompareSecondLegend) {
				s.Data = db.CompareSubData(s.Data, float64(param.Compare.CompareSubTime)*1000)
			}
		}
		_, tmpJsonMarshalErr := json.Marshal(s)
		if tmpJsonMarshalErr == nil {
			result.Series = append(result.Series, s)
		}
	}
	if param.ChartId == 0 && param.Title != "" {
		result.Title = param.Title
	}
	result.Xaxis = make(map[string]interface{})
	result.Yaxis = models.YaxisModel{Unit: param.Unit}
	return err
}

// GetComparisonChartData 获取同环比预览数据
func GetComparisonChartData(c *gin.Context) {
	var param models.ComparisonChartQueryParam
	var queryParam *models.ChartQueryParam
	var err error
	var metric *models.MetricTable
	var queryList []*models.QueryMonitorData
	var result = &models.EChartOption{Legend: []string{}, Series: []*models.SerialModel{}}
	now := time.Now()
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if strings.TrimSpace(param.MetricId) == "" {
		middleware.ReturnParamEmptyError(c, "metricId")
		return
	}
	if metric, err = db.GetMetric(param.MetricId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if metric == nil {
		middleware.ReturnValidateError(c, "metricId is invalid")
		return
	}
	queryParam = &models.ChartQueryParam{
		Start: now.Unix() + param.TimeSecond,
		End:   now.Unix(),
		Step:  10,
		Data: []*models.ChartQueryConfigObj{{
			Endpoint:     param.Endpoint,
			Metric:       metric.Metric,
			PromQl:       metric.PromExpr,
			AppObject:    metric.ServiceGroup,
			EndpointType: metric.MonitorType,
			MonitorType:  metric.MonitorType,
		}},
	}
	if queryList, err = GetChartConfigByCustom(queryParam); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(queryList) == 0 {
		middleware.ReturnSuccessData(c, result)
		return
	}
	if result, err = GetChartComparisonQueryData(queryList, param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}
