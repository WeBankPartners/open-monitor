package collector

import (
	"net/http"
	"sync"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"github.com/prometheus/common/log"
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
}

func (c *processMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _,v := range ProcessCacheObj.get() {
		ch <- prometheus.MustNewConstMetric(c.processMonitor,
			prometheus.GaugeValue,
			v.Value, v.Name, v.Command)
		ch <- prometheus.MustNewConstMetric(c.processCpuMonitor,
			prometheus.GaugeValue,
			v.CpuUsedPercent, v.Name, v.Command)
		ch <- prometheus.MustNewConstMetric(c.processMemMonitor,
			prometheus.GaugeValue,
			v.MemUsedByte, v.Name, v.Command)
	}
	return nil
}

func init() {
	registerCollector("process_num", defaultEnabled, NewProcessMonitorCollector)
	ProcessCacheObj.init()
}

func NewProcessMonitorCollector() (Collector, error) {
	return &processMonitorCollector{
		processMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "count_current"),
			"Count the process num with assign name.",
			[]string{"name", "command"}, nil,
		),
		processCpuMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "cpu"),
			"Process cpu used percent",
			[]string{"name", "command"}, nil,
		),
		processMemMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "mem"),
			"Process memory used byte",
			[]string{"name", "command"}, nil,
		),
	}, nil
}

type processMonitorObj struct {
	Name  string
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

func (c *processCache) init()  {
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
	t := time.NewTicker(time.Duration(5 * time.Second)).C
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
			for _,v := range c.ProcessMonitor {
				tmpName := v.Name
				//tmpTag := ""
				var tmpCount float64 = 0
				//if strings.Contains(v.Name, "(") {
				//	tmpNameList := strings.Split(v.Name, "(")
				//	tmpName = tmpNameList[0]
				//	tmpTag = strings.Replace(tmpNameList[1], ")", "", -1)
				//}
				for _,vv := range processUsedList {
					//if vv.Name == tmpName && strings.Contains(vv.Cmd, tmpTag) {
					if strings.Contains(vv.Name, tmpName) || strings.Contains(vv.Cmd, tmpName) {
						tmpCount = tmpCount + 1
						if len(vv.Cmd) > 100 {
							v.Command = vv.Cmd[:100]
						}else{
							v.Command = vv.Cmd
						}
						v.CpuUsedPercent = vv.Cpu
						v.MemUsedByte = vv.Mem
					}
				}
				v.Value = tmpCount
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
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return c.Running
}

func (c *processCache) update(names []string)  {
	c.Lock.Lock()
	c.ProcessMonitor = []*processMonitorObj{}
	for _,v := range names {
		c.ProcessMonitor = append(c.ProcessMonitor, &processMonitorObj{Name:v, Value:0, Command:""})
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
		log.Errorf("gob encode process monitor error : %v \n", err)
	}else{
		ioutil.WriteFile(processFilePath, tmpBuffer.Bytes(), 0644)
		log.Infof("write %s succeed \n", processFilePath)
	}
}

func (c *processCache) Load()  {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	file,err := os.Open(processFilePath)
	if err != nil {
		log.Infof("read %s file error %v \n", processFilePath, err)
	}else{
		dec := gob.NewDecoder(file)
		err = dec.Decode(&c.ProcessMonitor)
		if err != nil {
			log.Errorf("gob decode %s error %v \n", processFilePath, err)
		}else{
			log.Infof("load %s file succeed \n", processFilePath)
		}
	}
}

func (c *processCache) get() []*processMonitorObj {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return c.ProcessMonitor
}

func (c *processCache) checkNum(names []string) []int {
	processUseList := getProcessUsedResource()
	if len(processUseList) == 0 {
		return []int{}
	}
	var result []int
	for _,v := range names {
		count := 0
		for _,vv := range processUseList {
			if strings.Contains(vv.Name, v) || strings.Contains(vv.Cmd, v) {
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

func ProcessMonitorHttpHandle(w http.ResponseWriter, r *http.Request)  {
	buff,err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel process monitor http request fail,read body error: %v \n", err)
		log.Errorln(errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	var param processHttpDto
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel process monitor http request fail,json unmarshal error: %v \n", err)
		log.Errorln(errorMsg)
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
				w.Write([]byte(fmt.Sprintf("Process %s num = %d", param.Process[i], v)))
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
		log.Errorf("get process used resource error : %v \n", err)
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
						tmpProcessObj.Name = vv
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