package models

const (
	ServerToken             = `default-token-used-in-server-side`
	DatetimeFormat          = `2006-01-02 15:04:05`
	DatetimeDigitFormat     = `20060102150405`
	DateFormatWithZone      = `2006-01-02 15:04:05 MST`
	Version                 = `1.0.0`
	SystemRole              = `SUB_SYSTEM`
	PlatformUser            = `SYS_PLATFORM`
	UrlPrefix               = "/monitor"
	RsaPemPath              = "/data/certs/rsa_key"
	LogMetricName           = "node_log_metric_monitor_value"
	DBMonitorMetricName     = "db_monitor_value"
	SPAlertMailKey          = "alert_mail"
	SPMetricTemplate        = "metric_template"
	SPServiceMetricTemplate = "service_metric_template"
	MetricWorkspaceService  = "all_object"
	MetricWorkspaceAll      = "any_object"
	DefaultActiveWindow     = "00:00-23:59"

	AlarmNotifyAutoMode   = "auto"
	AlarmNotifyManualMode = "manual"

	AuthTokenHeader = "Authorization"
)

var (
	LogIgnorePath         = []string{"/monitor/webhook", "export/ping/source"}
	LogParamIgnorePath    = []string{"/dashboard/newchart", "/dashboard/pie/chart", "/problem/query", "/problem/history"}
	DashboardIgnoreTagKey = []string{"job", "__name__"}
)

type Permission string

const (
	PermissionMgmt Permission = "mgmt" //管理权限
	PermissionUse  Permission = "use"  //使用权限
)

type MetricType string

const (
	MetricTypeCommon   MetricType = "common"   //通用类型
	MetricTypeBusiness MetricType = "business" //业务配置
	MetricTypeCustom   MetricType = "custom"   // 自定义
)

type ImportRule string

const (
	ImportRuleCover  ImportRule = "cover"
	ImportRuleInsert ImportRule = "insert"
)

const (
	PreviewPointCount int = 6 // 预览默认给6个点
)
