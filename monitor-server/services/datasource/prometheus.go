package datasource

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"net/http"
	"strings"
	"golang.org/x/net/context/ctxhttp"
	"context"
	"io/ioutil"
	"encoding/json"
	"time"
	"strconv"
	"sort"
	"net/url"
)

var promDS DataSourceParam

func InitPrometheusDatasource()  {
	t := time.Now()
	cfg := *m.Config().Datasource.Servers[0]
	opentsdbDS := &DataSource{Id:cfg.Id,Name:cfg.Type,Url:fmt.Sprintf("http://%s", cfg.Host),IsDefault:true,Updated:t}
	promDS = DataSourceParam{DataSource:opentsdbDS, Host:cfg.Host, Token:cfg.Token}
}

var PieLegendBlackName = []string{"job", "instance"}

func PrometheusData(query *m.QueryMonitorData) []*m.SerialModel  {
	serials := []*m.SerialModel{}
	urlParams := url.Values{}
	requestUrl,err := url.Parse(fmt.Sprintf("http://%s/api/v1/query_range", promDS.Host))
	if err!=nil {
		mid.LogError("make url fail", err)
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
		mid.LogError("Failed to create request", err)
		return serials
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient,err := promDS.DataSource.GetHttpClient()
	if err != nil {
		mid.LogError("get httpClient fail", err)
		return serials
	}
	res, err := ctxhttp.Do(context.Background(), httpClient, req)
	if err != nil {
		mid.LogError("http request fail", err)
		return serials
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		mid.LogError("http request body read fail", err)
		return serials
	}
	if res.StatusCode/100 != 2 {
		mid.LogError(fmt.Sprintf("request status : %v", res.Status), nil)
		return serials
	}
	var data m.PrometheusResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		mid.LogError("unmarshal response fail", err)
		return serials
	}
	if data.Status != "success" {
		mid.LogError(fmt.Sprintf("query prometheus data fail : %s", data.Status), nil)
		return serials
	}
	if query.ChartType == "pie" {
		var pieData m.EChartPie
		for _,otr := range data.Data.Result {
			var tmpNameList []string
			for k,v := range otr.Metric {
				isBlack := false
				for _,vv := range PieLegendBlackName {
					if k == vv {
						isBlack = true
						break
					}
				}
				if isBlack {
					continue
				}
				tmpNameList = append(tmpNameList, fmt.Sprintf("%s=%s", k, v))
			}
			tmpName := strings.Join(tmpNameList, ",")
			if tmpName == "" {
				tmpName = query.Metric[0]
			}
			pieData.Legend = append(pieData.Legend, tmpName)
			if len(otr.Values) > 0 {
				tmpValue,_ := strconv.ParseFloat(otr.Values[len(otr.Values)-1][1].(string), 64)
				tmpValue,_ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
				pieData.Data = append(pieData.Data, &m.EChartPieObj{Name:tmpName, Value:tmpValue})
			}
		}
		query.PieData = pieData
		return serials
	}
	for _,otr := range data.Data.Result {
		var serial m.SerialModel
		serial.Type = "line"
		serial.Name = GetSerialName(query, otr.Metric, len(data.Data.Result))
		var sdata m.DataSort
		for _,v := range otr.Values {
			tmpTime := v[0].(float64) * 1000
			tmpValue,_ := strconv.ParseFloat(v[1].(string), 64)
			//tmpValue,_ = strconv.ParseFloat(fmt.Sprintf("%.3f", tmpValue), 64)
			sdata = append(sdata, []float64{tmpTime, tmpValue})
		}
		sort.Sort(sdata)
		serial.Data = sdata
		serials = append(serials, &serial)
	}
	return serials
}

func appendTagString(name string, metricMap map[string]string) string {
	var tmpList m.DefaultSortList
	for k,v := range metricMap {
		tmpList = append(tmpList, &m.DefaultSortObj{Key:k, Value:v})
	}
	sort.Sort(tmpList)
	tmpName := name + "{"
	for _,v := range tmpList {
		if v.Key == "job" && v.Value == "consul" {
			continue
		}
		tmpName += fmt.Sprintf("%s=%s,", v.Key, v.Value)
	}
	tmpName = tmpName[:len(tmpName)-1]
	tmpName += "}"
	return tmpName
}

func GetSerialName(query *m.QueryMonitorData,tagMap map[string]string,dataLength int) string {
	tmpName := query.Legend
	legend := query.Legend
	endpoint := query.Endpoint[0]
	metric := query.Metric[0]
	for k,v := range tagMap {
		if strings.Contains(legend, "$"+k) {
			tmpName = strings.Replace(tmpName, "$"+k, v, -1)
			if !query.SameEndpoint {
				tmpName = fmt.Sprintf("%s:%s", endpoint, tmpName)
			}
		}
	}
	if strings.Contains(legend, "$custom") {
		if legend == "$custom" {
			tmpName = fmt.Sprintf("%s:%s", endpoint, metric)
			if dataLength > 1 {
				tmpName = appendTagString(tmpName, tagMap)
			}
		}else if legend == "$custom_metric" {
			tmpName = metric
			if dataLength > 1 {
				tmpName = appendTagString(tmpName, tagMap)
			}
		}else if legend == "$custom_endpoint" {
			tmpName = endpoint
			if dataLength > 1 {
				tmpName = appendTagString(tmpName, tagMap)
			}
		}else if legend == "$custom_all" {
			tmpName = fmt.Sprintf("%s:%s", endpoint, metric)
			tmpName = appendTagString(tmpName, tagMap)
		}
	}
	if legend == "$metric" || legend == "$custom_metric" {
		if !strings.Contains(tmpName, ":") && !query.SameEndpoint {
			tmpName = fmt.Sprintf("%s:%s", endpoint, tmpName)
		}
	}
	if query.CompareLegend != "" {
		tmpName = fmt.Sprintf("%s_%s", query.CompareLegend, tmpName)
	}
	return tmpName
}