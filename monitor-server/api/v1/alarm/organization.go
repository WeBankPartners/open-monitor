package alarm

import (
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func GetOrganizaionList(c *gin.Context) {
	nameText := c.Query("name")
	endpointText := c.Query("endpoint")
	data, err := db.GetOrganizationList(nameText, endpointText)
	if err != nil {
		mid.ReturnQueryTableError(c, "panel_recursive", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdateOrgPanel(c *gin.Context) {
	var param m.UpdateOrgPanelParam
	operation := c.Param("name")
	if err := c.ShouldBindJSON(&param); err == nil {
		if operation == "delete" && param.Force != "yes" {
			result, getErr := db.GetDeleteServiceGroupAffectList(param.Guid)
			if getErr != nil {
				mid.ReturnHandleError(c, getErr.Error(), getErr)
			} else {
				if len(result) == 0 {
					err = db.UpdateOrganization(operation, param)
					if err != nil {
						mid.ReturnUpdateTableError(c, "panel_recursive", err)
					} else {
						mid.ReturnSuccessData(c, []string{})
					}
				} else {
					mid.ReturnSuccessData(c, result)
				}
			}
		} else {
			err = db.UpdateOrganization(operation, param)
			if err != nil {
				mid.ReturnUpdateTableError(c, "panel_recursive", err)
			} else {
				mid.ReturnSuccess(c)
			}
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetOrgPanelRole(c *gin.Context) {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	data, err := db.GetOrgRoleNew(guid)
	if err != nil {
		mid.ReturnFetchDataError(c, "panel_recursive", "guid", guid)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdateOrgPanelRole(c *gin.Context) {
	var param m.UpdateOrgPanelRoleParam
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateOrgRole(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetOrgPanelEndpoint(c *gin.Context) {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	data, err := db.GetOrgEndpoint(guid)
	if err != nil {
		mid.ReturnHandleError(c, "get organization endpoint fail", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdateOrgPanelEndpoint(c *gin.Context) {
	var param m.UpdateOrgPanelEndpointParam
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateOrgEndpoint(param, mid.GetOperateUser(c))
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateOrgConnect(c *gin.Context) {
	var param m.UpdateOrgConnectParam
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateOrgConnect(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetOrgConnect(c *gin.Context) {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	data, err := db.GetOrgConnect(guid)
	if err != nil {
		mid.ReturnFetchDataError(c, "panel_recursive", "guid", guid)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func IsPluginMode(c *gin.Context) {
	result := m.IsPluginModeResult{IsPlugin: true}
	if m.CoreUrl == "" {
		result.IsPlugin = false
	}
	mid.ReturnSuccessData(c, result)
}

func GetOrgPanelEventList(c *gin.Context) {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	eventList, err := db.GetCoreEventList(c.GetHeader(m.AuthTokenHeader))
	if err != nil {
		mid.ReturnHandleError(c, "get core event list fail", err)
		return
	}
	recordData, err := db.GetOrgCallback(guid)
	if err != nil {
		mid.ReturnFetchDataError(c, "panel_recursive", "guid", guid)
		return
	}
	var result m.GetOrgPanelCallbackData
	result.FiringCallback = append(result.FiringCallback, &m.OptionModel{OptionText: "default empty", OptionValue: ""})
	result.RecoverCallback = append(result.RecoverCallback, &m.OptionModel{OptionText: "default empty", OptionValue: ""})
	var firingFlag, recoverFlag bool
	for _, v := range eventList.Data {
		firingOption := m.OptionModel{OptionText: fmt.Sprintf("%s %s", v.ProcDefName, v.CreatedTime), OptionValue: v.ProcDefKey}
		recoverOption := m.OptionModel{OptionText: fmt.Sprintf("%s %s", v.ProcDefName, v.CreatedTime), OptionValue: v.ProcDefKey}
		if v.ProcDefKey == recordData.FiringCallbackKey && firingFlag == false {
			firingOption.Active = true
			firingFlag = true
		}
		if v.ProcDefKey == recordData.RecoverCallbackKey && recoverFlag == false {
			recoverOption.Active = true
			recoverFlag = true
		}
		result.FiringCallback = append(result.FiringCallback, &firingOption)
		result.RecoverCallback = append(result.RecoverCallback, &recoverOption)
	}
	if !firingFlag {
		result.FiringCallback[0].Active = true
	}
	if !recoverFlag {
		result.RecoverCallback[0].Active = true
	}
	mid.ReturnSuccessData(c, result)
}

func UpdateOrgPanelCallback(c *gin.Context) {
	var param m.UpdateOrgPanelEventParam
	if err := c.ShouldBindJSON(&param); err == nil {
		err = db.UpdateOrgCallback(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func SearchSysPanelData(c *gin.Context) {
	search := c.Query("search")
	endpoint := c.Query("endpoint")
	if !mid.IsIllegalNormalInput(search) {
		mid.ReturnValidateError(c, "Illegal input")
		return
	}
	if search == "." {
		search = ""
	}
	result := db.SearchPanelByName(search, endpoint)
	mid.ReturnSuccessData(c, result)
}

func BatchGetServiceGroup(c *gin.Context) {
	var param m.IdsParam
	if err := c.ShouldBindJSON(&param); err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	result, err := db.BatchGetServiceGroupByIds(param.Ids)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		mid.ReturnSuccessData(c, result)
	}
}
