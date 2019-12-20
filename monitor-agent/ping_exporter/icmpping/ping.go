package icmpping

import (
	"io/ioutil"
	"strings"
	"sync"
	"log"
	"time"
	m "github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/model"
	"net/http"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
)

var (
	Hosts map[string]string   // 存储主机IP和UUID对应关系,open-falcon中上报transfer需要用到
	resultMap map[string]int   // 保存本次结果
	lastResultMap map[string]int  // 保存上次结果，和本次进行对比
	retryMap map[string]int   // 保存需要重试的IP
	recordIps  []string   // 保存上一次检测的IP，以防本次检测获取IP失败而导致本次不跑，如果IP列表获取失败则跑上一次的IP列表
	localIp    []string   // 存储本机IP
	localUuid  string
	rLock  sync.RWMutex
	tLock  sync.RWMutex
	)

func Start(isTest bool) {
	timeout := funcs.Config().Timeout
	interval := funcs.Config().Interval
	if timeout > interval/12 {
		log.Printf("timeout %d > interval/12,max 3 try and 4 packages,reset to interval/12=%d \n",timeout,interval/12)
		timeout = interval/12
	}
	InitIcmpBytes()  // 初始化ICMP数据包的bytes
	localIp = funcs.GetIntranetIp()  // 初始化本机IP列表
	Hosts = make(map[string]string)
	resultMap = make(map[string]int)
	lastResultMap = make(map[string]int)
	retryMap = make(map[string]int)
	recordIps = []string{}


	if funcs.Config().OpenFalcon.Enabled {
		localUuid = funcs.Uuid()
	}else{
		ips,err := ioutil.ReadFile(funcs.Config().IpSource.File.Path)
		if err != nil {
			log.Fatalf("read file error : %v \n", err)
			return
		}
		addressIp := strings.Split(string(ips), "\n")
		task(addressIp, timeout, false)
	}

	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		<- t
		go task([]string{}, timeout, true)
	}
}

func getIps() []string {
	//if funcs.Config().Debug {
	//	log.Println("start get monitor ip")
	//}
	DebugLog("start get monitor ip")
	//hostname,err := os.Hostname()
	//if err!=nil{
	//	log.Fatalf("get monitor hostname error : %v", err)
	//	return []string{}
	//}
	url := funcs.Config().IpSource.Remote.Url
	resp,err := http.Get(url)
	if err!=nil {
		log.Fatalf("get monitor ip error : %v", err)
		return []string{}
	}
	body,err := ioutil.ReadAll(resp.Body)
	if err!=nil{
		log.Fatalf("get monitor ip read body error : %v", err)
		return []string{}
	}
	var result m.Response
	var ips []string
	json.Unmarshal(body, &result)
	for _,v :=range result.Data{
		tmp := strings.Split(v, ":")
		if len(tmp) > 1{
			Hosts[tmp[0]] = tmp[1]
			ips = append(ips, tmp[0])
		}
	}
	return ips
}

func task(addressIp []string, timeout int, isTransfer bool) {
	if isTransfer{
		if funcs.Config().IpSource.Remote.Enabled {
			addressIp = getIps()
		}
		if funcs.Config().IpSource.Const.Enabled {
			for _,v := range funcs.Config().IpSource.Const.Ips {
				flag := false
				for _,vv := range addressIp {
					if v == vv {
						flag = true
						break
					}
				}
				if !flag {
					addressIp = append(addressIp, v)
				}
			}
		}
		if len(addressIp)>0{  // 如果本次IP获取失败，则取上一次的IP列表，如果成功，则保存本次IP列表
			recordIps = addressIp
		}else{
			addressIp = recordIps
		}
		// 清空上次数据，把resultMap数据保存到lastResultMap
		ClearRetryIp()
		ClearDoneIp()
		ClearRetryMap()
		for k,v := range readResultMap(){
			lastResultMap[k] = v
		}
	}
	log.Println("start ping")
	startTime := time.Now()
	wg := sync.WaitGroup{}
	count := 0  // 计数器,纪录成功的数量
	all := 0  // 计数器,纪录全部可检测的IP的数量
	// 第一次检测
	for _,ip := range addressIp {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			all++
			if IsLocalIp(ip, localIp) {    // 判断是否本地IP
				DebugLog("%s is local ip", ip)
				writeResultMap(ip, 0)
				DebugLog("ping %s result %d ", ip ,0)
				count += 1
				continue
			}
			wg.Add(1)
			go func(ip string,timeout int) {
				defer wg.Done()
				d := StartPing(ip, timeout)
				if d<2 {
					writeResultMap(ip ,d)
				}
				DebugLog("ping %s result %d ", ip, d)
			}(ip,timeout)
		}
	}
	wg.Wait()

	// 对第一次执行有异常的IP执行第二次检测
	retryIp := GetRetryIp()
	retryLength := len(retryIp) // 重试IP的数量,如果第二次检测成功则数量减1,如果第二次检测完数量还大于0,则进行第三次检测
	if len(retryIp) > 0 {
		DebugLog("start second round, retry ip num : %d ", len(retryIp))
		wgs := sync.WaitGroup{}
		for _,v:= range retryIp{
			wgs.Add(1)
			go func(ip string,timeout int) {
				defer wgs.Done()
				DebugLog("second round , retry ip %s ", ip)
				d := StartPing(ip, timeout)
				if d==0{
					retryLength = retryLength - 1
				}
				if isTransfer {
					writeResultMap(ip, d)
				}else{
					DebugLog("ping %s result %d ", ip, d)
				}
			}(v,timeout)
		}
		wgs.Wait()
		DebugLog("end second round ")
	}

	// 第三次,最后一次检测,如果前面两次检测都还有需要重试的IP,说明此时的网络环境不稳定,需要对比上次检测的结果进行一次最后的重试
	if retryLength>0 {
		DebugLog("start third round, retry ip num : %d ", retryLength)
		ClearRetryIp()
		tmpRMap := readResultMap()
		for _,v := range addressIp{
			if lv,ok := lastResultMap[v]; ok {
				if lv!=tmpRMap[v]{  // 上次结果和这次的不一致，最后再检查一次，一般这种IP比较少
					addRetryIp(v)
					DebugLog("%s last result is different this : %d, last %d ", v, tmpRMap[v], lv)
					if tmpRMap[v]==0{
						count = count-1
					}
					continue
				}
			}
			if tmpRMap[v] >= 2{
				DebugLog("%s retry is still not work ", v)
				addRetryIp(v)
			}
		}
		finalRetryIp := GetRetryIp()
		if len(finalRetryIp) > 0 {
			DebugLog("final retry ip num : %d ", len(finalRetryIp))
			wgt := sync.WaitGroup{}
			for _,v := range finalRetryIp{
				wgt.Add(1)
				go func(ip string,timeout int) {
					defer wgt.Done()
					d := StartPing(ip, timeout)
					writeResultMap(ip, d)
					DebugLog("ping %s result %d ", ip ,d)
				}(v,timeout)
			}
			wgt.Wait()
		}
	}
	endTime := time.Now()
	useTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6
	doneIp := GetDoneIp()
	count = count + len(doneIp)
	failNum := all-count
	log.Printf("end ping, success num %d, fail num %d, use time %.3f ms \n", count, failNum, useTime)
	if isTransfer{
		sendData := []*m.MetricValue{}
		metric := funcs.Config().Metrics.Default
		interval := funcs.Config().Interval
		rMap := readResultMap()
		now := time.Now().Unix()
		for _,v := range addressIp{
			endpoint := Hosts[v]
			if endpoint!="" {
				metricData := m.MetricValue{Endpoint: endpoint, Metric: metric, Value: rMap[v], Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
				sendData = append(sendData, &metricData)
			}else{
				metricData := m.MetricValue{Endpoint: localUuid, Metric: fmt.Sprintf("%s_check", metric), Value: rMap[v], Step: int64(interval), Type: "GAUGE", Tags: fmt.Sprintf("ip=%s",v), Timestamp: now}
				sendData = append(sendData, &metricData)
			}
		}
		if localUuid!=""{
			metricOk := m.MetricValue{Endpoint: localUuid, Metric: "ip_ping_ok_num", Value: count, Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
			metricFa := m.MetricValue{Endpoint: localUuid, Metric: "ip_ping_fa_num", Value: failNum, Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
			metricAll := m.MetricValue{Endpoint: localUuid, Metric: "ip_ping_all_num", Value: all, Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
			sendData = append(sendData, &metricOk)
			sendData = append(sendData, &metricFa)
			sendData = append(sendData, &metricAll)
		}
		length := len(sendData)
		sn := funcs.Config().OpenFalcon.Transfer.Sn
		if sn<=0{
			sn = 500
		}
		if length > 0 {
			log.Printf("=> <Total=%d> %v\n", length, sendData[0])
		}
		if length>sn{
			var resps m.TransferResponse
			cut := length/sn
			if cut*sn<length{
				cut = cut+1
			}
			for i:=0;i<cut;i++{
				s := i*sn
				e := (i+1)*sn
				if i==cut-1{
					e = length
				}
				funcs.SendMetrics(sendData[s:e], &resps)
				log.Printf("s: %d , e : %d <= %v \n", s, e, &resps)
			}
		}else {
			var resp m.TransferResponse
			funcs.SendMetrics(sendData, &resp)
			log.Println("<=", &resp)
		}
	}
}

func writeResultMap(ip string, re int) {
	rLock.Lock()
	defer rLock.Unlock()
	resultMap[ip] = re
}

func readResultMap() map[string]int {
	rLock.RLock()
	defer rLock.RUnlock()
	return resultMap
}

func GetRetryMap(ip string, n int) int {
	tLock.Lock()
	defer tLock.Unlock()
	if v,ok := retryMap[ip]; ok{
		retryMap[ip] = v+n
		return v+n
	}else{
		retryMap[ip] = n
		return n
	}
}

func ClearRetryMap() {
	tLock.Lock()
	defer tLock.Unlock()
	for k,_ := range retryMap{
		retryMap[k] = 0
	}
}

func DebugLog(msg string, v ...interface{}){
	if funcs.Config().Debug {
		msg = msg + " \n"
		log.Printf(msg, v...)
	}
}