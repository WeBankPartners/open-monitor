package dashboard_new

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func PanelList(c *gin.Context)  {
	var id int
	if c.Query("id") != "" {
		id,_ = strconv.Atoi(c.Query("id"))
	}
	endpointType := c.Query("endpointType")
	result,err := db.PanelList(id,endpointType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccessData(c, result)
	}
}

func PanelCreate(c *gin.Context)  {
	endpointType := c.Param("endpointType")
	var param []*models.PanelTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.PanelCreate(endpointType, param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func PanelUpdate(c *gin.Context)  {
	var param []*models.PanelTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.PanelUpdate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}

func PanelDelete(c *gin.Context)  {
	ids := c.Query("ids")
	err := db.PanelDelete(strings.Split(ids,","))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	}else{
		middleware.ReturnSuccess(c)
	}
}