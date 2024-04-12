package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	//"github.com/dlclark/regexp2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"reflect"
	//"regexp"
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
	TailSession       *tail.Tail            `json:"-"`
	Lock              *sync.RWMutex         `json:"-"`
	Path              string                `json:"path"`
	TargetEndpoint    string                `json:"target_endpoint"`
	ServiceGroup      string                `json:"service_group"`
	JsonConfig        []*logMetricJsonNeObj `json:"config"`
	MetricConfig      []*logMetricNeObj     `json:"custom"`
	DataChan          chan string           `json:"-"`
	ReOpenHandlerChan chan int              `json:"-"`
	TailTimeLock      *sync.RWMutex         `json:"-"`
	TailLastUnixTime  int64                 `json:"-"`
}

type logMetricJsonNeObj struct {
	Regexp       *Regexp                     `json:"-"`
	DataChannel  chan map[string]interface{} `json:"-"`
	Regular      string                      `json:"regular"`
	Tags         string                      `json:"tags"`
	MetricConfig []*logMetricNeObj           `json:"metric_config"`
}

type logMetricNeObj struct {
	RegExp       *Regexp                    `json:"-"`
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
	Regexp            *Regexp `json:"-"`
	RegEnable         bool    `json:"reg_enable"`
	Regulation        string  `json:"regulation"`
	StringValue       string  `json:"string_value"`
	IntValue          float64 `json:"int_value"`
	TargetStringValue string  `json:"target_string_value"`
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
		//level.Info(monitorLogger).Log("get a new line:", lineText)
		for _, rule := range c.JsonConfig {
			if rule.Regexp == nil {
				continue
			}
			fetchList := pcreMatchSubString(rule.Regexp, lineText)
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
			if matchList := pcreMatchSubString(custom.RegExp, lineText); len(matchList) > 0 {
				//level.Info(monitorLogger).Log("get a match line:", lineText)
				custom.DataChannel <- matchList[0]
			}
			//if custom.RegExp == nil {
			//	continue
			//}
			//fetchList := custom.RegExp.FindStringMatch(lineText)
			//if len(fetchList) > 1 {
			//	if fetchList[1] != "" {
			//		custom.DataChannel <- fetchList[1]
			//	}
			//}
		}
		c.Lock.RUnlock()
	}
}

//func regexp2FindStringMatch(re *regexp2.Regexp, lineText string) (matchString string) {
//	if re == nil {
//		return
//	}
//	mat, err := re.FindStringMatch(lineText)
//	if err != nil || mat == nil {
//		return
//	}
//	level.Info(monitorLogger).Log("regexp2FindStringMatch:", len(mat.Groups()))
//	for i, v := range mat.Groups() {
//		groupString := v.String()
//		level.Info(monitorLogger).Log("mat.Groups:", groupString)
//		if (i == 0 && groupString == lineText) || groupString == "" {
//			continue
//		}
//		matchString = groupString
//		break
//	}
//	return
//}

func pcreMatchSubString(re *Regexp, lineText string) (matchList []string) {
	if re == nil {
		return
	}
	mat := re.MatcherString(lineText, 0)
	if mat != nil {
		for i := 0; i <= mat.Groups(); i++ {
			groupString := mat.GroupString(i)
			if (i == 0 && groupString == lineText) || groupString == "" {
				continue
			}
			matchList = append(matchList, groupString)
		}
	}
	return
}

func (c *logMetricMonitorNeObj) start() {
	level.Info(monitorLogger).Log("log_metric -> startLogMetricMonitorNeObj__start", c.Path)
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}})
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("start log metric collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	c.DataChan = make(chan string, logMetricChanLength)
	go c.startHandleTailData()
	go c.startFileHandlerCheck()
	breakFlag := false
	for {
		select {
		case <-c.ReOpenHandlerChan:
			breakFlag = true
		case line := <-c.TailSession.Lines:
			if line == nil {
				continue
			}
			c.DataChan <- line.Text
		}
		if breakFlag {
			break
		} else {
			c.TailTimeLock.Lock()
			c.TailLastUnixTime = time.Now().Unix()
			c.TailTimeLock.Unlock()
		}
	}
	c.TailSession.Stop()
	c.TailSession.Cleanup()
	level.Info(monitorLogger).Log("log_metric -> startLogMetricMonitorNeObj__end", c.Path)
	time.Sleep(2 * time.Second)
	go c.start()
}

func (c *logMetricMonitorNeObj) startFileHandlerCheck() {
	t := time.NewTicker(1 * time.Minute).C
	for {
		<-t
		if fileLastTime, err := getFileLastUpdatedTime(c.Path); err == nil {
			c.TailTimeLock.RLock()
			tailLastTime := c.TailLastUnixTime
			c.TailTimeLock.RUnlock()
			if tailLastTime == 0 {
				c.TailTimeLock.Lock()
				c.TailLastUnixTime = fileLastTime
				c.TailTimeLock.Unlock()
				tailLastTime = fileLastTime
			}
			if fileLastTime-tailLastTime > 60 {
				c.ReOpenHandlerChan <- 1
				level.Info(monitorLogger).Log(fmt.Sprintf("log_metric -> reopen_tail_with_time_check_fail,path:%s,fileLastTime:%d,tailLastTime:%d ", c.Path, fileLastTime, tailLastTime))
			} else {
				level.Info(monitorLogger).Log(fmt.Sprintf("log_metric -> reopen_tail_with_time_check_ok,path:%s,fileLastTime:%d,tailLastTime:%d ", c.Path, fileLastTime, tailLastTime))
			}
		} else {
			level.Error(monitorLogger).Log(fmt.Sprintf("log_metric -> check_file_handler_fail,path:%s,err:%s ", c.Path, err.Error()))
		}
	}
}

func (c *logMetricMonitorNeObj) new(input *logMetricMonitorNeObj) {
	level.Info(monitorLogger).Log("newLogMetricMonitorNeObj", c.Path)
	c.TargetEndpoint = input.TargetEndpoint
	c.ServiceGroup = input.ServiceGroup
	c.JsonConfig = []*logMetricJsonNeObj{}
	c.TailTimeLock = new(sync.RWMutex)
	c.ReOpenHandlerChan = make(chan int, 1)
	var err error
	for _, jsonObj := range input.JsonConfig {
		tmpReg, tmpErr := PcreCompile(jsonObj.Regular, 0)
		if tmpErr != nil {
			err = fmt.Errorf(tmpErr.Message)
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", err.Error()))
			continue
		}
		jsonObj.Regexp = &tmpReg
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
		tmpExp, tmpErr := PcreCompile(jsonObj.Regular, 0)
		if tmpErr != nil {
			err = fmt.Errorf(tmpErr.Message)
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", err.Error()))
			continue
		}
		jsonObj.Regexp = &tmpExp
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
		if newRegExp, compileErr := PcreCompile(metricObj.ValueRegular, 0); compileErr != nil {
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", compileErr.Message))
		} else {
			metricObj.RegExp = &newRegExp
		}
		//metricObj.RegExp, err = PcreCompile(metricObj.ValueRegular, 0)
		//if err != nil {
		//	level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", err.Error()))
		//}
		if metricObj.DataChannel == nil {
			metricObj.DataChannel = make(chan string, logMetricChanLength)
		}
	}
	for _, stringMapObj := range metricObj.StringMap {
		if stringMapObj.RegEnable {
			if tmpExp, tmpErr := PcreCompile(stringMapObj.StringValue, 0); tmpErr != nil {
				stringMapObj.RegEnable = false
				level.Error(monitorLogger).Log("string map regexp format fail", tmpErr.Message)
			} else {
				stringMapObj.Regexp = &tmpExp
			}
			//if tmpErr == nil {
			//	stringMapObj.Regexp = tmpExp
			//} else {
			//	stringMapObj.RegEnable = false
			//	level.Error(monitorLogger).Log("string map regexp format fail", tmpErr.Error())
			//}
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
		level.Warn(monitorLogger).Log("logMetricLoadConfig", err.Error())
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
			//tmpTagsKey := []string{}
			//if jsonObj.Tags != "" {
			//	tmpTagsKey = strings.Split(jsonObj.Tags, ",")
			//}
			for _, metricConfig := range jsonObj.MetricConfig {
				isMatchNewDataFlag := false
				for _, tmpMapData := range jsonDataList {
					// Get metric tags
					tmpTagsKey := []string{}
					for _, tmpJsonTagItem := range metricConfig.TagConfig {
						tmpTagsKey = append(tmpTagsKey, tmpJsonTagItem.Key)
					}
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
		if mat := v.Regexp.MatcherString(input, 0); mat != nil {
			if mat.Matches() {
				matchFlag = true
				output = v.TargetStringValue
				break
			}
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
		if mat := v.Regexp.MatcherString(input, 0); mat != nil {
			if mat.Matches() {
				matchFlag = true
				output = v.IntValue
				break
			}
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

func getFileLastUpdatedTime(filePath string) (unixTime int64, err error) {
	f, statErr := os.Stat(filePath)
	if statErr != nil {
		err = fmt.Errorf("check file update time fail with stat file:%s %s ", filePath, statErr.Error())
		return
	}
	unixTime = f.ModTime().Unix()
	return
}
