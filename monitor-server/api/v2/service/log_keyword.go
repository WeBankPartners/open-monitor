package service

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

func ListLogKeywordMonitor(c *gin.Context) {
	queryType := c.Query("type")
	guid := c.Query("guid")
	alarmName := c.Query("alarmName")
	if queryType == "endpoint" {
		result, err := db.GetLogKeywordByEndpoint(guid, alarmName, false)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	} else {
		result, err := db.GetLogKeywordByServiceGroup(guid, alarmName)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccessData(c, result)
		}
	}
}

func CreateLogKeywordMonitor(c *gin.Context) {
	var param models.LogKeywordMonitorCreateObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	var err error
	if len(param.LogPath) == 0 {
		err = fmt.Errorf("Param log_path is empty ")
		middleware.ReturnValidateError(c, err.Error())
		return
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
	err = db.CreateLogKeywordMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogKeywordMonitor(c *gin.Context) {
	var param models.LogKeywordMonitorObj
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if !strings.HasPrefix(param.LogPath, "/") {
		middleware.ReturnValidateError(c, fmt.Sprintf("Path:%s illegal ", param.LogPath))
		return
	}
	var endpointList []string
	for _, v := range db.ListLogKeywordEndpointRel(param.Guid) {
		endpointList = append(endpointList, v.SourceEndpoint)
	}
	for _, v := range param.EndpointRel {
		endpointList = append(endpointList, v.SourceEndpoint)
	}
	err := db.UpdateLogKeywordMonitor(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordNodeExporterConfig(endpointList)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func DeleteLogKeywordMonitor(c *gin.Context) {
	logKeywordMonitorGuid := c.Param("logKeywordMonitorGuid")
	err := db.DeleteLogKeywordMonitor(logKeywordMonitorGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func CreateLogKeyword(c *gin.Context) {
	var param models.LogKeywordConfigTable
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if len(param.ActiveWindowList) > 0 {
		param.ActiveWindow = strings.Join(param.ActiveWindowList, ",")
	}
	var sameNameList, sameKeywordList []*models.LogKeywordConfigTable
	sameNameList, sameKeywordList, err = db.GetLogKeywordConfigUniqueData(param.Guid, param.Name, param.Keyword, param.LogKeywordMonitor)
	if err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(sameNameList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).AlertNameRepeatError))
		return
	}
	if len(sameKeywordList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).AlertKeywordRepeatError))
		return
	}
	err = db.CreateLogKeyword(&param, middleware.GetOperateUser(c))
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordMonitorConfig(param.LogKeywordMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func UpdateLogKeyword(c *gin.Context) {
	var param models.LogKeywordConfigTable
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if len(param.ActiveWindowList) > 0 {
		param.ActiveWindow = strings.Join(param.ActiveWindowList, ",")
	}
	logKeywordConfig, getExistErr := db.GetSimpleLogKeywordConfig(param.Guid)
	if getExistErr != nil {
		middleware.ReturnValidateError(c, getExistErr.Error())
		return
	}
	param.LogKeywordMonitor = logKeywordConfig.LogKeywordMonitor
	var sameNameList, sameKeywordList []*models.LogKeywordConfigTable
	sameNameList, sameKeywordList, err = db.GetLogKeywordConfigUniqueData(param.Guid, param.Name, param.Keyword, param.LogKeywordMonitor)
	if err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(sameNameList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).AlertNameRepeatError))
		return
	}
	if len(sameKeywordList) > 0 {
		middleware.ReturnServerHandleError(c, fmt.Errorf(middleware.GetMessageMap(c).AlertKeywordRepeatError))
		return
	}
	if err = db.UpdateLogKeyword(&param, logKeywordConfig, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	err = syncLogKeywordMonitorConfig(param.LogKeywordMonitor)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func DeleteLogKeyword(c *gin.Context) {
	logKeywordConfigGuid := c.Query("guid")
	logKeywordConfig, getExistErr := db.GetSimpleLogKeywordConfig(logKeywordConfigGuid)
	if getExistErr != nil {
		middleware.ReturnValidateError(c, getExistErr.Error())
		return
	}
	err := db.DeleteLogKeyword(logKeywordConfigGuid)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		err = syncLogKeywordMonitorConfig(logKeywordConfig.LogKeywordMonitor)
		if err != nil {
			middleware.ReturnHandleError(c, err.Error(), err)
		} else {
			middleware.ReturnSuccess(c)
		}
	}
}

func syncLogKeywordMonitorConfig(logKeywordMonitor string) error {
	endpointList := []string{}
	endpointRel := db.ListLogKeywordEndpointRel(logKeywordMonitor)
	for _, v := range endpointRel {
		if v.TargetEndpoint != "" {
			endpointList = append(endpointList, v.SourceEndpoint)
		}
	}
	err := syncLogKeywordNodeExporterConfig(endpointList)
	return err
}

func syncLogKeywordNodeExporterConfig(endpointList []string) error {
	err := db.SyncLogKeywordExporterConfig(endpointList)
	return err
}

func ExportLogAndDbKeyword(c *gin.Context) {
	serviceGroup := c.Query("serviceGroup")
	result, err := db.GetLogKeywordByServiceGroup(serviceGroup, "")
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	if len(result) == 0 {
		middleware.ReturnHandleError(c, "log keyword config list is empty", nil)
		return
	}
	b, marshalErr := json.Marshal(result[0])
	if marshalErr != nil {
		middleware.ReturnHandleError(c, "export log metric fail, json marshal object error", marshalErr)
		return
	}
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%s.json", "log_keyword_", result[0].DisplayName, time.Now().Format("20060102150405")))
	c.Data(http.StatusOK, "application/octet-stream", b)
}

func ImportLogAndDbKeyword(c *gin.Context) {
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
	var paramObj models.LogKeywordServiceGroupObj
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
	for _, logKeyword := range paramObj.Config {
		logKeyword.ServiceGroup = serviceGroup
	}
	if err = db.ImportLogAndDbKeyword(&paramObj, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnServerHandleError(c, err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func UpdateLogKeywordNotify(c *gin.Context) {
	var param models.LogKeywordNotifyParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if param.LogKeywordMonitor == "" {
		middleware.ReturnValidateError(c, "param log_keyword_monitor illegal")
		return
	}
	err := db.UpdateLogKeywordNotify(&param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}
