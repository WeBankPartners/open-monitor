package alarm

import (
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v2/monitor"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io"
	"strconv"
	"strings"
)

// defaultEndpointGroup 默认对象组
var defaultEndpointGroup = []string{"default_pod_group", "default_host_group", "default_http_group", "default_java_group", "default_ping_group",
	"default_snmp_group", "default_mysql_group", "default_nginx_group", "default_redis_group", "default_telnet_group", "default_process_group"}

func ListEndpointGroup(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	search := c.Query("search")
	monitorType := c.Query("monitor_type")
	param := models.QueryRequestParam{Sorting: &models.QueryRequestSorting{
		Asc:   false,
		Field: "update_time",
	}}
	if size > 0 {
		param.Paging = true
		param.Pageable = &models.PageInfo{PageSize: size, StartIndex: page - 1}
	}
	if search != "" {
		param.Filters = []*models.QueryRequestFilterObj{{Name: "guid", Operator: "like", Value: search}}
	}
	if monitorType != "" {
		values := strings.Split(monitorType, ",")
		var interArr []interface{}
		for _, value := range values {
			interArr = append(interArr, value)
		}
		if len(param.Filters) == 0 {
			param.Filters = []*models.QueryRequestFilterObj{{Name: "monitor_type", Operator: "in", Value: interArr}}
		} else {
			param.Filters = append(param.Filters, &models.QueryRequestFilterObj{Name: "monitor_type", Operator: "in", Value: interArr})
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

func ImportEndpointGroup(c *gin.Context) {
	var list []*models.EndpointGroupTable
	file, err := c.FormFile("file")
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	f, err := file.Open()
	if err != nil {
		middleware.ReturnHandleError(c, "file open error ", err)
		return
	}
	b, err := io.ReadAll(f)
	defer f.Close()
	if err != nil {
		middleware.ReturnHandleError(c, "read content fail error ", err)
		return
	}
	if err = json.Unmarshal(b, &list); err != nil {
		middleware.ReturnHandleError(c, "json unmarshal fail error ", err)
		return
	}
	if len(list) == 0 {
		middleware.ReturnValidateError(c, "can not import empty file")
		return
	}
	defaultEndpointGroupMap := monitor.ConvertArr2Map(defaultEndpointGroup)
	for _, endpointGroup := range list {
		// 默认对象组直接过滤掉
		if defaultEndpointGroupMap[endpointGroup.DisplayName] {
			continue
		}
		if err = db.CreateEndpointGroup(endpointGroup, middleware.GetOperateUser(c)); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
	}
	middleware.ReturnSuccess(c)
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
