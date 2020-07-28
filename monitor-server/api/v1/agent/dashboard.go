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
	Parent  string  `json:"parent"`
	Endpoint  string  `json:"endpoint"`
	Email  string  `json:"email"`
	Phone  string  `json:"phone"`
	Role  string  `json:"role"`
	FiringCallback  string  `json:"firing_callback"`
	RecoverCallback  string  `json:"recover_callback"`
	Type  string  `json:"type"`
	DeleteAll  string  `json:"delete_all"`
}

func ExportPanelAdd(c *gin.Context)  {
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
			v.Endpoint = trimListString(v.Endpoint)
			v.Parent = trimListString(v.Parent)
			v.Email = trimListString(v.Email)
			v.Phone = trimListString(v.Phone)
			v.Role = trimListString(v.Role)
			tmpEndpoint := strings.Split(v.Endpoint, ",")
			tmpParent := strings.Split(v.Parent, ",")
			tmpRole := db.CheckRoleList(v.Role)
			var tmpMessage string
			if v.Guid == "" {
				tmpMessage = fmt.Sprintf("Index:%s guid is null", v.CallbackParameter)
			}
			//if len(v.Parent) == 0 && v.Endpoint == "" {
			//	tmpMessage = fmt.Sprintf("Index:%s children and endpoint both null", v.CallbackParameter)
			//}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
				continue
			}
			var endpointStringList []string
			for _,vv := range tmpEndpoint {
				if vv == "" {
					continue
				}
				endpointObj := m.EndpointTable{Guid:vv}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id > 0 {
					endpointStringList = append(endpointStringList, endpointObj.Guid)
				}else{
					var tmpAddress string
					tmpList := strings.Split(vv, ":")
					if len(tmpList) == 1 {
						tmpAddress = fmt.Sprintf("%s:9100", tmpList[0])
					}else if len(tmpList) == 2 {
						tmpAddress = fmt.Sprintf("%s:%s", tmpList[0], tmpList[1])
					}else{
						tmpMessage += fmt.Sprintf(" endpoint %s validate fail, ", vv)
						continue
					}
					endpointObj = m.EndpointTable{Address:tmpAddress}
					db.GetEndpoint(&endpointObj)
					if endpointObj.Id > 0 {
						endpointStringList = append(endpointStringList, endpointObj.Guid)
					}else{
						tmpMessage += fmt.Sprintf(" endpoint:%s with address %s can not find, ", vv, tmpAddress)
					}
				}
			}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
				continue
			}
			err := db.UpdateRecursivePanel(m.PanelRecursiveTable{Guid:v.Guid,DisplayName:v.DisplayName,Parent:strings.Join(tmpParent, "^"),Endpoint:strings.Join(endpointStringList, "^"),Email:v.Email,Phone:v.Phone,Role:tmpRole,FiringCallbackKey:v.FiringCallback,RecoverCallbackKey:v.RecoverCallback,ObjType:v.Type})
			if err != nil {
				tmpMessage = fmt.Sprintf("Index:%s update database error:%v", v.CallbackParameter, err)
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
			}else{
				tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
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

func trimListString(input string) string {
	input = strings.Replace(input, "[", "", -1)
	input = strings.Replace(input, "]", "", -1)
	return input
}

func GetPanelRecursive(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "Guid is null")
		return
	}
	err,result := db.GetRecursivePanel(guid)
	if err != nil {
		mid.ReturnError(c, "Get recursive panel error", err)
	}else{
		mid.ReturnData(c, result)
	}
}

func ExportPanelDelete(c *gin.Context)  {
	var param requestPanelObj
	var result resultObj
	if err := c.ShouldBindJSON(&param); err==nil {
		var tmpResult []resultOutputObj
		successFlag := "0"
		errorMessage := "Done"
		for _,v := range param.Inputs {
			var tmpMessage string
			if v.Guid == "" {
				tmpMessage = fmt.Sprintf("Index:%s guid is null", v.CallbackParameter)
			}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
				continue
			}
			var cErr error
			if strings.ToLower(v.DeleteAll) == "y" || strings.ToLower(v.DeleteAll) == "yes" {
				cErr = db.DeleteRecursivePanel(v.Guid)
			}else {
				v.Endpoint = trimListString(v.Endpoint)
				tmpEndpoint := strings.Split(v.Endpoint, ",")
				var endpointStringList []string
				for _,vv := range tmpEndpoint {
					if vv == "" {
						continue
					}
					endpointObj := m.EndpointTable{Guid:vv}
					db.GetEndpoint(&endpointObj)
					if endpointObj.Id > 0 {
						endpointStringList = append(endpointStringList, endpointObj.Guid)
					}else {
						var tmpAddress string
						tmpList := strings.Split(vv, ":")
						if len(tmpList) == 1 {
							tmpAddress = fmt.Sprintf("%s:9100", tmpList[0])
						} else if len(tmpList) == 2 {
							tmpAddress = fmt.Sprintf("%s:%s", tmpList[0], tmpList[1])
						} else {
							tmpMessage += fmt.Sprintf(" endpoint %s validate fail, ", vv)
							continue
						}
						endpointObj = m.EndpointTable{Address: tmpAddress}
						db.GetEndpoint(&endpointObj)
						if endpointObj.Id > 0 {
							endpointStringList = append(endpointStringList, endpointObj.Guid)
						} else {
							tmpMessage += fmt.Sprintf(" endpoint:%s with address %s can not find, ", vv, tmpAddress)
						}
					}
				}
				if tmpMessage != "" {
					errorMessage = tmpMessage
					tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
					successFlag = "1"
					continue
				}
				if len(endpointStringList) > 0 {
					cErr = db.UpdateRecursiveEndpoint(v.Guid, endpointStringList)
				}
			}
			if cErr != nil {
				tmpMessage = fmt.Sprintf("Index:%s update database error:%v", v.CallbackParameter, cErr)
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"1", ErrorMessage:tmpMessage})
				successFlag = "1"
			}else{
				tmpResult = append(tmpResult, resultOutputObj{Guid:v.Guid, CallbackParameter:v.CallbackParameter, ErrorCode:"0", ErrorMessage:""})
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