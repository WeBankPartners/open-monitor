package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

func ListMetric(c *gin.Context) {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	onlyService := c.Query("onlyService")
	result, err := db.MetricListNew(guid, monitorType, serviceGroup, onlyService)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func GetSysMetricTemplate(c *gin.Context) {
	workspace := c.Query("workspace")
	result, err := db.GetSysMetricTemplateConfig(workspace)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func ExportMetric(c *gin.Context) {
	serviceGroup := c.Query("serviceGroup")
	monitorType := c.Query("monitorType")
	result, err := db.MetricListNew("", monitorType, serviceGroup, "Y")
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	b, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		middleware.ReturnHandleError(c, "export metric fail, json marshal object error", marshalErr)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%s.json", "metric_", serviceGroup, time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func ImportMetric(c *gin.Context) {
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
	var paramObj []*models.MetricTable
	var result = &models.MetricImportResultDto{
		SuccessList: []string{},
		FailList:    []string{},
		Message:     "",
	}
	var nameList []string
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
	if len(paramObj) == 0 {
		middleware.ReturnValidateError(c, "can not import empty file")
		return
	}
	serviceGroup := c.Query("serviceGroup")
	if serviceGroup == "" {
		middleware.ReturnValidateError(c, "serviceGroup can not empty")
		return
	}
	for _, obj := range paramObj {
		nameList = append(nameList, obj.Metric)
	}
	if result.FailList, err = db.MetricImport(serviceGroup, middleware.GetOperateUser(c), paramObj); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(result.FailList) == 0 {
		result.SuccessList = nameList
	}
	middleware.ReturnSuccessData(c, result)
}

func QueryMetricTagValue(c *gin.Context) {
	var param models.QueryMetricTagParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	result := []*models.QueryMetricTagResultObj{}
	if param.MetricId == "" {
		middleware.ReturnSuccessData(c, result)
		return
	}
	// 查指标有哪些标签
	metricRow, err := db.GetSimpleMetric(param.MetricId)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	var tagList []string
	tagList, err = db.GetMetricTags(metricRow)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	log.Logger.Debug("QueryMetricTagValue", log.StringList("tagList", tagList))
	if len(tagList) == 0 {
		middleware.ReturnSuccessData(c, result)
		return
	}
	var endpointObj models.EndpointNewTable
	if param.Endpoint != "" {
		endpointObj, _ = db.GetEndpointNew(&models.EndpointNewTable{Guid: param.Endpoint})
	}
	if endpointObj.AgentAddress == "" {
		endpointObj.AgentAddress = ".*"
	}
	metricRow.PromExpr = db.ReplacePromQlKeyword(metricRow.PromExpr, "", &endpointObj, []*models.TagDto{})
	// 查标签值
	seriesMapList, getSeriesErr := datasource.QueryPromSeries(metricRow.PromExpr)
	if getSeriesErr != nil {
		err = fmt.Errorf("query prom series fail,%s ", getSeriesErr)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	for _, v := range tagList {
		tmpValueList := []string{}
		tmpValueDistinctMap := make(map[string]int)
		for _, seriesMap := range seriesMapList {
			if seriesMap == nil {
				continue
			}
			if tmpTagValue, ok := seriesMap[v]; ok {
				if _, existFlag := tmpValueDistinctMap[tmpTagValue]; !existFlag {
					tmpValueList = append(tmpValueList, tmpTagValue)
					tmpValueDistinctMap[tmpTagValue] = 1
				}
			}
		}
		valueObjList := []*models.MetricTagValueObj{}
		for _, tmpValue := range tmpValueList {
			valueObjList = append(valueObjList, &models.MetricTagValueObj{Key: tmpValue, Value: tmpValue})
		}
		result = append(result, &models.QueryMetricTagResultObj{Tag: v, Values: valueObjList})
	}
	middleware.ReturnSuccessData(c, result)
}
