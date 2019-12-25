package funcs

import (
	"sync"
	"fmt"
	"bytes"
	"sort"
)

type exportMetricObj struct {
	Ip  string
	Value  int
	Note  string
}

var (
	exportMetrics = make(map[string]*exportMetricObj)
	exportLock = new(sync.RWMutex)
)

func UpdateExportMetric(result map[string]int,successCount int)  {
	exportLock.Lock()
	exportMetrics = make(map[string]*exportMetricObj)
	for k,v := range result {
		exportMetrics[k] = &exportMetricObj{Value:v, Note:fmt.Sprintf("# HELP target ip %s \r", k)}
	}
	exportMetrics[Config().Metrics.CountNum] = &exportMetricObj{Value:len(result), Note:"# HELP task ip num \r"}
	exportMetrics[Config().Metrics.CountSuccess] = &exportMetricObj{Value:successCount, Note:"# HELP alive ip num \r"}
	exportMetrics[Config().Metrics.CountFail] = &exportMetricObj{Value:len(result)-successCount, Note:"# HELP dead ip num \r"}
	exportLock.Unlock()
}

func GetExportMetric() []byte {
	var buff bytes.Buffer
	buff.WriteString("# HELP ping check 0 -> alive, 1 -> dead, 2 -> problem.")
	metricString := Config().Metrics.Default
	var tmpExportMetric exportMetricList
	exportLock.RLock()
	for k,v := range exportMetrics {
		tmpExportMetric = append(tmpExportMetric, &exportMetricObj{Ip:k, Value:v.Value, Note:v.Note})
	}
	exportLock.RUnlock()
	sort.Sort(tmpExportMetric)
	for _,v := range tmpExportMetric {
		buff.WriteString(v.Note)
		if v.Ip == Config().Metrics.CountNum || v.Ip == Config().Metrics.CountSuccess || v.Ip == Config().Metrics.CountFail {
			buff.WriteString(fmt.Sprintf("%s %d \r\n", v.Ip, v.Value))
			continue
		}
		buff.WriteString(fmt.Sprintf("%s{target:\"%s\"} %d \r\n", metricString, v.Ip, v.Value))
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
