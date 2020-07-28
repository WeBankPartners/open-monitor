package funcs

import (
	"net/http"
	"log"
	"fmt"
	"time"
	"encoding/json"
)

func InitHttpHandles()  {
	if !Config().Http.Enable || Config().Http.Port<=0 {
		return
	}
	http.Handle("/archive/v1/1m/job", http.HandlerFunc(handleCustomJob))
	http.Handle("/archive/v1/5m/job", http.HandlerFunc(handleFiveMinJob))
	listenPort := fmt.Sprintf(":%d", Config().Http.Port)
	log.Printf("listening %s ...\n", listenPort)
	http.ListenAndServe(listenPort, nil)
}

func handleCustomJob(w http.ResponseWriter,r *http.Request)  {
	dateString := r.FormValue("date")
	_, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 CST", dateString))
	if err != nil {
		returnJson(r,w,err,nil)
	}else{
		CreateJob(dateString)
		returnJson(r,w,err,"start 1min job success")
	}
}

func handleFiveMinJob(w http.ResponseWriter,r *http.Request)  {
	dateString := r.FormValue("date")
	t, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 CST", dateString))
	if err != nil {
		returnJson(r,w,err,nil)
	}else{
		ArchiveFromMysql(t.Unix())
		returnJson(r,w,err,"start 5min job success")
	}
}

func returnJson(r *http.Request,w http.ResponseWriter,err error,result interface{})  {
	w.Header().Set("Content-Type", "application/json")
	var response HttpRespJson
	if err != nil {
		log.Printf(" %s  fail,error:%v \n", r.URL.String(), err)
		response.Code = 1
		response.Msg = err.Error()
	}else{
		log.Printf(" %s success! \n", r.URL.String())
		response.Code = 0
		response.Msg = "success"
	}
	response.Data = result
	w.WriteHeader(http.StatusOK)
	d,_ := json.Marshal(response)
	w.Write(d)
}