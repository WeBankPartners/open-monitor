package redirect

import (
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
	ListenPort           int    `json:"listenPort"`
	RemoteAgentManagerIp string `json:"remoteAgentManagerIp"`
}

func (h redirectHttpObj) Init() {
	http.ListenAndServe(fmt.Sprintf(":%d", h.ListenPort), http.HandlerFunc(h.httpHandleFunc))
}

func (h redirectHttpObj) httpHandleFunc(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d%s", h.RemoteAgentManagerIp, h.ListenPort, r.RequestURI))
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
	newParam := funcs.InitDeployParam{Config: param.Config}
	_, err = requestAgentMonitor(&newParam, fmt.Sprintf("http://%s", param.AgentManagerRemoteIp), "init")
	for _, v := range param.Config {
		if splitIndex := strings.LastIndex(v.AgentAddress, ":"); splitIndex > 0 {
			if addressPort, _ := strconv.Atoi(v.AgentAddress[splitIndex:]); addressPort > 0 {
				if _, ok := redirectHttpMap.Load(addressPort); !ok {
					rhObj := redirectHttpObj{ListenPort: addressPort, RemoteAgentManagerIp: param.AgentManagerRemoteIp}
					rhObj.Init()
					redirectHttpMap.Store(addressPort, &rhObj)
					log.Printf("start listen port:%d \n", addressPort)
				}
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
		port, _ = strconv.Atoi(resp.Message[splitIndex:])
	}
	if port > 0 {
		if _, ok := redirectHttpMap.Load(port); !ok {
			rhObj := redirectHttpObj{ListenPort: port, RemoteAgentManagerIp: remoteAddress}
			rhObj.Init()
			redirectHttpMap.Store(port, &rhObj)
			log.Printf("start listen port:%d \n", port)
		}
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
	return
}

func requestAgentMonitor(param interface{}, url, method string) (resp funcs.HttpResponse, err error) {
	postData, err := json.Marshal(param)
	if err != nil {
		log.Printf("Failed marshalling data:%s ", err.Error())
		return resp, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/deploy/%s", url, method), strings.NewReader(string(postData)))
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
