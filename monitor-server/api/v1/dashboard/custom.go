package dashboard

import (
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func ListCustomDashboard(c *gin.Context) {
	err, result := db.ListCustomDashboard(mid.GetOperateUser(c), mid.GetCoreToken(c))
	if err != nil {
		mid.ReturnQueryTableError(c, "custom_dashboard", err)
		return
	}
	mid.ReturnSuccessData(c, result)
}

func GetCustomDashboard(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	result, getErr := db.GetCustomDashboard(id)
	if getErr != nil {
		mid.ReturnFetchDataError(c, "custom_dashboard", "id", strconv.Itoa(id))
		return
	}
	mid.ReturnSuccessData(c, result)
}

func SaveCustomDashboard(c *gin.Context) {
	var param m.CustomDashboardObj
	if err := c.ShouldBindJSON(&param); err == nil {
		param.UpdateUser = mid.GetOperateUser(c)
		param.PanelGroups = strings.Join(param.PanelGroupList, ",")
		err = db.SaveCustomDashboard(&param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "custom_dashboard", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func DeleteCustomDashboard(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	query := m.CustomDashboardTable{Id: id}
	err = db.DeleteCustomDashboard(&query)
	if err != nil {
		mid.ReturnDeleteTableError(c, "custom_dashboard", "id", strconv.Itoa(id))
		return
	}
	mid.ReturnSuccess(c)
}

func GetCustomDashboardRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("dashboard_id"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "dashboard_id", "int")
		return
	}
	err, result := db.GetCustomDashboardRole(id)
	if err != nil {
		mid.ReturnQueryTableError(c, "custom_dashboard_role_rel", err)
	} else {
		mid.ReturnSuccessData(c, result)
	}
}

func SaveCustomDashboardRole(c *gin.Context) {
	var param m.CustomDashboardRoleDto
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.SaveCustomeDashboardRole(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "custom_dashboard_role_rel", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}
