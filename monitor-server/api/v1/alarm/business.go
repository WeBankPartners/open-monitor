package alarm

import (
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/gin-gonic/gin"
	"strconv"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"encoding/json"
	"net/http"
	"strings"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"io/ioutil"
)

func GetEndpointBusinessConfig(c *gin.Context)  {
	endpointId,err := strconv.Atoi(c.Query("id"))
	if err != nil || endpointId <= 0 {
		mid.ReturnValidateFail(c, fmt.Sprintf("Param id validate fail %v", err))
		return
	}
	err,data := db.GetBusinessList(endpointId, "")
	if err != nil {
		mid.ReturnError(c, "Get business list fail", err)
	}else{
		mid.ReturnData(c, data)
	}
}

func UpdateEndpointBusinessConfig(c *gin.Context)  {
	var param m.BusinessUpdateDto
	if err := c.ShouldBindJSON(&param); err==nil {
		for _,v := range param.PathList {
			if !mid.IsIllegalPath(v.Path) {
				mid.ReturnValidateFail(c, "Parameter validate fail, path illegal")
				return
			}
		}
		err = db.UpdateBusiness(param)
		if err != nil {
			mid.ReturnError(c, "Update business fail ", err)
		}else{
			err = UpdateNodeExporterBusinessConfig(param.EndpointId)
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

type businessHttpDto struct {
	Paths  []string  `json:"paths"`
}

func UpdateNodeExporterBusinessConfig(endpointId int) error {
	err,data := db.GetBusinessList(endpointId, "")
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
	postParam := businessHttpDto{Paths:[]string{}}
	for _,v := range data {
		postParam.Paths = append(postParam.Paths, v.Path)
	}
	postData,err := json.Marshal(postParam)
	if err != nil {
		log.Logger.Error("Update node_exporter fail, marshal post data fail", log.Error(err))
		return err
	}
	url := fmt.Sprintf("http://%s/business/config", endpointObj.Address)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Update node_exporter fail, http post fail", log.Error(err))
		return err
	}
	respBody,_ := ioutil.ReadAll(resp.Body)
	log.Logger.Info("", log.String("url", url), log.String("response", string(respBody)))
	resp.Body.Close()
	return nil
}
