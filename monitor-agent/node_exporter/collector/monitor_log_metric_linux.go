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
	logMetricHttpLock        = new(sync.RWMutex)
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
	Regexp       *regexp.Regexp                 `json:"-"`
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
	Step         int                        `json:"step"`
	StringMap    []*logMetricStringMapNeObj `json:"string_map"`
}

type logMetricStringMapNeObj struct {
	Regexp      pcre.Regexp `json:"-"`
	RegEnable   bool    `json:"reg_enable"`
	Regulation  string  `json:"regulation"`
	StringValue string  `json:"string_value"`
	IntValue    float64 `json:"int_value"`
}

type logMetricDisplayObj struct {
	Metric     string   `json:"metric"`
	Path       string   `json:"path"`
	Agg        string   `json:"agg"`
	TEndpoint  string   `json:"t_endpoint"`
	Tags       []string `json:"tags"`
	TagsString string   `json:"tags_string"`
	Value      float64  `json:"value"`
	ValueObj   logMetricValueObj `json:"value_obj"`
	Step       int      `json:"step"`
	Display    bool     `json:"display"`
	UpdateTime int64    `json:"update_time"`
}

type logMetricValueObj struct {
	Sum   float64
	Avg   float64
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

func (c *logMetricMonitorNeObj) new(input *logMetricMonitorNeObj)  {
	c.TargetEndpoint = input.TargetEndpoint
	c.JsonConfig = []*logMetricJsonNeObj{}
	for _,jsonObj := range input.JsonConfig {
		jsonObj.Regexp = regexp.MustCompile(jsonObj.Regular)
		jsonObj.DataChannel = make(chan map[string]interface{}, 10000)
		for _,metricObj := range jsonObj.MetricConfig {
			initLogMetricNeObj(metricObj)
		}
		c.JsonConfig = append(c.JsonConfig, jsonObj)
	}
	c.MetricConfig = []*logMetricNeObj{}
	for _,metricObj := range input.MetricConfig {
		initLogMetricNeObj(metricObj)
		c.MetricConfig = append(c.MetricConfig, metricObj)
	}
	go c.start()
}

func (c *logMetricMonitorNeObj) update(input *logMetricMonitorNeObj)  {
	c.Lock.Lock()
	newJsonConfigList := []*logMetricJsonNeObj{}
	for _,jsonObj := range input.JsonConfig {
		for _,existJson := range c.JsonConfig {
			if jsonObj.Regular == existJson.Regular {
				jsonObj.DataChannel = existJson.DataChannel
				break
			}
		}
		if jsonObj.DataChannel == nil {
			jsonObj.DataChannel = make(chan map[string]interface{}, 10000)
		}
		jsonObj.Regexp = regexp.MustCompile(jsonObj.Regular)
		for _,metricObj := range jsonObj.MetricConfig {
			initLogMetricNeObj(metricObj)
		}
		newJsonConfigList = append(newJsonConfigList, jsonObj)
	}
	c.JsonConfig = newJsonConfigList
	newMetricConfigList := []*logMetricNeObj{}
	for _,metricObj := range input.MetricConfig {
		for _,existMetricObj := range c.MetricConfig {
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

func initLogMetricNeObj(metricObj *logMetricNeObj)  {
	if metricObj.ValueRegular != "" {
		metricObj.RegExp = regexp.MustCompile(metricObj.ValueRegular)
		if metricObj.DataChannel == nil {
			metricObj.DataChannel = make(chan string, 10000)
		}
	}
	for _,stringMapObj := range metricObj.StringMap {
		if stringMapObj.RegEnable {
			tmpExp, tmpErr := pcre.Compile(stringMapObj.StringValue, 0)
			if tmpErr == nil {
				stringMapObj.Regexp = tmpExp
			}else{
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
		b,_ := json.Marshal(responseObj)
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
		}else{
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
		newLogMetricObj := logMetricMonitorNeObj{Path: paramObj.Path,Lock: new(sync.RWMutex)}
		newLogMetricObj.new(paramObj)
		tmpLogMetricObjJobs = append(tmpLogMetricObjJobs, &newLogMetricObj)
	}
	logMetricMonitorJobs = tmpLogMetricObjJobs
}

func StartCalcLogMetricCron()  {
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go calcLogMetricData()
	}
}

func calcLogMetricData()  {
	tmpLogMetricMetrics :=  []*logMetricDisplayObj{}
	logMetricHttpLock.RLock()
	existMetricMap := make(map[string]*logMetricDisplayObj)
	for _,displayObj := range logMetricMonitorMetrics {
		if !displayObj.Display {
			existMetricMap[fmt.Sprintf("%s^%s",displayObj.Metric,displayObj.Path)] = displayObj
		}
	}
	for _,lmObj := range logMetricMonitorJobs {
		for _,jsonObj := range lmObj.JsonConfig {
			dataLength := len(jsonObj.DataChannel)
			if dataLength == 0 {
				continue
			}
			tmpTagsKey := []string{}
			if jsonObj.Tags != "" {
				tmpTagsKey = strings.Split(jsonObj.Tags, ",")
			}
			for i:=0;i<dataLength;i++ {
				tmpMapData := <-jsonObj.DataChannel
				// Try to change value to float64
				for _,metricConfig := range jsonObj.MetricConfig {
					changedMapData, tmpTagString := getJsonMapValue(tmpMapData, tmpTagsKey, metricConfig.StringMap)
				}
			}
		}
	}


	for _, v := range businessMonitorJobs {
		for _, rule := range v.Rules {
			dataLength := len(rule.DataChannel)
			if dataLength == 0 {
				continue
			}
			valueCountMap := make(map[string]*businessValueObj)
			for i := 0; i < dataLength; i++ {
				tmpMapData := <-rule.DataChannel
				// Try to change value to float64
				changedMapData, tmpTagString := changeValueByStringMap(tmpMapData, rule.TagsKey, rule.StringMap)
				for _, metricConfig := range rule.MetricConfig {
					if metricValueFloat, b := changedMapData[metricConfig.Key]; b {
						tmpMapKey := fmt.Sprintf("%s^%s^%s^%s", metricConfig.Key, metricConfig.AggType, metricConfig.Metric, tmpTagString)
						if _, keyExist := valueCountMap[tmpMapKey]; keyExist {
							valueCountMap[tmpMapKey].Sum += metricValueFloat
							valueCountMap[tmpMapKey].Count++
						} else {
							valueCountMap[tmpMapKey] = &businessValueObj{Sum: metricValueFloat, Count: 1, Avg: metricValueFloat}
						}
					}
				}
			}
			for mapKey, mapValue := range valueCountMap {
				mapValue.Avg = mapValue.Sum / mapValue.Count
				keySplitList := strings.Split(mapKey, "^")
				tmpMetricObj := businessRuleMetricObj{Path: v.Path, Agg: keySplitList[1], TagsString: keySplitList[3], Metric: keySplitList[2]}
				if keySplitList[1] == "sum" {
					tmpMetricObj.Value = mapValue.Sum
				} else if keySplitList[1] == "avg" {
					tmpMetricObj.Value = mapValue.Avg
				} else if keySplitList[1] == "count" {
					tmpMetricObj.Value = mapValue.Count
				}
				newRuleData = append(newRuleData, &tmpMetricObj)
			}
		}
		for _, custom := range v.Custom {
			dataLength := len(custom.DataChannel)
			if dataLength == 0 {
				continue
			}
			tmpValueObj := businessValueObj{}
			for i := 0; i < dataLength; i++ {
				customFetchString := <-custom.DataChannel
				// Try to change value to float64
				tmpMapData := make(map[string]interface{})
				tmpMapData[custom.Metric] = customFetchString
				changedMapData, _ := changeValueByStringMap(tmpMapData, []string{}, custom.StringMap)
				tmpValueObj.Sum += changedMapData[custom.Metric]
				tmpValueObj.Count++
			}
			if tmpValueObj.Count > 0 {
				tmpValueObj.Avg = tmpValueObj.Sum / tmpValueObj.Count
				tmpMetricObj := businessRuleMetricObj{Path: v.Path, Agg: custom.AggType, Metric: custom.Metric, Tags: []string{}}
				if custom.AggType == "sum" {
					tmpMetricObj.Value = tmpValueObj.Sum
				} else if custom.AggType == "avg" {
					tmpMetricObj.Value = tmpValueObj.Avg
				} else if custom.AggType == "count" {
					tmpMetricObj.Value = tmpValueObj.Count
				}
				newRuleData = append(newRuleData, &tmpMetricObj)
			}
		}
	}
	logMetricHttpLock.RUnlock()
	logMetricMonitorMetricLock.Lock()
	logMetricMonitorMetrics = tmpLogMetricMetrics
	logMetricMonitorMetricLock.Unlock()
}

func transStringMapValue(config []*logMetricStringMapNeObj,input string) (output float64) {
	if len(config) == 0 {
		output, _ = strconv.ParseFloat(input, 64)
		return output
	}
	for _,v := range config {
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

func getJsonMapValue(input map[string]interface{}, tagKey []string, smConfig []*logMetricStringMapNeObj) (output map[string]float64, tagString string) {
	output = make(map[string]float64)
	for k, v := range input {
		if v == nil || reflect.TypeOf(v) == nil {
			continue
		}
		var newValue float64 = 0
		typeString := reflect.TypeOf(v).String()
		if strings.Contains(typeString, "string") {
			valueString := fmt.Sprintf("%s", v)
			newValue = transStringMapValue(smConfig, valueString)
		} else if strings.Contains(typeString, "int") {
			newValue, _ = strconv.ParseFloat(fmt.Sprintf("%d", v), 64)
		} else if strings.Contains(typeString, "float") {
			fmt.Sprintf("%.6f", input)
			newValue, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", v), 64)
		}
		output[k] = newValue
	}
	for _, v := range tagKey {
		if tmpTagValue, b := input[v]; b {
			tagString += fmt.Sprintf("%s=%s,", v, tmpTagValue)
		}
	}
	if tagString != "" {
		tagString = tagString[:len(tagString)-1]
	}
	return output, tagString
}