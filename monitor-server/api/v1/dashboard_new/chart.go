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
	if len(param.Data) == 0 {
		middleware.ReturnValidateError(c, "Param data can not empty ")
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
	var err error
	var queryList []*models.QueryMonitorData
	var result = models.EChartOption{Legend: []string{}, Series: []*models.SerialModel{}}
	if param.ChartId > 0 {
		// handle dashboard chart with config
		queryList, err = getChartConfigByChartId(&param, &result)
	} else {
		// handle custom chart
		queryList, err = getChartConfigByCustom(&param)
	}
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if len(queryList) == 0 {
		middleware.ReturnSuccessData(c, result)
		return
	}
	// query from prometheus
	err = getChartQueryData(queryList, &param, &result)
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
	param.Aggregate = chartList[0].AggType
	param.Unit = chartList[0].Unit
	result.Id = param.ChartId
	result.Title = chartList[0].Title
	queryList = []*models.QueryMonitorData{}
	for _, dataConfig := range param.Data {
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
			queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQl, Legend: chartList[0].Legend, Metric: []string{metric}, Endpoint: []string{dataConfig.Endpoint}, CompareLegend: param.Compare.CompareFirstLegend, SameEndpoint: true, Step: param.Step, Cluster: endpointObj.Cluster})
			if param.Compare.CompareFirstLegend != "" {
				queryList = append(queryList, &models.QueryMonitorData{Start: param.Compare.CompareSecondStartTimestamp, End: param.Compare.CompareSecondEndTimestamp, PromQ: tmpPromQl, Legend: chartList[0].Legend, Metric: []string{metric}, Endpoint: []string{dataConfig.Endpoint}, CompareLegend: param.Compare.CompareSecondLegend, SameEndpoint: true, Step: param.Step, Cluster: endpointObj.Cluster})
			}
		}
		if err != nil {
			break
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
	return nil
}

func getChartConfigByCustom(param *models.ChartQueryParam) (queryList []*models.QueryMonitorData, err error) {
	param.Compare = &models.ChartQueryCompareParam{CompareFirstLegend: ""}
	queryList = []*models.QueryMonitorData{}
	var endpointList []*models.EndpointTable
	for _, dataConfig := range param.Data {
		endpointList = []*models.EndpointTable{}
		tmpMonitorType := dataConfig.EndpointType
		// check endpoint if is service group
		if dataConfig.AppObject != "" {
			endpointList, err = db.GetRecursiveEndpointByTypeNew(dataConfig.AppObject, dataConfig.EndpointType)
			if err != nil {
				err = fmt.Errorf("Try to get endpoints from object:%s fail,%s ", dataConfig.AppObject, err.Error())
				break
			}
			if len(endpointList) == 0 {
				continue
			}
			param.Data[0].Endpoint = endpointList[0].Guid
		} else {
			endpointObj := models.EndpointTable{Guid: dataConfig.Endpoint}
			db.GetEndpoint(&endpointObj)
			if endpointObj.Id <= 0 {
				err = fmt.Errorf("Param data endpoint:%s can not find ", dataConfig.Endpoint)
				break
			}
			endpointList = append(endpointList, &endpointObj)
			tmpMonitorType = endpointObj.ExportType
		}
		metricLegend := "$custom"
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
				tmpPromQL, _ := db.GetPromQLByMetric(dataConfig.Metric, tmpMonitorType)
				if tmpPromQL == "" {
					if dataConfig.PromQl == "" {
						continue
					}
				} else {
					dataConfig.PromQl = tmpPromQL
				}
			}
		}
		for _, endpoint := range endpointList {
			tmpPromQL := db.ReplacePromQlKeyword(dataConfig.PromQl, dataConfig.Metric, *endpoint)
			queryList = append(queryList, &models.QueryMonitorData{Start: param.Start, End: param.End, PromQ: tmpPromQL, Legend: metricLegend, Metric: []string{dataConfig.Metric}, Endpoint: []string{endpoint.Guid}, Step: endpoint.Step, Cluster: endpoint.Cluster})
		}
	}
	return
}

func getChartQueryData(queryList []*models.QueryMonitorData, param *models.ChartQueryParam, result *models.EChartOption) error {
	serials := []*models.SerialModel{}
	var err error
	archiveQueryFlag := false
	if param.Start < (time.Now().Unix()-models.Config().ArchiveMysql.LocalStorageMaxDay*86400) && db.ArchiveEnable {
		archiveQueryFlag = true
	}
	for _, query := range queryList {
		log.Logger.Debug("Query param", log.StringList("endpoint", query.Endpoint), log.StringList("metric", query.Metric), log.Int64("start", query.Start), log.Int64("end", query.End), log.String("promQl", query.PromQ))
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
	agg := 0
	if param.Aggregate != "none" {
		agg = db.CheckAggregate(param.Start, param.End, "", param.Step, len(serials))
	}
	var processDisplayMap = make(map[string]string)
	if strings.HasPrefix(param.Data[0].Metric, "process_") {
		processDisplayMap = db.GetProcessDisplayMap(param.Data[0].Endpoint)
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
		if agg > 1 && len(s.Data) > 300 {
			s.Data = db.Aggregate(s.Data, agg, param.Aggregate)
		}
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
