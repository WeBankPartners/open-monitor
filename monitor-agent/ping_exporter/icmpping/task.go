package icmpping

import (
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"log"
	"net"
	"sync"
	"time"
)

func StartTask() {
	timeout = 2
	interval := funcs.Config().Interval
	if interval < 30 {
		log.Println("interval refresh to 30s")
		interval = 30
	}
	if timeout > interval/12 {
		log.Printf("timeout %d > interval/12,max 3 try and 4 packages,reset to interval/12=%d \n", timeout, interval/12)
		timeout = interval / 12
	}
	InitIcmpBytes()           // 初始化ICMP数据包的bytes
	localIp = getIntranetIp() // 初始化本机IP列表

	if funcs.Config().OpenFalcon.Enabled {
		funcs.InitTransfer()
	}

	if TestModel {
		doTask()
		return
	}
	t := time.NewTicker(time.Second * time.Duration(interval)).C
	for {
		go doTask()
		<-t
	}
}

func doPacketLossTask() {

}

func doTask() {
	// 清空上次数据，把resultMap数据保存到lastResultMap
	//ClearRetryIp()
	ClearSuccessIp()
	//ClearRetryMap()
	lastResultMap := readResultMap()
	//lastResultMap := make(map[string]funcs.PingResultObj)
	//for k,v := range readResultMap(){
	//	lastResultMap[k] = v
	//}
	log.Println("start")
	startTime := time.Now()
	wg := sync.WaitGroup{}
	var successCounter int
	var lastValue int
	var lastExist bool
	ipList := funcs.GetIpList()
	clearResultMap(ipList)
	// first check
	for _, ip := range ipList {
		if containString(ip, localIp) {
			funcs.DebugLog("%s is local ip", ip)
			writeResultMap(ip, 0, 0)
			successCounter += 1
			continue
		}
		if lv, ok := lastResultMap[ip]; ok {
			lastValue = lv.UpDown
			lastExist = true
		} else {
			lastValue = 0
			lastExist = false
		}
		wg.Add(1)
		go func(ip string, timeout int, lv int, le bool) {
			d, ut, isConfused := StartPing(ip, timeout)
			if isConfused {
				if le == true && lv == 1 {
					d = 1
				} else {
					addSuccessIp(ip)
				}
			}
			writeResultMap(ip, d, ut)
			funcs.DebugLog("ping %s result %d ", ip, d)
			wg.Done()
		}(ip, timeout, lastValue, lastExist)
	}
	wg.Wait()

	// 对第一次执行有异常的IP执行第二次检测
	//retryIp := GetRetryIp()
	//retryLength := len(retryIp) // 重试IP的数量,如果第二次检测成功则数量减1,如果第二次检测完数量还大于0,则进行第三次检测
	//if len(retryIp) > 0 {
	//	funcs.DebugLog("start second round, retry ip num : %d ", len(retryIp))
	//	wgs := sync.WaitGroup{}
	//	for _,v:= range retryIp{
	//		wgs.Add(1)
	//		go func(ip string,timeout int) {
	//			defer wgs.Done()
	//			funcs.DebugLog("second round , retry ip %s ", ip)
	//			d,ut := StartPing(ip, timeout)
	//			if d==0 {
	//				retryLength = retryLength - 1
	//			}
	//			writeResultMap(ip, d, ut)
	//		}(v,timeout)
	//	}
	//	wgs.Wait()
	//	funcs.DebugLog("end second round ")
	//}

	// 第三次,最后一次检测,如果前面两次检测都还有需要重试的IP,说明此时的网络环境不稳定,需要对比上次检测的结果进行一次最后的重试
	//if retryLength>0 {
	//	funcs.DebugLog("start third round, retry ip num : %d ", retryLength)
	//	ClearRetryIp()
	//	tmpRMap := readResultMap()
	//	for _,v := range ipList{
	//		if lv,ok := lastResultMap[v]; ok {
	//			if lv!=tmpRMap[v]{  // 上次结果和这次的不一致，最后再检查一次，一般这种IP比较少
	//				addRetryIp(v)
	//				funcs.DebugLog("%s last result is different this : %d, last %d ", v, tmpRMap[v], lv)
	//				if tmpRMap[v].UpDown==0{
	//					successCounter = successCounter - 1
	//				}
	//				continue
	//			}
	//		}
	//		if tmpRMap[v].UpDown >= 2{
	//			funcs.DebugLog("%s retry is still not work ", v)
	//			addRetryIp(v)
	//		}
	//	}
	//	finalRetryIp := GetRetryIp()
	//	if len(finalRetryIp) > 0 {
	//		funcs.DebugLog("final retry ip num : %d ", len(finalRetryIp))
	//		wgt := sync.WaitGroup{}
	//		for _,v := range finalRetryIp{
	//			wgt.Add(1)
	//			go func(ip string,timeout int) {
	//				defer wgt.Done()
	//				d,ut := StartPing(ip, timeout)
	//				writeResultMap(ip, d, ut)
	//				funcs.DebugLog("ping %s result %d ", ip ,d)
	//			}(v,timeout)
	//		}
	//		wgt.Wait()
	//	}
	//}
	endTime := time.Now()
	useTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6
	successIp := GetSuccessIp()
	successCounter = successCounter + len(successIp)
	log.Printf("end ping, success num %d, fail num %d, use time %.3f ms \n", successCounter, len(ipList)-successCounter, useTime)
	dealResult(successCounter)
}

func dealResult(successCounter int) {
	result := readResultMap()
	if funcs.Config().Prometheus.Enabled {
		go funcs.UpdatePingExportMetric(result, successCounter)
	}
	if funcs.Config().OpenFalcon.Enabled {
		go funcs.HandleTransferResult(result, successCounter)
	}
	if TestModel {
		successOutput := "alive ip : \n"
		for k, v := range result {
			if v.UpDown == 0 {
				successOutput = successOutput + k + ","
			}
		}
		log.Println(successOutput)
	}
}

func getIntranetIp() []string {
	addrs, err := net.InterfaceAddrs()
	re := []string{}
	if err != nil {
		log.Println(err)
		return re
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				re = append(re, ipNet.IP.String())
			}
		}
	}
	return re
}
