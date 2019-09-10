package user

import (
	"github.com/gin-gonic/gin"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"net/http"
	"time"
	"strings"
	"fmt"
)

type auth struct {
	Username  string  `form:"username" json:"username" binding:"required"`
	Password  string  `form:"password" json:"password" binding:"required"`
}

// @Summary 登录
// @Produce  json
// @Param username query string true "username"
// @Param password query string true "password"
// @Success 200 {string} json "{"message": "login success"}"
// @Router /login [post]
func Login(c *gin.Context)  {
	var authData auth
	if err := c.ShouldBindJSON(&authData); err==nil {
		if !mid.ValidatePost(c, authData, "Password") {return}
		if authData.Username=="test" && authData.Password=="123" {
			session := m.Session{User:authData.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.Return(c, mid.RespJson{Msg:"fail to save session"})
			}else{
				session.Token = sId
				mid.Return(c, mid.RespJson{Msg:"login success", Data:session})
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"auth fail", Code:http.StatusBadRequest})
		}
	}else{
		c.JSON(http.StatusBadRequest, gin.H{"error": "params validate fail"})
	}
}

// @Summary 登出
// @Produce  json
// @Success 200 {string} json "{"message": "successfully logout"}"
// @Router /login [get]
func Logout(c *gin.Context) {
	auToken := c.GetHeader("X-Auth-Token")
	if auToken!= ""{
		mid.DelSession(auToken)
		mid.Return(c, mid.RespJson{Msg:"successfully logout"})
	}else{
		mid.Return(c, mid.RespJson{Msg:"invalid session token", Code:http.StatusUnauthorized})
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		return
		if strings.Contains(c.Request.URL.Path, "alarm/list") {
			c.Next()
			return
		}
		auToken := c.GetHeader("X-Auth-Token")
		if auToken != ""{
			if mid.IsActive(auToken) {
				c.Next()
			}else{
				mid.Return(c, mid.RespJson{Msg:"invalid session token", Code:http.StatusUnauthorized})
				c.Abort()
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"no auth token", Code:http.StatusUnauthorized})
			c.Abort()
		}
	}
}

func AuthServer() gin.HandlerFunc  {
	return func(c *gin.Context) {
		auToken := c.GetHeader("X-Auth-Token")
		if auToken == m.ServerToken {
			realIp := c.ClientIP()
			pass := false
			for _,ip := range m.Config().LimitIp {
				if ip == realIp || ip == "*" {
					pass = true
					break
				}
			}
			if pass {
				mid.LogInfo(fmt.Sprintf("server %s request %s ", realIp, c.Request.URL.Path))
				c.Next()
			}else{
				mid.Return(c, mid.RespJson{Msg:"ip not allow", Code:http.StatusUnauthorized})
				c.Abort()
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"token not allow", Code:http.StatusUnauthorized})
			c.Abort()
		}
	}
}

func LdapLogin(c *gin.Context) {
	var authData auth
	if err := c.ShouldBindJSON(&authData); err==nil {
		if !mid.ValidatePost(c, authData, "Password") {return}
		if ldapAuth(authData.Username, authData.Password) {
			session := m.Session{User:authData.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.Return(c, mid.RespJson{Msg:"fail to save session"})
			}else{
				session.Token = sId
				mid.Return(c, mid.RespJson{Msg:"login success", Data:session})
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"auth fail", Code:http.StatusUnauthorized})
		}
	}else{
		mid.Return(c, mid.RespJson{Msg:"params validate fail", Code:http.StatusUnauthorized})
	}
}

func UserMsg(c *gin.Context) {
	auToken := c.GetHeader("X-Auth-Token")
	if auToken!= "" {
		re := mid.GetSessionData(auToken)
		mid.Return(c, mid.RespJson{Data:map[string]interface{}{"user":re.User}})
	}else{
		mid.Return(c, mid.RespJson{Msg:"invalid session token", Code:http.StatusBadRequest})
	}
}

func HealthCheck(c *gin.Context)  {
	ip := c.ClientIP()
	date := time.Now().Format("2006-01-02 15:04:05")
	mid.LogInfo(fmt.Sprintf("healthcheck request ip : %s , date : %s", ip, date))
	c.JSON(http.StatusOK, gin.H{"status": "ok", "request_ip": ip, "date": date})
}
