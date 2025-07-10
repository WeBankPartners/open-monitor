package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	//"github.com/dlclark/regexp2"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"

	//"regexp"
	"os/exec"
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
			[]string{"key", "tags", "path", "agg", "t_endpoint", "service_group", "code", "retcode"}, nil,
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
			v.Value, v.Metric, v.TagsString, v.Path, v.Agg, v.TEndpoint, v.ServiceGroup, v.Code, v.RetCode)
	}
	logMetricMonitorMetricLock.RUnlock()
	return nil
}

type logMetricNodeExporterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type logMetricMonitorNeObj struct {
	TailSession        *tail.Tail             `json:"-"`
	Lock               *sync.RWMutex          `json:"-"`
	Path               string                 `json:"path"`
	TargetEndpoint     string                 `json:"target_endpoint"`
	ServiceGroup       string                 `json:"service_group"`
	JsonConfig         []*logMetricJsonNeObj  `json:"config"`
	MetricConfig       []*logMetricNeObj      `json:"custom"`
	MetricGroupConfig  []*logMetricGroupNeObj `json:"metric_group_config"`
	DataChan           chan string            `json:"-"`
	ReOpenHandlerChan  chan int               `json:"-"`
	TailTimeLock       *sync.RWMutex          `json:"-"`
	TailLastUnixTime   int64                  `json:"-"`
	DestroyChan        chan int               `json:"-"`
	TailDataCancelChan chan int               `json:"-"`
	MultiPathNum       int                    `json:"-"`
}

type logMetricGroupNeObj struct {
	JsonRegexp     *Regexp                     `json:"-"`
	DataChannel    chan map[string]interface{} `json:"-"`
	LogMetricGroup string                      `json:"log_metric_group"`
	LogType        string                      `json:"log_type"`
	JsonRegular    string                      `json:"json_regular"`
	ParamList      []*logMetricParamNeObj      `json:"param_list"`
	MetricConfig   []*logMetricNeObj           `json:"custom"`
}

type logMetricParamNeObj struct {
	RegExp    *Regexp                    `json:"-"`
	Name      string                     `json:"name"`
	JsonKey   string                     `json:"json_key"`
	Regular   string                     `json:"regular"`
	StringMap []*logMetricStringMapNeObj `json:"string_map"`
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
	LogParamName string                     `json:"log_param_name"`
}

type LogMetricConfigTag struct {
	Key          string         `json:"key"`
	Regular      string         `json:"regular"`
	RegExp       *regexp.Regexp `json:"-"`
	LogParamName string         `json:"log_param_name"`
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
	Id             string            `json:"id"`
	Metric         string            `json:"metric"`
	Path           string            `json:"path"`
	Agg            string            `json:"agg"`
	TEndpoint      string            `json:"t_endpoint"`
	ServiceGroup   string            `json:"service_group"`
	Tags           []string          `json:"tags"`
	TagsString     string            `json:"tags_string"`
	Value          float64           `json:"value"`
	ValueObj       logMetricValueObj `json:"value_obj"`
	Step           int64             `json:"step"`
	Display        bool              `json:"display"` // 用来控制采集间隔,默认最小间隔10s,当间隔为30s时,通过display来控制30s才出现汇总一次数据
	UpdateTime     int64             `json:"update_time"`
	Code           string            `json:"code"`
	RetCode        string            `json:"ret_code"`
	LastActiveTime int64             `json:"last_active_time"`
	ByAvgFlag      bool              `json:"by_avg_flag"`
}

type logMetricValueObj struct {
	Sum   float64
	Count float64
	Max   float64
	Min   float64
}

func (c *logMetricMonitorNeObj) startHandleTailData() {
	for {
		var lineText string
		select {
		case lineText = <-c.DataChan:
		case <-c.TailDataCancelChan:
			level.Info(monitorLogger).Log("log_metric -> logMetricMonitorNeObj_tail_data_cancel", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
			return
		}
		//lineText := <-c.DataChan
		//level.Info(monitorLogger).Log("log_metric_get_new_line ->", lineText)
		//lineText = strings.ReplaceAll(lineText, "\\t", "    ")
		c.Lock.RLock()
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
		}
		for _, metricGroup := range c.MetricGroupConfig {
			fetchParamValueMap := make(map[string]interface{})
			if metricGroup.LogType == "json" {
				if metricGroup.JsonRegexp != nil {
					// 先用匹配出json串,可以匹配多段json
					fetchList := pcreMatchSubString(metricGroup.JsonRegexp, lineText)
					fetchJsonDataMap := make(map[string]interface{})
					for _, v := range fetchList {
						//level.Info(monitorLogger).Log("log_json_group -> fetch", v)
						tmpKeyMap := make(map[string]interface{})
						tmpErr := json.Unmarshal([]byte(v), &tmpKeyMap)
						if tmpErr != nil {
							level.Error(monitorLogger).Log("line fetch regexp fail", fmt.Sprintf("line:%s error:%s", v, tmpErr.Error()))
						} else {
							for tmpKeyMapKey, tmpKeyMapValue := range tmpKeyMap {
								fetchJsonDataMap[tmpKeyMapKey] = tmpKeyMapValue
							}
						}
					}
					// 再检测参数里的key是否都有,必须都有
					allMatchFlag := true
					for _, metricParam := range metricGroup.ParamList {
						if fetchValue, ok := fetchJsonDataMap[metricParam.JsonKey]; !ok {
							allMatchFlag = false
							//level.Info(monitorLogger).Log("log_json_group -> check key fail", metricParam.JsonKey)
							break
						} else {
							fetchParamValueMap[metricParam.Name] = transMetricGroupData(fetchValue, metricParam.StringMap)
						}
					}
					if allMatchFlag {
						//fetchParamValueMapByte, _ := json.Marshal(fetchParamValueMap)
						metricGroup.DataChannel <- fetchParamValueMap
					}
				}
			} else {
				allMatchFlag := true
				for _, metricParam := range metricGroup.ParamList {
					if metricParam.RegExp != nil {
						if matchList := pcreMatchSubString(metricParam.RegExp, lineText); len(matchList) > 0 {
							fetchParamValueMap[metricParam.Name] = transMetricGroupData(matchList[0], metricParam.StringMap)
						} else {
							allMatchFlag = false
							break
						}
					}
				}
				if allMatchFlag {
					metricGroup.DataChannel <- fetchParamValueMap
				}
			}
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
	lineText = strings.TrimSpace(lineText)
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
	level.Info(monitorLogger).Log("log_metric -> startLogMetricMonitorNeObj__start", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}, Poll: true})
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("start log metric collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	c.TailLastUnixTime = 0
	c.DataChan = make(chan string, logMetricChanLength)
	go c.startHandleTailData()
	//go c.startFileHandlerCheck()
	reopenFlag := false
	destroyFlag := false
	for {
		select {
		case <-c.ReOpenHandlerChan:
			reopenFlag = true
		case <-c.DestroyChan:
			destroyFlag = true
		case line := <-c.TailSession.Lines:
			if line == nil {
				destroyFlag = true
				level.Error(monitorLogger).Log("log_metric -> tailSessionBreak", fmt.Sprintf("path:%s,serviceGroup:%s reason:%v ", c.Path, c.ServiceGroup, c.TailSession.Err()))
				break
			}
			c.DataChan <- line.Text
		}
		if reopenFlag || destroyFlag {
			break
		}
	}
	c.TailSession.Stop()
	//c.TailSession.Cleanup()
	c.TailDataCancelChan <- 1
	level.Info(monitorLogger).Log("log_metric -> startLogMetricMonitorNeObj__end", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
	if destroyFlag {
		level.Info(monitorLogger).Log("log_metric -> destroy", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
		return
	}
	if reopenFlag {
		level.Info(monitorLogger).Log("log_metric -> reopen", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
		time.Sleep(500 * time.Millisecond)
		go c.start()
	}
}

func (c *logMetricMonitorNeObj) tailLogFile(logPath string, destroyChan chan int) {
	level.Info(monitorLogger).Log("log_metric_start -> tailMultiLog", fmt.Sprintf("path:%s,serviceGroup:%s", logPath, c.ServiceGroup))
	logTailSession, err := tail.TailFile(logPath, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}, Poll: true})
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("start multi log metric collector fail, path: %s, error: %v", logPath, err))
		return
	}
	destroyFlag := false
	for {
		select {
		case <-destroyChan:
			destroyFlag = true
		case line := <-logTailSession.Lines:
			if line == nil {
				destroyFlag = true
				level.Error(monitorLogger).Log("log_metric -> tailMultiLog -> tailSessionBreak", fmt.Sprintf("path:%s,serviceGroup:%s reason:%v ", logPath, c.ServiceGroup, c.TailSession.Err()))
				break
			}
			//level.Info(monitorLogger).Log("log_metric -> get_new_line", fmt.Sprintf("path:%s,serviceGroup:%s,text:%s", c.Path, c.ServiceGroup, line.Text))
			c.DataChan <- line.Text
		}
		if destroyFlag {
			break
		}
	}
	logTailSession.Stop()
	level.Info(monitorLogger).Log("log_metric_end -> tailMultiLog", fmt.Sprintf("path:%s,serviceGroup:%s", logPath, c.ServiceGroup))
}

func (c *logMetricMonitorNeObj) startMultiPath() {
	level.Info(monitorLogger).Log("log_metric -> startLogMetricMonitorNeObj__startMultiPath", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
	pathList := listMatchLogPath(c.Path)
	if len(pathList) == 0 {
		level.Warn(monitorLogger).Log("log_metric -> startMultiPath_cannotMatchAnyFile", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
		return
	}
	c.MultiPathNum = len(pathList)
	go c.startHandleTailData()
	var destroyChanList []chan int
	for _, targetFilePath := range pathList {
		tmpDestroyChan := make(chan int, 1)
		go c.tailLogFile(targetFilePath, tmpDestroyChan)
		destroyChanList = append(destroyChanList, tmpDestroyChan)
	}
	<-c.DestroyChan
	for _, tmpDestroyChan := range destroyChanList {
		tmpDestroyChan <- 1
	}
	c.TailDataCancelChan <- 1
	level.Info(monitorLogger).Log("log_metric -> startLogMetricMonitorNeObj__endMultiPath", fmt.Sprintf("path:%s,serviceGroup:%s", c.Path, c.ServiceGroup))
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
				break
			} else {
				//level.Info(monitorLogger).Log(fmt.Sprintf("log_metric -> reopen_tail_with_time_check_ok,path:%s,fileLastTime:%d,tailLastTime:%d ", c.Path, fileLastTime, tailLastTime))
			}
		} else {
			//level.Error(monitorLogger).Log(fmt.Sprintf("log_metric -> check_file_handler_fail,path:%s,err:%s ", c.Path, err.Error()))
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
	c.DestroyChan = make(chan int, 1)
	c.TailDataCancelChan = make(chan int, 1)
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
	c.MetricGroupConfig = []*logMetricGroupNeObj{}
	for _, metricGroupObj := range input.MetricGroupConfig {
		initLogMetricGroupNeObj(metricGroupObj)
		c.MetricGroupConfig = append(c.MetricGroupConfig, metricGroupObj)
	}
	if strings.Contains(c.Path, "*") {
		c.DataChan = make(chan string, logMetricChanLength)
		go c.startMultiPath()
	} else {
		go c.start()
	}
}

// 把所有正则初始化
func (c *logMetricMonitorNeObj) update(input *logMetricMonitorNeObj) {
	level.Info(monitorLogger).Log("do updateLogMetricMonitorNeObj", c.Path)
	c.Lock.Lock()
	level.Info(monitorLogger).Log("start updateLogMetricMonitorNeObj", c.Path)
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
	newMetricGroupList := []*logMetricGroupNeObj{}
	for _, metricGroupObj := range input.MetricGroupConfig {
		for _, existMetricGroupObj := range c.MetricGroupConfig {
			if metricGroupObj.LogMetricGroup == existMetricGroupObj.LogMetricGroup {
				metricGroupObj.DataChannel = existMetricGroupObj.DataChannel
				break
			}
		}
		initLogMetricGroupNeObj(metricGroupObj)
		newMetricGroupList = append(newMetricGroupList, metricGroupObj)
	}
	c.MetricConfig = newMetricConfigList
	c.TargetEndpoint = input.TargetEndpoint
	c.ServiceGroup = input.ServiceGroup
	c.MetricGroupConfig = newMetricGroupList
	level.Info(monitorLogger).Log("updateLogMetricMonitorNeObj_MetricGroupConfig: ", fmt.Sprintf("len:%d", len(c.MetricGroupConfig)))
	c.Lock.Unlock()
	if strings.Contains(c.Path, "*") {
		if c.MultiPathNum > 0 {
			level.Info(monitorLogger).Log("start_updateLogMetricMonitorNeObj_destroy_*: ", c.Path)
			c.DestroyChan <- 1
			level.Info(monitorLogger).Log("end_updateLogMetricMonitorNeObj_destroy_*: ", c.Path)
		}
		time.Sleep(500 * time.Millisecond)
		go c.startMultiPath()
	}
}

func initLogMetricGroupNeObj(metricGroupObj *logMetricGroupNeObj) {
	if metricGroupObj.DataChannel == nil {
		metricGroupObj.DataChannel = make(chan map[string]interface{}, logMetricChanLength)
	}
	if metricGroupObj.LogType == "json" {
		tmpExp, tmpErr := PcreCompile(metricGroupObj.JsonRegular, 0)
		if tmpErr != nil {
			err := fmt.Errorf(tmpErr.Message)
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("logType:json regexpError:%s ", err.Error()))
			return
		}
		metricGroupObj.JsonRegexp = &tmpExp
	} else {
		for _, metricParamObj := range metricGroupObj.ParamList {
			if newRegExp, compileErr := PcreCompile(metricParamObj.Regular, 0); compileErr != nil {
				level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("logParam:%s regexpError:%s ", metricParamObj.Name, compileErr.Message))
			} else {
				metricParamObj.RegExp = &newRegExp
			}
		}
	}
	for _, metricParamObj := range metricGroupObj.ParamList {
		for _, stringMapObj := range metricParamObj.StringMap {
			if stringMapObj.RegEnable {
				if tmpExp, tmpErr := PcreCompile(stringMapObj.StringValue, 0); tmpErr != nil {
					stringMapObj.RegEnable = false
					level.Error(monitorLogger).Log("string map regexp format fail", tmpErr.Message)
				} else {
					stringMapObj.Regexp = &tmpExp
				}
			}
		}
	}
}

func initLogMetricNeObj(metricObj *logMetricNeObj) {
	var err error
	if metricObj.ValueRegular != "" {
		if newRegExp, compileErr := PcreCompile(metricObj.ValueRegular, 0); compileErr != nil {
			level.Error(monitorLogger).Log("newLogMetricMonitorNeObj", fmt.Sprintf("regexpError:%s ", compileErr.Message))
		} else {
			metricObj.RegExp = &newRegExp
		}
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
	level.Info(monitorLogger).Log("start_log_metric_destroy:", c.Path)
	c.Lock.Lock()
	c.DestroyChan <- 1
	c.JsonConfig = []*logMetricJsonNeObj{}
	c.MetricConfig = []*logMetricNeObj{}
	c.MetricGroupConfig = []*logMetricGroupNeObj{}
	c.Lock.Unlock()
	level.Info(monitorLogger).Log("done_log_metric_destroy:", c.Path)
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
	deletePathMap := make(map[string]int)
	for _, logMetricMonitorJob := range logMetricMonitorJobs {
		delFlag := true
		for _, paramObj := range param {
			if paramObj.Path == logMetricMonitorJob.Path && paramObj.ServiceGroup == logMetricMonitorJob.ServiceGroup {
				delFlag = false
				// update config
				logMetricMonitorJob.update(paramObj)
				break
			}
		}
		if delFlag {
			// delete config
			logMetricMonitorJob.destroy()
			deletePathMap[logMetricMonitorJob.Path] = 1
		} else {
			tmpLogMetricObjJobs = append(tmpLogMetricObjJobs, logMetricMonitorJob)
		}
	}
	if len(deletePathMap) > 0 && len(tmpLogMetricObjJobs) > 0 {
		for _, existJob := range tmpLogMetricObjJobs {
			if _, ok := deletePathMap[existJob.Path]; ok {
				if !strings.Contains(existJob.Path, "*") {
					existJob.ReOpenHandlerChan <- 1
				}
			}
		}
	}
	for _, paramObj := range param {
		addFlag := true
		for _, logMetricMonitorJob := range logMetricMonitorJobs {
			if logMetricMonitorJob.Path == paramObj.Path && logMetricMonitorJob.ServiceGroup == paramObj.ServiceGroup {
				addFlag = false
				break
			}
		}
		if !addFlag {
			continue
		}
		if !checkPathLegal(paramObj.Path) {
			level.Warn(monitorLogger).Log("log metric checkPathLegal:", fmt.Sprintf("serviceGroup:%s,path:%s", paramObj.ServiceGroup, paramObj.Path))
			continue
		}
		// add config
		newLogMetricObj := logMetricMonitorNeObj{Path: paramObj.Path, ServiceGroup: paramObj.ServiceGroup, Lock: new(sync.RWMutex)}
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
	nowTimeUnix := time.Now().Unix()
	existMetricMap := make(map[string]*logMetricDisplayObj)
	for _, displayObj := range logMetricMonitorMetrics {
		// 如果一个小时内都没再出现过该数据，就剔除掉它
		if displayObj.LastActiveTime > 0 && (displayObj.LastActiveTime+86400) < nowTimeUnix {
			continue
		}
		existMetricMap[displayObj.Id] = displayObj
	}
	//appendDisplayMap := make(map[string]int)
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
				//isMatchNewDataFlag := false
				for _, tmpMapData := range jsonDataList {
					// Get metric tags
					tmpTagsKey := []string{}
					for _, tmpJsonTagItem := range metricConfig.TagConfig {
						tmpTagsKey = append(tmpTagsKey, tmpJsonTagItem.Key)
					}
					_, _, _, tmpTagString := getLogMetricJsonMapTags(tmpMapData, tmpTagsKey)
					changedMapData := getLogMetricJsonMapValue(tmpMapData, metricConfig.StringMap)
					if metricValueFloat, b := changedMapData[metricConfig.Key]; b {
						//isMatchNewDataFlag = true
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
							valueExistObj.LastActiveTime = nowTimeUnix
						} else {
							valueCountMap[tmpMetricKey] = &logMetricDisplayObj{Id: tmpMetricKey, Metric: metricConfig.Metric, Path: lmObj.Path, Agg: metricConfig.AggType, TEndpoint: lmObj.TargetEndpoint, ServiceGroup: lmObj.ServiceGroup, TagsString: tmpTagString, Step: metricConfig.Step, ValueObj: logMetricValueObj{Sum: metricValueFloat, Max: metricValueFloat, Min: metricValueFloat, Count: 1}, LastActiveTime: nowTimeUnix}
						}
					}
				}
				//if !isMatchNewDataFlag {
				//	if metricConfig.AggType == "avg" {
				//		appendDisplayMap[fmt.Sprintf("%s^%s^sum", lmObj.Path, metricConfig.Metric)] = 1
				//		appendDisplayMap[fmt.Sprintf("%s^%s^count", lmObj.Path, metricConfig.Metric)] = 1
				//	}
				//	appendDisplayMap[fmt.Sprintf("%s^%s^%s", lmObj.Path, metricConfig.Metric, metricConfig.AggType)] = 1
				//}
			}
		}
		for _, metricObj := range lmObj.MetricConfig {
			dataLength := len(metricObj.DataChannel)
			if dataLength == 0 {
				//if metricObj.AggType == "avg" {
				//	appendDisplayMap[fmt.Sprintf("%s^%s^sum", lmObj.Path, metricObj.Metric)] = 1
				//	appendDisplayMap[fmt.Sprintf("%s^%s^count", lmObj.Path, metricObj.Metric)] = 1
				//}
				//appendDisplayMap[fmt.Sprintf("%s^%s^%s", lmObj.Path, metricObj.Metric, metricObj.AggType)] = 1
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
					valueExistObj.LastActiveTime = nowTimeUnix
				} else {
					valueCountMap[tmpMetricKey] = &logMetricDisplayObj{Id: tmpMetricKey, Metric: metricObj.Metric, Path: lmObj.Path, Agg: metricObj.AggType, TEndpoint: lmObj.TargetEndpoint, ServiceGroup: lmObj.ServiceGroup, TagsString: tmpTagString, Step: metricObj.Step, ValueObj: logMetricValueObj{Sum: metricValueFloat, Max: metricValueFloat, Min: metricValueFloat, Count: 1}, LastActiveTime: nowTimeUnix}
				}
			}
			//valueCountMap[tmpMetricKey] = &tmpMetricObj
		}
		for _, metricGroupObj := range lmObj.MetricGroupConfig {
			// pull channel data list
			matchDataList := []map[string]interface{}{}
			dataLength := len(metricGroupObj.DataChannel)
			for i := 0; i < dataLength; i++ {
				tmpMapData := <-metricGroupObj.DataChannel
				matchDataList = append(matchDataList, tmpMapData)
			}
			calcMetricGroupFunc(lmObj.Path, lmObj.TargetEndpoint, lmObj.ServiceGroup, matchDataList, metricGroupObj.MetricConfig, valueCountMap)
		}
	}
	// appendDisplayMap是当数据上一次采集出现，但此次采集不出现，尝试把数据补个默认点上去，现在默认值是0
	for k, v := range existMetricMap {
		//level.Info(monitorLogger).Log("existMetricMap -> ", fmt.Sprintf("k:%s", k))
		appendFlag := true
		for id, _ := range valueCountMap {
			if id == k {
				appendFlag = false
				break
			}
		}
		//level.Info(monitorLogger).Log("existMetricMap append -> ", fmt.Sprintf("k:%s", k))
		if appendFlag {
			if v.ByAvgFlag {
				continue
			}
			if v.Display {
				v.ValueObj = logMetricValueObj{Sum: 0, Count: 0, Max: 0, Min: 0}
			}
			valueCountMap[k] = v
		}
	}
	//if len(appendDisplayMap) > 0 {
	//	for k, v := range existMetricMap {
	//		tmpKey := fmt.Sprintf("%s^%s^%s^%s", v.Path, v.Metric, v.Agg, v.TagsString)
	//		if _, b := appendDisplayMap[tmpKey]; b {
	//			if v.Display {
	//				v.ValueObj = logMetricValueObj{Sum: 0, Count: 0, Max: 0, Min: 0}
	//			}
	//			valueCountMap[k] = v
	//		}
	//	}
	//}
	tmpLogMetricMetrics := buildLogMetricDisplayMetrics(valueCountMap, existMetricMap)
	logMetricHttpLock.RUnlock()
	logMetricMonitorMetricLock.Lock()
	logMetricMonitorMetrics = tmpLogMetricMetrics
	logMetricMonitorMetricLock.Unlock()
}

func transMetricGroupData(input interface{}, stringMap []*logMetricStringMapNeObj) interface{} {
	if len(stringMap) == 0 {
		return input
	}
	output := input
	inputString := transInterfaceValueToString(input)
	for _, v := range stringMap {
		if !v.RegEnable {
			if v.StringValue == inputString {
				output = v.TargetStringValue
				break
			}
			continue
		}
		if mat := v.Regexp.MatcherString(inputString, 0); mat != nil {
			if mat.Matches() {
				output = v.TargetStringValue
				break
			}
		}
	}
	return output
}

func calcMetricGroupFunc(logPath, endpoint, serviceGroup string, dataList []map[string]interface{}, metricConfigList []*logMetricNeObj, valueCountMap map[string]*logMetricDisplayObj) {
	nowTimeUnix := time.Now().Unix()
	for _, metricConfig := range metricConfigList {
		for _, tmpMapData := range dataList {
			// Get metric tags
			tagsNameList := []string{}
			for _, tmpJsonTagItem := range metricConfig.TagConfig {
				tagsNameList = append(tagsNameList, tmpJsonTagItem.LogParamName)
			}
			tmpCode, tmpRetCode, _, tmpTagString := getLogMetricJsonMapTags(tmpMapData, tagsNameList)
			//level.Info(monitorLogger).Log("log_metric_group -> ", fmt.Sprintf("code:%s, retCode:%s, otherTag:%s, tagString:%s", tmpCode, tmpRetCode, tmpOtherTag, tmpTagString))
			// 根据值类型尝试转换成数值
			metricValueMap := getLogMetricJsonMapValue(tmpMapData, metricConfig.StringMap)
			if metricValueFloat, b := metricValueMap[metricConfig.LogParamName]; b {
				tmpMetricKey := fmt.Sprintf("%s^%s^%s^%s", logPath, metricConfig.Metric, metricConfig.AggType, tmpTagString)
				if valueExistObj, keyExist := valueCountMap[tmpMetricKey]; keyExist {
					valueExistObj.ValueObj.Sum += metricValueFloat
					valueExistObj.ValueObj.Count++
					if valueExistObj.ValueObj.Max < metricValueFloat {
						valueExistObj.ValueObj.Max = metricValueFloat
					}
					if valueExistObj.ValueObj.Min > metricValueFloat {
						valueExistObj.ValueObj.Min = metricValueFloat
					}
					valueExistObj.LastActiveTime = nowTimeUnix
				} else {
					valueCountMap[tmpMetricKey] = &logMetricDisplayObj{Id: tmpMetricKey, Metric: metricConfig.Metric, Path: logPath, Agg: metricConfig.AggType, TEndpoint: endpoint, ServiceGroup: serviceGroup, TagsString: tmpTagString, Step: metricConfig.Step, ValueObj: logMetricValueObj{Sum: metricValueFloat, Max: metricValueFloat, Min: metricValueFloat, Count: 1}, Code: tmpCode, RetCode: tmpRetCode, LastActiveTime: nowTimeUnix}
				}
			}
		}
	}
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
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, &logMetricDisplayObj{Id: fmt.Sprintf("%s^%s^%s^%s", v.Path, v.Metric, "avg", v.TagsString), ByAvgFlag: false, Metric: v.Metric, Path: v.Path, Agg: "avg", TEndpoint: v.TEndpoint, ServiceGroup: v.ServiceGroup, Tags: v.Tags, TagsString: v.TagsString, Value: v.Value, Step: v.Step, Display: v.Display, UpdateTime: v.UpdateTime, Code: v.Code, RetCode: v.RetCode, LastActiveTime: v.LastActiveTime})
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, &logMetricDisplayObj{Id: fmt.Sprintf("%s^%s^%s^%s", v.Path, v.Metric, "sum", v.TagsString), ByAvgFlag: true, Metric: v.Metric, Path: v.Path, Agg: "sum", TEndpoint: v.TEndpoint, ServiceGroup: v.ServiceGroup, Tags: v.Tags, TagsString: v.TagsString, Value: v.ValueObj.Sum, Step: v.Step, Display: v.Display, UpdateTime: v.UpdateTime, Code: v.Code, RetCode: v.RetCode, LastActiveTime: v.LastActiveTime})
			tmpLogMetricMetrics = append(tmpLogMetricMetrics, &logMetricDisplayObj{Id: fmt.Sprintf("%s^%s^%s^%s", v.Path, v.Metric, "count", v.TagsString), ByAvgFlag: true, Metric: v.Metric, Path: v.Path, Agg: "count", TEndpoint: v.TEndpoint, ServiceGroup: v.ServiceGroup, Tags: v.Tags, TagsString: v.TagsString, Value: v.ValueObj.Count, Step: v.Step, Display: v.Display, UpdateTime: v.UpdateTime, Code: v.Code, RetCode: v.RetCode, LastActiveTime: v.LastActiveTime})
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

func transInterfaceValueToString(input interface{}) (output string) {
	if input == nil {
		return
	}
	rt := reflect.TypeOf(input)
	switch rt.String() {
	case "string":
		output = fmt.Sprintf("%s", input)
		break
	case "float64":
		output = fmt.Sprintf("%.0f", input.(float64))
		break
	case "int64":
		output = fmt.Sprintf("%d", input.(int64))
		break
	case "bool":
		output = fmt.Sprintf("%t", input.(bool))
		break
	default:
		output = fmt.Sprintf("%v", input)
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

func getLogMetricJsonMapTags(input map[string]interface{}, tagKey []string) (code, retCode, otherTag, tagString string) {
	tagString = ""
	for _, v := range tagKey {
		if tmpTagValue, b := input[v]; b {
			tmpValueString := fmt.Sprintf("%s", tmpTagValue)
			rt := reflect.TypeOf(tmpTagValue)
			if rt.String() == "float64" {
				tmpValueString = fmt.Sprintf("%.0f", tmpTagValue.(float64))
			}
			tagString += fmt.Sprintf("%s=%s,", v, tmpValueString)
			if v == "code" {
				code = fmt.Sprintf("%s", tmpValueString)
			} else if v == "retcode" {
				retCode = fmt.Sprintf("%s", tmpValueString)
			} else {
				otherTag += fmt.Sprintf("%s=%s,", v, tmpValueString)
			}
		}
	}
	if tagString != "" {
		tagString = tagString[:len(tagString)-1]
	}
	if otherTag != "" {
		otherTag = otherTag[:len(otherTag)-1]
	}
	return
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

func listMatchLogPath(inputPath string) (result []string) {
	var dirPath, fileName string
	if lastPathIndex := strings.LastIndex(inputPath, "/"); lastPathIndex >= 0 {
		dirPath = inputPath[:lastPathIndex+1]
		fileName = inputPath[lastPathIndex+1:]
	}
	if fileName == "" {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("log path illgal : %s ", inputPath))
		return
	}
	fileName = strings.ReplaceAll(fileName, ".", "\\.")
	fileName = strings.ReplaceAll(fileName, "*", ".*")
	cmdString := fmt.Sprintf("ls %s |grep \"^%s$\"", dirPath, fileName)
	cmd := exec.Command("bash", "-c", cmdString)
	b, err := cmd.Output()
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("list log path:%s fail : %v ", cmdString, err))
		return
	}
	for _, row := range strings.Split(string(b), "\n") {
		if row != "" {
			result = append(result, dirPath+row)
		}
	}
	return
}

func checkPathLegal(path string) bool {
	if strings.HasSuffix(path, "/") {
		return false
	}
	re := regexp.MustCompile(`^\/([\w|\.|\-|\*]+\/?)+$`)
	if re.MatchString(path) {
		return true
	}
	return false
}
