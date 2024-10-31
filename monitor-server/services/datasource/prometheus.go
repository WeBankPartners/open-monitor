package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var promDS DataSourceParam

func InitPrometheusDatasource() {
	t := time.Now()
	cfg := *m.Config().Datasource.Servers[0]
	opentsdbDS := &DataSource{Id: cfg.Id, Name: cfg.Type, Url: fmt.Sprintf("http://%s", cfg.Host), IsDefault: true, Updated: t}
	promDS = DataSourceParam{DataSource: opentsdbDS, Host: cfg.Host, Token: cfg.Token}
}

var PieLegendBlackName = []string{"job", "instance", "__name__", "e_guid"}

func PrometheusData(query *m.QueryMonitorData) []*m.SerialModel {
	log.Logger.Debug("prometheus data query", log.JsonObj("queryParam", query))
	serials := []*m.SerialModel{}
	urlParams := url.Values{}
	hostAddress := promDS.Host
	if query.Cluster != "" && query.Cluster != "default" {
		hostAddress = query.Cluster
	}
	requestUrl, err := url.Parse(fmt.Sprintf("http://%s/api/v1/query_range", hostAddress))
	if err != nil {
		log.Logger.Error("Make url fail", log.Error(err))
		return serials
	}
	var tmpStep int64
	tmpStep = 10
	if query.Step > 0 && query.Step != 10 {
		tmpStep = int64(query.Step)
		if strings.Contains(query.PromQ, "20s") {
			query.PromQ = strings.Replace(query.PromQ, "20s", fmt.Sprintf("%ds", tmpStep*2), -1)
		}
	}
	subSec := query.End - query.Start
	if subSec > 86400 {
		tmpStep = tmpStep * (subSec/86400 + 1)
	}
	urlParams.Set("start", strconv.FormatInt(query.Start, 10))
	urlParams.Set("end", strconv.FormatInt(query.End, 10))
	urlParams.Set("step", fmt.Sprintf("%d", tmpStep))
	urlParams.Set("query", query.PromQ)
	requestUrl.RawQuery = urlParams.Encode()
	req, err := http.NewRequest(http.MethodGet, requestUrl.String(), strings.NewReader(""))
	if err != nil {
		log.Logger.Error("Failed to create request", log.Error(err))
		return serials
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient, err := promDS.DataSource.GetHttpClient()
	if err != nil {
		log.Logger.Error("Get httpClient fail", log.Error(err))
		return serials
	}
	res, err := ctxhttp.Do(context.Background(), httpClient, req)
	if err != nil {
		log.Logger.Error("Http request fail", log.Error(err))
		return serials
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Logger.Error("Http request body read fail", log.Error(err))
		return serials
	}
	//log.Logger.Debug("prometheus data result", log.String("response", string(body)))
	if res.StatusCode/100 != 2 {
		log.Logger.Warn("Request fail with bad status", log.String("status", res.Status))
		return serials
	}
	var data m.PrometheusResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Logger.Error("Unmarshal response fail", log.Error(err))
		return serials
	}
	if data.Status != "success" {
		log.Logger.Warn("Query prometheus data fail", log.String("status", data.Status))
		return serials
	}
	if query.ChartType == "pie" {
		buildPieData(query, data.Data.Result)
		return serials
	}
	for _, otr := range data.Data.Result {
		//if len(otr.Metric) == 0 {
		//	continue
		//}
		var serial m.SerialModel
		serial.Type = "line"
		serial.Name = GetSerialName(query, otr.Metric, len(data.Data.Result), query.CustomDashboard)
		// 同环比 指标
		if query.ComparisonFlag == "Y" {
			if otr.Metric["calc_type"] == "diff" {
				serial.Type = "bar"
			} else if otr.Metric["calc_type"] == "diff_percent" {
				serial.YAxisIndex = 1
			}
		}
		var sdata m.DataSort
		for _, v := range otr.Values {
			tmpTime := v[0].(float64) * 1000
			tmpValue, _ := strconv.ParseFloat(v[1].(string), 64)
			//tmpValue,_ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
			sdata = append(sdata, []float64{tmpTime, tmpValue})
		}
		sort.Sort(sdata)
		serial.Data = sdata
		serials = append(serials, &serial)
	}
	return serials
}

func CheckPrometheusQL(promQl string) error {
	hostAddress := promDS.Host
	requestUrl, _ := url.Parse(fmt.Sprintf("http://%s/api/v1/query_range", hostAddress))
	nowTime := time.Now().Unix()
	urlParams := url.Values{}
	urlParams.Set("start", strconv.FormatInt(nowTime-10, 10))
	urlParams.Set("end", strconv.FormatInt(nowTime, 10))
	urlParams.Set("step", "10")
	urlParams.Set("query", promQl)
	requestUrl.RawQuery = urlParams.Encode()
	req, err := http.NewRequest(http.MethodGet, requestUrl.String(), strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("Failed to create request:%s ", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient, getClientErr := promDS.DataSource.GetHttpClient()
	if getClientErr != nil {
		return fmt.Errorf("Get httpClient fail:%s ", getClientErr.Error())
	}
	res, doHttpErr := ctxhttp.Do(context.Background(), httpClient, req)
	if doHttpErr != nil {
		return fmt.Errorf("Http request fail:%s ", doHttpErr.Error())
	}
	_, err = ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("Http request body read fail:%s ", err.Error())
	}
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Request fail with bad statusCode: %d ", res.StatusCode)
	}
	return nil
}

func buildPieData(query *m.QueryMonitorData, dataList []m.PrometheusResult) {
	pieData := m.EChartPie{}
	useNewValue := true
	if query.PieAggType != "new" {
		useNewValue = false
	}
	//log.Logger.Debug("buildPieData", log.String("pieAggType", query.PieAggType), log.JsonObj("dataList", dataList))
	for _, otr := range dataList {
		var tmpNameList []string
		for k, v := range otr.Metric {
			// 标签黑名单
			ignoreFlag := false
			for _, vv := range PieLegendBlackName {
				if k == vv {
					ignoreFlag = true
					break
				}
			}
			if ignoreFlag {
				continue
			}
			if len(query.Tags) > 0 {
				legalFlag := false
				for _, legalTag := range query.Tags {
					if legalTag == k {
						legalFlag = true
						break
					}
				}
				if !legalFlag {
					continue
				}
			}
			tmpName := v
			if k != "tags" {
				tmpName = fmt.Sprintf("%s=%s", k, v)
			}
			tmpNameList = append(tmpNameList, tmpName)
		}
		pieObj := m.EChartPieObj{}
		pieObj.Name = strings.Join(tmpNameList, ",")
		pieObj.NameList = tmpNameList
		log.Logger.Debug("pieData", log.String("pieName", pieObj.Name), log.StringList("tmpNameList", tmpNameList))
		if pieObj.Name == "" {
			pieObj.Name = query.Endpoint[0] + "__" + query.Metric[0]
		}
		if len(otr.Values) > 0 {
			if useNewValue {
				// 取最新值
				pieObj.Value, _ = strconv.ParseFloat(otr.Values[len(otr.Values)-1][1].(string), 64)
				pieObj.Value, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", pieObj.Value), 64)
			} else {
				// 按合并规则取值
				var valueDataList []float64
				for _, v := range otr.Values {
					tmpValue, _ := strconv.ParseFloat(v[1].(string), 64)
					valueDataList = append(valueDataList, tmpValue)
				}
				pieObj.SourceValue = valueDataList
				pieObj.Value = m.CalcData(valueDataList, query.PieAggType)
			}
		}
		//log.Logger.Info("buildPidData otr", log.String("name", pieObj.Name), log.Float64("value", pieObj.Value))
		//if existPie, ok := pieMap[pieObj.Name]; ok {
		//	existPie.Value = m.CalcData([]float64{existPie.Value, pieObj.Value}, query.PieAggType)
		//	continue
		//} else {
		//	pieMap[pieObj.Name] = &pieObj
		//}
		//log.Logger.Info("buildPidData otr append", log.String("name", pieObj.Name))
		pieData.Legend = append(pieData.Legend, pieObj.Name)
		pieData.Data = append(pieData.Data, &pieObj)
	}
	if query.PieDisplayTag != "" {
		log.Logger.Debug("start build pie display tag value", log.String("PieDisplayTag", query.PieDisplayTag))
		displayPieData := m.EChartPie{Title: pieData.Title}
		dataValueMap := make(map[string]float64)
		for _, dataObj := range pieData.Data {
			matchTagKey := ""
			for _, v := range dataObj.NameList {
				if strings.HasPrefix(v, query.PieDisplayTag+"=") {
					matchTagKey = v
					break
				}
			}
			if existValue, ok := dataValueMap[matchTagKey]; ok {
				dataValueMap[matchTagKey] = existValue + dataObj.Value
			} else {
				displayPieData.Legend = append(displayPieData.Legend, matchTagKey)
				dataValueMap[matchTagKey] = dataObj.Value
			}
		}
		for _, legend := range displayPieData.Legend {
			tmpPieObj := m.EChartPieObj{Name: legend, Value: dataValueMap[legend]}
			displayPieData.Data = append(displayPieData.Data, &tmpPieObj)
		}
		pieData = displayPieData
	}
	query.PieData = pieData
}

func appendTagString(name string, metricMap map[string]string, tagList []string) string {
	var tmpList m.DefaultSortList
	if len(tagList) == 0 {
		for k, v := range metricMap {
			if k == "exported_instance" || k == "exported_job" {
				continue
			}
			if k == "instance" && v == "127.0.0.1:8181" {
				continue
			}
			tmpList = append(tmpList, &m.DefaultSortObj{Key: k, Value: v})
		}
		sort.Sort(tmpList)
	} else {
		for _, v := range tagList {
			if tagValue, b := metricMap[v]; b {
				tmpList = append(tmpList, &m.DefaultSortObj{Key: v, Value: tagValue})
			}
		}
	}
	tmpName := name + "{"
	for _, v := range tmpList {
		ignoreFlag := false
		for _, vv := range m.DashboardIgnoreTagKey {
			if v.Key == vv {
				ignoreFlag = true
				break
			}
		}
		if ignoreFlag {
			continue
		}
		tmpName += fmt.Sprintf("%s=%s,", v.Key, v.Value)
	}
	tmpName = tmpName[:len(tmpName)-1]
	if tmpName != name {
		tmpName += "}"
	}
	return tmpName
}

func GetSerialName(query *m.QueryMonitorData, tagMap map[string]string, dataLength int, metricFirst bool) string {
	tmpName := query.Legend
	legend := query.Legend
	var endpoint, metric string
	if len(query.Endpoint) > 0 {
		endpoint = query.Endpoint[0]
	}
	if len(query.Metric) > 0 {
		metric = query.Metric[0]
	}
	for k, v := range tagMap {
		if metric == "" && k == "__name__" {
			metric = v
		}
		if strings.Contains(legend, "$"+k) {
			tmpName = strings.Replace(tmpName, "$"+k, k+"="+v, -1)
			if !query.SameEndpoint {
				tmpName = fmt.Sprintf("%s:%s", endpoint, tmpName)
			}
		}
	}
	if strings.Contains(legend, "$custom") {
		if legend == "$custom" {
			if metricFirst {
				tmpName = fmt.Sprintf("%s:%s", metric, endpoint)
			} else {
				tmpName = fmt.Sprintf("%s:%s", endpoint, metric)
			}
			tmpName = appendTagString(tmpName, tagMap, []string{})
		} else if legend == "$custom_metric" {
			tmpName = metric
			if dataLength > 1 {
				tmpName = appendTagString(tmpName, tagMap, []string{})
			}
		} else if legend == "$custom_endpoint" {
			tmpName = endpoint
			if dataLength > 1 {
				tmpName = appendTagString(tmpName, tagMap, []string{})
			}
		} else if legend == "$custom_all" {
			if metricFirst {
				tmpName = fmt.Sprintf("%s:%s", metric, endpoint)
			} else {
				tmpName = fmt.Sprintf("%s:%s", endpoint, metric)
			}
			tmpName = appendTagString(tmpName, tagMap, []string{})
		} else if legend == "$custom_with_tag" {
			if metricFirst {
				tmpName = fmt.Sprintf("%s:%s", metric, endpoint)
			} else {
				tmpName = fmt.Sprintf("%s:%s", endpoint, metric)
			}
			tagKeyList := []string{}
			for k, _ := range tagMap {
				tagKeyList = append(tagKeyList, k)
			}
			sort.Strings(tagKeyList)
			tmpName = appendTagString(tmpName, tagMap, tagKeyList)
		}
	}
	if legend == "$metric" || legend == "$custom_metric" {
		if !strings.Contains(tmpName, ":") && !query.SameEndpoint {
			tmpName = fmt.Sprintf("%s:%s", endpoint, tmpName)
		}
	}
	if legend == "$app_metric" {
		if serviceGroup, b := tagMap["service_group"]; b {
			if serviceGroupName, bb := m.GlobalSGDisplayNameMap[serviceGroup]; bb {
				if metricFirst {
					tmpName = fmt.Sprintf("%s:%s", metric, serviceGroupName)
				} else {
					tmpName = fmt.Sprintf("%s:%s", serviceGroupName, metric)
				}
			} else {
				if metricFirst {
					tmpName = fmt.Sprintf("%s:%s", metric, serviceGroup)
				} else {
					tmpName = fmt.Sprintf("%s:%s", serviceGroup, metric)
				}
			}
			tmpTagList := query.Tags
			tmpTagList = append(tmpTagList, "t_endpoint", "instance", "calc_type")
			if query.ServiceConfiguration == "custom" {
				tagMap = ResetPrometheusMetricMap(tagMap)
			}
			tmpName = appendTagString(tmpName, tagMap, tmpTagList)
		} else {
			tmpName = metric
		}
	}
	if query.CompareLegend != "" {
		tmpName = fmt.Sprintf("%s_%s", query.CompareLegend, tmpName)
	}
	return tmpName
}

func QueryPromQLMetric(promQl, address string, start, end int64) (metricList []string, err error) {
	requestUrl, _ := url.Parse(fmt.Sprintf("http://%s/api/v1/query_range", address))
	urlParams := url.Values{}
	urlParams.Set("start", strconv.FormatInt(start, 10))
	urlParams.Set("end", strconv.FormatInt(end, 10))
	urlParams.Set("step", "10")
	urlParams.Set("query", promQl)
	requestUrl.RawQuery = urlParams.Encode()
	req, newReqErr := http.NewRequest(http.MethodGet, requestUrl.String(), strings.NewReader(""))
	if newReqErr != nil {
		err = fmt.Errorf("new request error:%s ", newReqErr.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = fmt.Errorf("http response error:%s ", respErr.Error())
		return
	}
	body, readBodyErr := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if readBodyErr != nil {
		err = fmt.Errorf("http request body read fail,%s ", readBodyErr.Error())
		return
	}
	if resp.StatusCode/100 != 2 {
		err = fmt.Errorf("request prometheus fail with bad status:%d ", resp.StatusCode)
		return
	}
	var data m.PrometheusResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		err = fmt.Errorf("unmarshal prometheus response data fail,%s ", err.Error())
		return
	}
	if data.Status != "success" {
		err = fmt.Errorf("prometheus response with error status:%s ", data.Status)
		return
	}
	for _, otr := range data.Data.Result {
		tmpMapSort := m.PromMapSort{}
		for k, v := range otr.Metric {
			if k == "job" || k == "instance" || k == "e_guid" {
				continue
			}
			tmpMapSort = append(tmpMapSort, &m.SimpleMapObj{Key: k, Value: v})
		}
		metricList = append(metricList, tmpMapSort.String())
	}
	return
}

// QueryLogKeywordData keywordMode -> log | db
func QueryLogKeywordData(keywordMode string) (result map[string]float64, err error) {
	result = make(map[string]float64)
	queryQl := "node_log_monitor_count_total"
	if keywordMode == "db" {
		queryQl = "db_keyword_value"
	}
	nowTime := time.Now().Unix()
	queryResult, queryErr := QueryPrometheusRange(queryQl, nowTime-10, nowTime, 10)
	if queryErr != nil {
		err = queryErr
		return
	}
	if keywordMode == "db" {
		for _, otr := range queryResult.Result {
			key := fmt.Sprintf("service_group:%s^db_keyword_guid:%s^t_endpoint:%s", otr.Metric["service_group"], otr.Metric["db_keyword_guid"], otr.Metric["t_endpoint"])
			tmpValue := float64(0)
			if len(otr.Values) > 0 {
				tmpValue, _ = strconv.ParseFloat(otr.Values[len(otr.Values)-1][1].(string), 64)
			}
			result[key] = tmpValue
		}
		return
	}
	for _, otr := range queryResult.Result {
		key := fmt.Sprintf("e_guid:%s^t_guid:%s^file:%s^keyword:%s", otr.Metric["e_guid"], otr.Metric["t_guid"], otr.Metric["file"], otr.Metric["keyword"])
		tmpValue := float64(0)
		if len(otr.Values) > 0 {
			tmpValue, _ = strconv.ParseFloat(otr.Values[len(otr.Values)-1][1].(string), 64)
		}
		result[key] = tmpValue
	}
	return
}

func QueryPromSeries(promQL string) (result []map[string]string, err error) {
	//if strings.Contains(promQL, "$") {
	//	re, _ := regexp.Compile("=\"[\\$]+[^\"]+\"")
	//	fetchTag := re.FindAll([]byte(promQL), -1)
	//	for _, vv := range fetchTag {
	//		promQL = strings.Replace(promQL, string(vv), "=~\".*\"", -1)
	//	}
	//}
	promQL = getPromQlMainExpr(promQL)
	requestUrl, urlParseErr := url.Parse(fmt.Sprintf("http://%s/api/v1/series", promDS.Host))
	if urlParseErr != nil {
		return result, fmt.Errorf("Url parse fail,%s ", urlParseErr.Error())
	}
	urlParams := url.Values{}
	urlParams.Set("match[]", promQL)
	requestUrl.RawQuery = urlParams.Encode()
	req, _ := http.NewRequest(http.MethodGet, requestUrl.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	httpClient, getClientErr := promDS.DataSource.GetHttpClient()
	if getClientErr != nil {
		return result, fmt.Errorf("Get httpClient fail,%s ", getClientErr.Error())
	}
	res, reqErr := ctxhttp.Do(context.Background(), httpClient, req)
	if reqErr != nil {
		return result, fmt.Errorf("http do request fail,%s ", reqErr.Error())
	}
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode/100 != 2 {
		return result, fmt.Errorf("Request fail with bad status:%d ", res.StatusCode)
	}
	var data m.PromSeriesResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return result, fmt.Errorf("Json unmarshal response fail,%s ", err.Error())
	}
	if data.Status != "success" {
		return result, fmt.Errorf("Query prometheus data fail,status:%s ", data.Status)
	}
	result = data.Data
	return
}

func getPromQlMainExpr(input string) (output string) {
	if rightIndex := strings.Index(input, "}"); rightIndex > 0 {
		input = input[:rightIndex+1]
		if leftIndex := strings.LastIndex(input, "("); leftIndex > 0 {
			input = input[leftIndex+1:]
		}
		output = input
	} else {
		output = input
	}
	return
}

// QueryPrometheusRange start/end/step second value
func QueryPrometheusRange(promQL string, start, end, step int64) (result *m.PrometheusData, err error) {
	requestUrl, urlParseErr := url.Parse(fmt.Sprintf("http://%s/api/v1/query_range", promDS.Host))
	if urlParseErr != nil {
		return result, fmt.Errorf("Url parse fail,%s ", urlParseErr.Error())
	}
	urlParams := url.Values{}
	urlParams.Set("start", strconv.FormatInt(start, 10))
	urlParams.Set("end", strconv.FormatInt(end, 10))
	urlParams.Set("step", strconv.FormatInt(step, 10))
	urlParams.Set("query", promQL)
	requestUrl.RawQuery = urlParams.Encode()
	req, _ := http.NewRequest(http.MethodGet, requestUrl.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	httpClient, getClientErr := promDS.DataSource.GetHttpClient()
	if getClientErr != nil {
		return result, fmt.Errorf("Get httpClient fail,%s ", getClientErr.Error())
	}
	res, reqErr := ctxhttp.Do(context.Background(), httpClient, req)
	if reqErr != nil {
		return result, fmt.Errorf("http do request fail,%s ", reqErr.Error())
	}
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode/100 != 2 {
		return result, fmt.Errorf("Request fail with bad status:%d ", res.StatusCode)
	}
	var data m.PrometheusResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return result, fmt.Errorf("Json unmarshal response fail,%s ", err.Error())
	}
	if data.Status != "success" {
		return result, fmt.Errorf("Query prometheus data fail,status:%s ", data.Status)
	}
	result = &data.Data
	return
}

// ResetPrometheusMetricMap 重置 Prometheus返回的metric
func ResetPrometheusMetricMap(tagMap map[string]string) map[string]string {
	// 此处查询指标 对应的业务配置,如果是自定义业务配置, tags内容: tags="test_service_code=addUser,test_retcode=200",需要做特殊解析处理
	if tagMap["tags"] != "" {
		strArr := strings.Split(tagMap["tags"], ",")
		for _, str := range strArr {
			if index := strings.Index(str, "="); index > 0 {
				tagMap[str[:index]] = str[index+1:]
			}
		}
	}
	return tagMap
}
