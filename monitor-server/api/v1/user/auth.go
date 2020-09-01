package user

import (
	"github.com/gin-gonic/gin"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"time"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"encoding/base64"
	"strconv"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
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
			mid.ReturnQueryTableError(c, err.Error(), err)
			return
		}
		if user.Id == 0 {
			mid.ReturnFetchDataError(c, "user", "name", authData.Username)
			return
		}
		authPassword,err := base64.StdEncoding.DecodeString(authData.Password)
		if err != nil {
			mid.ReturnValidateError(c, "password is not base64 encode")
			return
		}
		savePassword,_ := mid.Dncrypt(user.Passwd)
		if string(authPassword) == savePassword {
			session := m.Session{User:authData.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.ReturnHandleError(c, "save session failed", nil)
			}else{
				session.Token = sId
				mid.ReturnSuccessData(c, session)
			}
		}else{
			mid.ReturnPasswordError(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
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
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnError(c, http.StatusUnauthorized, "Invalid session token", nil)
	}
}

func Register(c *gin.Context)  {
	var param auth
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Password != param.RePassword {
			mid.ReturnValidateError(c, "password and re_password is different")
			return
		}
		tmpPassword,err := base64.StdEncoding.DecodeString(param.Password)
		if err != nil {
			mid.ReturnValidateError(c, "password is not base64 encode")
			return
		}
		newPassword,err := mid.Encrypt(tmpPassword)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		err = db.AddUser(m.UserTable{Name:param.Username, Passwd:string(newPassword), DisplayName:param.DisplayName, Email:param.Email, Phone:param.Phone}, "")
		if err != nil {
			mid.ReturnUpdateTableError(c, "user", err)
		}else{
			session := m.Session{User:param.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.ReturnSuccessWithMessage(c, "Register success,but login with save session failed,please login")
			}else{
				session.Token = sId
				mid.ReturnSuccessData(c, session)
			}
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateUserMsg(c *gin.Context)  {
	var param m.UpdateUserDto
	if err := c.ShouldBindJSON(&param); err==nil {
		if mid.GetOperateUser(c) == "admin" {
			mid.ReturnHandleError(c, "admin message can not change", nil)
			return
		}
		if !mid.ValidatePost(c, param, "NewPassword", "ReNewPassword") {
			return
		}
		operator := mid.GetOperateUser(c)
		var userObj m.UserTable
		userObj.Name = operator
		if param.NewPassword != "" && param.ReNewPassword != "" {
			if param.NewPassword != param.ReNewPassword {
				mid.ReturnValidateError(c, "password and re_password is different")
				return
			}
			tmpPassword,err := base64.StdEncoding.DecodeString(param.NewPassword)
			if err != nil {
				mid.ReturnValidateError(c, "password is not base64 encode")
				return
			}
			newPassword,err := mid.Encrypt(tmpPassword)
			userObj.Passwd = newPassword
		}else{
			userObj.Phone = param.Phone
			userObj.Email = param.Email
			userObj.DisplayName = param.DisplayName
		}
		err = db.UpdateUser(userObj)
		if err != nil {
			mid.ReturnUpdateTableError(c, "user", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetUserMsg(c *gin.Context)  {
	operator := mid.GetOperateUser(c)
	err,userObj := db.GetUser(operator)
	if err != nil {
		mid.ReturnQueryTableError(c, err.Error(), err)
		return
	}
	userObj.Passwd = ""
	mid.ReturnSuccessData(c, userObj)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.Config().Http.Session.Enable {
			//auToken := c.GetHeader("Authorization")
			//if auToken != "" {
			//	_,err := mid.DecodeCoreToken(auToken, m.CoreJwtKey)
			//	if err != nil {
			//		mid.ReturnTokenError(c)
			//		c.Abort()
			//	}else{
			//		c.Next()
			//	}
			//}else{
			//	mid.ReturnTokenError(c)
			//	c.Abort()
			//}
			c.Next()
			return
		}else {
			auToken := c.GetHeader("X-Auth-Token")
			if auToken != "" {
				if mid.IsActive(auToken) {
					c.Next()
				} else {
					mid.ReturnTokenError(c)
					c.Abort()
				}
			} else {
				mid.ReturnTokenError(c)
				c.Abort()
			}
		}
	}
}

func AuthServerRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.Config().Http.Session.ServerEnable {
			if c.GetHeader("X-Auth-Token") == m.Config().Http.Session.ServerToken {
				c.Next()
			}else{
				mid.ReturnTokenError(c)
				c.Abort()
			}
		}else{
			mid.ReturnTokenError(c)
			c.Abort()
		}
	}
}

func LdapLogin(c *gin.Context) {
	var authData auth
	if err := c.ShouldBindJSON(&authData); err==nil {
		if !mid.ValidatePost(c, authData, "Password") {
			return
	    }
		if ldapAuth(authData.Username, authData.Password) {
			session := m.Session{User:authData.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.ReturnHandleError(c, "save session failed", nil)
			}else{
				session.Token = sId
				mid.ReturnSuccessData(c, session)
			}
		}else{
			mid.ReturnValidateError(c, "ldap auth fail")
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func HealthCheck(c *gin.Context)  {
	ip := c.ClientIP()
	date := time.Now().Format(m.DatetimeFormat)
	log.Logger.Info("Health check", log.String("requestIp", ip), log.String("date", date))
	c.JSON(http.StatusOK, gin.H{"status": "ok", "request_ip": ip, "date": date})
}

func ListUser(c *gin.Context)  {
	search := c.Query("search")
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	role,_ := strconv.Atoi(c.Query("role"))
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	err,data := db.ListUser(search, role, page, size)
	if err != nil {
		mid.ReturnQueryTableError(c, "user", err)
	}else{
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateRole(c *gin.Context)  {
	var param m.UpdateRoleDto
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.Operation != "add" && param.Operation != "update" && param.Operation != "delete" {
			mid.ReturnValidateError(c, "operation should be add,update,delete")
			return
		}
		param.Operator = mid.GetOperateUser(c)
		err = db.UpdateRole(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "role", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func ListRole(c *gin.Context)  {
	search := c.Query("search")
	page,_ := strconv.Atoi(c.Query("page"))
	size,_ := strconv.Atoi(c.Query("size"))
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	db.SyncCoreRole()
	err,data := db.ListRole(search, page, size)
	if err != nil {
		mid.ReturnQueryTableError(c, "role", err)
	}else{
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateRoleUser(c *gin.Context)  {
	var param m.UpdateRoleUserDto
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateRoleUser(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "rel_role_user", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}