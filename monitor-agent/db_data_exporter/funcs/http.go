package funcs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func StartHttpServer(port int) {
	http.Handle("/db/check", http.HandlerFunc(handleCheckIllegal))
	http.Handle("/db/config", http.HandlerFunc(handleAcceptConfig))
	http.Handle("/db/lastkeyword", http.HandlerFunc(handleGetLastKeyword))
	http.Handle("/metrics", http.HandlerFunc(handlePrometheus))
	//http.Handle("/metrics_60", http.HandlerFunc(handlePrometheusWith1min))
	//http.Handle("/metrics_300", http.HandlerFunc(handlePrometheusWith5min))
	//http.Handle("/metrics_1800", http.HandlerFunc(handlePrometheusWith30min))
	//http.Handle("/metrics_3600", http.HandlerFunc(handlePrometheusWith1hour))
	//http.Handle("/metrics_43200", http.HandlerFunc(handlePrometheusWith12hour))
	//http.Handle("/metrics_86400", http.HandlerFunc(handlePrometheusWith24hour))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handleCheckIllegal(w http.ResponseWriter, r *http.Request) {
	var param DbMonitorTaskObj
	var respMessage string
	requestByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respMessage = fmt.Sprintf("handle config read body error : %s \n", err.Error())
		log.Printf(respMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(respMessage))
		return
	}
	log.Printf("check illegal param:%s\n", string(requestByte))
	err = json.Unmarshal(requestByte, &param)
	if err != nil {
		respMessage = fmt.Sprintf("handle config json unmarshal error : %s \n", err.Error())
		log.Printf(respMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(respMessage))
		return
	}
	err = checkIllegal(param)
	if err != nil {
		respMessage = err.Error()
		log.Printf(respMessage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(respMessage))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func handleAcceptConfig(w http.ResponseWriter, r *http.Request) {
	var param []*DbMonitorTaskObj
	var respMessage string
	requestByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respMessage = fmt.Sprintf("handle config read body error : %s \n", err.Error())
		log.Printf(respMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(respMessage))
		return
	}
	log.Printf("accept config param:%s\n", string(requestByte))
	err = json.Unmarshal(requestByte, &param)
	if err != nil {
		respMessage = fmt.Sprintf("handle config json unmarshal error : %s \n", err.Error())
		log.Printf(respMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(respMessage))
		return
	}
	taskLock.Lock()
	for _, v := range param {
		existTaskObj := &DbMonitorTaskObj{}
		for _, existTask := range taskList {
			if existTask.KeywordGuid == v.KeywordGuid {
				existTaskObj = existTask
				break
			}
		}
		if existTaskObj.KeywordGuid != "" {
			if existTaskObj.KeywordCount != 0 {
				v.KeywordCount = existTaskObj.KeywordCount
				v.KeywordContent = existTaskObj.KeywordContent
			}
		}
	}
	taskList = param
	taskLock.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func handlePrometheus(w http.ResponseWriter, r *http.Request) {
	//log.Println("--> get 10s data")
	w.Write(GetExportMetric(10))
}

func handlePrometheusWith1min(w http.ResponseWriter, r *http.Request) {
	log.Println("--> get 1min data")
	w.Write(GetExportMetric(60))
}

func handlePrometheusWith5min(w http.ResponseWriter, r *http.Request) {
	log.Println("--> get 5min data")
	w.Write(GetExportMetric(300))
}

func handlePrometheusWith30min(w http.ResponseWriter, r *http.Request) {
	log.Println("--> get 30min data")
	w.Write(GetExportMetric(1800))
}

func handlePrometheusWith1hour(w http.ResponseWriter, r *http.Request) {
	log.Println("--> get 1h data")
	w.Write(GetExportMetric(3600))
}

func handlePrometheusWith12hour(w http.ResponseWriter, r *http.Request) {
	log.Println("--> get 12h data")
	w.Write(GetExportMetric(43200))
}

func handlePrometheusWith24hour(w http.ResponseWriter, r *http.Request) {
	log.Println("--> get 24h data")
	w.Write(GetExportMetric(86400))
}

func handleGetLastKeyword(w http.ResponseWriter, r *http.Request) {
	var param []*DbLastKeywordDto
	var respMessage string
	requestByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respMessage = fmt.Sprintf("handle get last keyword read body error : %s \n", err.Error())
		log.Printf(respMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(respMessage))
		return
	}
	log.Printf("accept get last keyword param:%s\n", string(requestByte))
	err = json.Unmarshal(requestByte, &param)
	if err != nil {
		respMessage = fmt.Sprintf("handle get last keyword json unmarshal error : %s \n", err.Error())
		log.Printf(respMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(respMessage))
		return
	}
	taskLock.RLock()
	for _, input := range param {
		for _, v := range taskList {
			if v.KeywordGuid == input.KeywordGuid && v.Endpoint == input.Endpoint {
				input.KeywordContent = v.KeywordContent
				break
			}
		}
	}
	taskLock.RUnlock()
	respBytes, _ := json.Marshal(param)
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
