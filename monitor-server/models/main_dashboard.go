package models

type MainDashboard struct {
	Guid            string `json:"id" xorm:"'guid' pk"`
	RoleId          string `json:"roleId" xorm:"role_id"`
	CustomDashboard *int   `json:"customDashboard" xorm:"custom_dashboard"` // 首页看板表
}
