package agent

import (
	"github.com/gin-gonic/gin"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"net/http"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"encoding/json"
)

type requestPanelObj struct {
	RequestId  string  	`json:"requestId"`
	Inputs  []panelRequestObj  `json:"inputs"`
}

type panelRequestObj struct {
	CallbackParameter  string  `json:"callbackParameter"`
	Guid  string  `json:"guid"`
	DisplayName  string  `json:"display_name"`
	Parent  []string  `json:"parent"`
	Endpoint  string  `json:"endpoint"`
}

func ExportPanel(c *gin.Context)  {
	var param requestPanelObj
	var result resultObj
	if err := c.ShouldBindJSON(&param); err==nil {
		if len(param.Inputs) == 0 {
			result = resultObj{ResultCode:"0", ResultMessage:"inputs length is zero,do nothing"}
			mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
			c.JSON(http.StatusOK, result)
			return
		}
		var tmpResult []resultOutputObj
		successFlag := "0"
		errorMessage := "Done"
		for _,v := range param.Inputs {
			var tmpMessage string
			if v.Guid == "" {
				tmpMessage = fmt.Sprintf("Index:%s guid is null", v.CallbackParameter)
			}
			if len(v.Parent) == 0 && v.Endpoint == "" {
				tmpMessage = fmt.Sprintf("Index:%s children and endpoint both null", v.CallbackParameter)
			}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
				continue
			}
			err := db.UpdateRecursivePanel(m.PanelRecursiveTable{Guid:v.Guid,DisplayName:v.DisplayName,Parent:strings.Join(v.Parent, "^"),Endpoint:v.Endpoint})
			if err != nil {
				tmpMessage = fmt.Sprintf("Index:%s update database error:%v", v.CallbackParameter, err)
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
			}else{
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
			}
		}
		result = resultObj{ResultCode: successFlag, ResultMessage: errorMessage, Results: resultOutput{Outputs: tmpResult}}
		resultString,_ := json.Marshal(result)
		mid.LogInfo(string(resultString))
		mid.ReturnData(c, result)
	}else{
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("Param validate fail : %v", err)}
		mid.LogInfo(fmt.Sprintf("result : code %s , message %s", result.ResultCode, result.ResultMessage))
		c.JSON(http.StatusBadRequest, result)
	}
}
