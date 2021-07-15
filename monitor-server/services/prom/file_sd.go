package prom

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"sync"
	"fmt"
	"io/ioutil"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

var fileSdList m.ServiceDiscoverFileList
var fileSdLock = new(sync.RWMutex)
var fileSdPath string

func AddSdEndpoint(param m.ServiceDiscoverFileObj) []int {
	log.Logger.Debug("add sd endpoint", log.String("guid", param.Guid), log.String("address", param.Address), log.String("cluster", param.Cluster), log.Int("step", param.Step))
	var stepList []int
	if param.Guid == "" || param.Address == "" || param.Step == 0 {
		return stepList
	}
	stepList = append(stepList, param.Step)
	fileSdLock.Lock()
	exist := false
	for _,v := range fileSdList {
		if v.Guid == param.Guid {
			if v.Step != param.Step {
				stepList = append(stepList, v.Step)
				v.Step = param.Step
			}
			if v.Address != param.Address {
				v.Address = param.Address
			}
			exist = true
			break
		}
	}
	if !exist {
		fileSdList = append(fileSdList, &m.ServiceDiscoverFileObj{Guid:param.Guid, Address:param.Address, Step:param.Step, Cluster: param.Cluster})
	}
	fileSdLock.Unlock()
	return stepList
}

func DeleteSdEndpoint(guid string)  {
	if guid == "" {
		return
	}
	fileSdLock.Lock()
	var newFileSdList m.ServiceDiscoverFileList
	for _,v := range fileSdList {
		if v.Guid != guid {
			newFileSdList = append(newFileSdList, v)
		}
	}
	fileSdList = newFileSdList
	fileSdLock.Unlock()
}

func InitSdConfig(param []*m.ServiceDiscoverFileObj)  {
	fileSdLock.Lock()
	fileSdList = param
	fileSdLock.Unlock()
	fileSdPath = m.Config().SdFile.Path
	if fileSdPath != "" {
		if fileSdPath[len(fileSdPath)-1:] != "/" {
			fileSdPath = fileSdPath + "/"
		}
	}
}

func SyncSdConfigFile(step int) error {
	var configFile string
	if fileSdPath == "" {
		return fmt.Errorf("config file path is empty ")
	}
	configFile = fmt.Sprintf("%ssd_file_%d.json", fileSdPath, step)
	var sdConfigByte []byte
	fileSdLock.RLock()
	sdConfigByte = fileSdList.TurnToFileSdConfigByte(step)
	fileSdLock.RUnlock()
	if len(sdConfigByte) == 0 {
		return fmt.Errorf("step %d service discover config byte empty,please check ", step)
	}
	err := ioutil.WriteFile(configFile, sdConfigByte, 0644)
	if err != nil {
		return err
	}
	return nil
}

// New
var consumeSdConfigChan = make(chan []*m.SdConfigSyncObj, 50)

func StartConsumeSdConfig()  {
	if strings.HasSuffix(m.Config().Prometheus.SdConfigPath, "/") {
		m.Config().Prometheus.SdConfigPath = m.Config().Prometheus.SdConfigPath[:len(m.Config().Prometheus.SdConfigPath)-1]
	}
	output,err := execCommand("mkdir -p " + m.Config().Prometheus.SdConfigPath)
	if err != nil {
		log.Logger.Error("Run make sd dir command fail", log.String("output", output), log.Error(err))
	}
	log.Logger.Info("Start consume prometheus service discover config job")
	for {
		sdConfig := <- consumeSdConfigChan
		consumeSdConfig(sdConfig)
	}
}

func consumeSdConfig(params []*m.SdConfigSyncObj)  {
	log.Logger.Info("start consume sd config")
	var err error
	for _, param := range params {
		configFile := fmt.Sprintf("%ssd_file_%d.json", m.Config().Prometheus.SdConfigPath, param.Step)
		writeErr := ioutil.WriteFile(configFile, []byte(param.Content), 0644)
		if writeErr != nil {
			err = fmt.Errorf("Try to write sd file fail,%s ", writeErr.Error())
			break
		}
	}
	if err != nil {
		log.Logger.Error("Consume sd config fail", log.Error(err))
	}else{
		ReloadConfig()
	}
}

func SyncLocalSdConfig(params []*m.SdConfigSyncObj)  {
	if len(params) > 0 {
		consumeSdConfigChan <- params
	}
}