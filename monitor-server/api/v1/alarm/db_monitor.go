package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"fmt"
	"strings"
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
		param.Sql = strings.TrimSpace(param.Sql)
		param.Sql = strings.Replace(param.Sql, "\n", " ", -1)
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
		param.Sql = strings.TrimSpace(param.Sql)
		param.Sql = strings.Replace(param.Sql, "\n", " ", -1)
		sql := strings.ToLower(param.Sql)
		if !strings.HasPrefix(sql, "select") {
			mid.ReturnValidateError(c, "SQL must start with select")
			return
		}
		if strings.Contains(sql, ";") || strings.Contains(sql, "insert") || strings.Contains(sql, "update") || strings.Contains(sql, "delete") || strings.Contains(sql, "alter") || strings.Contains(sql, "drop") {
			mid.ReturnValidateError(c, "SQL contains illegal character")
			return
		}
		nameExists := false
		dbMonitorObjs,_ := db.ListDbMonitor(param.EndpointId)
		for _,v := range dbMonitorObjs {
			for _,vv := range v.Data {
				if vv.Name == param.Name && vv.Id != param.Id {
					nameExists = true
				}
			}
		}
		if nameExists {
			mid.ReturnValidateError(c, "Name already used")
			return
		}
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
		param.Sql = strings.TrimSpace(param.Sql)
		param.Sql = strings.Replace(param.Sql, "\n", " ", -1)
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

func UpdateDbMonitorSysName(c *gin.Context)  {
	var param m.DbMonitorSysNameDto
	if err := c.ShouldBindJSON(&param);err == nil {
		err = db.UpdateDbMonitorSysName(param)
		if err != nil {
			mid.ReturnHandleError(c, fmt.Sprintf("Update db_monitor sys_panel fail,%s ", err.Error()), err)
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
