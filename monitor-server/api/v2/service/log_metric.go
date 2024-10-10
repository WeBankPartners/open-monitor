package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/WeBankPartners/go-common-lib/pcre"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ListLogMetricMonitor(c *gin.Context) {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	metricKey := c.Query("metricKey")
	if queryType == "endpoint" {
		result, err := db.GetLogMetricByEndpoint(guid, metricKey, false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.GetLogMetricByServiceGroup(guid, metricKey)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, []*models.LogMetricQueryObj{&result})
		}
	}
}

func GetLogMetricMonitor(c *gin.Context) {
	logMonitorMonitorGuid := c.Param("logMonitorGuid")
	result, err := db.GetLogMetricMonitor(logMonitorMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricMonitor(c *gin.Context) {
	var param models.LogMetricMonitorCreateDto
	var list []*models.LogMetricMonitorTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	if len(param.LogPath) == 0 {
		err = fmt.Errorf("Param log_path is empty ")
	}
	for _, v := range param.LogPath {
		if !strings.HasPrefix(v, "/") {
			err = fmt.Errorf("Path:%s illegal ", v)
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	// 校验路径是否重复
	if list, err = db.GetLogMetricMonitorByCond(param.LogPath, "", param.ServiceGroup); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(list) > 0 {
		middleware.ReturnValidateError(c, fmt.Errorf("path:%s Already exists", list[0].LogPath).Error())
		return
	}
	err = db.CreateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogMetricMonitor(c *gin.Context) {
	var param models.LogMetricMonitorObj
	var list []*models.LogMetricMonitorTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result, err := db.GetLogMetricMonitor(param.Guid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	hostEndpointList := []string{}
	for _, v := range result.EndpointRel {
		hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
	}
	for _, v := range param.EndpointRel {
		hostEndpointList = append(hostEndpointList, v.SourceEndpoint)
	}
	// 校验路径是否重复
	if list, err = db.GetLogMetricMonitorByCond([]string{param.LogPath}, param.Guid, param.ServiceGroup); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(list) > 0 {
		middleware.ReturnValidateError(c, fmt.Errorf("path:%s Already exists", list[0].LogPath).Error())
		return
	}
	err = db.UpdateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricNodeExporterConfig(hostEndpointList)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricMonitor(c *gin.Context) {
	logMonitorGuid := c.Param("logMonitorGuid")
	err := db.DeleteLogMetricMonitor(logMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func GetLogMetricJson(c *gin.Context) {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	result, err := db.GetLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricJson(c *gin.Context) {
	var param models.LogMetricJsonObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	for _, v := range param.MetricList {
		if middleware.IsIllegalName(v.Metric) {
			err = fmt.Errorf("metric name illegal")
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err = db.CreateLogMetricJson(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateLogMetricJson(c *gin.Context) {
	var param models.LogMetricJsonObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	for _, v := range param.MetricList {
		if middleware.IsIllegalName(v.Metric) {
			err = fmt.Errorf("metric name illegal")
			break
		}
	}
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err = db.UpdateLogMetricJson(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricJson(c *gin.Context) {
	logMonitorJsonGuid := c.Param("logMonitorJsonGuid")
	logMetricMonitor, err := db.DeleteLogMetricJson(logMonitorJsonGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if logMetricMonitor != "" {
			err = syncLogMetricMonitorConfig(logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func GetLogMetricConfig(c *gin.Context) {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	result, err := db.GetLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricConfig(c *gin.Context) {
	var param models.LogMetricConfigObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if middleware.IsIllegalName(param.Metric) {
		middleware.ReturnValidateError(c, "metric name illegal")
		return
	}
	if strings.Contains(param.Regular, "\n") {
		middleware.ReturnValidateError(c, "regular illegal")
		return
	}
	err := db.CreateLogMetricConfig(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateLogMetricConfig(c *gin.Context) {
	var param models.LogMetricConfigObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if middleware.IsIllegalName(param.Metric) {
		middleware.ReturnValidateError(c, "metric name illegal")
		return
	}
	if strings.Contains(param.Regular, "\n") {
		middleware.ReturnValidateError(c, "regular illegal")
		return
	}
	err := db.UpdateLogMetricConfig(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricConfig(c *gin.Context) {
	logMonitorConfigGuid := c.Param("logMonitorConfigGuid")
	logMetricMonitor, err := db.DeleteLogMetricConfig(logMonitorConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if logMetricMonitor != "" {
			err = syncLogMetricMonitorConfig(logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func syncLogMetricMonitorConfig(logMetricMonitor string) error {
	endpointList := []string{}
	endpointRel := db.ListLogMetricEndpointRel(logMetricMonitor)
	for _, v := range endpointRel {
		if v.TargetEndpoint != "" {
			endpointList = append(endpointList, v.SourceEndpoint)
		}
	}
	err := syncLogMetricNodeExporterConfig(endpointList)
	return err
}

func syncLogMetricNodeExporterConfig(endpointList []string) error {
	err := db.SyncLogMetricExporterConfig(endpointList)
	return err
}

func CheckRegExpMatch(c *gin.Context) {
	var param models.CheckRegExpParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result := models.CheckRegExpResult{}
	var matchString string
	result.MatchText, matchString = db.CheckRegExpMatchPCRE(param)
	if strings.HasPrefix(matchString, "{") {
		resultJsonMap := make(map[string]interface{})
		if unmarshalErr := json.Unmarshal([]byte(matchString), &resultJsonMap); unmarshalErr == nil {
			result.JsonObj = resultJsonMap
			for k, _ := range resultJsonMap {
				result.JsonKeyList = append(result.JsonKeyList, k)
			}
			sort.Strings(result.JsonKeyList)
		}
	}
	middleware.ReturnSuccessData(c, result)
}

func GetServiceGroupEndpointRel(c *gin.Context) {
	serviceGroup := c.Query("serviceGroup")
	sourceType := c.Query("sourceType")
	targetType := c.Query("targetType")
	result, err := db.GetServiceGroupEndpointRel(serviceGroup, sourceType, targetType)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func ExportLogMetric(c *gin.Context) {
	serviceGroup := c.Query("serviceGroup")
	result, err := db.GetLogMetricByServiceGroup(serviceGroup, "")
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	b, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		middleware.ReturnHandleError(c, "export log metric fail, json marshal object error", marshalErr)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%s.json", "log_metric_", result.DisplayName, time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func ImportLogMetric(c *gin.Context) {
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
	var paramObj models.LogMetricQueryObj
	b, err := io.ReadAll(f)
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
	serviceGroup := c.Query("serviceGroup")
	if serviceGroup == "" {
		middleware.ReturnValidateError(c, "serviceGroup can not empty")
		return
	}
	paramObj.Guid = serviceGroup
	for _, logMonitor := range paramObj.Config {
		logMonitor.ServiceGroup = serviceGroup
		for _, logMetric := range logMonitor.MetricConfigList {
			logMetric.ServiceGroup = serviceGroup
		}
		for _, logJson := range logMonitor.JsonConfigList {
			for _, logMetric := range logJson.MetricList {
				logMetric.ServiceGroup = serviceGroup
			}
		}
		for _, metricGroup := range logMonitor.MetricGroups {
			metricGroup.ServiceGroup = serviceGroup
		}
	}
	for _, dbMonitor := range paramObj.DBConfig {
		dbMonitor.ServiceGroup = serviceGroup
	}
	if err = db.ImportLogMetric(&paramObj, middleware.GetOperateUser(c), middleware.GetMessageMap(c)); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func ImportLogMetricExcel(c *gin.Context) {
	logMonitorGuid := c.Param("logMonitorGuid")
	if logMonitorGuid == "" {
		middleware.ReturnValidateError(c, "url param logMonitorGuid can not empty")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	f, fileOpenErr := file.Open()
	if fileOpenErr != nil {
		middleware.ReturnHandleError(c, "file open error ", fileOpenErr)
		return
	}
	defer f.Close()
	b, readFileErr := ioutil.ReadAll(f)
	if readFileErr != nil {
		middleware.ReturnHandleError(c, "read content fail error ", readFileErr)
		return
	}
	excelObj, readExcelDataErr := excelize.OpenReader(bytes.NewReader(b))
	if readExcelDataErr != nil {
		middleware.ReturnHandleError(c, "read excel data error ", readExcelDataErr)
		return
	}
	var logMetricConfigList []*models.LogMetricConfigObj
	for rowIndex, row := range excelObj.GetRows(excelObj.GetSheetName(1)) {
		if rowIndex == 0 || len(row) < 4 {
			continue
		}
		logMetricConfigList = append(logMetricConfigList, &models.LogMetricConfigObj{Metric: row[0], DisplayName: row[1], Regular: row[2], AggType: row[3], LogMetricMonitor: logMonitorGuid})
	}
	if len(logMetricConfigList) == 0 {
		err = fmt.Errorf("can not find any enable excel row config")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if err = db.ImportLogMetricExcel(logMonitorGuid, middleware.GetOperateUser(c), logMetricConfigList); err != nil {
		middleware.ReturnHandleError(c, "import log metric from excel data fail", err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func ListLogMonitorTemplate(c *gin.Context) {
	var param models.LogMonitorTemplateListParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result, err := db.ListLogMonitorTemplate(&param, middleware.GetOperateUserRoles(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func ListLogMonitorTemplateOptions(c *gin.Context) {
	var result []string
	var err error
	if result, err = db.ListLogMonitorTemplateOptions(); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func GetLogMonitorTemplate(c *gin.Context) {
	logMonitorTemplateGuid := c.Param("logMonitorTemplateGuid")
	result, err := db.GetLogMonitorTemplate(logMonitorTemplateGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMonitorTemplate(c *gin.Context) {
	var param models.LogMonitorTemplateDto
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := validateLogMonitorTemplateParam(&param)
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err = db.CreateLogMonitorTemplate(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func validateLogMonitorTemplateParam(param *models.LogMonitorTemplateDto) (err error) {
	if param.LogType != models.LogMonitorJsonType && param.LogType != models.LogMonitorRegularType && param.LogType != models.LogMonitorCustomType {
		err = fmt.Errorf("param json type illegal")
		return
	}
	if param.CalcResultObj == nil {
		err = fmt.Errorf("calc result can not empty")
		return
	}
	param.Name = strings.TrimSpace(param.Name)
	if existLogMonitorTemplate, getErr := db.GetLogMonitorTemplateByName(param.Guid, param.Name); getErr != nil {
		err = getErr
		return
	} else if existLogMonitorTemplate != nil {
		err = fmt.Errorf("log monitor template name:%s duplicate", param.Name)
		return
	}
	calcResultBytes, _ := json.Marshal(param.CalcResultObj)
	param.CalcResult = string(calcResultBytes)
	if len(param.CalcResult) > 50000 {
		err = fmt.Errorf("calc result too long")
		return
	}
	for _, v := range param.ParamList {
		if middleware.IsIllegalDisplayName(v.DisplayName) {
			err = fmt.Errorf("log param display name:%s illegal", v.DisplayName)
			return
		}
		if param.LogType == "json" && middleware.IsIllegalNameNew(v.JsonKey) {
			err = fmt.Errorf("log param jsonKey:%s illegal", v.JsonKey)
			return
		}
		if param.LogType == "regular" && v.Regular == "" {
			err = fmt.Errorf("log param regular can not empty ")
			return
		}
	}
	for _, v := range param.MetricList {
		if middleware.IsIllegalMetric(v.Metric) {
			err = fmt.Errorf("metric : %s illegal", v.Metric)
			return
		}
		if middleware.IsIllegalDisplayName(v.DisplayName) {
			err = fmt.Errorf("metric: %s metric displayName: %s illegal", v.Metric, v.DisplayName)
			return
		}
		tagConfigByte, _ := json.Marshal(v.TagConfigList)
		v.TagConfig = string(tagConfigByte)
		if v.LogParamName == "" {
			err = fmt.Errorf("metric: %s log param name can not empty", v.Metric)
			return
		}
		if v.AutoAlarm {
			if err = checkThresholdWarnConfigInvalid(v.Metric, v.RangeConfig); err != nil {
				return
			}
		}
		if v.Step == 0 {
			v.Step = 10
		}
	}
	return
}

func checkThresholdWarnConfigInvalid(metric, rangeConfig string) error {
	temp := &models.ThresholdConfig{}
	var intTime int
	var err error
	if err = json.Unmarshal([]byte(rangeConfig), temp); err != nil {
		return getError(metric, rangeConfig)
	}
	if temp == nil {
		return getError(metric, rangeConfig)
	}
	if temp.Operator == "" || temp.Time == "" || temp.TimeUnit == "" || temp.Threshold == "" {
		return getError(metric, rangeConfig)
	}
	if _, err = strconv.ParseFloat(temp.Threshold, 64); err != nil {
		return getError(metric, rangeConfig)
	}
	if intTime, err = strconv.Atoi(temp.Time); err != nil || intTime < 0 {
		return getError(metric, rangeConfig)
	}
	return nil
}

func getError(metric, rangeConfig string) error {
	return fmt.Errorf("metric: %s alarm config:%+v illegal", metric, rangeConfig)
}

func UpdateLogMonitorTemplate(c *gin.Context) {
	var param models.LogMonitorTemplateDto
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := validateLogMonitorTemplateParam(&param)
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var affectEndpoints []string
	affectEndpoints, err = db.UpdateLogMonitorTemplate(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricNodeExporterConfig(affectEndpoints)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMonitorTemplate(c *gin.Context) {
	logMonitorTemplateGuid := c.Param("logMonitorTemplateGuid")
	err := db.DeleteLogMonitorTemplate(logMonitorTemplateGuid, middleware.GetMessageMap(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func CheckLogMonitorRegExpMatch(c *gin.Context) {
	var param models.LogMonitorRegMatchParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result := []*models.LogParamTemplateObj{}
	for _, v := range param.ParamList {
		_, v.DemoMatchValue = db.CheckRegExpMatchPCRE(models.CheckRegExpParam{RegString: v.Regular, TestContext: param.DemoLog})
		result = append(result, v)
	}
	middleware.ReturnSuccessData(c, result)
}

func GetLogMetricGroup(c *gin.Context) {
	logMetricGroupGuid := c.Param("logMetricGroupGuid")
	result, err := db.GetLogMetricGroup(logMetricGroupGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricGroup(c *gin.Context) {
	var param models.LogMetricGroupWithTemplate
	var prefixMap map[string]int
	var result *models.CreateLogMetricGroupDto
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.LogMetricMonitorGuid == "" || param.LogMonitorTemplateGuid == "" {
		err := fmt.Errorf("LogMetricMonitorGuid and LogMonitorTemplateGuid can not empty")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalMetricPrefixCode(param.MetricPrefixCode) {
		err := fmt.Errorf("param MetricPrefixCode validate fail")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	param.Name = strings.TrimSpace(param.Name)
	if err = db.ValidateLogMetricGroupName("", param.Name, param.LogMetricMonitorGuid); err != nil {
		err = fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameDuplicateError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalDisplayName(param.Name) {
		err := fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameIllegalError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalTargetValueCode(param.CodeStringMap) {
		middleware.ReturnServerHandleError(c, fmt.Errorf("target_value code repeat"))
		return
	}
	if prefixMap, err = db.GetLogMetricMonitorMetricPrefixMap(param.LogMetricMonitorGuid); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if _, ok := prefixMap[param.MetricPrefixCode]; ok {
		err := fmt.Errorf("Prefix: %s duplidate ", param.MetricPrefixCode)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if result, err = db.CreateLogMetricGroup(&param, middleware.GetOperateUser(c), middleware.GetOperateUserRoles(c), middleware.GetMessageMap(c)); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if err = syncLogMetricMonitorConfig(param.LogMetricMonitorGuid); err != nil {
		middleware.ReturnError(c, 200, middleware.GetMessageMap(c).SaveDoneButSyncFail, err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func UpdateLogMetricGroup(c *gin.Context) {
	var param models.LogMetricGroupWithTemplate
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.LogMetricGroupGuid == "" || param.LogMetricMonitorGuid == "" {
		err := fmt.Errorf("LogMetricGroupGuid and LogMetricMonitorGuid can not empty")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	param.Name = strings.TrimSpace(param.Name)
	if err := db.ValidateLogMetricGroupName(param.LogMetricGroupGuid, param.Name, param.LogMetricMonitorGuid); err != nil {
		err = fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameDuplicateError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalDisplayName(param.Name) {
		err := fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameIllegalError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalTargetValueCode(param.CodeStringMap) {
		middleware.ReturnServerHandleError(c, fmt.Errorf("target_value code repeat"))
		return
	}
	err := db.UpdateLogMetricGroup(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitorGuid)
		if err != nil {
			middleware.ReturnError(c, 200, middleware.GetMessageMap(c).SaveDoneButSyncFail, err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogMetricGroup(c *gin.Context) {
	logMetricGroupGuid := c.Param("logMetricGroupGuid")
	logMetricMonitor, err := db.DeleteLogMetricGroup(logMetricGroupGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		if logMetricMonitor != "" {
			err = syncLogMetricMonitorConfig(logMetricMonitor)
			if err != nil {
				middleware.ReturnHandleError(c, err.Error(), err)
			} else {
				middleware.ReturnSuccess(c)
			}
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func GetLogMetricCustomGroup(c *gin.Context) {
	logMetricGroupGuid := c.Param("logMetricGroupGuid")
	result, err := db.GetLogMetricCustomGroup(logMetricGroupGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func CreateLogMetricCustomGroup(c *gin.Context) {
	var param models.LogMetricGroupObj
	var prefixMap map[string]int
	var result *models.CreateLogMetricGroupDto
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.LogMetricMonitor == "" {
		err = fmt.Errorf("LogMetricMonitor can not empty")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if param.LogMetricMonitor != "" && param.MetricPrefixCode != "" {
		if prefixMap, err = db.GetLogMetricMonitorMetricPrefixMap(param.LogMetricGroup.LogMetricMonitor); err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
		if _, ok := prefixMap[param.MetricPrefixCode]; ok {
			err := fmt.Errorf("Prefix: %s duplidate ", param.MetricPrefixCode)
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
	} else {
		existMetricMap, err := db.GetServiceGroupMetricMap(param.ServiceGroup)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
			return
		}
		for _, v := range param.MetricList {
			if _, existFlag := existMetricMap[v.Metric]; existFlag {
				err = fmt.Errorf(middleware.GetMessageMap(c).MetricDuplicateError, v.Metric)
				break
			} else {
				existMetricMap[v.Metric] = param.MonitorType
			}
		}
		if err != nil {
			middleware.ReturnError(c, 200, err.Error(), err)
			return
		}
	}
	if len(param.MetricList) > 0 {
		for _, metric := range param.MetricList {
			if middleware.IsIllegalLogParamNameOrMetric(metric.LogParamName) || middleware.IsIllegalLogParamNameOrMetric(metric.Metric) {
				middleware.ReturnValidateError(c, "log_param_name or metric param invalid")
				return
			}
		}
	}
	if err := db.ValidateLogMetricGroupName(param.Guid, param.Name, param.LogMetricMonitor); err != nil {
		err = fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameDuplicateError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalDisplayName(param.Name) {
		err := fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameIllegalError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	param.ServiceGroup, param.MonitorType = db.GetLogMetricServiceGroup(param.LogMetricMonitor)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if result, err = db.CreateLogMetricCustomGroup(&param, middleware.GetOperateUser(c), middleware.GetOperateUserRoles(c), middleware.GetMessageMap(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if err = syncLogMetricMonitorConfig(param.LogMetricMonitor); err != nil {
		middleware.ReturnError(c, 200, middleware.GetMessageMap(c).SaveDoneButSyncFail, err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func UpdateLogMetricCustomGroup(c *gin.Context) {
	var param models.LogMetricGroupObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.Guid == "" {
		err := fmt.Errorf("guid can not empty")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if len(param.MetricList) > 0 {
		for _, metric := range param.MetricList {
			if middleware.IsIllegalLogParamNameOrMetric(metric.LogParamName) || middleware.IsIllegalLogParamNameOrMetric(metric.Metric) {
				middleware.ReturnValidateError(c, "log_param_name or metric param invalid")
				return
			}
		}
	}
	if err := db.ValidateLogMetricGroupName(param.Guid, param.Name, param.LogMetricMonitor); err != nil {
		err = fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameDuplicateError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if middleware.IsIllegalDisplayName(param.Name) {
		err := fmt.Errorf(middleware.GetMessageMap(c).LogGroupNameIllegalError, param.Name)
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	param.ServiceGroup, param.MonitorType = db.GetLogMetricServiceGroup(param.LogMetricMonitor)
	existMetricMap, err := db.GetServiceGroupMetricMap(param.ServiceGroup)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	for _, v := range param.MetricList {
		if existLogMetricGroup, existFlag := existMetricMap[v.Metric]; existFlag {
			if existLogMetricGroup != "" && existLogMetricGroup != param.Guid {
				err = fmt.Errorf(middleware.GetMessageMap(c).MetricDuplicateError, v.Metric)
				break
			}
		} else {
			existMetricMap[v.Metric] = param.MonitorType
		}
	}
	if err != nil {
		middleware.ReturnError(c, 200, err.Error(), err)
		return
	}
	err = db.UpdateLogMetricCustomGroup(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogMetricMonitorConfig(param.LogMetricMonitor)
		if err != nil {
			middleware.ReturnError(c, 200, middleware.GetMessageMap(c).SaveDoneButSyncFail, err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func GetLogMonitorTemplateServiceGroup(c *gin.Context) {
	logMonitorTemplateGuid := c.Param("logMonitorTemplateGuid")
	result, err := db.GetLogMonitorTemplateServiceGroup(logMonitorTemplateGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func LogMonitorTemplateExport(c *gin.Context) {
	var param models.LogTemplateExportParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	if len(param.GuidList) == 0 {
		err = fmt.Errorf("param empty")
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	resultData := []*models.LogMonitorTemplateDto{}
	for _, v := range param.GuidList {
		templateObj, tmpErr := db.GetLogMonitorTemplate(v)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		resultData = append(resultData, templateObj)
	}
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	b, marshalErr := json.Marshal(resultData)
	if marshalErr != nil {
		middleware.ReturnHandleError(c, "export log metric fail, json marshal object error", marshalErr)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.json", "log_template_config_", time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func LogMonitorTemplateImport(c *gin.Context) {
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
	var paramObj []*models.LogMonitorTemplateDto
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
	var affectEndpoints []string
	affectEndpoints, err = db.ImportLogMonitorTemplate(paramObj, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	err = syncLogMetricNodeExporterConfig(affectEndpoints)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func LogMonitorDataMapRegMatch(c *gin.Context) {
	var param models.LogMetricDataMapMatchDto
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if !param.IsRegexp {
		param.Match = true
		middleware.ReturnSuccessData(c, param)
		return
	}
	ce, err := pcre.Compile(param.Regexp, 0)
	if err != nil {
		middleware.ReturnHandleError(c, err.Message, fmt.Errorf(err.Message))
		return
	}
	if mat := ce.MatcherString(param.Content, 0); mat != nil {
		if mat.Matches() {
			param.Match = true
		}
	}
	middleware.ReturnSuccessData(c, param)
}
