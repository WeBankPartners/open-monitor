package datasource

import (
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
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

func PrometheusData(query m.QueryMonitorData) []*m.SerialModel  {
	serials := []*m.SerialModel{}
	urlParams := url.Values{}
	requestUrl,err := url.Parse(fmt.Sprintf("http://%s/api/v1/query_range", promDS.Host))
	if err!=nil {
		mid.LogError("make url fail", err)
		return serials
	}
	var tmpStep int64
	tmpStep = 10
	subSec := query.End - query.Start
	if subSec > 100000 {
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
	for _,otr := range data.Data.Result {
		var serial m.SerialModel
		serial.Type = "line"
		tmpName := query.Legend
		for k,v := range otr.Metric {
			if strings.Contains(query.Legend, "$"+k) {
				tmpName = strings.Replace(tmpName, "$"+k, v, -1)
			}
		}
		if strings.Contains(query.Legend, "$custom") {
			if query.Legend == "$custom" {
				tmpName = fmt.Sprintf("%s:%s", query.Endpoint[0], query.Metric[0])
				if len(data.Data.Result) > 1 {
					tmpName = appendTagString(tmpName, otr.Metric)
				}
			}else if query.Legend == "$custom_metric" {
				tmpName = query.Metric[0]
				if len(data.Data.Result) > 1 {
					tmpName = appendTagString(tmpName, otr.Metric)
				}
			}else if query.Legend == "$custom_endpoint" {
				tmpName = query.Endpoint[0]
				if len(data.Data.Result) > 1 {
					tmpName = appendTagString(tmpName, otr.Metric)
				}
			}
		}
		serial.Name = tmpName
		var sdata m.DataSort
		for _,v := range otr.Values {
			tmpTime := v[0].(float64) * 1000
			tmpValue,_ := strconv.ParseFloat(v[1].(string), 64)
			sdata = append(sdata, []float64{tmpTime, tmpValue})
		}
		sort.Sort(sdata)
		serial.Data = sdata
		serials = append(serials, &serial)
	}
	return serials
}

func appendTagString(name string, metricMap map[string]string) string {
	tmpName := name + "{"
	for k,v := range metricMap {
		if k == "job" && v == "consul" {
			continue
		}
		tmpName += fmt.Sprintf("%s=%s,", k, v)
	}
	tmpName = tmpName[:len(tmpName)-1]
	tmpName += "}"
	return tmpName
}