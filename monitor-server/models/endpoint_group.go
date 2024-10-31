package models

import "time"

type EndpointGroupTable struct {
	Guid         string `json:"guid" xorm:"guid"`
	DisplayName  string `json:"display_name" xorm:"display_name"`
	Description  string `json:"description" xorm:"description"`
	MonitorType  string `json:"monitor_type" xorm:"monitor_type"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
	AlarmWindow  string `json:"alarm_window" xorm:"alarm_window"`
	UpdateTime   string `json:"update_time" xorm:"update_time"`
	CreateUser   string `json:"create_user" xorm:"create_user"`
	UpdateUser   string `json:"update_user" xorm:"update_user"`
}

type EndpointGroupRelTable struct {
	Guid          string `json:"guid" xorm:"guid"`
	Endpoint      string `json:"endpoint" xorm:"endpoint"`
	EndpointGroup string `json:"endpoint_group" xorm:"endpoint_group"`
}

type NotifyTable struct {
	Guid               string   `json:"guid" xorm:"guid"`
	EndpointGroup      string   `json:"endpoint_group" xorm:"endpoint_group"`
	ServiceGroup       string   `json:"service_group" xorm:"service_group"`
	AlarmStrategy      string   `json:"alarm_strategy" xorm:"alarm_strategy"`
	AlarmAction        string   `json:"alarm_action" xorm:"alarm_action"`
	AlarmPriority      string   `json:"alarm_priority" xorm:"alarm_priority"`
	NotifyNum          int      `json:"notify_num" xorm:"notify_num"`
	ProcCallbackName   string   `json:"proc_callback_name" xorm:"proc_callback_name"`
	ProcCallbackKey    string   `json:"proc_callback_key" xorm:"proc_callback_key"`
	CallbackUrl        string   `json:"callback_url" xorm:"callback_url"`
	CallbackParam      string   `json:"callback_param" xorm:"callback_param"`
	ProcCallbackMode   string   `json:"proc_callback_mode" xorm:"proc_callback_mode"` // 回调模式 -> manual(手动) | auto(自动)
	Description        string   `json:"description" xorm:"description"`
	AffectServiceGroup []string `json:"-" xorm:"-"`
}

type NotifyRoleRelTable struct {
	Guid   string `json:"guid" xorm:"guid"`
	Notify string `json:"notify" xorm:"notify"`
	Role   string `json:"role" xorm:"role"`
}

type NotifyObj struct {
	Guid             string   `json:"guid" xorm:"guid"`
	EndpointGroup    string   `json:"endpoint_group" xorm:"endpoint_group"`
	ServiceGroup     string   `json:"service_group" xorm:"service_group"`
	AlarmStrategy    string   `json:"alarm_strategy" xorm:"alarm_strategy"`
	AlarmAction      string   `json:"alarm_action" xorm:"alarm_action"`
	AlarmPriority    string   `json:"alarm_priority" xorm:"alarm_priority"`
	NotifyNum        int      `json:"notify_num" xorm:"notify_num"`
	ProcCallbackName string   `json:"proc_callback_name" xorm:"proc_callback_name"`
	ProcCallbackKey  string   `json:"proc_callback_key" xorm:"proc_callback_key"`
	CallbackUrl      string   `json:"callback_url" xorm:"callback_url"`
	CallbackParam    string   `json:"callback_param" xorm:"callback_param"`
	NotifyRoles      []string `json:"notify_roles"`
	ProcCallbackMode string   `json:"proc_callback_mode" xorm:"proc_callback_mode"` // 回调模式 -> manual(手动) | auto(自动)
	Description      string   `json:"description" xorm:"description"`
}

type PageInfo struct {
	StartIndex int `json:"startIndex"`
	PageSize   int `json:"pageSize"`
	TotalRows  int `json:"totalRows"`
}

type QueryRequestFilterObj struct {
	Name     string      `json:"name"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type QueryRequestSorting struct {
	Asc   bool   `json:"asc"`
	Field string `json:"field"`
}

type QueryRequestParam struct {
	Filters       []*QueryRequestFilterObj `json:"filters"`
	Paging        bool                     `json:"paging"`
	Pageable      *PageInfo                `json:"pageable"`
	Sorting       *QueryRequestSorting     `json:"sorting"`
	ResultColumns []string                 `json:"resultColumns"`
}

type TransFiltersParam struct {
	IsStruct   bool
	StructObj  interface{}
	Prefix     string
	KeyMap     map[string]string
	PrimaryKey string
}

type UpdateGroupEndpointParam struct {
	GroupGuid        string   `json:"group_guid"`
	EndpointGuidList []string `json:"endpoint_guid_list"`
}

type AlarmStrategyTable struct {
	Guid              string `json:"guid" xorm:"guid"`
	EndpointGroup     string `json:"endpoint_group" xorm:"endpoint_group"`
	Metric            string `json:"metric" xorm:"metric"`
	Condition         string `json:"condition" xorm:"condition"`
	Last              string `json:"last" xorm:"last"`
	Priority          string `json:"priority" xorm:"priority"`
	Content           string `json:"content" xorm:"content"`
	NotifyEnable      int    `json:"notify_enable" xorm:"notify_enable"`
	NotifyDelaySecond int    `json:"notify_delay_second" xorm:"notify_delay_second"`
	ActiveWindow      string `json:"active_window" xorm:"active_window"`
	UpdateTime        string `json:"update_time" xorm:"update_time"`
	Name              string `json:"name" xorm:"name"`
	UpdateUser        string `json:"update_user" xorm:"update_user"`
}

type AlarmStrategyMetricObj struct {
	Guid                    string       `json:"guid" xorm:"guid"`
	Name                    string       `json:"name" xorm:"name"`
	EndpointGroup           string       `json:"endpoint_group" xorm:"endpoint_group"`
	Metric                  string       `json:"metric" xorm:"metric"`
	Condition               string       `json:"condition" xorm:"condition"`
	Last                    string       `json:"last" xorm:"last"`
	Priority                string       `json:"priority" xorm:"priority"`
	Content                 string       `json:"content" xorm:"content"`
	NotifyEnable            int          `json:"notify_enable" xorm:"notify_enable"`
	NotifyDelaySecond       int          `json:"notify_delay_second" xorm:"notify_delay_second"`
	UpdateTime              string       `json:"update_time" xorm:"update_time"`
	MetricName              string       `json:"metric_name" xorm:"metric_name"`
	MetricExpr              string       `json:"metric_expr" xorm:"metric_expr"`
	MetricType              string       `json:"metric_type" xorm:"metric_type"`
	ActiveWindow            string       `json:"active_window" xorm:"active_window"`
	ConditionCrc            string       `json:"condition_crc"`
	Tags                    []*MetricTag `json:"tags"`
	UpdateUser              string       `json:"update_user" xorm:"update_user"`
	LogMetricGroup          string       `json:"log_metric_group" xorm:"log_metric_group"`
	AlarmStrategyMetricGuid string       `json:"alarm_strategy_metric_guid" xorm:"-"`
}

type GroupStrategyObj struct {
	Guid              string                  `json:"guid"`
	Name              string                  `json:"name"`
	EndpointGroup     string                  `json:"endpoint_group"`
	Metric            string                  `json:"metric"`
	MetricName        string                  `json:"metric_name"`
	Condition         string                  `json:"condition"`
	Last              string                  `json:"last"`
	Priority          string                  `json:"priority"`
	Content           string                  `json:"content"`
	NotifyEnable      int                     `json:"notify_enable"`
	NotifyDelaySecond int                     `json:"notify_delay_second"`
	ActiveWindow      string                  `json:"active_window"`
	NotifyList        []*NotifyObj            `json:"notify"`
	Conditions        []*StrategyConditionObj `json:"conditions"`
	UpdateTime        string                  `json:"update_time"`
	UpdateUser        string                  `json:"update_user"`
	LogMetricGroup    *string                 `json:"log_metric_group"`
	ActiveWindowList  []string                `json:"active_window_list"`
}

type EndpointStrategyObj struct {
	EndpointGroup string              `json:"endpoint_group"`
	DisplayName   string              `json:"display_name"`
	MonitorType   string              `json:"monitor_type"`
	ServiceGroup  string              `json:"service_group"`
	Strategy      []*GroupStrategyObj `json:"strategy"`
	NotifyList    []*NotifyObj        `json:"notify"`
}

type SysParameterTable struct {
	Guid       string `json:"guid" xorm:"guid"`
	ParamKey   string `json:"param_key" xorm:"param_key"`
	ParamValue string `json:"param_value" xorm:"param_value"`
}

type SysAlertMailParameter struct {
	SenderName   string `json:"sender_name"`
	SenderMail   string `json:"sender_mail"`
	AuthServer   string `json:"auth_server"`
	AuthPassword string `json:"auth_password"`
	SSL          string `json:"ssl"`
	AuthUser     string `json:"auth_user"`
}

type SysMetricTemplateParameter struct {
	Name     string `json:"name"`
	PromExpr string `json:"prom_expr"`
	Param    string `json:"param"`
}

type AlarmNotifyTable struct {
	Id                int       `json:"id" xorm:"id"`
	AlarmId           int       `json:"alarm_id" xorm:"alarm_id"`
	NotifyId          string    `json:"notify_id" xorm:"notify_id"`
	Endpoint          string    `json:"endpoint" xorm:"endpoint"`
	Metric            string    `json:"metric" xorm:"metric"`
	Status            string    `json:"status" xorm:"status"`
	ProcDefKey        string    `json:"proc_def_key" xorm:"proc_def_key"`
	ProcDefName       string    `json:"proc_def_name" xorm:"proc_def_name"`
	NotifyDescription string    `json:"notify_description" xorm:"notify_description"`
	ProcInsId         string    `json:"proc_ins_id" xorm:"proc_ins_id"`
	CreatedUser       string    `json:"created_user" xorm:"created_user"`
	CreatedTime       time.Time `json:"created_time" xorm:"created_time"`
	UpdatedTime       time.Time `json:"updated_time" xorm:"updated_time"`
}

type StrategyConditionObj struct {
	Metric     string       `json:"metric"`
	MetricName string       `json:"metric_name"`
	Condition  string       `json:"condition"`
	Last       string       `json:"last"`
	Tags       []*MetricTag `json:"tags"`
}

type MetricTag struct {
	TagName  string   `json:"tagName"`
	Equal    string   `json:"equal"` //  in/not in
	TagValue []string `json:"tagValue"`
}

type AlarmStrategyMetric struct {
	Guid              string    `json:"guid" xorm:"guid"`                    // 唯一标识
	AlarmStrategy     string    `json:"alarmStrategy" xorm:"alarm_strategy"` // 告警配置表
	Metric            string    `json:"metric" xorm:"metric"`                // 指标
	Condition         string    `json:"condition" xorm:"condition"`          // 条件
	Last              string    `json:"last" xorm:"last"`                    // 持续时间
	CrcHash           string    `json:"crc_hash" xorm:"crc_hash"`            // hash
	CreateTime        time.Time `json:"createTime" xorm:"create_time"`       // 创建时间
	UpdateTime        time.Time `json:"updateTime" xorm:"update_time"`       // 更新时间
	MonitorEngine     int       `json:"monitor_engine" xorm:"monitor_engine"`
	MonitorEngineExpr string    `json:"monitor_engine_expr" xorm:"monitor_engine_expr"`
}

type AlarmStrategyMetricQueryRow struct {
	Guid          string `json:"guid" xorm:"guid"`                    // 唯一标识
	AlarmStrategy string `json:"alarmStrategy" xorm:"alarm_strategy"` // 告警配置表
	Metric        string `json:"metric" xorm:"metric"`                // 指标
	Condition     string `json:"condition" xorm:"condition"`          // 条件
	Last          string `json:"last" xorm:"last"`                    // 持续时间
	CrcHash       string `json:"crc_hash" xorm:"crc_hash"`            // hash
	MetricName    string `json:"metric_name" xorm:"metric_name"`
}

type AlarmStrategyTag struct {
	Guid                string `json:"guid" xorm:"guid"`                                 // 唯一标识
	AlarmStrategyMetric string `json:"alarmStrategyMetric" xorm:"alarm_strategy_metric"` // 告警配置指标
	Name                string `json:"name" xorm:"name"`                                 // 标签名
	Equal               string `json:"equal" xorm:"equal"`                               // 标签名
}

type AlarmStrategyTagValue struct {
	Id               int    `json:"id" xorm:"id"`                               // 自增id
	AlarmStrategyTag string `json:"alarmStrategyTag" xorm:"alarm_strategy_tag"` // 告警配置标签值
	Value            string `json:"value" xorm:"value"`                         // 标签值
}

type AlarmStrategyMetricWithExpr struct {
	Guid          string `json:"guid" xorm:"guid"`
	AlarmStrategy string `json:"alarm_strategy" xorm:"alarm_strategy"`
	Metric        string `json:"metric" xorm:"metric"`
	Condition     string `json:"condition" xorm:"condition"`
	Last          string `json:"last" xorm:"last"`
	CrcHash       string `json:"crc_hash" xorm:"crc_hash"`
	MetricName    string `json:"metric_name" xorm:"metric_name"`
	MetricExpr    string `json:"metric_expr" xorm:"metric_expr"`
	MetricType    string `json:"metric_type" xorm:"metric_type"`
	MonitorEngine int    `json:"monitor_engine" xorm:"monitor_engine"`
}

type AlarmStrategyQueryParam struct {
	QueryType string `json:"queryType"`
	Guid      string `json:"guid"`
	Show      string `json:"show"`
	AlarmName string `json:"alarmName"`
}

type WorkflowDto struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Key     string `json:"key"`
}
