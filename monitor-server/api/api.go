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
			dashboardApi.GET("/register", dashboard.RegisterEndpoint)
		}
		port := m.Config().Http.Port
		r.Run(fmt.Sprintf(":%d", port))
	}
}