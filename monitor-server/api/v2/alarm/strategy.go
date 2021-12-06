package alarm

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func QueryAlarmStrategy(c *gin.Context) {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	if queryType == "endpoint" {
		result, err := db.QueryAlarmStrategyByEndpoint(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.QueryAlarmStrategyByGroup(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	}
}

func CreateAlarmStrategy(c *gin.Context) {
	var param models.GroupStrategyObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateAlarmStrategy(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateAlarmStrategy(c *gin.Context) {
	var param models.GroupStrategyObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateAlarmStrategy(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func DeleteAlarmStrategy(c *gin.Context) {
	strategyGuid := c.Param("strategyGuid")
	err := db.DeleteAlarmStrategy(strategyGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func ListStrategyQueryOptions(c *gin.Context) {
	searchType := c.Query("type")
	searchMsg := c.Query("search")
	if searchType == "" {
		middleware.ReturnParamEmptyError(c, "type and search")
		return
	}
	var err error
	var data []*models.OptionModel
	if searchType == "endpoint" {
		data, err = db.ListEndpointOptions(searchMsg)
	} else {
		data, err = db.ListEndpointGroupOptions(searchMsg)
	}
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	for _, v := range data {
		v.OptionTypeName = v.OptionType
	}
	middleware.ReturnSuccessData(c, data)
}
