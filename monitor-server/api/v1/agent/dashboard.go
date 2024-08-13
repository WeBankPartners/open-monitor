package agent

import (
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type requestPanelObj struct {
	RequestId string            `json:"requestId"`
	Inputs    []panelRequestObj `json:"inputs"`
}

type panelRequestObj struct {
	ConfirmToken      string      `json:"confirmToken"`
	CallbackParameter string      `json:"callbackParameter"`
	Guid              string      `json:"guid"`
	DisplayName       string      `json:"display_name"`
	Parent            string      `json:"parent"`
	Endpoint          interface{} `json:"endpoint"`
	Email             string      `json:"email"`
	Phone             string      `json:"phone"`
	Role              interface{} `json:"role"`
	FiringCallback    string      `json:"firing_callback"`
	RecoverCallback   string      `json:"recover_callback"`
	Type              string      `json:"type"`
	DeleteAll         string      `json:"delete_all"`
}

func ExportPanelAdd(c *gin.Context) {
	var param requestPanelObj
	var result resultObj
	if err := c.ShouldBindJSON(&param); err == nil {
		if len(param.Inputs) == 0 {
			result = resultObj{ResultCode: "0", ResultMessage: fmt.Sprintf(mid.GetMessageMap(c).ParamEmptyError, "inputs")}
			log.Logger.Warn(result.ResultMessage)
			c.JSON(http.StatusOK, result)
			return
		}
		roleMap := db.GetRoleMap()
		var tmpResult []resultOutputObj
		successFlag := "0"
		errorMessage := "Done"
		for _, v := range param.Inputs {
			var tmpMessage string
			v.Parent = trimListString(v.Parent)
			v.Email = trimListString(v.Email)
			v.Phone = trimListString(v.Phone)
			inputRoleList := m.TransPluginMultiStringParam(v.Role)
			tmpEndpoint := m.TransPluginMultiStringParam(v.Endpoint)
			tmpParent := strings.Split(v.Parent, ",")
			checkRoleErr := db.CheckRoleIllegal(inputRoleList, roleMap)
			if checkRoleErr != nil {
				tmpMessage = checkRoleErr.Error()
			}
			if v.Guid == "" {
				tmpMessage = fmt.Sprintf(mid.GetMessageMap(c).ParamEmptyError, "guid")
			}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpMessage})
				successFlag = "1"
				continue
			}
			var endpointStringList []string
			for _, vv := range tmpEndpoint {
				if vv == "" {
					continue
				}
				endpointObj := m.EndpointTable{Guid: vv}
				db.GetEndpoint(&endpointObj)
				if endpointObj.Id > 0 {
					endpointStringList = append(endpointStringList, endpointObj.Guid)
				} else {
					var tmpAddress string
					tmpList := strings.Split(vv, ":")
					if len(tmpList) == 2 {
						tmpAddress = fmt.Sprintf("%s:%s", tmpList[0], tmpList[1])
					} else {
						tmpMessage += fmt.Sprintf(mid.GetMessageMap(c).ParamTypeError, "endpoint", "[guid] or [ip:port]")
						continue
					}
					endpointObj = m.EndpointTable{Address: tmpAddress}
					db.GetEndpoint(&endpointObj)
					if endpointObj.Id > 0 {
						endpointStringList = append(endpointStringList, endpointObj.Guid)
					} else {
						tmpMessage += fmt.Sprintf(mid.GetMessageMap(c).FetchTableDataError, "endpoint", "address", tmpAddress)
					}
				}
			}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpMessage})
				successFlag = "1"
				continue
			}
			err := db.UpdateRecursivePanel(m.PanelRecursiveTable{Guid: v.Guid, DisplayName: v.DisplayName, Parent: strings.Join(tmpParent, "^"), Endpoint: strings.Join(endpointStringList, "^"), Email: v.Email, Phone: v.Phone, Role: strings.Join(inputRoleList, ","), FiringCallbackKey: v.FiringCallback, RecoverCallbackKey: v.RecoverCallback, ObjType: v.Type}, mid.GetOperateUser(c))
			if err != nil {
				tmpMessage = fmt.Sprintf(mid.GetMessageMap(c).UpdateTableError, "recursive_panel")
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpMessage, ErrorDetail: err.Error()})
				successFlag = "1"
			} else {
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "0", ErrorMessage: ""})
			}
		}
		result = resultObj{ResultCode: successFlag, ResultMessage: errorMessage, Results: resultOutput{Outputs: tmpResult}}
		log.Logger.Info("Plugin result", log.JsonObj("result", result))
		mid.ReturnData(c, result)
	} else {
		result = resultObj{ResultCode: "1", ResultMessage: fmt.Sprintf(mid.GetMessageMap(c).ParamValidateError, err.Error())}
		log.Logger.Warn(result.ResultMessage)
		c.JSON(http.StatusBadRequest, result)
	}
}

func trimListString(input string) string {
	if strings.HasPrefix(input, "[") {
		input = input[1:]
	}
	if strings.HasSuffix(input, "]") {
		input = input[:len(input)-1]
	}
	return input
}

func GetPanelRecursive(c *gin.Context) {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	err, result := db.GetRecursivePanel(guid)
	if err != nil {
		mid.ReturnQueryTableError(c, "panel_recursive", err)
	} else {
		mid.ReturnSuccessData(c, result)
	}
}

func GetPanelRecursiveEndpointType(c *gin.Context) {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	result, err := db.ListRecursiveEndpointType(guid)
	if err != nil {
		mid.ReturnQueryTableError(c, "panel_recursive", err)
	} else {
		mid.ReturnSuccessData(c, result)
	}
}

func ExportPanelDelete(c *gin.Context) {
	var param requestPanelObj
	var result resultObj
	if err := c.ShouldBindJSON(&param); err == nil {
		var tmpResult []resultOutputObj
		successFlag := "0"
		errorMessage := "Done"
		for _, v := range param.Inputs {
			var tmpMessage string
			affectList, affectErr := db.GetDeleteServiceGroupAffectList(v.Guid)
			if affectErr != nil {
				tmpMessage = fmt.Sprintf("Try to get affect object list fail,%s ", affectErr.Error())
			}
			if v.ConfirmToken != "Y" && len(affectList) > 0 {
				tmpMessage = fmt.Sprintf("This action will delete these config:%s ", strings.Join(affectList, " \n "))
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "-1", ErrorMessage: tmpMessage})
				successFlag = "1"
				continue
			}
			if v.Guid == "" {
				tmpMessage = fmt.Sprintf(mid.GetMessageMap(c).ParamEmptyError, "guid")
			}
			if tmpMessage != "" {
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpMessage})
				successFlag = "1"
				continue
			}
			var cErr error
			if strings.ToLower(v.DeleteAll) == "y" || strings.ToLower(v.DeleteAll) == "yes" {
				cErr = db.DeleteRecursivePanel(v.Guid)
			} else {
				tmpEndpoint := m.TransPluginMultiStringParam(v.Endpoint)
				var endpointStringList []string
				for _, vv := range tmpEndpoint {
					if vv == "" {
						continue
					}
					endpointObj := m.EndpointTable{Guid: vv}
					db.GetEndpoint(&endpointObj)
					if endpointObj.Id > 0 {
						endpointStringList = append(endpointStringList, endpointObj.Guid)
					} else {
						var tmpAddress string
						tmpList := strings.Split(vv, ":")
						if len(tmpList) == 2 {
							tmpAddress = fmt.Sprintf("%s:%s", tmpList[0], tmpList[1])
						} else {
							tmpMessage += fmt.Sprintf(mid.GetMessageMap(c).ParamTypeError, "endpoint", "[guid] or [ip:port]")
							continue
						}
						endpointObj = m.EndpointTable{Address: tmpAddress}
						db.GetEndpoint(&endpointObj)
						if endpointObj.Id > 0 {
							endpointStringList = append(endpointStringList, endpointObj.Guid)
						} else {
							tmpMessage += fmt.Sprintf(mid.GetMessageMap(c).FetchTableDataError, "endpoint", "address", tmpAddress)
						}
					}
				}
				if tmpMessage != "" {
					errorMessage = tmpMessage
					tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpMessage})
					successFlag = "1"
					continue
				}
				if len(endpointStringList) > 0 {
					cErr = db.UpdateRecursiveEndpoint(v.Guid, mid.GetOperateUser(c), endpointStringList)
				}
			}
			if cErr != nil {
				tmpMessage = fmt.Sprintf(mid.GetMessageMap(c).UpdateTableError, "recursive_panel")
				errorMessage = tmpMessage
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpMessage, ErrorDetail: cErr.Error()})
				successFlag = "1"
			} else {
				tmpResult = append(tmpResult, resultOutputObj{Guid: v.Guid, CallbackParameter: v.CallbackParameter, ErrorCode: "0", ErrorMessage: ""})
			}
		}
		result = resultObj{ResultCode: successFlag, ResultMessage: errorMessage, Results: resultOutput{Outputs: tmpResult}}
		log.Logger.Info("Plugin result", log.JsonObj("result", result))
		mid.ReturnData(c, result)
	} else {
		result = resultObj{ResultCode: "1", ResultMessage: fmt.Sprintf(mid.GetMessageMap(c).ParamValidateError, err.Error())}
		log.Logger.Warn(result.ResultMessage)
		c.JSON(http.StatusBadRequest, result)
	}
}
