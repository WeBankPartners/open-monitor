package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-agent/metric_comparison/models"
	"github.com/WeBankPartners/open-monitor/monitor-agent/metric_comparison/rpc"
)

var (
	metricComparisonHttpLock   = new(sync.RWMutex)
	metricComparisonResultLock = new(sync.RWMutex)
	metricComparisonRes        []*models.MetricComparisonRes
	metricComparisonList       []*models.MetricComparisonDto
	exposeCycleMap             = make(map[int]bool) //暴露周期map
)

const (
	metricComparisonFilePath = "config/metric_comparison_cache.json"
)

// HandlePrometheus 封装数据给Prometheus采集
func HandlePrometheus(w http.ResponseWriter, r *http.Request) {
	var buff bytes.Buffer
	var i int
	metricComparisonResultLock.RLock()
	for _, v := range metricComparisonRes {
		// 简单校验同环比暴露指标名称合法性
		if !checkPrometheusQL(v.Name) {
			log.Printf("Prometheus %s is invalid\n", v.Name)
			continue
		}
		if len(v.MetricMap) > 0 {
			buff.WriteString(fmt.Sprintf("%s{", v.Name))
			i = 0
			for key, value := range v.MetricMap {
				if i < len(v.MetricMap)-1 {
					buff.WriteString(fmt.Sprintf("%s=\"%s\",", key, value))
				} else {
					buff.WriteString(fmt.Sprintf("%s=\"%s\"} %0.2f \n", key, value, v.Value))
				}
				i++
			}
		}
	}
	metricComparisonResultLock.RUnlock()
	if buff.Len() > 0 {
		log.Printf("%s\n", buff.Bytes())
	}
	w.Write(buff.Bytes())
}

// checkPrometheusQL 简单校验同环比暴露指标名称合法性
func checkPrometheusQL(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	// 不能以数字开头,不能出现点号和中文
	regex := regexp.MustCompile(`^\d`)
	// 判断是否以数字开头
	if regex.MatchString(name) {
		return false
	}
	if strings.Contains(name, ".") {
		return false
	}
	if containsChinese(name) {
		return false
	}
	return true
}

func containsChinese(s string) bool {
	return regexp.MustCompile("[\u4e00-\u9fa5]+").MatchString(s)
}

func StartCalcMetricComparisonCron() {
	if len(metricComparisonList) == 0 {
		LoadMetricComparisonConfig()
	}
	now := time.Now()
	// 计算到下一分钟整点的时间差
	nextMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	timeDiff := nextMinute.Sub(now)
	time.Sleep(time.Duration(timeDiff.Seconds()+1) * time.Second)
	// 整点1分钟执行
	t := time.NewTicker(time.Minute).C
	for {
		<-t
		newNow := time.Now()
		exposeCycleMap = make(map[int]bool)
		exposeCycleMap[1] = true
		if newNow.Minute() == 0 {
			exposeCycleMap[60] = true
			exposeCycleMap[30] = true
		} else if newNow.Minute() == 30 {
			exposeCycleMap[30] = true
		}
		if newNow.Minute()%10 == 0 {
			exposeCycleMap[5] = true
			exposeCycleMap[10] = true
		} else if newNow.Minute()%5 == 0 {
			exposeCycleMap[5] = true
		}
		log.Printf("exposeCycleMap:%v", exposeCycleMap)
		go calcMetricComparisonData(exposeCycleMap)
	}
}

// ReceiveMetricComparisonData 接受同环比数据
func ReceiveMetricComparisonData(w http.ResponseWriter, r *http.Request) {
	log.Println("start receive metric comparison data!")
	var err error
	var requestParamBuff []byte
	var response models.Response
	metricComparisonHttpLock.Lock()
	defer func(retErr error) {
		metricComparisonHttpLock.Unlock()
		response = models.Response{Status: "OK", Message: "success"}
		if retErr != nil {
			response = models.Response{Status: "ERROR", Message: retErr.Error()}
		}
		b, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}(err)
	if requestParamBuff, err = ioutil.ReadAll(r.Body); err != nil {
		return
	}
	if err = json.Unmarshal(requestParamBuff, &metricComparisonList); err != nil {
		log.Printf("json Unmarshal err:%+v\n", err)
		return
	}
	if err = MetricComparisonSaveConfig(requestParamBuff); err != nil {
		log.Printf("metricComparison config err:%+v\n", err)
		return
	} else {
		log.Println("metricComparison config save success!")
	}
}

func MetricComparisonSaveConfig(requestParamBuff []byte) (err error) {
	err = ioutil.WriteFile(metricComparisonFilePath, requestParamBuff, 0644)
	return
}

func LoadMetricComparisonConfig() {
	if requestParamBuff, err := ioutil.ReadFile(metricComparisonFilePath); err == nil {
		json.Unmarshal(requestParamBuff, &metricComparisonList)
	} else {
		log.Printf("read metric_comparison_cache.json err:%+v", err)
	}
}

func calcMetricComparisonData(exposeCycleMap map[int]bool) {
	var curResultList, historyResultList []*models.PrometheusQueryObj
	var err error
	var historyEnd int64
	var tempMetricComparisonList []*models.MetricComparisonRes
	metricComparisonHttpLock.RLock()
	defer metricComparisonHttpLock.RUnlock()
	if len(metricComparisonList) == 0 {
		return
	}
	// 根据数据查询原始指标数据
	for _, metricComparison := range metricComparisonList {
		// 需要整点计算数据并且暴露,没到对应整点时间不参与本次计算
		if !exposeCycleMap[metricComparison.CalcPeriod/60] {
			continue
		}
		if metricComparison.OriginPromExpr == "" {
			continue
		}
		now := time.Now()
		calcTypeMap := getCalcTypeMap(metricComparison.CalcType)
		curResultList = []*models.PrometheusQueryObj{}
		historyResultList = []*models.PrometheusQueryObj{}
		// 查询范围,先扩大查询范围,然后过滤 时间范围精准控制,保证是整点分钟
		if curResultList, err = QueryPrometheusData(&models.PrometheusQueryParam{
			Start:  now.Unix() - 2*int64(metricComparison.CalcPeriod) - 1,
			End:    now.Unix(),
			PromQl: parsePromQL(metricComparison.OriginPromExpr),
		}); err != nil {
			log.Printf("prometheus query_range err:%+v", err)
			continue
		}
		// 根据数据计算 同环比
		switch metricComparison.ComparisonType {
		case "day":
			historyEnd = now.Unix() - 86400
		case "week":
			historyEnd = now.Unix() - 86400*7
		case "month":
			historyEnd = now.AddDate(0, -1, 0).Unix()
		}
		// 查询对比历史数据
		// 查询范围,历史数据 先扩大查询范围,然后过滤 时间范围精准控制,保证是整点分钟
		if historyResultList, err = QueryPrometheusData(&models.PrometheusQueryParam{
			Start:  historyEnd - 2*int64(metricComparison.CalcPeriod) - 1,
			End:    historyEnd + 60,
			PromQl: parsePromQL(metricComparison.OriginPromExpr),
		}); err != nil {
			log.Printf("prometheus query_range err:%+v\n", err)
			continue
		}
		if len(curResultList) == 0 || len(historyResultList) == 0 {
			if len(curResultList) == 0 {
				log.Printf("%s current prometheus query data empty\n", metricComparison.Metric)
			} else {
				log.Printf("%s history prometheus query data empty\n", metricComparison.Metric)
			}
			continue
		}
		// 数据过滤,过滤整点时间范围内数据
		curResultList = filterData(curResultList, now.Unix(), int64(metricComparison.CalcPeriod))
		historyResultList = filterData(historyResultList, historyEnd, int64(metricComparison.CalcPeriod))
		for _, data := range curResultList {
			for _, historyData := range historyResultList {
				if len(data.Metric) > 0 && len(historyData.Metric) > 0 && data.Metric.ToTagString() == historyData.Metric.ToTagString() {
					if len(data.Values) == 0 || len(historyData.Values) == 0 {
						break
					}
					var dataVal, historyDataVal float64
					metricComparisonRes1 := &models.MetricComparisonRes{
						MetricMap:  make(map[string]string),
						Name:       metricComparison.Metric,
						CalcPeriod: metricComparison.CalcPeriod,
					}
					metricComparisonRes2 := &models.MetricComparisonRes{
						MetricMap:  make(map[string]string),
						Name:       metricComparison.Metric,
						CalcPeriod: metricComparison.CalcPeriod,
					}
					for _, metricObj := range data.Metric {
						if metricObj.Key == "__name__" {
							continue
						}
						metricComparisonRes1.MetricMap[metricObj.Key] = metricObj.Value
						metricComparisonRes2.MetricMap[metricObj.Key] = metricObj.Value
					}
					switch metricComparison.CalcMethod {
					case "avg":
						var sum1, sum2 float64
						for _, arr := range data.Values {
							if len(arr) == 2 {
								sum1 = sum1 + arr[1]
							}
						}
						for _, arr := range historyData.Values {
							if len(arr) == 2 {
								sum2 = sum2 + arr[1]
							}
						}
						dataVal = sum1 / float64(len(data.Values))
						historyDataVal = sum2 / float64(len(historyData.Values))
					case "sum":
						for _, arr := range data.Values {
							if len(arr) == 2 {
								dataVal = dataVal + arr[1]
							}
						}
						for _, arr := range historyData.Values {
							if len(arr) == 2 {
								historyDataVal = historyDataVal + arr[1]
							}
						}
					case "max":
						for i, arr := range data.Values {
							// 初始化值
							if i == 0 && len(arr) == 2 {
								dataVal = arr[1]
								continue
							}
							if len(arr) == 2 && dataVal < arr[1] {
								dataVal = arr[1]
							}
						}
						for i, arr := range historyData.Values {
							// 初始化值
							if i == 0 && len(arr) == 2 {
								historyDataVal = arr[1]
								continue
							}
							if len(arr) == 2 && historyDataVal < arr[1] {
								historyDataVal = arr[1]
							}
						}
					case "min":
						for i, arr := range data.Values {
							// 初始化值
							if i == 0 && len(arr) == 2 {
								dataVal = arr[1]
								continue
							}
							if len(arr) == 2 && dataVal > arr[1] {
								dataVal = arr[1]
							}
						}
						for i, arr := range historyData.Values {
							// 初始化值
							if i == 0 && len(arr) == 2 {
								historyDataVal = arr[1]
								continue
							}
							if len(arr) == 2 && historyDataVal > arr[1] {
								historyDataVal = arr[1]
							}
						}
					}
					if calcTypeMap["diff"] {
						metricComparisonRes1.MetricMap["calc_type"] = "diff"
						metricComparisonRes1.Value = dataVal - historyDataVal
						tempMetricComparisonList = append(tempMetricComparisonList, metricComparisonRes1)
					}
					if calcTypeMap["diff_percent"] {
						metricComparisonRes2.MetricMap["calc_type"] = "diff_percent"
						if historyDataVal != 0 {
							metricComparisonRes2.Value = (dataVal - historyDataVal) * 100 / historyDataVal
						}
						tempMetricComparisonList = append(tempMetricComparisonList, metricComparisonRes2)
					}
					break
				}
			}
		}
	}
	// 写数据
	metricComparisonResultLock.Lock()
	metricComparisonRes = tempMetricComparisonList
	metricComparisonResultLock.Unlock()
	// 数据有效期10s钟,保证数据只被采集到一次,10s后清空数据
	time.Sleep(10 * time.Second)
	metricComparisonResultLock.Lock()
	metricComparisonRes = []*models.MetricComparisonRes{}
	metricComparisonResultLock.Unlock()
}

func parsePromQL(promQl string) string {
	if strings.Contains(promQl, "$") {
		re, _ := regexp.Compile("=\"[\\$]+[^\"]+\"")
		fetchTag := re.FindAll([]byte(promQl), -1)
		for _, vv := range fetchTag {
			promQl = strings.Replace(promQl, string(vv), "=~\".*\"", -1)
		}
	}
	return promQl
}

func getCalcTypeMap(calcType string) map[string]bool {
	hashMap := make(map[string]bool)
	if strings.TrimSpace(calcType) != "" {
		arr := strings.Split(calcType, ",")
		for _, s := range arr {
			hashMap[s] = true
		}
	}
	return hashMap
}

func QueryPrometheusData(param *models.PrometheusQueryParam) (resultList []*models.PrometheusQueryObj, err error) {
	var result models.PrometheusResponse
	var resByteArr []byte
	resultList = []*models.PrometheusQueryObj{}
	requestUrl, _ := url.Parse("http://127.0.0.1:9090/api/v1/query_range")
	//requestUrl, _ := url.Parse("http://106.52.160.142:9090/api/v1/query_range")
	urlParams := url.Values{}
	urlParams.Set("start", fmt.Sprintf("%d", param.Start))
	urlParams.Set("end", fmt.Sprintf("%d", param.End))
	urlParams.Set("step", "10")
	urlParams.Set("query", param.PromQl)
	requestUrl.RawQuery = urlParams.Encode()
	if param == nil || strings.TrimSpace(param.PromQl) == "" {
		err = fmt.Errorf("promQl is empty")
		return
	}
	if resByteArr, err = rpc.HttpGet(requestUrl.String()); err != nil {
		return
	}
	if err = json.Unmarshal(resByteArr, &result); err != nil {
		return
	}
	if result.Status != "success" {
		log.Printf("param:%s,%+v\n", requestUrl.String(), string(resByteArr))
		err = fmt.Errorf("prometheus response status=%s \n", result.Status)
		return
	}
	if len(result.Data.Result) > 0 {
		for _, v := range result.Data.Result {
			tmpResultObj := models.PrometheusQueryObj{Start: param.Start, End: param.End}
			var tmpValues [][]float64
			var tmpTagSortList models.DefaultSortList
			for kk, vv := range v.Metric {
				tmpTagSortList = append(tmpTagSortList, &models.DefaultSortObj{Key: kk, Value: vv})
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
					continue
				}

				// 安全解析数值
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Panic in metric comparison parseFloat: value=%s, panic=%v\n", tmpValueStr, r)
					}
				}()

				var err error
				tmpV, err = strconv.ParseFloat(tmpValueStr, 64)
				if err != nil {
					continue
				}

				tmpValues = append(tmpValues, []float64{vv[0].(float64), tmpV})
			}
			tmpResultObj.Values = tmpValues
			resultList = append(resultList, &tmpResultObj)
		}
	}
	return
}

func filterData(resultList []*models.PrometheusQueryObj, timestamp, calcPeriod int64) []*models.PrometheusQueryObj {
	var newResultList []*models.PrometheusQueryObj
	var timestampStart, timestampEnd int64
	switch calcPeriod {
	case 60:
		// 将整点分钟的时间转换回Unix时间戳
		timestampEnd = time.Unix(timestamp, 0).Truncate(time.Minute).Unix()
	default:
		// 转成 最近N分钟倍数
		// 获取当前的分钟数
		t := time.Unix(timestamp, 0)
		currentMinute := t.Minute()
		// 计算离当前时间最近的N分钟倍数的分钟数
		adjustedMinute := currentMinute - (currentMinute % int(calcPeriod/60))
		// 创建一个新的时间点，使得分钟数为调整后的值
		adjustedTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), adjustedMinute, 0, 0, t.Location())
		timestampEnd = adjustedTime.Unix()
	}
	timestampStart = timestampEnd - calcPeriod
	for _, obj := range resultList {
		newObj := &models.PrometheusQueryObj{
			Start:  timestampStart,
			End:    timestampEnd,
			Metric: obj.Metric,
			Values: [][]float64{},
		}
		if len(obj.Values) > 0 {
			for _, valArr := range obj.Values {
				if len(valArr) == 2 && valArr[0] >= float64(timestampStart) && valArr[0] <= float64(timestampEnd) {
					newObj.Values = append(newObj.Values, valArr)
				}
			}
		}
		if len(newObj.Values) > 0 {
			newResultList = append(newResultList, newObj)
		}
	}
	return newResultList
}
