package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func ListLogMetricMonitor(c *gin.Context) {
	queryType := c.Param("queryType")
	guid := c.Param("guid")
	if queryType == "endpoint" {
		result, err := db.GetLogMetricByEndpoint(guid, false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.GetLogMetricByServiceGroup(guid)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
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
	err = db.CreateLogMetricMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogMetricMonitor(c *gin.Context) {
	var param models.LogMetricMonitorObj
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
	err = db.CreateLogMetricJson(&param)
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
	err = db.UpdateLogMetricJson(&param)
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
	err := db.CreateLogMetricConfig(&param)
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
	err := db.UpdateLogMetricConfig(&param)
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
	result := db.CheckRegExpMatch(param)
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
	result, err := db.GetLogMetricByServiceGroup(serviceGroup)
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
	}
	if err = db.ImportLogMetric(&paramObj); err != nil {
		middleware.ReturnHandleError(c, "import log metric fail", err)
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
	if err = db.ImportLogMetricExcel(logMonitorGuid, logMetricConfigList); err != nil {
		middleware.ReturnHandleError(c, "import log metric from excel data fail", err)
	} else {
		middleware.ReturnSuccess(c)
	}
}
