package prom

import (
	"fmt"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"net/http"
	"strings"
	"io/ioutil"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"sync"
	"time"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

var (
	AgentManagerInitFlag = false
	AgentManagerLock = new(sync.RWMutex)
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
	if !AgentManagerInitFlag {
		time.Sleep(1*time.Second)
		AgentManagerLock.RLock()
		AgentManagerLock.RUnlock()
	}
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

func InitAgentManager(url string) {
	param, err := db.GetAgentManager("")
	if err != nil {
		log.Logger.Error("Get agent manager table fail", log.Error(err))
		return
	}
	count := 0
	AgentManagerLock.Lock()
	for {
		time.Sleep(30*time.Second)
		resp, err := requestAgentMonitor(param, url, "init")
		if err != nil {
			log.Logger.Error("Init agent manager, request error", log.Error(err))
		}
		if resp.Code == 200 {
			log.Logger.Info("Init agent manager success")
			break
		}else{
			log.Logger.Warn("Init agent manager, response error", log.String("message", resp.Message))
		}
		count++
		if count >= 10 {
			log.Logger.Warn("Init agent manager fail, retry max time")
			break
		}
	}
	AgentManagerLock.Unlock()
	AgentManagerInitFlag = true
}

func StartSyncAgentManagerJob(url string)  {
	intervalSecond := 86400
	timeStartValue, _ := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02")))
	time.Sleep(time.Duration(timeStartValue.Unix()+86400-time.Now().Unix()) * time.Second)
	t := time.NewTicker(time.Duration(intervalSecond) * time.Second).C
	for {
		go doSyncAgentManagerJob(url)
		<- t
	}
}

func doSyncAgentManagerJob(url string)  {
	log.Logger.Info("Start init agent manager ")
	param, _ := db.GetAgentManager("")
	resp, err := requestAgentMonitor(param, url, "init")
	if err != nil {
		log.Logger.Error("Init agent manager, request error", log.Error(err))
	}
	if resp.Code == 200 {
		log.Logger.Info("Init agent manager success")
	}else{
		log.Logger.Warn("Init agent manager, response error", log.String("message", resp.Message))
	}
}

func requestAgentMonitor(param interface{},url,method string) (resp agentManagerResponse,err error) {
	postData,err := json.Marshal(param)
	if err != nil {
		log.Logger.Error("Failed marshalling data", log.Error(err))
		return resp,err
	}
	req,err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/deploy/%s", url, method), strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Curl agent_monitor http request error", log.Error(err))
		return resp,err
	}
	res,err := http.DefaultClient.Do(req)
	if err != nil {
		log.Logger.Error("Curl agent_monitor http response error", log.Error(err))
		return resp,err
	}
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	log.Logger.Debug(fmt.Sprintf("Curl %s agent_monitor response : %s ", method, string(body)))
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Logger.Error("Curl agent_monitor unmarshal error", log.Error(err))
		return resp,err
	}
	return resp,nil
}