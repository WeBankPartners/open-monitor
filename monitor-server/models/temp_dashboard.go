package models

type HistoryChart struct {
	PanalTitle string      `json:"panalTitle"`
	PanalUnit  string      `json:"panalUnit"`
	ChartType  string      `json:"chartType"`
	LineType   int         `json:"lineType"`
	Aggregate  string      `json:"aggregate"`
	AggStep    int         `json:"agg_step"`
	Query      []*Query    `json:"query"`
	ViewConfig *ViewConfig `json:"viewConfig"`
}
type MetricToColor struct {
	Metric string `json:"metric"`
	Color  string `json:"color"`
}
type Query struct {
	Endpoint      string           `json:"endpoint"`
	Metric        string           `json:"metric"`
	ChartType     string           `json:"chartType"`
	LineType      int              `json:"lineType"`
	Aggregate     string           `json:"aggregate"`
	AggStep       int              `json:"agg_step"`
	EndpointType  string           `json:"endpoint_type"`
	AppObject     string           `json:"app_object"`
	MetricToColor []*MetricToColor `json:"metricToColor"`
	EndpointName  string           `json:"endpointName"`
	DefaultColor  string           `json:"defaultColor"`
}
type Data struct {
	Endpoint      string           `json:"endpoint"`
	Metric        string           `json:"metric"`
	ChartType     string           `json:"chartType"`
	LineType      int              `json:"lineType"`
	Aggregate     string           `json:"aggregate"`
	AggStep       int              `json:"agg_step"`
	EndpointType  string           `json:"endpoint_type"`
	AppObject     string           `json:"app_object"`
	MetricToColor []*MetricToColor `json:"metricToColor"`
	EndpointName  string           `json:"endpointName"`
	DefaultColor  string           `json:"defaultColor"`
}
type ChartParams struct {
	Aggregate  string `json:"aggregate"`
	AggStep    int    `json:"agg_step"`
	TimeSecond int    `json:"time_second"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
	Title      string `json:"title"`
	Unit       string `json:"unit"`
	Data       []Data `json:"data"`
}
type ActiveCharts struct {
	Style       string       `json:"style"`
	PanalUnit   string       `json:"panalUnit"`
	ElID        string       `json:"elId"`
	ChartParams *ChartParams `json:"chartParams"`
	ChartType   string       `json:"chartType"`
	Aggregate   string       `json:"aggregate"`
	AggStep     int          `json:"agg_step"`
	TimeSecond  int          `json:"time_second"`
	Start       int          `json:"start"`
	End         int          `json:"end"`
}
type ViewConfig struct {
	X            int             `json:"x"`
	Y            int             `json:"y"`
	W            int             `json:"w"`
	H            int             `json:"h"`
	I            string          `json:"i"`
	ID           string          `json:"id"`
	Moved        bool            `json:"moved"`
	ActiveCharts []*ActiveCharts `json:"_activeCharts"`
	Group        string          `json:"group"`
}

type NewDisplayConfig struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}
