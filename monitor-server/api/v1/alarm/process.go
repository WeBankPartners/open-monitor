package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"fmt"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
)

func GetEndpointProcessConfig(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnValidateFail(c, fmt.Sprintf("Param id validate fail %v", err))
		return
	}
	err,data := db.GetProcessList(endpointId)
	if err != nil {
		mid.ReturnError(c, "Get process list fail", err)
	}else{
		mid.ReturnData(c, data)
	}
}

func UpdateEndpointProcessConfig(c *gin.Context)  {
	var param m.ProcessUpdateDto
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateProcess(param)
		if err != nil {
			mid.ReturnError(c, "Update process fail ", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Param validate fail %v \n", err))
	}
}


