package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"fmt"
)

func GetDbMonitorList(c *gin.Context) {
	endpointId,_ := strconv.Atoi(c.Query("endpoint_id"))
	if endpointId <= 0 {
		mid.ReturnParamTypeError(c, "endpoint_id", "int")
		return
	}
	result,err := db.ListDbMonitor(endpointId)
	if err != nil {
		mid.ReturnQueryTableError(c, "db_monitor", err)
	}else{
		mid.ReturnSuccessData(c, result)
	}
}

func AddDbMonitor(c *gin.Context)  {
	var param m.DbMonitorUpdateDto
	if err := c.ShouldBindJSON(&param);err == nil {
		err = db.AddDbMonitor(param)
		if err != nil {
			mid.ReturnHandleError(c, fmt.Sprintf("Add db_monitor table fail,%s ", err.Error()), err)
			return
		}
		err = db.SendConfigToDbManager()
		if err != nil {
			mid.ReturnHandleError(c, "Send config to db_data_exporter fail", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func CheckDbMonitor(c *gin.Context)  {
	var param m.DbMonitorUpdateDto
	if err := c.ShouldBindJSON(&param);err == nil {
		err = db.CheckDbMonitor(param)
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
			return
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateDbMonitor(c *gin.Context)  {
	var param m.DbMonitorUpdateDto
	if err := c.ShouldBindJSON(&param);err == nil {
		if param.Id <= 0 {
			mid.ReturnParamEmptyError(c, "id")
			return
		}
		err = db.UpdateDbMonitor(param)
		if err != nil {
			mid.ReturnHandleError(c, "Update db_monitor table fail", err)
			return
		}
		err = db.SendConfigToDbManager()
		if err != nil {
			mid.ReturnHandleError(c, "Send config to db_data_exporter fail", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func DeleteDbMonitor(c *gin.Context)  {
	var param m.DbMonitorTable
	if err := c.ShouldBindJSON(&param);err == nil {
		if param.Id <= 0 {
			mid.ReturnParamEmptyError(c, "id")
			return
		}
		err = db.DeleteDbMonitor(param.Id)
		if err != nil {
			mid.ReturnHandleError(c, "Delete db_monitor table fail", err)
			return
		}
		err = db.SendConfigToDbManager()
		if err != nil {
			mid.ReturnHandleError(c, "Send config to db_data_exporter fail", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}
