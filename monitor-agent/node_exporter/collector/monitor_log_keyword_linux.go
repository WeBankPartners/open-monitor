package collector

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type logMonitorCollector struct {
	logMonitor *prometheus.Desc
	logger     log.Logger
}

const (
	logMonitorCollectorName = "log_monitor"
	logMonitorFilePath      = "data/log_monitor_cache.data"
)

func init() {
	registerCollector("log_monitor", defaultEnabled, NewLogMonitorCollector)
}

func NewLogMonitorCollector(logger log.Logger) (Collector, error) {
	return &logMonitorCollector{
		logMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, logMonitorCollectorName, "count_total"),
			"Count the keyword from log file.",
			[]string{"file", "keyword"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *logMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _, v := range logCollectorJobs {
		for _, vv := range v.get() {
			ch <- prometheus.MustNewConstMetric(c.logMonitor,
				prometheus.GaugeValue,
				vv.Value, vv.Path, vv.Keyword)
		}
	}
	return nil
}

type logKeywordFetchObj struct {
	Index   float64 `json:"index"`
	Content string  `json:"content"`
}

type logKeywordObj struct {
	Keyword  string
	RegExp   *pcre.Regexp
	Count    float64
	FetchRow []logKeywordFetchObj
}

type logCollectorObj struct {
	Path        string
	Rule        []*logKeywordObj
	TailSession *tail.Tail
	Lock        *sync.RWMutex
}

type logMetricObj struct {
	Path    string
	Keyword string
	Value   float64
}

type logHttpDto struct {
	Path     string   `json:"path"`
	Keywords []string `json:"keywords"`
}

type logRowsHttpDto struct {
	Path      string  `json:"path"`
	Keyword   string  `json:"keyword"`
	Value     float64 `json:"value"`
	LastValue float64 `json:"last_value"`
}

type logRowsHttpResult struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    []logKeywordFetchObj `json:"data"`
}

var logCollectorJobs []*logCollectorObj

func (c *logCollectorObj) update(rule []*logKeywordObj) {
	c.Lock.Lock()
	for _, v := range rule {
		for _, vv := range c.Rule {
			if v.Keyword == vv.Keyword {
				v.Count = vv.Count
			}
		}
	}
	c.Rule = rule
	c.Lock.Unlock()
}

func (c *logCollectorObj) start() {
	level.Info(monitorLogger).Log("start", c.Path)
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("start log collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	firstFlag := true
	timeNow := time.Now()
	for line := range c.TailSession.Lines {
		if firstFlag {
			if time.Now().Sub(timeNow).Seconds() >= 5 {
				firstFlag = false
			} else {
				continue
			}
		}
		c.Lock.Lock()
		for _, v := range c.Rule {
			level.Info(monitorLogger).Log("rule", fmt.Sprintf("k:%s regExp:%v ", v.Keyword, v.RegExp))
			if v.RegExp != nil {
				if len(v.RegExp.FindIndex([]byte(line.Text), 0)) > 0 {
					v.Count++
					v.FetchRow = append(v.FetchRow, logKeywordFetchObj{Content: line.Text, Index: v.Count})
				}
			} else {
				if strings.Contains(line.Text, v.Keyword) {
					v.Count++
					v.FetchRow = append(v.FetchRow, logKeywordFetchObj{Content: line.Text, Index: v.Count})
				}
			}
		}
		c.Lock.Unlock()
	}
}

func (c *logCollectorObj) destroy() {
	c.TailSession.Stop()
	c.Rule = []*logKeywordObj{}
}

func (c *logCollectorObj) get() []logMetricObj {
	var data []logMetricObj
	c.Lock.RLock()
	for _, v := range c.Rule {
		data = append(data, logMetricObj{Path: c.Path, Keyword: v.Keyword, Value: v.Count})
	}
	c.Lock.RUnlock()
	return data
}

func (c *logCollectorObj) getRows(keyword string, value, lastValue float64) []logKeywordFetchObj {
	var data []logKeywordFetchObj
	//nowTimestamp := time.Now().Unix()
	c.Lock.RLock()
	for _, v := range c.Rule {
		if v.Keyword == keyword {
			if len(v.FetchRow) == 0 {
				continue
			}
			data = append(data, logKeywordFetchObj{Content: v.FetchRow[len(v.FetchRow)-1].Content})
		}
	}
	c.Lock.RUnlock()
	return data
}

type logCollectorStore struct {
	Data []*logCollectorStoreObj
}

type logCollectorStoreObj struct {
	Path string
	Rule []*logKeywordObj
}

func (c *logCollectorStore) Save() {
	c.Data = []*logCollectorStoreObj{}
	for _, v := range logCollectorJobs {
		lmo := v.get()
		if len(lmo) == 0 {
			continue
		}
		tmpLogStoreObj := logCollectorStoreObj{Path: lmo[0].Path}
		tmpRule := []*logKeywordObj{}
		for _, vv := range lmo {
			tmpRule = append(tmpRule, &logKeywordObj{Keyword: vv.Keyword, Count: vv.Value})
		}
		tmpLogStoreObj.Rule = tmpRule
		c.Data = append(c.Data, &tmpLogStoreObj)
	}
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(c.Data)
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("gob encode log monitor error : %v ", err))
	} else {
		ioutil.WriteFile(logMonitorFilePath, tmpBuffer.Bytes(), 0644)
		level.Info(monitorLogger).Log("msg", fmt.Sprintf("write %s succeed ", logMonitorFilePath))
	}
}

func (c *logCollectorStore) Load() {
	file, err := os.Open(logMonitorFilePath)
	if err != nil {
		level.Info(monitorLogger).Log("msg", fmt.Sprintf("read %s file error %v ", logMonitorFilePath, err))
	} else {
		dec := gob.NewDecoder(file)
		err = dec.Decode(&c.Data)
		if err != nil {
			level.Error(monitorLogger).Log("msg", fmt.Sprintf("gob decode %s error %v ", logMonitorFilePath, err))
		} else {
			level.Info(monitorLogger).Log("msg", fmt.Sprintf("load %s file succeed ", logMonitorFilePath))
		}
	}
	for _, v := range c.Data {
		lco := logCollectorObj{Path: v.Path}
		lco.Lock = new(sync.RWMutex)
		lco.Rule = v.Rule
		logCollectorJobs = append(logCollectorJobs, &lco)
		go lco.start()
	}
}

var LogCollectorStore logCollectorStore

func LogMonitorHttpHandle(w http.ResponseWriter, r *http.Request) {
	buff, err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor http request fail,read body error: %v ", err)
		level.Error(monitorLogger).Log("msg", errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	level.Info(monitorLogger).Log("config", string(buff))
	var param []logHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor http request fail,json unmarshal error: %v ", err)
		level.Error(monitorLogger).Log("msg", errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	for _, v := range logCollectorJobs {
		exist := false
		for _, vv := range param {
			if v.Path == vv.Path {
				exist = true
				var tmp []*logKeywordObj
				for _, vvv := range vv.Keywords {
					//tmpRegExp,_ := regexp.Compile("("+vvv+")")
					tmpRegExp, tmpRegErr := pcre.Compile(vvv, 0)
					if tmpRegErr != nil {
						level.Error(monitorLogger).Log("reg compile error", fmt.Sprintf("%v", tmpRegErr))
					}
					tmp = append(tmp, &logKeywordObj{Keyword: vvv, RegExp: &tmpRegExp})
				}
				v.update(tmp)
			}
		}
		if !exist {
			v.destroy()
		}
	}
	for _, v := range param {
		exist := false
		for _, vv := range logCollectorJobs {
			if v.Path == vv.Path {
				exist = true
				break
			}
		}
		if !exist {
			lco := logCollectorObj{Path: v.Path}
			lco.Lock = new(sync.RWMutex)
			var tmp []*logKeywordObj
			for _, vv := range v.Keywords {
				//tmpRegExp,_ := regexp.Compile("("+vv+")")
				tmpRegExp, tmpRegErr := pcre.Compile(vv, 0)
				if tmpRegErr != nil {
					level.Error(monitorLogger).Log("reg compile error", fmt.Sprintf("%v", tmpRegErr))
				}
				tmp = append(tmp, &logKeywordObj{Keyword: vv, RegExp: &tmpRegExp, Count: 0})
			}
			lco.Rule = tmp
			logCollectorJobs = append(logCollectorJobs, &lco)
			go lco.start()
		}
	}
	w.Write([]byte("success"))
}

func LogMonitorRowsHttpHandle(w http.ResponseWriter, r *http.Request) {
	var result logRowsHttpResult
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
	level.Info(monitorLogger).Log("getRows", string(buff))
	var param logRowsHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor rows http request fail,json unmarshal error: %v ", err)
		level.Error(monitorLogger).Log("msg", errorMsg)
		result.Status = "error"
		result.Message = errorMsg
		return
	}
	for _, v := range logCollectorJobs {
		if v.Path == param.Path {
			result.Data = v.getRows(param.Keyword, param.Value, param.LastValue)
		}
	}
	result.Status = "ok"
	result.Message = "success"
}
