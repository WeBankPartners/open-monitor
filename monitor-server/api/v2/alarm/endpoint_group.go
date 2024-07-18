package alarm

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func ListEndpointGroup(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	search := c.Query("search")
	monitorType := c.Query("monitor_type")
	param := models.QueryRequestParam{}
	if size > 0 {
		param.Paging = true
		param.Pageable = &models.PageInfo{PageSize: size, StartIndex: page - 1}
	}
	if search != "" {
		param.Filters = []*models.QueryRequestFilterObj{{Name: "guid", Operator: "like", Value: search}}
	}
	if monitorType != "" {
		values := strings.Split(monitorType, ",")
		if len(param.Filters) == 0 {
			param.Filters = []*models.QueryRequestFilterObj{{Name: "monitor_type", Operator: "in", Value: values}}
		} else {
			param.Filters = append(param.Filters, &models.QueryRequestFilterObj{Name: "monitor_type", Operator: "in", Value: values})
		}
	}
	pageInfo, rowData, err := db.ListEndpointGroup(&param)
	returnData := models.TableData{Data: rowData, Page: page, Size: size, Num: pageInfo.TotalRows}
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, returnData)
	}
}

func EndpointGroupOptions(c *gin.Context) {
	var err error
	var result []string
	if result, err = db.ListEndpointGroupMonitoryType(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func CreateEndpointGroup(c *gin.Context) {
	var param models.EndpointGroupTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateEndpointGroup(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateEndpointGroup(c *gin.Context) {
	var param models.EndpointGroupTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateEndpointGroup(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func DeleteEndpointGroup(c *gin.Context) {
	endpointGroupGuid := c.Param("groupGuid")
	err := db.DeleteEndpointGroup(endpointGroupGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func GetGroupEndpointRel(c *gin.Context) {
	endpointGroupGuid := c.Param("groupGuid")
	result, err := db.GetGroupEndpointRel(endpointGroupGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func UpdateGroupEndpoint(c *gin.Context) {
	var param models.UpdateGroupEndpointParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateGroupEndpoint(&param, middleware.GetOperateUser(c), false)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncPrometheusRuleFile(param.GroupGuid, false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func GetGroupEndpointNotify(c *gin.Context) {
	endpointGroupGuid := c.Param("groupGuid")
	result, err := db.GetGroupEndpointNotify(endpointGroupGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func UpdateGroupEndpointNotify(c *gin.Context) {
	endpointGroupGuid := c.Param("groupGuid")
	var param []*models.NotifyObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateGroupEndpointNotify(endpointGroupGuid, param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}
