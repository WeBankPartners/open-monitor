package middleware

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type RespJson struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResponsePageData struct {
	PageInfo models.PageInfo `json:"pageInfo"`
	Contents interface{}     `json:"contents"`
}

func ReturnError(c *gin.Context, code int, msg string, err error) {
	log.Logger.Error(fmt.Sprintf("Request %s fail", c.FullPath()), log.Int("Code", code), log.String("Message", msg), log.Error(err))
	if models.Config().Log.Level == "debug" {
		c.JSON(code, RespJson{Status: "ERROR", Code: 200, Message: msg, Data: err})
	} else {
		c.JSON(code, RespJson{Status: "ERROR", Code: 200, Message: msg})
	}
}

func ReturnSuccessWithMessage(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: msg})
}

func ReturnSuccess(c *gin.Context) {
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: GetMessageMap(c).Success})
}

func ReturnData(c *gin.Context, data interface{}) {
	if fmt.Sprintf("%v", data) == `[]` {
		data = []string{}
	}
	c.JSON(http.StatusOK, data)
}

func ReturnSuccessData(c *gin.Context, data interface{}) {
	if fmt.Sprintf("%v", data) == `[]` {
		data = []string{}
	}
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: GetMessageMap(c).Success, Data: data})
}

func ReturnPageData(c *gin.Context, pageInfo models.PageInfo, data interface{}) {
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: GetMessageMap(c).Success, Data: ResponsePageData{PageInfo: pageInfo, Contents: data}})
}

func ReturnValidateError(c *gin.Context, msg string) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).ParamValidateError, msg), nil)
}

func ReturnParamTypeError(c *gin.Context, paramName, typeName string) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).ParamTypeError, paramName, typeName), nil)
}

func ReturnParamEmptyError(c *gin.Context, key string) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).ParamEmptyError, key), nil)
}

func ReturnQueryTableError(c *gin.Context, table string, err error) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).QueryTableError, table), err)
}

func ReturnFetchDataError(c *gin.Context, table, key, value string) {
	tmpErrorMessage := GetMessageMap(c).FetchTableDataError
	tmpErrorMessage = strings.Replace(tmpErrorMessage, "%t", table, -1)
	tmpErrorMessage = strings.Replace(tmpErrorMessage, "%k", key, -1)
	tmpErrorMessage = strings.Replace(tmpErrorMessage, "%v", value, -1)
	ReturnError(c, http.StatusOK, tmpErrorMessage, nil)
}

func ReturnUpdateTableError(c *gin.Context, table string, err error) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).UpdateTableError, table)+" detail: "+err.Error(), err)
}

func ReturnDeleteTableError(c *gin.Context, table, key, value string, err error) {
	tmpErrorMessage := GetMessageMap(c).DeleteTableDataError
	tmpErrorMessage = strings.Replace(tmpErrorMessage, "%t", table, -1)
	tmpErrorMessage = strings.Replace(tmpErrorMessage, "%k", key, -1)
	tmpErrorMessage = strings.Replace(tmpErrorMessage, "%v", value, -1)
	ReturnError(c, http.StatusOK, tmpErrorMessage, err)
}

func ReturnBodyError(c *gin.Context, err error) {
	ReturnError(c, http.StatusOK, GetMessageMap(c).RequestBodyError, err)
}

func ReturnRequestJsonError(c *gin.Context, err error) {
	ReturnError(c, http.StatusOK, GetMessageMap(c).RequestJsonUnmarshalError, err)
}

func ReturnHandleError(c *gin.Context, msg string, err error) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).HandleError, msg), err)
}

func ReturnServerHandleError(c *gin.Context, err error) {
	ReturnError(c, http.StatusOK, fmt.Sprintf(GetMessageMap(c).HandleError, err.Error()), err)
}

func ReturnPasswordError(c *gin.Context) {
	ReturnError(c, http.StatusOK, GetMessageMap(c).PasswordError, nil)
}

func ReturnTokenError(c *gin.Context) {
	ReturnError(c, 401, GetMessageMap(c).TokenError, nil)
}

func ReturnDashboardNameRepeatError(c *gin.Context) {
	ReturnError(c, http.StatusOK, GetMessageMap(c).DashboardNameRepeatError, nil)
}
