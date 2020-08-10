package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
	m "github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/models"
	"time"
)

type RespJson struct {
	Code  int  `json:"code"`
	Msg  string  `json:"msg"`
}

func ReturnMessage(c *gin.Context, resp RespJson)  {
	var statusCode int
	if resp.Code == 0 {
		statusCode = http.StatusOK
	}else if resp.Code == 1 {
		statusCode = http.StatusBadRequest
	}else{
		statusCode = http.StatusInternalServerError
	}
	c.JSON(statusCode, m.TransResult{ResultCode:resp.Code, ResultMsg:resp.Msg, SystemTime:time.Now().Format(m.DatetimeFormat)})
}
