package models

type CustomChartSeriesTagValue struct {
	Id                int     `json:"id" xorm:"'id' pk"`
	DashboardChartTag *string `json:"dashboardChartTag" xorm:"dashboard_chart_tag"`
	Value             string  `json:"value" xorm:"value"`
}
