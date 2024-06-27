package models

import "time"

type ClusterTableNew struct {
	Guid               string `json:"guid" xorm:"guid" binding:"required"`
	DisplayName        string `json:"display_name" xorm:"display_name" binding:"required"`
	RemoteAgentAddress string `json:"remote_agent_address" xorm:"remote_agent_address" binding:"required"`
	PrometheusAddress  string `json:"prometheus_address" xorm:"prometheus_address" binding:"required"`
}

type MonitorTypeTable struct {
	Guid        string `json:"guid" xorm:"guid" binding:"required"`
	DisplayName string `json:"display_name" xorm:"display_name"`
	Description string `json:"description" xorm:"description"`
}

type EndpointNewTable struct {
	Guid            string    `json:"guid" xorm:"guid"`
	Name            string    `json:"name" xorm:"name"`
	Ip              string    `json:"ip" xorm:"ip"`
	MonitorType     string    `json:"monitor_type" xorm:"monitor_type"`
	AgentVersion    string    `json:"agent_version" xorm:"agent_version"`
	AgentAddress    string    `json:"agent_address" xorm:"agent_address"`
	Step            int       `json:"step" xorm:"step"`
	EndpointVersion string    `json:"endpoint_version" xorm:"endpoint_version"`
	EndpointAddress string    `json:"endpoint_address" xorm:"endpoint_address"`
	Cluster         string    `json:"cluster" xorm:"cluster"`
	AlarmEnable     int       `json:"alarm_enable" xorm:"alarm_enable"`
	Tags            string    `json:"tags" xorm:"tags"`
	ExtendParam     string    `json:"extend_param" xorm:"extend_param"`
	Description     string    `json:"description" xorm:"description"`
	UpdateTime      time.Time `json:"update_time" xorm:"update_time"`
}

type EndpointExtendParamObj struct {
	Enable        bool   `json:"-"`
	Ip            string `json:"ip,omitempty"`
	Port          string `json:"port,omitempty"`
	User          string `json:"user,omitempty"`
	Password      string `json:"password,omitempty"`
	BinPath       string `json:"bin_path,omitempty"`
	ConfigPath    string `json:"config_path,omitempty"`
	HttpMethod    string `json:"http_method,omitempty"`
	HttpUrl       string `json:"http_url,omitempty"`
	ProcessName   string `json:"process_name,omitempty"`
	ProcessTags   string `json:"process_tags,omitempty"`
	ExportAddress string `json:"export_address,omitempty"`
	ProxyExporter string `json:"proxy_exporter,omitempty"`
}

type MetricTable struct {
	Guid               string `json:"guid" xorm:"guid"`
	Metric             string `json:"metric" xorm:"metric"`
	MonitorType        string `json:"monitor_type" xorm:"monitor_type"`
	PromExpr           string `json:"prom_expr" xorm:"prom_expr"`
	TagOwner           string `json:"tag_owner" xorm:"tag_owner"`
	ServiceGroup       string `json:"service_group" xorm:"service_group"`
	Workspace          string `json:"workspace" xorm:"workspace"`
	CreateTime         string `json:"create_time" xorm:"create_time"`
	UpdateTime         string `json:"update_time" xorm:"update_time"`
	CreateUser         string `json:"create_user" xorm:"create_user"`
	UpdateUser         string `json:"update_user" xorm:"update_user"`
	LogMetricConfig    string `json:"log_metric_config" xorm:"log_metric_config"`
	LogMetricTemplate  string `json:"log_metric_template" xorm:"log_metric_template"`
	LogMetricGroup     string `json:"log_metric_group" xorm:"log_metric_group"`
	MetricType         string `json:"metric_type" xorm:"-"`           // 指标类型
	LogMetricGroupName string `json:"log_metric_group_name" xorm:"-"` // 配置的模版名
}

type MetricImportResultDto struct {
	SuccessList []string `json:"success_list"` // 成功
	FailList    []string `json:"fail_list"`    // 失败
	Message     string   `json:"message"`      // 描述
}

type MetricComparisonExtend struct {
	MetricId           string `json:"metricId" xorm:"metric_id"` // 指标Id
	Metric             string `json:"metric" xorm:"metric"`
	MonitorType        string `json:"monitor_type" xorm:"monitor_type"`
	TagOwner           string `json:"tag_owner" xorm:"tag_owner"`
	ServiceGroup       string `json:"service_group" xorm:"service_group"`
	Workspace          string `json:"workspace" xorm:"workspace"`
	CreateTime         string `json:"create_time" xorm:"create_time"`
	UpdateTime         string `json:"update_time" xorm:"update_time"`
	CreateUser         string `json:"create_user" xorm:"create_user"`
	UpdateUser         string `json:"update_user" xorm:"update_user"`
	LogMetricConfig    string `json:"log_metric_config" xorm:"log_metric_config"`
	LogMetricTemplate  string `json:"log_metric_template" xorm:"log_metric_template"`
	LogMetricGroup     string `json:"log_metric_group" xorm:"log_metric_group"`
	MetricType         string `json:"metric_type" xorm:"-"`           // 指标类型
	LogMetricGroupName string `json:"log_metric_group_name" xorm:"-"` // 配置的模版名
	PromExpr           string `json:"promExpr" xorm:"prom_expr"`      // 同环比指标prom表达式
	CalcType           string `json:"calcType" xorm:"calc_type"`      // 计算数值: diff 差值,diff_percent 差值百分比
	CalcMethod         string `json:"calcMethod" xorm:"calc_method"`  // 计算方法: avg平均,sum求和
	CalcPeriod         string `json:"calcPeriod" xorm:"calc_period"`  // 计算周期
}
