package models

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
}

type CustomChartSeriesDto struct {
	Endpoint      string            `json:"endpoint"`      // 监控对象
	ServiceGroup  string            `json:"serviceGroup"`  // 层级对象
	EndpointName  string            `json:"endpointName" ` // 层级对象
	MonitorType   string            `json:"monitorType" `  // 监控类型
	ColorGroup    string            `json:"colorGroup" `   // 默认色系
	PieDisplayTag string            `json:"pieDisplayTag"` // 饼图展示标签
	Metric        string            `json:"metric"`        // 指标
	Tags          []*TagDto         `json:"tags"`          // 标签
	ColorConfig   []*ColorConfigDto `json:"series"`        // 颜色
}

type TagDto struct {
	TagName  string   `json:"tagName"`  // 标签名称
	TagValue []string `json:"tagValue"` // 标签值
}

type ColorConfigDto struct {
	SeriesName string `json:"seriesName"`
	Color      string `json:"color"`
}
