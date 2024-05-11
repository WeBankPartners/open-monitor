package models

type QueryMetricTagParam struct {
	MetricId     string `json:"metricId"`
	Endpoint     string `json:"endpoint"`
	ServiceGroup string `json:"serviceGroup"`
}

type QueryMetricTagResultObj struct {
	Tag    string   `json:"tag"`
	Values []string `json:"values"`
}
