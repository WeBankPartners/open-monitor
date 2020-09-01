package middleware

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"encoding/json"
	"strings"
)

var errorMessageList []*models.ErrorMessageObj

func InitErrorMessageList()  {
	fs,err := ioutil.ReadDir("./conf/i18n")
	if err != nil {
		log.Logger.Error("Init errorMessage fail", log.Error(err))
		return
	}
	if len(fs) == 0 {
		log.Logger.Error("Init errorMessage fail, conf/i18n is empty dir")
		return
	}
	for _,v := range fs {
		tmpFileBytes,tmpErr := ioutil.ReadFile("./conf/i18n/"+v.Name())
		if tmpErr != nil {
			log.Logger.Error("Init errorMessage,read " + v.Name() + " fail", log.Error(tmpErr))
			continue
		}
		var tmpErrorMessageObj models.ErrorMessageObj
		tmpErr = json.Unmarshal(tmpFileBytes, &tmpErrorMessageObj)
		if err != nil {
			log.Logger.Error("Init errorMessage,unmarshal file " + v.Name() + " fail", log.Error(tmpErr))
			continue
		}
		tmpErrorMessageObj.Language = strings.Replace(v.Name(), ".json", "", -1)
		errorMessageList = append(errorMessageList, &tmpErrorMessageObj)
	}
	if len(errorMessageList) == 0 {
		log.Logger.Error("Init errorMessage fail, errorMessageList is empty")
	}else{
		log.Logger.Info("Init errorMessage success")
	}
}

func GetMessageMap(c *gin.Context) *models.ErrorMessageObj {
	acceptLanguage := c.GetHeader("Accept-Language")
	if len(errorMessageList) == 0 {
		return &models.ErrorMessageObj{}
	}
	if acceptLanguage != "" {
		acceptLanguage = strings.Replace(acceptLanguage, ";", ",", -1)
		for _, v := range strings.Split(acceptLanguage, ",") {
			if strings.HasPrefix(v, "q=") {
				continue
			}
			lowerV := strings.ToLower(v)
			for _, vv := range errorMessageList {
				if vv.Language == lowerV {
					return vv
				}
			}
		}
	}
	for _,v := range errorMessageList {
		if v.Language == models.Config().Http.DefaultLanguage {
			return v
		}
	}
	return errorMessageList[0]
}
