package models

type TypeConfig struct {
	Guid        string `json:"guid" xorm:"guid"`
	DisplayName string `json:"displayName"  xorm:"display_name"`
	Description string `json:"description"  xorm:"description"`
	SystemType  int    `json:"systemType" xorm:"system_type"` //系统类型,0为非系统类型,1为系统类型
	CreateUser  string `json:"createUser" xorm:"create_user"`
	CreateTime  string `json:"createTime" xorm:"create_time"`
	ObjectCount int    `json:"objectCount" xorm:"-"` // 对象数
}
