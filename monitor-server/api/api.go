package api

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"github.com/gin-contrib/cors"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/api/v1/user"
	"net/http"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"fmt"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/api/v1/dashboard"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/api/v1/agent"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/api/v1/alarm"
	_ "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/docs"
)

func InitHttpServer() {
	urlPrefix := "/wecube-monitor"
	r := gin.Default()
	r.LoadHTMLGlob("public/*.html")
	r.Static(fmt.Sprintf("%s/js", urlPrefix), fmt.Sprintf("public%s/js", urlPrefix))
	r.Static(fmt.Sprintf("%s/css", urlPrefix), fmt.Sprintf("public%s/css", urlPrefix))
	r.Static(fmt.Sprintf("%s/img", urlPrefix), fmt.Sprintf("public%s/img", urlPrefix))
	r.Static(fmt.Sprintf("%s/fonts", urlPrefix), fmt.Sprintf("public%s/fonts", urlPrefix))
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
	r.GET(fmt.Sprintf("%s/logout", urlPrefix), user.Logout)
	r.GET(fmt.Sprintf("%s/check", urlPrefix), user.HealthCheck)
	if m.Config().Http.Swagger {
		r.GET(fmt.Sprintf("%s/swagger/*any", urlPrefix), ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	// auth api
	authApi := r.Group(fmt.Sprintf("%s/api/v1", urlPrefix), user.AuthRequired())
	{
		dashboardApi := authApi.Group("/dashboard")
		{
			dashboardApi.GET("/main", dashboard.MainDashboard)
			dashboardApi.GET("/panels", dashboard.GetPanels)
			dashboardApi.GET("/chart", dashboard.GetChart)
			dashboardApi.GET("/tags", dashboard.GetTags)
			dashboardApi.GET("/search", dashboard.MainSearch)
			dashboardApi.GET("/config/metric/list", dashboard.GetPromMetric)
			dashboardApi.GET("/newchart", dashboard.GetChartNew)
			dashboardApi.POST("/config/metric/update", dashboard.UpdatePromMetric)
			dashboardApi.GET("/endpoint/metric/list", dashboard.GetEndpointMetric)
			dashboardApi.GET("/custom/list", dashboard.ListCustomDashboard)
			dashboardApi.GET("/custom/get", dashboard.GetCustomDashboard)
			dashboardApi.POST("/custom/save", dashboard.SaveCustomDashboard)
			dashboardApi.GET("/custom/delete", dashboard.DeleteCustomDashboard)
			dashboardApi.GET("/server/chart", dashboard.GetChartsByEndpoint)
			dashboardApi.GET("/custom/main/get", dashboard.GetMainPage)
			dashboardApi.GET("/custom/main/set", dashboard.SetMainPage)
		}
		agentApi := authApi.Group("/agent")
		{
			agentApi.POST("/register", agent.RegisterAgent)
			agentApi.GET("/deregister", agent.DeregisterAgent)
			agentApi.GET("/host/register", agent.StartHostAgent)
			agentApi.GET("/host/deregister", agent.StopHostAgent)
			agentApi.GET("/mysql/register", agent.StartMysqlAgent)
			agentApi.GET("/mysql/deregister", agent.StopMysqlAgent)
			agentApi.GET("/redis/register", agent.StartRedisAgent)
			agentApi.GET("/redis/deregister", agent.StopRedisAgent)
			agentApi.GET("/tomcat/register", agent.StartTomcatAgent)
			agentApi.GET("/tomcat/deregister", agent.StopTomcatAgent)
		}
		alarmApi := authApi.Group("/alarm")
		{
			alarmApi.GET("/grp/list", alarm.ListGrp)
			alarmApi.POST("/grp/add", alarm.AddGrp)
			alarmApi.POST("/grp/update", alarm.UpdateGrp)
			alarmApi.GET("/grp/delete", alarm.DeleteGrp)
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
		}
		port := m.Config().Http.Port
		r.Run(fmt.Sprintf(":%s", port))
	}
}