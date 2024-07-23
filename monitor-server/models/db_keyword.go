package models

import "time"

type DbKeywordMonitor struct {
	Guid         string    `json:"guid" xorm:"guid"`                  // 唯一标识
	ServiceGroup string    `json:"serviceGroup" xorm:"service_group"` // 业务监控组
	Name         string    `json:"name" xorm:"name"`                  // 名称
	QuerySql     string    `json:"querySql" xorm:"query_sql"`         // 查询sql
	Priority     string    `json:"priority" xorm:"priority"`          // 告警级别
	Content      string    `json:"content" xorm:"content"`            // 告警内容
	NotifyEnable int8      `json:"notifyEnable" xorm:"notify_enable"` // 是否通知
	ActiveWindow string    `json:"activeWindow" xorm:"active_window"` // 生效时间段
	Step         int       `json:"step" xorm:"step"`                  // 采集间隔
	MonitorType  string    `json:"monitorType" xorm:"monitor_type"`   // 监控类型
	CreateUser   string    `json:"createUser" xorm:"create_user"`     // 创建人
	UpdateUser   string    `json:"updateUser" xorm:"update_user"`     // 更新人
	CreateTime   time.Time `json:"createTime" xorm:"create_time"`     // 创建时间
	UpdateTime   time.Time `json:"updateTime" xorm:"update_time"`     // 更新时间
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
	Guid              string `json:"guid" xorm:"guid"`                             // 唯一标识
	LogKeywordMonitor string `json:"logKeywordMonitor" xorm:"log_keyword_monitor"` // 业务关键字监控
	LogKeywordConfig  string `json:"logKeywordConfig" xorm:"log_keyword_config"`   // 业务关键字配置
	Notify            string `json:"notify" xorm:"notify"`                         // 通知表
}

type DbKeywordEndpointRel struct {
	Guid             string `json:"guid" xorm:"guid"`                           // 唯一标识
	DbKeywordMonitor string `json:"dbKeywordMonitor" xorm:"db_keyword_monitor"` // 数据库关键字监控
	SourceEndpoint   string `json:"sourceEndpoint" xorm:"source_endpoint"`      // 源对象
	TargetEndpoint   string `json:"targetEndpoint" xorm:"target_endpoint"`      // 目标对象
}

type DbKeywordNotifyRel struct {
	Guid             string `json:"guid" xorm:"guid"`                           // 唯一标识
	DbKeywordMonitor string `json:"dbKeywordMonitor" xorm:"db_keyword_monitor"` // 数据库关键字监控
	Notify           string `json:"notify" xorm:"notify"`                       // 通知表
}
