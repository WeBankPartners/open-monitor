package models

import "strings"

type CustomChartSeries struct {
	Guid           string `json:"guid" xorm:"'guid' pk"`
	DashboardChart string `json:"dashboardChart" xorm:"dashboard_chart"` // 所属看板图表
	Endpoint       string `json:"endpoint" xorm:"endpoint"`              // 监控对象
	ServiceGroup   string `json:"serviceGroup" xorm:"service_group"`     // 层级对象
	EndpointName   string `json:"endpointName" xorm:"endpoint_name"`     // 层级对象
	MonitorType    string `json:"monitorType" xorm:"monitor_type"`       // 监控类型
	Metric         string `json:"metric" xorm:"metric"`                  // 指标
	ColorGroup     string `json:"colorGroup" xorm:"color_group"`         // 默认色系
	PieDisplayTag  string `json:"pieDisplayTag" xorm:"pie_display_tag"`  // 饼图展示标签
	EndpointType   string `json:"endpointType" xorm:"endpoint_type"`     // 对象类型
	MetricType     string `json:"metricType" xorm:"metric_type"`         // 指标类型
	MetricGuid     string `json:"metricGuid" xorm:"metric_guid"`         // 指标Id
}

type CustomChartSeriesDto struct {
	Guid          string            `json:"chartSeriesGuid"`
	Endpoint      string            `json:"endpoint"`      // 监控对象
	ServiceGroup  string            `json:"serviceGroup"`  // 层级对象
	EndpointName  string            `json:"endpointName" ` // 层级对象
	MonitorType   string            `json:"monitorType" `  // 监控类型
	ColorGroup    string            `json:"colorGroup" `   // 默认色系
	PieDisplayTag string            `json:"pieDisplayTag"` // 饼图展示标签
	EndpointType  string            `json:"endpointType"`  // 对象类型
	MetricType    string            `json:"metricType"`    // 指标类型
	MetricGuid    string            `json:"metricGuid"`    // 指标Id
	Metric        string            `json:"metric"`        // 指标
	Comparison    bool              `json:"comparison"`    // 指标
	Tags          []*TagDto         `json:"tags"`          // 标签
	ColorConfig   []*ColorConfigDto `json:"series"`        // 颜色
}

type TagDto struct {
	TagName  string   `json:"tagName"`  // 标签名称
	Equal    string   `json:"equal"`    //  in/not in
	TagValue []string `json:"tagValue"` // 标签值
}

type ColorConfigDto struct {
	SeriesName string `json:"seriesName"`
	Color      string `json:"color"`
	New        bool   `json:"new"`
}

type CustomChartSeriesDtoSort []*CustomChartSeriesDto

func (s CustomChartSeriesDtoSort) Len() int {
	return len(s)
}

func (s CustomChartSeriesDtoSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s CustomChartSeriesDtoSort) Less(i, j int) bool {
	return strings.Compare(s[i].Guid, s[j].Guid) < 0
}

type GetChartSeriesColorParam struct {
	ChartSeriesGuid string    `json:"chartSeriesGuid"`
	Endpoint        string    `json:"endpoint" binding:"required"`
	Metric          string    `json:"metric" binding:"required"`
	ServiceGroup    string    `json:"serviceGroup"`
	MonitorType     string    `json:"monitorType"`
	Tags            []*TagDto `json:"tags"`
}
