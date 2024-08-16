package models

type DbMetricMonitorTable struct {
	Guid         string `json:"guid" xorm:"guid"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
	MetricSql    string `json:"metric_sql" xorm:"metric_sql"`
	Metric       string `json:"metric" xorm:"metric"`
	DisplayName  string `json:"display_name" xorm:"display_name"`
	Step         int64  `json:"step" xorm:"step"`
	MonitorType  string `json:"monitor_type" xorm:"monitor_type"`
	UpdateTime   string `json:"update_time" xorm:"update_time"`
	UpdateUser   string `json:"update_user" xorm:"update_user"`
}

type DbMetricEndpointRelTable struct {
	Guid            string `json:"guid" xorm:"guid"`
	DbMetricMonitor string `json:"db_metric_monitor" xorm:"db_metric_monitor"`
	SourceEndpoint  string `json:"source_endpoint" xorm:"source_endpoint"`
	TargetEndpoint  string `json:"target_endpoint" xorm:"target_endpoint"`
}

type DbMetricQueryObj struct {
	ServiceGroupTable
	Config []*DbMetricMonitorObj `json:"config"`
}

type DbMetricMonitorObj struct {
	Guid             string                      `json:"guid"`
	ServiceGroup     string                      `json:"service_group"`
	ServiceGroupName string                      `json:"service_group_name"`
	MetricSql        string                      `json:"metric_sql"`
	Metric           string                      `json:"metric"`
	DisplayName      string                      `json:"display_name"`
	Step             int64                       `json:"step"`
	MonitorType      string                      `json:"monitor_type"`
	UpdateTime       string                      `json:"update_time"`
	UpdateUser       string                      `json:"update_user"`
	EndpointRel      []*DbMetricEndpointRelTable `json:"endpoint_rel"`
}

type DbMetricMonitorQueryObj struct {
	Guid           string `json:"guid" xorm:"guid"`
	ServiceGroup   string `json:"service_group" xorm:"service_group"`
	MetricSql      string `json:"metric_sql" xorm:"metric_sql"`
	Metric         string `json:"metric" xorm:"metric"`
	DisplayName    string `json:"display_name" xorm:"display_name"`
	Step           int64  `json:"step" xorm:"step"`
	MonitorType    string `json:"monitor_type" xorm:"monitor_type"`
	SourceEndpoint string `json:"source_endpoint" xorm:"source_endpoint"`
	TargetEndpoint string `json:"target_endpoint" xorm:"target_endpoint"`
}

type DbKeywordMonitorQueryObj struct {
	Guid           string `json:"guid" xorm:"guid"`
	ServiceGroup   string `json:"service_group" xorm:"service_group"`
	QuerySql       string `json:"query_sql" xorm:"query_sql"`
	Step           int64  `json:"step" xorm:"step"`
	MonitorType    string `json:"monitor_type" xorm:"monitor_type"`
	Content        string `json:"content" xorm:"content"`
	Priority       string `json:"priority" xorm:"priority"`
	Name           string `json:"name" xorm:"name"`
	ActiveWindow   string `json:"active_window" xorm:"active_window"`
	SourceEndpoint string `json:"source_endpoint" xorm:"source_endpoint"`
	TargetEndpoint string `json:"target_endpoint" xorm:"target_endpoint"`
	NotifyEnable   int8   `json:"notify_enable" xorm:"notify_enable"` // 是否通知
}

type MetricComparison struct {
	Guid           string `json:"guid" xorm:"guid"`
	ComparisonType string `json:"comparisonType" xorm:"comparison_type"`
	CalcType       string `json:"calcType" xorm:"calc_type"`
	CalcMethod     string `json:"calcMethod" xorm:"calc_method"`
	CalcPeriod     string `json:"calcPeriod" xorm:"calc_period"`
	MetricId       string `json:"metricId" xorm:"metric_id"`
	OriginMetricId string `json:"originMetricId" xorm:"origin_metric_id"`
	CreateUser     string `json:"createUser" xorm:"create_user"`
	CreateTime     string `json:"createTime" xorm:"create_time"`
}

type MetricComparisonDto struct {
	Metric         string `json:"metric" xorm:"metric"`                   // 指标名称
	MonitorType    string `json:"monitorType" xorm:"monitor_type"`        // 原始指标类型
	ComparisonType string `json:"comparisonType" xorm:"comparison_type"`  // 对比类型: day 日环比, week 周, 月周比 month
	OriginPromExpr string `json:"originPromExpr" xorm:"origin_prom_expr"` // 原始指标prom表达式
	PromExpr       string `json:"promExpr" xorm:"prom_expr"`              // 同环比指标prom表达式
	CalcType       string `json:"calcType" xorm:"calc_type"`              // 计算数值: diff 差值,diff_percent 差值百分比
	CalcMethod     string `json:"calcMethod" xorm:"calc_method"`          // 计算方法: avg平均,sum求和,max最大,min最小
	CalcPeriod     int    `json:"calcPeriod" xorm:"calc_period"`          // 计算周期
	MetricId       string `json:"metricId" xorm:"metric_id"`              // 指标Id
	CreateUser     string `json:"createUser" xorm:"create_user"`
	CreateTime     string `json:"createTime" xorm:"create_time"`
}

type MetricComparisonParam struct {
	Metric             string   `json:"metric"`             // 原始指标名称
	MonitorType        string   `json:"monitorType"`        // 原始指标类型
	ComparisonType     string   `json:"comparisonType"`     // 对比类型: day 日环比, week 周, 月周比 month
	OriginPromExpr     string   `json:"originPromExpr"`     // 原始指标prom表达式
	PromExpr           string   `json:"promExpr"`           // 同环比指标prom表达式
	CalcType           []string `json:"calcType"`           // 计算数值: diff 差值,diff_percent 差值百分比
	CalcMethod         string   `json:"calcMethod"`         // 计算方法: avg平均,sum求和,max最大,min最小
	CalcPeriod         int      `json:"calcPeriod"`         // 计算周期,单位s
	MetricId           string   `json:"metricId"`           // 指标Id
	MetricComparisonId string   `json:"metricComparisonId"` // 同环比指标Id
}
