package models

import "fmt"

type MetricComparisonDto struct {
	Guid           string `json:"guid"`
	Metric         string `json:"metric"`         // 原始指标名称
	MonitorType    string `json:"monitorType"`    // 原始指标类型
	ComparisonType string `json:"comparisonType"` // 对比类型: day 日环比, week 周, 月周比 month
	OriginPromExpr string `json:"originPromExpr"` // 原始指标prom表达式
	PromExpr       string `json:"promExpr"`       // 同环比指标prom表达式
	CalcType       string `json:"calcType"`
	CalcMethod     string `json:"calcMethod"`
	CalcPeriod     string `json:"calcPeriod"`
	MetricId       string `json:"metricId"` // 原始指标Id
	CreateUser     string `json:"createUser"`
	CreateTime     string `json:"createTime"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type PrometheusQueryParam struct {
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	PromQl string `json:"prom_ql"`
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

type PrometheusQueryObj struct {
	Start  int64           `json:"start"`
	End    int64           `json:"end"`
	Metric DefaultSortList `json:"metric"`
	Values [][]float64     `json:"values"`
}

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
