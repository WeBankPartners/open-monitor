package agent

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"fmt"
	"net/http"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

type installRequestObj struct {
	RequestId  string  	`json:"requestId"`
	Inputs  []installObj  `json:"inputs"`
}

type installObj struct {
	CallbackParameter  string  `json:"callbackParameter"`
	Host  string  `json:"host"`
	Port  string  `json:"port"`
	User  string  `json:"user"`
	Password  string  `json:"password"`
}

func InstallAgent(c *gin.Context)  {
	var result resultObj
	name := c.Param("name")
	var exporterBin string
	illegal := true
	for _,v := range m.Config().Agent {
		if v.AgentType == name {
			exporterBin = v.AgentBin
			illegal = false
			break
		}
	}
	if illegal {
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("No such monitor type like %s", name)}
		log.Logger.Warn(result.ResultMessage)
		c.JSON(http.StatusBadRequest, result)
		return
	}
	requestBody,_ := ioutil.ReadAll(c.Request.Body)
	log.Logger.Info("request param", log.String("param", string(requestBody)))
	var param installRequestObj
	err := json.Unmarshal(requestBody, &param)
	if name != "host" && err != nil {
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("Param unmarshal fail : %v", err)}
		log.Logger.Error("Param unmarshal fail", log.Error(err))
		c.JSON(http.StatusBadRequest, result)
		return
	}
	data,err := ioutil.ReadFile(fmt.Sprintf("conf/agent/install_%s.sh", name))
	if err != nil {
		result = resultObj{ResultCode:"1", ResultMessage:fmt.Sprintf("Read install file fail : %v", err)}
		log.Logger.Error("Read install file fail", log.Error(err))
		c.JSON(http.StatusBadRequest, result)
		return
	}
	dataString := strings.Replace(string(data), "{{server_address}}", c.Request.Host, -1)
	dataString = strings.Replace(dataString, "{{exporter_type}}", name, -1)
	dataString = strings.Replace(dataString, "{{bin_name}}", exporterBin, -1)
	if name != "host" {
		if len(param.Inputs) == 0 {
			result = resultObj{ResultCode:"1", ResultMessage:"Param inputs is null"}
			log.Logger.Warn(result.ResultMessage)
			c.JSON(http.StatusBadRequest, result)
			return
		}
		dataString = strings.Replace(dataString, "{{mysql_host}}", param.Inputs[0].Host, -1)
		dataString = strings.Replace(dataString, "{{mysql_port}}", param.Inputs[0].Port, -1)
		dataString = strings.Replace(dataString, "{{redis_host}}", param.Inputs[0].Host, -1)
		dataString = strings.Replace(dataString, "{{redis_port}}", param.Inputs[0].Port, -1)
		dataString = strings.Replace(dataString, "{{redis_pwd}}", param.Inputs[0].Password, -1)
	}
	c.Data(http.StatusOK, "application/octet-stream", []byte(dataString))
}
