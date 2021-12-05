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
	http.Handle("/metrics", http.HandlerFunc(handlePrometheus))
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
	taskList = param
	taskLock.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func handlePrometheus(w http.ResponseWriter, r *http.Request) {
	w.Write(GetExportMetric())
}
