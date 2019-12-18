package user

import (
	"github.com/gin-gonic/gin"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"time"
	"strings"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"encoding/base64"
)

type auth struct {
	Username  string  `form:"username" json:"username" binding:"required"`
	Password  string  `form:"password" json:"password" binding:"required"`
	RePassword  string  `form:"re_password" json:"re_password"`
	DisplayName  string  `form:"display_name" json:"display_name"`
	Email  string  `form:"email" json:"email"`
	Phone  string  `form:"phone" json:"phone"`
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
		if !mid.ValidatePost(c, authData, "Password", "RePassword") {return}
		err,user := db.GetUser(authData.Username)
		if err != nil {
			mid.ReturnError(c ,"Query db fail ", err)
			return
		}
		if user.Id == 0 {
			mid.ReturnValidateFail(c, fmt.Sprintf("Username not exist"))
			return
		}
		authPassword,err := base64.StdEncoding.DecodeString(authData.Password)
		if err != nil {
			mid.ReturnValidateFail(c, "Password is not base64 encode")
			return
		}
		savePassword,_ := mid.Dncrypt(user.Passwd)
		if string(authPassword) == savePassword {
			session := m.Session{User:authData.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.Return(c, mid.RespJson{Msg:"Save session failed"})
			}else{
				session.Token = sId
				mid.Return(c, mid.RespJson{Msg:"Login successfully", Data:session})
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"Authorization failed", Code:http.StatusBadRequest})
		}
	}else{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter validation failed"})
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
		mid.Return(c, mid.RespJson{Msg:"Logout successfully"})
	}else{
		mid.Return(c, mid.RespJson{Msg:"Invalid session token", Code:http.StatusUnauthorized})
	}
}

func Register(c *gin.Context)  {
	var param auth
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Password != param.RePassword {
			mid.ReturnValidateFail(c, "Password and RePassword is different")
			return
		}
		tmpPassword,err := base64.StdEncoding.DecodeString(param.Password)
		if err != nil {
			mid.ReturnValidateFail(c, "Password is not base64 encode")
			return
		}
		newPassword,err := mid.Encrypt(tmpPassword)
		if err != nil {
			mid.ReturnError(c, "Register user fail", err)
			return
		}
		err = db.AddUser(m.UserTable{Name:param.Username, Passwd:string(newPassword), DisplayName:param.DisplayName, Email:param.Email, Phone:param.Phone}, "")
		if err != nil {
			mid.ReturnError(c, "Register user fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validate failed %v", err))
	}
}

func UpdateUserMsg(c *gin.Context)  {
	var param m.UpdateUserDto
	if err := c.ShouldBindJSON(&param); err==nil {
		if !mid.ValidatePost(c, param, "NewPassword", "ReNewPassword") {return}
		operator := mid.GetOperateUser(c)
		var userObj m.UserTable
		userObj.Name = operator
		if param.NewPassword != "" && param.ReNewPassword != "" {
			if param.NewPassword != param.ReNewPassword {
				mid.ReturnValidateFail(c, "Password and RePassword is different")
				return
			}
			tmpPassword,err := base64.StdEncoding.DecodeString(param.NewPassword)
			if err != nil {
				mid.ReturnValidateFail(c, "Password is not base64 encode")
				return
			}
			newPassword,err := mid.Encrypt(tmpPassword)
			userObj.Passwd = newPassword
		}
		if param.Phone != "" {
			userObj.Phone = param.Phone
		}
		if param.Email != "" {
			userObj.Email = param.Email
		}
		if param.DisplayName != "" {
			userObj.DisplayName = param.DisplayName
		}
		err = db.UpdateUser(userObj)
		if err != nil {
			mid.ReturnError(c, "Update user msg fail ", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validate failed %v", err))
	}
}

func GetUserMsg(c *gin.Context)  {
	operator := mid.GetOperateUser(c)
	err,userObj := db.GetUser(operator)
	if err != nil {
		mid.ReturnError(c, "Get user message fail ", err)
		return
	}
	userObj.Passwd = "********"
	mid.ReturnData(c, userObj)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.Config().Http.Session.Enable {
			c.Next()
			return
		}
		if strings.Contains(c.Request.URL.Path, "alarm/problem/list") || strings.Contains(c.Request.URL.Path, "alarm/webhook") {
			c.Next()
			return
		}
		auToken := c.GetHeader("X-Auth-Token")
		if auToken != ""{
			if mid.IsActive(auToken) {
				c.Next()
			}else{
				mid.Return(c, mid.RespJson{Msg:"Invalid session token", Code:http.StatusUnauthorized})
				c.Abort()
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"Token is not authorized", Code:http.StatusUnauthorized})
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
				mid.Return(c, mid.RespJson{Msg:"Save session failed"})
			}else{
				session.Token = sId
				mid.Return(c, mid.RespJson{Msg:"Login successfully", Data:session})
			}
		}else{
			mid.Return(c, mid.RespJson{Msg:"Auth fail", Code:http.StatusUnauthorized})
		}
	}else{
		mid.Return(c, mid.RespJson{Msg:"Params validation fail", Code:http.StatusUnauthorized})
	}
}

func HealthCheck(c *gin.Context)  {
	ip := c.ClientIP()
	date := time.Now().Format(m.DatetimeFormat)
	mid.LogInfo(fmt.Sprintf("healthcheck request ip : %s , date : %s", ip, date))
	c.JSON(http.StatusOK, gin.H{"status": "ok", "request_ip": ip, "date": date})
}
