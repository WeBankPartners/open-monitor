package alarm

import (
	"encoding/json"
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func GetEndpointBusinessConfig(c *gin.Context) {
	endpointId, err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	err, data := db.GetBusinessListNew(endpointId, "")
	if err != nil {
		mid.ReturnQueryTableError(c, "business_monitor", err)
	} else {
		mid.ReturnSuccessData(c, data)
	}
}

func AddEndpointBusinessConfig(c *gin.Context) {
	var param m.BusinessUpdateDto
	if err := c.ShouldBindJSON(&param); err == nil {
		for _, v := range param.PathList {
			if !mid.IsIllegalPath(v.Path) {
				mid.ReturnValidateError(c, "path illegal")
				return
			}
		}
		err = db.AddBusinessTable(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "business_monitor", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func UpdateEndpointBusinessConfig(c *gin.Context) {
	var param m.BusinessUpdateDto
	if err := c.ShouldBindJSON(&param); err == nil {
		pathMap := make(map[string]int)
		for _, v := range param.PathList {
			if !mid.IsIllegalPath(v.Path) {
				mid.ReturnValidateError(c, "path illegal")
				return
			}
			if _, b := pathMap[v.Path]; b {
				mid.ReturnValidateError(c, "path "+v.Path+" is duplicated")
				return
			} else {
				pathMap[v.Path] = 1
			}
			for _, vv := range v.Rules {
				if vv.Regular == "" {
					mid.ReturnValidateError(c, "path "+v.Path+" rules regular can not empty")
					return
				}
				if len(vv.MetricConfig) == 0 {
					mid.ReturnValidateError(c, "path "+v.Path+" rules metric_config can not empty")
					return
				}
			}
		}
		err = db.UpdateBusinessNew(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "business_monitor", err)
		} else {
			err = UpdateNodeExporterBusinessConfig(param.EndpointId)
			if err != nil {
				mid.ReturnHandleError(c, err.Error(), err)
				return
			}
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

type businessHttpDto struct {
	Paths []string `json:"paths"`
}

func UpdateNodeExporterBusinessConfig(endpointId int) error {
	err, data := db.GetBusinessListNew(endpointId, "")
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail", zap.Error(err))
		return err
	}
	endpointObj := m.EndpointTable{Id: endpointId}
	err = db.GetEndpoint(&endpointObj)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, get endpoint msg fail", zap.Error(err))
		return err
	}
	postParam := []*m.BusinessAgentDto{}
	for _, v := range data.PathList {
		postParam = append(postParam, &m.BusinessAgentDto{Path: v.Path, Config: v.Rules, Custom: v.CustomMetrics})
	}
	postData, err := json.Marshal(postParam)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, marshal post data fail", zap.Error(err))
		return err
	}
	url := fmt.Sprintf("http://%s/business/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update node_exporter fail, http post fail", zap.Error(err))
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Info(nil, log.LOGGER_APP, "", zap.String("url", url), zap.String("response", string(respBody)))
	resp.Body.Close()
	return nil
}

func PluginBusinessHandle(c *gin.Context) {
	response := m.PluginBusinessResp{ResultCode: "0", ResultMessage: "success", Results: m.PluginBusinessOutput{}}
	var err error
	defer func() {
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Plugin ci data operation handle fail", zap.Error(err))
			response.ResultCode = "1"
			response.ResultMessage = err.Error()
		}
		//bodyBytes, _ := json.Marshal(response)
		//c.Set("responseBody", string(bodyBytes))
		c.JSON(http.StatusOK, response)
	}()
	var param m.PluginBusinessRequest
	if err = c.ShouldBindJSON(&param); err != nil {
		return
	}
	if len(param.Inputs) == 0 {
		return
	}
	for _, input := range param.Inputs {
		output, endpointId, tmpErr := db.PluginBusinessAction(input)
		if tmpErr == nil {
			tmpErr = UpdateNodeExporterBusinessConfig(endpointId)
		}
		if tmpErr != nil {
			output.ErrorCode = "1"
			output.ErrorMessage = tmpErr.Error()
			err = tmpErr
		}
		response.Results.Outputs = append(response.Results.Outputs, output)
	}
	//logParam, _ := json.Marshal(param)
	//c.Set("requestBody", string(logParam))
}
