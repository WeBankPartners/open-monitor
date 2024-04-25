package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var (
	invalidData      = []string{"select", "insert", "update", "alter", "delete", "drop", "truncate", "show"}
	regCond          = regexp.MustCompile(`^([<=|>=|!=|==|<|>]*)-?\d+(\.\d+)?$`)
	regLast          = regexp.MustCompile(`^\d+[s|m|h]$`)
	regPath          = regexp.MustCompile(`^\/([\w|\.|\-]+\/?)+$`)
	regNormal        = regexp.MustCompile(`^[\w|\.|\-|\~|\!|\@|\#|\$|\%|\^|\[|\]|\{|\}|\(|\)|\,|\s]+$`)
	regIp            = regexp.MustCompile(`^((25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))\.){3}(25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))$`)
	regActiveWindow  = regexp.MustCompile(`^\d{2}:\d{2}-\d{2}:\d{2}$`)
	regName          = regexp.MustCompile(`^[\w|\-|\.|:]+$`)
	regMetric        = regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`)
	regDisplayName   = regexp.MustCompile(`.*`)
	roleEndpointMap  []map[string]int
	roleEndpointLock sync.RWMutex
)

func ValidateGet(c *gin.Context) {
	isOk := true
	if c.Request.Method == "GET" {
		index := strings.Index(c.Request.URL.String(), "?")
		if index > 0 {
			url := c.Request.URL.String()
			params := strings.ToLower(url[index:])
			for _, inv := range invalidData {
				if strings.Contains(params, inv) {
					isOk = false
					break
				}
			}
		}
	}
	if isOk {
		c.Next()
	} else {
		ReturnError(c, http.StatusBadRequest, "request get param validate fail", nil)
		c.Abort()
		return
	}
}

func ValidatePost(c *gin.Context, obj interface{}, ex ...string) bool {
	isOk := true
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for k := 0; k < t.NumField(); k++ {
		isEx := false
		for _, key := range ex {
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
			for _, inv := range invalidData {
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
		ReturnError(c, http.StatusBadRequest, "request post param validate fail", nil)
		c.Abort()
	}
	return isOk
}

func IsIllegalCond(str string) bool {
	return regCond.MatchString(str)
}

func IsIllegalLast(str string) bool {
	return regLast.MatchString(str)
}

func IsIllegalPath(str string) bool {
	return regPath.MatchString(str)
}

func IsIllegalNormalInput(str string) bool {
	return regNormal.MatchString(str)
}

func IsIllegalIp(str string) bool {
	return !regIp.MatchString(str)
}

func IsIllegalName(str string) bool {
	if str == "" {
		return true
	}
	return false
	//if len(str) > 50 {
	//	return true
	//}
	//return !regName.MatchString(str)
}

func IsIllegalNameNew(str string) bool {
	if str == "" || len(str) > 50 {
		return true
	}
	return !regName.MatchString(str)
}

func IsIllegalDisplayName(str string) bool {
	if str == "" || len(str) > 250 {
		return true
	}
	return !regDisplayName.MatchString(str)
}

func IsIllegalMetric(str string) bool {
	if str == "" || len(str) > 50 {
		return true
	}
	return !regMetric.MatchString(str)
}

func ValidateActiveWindowString(str string) bool {
	return regActiveWindow.MatchString(str)
}

func InitRoleEndpointMap() {

}

func CheckRoleEndpointOwner(roles []string) {

}

func UpdateRoleEndpointMap() {

}
