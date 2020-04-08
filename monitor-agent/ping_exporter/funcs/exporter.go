package funcs

import (
	"sync"
	"fmt"
	"bytes"
	"sort"
)

type exportPingMetricObj struct {
	Ip  string
	Value  int
	Note  string
}

type exportTelnetMetricObj struct {
	Ip string
	Port int
	Value  int
	Note  string
}

var (
	exportPingMetrics = make(map[string]*exportPingMetricObj)
	exportTelnetMetrics = make(map[string]*exportTelnetMetricObj)
	exportPingLock = new(sync.RWMutex)
	exportTelnetLock = new(sync.RWMutex)
)

func UpdatePingExportMetric(result map[string]int,successCount int)  {
	exportPingLock.Lock()
	exportPingMetrics = make(map[string]*exportPingMetricObj)
	for k,v := range result {
		exportPingMetrics[k] = &exportPingMetricObj{Value:v, Note:fmt.Sprintf("# HELP target ip %s \r", k)}
	}
	exportPingMetrics[Config().Metrics.PingCountNum] = &exportPingMetricObj{Value:len(result), Note:"# HELP task ip num \r"}
	exportPingMetrics[Config().Metrics.PingCountSuccess] = &exportPingMetricObj{Value:successCount, Note:"# HELP alive ip num \r"}
	exportPingMetrics[Config().Metrics.PingCountFail] = &exportPingMetricObj{Value:len(result)-successCount, Note:"# HELP dead ip num \r"}
	exportPingLock.Unlock()
}

func UpdateTelnetExportMetric(result []*TelnetObj)  {
	exportTelnetLock.Lock()
	exportTelnetMetrics = make(map[string]*exportTelnetMetricObj)
	for _,v := range result {
		tmpValue := 1
		if v.Success {
			tmpValue = 0
		}
		exportTelnetMetrics[fmt.Sprintf("%s_%d", v.Ip, v.Port)] = &exportTelnetMetricObj{Ip:v.Ip, Port:v.Port, Value:tmpValue, Note:fmt.Sprintf("# HELP target ip %s port %d \r", v.Ip, v.Port)}
	}
	exportTelnetLock.Unlock()
}

func GetExportMetric() []byte {
	pingByte := getPingExportMetric()
	telnetByte := getTelnetExportMetric()
	return append(pingByte, telnetByte...)
}

func getPingExportMetric() []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem.")
	metricString := &Config().Metrics.Ping
	var tmpExportMetric exportMetricList
	exportPingLock.RLock()
	for k,v := range exportPingMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportPingMetricObj{Ip:k, Value:v.Value, Note:v.Note})
	}
	exportPingLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _,v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.PingCountNum || v.Ip == Config().Metrics.PingCountSuccess || v.Ip == Config().Metrics.PingCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \r\n", v.Ip, v.Value))
			continue
		}
		buff.WriteString(fmt.Sprintf("%s{target:\"%s\"} %d \r\n", metricString, v.Ip, v.Value))
	}
	return buff.Bytes()
}

func getTelnetExportMetric() []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP telnet check 0 -> alive, 1 -> dead \r")

	return buff.Bytes()
}

type exportMetricList []*exportPingMetricObj

func (p exportMetricList) Len() int {
	return len(p)
}

func (p exportMetricList) Swap(i,j int) {
	p[i],p[j] = p[j],p[i]
}

func (p exportMetricList) Less(i,j int) bool {
	return p[i].Ip < p[j].Ip
}
