package monitor

import (
	"bytes"
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
	"strings"
	"time"
)

func ListMetric(c *gin.Context) {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	onlyService := c.Query("onlyService")
	endpointGroup := c.Query("endpointGroup")
	endpoint := c.Query("endpoint")
	result, err := db.MetricListNew(guid, monitorType, serviceGroup, onlyService, endpointGroup, endpoint)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}
func ListMetricComparison(c *gin.Context) {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	onlyService := c.Query("onlyService")
	endpointGroup := c.Query("endpointGroup")
	endpoint := c.Query("endpoint")
	result, err := db.MetricComparisonListNew(guid, monitorType, serviceGroup, onlyService, endpointGroup, endpoint)
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
	var err error
	var result interface{}
	serviceGroup := c.Query("serviceGroup")
	monitorType := c.Query("monitorType")
	endpointGroup := c.Query("endpointGroup")
	comparison := c.Query("comparison")
	if comparison == "Y" {
		result, err = db.MetricComparisonListNew("", monitorType, serviceGroup, "Y", endpointGroup, "")
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
	} else {
		result, err = db.MetricListNew("", monitorType, serviceGroup, "Y", endpointGroup, "")
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
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
	var paramObj, newParamObj []*models.MetricComparisonExtend
	var comparison bool
	var metricMap = make(map[string]bool)
	var result = &models.MetricImportResultDto{
		SuccessList: []string{},
		FailList:    []string{},
		Message:     "",
	}
	var nameList, subFaiList []string
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
	endPointGroup := c.Query("endpointGroup")
	if serviceGroup == "" && endPointGroup == "" {
		middleware.ReturnValidateError(c, "serviceGroup or endpointGroup can not empty")
		return
	}
	for i, obj := range paramObj {
		if !metricMap[obj.Metric] {
			metricMap[obj.Metric] = true
			newParamObj = append(newParamObj, obj)
			nameList = append(nameList, obj.Metric)
		} else {
			result.FailList = append(result.FailList, obj.Metric)
		}
		if i == 0 && strings.TrimSpace(obj.MetricId) != "" {
			comparison = true
		}
	}
	if comparison {
		// 走同环比导入逻辑
		if subFaiList, err = db.MetricComparisonImport(middleware.GetOperateUser(c), newParamObj); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
		if err = syncMetricComparisonData(); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
	} else {
		// 走原始指标的导入逻辑
		if subFaiList, err = db.MetricImport(serviceGroup, endPointGroup, middleware.GetOperateUser(c), ConvertMetricComparison2MetricList(newParamObj)); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
	}

	if len(subFaiList) > 0 {
		result.FailList = append(result.FailList, subFaiList...)
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

// AddOrUpdateComparisonMetric 添加更新同环比监控配置
func AddOrUpdateComparisonMetric(c *gin.Context) {
	var param models.MetricComparisonParam
	var metric *models.MetricTable
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if strings.TrimSpace(param.MetricId) == "" {
		middleware.ReturnParamEmptyError(c, "metricId")
		return
	}
	if strings.TrimSpace(param.ComparisonType) == "" || len(param.CalcType) == 0 {
		middleware.ReturnParamEmptyError(c, "comparisonType or calcType")
		return
	}
	if metric, err = db.GetMetric(param.MetricId); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if metric == nil {
		middleware.ReturnValidateError(c, "metricId is invalid")
		return
	}
	if strings.TrimSpace(param.MetricComparisonId) == "" {
		// 新增同环比
		if err = db.AddComparisonMetric(param, metric, middleware.GetOperateUser(c)); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
	} else {
		// 更新同环比
		if err = db.UpdateComparisonMetric(param.MetricComparisonId, param.CalcType); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
	}
	if err = syncMetricComparisonData(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func DeleteComparisonMetric(c *gin.Context) {
	var err error
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	if err = db.DeleteComparisonMetric(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if err = syncMetricComparisonData(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

// syncMetricComparisonData 同步同环比指标数据
func syncMetricComparisonData() (err error) {
	var list []*models.MetricComparisonDto
	var resByteArr []byte
	var response models.Response
	if list, err = db.GetComparisonMetricDtoList(); err != nil {
		return
	}
	if len(list) > 0 {
		param, _ := json.Marshal(list)
		if resByteArr, err = HttpPost("http://127.0.0.1:8181/receive", param); err != nil {
			return
		}
		if err = json.Unmarshal(resByteArr, &response); err != nil {
			return
		}
		if response.Status != "OK" {
			err = fmt.Errorf(response.Message)
		}
	}
	return
}

// HttpPost Post请求
func HttpPost(url string, postBytes []byte) (byteArr []byte, err error) {
	req, reqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(postBytes))
	if reqErr != nil {
		err = fmt.Errorf("new http reqeust fail,%s ", reqErr.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = fmt.Errorf("do http reqeust fail,%s ", respErr.Error())
		return
	}
	byteArr, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}

func ConvertMetricComparison2MetricList(metricComparisonList []*models.MetricComparisonExtend) []*models.MetricTable {
	var list []*models.MetricTable
	if len(metricComparisonList) > 0 {
		for _, metricComparison := range metricComparisonList {
			list = append(list, &models.MetricTable{
				Guid:               metricComparison.Guid,
				Metric:             metricComparison.MetricId,
				MonitorType:        metricComparison.MonitorType,
				PromExpr:           metricComparison.PromExpr,
				TagOwner:           metricComparison.TagOwner,
				ServiceGroup:       metricComparison.ServiceGroup,
				Workspace:          metricComparison.Workspace,
				CreateTime:         metricComparison.CreateTime,
				UpdateTime:         metricComparison.UpdateTime,
				CreateUser:         metricComparison.CreateUser,
				UpdateUser:         metricComparison.UpdateUser,
				LogMetricConfig:    metricComparison.LogMetricConfig,
				LogMetricTemplate:  metricComparison.LogMetricTemplate,
				LogMetricGroup:     metricComparison.LogMetricGroup,
				EndpointGroup:      metricComparison.EndpointGroup,
				MetricType:         metricComparison.MetricType,
				LogMetricGroupName: metricComparison.LogMetricGroupName,
				GroupType:          metricComparison.GroupType,
				GroupName:          metricComparison.GroupName,
			})
		}
	}
	return list
}
