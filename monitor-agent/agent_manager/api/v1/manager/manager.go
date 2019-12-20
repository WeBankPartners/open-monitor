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
			var exporter,configFile string
			if v,b := tmpParamMap["exporter"]; !b {
				resp.Code = 400
				resp.Message = "param exporter can not find!"
			}else {
				exporter = v
				if v, b := tmpParamMap["config"]; b {
					configFile = v
				}
				err = funcs.AddDeploy(exporter, configFile, tmpParamMap)
				if err != nil {
					resp.Code = 500
					resp.Message = fmt.Sprintf("error:%v", err)
				}else{
					resp.Code = 200
					resp.Message = "success"
				}
			}
		}
	}
	w.Write(resp.byte())
}

func DelDeploy(w http.ResponseWriter,r *http.Request)  {

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