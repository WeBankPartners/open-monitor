package alarm

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func QueryAlarmStrategy(c *gin.Context) {
	var param models.AlarmStrategyQueryParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.QueryType == "endpoint" {
		result, err := db.QueryAlarmStrategyByEndpoint(param.Guid, param.AlarmName, param.Show, middleware.GetOperateUser(c))
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else if param.QueryType == "group" {
		result, err := db.QueryAlarmStrategyByGroup(param.Guid, param.AlarmName, param.Show, middleware.GetOperateUser(c))
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else if param.QueryType == "service" {
		result, err := db.QueryAlarmStrategyByServiceGroup(param.Guid, param.AlarmName, param.Show, middleware.GetOperateUser(c))
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		middleware.ReturnValidateError(c, "queryType illegal")
	}
}

func CreateAlarmStrategy(c *gin.Context) {
	var param models.GroupStrategyObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	param.Guid = ""
	if len(param.ActiveWindowList) > 0 {
		param.ActiveWindow = strings.Join(param.ActiveWindowList, ",")
	}
	if param.ActiveWindow != "" {
		if !middleware.ValidateActiveWindowString(param.ActiveWindow) {
			middleware.ReturnValidateError(c, "Param active_window validate fail")
			return
		}
	} else {
		param.ActiveWindow = models.DefaultActiveWindow
	}
	param.Name = strings.TrimSpace(param.Name)
	if err := validateStrategyCondition(param.Conditions); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := db.ValidateAlarmStrategyName(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.CreateAlarmStrategy(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncPrometheusRuleFile(param.EndpointGroup, false)
		if err != nil {
			middleware.ReturnError(c, 200, middleware.GetMessageMap(c).SaveDoneButSyncFail, err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateAlarmStrategy(c *gin.Context) {
	var param models.GroupStrategyObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if len(param.ActiveWindowList) > 0 {
		param.ActiveWindow = strings.Join(param.ActiveWindowList, ",")
	}
	if param.ActiveWindow != "" {
		if !middleware.ValidateActiveWindowString(param.ActiveWindow) {
			middleware.ReturnValidateError(c, "Param active_window validate fail")
			return
		}
	} else {
		param.ActiveWindow = models.DefaultActiveWindow
	}
	param.Name = strings.TrimSpace(param.Name)
	if err := validateStrategyCondition(param.Conditions); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := db.ValidateAlarmStrategyName(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.UpdateAlarmStrategy(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncPrometheusRuleFile(param.EndpointGroup, false)
		if err != nil {
			middleware.ReturnError(c, 200, middleware.GetMessageMap(c).SaveDoneButSyncFail, err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func validateStrategyCondition(strategyList []*models.StrategyConditionObj) (err error) {
	for _, v := range strategyList {
		if !middleware.IsIllegalCond(v.Condition) || !middleware.IsIllegalLast(v.Last) {
			err = fmt.Errorf("condition: %s or last: %s illegal", v.Condition, v.Last)
			return
		}
	}
	return
}

func DeleteAlarmStrategy(c *gin.Context) {
	strategyGuid := c.Param("strategyGuid")
	endpointGroup, err := db.DeleteAlarmStrategy(strategyGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = db.SyncPrometheusRuleFile(endpointGroup, false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func ExportAlarmStrategy(c *gin.Context) {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	var result []*models.EndpointStrategyObj
	var err error
	if queryType == "group" {
		result, err = db.QueryAlarmStrategyByGroup(guid, "", "", middleware.GetOperateUser(c))
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
	} else if queryType == "service" {
		result, err = db.QueryAlarmStrategyByServiceGroup(guid, "", "", middleware.GetOperateUser(c))
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
	} else {
		middleware.ReturnHandleError(c, "queryType:"+queryType+" can not export strategy ", nil)
		return
	}
	b, err := json.Marshal(result)
	if err != nil {
		middleware.ReturnHandleError(c, "export alarm strategy fail, json marshal object error", err)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.json", "monitor_strategy_", time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func ImportAlarmStrategy(c *gin.Context) {
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
	var paramObj []*models.EndpointStrategyObj
	b, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		middleware.ReturnHandleError(c, "read content fail error ", err)
		return
	}
	err = json.Unmarshal(b, &paramObj)
	if err != nil {
		middleware.ReturnHandleError(c, "json unmarshal fail error ", err)
		return
	}
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	var metricNotFound, nameDuplicate []string
	err, metricNotFound, nameDuplicate = db.ImportAlarmStrategy(queryType, guid, paramObj, middleware.GetOperateUser(c))
	if err != nil {
		if len(metricNotFound) > 0 {
			err = fmt.Errorf(middleware.GetMessageMap(c).MetricNotFound, strings.Join(metricNotFound, ","))
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
		if len(nameDuplicate) > 0 {
			err = fmt.Errorf(middleware.GetMessageMap(c).StrategyNameImportDuplicateError, strings.Join(nameDuplicate, ","))
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
		middleware.ReturnHandleError(c, "import alarm strategy error:"+err.Error(), err)
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
	} else if searchType == "group" {
		data, err = db.ListEndpointGroupOptions(searchMsg)
	} else if searchType == "service" {
		data, err = db.ListServiceGroupOptions(searchMsg)
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

func ListCallbackEvent(c *gin.Context) {
	eventList, err := db.GetCoreEventList(c.GetHeader(models.AuthTokenHeader))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, eventList.Data)
	}
}

func ListAlarmStrategyWorkFlow(c *gin.Context) {
	var result []*models.WorkflowDto
	var err error
	if result, err = db.GetAlarmStrategyNotifyWorkflowList(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	for _, dto := range result {
		index := strings.Index(dto.Name, "[")
		name := dto.Name
		if index >= 0 {
			dto.Name = name[:index]
			dto.Version = name[index+1 : len(name)-1]
		}
	}
	middleware.ReturnSuccessData(c, result)
}
