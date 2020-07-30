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
		mid.ReturnValidateFail(c, fmt.Sprintf("Param id validate fail %v", err))
		return
	}
	err,data := db.GetProcessList(endpointId)
	if err != nil {
		mid.ReturnError(c, "Get process list fail", err)
	}else{
		mid.ReturnData(c, data)
	}
}

func UpdateEndpointProcessConfig(c *gin.Context)  {
	var param m.ProcessUpdateDto
	if err := c.ShouldBindJSON(&param); err==nil {
		err = db.UpdateProcess(param)
		if err != nil {
			mid.ReturnError(c, "Update process fail ", err)
		}else{
			err = updateNodeExporterProcessConfig(param.EndpointId)
			if err != nil {
				mid.ReturnError(c, "Update node exporter config fail ", err)
				return
			}
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Param validate fail %v \n", err))
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
