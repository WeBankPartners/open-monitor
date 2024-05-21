package models

const (
	ServerToken             = `default-token-used-in-server-side`
	DatetimeFormat          = `2006-01-02 15:04:05`
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
