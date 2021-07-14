package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type fileObj struct {
	RWLock  sync.RWMutex
	Name  string
}

var FileMap map[string]fileObj
var PathEnbale bool

func InitPrometheusRuleFile()  {
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
		log.Logger.Warn("Init prometheus fail,path illegal", log.String("path", path))
		return
	}
	files,_ := ioutil.ReadDir(path)
	for _,v := range files {
		name := strings.Split(v.Name(), ".yml")[0]
		FileMap[name] = fileObj{RWLock:*new(sync.RWMutex), Name:name}
		log.Logger.Info(fmt.Sprintf("prometheus rule file : %s", v.Name()))
	}
	log.Logger.Info("Success init prometheus config file")
}

func GetConfig(name string) (error,bool,m.RFGroup) {
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
		log.Logger.Error("Get prometheus rule,read file fail", log.Error(err))
		return err,isExist,m.RFGroup{}
	}
	var rf m.RuleFile
	err = yaml.Unmarshal(data, &rf)
	if err != nil {
		log.Logger.Error("Get prometheus rule,unmarshal fail", log.Error(err))
		return err,isExist,m.RFGroup{}
	}
	if len(rf.Groups) <= 0 {
		return nil,isExist,m.RFGroup{}
	}
	return nil,isExist,*rf.Groups[0]
}

func SetConfig(name string, config m.RFGroup) error {
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
		log.Logger.Error("Set prometheus rule,marshal fail", log.Error(err))
		return err
	}
	if fo,b := FileMap[name]; b {
		fo.RWLock.Lock()
		defer fo.RWLock.Unlock()
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		log.Logger.Error("Set prometheus rule,write file fail", log.Error(err))
		return err
	}
	return nil
}

func ReloadConfig() error {
	_,err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
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
		log.Logger.Error("Prometheus alive check: error", log.Error(err))
		restartPrometheus()
	}
}

func restartPrometheus()  {
	log.Logger.Info("Try to start prometheus . . . . . .")
	lastLog,_ := execCommand("tail -n 30 /app/monitor/prometheus/logs/prometheus.log")
	if lastLog != "" {
		for _,v := range strings.Split(lastLog, "\n") {
			if strings.Contains(v, "err=\"/app/monitor/prometheus/rules/") {
				errorFile := strings.Split(strings.Split(v, "err=\"/app/monitor/prometheus/rules/")[1], ":")[0]
				err := os.Remove(fmt.Sprintf("/app/monitor/prometheus/rules/%s", errorFile))
				if err != nil {
					log.Logger.Error(fmt.Sprintf("Remove problem file %s error ", errorFile), log.Error(err))
				}else{
					log.Logger.Info(fmt.Sprintf("Remove problem file %s success", errorFile))
				}
			}
		}
	}else{
		log.Logger.Info("Prometheus last log is empty ??")
	}
	startCommand,_ := execCommand("cat /app/monitor/start.sh |grep prometheus")
	if startCommand != "" {
		startCommand = strings.Replace(startCommand, "\n", " && ", -1)
		startCommand = startCommand[:len(startCommand)-3]
		_,err := execCommand(startCommand)
		if err != nil {
			log.Logger.Error("Start prometheus fail,error", log.Error(err))
		}else{
			log.Logger.Info("Start prometheus success")
		}
	}else{
		log.Logger.Warn("Start prometheus fail, the start command is empty!!")
	}
}

func StartCheckProcessList(interval int)  {
	if len(m.Config().ProcessCheckList) == 0 {
		return
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		<- t
		for _,v := range m.Config().ProcessCheckList {
			go checkSubProcessAlive(v)
		}
	}
}

func checkSubProcessAlive(name string)  {
	if name == "" {
		return
	}
	_,err := execCommand("ps aux|grep -v grep|grep "+name)
	if err != nil {
		if strings.Contains(err.Error(), "status 1") {
			startCommand, _ := execCommand("cat /app/monitor/start.sh |grep " + name)
			if startCommand != "" {
				startCommand = strings.Replace(startCommand, "\n", " && ", -1)
				startCommand = startCommand[:len(startCommand)-3]
				_, err := execCommand(startCommand)
				if err != nil {
					log.Logger.Error("Start sub process fail,error", log.String("process", name), log.String("command", startCommand), log.Error(err))
				} else {
					log.Logger.Info("Start sub process success", log.String("process", name))
				}
			}
		}
	}
}

func execCommand(str string) (string,error) {
	b,err := exec.Command("/bin/sh", "-c", str).Output()
	if err != nil {
		log.Logger.Error(fmt.Sprintf("Exec command %s fail,error", str), log.Error(err))
	}
	return string(b),err
}