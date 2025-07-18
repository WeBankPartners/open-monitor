package models

type RoleNewTable struct {
	Guid        string `json:"guid" xorm:"guid"`
	DisplayName string `json:"display_name" xorm:"display_name"`
	Email       string `json:"email" xorm:"email"`
	Phone       string `json:"phone" xorm:"phone"`
	Disable     int    `json:"disable" xorm:"disable"`
	UpdateTime  string `json:"update_time" xorm:"update_time"`
}
