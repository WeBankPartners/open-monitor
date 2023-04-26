package funcs

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
)

type exportMetricObj struct {
	Ip      string
	Port    int
	Url     string
	Method  string
	Value   int
	Note    string
	UseTime float64
}

var (
	exportPingMetrics      = make(map[string]*exportMetricObj)
	exportLossPingMetrics  = make(map[string]*exportMetricObj)
	exportTelnetMetrics    = make(map[string]*exportMetricObj)
	exportHttpCheckMetrics = make(map[string]*exportMetricObj)
	exportPingLock         = new(sync.RWMutex)
	exportLossPingLock     = new(sync.RWMutex)
	exportTelnetLock       = new(sync.RWMutex)
	exportHttpCheckLock    = new(sync.RWMutex)
)

func UpdatePingExportMetric(result map[string]PingResultObj, successCount int) {
	exportPingLock.Lock()
	exportPingMetrics = make(map[string]*exportMetricObj)
	for k, v := range result {
		exportPingMetrics[k] = &exportMetricObj{Value: v.UpDown, UseTime: v.UseTime, Note: fmt.Sprintf("# HELP ping target ip %s \n", k)}
	}
	exportPingMetrics[Config().Metrics.PingCountNum] = &exportMetricObj{Value: len(result), Note: "# HELP ping task ip num \n"}
	exportPingMetrics[Config().Metrics.PingCountSuccess] = &exportMetricObj{Value: successCount, Note: "# HELP ping success ip num \n"}
	exportPingMetrics[Config().Metrics.PingCountFail] = &exportMetricObj{Value: len(result) - successCount, Note: "# HELP ping fail ip num \n"}
	exportPingLock.Unlock()
}

func UpdateLossPingExportMetric(result map[string]PingResultObj) {
	exportLossPingLock.Lock()
	exportLossPingMetrics = make(map[string]*exportMetricObj)
	for k, v := range result {
		exportLossPingMetrics[k] = &exportMetricObj{Value: int(v.LossPercent), Note: fmt.Sprintf("# HELP loss ping target ip %s \n", k)}
	}
	exportLossPingLock.Unlock()
}

func UpdateTelnetExportMetric(result []*TelnetObj, successCount int) {
	exportTelnetLock.Lock()
	exportTelnetMetrics = make(map[string]*exportMetricObj)
	for _, v := range result {
		tmpValue := 1
		if v.Success {
			tmpValue = 0
		}
		exportTelnetMetrics[fmt.Sprintf("%s_%d", v.Ip, v.Port)] = &exportMetricObj{Ip: v.Ip, Port: v.Port, Value: tmpValue, Note: fmt.Sprintf("# HELP telnet target ip %s port %d \n", v.Ip, v.Port)}
	}
	exportTelnetMetrics[Config().Metrics.TelnetCountNum] = &exportMetricObj{Ip: Config().Metrics.TelnetCountNum, Value: len(result), Note: "# HELP telnet task num \n"}
	exportTelnetMetrics[Config().Metrics.TelnetCountSuccess] = &exportMetricObj{Ip: Config().Metrics.TelnetCountSuccess, Value: successCount, Note: "# HELP telnet success num \n"}
	exportTelnetMetrics[Config().Metrics.TelnetCountFail] = &exportMetricObj{Ip: Config().Metrics.TelnetCountFail, Value: len(result) - successCount, Note: "# HELP telnet fail num \n"}
	exportTelnetLock.Unlock()
}

func UpdateHttpCheckExportMetric(result []*HttpCheckObj, successCount int) {
	exportHttpCheckLock.Lock()
	exportHttpCheckMetrics = make(map[string]*exportMetricObj)
	for _, v := range result {
		exportHttpCheckMetrics[v.Url] = &exportMetricObj{Ip: v.Url, Url: v.Url, Method: v.Method, Value: v.StatusCode, Note: fmt.Sprintf("# HELP http check target method %s url %s \n", v.Method, v.Url)}
	}
	exportHttpCheckMetrics[Config().Metrics.HttpCheckCountNum] = &exportMetricObj{Ip: Config().Metrics.HttpCheckCountNum, Url: Config().Metrics.HttpCheckCountNum, Value: len(result), Note: "# HELP http check task num \n"}
	exportHttpCheckMetrics[Config().Metrics.HttpCheckCountSuccess] = &exportMetricObj{Ip: Config().Metrics.HttpCheckCountSuccess, Url: Config().Metrics.HttpCheckCountSuccess, Value: successCount, Note: "# HELP http check success num \n"}
	exportHttpCheckMetrics[Config().Metrics.HttpCheckCountFail] = &exportMetricObj{Ip: Config().Metrics.HttpCheckCountFail, Url: Config().Metrics.HttpCheckCountFail, Value: len(result) - successCount, Note: "# HELP http check fail num \n"}
	exportHttpCheckLock.Unlock()
}

func GetExportMetric() []byte {
	var result []byte
	guidMap := GetSourceGuidMap()
	if Config().PingEnable {
		pingByte := getPingExportMetric(guidMap)
		result = append(result, pingByte...)
		lossPingByte := getLossPingExportMetric(guidMap)
		result = append(result, lossPingByte...)
	}
	if Config().TelnetEnable {
		telnetByte := getTelnetExportMetric(guidMap)
		result = append(result, telnetByte...)
	}
	if Config().HttpCheckEnable {
		httpCheckByte := getHttpCheckExportMetric(guidMap)
		result = append(result, httpCheckByte...)
	}
	return result
}

func getPingExportMetric(guidMap map[string][]string) []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem. \n")
	metricString := Config().Metrics.Ping
	metricTimeString := Config().Metrics.PingUseTime
	var tmpExportMetric exportMetricList
	exportPingLock.RLock()
	for k, v := range exportPingMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Ip: k, Value: v.Value, Note: v.Note, UseTime: v.UseTime})
	}
	exportPingLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _, v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.PingCountNum || v.Ip == Config().Metrics.PingCountSuccess || v.Ip == Config().Metrics.PingCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \n", v.Ip, v.Value))
			continue
		}
		if len(guidMap[v.Ip]) > 0 {
			for _, vv := range guidMap[v.Ip] {
				buff.WriteString(fmt.Sprintf("%s{target=\"%s\",guid=\"%s\"} %d \n", metricString, v.Ip, vv, v.Value))
				buff.WriteString(fmt.Sprintf("%s{target=\"%s\",guid=\"%s\"} %.3f \n", metricTimeString, v.Ip, vv, v.UseTime))
			}
		} else {
			buff.WriteString(fmt.Sprintf("%s{target=\"%s\"} %d \n", metricString, v.Ip, v.Value))
			buff.WriteString(fmt.Sprintf("%s{target=\"%s\"} %.3f \n", metricTimeString, v.Ip, v.UseTime))
		}
	}
	return buff.Bytes()
}

func getLossPingExportMetric(guidMap map[string][]string) []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem. \n")
	metricString := Config().Metrics.PingLossPercent
	var tmpExportMetric exportMetricList
	exportPingLock.RLock()
	for k, v := range exportLossPingMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Ip: k, Value: v.Value, Note: v.Note})
	}
	exportPingLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _, v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.PingCountNum || v.Ip == Config().Metrics.PingCountSuccess || v.Ip == Config().Metrics.PingCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \n", v.Ip, v.Value))
			continue
		}
		if len(guidMap[v.Ip]) > 0 {
			for _, vv := range guidMap[v.Ip] {
				buff.WriteString(fmt.Sprintf("%s{target=\"%s\",guid=\"%s\"} %d \n", metricString, v.Ip, vv, v.Value))
			}
		} else {
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
	for _, v := range exportTelnetMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Ip: v.Ip, Port: v.Port, Value: v.Value, Note: v.Note})
	}
	exportTelnetLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _, v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.TelnetCountNum || v.Ip == Config().Metrics.TelnetCountSuccess || v.Ip == Config().Metrics.TelnetCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \n", v.Ip, v.Value))
			continue
		}
		tmpIpPort := fmt.Sprintf("%s:%d", v.Ip, v.Port)
		if len(guidMap[tmpIpPort]) > 0 {
			for _, vv := range guidMap[tmpIpPort] {
				buff.WriteString(fmt.Sprintf("%s{ip=\"%s\",port=\"%d\",guid=\"%s\"} %d \n", metricString, v.Ip, v.Port, vv, v.Value))
			}
		} else {
			buff.WriteString(fmt.Sprintf("%s{ip=\"%s\",port=\"%d\"} %d \n", metricString, v.Ip, v.Port, v.Value))
		}
	}
	return buff.Bytes()
}

func getHttpCheckExportMetric(guidMap map[string][]string) []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP http check 1 -> request error, 2 -> response error, other -> http response code \n")
	metricString := Config().Metrics.HttpCheck
	var tmpExportMetric exportMetricList
	exportHttpCheckLock.RLock()
	for _, v := range exportHttpCheckMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Url: v.Url, Method: v.Method, Ip: v.Ip, Value: v.Value, Note: v.Note})
	}
	exportHttpCheckLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _, v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Url == Config().Metrics.HttpCheckCountNum || v.Url == Config().Metrics.HttpCheckCountSuccess || v.Url == Config().Metrics.HttpCheckCountFail {
			buff.WriteString(fmt.Sprintf("%s %d \n", v.Url, v.Value))
			continue
		}
		tmpMethodUrl := fmt.Sprintf("%s_%s", v.Method, v.Url)
		if len(guidMap[tmpMethodUrl]) > 0 {
			for _, vv := range guidMap[tmpMethodUrl] {
				buff.WriteString(fmt.Sprintf("%s{url=\"%s\",method=\"%s\",guid=\"%s\"} %d \n", metricString, v.Url, v.Method, vv, v.Value))
			}
		} else {
			buff.WriteString(fmt.Sprintf("%s{url=\"%s\",method=\"%s\"} %d \n", metricString, v.Url, v.Method, v.Value))
		}
	}
	return buff.Bytes()
}

type exportMetricList []*exportMetricObj

func (p exportMetricList) Len() int {
	return len(p)
}

func (p exportMetricList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p exportMetricList) Less(i, j int) bool {
	return p[i].Ip < p[j].Ip
}
