package user

import (
	"encoding/base64"
	"errors"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	whitePathMap = map[string]bool{
		"/monitor/entities/${model}/query": true,
	}
	//  需要放行的 ApiCode
	whiteCodeList = []string{
		"trans_export_analyze", "monitor_metric_export", "alarm_endpoint_group_query", "alarm_strategy_workflow", "service_log_metric_export", "service_log_metric_log_monitor_template_export",
		"alarm_strategy_export_by_query_type_and_guid", "service_log_keyword_export", "dashboard_custom_export", "dashboard_custom_get", "chart_custom_permission_batch", "trans_export_log_monitor_template_batch",
		"trans_export_dashboard_batch", "trans_export_service_group_batch", "trans_export_config_type_batch", "config_type_batch_add", "alarm_endpoint_group_import", "monitor_metric_import",
		"alarm_strategy_import_by_query_type_and_guid", "service_log_metric_log_monitor_template_import", "service_log_metric_import", "service_log_keyword_import", "dashboard_custom_trans_import",
		"dashboard_data_sync", "dashboard_system_add", "agent_export_stop", "agent_export_snmp_exporter", "alarm_problem_query", "user_list", "alarm_endpoint_group_notify_list_by_group_guid",
		"alarm_org_plugin", "dashboard_new_metric_get", "user_role_update", "agent_export_process", "alarm_problem_list", "user_message_get", "user_role_user_update", "agent_export_log_monitor",
		"agent_export_start", "agent_export_dregister", "dashboard_config_metric_list", "agent_export_kubernetes_cluster", "agent_export_register", "agent_export_kubernetes_pod",
		"dashboard_server_chart", "user_message_update", "agent_custom_metric_add", "dashboard_system_delete", "service_plugin_update_path", "dashboard_custom_main_get",
	}
	ApiMenuMap = make(map[string][]string)
)

type auth struct {
	Username    string `form:"username" json:"username" binding:"required"`
	Password    string `form:"password" json:"password" binding:"required"`
	RePassword  string `form:"re_password" json:"re_password"`
	DisplayName string `form:"display_name" json:"display_name"`
	Email       string `form:"email" json:"email"`
	Phone       string `form:"phone" json:"phone"`
}

// @Summary 登录
// @Produce  json
// @Param username query string true "username"
// @Param password query string true "password"
// @Success 200 {string} json "{"message": "login success"}"
// @Router /login [post]
func Login(c *gin.Context) {
	var authData auth
	if err := c.ShouldBindJSON(&authData); err == nil {
		if !mid.ValidatePost(c, authData, "Password", "RePassword") {
			return
		}
		err, user := db.GetUser(authData.Username)
		if err != nil {
			mid.ReturnQueryTableError(c, err.Error(), err)
			return
		}
		if user.Id == 0 {
			mid.ReturnFetchDataError(c, "user", "name", authData.Username)
			return
		}
		authPassword, err := base64.StdEncoding.DecodeString(authData.Password)
		if err != nil {
			mid.ReturnValidateError(c, "password is not base64 encode")
			return
		}
		savePassword, _ := mid.Dncrypt(user.Passwd)
		if string(authPassword) == savePassword {
			session := m.Session{User: authData.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.ReturnHandleError(c, "save session failed", nil)
			} else {
				session.Token = sId
				mid.ReturnSuccessData(c, session)
			}
		} else {
			mid.ReturnPasswordError(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

// Logout @Summary 登出
// @Produce  json
// @Success 200 {string} json "{"message": "successfully logout"}"
// @Router /login [get]
func Logout(c *gin.Context) {
	auToken := c.GetHeader("X-Auth-Token")
	if auToken != "" {
		mid.DelSession(auToken)
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnError(c, errors.New("invalid session token"), http.StatusUnauthorized)
	}
}

func Register(c *gin.Context) {
	var param auth
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.Password != param.RePassword {
			mid.ReturnValidateError(c, "password and re_password is different")
			return
		}
		tmpPassword, err := base64.StdEncoding.DecodeString(param.Password)
		if err != nil {
			mid.ReturnValidateError(c, "password is not base64 encode")
			return
		}
		newPassword, err := mid.Encrypt(tmpPassword)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		err = db.AddUser(m.UserTable{Name: param.Username, Passwd: string(newPassword), DisplayName: param.DisplayName, Email: param.Email, Phone: param.Phone})
		if err != nil {
			mid.ReturnUpdateTableError(c, "user", err)
		} else {
			session := m.Session{User: param.Username}
			isOk, sId := mid.SaveSession(session)
			if !isOk {
				mid.ReturnSuccessWithMessage(c, "Register success,but login with save session failed,please login")
			} else {
				session.Token = sId
				mid.ReturnSuccessData(c, session)
			}
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateUserMsg(c *gin.Context) {
	var param m.UpdateUserDto
	if err := c.ShouldBindJSON(&param); err == nil {
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
			tmpPassword, err := base64.StdEncoding.DecodeString(param.NewPassword)
			if err != nil {
				mid.ReturnValidateError(c, "password is not base64 encode")
				return
			}
			newPassword, err := mid.Encrypt(tmpPassword)
			userObj.Passwd = newPassword
		} else {
			userObj.Phone = param.Phone
			userObj.Email = param.Email
			userObj.DisplayName = param.DisplayName
		}
		err = db.UpdateUser(userObj)
		if err != nil {
			mid.ReturnUpdateTableError(c, "user", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetUserMsg(c *gin.Context) {
	operator := mid.GetOperateUser(c)
	err, userObj := db.GetUser(operator)
	if err != nil {
		mid.ReturnQueryTableError(c, err.Error(), err)
		return
	}
	userObj.Passwd = ""
	mid.ReturnSuccessData(c, userObj)
}

type pluginInterfaceResultObj struct {
	ResultCode    string                      `json:"resultCode"`
	ResultMessage string                      `json:"resultMessage"`
	Results       pluginInterfaceResultOutput `json:"results"`
}

type pluginInterfaceResultOutput struct {
	Outputs []string `json:"outputs"`
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.RequestURI, "/export/ping/source") || strings.Contains(c.Request.RequestURI, "/agent/export/custom") {
			c.Next()
		} else {
			if m.Config().Http.Session.Enable != "true" {
				auToken := c.GetHeader("Authorization")
				if auToken != "" {
					coreToken, err := mid.DecodeCoreToken(auToken, m.CoreJwtKey)
					if err != nil {
						mid.ReturnTokenError(c)
						c.Abort()
					} else {
						c.Set("operatorName", coreToken.User)
						c.Set("operatorRoles", coreToken.Roles)
						// 子系统直接放行
						if strings.Contains(strings.Join(mid.GetOperateUserRoles(c), ","), "SUB_SYSTEM") {
							c.Next()
							return
						}

						// 白名单URL直接放行
						for path, _ := range whitePathMap {
							re := regexp.MustCompile(BuildRegexPattern(path))
							if re.MatchString(c.Request.URL.Path) {
								c.Next()
								return
							}
						}
						// 白名单 code直接放行
						for _, code := range whiteCodeList {
							if code == c.GetString(m.ContextApiCode) {
								c.Next()
								return
							}
						}
						for _, code := range whiteCodeList {
							if code == c.GetString(m.ContextApiCode) {
								c.Next()
								return
							}
						}
						if m.Config().MenuApiMap.Enable == "true" || strings.TrimSpace(m.Config().MenuApiMap.Enable) == "" || strings.ToUpper(m.Config().MenuApiMap.Enable) == "Y" {
							legal := false
							if allowMenuList, ok := ApiMenuMap[c.GetString(m.ContextApiCode)]; ok {
								legal = compareStringList(mid.GetOperateUserRoles(c), allowMenuList)
							} else {
								legal = validateMenuApi(mid.GetOperateUserRoles(c), c.Request.URL.Path, c.Request.Method)
							}
							if legal {
								c.Next()
							} else {
								mid.ReturnApiPermissionError(c)
								c.Abort()
							}
						} else {
							c.Next()
						}
						c.Next()
					}
				} else {
					mid.ReturnTokenError(c)
					c.Abort()
				}
			} else {
				auToken := c.GetHeader("X-Auth-Token")
				if auToken != "" {
					if auToken == m.Config().Http.Session.ServerToken {
						c.Next()
					} else {
						isOk, operator := mid.IsActive(auToken, c.ClientIP())
						if isOk {
							c.Set("operatorName", operator)
							c.Next()
						} else {
							mid.ReturnTokenError(c)
							c.Abort()
						}
					}
				} else {
					mid.ReturnTokenError(c)
					c.Abort()
				}
			}
		}
	}
}

func AuthServerRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.Config().Http.Session.ServerEnable {
			if c.GetHeader("X-Auth-Token") == m.Config().Http.Session.ServerToken {
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

func HealthCheck(c *gin.Context) {
	ip := c.ClientIP()
	date := time.Now().Format(m.DatetimeFormat)
	log.Info(nil, log.LOGGER_APP, "Health check", zap.String("requestIp", ip), zap.String("date", date))
	c.JSON(http.StatusOK, gin.H{"status": "ok", "request_ip": ip, "date": date})
}

func ListUser(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	role, _ := strconv.Atoi(c.Query("role"))
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	err, data := db.ListUser(search, role, page, size)
	if err != nil {
		mid.ReturnQueryTableError(c, "user", err)
	} else {
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateRole(c *gin.Context) {
	var param m.UpdateRoleDto
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.Operation != "add" && param.Operation != "update" && param.Operation != "delete" {
			mid.ReturnValidateError(c, "operation should be add,update,delete")
			return
		}
		param.Operator = mid.GetOperateUser(c)
		err = db.UpdateRoleNew(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "role", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func ListRole(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	db.SyncCoreRole()
	db.SyncCoreRoleList()
	err, data := db.ListRole(search, page, size)
	if err != nil {
		mid.ReturnQueryTableError(c, "role", err)
	} else {
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateRoleUser(c *gin.Context) {
	var param m.UpdateRoleUserDto
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateRoleUser(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "rel_role_user", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func ListManageRole(c *gin.Context) {
	var result []*m.RoleTable
	var err error
	db.SyncCoreRole()
	db.SyncCoreRoleList()
	if result, err = db.ListManageRole(mid.GetOperateUserRoles(c)); err != nil {
		mid.ReturnServerHandleError(c, err)
	}
	mid.ReturnSuccessData(c, result)
}

func validateMenuApi(roles []string, path, method string) (legal bool) {
	// 防止ip 之类数据配置不上
	path = strings.ReplaceAll(path, ".", "")
	path = strings.ReplaceAll(path, "_", "")
	path = strings.ReplaceAll(path, "-", "")
	for _, menuApi := range m.MenuApiGlobalList {
		for _, role := range roles {
			if strings.ToLower(menuApi.Menu) == strings.ToLower(role) {
				for _, item := range menuApi.Urls {
					if strings.TrimSpace(item.Url) == "" {
						continue
					}
					if strings.ToLower(item.Method) == strings.ToLower(method) {
						re := regexp.MustCompile(BuildRegexPattern(item.Url))
						if re.MatchString(path) {
							legal = true
							return
						}
					}
				}
			}
		}
	}
	return
}

func BuildRegexPattern(template string) string {
	// 将 ${variable} 替换为 (\w+) ,并且严格匹配
	return "^" + regexp.MustCompile(`\$\{[\w.-]+\}`).ReplaceAllString(template, `(\w+)`) + "$"
}

func InitApiMenuMap(apiMenuCodeMap map[string]string) {
	var exist bool
	matchUrlMap := make(map[string]int)
	for k, code := range apiMenuCodeMap {
		exist = false
		re := regexp.MustCompile("^" + regexp.MustCompile(":[\\w\\-]+").ReplaceAllString(strings.ToLower(k), "([\\w\\.\\-\\$\\{\\}:\\[\\]]+)") + "$")
		for _, menuApi := range m.MenuApiGlobalList {
			for _, item := range menuApi.Urls {
				key := strings.ToLower(item.Method + "_" + item.Url)
				if re.MatchString(key) {
					exist = true
					if existList, existFlag := ApiMenuMap[code]; existFlag {
						ApiMenuMap[code] = append(existList, menuApi.Menu)
					} else {
						ApiMenuMap[code] = []string{menuApi.Menu}
					}
					matchUrlMap[item.Method+"_"+item.Url] = 1
				}
			}
		}
		if !exist {
			log.Info(nil, log.LOGGER_APP, "InitApiMenuMap menu-api-json lack url", zap.String("path", k), zap.String("code", code))
		}
	}
	for _, menuApi := range m.MenuApiGlobalList {
		for _, item := range menuApi.Urls {
			if _, ok := matchUrlMap[item.Method+"_"+item.Url]; !ok {
				log.Info(nil, log.LOGGER_APP, "InitApiMenuMap can not match menuUrl", zap.String("menu", menuApi.Menu), zap.String("method", item.Method), zap.String("url", item.Url))
			}
		}
	}
	for k, v := range ApiMenuMap {
		if len(v) > 1 {
			ApiMenuMap[k] = DistinctStringList(v, []string{})
		}
	}
	log.Debug(nil, log.LOGGER_APP, "InitApiMenuMap done", log.JsonObj("ApiMenuMap", ApiMenuMap))
}

func DistinctStringList(input, excludeList []string) (output []string) {
	if len(input) == 0 {
		return
	}
	existMap := make(map[string]int)
	for _, v := range excludeList {
		existMap[v] = 1
	}
	for _, v := range input {
		if _, ok := existMap[v]; !ok {
			output = append(output, v)
			existMap[v] = 1
		}
	}
	return
}

func compareStringList(from, target []string) bool {
	match := false
	for _, f := range from {
		for _, t := range target {
			if f == t {
				match = true
				break
			}
		}
		if match {
			break
		}
	}
	return match
}
