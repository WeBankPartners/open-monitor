package models

type CustomChartPermission struct {
	Guid           string `json:"guid" xorm:"'guid' pk"`
	DashboardChart string `json:"dashboardChart" xorm:"dashboard_chart"` // 所属看板图表
	RoleId         string `json:"roleId" xorm:"role_id"`                 // 角色id
	Permission     string `json:"permission" xorm:"permission"`          // 权限,MGMT/USE
}

type SharedChartPermissionDto struct {
	UseRoles  []string `json:"useRoles"`
	MgmtRoles []string `json:"mgmtRoles"`
}

type ChartPermissionBatchParam struct {
	Ids []string `json:"ids"`
}
