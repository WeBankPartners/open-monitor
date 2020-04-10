package funcs

import (
	"sync"
	"fmt"
	"bytes"
	"sort"
)

type exportMetricObj struct {
	Ip  string
	Port  int
	Value  int
	Note  string
}

var (
	exportPingMetrics = make(map[string]*exportMetricObj)
	exportTelnetMetrics = make(map[string]*exportMetricObj)
	exportPingLock = new(sync.RWMutex)
	exportTelnetLock = new(sync.RWMutex)
)

func UpdatePingExportMetric(result map[string]int,successCount int)  {
	exportPingLock.Lock()
	exportPingMetrics = make(map[string]*exportMetricObj)
	for k,v := range result {
		exportPingMetrics[k] = &exportMetricObj{Value:v, Note:fmt.Sprintf("# HELP ping target ip %s \n", k)}
	}
	exportPingMetrics[Config().Metrics.PingCountNum] = &exportMetricObj{Value:len(result), Note:"# HELP ping task ip num \n"}
	exportPingMetrics[Config().Metrics.PingCountSuccess] = &exportMetricObj{Value:successCount, Note:"# HELP ping success ip num \n"}
	exportPingMetrics[Config().Metrics.PingCountFail] = &exportMetricObj{Value:len(result)-successCount, Note:"# HELP ping fail ip num \n"}
	exportPingLock.Unlock()
}

func UpdateTelnetExportMetric(result []*TelnetObj,successCount int)  {
	exportTelnetLock.Lock()
	exportTelnetMetrics = make(map[string]*exportMetricObj)
	for _,v := range result {
		tmpValue := 1
		if v.Success {
			tmpValue = 0
		}
		exportTelnetMetrics[fmt.Sprintf("%s_%d", v.Ip, v.Port)] = &exportMetricObj{Ip:v.Ip, Port:v.Port, Value:tmpValue, Note:fmt.Sprintf("# HELP telnet target ip %s port %d \n", v.Ip, v.Port)}
	}
	exportTelnetMetrics[Config().Metrics.TelnetCountNum] = &exportMetricObj{Ip:Config().Metrics.TelnetCountNum, Value:len(result), Note:"# HELP telnet task num \n"}
	exportTelnetMetrics[Config().Metrics.TelnetCountSuccess] = &exportMetricObj{Ip:Config().Metrics.TelnetCountSuccess, Value:successCount, Note:"# HELP telnet success num \n"}
	exportTelnetMetrics[Config().Metrics.TelnetCountFail] = &exportMetricObj{Ip:Config().Metrics.TelnetCountFail, Value:len(result)-successCount, Note:"# HELP telnet fail num \n"}
	exportTelnetLock.Unlock()
}

func GetExportMetric() []byte {
	guidMap := GetSourceGuidMap()
	pingByte := getPingExportMetric(guidMap)
	telnetByte := getTelnetExportMetric(guidMap)
	return append(pingByte, telnetByte...)
}

func getPingExportMetric(guidMap map[string][]string) []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem. \n")
	metricString := Config().Metrics.Ping
	var tmpExportMetric exportMetricList
	exportPingLock.RLock()
	for k,v := range exportPingMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Ip:k, Value:v.Value, Note:v.Note})
	}
	exportPingLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _,v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.PingCountNum || v.Ip == Config().Metrics.PingCountSuccess || v.Ip == Config().Metrics.PingCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \n", v.Ip, v.Value))
			continue
		}
		if len(guidMap[v.Ip]) > 0 {
			for _,vv := range guidMap[v.Ip] {
				buff.WriteString(fmt.Sprintf("%s{target=\"%s\",guid=\"%s\"} %d \n", metricString, v.Ip, vv, v.Value))
			}
		}else {
			buff.WriteString(fmt.Sprintf("%s{target=\"%s\"} %d \n", metricString, v.Ip, v.Value))
		}
	}
	return buff.Bytes()
}

func getTelnetExportMetric(guidMap map[string][]string) []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP telnet check 0 -> alive, 1 -> dead \n")
	metricString := Config().Metrics.Telnet
	var tmpExportMetric exportMetricList
	exportTelnetLock.RLock()
	for _,v := range exportTelnetMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Ip:v.Ip, Port:v.Port, Value:v.Value, Note:v.Note})
	}
	exportTelnetLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _,v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.TelnetCountNum || v.Ip == Config().Metrics.TelnetCountSuccess || v.Ip == Config().Metrics.TelnetCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \n", v.Ip, v.Value))
			continue
		}
		tmpIpPort := fmt.Sprintf("%s:%d", v.Ip, v.Port)
		if len(guidMap[tmpIpPort]) > 0 {
			for _,vv := range guidMap[tmpIpPort] {
				buff.WriteString(fmt.Sprintf("%s{ip=\"%s\",port=\"%d\",guid=\"%s\"} %d \n", metricString, v.Ip, v.Port, vv, v.Value))
			}
		}else {
			buff.WriteString(fmt.Sprintf("%s{ip=\"%s\",port=\"%d\"} %d \n", metricString, v.Ip, v.Port, v.Value))
		}
	}
	return buff.Bytes()
}

type exportMetricList []*exportMetricObj

func (p exportMetricList) Len() int {
	return len(p)
}

func (p exportMetricList) Swap(i,j int) {
	p[i],p[j] = p[j],p[i]
}

func (p exportMetricList) Less(i,j int) bool {
	return p[i].Ip < p[j].Ip
}
