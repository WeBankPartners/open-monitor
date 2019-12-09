package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"strconv"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
)

func ListCustomDashboard(c *gin.Context)  {
	err,result := db.ListCustomDashboard()
	if err != nil {
		mid.ReturnError(c, "List customized dashboard failed", err)
		return
	}
	mid.ReturnData(c, result)
}

func GetCustomDashboard(c *gin.Context)  {
	id,err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnValidateFail(c, "Parameter \"id\" validation failed")
		return
	}
	query := m.CustomDashboardTable{Id:id}
	err = db.GetCustomDashboard(&query)
	if err != nil {
		mid.ReturnError(c, "Get customized dashboard failed", err)
		return
	}
	mid.ReturnData(c, query)
}

func SaveCustomDashboard(c *gin.Context)  {
	var param m.CustomDashboardTable
	if err := c.ShouldBindJSON(&param);err==nil {
		err = db.SaveCustomDashboard(&param)
		if err != nil {
			mid.ReturnError(c, "Save customized dashboard failed", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func DeleteCustomDashboard(c *gin.Context)  {
	id,err := strconv.Atoi(c.Query("id"))
	if err != nil || id <= 0 {
		mid.ReturnValidateFail(c, "Parameter \"id\" validation failed")
		return
	}
	query := m.CustomDashboardTable{Id:id}
	err = db.DeleteCustomDashboard(&query)
	if err != nil {
		mid.ReturnError(c, "Delete customized dashboard failed", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}
