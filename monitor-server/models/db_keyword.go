package models

import "time"

type DbKeywordMonitor struct {
	Guid         string    `json:"guid" xorm:"guid"`                   // 唯一标识
	ServiceGroup string    `json:"service_group" xorm:"service_group"` // 业务监控组
	Name         string    `json:"name" xorm:"name"`                   // 名称
	QuerySql     string    `json:"query_sql" xorm:"query_sql"`         // 查询sql
	Priority     string    `json:"priority" xorm:"priority"`           // 告警级别
	Content      string    `json:"content" xorm:"content"`             // 告警内容
	NotifyEnable int8      `json:"notify_enable" xorm:"notify_enable"` // 是否通知
	ActiveWindow string    `json:"active_window" xorm:"active_window"` // 生效时间段
	Step         int       `json:"step" xorm:"step"`                   // 采集间隔
	MonitorType  string    `json:"monitor_type" xorm:"monitor_type"`   // 监控类型
	CreateUser   string    `json:"create_user" xorm:"create_user"`     // 创建人
	UpdateUser   string    `json:"update_user" xorm:"update_user"`     // 更新人
	CreateTime   time.Time `json:"create_time" xorm:"create_time"`     // 创建时间
	UpdateTime   time.Time `json:"update_time" xorm:"update_time"`     // 更新时间
}

type ListDbKeywordData struct {
	Guid        string                `json:"guid"`
	DisplayName string                `json:"display_name"`
	Description string                `json:"description"`
	ServiceType string                `json:"service_type"`
	UpdateTime  string                `json:"update_time"`
	Config      []*DbKeywordConfigObj `json:"config"`
}

type DbKeywordConfigObj struct {
	DbKeywordMonitor
	EndpointRel []*DbKeywordEndpointRel `json:"endpoint_rel"`
	Notify      *NotifyObj              `json:"notify"`
}

type LogKeywordNotifyRel struct {
	Guid              string `json:"guid" xorm:"guid"`                               // 唯一标识
	LogKeywordMonitor string `json:"log_keyword_monitor" xorm:"log_keyword_monitor"` // 业务关键字监控
	LogKeywordConfig  string `json:"log_keyword_config" xorm:"log_keyword_config"`   // 业务关键字配置
	Notify            string `json:"notify" xorm:"notify"`                           // 通知表
}

type DbKeywordEndpointRel struct {
	Guid             string `json:"guid" xorm:"guid"`                             // 唯一标识
	DbKeywordMonitor string `json:"db_keyword_monitor" xorm:"db_keyword_monitor"` // 数据库关键字监控
	SourceEndpoint   string `json:"source_endpoint" xorm:"source_endpoint"`       // 源对象
	TargetEndpoint   string `json:"target_endpoint" xorm:"target_endpoint"`       // 目标对象
}

type DbKeywordNotifyRel struct {
	Guid             string `json:"guid" xorm:"guid"`                             // 唯一标识
	DbKeywordMonitor string `json:"db_keyword_monitor" xorm:"db_keyword_monitor"` // 数据库关键字监控
	Notify           string `json:"notify" xorm:"notify"`                         // 通知表
}
