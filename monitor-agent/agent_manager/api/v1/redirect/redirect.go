package redirect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/funcs"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type redirectHttpObj struct {
	Server               *http.Server
	ListenPort           int    `json:"listenPort"`
	RemoteAgentManagerIp string `json:"remoteAgentManagerIp"`
	RemoteAgentPort      int    `json:"remoteAgentPort"`
}

func (h *redirectHttpObj) Init() {
	log.Printf("start remoteHttp init,remoteIP: %s , port:%d \n", h.RemoteAgentManagerIp, h.ListenPort)
	h.Server = &http.Server{Addr: fmt.Sprintf(":%d", h.ListenPort), Handler: http.HandlerFunc(h.httpHandleFunc)}
	go h.Server.ListenAndServe()
}

func (h *redirectHttpObj) Destroy() {
	if h.Server != nil {
		err := h.Server.Shutdown(context.Background())
		if err != nil {
			log.Printf("listen server shutdown err:%s \n", err.Error())
		}
	} else {
		log.Printf("listen server %d already shutdown \n", h.ListenPort)
	}
}

func (h *redirectHttpObj) httpHandleFunc(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d%s", h.RemoteAgentManagerIp, h.RemoteAgentPort, r.RequestURI))
	if err != nil {
		log.Printf("reqeust error: %s \n", err.Error())
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("response code:%d data:%s \n", resp.StatusCode, string(b))
		return
	}
	w.Write(b)
}

var redirectHttpMap = new(sync.Map)

func Init(param *funcs.InitDeployParam) (err error) {
	newParam := funcs.InitDeployParam{}
	for _, v := range param.Config {
		newConfigObj := funcs.AgentManagerTable{
			EndpointGuid:    v.EndpointGuid,
			Name:            v.Name,
			User:            v.User,
			Password:        v.Password,
			InstanceAddress: v.InstanceAddress,
			AgentAddress:    v.AgentAddress,
			ConfigFile:      v.ConfigFile,
			BinPath:         v.BinPath,
			AgentRemotePort: v.AgentRemotePort,
		}
		if remoteAgentPort, _ := strconv.Atoi(v.AgentRemotePort); remoteAgentPort > 0 {
			newConfigObj.AgentAddress = fmt.Sprintf("%s:%d", strings.Split(v.AgentAddress, ":")[0], remoteAgentPort)
		}
		newParam.Config = append(newParam.Config, &newConfigObj)
	}
	_, err = requestAgentMonitor(&newParam, fmt.Sprintf("http://%s", param.AgentManagerRemoteIp), "init")
	for _, v := range param.Config {
		if splitIndex := strings.LastIndex(v.AgentAddress, ":"); splitIndex > 0 {
			if addressPort, _ := strconv.Atoi(v.AgentAddress[splitIndex+1:]); addressPort > 0 {
				log.Printf("port: %d \n", addressPort)
				if existRh, ok := redirectHttpMap.Load(v.EndpointGuid); ok {
					rhObj := existRh.(*redirectHttpObj)
					if rhObj != nil {
						rhObj.Destroy()
						redirectHttpMap.Delete(v.EndpointGuid)
					}
				}
				remoteAgentPort, _ := strconv.Atoi(v.AgentRemotePort)
				if remoteAgentPort == 0 {
					remoteAgentPort = addressPort
				}
				rhObj := redirectHttpObj{ListenPort: addressPort, RemoteAgentManagerIp: param.AgentManagerRemoteIp, RemoteAgentPort: remoteAgentPort}
				rhObj.Init()
				redirectHttpMap.Store(v.EndpointGuid, &rhObj)
				log.Printf("start listen port:%d \n", addressPort)
			}
		}
	}
	return
}

func Add(remoteAddress string, param map[string]string) (port int, err error) {
	param["agentManagerRemoteIp"] = ""
	resp, respErr := requestAgentMonitor(param, fmt.Sprintf("http://%s", remoteAddress), "add")
	if respErr != nil {
		err = respErr
		return
	}
	if resp.Code != 200 {
		err = fmt.Errorf(resp.Message)
		return
	}
	if splitIndex := strings.LastIndex(resp.Message, ":"); splitIndex > 0 {
		port, _ = strconv.Atoi(resp.Message[splitIndex+1:])
	}
	if port > 0 {
		endpointGuid := param["guid"]
		if existRh, ok := redirectHttpMap.Load(endpointGuid); ok {
			rhObj := existRh.(*redirectHttpObj)
			if rhObj != nil {
				rhObj.Destroy()
				redirectHttpMap.Delete(endpointGuid)
			}
		}
		rhObj := redirectHttpObj{ListenPort: port, RemoteAgentManagerIp: remoteAddress, RemoteAgentPort: port}
		rhObj.Init()
		redirectHttpMap.Store(endpointGuid, &rhObj)
		log.Printf("start listen port:%d \n", port)
	}
	return
}

func Delete(remoteAddress string, param map[string]string) (err error) {
	param["agentManagerRemoteIp"] = ""
	resp, respErr := requestAgentMonitor(param, fmt.Sprintf("http://%s", remoteAddress), "delete")
	if respErr != nil {
		err = respErr
		return
	}
	if resp.Code != 200 {
		err = fmt.Errorf(resp.Message)
	}
	if err == nil {
		endpointGuid := param["guid"]
		if v, ok := redirectHttpMap.Load(endpointGuid); ok {
			rhObj := v.(*redirectHttpObj)
			if rhObj != nil {
				log.Printf("rhObj -> port:%d rap:%d rami:%s\n", rhObj.ListenPort, rhObj.RemoteAgentPort, rhObj.RemoteAgentManagerIp)
				rhObj.Destroy()
				redirectHttpMap.Delete(endpointGuid)
				log.Printf("delete redirect http done,guid:%s \n", endpointGuid)
			}
		}
	}
	return
}

func requestAgentMonitor(param interface{}, url, method string) (resp funcs.HttpResponse, err error) {
	postData, err := json.Marshal(param)
	if err != nil {
		log.Printf("Failed marshalling data:%s ", err.Error())
		return resp, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s:19999/deploy/%s", url, method), strings.NewReader(string(postData)))
	if err != nil {
		log.Printf("Curl agent_monitor http request error:%s ", err.Error())
		return resp, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Curl agent_monitor http response error:%s ", err.Error())
		return resp, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Printf("Curl %s agent_monitor response : %s ", method, string(body))
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Printf("Curl agent_monitor unmarshal error:%s ", err.Error())
		return resp, err
	}
	return resp, nil
}
