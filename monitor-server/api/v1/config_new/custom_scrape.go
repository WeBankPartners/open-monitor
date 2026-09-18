package config_new

import (
	"fmt"
	"strings"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
)

func CustomScrapeConfigList(c *gin.Context) {
	result, err := db.CustomScrapeConfigList()
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func CustomScrapeConfigCheck(c *gin.Context) {
	var param models.CustomScrapeConfigParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := validateCustomScrapeParam(param, false); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	result, err := db.CustomScrapeConfigCheck(param)
	if err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	middleware.ReturnSuccessData(c, result)
}

func CustomScrapeConfigCreate(c *gin.Context) {
	var param models.CustomScrapeConfigParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := validateCustomScrapeParam(param, false); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := db.CustomScrapeConfigCreate(param, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	middleware.ReturnSuccess(c)
}

func CustomScrapeConfigUpdate(c *gin.Context) {
	var param models.CustomScrapeConfigParam
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := validateCustomScrapeParam(param, true); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := db.CustomScrapeConfigUpdate(param, middleware.GetOperateUser(c)); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	middleware.ReturnSuccess(c)
}

func CustomScrapeConfigDelete(c *gin.Context) {
	id := c.Query("guid")
	if id == "" {
		middleware.ReturnParamEmptyError(c, "guid")
		return
	}
	if err := db.CustomScrapeConfigDelete(id); err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
		return
	}
	middleware.ReturnSuccess(c)
}

func validateCustomScrapeParam(param models.CustomScrapeConfigParam, needGuid bool) error {
	if needGuid && param.Guid == "" {
		return fmt.Errorf("guid is empty")
	}
	name := strings.TrimSpace(param.Name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 64 {
		return fmt.Errorf("name is too long")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("name is illegal")
	}
	if strings.TrimSpace(param.YamlContent) == "" {
		return fmt.Errorf("yaml_content is required")
	}
	return nil
}
