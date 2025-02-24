package config_new

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func SnmpExporterList(c *gin.Context) {
	result, err := db.SnmpExporterList()
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccessData(c, result)
	}
}

func SnmpExporterCreate(c *gin.Context) {
	var param models.SnmpExporterTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if !middleware.IsIllegalNormalInput(param.Id) {
		middleware.ReturnValidateError(c, "Param id is illegal")
		return
	}
	err := db.SnmpExporterCreate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func SnmpExporterUpdate(c *gin.Context) {
	var param models.SnmpExporterTable
	if err := c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	err := db.SnmpExporterUpdate(param)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func SnmpExporterDelete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	err := db.SnmpExporterDelete(id)
	if err != nil {
		middleware.ReturnHandleError(c, err.Error(), err)
	} else {
		middleware.ReturnSuccess(c)
	}
}

func PluginSnmpExporterHandle(c *gin.Context) {
	action := c.Param("action")
	response := models.PluginSnmpExporterResp{ResultCode: "0", ResultMessage: "success", Results: models.PluginSnmpExporterOutput{}}
	var err error
	defer func() {
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Plugin snmp exporter handle fail", zap.Error(err))
			response.ResultCode = "1"
			response.ResultMessage = err.Error()
		}
		c.JSON(http.StatusOK, response)
	}()
	var param models.PluginSnmpExporterRequest
	c.ShouldBindJSON(&param)
	if len(param.Inputs) == 0 {
		return
	}
	if action != "add" && action != "delete" {
		err = fmt.Errorf("Url action is illegal ")
		return
	}
	nowSnmpList, queryErr := db.SnmpExporterList()
	if queryErr != nil {
		err = fmt.Errorf("Try to query snmp list fail,%s ", queryErr.Error())
		return
	}
	for _, input := range param.Inputs {
		output, tmpErr := PluginSnmpExporter(input, action, nowSnmpList)
		if tmpErr != nil {
			output.ErrorCode = "1"
			output.ErrorMessage = tmpErr.Error()
			err = tmpErr
		}
		response.Results.Outputs = append(response.Results.Outputs, output)
	}
}

func PluginSnmpExporter(input *models.PluginSnmpExporterRequestObj, action string, nowSnmpList []*models.SnmpExporterTable) (result *models.PluginSnmpExporterOutputObj, err error) {
	result = &models.PluginSnmpExporterOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", Id: input.Id}
	if input.Id == "" {
		err = fmt.Errorf("Param id can not empty ")
		return
	}
	if action == "add" {
		if input.Address == "" {
			err = fmt.Errorf("Param address can not empty ")
			return
		}
		defaultInterval := 10
		if input.ScrapeInterval != "" {
			defaultInterval, err = strconv.Atoi(input.ScrapeInterval)
			if err != nil {
				err = fmt.Errorf("Param scrapeInterval can not parse to int ")
				return
			}
		}
		param := models.SnmpExporterTable{Id: input.Id, Address: input.Address, ScrapeInterval: defaultInterval}
		existFlag := 0
		for _, nowData := range nowSnmpList {
			if nowData.Id == input.Id {
				existFlag = 1
				if nowData.Address == input.Address && nowData.ScrapeInterval == defaultInterval {
					existFlag = 2
				}
				break
			}
		}
		if existFlag == 2 {
			return
		}
		if existFlag == 1 {
			err = db.SnmpExporterUpdate(param)
		} else {
			err = db.SnmpExporterCreate(param)
		}
	} else {
		existFlag := false
		for _, nowData := range nowSnmpList {
			if nowData.Id == input.Id {
				existFlag = true
				break
			}
		}
		if existFlag {
			err = db.SnmpExporterDelete(input.Id)
		}
	}
	return
}
