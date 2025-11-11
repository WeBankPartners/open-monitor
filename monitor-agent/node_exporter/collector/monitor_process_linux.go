package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	processFilePath = "data/process_cache.json"
)

var ProcessJob processMonitorJob

type processMonitorCollector struct {
	processMonitor    *prometheus.Desc
	processCpuMonitor *prometheus.Desc
	processMemMonitor *prometheus.Desc
	processPidMonitor *prometheus.Desc
	logger            log.Logger
}

func (c *processMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	for _, v := range ProcessJob.GetResult() {
		ch <- prometheus.MustNewConstMetric(c.processMonitor,
			prometheus.GaugeValue,
			v.Value, v.DisplayName, v.Command, v.EndpointGuid)
		ch <- prometheus.MustNewConstMetric(c.processCpuMonitor,
			prometheus.GaugeValue,
			v.CpuUsedPercent, v.DisplayName, v.Command, v.EndpointGuid)
		ch <- prometheus.MustNewConstMetric(c.processMemMonitor,
			prometheus.GaugeValue,
			v.MemUsedByte, v.DisplayName, v.Command, v.EndpointGuid)
		ch <- prometheus.MustNewConstMetric(c.processPidMonitor,
			prometheus.GaugeValue,
			v.Pid, v.DisplayName, v.Command, v.EndpointGuid)
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
		processPidMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process_monitor", "pid"),
			"Process pid",
			[]string{"name", "command", "process_guid"}, nil,
		),
		logger: logger,
	}, nil
}

type processMonitorObj struct {
	Pid            float64
	Name           string
	Tags           string
	EndpointGuid   string
	DisplayName    string
	Command        string
	Value          float64
	CpuUsedPercent float64
	MemUsedByte    float64
}

type processUsedResource struct {
	Pid  int
	Name string
	Cmd  string
	Cpu  float64
	Mem  float64
}

type processMonitorJob struct {
	ConfigLock *sync.RWMutex
	Config     []*processConfigObj
	ResultLock *sync.RWMutex
	ResultList []*processMonitorObj
}

func (c *processMonitorJob) Init() {
	c.ConfigLock = new(sync.RWMutex)
	c.Config = []*processConfigObj{}
	c.ResultLock = new(sync.RWMutex)
	c.ResultList = []*processMonitorObj{}
}

func (c *processMonitorJob) ContainConfig() bool {
	containFlag := false
	c.ConfigLock.RLock()
	if len(c.Config) > 0 {
		containFlag = true
	}
	c.ConfigLock.RUnlock()
	return containFlag
}

func (c *processMonitorJob) UpdateConfig(input []*processConfigObj) {
	c.ConfigLock.Lock()
	c.Config = input
	c.ConfigLock.Unlock()
}

func (c *processMonitorJob) GetResult() []*processMonitorObj {
	var output []*processMonitorObj
	c.ResultLock.RLock()
	for _, v := range c.ResultList {
		output = append(output, &processMonitorObj{Pid: v.Pid, Name: v.Name, Tags: v.Tags, EndpointGuid: v.EndpointGuid, DisplayName: v.DisplayName, Command: v.Command, Value: v.Value, CpuUsedPercent: v.CpuUsedPercent, MemUsedByte: v.MemUsedByte})
	}
	c.ResultLock.RUnlock()
	return output
}

func StartProcessMonitorCron() {
	ProcessJob.Init()
	loadProcessConfig()
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go doProcessMonitor()
	}
}

func doProcessMonitor() {
	if !ProcessJob.ContainConfig() {
		// 当配置为空时，清空结果列表
		ProcessJob.ResultLock.Lock()
		ProcessJob.ResultList = []*processMonitorObj{}
		ProcessJob.ResultLock.Unlock()
		return
	}
	processUsedList := getProcessUsedResource()
	if len(processUsedList) == 0 {
		return
	}
	var resultList []*processMonitorObj
	ProcessJob.ConfigLock.RLock()
	for _, config := range ProcessJob.Config {
		matchList := matchProcess(processUsedList, config)
		if len(matchList) > 0 {
			resultList = append(resultList, matchList...)
		} else {
			resultList = append(resultList, &processMonitorObj{Name: config.ProcessName, DisplayName: config.ProcessName, Tags: config.ProcessTags, EndpointGuid: config.ProcessGuid, Value: 0, CpuUsedPercent: 0, MemUsedByte: 0, Pid: 0})
		}
	}
	ProcessJob.ConfigLock.RUnlock()
	ProcessJob.ResultLock.Lock()
	ProcessJob.ResultList = resultList
	ProcessJob.ResultLock.Unlock()
}

func matchProcess(processList []*processUsedResource, config *processConfigObj) (result []*processMonitorObj) {
	nameList := strings.Split(strings.ToLower(config.ProcessName), ",")
	for _, v := range processList {
		nameMatchFlag := false
		for _, name := range nameList {
			if v.Name == name {
				nameMatchFlag = true
				break
			}
		}
		if !nameMatchFlag {
			continue
		}
		if !strings.Contains(v.Cmd, config.ProcessTags) {
			continue
		}
		matchObj := processMonitorObj{Pid: float64(v.Pid), Value: 1, CpuUsedPercent: v.Cpu, MemUsedByte: v.Mem, DisplayName: config.ProcessName, EndpointGuid: config.ProcessGuid}
		if config.ProcessTags != "" {
			matchObj.DisplayName = fmt.Sprintf("%s(%s)", v.Name, config.ProcessTags)
		}
		if len(v.Cmd) > 50 {
			matchObj.Command = v.Cmd[:50]
		} else {
			matchObj.Command = v.Cmd
		}
		result = append(result, &matchObj)
	}
	return result
}

type processConfigObj struct {
	ProcessGuid string `json:"process_guid"`
	ProcessName string `json:"process_name"`
	ProcessTags string `json:"process_tags"`
}

type syncProcessConfigParam struct {
	Check   int                 `json:"check"`
	Process []*processConfigObj `json:"process"`
}

type syncProcessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func ProcessHttpHandle(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func(returnErr error) {
		responseObj := syncProcessResponse{Status: "OK", Message: "success"}
		if returnErr != nil {
			returnErr = fmt.Errorf("Handel process monitor http request fail,%s ", returnErr.Error())
			responseObj = syncProcessResponse{Status: "ERROR", Message: returnErr.Error()}
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
	level.Info(monitorLogger).Log("processConfig", string(requestParamBuff))
	isCheck := false
	isCheck, err = HandleProcessAction(requestParamBuff)
	if isCheck == false && err == nil {
		saveProcessConfig(requestParamBuff)
	}
}

func HandleProcessAction(requestParamBuff []byte) (isCheck bool, err error) {
	var param syncProcessConfigParam
	err = json.Unmarshal(requestParamBuff, &param)
	if err != nil {
		return
	}
	if param.Check > 0 {
		isCheck = true
		return
	}
	ProcessJob.UpdateConfig(param.Process)
	return
}

func saveProcessConfig(requestParamBuff []byte) {
	err := ioutil.WriteFile(processFilePath, requestParamBuff, 0644)
	if err != nil {
		level.Error(monitorLogger).Log("processSaveConfig", err.Error())
	} else {
		level.Info(monitorLogger).Log("processSaveConfig", "success")
	}
}

func loadProcessConfig() {
	b, err := ioutil.ReadFile(processFilePath)
	if err != nil {
		level.Warn(monitorLogger).Log("processLoadConfig", err.Error())
	} else {
		_, err = HandleProcessAction(b)
		if err != nil {
			level.Error(monitorLogger).Log("processLoadConfigAction", err.Error())
		} else {
			level.Info(monitorLogger).Log("processLoadConfig", "success")
		}
	}
}

func getProcessUsedResource() (result []*processUsedResource) {
	cmd := exec.Command("bash", "-c", "ps -eo 'pid,comm,pcpu,rsz,args'")
	b, err := cmd.Output()
	if err != nil {
		level.Error(monitorLogger).Log("msg", fmt.Sprintf("get process used resource error : %v ", err))
		return
	}
	for _, v := range strings.Split(string(b), "\n") {
		tmpList := strings.Split(v, " ")
		tmpIndex := 1
		strIndex := 0
		var tmpProcessObj processUsedResource
		for _, vv := range tmpList {
			strIndex += len(vv) + 1
			if vv != "" {
				if tmpIndex == 1 {
					tmpPid, _ := strconv.Atoi(vv)
					if tmpPid > 0 {
						tmpProcessObj.Pid = tmpPid
					}
				} else if tmpIndex == 2 {
					tmpProcessObj.Name = strings.ToLower(vv)
				} else if tmpIndex == 3 {
					tmpCpu, _ := strconv.ParseFloat(vv, 64)
					tmpProcessObj.Cpu = tmpCpu
				} else if tmpIndex == 4 {
					tmpMem, _ := strconv.ParseFloat(vv, 64)
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
			result = append(result, &tmpProcessObj)
		}
	}
	return
}
