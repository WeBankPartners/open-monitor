package models

import "time"

type LogMonitorTemplate struct {
	Guid             string    `json:"guid" xorm:"guid"`
	Name             string    `json:"name" xorm:"name"`
	LogType          string    `json:"log_type" xorm:"log_type"`
	DemoLog          string    `json:"demo_log" xorm:"demo_log"`
	JsonRegular      string    `json:"json_regular" xorm:"json_regular"`
	CalcResult       string    `json:"-" xorm:"calc_result"`
	CreateUser       string    `json:"create_user" xorm:"create_user"`
	UpdateUser       string    `json:"update_user" xorm:"update_user"`
	CreateTime       time.Time `json:"-" xorm:"create_time"`
	CreateTimeString string    `json:"create_time"`
	UpdateTime       time.Time `json:"-" xorm:"update_time"`
	UpdateTimeString string    `json:"update_time"`
}

type LogMonitorTemplateRole struct {
	Guid               string `json:"guid" xorm:"guid"`
	LogMonitorTemplate string `json:"log_monitor_template" xorm:"log_monitor_template"`
	Role               string `json:"role" xorm:"role"`
	Permission         string `json:"permission" xorm:"permission"`
}

type LogParamTemplate struct {
	Guid               string    `json:"guid" xorm:"guid"`
	LogMonitorTemplate string    `json:"log_monitor_template" xorm:"log_monitor_template"`
	Name               string    `json:"name" xorm:"name"`
	DisplayName        string    `json:"display_name" xorm:"display_name"`
	JsonKey            string    `json:"json_key" xorm:"json_key"`
	Regular            string    `json:"regular" xorm:"regular"`
	DemoMatchValue     string    `json:"demo_match_value" xorm:"demo_match_value"`
	CreateUser         string    `json:"create_user" xorm:"create_user"`
	UpdateUser         string    `json:"update_user" xorm:"update_user"`
	CreateTime         time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime         time.Time `json:"update_time" xorm:"update_time"`
}

type LogMetricTemplate struct {
	Guid               string    `json:"guid" xorm:"guid"`
	LogMonitorTemplate string    `json:"log_monitor_template" xorm:"log_monitor_template"`
	LogParamName       string    `json:"log_param_name" xorm:"log_param_name"`
	Metric             string    `json:"metric" xorm:"metric"`
	DisplayName        string    `json:"display_name" xorm:"display_name"`
	Step               int       `json:"step" xorm:"step"`
	AggType            string    `json:"agg_type" xorm:"agg_type"`
	TagConfig          string    `json:"-" xorm:"tag_config"`
	TagConfigList      []string  `json:"tag_config" xorm:"-"`
	CreateUser         string    `json:"create_user" xorm:"create_user"`
	UpdateUser         string    `json:"update_user" xorm:"update_user"`
	CreateTime         time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime         time.Time `json:"update_time" xorm:"update_time"`
}

type LogMonitorTemplateDto struct {
	LogMonitorTemplate
	CalcResultObj *CheckRegExpResult            `json:"calc_result"`
	ParamList     []*LogParamTemplate           `json:"param_list"`
	MetricList    []*LogMetricTemplate          `json:"metric_list"`
	Permission    *LogMonitorTemplatePermission `json:"permission"`
}

type LogMonitorTemplatePermission struct {
	MgmtRoles []string `json:"mgmt_roles"`
	UseRoles  []string `json:"use_roles"`
}

type LogMonitorTemplateListResp struct {
	JsonList    []*LogMonitorTemplate `json:"json_list"`
	RegularList []*LogMonitorTemplate `json:"regular_list"`
}

type LogMonitorRegMatchParam struct {
	DemoLog   string              `json:"demo_log"`
	ParamList []*LogParamTemplate `json:"param_list"`
}

type LogMetricGroup struct {
	Guid               string    `json:"guid" xorm:"guid"`
	Name               string    `json:"name" xorm:"name"`
	LogType            string    `json:"log_type" xorm:"log_type"`
	LogMetricMonitor   string    `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	LogMonitorTemplate string    `json:"log_monitor_template" xorm:"log_monitor_template"`
	DemoLog            string    `json:"demo_log" xorm:"demo_log"`
	CalcResult         string    `json:"calc_result" xorm:"calc_result"`
	CreateUser         string    `json:"create_user" xorm:"create_user"`
	UpdateUser         string    `json:"update_user" xorm:"update_user"`
	CreateTime         time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime         time.Time `json:"update_time" xorm:"update_time"`
}

type LogMetricParam struct {
	Guid           string    `json:"guid" xorm:"guid"`
	Name           string    `json:"name" xorm:"name"`
	DisplayName    string    `json:"display_name" xorm:"display_name"`
	LogMetricGroup string    `json:"log_metric_group" xorm:"log_metric_group"`
	Regular        string    `json:"regular" xorm:"regular"`
	DemoMatchValue string    `json:"demo_match_value" xorm:"demo_match_value"`
	DemoLog        string    `json:"demo_log" xorm:"demo_log"`
	CalcResult     string    `json:"-" xorm:"calc_result"`
	CreateUser     string    `json:"create_user" xorm:"create_user"`
	UpdateUser     string    `json:"update_user" xorm:"update_user"`
	CreateTime     time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime     time.Time `json:"update_time" xorm:"update_time"`
}
