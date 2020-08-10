package alarm

import (
	"github.com/gin-gonic/gin"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
)

func GetOrganizaionList(c *gin.Context)  {
	data,err := db.GetOrganizationList()
	if err != nil {
		mid.ReturnQueryTableError(c, "panel_recursive", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdateOrgPanel(c *gin.Context)  {
	var param m.UpdateOrgPanelParam
	operation := c.Param("name")
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrganization(operation, param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
		}else{
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetOrgPanelRole(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	data,err := db.GetOrgRole(guid)
	if err != nil {
		mid.ReturnFetchDataError(c, "panel_recursive", "guid", guid)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdateOrgPanelRole(c *gin.Context)  {
	var param m.UpdateOrgPanelRoleParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgRole(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetOrgPanelEndpoint(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	data,err := db.GetOrgEndpoint(guid)
	if err != nil {
		mid.ReturnHandleError(c, "get organization endpoint fail", err)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func UpdateOrgPanelEndpoint(c *gin.Context)  {
	var param m.UpdateOrgPanelEndpointParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgEndpoint(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateOrgConnect(c *gin.Context)  {
	var param m.UpdateOrgConnectParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgConnect(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

func GetOrgConnect(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	data,err := db.GetOrgConnect(guid)
	if err != nil {
		mid.ReturnFetchDataError(c, "panel_recursive", "guid", guid)
		return
	}
	mid.ReturnSuccessData(c, data)
}

func IsPluginMode(c *gin.Context)  {
	result := m.IsPluginModeResult{IsPlugin:true}
	if m.CoreUrl == "" {
		result.IsPlugin = false
	}
	mid.ReturnSuccessData(c, result)
}

func GetOrgPanelEventList(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnParamEmptyError(c, "guid")
		return
	}
	eventList,err := db.GetCoreEventList()
	if err != nil {
		mid.ReturnHandleError(c, "get core event list fail", err)
		return
	}
	recordData,err := db.GetOrgCallback(guid)
	if err != nil {
		mid.ReturnFetchDataError(c, "panel_recursive", "guid", guid)
		return
	}
	var result m.GetOrgPanelCallbackData
	result.FiringCallback = append(result.FiringCallback, &m.OptionModel{OptionText:"default empty",OptionValue:""})
	result.RecoverCallback = append(result.RecoverCallback, &m.OptionModel{OptionText:"default empty",OptionValue:""})
	var firingFlag,recoverFlag bool
	for _,v := range eventList.Data {
		firingOption := m.OptionModel{OptionText:fmt.Sprintf("%s %s", v.ProcDefName, v.CreatedTime), OptionValue:v.ProcDefKey}
		recoverOption := m.OptionModel{OptionText:fmt.Sprintf("%s %s", v.ProcDefName, v.CreatedTime), OptionValue:v.ProcDefKey}
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

func UpdateOrgPanelCallback(c *gin.Context)  {
	var param m.UpdateOrgPanelEventParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgCallback(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "panel_recursive", err)
			return
		}
		mid.ReturnSuccess(c)
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}