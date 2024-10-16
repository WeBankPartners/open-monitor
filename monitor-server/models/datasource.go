package models

import (
	"fmt"
	"sort"
)

type QueryMonitorData struct {
	Start                int64     `json:"start"`
	End                  int64     `json:"end"`
	Endpoint             []string  `json:"endpoint"`
	Metric               []string  `json:"metric"`
	PromQ                string    `json:"prom_q"`
	Legend               string    `json:"legend"`
	CompareLegend        string    `json:"compare_legend"`
	ChartType            string    `json:"chart_type"`
	PieData              EChartPie `json:"pie_data"`
	SameEndpoint         bool      `json:"same_endpoint"`
	Step                 int       `json:"step"`
	Cluster              string    `json:"cluster"`
	ServiceGroupName     string    `json:"service_group_name"`
	CustomDashboard      bool      `json:"custom_dashboard"`
	PieMetricType        string    `json:"pie_metric_type"`
	PieAggType           string    `json:"pie_agg_type"`
	Tags                 []string  `json:"tags"`
	PieDisplayTag        string    `json:"pie_display_tag"`
	ComparisonFlag       string    `json:"comparison_flag"`
	ServiceConfiguration string    `json:"service_configuration"` // 业务配置, custom 表示自定义
}

type PrometheusParam struct {
	Start int64  `json:"start"`
	End   int64  `json:"end"`
	Step  int64  `json:"step"`
	Query string `json:"query"`
}

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

type DataSort [][]float64

func (s DataSort) Len() int {
	return len(s)
}

func (s DataSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DataSort) Less(i, j int) bool {
	return s[i][0] < s[j][0]
}

type ArchiveQueryTable struct {
	Endpoint string  `json:"endpoint"`
	Metric   string  `json:"metric"`
	Tags     string  `json:"tags"`
	UnixTime int64   `json:"unix_time"`
	Value    float64 `json:"value"`
}

type SimpleMapObj struct {
	Key   string
	Value string
}

type PromMapSort []*SimpleMapObj

func (s PromMapSort) Len() int {
	return len(s)
}

func (s PromMapSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s PromMapSort) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func (s PromMapSort) String() string {
	sort.Sort(s)
	var name, tags string
	for _, v := range s {
		if v.Key == "__name__" {
			name = v.Value
		} else {
			tags += fmt.Sprintf("%s=\"%s\",", v.Key, v.Value)
		}
	}
	if tags != "" {
		tags = tags[:len(tags)-1]
		name = fmt.Sprintf("%s{%s}", name, tags)
	}
	return name
}

type PromSeriesResponse struct {
	Status string              `json:"status"`
	Data   []map[string]string `json:"data"`
}
