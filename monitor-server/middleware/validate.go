package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"reflect"
	"fmt"
)

func ValidateGet(c *gin.Context)  {
	isOk := true
	if c.Request.Method == "GET" {
		index := strings.Index(c.Request.URL.String(), "?")
		if index > 0 {
			invalidData := []string{"select", "insert", "update", "alter", "delete", "drop", "truncate", "show"}
			url := c.Request.URL.String()
			params := strings.ToLower(url[index:])
			for _,inv := range invalidData {
				if strings.Contains(params, inv) {
					isOk = false
					break
				}
			}
		}
	}
	if isOk {
		c.Next()
	}else{
		Return(c, RespJson{Msg:"request validate fail", Code:http.StatusBadRequest})
		c.Abort()
		return
	}
}

func ValidatePost(c *gin.Context, obj interface{}, ex ...string) bool {
	isOk := true
	invalidData := []string{"select", "insert", "update", "alter", "delete", "drop", "truncate", "show"}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for k:=0;k < t.NumField();k++ {
		isEx := false
		for _,key := range ex {
			if key == t.Field(k).Name {
				isEx = true
				break
			}
		}
		if isEx {
			continue
		}
		fieldType := fmt.Sprintf("%v", reflect.TypeOf(v.Field(k).Interface()))
		if fieldType == "string" || fieldType == "[]string" {
			for _,inv := range invalidData {
				if strings.Contains(fmt.Sprintf("%v", v.Field(k).Interface()), inv) {
					isOk = false
					break
				}
			}
		}
		if !isOk {
			break
		}
	}
	if !isOk {
		Return(c, RespJson{Msg:"request validate fail", Code:http.StatusBadRequest})
		c.Abort()
	}
	return isOk
}

func IsIllegalName(str string) bool {
	re := false
	if len(str) > 50 {
		re = true
	}
	if strings.TrimSpace(str)=="" {
		re = true
	}
	return re
}