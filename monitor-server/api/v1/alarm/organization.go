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
		mid.ReturnError(c, "Get organization list fail", err)
		return
	}
	mid.ReturnData(c, data)
}

func UpdateOrgPanel(c *gin.Context)  {
	var param m.UpdateOrgPanelParam
	operation := c.Param("name")
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrganization(operation, param)
		if err != nil {
			mid.ReturnError(c, operation + " organization panel fail", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func GetOrgPanelRole(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "Parameter guid can not be empty")
		return
	}
	data,err := db.GetOrgRole(guid)
	if err != nil {
		mid.ReturnError(c, "Get organization role fail", err)
		return
	}
	mid.ReturnData(c, data)
}

func UpdateOrgPanelRole(c *gin.Context)  {
	var param m.UpdateOrgPanelRoleParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgRole(param)
		if err != nil {
			mid.ReturnError(c, "Update organization role fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func GetOrgPanelEndpoint(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "Parameter guid can not be empty")
		return
	}
	data,err := db.GetOrgEndpoint(guid)
	if err != nil {
		mid.ReturnError(c, "Get organization endpoint fail", err)
		return
	}
	mid.ReturnData(c, data)
}

func UpdateOrgPanelEndpoint(c *gin.Context)  {
	var param m.UpdateOrgPanelEndpointParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgEndpoint(param)
		if err != nil {
			mid.ReturnError(c, "Update organization endpoint fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func IsPluginMode(c *gin.Context)  {
	result := m.IsPluginModeResult{IsPlugin:true}
	if m.CoreUrl == "" {
		result.IsPlugin = false
	}
	mid.ReturnData(c, result)
}

func GetOrgPanelEventList(c *gin.Context)  {
	guid := c.Query("guid")
	if guid == "" {
		mid.ReturnValidateFail(c, "Parameter guid can not be empty")
		return
	}
	eventList,err := db.GetCoreEventList()
	if err != nil {
		mid.ReturnError(c, "Get core event list fail", err)
		return
	}
	recordData,err := db.GetOrgCallback(guid)
	if err != nil {
		mid.ReturnError(c, "Get organization callback fail", err)
		return
	}
	var result m.GetOrgPanelCallbackData
	result.FiringCallback = append(result.FiringCallback, &m.OptionModel{OptionText:"default null",OptionValue:""})
	result.RecoverCallback = append(result.RecoverCallback, &m.OptionModel{OptionText:"default null",OptionValue:""})
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
	mid.ReturnData(c, result)
}

func UpdateOrgPanelCallback(c *gin.Context)  {
	var param m.UpdateOrgPanelEventParam
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateOrgCallback(param)
		if err != nil {
			mid.ReturnError(c, "Update organization callback fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}