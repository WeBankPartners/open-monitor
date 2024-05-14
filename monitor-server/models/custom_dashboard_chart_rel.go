package models

type CustomDashboardChartRel struct {
	Guid            string  `json:"guid" xorm:"'guid' pk"`
	CustomDashboard *int    `json:"customDashboard" xorm:"custom_dashboard"` // 所属看板
	DashboardChart  *string `json:"dashboardChart" xorm:"dashboard_chart"`   // 所属看板图表
	Group           string  `json:"group" xorm:"group"`                      // 所属分组
	DisplayConfig   string  `json:"displayConfig" xorm:"display_config"`     // 视图位置与长宽
	CreateUser      string  `json:"createUser" xorm:"create_user"`           // 创建人
	UpdatedUser     string  `json:"updatedUser" xorm:"updated_user"`         // 更新人
	CreateTime      string  `json:"createTime" xorm:"create_time"`           // 创建时间
	UpdateTime      string  `json:"updateTime" xorm:"update_time"`           // 更新时间
}
