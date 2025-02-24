package alarm

import (
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func PluginCloseAlarm(c *gin.Context) {
	response := models.PluginCloseAlarmResp{ResultCode: "0", ResultMessage: "success", Results: models.PluginCloseAlarmOutput{}}
	var err error
	defer func() {
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update service path handle fail", zap.Error(err))
			response.ResultCode = "1"
			response.ResultMessage = err.Error()
		}
		bodyBytes, _ := json.Marshal(response)
		c.Set("responseBody", string(bodyBytes))
		c.JSON(http.StatusOK, response)
	}()
	var param models.PluginCloseAlarmRequest
	c.ShouldBindJSON(&param)
	if len(param.Inputs) == 0 {
		return
	}
	for _, input := range param.Inputs {
		output, tmpErr := db.PluginCloseAlarmAction(input)
		if tmpErr != nil {
			output.ErrorCode = "1"
			output.ErrorMessage = tmpErr.Error()
			err = tmpErr
		}
		response.Results.Outputs = append(response.Results.Outputs, output)
	}
}
