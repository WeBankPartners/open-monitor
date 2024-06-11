package models

import "time"

type LogKeywordMonitorTable struct {
	Guid         string `json:"guid"`
	ServiceGroup string `json:"service_group"`
	LogPath      string `json:"log_path"`
	MonitorType  string `json:"monitor_type"`
	UpdateTime   string `json:"update_time"`
}

type LogKeywordConfigTable struct {
	Guid              string `json:"guid"`
	LogKeywordMonitor string `json:"log_keyword_monitor"`
	Keyword           string `json:"keyword"`
	Regulative        int    `json:"regulative"`
	NotifyEnable      int    `json:"notify_enable"`
	Priority          string `json:"priority"`
	UpdateTime        string `json:"update_time"`
	Content           string `json:"content"`
	Name              string `json:"name"`
}

type LogKeywordEndpointRelTable struct {
	Guid              string `json:"guid"`
	LogKeywordMonitor string `json:"log_keyword_monitor"`
	SourceEndpoint    string `json:"source_endpoint"`
	TargetEndpoint    string `json:"target_endpoint"`
}

type LogKeywordServiceGroupObj struct {
	ServiceGroupTable
	Config []*LogKeywordMonitorObj `json:"config"`
}

type LogKeywordMonitorObj struct {
	Guid         string                        `json:"guid"`
	ServiceGroup string                        `json:"service_group"`
	LogPath      string                        `json:"log_path"`
	MonitorType  string                        `json:"monitor_type"`
	KeywordList  []*LogKeywordConfigTable      `json:"keyword_list"`
	EndpointRel  []*LogKeywordEndpointRelTable `json:"endpoint_rel"`
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
	Guid           string `xorm:"guid"`
	ServiceGroup   string `xorm:"service_group"`
	LogPath        string `xorm:"log_path"`
	MonitorType    string `xorm:"monitor_type"`
	Keyword        string `xorm:"keyword"`
	NotifyEnable   int    `xorm:"notify_enable"`
	Priority       string `xorm:"priority"`
	SourceEndpoint string `xorm:"source_endpoint"`
	TargetEndpoint string `xorm:"target_endpoint"`
	AgentAddress   string `xorm:"agent_address"`
	Content        string `xorm:"content"`
	Name           string `xorm:"name"`
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
	Id          int       `json:"id" xorm:"id"`
	AlarmId     int       `json:"alarmId" xorm:"alarm_id"`
	Endpoint    string    `json:"endpoint" xorm:"endpoint"`
	Status      string    `json:"status" xorm:"status"`
	Content     string    `json:"content" xorm:"content"`
	Tags        string    `json:"tags" xorm:"tags"`
	StartValue  float64   `json:"startValue" xorm:"start_value"`
	EndValue    float64   `json:"endValue" xorm:"end_value"`
	UpdatedTime time.Time `json:"updatedTime" xorm:"updated_time"`
}
