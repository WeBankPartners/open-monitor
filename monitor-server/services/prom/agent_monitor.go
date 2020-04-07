package prom

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strings"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"context"
)

type agentManagerRequest struct {
	Guid  string  `json:"guid"`
	Exporter  string  `json:"exporter"`
	Config  string  `json:"config"`
	InstanceServer  string  `json:"instance_server"`
	InstancePort  string  `json:"instance_port"`
	AuthUser  string  `json:"auth_user"`
	AuthPassword  string  `json:"auth_password"`
}

type agentManagerResponse struct {
	Code  int  `json:"code"`
	Message  string  `json:"message"`
	Data  interface{}  `json:"data"`
}

func DeployAgent(agentType,instance,bin,ip,port,user,pwd,url string) (address string,err error) {
	var param agentManagerRequest
	param.Guid = fmt.Sprintf("%s_%s_%s", instance, ip, agentType)
	param.Exporter = bin
	if agentType == "mysql" {
		param.Config = "my.cnf"
	}
	if agentType == "tomcat" {
		param.Config = "config.yaml"
	}
	param.InstanceServer = ip
	param.InstancePort = port
	param.AuthUser = user
	param.AuthPassword = pwd
	resp,err := requestAgentMonitor(param,url,"add")
	if err != nil {
		return address,err
	}
	if resp.Code == 200 {
		return resp.Message,nil
	}else{
		return address,fmt.Errorf(resp.Message)
	}
}

func StopAgent(agentType,instance,ip,url string) error {
	if agentType == "java" {
		agentType = "tomcat"
	}
	var param agentManagerRequest
	param.Guid = fmt.Sprintf("%s_%s_%s", instance, ip, agentType)
	resp,err := requestAgentMonitor(param,url,"delete")
	if err != nil {
		return err
	}
	if resp.Code == 200 {
		return nil
	}else{
		return fmt.Errorf(resp.Message)
	}
}

func requestAgentMonitor(param agentManagerRequest,url,method string) (resp agentManagerResponse,err error) {
	postData,err := json.Marshal(param)
	if err != nil {
		mid.LogError("Failed marshalling data", err)
		return resp,err
	}
	req,err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/deploy/%s", url, method), strings.NewReader(string(postData)))
	if err != nil {
		mid.LogError("curl agent_monitor http request error ", err)
		return resp,err
	}
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		mid.LogError("curl agent_monitor http response error ", err)
		return resp,err
	}
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	mid.LogInfo(fmt.Sprintf("guid: %s, curl %s agent_monitor response : %s ", param.Guid, method, string(body)))
	err = json.Unmarshal(body, &resp)
	if err != nil {
		mid.LogError("curl agent_monitor unmarshal error ", err)
		return resp,err
	}
	return resp,nil
}