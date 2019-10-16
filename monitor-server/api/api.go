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
	r := gin.Default()
	r.LoadHTMLGlob("public/*.html")
	r.Static("/js", "public/js")
	r.Static("/css", "public/css")
	r.Static("/img", "public/img")
	r.Static("/fonts", "public/fonts")
	if m.Config().Http.Cross {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Token", "X-Auth-Token"}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
		corsConfig.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
		corsConfig.AllowCredentials = true
		r.Use(cors.New(corsConfig))
	}
	// public api
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.Use(mid.ValidateGet)
	if m.Config().Http.Ldap.Enable {
		r.POST("/login", user.LdapLogin)
	}else{
		r.POST("/login", user.Login)
	}
	r.GET("/logout", user.Logout)
	r.GET("/check", user.HealthCheck)
	if m.Config().Http.Swagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	// auth api
	authApi := r.Group("/api/v1", user.AuthRequired())
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
		}
		agentApi := authApi.Group("/agent")
		{
			agentApi.POST("/register", agent.RegisterAgent)
			agentApi.GET("/deregister", agent.DeregisterAgent)
			agentApi.GET("/host/start", agent.StartHostAgent)
			agentApi.GET("/host/stop", agent.StopHostAgent)
			agentApi.GET("/mysql/start", agent.StartMysqlAgent)
			agentApi.GET("/mysql/stop", agent.StopMysqlAgent)
			agentApi.GET("/redis/start", agent.StartRedisAgent)
			agentApi.GET("/redis/stop", agent.StopRedisAgent)
			agentApi.GET("/tomcat/start", agent.StartTomcatAgent)
			agentApi.GET("/tomcat/stop", agent.StopTomcatAgent)
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
		r.Run(fmt.Sprintf(":%d", port))
	}
}