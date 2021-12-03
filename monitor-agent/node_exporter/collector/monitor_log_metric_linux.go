package collector

import (
	"encoding/json"
	"fmt"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	log_metricCollectorName   = "log_metric_monitor"
	log_metricMonitorFilePath = "data/log_metric_monitor_cache.data"
)

var (
	logMetricMonitorJobs       []*logMetricMonitorNeObj
	logMetricHttpLock          = new(sync.RWMutex)
	logMetricMonitorMetrics    []*logMetricDisplayObj
	logMetricMonitorMetricLock = new(sync.RWMutex)
	monitorLogger              log.Logger
)

type logMetricMonitorCollector struct {
	logMetricMonitor *prometheus.Desc
	logger           log.Logger
}

func InitMonitorLogger(logger log.Logger) {
	monitorLogger = logger
}

func init() {
	registerCollector(log_metricCollectorName, defaultEnabled, LogMetricMonitorCollector)
}

func LogMetricMonitorCollector(logger log.Logger) (Collector, error) {
	return &logMetricMonitorCollector{
		logMetricMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, log_metricCollectorName, "value"),
			"Show log_metric data from log file.",
			[]string{"key", "tags", "path", "agg", "t_endpoint"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *logMetricMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	logMetricMonitorMetricLock.RLock()
	for _, v := range logMetricMonitorMetrics {
		if !v.Display {
			continue
		}
		ch <- prometheus.MustNewConstMetric(c.logMetricMonitor,
			prometheus.GaugeValue,
			v.Value, v.Metric, v.TagsString, v.Path, v.Agg, v.TEndpoint)
	}
	logMetricMonitorMetricLock.RUnlock()
	return nil
}

type logMetricNodeExporterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type logMetricMonitorNeObj struct {
	TailSession    *tail.Tail            `json:"-"`
	Lock           *sync.RWMutex         `json:"-"`
	Path           string                `json:"path"`
	TargetEndpoint string                `json:"target_endpoint"`
	JsonConfig     []*logMetricJsonNeObj `json:"config"`
	MetricConfig   []*logMetricNeObj     `json:"custom"`
}

type logMetricJsonNeObj struct {
	Regexp       *regexp.Regexp              `json:"-"`
	DataChannel  chan map[string]interface{} `json:"-"`
	Regular      string                      `json:"regular"`
	Tags         string                      `json:"tags"`
	MetricConfig []*logMetricNeObj           `json:"metric_config"`
}

type logMetricNeObj struct {
	RegExp       *regexp.Regexp             `json:"-"`
	DataChannel  chan string                `json:"-"`
	Key          string                     `json:"key"`
	Metric       string                     `json:"metric"`
	ValueRegular string                     `json:"value_regular"`
	Title        string                     `json:"title"`
	AggType      string                     `json:"agg_type"`
	Step         int64                      `json:"step"`
	StringMap    []*logMetricStringMapNeObj `json:"string_map"`
}

type logMetricStringMapNeObj struct {
	Regexp      pcre.Regexp `json:"-"`
	RegEnable   bool        `json:"reg_enable"`
	Regulation  string      `json:"regulation"`
	StringValue string      `json:"string_value"`
	IntValue    float64     `json:"int_value"`
}

type logMetricDisplayObj struct {
	Metric     string            `json:"metric"`
	Path       string            `json:"path"`
	Agg        string            `json:"agg"`
	TEndpoint  string            `json:"t_endpoint"`
	Tags       []string          `json:"tags"`
	TagsString string            `json:"tags_string"`
	Value      float64           `json:"value"`
	ValueObj   logMetricValueObj `json:"value_obj"`
	Step       int64             `json:"step"`
	Display    bool              `json:"display"`
	UpdateTime int64             `json:"update_time"`
}

type logMetricValueObj struct {
	Sum   float64
	Count float64
	Max   float64
	Min   float64
}

func (c *logMetricMonitorNeObj) start() {
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("start log metric collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	firstFlag := true
	timeNow := time.Now()
	for line := range c.TailSession.Lines {
		if firstFlag {
			// load log file wait 5 sec to ignore exist log context
			if time.Now().Sub(timeNow).Seconds() >= 5 {
				firstFlag = false
			} else {
				continue
			}
		}
		c.Lock.RLock()
		for _, rule := range c.JsonConfig {
			if rule.Regexp == nil {
				continue
			}
			fetchList := rule.Regexp.FindStringSubmatch(line.Text)
			if len(fetchList) <= 1 {
				continue
			}
			fetchList = fetchList[1:]
			fetchKeyMap := make(map[string]interface{})
			for _, v := range fetchList {
				tmpKeyMap := make(map[string]interface{})
				tmpErr := json.Unmarshal([]byte(v), &tmpKeyMap)
				if tmpErr != nil {
					level.Error(monitorLogger).Log("line fetch regexp fail", fmt.Sprintf("line:%s error:%s", v, tmpErr.Error()))
				} else {
					for tmpKeyMapKey, tmpKeyMapValue := range tmpKeyMap {
						fetchKeyMap[tmpKeyMapKey] = tmpKeyMapValue
					}
				}
			}
			if len(fetchKeyMap) > 0 {
				rule.DataChannel <- fetchKeyMap
			}
		}
		for _, custom := range c.MetricConfig {
			if custom.RegExp == nil {
				continue
			}
			fetchList := custom.RegExp.FindStringSubmatch(line.Text)
			if len(fetchList) > 1 {
				if fetchList[1] != "" {
					custom.DataChannel <- fetchList[1]
				}
			}
		}
		c.Lock.RUnlock()
	}
}

func (c *logMetricMonitorNeObj) new(input *logMetricMonitorNeObj) {
	c.TargetEndpoint = input.TargetEndpoint
	c.JsonConfig = []*logMetricJsonNeObj{}
	for _, jsonObj := range input.JsonConfig {
		jsonObj.Regexp = regexp.MustCompile(jsonObj.Regular)
		jsonObj.DataChannel = make(chan map[string]interface{}, 10000)
		for _, metricObj := range jsonObj.MetricConfig {
			initLogMetricNeObj(metricObj)
		}
		c.JsonConfig = append(c.JsonConfig, jsonObj)
	}
	c.MetricConfig = []*logMetricNeObj{}
	for _, metricObj := range input.MetricConfig {
		initLogMetricNeObj(metricObj)
		c.MetricConfig = append(c.MetricConfig, metricObj)
	}
	go c.start()
}

func (c *logMetricMonitorNeObj) update(input *logMetricMonitorNeObj) {
	c.Lock.Lock()
	newJsonConfigList := []*logMetricJsonNeObj{}
	for _, jsonObj := range input.JsonConfig {
		for _, existJson := range c.JsonConfig {
			if jsonObj.Regular == existJson.Regular {
				jsonObj.DataChannel = existJson.DataChannel
				break
			}
		}
		if jsonObj.DataChannel == nil {
			jsonObj.DataChannel = make(chan map[string]interface{}, 10000)
		}
		jsonObj.Regexp = regexp.MustCompile(jsonObj.Regular)
		for _, metricObj := range jsonObj.MetricConfig {
			initLogMetricNeObj(metricObj)
		}
		newJsonConfigList = append(newJsonConfigList, jsonObj)
	}
	c.JsonConfig = newJsonConfigList
	newMetricConfigList := []*logMetricNeObj{}
	for _, metricObj := range input.MetricConfig {
		for _, existMetricObj := range c.MetricConfig {
			if metricObj.ValueRegular == existMetricObj.ValueRegular && metricObj.Metric == existMetricObj.Metric {
				metricObj.DataChannel = existMetricObj.DataChannel
				break
			}
		}
		initLogMetricNeObj(metricObj)
		newMetricConfigList = append(newMetricConfigList, metricObj)
	}
	c.MetricConfig = newMetricConfigList
	c.Lock.Unlock()
}

func initLogMetricNeObj(metricObj *logMetricNeObj) {
	if metricObj.ValueRegular != "" {
		metricObj.RegExp = regexp.MustCompile(metricObj.ValueRegular)
		if metricObj.DataChannel == nil {
			metricObj.DataChannel = make(chan string, 10000)
		}
	}
	for _, stringMapObj := range metricObj.StringMap {
		if stringMapObj.RegEnable {
			tmpExp, tmpErr := pcre.Compile(stringMapObj.StringValue, 0)
			if tmpErr == nil {
				stringMapObj.Regexp = tmpExp
			} else {
				stringMapObj.RegEnable = false
				level.Error(monitorLogger).Log("string map regexp format fail", tmpErr.Message)
			}
		}
	}
}

func (c *logMetricMonitorNeObj) destroy() {
	c.Lock.Lock()
	c.TailSession.Stop()
	c.JsonConfig = []*logMetricJsonNeObj{}
	c.MetricConfig = []*logMetricNeObj{}
	c.Lock.Unlock()
}

func LogMetricMonitorHttpHandle(w http.ResponseWriter, r *http.Request) {
	logMetricHttpLock.Lock()
	var err error
	defer func(returnErr error) {
		logMetricHttpLock.Unlock()
		responseObj := logMetricNodeExporterResponse{Status: "OK", Message: "success"}
		if returnErr != nil {
			returnErr = fmt.Errorf("Handel log metric monitor http request fail,%s ", returnErr.Error())
			responseObj = logMetricNodeExporterResponse{Status: "ERROR", Message: returnErr.Error()}
			level.Error(monitorLogger).Log("error", returnErr.Error())
		}
		b, _ := json.Marshal(responseObj)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}(err)
	var requestParamBuff []byte
	requestParamBuff, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	var param []*logMetricMonitorNeObj
	err = json.Unmarshal(requestParamBuff, &param)
	if err != nil {
		return
	}
	var tmpLogMetricObjJobs []*logMetricMonitorNeObj
	for _, logMetricMonitorJob := range logMetricMonitorJobs {
		delFlag := true
		for _, paramObj := range param {
			if paramObj.Path == logMetricMonitorJob.Path {
				delFlag = false
				// update config
				logMetricMonitorJob.update(paramObj)
				break
			}
		}
		if delFlag {
			// delete config
			logMetricMonitorJob.destroy()
		} else {
			tmpLogMetricObjJobs = append(tmpLogMetricObjJobs, logMetricMonitorJob)
		}
	}
	for _, paramObj := range param {
		addFlag := true
		for _, logMetricMonitorJob := range logMetricMonitorJobs {
			if logMetricMonitorJob.Path == paramObj.Path {
				addFlag = false
				break
			}
		}
		if !addFlag {
			continue
		}
		// add config
		newLogMetricObj := logMetricMonitorNeObj{Path: paramObj.Path, Lock: new(sync.RWMutex)}
		newLogMetricObj.new(paramObj)
		tmpLogMetricObjJobs = append(tmpLogMetricObjJobs, &newLogMetricObj)
	}
	logMetricMonitorJobs = tmpLogMetricObjJobs
}

func StartCalcLogMetricCron() {
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go calcLogMetricData()
	}
}

func calcLogMetricData() {
	logMetricHttpLock.RLock()
	existMetricMap := make(map[string]*logMetricDisplayObj)
	for _, displayObj := range logMetricMonitorMetrics {
		existMetricMap[fmt.Sprintf("%s^%s^%s^%s", displayObj.Path, displayObj.Metric, displayObj.Agg, displayObj.TagsString)] = displayObj
	}
	valueCountMap := make(map[string]*logMetricDisplayObj)
	for _, lmObj := range logMetricMonitorJobs {
		for _, jsonObj := range lmObj.JsonConfig {
			dataLength := len(jsonObj.DataChannel)
			if dataLength == 0 {
				continue
			}
			tmpTagsKey := []string{}
			if jsonObj.Tags != "" {
				tmpTagsKey = strings.Split(jsonObj.Tags, ",")
			}
			for i := 0; i < dataLength; i++ {
				tmpMapData := <-jsonObj.DataChannel
				// Get metric tags
				tmpTagString := getLogMetricJsonMapTags(tmpMapData, tmpTagsKey)
				// Try to change value to float64
				for _, metricConfig := range jsonObj.MetricConfig {
					changedMapData := getLogMetricJsonMapValue(tmpMapData, metricConfig.StringMap)
					if metricValueFloat, b := changedMapData[metricConfig.Key]; b {
						tmpMetricKey := fmt.Sprintf("%s^%s^%s^%s", lmObj.Path, metricConfig.Metric, metricConfig.AggType, tmpTagString)
						if valueExistObj, keyExist := valueCountMap[tmpMetricKey]; keyExist {
							valueExistObj.ValueObj.Sum += metricValueFloat
							valueExistObj.ValueObj.Count++
							if valueExistObj.ValueObj.Max < metricValueFloat {
								valueExistObj.ValueObj.Max = metricValueFloat
							}
							if valueExistObj.ValueObj.Min > metricValueFloat {
								valueExistObj.ValueObj.Min = metricValueFloat
							}
						} else {
							valueCountMap[tmpMetricKey] = &logMetricDisplayObj{Metric: metricConfig.Metric, Path: lmObj.Path, Agg: metricConfig.AggType, TEndpoint: lmObj.TargetEndpoint, TagsString: tmpTagString, Step: metricConfig.Step, ValueObj: logMetricValueObj{Sum: metricValueFloat, Max: metricValueFloat, Min: metricValueFloat, Count: 1}}
						}
					}
				}
			}
		}
		for _, metricObj := range lmObj.MetricConfig {
			dataLength := len(metricObj.DataChannel)
			if dataLength == 0 {
				continue
			}
			tmpMetricKey := fmt.Sprintf("%s^%s^%s^%s", lmObj.Path, metricObj.Metric, metricObj.AggType, "")
			tmpMetricObj := logMetricDisplayObj{Metric: metricObj.Metric, Path: lmObj.Path, Agg: metricObj.AggType, TEndpoint: lmObj.TargetEndpoint, TagsString: "", Step: metricObj.Step, ValueObj: logMetricValueObj{Sum: 0, Max: 0, Min: 0, Count: 0}}
			for i := 0; i < dataLength; i++ {
				customFetchString := <-metricObj.DataChannel
				metricValueFloat := transLogMetricStringMapValue(metricObj.StringMap, customFetchString)
				tmpMetricObj.ValueObj.Sum += metricValueFloat
				tmpMetricObj.ValueObj.Count++
				if tmpMetricObj.ValueObj.Max < metricValueFloat {
					tmpMetricObj.ValueObj.Max = metricValueFloat
				}
				if tmpMetricObj.ValueObj.Min > metricValueFloat {
					tmpMetricObj.ValueObj.Min = metricValueFloat
				}
			}
			valueCountMap[tmpMetricKey] = &tmpMetricObj
		}
	}
	tmpLogMetricMetrics := buildLogMetricDisplayMetrics(valueCountMap, existMetricMap)
	logMetricHttpLock.RUnlock()
	logMetricMonitorMetricLock.Lock()
	logMetricMonitorMetrics = tmpLogMetricMetrics
	logMetricMonitorMetricLock.Unlock()
}

func buildLogMetricDisplayMetrics(valueCountMap, existMetricMap map[string]*logMetricDisplayObj) (tmpLogMetricMetrics []*logMetricDisplayObj) {
	tmpLogMetricMetrics = []*logMetricDisplayObj{}
	nowTime := time.Now().Unix()
	for k, v := range valueCountMap {
		lastTimestamp := nowTime
		firstDisplay := true
		if existObj, b := existMetricMap[k]; b {
			firstDisplay = false
			if !existObj.Display {
				// keep append old data
				v.ValueObj.Sum += existObj.ValueObj.Sum
				v.ValueObj.Count += existObj.ValueObj.Count
				if v.ValueObj.Max < existObj.ValueObj.Max {
					v.ValueObj.Max = existObj.ValueObj.Max
				}
				if v.ValueObj.Min > existObj.ValueObj.Min {
					v.ValueObj.Min = existObj.ValueObj.Min
				}
				lastTimestamp = existObj.UpdateTime
			}
		}
		// check display or not
		if v.Step < 20 || firstDisplay {
			v.Display = true
		} else {
			if (nowTime - lastTimestamp) >= v.Step {
				v.Display = true
			} else {
				v.Display = false
			}
		}
		if v.Display {
			v.UpdateTime = nowTime
			// match value
			switch v.Agg {
			case "sum":
				v.Value = v.ValueObj.Sum
			case "count":
				v.Value = v.ValueObj.Count
			case "max":
				v.Value = v.ValueObj.Max
			case "min":
				v.Value = v.ValueObj.Min
			case "avg":
				v.Value = v.ValueObj.Sum / v.ValueObj.Count
			}
		} else {
			v.UpdateTime = lastTimestamp
		}
		tmpLogMetricMetrics = append(tmpLogMetricMetrics, v)
	}
	return tmpLogMetricMetrics
}

func transLogMetricStringMapValue(config []*logMetricStringMapNeObj, input string) (output float64) {
	if len(config) == 0 {
		output, _ = strconv.ParseFloat(input, 64)
		return output
	}
	for _, v := range config {
		if !v.RegEnable {
			if v.StringValue == input {
				output = v.IntValue
				break
			}
			continue
		}
		if len(v.Regexp.FindIndex([]byte(input), 0)) > 0 {
			output = v.IntValue
			break
		}
	}
	return output
}

func getLogMetricJsonMapValue(input map[string]interface{}, smConfig []*logMetricStringMapNeObj) (output map[string]float64) {
	output = make(map[string]float64)
	for k, v := range input {
		if v == nil || reflect.TypeOf(v) == nil {
			continue
		}
		var newValue float64 = 0
		typeString := reflect.TypeOf(v).String()
		if strings.Contains(typeString, "string") {
			valueString := fmt.Sprintf("%s", v)
			newValue = transLogMetricStringMapValue(smConfig, valueString)
		} else if strings.Contains(typeString, "int") {
			newValue, _ = strconv.ParseFloat(fmt.Sprintf("%d", v), 64)
		} else if strings.Contains(typeString, "float") {
			fmt.Sprintf("%.6f", input)
			newValue, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", v), 64)
		}
		output[k] = newValue
	}
	return output
}

func getLogMetricJsonMapTags(input map[string]interface{}, tagKey []string) (tagString string) {
	tagString = ""
	for _, v := range tagKey {
		if tmpTagValue, b := input[v]; b {
			tagString += fmt.Sprintf("%s=%s,", v, tmpTagValue)
		}
	}
	if tagString != "" {
		tagString = tagString[:len(tagString)-1]
	}
	return tagString
}
