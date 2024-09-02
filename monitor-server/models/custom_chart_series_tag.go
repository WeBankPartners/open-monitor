package models

type CustomChartSeriesTag struct {
	Guid                 string  `json:"guid" xorm:"'guid' pk"`
	DashboardChartConfig *string `json:"dashboardChartConfig" xorm:"dashboard_chart_config"` // 图表配置表
	Name                 string  `json:"name" xorm:"name"`                                   // 标签名
	Equal                string  `json:"equal" xorm:"equal"`                                 // in | notin
}

type CustomChartTagValueRow struct {
	ChartGuid string `json:"chartGuid" xorm:"chart_guid"`
	Name      string `json:"name" xorm:"name"`
	Value     string `json:"value" xorm:"value"`
	Equal     string `json:"equal" xorm:"equal"` // in | notin
}
