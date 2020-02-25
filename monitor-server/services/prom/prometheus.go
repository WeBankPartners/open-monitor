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
	resp,err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
	mid.LogInfo(fmt.Sprintf("reload config resp : %v", resp.Body))
	defer resp.Body.Close()
	return err
}
