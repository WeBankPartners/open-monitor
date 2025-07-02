package models

import "time"

type LogKeywordMonitorTable struct {
	Guid         string `json:"guid"`
	ServiceGroup string `json:"service_group"`
	LogPath      string `json:"log_path"`
	MonitorType  string `json:"monitor_type"`
	UpdateTime   string `json:"update_time"`
}

func (LogKeywordMonitorTable) TableName() string {
	return "log_keyword_monitor"
}

type LogKeywordConfigTable struct {
	Guid              string     `json:"guid" xorm:"guid"`
	LogKeywordMonitor string     `json:"log_keyword_monitor" xorm:"log_keyword_monitor"`
	Keyword           string     `json:"keyword" xorm:"keyword"`
	Regulative        int        `json:"regulative" xorm:"regulative"`
	NotifyEnable      int        `json:"notify_enable" xorm:"notify_enable"`
	Priority          string     `json:"priority" xorm:"priority"`
	UpdateTime        string     `json:"update_time" xorm:"update_time"`
	Content           string     `json:"content" xorm:"content"`
	Name              string     `json:"name" xorm:"name"`
	ActiveWindow      string     `json:"active_window" xorm:"active_window"`
	ActiveWindowList  []string   `json:"active_window_list" xorm:"-"`
	Notify            *NotifyObj `json:"notify" xorm:"-"`
	UpdateUser        string     `json:"update_user" xorm:"update_user"`
}

func (LogKeywordConfigTable) TableName() string {
	return "log_keyword_config"
}

type LogKeywordEndpointRelTable struct {
	Guid              string `json:"guid"`
	LogKeywordMonitor string `json:"log_keyword_monitor"`
	SourceEndpoint    string `json:"source_endpoint"`
	TargetEndpoint    string `json:"target_endpoint"`
}

func (LogKeywordEndpointRelTable) TableName() string {
	return "log_keyword_endpoint_rel"
}

type LogKeywordServiceGroupObj struct {
	ServiceGroupTable
	Config   []*LogKeywordMonitorObj `json:"config"`
	DbConfig []*DbKeywordConfigObj   `json:"dbConfig"`
}

type LogKeywordMonitorObj struct {
	Guid         string                        `json:"guid"`
	ServiceGroup string                        `json:"service_group"`
	LogPath      string                        `json:"log_path"`
	MonitorType  string                        `json:"monitor_type"`
	KeywordList  []*LogKeywordConfigTable      `json:"keyword_list"`
	EndpointRel  []*LogKeywordEndpointRelTable `json:"endpoint_rel"`
	Notify       *NotifyObj                    `json:"notify"`
}

type LogKeywordMonitorCreateObj struct {
	Guid         string                        `json:"guid"`
	ServiceGroup string                        `json:"service_group"`
	LogPath      []string                      `json:"log_path"`
	MonitorType  string                        `json:"monitor_type"`
	KeywordList  []*LogKeywordConfigTable      `json:"keyword_list"`
	EndpointRel  []*LogKeywordEndpointRelTable `json:"endpoint_rel"`
}

type LogKeywordHttpRuleObj struct {
	RegularEnable  bool    `json:"regular_enable"`
	Keyword        string  `json:"keyword"`
	Count          float64 `json:"count"`
	TargetEndpoint string  `json:"target_endpoint"`
}

type LogKeywordHttpDto struct {
	Path     string                   `json:"path"`
	Keywords []*LogKeywordHttpRuleObj `json:"keywords"`
}

type LogKeywordFetchObj struct {
	Index   float64 `json:"index"`
	Content string  `json:"content"`
}

type LogKeywordHttpResult struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    []*LogKeywordFetchObj `json:"data"`
}

type LogKeywordCronJobQuery struct {
	Guid                 string `xorm:"guid"`
	ServiceGroup         string `xorm:"service_group"`
	LogPath              string `xorm:"log_path"`
	MonitorType          string `xorm:"monitor_type"`
	Keyword              string `xorm:"keyword"`
	NotifyEnable         int    `xorm:"notify_enable"`
	Priority             string `xorm:"priority"`
	SourceEndpoint       string `xorm:"source_endpoint"`
	TargetEndpoint       string `xorm:"target_endpoint"`
	AgentAddress         string `xorm:"agent_address"`
	Content              string `xorm:"content"`
	Name                 string `xorm:"name"`
	LogKeywordConfigGuid string `xorm:"log_keyword_config_guid"`
	ActiveWindow         string `xorm:"active_window"`
}

type LogKeywordRowsHttpDto struct {
	Path    string `json:"path"`
	Keyword string `json:"keyword"`
}

type LogKeywordRowsHttpResult struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    []*LogKeywordFetchObj `json:"data"`
}

type LogKeywordAlarmTable struct {
	Id               int       `json:"id" xorm:"id"`
	AlarmId          int       `json:"alarmId" xorm:"alarm_id"`
	Endpoint         string    `json:"endpoint" xorm:"endpoint"`
	Status           string    `json:"status" xorm:"status"`
	Content          string    `json:"content" xorm:"content"`
	Tags             string    `json:"tags" xorm:"tags"`
	StartValue       float64   `json:"startValue" xorm:"start_value"`
	EndValue         float64   `json:"endValue" xorm:"end_value"`
	UpdatedTime      time.Time `json:"updatedTime" xorm:"updated_time"`
	LogKeywordConfig string    `json:"logKeywordConfig" xorm:"log_keyword_config"`
}

type LogKeywordNotifyParam struct {
	LogKeywordMonitor string     `json:"log_keyword_monitor"`
	Notify            *NotifyObj `json:"notify"`
}
