package collector

import (
	"github.com/go-kit/kit/log/level"
	"net/http"
	"sync"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"github.com/go-kit/kit/log"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"bytes"
	"encoding/gob"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

const (
	processFilePath = "data/process_cache.data"
)

type processMonitorCollector struct {
	processMonitor  *prometheus.Desc
	processCpuMonitor  *prometheus.Desc
	processMemMonitor  *prometheus.Desc
	logger  log.Logger
}

func (c *processMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _,v := range ProcessCacheObj.get() {
		ch <- prometheus.MustNewConstMetric(c.processMonitor,
			prometheus.GaugeValue,
			v.Value, v.DisplayName, v.Command, v.EndpointGuid)
		ch <- prometheus.MustNewConstMetric(c.processCpuMonitor,
			prometheus.GaugeValue,
			v.CpuUsedPercent, v.DisplayName, v.Command, v.EndpointGuid)
		ch <- prometheus.MustNewConstMetric(c.processMemMonitor,
			prometheus.GaugeValue,
			v.MemUsedByte, v.DisplayName, v.Command, v.EndpointGuid)
	}
	return nil
}

func init() {
	registerCollector("process_num", defaultEnabled, NewProcessMonitorCollector)
}

func NewProcessMonitorCollector(logger log.Logger) (Collector, error) {
	return &processMonitorCollector{
		processMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "count_current"),
			"Count the process num with assign name.",
			[]string{"name", "command", "process_guid"}, nil,
		),
		processCpuMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "cpu"),
			"Process cpu used percent",
			[]string{"name", "command", "process_guid"}, nil,
		),
		processMemMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "mem"),
			"Process memory used byte",
			[]string{"name", "command", "process_guid"}, nil,
		),
		logger: logger,
	}, nil
}

type processMonitorObj struct {
	Name  string
	Tags  string
	EndpointGuid string
	DisplayName string
	Value   float64
	Command  string
	CpuUsedPercent  float64
	MemUsedByte  float64
}

type processUsedResource struct {
	Pid  int
	Name  string
	Cmd  string
	Cpu  float64
	Mem  float64
}

type processCache struct {
	Lock *sync.RWMutex
	Running  bool
	ProcessMonitor  []*processMonitorObj
}

func (c *processCache) Init()  {
	c.Running = false
	c.Lock = new(sync.RWMutex)
	c.ProcessMonitor = []*processMonitorObj{}
	c.Load()
	if len(c.ProcessMonitor) > 0 {
		go c.start()
	}
}

func (c *processCache) start()  {
	c.Running = true
	t := time.NewTicker(time.Duration(10 * time.Second)).C
	for {
		<- t
		c.Lock.RLock()
		isRunning := c.Running
		c.Lock.RUnlock()
		if !isRunning {
			break
		}
		c.Lock.Lock()
		processUsedList := getProcessUsedResource()
		if len(processUsedList) > 0 {
			for _,tmpProcessMonitorObj := range c.ProcessMonitor {
				tmpTag := tmpProcessMonitorObj.Tags
				nameSplit := strings.Split(tmpProcessMonitorObj.Name, ",")
				if tmpTag != "" {
					tmpProcessMonitorObj.DisplayName = fmt.Sprintf("%s(%s)", tmpProcessMonitorObj.Name, tmpTag)
					tmpTag = strings.ToLower(tmpTag)
				}else{
					tmpProcessMonitorObj.DisplayName = tmpProcessMonitorObj.Name
				}
				var tmpCount float64 = 0
				for _,vv := range processUsedList {
					nameMatch := ""
					for _,nameSplitObj := range nameSplit {
						if vv.Name == strings.ToLower(nameSplitObj) {
							nameMatch = nameSplitObj
							break
						}
					}
					if nameMatch != "" && strings.Contains(vv.Cmd, tmpTag) {
						tmpCount = tmpCount + 1
						if len(vv.Cmd) > 100 {
							tmpProcessMonitorObj.Command = vv.Cmd[:100]
						}else{
							tmpProcessMonitorObj.Command = vv.Cmd
						}
						tmpProcessMonitorObj.CpuUsedPercent = vv.Cpu
						tmpProcessMonitorObj.MemUsedByte = vv.Mem
					}
				}
				tmpProcessMonitorObj.Value = tmpCount
			}
		}
		c.Lock.Unlock()
	}
}

func (c *processCache) stop()  {
	c.Lock.Lock()
	c.ProcessMonitor = []*processMonitorObj{}
	c.Running = false
	c.Lock.Unlock()
}

func (c *processCache) isRunning() bool {
	isRunning := false
	c.Lock.RLock()
	isRunning = c.Running
	c.Lock.RUnlock()
	return isRunning
}

func (c *processCache) update(names []*SyncProcessObj)  {
	c.Lock.Lock()
	c.ProcessMonitor = []*processMonitorObj{}
	for _,v := range names {
		c.ProcessMonitor = append(c.ProcessMonitor, &processMonitorObj{Name:v.ProcessName, Tags: v.ProcessTags, EndpointGuid: v.ProcessGuid, Value:0, Command:""})
	}
	c.Lock.Unlock()
}

func (c *processCache) Save()  {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(c.ProcessMonitor)
	if err != nil {
		level.Error(newLogger).Log("msg",fmt.Sprintf("gob encode process monitor error : %v ", err))
	}else{
		ioutil.WriteFile(processFilePath, tmpBuffer.Bytes(), 0644)
		level.Info(newLogger).Log("msg",fmt.Sprintf("write %s succeed ", processFilePath))
	}
}

func (c *processCache) Load()  {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	file,err := os.Open(processFilePath)
	if err != nil {
		level.Info(newLogger).Log("msg",fmt.Sprintf("read %s file error %v ", processFilePath, err))
	}else{
		dec := gob.NewDecoder(file)
		err = dec.Decode(&c.ProcessMonitor)
		if err != nil {
			level.Error(newLogger).Log("msg",fmt.Sprintf("gob decode %s error %v ", processFilePath, err))
		}else{
			level.Info(newLogger).Log("msg",fmt.Sprintf("load %s file succeed ", processFilePath))
		}
	}
}

func (c *processCache) get() []*processMonitorObj {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return c.ProcessMonitor
}

func (c *processCache) checkNum(processList []*SyncProcessObj) []int {
	processUseList := getProcessUsedResource()
	if len(processUseList) == 0 {
		return []int{}
	}
	var result []int
	for _,v := range processList {
		count := 0
		nameSplit := strings.Split(v.ProcessName, ",")
		for _,vv := range processUseList {
			nameMatch := ""
			for _,nameSplitObj := range nameSplit {
				if vv.Name == strings.ToLower(nameSplitObj) {
					nameMatch = nameSplitObj
					break
				}
			}
			if nameMatch != "" && strings.Contains(vv.Cmd, v.ProcessTags) {
				count = count + 1
			}
		}
		result = append(result, count)
	}
	return result
}

var ProcessCacheObj processCache

type processHttpDto struct {
	Process  []string  `json:"process"`
	Check    int       `json:"check"`
}

type SyncProcessObj struct {
	ProcessGuid string `json:"process_guid"`
	ProcessName string `json:"process_name"`
	ProcessTags string `json:"process_tags"`
}

type SyncProcessDto struct {
	Check int `json:"check"`
	Process []*SyncProcessObj `json:"process"`
}

func ProcessMonitorHttpHandle(w http.ResponseWriter, r *http.Request)  {
	buff,err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel process monitor http request fail,read body error: %v \n", err)
		level.Error(newLogger).Log("msg",errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	var param SyncProcessDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel process monitor http request fail,json unmarshal error: %v \n", err)
		level.Error(newLogger).Log("msg",errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	if len(param.Process) == 0 {
		ProcessCacheObj.stop()
		w.Write([]byte("Success"))
		return
	}
	if param.Check > 0 {
		illegalFlag := false
		checkNumResult := ProcessCacheObj.checkNum(param.Process)
		for i,v := range checkNumResult {
			if v != 1 {
				w.WriteHeader(http.StatusBadRequest)
				tmpProcessName := param.Process[i].ProcessName
				if param.Process[i].ProcessTags != "" {
					tmpProcessName += "(" + param.Process[i].ProcessTags + ")"
				}
				w.Write([]byte(fmt.Sprintf("Process %s num = %d", tmpProcessName, v)))
				illegalFlag = true
				break
			}
		}
		if illegalFlag {
			return
		}
	}
	ProcessCacheObj.update(param.Process)
	if !ProcessCacheObj.isRunning() {
		go ProcessCacheObj.start()
	}
	w.Write([]byte("Success"))
}

func getProcessUsedResource() []processUsedResource {
	var result []processUsedResource
	cmd := exec.Command("bash", "-c", "ps -eo 'pid,comm,pcpu,rsz,args'")
	b,err := cmd.Output()
	if err != nil {
		level.Error(newLogger).Log("msg",fmt.Sprintf("get process used resource error : %v ", err))
	}else{
		outputList := strings.Split(string(b), "\n")
		for _,v := range outputList {
			tmpList := strings.Split(v, " ")
			tmpIndex := 1
			strIndex := 0
			var tmpProcessObj processUsedResource
			for _,vv := range tmpList {
				strIndex += len(vv)+1
				if vv != "" {
					if tmpIndex == 1 {
						tmpPid,_ := strconv.Atoi(vv)
						if tmpPid > 0 {
							tmpProcessObj.Pid = tmpPid
						}
					}else if tmpIndex == 2 {
						tmpProcessObj.Name = strings.ToLower(vv)
					}else if tmpIndex == 3 {
						tmpCpu,_ := strconv.ParseFloat(vv, 64)
						tmpProcessObj.Cpu = tmpCpu
					}else if tmpIndex == 4 {
						tmpMem,_ := strconv.ParseFloat(vv, 64)
						tmpProcessObj.Mem = tmpMem
						break
					}
					tmpIndex++
				}
			}
			if len(v) > strIndex {
				tmpProcessObj.Cmd = v[strIndex:]
			}
			if tmpProcessObj.Pid > 0 {
				result = append(result, tmpProcessObj)
			}
		}
	}
	return result
}