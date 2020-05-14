package prom

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strings"
	"io/ioutil"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
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

func DeployAgent(agentType,instance,bin,ip,port,user,pwd,url,configFile string) (address string,err error) {
	var param agentManagerRequest
	param.Guid = fmt.Sprintf("%s_%s_%s", instance, ip, agentType)
	param.Exporter = bin
	param.Config = configFile
	param.InstanceServer = ip
	param.InstancePort = port
	param.AuthUser = user
	param.AuthPassword = pwd
	resp,err := requestAgentMonitor(param,url,"add")
	if err != nil {
		return address,err
	}
	if resp.Code == 200 {
		if strings.Contains(resp.Message, ":") {
			tmpAddress := resp.Message
			if strings.Contains(url, "127.0.0.1") {
				tmpAddress = "127.0.0.1" + tmpAddress[strings.Index(tmpAddress, ":"):]
			}
			return tmpAddress,nil
		}else{
			return "", fmt.Errorf("agent manager response message is illegal address: %s ", resp.Message)
		}
	}else{
		return address,fmt.Errorf(resp.Message)
	}
}

func StopAgent(agentType,instance,ip,url string) error {
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

func InitAgentManager(param []*m.AgentManagerTable, url string) {
	count := 0
	for {
		time.Sleep(3*time.Second)
		resp, err := requestAgentMonitor(param, url, "init")
		if err != nil {
			mid.LogError("init agent manager, request error -> ", err)
		}
		if resp.Code == 200 {
			mid.LogInfo("init agent manager success")
			break
		}else{
			mid.LogError("init agent manager, response error -> ", fmt.Errorf(resp.Message))
		}
		count++
		if count >= 10 {
			mid.LogError("init agent manager fail, retry max time ", nil)
			break
		}
	}
}

func requestAgentMonitor(param interface{},url,method string) (resp agentManagerResponse,err error) {
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
	res,err := http.DefaultClient.Do(req)
	if err != nil {
		mid.LogError("curl agent_monitor http response error ", err)
		return resp,err
	}
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	mid.LogInfo(fmt.Sprintf("curl %s agent_monitor response : %s ", method, string(body)))
	err = json.Unmarshal(body, &resp)
	if err != nil {
		mid.LogError("curl agent_monitor unmarshal error ", err)
		return resp,err
	}
	return resp,nil
}