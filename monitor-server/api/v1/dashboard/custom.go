package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"strconv"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func ListCustomDashboard(c *gin.Context)  {
	err,result := db.ListCustomDashboard()
	if err != nil {
		mid.ReturnQueryTableError(c, "custom_dashboard", err)
		return
	}
	mid.ReturnSuccessData(c, result)
}

func GetCustomDashboard(c *gin.Context)  {
	id,err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	query := m.CustomDashboardTable{Id:id}
	err = db.GetCustomDashboard(&query)
	if err != nil {
		mid.ReturnFetchDataError(c, "custom_dashboard", "id", strconv.Itoa(id))
		return
	}
	mid.ReturnSuccessData(c, query)
}

func SaveCustomDashboard(c *gin.Context)  {
	var param m.CustomDashboardTable
	if err := c.ShouldBindJSON(&param);err==nil {
		err = db.SaveCustomDashboard(&param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "custom_dashboard", err)
			return
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func DeleteCustomDashboard(c *gin.Context)  {
	id,err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	query := m.CustomDashboardTable{Id:id}
	err = db.DeleteCustomDashboard(&query)
	if err != nil {
		mid.ReturnDeleteTableError(c, "custom_dashboard", "id", strconv.Itoa(id), err)
		return
	}
	mid.ReturnSuccess(c)
}