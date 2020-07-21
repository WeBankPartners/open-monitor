package funcs

import "fmt"

type PrometheusResponse struct {
	Status  string  `json:"status"`
	Data  PrometheusData  `json:"data"`
}

type PrometheusData struct {
	Result  []PrometheusResult  `json:"result"`
	ResultType  string  `json:"resultType"`
}

type PrometheusResult struct {
	Metric  map[string]string  `json:"metric"`
	Values  [][]interface{}  `json:"values"`
}

type PrometheusQueryObj struct {
	Start   int64  `json:"start"`
	End     int64  `json:"end"`
	Metric  DefaultSortList  `json:"metric"`
	Values  [][]float64  `json:"values"`
}

type PrometheusQueryParam struct {
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	PromQl   string  `json:"prom_ql"`
	Data   []*PrometheusQueryObj  `json:"data"`
}

type ArchiveTable struct {
	Endpoint  string  `json:"endpoint"`
	Metric    string  `json:"metric"`
	Tags      string  `json:"tags"`
	UnixTime  int64     `json:"unix_time"`
	Avg       float64 `json:"avg"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	P95       float64 `json:"p_95"`
}

type PrometheusArchiveTables struct {
	TablesInPrometheusArchive string  `xorm:"Tables_in_prometheus_archive"`
}

type MonitorEndpointTable struct {
	Guid  string  `json:"guid"`
	ExportType  string  `json:"export_type"`
	Step  int  `json:"step"`
	Address  string  `json:"address"`
	AddressAgent  string  `json:"address_agent"`
}

type MonitorPromMetricTable struct {
	Metric  string  `json:"metric"`
	MetricType  string  `json:"metric_type"`
	PromQl  string  `json:"prom_ql"`
}

type MonitorArchiveObj struct {
	Endpoint  string  `json:"endpoint"`
	Metrics   []*MonitorPromMetricTable  `json:"metrics"`
}

type ArchiveActionParamObj struct {
	Endpoint  string  `json:"endpoint"`
	Metric  string  `json:"metric"`
	PromQl  string  `json:"prom_ql"`
	TableName string `json:"table_name"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
}

type ArchiveActionList []*ArchiveActionParamObj

type DefaultSortObj struct {
	Key  string
	Value  string
}

type DefaultSortList []*DefaultSortObj

func (s DefaultSortList) Len() int {
	return len(s)
}

func (s DefaultSortList) Swap(i,j int)  {
	s[i], s[j] = s[j], s[i]
}

func (s DefaultSortList) Less(i,j int) bool {
	return s[i].Key < s[j].Key
}

func (s DefaultSortList) ToTagString() string {
	var output string
	for i,v := range s {
		output += fmt.Sprintf("%s=\"%s\"", v.Key, v.Value)
		if i < len(s)-1 {
			output += ","
		}
	}
	return output
}

type HttpRespJson struct {
	Code  int  `json:"code"`
	Msg   string    `json:"msg"`
	Data  interface{}  `json:"data"`
}