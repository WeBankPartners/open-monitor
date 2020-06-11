package prom

import (
	"sync"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"os"
	"fmt"
	"io/ioutil"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"gopkg.in/yaml.v2"
	"strings"
	"net/http"
	"time"
	"os/exec"
)

type fileObj struct {
	RWLock  sync.RWMutex
	Name  string
}

var FileMap map[string]fileObj
var PathEnbale bool

func InitPrometheusConfigFile()  {
	PathEnbale = true
	FileMap = make(map[string]fileObj)
	path := m.Config().Prometheus.ConfigPath
	s, err := os.Stat(path)
	if err != nil {
		PathEnbale = false
	}
	if s != nil {
		if !s.IsDir() {
			PathEnbale = false
		}
	}else{
		PathEnbale = false
	}
	if !PathEnbale {
		mid.LogError("init prometheus fail", fmt.Errorf("path %s is illeagal", path))
		return
	}
	files,_ := ioutil.ReadDir(path)
	for _,v := range files {
		name := strings.Split(v.Name(), ".yml")[0]
		FileMap[name] = fileObj{RWLock:*new(sync.RWMutex), Name:name}
		mid.LogInfo(fmt.Sprintf("prometheus rule file : %s", v.Name()))
	}
	mid.LogInfo("Success init prometheus config file")
}

func GetConfig(name string, isGrp bool) (error,bool,m.RFGroup) {
	if isGrp {
		name = "g_" + name
	}else{
		name = "e_" + name
	}
	path := fmt.Sprintf("%s/%s.yml", m.Config().Prometheus.ConfigPath, name)
	isExist := false
	_,err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			isExist = true
		}
	}else{
		isExist = true
	}
	if !isExist {
		return nil,isExist,m.RFGroup{}
	}
	if fo,b := FileMap[name]; b {
		fo.RWLock.RLock()
		defer fo.RWLock.RUnlock()
	}
	data,err := ioutil.ReadFile(path)
	if err != nil {
		mid.LogError("get prometheus rule,read file fail", err)
		return err,isExist,m.RFGroup{}
	}
	var rf m.RuleFile
	err = yaml.Unmarshal(data, &rf)
	if err != nil {
		mid.LogError("get prometheus rule,unmarshal fail", err)
		return err,isExist,m.RFGroup{}
	}
	if len(rf.Groups) <= 0 {
		return nil,isExist,m.RFGroup{}
	}
	return nil,isExist,*rf.Groups[0]
}

func SetConfig(name string, isGrp bool, config m.RFGroup, exist bool) error {
	if isGrp {
		name = "g_" + name
	}else{
		name = "e_" + name
	}
	path := fmt.Sprintf("%s/%s.yml", m.Config().Prometheus.ConfigPath, name)
	if len(config.Rules) == 0 {
		err := os.Remove(path)
		if err == nil {
			return nil
		}
	}
	rf := m.RuleFile{Groups:[]*m.RFGroup{&config}}
	data,err := yaml.Marshal(&rf)
	if err != nil {
		mid.LogError("set prometheus rule,marshal fail", err)
		return err
	}
	if fo,b := FileMap[name]; b {
		fo.RWLock.Lock()
		defer fo.RWLock.Unlock()
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		mid.LogError("set prometheus rule,write file fail", err)
		return err
	}
	//err = reloadConfig()
	//if err != nil {
	//	mid.LogError("set prometheus rule,reload config fail", err)
	//	return err
	//}
	return nil
}

func ReloadConfig() error {
	_,err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
	//mid.LogInfo(fmt.Sprintf("reload config resp : %v", resp.Body))
	//defer resp.Body.Close()
	return err
}

func StartCheckPrometheusJob(interval int)  {
	// Check prometheus
	var prometheusAddress string
	for _,v := range m.Config().Datasource.Servers {
		if v.Type == "prometheus" {
			prometheusAddress = v.Host
			break
		}
	}
	if prometheusAddress == "" {
		return
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		go checkPrometheusAlive(prometheusAddress)
		<- t
	}
}

func checkPrometheusAlive(address string)  {
	_,err := http.Get(fmt.Sprintf("http://%s", address))
	if err != nil {
		mid.LogError("prometheus alive check: error ", err)
		restartPrometheus()
	}
}

func restartPrometheus()  {
	mid.LogInfo("try to start prometheus . . . . . .")
	lastLog,_ := execCommand("tail -n 30 /app/monitor/prometheus/logs/prometheus.log")
	if lastLog != "" {
		for _,v := range strings.Split(lastLog, "\n") {
			if strings.Contains(v, "err=\"/app/monitor/prometheus/rules/") {
				errorFile := strings.Split(strings.Split(v, "err=\"/app/monitor/prometheus/rules/")[1], ":")[0]
				err := os.Remove(fmt.Sprintf("/app/monitor/prometheus/rules/%s", errorFile))
				if err != nil {
					mid.LogError(fmt.Sprintf("remove problem file %s error ", errorFile), err)
				}else{
					mid.LogInfo(fmt.Sprintf("remove problem file %s success", errorFile))
				}
			}
		}
	}else{
		mid.LogInfo("prometheus last log is empty ??")
	}
	startCommand,_ := execCommand("cat /app/monitor/start.sh |grep prometheus")
	if startCommand != "" {
		startCommand = strings.Replace(startCommand, "\n", " && ", -1)
		startCommand = startCommand[:len(startCommand)-3]
		_,err := execCommand(startCommand)
		if err != nil {
			mid.LogError("start prometheus fail,error ", err)
		}else{
			mid.LogInfo("start prometheus success ")
		}
	}else{
		mid.LogError("start prometheus fail, the start command is empty!!", nil)
	}
}

func execCommand(str string) (string,error) {
	b,err := exec.Command("/bin/sh", "-c", str).Output()
	if err != nil {
		mid.LogError(fmt.Sprintf("exec command %s fail,error", str), err)
	}
	return string(b),err
}