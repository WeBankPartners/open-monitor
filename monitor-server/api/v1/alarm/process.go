package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"encoding/json"
	"strings"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"io/ioutil"
)

func GetEndpointProcessConfig(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	err,data := db.GetProcessList(endpointId)
	if err != nil {
		mid.ReturnFetchDataError(c, "process_monitor", "endpoint_id", strconv.Itoa(endpointId))
	}else{
		mid.ReturnSuccessData(c, data)
	}
}

func UpdateEndpointProcessConfig(c *gin.Context)  {
	var param m.ProcessUpdateDto
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateProcess(param)
		if err != nil {
			mid.ReturnUpdateTableError(c, "process_monitor", err)
		}else{
			err = updateNodeExporterProcessConfig(param.EndpointId)
			if err != nil {
				mid.ReturnHandleError(c, "update node exporter config fail ", err)
				return
			}
			mid.ReturnSuccess(c)
		}
	}else{
		mid.ReturnValidateError(c, err.Error())
	}
}

type processHttpDto struct {
	Process  []string  `json:"process"`
}

func updateNodeExporterProcessConfig(endpointId int) error {
	err,data := db.GetProcessList(endpointId)
	if err != nil {
		log.Logger.Error("Update node_exporter fail", log.Error(err))
		return err
	}
	endpointObj := m.EndpointTable{Id:endpointId}
	err = db.GetEndpoint(&endpointObj)
	if err != nil {
		log.Logger.Error("Update node_exporter fail, get endpoint msg fail", log.Error(err))
		return err
	}
	postParam := processHttpDto{Process:[]string{}}
	for _,v := range data {
		postParam.Process = append(postParam.Process, v.Name)
	}
	postData,err := json.Marshal(postParam)
	if err != nil {
		log.Logger.Error("Update node_exporter fail, marshal post data fail", log.Error(err))
		return err
	}
	url := fmt.Sprintf("http://%s/process/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Update node_exporter fail, http post fail", log.Error(err))
		return err
	}
	responseBody,_ := ioutil.ReadAll(resp.Body)
	log.Logger.Info("curl "+url, log.String("response", string(responseBody)))
	resp.Body.Close()
	return nil
}
