package models

type DbMetricMonitorTable struct {
	Guid string `json:"guid" xorm:"guid"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
	MetricSql string `json:"metric_sql" xorm:"metric_sql"`
	Metric string `json:"metric" xorm:"metric"`
	DisplayName string `json:"display_name" xorm:"display_name"`
	Step int `json:"step" xorm:"step"`
	MonitorType string `json:"monitor_type" xorm:"monitor_type"`
	UpdateTime string `json:"update_time" xorm:"update_time"`
}

type DbMetricEndpointRelTable struct {
	Guid string `json:"guid" xorm:"guid"`
	DbMetricMonitor string `json:"db_metric_monitor" xorm:"db_metric_monitor"`
	SourceEndpoint string `json:"source_endpoint" xorm:"source_endpoint"`
	TargetEndpoint string `json:"target_endpoint" xorm:"target_endpoint"`
}

type DbMetricMonitorObj struct {
	Guid string `json:"guid"`
	ServiceGroup string `json:"service_group"`
	ServiceGroupName string `json:"service_group_name"`
	MetricSql string `json:"metric_sql"`
	Metric string `json:"metric"`
	DisplayName string `json:"display_name"`
	Step int `json:"step"`
	MonitorType string `json:"monitor_type"`
	EndpointRel []*DbMetricEndpointRelTable `json:"endpoint_rel"`
}