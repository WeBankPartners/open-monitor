package prom

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"sync"
	"fmt"
	"io/ioutil"
)

var fileSdList m.ServiceDiscoverFileList
var fileSdLock = new(sync.RWMutex)
var fileSdPath string

func AddSdEndpoint(param m.ServiceDiscoverFileObj) {
	if param.Guid == "" || param.Address == "" || param.Step == 0 {
		return
	}
	fileSdLock.Lock()
	exist := false
	for _,v := range fileSdList {
		if v.Guid == param.Guid {
			exist = true
			break
		}
	}
	if !exist {
		fileSdList = append(fileSdList, &m.ServiceDiscoverFileObj{Guid:param.Guid, Address:param.Address, Step:param.Step})
	}
	fileSdLock.Unlock()
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
