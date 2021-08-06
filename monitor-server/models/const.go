package models

const (
	ServerToken        = `default-token-used-in-server-side`
	DatetimeFormat     = `2006-01-02 15:04:05`
	DateFormatWithZone = `2006-01-02 15:04:05 MST`
	Version            = `1.0.0`
	SystemRole         = `SUB_SYSTEM`
	PlatformUser       = `SYS_PLATFORM`
	UrlPrefix          = "/monitor"
	RsaPemPath         = "/data/certs/rsa_key"
)

var (
	LogIgnorePath         = []string{"/monitor/webhook", "export/ping/source"}
	LogParamIgnorePath    = []string{"/dashboard/newchart", "/dashboard/pie/chart", "/problem/query", "/problem/history"}
	DashboardIgnoreTagKey = []string{"job", "__name__"}
)
