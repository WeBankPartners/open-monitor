package collector

import (
	"net/http"
	"sync"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"github.com/toolkits/nux"
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
	Name  string
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
		processList,err := nux.AllProcs()
		processUsedMap := getProcessUsedResource()
		if err != nil {
			log.Errorf("Get nux process list fail : %v \n", err)
		}else{
			//log.Infof("Get process list success,num : %d \n", len(processList))
			for _,v := range c.ProcessMonitor {
				v.Value = 0
			}
			for _,v := range processList {
				for _,vv := range c.ProcessMonitor {
					if v.Name == vv.Name {
						vv.Value = vv.Value + 1
						if len(v.Cmdline) > 100 {
							vv.Command = v.Cmdline[:100]
						}else {
							vv.Command = v.Cmdline
						}
					}
				}
			}
		}
		if len(processUsedMap) > 0 {
			for _,v := range c.ProcessMonitor {
				v.CpuUsedPercent = processUsedMap[v.Name].Cpu
				v.MemUsedByte = processUsedMap[v.Name].Mem
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

var ProcessCacheObj processCache

type processHttpDto struct {
	Process  []string  `json:"process"`
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
	ProcessCacheObj.update(param.Process)
	if !ProcessCacheObj.isRunning() {
		go ProcessCacheObj.start()
	}
	w.Write([]byte("Success"))
}

func getProcessUsedResource() map[string]processUsedResource {
	result := make(map[string]processUsedResource)
	cmd := exec.Command("bash", "-c", "ps -eo 'comm,pcpu,rsz'")
	b,err := cmd.Output()
	if err != nil {
		log.Errorf("get process used resource error : %v \n", err)
	}else{
		outputList := strings.Split(string(b), "\n")
		for _,v := range outputList {
			tmpList := strings.Split(v, " ")
			tmpIndex := 1
			var tmpProcessObj processUsedResource
			for _,vv := range tmpList {
				if vv != "" {
					if tmpIndex == 1 {
						tmpProcessObj.Name = vv
					}else if tmpIndex == 2 {
						tmpCpu,_ := strconv.ParseFloat(vv, 64)
						tmpProcessObj.Cpu = tmpCpu
					}else if tmpIndex == 3 {
						tmpMem,_ := strconv.ParseFloat(vv, 64)
						tmpProcessObj.Mem = tmpMem
					}
					tmpIndex++
				}
			}
			result[tmpProcessObj.Name] = tmpProcessObj
		}
	}
	return result
}