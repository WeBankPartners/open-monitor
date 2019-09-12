package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
	"net/http"
	"fmt"
)

const (
	ok = "success"
	serverError = "fail"
	serverErrorCn = "失败"
)

type RespJson struct {
	Code  int  `json:"code"`
	Msg   string    `json:"msg"`
	Data  interface{}  `json:"data"`
}

func Return(c *gin.Context, j RespJson)  {
	if j.Code <= 0{
		if strings.Contains(j.Msg, serverError) || strings.Contains(j.Msg, serverErrorCn) {
			j.Code = http.StatusInternalServerError
		}else{
			j.Code = http.StatusOK
		}
	}
	c.JSON(j.Code, gin.H{"code": j.Code, "msg": j.Msg, "data": j.Data})
}

func ReturnError(c *gin.Context, msg string, err error) {
	LogError(fmt.Sprintf("request %s fail", c.FullPath()), err)
	c.JSON(http.StatusInternalServerError, gin.H{"msg": msg, "data": err})
}

func ReturnSuccess(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, RespJson{Msg:msg})
}

func ReturnData(c *gin.Context, data interface{}) {
	if fmt.Sprintf("%v", data) == `[]` {
		data = []string{}
	}
	c.JSON(http.StatusOK, data)
}

func ReturnValidateFail(c *gin.Context, msg string)  {
	c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
}