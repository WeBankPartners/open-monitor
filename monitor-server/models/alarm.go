package models

import "time"

type GrpTable struct {
	Id  int  `json:"id"`
	Name  string  `json:"name"`
	Parent  int  `json:"parent"`
	Description  string  `json:"description"`
	CreateUser  string  `json:"create_user"`
	UpdateUser  string  `json:"update_user"`
	CreateAt  time.Time  `json:"create_at"`
	UpdateAt  time.Time  `json:"update_at"`
}

type TplTable struct {
	Id  int  `json:"id"`
	GrpId  int  `json:"grp_id"`
	EndpointId  int  `json:"endpoint_id"`
	NotifyUrl  string  `json:"notify_url"`
	CreateUser  string  `json:"create_user"`
	UpdateUser  string  `json:"update_user"`
	ActionUser  string  `json:"action_user"`
	ActionRole  string  `json:"action_role"`
	ExtraMail   string  `json:"extra_mail"`
	ExtraPhone  string  `json:"extra_phone"`
	CreateAt  time.Time  `json:"create_at"`
	UpdateAt  time.Time  `json:"update_at"`
}

type StrategyTable struct {
	Id  int  `json:"id"`
	TplId  int  `json:"tpl_id"`
	Metric  string  `json:"metric" binding:"required"`
	Expr  string  `json:"expr" binding:"required"`
	Cond  string  `json:"cond" binding:"required"`
	Last  string  `json:"last" binding:"required"`
	Priority  string  `json:"priority" binding:"required"`
	Content  string  `json:"content" binding:"required"`
	ConfigType  string  `json:"config_type"`
}

type AlarmTable struct {
	Id  int  `json:"id"`
	StrategyId  int  `json:"strategy_id"`
	Endpoint  string  `json:"endpoint"`
	Status  string  `json:"status"`
	SMetric  string  `json:"s_metric"`
	SExpr  string  `json:"s_expr"`
	SCond  string  `json:"s_cond"`
	SLast  string  `json:"s_last"`
	SPriority  string  `json:"s_priority"`
	Content  string  `json:"content"`
	Tags  string  `json:"tags"`
	StartValue  float64  `json:"start_value"`
	Start  time.Time  `json:"start"`
	EndValue  float64  `json:"end_value"`
	End  time.Time  `json:"end"`
}

type AlarmProblemQuery struct {
	Id  int  `json:"id"`
	StrategyId  int  `json:"strategy_id"`
	Endpoint  string  `json:"endpoint"`
	Status  string  `json:"status"`
	SMetric  string  `json:"s_metric"`
	SExpr  string  `json:"s_expr"`
	SCond  string  `json:"s_cond"`
	SLast  string  `json:"s_last"`
	SPriority  string  `json:"s_priority"`
	Content  string  `json:"content"`
	Tags  string  `json:"tags"`
	StartValue  float64  `json:"start_value"`
	Start  time.Time  `json:"start"`
	StartString  string  `json:"start_string"`
	End  time.Time  `json:"end"`
	EndString  string  `json:"end_string"`
	IsLogMonitor  bool  `json:"is_log_monitor"`
	Path  string  `json:"path"`
	Keyword  string  `json:"keyword"`
	IsCustom  bool  `json:"is_custom"`
}

type AlarmProblemList []*AlarmProblemQuery

func (s AlarmProblemList) Len() int {
	return len(s)
}

func (s AlarmProblemList) Swap(i,j int) {
	s[i],s[j] = s[j],s[i]
}

func (s AlarmProblemList) Less(i,j int) bool {
	return s[i].Start.Unix() > s[j].Start.Unix()
}

type GrpEndpointTable struct {
	GrpId  int  `json:"grp_id"`
	EndpointId  int  `json:"endpoint_id"`
}

type GrpQuery struct {
	Id  int
	Name  string
	Search  string
	User  string
	Page  int
	Size  int
	Result  []*GrpTable
	ResultNum  int
}

type UpdateGrp struct {
	Groups  []*GrpTable
	Operation  string
	OperateUser  string
}

type TableData struct {
	Data  interface{} `json:"data"`
	Page  int  `json:"page"`
	Size  int  `json:"size"`
	Num   int  `json:"num"`
}

type AlarmEndpointQuery struct {
	Search string
	Page  int
	Size  int
	Grp  int
	Result  []*AlarmEndpointObj
	ResultNum  int
}

type AlarmEndpointObj struct {
	Id  string  `json:"id"`
	Guid  string  `json:"guid"`
	Type  string  `json:"type"`
	GroupsName  string  `json:"groups_name"`
}

type GrpEndpointParam struct {
	Grp  int  `json:"grp" binding:"required"`
	Endpoints  []string  `json:"endpoints"`
}

type GrpEndpointParamNew struct {
	Grp  int  `json:"grp" binding:"required"`
	Endpoints  []int  `json:"endpoints"`
	Operation  string  `json:"operation" binding:"required"`
}

type AcceptObj struct {
	Employ  []string  `json:"employ"`
}

type TplObj struct {
	TplId  int  `json:"tpl_id"`
	ObjId  int  `json:"obj_id"`
	ObjName  string  `json:"obj_name"`
	ObjType  string  `json:"obj_type"`
	Operation  bool  `json:"operation"`
	Accept  []*OptionModel  `json:"accept"`
	Strategy  []*StrategyTable  `json:"strategy"`
	LogMonitor  []*LogMonitorDto  `json:"log_monitor"`
}

type TplQuery struct {
	SearchType  string  `json:"search_type"`
	SearchId  int  `json:"search_id"`
	Tpl  []*TplObj  `json:"tpl"`
}

type TplStrategyTable struct {
	TplId  int  `json:"tpl_id"`
	GrpId  int  `json:"grp_id"`
	EndpointId  int  `json:"endpoint_id"`
	StrategyId  int  `json:"strategy_id"`
	Metric  string  `json:"metric" binding:"required"`
	Expr  string  `json:"expr" binding:"required"`
	Cond  string  `json:"cond" binding:"required"`
	Last  string  `json:"last" binding:"required"`
	Priority  string  `json:"priority" binding:"required"`
	Content  string  `json:"content" binding:"required"`
}

type UpdateStrategy struct {
	Strategy  []*StrategyTable
	Operation  string
	OperateUser  string
}

type AlterManagerRespObj struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []AMRespAlert `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL string `json:"externalURL"`
	Version  string `json:"version"`
	GroupKey string `json:"groupKey"`
}

type AMRespAlert struct {
	Status       string    `json:"status"`
	Labels       map[string]string    `json:"labels"`
	Annotations  map[string]string    `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}

type LogMonitorTable struct {
	Id  int  `json:"id"`
	StrategyId  int  `json:"strategy_id"`
	Path  string  `json:"path"`
	Keyword  string `json:"keyword"`
	Priority  string  `json:"priority"`
}

type LogMonitorDto struct {
	Id  int  `json:"id"`
	TplId  int  `json:"tpl_id"`
	GrpId  int  `json:"grp_id"`
	EndpointId  int  `json:"endpoint_id"`
	Path  string  `json:"path" binding:"required"`
	Strategy []*LogMonitorStrategyDto  `json:"strategy"`
}

type LogMonitorStrategyDto struct {
	Id  int  `json:"id"`
	StrategyId  int  `json:"strategy_id"`
	Keyword  string  `json:"keyword"`
	Cond  string  `json:"cond"`
	Last  string  `json:"last"`
	Priority  string  `json:"priority"`
}

type UpdateLogMonitor struct {
	LogMonitor  []*LogMonitorTable
	Operation  string
	OperateUser  string
}

type TplStrategyLogMonitorTable struct {
	TplId  int  `json:"tpl_id"`
	GrpId  int  `json:"grp_id"`
	LogMonitorId  int  `json:"log_monitor_id"`
	EndpointId  int  `json:"endpoint_id"`
	StrategyId  int  `json:"strategy_id"`
	Expr  string  `json:"expr" binding:"required"`
	Cond  string  `json:"cond" binding:"required"`
	Last  string  `json:"last"`
	Priority  string  `json:"priority" binding:"required"`
	Path  string  `json:"path"`
	Keyword  string  `json:"keyword"`
}

type GrpStrategyExportObj struct {
	GrpName  string  `json:"grp_name"`
	Description  string  `json:"description"`
	Strategy  []StrategyTable  `json:"strategy"`
}

type GrpStrategyQuery struct {
	Name  string
	Description  string
	Metric  string
	Expr  string
	Cond  string
	Last  string
	Priority  string
	Content  string
	ConfigType  string
}

type OpenAlarmObj struct {
	Id  int  `json:"id"`
	AlertInfo  string  `json:"alert_info"`
	AlertIp  string  `json:"alert_ip"`
	AlertLevel  string  `json:"alert_level"`
	AlertObj  string  `json:"alert_obj"`
	AlertTitle  string  `json:"alert_title"`
	AlertReciver  string  `json:"alert_reciver"`
	RemarkInfo  string  `json:"remark_info"`
	SubSystemId  string  `json:"sub_system_id"`
	UpdateAt  time.Time  `json:"update_at"`
}

type OpenAlarmRequest struct {
	AlertList  []OpenAlarmObj  `json:"alertList"`
}

type OpenAlarmResponse struct {
	ResultCode  int  `json:"resultCode"`
	ResultMsg  string  `json:"resultMsg"`
}

type UpdateActionDto struct {
	TplId  int  `json:"tpl_id" binding:"required"`
	Accept  []OptionModel  `json:"accept"`
}

type SyncConsulDto struct {
	IsRegister  bool  `json:"is_register"`
	Guid  string  `json:"guid"`
	Ip  string  `json:"ip"`
	Port  string  `json:"port"`
	Tags  []string  `json:"tags"`
	Interval  int  `json:"interval"`
}

type EndpointHttpTable struct {
	Id  int  `json:"id"`
	EndpointGuid  string  `json:"endpoint_guid"`
	Method  string  `json:"method"`
	Url  string  `json:"url"`
}

type LogMonitorTags struct {
	Endpoint  string  `json:"endpoint"`
	FilePath  string  `json:"file_path"`
	Keyword   string  `json:"keyword"`
	Tags      string  `json:"tags"`
}