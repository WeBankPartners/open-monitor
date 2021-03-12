package telnet

import (
	"net"
	"fmt"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"time"
	"sync"
)

var (
	telnetResultList []*funcs.TelnetObj
	resultLock = new(sync.RWMutex)
)

func StartTelnetTask()  {
	interval := funcs.Config().Interval
	if interval < 30 {
		log.Println("telnet interval refresh to 30s")
		interval = 30
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		go telnetTask()
		<- t
	}
}

func telnetTask()  {
	startTime := time.Now()
	telnetList := funcs.GetTelnetList()
	clearTelnetResult(telnetList)
	wg := sync.WaitGroup{}
	//var successCounter int
	for _,v := range telnetList {
		wg.Add(1)
		go func(ip string,port int) {
			defer wg.Done()
			b := doTelnet(ip, port)
			writeTelnetResult(ip, port, b)
			funcs.DebugLog("telnet %s:%d result %b ", ip, port, b)
		}(v.Ip,v.Port)
	}
	wg.Wait()
	endTime := time.Now()
	useTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6
	resultList,successCount := getTelnetResult()
	log.Printf("end telnet, success num %d, fail num %d, use time %.3f ms \n", successCount, len(resultList)-successCount, useTime)
	funcs.UpdateTelnetExportMetric(resultList, successCount)
}

func doTelnet(ip string,port int) bool {
	conn,err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return false
	}else{
		conn.Close()
		return true
	}
}

func writeTelnetResult(ip string,port int,success bool)  {
	resultLock.Lock()
	for _,v := range telnetResultList {
		if v.Ip == ip && v.Port == port {
			v.Success = success
			break
		}
	}
	resultLock.Unlock()
}

func clearTelnetResult(param []*funcs.TelnetObj)  {
	resultLock.Lock()
	telnetResultList = []*funcs.TelnetObj{}
	for _,v := range param {
		telnetResultList = append(telnetResultList, &funcs.TelnetObj{Ip: v.Ip, Port: v.Port, Success: false})
	}
	resultLock.Unlock()
}

func getTelnetResult() (result []*funcs.TelnetObj,successCount int) {
	resultLock.RLock()
	for _,v := range telnetResultList {
		if v.Success {
			successCount += 1
		}
		result = append(result, &funcs.TelnetObj{Ip:v.Ip, Port:v.Port, Success:v.Success})
	}
	resultLock.RUnlock()
	return result,successCount
}