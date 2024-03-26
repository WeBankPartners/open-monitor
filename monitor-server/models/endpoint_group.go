package models

import "time"

type EndpointGroupTable struct {
	Guid         string    `json:"guid" xorm:"guid"`
	DisplayName  string    `json:"display_name" xorm:"display_name"`
	Description  string    `json:"description" xorm:"description"`
	MonitorType  string    `json:"monitor_type" xorm:"monitor_type"`
	ServiceGroup string    `json:"service_group" xorm:"service_group"`
	AlarmWindow  string    `json:"alarm_window" xorm:"alarm_window"`
	UpdateTime   time.Time `json:"update_time" xorm:"update_time"`
}

type EndpointGroupRelTable struct {
	Guid          string `json:"guid" xorm:"guid"`
	Endpoint      string `json:"endpoint" xorm:"endpoint"`
	EndpointGroup string `json:"endpoint_group" xorm:"endpoint_group"`
}

type NotifyTable struct {
	Guid             string `json:"guid" xorm:"guid"`
	EndpointGroup    string `json:"endpoint_group" xorm:"endpoint_group"`
	ServiceGroup     string `json:"service_group" xorm:"service_group"`
	AlarmStrategy    string `json:"alarm_strategy" xorm:"alarm_strategy"`
	AlarmAction      string `json:"alarm_action" xorm:"alarm_action"`
	AlarmPriority    string `json:"alarm_priority" xorm:"alarm_priority"`
	NotifyNum        int    `json:"notify_num" xorm:"notify_num"`
	ProcCallbackName string `json:"proc_callback_name" xorm:"proc_callback_name"`
	ProcCallbackKey  string `json:"proc_callback_key" xorm:"proc_callback_key"`
	CallbackUrl      string `json:"callback_url" xorm:"callback_url"`
	CallbackParam    string `json:"callback_param" xorm:"callback_param"`
	ProcCallbackMode string `json:"proc_callback_mode" xorm:"proc_callback_mode"` // 回调模式 -> manual(手动) | auto(自动)
	Description      string `json:"description" xorm:"description"`
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
}

type AlarmStrategyMetricObj struct {
	Guid              string `json:"guid" xorm:"guid"`
	EndpointGroup     string `json:"endpoint_group" xorm:"endpoint_group"`
	Metric            string `json:"metric" xorm:"metric"`
	Condition         string `json:"condition" xorm:"condition"`
	Last              string `json:"last" xorm:"last"`
	Priority          string `json:"priority" xorm:"priority"`
	Content           string `json:"content" xorm:"content"`
	NotifyEnable      int    `json:"notify_enable" xorm:"notify_enable"`
	NotifyDelaySecond int    `json:"notify_delay_second" xorm:"notify_delay_second"`
	UpdateTime        string `json:"update_time" xorm:"update_time"`
	MetricName        string `json:"metric_name" json:"metric_name"`
	MetricExpr        string `json:"metric_expr" json:"metric_expr"`
	MetricType        string `json:"metric_type" json:"metric_type"`
	ActiveWindow      string `json:"active_window" xorm:"active_window"`
}

type GroupStrategyObj struct {
	Guid              string       `json:"guid"`
	EndpointGroup     string       `json:"endpoint_group"`
	Metric            string       `json:"metric"`
	MetricName        string       `json:"metric_name"`
	Condition         string       `json:"condition"`
	Last              string       `json:"last"`
	Priority          string       `json:"priority"`
	Content           string       `json:"content"`
	NotifyEnable      int          `json:"notify_enable"`
	NotifyDelaySecond int          `json:"notify_delay_second"`
	ActiveWindow      string       `json:"active_window"`
	NotifyList        []*NotifyObj `json:"notify"`
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
}

type SysMetricTemplateParameter struct {
	Name     string `json:"name"`
	PromExpr string `json:"prom_expr"`
	Param    string `json:"param"`
}
