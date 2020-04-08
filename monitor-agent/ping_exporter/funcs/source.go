package funcs

import (
	"sync"
	"io/ioutil"
	"log"
	"strings"
	"time"
	"net/http"
	"encoding/json"
	"strconv"
)

var (
	sourceMap map[string]int
	sourceLock sync.RWMutex
)

type remoteResponse struct {
	Ip  []string  `json:"ip"`
}

//Note: weight参数是为了在众多数据源中识别当前数据源的数据并更新,weight越小权重越高,各数据源之间的关系是并集
//Note: 比如说remote的weight是3,file的weight是2,remote数据更新后不会覆盖file的数据
func InitIpList()  {
	sourceLock = *new(sync.RWMutex)
	sourceMap = make(map[string]int)
	var weight int
	if Config().Source.Const.Enabled {
		weight = 1
		if Config().Source.Const.Weight > 0 {
			weight = Config().Source.Const.Weight
		}
		for _,v := range Config().Source.Const.Ips {
			sourceMap[strings.TrimSpace(v)] = weight
		}
	}
	if Config().Source.File.Enabled {
		weight = 2
		if Config().Source.File.Weight > 0 {
			weight = Config().Source.File.Weight
		}
		ips,err := ioutil.ReadFile(Config().Source.File.Path)
		if err != nil {
			log.Printf("read file %s error: %v \n", Config().Source.File.Path, err)
		}else{
			for _,v := range strings.Split(string(ips), "\n") {
				sourceMap[strings.TrimSpace(v)] = weight
			}
		}
	}
	if Config().Source.Remote.Enabled && Config().Source.Remote.Url != "" {
		go startRemoteCurl()
	}
}

func startRemoteCurl()  {
	interval := 120
	if Config().Source.Remote.Interval > 0 {
		interval = Config().Source.Remote.Interval
	}
	weight := 3
	if Config().Source.Remote.Weight > 0 {
		weight = Config().Source.Remote.Weight
	}
	url := Config().Source.Remote.Url
	if Config().Source.Remote.GroupTag != "" {
		url = url + "?" + Config().Source.Remote.GroupTag
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
	sourceLock.Lock()
	for _,v := range ips {
		if vv,b := sourceMap[strings.TrimSpace(v)];b {
			if vv <= sourceType {
				continue
			}
		}
		sourceMap[strings.TrimSpace(v)] = sourceType
	}
	var tmpList []string
	for k,v := range sourceMap {
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
		delete(sourceMap, v)
	}
	sourceLock.Unlock()
}

func GetIpList() []string {
	var tmpList []string
	sourceLock.RLock()
	for k,_ := range sourceMap {
		if k == "" || strings.Contains(k, ":") {
			continue
		}
		tmpList = append(tmpList, k)
	}
	sourceLock.RUnlock()
	return tmpList
}

func GetTelnetList() []*TelnetObj {
	var tmpList []*TelnetObj
	sourceLock.RLock()
	for k,_ := range sourceMap {
		if strings.Contains(k, ":") {
			tmpSplit := strings.Split(k, ":")
			if len(tmpSplit) > 1 {
				i,_ := strconv.Atoi(tmpSplit[1])
				if i > 0 {
					tmpList = append(tmpList, &TelnetObj{Ip:tmpSplit[0], Port:i})
				}
			}
		}
	}
	sourceLock.RUnlock()
	return tmpList
}
