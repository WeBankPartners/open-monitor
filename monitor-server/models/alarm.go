package models

import "time"

type GrpTable struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Parent       int       `json:"parent"`
	Description  string    `json:"description"`
	CreateUser   string    `json:"create_user"`
	UpdateUser   string    `json:"update_user"`
	EndpointType string    `json:"endpoint_type"`
	CreateAt     time.Time `json:"create_at"`
	UpdateAt     time.Time `json:"update_at"`
}

type TplTable struct {
	Id         int       `json:"id"`
	GrpId      int       `json:"grp_id"`
	EndpointId int       `json:"endpoint_id"`
	NotifyUrl  string    `json:"notify_url"`
	CreateUser string    `json:"create_user"`
	UpdateUser string    `json:"update_user"`
	ActionUser string    `json:"action_user"`
	ActionRole string    `json:"action_role"`
	ExtraMail  string    `json:"extra_mail"`
	ExtraPhone string    `json:"extra_phone"`
	CreateAt   time.Time `json:"create_at"`
	UpdateAt   time.Time `json:"update_at"`
}

type StrategyTable struct {
	Id           int    `json:"id"`
	TplId        int    `json:"tpl_id"`
	Metric       string `json:"metric" binding:"required"`
	Expr         string `json:"expr" binding:"required"`
	Cond         string `json:"cond" binding:"required"`
	Last         string `json:"last" binding:"required"`
	Priority     string `json:"priority" binding:"required"`
	Content      string `json:"content" binding:"required"`
	ConfigType   string `json:"config_type"`
	NotifyEnable int    `json:"notify_enable"`
	NotifyDelay  int    `json:"notify_delay"`
}

type AlarmTable struct {
	Id            int       `json:"id"`
	StrategyId    int       `json:"strategy_id"`
	Endpoint      string    `json:"endpoint"`
	Status        string    `json:"status"`
	SMetric       string    `json:"s_metric"`
	SExpr         string    `json:"s_expr"`
	SCond         string    `json:"s_cond"`
	SLast         string    `json:"s_last"`
	SPriority     string    `json:"s_priority"`
	Content       string    `json:"content"`
	Tags          string    `json:"tags"`
	StartValue    float64   `json:"start_value"`
	Start         time.Time `json:"start"`
	EndValue      float64   `json:"end_value"`
	End           time.Time `json:"end"`
	CloseType     string    `json:"close_type"`
	CloseMsg      string    `json:"close_msg"`
	CloseUser     string    `json:"close_user"`
	CustomMessage string    `json:"custom_message"`
	EndpointTags  string    `json:"endpoint_tags"`
	AlarmStrategy string    `json:"alarm_strategy"`
}

type AlarmHandleObj struct {
	AlarmTable
	NotifyEnable int `json:"notify_enable"`
	NotifyDelay  int `json:"notify_delay"`
}

type AlarmProblemQuery struct {
	Id            int       `json:"id"`
	StrategyId    int       `json:"strategy_id"`
	Endpoint      string    `json:"endpoint"`
	Status        string    `json:"status"`
	SMetric       string    `json:"s_metric"`
	SExpr         string    `json:"s_expr"`
	SCond         string    `json:"s_cond"`
	SLast         string    `json:"s_last"`
	SPriority     string    `json:"s_priority"`
	Content       string    `json:"content"`
	Tags          string    `json:"tags"`
	StartValue    float64   `json:"start_value"`
	Start         time.Time `json:"start"`
	StartString   string    `json:"start_string"`
	EndValue      float64   `json:"end_value"`
	End           time.Time `json:"end"`
	EndString     string    `json:"end_string"`
	IsLogMonitor  bool      `json:"is_log_monitor"`
	Path          string    `json:"path"`
	Keyword       string    `json:"keyword"`
	IsCustom      bool      `json:"is_custom"`
	CloseType     string    `json:"close_type"`
	CloseMsg      string    `json:"close_msg"`
	CloseUser     string    `json:"close_user"`
	CustomMessage string    `json:"custom_message"`
	EndpointTags  string    `json:"endpoint_tags"`
	AlarmStrategy string    `json:"alarm_strategy"`
	Title         string    `json:"title"`
	SystemId      string    `json:"system_id"`
}

type UpdateAlarmCustomMessageDto struct {
	Id       int    `json:"id" binding:"required"`
	IsCustom bool   `json:"is_custom"`
	Message  string `json:"message"`
}

type AlarmProblemQueryResult struct {
	Data  AlarmProblemList        `json:"data"`
	High  int                     `json:"high"`
	Mid   int                     `json:"mid"`
	Low   int                     `json:"low"`
	Count []*AlarmProblemCountObj `json:"count"`
	Page  *PageInfo               `json:"page"`
}

type AlarmProblemCountObj struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Value      int    `json:"value"`
	FilterType string `json:"filterType"`
}

type AlarmProblemCountList []*AlarmProblemCountObj

func (s AlarmProblemCountList) Len() int {
	return len(s)
}

func (s AlarmProblemCountList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s AlarmProblemCountList) Less(i, j int) bool {
	return s[i].Name+s[i].Type > s[j].Name+s[j].Type
}

type AlarmProblemList []*AlarmProblemQuery

func (s AlarmProblemList) Len() int {
	return len(s)
}

func (s AlarmProblemList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s AlarmProblemList) Less(i, j int) bool {
	return s[i].Start.Unix() > s[j].Start.Unix()
}

type AlarmHistoryReturnData struct {
	Endpoint    string           `json:"endpoint"`
	ProblemList AlarmProblemList `json:"problem_list"`
}

type GrpEndpointTable struct {
	GrpId      int `json:"grp_id"`
	EndpointId int `json:"endpoint_id"`
}

type GrpQuery struct {
	Id        int
	Name      string
	Search    string
	User      string
	Page      int
	Size      int
	Result    []*GrpTable
	ResultNum int
}

type UpdateGrp struct {
	Groups      []*GrpTable
	Operation   string
	OperateUser string
}

type TableData struct {
	Data interface{} `json:"data"`
	Page int         `json:"page"`
	Size int         `json:"size"`
	Num  int         `json:"num"`
}

type AlarmEndpointQuery struct {
	Search    string
	Page      int
	Size      int
	Grp       int
	Result    []*AlarmEndpointObj
	ResultNum int
}

type AlarmEndpointObj struct {
	Id        string      `json:"id"`
	Guid      string      `json:"guid"`
	Type      string      `json:"type"`
	GroupsIds string      `json:"groups_ids"`
	Tags      string      `json:"tags"`
	Groups    []*GrpTable `json:"groups"`
}

type GrpEndpointParam struct {
	Grp       int      `json:"grp" binding:"required"`
	Endpoints []string `json:"endpoints"`
}

type GrpEndpointParamNew struct {
	Grp       int    `json:"grp" binding:"required"`
	Endpoints []int  `json:"endpoints"`
	Operation string `json:"operation" binding:"required"`
}

type EndpointGrpParam struct {
	EndpointId int   `json:"endpoint_id"`
	GroupIds   []int `json:"group_ids"`
}

type AcceptObj struct {
	Employ []string `json:"employ"`
}

type TplObj struct {
	TplId      int              `json:"tpl_id"`
	ObjId      int              `json:"obj_id"`
	ObjName    string           `json:"obj_name"`
	ObjType    string           `json:"obj_type"`
	Operation  bool             `json:"operation"`
	Accept     []*OptionModel   `json:"accept"`
	Strategy   []*StrategyTable `json:"strategy"`
	LogMonitor []*LogMonitorDto `json:"log_monitor"`
}

type TplQuery struct {
	SearchType string    `json:"search_type"`
	SearchId   int       `json:"search_id"`
	Tpl        []*TplObj `json:"tpl"`
}

type TplStrategyTable struct {
	TplId        int    `json:"tpl_id"`
	GrpId        int    `json:"grp_id"`
	EndpointId   int    `json:"endpoint_id"`
	StrategyId   int    `json:"strategy_id"`
	Metric       string `json:"metric" binding:"required"`
	Expr         string `json:"expr" binding:"required"`
	Cond         string `json:"cond" binding:"required"`
	Last         string `json:"last" binding:"required"`
	Priority     string `json:"priority" binding:"required"`
	Content      string `json:"content" binding:"required"`
	NotifyEnable int    `json:"notify_enable"`
	NotifyDelay  int    `json:"notify_delay"`
}

type UpdateStrategy struct {
	Strategy    []*StrategyTable
	Operation   string
	OperateUser string
}

type AlterManagerRespObj struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []AMRespAlert     `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
}

type AMRespAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type LogMonitorTable struct {
	Id            int    `json:"id"`
	StrategyId    int    `json:"strategy_id"`
	Path          string `json:"path"`
	Keyword       string `json:"keyword"`
	Priority      string `json:"priority"`
	NotifyEnable  int    `json:"notify_enable"`
	OwnerEndpoint string `json:"owner_endpoint"`
}

type LogMonitorDto struct {
	Id            int                      `json:"id"`
	TplId         int                      `json:"tpl_id"`
	GrpId         int                      `json:"grp_id"`
	EndpointId    int                      `json:"endpoint_id"`
	Path          string                   `json:"path" binding:"required"`
	OwnerEndpoint string                   `json:"owner_endpoint"`
	Strategy      []*LogMonitorStrategyDto `json:"strategy"`
}

type LogMonitorStrategyDto struct {
	Id           int    `json:"id"`
	StrategyId   int    `json:"strategy_id"`
	Keyword      string `json:"keyword"`
	Cond         string `json:"cond"`
	Last         string `json:"last"`
	Priority     string `json:"priority"`
	NotifyEnable int    `json:"notify_enable"`
}

type UpdateLogMonitor struct {
	LogMonitor  []*LogMonitorTable
	Operation   string
	OperateUser string
}

type TplStrategyLogMonitorTable struct {
	TplId        int    `json:"tpl_id"`
	GrpId        int    `json:"grp_id"`
	LogMonitorId int    `json:"log_monitor_id"`
	EndpointId   int    `json:"endpoint_id"`
	StrategyId   int    `json:"strategy_id"`
	Expr         string `json:"expr" binding:"required"`
	Cond         string `json:"cond" binding:"required"`
	Last         string `json:"last"`
	Priority     string `json:"priority" binding:"required"`
	Path         string `json:"path"`
	Keyword      string `json:"keyword"`
}

type GrpStrategyExportObj struct {
	GrpName     string          `json:"grp_name"`
	Description string          `json:"description"`
	Strategy    []StrategyTable `json:"strategy"`
}

type GrpStrategyQuery struct {
	Name        string
	Description string
	Metric      string
	Expr        string
	Cond        string
	Last        string
	Priority    string
	Content     string
	ConfigType  string
}

type AlarmCustomTable struct {
	Id           int       `json:"id"`
	AlertInfo    string    `json:"alert_info"`
	AlertIp      string    `json:"alert_ip"`
	AlertLevel   int       `json:"alert_level"`
	AlertObj     string    `json:"alert_obj"`
	AlertTitle   string    `json:"alert_title"`
	AlertReciver string    `json:"alert_reciver"`
	RemarkInfo   string    `json:"remark_info"`
	SubSystemId  string    `json:"sub_system_id"`
	Closed       int       `json:"closed"`
	UpdateAt     time.Time `json:"update_at"`
}

type OpenAlarmObj struct {
	Id            int       `json:"id"`
	AlertInfo     string    `json:"alert_info"`
	AlertIp       string    `json:"alert_ip"`
	AlertLevel    string    `json:"alert_level"`
	AlertObj      string    `json:"alert_obj"`
	AlertTitle    string    `json:"alert_title"`
	UseUmgPolicy  string    `json:"use_umg_policy"`
	AlertWay      string    `json:"alert_way"`
	AlertReciver  string    `json:"alert_reciver"`
	RemarkInfo    string    `json:"remark_info"`
	SubSystemId   string    `json:"sub_system_id"`
	UpdateAt      time.Time `json:"update_at"`
	CustomMessage string    `json:"custom_message"`
}

type OpenAlarmRequest struct {
	AlertList []OpenAlarmObj `json:"alertList"`
}

type OpenAlarmResponse struct {
	ResultCode int    `json:"resultCode"`
	ResultMsg  string `json:"resultMsg"`
}

type UpdateActionDto struct {
	TplId  int           `json:"tpl_id" binding:"required"`
	Accept []OptionModel `json:"accept"`
}

type SyncConsulDto struct {
	IsRegister bool     `json:"is_register"`
	Guid       string   `json:"guid"`
	Ip         string   `json:"ip"`
	Port       string   `json:"port"`
	Tags       []string `json:"tags"`
	Interval   int      `json:"interval"`
}

type SyncSdConfigDto struct {
	IsRegister bool     `json:"is_register"`
	Guid       string   `json:"guid"`
	Ip         string   `json:"ip"`
	Port       string   `json:"port"`
	Tags       []string `json:"tags"`
	Step       int      `json:"step"`
	StepList   []int    `json:"step_list"`
}

type EndpointHttpTable struct {
	Id           int    `json:"id"`
	EndpointGuid string `json:"endpoint_guid"`
	Method       string `json:"method"`
	Url          string `json:"url"`
}

type LogMonitorTags struct {
	Endpoint string `json:"endpoint"`
	FilePath string `json:"file_path"`
	Keyword  string `json:"keyword"`
	Tags     string `json:"tags"`
}

type QueryProblemAlarmDto struct {
	Endpoint string    `json:"endpoint"`
	Metric   string    `json:"metric"`
	Priority string    `json:"priority"`
	Page     *PageInfo `json:"page"`
}

type QueryHistoryAlarmParam struct {
	Start    int64     `json:"start" binding:"required"`
	End      int64     `json:"end" binding:"required"`
	Filter   string    `json:"filter" binding:"required"`
	Endpoint string    `json:"endpoint"`
	Metric   string    `json:"metric"`
	Priority string    `json:"priority"`
	Page     *PageInfo `json:"page"`
}

type AlertWindowTable struct {
	Id         int    `json:"id"`
	Endpoint   string `json:"endpoint"`
	Start      string `json:"start"`
	End        string `json:"end"`
	Weekday    string `json:"weekday"`
	UpdateUser string `json:"update_user"`
}

type AlertWindowObj struct {
	Start    string   `json:"start"`
	End      string   `json:"end"`
	TimeList []string `json:"time_list"`
	Weekday  string   `json:"weekday"`
}

type AlertWindowParam struct {
	Endpoint string            `json:"endpoint" binding:"required"`
	Data     []*AlertWindowObj `json:"data"`
}

type CustomAlarmQueryParam struct {
	Enable bool
	Level  string
	Start  string
	End    string
	Status string
}

type EventTreeventNotifyDto struct {
	Type string                    `json:"type"`
	Data []*EventTreeventNodeParam `json:"data"`
}

type EventTreeventNodeParam struct {
	EventId   string `json:"event_id"`
	Status    string `json:"status"`
	Endpoint  string `json:"endpoint"`
	Message   string `json:"message"`
	StartUnix int64  `json:"start_unix"`
}

type EventTreeventResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Msg    string `json:"message"`
}

type PluginCloseAlarmRequest struct {
	RequestId string                        `json:"requestId"`
	Inputs    []*PluginCloseAlarmRequestObj `json:"inputs"`
}

type PluginCloseAlarmRequestObj struct {
	CallbackParameter string `json:"callbackParameter"`
	Guid              string `json:"guid"`
	AlarmId           string `json:"alarmId"`
	Message           string `json:"message"`
}

type PluginCloseAlarmResp struct {
	ResultCode    string                 `json:"resultCode"`
	ResultMessage string                 `json:"resultMessage"`
	Results       PluginCloseAlarmOutput `json:"results"`
}

type PluginCloseAlarmOutput struct {
	RequestId      string                       `json:"requestId"`
	AllowedOptions []string                     `json:"allowedOptions,omitempty"`
	Outputs        []*PluginCloseAlarmOutputObj `json:"outputs"`
}

type PluginCloseAlarmOutputObj struct {
	CallbackParameter string `json:"callbackParameter"`
	Guid              string `json:"guid"`
	AlarmId           string `json:"alarmId"`
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
	ErrorDetail       string `json:"errorDetail,omitempty"`
}

type AlarmCloseParam struct {
	Id     int    `json:"id"`
	Custom bool   `json:"custom"`
	Metric string `json:"metric"`
}
