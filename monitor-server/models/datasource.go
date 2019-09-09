package models

type QueryMonitorData struct{
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Endpoint  []string  `json:"endpoint"`
	Metric  []string  `json:"metric"`
	ComputeRate  bool  `json:"compute_rate"`
	PromQ  string  `json:"prom_q"`
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