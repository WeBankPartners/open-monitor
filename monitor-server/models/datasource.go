package models

type QueryMonitorData struct{
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Endpoint  []string  `json:"endpoint"`
	Metric  []string  `json:"metric"`
	PromQ  string  `json:"prom_q"`
	Legend  string  `json:"legend"`
	CompareLegend  string  `json:"compare_legend"`
	ChartType  string  `json:"chart_type"`
	PieData  EChartPie  `json:"pie_data"`
	SameEndpoint bool `json:"same_endpoint"`
	Step  int  `json:"step"`
	Cluster string `json:"cluster"`
}

type PrometheusParam struct {
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Step   int64  `json:"step"`
	Query  string  `json:"query"`
}

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

type DataSort [][]float64

func (s DataSort) Len() int {
	return len(s)
}

func (s DataSort) Swap(i,j int)  {
	s[i], s[j] = s[j], s[i]
}

func (s DataSort) Less(i,j int) bool {
	return s[i][0] < s[j][0]
}

type ArchiveQueryTable struct {
	Endpoint  string  `json:"endpoint"`
	Metric  string  `json:"metric"`
	Tags    string  `json:"tags"`
	UnixTime  int64  `json:"unix_time"`
	Value  float64  `json:"value"`
}