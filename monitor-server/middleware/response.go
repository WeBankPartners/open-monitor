package middleware

import (
	"errors"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

func ReturnError(c *gin.Context, err error, statusCode int) {
	errorCode, errorKey, errorMessage := models.GetErrorResult(c.GetHeader("Accept-Language"), err, -1)
	if !models.IsBusinessErrorCode(errorCode) {
		log.Logger.Error("systemError", log.String("url", c.FullPath()), log.Int("errorCode", errorCode), log.String("errorKey", errorKey), log.String("message", errorMessage), log.Error(err))
	} else {
		log.Logger.Error("businessError", log.String("url", c.FullPath()), log.Int("errorCode", errorCode), log.String("errorKey", errorKey), log.String("message", errorMessage), log.Error(err))
	}
	c.Writer.Header().Add("Error-Code", strconv.Itoa(errorCode))
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	c.JSON(statusCode, RespJson{Status: "ERROR", Code: errorCode, Message: errorMessage})
}

func ReturnSuccessWithMessage(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: msg})
}

func ReturnSuccess(c *gin.Context) {
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: models.GetMessageMap(c).Success})
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
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: models.GetMessageMap(c).Success, Data: data})
}

func ReturnPageData(c *gin.Context, pageInfo models.PageInfo, data interface{}) {
	c.JSON(http.StatusOK, RespJson{Status: "OK", Code: 200, Message: models.GetMessageMap(c).Success, Data: ResponsePageData{PageInfo: pageInfo, Contents: data}})
}

func ReturnValidateError(c *gin.Context, msg string) {
	ReturnError(c, models.GetMessageMap(c).ParamValidateError.WithParam(msg), http.StatusOK)
}

func ReturnParamTypeError(c *gin.Context, paramName, typeName string) {
	ReturnError(c, models.GetMessageMap(c).ParamTypeError.WithParam(paramName, typeName), http.StatusOK)
}

func ReturnParamEmptyError(c *gin.Context, key string) {
	ReturnError(c, models.GetMessageMap(c).ParamEmptyError.WithParam(key), http.StatusOK)
}

func ReturnQueryTableError(c *gin.Context, table string, err error) {
	errMsg := fmt.Errorf("%s,%s", table, err.Error())
	ReturnError(c, models.GetMessageMap(c).QueryTableError.WithParam(errMsg), http.StatusOK)
}

func ReturnUpdateTableError(c *gin.Context, table string, err error) {
	ReturnError(c, models.GetMessageMap(c).UpdateTableError.WithParam(table+" detail: "+err.Error()), http.StatusOK)
}

func ReturnFetchDataError(c *gin.Context, table, key, value string) {
	ReturnError(c, models.GetMessageMap(c).FetchTableDataError.WithParam(table, key, value), http.StatusOK)
}

func ReturnDeleteTableError(c *gin.Context, table, key, value string) {
	ReturnError(c, models.GetMessageMap(c).DeleteTableDataError.WithParam(table, key, value), http.StatusOK)
}

func ReturnBodyError(c *gin.Context, err error) {
	if err != nil {
		var customErr models.CustomError
		if errors.As(err, &customErr) {
			ReturnError(c, customErr, http.StatusOK)
		} else {
			ReturnError(c, models.GetMessageMap(c).RequestBodyError, http.StatusOK)
		}
		return
	}
	ReturnError(c, models.GetMessageMap(c).RequestBodyError, http.StatusOK)
}

func ReturnRequestJsonError(c *gin.Context, err error) {
	if err != nil {
		var customErr models.CustomError
		if errors.As(err, &customErr) {
			ReturnError(c, customErr, http.StatusOK)
		} else {
			ReturnError(c, models.GetMessageMap(c).RequestJsonUnmarshalError, http.StatusOK)
		}
		return
	}
	ReturnError(c, models.GetMessageMap(c).RequestJsonUnmarshalError, http.StatusOK)
}

func ReturnHandleError(c *gin.Context, msg string, err error) {
	if err != nil {
		var customErr models.CustomError
		if errors.As(err, &customErr) {
			ReturnError(c, customErr.WithParam(msg), http.StatusOK)
		} else {
			ReturnError(c, models.GetMessageMap(c).HandleError.WithParam(msg), http.StatusOK)
		}
		return
	}
	ReturnError(c, models.GetMessageMap(c).HandleError.WithParam(msg), http.StatusOK)
}

func ReturnServerHandleError(c *gin.Context, err error) {
	if err != nil {
		var customErr models.CustomError
		if errors.As(err, &customErr) {
			ReturnError(c, customErr, http.StatusOK)
		} else {
			ReturnError(c, models.GetMessageMap(c).HandleError, http.StatusOK)
		}
		return
	}
	ReturnError(c, models.GetMessageMap(c).HandleError.WithParam(err.Error()), http.StatusOK)
}

func ReturnPasswordError(c *gin.Context) {
	ReturnError(c, models.GetMessageMap(c).PasswordError, http.StatusOK)
}

func ReturnTokenError(c *gin.Context) {
	ReturnError(c, models.GetMessageMap(c).TokenError, http.StatusUnauthorized)
}

func ReturnDashboardNameRepeatError(c *gin.Context) {
	ReturnError(c, models.GetMessageMap(c).DashboardNameRepeatError, http.StatusOK)
}

func InitHttpError() {
	err := models.InitErrorTemplateList("./conf/i18n")
	if err != nil {
		log.Logger.Error("Init error template list fail", log.Error(err))
	}
}
