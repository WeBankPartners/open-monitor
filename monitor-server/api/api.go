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
	httpHandlerFuncList []*handlerFuncObj
)

func init() {
	// Dashboard 视图
	httpHandlerFuncList = append(httpHandlerFuncList,
		// 对象视图
		&handlerFuncObj{Url: "/dashboard/main", Method: http.MethodGet, HandlerFunc: dashboard.MainDashboard},
		&handlerFuncObj{Url: "/dashboard/panels", Method: http.MethodGet, HandlerFunc: dashboard.GetPanels},
		&handlerFuncObj{Url: "/dashboard/tags", Method: http.MethodGet, HandlerFunc: dashboard.GetTags},
		&handlerFuncObj{Url: "/dashboard/search", Method: http.MethodGet, HandlerFunc: dashboard.MainSearch},
		&handlerFuncObj{Url: "/dashboard/newchart", Method: http.MethodPost, HandlerFunc: dashboard.GetChart},
		&handlerFuncObj{Url: "/dashboard/chart", Method: http.MethodPost, HandlerFunc: dashboard_new.GetChartData},
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
		&handlerFuncObj{Url: "/dashboard/custom/alarm/list", Method: http.MethodGet, HandlerFunc: alarm.GetCustomDashboardAlarm},
		&handlerFuncObj{Url: "/dashboard/config/metric/list", Method: http.MethodGet, HandlerFunc: dashboard.GetPromMetric},
		// 层级对象
		&handlerFuncObj{Url: "/dashboard/system/add", Method: http.MethodPost, HandlerFunc: agent.ExportPanelAdd},
		&handlerFuncObj{Url: "/dashboard/system/delete", Method: http.MethodPost, HandlerFunc: agent.ExportPanelDelete},
		&handlerFuncObj{Url: "/dashboard/recursive/get", Method: http.MethodGet, HandlerFunc: agent.GetPanelRecursive},
		&handlerFuncObj{Url: "/dashboard/recursive/endpoint_type/list", Method: http.MethodGet, HandlerFunc: agent.GetPanelRecursiveEndpointType},
		// 指标配置
		&handlerFuncObj{Url: "/dashboard/endpoint/type", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointTypeList},
		&handlerFuncObj{Url: "/dashboard/endpoint", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointList},
		&handlerFuncObj{Url: "/dashboard/endpoint/metric/list", Method: http.MethodGet, HandlerFunc: dashboard.GetEndpointMetric},
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
	)
	// Agent 对象管理
	httpHandlerFuncList = append(httpHandlerFuncList,
		&handlerFuncObj{Url: "/agent/register", Method: http.MethodPost, HandlerFunc: agent.RegisterAgentNew},
		&handlerFuncObj{Url: "/agent/deregister", Method: http.MethodGet, HandlerFunc: agent.DeregisterAgent},
		&handlerFuncObj{Url: "/agent/custom/endpoint/add", Method: http.MethodPost, HandlerFunc: agent.CustomRegister},
		&handlerFuncObj{Url: "/agent/custom/metric/add", Method: http.MethodPost, HandlerFunc: agent.CustomMetricPush},
		&handlerFuncObj{Url: "/agent/endpoint/telnet/get", Method: http.MethodGet, HandlerFunc: agent.GetEndpointTelnet},
		&handlerFuncObj{Url: "/agent/endpoint/telnet/update", Method: http.MethodPost, HandlerFunc: agent.UpdateEndpointTelnet},
		&handlerFuncObj{Url: "/agent/kubernetes/cluster/:operation", Method: http.MethodPost, HandlerFunc: agent.UpdateKubernetesCluster},
	)
	// Config 配置
	httpHandlerFuncList = append(httpHandlerFuncList,
		// 对象配置
		&handlerFuncObj{Url: "/alarm/endpoint/list", Method: http.MethodGet, HandlerFunc: alarm.ListGrpEndpoint},
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
		&handlerFuncObj{Url: "/alarm/problem/list", Method: http.MethodGet, HandlerFunc: alarm.GetProblemAlarm},
		&handlerFuncObj{Url: "/alarm/problem/query", Method: http.MethodPost, HandlerFunc: alarm.QueryProblemAlarm},
		&handlerFuncObj{Url: "/alarm/problem/close", Method: http.MethodGet, HandlerFunc: alarm.CloseAlarm},
		&handlerFuncObj{Url: "/alarm/problem/history", Method: http.MethodPost, HandlerFunc: alarm.QueryHistoryAlarm},
		&handlerFuncObj{Url: "/alarm/problem/message", Method: http.MethodPost, HandlerFunc: alarm.UpdateAlarmCustomMessage},
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
	r.POST(fmt.Sprintf("%s/webhook", urlPrefix), alarm.AcceptAlertMsg)
	r.POST(fmt.Sprintf("%s/openapi/alarm/send", urlPrefix), alarm.OpenAlarmApi)
	entityApi := r.Group(fmt.Sprintf("%s/entities", urlPrefix), user.AuthRequired())
	{
		entityApi.POST("/alarm/query", alarm.QueryEntityAlarm)
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
	netIp, ok := c.RemoteIP()
	if ok {
		return netIp.String()
	}
	return c.ClientIP()
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
