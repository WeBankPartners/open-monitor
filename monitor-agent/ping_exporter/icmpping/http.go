package icmpping

import (
	"net/http"
	"io/ioutil"
	"log"
	"fmt"
	"encoding/json"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
)

func StartHttpServer() {
	if funcs.Config().Prometheus.Enabled || funcs.Config().IpSource.Listen.Enabled {
		port := funcs.Config().IpSource.Listen.Port
		if funcs.Config().Prometheus.Enabled {
			port = funcs.Config().Prometheus.Port
			http.Handle(funcs.Config().Prometheus.Path, http.HandlerFunc(handlePrometheus))
		}
		if funcs.Config().IpSource.Listen.Enabled {
			http.Handle(funcs.Config().IpSource.Listen.Path, http.HandlerFunc(handleIpSource))
		}
		http.ListenAndServe(":"+port, nil)
	}
}

func handlePrometheus(w http.ResponseWriter,r *http.Request)  {
	w.Write(funcs.GetExportMetric())
}

type ipRequest struct {
	Ip  []string  `json:"ip"`
}

func handleIpSource(w http.ResponseWriter,r *http.Request)  {
	var respMessage string
	var param ipRequest
	b,err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respMessage = fmt.Sprintf("handle ip source read body error : %v \n", err)
		log.Printf(respMessage)
		w.Write([]byte(respMessage))
		return
	}
	err = json.Unmarshal(b, &param)
	if err != nil {
		respMessage = fmt.Sprintf("handle ip source unmarshal json error : %v \n", err)
		log.Printf(respMessage)
		w.Write([]byte(respMessage))
		return
	}
	UpdateIpList(param.Ip, funcs.Config().IpSource.Listen.Weight)
	w.Write([]byte("success"))
}