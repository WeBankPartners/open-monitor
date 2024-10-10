package monitor

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strings"
)

func QueryTypeConfigList(c *gin.Context) {
	var err error
	var list []*models.TypeConfig
	name := c.Query("name")
	if list, err = db.GetTypeConfigList(name); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, list)
}

func AddTypeConfig(c *gin.Context) {
	var param models.TypeConfig
	var typeConfigList []*models.TypeConfig
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	param.Guid = param.DisplayName
	param.CreateUser = middleware.GetOperateUser(c)
	if typeConfigList, err = db.QueryTypeConfigByName(param.DisplayName); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(typeConfigList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).TypeConfigNameRepeatError))
		return
	}
	if err = db.AddTypeConfig(param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func DeleteTypeConfig(c *gin.Context) {
	var typeConfig *models.TypeConfig
	var endpointList []*models.EndpointNewTable
	var endpointGroupList []*models.EndpointGroupTable
	var err error
	id := c.Query("id")
	if strings.TrimSpace(id) == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	if typeConfig, err = db.GetTypeConfig(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if typeConfig == nil {
		middleware.ReturnValidateError(c, "id invalid")
		return
	}
	if typeConfig.SystemType == 1 {
		middleware.ReturnValidateError(c, "system config not allow delete")
		return
	}
	if endpointList, err = db.GetEndpointByMonitorType(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(endpointList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).TypeConfigNameAssociationObjectError))
		return
	}
	if endpointGroupList, err = db.GetEndpointGroupByMonitorType(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(endpointGroupList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).TypeConfigNameAssociationObjectGroupError))
		return
	}
	if err = db.DeleteTypeConfig(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}
