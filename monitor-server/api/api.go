package api

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/gin-contrib/cors"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/user"
	"net/http"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/dashboard"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/agent"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
	_ "github.com/WeBankPartners/open-monitor/monitor-server/docs"
)

func InitHttpServer(exportAgent bool) {
	urlPrefix := "/wecube-monitor"
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
	if m.Config().Http.Cross {
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
	// public api
	r.GET(fmt.Sprintf("%s/", urlPrefix), func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.Use(mid.ValidateGet)
	if m.Config().Http.Ldap.Enable {
		r.POST(fmt.Sprintf("%s/login", urlPrefix), user.LdapLogin)
	}else{
		r.POST(fmt.Sprintf("%s/login", urlPrefix), user.Login)
	}
	r.POST(fmt.Sprintf("%s/register", urlPrefix), user.Register)
	r.GET(fmt.Sprintf("%s/logout", urlPrefix), user.Logout)
	r.GET(fmt.Sprintf("%s/check", urlPrefix), user.HealthCheck)
	if m.Config().Http.Swagger {
		r.GET(fmt.Sprintf("%s/swagger/*any", urlPrefix), ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	entityApi := r.Group(fmt.Sprintf("%s/entities", urlPrefix), user.AuthRequired())
	{
		entityApi.GET("/alarm", alarm.GetEntityAlarm)
		entityApi.GET("/test", alarm.TestNotifyAlarm)
		entityApi.POST("/alarm/query", alarm.QueryEntityAlarm)
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
			dashboardApi.POST("/config/metric/update", dashboard.UpdatePromMetric)
			dashboardApi.GET("/config/metric/reload", agent.ReloadEndpointMetric)
			dashboardApi.GET("/endpoint/metric/list", dashboard.GetEndpointMetric)
			dashboardApi.GET("/custom/list", dashboard.ListCustomDashboard)
			dashboardApi.GET("/custom/get", dashboard.GetCustomDashboard)
			dashboardApi.POST("/custom/save", dashboard.SaveCustomDashboard)
			dashboardApi.GET("/custom/delete", dashboard.DeleteCustomDashboard)
			dashboardApi.GET("/server/chart", dashboard.GetChartsByEndpoint)
			dashboardApi.GET("/custom/main/get", dashboard.GetMainPage)
			dashboardApi.GET("/custom/main/set", dashboard.SetMainPage)
			dashboardApi.GET("/custom/endpoint/get", dashboard.GetEndpointsByIp)
			dashboardApi.POST("/system/add", agent.ExportPanelAdd)
			dashboardApi.POST("/system/delete", agent.ExportPanelDelete)
			dashboardApi.GET("/recursive/get", agent.GetPanelRecursive)
			dashboardApi.POST("/config/chart/title", dashboard.UpdateChartsTitle)
		}
		agentApi := authApi.Group("/agent")
		{
			agentApi.POST("/register", agent.RegisterAgent)
			agentApi.POST("/register_new", agent.RegisterAgentNew)
			agentApi.GET("/deregister", agent.DeregisterAgent)
			agentApi.POST("/export/register/:name", agent.ExportAgent)
			agentApi.POST("/export/deregister/:name", agent.ExportAgent)
			agentApi.POST("/export/start/:name", agent.AlarmControl)
			agentApi.POST("/export/stop/:name", agent.AlarmControl)
			agentApi.POST("/export/install/:name", agent.InstallAgent)
			agentApi.POST("/export/custom/endpoint/add", agent.CustomRegister)
			agentApi.POST("/export/custom/metric/add", agent.CustomMetricPush)
			agentApi.POST("/export/endpoint/telnet/update", agent.UpdateEndpointTelnet)
			agentApi.GET("/export/ping/source", agent.ExportPingSource)
			agentApi.GET("/export/endpoint/telnet/get", agent.GetEndpointTelnet)
		}
		alarmApi := authApi.Group("/alarm")
		{
			alarmApi.GET("/grp/list", alarm.ListGrp)
			alarmApi.POST("/grp/add", alarm.AddGrp)
			alarmApi.POST("/grp/update", alarm.UpdateGrp)
			alarmApi.GET("/grp/delete", alarm.DeleteGrp)
			alarmApi.POST("/grp/role/update", alarm.UpdateGrpRole)
			alarmApi.GET("/grp/role/get", alarm.GetGrpRole)
			alarmApi.GET("/endpoint/list", alarm.ListEndpoint)
			alarmApi.POST("/endpoint/update", alarm.EditGrpEndpoint)
			alarmApi.GET("/strategy/search", alarm.SearchObjOption)
			alarmApi.GET("/strategy/list", alarm.ListTpl)
			alarmApi.POST("/strategy/add", alarm.AddStrategy)
			alarmApi.POST("/strategy/update", alarm.EditStrategy)
			alarmApi.GET("/strategy/delete", alarm.DeleteStrategy)
			alarmApi.POST("/webhook", alarm.AcceptAlertMsg)
			alarmApi.GET("/history", alarm.GetHistoryAlarm)
			alarmApi.GET("/problem/list", alarm.GetProblemAlarm)
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
			alarmApi.POST("/business/update", alarm.UpdateEndpointBusinessConfig)
			alarmApi.GET("/grp/export", alarm.ExportGrpStrategy)
			alarmApi.POST("/grp/import", alarm.ImportGrpStrategy)
			alarmApi.POST("/send", alarm.OpenAlarmApi)
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
	if !m.Config().Cluster.Enable {
		return
	}
	http.Handle("/sync/config", http.HandlerFunc(alarm.SyncConfigHandle))
	http.Handle("/sync/consul", http.HandlerFunc(alarm.SyncConsulHandle))
	http.ListenAndServe(fmt.Sprintf(":%s", m.Config().Cluster.HttpPort), nil)
}

func InitDependenceParam()  {
	agent.InitAgentManager()
}