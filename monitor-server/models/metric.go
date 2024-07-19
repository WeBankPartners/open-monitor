package models

type QueryMetricTagParam struct {
	MetricId     string `json:"metricId"`
	Endpoint     string `json:"endpoint"`
	ServiceGroup string `json:"serviceGroup"`
}

type QueryMetricTagResultObj struct {
	Tag    string               `json:"tag"`
	Values []*MetricTagValueObj `json:"values"`
}

type MetricTagValueObj struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MetricCountRes struct {
	Count           int `json:"count"`
	ComparisonCount int `json:"comparisonCount"`
}
