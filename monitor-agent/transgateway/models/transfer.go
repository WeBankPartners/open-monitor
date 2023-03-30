package models

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var DataCache []*Member
var DataStore []*MemberStore
var TokenCache = make(map[string]string)
var DataCacheFile = `cache.data`
var TokenCacheFile = `token.data`
var MonitorUrl string
var LocalIp string
var LocalPort string

const (
	DatetimeFormat = `2006-01-02 15:04:05`
)

type TransRequest struct {
	UserAuthKey    string               `json:"userAuthKey" form:"userAuthKey" binding:"required"`
	MetricDataList []*RequestMetricData `json:"metricDataList"`
}

type RequestMetricData struct {
	SubSystemId      int         `json:"subSystemId" form:"subSystemId" binding:"required"`
	InterfaceName    string      `json:"interfaceName" form:"interfaceName" binding:"required"`
	AttrName         string      `json:"attrName" form:"attrName" binding:"required"`
	CollectTimestamp int64       `json:"collectTimestamp" form:"collectTimestamp"`
	MetricValue      interface{} `json:"metricValue" form:"metricValue" binding:"required"`
	HostIp           string      `json:"hostIp" form:"hostIp" binding:"required"`
	Object           interface{} `json:"object" form:"object"`
}

type TransResult struct {
	ResultCode int    `json:"resultCode"`
	ResultMsg  string `json:"resultMsg"`
	SystemTime string `json:"systemTime"`
}

type Member struct {
	Id         int
	Name       string
	Token      string
	Metrics    []*MetricObj
	LastUpdate time.Time
	Lock       sync.RWMutex
	Active     bool
}

type MemberStore struct {
	Id         int
	Name       string
	Token      string
	Metrics    []MetricObj
	LastUpdate time.Time
	Active     bool
}

type MetricObj struct {
	Id            string
	Metric        string
	Value         float64
	InterfaceName string
	Object        string
	AttrName      string
	HostIp        string
	LastUpdate    time.Time
	Active        bool
}

type TransGatewayRequestDto struct {
	Name    string `json:"name"`
	HostIp  string `json:"host_ip"`
	Address string `json:"address"`
}

func CleanTimeoutData(timeout int64) {
	t := time.NewTicker(time.Duration(60) * time.Second).C
	for {
		<-t
		tNow := time.Now().Unix()
		for i, v := range DataCache {
			v.Lock.Lock()
			if (tNow - v.LastUpdate.Unix()) > timeout {
				if !v.Active {
					DataCache[i].Metrics = []*MetricObj{}
				} else {
					DataCache[i].Active = false
				}
			} else {
				for _, vv := range v.Metrics {
					if (tNow - vv.LastUpdate.Unix()) > timeout {
						vv.Active = false
					}
				}
			}
			v.Lock.Unlock()
		}
	}
}

func SaveCacheData() {
	var dataBuffer bytes.Buffer
	var tokenBuffer bytes.Buffer
	DataStore = []*MemberStore{}
	for _, v := range DataCache {
		var tmpMetrics []MetricObj
		for _, vv := range v.Metrics {
			tmpMetrics = append(tmpMetrics, MetricObj{Id: vv.Id, Metric: vv.Metric, Value: vv.Value, InterfaceName: vv.InterfaceName, Object: vv.Object, AttrName: vv.AttrName, HostIp: vv.HostIp, LastUpdate: vv.LastUpdate, Active: vv.Active})
		}
		DataStore = append(DataStore, &MemberStore{Id: v.Id, Name: v.Name, Token: v.Token, Metrics: tmpMetrics, LastUpdate: v.LastUpdate, Active: v.Active})
	}
	enc := gob.NewEncoder(&dataBuffer)
	err := enc.Encode(DataStore)
	if err != nil {
		log.Println("data cache gob encode error : ", err)
	} else {
		ioutil.WriteFile(DataCacheFile, dataBuffer.Bytes(), 0644)
		log.Println("write cache.data succeed !")
	}
	ent := gob.NewEncoder(&tokenBuffer)
	err = ent.Encode(TokenCache)
	if err != nil {
		log.Println("token cache gob encode error : ", err)
	} else {
		ioutil.WriteFile(TokenCacheFile, tokenBuffer.Bytes(), 0644)
		log.Println("write token.data succeed !!")
	}
}

func LoadCacheData(dataDir string) {
	if dataDir != "" {
		DataCacheFile = dataDir + "/" + DataCacheFile
		TokenCacheFile = dataDir + "/" + TokenCacheFile
	}
	successLoadToken := false
	successLoadData := false
	tokenFile, err := os.Open(TokenCacheFile)
	if err != nil {
		log.Println("os open token.data fail ", err)
	} else {
		dec := gob.NewDecoder(tokenFile)
		err = dec.Decode(&TokenCache)
		if err != nil {
			log.Println("gob decode token.data fail : ", err)
		} else {
			log.Println("gob decode token.data success")
			successLoadToken = true
		}
	}
	dataFile, err := os.Open(DataCacheFile)
	if err != nil {
		log.Println("os open cache.data fail ", err)
	} else {
		dec := gob.NewDecoder(dataFile)
		err = dec.Decode(&DataStore)
		if err != nil {
			log.Println("gob decode cache.data fail : ", err)
		} else {
			log.Println("gob decode cache.data success")
			successLoadData = true
			for _, v := range DataStore {
				var member Member
				member.Token = v.Token
				member.Name = v.Name
				member.Active = v.Active
				member.LastUpdate = v.LastUpdate
				var tmpMetrics []*MetricObj
				for _, vv := range v.Metrics {
					if vv.Id == "" {
						continue
					}
					tmpMetrics = append(tmpMetrics, &MetricObj{Id: vv.Id, Metric: vv.Metric, Value: vv.Value, InterfaceName: vv.InterfaceName, Object: vv.Object, AttrName: vv.AttrName, HostIp: vv.HostIp, LastUpdate: vv.LastUpdate, Active: vv.Active})
				}
				member.Metrics = tmpMetrics
				member.Lock = *new(sync.RWMutex)
				DataCache = append(DataCache, &member)
				log.Println("load ", v.Name)
			}
		}
	}
	if successLoadToken && !successLoadData {
		for k, v := range TokenCache {
			var member Member
			member.Token = k
			member.Name = v
			member.Active = true
			member.LastUpdate = time.Now()
			member.Lock = *new(sync.RWMutex)
			DataCache = append(DataCache, &member)
		}
	}
}

func InitMonitorUrl(url, port string) {
	if url != "" {
		MonitorUrl = url + "/monitor/api/v1/agent/export/custom/endpoint/add"
		LocalPort = port
		LocalIp = getIntranetIp()
	}
}

func getIntranetIp() string {
	addrs, err := net.InterfaceAddrs()
	var re string
	if err != nil {
		log.Println(err)
		return re
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				re = ipNet.IP.String()
				break
			}
		}
	}
	return re
}
