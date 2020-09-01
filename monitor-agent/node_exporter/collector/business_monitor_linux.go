package collector

import (
	"sync"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"github.com/hpcloud/tail"
	"strings"
	"github.com/prometheus/common/log"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"fmt"
	"time"
)

const (
	businessCollectorName = "business_monitor"
)

type businessMonitorCollector struct {
	businessMonitor  *prometheus.Desc
}

func init() {
	registerCollector(businessCollectorName, defaultEnabled, BusinessMonitorCollector)
}

func BusinessMonitorCollector() (Collector, error) {
	return &businessMonitorCollector{
		businessMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, businessCollectorName, "value"),
			"Show business data from log file.",
			[]string{"sys", "msg", "key", "path"}, nil,
		),
	}, nil
}

func (c *businessMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _,v := range businessMonitorJobs {
		for _,vv := range v.get() {
			ch <- prometheus.MustNewConstMetric(c.businessMonitor,
				prometheus.GaugeValue,
				vv.Value, vv.SystemNum, vv.Name, vv.Key, vv.Path)
		}
	}
	return nil
}

type businessMetricObj struct {
	SystemNum  string
	Name  string
	Path  string
	Key  string
	Value  float64
}

type businessMonitorObj struct {
	SystemNum  string
	Path  string
	Name  string
	LastDate  string
	Data  map[string]string
	TailSession  *tail.Tail
	Lock  sync.RWMutex
}

type businessHttpDto struct {
	Paths  []string  `json:"paths"`
}

func (c *businessMonitorObj) start()  {
	var err error
	log.Infof("start business collector, path: %s \n", c.Path)
	c.TailSession,err = tail.TailFile(c.Path, tail.Config{Follow:true, ReOpen:true})
	if err != nil {
		log.Errorf("start business collector fail, path: %s, error: %v", c.Path, err)
		return
	}
	var tmpList []string
	if strings.Contains(c.Path, "/") {
		tmpList = strings.Split(strings.Split(c.Path, "/")[strings.Count(c.Path, "/")], "_")
	}else{
		tmpList = strings.Split(c.Path, "_")
	}
	if len(tmpList) > 3 {
		c.SystemNum = tmpList[1]
		c.Name = tmpList[2]
	}
	for line := range c.TailSession.Lines {
		c.Lock.Lock()
		textList := strings.Split(line.Text, "][")
		//log.Infof("Get a new line : %v \n", textList)
		if len(textList) > 8 {
			c.LastDate = textList[1]
			mapData := make(map[string]string)
			err := json.Unmarshal([]byte(textList[7]), &mapData)
			if err == nil {
				//log.Infof("Update new data : %v \n", mapData)
				c.Data = mapData
			}
		}
		c.Lock.Unlock()
	}
}

func (c *businessMonitorObj) get() []businessMetricObj {
	data := []businessMetricObj{}
	c.Lock.RLock()
	zeroFlag := true
	if checkIllegalDate(c.LastDate) {
		zeroFlag = false
	}
	for k,v := range c.Data {
		value,err := strconv.ParseFloat(v, 64)
		if !zeroFlag {
			value = 0
		}
		if err == nil {
			data = append(data, businessMetricObj{SystemNum:c.SystemNum, Name:c.Name, Path:c.Path, Key:k, Value:value})
		}
	}
	c.Lock.RUnlock()
	return data
}

func (c *businessMonitorObj) destroy()  {
	c.TailSession.Stop()
	c.Data = make(map[string]string)
}

var businessMonitorJobs = make(map[string]*businessMonitorObj)

func BusinessMonitorHttpHandle(w http.ResponseWriter, r *http.Request)  {
	buff,err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel business monitor http request fail,read body error: %v \n", err)
		log.Errorln(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	var param businessHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel business monitor http request fail,json unmarshal error: %v \n", err)
		log.Errorln(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	for k,v := range businessMonitorJobs {
		exist := false
		for _,vv := range param.Paths {
			if v.Path == vv {
				exist = true
				break
			}
		}
		if !exist {
			businessMonitorJobs[k].destroy()
			delete(businessMonitorJobs, k)
		}
	}
	for _,v := range param.Paths {
		if _,b := businessMonitorJobs[v];!b {
			tmpBusinessObj := businessMonitorObj{Path:v}
			businessMonitorJobs[v] = &tmpBusinessObj
			go businessMonitorJobs[v].start()
		}
	}
	log.Infoln("success")
	w.Write([]byte("success"))
}

func checkIllegalDate(input string) bool {
	tmpList := strings.Split(input, " ")
	if len(tmpList) < 2 {
		return true
	}
	t,err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s %s CST", tmpList[0], tmpList[1]))
	if err != nil {
		return true
	}
	sub := time.Now().Unix()-t.Unix()
	if sub >= 10 {
		return true
	}
	return false
}