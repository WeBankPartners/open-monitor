package icmpping

import (
	"sync"
)

var (
	TestModel bool
	localIp    []string   // 存储本机IP
	timeout int
	lastResultMap = make(map[string]int)  // 保存上次结果，和本次进行对比

	resultMap = make(map[string]int)   // 保存本次结果
	resultMapLock = new(sync.RWMutex)

	retryMap = make(map[string]int)   // 保存需要重试的IP
	retryMapLock = new(sync.RWMutex)

	successIpList []string
	successListLock = new(sync.RWMutex)

	retryIpList  []string
	retryListLock = new(sync.RWMutex)
)

func addSuccessIp(ip string) {
	successListLock.Lock()
	defer successListLock.Unlock()
	successIpList = append(successIpList, ip)
}

func GetSuccessIp() []string {
	successListLock.RLock()
	defer successListLock.RUnlock()
	return successIpList
}

func ClearSuccessIp() {
	successListLock.Lock()
	successIpList = []string{}
	successListLock.Unlock()
}

func addRetryIp(ip string) {
	retryListLock.Lock()
	defer retryListLock.Unlock()
	retryIpList = append(retryIpList, ip)
}

func GetRetryIp() []string {
	retryListLock.RLock()
	defer retryListLock.RUnlock()
	return retryIpList
}

func ClearRetryIp(){
	retryListLock.Lock()
	retryIpList = []string{}
	retryListLock.Unlock()
}

func writeResultMap(ip string, re int) {
	resultMapLock.Lock()
	defer resultMapLock.Unlock()
	resultMap[ip] = re
}

func readResultMap() map[string]int {
	resultMapLock.RLock()
	defer resultMapLock.RUnlock()
	return resultMap
}

func GetRetryMap(ip string, n int) int {
	retryMapLock.Lock()
	defer retryMapLock.Unlock()
	if v,ok := retryMap[ip]; ok{
		retryMap[ip] = v+n
		return v+n
	}else{
		retryMap[ip] = n
		return n
	}
}

func ClearRetryMap() {
	retryMapLock.Lock()
	defer retryMapLock.Unlock()
	for k,_ := range retryMap{
		retryMap[k] = 0
	}
}

func containString(ip string, ips []string) bool {
	for _,v := range ips {
		if ip == v {
			return true
		}
	}
	return false
}