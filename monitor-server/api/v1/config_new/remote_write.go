package config_new

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func RemoteWriteConfigList(c *gin.Context) {
	result, err := db.RemoteWriteConfigList()
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func RemoteWriteConfigCreate(c *gin.Context) {
	var param models.RemoteWriteConfigTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if !middleware.IsIllegalNormalInput(param.Id) {
		middleware.ReturnValidateError(c, "Param id is illegal")
		return
	}
	err := db.RemoteWriteConfigCreate(param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func RemoteWriteConfigUpdate(c *gin.Context) {
	var param models.RemoteWriteConfigTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.RemoteWriteConfigUpdate(param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func RemoteWriteConfigDelete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	err := db.RemoteWriteConfigDelete(id)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}
