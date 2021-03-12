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
	if funcs.Config().Prometheus.Enabled || funcs.Config().Source.Listen.Enabled {
		port := funcs.Config().Source.Listen.Port
		if funcs.Config().Prometheus.Enabled {
			port = funcs.Config().Prometheus.Port
			http.Handle(funcs.Config().Prometheus.Path, http.HandlerFunc(handlePrometheus))
		}
		if funcs.Config().Source.Listen.Enabled {
			loadHttpConfigData()
			http.Handle(funcs.Config().Source.Listen.Path, http.HandlerFunc(handleIpSource))
		}
		http.ListenAndServe(":"+port, nil)
	}
}

func handlePrometheus(w http.ResponseWriter,r *http.Request)  {
	w.Write(funcs.GetExportMetric())
}

func handleIpSource(w http.ResponseWriter,r *http.Request)  {
	var respMessage string
	var param funcs.RemoteResponse
	b,err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		respMessage = fmt.Sprintf("handle ip source read body error : %v \n", err)
		log.Printf(respMessage)
		w.Write([]byte(respMessage))
		return
	}
	err = json.Unmarshal(b, &param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		respMessage = fmt.Sprintf("handle ip source unmarshal json error : %v \n", err)
		log.Printf(respMessage)
		w.Write([]byte(respMessage))
		return
	}
	var ips []string
	for _,v := range param.Config {
		ips = append(ips, v.Ip)
	}
	funcs.UpdateIpList(ips, funcs.Config().Source.Listen.Weight)
	funcs.UpdateSourceRemoteData(param.Config)
	saveHttpConfigData(b)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func saveHttpConfigData(bodyData []byte)  {
	err := ioutil.WriteFile("http_config_data.json", bodyData, 0644)
	if err != nil {
		log.Printf("Save http config data fail,error: %s \n", err.Error())
	}else{
		log.Println("Save http config data success")
	}
}

func loadHttpConfigData()  {
	b,err := ioutil.ReadFile("http_config_data.json")
	if err != nil {
		log.Printf("Load http config data fail,error: %s \n", err.Error())
		return
	}
	var param funcs.RemoteResponse
	err = json.Unmarshal(b, &param)
	if err != nil {
		log.Printf("Load http config data fail,json unmarshal error: %s \n", err.Error())
		return
	}
	var ips []string
	for _,v := range param.Config {
		ips = append(ips, v.Ip)
	}
	funcs.UpdateIpList(ips, funcs.Config().Source.Listen.Weight)
	funcs.UpdateSourceRemoteData(param.Config)
}