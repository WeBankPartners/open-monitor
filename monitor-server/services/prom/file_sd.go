package prom

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"sync"
	"fmt"
	"io/ioutil"
)

var fileSdList m.ServiceDiscoverFileList
var fileSdLock = new(sync.RWMutex)

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
}

func SyncSdConfigFile(step int) error {
	var configFile string
	if step == 10 {
		configFile = m.Config().SdFile.TenSecFile
	}else if step == 60 {
		configFile = m.Config().SdFile.OneMinFile
	}
	if configFile == "" {
		return fmt.Errorf("config file can not find ")
	}
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
