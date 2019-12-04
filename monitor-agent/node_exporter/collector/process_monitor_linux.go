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
)

type processMonitorCollector struct {
	processMonitor  *prometheus.Desc
}

func (c *processMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _,v := range processCacheObj.get() {
		ch <- prometheus.MustNewConstMetric(c.processMonitor,
			prometheus.GaugeValue,
			v.Value, v.Name, v.Command)
	}
	return nil
}

func init() {
	registerCollector("process_num", defaultEnabled, NewProcessMonitorCollector)
	processCacheObj.init()
}

func NewProcessMonitorCollector() (Collector, error) {
	return &processMonitorCollector{
		processMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "count_current"),
			"Count the process num with assign name.",
			[]string{"name", "command"}, nil,
		),
	}, nil
}

type processMonitorObj struct {
	Name  string
	Value   float64
	Command  string
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

func (c *processCache) get() []*processMonitorObj {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return c.ProcessMonitor
}

var processCacheObj processCache

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
		processCacheObj.stop()
		w.Write([]byte("Success"))
		return
	}
	processCacheObj.update(param.Process)
	if !processCacheObj.isRunning() {
		go processCacheObj.start()
	}
	w.Write([]byte("Success"))
}
