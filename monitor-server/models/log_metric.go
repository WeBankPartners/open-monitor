package models

import "time"

const (
	LogMonitorJsonType    = "json"
	LogMonitorRegularType = "regular"
	LogMonitorCustomType  = "custom"
)

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
	Name             string `json:"name" xorm:"name"`
	LogMetricMonitor string `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	JsonRegular      string `json:"json_regular" xorm:"json_regular"`
	Tags             string `json:"tags" xorm:"tags"`
	UpdateTime       string `json:"update_time" xorm:"update_time"`
	DemoLog          string `json:"demo_log" xorm:"demo_log"`
	CalcResult       string `json:"calc_result" xorm:"calc_result"`
}

type LogMetricConfigTable struct {
	Guid             string    `json:"guid" xorm:"guid"`
	LogMetricMonitor string    `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	LogMetricGroup   string    `json:"log_metric_group" xorm:"log_metric_group"`
	LogParamName     string    `json:"log_param_name" xorm:"log_param_name"`
	LogMetricJson    string    `json:"log_metric_json" xorm:"log_metric_json"`
	Metric           string    `json:"metric" xorm:"metric"`
	DisplayName      string    `json:"display_name" xorm:"display_name"`
	JsonKey          string    `json:"json_key" xorm:"json_key"`
	Regular          string    `json:"regular" xorm:"regular"`
	AggType          string    `json:"agg_type" xorm:"agg_type"`
	Step             int64     `json:"step" xorm:"step"`
	TagConfig        string    `json:"-" xorm:"tag_config"`
	TagConfigList    []string  `json:"tag_config" xorm:"-"`
	CreateUser       string    `json:"create_user" xorm:"create_user"`
	UpdateUser       string    `json:"update_user" xorm:"update_user"`
	CreateTime       time.Time `json:"create_time" xorm:"create_time"`
	UpdateTime       time.Time `json:"update_time" xorm:"update_time"`
	AutoAlarm        bool      `json:"auto_alarm" xorm:"auto_alarm"`
	RangeConfig      string    `json:"range_config" xorm:"range_config"`
	ColorGroup       string    `json:"color_group" xorm:"color_group"`
	FullMetric       string    `json:"full_metric" xorm:"-"`
}

type LogMetricConfigDto struct {
	Guid             string      `json:"guid"`
	LogMetricMonitor string      `json:"log_metric_monitor"`
	LogMetricGroup   string      `json:"log_metric_group"`
	LogParamName     string      `json:"log_param_name"`
	LogMetricJson    string      `json:"log_metric_json"`
	Metric           string      `json:"metric"`
	DisplayName      string      `json:"display_name"`
	JsonKey          string      `json:"json_key"`
	TagConfig        string      `json:"-"`
	Regular          string      `json:"regular"`
	AggType          string      `json:"agg_type"`
	Step             int64       `json:"step"`
	TagConfigList    []string    `json:"tag_config"`
	CreateUser       string      `json:"create_user"`
	UpdateUser       string      `json:"update_user"`
	CreateTime       string      `json:"create_time"`
	UpdateTime       string      `json:"update_time"`
	AutoAlarm        bool        `json:"auto_alarm"`
	RangeConfig      interface{} `json:"range_config"`
	ColorGroup       string      `json:"color_group"`
	FullMetric       string      `json:"full_metric"`
}

type LogMetricStringMapTable struct {
	Guid               string `json:"guid" xorm:"guid"`
	LogMetricConfig    string `json:"log_metric_config" xorm:"log_metric_config"`
	LogMetricGroup     string `json:"log_metric_group" xorm:"log_metric_group"`
	LogMonitorTemplate string `json:"log_monitor_template" xorm:"log_monitor_template"`
	LogParamName       string `json:"log_param_name" xorm:"log_param_name"`
	ValueType          string `json:"value_type" xorm:"value_type"`
	SourceValue        string `json:"source_value" xorm:"source_value"`
	Regulative         int    `json:"regulative" xorm:"regulative"`
	TargetValue        string `json:"target_value" xorm:"target_value"`
	UpdateTime         string `json:"update_time" xorm:"update_time"`
}

type LogMetricEndpointRelTable struct {
	Guid             string `json:"guid" json:"guid"`
	LogMetricMonitor string `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	SourceEndpoint   string `json:"source_endpoint" xorm:"source_endpoint"`
	TargetEndpoint   string `json:"target_endpoint" xorm:"target_endpoint"`
}

type LogMetricQueryObj struct {
	ServiceGroupTable
	Config   []*LogMetricMonitorObj `json:"config"`
	DBConfig []*DbMetricMonitorObj  `json:"db_config"`
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
	MetricGroups     []*LogMetricGroupObj         `json:"metric_groups"`
}

type LogMetricJsonObj struct {
	Guid                   string                `json:"guid" xorm:"guid"`
	Name                   string                `json:"name" xorm:"name"`
	LogMetricMonitor       string                `json:"log_metric_monitor" xorm:"log_metric_monitor"`
	JsonRegular            string                `json:"json_regular" xorm:"json_regular"`
	Tags                   string                `json:"tags" xorm:"tags"`
	MetricList             []*LogMetricConfigObj `json:"metric_list"`
	DemoLog                string                `json:"log_sample" xorm:"demo_log"`
	CalcResult             string                `json:"calc_result" xorm:"calc_result"`
	TrialCalculationResult []string              `json:"trialCalculationResult"`
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
	ServiceGroup     string                     `json:"service_group"`
	MonitorType      string                     `json:"monitor_type"`
	TagConfig        []*LogMetricConfigTag      `json:"tag_config" xorm:"tag_config"`
	JsonTagList      []string                   `json:"json_tag_list"`
}

type LogMetricConfigTag struct {
	Key          string `json:"key"`
	Regular      string `json:"regular"`
	LogParamName string `json:"log_param_name"`
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
	Path              string                 `json:"path"`
	TargetEndpoint    string                 `json:"target_endpoint"`
	ServiceGroup      string                 `json:"service_group"`
	JsonConfig        []*LogMetricJsonNeObj  `json:"config"`
	MetricConfig      []*LogMetricNeObj      `json:"custom"`
	MetricGroupConfig []*LogMetricGroupNeObj `json:"metric_group_config"`
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
	TagConfig    []*LogMetricConfigTag      `json:"tag_config"`
	LogParamName string                     `json:"log_param_name"`
}

type LogMetricStringMapNeObj struct {
	Regulation        string  `json:"regulation"`
	StringValue       string  `json:"string_value"`
	IntValue          float64 `json:"int_value"`
	RegEnable         bool    `json:"reg_enable"`
	TargetStringValue string  `json:"target_string_value"`
}

type CheckRegExpParam struct {
	RegString   string `json:"reg_string" binding:"required"`
	TestContext string `json:"test_context" binding:"required"`
}

type CheckRegExpResult struct {
	MatchText   string                 `json:"match_text"`
	JsonKeyList []string               `json:"json_key_list"`
	JsonObj     map[string]interface{} `json:"json_obj"`
}

type LogMetricGroupObj struct {
	LogMonitorTemplateGuid string `json:"log_monitor_template_guid"`
	LogMetricGroup
	LogMonitorTemplateName    string                 `json:"log_monitor_template_name"`
	ServiceGroup              string                 `json:"service_group"`
	MonitorType               string                 `json:"monitor_type"`
	JsonRegular               string                 `json:"json_regular"`
	ParamList                 []*LogMetricParamObj   `json:"param_list"`
	MetricList                []*LogMetricConfigDto  `json:"metric_list"`
	AutoCreateWarn            bool                   `json:"auto_create_warn"`      //自动创建告警
	AutoCreateDashboard       bool                   `json:"auto_create_dashboard"` //自动创建自定义看板
	LogMonitorTemplateDto     *LogMonitorTemplateDto `json:"log_monitor_template_data"`
	LogMonitorTemplateVersion string                 `json:"log_monitor_template_version"`
}

type LogMetricGroupWithTemplate struct {
	Name                      string                     `json:"name"`
	LogMetricMonitorGuid      string                     `json:"log_metric_monitor_guid"`
	LogMetricGroupGuid        string                     `json:"log_metric_group_guid"`
	LogMonitorTemplateGuid    string                     `json:"log_monitor_template_guid"`
	CodeStringMap             []*LogMetricStringMapTable `json:"code_string_map"`
	RetCodeStringMap          []*LogMetricStringMapTable `json:"retcode_string_map"`
	MetricPrefixCode          string                     `json:"metric_prefix_code"`
	ServiceGroup              string                     `json:"service_group"`
	MonitorType               string                     `json:"monitor_type"`
	AutoCreateWarn            bool                       `json:"auto_create_warn"`      //自动创建告警
	AutoCreateDashboard       bool                       `json:"auto_create_dashboard"` //自动创建自定义看板
	LogMonitorTemplate        *LogMonitorTemplateDto     `json:"log_monitor_template"`
	LogMonitorTemplateVersion string                     `json:"log_monitor_template_version"`
}

type LogMetricThreshold struct {
	MetricId    string
	Metric      string
	DisplayName string
	TagConfig   []string
	*ThresholdConfig
}

type LogMetricGroupNeObj struct {
	LogMetricGroup string                 `json:"log_metric_group"`
	LogType        string                 `json:"log_type"`
	JsonRegular    string                 `json:"json_regular"`
	ParamList      []*LogMetricParamNeObj `json:"param_list"`
	MetricConfig   []*LogMetricNeObj      `json:"custom"`
}

type LogMetricParamNeObj struct {
	Name      string                     `json:"name"`
	JsonKey   string                     `json:"json_key"`
	Regular   string                     `json:"regular"`
	StringMap []*LogMetricStringMapNeObj `json:"string_map"`
}

type LogMetricDataMapMatchDto struct {
	IsRegexp bool   `json:"is_regexp"`
	Content  string `json:"content"`
	Regexp   string `json:"regexp"`
	Match    bool   `json:"match"`
}

type CreateLogMetricGroupDto struct {
	AlarmList       []string `json:"alarm_list"`
	CustomDashboard string   `json:"custom_dashboard"`
}

type IdsParam struct {
	Ids []string `json:"ids"`
}
