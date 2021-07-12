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
	_ "github.com/WeBankPartners/open-monitor/monitor-server/docs"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func InitHttpServer(exportAgent bool) {
	urlPrefix := "/monitor"
	r := gin.Default()
	r.LoadHTMLGlob("public/*.html")
	r.Static(fmt.Sprintf("%s/js", urlPrefix), fmt.Sprintf("public%s/js", urlPrefix))
	r.Static(fmt.Sprintf("%s/css", urlPrefix), fmt.Sprintf("public%s/css", urlPrefix))
	r.Static(fmt.Sprintf("%s/img", urlPrefix), fmt.Sprintf("public%s/img", urlPrefix))
	r.Static(fmt.Sprintf("%s/fonts", urlPrefix), fmt.Sprintf("public%s/fonts", urlPrefix))
	r.StaticFile("/favicon.ico", "public/favicon.ico")
	if exportAgent {
		r.Static(fmt.Sprintf("%s/exporter", urlPrefix), "exporter")
	}
	if m.Config().Http.Cross {
		r.Use(func(c *gin.Context) {
			// Deal with options request
			if c.Request.Method == "OPTIONS" {
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, authorization, Token, X-Auth-Token")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
				if c.GetHeader("Origin") != "" {
					c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
				}else{
					c.Header("Access-Control-Allow-Origin", "*")
				}
				c.AbortWithStatus(http.StatusNoContent)
			}else{
				c.Next()
			}
		})
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "authorization", "Token", "X-Auth-Token"}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
		corsConfig.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
		corsConfig.AllowCredentials = true
		r.Use(cors.New(corsConfig))
		r.Use(func(c *gin.Context) {
			if c.GetHeader("Origin") != "" {
				c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
			}
		})
	}
	r.Use(httpLogHandle())
	// public api
	r.GET(fmt.Sprintf("%s/", urlPrefix), func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	if m.Config().Http.Ldap.Enable {
		r.POST(fmt.Sprintf("%s/login", urlPrefix), user.LdapLogin)
	}else{
		r.POST(fmt.Sprintf("%s/login", urlPrefix), user.Login)
	}
	r.POST(fmt.Sprintf("%s/register", urlPrefix), user.Register)
	r.GET(fmt.Sprintf("%s/logout", urlPrefix), user.Logout)
	r.GET(fmt.Sprintf("%s/check", urlPrefix), user.HealthCheck)
	r.GET(fmt.Sprintf("%s/demo", urlPrefix), dashboard.DisplayWatermark)
	if m.Config().Http.Swagger {
		r.GET(fmt.Sprintf("%s/swagger/*any", urlPrefix), ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	entityApi := r.Group(fmt.Sprintf("%s/entities", urlPrefix), user.AuthRequired())
	{
		entityApi.GET("/alarm", alarm.GetEntityAlarm)
		entityApi.GET("/test", alarm.TestNotifyAlarm)
		entityApi.POST("/alarm/query", alarm.QueryEntityAlarm)
	}
	r.POST(fmt.Sprintf("%s/webhook", urlPrefix), alarm.AcceptAlertMsg)
	// auth server side
	serverApi := r.Group(fmt.Sprintf("%s/openapi", urlPrefix))
	{
		serverApi.POST("/alarm/send", alarm.OpenAlarmApi)
	}
	// auth api
	authApi := r.Group(fmt.Sprintf("%s/api/v1", urlPrefix), user.AuthRequired())
	{
		dashboardApi := authApi.Group("/dashboard")
		{
			dashboardApi.GET("/main", dashboard.MainDashboard)
			dashboardApi.GET("/panels", dashboard.GetPanels)
			dashboardApi.GET("/chart", dashboard.GetChartOld)
			dashboardApi.GET("/tags", dashboard.GetTags)
			dashboardApi.GET("/search", dashboard.MainSearch)
			dashboardApi.GET("/config/metric/list", dashboard.GetPromMetric)
			dashboardApi.POST("/newchart", dashboard.GetChart)
			dashboardApi.POST("/pie/chart", dashboard.GetPieChart)
			dashboardApi.POST("/config/panel/update", dashboard.UpdatePanelChartMetric)
			dashboardApi.POST("/config/metric/update", dashboard.UpdatePromMetric)
			dashboardApi.GET("/config/metric/reload", agent.ReloadEndpointMetric)
			dashboardApi.GET("/endpoint/metric/list", dashboard.GetEndpointMetric)
			dashboardApi.GET("/custom/list", dashboard.ListCustomDashboard)
			dashboardApi.GET("/custom/get", dashboard.GetCustomDashboard)
			dashboardApi.POST("/custom/save", dashboard.SaveCustomDashboard)
			dashboardApi.GET("/custom/delete", dashboard.DeleteCustomDashboard)
			dashboardApi.GET("/server/chart", dashboard.GetChartsByEndpoint)
			dashboardApi.GET("/custom/main/get", dashboard.GetMainPage)
			dashboardApi.GET("/custom/main/list", dashboard.ListMainPageRole)
			dashboardApi.POST("/custom/main/set", dashboard.UpdateMainPage)
			dashboardApi.GET("/custom/endpoint/get", dashboard.GetEndpointsByIp)
			dashboardApi.POST("/system/add", agent.ExportPanelAdd)
			dashboardApi.POST("/system/delete", agent.ExportPanelDelete)
			dashboardApi.GET("/recursive/get", agent.GetPanelRecursive)
			dashboardApi.POST("/config/chart/title", dashboard.UpdateChartsTitle)
			dashboardApi.GET("/custom/role/get", dashboard.GetCustomDashboardRole)
			dashboardApi.POST("/custom/role/save", dashboard.SaveCustomDashboardRole)
			dashboardApi.GET("/custom/alarm/list", alarm.GetCustomDashboardAlarm)
			dashboardApi.GET("/panel/list", dashboard.GetDashboardPanelList)
			dashboardApi.GET("/endpoint/type", dashboard.GetEndpointTypeList)
			dashboardApi.GET("/endpoint", dashboard.GetEndpointList)
		}
		dashboardNewApi := authApi.Group("/dashboard/new")
		{
			// dashboard metric curd
			dashboardNewApi.GET("/metric", dashboard_new.MetricList)
			dashboardNewApi.POST("/metric", dashboard_new.MetricCreate)
			dashboardNewApi.PUT("/metric", dashboard_new.MetricUpdate)
			dashboardNewApi.DELETE("/metric", dashboard_new.MetricDelete)
			// dashboard panel curd
			dashboardNewApi.GET("/panel", dashboard_new.PanelList)
			dashboardNewApi.POST("/panel/:endpointType", dashboard_new.PanelCreate)
			dashboardNewApi.PUT("/panel", dashboard_new.PanelUpdate)
			dashboardNewApi.DELETE("/panel", dashboard_new.PanelDelete)
			// dashboard chart curd
			dashboardNewApi.GET("/chart", dashboard_new.ChartList)
			dashboardNewApi.POST("/chart", dashboard_new.ChartCreate)
			dashboardNewApi.PUT("/chart", dashboard_new.ChartUpdate)
			dashboardNewApi.DELETE("/chart", dashboard_new.ChartDelete)
		}
		configNewApi := authApi.Group("/config/new")
		{
			// snmp config curd
			configNewApi.GET("/snmp", config_new.SnmpExporterList)
			configNewApi.POST("/snmp", config_new.SnmpExporterCreate)
			configNewApi.PUT("/snmp", config_new.SnmpExporterUpdate)
			configNewApi.DELETE("/snmp", config_new.SnmpExporterDelete)
		}
		agentApi := authApi.Group("/agent")
		{
			agentApi.POST("/register", agent.RegisterAgentNew)
			agentApi.GET("/deregister", agent.DeregisterAgent)
			agentApi.POST("/export/register/:name", agent.ExportAgentNew)
			agentApi.POST("/export/deregister/:name", agent.ExportAgentNew)
			agentApi.POST("/export/start/:name", agent.AlarmControl)
			agentApi.POST("/export/stop/:name", agent.AlarmControl)
			agentApi.POST("/custom/endpoint/add", agent.CustomRegister)
			agentApi.POST("/custom/metric/add", agent.CustomMetricPush)
			agentApi.POST("/endpoint/telnet/update", agent.UpdateEndpointTelnet)
			agentApi.GET("/export/ping/source", agent.ExportPingSource)
			agentApi.GET("/endpoint/telnet/get", agent.GetEndpointTelnet)
			agentApi.POST("/export/process/:operation", agent.AutoUpdateProcessMonitor)
			agentApi.POST("/export/log_monitor/:operation", agent.AutoUpdateLogMonitor)
			agentApi.POST("/kubernetes/cluster/:operation", agent.UpdateKubernetesCluster)
			agentApi.POST("/export/kubernetes/cluster/:action", agent.PluginKubernetesCluster)
			agentApi.POST("/export/kubernetes/pod/:action", agent.PluginKubernetesPod)
		}
		alarmApi := authApi.Group("/alarm")
		{
			alarmApi.GET("/grp/list", alarm.ListGrp)
			alarmApi.POST("/grp/add", alarm.AddGrp)
			alarmApi.POST("/grp/update", alarm.UpdateGrp)
			alarmApi.GET("/grp/delete", alarm.DeleteGrp)
			alarmApi.POST("/grp/role/update", alarm.UpdateGrpRole)
			alarmApi.GET("/grp/role/get", alarm.GetGrpRole)
			alarmApi.GET("/endpoint/list", alarm.ListGrpEndpoint)
			alarmApi.POST("/endpoint/update", alarm.EditGrpEndpoint)
			alarmApi.POST("/endpoint/grp/update", alarm.EditEndpointGrp)
			alarmApi.GET("/strategy/search", alarm.SearchObjOption)
			alarmApi.GET("/strategy/list", alarm.ListTpl)
			alarmApi.POST("/strategy/add", alarm.AddStrategy)
			alarmApi.POST("/strategy/update", alarm.EditStrategy)
			alarmApi.GET("/strategy/delete", alarm.DeleteStrategy)
			alarmApi.GET("/history", alarm.GetHistoryAlarm)
			alarmApi.GET("/problem/list", alarm.GetProblemAlarm)
			alarmApi.POST("/problem/query", alarm.QueryProblemAlarm)
			alarmApi.GET("/problem/close", alarm.CloseALarm)
			alarmApi.GET("/log/monitor/list", alarm.ListLogTpl)
			alarmApi.POST("/log/monitor/add", alarm.AddLogStrategy)
			alarmApi.POST("/log/monitor/update", alarm.EditLogStrategy)
			alarmApi.POST("/log/monitor/update_path", alarm.EditLogPath)
			alarmApi.GET("/log/monitor/delete", alarm.DeleteLogStrategy)
			alarmApi.GET("/log/monitor/delete_path", alarm.DeleteLogPath)
			alarmApi.GET("/process/list", alarm.GetEndpointProcessConfig)
			alarmApi.POST("/process/update", alarm.UpdateEndpointProcessConfig)
			alarmApi.GET("/business/list", alarm.GetEndpointBusinessConfig)
			alarmApi.POST("/business/add", alarm.AddEndpointBusinessConfig)
			alarmApi.POST("/business/update", alarm.UpdateEndpointBusinessConfig)
			alarmApi.GET("/grp/export", alarm.ExportGrpStrategy)
			alarmApi.POST("/grp/import", alarm.ImportGrpStrategy)
			alarmApi.GET("/action/search", alarm.SearchUserRole)
			alarmApi.POST("/action/update", alarm.UpdateTplAction)
			alarmApi.GET("/org/panel/get", alarm.GetOrganizaionList)
			alarmApi.POST("/org/panel/:name", alarm.UpdateOrgPanel)
			alarmApi.GET("/org/role/get", alarm.GetOrgPanelRole)
			alarmApi.POST("/org/role/update", alarm.UpdateOrgPanelRole)
			alarmApi.GET("/org/endpoint/get", alarm.GetOrgPanelEndpoint)
			alarmApi.POST("/org/endpoint/update", alarm.UpdateOrgPanelEndpoint)
			alarmApi.GET("/org/plugin", alarm.IsPluginMode)
			alarmApi.GET("/org/callback/get", alarm.GetOrgPanelEventList)
			alarmApi.POST("/org/callback/update", alarm.UpdateOrgPanelCallback)
			alarmApi.GET("/org/connect/get", alarm.GetOrgConnect)
			alarmApi.POST("/org/connect/update", alarm.UpdateOrgConnect)
			alarmApi.GET("/org/search", alarm.SearchSysPanelData)
			alarmApi.GET("/db/monitor/list", alarm.GetDbMonitorList)
			alarmApi.POST("/db/monitor/add", alarm.AddDbMonitor)
			alarmApi.POST("/db/monitor/update", alarm.UpdateDbMonitor)
			alarmApi.POST("/db/monitor/check", alarm.CheckDbMonitor)
			alarmApi.POST("/db/monitor/delete", alarm.DeleteDbMonitor)
			alarmApi.POST("/db/monitor/sys/update", alarm.UpdateDbMonitorSysName)
			alarmApi.POST("/problem/history", alarm.QueryHistoryAlarm)
			alarmApi.GET("/window/get", alarm.GetAlertWindowList)
			alarmApi.POST("/window/update", alarm.UpdateAlertWindow)
		}
		userApi := authApi.Group("/user")
		{
			userApi.GET("/message/get", user.GetUserMsg)
			userApi.POST("/message/update", user.UpdateUserMsg)
			userApi.GET("/list", user.ListUser)
			userApi.POST("/role/update", user.UpdateRole)
			userApi.GET("/role/list", user.ListRole)
			userApi.POST("/role/user/update", user.UpdateRoleUser)
		}
		port := m.Config().Http.Port
		r.Run(fmt.Sprintf(":%s", port))
	}
}

func InitClusterApi()  {
	if !m.Config().Peer.Enable {
		return
	}
	http.Handle("/sync/config", http.HandlerFunc(alarm.SyncConfigHandle))
	http.Handle("/sync/sd", http.HandlerFunc(alarm.SyncSdFileHandle))
	http.ListenAndServe(fmt.Sprintf(":%s", m.Config().Peer.HttpPort), nil)
}

func InitDependenceParam()  {
	agent.InitAgentManager()
}

func httpLogHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ignoreLog := false
		for _,v := range m.LogIgnorePath {
			if strings.Contains(c.Request.RequestURI, v) {
				ignoreLog = true
				break
			}
		}
		if ignoreLog {
			c.Next()
		}else {
			start := time.Now()
			var bodyBytes []byte
			if c.Request.Method == http.MethodPost {
				ignore := false
				for _, v := range m.LogParamIgnorePath {
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
			log.AccessLogger.Info("request", log.String("url", c.Request.RequestURI), log.String("method", c.Request.Method), log.Int("code", c.Writer.Status()), log.String("operator", c.GetString("operatorName")), log.String("ip", c.ClientIP()), log.Float64("cost_second", time.Now().Sub(start).Seconds()), log.String("body", string(bodyBytes)))
		}
	}
}