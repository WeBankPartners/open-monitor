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
	log_metricMonitorFilePath = "data/log_metric_monitor_cache.json"
)

var (
	logMetricMonitorJobs       []*logMetricMonitorNeObj
	logMetricHttpLock          = new(sync.RWMutex)
	logMetricMonitorMetrics    []*logMetricDisplayObj
	logMetricMonitorMetricLock = new(sync.RWMutex)
	monitorLogger              log.Logger
	logMetricChanLength        = 100000
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
			[]string{"key", "tags", "path", "agg", "t_endpoint", "service_group"}, nil,
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
			v.Value, v.Metric, v.TagsString, v.Path, v.Agg, v.TEndpoint, v.ServiceGroup)
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
	ServiceGroup   string                `json:"service_group"`
	JsonConfig     []*logMetricJsonNeObj `json:"config"`
	MetricConfig   []*logMetricNeObj     `json:"custom"`
	DataChan       chan string           `json:"-"`
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
	TagConfig    []*LogMetricConfigTag      `json:"tag_config"`
}

type LogMetricConfigTag struct {
	Key     string         `json:"key"`
	Regular string         `json:"regular"`
	RegExp  *regexp.Regexp `json:"-"`
}

type logMetricStringMapNeObj struct {
	Regexp            pcre.Regexp `json:"-"`
	RegEnable         bool        `json:"reg_enable"`
	Regulation        string      `json:"regulation"`
	StringValue       string      `json:"string_value"`
	IntValue          float64     `json:"int_value"`
	TargetStringValue string      `json:"target_string_value"`
}

type logMetricDisplayObj struct {
	Metric       string            `json:"metric"`
	Path         string            `json:"path"`
	Agg          string            `json:"agg"`
	TEndpoint    string            `json:"t_endpoint"`
	ServiceGroup string            `json:"service_group"`
	Tags         []string          `json:"tags"`
	TagsString   string            `json:"tags_string"`
	Value        float64           `json:"value"`
	ValueObj     logMetricValueObj `json:"value_obj"`
	Step         int64             `json:"step"`
	Display      bool              `json:"display"`
	UpdateTime   int64             `json:"update_time"`
}

type logMetricValueObj struct {
	Sum   float64
	Count float64
	Max   float64
	Min   float64
}

func (c *logMetricMonitorNeObj) startHandleTailData() {
	for {
		lineText := <-c.DataChan
		c.Lock.RLock()
		for _, rule := range c.JsonConfig {
			if rule.Regexp == nil {
				continue
			}
			fetchList := rule.Regexp.FindStringSubmatch(lineText)
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
			fetchList := custom.RegExp.FindStringSubmatch(lineText)
			if len(fetchList) > 1 {
				if fetchList[1] != "" {
					custom.DataChannel <- fetchList[1]
				}
			}
		}
		c.Lock.RUnlock()
	}
}

func (c *logMetricMonitorNeObj) start() {
	level.Info(monitorLogger).Log("startLogMetricMonitorNeObj", c.Path)
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("start log metric collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	c.DataChan = make(chan string, logMetricChanLength)
	go c.startHandleTailData()
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
		if len(c.DataChan) == logMetricChanLength {
			level.Info(monitorLogger).Log("Log metric queue is full,file:", c.Path)
		}
		c.DataChan <- line.Text
	}
}

func (c *logMetricMonitorNeObj) new(input *logMetricMonitorNeObj) {
	level.Info(monitorLogger).Log("newLogMetricMonitorNeObj", c.Path)
	c.TargetEndpoint = input.TargetEndpoint
	c.ServiceGroup = input.ServiceGroup
	c.JsonConfig = []*logMetricJsonNeObj{}
	var err error
	for _, jsonObj := range input.JsonConfig {
		jsonObj.Regexp, err = regexp.Compile(jsonObj.Regular)
		if err != nil {
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", err.Error()))
			continue
		}
		jsonObj.DataChannel = make(chan map[string]interface{}, logMetricChanLength)
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
	level.Info(monitorLogger).Log("updateLogMetricMonitorNeObj", c.Path)
	newJsonConfigList := []*logMetricJsonNeObj{}
	var err error
	for _, jsonObj := range input.JsonConfig {
		jsonObj.Regexp, err = regexp.Compile(jsonObj.Regular)
		if err != nil {
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", err.Error()))
			continue
		}
		for _, existJson := range c.JsonConfig {
			if jsonObj.Regular == existJson.Regular {
				jsonObj.DataChannel = existJson.DataChannel
				break
			}
		}
		if jsonObj.DataChannel == nil {
			jsonObj.DataChannel = make(chan map[string]interface{}, logMetricChanLength)
		}
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
	c.TargetEndpoint = input.TargetEndpoint
	c.ServiceGroup = input.ServiceGroup
	c.Lock.Unlock()
}

func initLogMetricNeObj(metricObj *logMetricNeObj) {
	var err error
	if metricObj.ValueRegular != "" {
		metricObj.RegExp, err = regexp.Compile(metricObj.ValueRegular)
		if err != nil {
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", err.Error()))
			metricObj.RegExp, _ = regexp.Compile(".*")
		}
		if metricObj.DataChannel == nil {
			metricObj.DataChannel = make(chan string, logMetricChanLength)
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
	for _, tagConfigObj := range metricObj.TagConfig {
		tagConfigObj.RegExp, err = regexp.Compile(tagConfigObj.Regular)
		if err != nil {
			level.Error(monitorLogger).Log("newLogMetricMonitorTagConfig", fmt.Sprintf("regexpError:%s ", err.Error()))
			tagConfigObj.RegExp = nil
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
	err = LogMetricMonitorHandleAction(requestParamBuff)
	if err == nil {
		LogMetricSaveConfig(requestParamBuff)
	}
}

func LogMetricMonitorHandleAction(requestParamBuff []byte) error {
	var param []*logMetricMonitorNeObj
	err := json.Unmarshal(requestParamBuff, &param)
	if err != nil {
		return err
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
	return nil
}

func LogMetricSaveConfig(requestParamBuff []byte) {
	err := ioutil.WriteFile(log_metricMonitorFilePath, requestParamBuff, 0644)
	if err != nil {
		level.Error(monitorLogger).Log("logMetricSaveConfig", err.Error())
	} else {
		level.Info(monitorLogger).Log("logMetricSaveConfig", "success")
	}
}

func LogMetricLoadConfig() {
	b, err := ioutil.ReadFile(log_metricMonitorFilePath)
	if err != nil {
		level.Error(monitorLogger).Log("logMetricLoadConfig", err.Error())
	} else {
		err = LogMetricMonitorHandleAction(b)
		if err != nil {
			level.Error(monitorLogger).Log("logMetricLoadConfigAction", err.Error())
		} else {
			level.Info(monitorLogger).Log("logMetricLoadConfig", "success")
		}
	}
}

func StartCalcLogMetricCron() {
	LogMetricLoadConfig()
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go calcLogMetricData()
	}
}

func calcLogMetricData() {
	logMetricHttpLock.RLock()
	if len(logMetricMonitorJobs) == 0 {
		logMetricHttpLock.RUnlock()
		return
	}
	existMetricMap := make(map[string]*logMetricDisplayObj)
	for _, displayObj := range logMetricMonitorMetrics {
		existMetricMap[fmt.Sprintf("%s^%s^%s^%s", displayObj.Path, displayObj.Metric, displayObj.Agg, displayObj.TagsString)] = displayObj
	}
	appendDisplayMap := make(map[string]int)
	valueCountMap := make(map[string]*logMetricDisplayObj)
	for _, lmObj := range logMetricMonitorJobs {
		for _, jsonObj := range lmObj.JsonConfig {
			// pull channel data list
			jsonDataList := []map[string]interface{}{}
			dataLength := len(jsonObj.DataChannel)
			for i := 0; i < dataLength; i++ {
				tmpMapData := <-jsonObj.DataChannel
				jsonDataList = append(jsonDataList, tmpMapData)
			}
			tmpTagsKey := []string{}
			if jsonObj.Tags != "" {
				tmpTagsKey = strings.Split(jsonObj.Tags, ",")
			}
			for _, metricConfig := range jsonObj.MetricConfig {
				isMatchNewDataFlag := false
				for _, tmpMapData := range jsonDataList {
					// Get metric tags
					tmpTagString := getLogMetricJsonMapTags(tmpMapData, tmpTagsKey)
					changedMapData := getLogMetricJsonMapValue(tmpMapData, metricConfig.StringMap)
					if metricValueFloat, b := changedMapData[metricConfig.Key]; b {
						isMatchNewDataFlag = true
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
							valueCountMap[tmpMetricKey] = &logMetricDisplayObj{Metric: metricConfig.Metric, Path: lmObj.Path, Agg: metricConfig.AggType, TEndpoint: lmObj.TargetEndpoint, ServiceGroup: lmObj.ServiceGroup, TagsString: tmpTagString, Step: metricConfig.Step, ValueObj: logMetricValueObj{Sum: metricValueFloat, Max: metricValueFloat, Min: metricValueFloat, Count: 1}}
						}
					}
				}
				if !isMatchNewDataFlag {
					if metricConfig.AggType == "avg" {
						appendDisplayMap[fmt.Sprintf("%s^%s^sum", lmObj.Path, metricConfig.Metric)] = 1
						appendDisplayMap[fmt.Sprintf("%s^%s^count", lmObj.Path, metricConfig.Metric)] = 1
					}
					appendDisplayMap[fmt.Sprintf("%s^%s^%s", lmObj.Path, metricConfig.Metric, metricConfig.AggType)] = 1
				}
			}
		}
		for _, metricObj := range lmObj.MetricConfig {
			dataLength := len(metricObj.DataChannel)
			if dataLength == 0 {
				if metricObj.AggType == "avg" {
					appendDisplayMap[fmt.Sprintf("%s^%s^sum", lmObj.Path, metricObj.Metric)] = 1
					appendDisplayMap[fmt.Sprintf("%s^%s^count", lmObj.Path, metricObj.Metric)] = 1
				}
				appendDisplayMap[fmt.Sprintf("%s^%s^%s", lmObj.Path, metricObj.Metric, metricObj.AggType)] = 1
				continue
			}
			//tmpMetricKey := fmt.Sprintf("%s^%s^%s^%s", lmObj.Path, metricObj.Metric, metricObj.AggType, "")
			//tmpMetricObj := logMetricDisplayObj{Metric: metricObj.Metric, Path: lmObj.Path, Agg: metricObj.AggType, TEndpoint: lmObj.TargetEndpoint, ServiceGroup: lmObj.ServiceGroup, TagsString: "", Step: metricObj.Step, ValueObj: logMetricValueObj{Sum: 0, Max: 0, Min: 0, Count: 0}}
			for i := 0; i < dataLength; i++ {
				customFetchString := <-metricObj.DataChannel
				//tmpMetricObj := logMetricDisplayObj{Metric: metricObj.Metric, Path: lmObj.Path, Agg: metricObj.AggType, TEndpoint: lmObj.TargetEndpoint, ServiceGroup: lmObj.ServiceGroup, TagsString: "", Step: metricObj.Step, ValueObj: logMetricValueObj{Sum: 0, Max: 0, Min: 0, Count: 0}}
				// check if tag match
				illegalFlag, tmpTagString := getLogMetricTags(customFetchString, metricObj.TagConfig, metricObj.StringMap)
				if illegalFlag {
					continue
				}
				tmpMetricKey := fmt.Sprintf("%s^%s^%s^%s", lmObj.Path, metricObj.Metric, metricObj.AggType, tmpTagString)
				_, metricValueFloat := transLogMetricStringMapValue(metricObj.StringMap, customFetchString)
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
					valueCountMap[tmpMetricKey] = &logMetricDisplayObj{Metric: metricObj.Metric, Path: lmObj.Path, Agg: metricObj.AggType, TEndpoint: lmObj.TargetEndpoint, ServiceGroup: lmObj.ServiceGroup, TagsString: tmpTagString, Step: metricObj.Step, ValueObj: logMetricValueObj{Sum: metricValueFloat, Max: metricValueFloat, Min: metricValueFloat, Count: 1}}
				}
			}
			//valueCountMap[tmpMetricKey] = &tmpMetricObj
		}
	}
	if len(appendDisplayMap) > 0 {
		for k, v := range existMetricMap {
			tmpKey := fmt.Sprintf("%s^%s^%s", v.Path, v.Metric, v.Agg)
			if _, b := appendDisplayMap[tmpKey]; b {
				if v.Display {
					v.ValueObj = logMetricValueObj{Sum: 0, Count: 0, Max: 0, Min: 0}
				}
				valueCountMap[k] = v
			}
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
			if (nowTime - lastTimestamp + 10) >= v.Step {
				v.Display = true
			} else {
				v.Display = false
			}
		}
		avgFlag := false
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
				avgFlag = true
			}
		} else {
			v.UpdateTime = lastTimestamp
		}
		if avgFlag {
			if v.ValueObj.Count > 0 {
				v.Value = v.ValueObj.Sum / v.ValueObj.Count
			} else {
				v.Value = 0
			}
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, &logMetricDisplayObj{Metric: v.Metric, Path: v.Path, Agg: "avg", TEndpoint: v.TEndpoint, ServiceGroup: v.ServiceGroup, Tags: v.Tags, TagsString: v.TagsString, Value: v.Value, Step: v.Step, Display: v.Display, UpdateTime: v.UpdateTime})
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, &logMetricDisplayObj{Metric: v.Metric, Path: v.Path, Agg: "sum", TEndpoint: v.TEndpoint, ServiceGroup: v.ServiceGroup, Tags: v.Tags, TagsString: v.TagsString, Value: v.ValueObj.Sum, Step: v.Step, Display: v.Display, UpdateTime: v.UpdateTime})
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, &logMetricDisplayObj{Metric: v.Metric, Path: v.Path, Agg: "count", TEndpoint: v.TEndpoint, ServiceGroup: v.ServiceGroup, Tags: v.Tags, TagsString: v.TagsString, Value: v.ValueObj.Count, Step: v.Step, Display: v.Display, UpdateTime: v.UpdateTime})
		} else {
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, v)
		}
	}
	return tmpLogMetricMetrics
}

func transLogMetricTagString(config []*logMetricStringMapNeObj, input string) (matchFlag bool, output string) {
	output = input
	if len(config) == 0 {
		return
	}
	for _, v := range config {
		if !v.RegEnable {
			if v.StringValue == input {
				matchFlag = true
				output = v.TargetStringValue
				break
			}
			continue
		}
		if len(v.Regexp.FindIndex([]byte(input), 0)) > 0 {
			matchFlag = true
			output = v.TargetStringValue
			break
		}
	}
	return
}

func transLogMetricStringMapValue(config []*logMetricStringMapNeObj, input string) (matchFlag bool, output float64) {
	if len(config) == 0 {
		output, _ = strconv.ParseFloat(input, 64)
		return
	}
	for _, v := range config {
		if !v.RegEnable {
			if v.StringValue == input {
				matchFlag = true
				output = v.IntValue
				break
			}
			continue
		}
		if len(v.Regexp.FindIndex([]byte(input), 0)) > 0 {
			matchFlag = true
			output = v.IntValue
			break
		}
	}
	return
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
			_, newValue = transLogMetricStringMapValue(smConfig, valueString)
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

func getLogMetricTags(lineText string, tagConfigList []*LogMetricConfigTag, stringMap []*logMetricStringMapNeObj) (illegalFlag bool, tagString string) {
	if len(tagConfigList) == 0 {
		return
	}
	var tagList []string
	for _, rule := range tagConfigList {
		if rule.RegExp == nil {
			continue
		}
		fetchList := rule.RegExp.FindStringSubmatch(lineText)
		tmpFetchString := ""
		if len(fetchList) > 1 {
			tmpFetchString = fetchList[1]
			if tmpFetchString != "" {
				if len(stringMap) > 0 {
					if tmpMatchFlag, tmpMatchData := transLogMetricTagString(stringMap, tmpFetchString); tmpMatchFlag {
						tmpFetchString = tmpMatchData
					}
				}
				tagList = append(tagList, fmt.Sprintf("%s=%s", rule.Key, tmpFetchString))
			}
		}
		if tmpFetchString == "" {
			illegalFlag = true
			break
		}
	}
	tagString = strings.Join(tagList, ",")
	return
}
