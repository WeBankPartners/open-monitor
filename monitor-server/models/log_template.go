package models

import "time"

type LogMonitorTemplate struct {
	Guid       string    `json:"guid" xorm:"guid"`
	Name       string    `json:"name" xorm:"name"`
	MetricType string    `json:"metric_type" xorm:"metric_type"`
	CreateUser string    `json:"create_user" xorm:"create_user"`
	UpdateUser string    `json:"update_user" xorm:"update_user"`
	CreateTime time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime time.Time `json:"update_time" xorm:"update_time"`
}

type LogMonitorTemplateRole struct {
	Guid               string `json:"guid" xorm:"guid"`
	LogMonitorTemplate string `json:"log_monitor_template" xorm:"log_monitor_template"`
	Role               string `json:"role" xorm:"role"`
	Permission         string `json:"permission" xorm:"permission"`
}

type LogJsonTemplate struct {
	Guid               string    `json:"guid" xorm:"guid"`
	LogMonitorTemplate string    `json:"log_monitor_template" xorm:"log_monitor_template"`
	Name               string    `json:"name" xorm:"name"`
	JsonRegular        string    `json:"json_regular" xorm:"json_regular"`
	DemoLog            string    `json:"demo_log" xorm:"demo_log"`
	CalcResult         string    `json:"calc_result" xorm:"calc_result"`
	CreateUser         string    `json:"create_user" xorm:"create_user"`
	UpdateUser         string    `json:"update_user" xorm:"update_user"`
	CreateTime         time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime         time.Time `json:"update_time" xorm:"update_time"`
}

type LogMetricTemplate struct {
	Guid               string    `json:"guid" xorm:"guid"`
	LogMonitorTemplate string    `json:"log_monitor_template" xorm:"log_monitor_template"`
	LogJsonTemplate    string    `json:"log_json_template" xorm:"log_json_template"`
	Metric             string    `json:"metric" xorm:"metric"`
	DisplayName        string    `json:"display_name" xorm:"display_name"`
	JsonKey            string    `json:"json_key" xorm:"json_key"`
	Regular            string    `json:"regular" xorm:"regular"`
	DemoLog            string    `json:"demo_log" xorm:"demo_log"`
	CalcResult         string    `json:"calc_result" xorm:"calc_result"`
	Step               int       `json:"step" xorm:"step"`
	AggType            string    `json:"agg_type" xorm:"agg_type"`
	TagConfig          string    `json:"tag_config" xorm:"tag_config"`
	CreateUser         string    `json:"create_user" xorm:"create_user"`
	UpdateUser         string    `json:"update_user" xorm:"update_user"`
	CreateTime         time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime         time.Time `json:"update_time" xorm:"update_time"`
}

type LogMonitorTemplateDto struct {
	Guid       string                        `json:"guid" xorm:"guid"`
	Name       string                        `json:"name" xorm:"name"`
	MetricType string                        `json:"metric_type" xorm:"metric_type"`
	JsonConfig *LogJsonTemplate              `json:"json_config"`
	Metrics    []*LogMetricTemplate          `json:"metrics"`
	Permission *LogMonitorTemplatePermission `json:"permission"`
}

type LogMonitorTemplatePermission struct {
	MgmtRoles []string `json:"mgmt_roles"`
	UseRoles  []string `json:"use_roles"`
}
