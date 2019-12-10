package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"github.com/hpcloud/tail"
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/prometheus/common/log"
	"encoding/json"
	"bytes"
	"encoding/gob"
	"os"
)

type logMonitorCollector struct {
	logMonitor  *prometheus.Desc
}

const (
	logMonitorCollectorName = "log_monitor"
	logMonitorFilePath = "data/log_monitor_cache.data"
)

func init() {
	registerCollector("log_monitor", defaultEnabled, NewLogMonitorCollector)
	LogCollectorStore.Load()
}

func NewLogMonitorCollector() (Collector, error) {
	return &logMonitorCollector{
		logMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, logMonitorCollectorName, "count_total"),
			"Count the keyword from log file.",
			[]string{"file", "keyword"}, nil,
		),
	}, nil
}

func (c *logMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _,v := range logCollectorJobs {
		for _,vv := range v.get() {
			ch <- prometheus.MustNewConstMetric(c.logMonitor,
				prometheus.GaugeValue,
				vv.Value, vv.Path,vv.Keyword)
		}
	}
	return nil
}

type logKeywordObj struct {
	Keyword  string
	Count  float64
}

type logCollectorObj struct {
	Path  string
	Rule  []*logKeywordObj
	TailSession  *tail.Tail
	Lock  *sync.RWMutex
}

type logMetricObj struct {
	Path  string
	Keyword  string
	Value  float64
}

type logHttpDto struct {
	Path  string  `json:"path"`
	Keywords  []string  `json:"keywords"`
}

var logCollectorJobs []*logCollectorObj

func (c *logCollectorObj) update(rule []*logKeywordObj)  {
	c.Lock.Lock()
	for _,v := range rule {
		for _,vv := range c.Rule {
			if v.Keyword == vv.Keyword {
				v.Count = vv.Count
			}
		}
	}
	c.Rule = rule
	c.Lock.Unlock()
}

func (c *logCollectorObj) start() {
	var err error
	c.TailSession,err = tail.TailFile(c.Path, tail.Config{Follow:true})
	if err != nil {
		log.Errorf("start log collector fail, path: %s, error: %v", c.Path, err)
		return
	}
	for line := range c.TailSession.Lines {
		c.Lock.Lock()
		for _,v := range c.Rule {
			if strings.Contains(line.Text, v.Keyword) {
				v.Count++
			}
		}
		c.Lock.Unlock()
	}
}

func (c *logCollectorObj) destroy()  {
	c.TailSession.Dead()
	c.Rule = []*logKeywordObj{}
}

func (c *logCollectorObj) get() []logMetricObj {
	var data []logMetricObj
	c.Lock.RLock()
	for _,v := range c.Rule {
		data = append(data, logMetricObj{Path:c.Path,Keyword:v.Keyword,Value:v.Count})
	}
	c.Lock.RUnlock()
	return data
}

type logCollectorStrore struct {
	Data  []*logCollectorStoreObj
}

type logCollectorStoreObj struct {
	Path  string
	Rule  []*logKeywordObj
}

func (c *logCollectorStrore) Save()  {
	c.Data = []*logCollectorStoreObj{}
	for _,v := range logCollectorJobs {
		lmo := v.get()
		if len(lmo) == 0 {
			continue
		}
		tmpLogStoreObj := logCollectorStoreObj{Path:lmo[0].Path}
		tmpRule := []*logKeywordObj{}
		for _,vv := range lmo {
			tmpRule = append(tmpRule, &logKeywordObj{Keyword:vv.Keyword, Count:vv.Value})
		}
		tmpLogStoreObj.Rule = tmpRule
		c.Data = append(c.Data, &tmpLogStoreObj)
	}
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(c.Data)
	if err != nil {
		log.Errorf("gob encode log monitor error : %v \n", err)
	}else{
		ioutil.WriteFile(logMonitorFilePath, tmpBuffer.Bytes(), 0644)
		log.Infof("write %s succeed \n", logMonitorFilePath)
	}
}

func (c *logCollectorStrore) Load()  {
	file,err := os.Open(logMonitorFilePath)
	if err != nil {
		log.Infof("read %s file error %v \n", logMonitorFilePath, err)
	}else{
		dec := gob.NewDecoder(file)
		err = dec.Decode(&c.Data)
		if err != nil {
			log.Errorf("gob decode %s error %v \n", logMonitorFilePath, err)
		}else{
			log.Infof("load %s file succeed \n", logMonitorFilePath)
		}
	}
	for _,v := range c.Data {
		lco := logCollectorObj{Path:v.Path}
		lco.Lock = new(sync.RWMutex)
		lco.Rule = v.Rule
		logCollectorJobs = append(logCollectorJobs, &lco)
		go lco.start()
	}
}

var LogCollectorStore logCollectorStrore

func LogMonitorHttpHandle(w http.ResponseWriter, r *http.Request)  {
	buff,err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor http request fail,read body error: %v \n", err)
		log.Errorln(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	var param []logHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel log monitor http request fail,json unmarshal error: %v \n", err)
		log.Errorln(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	for _,v := range logCollectorJobs{
		exist := false
		for _,vv := range param {
			if v.Path == vv.Path {
				exist = true
				var tmp []*logKeywordObj
				for _,vvv := range vv.Keywords {
					tmp = append(tmp, &logKeywordObj{Keyword:vvv})
				}
				v.update(tmp)
			}
		}
		if !exist {
			v.destroy()
		}
	}
	for _,v := range param {
		exist := false
		for _,vv := range logCollectorJobs {
			if v.Path == vv.Path {
				exist = true
				break
			}
		}
		if !exist {
			lco := logCollectorObj{Path:v.Path}
			lco.Lock = new(sync.RWMutex)
			var tmp []*logKeywordObj
			for _,vv := range v.Keywords {
				tmp = append(tmp, &logKeywordObj{Keyword:vv, Count:0})
			}
			lco.Rule = tmp
			logCollectorJobs = append(logCollectorJobs, &lco)
			go lco.start()
		}
	}
	w.Write([]byte("success"))
}