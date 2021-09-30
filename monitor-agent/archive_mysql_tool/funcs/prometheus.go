package funcs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	gTransport *http.Transport
	prometheusAddress string
	prometheusQueryUrl string
	queryStep int
)

func InitHttpTransport() error {
	if Config().Prometheus.Server == "" || Config().Prometheus.Port == 0 {
		return fmt.Errorf("init prometheus http config fail,address illegal ")
	}
	prometheusAddress = fmt.Sprintf("%s:%d", Config().Prometheus.Server, Config().Prometheus.Port)
	prometheusQueryUrl = fmt.Sprintf("http://%s/api/v1/query_range", prometheusAddress)
	maxOpen := 100
	maxIdle := 10
	timeout := 60
	queryStep = 10
	if Config().Prometheus.QueryStep > 0 {
		queryStep = Config().Prometheus.QueryStep
	}
	if Config().Prometheus.MaxHttpOpen > 0 {
		maxOpen = Config().Prometheus.MaxHttpOpen
	}
	if Config().Prometheus.MaxHttpIdle > 0 {
		maxIdle = Config().Prometheus.MaxHttpIdle
	}
	if Config().Prometheus.HttpIdleTimeout > 0 {
		timeout = Config().Prometheus.HttpIdleTimeout
	}
	if maxIdle > maxOpen {
		return fmt.Errorf("init prometheus http config fail,idle > open? ")
	}
	gTransport = &http.Transport{
		MaxConnsPerHost: maxOpen,
		MaxIdleConnsPerHost: maxIdle,
		IdleConnTimeout: time.Duration(timeout)*time.Second,
	}
	return nil
}

func getPrometheusData(param *PrometheusQueryParam) error {
	requestUrl,_ := url.Parse(prometheusQueryUrl)
	urlParams := url.Values{}
	urlParams.Set("start", fmt.Sprintf("%d", param.Start))
	urlParams.Set("end", fmt.Sprintf("%d", param.End))
	urlParams.Set("step", fmt.Sprintf("%d", queryStep))
	urlParams.Set("query", param.PromQl)
	requestUrl.RawQuery = urlParams.Encode()
	req,_ := http.NewRequest(http.MethodGet, requestUrl.String(), strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	resp,err := gTransport.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("http request error, %v \n", err)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("http response status %d \n", resp.StatusCode)
	}
	bodyBytes,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http read body error, %v \n", err)
	}
	resp.Body.Close()
	var result PrometheusResponse
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return fmt.Errorf("http body json unmarshal error, %v \n", err)
	}
	if result.Status != "success" {
		return fmt.Errorf("prometheus response status=%s \n", result.Status)
	}
	for _,v := range result.Data.Result {
		tmpResultObj := PrometheusQueryObj{Start:param.Start, End:param.End}
		var tmpValues [][]float64
		var tmpTagSortList DefaultSortList
		for kk,vv := range v.Metric {
			if !isIgnoreTag(kk) {
				tmpTagSortList = append(tmpTagSortList, &DefaultSortObj{Key:kk, Value:vv})
			}
		}
		sort.Sort(tmpTagSortList)
		tmpResultObj.Metric = tmpTagSortList
		for _,vv := range v.Values {
			tmpV,_ := strconv.ParseFloat(vv[1].(string), 64)
			tmpValues = append(tmpValues, []float64{vv[0].(float64), tmpV})
		}
		tmpResultObj.Values = tmpValues
		param.Data = append(param.Data, &tmpResultObj)
	}
	return nil
}

func isIgnoreTag(tag string) bool {
	flag := false
	for _,v := range Config().Prometheus.IgnoreTags {
		if v == tag {
			flag = true
			break
		}
	}
	return flag
}