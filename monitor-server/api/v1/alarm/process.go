package alarm

import (
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetEndpointProcessConfig(c *gin.Context) {
	endpointId, err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	err, data := db.GetProcessList(endpointId)
	if err != nil {
		mid.ReturnFetchDataError(c, "process_monitor", "endpoint_id", strconv.Itoa(endpointId))
	} else {
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateEndpointProcessConfig(c *gin.Context) {
	var param m.ProcessUpdateDto
	if err := c.ShouldBindJSON(&param); err == nil {
		//if param.Check && len(param.ProcessList) > 0 {
		//	err, illegal, msg := db.CheckNodeExporterProcessConfig(param.EndpointId, param.ProcessList)
		//	if err != nil {
		//		mid.ReturnHandleError(c, "check node exporter config fail ", err)
		//		return
		//	}
		//	if illegal {
		//		c.JSON(http.StatusBadRequest, mid.RespJson{Status: "WARN", Code: 300, Message: msg})
		//		return
		//	}
		//}
		var processDtoNew m.ProcessUpdateDtoNew
		processDtoNew.EndpointId = param.EndpointId
		for _, v := range param.ProcessList {
			processDtoNew.ProcessList = append(processDtoNew.ProcessList, m.ProcessMonitorTable{ProcessName: v.ProcessName, Tags: v.Tags, DisplayName: v.DisplayName})
		}
		err = db.UpdateProcess(processDtoNew, "update")
		if err != nil {
			mid.ReturnUpdateTableError(c, "process_monitor", err)
		} else {
			err = db.UpdateNodeExporterProcessConfig(param.EndpointId)
			if err != nil {
				mid.ReturnHandleError(c, "update node exporter config fail ", err)
				return
			}
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateEndpointProcessConfigNew(c *gin.Context) {
	var param m.ProcessUpdateDtoNew
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.Check && len(param.ProcessList) > 0 {
			var processList []m.ProcessMonitorTable
			for _, v := range param.ProcessList {
				processList = append(processList, m.ProcessMonitorTable{ProcessName: v.ProcessName, DisplayName: v.DisplayName, Tags: v.Tags})
			}
			err, illegal, msg := db.CheckNodeExporterProcessConfig(param.EndpointId, processList)
			if err != nil {
				mid.ReturnHandleError(c, "check node exporter config fail ", err)
				return
			}
			if illegal {
				mid.ReturnValidateError(c, msg)
				return
			}
		}
		err = db.UpdateProcess(param, "update")
		if err != nil {
			mid.ReturnUpdateTableError(c, "process_monitor", err)
		} else {
			err = db.UpdateNodeExporterProcessConfig(param.EndpointId)
			if err != nil {
				mid.ReturnHandleError(c, "update node exporter config fail ", err)
				return
			}
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}
