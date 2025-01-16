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
	"strconv"
	"strings"
	"time"
)

var defaultMonitorTypeMetricMap = map[string][]string{
	"host": {"file_handler_free_percent", "mem_used", "disk_iops", "load_1min", "mem_total", "ping_loss", "ping_time",
		"ping_alive", "telnet_alive", "disk_read_bytes", "net_if_in_bytes", "cpu_used_percent", "disk_write_bytes", "mem_used_percent",
		"net_if_out_bytes", "process_mem_byte", "cpu_detail_percent", "process_alive_count", "volume_used_percent", "process_cpu_used_percent"},
	"telnet":  {"telnet_alive"},
	"redis":   {"redis_alive", "redis_cmd_num", "redis_db_keys", "redis_mem_used", "redis_expire_key", "redis_client_used_percent"},
	"process": {"process_mem_byte", "process_alive_count", "process_cpu_used_percent"},
	"pod":     {"pod_cpu_used_percent", "pod_mem_used_percent"},
	"ping":    {"ping_loss", "ping_alive"},
	"nginx":   {"nginx_connect_active", "nginx_handle_request"},
	"mysql":   {"mysql_alive", "mysql_requests", "db_count_change", "db_monitor_count", "mysql_threads_max", "mysql_buffer_status", "mysql_threads_connected", "mysql_connect_used_percent"},
	"jvm":     {"jvm_gc_time", "tomcat_request", "jvm_thread_count", "gc_marksweep_time", "tomcat_connection", "jvm_memory_heap_max", "jvm_memory_heap_used", "heap_mem_used_percent"},
	"http":    {"http_status"},
}

func ListMetric(c *gin.Context) {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	onlyService := c.Query("onlyService")
	endpointGroup := c.Query("endpointGroup")
	endpoint := c.Query("endpoint")
	query := c.Query("query")
	metric := c.Query("metric")
	startIndex, _ := strconv.Atoi(c.Query("startIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	pageInfo, result, err := db.MetricListNew(guid, monitorType, serviceGroup, onlyService, endpointGroup, endpoint, query, metric, startIndex, pageSize)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if pageSize == 0 {
		middleware.ReturnSuccessData(c, result)
		return
	}
	// 分页返回
	middleware.ReturnPageData(c, pageInfo, result)
}

func ListMetricCount(c *gin.Context) {
	var err error
	var countRes models.MetricCountRes
	var metricList []*models.MetricTable
	var metricComparisonList []*models.MetricComparisonExtend
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	endpointGroup := c.Query("endpointGroup")
	onlyService := c.Query("onlyService")
	endpoint := c.Query("endpoint")
	metric := c.Query("metric")
	if _, metricList, err = db.MetricListNew("", monitorType, serviceGroup, onlyService, endpointGroup, endpoint, "", metric, 0, 0); err != nil {
		middleware.ReturnServerHandleError(c, err)
	}
	if _, metricComparisonList, err = db.MetricComparisonListNew("", monitorType, serviceGroup, onlyService, endpointGroup, endpoint, "", 0, 0); err != nil {
		middleware.ReturnServerHandleError(c, err)
	}
	countRes.Count = len(metricList)
	countRes.ComparisonCount = len(metricComparisonList)
	middleware.ReturnSuccessData(c, countRes)
}

func ListMetricComparison(c *gin.Context) {
	guid := c.Query("guid")
	monitorType := c.Query("monitorType")
	serviceGroup := c.Query("serviceGroup")
	onlyService := c.Query("onlyService")
	endpointGroup := c.Query("endpointGroup")
	endpoint := c.Query("endpoint")
	metric := c.Query("metric")
	startIndex, _ := strconv.Atoi(c.Query("startIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	pageInfo, result, err := db.MetricComparisonListNew(guid, monitorType, serviceGroup, onlyService, endpointGroup, endpoint, metric, startIndex, pageSize)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if pageSize > 0 {
		middleware.ReturnPageData(c, pageInfo, result)
		return
	}
	middleware.ReturnSuccessData(c, result)
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
	var fileNamePrefix = "metric_"
	serviceGroup := c.Query("serviceGroup")
	monitorType := c.Query("monitorType")
	endpointGroup := c.Query("endpointGroup")
	comparison := c.Query("comparison")
	if comparison == "Y" {
		fileNamePrefix = "metric_comparison_"
		_, result, err = db.MetricComparisonListNew("", monitorType, serviceGroup, "Y", endpointGroup, "", "", 0, 0)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
	} else {
		_, result, err = db.MetricListNew("", monitorType, serviceGroup, "Y", endpointGroup, "", "", "", 0, 0)
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
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%s.json", fileNamePrefix, serviceGroup, time.Now().Format("20060102150405")))
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
	monitorType := c.Query("monitorType")
	if serviceGroup == "" && endPointGroup == "" && monitorType == "" {
		middleware.ReturnValidateError(c, "serviceGroup or endpointGroup  or monitorType can not empty")
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
		if err = db.SyncMetricComparisonData(); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
	} else {
		// 走原始指标的导入逻辑
		if subFaiList, err = db.MetricImport(monitorType, serviceGroup, endPointGroup, middleware.GetOperateUser(c), ConvertMetricComparison2MetricList(newParamObj)); err != nil {
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
	var orginMetricRow *models.MetricTable
	var logType string
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	var result []*models.QueryMetricTagResultObj
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
	if metricRow == nil {
		middleware.ReturnServerHandleError(c, fmt.Errorf("metricId %s is invalid", param.MetricId))
		return
	}
	if logType, err = db.GetLogTypeByLogMetricGroup(metricRow.LogMetricGroup); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	var tagList []string
	// 如果是同环比指标需要用原始指标进去查询
	if orginMetricRow, err = db.GetOriginMetricByComparisonId(param.MetricId); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	tagConfigValueMap := make(map[string][]string)
	if orginMetricRow != nil {
		// 同环比指标 默认新增 calc_type标签
		tagList, tagConfigValueMap, err = db.GetMetricTags(orginMetricRow)
		tagList = append(tagList, "calc_type")
	} else {
		tagList, tagConfigValueMap, err = db.GetMetricTags(metricRow)
	}
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
	if param.Endpoint != "" && param.Endpoint != param.ServiceGroup {
		endpointObj, _ = db.GetEndpointNew(&models.EndpointNewTable{Guid: param.Endpoint})
	} else if param.ServiceGroup != "" {
		endpointList, getEndpointListErr := db.GetRecursiveEndpointByTypeNew(param.ServiceGroup, metricRow.MonitorType)
		if getEndpointListErr != nil {
			err = fmt.Errorf("Try to get endpoints from object:%s fail,%s ", param.ServiceGroup, getEndpointListErr.Error())
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
		if len(endpointList) > 0 {
			endpointObj = *endpointList[0]
		}
	}
	if endpointObj.AgentAddress == "" {
		endpointObj.AgentAddress = ".*"
	}
	metricRow.PromExpr = db.ReplacePromQlKeyword(metricRow.PromExpr, "", &endpointObj, []*models.TagDto{})
	// 查标签值
	log.Logger.Debug("QueryPromSeries start", log.String("promExpr", metricRow.PromExpr))
	seriesMapList, getSeriesErr := datasource.QueryPromSeries(metricRow.PromExpr)
	if getSeriesErr != nil {
		err = fmt.Errorf("query prom series fail,%s ", getSeriesErr)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	log.Logger.Debug("QueryPromSeries end", log.JsonObj("result", seriesMapList))
	for _, v := range tagList {
		tmpValueList := []string{}
		tmpValueDistinctMap := make(map[string]int)
		for _, seriesMap := range seriesMapList {
			if seriesMap == nil {
				continue
			}
			// 如果该指标为自定义类型的业务配置创建,tags内容: tags="test_service_code=addUser,test_retcode=200",需要做特殊解析处理
			if logType == models.LogMonitorCustomType && seriesMap["tags"] != "" {
				seriesMap = datasource.ResetPrometheusMetricMap(seriesMap)
			}
			if tmpTagValue, ok := seriesMap[v]; ok {
				if _, existFlag := tmpValueDistinctMap[tmpTagValue]; !existFlag {
					tmpValueList = append(tmpValueList, tmpTagValue)
					tmpValueDistinctMap[tmpTagValue] = 1
				}
			}
		}
		if configValueList, ok := tagConfigValueMap[v]; ok {
			for _, configValue := range configValueList {
				if _, existFlag := tmpValueDistinctMap[configValue]; !existFlag {
					tmpValueList = append(tmpValueList, configValue)
					tmpValueDistinctMap[configValue] = 1
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
	var comparison models.MetricComparison
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
	if metric == nil || metric.Guid == "" {
		middleware.ReturnValidateError(c, "metricId is invalid")
		return
	}
	promQl := db.NewPromExpr(db.GetComparisonMetricId(metric.Metric, param.ComparisonType, param.CalcMethod, param.CalcPeriod))
	log.Logger.Debug("CheckPrometheusQL", log.String("promQl", promQl))
	if err = datasource.CheckPrometheusQL(promQl); err != nil {
		log.Logger.Debug("CheckPrometheusQL error", log.Error(err))
		middleware.ReturnValidateError(c, "metric is invalid")
		return
	}
	if strings.TrimSpace(param.MetricComparisonId) == "" {
		// 查询同环比数据
		newMetricId := db.GetComparisonMetricId(metric.Guid, param.ComparisonType, param.CalcMethod, param.CalcPeriod)
		if comparison, err = db.GetComparisonMetricByMetricId(newMetricId); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
		if comparison.Guid != "" {
			middleware.ReturnServerHandleError(c, models.GetMessageMap(c).AddComparisonMetricRepeatError)
			return
		}
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
	if err = db.SyncMetricComparisonData(); err != nil {
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
	if err = db.SyncMetricComparisonData(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func ConvertMetricComparison2MetricList(metricComparisonList []*models.MetricComparisonExtend) []*models.MetricTable {
	var list []*models.MetricTable
	if len(metricComparisonList) > 0 {
		for _, metricComparison := range metricComparisonList {
			var promExpr = metricComparison.ImportPromExpr
			if promExpr == "" {
				promExpr = metricComparison.PromExpr
			}
			list = append(list, &models.MetricTable{
				Guid:               metricComparison.Guid,
				Metric:             metricComparison.Metric,
				MonitorType:        metricComparison.MonitorType,
				PromExpr:           promExpr,
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
