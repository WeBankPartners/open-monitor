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
	gTransport         *http.Transport
	prometheusAddress  string
	prometheusQueryUrl string
	queryStep          int
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
		MaxConnsPerHost:     maxOpen,
		MaxIdleConnsPerHost: maxIdle,
		IdleConnTimeout:     time.Duration(timeout) * time.Second,
	}
	return nil
}

func getPrometheusData(param *PrometheusQueryParam) error {
	requestUrl, _ := url.Parse(prometheusQueryUrl)
	urlParams := url.Values{}
	urlParams.Set("start", fmt.Sprintf("%d", param.Start))
	urlParams.Set("end", fmt.Sprintf("%d", param.End))
	urlParams.Set("step", fmt.Sprintf("%d", queryStep))
	urlParams.Set("query", param.PromQl)
	requestUrl.RawQuery = urlParams.Encode()
	req, _ := http.NewRequest(http.MethodGet, requestUrl.String(), strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	resp, err := gTransport.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("http request error, %v \n", err)
	}
	if resp.StatusCode/100 != 2 {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return fmt.Errorf("http response status %d \n", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("http read body error, %v \n", err)
	}
	var result PrometheusResponse
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return fmt.Errorf("http body json unmarshal error, %v \n", err)
	}
	if result.Status != "success" {
		return fmt.Errorf("prometheus response status=%s \n", result.Status)
	}
	for _, v := range result.Data.Result {
		tmpResultObj := PrometheusQueryObj{Start: param.Start, End: param.End}
		var tmpValues [][]float64
		var tmpTagSortList DefaultSortList
		for kk, vv := range v.Metric {
			if !isIgnoreTag(kk) {
				tmpTagSortList = append(tmpTagSortList, &DefaultSortObj{Key: kk, Value: vv})
			}
		}
		sort.Sort(tmpTagSortList)
		tmpResultObj.Metric = tmpTagSortList
		for _, vv := range v.Values {
			tmpValueStr := vv[1].(string)

			// 处理 NaN 值 - 添加异常过滤
			var tmpV float64

			// 安全检查：确保输入参数不为空
			if tmpValueStr == "" {
				continue
			}

			// 快速检查 NaN 和 inf 值
			if (len(tmpValueStr) == 3 && (tmpValueStr == "NaN" || tmpValueStr == "nan" || tmpValueStr == "inf")) ||
				(len(tmpValueStr) == 4 && tmpValueStr == "-inf") {
				// 对于成功率指标，NaN/inf 时返回 100（避免告警）
				if strings.HasSuffix(param.Metric, "req_suc_rate") {
					tmpV = 100.0
				} else {
					continue
				}
			} else {
				// 安全解析数值
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Panic in archive prometheus parseFloat: value=%s, panic=%v\n", tmpValueStr, r)
					}
				}()

				var err error
				tmpV, err = strconv.ParseFloat(tmpValueStr, 64)
				if err != nil {
					continue
				}
			}

			tmpValues = append(tmpValues, []float64{vv[0].(float64), tmpV})
		}
		tmpResultObj.Values = tmpValues
		param.Data = append(param.Data, &tmpResultObj)
	}
	return nil
}

func isIgnoreTag(tag string) bool {
	flag := false
	for _, v := range Config().Prometheus.IgnoreTags {
		if v == tag {
			flag = true
			break
		}
	}
	return flag
}
