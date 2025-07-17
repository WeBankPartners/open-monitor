package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
)

type logMonitorCollector struct {
	logMonitor *prometheus.Desc
	logger     log.Logger
}

const (
	logMonitorCollectorName = "log_monitor"
	logMonitorFilePath      = "data/log_monitor_cache.json"
)

var (
	logKeywordCollectorJobs []*logKeywordCollector
	logKeywordChanLength    = 100000
)

func init() {
	registerCollector("log_monitor", defaultEnabled, NewLogMonitorCollector)
}

func NewLogMonitorCollector(logger log.Logger) (Collector, error) {
	return &logMonitorCollector{
		logMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, logMonitorCollectorName, "count_total"),
			"Count the keyword from log file.",
			[]string{"file", "keyword", "t_guid"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *logMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _, v := range logKeywordCollectorJobs {
		for _, vv := range v.get() {
			ch <- prometheus.MustNewConstMetric(c.logMonitor,
				prometheus.GaugeValue,
				vv.Value, vv.Path, vv.Keyword, vv.TargetEndpoint)
		}
	}
	return nil
}

type logKeywordMetricObj struct {
	Path           string
	Keyword        string
	TargetEndpoint string
	Value          float64
}

type logRowsHttpDto struct {
	Path      string  `json:"path"`
	Keyword   string  `json:"keyword"`
	Value     float64 `json:"value"`
	LastValue float64 `json:"last_value"`
}

type logKeywordFetchObj struct {
	Index   float64 `json:"index"`
	Content string  `json:"content"`
}

type logKeywordObj struct {
	Keyword        string
	RegExp         *Regexp
	Count          float64
	LastMatchRow   string
	TargetEndpoint string
}

type logKeywordCollector struct {
	Path               string
	Rule               []*logKeywordObj
	TailSession        *tail.Tail
	Lock               *sync.RWMutex
	DataChan           chan string
	ReOpenHandlerChan  chan int      `json:"-"`
	TailTimeLock       *sync.RWMutex `json:"-"`
	TailLastUnixTime   int64         `json:"-"`
	DestroyChan        chan int      `json:"-"`
	TailDataCancelChan chan int      `json:"-"`
}

func (c *logKeywordCollector) update(rule []*logKeywordObj) {
	c.Lock.Lock()
	for _, inputRule := range rule {
		for _, existRule := range c.Rule {
			if inputRule.Keyword == existRule.Keyword {
				inputRule.Count = existRule.Count
				inputRule.LastMatchRow = existRule.LastMatchRow
				break
			}
		}
	}
	c.Rule = rule
	c.Lock.Unlock()
}

func (c *logKeywordCollector) startHandleTailData() {
	for {
		var lineText string
		select {
		case lineText = <-c.DataChan:
		case <-c.TailDataCancelChan:
			return
		}
		//lineText := <-c.DataChan
		c.Lock.Lock()
		for _, v := range c.Rule {
			if v.RegExp != nil {
				if pcreMatch(v.RegExp, lineText) {
					v.Count++
					v.LastMatchRow = lineText
				}
				//if ok, _ := v.RegExp.MatchString(lineText); ok {
				//	v.Count++
				//	v.LastMatchRow = lineText
				//}
			} else {
				if strings.Contains(lineText, v.Keyword) {
					v.Count++
					v.LastMatchRow = lineText
				}
			}
		}
		c.Lock.Unlock()
	}
}

func (c *logKeywordCollector) init() {
	c.TailTimeLock = new(sync.RWMutex)
	c.ReOpenHandlerChan = make(chan int, 1)
	c.TailLastUnixTime = 0
	c.DestroyChan = make(chan int, 1)
	c.TailDataCancelChan = make(chan int, 1)
	go c.start()
}

func (c *logKeywordCollector) start() {
	level.Info(monitorLogger).Log("log_keyword -> logKeywordCollectorStart", c.Path)
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true, Poll: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}})
	if err != nil {
		level.Error(monitorLogger).Log("error", fmt.Sprintf("start log keyword collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	c.TailLastUnixTime = 0
	c.DataChan = make(chan string, logKeywordChanLength)
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
				level.Error(monitorLogger).Log("log_keyword -> tailSessionBreak", fmt.Sprintf("path:%s reason:%v ", c.Path, c.TailSession.Err()))
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
	level.Info(monitorLogger).Log("log_keyword -> startLogMetricMonitorNeObj__end", c.Path)
	if destroyFlag {
		return
	}
	//time.Sleep(60 * time.Second)
	//go c.start()
}

func (c *logKeywordCollector) startFileHandlerCheck() {
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
				level.Info(monitorLogger).Log(fmt.Sprintf("log_keyword -> reopen_tail_with_time_check_fail,path:%s,fileLastTime:%d,tailLastTime:%d ", c.Path, fileLastTime, tailLastTime))
				break
			} else {
				//level.Info(monitorLogger).Log(fmt.Sprintf("log_keyword -> reopen_tail_with_time_check_ok,path:%s,fileLastTime:%d,tailLastTime:%d ", c.Path, fileLastTime, tailLastTime))
			}
		} else {
			//level.Error(monitorLogger).Log(fmt.Sprintf("log_keyword -> check_file_handler_fail,path:%s,err:%s ", c.Path, err.Error()))
		}
	}
}

func (c *logKeywordCollector) destroy() {
	level.Info(monitorLogger).Log("start_log_keyword_destroy:", c.Path)
	c.Lock.Lock()
	c.DestroyChan <- 1
	c.Rule = []*logKeywordObj{}
	c.Lock.Unlock()
	level.Info(monitorLogger).Log("done_log_keyword_destroy:", c.Path)
}

func (c *logKeywordCollector) get() (data []*logKeywordMetricObj) {
	for _, v := range c.Rule {
		data = append(data, &logKeywordMetricObj{Path: c.Path, Keyword: v.Keyword, Value: v.Count, TargetEndpoint: v.TargetEndpoint})
	}
	return data
}

func (c *logKeywordCollector) getRows(keyword string) (data []*logKeywordFetchObj) {
	data = []*logKeywordFetchObj{}
	for _, v := range c.Rule {
		if v.Keyword == keyword {
			//level.Info(monitorLogger).Log("getRows:", keyword, " count:", v.Count)
			data = append(data, &logKeywordFetchObj{Content: v.LastMatchRow, Index: v.Count})
			break
		}
	}
	return data
}

type logKeywordHttpRuleObj struct {
	RegularEnable  bool    `json:"regular_enable"`
	Keyword        string  `json:"keyword"`
	Count          float64 `json:"count"`
	TargetEndpoint string  `json:"target_endpoint"`
}

type logKeywordHttpDto struct {
	Path     string                   `json:"path"`
	Keywords []*logKeywordHttpRuleObj `json:"keywords"`
}

type logKeywordHttpResult struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    []*logKeywordFetchObj `json:"data"`
}

func LogKeywordHttpHandle(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func(returnErr error) {
		responseObj := logKeywordHttpResult{Status: "OK", Message: "success"}
		if returnErr != nil {
			returnErr = fmt.Errorf("Handel log keyword monitor http request fail,%s ", returnErr.Error())
			responseObj = logKeywordHttpResult{Status: "ERROR", Message: returnErr.Error()}
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
	level.Info(monitorLogger).Log("logKeywordConfig", string(requestParamBuff))
	err = logKeywordHttpAction(requestParamBuff)
	if err == nil {
		logKeywordSaveConfig(requestParamBuff)
	}
}

func logKeywordHttpAction(requestParamBuff []byte) (err error) {
	var param []*logKeywordHttpDto
	err = json.Unmarshal(requestParamBuff, &param)
	if err != nil {
		return
	}
	var newCollectorList []*logKeywordCollector
	var removePathList []string
	for _, existCollector := range logKeywordCollectorJobs {
		exist := false
		for _, inputParam := range param {
			if existCollector.Path == inputParam.Path {
				// Update collector
				exist = true
				var tmpKeywordList []*logKeywordObj
				for _, inputKeyword := range inputParam.Keywords {
					if inputKeyword.RegularEnable {
						tmpRegExp, tmpRegErr := PcreCompile(inputKeyword.Keyword, 0)
						if tmpRegErr != nil {
							err = fmt.Errorf("path:%s pcre regexp compile %s fail:%s", inputParam.Path, inputKeyword.Keyword, tmpRegErr.Message)
							continue
						}
						tmpKeywordList = append(tmpKeywordList, &logKeywordObj{Keyword: inputKeyword.Keyword, RegExp: &tmpRegExp, TargetEndpoint: inputKeyword.TargetEndpoint})
					} else {
						tmpKeywordList = append(tmpKeywordList, &logKeywordObj{Keyword: inputKeyword.Keyword, TargetEndpoint: inputKeyword.TargetEndpoint})
					}
				}
				existCollector.update(tmpKeywordList)
			}
		}
		if !exist {
			// Remove collector
			existCollector.destroy()
			removePathList = append(removePathList, existCollector.Path)
		}
	}
	if err != nil {
		return
	}
	if len(removePathList) > 0 {
		for _, collector := range logKeywordCollectorJobs {
			deleteFlag := false
			for _, v := range removePathList {
				if collector.Path == v {
					deleteFlag = true
					break
				}
			}
			if !deleteFlag {
				newCollectorList = append(newCollectorList, collector)
			}
		}
		logKeywordCollectorJobs = newCollectorList
	}
	for _, inputParam := range param {
		exist := false
		for _, existCollector := range logKeywordCollectorJobs {
			if inputParam.Path == existCollector.Path {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		if !checkPathLegal(inputParam.Path) {
			level.Warn(monitorLogger).Log("log keyword checkPathLegal:", fmt.Sprintf("serviceGroup:%s,path:%s", inputParam.Keywords, inputParam.Path))
			continue
		}
		// Add collector
		newCollector := logKeywordCollector{Path: inputParam.Path}
		newCollector.Lock = new(sync.RWMutex)
		var tmpKeywordList []*logKeywordObj
		for _, inputKeyword := range inputParam.Keywords {
			if inputKeyword.RegularEnable {
				tmpRegExp, tmpRegErr := PcreCompile(inputKeyword.Keyword, 0)
				if tmpRegErr != nil {
					err = fmt.Errorf("path:%s regexp2 regexp compile %s fail:%s", inputParam.Path, inputKeyword.Keyword, tmpRegErr.Message)
					continue
				}
				tmpKeywordList = append(tmpKeywordList, &logKeywordObj{Keyword: inputKeyword.Keyword, RegExp: &tmpRegExp, Count: inputKeyword.Count, TargetEndpoint: inputKeyword.TargetEndpoint})
			} else {
				tmpKeywordList = append(tmpKeywordList, &logKeywordObj{Keyword: inputKeyword.Keyword, Count: inputKeyword.Count, TargetEndpoint: inputKeyword.TargetEndpoint})
			}
		}
		newCollector.Rule = tmpKeywordList
		logKeywordCollectorJobs = append(logKeywordCollectorJobs, &newCollector)
		newCollector.init()
	}
	return err
}

func logKeywordSaveConfig(requestParamBuff []byte) {
	err := ioutil.WriteFile(logMonitorFilePath, requestParamBuff, 0644)
	if err != nil {
		level.Error(monitorLogger).Log("logKeywordSaveConfig", err.Error())
	} else {
		level.Info(monitorLogger).Log("logKeywordSaveConfig", "success")
	}
}

func LogKeyWordLoadConfig() {
	b, err := ioutil.ReadFile(logMonitorFilePath)
	if err != nil {
		level.Warn(monitorLogger).Log("logKeywordLoadConfig", err.Error())
	} else {
		err = logKeywordHttpAction(b)
		if err != nil {
			level.Error(monitorLogger).Log("logKeywordLoadConfigAction", err.Error())
		} else {
			level.Info(monitorLogger).Log("logKeywordLoadConfig", "success")
		}
	}
}

func LogMonitorRowsHttpHandle(w http.ResponseWriter, r *http.Request) {
	var result logKeywordHttpResult
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		d, _ := json.Marshal(result)
		w.Write(d)
	}()
	buff, err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor rows http request fail,read body error: %v ", err)
		level.Error(monitorLogger).Log("msg", errorMsg)
		result.Status = "error"
		result.Message = errorMsg
		return
	}
	var param logRowsHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor rows http request fail,json unmarshal error: %v ", err)
		level.Error(monitorLogger).Log("msg", errorMsg)
		result.Status = "error"
		result.Message = errorMsg
		return
	}
	for _, v := range logKeywordCollectorJobs {
		if v.Path == param.Path {
			result.Data = v.getRows(param.Keyword)
			break
		}
	}
	result.Status = "ok"
	result.Message = "success"
}

func pcreMatch(re *Regexp, lineText string) (match bool) {
	if re == nil {
		return
	}
	mat := re.MatcherString(lineText, 0)
	if mat != nil {
		match = mat.Matches()
	}
	return
}
