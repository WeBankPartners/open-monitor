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
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
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
	ApiCode      string
}

var (
	httpHandlerFuncList   []*handlerFuncObj
	httpHandlerFuncListV2 []*handlerFuncObj
	apiCodeMap            = make(map[string]string)
)

func init() {
	// Dashboard 视图
	httpHandlerFuncList = append(httpHandlerFuncList,
		// 对象视图
		&handlerFuncObj{Url: "/dashboard/main", Method: http.MethodGet, HandlerFunc: dashboard.MainDashboard, ApiCode: "dashboard_main"},
		&handlerFuncObj{Url: "/dashboard/panels", Method: http.MethodGet, HandlerFunc: dashboard.GetPanels, ApiCode: "dashboard_panels"},
		&handlerFuncObj{Url: "/dashboard/tags", Method: http.MethodGet, HandlerFunc: dashboard.GetTags, ApiCode: "dashboard_tags"},
		&handlerFuncObj{Url: "/dashboard/search", Method: http.MethodGet, HandlerFunc: dashboard.MainSearch, ApiCode: "dashboard_search"},
		&handlerFuncObj{Url: "/dashboard/chart", Method: http.MethodPost, HandlerFunc: dashboard_new.GetChartData, ApiCode: "dashboard_chart"},
		&handlerFuncObj{Url: "/dashboard/comparison_chart", Method: http.MethodPost, HandlerFunc: dashboard_new.GetComparisonChartData, ApiCode: "dashboard_comparison"},
		&handlerFuncObj{Url: "/dashboard/config/chart/title", Method: http.MethodPost, HandlerFunc: dashboard.UpdateChartsTitle, ApiCode: "dashboard_config"},
		// 自定义视图
		&handlerFuncObj{Url: "/dashboard/pie/chart", Method: http.MethodPost, HandlerFunc: dashboard.GetPieChart, ApiCode: "dashboard_pie_chart"},
		&handlerFuncObj{Url: "/dashboard/custom/list", Method: http.MethodGet, HandlerFunc: dashboard.ListCustomDashboard, ApiCode: "dashboard_custom_list"},
		&handlerFuncObj{Url: "/dashboard/custom/get", Method: http.MethodGet, HandlerFunc: dashboard.GetCustomDashboard, ApiCode: "dashboard_custom_get"},
		&handlerFuncObj{Url: "/dashboard/custom/save", Method: http.MethodPost, HandlerFunc: dashboard.SaveCustomDashboard, ApiCode: "dashboard_custom_save"},
		&handlerFuncObj{Url: "/dashboard/custom/delete", Method: http.MethodGet, HandlerFunc: dashboard.DeleteCustomDashboard, ApiCode: "dashboard_custom_delete"},
		&handlerFuncObj{Url: "/dashboard/server/chart", Method: http.MethodGet, HandlerFunc: dashboard.GetChartsByEndpoint, ApiCode: "dashboard_server_chart"},
		&handlerFuncObj{Url: "/dashboard/custom/main/get", Method: http.MethodGet, HandlerFunc: dashboard.GetMainPage, ApiCode: "dashboard_custom_main_get"},
		&handlerFuncObj{Url: "/dashboard/custom/main/list", Method: http.MethodGet, HandlerFunc: dashboard.ListMainPageRole, ApiCode: "dashboard_custom_main_list"},
		&handlerFuncObj{Url: "/dashboard/custom/main/set", Method: http.MethodPost, HandlerFunc: dashboard.UpdateMainPage, ApiCode: "dashboard_custom_main_set"},
		&handlerFuncObj{Url: "/dashboard/custom/endpoint/get", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointsByIp, ApiCode: "dashboard_custom_endpoint_get"},
		&handlerFuncObj{Url: "/dashboard/custom/role/get", Method: http.MethodGet, HandlerFunc: dashboard.GetCustomDashboardRole, ApiCode: "dashboard_custom_role_get"},
		&handlerFuncObj{Url: "/dashboard/custom/role/save", Method: http.MethodPost, HandlerFunc: dashboard.SaveCustomDashboardRole, ApiCode: "dashboard_custom_role_save"},
		&handlerFuncObj{Url: "/dashboard/custom/alarm/list/:customDashboardId", Method: http.MethodPost, HandlerFunc: alarm.GetCustomDashboardAlarm, ApiCode: "dashboard_custom_alarm_list"},
		&handlerFuncObj{Url: "/dashboard/config/metric/list", Method: http.MethodGet, HandlerFunc: dashboard.GetPromMetric, ApiCode: "dashboard_config_metric_list"},
		// 层级对象
		&handlerFuncObj{Url: "/dashboard/system/add", Method: http.MethodPost, HandlerFunc: agent.ExportPanelAdd, ApiCode: "dashboard_system_add"},
		&handlerFuncObj{Url: "/dashboard/system/delete", Method: http.MethodPost, HandlerFunc: agent.ExportPanelDelete, ApiCode: "dashboard_system_delete"},
		&handlerFuncObj{Url: "/dashboard/recursive/get", Method: http.MethodGet, HandlerFunc: agent.GetPanelRecursive, ApiCode: "dashboard_recursive_get"},
		&handlerFuncObj{Url: "/dashboard/recursive/endpoint_type/list", Method: http.MethodGet, HandlerFunc: agent.GetPanelRecursiveEndpointType, ApiCode: "dashboard_recursive_endpoint_list"},
		// 指标配置
		&handlerFuncObj{Url: "/dashboard/endpoint/type", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointTypeList, ApiCode: "dashboard_endpoint_type"},
		&handlerFuncObj{Url: "/dashboard/endpoint/type_new", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointTypeNewList, ApiCode: "dashboard_endpoint_type_new"},
		&handlerFuncObj{Url: "/dashboard/endpoint", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointList, ApiCode: "dashboard_endpoint"},
		&handlerFuncObj{Url: "/dashboard/endpoint/metric/list", Method: http.MethodPost, HandlerFunc: dashboard.GetEndpointMetric, ApiCode: "dashboard_endpoint_metric_list"},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodGet, HandlerFunc: dashboard_new.MetricList, ApiCode: "dashboard_new_metric_get"},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodPost, HandlerFunc: dashboard_new.MetricCreate, ApiCode: "dashboard_new_metric_post"},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodPut, HandlerFunc: dashboard_new.MetricUpdate, ApiCode: "dashboard_new_metric_put"},
		&handlerFuncObj{Url: "/dashboard/new/metric", Method: http.MethodDelete, HandlerFunc: dashboard_new.MetricDelete, ApiCode: "dashboard_new_metric_delete"},
		&handlerFuncObj{Url: "/dashboard/new/panel", Method: http.MethodGet, HandlerFunc: dashboard_new.PanelList, ApiCode: "dashboard_new_panel_get"},
		&handlerFuncObj{Url: "/dashboard/new/panel/:endpointType", Method: http.MethodPost, HandlerFunc: dashboard_new.PanelCreate, ApiCode: "dashboard_new_panel_endpoint"},
		&handlerFuncObj{Url: "/dashboard/new/panel", Method: http.MethodPut, HandlerFunc: dashboard_new.PanelUpdate, ApiCode: "dashboard_new_panel_put"},
		&handlerFuncObj{Url: "/dashboard/new/panel", Method: http.MethodDelete, HandlerFunc: dashboard_new.PanelDelete, ApiCode: "dashboard_new_panel_delete"},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodGet, HandlerFunc: dashboard_new.ChartList, ApiCode: "dashboard_new_panel_get"},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodPost, HandlerFunc: dashboard_new.ChartCreate, ApiCode: "dashboard_new_panel_post"},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodPut, HandlerFunc: dashboard_new.ChartUpdate, ApiCode: "dashboard_new_panel_put"},
		&handlerFuncObj{Url: "/dashboard/new/chart", Method: http.MethodDelete, HandlerFunc: dashboard_new.ChartDelete, ApiCode: "dashboard_new_panel_delete"},
		&handlerFuncObj{Url: "/dashboard/new/comparison_metric", Method: http.MethodPost, HandlerFunc: monitor.AddOrUpdateComparisonMetric, ApiCode: "dashboard_new_comparison"},
		&handlerFuncObj{Url: "/dashboard/new/comparison_metric/:id", Method: http.MethodDelete, HandlerFunc: monitor.DeleteComparisonMetric, ApiCode: "dashboard_new_comparison_delete"},
	)
	// Agent 对象管理
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/agent/register", Method: http.MethodPost, HandlerFunc: agent.RegisterAgentNew, ApiCode: "agent_register"},
		&handlerFuncObj{Url: "/agent/deregister", Method: http.MethodPost, HandlerFunc: agent.DeregisterAgent, ApiCode: "agent_deregister"},
		&handlerFuncObj{Url: "/agent/export/custom/endpoint/add", Method: http.MethodPost, HandlerFunc: agent.CustomRegister, ApiCode: "agent_export_custom_endpoint"},
		&handlerFuncObj{Url: "/agent/custom/metric/add", Method: http.MethodPost, HandlerFunc: agent.CustomMetricPush, ApiCode: "agent_custom_metric_add"},
		&handlerFuncObj{Url: "/agent/endpoint/telnet/get", Method: http.MethodGet, HandlerFunc: agent.GetEndpointTelnet, ApiCode: "agent_endpoint_telnet_get"},
		&handlerFuncObj{Url: "/agent/endpoint/telnet/update", Method: http.MethodPost, HandlerFunc: agent.UpdateEndpointTelnet, ApiCode: "agent_endpoint_telnet_update"},
		&handlerFuncObj{Url: "/agent/kubernetes/cluster/:operation", Method: http.MethodPost, HandlerFunc: agent.UpdateKubernetesCluster, ApiCode: "agent_kubernetes_cluster_update"},
	)
	// Config 配置
	httpHandlerFuncList = append(httpHandlerFuncList,
		// 对象配置
		&handlerFuncObj{Url: "/alarm/endpoint/list", Method: http.MethodPost, HandlerFunc: alarm.ListGrpEndpoint, ApiCode: "alarm_endpoint_list"},
		&handlerFuncObj{Url: "/alarm/endpoint/options", Method: http.MethodGet, HandlerFunc: alarm.ListGrpEndpointOptions, ApiCode: "alarm_endpoint_options"},
		&handlerFuncObj{Url: "/alarm/endpoint/update", Method: http.MethodPost, HandlerFunc: alarm.EditGrpEndpoint, ApiCode: "alarm_endpoint_update"},
		&handlerFuncObj{Url: "/alarm/process/list", Method: http.MethodGet, HandlerFunc: alarm.GetEndpointProcessConfig, ApiCode: "alarm_process_list"},
		&handlerFuncObj{Url: "/alarm/process/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateEndpointProcessConfig, ApiCode: "alarm_process_update"},
		&handlerFuncObj{Url: "/alarm/window/get", Method: http.MethodGet, HandlerFunc: alarm.GetAlertWindowList, ApiCode: "alarm_window_get"},
		&handlerFuncObj{Url: "/alarm/window/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateAlertWindow, ApiCode: "alarm_window_update"},
		// db查询监控配置
		//&handlerFuncObj{Url: "/alarm/db/monitor/list", Method: http.MethodGet, HandlerFunc: alarm.GetDbMonitorList, ApiCode: "alarm_db_monitor_list"},
		//&handlerFuncObj{Url: "/alarm/db/monitor/add", Method: http.MethodPost, HandlerFunc: alarm.AddDbMonitor, ApiCode: "alarm_db_monitor_add"},
		//&handlerFuncObj{Url: "/alarm/db/monitor/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateDbMonitor, ApiCode: "alarm_db_monitor_update"},
		//&handlerFuncObj{Url: "/alarm/db/monitor/check", Method: http.MethodPost, HandlerFunc: alarm.CheckDbMonitor, ApiCode: "alarm_db_monitor_check"},
		//&handlerFuncObj{Url: "/alarm/db/monitor/delete", Method: http.MethodPost, HandlerFunc: alarm.DeleteDbMonitor, ApiCode: "alarm_db_monitor_delete"},
		//&handlerFuncObj{Url: "/alarm/db/monitor/sys/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateDbMonitorSysName, ApiCode: "alarm_db_monitor_sys_update"},
		// 组配置
		&handlerFuncObj{Url: "/alarm/grp/list", Method: http.MethodGet, HandlerFunc: alarm.ListGrp, ApiCode: "alarm_grp_list"},
		&handlerFuncObj{Url: "/alarm/grp/add", Method: http.MethodPost, HandlerFunc: alarm.AddGrp, ApiCode: "alarm_grp_add"},
		&handlerFuncObj{Url: "/alarm/grp/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateGrp, ApiCode: "alarm_grp_update"},
		&handlerFuncObj{Url: "/alarm/grp/delete", Method: http.MethodGet, HandlerFunc: alarm.DeleteGrp, ApiCode: "alarm_grp_delete"},
		&handlerFuncObj{Url: "/alarm/grp/role/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateGrpRole, ApiCode: "alarm_grp_role_update"},
		&handlerFuncObj{Url: "/alarm/grp/role/get", Method: http.MethodGet, HandlerFunc: alarm.GetGrpRole, ApiCode: "alarm_grp_role_get"},
		&handlerFuncObj{Url: "/alarm/endpoint/grp/update", Method: http.MethodPost, HandlerFunc: alarm.EditEndpointGrp, ApiCode: "alarm_endpoint_grp_update"},
		&handlerFuncObj{Url: "/alarm/grp/export", Method: http.MethodGet, HandlerFunc: alarm.ExportGrpStrategy, ApiCode: "alarm_grp_export"},
		&handlerFuncObj{Url: "/alarm/grp/import", Method: http.MethodPost, HandlerFunc: alarm.ImportGrpStrategy, ApiCode: "alarm_grp_import"},
		// 阈值配置
		&handlerFuncObj{Url: "/alarm/strategy/search", Method: http.MethodGet, HandlerFunc: alarm.SearchObjOption, ApiCode: "alarm_strategy_search"},
		&handlerFuncObj{Url: "/alarm/strategy/list", Method: http.MethodGet, HandlerFunc: alarm.ListTpl, ApiCode: "alarm_strategy_list"},
		&handlerFuncObj{Url: "/alarm/strategy/add", Method: http.MethodPost, HandlerFunc: alarm.AddStrategy, ApiCode: "alarm_strategy_add"},
		&handlerFuncObj{Url: "/alarm/strategy/update", Method: http.MethodPost, HandlerFunc: alarm.EditStrategy, ApiCode: "alarm_strategy_update"},
		&handlerFuncObj{Url: "/alarm/strategy/delete", Method: http.MethodGet, HandlerFunc: alarm.DeleteStrategy, ApiCode: "alarm_strategy_delete"},
		&handlerFuncObj{Url: "/alarm/action/search", Method: http.MethodGet, HandlerFunc: alarm.SearchUserRole, ApiCode: "alarm_action_search"},
		&handlerFuncObj{Url: "/alarm/action/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateTplAction, ApiCode: "alarm_action_update"},
		// 告警列表
		&handlerFuncObj{Url: "/alarm/history", Method: http.MethodGet, HandlerFunc: alarm.GetHistoryAlarm, ApiCode: "alarm_history"},
		&handlerFuncObj{Url: "/alarm/problem/options", Method: http.MethodPost, HandlerFunc: alarm.GetProblemAlarmOptions, ApiCode: "alarm_problem_options"},
		&handlerFuncObj{Url: "/alarm/problem/list", Method: http.MethodGet, HandlerFunc: alarm.GetProblemAlarm, ApiCode: "alarm_problem_list"},
		&handlerFuncObj{Url: "/alarm/problem/query", Method: http.MethodPost, HandlerFunc: alarm.QueryProblemAlarm, ApiCode: "alarm_problem_query"},
		&handlerFuncObj{Url: "/alarm/problem/page", Method: http.MethodPost, HandlerFunc: alarm.QueryProblemAlarmByPage, ApiCode: "alarm_problem_page"},
		&handlerFuncObj{Url: "/alarm/problem/close", Method: http.MethodPost, HandlerFunc: alarm.CloseAlarm, ApiCode: "alarm_problem_close"},
		&handlerFuncObj{Url: "/alarm/problem/history", Method: http.MethodPost, HandlerFunc: alarm.QueryHistoryAlarm, ApiCode: "alarm_problem_history"},
		&handlerFuncObj{Url: "/alarm/problem/message", Method: http.MethodPost, HandlerFunc: alarm.UpdateAlarmCustomMessage, ApiCode: "alarm_problem_message"},
		&handlerFuncObj{Url: "/alarm/problem/notify", Method: http.MethodPost, HandlerFunc: alarm.NotifyAlarm, ApiCode: "alarm_problem_notify"},
		// 关键字监控配置
		&handlerFuncObj{Url: "/alarm/log/monitor/list", Method: http.MethodGet, HandlerFunc: alarm.ListLogTpl, ApiCode: "alarm_log_monitor_list"},
		&handlerFuncObj{Url: "/alarm/log/monitor/add", Method: http.MethodPost, HandlerFunc: alarm.AddLogStrategy, ApiCode: "alarm_log_monitor_add"},
		&handlerFuncObj{Url: "/alarm/log/monitor/update", Method: http.MethodPost, HandlerFunc: alarm.EditLogStrategy, ApiCode: "alarm_log_monitor_update"},
		&handlerFuncObj{Url: "/alarm/log/monitor/update_path", Method: http.MethodPost, HandlerFunc: alarm.EditLogPath, ApiCode: "alarm_log_monitor_update_path"},
		&handlerFuncObj{Url: "/alarm/log/monitor/delete", Method: http.MethodGet, HandlerFunc: alarm.DeleteLogStrategy, ApiCode: "alarm_log_monitor_delete"},
		&handlerFuncObj{Url: "/alarm/log/monitor/delete_path", Method: http.MethodGet, HandlerFunc: alarm.DeleteLogPath, ApiCode: "alarm_log_monitor_delete_path"},
		// 业务日志监控配置
		&handlerFuncObj{Url: "/alarm/business/list", Method: http.MethodGet, HandlerFunc: alarm.GetEndpointBusinessConfig, ApiCode: "alarm_business_list"},
		&handlerFuncObj{Url: "/alarm/business/add", Method: http.MethodPost, HandlerFunc: alarm.AddEndpointBusinessConfig, ApiCode: "alarm_business_add"},
		&handlerFuncObj{Url: "/alarm/business/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateEndpointBusinessConfig, ApiCode: "alarm_business_update"},
		// 层级对象配置
		&handlerFuncObj{Url: "/alarm/org/panel/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrganizaionList, ApiCode: "alarm_org_panel_get"},
		&handlerFuncObj{Url: "/alarm/org/panel/:name", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanel, ApiCode: "alarm_org_panel_update"},
		&handlerFuncObj{Url: "/alarm/org/role/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgPanelRole, ApiCode: "alarm_org_role_get"},
		&handlerFuncObj{Url: "/alarm/org/role/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanelRole, ApiCode: "alarm_org_role_update"},
		&handlerFuncObj{Url: "/alarm/org/endpoint/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgPanelEndpoint, ApiCode: "alarm_org_endpoint_get"},
		&handlerFuncObj{Url: "/alarm/org/endpoint/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanelEndpoint, ApiCode: "alarm_org_endpoint_update"},
		&handlerFuncObj{Url: "/alarm/org/plugin", Method: http.MethodGet, HandlerFunc: alarm.IsPluginMode, ApiCode: "alarm_org_plugin"},
		&handlerFuncObj{Url: "/alarm/org/callback/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgPanelEventList, ApiCode: "alarm_org_callback_get"},
		&handlerFuncObj{Url: "/alarm/org/callback/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgPanelCallback, ApiCode: "alarm_org_callback_update"},
		&handlerFuncObj{Url: "/alarm/org/connect/get", Method: http.MethodGet, HandlerFunc: alarm.GetOrgConnect, ApiCode: "alarm_org_connect_get"},
		&handlerFuncObj{Url: "/alarm/org/connect/update", Method: http.MethodPost, HandlerFunc: alarm.UpdateOrgConnect, ApiCode: "alarm_org_connect_update"},
		&handlerFuncObj{Url: "/alarm/org/search", Method: http.MethodGet, HandlerFunc: alarm.SearchSysPanelData, ApiCode: "alarm_org_search"},
		// 采集器配置
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodGet, HandlerFunc: config_new.SnmpExporterList, ApiCode: "config_snmp_get"},
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodPost, HandlerFunc: config_new.SnmpExporterCreate, ApiCode: "config_snmp_post"},
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodPut, HandlerFunc: config_new.SnmpExporterUpdate, ApiCode: "config_snmp_put"},
		&handlerFuncObj{Url: "/config/new/snmp", Method: http.MethodDelete, HandlerFunc: config_new.SnmpExporterDelete, ApiCode: "config_snmp_delete"},
	)
	// User
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/user/message/get", Method: http.MethodGet, HandlerFunc: user.GetUserMsg, ApiCode: "user_message_get"},
		&handlerFuncObj{Url: "/user/message/update", Method: http.MethodPost, HandlerFunc: user.UpdateUserMsg, ApiCode: "user_message_update"},
		&handlerFuncObj{Url: "/user/list", Method: http.MethodGet, HandlerFunc: user.ListUser, ApiCode: "user_list"},
		&handlerFuncObj{Url: "/user/role/update", Method: http.MethodPost, HandlerFunc: user.UpdateRole, ApiCode: "user_role_update"},
		&handlerFuncObj{Url: "/user/role/list", Method: http.MethodGet, HandlerFunc: user.ListRole, ApiCode: "user_role_list"},
		&handlerFuncObj{Url: "/user/manage_role/list", Method: http.MethodGet, HandlerFunc: user.ListManageRole, ApiCode: "user_manage_role_list"},
		&handlerFuncObj{Url: "/user/role/user/update", Method: http.MethodPost, HandlerFunc: user.UpdateRoleUser, ApiCode: "user_role_user_update"},
	)
	// Export plugin interface
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/agent/export/register/:name", Method: http.MethodPost, HandlerFunc: agent.ExportAgentNew, ApiCode: "agent_export_register"},
		&handlerFuncObj{Url: "/agent/export/deregister/:name", Method: http.MethodPost, HandlerFunc: agent.ExportAgentNew, ApiCode: "agent_export_dregister"},
		&handlerFuncObj{Url: "/agent/export/start/:name", Method: http.MethodPost, HandlerFunc: agent.AlarmControl, ApiCode: "agent_export_start"},
		&handlerFuncObj{Url: "/agent/export/stop/:name", Method: http.MethodPost, HandlerFunc: agent.AlarmControl, ApiCode: "agent_export_stop"},
		&handlerFuncObj{Url: "/agent/export/ping/source", Method: http.MethodGet, HandlerFunc: agent.ExportPingSource, ApiCode: "agent_export_ping_source"},
		&handlerFuncObj{Url: "/agent/export/process/:operation", Method: http.MethodPost, HandlerFunc: agent.AutoUpdateProcessMonitor, ApiCode: "agent_export_process"},
		&handlerFuncObj{Url: "/agent/export/log_monitor/:operation", Method: http.MethodPost, HandlerFunc: agent.AutoUpdateLogMonitor, ApiCode: "agent_export_log_monitor"},
		&handlerFuncObj{Url: "/agent/export/kubernetes/cluster/:action", Method: http.MethodPost, HandlerFunc: agent.PluginKubernetesCluster, ApiCode: "agent_export_kubernetes_cluster"},
		&handlerFuncObj{Url: "/agent/export/kubernetes/pod/:action", Method: http.MethodPost, HandlerFunc: agent.PluginKubernetesPod, ApiCode: "agent_export_kubernetes_pod"},
		&handlerFuncObj{Url: "/agent/export/snmp/exporter/:action", Method: http.MethodPost, HandlerFunc: config_new.PluginSnmpExporterHandle, ApiCode: "agent_export_snmp_exporter"},
	)
	// V2
	httpHandlerFuncListV2 = append(httpHandlerFuncListV2,
		// service
		&handlerFuncObj{Url: "/service_endpoint/search/:searchType", Method: http.MethodGet, HandlerFunc: service.GetServiceGroupEndpointList, ApiCode: "service_endpoint_search"},
		&handlerFuncObj{Url: "/service/log_metric/list/:queryType/:guid", Method: http.MethodGet, HandlerFunc: service.ListLogMetricMonitor, ApiCode: "service_log_metric_list"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor/:logMonitorGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricMonitor, ApiCode: "service_log_metric_monitor_get"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricMonitor, ApiCode: "service_log_metric_monitor_create"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricMonitor, ApiCode: "service_log_metric_monitor_update"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_monitor/:logMonitorGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricMonitor, ApiCode: "service_log_metric_monitor_delete"},
		&handlerFuncObj{Url: "/service/service_group/endpoint_rel", Method: http.MethodGet, HandlerFunc: service.GetServiceGroupEndpointRel, ApiCode: "service_service_group_endpoint_rel"},
		&handlerFuncObj{Url: "/service/log_metric/export", Method: http.MethodGet, HandlerFunc: service.ExportLogMetric, ApiCode: "service_log_metric_export"},
		&handlerFuncObj{Url: "/service/log_metric/import", Method: http.MethodPost, HandlerFunc: service.ImportLogMetric, ApiCode: "service_log_metric_import"},

		&handlerFuncObj{Url: "/service/log_metric/log_metric_json/:logMonitorJsonGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricJson, ApiCode: "service_log_metric_json_get"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_json", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricJson, ApiCode: "service_log_metric_json_create"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_json", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricJson, ApiCode: "service_log_metric_json_update"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_json/:logMonitorJsonGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricJson, ApiCode: "service_log_metric_json_delete"},

		&handlerFuncObj{Url: "/service/log_metric/log_metric_config/:logMonitorConfigGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricConfig, ApiCode: "service_log_metric_config_get"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_config", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricConfig, ApiCode: "service_log_metric_config_create"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_config", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricConfig, ApiCode: "service_log_metric_config_update"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_config/:logMonitorConfigGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricConfig, ApiCode: "service_log_metric_config_delete"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_import/excel/:logMonitorGuid", Method: http.MethodPost, HandlerFunc: service.ImportLogMetricExcel, ApiCode: "service_log_metric_import_excel"},
		&handlerFuncObj{Url: "/service/service_group/:serviceGroup/endpoint/:monitorType", Method: http.MethodGet, HandlerFunc: service.ListServiceGroupEndpoint, ApiCode: "service_service_group_endpoint_list"},

		&handlerFuncObj{Url: "/service/db_metric/list/:queryType/:guid", Method: http.MethodGet, HandlerFunc: service.ListDbMetricMonitor, ApiCode: "service_db_metric_list"},
		&handlerFuncObj{Url: "/service/db_metric/:dbMonitorGuid", Method: http.MethodGet, HandlerFunc: service.GetDbMetricMonitor, ApiCode: "service_db_metric_monitor_get"},
		&handlerFuncObj{Url: "/service/db_metric", Method: http.MethodPost, HandlerFunc: service.CreateDbMetricMonitor, ApiCode: "service_db_metric_monitor_create"},
		&handlerFuncObj{Url: "/service/db_metric", Method: http.MethodPut, HandlerFunc: service.UpdateDbMetricMonitor, ApiCode: "service_db_metric_monitor_update"},
		&handlerFuncObj{Url: "/service/db_metric/:dbMonitorGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteDbMetricMonitor, ApiCode: "service_db_metric_monitor_delete"},
		&handlerFuncObj{Url: "/regexp/test/match", Method: http.MethodPost, HandlerFunc: service.CheckRegExpMatch, ApiCode: "service_regexp_test_match"},
		// 关键字告警配置
		&handlerFuncObj{Url: "/service/log_keyword/list", Method: http.MethodGet, HandlerFunc: service.ListLogKeywordMonitor, ApiCode: "service_log_keyword_list"},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_monitor", Method: http.MethodPost, HandlerFunc: service.CreateLogKeywordMonitor, ApiCode: "service_log_keyword_monitor_create"},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_monitor", Method: http.MethodPut, HandlerFunc: service.UpdateLogKeywordMonitor, ApiCode: "service_log_keyword_monitor_update"},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_monitor/:logKeywordMonitorGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogKeywordMonitor, ApiCode: "service_log_keyword_monitor_delete"},
		&handlerFuncObj{Url: "/service/log_keyword/export", Method: http.MethodGet, HandlerFunc: service.ExportLogAndDbKeyword, ApiCode: "service_log_keyword_export"},
		&handlerFuncObj{Url: "/service/log_keyword/import", Method: http.MethodPost, HandlerFunc: service.ImportLogAndDbKeyword, ApiCode: "service_log_keyword_import"},

		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_config", Method: http.MethodPost, HandlerFunc: service.CreateLogKeyword, ApiCode: "service_log_keyword_config_create"},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_config", Method: http.MethodPut, HandlerFunc: service.UpdateLogKeyword, ApiCode: "service_log_keyword_config_update"},
		&handlerFuncObj{Url: "/service/log_keyword/log_keyword_config", Method: http.MethodDelete, HandlerFunc: service.DeleteLogKeyword, ApiCode: "service_log_keyword_config_delete"},

		&handlerFuncObj{Url: "/service/log_keyword/notify", Method: http.MethodPost, HandlerFunc: service.UpdateLogKeywordNotify, ApiCode: "service_log_keyword_notify_update"},
		// 数据库关键字配置
		&handlerFuncObj{Url: "/service/db_keyword/list", Method: http.MethodGet, HandlerFunc: service.ListDBKeywordConfig, ApiCode: "service_db_keyword_list"},
		&handlerFuncObj{Url: "/service/db_keyword/db_keyword_config", Method: http.MethodPost, HandlerFunc: service.CreateDBKeywordConfig, ApiCode: "service_db_keyword_config_create"},
		&handlerFuncObj{Url: "/service/db_keyword/db_keyword_config", Method: http.MethodPut, HandlerFunc: service.UpdateDBKeywordConfig, ApiCode: "service_db_keyword_config_update"},
		&handlerFuncObj{Url: "/service/db_keyword/db_keyword_config", Method: http.MethodDelete, HandlerFunc: service.DeleteDBKeywordConfig, ApiCode: "service_db_keyword_config_delete"},
		// service plugin
		&handlerFuncObj{Url: "/service/plugin/update/path", Method: http.MethodPost, HandlerFunc: service.PluginUpdateServicePath, ApiCode: "service_plugin_update_path"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/query", Method: http.MethodGet, HandlerFunc: alarmv2.ListEndpointGroup, ApiCode: "alarm_endpoint_group_query"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/options", Method: http.MethodGet, HandlerFunc: alarmv2.EndpointGroupOptions, ApiCode: "alarm_endpoint_group_options"},
		&handlerFuncObj{Url: "/alarm/endpoint_group", Method: http.MethodPost, HandlerFunc: alarmv2.CreateEndpointGroup, ApiCode: "alarm_endpoint_group_create"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/import", Method: http.MethodPost, HandlerFunc: alarmv2.ImportEndpointGroup, ApiCode: "alarm_endpoint_group_import"},
		&handlerFuncObj{Url: "/alarm/endpoint_group", Method: http.MethodPut, HandlerFunc: alarmv2.UpdateEndpointGroup, ApiCode: "alarm_endpoint_group_update"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid", Method: http.MethodDelete, HandlerFunc: alarmv2.DeleteEndpointGroup, ApiCode: "alarm_endpoint_group_delete_by_group_guid"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/endpoint/list", Method: http.MethodGet, HandlerFunc: alarmv2.GetGroupEndpointRel, ApiCode: "alarm_endpoint_group_endpoint_list_by_group_guid"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/endpoint/update", Method: http.MethodPost, HandlerFunc: alarmv2.UpdateGroupEndpoint, ApiCode: "alarm_endpoint_group_endpoint_update_by_group_guid"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/notify/list", Method: http.MethodGet, HandlerFunc: alarmv2.GetGroupEndpointNotify, ApiCode: "alarm_endpoint_group_notify_list_by_group_guid"},
		&handlerFuncObj{Url: "/alarm/endpoint_group/:groupGuid/notify/update", Method: http.MethodPost, HandlerFunc: alarmv2.UpdateGroupEndpointNotify, ApiCode: "alarm_endpoint_group_notify_update_by_group_guid"},
		&handlerFuncObj{Url: "/alarm/strategy/search", Method: http.MethodGet, HandlerFunc: alarmv2.ListStrategyQueryOptions, ApiCode: "alarm_strategy_search"},
		&handlerFuncObj{Url: "/alarm/strategy/query", Method: http.MethodPost, HandlerFunc: alarmv2.QueryAlarmStrategy, ApiCode: "alarm_strategy_query"},
		&handlerFuncObj{Url: "/alarm/strategy/workflow", Method: http.MethodGet, HandlerFunc: alarmv2.ListAlarmStrategyWorkFlow, ApiCode: "alarm_strategy_workflow"},
		&handlerFuncObj{Url: "/alarm/strategy", Method: http.MethodPost, HandlerFunc: alarmv2.CreateAlarmStrategy, ApiCode: "alarm_strategy_create"},
		&handlerFuncObj{Url: "/alarm/strategy", Method: http.MethodPut, HandlerFunc: alarmv2.UpdateAlarmStrategy, ApiCode: "alarm_strategy_update"},
		&handlerFuncObj{Url: "/alarm/strategy/:strategyGuid", Method: http.MethodDelete, HandlerFunc: alarmv2.DeleteAlarmStrategy, ApiCode: "alarm_strategy_delete_by_strategy_guid"},
		&handlerFuncObj{Url: "/alarm/event/callback/list", Method: http.MethodGet, HandlerFunc: alarmv2.ListCallbackEvent, ApiCode: "alarm_event_callback_list"},
		&handlerFuncObj{Url: "/alarm/strategy/export/:queryType/:guid", Method: http.MethodGet, HandlerFunc: alarmv2.ExportAlarmStrategy, ApiCode: "alarm_strategy_export_by_query_type_and_guid"},
		&handlerFuncObj{Url: "/alarm/strategy/import/:queryType/:guid", Method: http.MethodPost, HandlerFunc: alarmv2.ImportAlarmStrategy, ApiCode: "alarm_strategy_import_by_query_type_and_guid"},
		&handlerFuncObj{Url: "/monitor/endpoint/query", Method: http.MethodGet, HandlerFunc: monitor.ListEndpoint, ApiCode: "monitor_endpoint_query"},
		&handlerFuncObj{Url: "/monitor/metric/list", Method: http.MethodGet, HandlerFunc: monitor.ListMetric, ApiCode: "monitor_metric_list"},
		&handlerFuncObj{Url: "/monitor/metric/list/count", Method: http.MethodGet, HandlerFunc: monitor.ListMetricCount, ApiCode: "monitor_metric_list_count"},
		&handlerFuncObj{Url: "/monitor/metric_comparison/list", Method: http.MethodGet, HandlerFunc: monitor.ListMetricComparison, ApiCode: "monitor_metric_comparison_list"},
		&handlerFuncObj{Url: "/sys/parameter/metric_template", Method: http.MethodGet, HandlerFunc: monitor.GetSysMetricTemplate, ApiCode: "sys_parameter_metric_template"},
		&handlerFuncObj{Url: "/monitor/endpoint/get/:guid", Method: http.MethodGet, HandlerFunc: monitor.GetEndpoint, ApiCode: "monitor_endpoint_get_by_guid"},
		&handlerFuncObj{Url: "/monitor/endpoint/update", Method: http.MethodPut, HandlerFunc: monitor.UpdateEndpoint, ApiCode: "monitor_endpoint_update"},
		&handlerFuncObj{Url: "/monitor/metric/export", Method: http.MethodGet, HandlerFunc: monitor.ExportMetric, ApiCode: "monitor_metric_export"},
		&handlerFuncObj{Url: "/monitor/metric/import", Method: http.MethodPost, HandlerFunc: monitor.ImportMetric, ApiCode: "monitor_metric_import"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/options", Method: http.MethodGet, HandlerFunc: service.ListLogMonitorTemplateOptions, ApiCode: "service_log_metric_log_monitor_template_options"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/list", Method: http.MethodPost, HandlerFunc: service.ListLogMonitorTemplate, ApiCode: "service_log_metric_log_monitor_template_list"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/:logMonitorTemplateGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMonitorTemplate, ApiCode: "service_log_metric_log_monitor_template_get_by_log_monitor_template_guid"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template", Method: http.MethodPost, HandlerFunc: service.CreateLogMonitorTemplate, ApiCode: "service_log_metric_log_monitor_template_create"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template", Method: http.MethodPut, HandlerFunc: service.UpdateLogMonitorTemplate, ApiCode: "service_log_metric_log_monitor_template_update"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/:logMonitorTemplateGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMonitorTemplate, ApiCode: "service_log_metric_log_monitor_template_delete_by_log_monitor_template_guid"},
		&handlerFuncObj{Url: "/service/log_metric/affect_service_group/:logMonitorTemplateGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMonitorTemplateServiceGroup, ApiCode: "service_log_metric_affect_service_group_by_log_monitor_template_guid"},
		&handlerFuncObj{Url: "/service/log_metric/regexp/match", Method: http.MethodPost, HandlerFunc: service.CheckLogMonitorRegExpMatch, ApiCode: "service_log_metric_regexp_match"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/export", Method: http.MethodPost, HandlerFunc: service.LogMonitorTemplateExport, ApiCode: "service_log_metric_log_monitor_template_export"},
		&handlerFuncObj{Url: "/service/log_metric/log_monitor_template/import", Method: http.MethodPost, HandlerFunc: service.LogMonitorTemplateImport, ApiCode: "service_log_metric_log_monitor_template_import"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group/:logMetricGroupGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricGroup, ApiCode: "service_log_metric_log_metric_group_get_by_log_metric_group_guid"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricGroup, ApiCode: "service_log_metric_log_metric_group_create"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricGroup, ApiCode: "service_log_metric_log_metric_group_update"},
		&handlerFuncObj{Url: "/service/log_metric/log_metric_group/:logMetricGroupGuid", Method: http.MethodDelete, HandlerFunc: service.DeleteLogMetricGroup, ApiCode: "service_log_metric_log_metric_group_delete_by_log_metric_group_guid"},
		&handlerFuncObj{Url: "/service/log_metric/custom/log_metric_group/:logMetricGroupGuid", Method: http.MethodGet, HandlerFunc: service.GetLogMetricCustomGroup, ApiCode: "service_log_metric_custom_log_metric_group_get_by_log_metric_group_guid"},
		&handlerFuncObj{Url: "/service/log_metric/custom/log_metric_group", Method: http.MethodPost, HandlerFunc: service.CreateLogMetricCustomGroup, ApiCode: "service_log_metric_custom_log_metric_group_create"},
		&handlerFuncObj{Url: "/service/log_metric/custom/log_metric_group", Method: http.MethodPut, HandlerFunc: service.UpdateLogMetricCustomGroup, ApiCode: "service_log_metric_custom_log_metric_group_update"},
		&handlerFuncObj{Url: "/service/log_metric/data_map/regexp/match", Method: http.MethodPost, HandlerFunc: service.LogMonitorDataMapRegMatch, ApiCode: "service_log_metric_data_map_regexp_match"},
		&handlerFuncObj{Url: "/metric/tag/value-list", Method: http.MethodPost, HandlerFunc: monitor.QueryMetricTagValue, ApiCode: "metric_tag_value_list"},
		&handlerFuncObj{Url: "/dashboard/all", Method: http.MethodGet, HandlerFunc: monitor.GetAllCustomDashboardList, ApiCode: "dashboard_all"},
		&handlerFuncObj{Url: "/dashboard/custom/list", Method: http.MethodPost, HandlerFunc: monitor.QueryCustomDashboardList, ApiCode: "dashboard_custom_list"},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodGet, HandlerFunc: monitor.GetCustomDashboard, ApiCode: "dashboard_custom_get"},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodPost, HandlerFunc: monitor.AddCustomDashboard, ApiCode: "dashboard_custom_add"},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodPut, HandlerFunc: monitor.UpdateCustomDashboard, ApiCode: "dashboard_custom_update"},
		&handlerFuncObj{Url: "/dashboard/custom", Method: http.MethodDelete, HandlerFunc: monitor.DeleteCustomDashboard, ApiCode: "dashboard_custom_delete"},
		&handlerFuncObj{Url: "/dashboard/custom/copy", Method: http.MethodPost, HandlerFunc: monitor.CopyCustomDashboard, ApiCode: "dashboard_custom_copy"},
		&handlerFuncObj{Url: "/dashboard/custom/permission", Method: http.MethodPost, HandlerFunc: monitor.UpdateCustomDashboardPermission, ApiCode: "dashboard_custom_permission_update"},
		&handlerFuncObj{Url: "/dashboard/custom/export", Method: http.MethodPost, HandlerFunc: monitor.ExportCustomDashboard, ApiCode: "dashboard_custom_export"},
		&handlerFuncObj{Url: "/dashboard/custom/import", Method: http.MethodPost, HandlerFunc: monitor.ImportCustomDashboard, ApiCode: "dashboard_custom_import"},
		&handlerFuncObj{Url: "/dashboard/custom/trans_import", Method: http.MethodPost, HandlerFunc: monitor.TransImportCustomDashboard, ApiCode: "dashboard_custom_trans_import"},
		&handlerFuncObj{Url: "/chart/shared/list", Method: http.MethodPost, HandlerFunc: monitor.GetSharedChartList, ApiCode: "chart_shared_list"},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodPost, HandlerFunc: monitor.AddCustomChart, ApiCode: "chart_custom_add"},
		&handlerFuncObj{Url: "/chart/custom/copy", Method: http.MethodPost, HandlerFunc: monitor.CopyCustomChart, ApiCode: "chart_custom_copy"},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodPut, HandlerFunc: monitor.UpdateCustomChart, ApiCode: "chart_custom_update"},
		&handlerFuncObj{Url: "/chart/custom/name", Method: http.MethodPut, HandlerFunc: monitor.UpdateCustomChartName, ApiCode: "chart_custom_name_update"},
		&handlerFuncObj{Url: "/chart/custom/name/exist", Method: http.MethodGet, HandlerFunc: monitor.QueryCustomChartNameExist, ApiCode: "chart_custom_name_exist"},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodGet, HandlerFunc: monitor.GetCustomChart, ApiCode: "chart_custom_get"},
		&handlerFuncObj{Url: "/chart/custom", Method: http.MethodDelete, HandlerFunc: monitor.DeleteCustomChart, ApiCode: "chart_custom_delete"},
		&handlerFuncObj{Url: "/chart/custom/permission", Method: http.MethodPost, HandlerFunc: monitor.SharedCustomChart, ApiCode: "chart_custom_permission_post"},
		&handlerFuncObj{Url: "/chart/custom/permission", Method: http.MethodGet, HandlerFunc: monitor.GetSharedChartPermission, ApiCode: "chart_custom_permission_get"},
		&handlerFuncObj{Url: "/chart/custom/permission/batch", Method: http.MethodPost, HandlerFunc: monitor.GetSharedChartPermissionBatch, ApiCode: "chart_custom_permission_batch"},
		&handlerFuncObj{Url: "/chart/manage/list", Method: http.MethodPost, HandlerFunc: monitor.QueryCustomChart, ApiCode: "chart_manage_list"},
		&handlerFuncObj{Url: "/dashboard/data/sync", Method: http.MethodPost, HandlerFunc: monitor.SyncData, ApiCode: "dashboard_data_sync"},
		&handlerFuncObj{Url: "/chart/custom/series/config", Method: http.MethodPost, HandlerFunc: monitor.GetChartSeriesColor, ApiCode: "chart_custom_series_config"},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodGet, HandlerFunc: config_new.RemoteWriteConfigList, ApiCode: "config_remote_write_get"},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodPost, HandlerFunc: config_new.RemoteWriteConfigCreate, ApiCode: "config_remote_write_post"},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodPut, HandlerFunc: config_new.RemoteWriteConfigUpdate, ApiCode: "config_remote_write_put"},
		&handlerFuncObj{Url: "/config/remote/write", Method: http.MethodDelete, HandlerFunc: config_new.RemoteWriteConfigDelete, ApiCode: "config_remote_write_delete"},
		&handlerFuncObj{Url: "/config/type/query", Method: http.MethodGet, HandlerFunc: monitor.QueryTypeConfigList, ApiCode: "config_type_query"},
		&handlerFuncObj{Url: "/config/type", Method: http.MethodPost, HandlerFunc: monitor.AddTypeConfig, ApiCode: "config_type_add"},
		&handlerFuncObj{Url: "/config/type-batch", Method: http.MethodPost, HandlerFunc: monitor.BatchAddTypeConfig, ApiCode: "config_type_batch_add"},
		&handlerFuncObj{Url: "/config/type", Method: http.MethodDelete, HandlerFunc: monitor.DeleteTypeConfig, ApiCode: "config_type_delete"},
		&handlerFuncObj{Url: "/seed", Method: http.MethodGet, HandlerFunc: monitor.GetEncryptSeed, ApiCode: "seed_get"},
		&handlerFuncObj{Url: "/trans-export/analyze", Method: http.MethodPost, HandlerFunc: monitor.AnalyzeTransExportData, ApiCode: "trans_export_analyze"},
		&handlerFuncObj{Url: "/trans-export/log_monitor_template/batch", Method: http.MethodPost, HandlerFunc: service.BatchGetLogMonitorTemplate, ApiCode: "trans_export_log_monitor_template_batch"},
		&handlerFuncObj{Url: "/trans-export/dashboard/batch", Method: http.MethodPost, HandlerFunc: monitor.BatchGetDashboard, ApiCode: "trans_export_dashboard_batch"},
		&handlerFuncObj{Url: "/trans-export/service_group/batch", Method: http.MethodPost, HandlerFunc: alarm.BatchGetServiceGroup, ApiCode: "trans_export_service_group_batch"},
		&handlerFuncObj{Url: "/trans-export/config/type/batch", Method: http.MethodPost, HandlerFunc: monitor.BatchGetTypeConfigList, ApiCode: "trans_export_config_type_batch"},
	)
}

func InitHttpServer() {
	middleware.InitHttpError()
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
		entityApi.POST("/endpoint/query", monitor.QueryEntityEndpoint)
		entityApi.POST("/service_group/query", monitor.QueryEntityServiceGroup)
		entityApi.POST("/endpoint_group/query", monitor.QueryEntityEndpointGroup)
		entityApi.POST("/monitor_type/query", monitor.QueryEntityMonitorType)
		entityApi.POST("/log_monitor_template/query", monitor.QueryEntityLogMonitorTemplate)
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
		apiCodeMap[fmt.Sprintf("%s_%s%s", funcObj.Method, models.UrlPrefix+"/api/v1", funcObj.Url)] = funcObj.ApiCode
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
		apiCodeMap[fmt.Sprintf("%s_%s%s", funcObj.Method, models.UrlPrefix+"/api/v2", funcObj.Url)] = funcObj.ApiCode
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
		apiCode := apiCodeMap[c.Request.Method+"_"+c.FullPath()]
		c.Writer.Header().Add("Api-Code", apiCode)
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
