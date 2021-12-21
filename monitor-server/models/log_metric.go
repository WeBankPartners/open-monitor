package models

type LogMetricMonitorTable struct {
	Guid         string `json:"guid" xorm:"guid"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
	LogPath      string `json:"log_path" xorm:"log_path"`
	MetricType   string `json:"metric_type" xorm:"metric_type"`
	MonitorType  string `json:"monitor_type" xorm:"monitor_type"`
	UpdateTime   string `json:"update_time" xorm:"update_time"`
}

type LogMetricJsonTable struct {
	Guid             string `json:"guid" xorm:"guid"`
	LogMetricMonitor string `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	JsonRegular      string `json:"json_regular" xorm:"json_regular"`
	Tags             string `json:"tags" xorm:"tags"`
	UpdateTime       string `json:"update_time" xorm:"update_time"`
}

type LogMetricConfigTable struct {
	Guid             string `json:"guid" xorm:"guid"`
	LogMetricMonitor string `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	LogMetricJson    string `json:"log_metric_json" xorm:"log_metric_json"`
	Metric           string `json:"metric" xorm:"metric"`
	DisplayName      string `json:"display_name" xorm:"display_name"`
	JsonKey          string `json:"json_key" xorm:"json_key"`
	Regular          string `json:"regular" xorm:"regular"`
	AggType          string `json:"agg_type" xorm:"agg_type"`
	Step             int64  `json:"step" xorm:"step"`
	UpdateTime       string `json:"update_time" xorm:"update_time"`
}

type LogMetricStringMapTable struct {
	Guid            string `json:"guid" xorm:"guid"`
	LogMetricConfig string `json:"log_metric_config" xorm:"log_metric_config"`
	SourceValue     string `json:"source_value" xorm:"source_value"`
	Regulative      int    `json:"regulative" xorm:"regulative"`
	TargetValue     string `json:"target_value" xorm:"target_value"`
	UpdateTime      string `json:"update_time" xorm:"update_time"`
}

type LogMetricEndpointRelTable struct {
	Guid             string `json:"guid" json:"guid"`
	LogMetricMonitor string `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	SourceEndpoint   string `json:"source_endpoint" xorm:"source_endpoint"`
	TargetEndpoint   string `json:"target_endpoint" xorm:"target_endpoint"`
}

type LogMetricQueryObj struct {
	ServiceGroupTable
	Config []*LogMetricMonitorObj `json:"config"`
}

type LogMetricMonitorObj struct {
	Guid             string                       `json:"guid" xorm:"guid"`
	ServiceGroup     string                       `json:"service_group" xorm:"service_group"`
	LogPath          string                       `json:"log_path" xorm:"log_path"`
	MetricType       string                       `json:"metric_type" xorm:"metric_type"`
	MonitorType      string                       `json:"monitor_type" xorm:"monitor_type"`
	JsonConfigList   []*LogMetricJsonObj          `json:"json_config_list"`
	MetricConfigList []*LogMetricConfigObj        `json:"metric_config_list"`
	EndpointRel      []*LogMetricEndpointRelTable `json:"endpoint_rel"`
}

type LogMetricJsonObj struct {
	Guid             string                `json:"guid" xorm:"guid"`
	LogMetricMonitor string                `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	JsonRegular      string                `json:"json_regular" xorm:"json_regular"`
	Tags             string                `json:"tags" xorm:"tags"`
	MetricList       []*LogMetricConfigObj `json:"metric_list"`
}

type LogMetricConfigObj struct {
	Guid             string                     `json:"guid" xorm:"guid"`
	LogMetricMonitor string                     `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	LogMetricJson    string                     `json:"log_metric_json" xorm:"log_metric_json"`
	Metric           string                     `json:"metric" xorm:"metric"`
	DisplayName      string                     `json:"display_name" xorm:"display_name"`
	JsonKey          string                     `json:"json_key" xorm:"json_key"`
	Regular          string                     `json:"regular" xorm:"regular"`
	AggType          string                     `json:"agg_type" xorm:"agg_type"`
	Step             int64                      `json:"step" xorm:"step"`
	StringMap        []*LogMetricStringMapTable `json:"string_map"`
}

type LogMetricMonitorCreateDto struct {
	ServiceGroup string                       `json:"service_group" xorm:"service_group"`
	LogPath      []string                     `json:"log_path" xorm:"log_path"`
	MetricType   string                       `json:"metric_type" xorm:"metric_type"`
	MonitorType  string                       `json:"monitor_type" xorm:"monitor_type"`
	EndpointRel  []*LogMetricEndpointRelTable `json:"endpoint_rel"`
}

type LogMetricNodeExporterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type LogMetricMonitorNeObj struct {
	Path           string                `json:"path"`
	TargetEndpoint string                `json:"target_endpoint"`
	ServiceGroup   string                `json:"service_group"`
	JsonConfig     []*LogMetricJsonNeObj `json:"config"`
	MetricConfig   []*LogMetricNeObj     `json:"custom"`
}

type LogMetricJsonNeObj struct {
	Regular      string            `json:"regular"`
	Tags         string            `json:"tags"`
	MetricConfig []*LogMetricNeObj `json:"metric_config"`
}

type LogMetricNeObj struct {
	Key          string                     `json:"key"`
	Metric       string                     `json:"metric"`
	ValueRegular string                     `json:"value_regular"`
	Title        string                     `json:"title"`
	AggType      string                     `json:"agg_type"`
	Step         int64                      `json:"step"`
	StringMap    []*LogMetricStringMapNeObj `json:"string_map"`
}

type LogMetricStringMapNeObj struct {
	Regulation  string  `json:"regulation"`
	StringValue string  `json:"string_value"`
	IntValue    float64 `json:"int_value"`
	RegEnable   bool    `json:"reg_enable"`
}

type CheckRegExpParam struct {
	RegString   string `json:"reg_string" binding:"required"`
	TestContext string `json:"test_context" binding:"required"`
}
