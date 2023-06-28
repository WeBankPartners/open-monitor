package api

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/api/v1/transfer"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitHttpServer(port string) {
	urlPrefix := "/monitor-gateway"
	r := gin.Default()
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
	authApi := r.Group(fmt.Sprintf("%s/api/v1", urlPrefix))
	{
		authApi.POST("/push", transfer.AcceptPostData)
		authApi.POST("/push/:first", transfer.AcceptPostData)
		authApi.POST("/push/:first/:second", transfer.AcceptPostData)
		authApi.GET("/register", transfer.AddMember)
	}
	r.GET("/metrics", transfer.DisplayMetrics)
	r.Run(fmt.Sprintf(":%s", port))
}
