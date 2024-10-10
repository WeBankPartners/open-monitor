package service

import (
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

func PluginUpdateServicePath(c *gin.Context) {
	response := models.PluginUpdateServicePathResp{ResultCode: "0", ResultMessage: "success", Results: models.PluginUpdateServicePathOutput{}}
	var err error
	defer func() {
		if err != nil {
			log.Logger.Error("Update service path handle fail", log.Error(err))
			response.ResultCode = "1"
			response.ResultMessage = err.Error()
		}
		bodyBytes, _ := json.Marshal(response)
		c.Set("responseBody", string(bodyBytes))
		c.JSON(http.StatusOK, response)
	}()
	var param models.PluginUpdateServicePathRequest
	c.ShouldBindJSON(&param)
	if len(param.Inputs) == 0 {
		return
	}
	for _, input := range param.Inputs {
		output, tmpErr := db.PluginUpdateServicePathAction(input, middleware.GetOperateUser(c), middleware.GetOperateUserRoles(c), middleware.GetMessageMap(c))
		if tmpErr != nil {
			output.ErrorCode = "1"
			output.ErrorMessage = tmpErr.Error()
			err = tmpErr
		}
		response.Results.Outputs = append(response.Results.Outputs, output)
	}
}
