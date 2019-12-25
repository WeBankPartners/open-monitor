package icmpping

import (
	"sync"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"io/ioutil"
	"log"
	"strings"
	"time"
	"net/http"
	"encoding/json"
)

var (
	ipListMap map[string]int
	ipLock sync.RWMutex
)

type remoteResponse struct {
	Ip  []string  `json:"ip"`
}

func InitIpList()  {
	ipLock = *new(sync.RWMutex)
	ipListMap = make(map[string]int)
	var weight int
	if funcs.Config().IpSource.Const.Enabled {
		weight = 1
		if funcs.Config().IpSource.Const.Weight > 0 {
			weight = funcs.Config().IpSource.Const.Weight
		}
		for _,v := range funcs.Config().IpSource.Const.Ips {
			ipListMap[strings.TrimSpace(v)] = weight
		}
	}
	if funcs.Config().IpSource.File.Enabled {
		weight = 2
		if funcs.Config().IpSource.File.Weight > 0 {
			weight = funcs.Config().IpSource.File.Weight
		}
		ips,err := ioutil.ReadFile(funcs.Config().IpSource.File.Path)
		if err != nil {
			log.Printf("read file %s error: %v \n", funcs.Config().IpSource.File.Path, err)
		}else{
			for _,v := range strings.Split(string(ips), "\n") {
				ipListMap[strings.TrimSpace(v)] = weight
			}
		}
	}
	if funcs.Config().IpSource.Remote.Enabled && funcs.Config().IpSource.Remote.Url != "" {
		go startRemoteCurl()
	}
}

func startRemoteCurl()  {
	interval := 120
	if funcs.Config().IpSource.Remote.Interval > 0 {
		interval = funcs.Config().IpSource.Remote.Interval
	}
	weight := 3
	if funcs.Config().IpSource.Remote.Weight > 0 {
		weight = funcs.Config().IpSource.Remote.Weight
	}
	url := funcs.Config().IpSource.Remote.Url
	if funcs.Config().IpSource.Remote.GroupTag != "" {
		url = url + "?" + funcs.Config().IpSource.Remote.GroupTag
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		<- t
		resp,err := http.Get(url)
		if err != nil {
			log.Printf("curl %s fail,error: %v \n", url, err)
		}else{
			b,_ := ioutil.ReadAll(resp.Body)
			if resp.StatusCode >= 300 {
				log.Printf("curl %s fail,resp code %d %s \n", url, resp.StatusCode, string(b))
			}else{
				var responseData remoteResponse
				err = json.Unmarshal(b, &responseData)
				if err != nil {
					log.Printf("curl %s fail,body unmarshal fail: %s", url, err)
				}else{
					UpdateIpList(responseData.Ip, weight)
				}
			}
		}
	}
}

func UpdateIpList(ips []string,sourceType int) {
	if len(ips) == 0 {
		return
	}
	ipLock.Lock()
	for _,v := range ips {
		if vv,b := ipListMap[strings.TrimSpace(v)];b {
			if vv <= sourceType {
				continue
			}
		}
		ipListMap[strings.TrimSpace(v)] = sourceType
	}
	var tmpList []string
	for k,v := range ipListMap {
		if v == sourceType {
			alive := false
			for _,vv := range ips {
				if k == vv {
					alive = true
					break
				}
			}
			if !alive {
				tmpList = append(tmpList, k)
			}
		}
	}
	for _,v := range tmpList {
		delete(ipListMap, v)
	}
	ipLock.Unlock()
}

func getIpList() []string {
	var tmpList []string
	ipLock.RLock()
	for k,_ := range ipListMap {
		if k == "" {
			continue
		}
		tmpList = append(tmpList, k)
	}
	ipLock.RUnlock()
	return tmpList
}