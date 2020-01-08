package manager

import (
	"net/http"
	"io/ioutil"
	"log"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/funcs"
)

func AddDeploy(w http.ResponseWriter,r *http.Request)  {
	var resp httpResponse
	b,err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error : %v \n", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("error:%v",err)
	}else{
		var tmpParamMap map[string]string
		err = json.Unmarshal(b, &tmpParamMap)
		if err != nil {
			resp.Code = 500
			resp.Message = fmt.Sprintf("error:%v",err)
		}else{
			var exporter,configFile,guid string
			if _,b := tmpParamMap["guid"]; !b {
				resp.Code = 400
				resp.Message = "param guid can not find!"
			}
			guid = tmpParamMap["guid"]
			if v,b := tmpParamMap["exporter"]; !b {
				resp.Code = 400
				resp.Message = "param exporter can not find!"
			}else {
				exporter = v
				if v, b := tmpParamMap["config"]; b {
					configFile = v
				}
				port,err := funcs.AddDeploy(exporter, configFile, guid, tmpParamMap)
				if err != nil {
					resp.Code = 500
					resp.Message = fmt.Sprintf("error:%v", err)
				}else{
					resp.Code = 200
					if port > 0 {
						resp.Message = fmt.Sprintf("%s:%d", funcs.LocalIp, port)
					}else{
						resp.Message = "exist"
					}
				}
			}
		}
	}
	funcs.SaveDeployProcess()
	w.Write(resp.byte())
}

func DelDeploy(w http.ResponseWriter,r *http.Request)  {
	var resp httpResponse
	b,err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error : %v \n", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("error:%v",err)
	}else {
		var tmpParamMap map[string]string
		err = json.Unmarshal(b, &tmpParamMap)
		if err != nil {
			resp.Code = 500
			resp.Message = fmt.Sprintf("error:%v", err)
		} else {
			if v,b := tmpParamMap["guid"];b {
				err = funcs.DeleteDeploy(v)
				if err != nil {
					resp.Code = 500
					resp.Message = fmt.Sprintf("error:%v", err)
				}else{
					resp.Code = 200
					resp.Message = "success"
				}
			}else{
				resp.Code = 400
				resp.Message = "Param guid not exist"
			}
		}
	}
	funcs.SaveDeployProcess()
	w.Write(resp.byte())
}

func DisplayProcess(w http.ResponseWriter,r *http.Request)  {
	w.Write(funcs.PrintProcessList())
}

type httpResponse struct {
	Code  int  `json:"code"`
	Message  string  `json:"message"`
	Data  interface{}  `json:"data"`
}

func (h *httpResponse) byte() []byte {
	d,err := json.Marshal(h)
	if err == nil {
		return d
	}else{
		return []byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\",\"data\":%v}", h.Code, h.Message, h.Data))
	}
}