package collector

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/go-kit/kit/log"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	businessCollectorName = "business_monitor"
	businessMonitorFilePath = "data/business_monitor_cache.data"
)

var (
	regDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	businessMonitorJobs []*businessMonitorObj
	businessMonitorLock = new(sync.RWMutex)
	newLogger  log.Logger
)

type businessMonitorCollector struct {
	businessMonitor  *prometheus.Desc
	logger  log.Logger
}

func InitNewLogger(logger  log.Logger)  {
	newLogger = logger
}

func init() {
	registerCollector(businessCollectorName, defaultEnabled, BusinessMonitorCollector)
	BusinessCollectorStore.Load()
}

func BusinessMonitorCollector(logger log.Logger) (Collector, error) {
	return &businessMonitorCollector{
		businessMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, businessCollectorName, "value"),
			"Show business data from log file.",
			[]string{"sys", "msg", "key", "path"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *businessMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	businessMonitorLock.RLock()
	for _,v := range businessMonitorJobs {
		for _,vv := range v.get() {
			ch <- prometheus.MustNewConstMetric(c.businessMonitor,
				prometheus.GaugeValue,
				vv.Value, vv.SystemNum, vv.Name, vv.Key, vv.Path)
		}
	}
	businessMonitorLock.RUnlock()
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
	SystemNum  string  `json:"system_num"`
	Path  string  `json:"path"`
	Name  string  `json:"name"`
	LastDate  string  `json:"last_date"`
	Data  []*businessMetricObj  `json:"data"`
	TailSession  *tail.Tail  `json:"-"`
	Lock  *sync.RWMutex  `json:"-"`
}

type businessHttpDto struct {
	Paths  []string  `json:"paths"`
}

func (c *businessMonitorObj) start()  {
	var err error
	c.TailSession,err = tail.TailFile(c.Path, tail.Config{Follow:true, ReOpen:true})
	if err != nil {
		level.Error(newLogger).Log(fmt.Sprintf("start business collector fail, path: %s, error: %v", c.Path, err))
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
		for _,textSplit := range textList {
			if strings.HasPrefix(textSplit, "[") {
				textSplit = textSplit[1:]
			}
			if strings.HasSuffix(textSplit, "]") {
				textSplit = textSplit[:len(textSplit)-1]
			}
			if regDate.MatchString(textSplit) {
				c.LastDate = textSplit
			}
			if strings.Contains(textSplit, "{") &&  strings.Contains(textSplit, "}") {
				mapData := make(map[string]string)
				err := json.Unmarshal([]byte(textSplit), &mapData)
				if err == nil {
					if len(mapData) > 0 {
						c.Data = []*businessMetricObj{}
						for k, v := range mapData {
							floatValue,err := strconv.ParseFloat(v, 64)
							if err != nil {
								continue
							}
							c.Data = append(c.Data, &businessMetricObj{Key:k, Value:floatValue})
						}
					}
				}else{
					level.Info(newLogger).Log(fmt.Sprintf("json unmarshal %s error:%v ", textSplit, err))
				}
			}
		}
		c.Lock.Unlock()
	}
}

func (c *businessMonitorObj) get() []businessMetricObj {
	data := []businessMetricObj{}
	c.Lock.RLock()
	zeroFlag := true
	if c.LastDate != "" {
		if checkIllegalDate(c.LastDate) {
			zeroFlag = false
		}
	}
	for _,v := range c.Data {
		tmpValue := v.Value
		if !zeroFlag {
			tmpValue = 0
		}
		data = append(data, businessMetricObj{SystemNum:c.SystemNum, Name:c.Name, Path:c.Path, Key:v.Key, Value:tmpValue})
	}
	c.Lock.RUnlock()
	return data
}

func (c *businessMonitorObj) destroy()  {
	c.TailSession.Stop()
	c.Data = []*businessMetricObj{}
}

func BusinessMonitorHttpHandle(w http.ResponseWriter, r *http.Request)  {
	buff,err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel business monitor http request fail,read body error: %v \n", err)
		level.Error(newLogger).Log(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	var param businessHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel business monitor http request fail,json unmarshal error: %v \n", err)
		level.Error(newLogger).Log(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	businessMonitorLock.Lock()

	var newBusinessMonitorJobs []*businessMonitorObj
	for _,v := range businessMonitorJobs {
		exist := false
		for _,vv := range param.Paths {
			if v.Path == vv {
				exist = true
				break
			}
		}
		if !exist {
			v.destroy()
		}else{
			newBusinessMonitorJobs = append(newBusinessMonitorJobs, v)
		}
	}
	for _,v := range param.Paths {
		exist := false
		for _,vv := range businessMonitorJobs {
			if vv.Path == v {
				exist = true
				break
			}
		}
		if !exist {
			tmpBusinessObj := businessMonitorObj{Path:v}
			tmpBusinessObj.Data = []*businessMetricObj{}
			tmpBusinessObj.Lock = new(sync.RWMutex)
			go tmpBusinessObj.start()
			newBusinessMonitorJobs = append(newBusinessMonitorJobs, &tmpBusinessObj)
		}
	}
	businessMonitorJobs = newBusinessMonitorJobs
	businessMonitorLock.Unlock()
	level.Info(newLogger).Log("success")
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
	if time.Now().Sub(t).Seconds() > 10 {
		return true
	}
	return false
}

type businessCollectorStore struct {
	Data  []*businessMetricObj  `json:"data"`
}

var BusinessCollectorStore businessCollectorStore

func (c *businessCollectorStore) Save()  {
	for _,v := range businessMonitorJobs {
		BusinessCollectorStore.Data = append(BusinessCollectorStore.Data, &businessMetricObj{SystemNum:v.SystemNum,Path:v.Path,Name:v.Name})
	}
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(c.Data)
	if err != nil {
		level.Error(newLogger).Log(fmt.Sprintf("gob encode business monitor error : %v ", err))
	}else{
		ioutil.WriteFile(businessMonitorFilePath, tmpBuffer.Bytes(), 0644)
		level.Info(newLogger).Log(fmt.Sprintf("write %s succeed ", businessMonitorFilePath))
	}
}

func (c *businessCollectorStore) Load()  {
	file,err := os.Open(businessMonitorFilePath)
	if err != nil {
		level.Info(newLogger).Log(fmt.Sprintf("read %s file error %v ", businessMonitorFilePath, err))
	}else{
		dec := gob.NewDecoder(file)
		err = dec.Decode(&c.Data)
		if err != nil {
			level.Error(newLogger).Log(fmt.Sprintf("gob decode %s error %v ", businessMonitorFilePath, err))
		}else{
			level.Info(newLogger).Log(fmt.Sprintf("load %s file succeed ", businessMonitorFilePath))
		}
	}
	tmpMap := make(map[string]int)
	businessMonitorLock.Lock()
	businessMonitorJobs = []*businessMonitorObj{}
	for _,v := range c.Data {
		if v.Path != "" {
			if _,b := tmpMap[v.Path];!b {
				tmpMap[v.Path] = 1
				newBusinessMonitorObj := businessMonitorObj{Path: v.Path, SystemNum: v.SystemNum, Name: v.Name}
				newBusinessMonitorObj.Data = []*businessMetricObj{}
				newBusinessMonitorObj.Lock = new(sync.RWMutex)
				businessMonitorJobs = append(businessMonitorJobs, &newBusinessMonitorObj)
			}
		}
	}
	for _,v := range businessMonitorJobs {
		go v.start()
	}
	businessMonitorLock.Unlock()
}