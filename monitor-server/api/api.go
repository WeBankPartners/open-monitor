package api

import (
	"bytes"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/agent"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/config_new"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/dashboard"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/dashboard_new"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/user"
	alarmv2 "github.com/WeBankPartners/open-monitor/monitor-server/api/v2/alarm"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v2/monitor"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v2/service"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type handlerFuncObj struct {
	HandlerFunc  func(c *gin.Context)
	Method       string
	Url          string
	LogOperation bool
	PreHandle    func(c *gin.Context)
}

var (
	httpHandlerFuncList   []*handlerFuncObj
	httpHandlerFuncListV2 []*handlerFuncObj
)

func init() {
	// Dashboard 视图
	httpHandlerFuncList = append(httpHandlerFuncList,
		// 对象视图
		&handlerFuncObj{Url: "/dashboard/main", Method: http.MethodGet, HandlerFunc: dashboard.MainDashboard},
		&handlerFuncObj{Url: "/dashboard/panels", Method: http.MethodGet, HandlerFunc: dashboard.GetPanels},
		&handlerFuncObj{Url: "/dashboard/tags", Method: http.MethodGet, HandlerFunc: dashboard.GetTags},
		&handlerFuncObj{Url: "/dashboard/search", Method: http.MethodGet, HandlerFunc: dashboard.MainSearch},
		&handlerFuncObj{Url: "/dashboard/chart", Method: http.MethodPost, HandlerFunc: dashboard_new.GetChartData},
		&handlerFuncObj{Url: "/dashboard/comparison_chart", Method: http.MethodPost, HandlerFunc: dashboard_new.GetComparisonChartData},
		&handlerFuncObj{Url: "/dashboard/config/chart/title", Method: http.MethodPost, HandlerFunc: dashboard.UpdateChartsTitle},
		// 自定义视图
		&handlerFuncObj{Url: "/dashboard/pie/chart", Method: http.MethodPost, HandlerFunc: dashboard.GetPieChart},
		&handlerFuncObj{Url: "/dashboard/custom/list", Method: http.MethodGet, HandlerFunc: dashboard.ListCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom/get", Method: http.MethodGet, HandlerFunc: dashboard.GetCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom/save", Method: http.MethodPost, HandlerFunc: dashboard.SaveCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom/delete", Method: http.MethodGet, HandlerFunc: dashboard.DeleteCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/server/chart", Method: http.MethodGet, HandlerFunc: dashboard.GetChartsByEndpoint},
		&handlerFuncObj{Url: "/dashboard/custom/main/get", Method: http.MethodGet, HandlerFunc: dashboard.GetMainPage},
		&handlerFuncObj{Url: "/dashboard/custom/main/list", Method: http.MethodGet, HandlerFunc: dashboard.ListMainPageRole},
		&handlerFuncObj{Url: "/dashboard/custom/main/set", Method: http.MethodPost, HandlerFunc: dashboard.UpdateMainPage},
		&handlerFuncObj{Url: "/dashboard/custom/endpoint/get", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointsByIp},
		&handlerFuncObj{Url: "/dashboard/custom/role/get", Method: http.MethodGet, HandlerFunc: dashboard.GetCustomDashboardRole},
		&handlerFuncObj{Url: "/dashboard/custom/role/save", Method: http.MethodPost, HandlerFunc: dashboard.SaveCustomDashboardRole},
		&handlerFuncObj{Url: "/dashboard/custom/alarm/list/:customDashboardId", Method: http.MethodPost, HandlerFunc: alarm.GetCustomDashboardAlarm},
		&handlerFuncObj{Url: "/dashboard/config/metric/list", Method: http.MethodGet, HandlerFunc: dashboard.GetPromMetric},
		// 层级对象
		&handlerFuncObj{Url: "/dashboard/system/add", Method: http.MethodPost, HandlerFunc: agent.ExportPanelAdd},
		&handlerFuncObj{Url: "/dashboard/system/delete", Method: http.MethodPost, HandlerFunc: agent.ExportPanelDelete},
		&handlerFuncObj{Url: "/dashboard/recursive/get", Method: http.MethodGet, HandlerFunc: agent.GetPanelRecursive},
		&handlerFuncObj{Url: "/dashboard/recursive/endpoint_type/list", Method: http.MethodGet, HandlerFunc: agent.GetPanelRecursiveEndpointType},
		// 指标配置
		&handlerFuncObj{Url: "/dashboard/endpoint/type", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointTypeList},
		&handlerFuncObj{Url: "/dashboard/endpoint", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointList},
		&handlerFuncObj{Url: "/dashboard/endpoint/metric/list", Method: http.MethodPost, HandlerFunc: dashboard.GetEndpointMetric},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodGet, HandlerFunc: dashboard_new.MetricList},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodPost, HandlerFunc: dashboard_new.MetricCreate},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodPut, HandlerFunc: dashboard_new.MetricUpdate},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodDelete, HandlerFunc: dashboard_new.MetricDelete},
		&handlerFuncObj{Url: "/dashboard/new/panel", Method: http.MethodGet, HandlerFunc: dashboard_new.PanelList},
		&handlerFuncObj{Url: "/dashboard/new/panel/:endpointType", Method: http.MethodPost, HandlerFunc: dashboard_new.PanelCreate},
		&handlerFuncObj{Url: "/dashboard/new/panel", Method: http.MethodPut, HandlerFunc: dashboard_new.PanelUpdate},
		&handlerFuncObj{Url: "/dashboard/new/panel", Method: http.MethodDelete, HandlerFunc: dashboard_new.PanelDelete},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodGet, HandlerFunc: dashboard_new.ChartList},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodPost, HandlerFunc: dashboard_new.ChartCreate},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodPut, HandlerFunc: dashboard_new.ChartUpdate},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodDelete, HandlerFunc: dashboard_new.ChartDelete},
		&handlerFuncObj{Url: "/dashboard/new/comparison_metric", Method: http.MethodPost, HandlerFunc: monitor.AddOrUpdateComparisonMetric},
		&handlerFuncObj{Url: "/dashboard/new/comparison_metric/:id", Method: http.MethodDelete, HandlerFunc: monitor.DeleteComparisonMetric},
	)
	// Agent 对象管理
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/agent/register", Method: http.MethodPost, HandlerFunc: agent.RegisterAgentNew},
		&handlerFuncObj{Url: "/agent/deregister", Method: http.MethodPost, HandlerFunc: agent.DeregisterAgent},
		&handlerFuncObj{Url: "/agent/export/custom/endpoint/add", Method: http.MethodPost, HandlerFunc: agent.CustomRegister},
		&handlerFuncObj{Url: "/agent/custom/metric/add", Method: http.MethodPost, HandlerFunc: agent.CustomMetricPush},
		&handlerFuncObj{Url: "/agent/endpoint/telnet/get", Method: http.MethodGet, HandlerFunc: agent.GetEndpointTelnet},
		&handlerFuncObj{Url: "/agent/endpoint/telnet/update", Method: http.MethodPost, HandlerFunc: agent.UpdateEndpointTelnet},
		&handlerFuncObj{Url: "/agent/kubernetes/cluster/:operation", Method: http.MethodPost, HandlerFunc: agent.UpdateKubernetesCluster},
	)
	// Config 配置
	httpHandlerFuncList = append(httpHandlerFuncList,
		// 对象配置
		&handlerFuncObj{Url: "/alarm/endpoint/list", Method: http.MethodPost, HandlerFunc: alarm.ListGrpEndpoint},
		&handlerFuncObj{Url: "/alarm/endpoint/options", Method: http.MethodGet, HandlerFunc: alarm.ListGrpEndpointOptions},
		&handlerFuncObj{Url: "/alarm/endpoint/update", Method: http.MethodPost, HandlerFunc: alarm.EditGrpEndpoint},
		&handlerFuncObj{Url: "/alarm/process/list", Method: http.MethodGet, HandlerFunc: alarm.GetEndpointProcessConfig},
		&handlerFuncObj{Url: "/alarm/process/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateEndpointProcessConfig},
		&handlerFuncObj{Url: "/alarm/window/get", Method: http.MethodGet, HandlerFunc: alarm.GetAlertWindowList},
		&handlerFuncObj{Url: "/alarm/window/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateAlertWindow},
		// db查询监控配置
		&handlerFuncObj{Url: "/alarm/db/monitor/list", Method: http.MethodGet, HandlerFunc: alarm.GetDbMonitorList},
		&handlerFuncObj{Url: "/alarm/db/monitor/add", Method: http.MethodPost, HandlerFunc: alarm.AddDbMonitor},
		&handlerFuncObj{Url: "/alarm/db/monitor/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateDbMonitor},
		&handlerFuncObj{Url: "/alarm/db/monitor/check", Method: http.MethodPost, HandlerFunc: alarm.CheckDbMonitor},
		&handlerFuncObj{Url: "/alarm/db/monitor/delete", Method: http.MethodPost, HandlerFunc: alarm.DeleteDbMonitor},
		&handlerFuncObj{Url: "/alarm/db/monitor/sys/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateDbMonitorSysName},
		// 组配置
		&handlerFuncObj{Url: "/alarm/grp/list", Method: http.MethodGet, HandlerFunc: alarm.ListGrp},
		&handlerFuncObj{Url: "/alarm/grp/add", Method: http.MethodPost, HandlerFunc: alarm.AddGrp},
		&handlerFuncObj{Url: "/alarm/grp/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateGrp},
		&handlerFuncObj{Url: "/alarm/grp/delete", Method: http.MethodGet, HandlerFunc: alarm.DeleteGrp},
		&handlerFuncObj{Url: "/alarm/grp/role/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateGrpRole},
		&handlerFuncObj{Url: "/alarm/grp/role/get", Method: http.MethodGet, HandlerFunc: alarm.GetGrpRole},
		&handlerFuncObj{Url: "/alarm/endpoint/grp/update", Method: http.MethodPost, HandlerFunc: alarm.EditEndpointGrp},
		&handlerFuncObj{Url: "/alarm/grp/export", Method: http.MethodGet, HandlerFunc: alarm.ExportGrpStrategy},
		&handlerFuncObj{Url: "/alarm/grp/import", Method: http.MethodPost, HandlerFunc: alarm.ImportGrpStrategy},
		// 阈值配置
		&handlerFuncObj{Url: "/alarm/strategy/search", Method: http.MethodGet, HandlerFunc: alarm.SearchObjOption},
		&handlerFuncObj{Url: "/alarm/strategy/list", Method: http.MethodGet, HandlerFunc: alarm.ListTpl},
		&handlerFuncObj{Url: "/alarm/strategy/add", Method: http.MethodPost, HandlerFunc: alarm.AddStrategy},
		&handlerFuncObj{Url: "/alarm/strategy/update", Method: http.MethodPost, HandlerFunc: alarm.EditStrategy},
		&handlerFuncObj{Url: "/alarm/strategy/delete", Method: http.MethodGet, HandlerFunc: alarm.DeleteStrategy},
		&handlerFuncObj{Url: "/alarm/action/search", Method: http.MethodGet, HandlerFunc: alarm.SearchUserRole},
		&handlerFuncObj{Url: "/alarm/action/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateTplAction},
		// 告警列表
		&handlerFuncObj{Url: "/alarm/history", Method: http.MethodGet, HandlerFunc: alarm.GetHistoryAlarm},
		&handlerFuncObj{Url: "/alarm/problem/options", Method: http.MethodGet, HandlerFunc: alarm.GetProblemAlarmOptions},
		&handlerFuncObj{Url: "/alarm/problem/list", Method: http.MethodGet, HandlerFunc: alarm.GetProblemAlarm},
		&handlerFuncObj{Url: "/alarm/problem/query", Method: http.MethodPost, HandlerFunc: alarm.QueryProblemAlarm},
		&handlerFuncObj{Url: "/alarm/problem/page", Method: http.MethodPost, HandlerFunc: alarm.QueryProblemAlarmByPage},
		&handlerFuncObj{Url: "/alarm/problem/close", Method: http.MethodPost, HandlerFunc: alarm.CloseAlarm},
		&handlerFuncObj{Url: "/alarm/problem/history", Method: http.MethodPost, HandlerFunc: alarm.QueryHistoryAlarm},
		&handlerFuncObj{Url: "/alarm/problem/message", Method: http.MethodPost, HandlerFunc: alarm.UpdateAlarmCustomMessage},
		&handlerFuncObj{Url: "/alarm/problem/notify", Method: http.MethodPost, HandlerFunc: alarm.NotifyAlarm},
		// 关键字监控配置
		&handlerFuncObj{Url: "/alarm/log/monitor/list", Method: http.MethodGet, HandlerFunc: alarm.ListLogTpl},
		&handlerFuncObj{Url: "/alarm/log/monitor/add", Method: http.MethodPost, HandlerFunc: alarm.AddLogStrategy},
		&handlerFuncObj{Url: "/alarm/log/monitor/update", Method: http.MethodPost, HandlerFunc: alarm.EditLogStrategy},
		&handlerFuncObj{Url: "/alarm/log/monitor/update_path", Method: http.MethodPost, HandlerFunc: alarm.EditLogPath},
		&handlerFuncObj{Url: "/alarm/log/monitor/delete", Method: http.MethodGet, HandlerFunc: alarm.DeleteLogStrategy},
		&handlerFuncObj{Url: "/alarm/log/monitor/delete_path", Method: http.MethodGet, HandlerFunc: alarm.DeleteLogPath},
		// 业务日志监控配置
		&handlerFuncObj{Url: "/alarm/business/list", Method: http.MethodGet, HandlerFunc: alarm.GetEndpointBusinessConfig},
		&handlerFuncObj{Url: "/alarm/business/add", Method: http.MethodPost, HandlerFunc: alarm.AddEndpointBusinessConfig},
		&handlerFuncObj{Url: "/alarm/business/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateEndpointBusinessConfig},
		// 层级对象配置
		&handlerFuncObj{Url: "/alarm/org/panel/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrganizaionList},
		&handlerFuncObj{Url: "/alarm/org/panel/:name", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanel},
		&handlerFuncObj{Url: "/alarm/org/role/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgPanelRole},
		&handlerFuncObj{Url: "/alarm/org/role/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanelRole},
		&handlerFuncObj{Url: "/alarm/org/endpoint/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgPanelEndpoint},
		&handlerFuncObj{Url: "/alarm/org/endpoint/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanelEndpoint},
		&handlerFuncObj{Url: "/alarm/org/plugin", Method: http.MethodGet, HandlerFunc: alarm.IsPluginMode},
		&handlerFuncObj{Url: "/alarm/org/callback/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgPanelEventList},
		&handlerFuncObj{Url: "/alarm/org/callback/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanelCallback},
		&handlerFuncObj{Url: "/alarm/org/connect/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgConnect},
		&handlerFuncObj{Url: "/alarm/org/connect/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgConnect},
		&handlerFuncObj{Url: "/alarm/org/search", Method: http.MethodGet, HandlerFunc: alarm.SearchSysPanelData},
		// 采集器配置
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodGet, HandlerFunc: config_new.SnmpExporterList},
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodPost, HandlerFunc: config_new.SnmpExporterCreate},
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodPut, HandlerFunc: config_new.SnmpExporterUpdate},
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodDelete, HandlerFunc: config_new.SnmpExporterDelete},
	)
	// User
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/user/message/get", Method: http.MethodGet, HandlerFunc: user.GetUserMsg},
		&handlerFuncObj{Url: "/user/message/update", Method: http.MethodPost, HandlerFunc: user.UpdateUserMsg},
		&handlerFuncObj{Url: "/user/list", Method: http.MethodGet, HandlerFunc: user.ListUser},
		&handlerFuncObj{Url: "/user/role/update", Method: http.MethodPost, HandlerFunc: user.UpdateRole},
		&handlerFuncObj{Url: "/user/role/list", Method: http.MethodGet, HandlerFunc: user.ListRole},
		&handlerFuncObj{Url: "/user/manage_role/list", Method: http.MethodGet, HandlerFunc: user.ListManageRole},
		&handlerFuncObj{Url: "/user/role/user/update", Method: http.MethodPost, HandlerFunc: user.UpdateRoleUser},
	)
	// Export plugin interface
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/agent/export/register/:name", Method: http.MethodPost, HandlerFunc: agent.ExportAgentNew},
		&handlerFuncObj{Url: "/agent/export/deregister/:name", Method: http.MethodPost, HandlerFunc: agent.ExportAgentNew},
		&handlerFuncObj{Url: "/agent/export/start/:name", Method: http.MethodPost, HandlerFunc: agent.AlarmControl},
		&handlerFuncObj{Url: "/agent/export/stop/:name", Method: http.MethodPost, HandlerFunc: agent.AlarmControl},
		&handlerFuncObj{Url: "/agent/export/ping/source", Method: http.MethodGet, HandlerFunc: agent.ExportPingSource},
		&handlerFuncObj{Url: "/agent/export/process/:operation", Method: http.MethodPost, HandlerFunc: agent.AutoUpdateProcessMonitor},
		&handlerFuncObj{Url: "/agent/export/log_monitor/:operation", Method: http.MethodPost, HandlerFunc: agent.AutoUpdateLogMonitor},
		&handlerFuncObj{Url: "/agent/export/kubernetes/cluster/:action", Method: http.MethodPost, HandlerFunc: agent.PluginKubernetesCluster},
		&handlerFuncObj{Url: "/agent/export/kubernetes/pod/:action", Method: http.MethodPost, HandlerFunc: agent.PluginKubernetesPod},
		&handlerFuncObj{Url: "/agent/export/snmp/exporter/:action", Method: http.MethodPost, HandlerFunc: config_new.PluginSnmpExporterHandle},
	)
	// V2
	httpHandlerFuncListV2 = append(httpHandlerFuncListV2,
		// service
		&handlerFuncObj{Url: "/service_endpoint/search/:searchType", Method: http.MethodGet, HandlerFunc: service.GetServiceGroupEndpointList},
		&handlerFuncObj{Url: "/service/log_metric/list/:queryType/:guid", Method: http.MethodGet, HandlerFunc: service.ListLogMetricMonitor},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor/:logMonitorGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricMonitor},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricMonitor},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricMonitor},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor/:logMonitorGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricMonitor},
		&handlerFuncObj{Url: "/service/service_group/endpoint_rel", Method: http.MethodGet, HandlerFunc: service.GetServiceGroupEndpointRel},
		&handlerFuncObj{Url: "/service/log_metric/export", Method: http.MethodGet, HandlerFunc: service.ExportLogMetric},
		&handlerFuncObj{Url: "/service/log_metric/import", Method: http.MethodPost, HandlerFunc: service.ImportLogMetric},

		&handlerFuncObj{Url: "/service/log_metric/log_metric_json/:logMonitorJsonGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricJson},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_json", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricJson},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_json", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricJson},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_json/:logMonitorJsonGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricJson},

		&handlerFuncObj{Url: "/service/log_metric/log_metric_config/:logMonitorConfigGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricConfig},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_config", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricConfig},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_config", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricConfig},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_config/:logMonitorConfigGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricConfig},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_import/excel/:logMonitorGuid", Method: http.MethodPost, HandlerFunc: service.ImportLogMetricExcel},
		&handlerFuncObj{Url: "/service/service_group/:serviceGroup/endpoint/:monitorType", Method: http.MethodGet, HandlerFunc: service.ListServiceGroupEndpoint},

		&handlerFuncObj{Url: "/service/db_metric/list/:queryType/:guid", Method: http.MethodGet, HandlerFunc: service.ListDbMetricMonitor},
		&handlerFuncObj{Url: "/service/db_metric/:dbMonitorGuid", Method: http.MethodGet, HandlerFunc: service.GetDbMetricMonitor},
		&handlerFuncObj{Url: "/service/db_metric", Method: http.MethodPost, HandlerFunc: service.CreateDbMetricMonitor},
		&handlerFuncObj{Url: "/service/db_metric", Method: http.MethodPut, HandlerFunc: service.UpdateDbMetricMonitor},
		&handlerFuncObj{Url: "/service/db_metric/:dbMonitorGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteDbMetricMonitor},
		&handlerFuncObj{Url: "/regexp/test/match", Method: http.MethodPost, HandlerFunc: service.CheckRegExpMatch},

		&handlerFuncObj{Url: "/service/log_keyword/list/:queryType/:guid", Method: http.MethodGet, HandlerFunc: service.ListLogKeywordMonitor},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_monitor", Method: http.MethodPost, HandlerFunc: service.CreateLogKeywordMonitor},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_monitor", Method: http.MethodPut, HandlerFunc: service.UpdateLogKeywordMonitor},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_monitor/:logKeywordMonitorGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogKeywordMonitor},
		&handlerFuncObj{Url: "/service/log_keyword/export", Method: http.MethodGet, HandlerFunc: service.ExportLogKeyword},
		&handlerFuncObj{Url: "/service/log_keyword/import", Method: http.MethodPost, HandlerFunc: service.ImportLogKeyword},

		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_config", Method: http.MethodPost, HandlerFunc: service.CreateLogKeyword},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_config", Method: http.MethodPut, HandlerFunc: service.UpdateLogKeyword},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_config/:logKeywordGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogKeyword},
		// service plugin
		&handlerFuncObj{Url: "/service/plugin/update/path", Method: http.MethodPost, HandlerFunc: service.PluginUpdateServicePath},
		// alarm
		&handlerFuncObj{Url: "/alarm/endpoint_group/query", Method: http.MethodGet, HandlerFunc: alarmv2.ListEndpointGroup},
		&handlerFuncObj{Url: "/alarm/endpoint_group/options", Method: http.MethodGet, HandlerFunc: alarmv2.EndpointGroupOptions},
		&handlerFuncObj{Url: "/alarm/endpoint_group", Method: http.MethodPost, HandlerFunc: alarmv2.CreateEndpointGroup},
		&handlerFuncObj{Url: "/alarm/endpoint_group", Method: http.MethodPut, HandlerFunc: alarmv2.UpdateEndpointGroup},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid", Method: http.MethodDelete, HandlerFunc: alarmv2.DeleteEndpointGroup},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/endpoint/list", Method: http.MethodGet, HandlerFunc: alarmv2.GetGroupEndpointRel},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/endpoint/update", Method: http.MethodPost, HandlerFunc: alarmv2.UpdateGroupEndpoint},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/notify/list", Method: http.MethodGet, HandlerFunc: alarmv2.GetGroupEndpointNotify},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/notify/update", Method: http.MethodPost, HandlerFunc: alarmv2.UpdateGroupEndpointNotify},

		&handlerFuncObj{Url: "/alarm/strategy/search", Method: http.MethodGet, HandlerFunc: alarmv2.ListStrategyQueryOptions},
		&handlerFuncObj{Url: "/alarm/strategy/list/:queryType/:guid", Method: http.MethodGet, HandlerFunc: alarmv2.QueryAlarmStrategy},
		&handlerFuncObj{Url: "/alarm/strategy", Method: http.MethodPost, HandlerFunc: alarmv2.CreateAlarmStrategy},
		&handlerFuncObj{Url: "/alarm/strategy", Method: http.MethodPut, HandlerFunc: alarmv2.UpdateAlarmStrategy},
		&handlerFuncObj{Url: "/alarm/strategy/:strategyGuid", Method: http.MethodDelete, HandlerFunc: alarmv2.DeleteAlarmStrategy},
		&handlerFuncObj{Url: "/alarm/event/callback/list", Method: http.MethodGet, HandlerFunc: alarmv2.ListCallbackEvent},
		&handlerFuncObj{Url: "/alarm/strategy/export/:queryType/:guid", Method: http.MethodGet, HandlerFunc: alarmv2.ExportAlarmStrategy},
		&handlerFuncObj{Url: "/alarm/strategy/import/:queryType/:guid", Method: http.MethodPost, HandlerFunc: alarmv2.ImportAlarmStrategy},
		// monitor
		&handlerFuncObj{Url: "/monitor/endpoint/query", Method: http.MethodGet, HandlerFunc: monitor.ListEndpoint},
		&handlerFuncObj{Url: "/monitor/metric/list", Method: http.MethodGet, HandlerFunc: monitor.ListMetric},
		&handlerFuncObj{Url: "/monitor/metric/list/count", Method: http.MethodGet, HandlerFunc: monitor.ListMetricCount},
		&handlerFuncObj{Url: "/monitor/metric_comparison/list", Method: http.MethodGet, HandlerFunc: monitor.ListMetricComparison},
		&handlerFuncObj{Url: "/sys/parameter/metric_template", Method: http.MethodGet, HandlerFunc: monitor.GetSysMetricTemplate},
		&handlerFuncObj{Url: "/monitor/endpoint/get/:guid", Method: http.MethodGet, HandlerFunc: monitor.GetEndpoint},
		&handlerFuncObj{Url: "/monitor/endpoint/update", Method: http.MethodPut, HandlerFunc: monitor.UpdateEndpoint},
		&handlerFuncObj{Url: "/monitor/metric/export", Method: http.MethodGet, HandlerFunc: monitor.ExportMetric},
		&handlerFuncObj{Url: "/monitor/metric/import", Method: http.MethodPost, HandlerFunc: monitor.ImportMetric},
		// log monitor template
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/list", Method: http.MethodPost, HandlerFunc: service.ListLogMonitorTemplate},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/:logMonitorTemplateGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMonitorTemplate},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template", Method: http.MethodPost, HandlerFunc: service.CreateLogMonitorTemplate},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template", Method: http.MethodPut, HandlerFunc: service.UpdateLogMonitorTemplate},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/:logMonitorTemplateGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMonitorTemplate},
		&handlerFuncObj{Url: "/service/log_metric/affect_service_group/:logMonitorTemplateGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMonitorTemplateServiceGroup},
		&handlerFuncObj{Url: "/service/log_metric/regexp/match", Method: http.MethodPost, HandlerFunc: service.CheckLogMonitorRegExpMatch},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/export", Method: http.MethodPost, HandlerFunc: service.LogMonitorTemplateExport},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/import", Method: http.MethodPost, HandlerFunc: service.LogMonitorTemplateImport},

		&handlerFuncObj{Url: "/service/log_metric/log_metric_group/:logMetricGroupGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricGroup},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricGroup},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricGroup},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group/:logMetricGroupGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricGroup},
		&handlerFuncObj{Url: "/service/log_metric/custom/log_metric_group/:logMetricGroupGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricCustomGroup},
		&handlerFuncObj{Url: "/service/log_metric/custom/log_metric_group", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricCustomGroup},
		&handlerFuncObj{Url: "/service/log_metric/custom/log_metric_group", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricCustomGroup},
		// 标签
		&handlerFuncObj{Url: "/metric/tag/value-list", Method: http.MethodPost, HandlerFunc: monitor.QueryMetricTagValue},

		//自定义视图
		&handlerFuncObj{Url: "/dashboard/all", Method: http.MethodGet, HandlerFunc: monitor.GetAllCustomDashboardList},
		&handlerFuncObj{Url: "/dashboard/custom/list", Method: http.MethodPost, HandlerFunc: monitor.QueryCustomDashboardList},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodGet, HandlerFunc: monitor.GetCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodPost, HandlerFunc: monitor.AddCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodPut, HandlerFunc: monitor.UpdateCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom/permission", Method: http.MethodPost, HandlerFunc: monitor.UpdateCustomDashboardPermission},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodDelete, HandlerFunc: monitor.DeleteCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom/export", Method: http.MethodPost, HandlerFunc: monitor.ExportCustomDashboard},
		&handlerFuncObj{Url: "/dashboard/custom/import", Method: http.MethodPost, HandlerFunc: monitor.ImportCustomDashboard},
		&handlerFuncObj{Url: "/chart/shared/list", Method: http.MethodGet, HandlerFunc: monitor.GetSharedChartList},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodPost, HandlerFunc: monitor.AddCustomChart},
		&handlerFuncObj{Url: "/chart/custom/copy", Method: http.MethodPost, HandlerFunc: monitor.CopyCustomChart},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodPut, HandlerFunc: monitor.UpdateCustomChart},
		&handlerFuncObj{Url: "/chart/custom/name", Method: http.MethodPut, HandlerFunc: monitor.UpdateCustomChartName},
		&handlerFuncObj{Url: "/chart/custom/name/exist", Method: http.MethodGet, HandlerFunc: monitor.QueryCustomChartNameExist},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodGet, HandlerFunc: monitor.GetCustomChart},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodDelete, HandlerFunc: monitor.DeleteCustomChart},
		&handlerFuncObj{Url: "/chart/custom/permission", Method: http.MethodPost, HandlerFunc: monitor.SharedCustomChart},
		&handlerFuncObj{Url: "/chart/custom/permission", Method: http.MethodGet, HandlerFunc: monitor.GetSharedChartPermission},
		&handlerFuncObj{Url: "/chart/manage/list", Method: http.MethodPost, HandlerFunc: monitor.QueryCustomChart},
		&handlerFuncObj{Url: "/dashboard/data/sync", Method: http.MethodPost, HandlerFunc: monitor.SyncData},
		&handlerFuncObj{Url: "/chart/custom/series/config", Method: http.MethodPost, HandlerFunc: monitor.GetChartSeriesColor},

		// 远程读写
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodGet, HandlerFunc: config_new.RemoteWriteConfigList},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodPost, HandlerFunc: config_new.RemoteWriteConfigCreate},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodPut, HandlerFunc: config_new.RemoteWriteConfigUpdate},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodDelete, HandlerFunc: config_new.RemoteWriteConfigDelete},

		// 类型配置
		&handlerFuncObj{Url: "/config/type/query", Method: http.MethodGet, HandlerFunc: monitor.QueryTypeConfigList},
		&handlerFuncObj{Url: "/config/type", Method: http.MethodPost, HandlerFunc: monitor.AddTypeConfig},
		&handlerFuncObj{Url: "/config/type", Method: http.MethodDelete, HandlerFunc: monitor.DeleteTypeConfig},

		// 获取seed
		&handlerFuncObj{Url: "/seed", Method: http.MethodGet, HandlerFunc: monitor.GetEncryptSeed},
	)
}

func InitHttpServer() {
	urlPrefix := models.UrlPrefix
	r := gin.New()
	if !models.PluginRunningMode {
		// reflect ui resource
		r.LoadHTMLGlob("public/*.html")
		r.Static(fmt.Sprintf("%s/js", urlPrefix), fmt.Sprintf("public%s/js", urlPrefix))
		r.Static(fmt.Sprintf("%s/css", urlPrefix), fmt.Sprintf("public%s/css", urlPrefix))
		r.Static(fmt.Sprintf("%s/img", urlPrefix), fmt.Sprintf("public%s/img", urlPrefix))
		r.Static(fmt.Sprintf("%s/fonts", urlPrefix), fmt.Sprintf("public%s/fonts", urlPrefix))
		r.StaticFile("/favicon.ico", "public/favicon.ico")
		r.GET(fmt.Sprintf("%s/", urlPrefix), func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{})
		})
		// allow cross request
		r.Use(func(c *gin.Context) {
			if c.Request.Method == "OPTIONS" {
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, authorization, Token, X-Auth-Token")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
				if c.GetHeader("Origin") != "" {
					c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
				} else {
					c.Header("Access-Control-Allow-Origin", "*")
				}
				c.AbortWithStatus(http.StatusNoContent)
			}
			if c.GetHeader("Origin") != "" {
				c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
			}
		})
	}
	// access log
	r.Use(httpLogHandle())
	// const handler func
	r.POST(fmt.Sprintf("%s/login", urlPrefix), user.Login)
	r.POST(fmt.Sprintf("%s/register", urlPrefix), user.Register)
	r.GET(fmt.Sprintf("%s/logout", urlPrefix), user.Logout)
	r.GET(fmt.Sprintf("%s/check", urlPrefix), user.HealthCheck)
	r.GET(fmt.Sprintf("%s/demo", urlPrefix), dashboard.DisplayWatermark)
	r.POST(fmt.Sprintf("%s/webhook", urlPrefix), alarm.AcceptAlert)
	r.POST(fmt.Sprintf("%s/openapi/alarm/send", urlPrefix), alarm.OpenAlarmApi)
	entityApi := r.Group(fmt.Sprintf("%s/entities", urlPrefix), user.AuthRequired())
	{
		entityApi.POST("/alarm/query", alarm.QueryEntityAlarm)
		entityApi.POST("/alarm/close", alarmv2.PluginCloseAlarm)
		entityApi.POST("/alarm_event/query", alarm.QueryEntityAlarmEvent)
		entityApi.POST("/alarm_event/update", alarm.UpdateEntityAlarm)
	}
	// register handler func with auth
	authRouter := r.Group(urlPrefix+"/api/v1", user.AuthRequired())
	for _, funcObj := range httpHandlerFuncList {
		handleFuncList := []gin.HandlerFunc{funcObj.HandlerFunc}
		if funcObj.PreHandle != nil {
			log.Logger.Info("Append pre handle", log.String("url", funcObj.Url))
			handleFuncList = append([]gin.HandlerFunc{funcObj.PreHandle}, funcObj.HandlerFunc)
		}
		switch funcObj.Method {
		case "GET":
			authRouter.GET(funcObj.Url, funcObj.HandlerFunc)
			break
		case "POST":
			authRouter.POST(funcObj.Url, handleFuncList...)
			break
		case "PUT":
			authRouter.PUT(funcObj.Url, funcObj.HandlerFunc)
			break
		case "DELETE":
			authRouter.DELETE(funcObj.Url, funcObj.HandlerFunc)
			break
		}
	}
	authRouterV2 := r.Group(urlPrefix+"/api/v2", user.AuthRequired())
	for _, funcObj := range httpHandlerFuncListV2 {
		handleFuncList := []gin.HandlerFunc{funcObj.HandlerFunc}
		if funcObj.PreHandle != nil {
			handleFuncList = append([]gin.HandlerFunc{funcObj.PreHandle}, funcObj.HandlerFunc)
		}
		switch funcObj.Method {
		case "GET":
			authRouterV2.GET(funcObj.Url, handleFuncList...)
			break
		case "POST":
			authRouterV2.POST(funcObj.Url, handleFuncList...)
			break
		case "PUT":
			authRouterV2.PUT(funcObj.Url, handleFuncList...)
			break
		case "DELETE":
			authRouterV2.DELETE(funcObj.Url, handleFuncList...)
			break
		}
	}
	r.Run(":" + models.Config().Http.Port)
}

func httpLogHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ignoreLog := false
		for _, v := range models.LogIgnorePath {
			if strings.Contains(c.Request.RequestURI, v) {
				ignoreLog = true
				break
			}
		}
		if ignoreLog {
			c.Next()
		} else {
			start := time.Now()
			var bodyBytes []byte
			if c.Request.Method == http.MethodPost {
				ignore := false
				for _, v := range models.LogParamIgnorePath {
					if strings.Contains(c.Request.RequestURI, v) {
						ignore = true
						break
					}
				}
				if !ignore {
					bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
					c.Request.Body.Close()
					c.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
				}
			}
			c.Next()
			log.AccessLogger.Info("request", log.String("url", c.Request.RequestURI), log.String("method", c.Request.Method), log.Int("code", c.Writer.Status()), log.String("operator", c.GetString("operatorName")), log.String("ip", getRemoteIp(c)), log.Float64("cost_second", time.Now().Sub(start).Seconds()), log.String("body", string(bodyBytes)))
		}
	}
}

func getRemoteIp(c *gin.Context) string {
	return c.RemoteIP()
}

func InitClusterApi() {
	if !models.Config().Peer.Enable {
		return
	}
	http.Handle("/sync/config", http.HandlerFunc(alarm.SyncConfigHandle))
	http.Handle("/sync/sd", http.HandlerFunc(alarm.AcceptPeerSdConfigHandle))
	http.ListenAndServe(fmt.Sprintf(":%s", models.Config().Peer.HttpPort), nil)
}

func InitDependenceParam() {
	agent.InitAgentManager()
}
