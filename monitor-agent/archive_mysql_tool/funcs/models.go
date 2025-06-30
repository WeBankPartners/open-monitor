package funcs

import "fmt"

type PrometheusResponse struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}

type PrometheusData struct {
	Result     []PrometheusResult `json:"result"`
	ResultType string             `json:"resultType"`
}

type PrometheusResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

type PrometheusQueryObj struct {
	Start  int64           `json:"start"`
	End    int64           `json:"end"`
	Metric DefaultSortList `json:"metric"`
	Values [][]float64     `json:"values"`
}

type PrometheusQueryParam struct {
	Start  int64                 `json:"start"`
	End    int64                 `json:"end"`
	PromQl string                `json:"prom_ql"`
	Data   []*PrometheusQueryObj `json:"data"`
}

type ArchiveTable struct {
	Endpoint   string  `json:"endpoint"`
	Metric     string  `json:"metric"`
	Tags       string  `json:"tags"`
	UnixTime   int64   `json:"unix_time"`
	Avg        float64 `json:"avg"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	P95        float64 `json:"p_95"`
	Sum        float64 `json:"sum"`
	CreateTime string  `json:"create_time"`
}

type ArchiveCountQueryObj struct {
	Endpoint string `json:"endpoint"`
	Metric   string `json:"metric"`
}

type PrometheusArchiveTables struct {
	TableName string `xorm:"TABLE_NAME"`
}

type MonitorEndpointTable struct {
	Guid       string `json:"guid"`
	ExportType string `json:"export_type"`
	Step       int    `json:"step"`
	Address    string `json:"address"`
}

type MonitorPromMetricTable struct {
	Metric     string `json:"metric"`
	MetricType string `json:"metric_type"`
	PromQl     string `json:"prom_ql"`
}

type MonitorMetricTable struct {
	Guid         string `json:"guid" xorm:"guid"`
	Metric       string `json:"metric" xorm:"metric"`
	MonitorType  string `json:"monitor_type" xorm:"monitor_type"`
	PromExpr     string `json:"prom_expr" xorm:"prom_expr"`
	TagOwner     string `json:"tag_owner" xorm:"tag_owner"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
	Workspace    string `json:"workspace" xorm:"workspace"`
	UpdateTime   string `json:"update_time" xorm:"update_time"`
}

type MonitorArchiveObj struct {
	Endpoint string                    `json:"endpoint"`
	Metrics  []*MonitorPromMetricTable `json:"metrics"`
}

type ArchiveActionParamObj struct {
	Endpoint  string `json:"endpoint"`
	Metric    string `json:"metric"`
	PromQl    string `json:"prom_ql"`
	TableName string `json:"table_name"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
}

type ArchiveActionList []*ArchiveActionParamObj

type DefaultSortObj struct {
	Key   string
	Value string
}

type DefaultSortList []*DefaultSortObj

func (s DefaultSortList) Len() int {
	return len(s)
}

func (s DefaultSortList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DefaultSortList) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func (s DefaultSortList) ToTagString() string {
	var output string
	for i, v := range s {
		output += fmt.Sprintf("%s=\"%s\"", v.Key, v.Value)
		if i < len(s)-1 {
			output += ","
		}
	}
	return output
}

type HttpRespJson struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ArchiveFiveRowObj struct {
	Endpoint string    `json:"endpoint"`
	Metric   string    `json:"metric"`
	Tags     string    `json:"tags"`
	UnixTime int64     `json:"unix_time"`
	Avg      []float64 `json:"avg"`
	Min      []float64 `json:"min"`
	Max      []float64 `json:"max"`
	P95      []float64 `json:"p_95"`
	Sum      []float64 `json:"sum"`
}

func (a ArchiveFiveRowObj) CalcArchiveTable() ArchiveTable {
	tmpAvg, _, _, _, _ := calcData(a.Avg)
	_, tmpMin, _, _, _ := calcData(a.Min)
	_, _, tmpMax, _, _ := calcData(a.Max)
	_, _, _, tmpP95, _ := calcData(a.P95)
	_, _, _, _, tmpSum := calcData(a.Sum)
	return ArchiveTable{Endpoint: a.Endpoint, Metric: a.Metric, Tags: a.Tags, UnixTime: a.UnixTime, Avg: tmpAvg, Min: tmpMin, Max: tmpMax, P95: tmpP95, Sum: tmpSum}
}

type JobRecordTable struct {
	Id      int    `json:"id" xorm:"id"`
	JobTime string `json:"job_time" xorm:"job_time"`
	HostIp  string `json:"host_ip" xorm:"host_ip"`
}
