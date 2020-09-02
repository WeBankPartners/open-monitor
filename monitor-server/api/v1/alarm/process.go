package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func GetEndpointProcessConfig(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	err,data := db.GetProcessList(endpointId)
	if err != nil {
		mid.ReturnFetchDataError(c, "process_monitor", "endpoint_id", strconv.Itoa(endpointId))
	}else{
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateEndpointProcessConfig(c *gin.Context)  {
	var param m.ProcessUpdateDto
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateProcess(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "process_monitor", err)
		}else{
			err = db.UpdateNodeExporterProcessConfig(param.EndpointId)
			if err != nil {
				mid.ReturnHandleError(c, "update node exporter config fail ", err)
				return
			}
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}
