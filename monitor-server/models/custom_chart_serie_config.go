package models

type CustomChartSeriesConfig struct {
	Guid                 string  `json:"guid" xorm:"'guid' pk"`
	DashboardChartConfig *string `json:"dashboardChartConfig" xorm:"dashboard_chart_config"` // 图表配置表
	Tags                 string  `json:"tags" xorm:"tags"`                                   // 标签
	Color                string  `json:"color" xorm:"color"`                                 // 颜色
	SeriesName           string  `json:"series_name" xorm:"series_name"`                     // 指标+对象+标签值
}
