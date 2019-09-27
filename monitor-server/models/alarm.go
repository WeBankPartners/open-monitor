package models

import "time"

type GrpTable struct {
	Id  int  `json:"id"`
	Name  string  `json:"name"`
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
	StartValue  float64  `json:"start_value"`
	Start  time.Time  `json:"start"`
	EndValue  float64  `json:"end_value"`
	End  time.Time  `json:"end"`
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
	Accept  AcceptObj  `json:"accept"`
	Strategy  []*StrategyTable  `json:"strategy"`
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