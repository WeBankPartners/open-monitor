package funcs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func InitHttpHandles() {
	if !Config().Http.Enable || Config().Http.Port <= 0 {
		return
	}
	http.Handle("/archive/v1/1m/job", http.HandlerFunc(handleCustomJob))
	http.Handle("/archive/v1/5m/job", http.HandlerFunc(handleFiveMinJob))
	http.Handle("/archive/v1/status/db", http.HandlerFunc(handleDbStatus))
	http.Handle("/archive/v1/health/db", http.HandlerFunc(handleDbHealth))
	listenPort := fmt.Sprintf(":%d", Config().Http.Port)
	log.Printf("listening %s ...\n", listenPort)
	http.ListenAndServe(listenPort, nil)
}

func handleCustomJob(w http.ResponseWriter, r *http.Request) {
	dateString := r.FormValue("date")
	if strings.Contains(dateString, "_") {
		_, err := time.Parse("2006-01-02_15:04:05 MST", fmt.Sprintf("%s:00:00 "+DefaultLocalTimeZone, dateString))
		if err != nil {
			returnJson(r, w, err, nil)
		} else {
			go CreateJob(dateString)
			returnJson(r, w, err, "start 1min job success")
		}
		return
	}
	_, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+DefaultLocalTimeZone, dateString))
	if err != nil {
		returnJson(r, w, err, nil)
	} else {
		go handleDayJob(dateString)
		returnJson(r, w, err, "start 1min job success")
	}
}

func handleDayJob(dayString string) {
	for i := 0; i < 24; i++ {
		jobId := fmt.Sprintf("%s_%d", dayString, i)
		if i < 10 {
			jobId = fmt.Sprintf("%s_0%d", dayString, i)
		}
		CreateJob(jobId)
	}
}

func handleFiveMinJob(w http.ResponseWriter, r *http.Request) {
	dateString := r.FormValue("date")
	t, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+DefaultLocalTimeZone, dateString))
	if err != nil {
		returnJson(r, w, err, nil)
	} else {
		ArchiveFromMysql(t.Unix())
		returnJson(r, w, err, "start 5min job success")
	}
}

func returnJson(r *http.Request, w http.ResponseWriter, err error, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var response HttpRespJson
	if err != nil {
		log.Printf(" %s  fail,error:%v \n", r.URL.String(), err)
		response.Code = 1
		response.Msg = err.Error()
	} else {
		log.Printf(" %s success! \n", r.URL.String())
		response.Code = 0
		response.Msg = "success"
	}
	response.Data = result
	w.WriteHeader(http.StatusOK)
	d, _ := json.Marshal(response)
	w.Write(d)
}

// 数据库状态响应结构体
type DbStatusResponse struct {
	ArchiveDB *DbConnectionStats `json:"archive_db,omitempty"`
	MonitorDB *DbConnectionStats `json:"monitor_db,omitempty"`
	Timestamp string             `json:"timestamp"`
}

func handleDbStatus(w http.ResponseWriter, r *http.Request) {
	response := DbStatusResponse{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 获取归档数据库状态
	if mysqlEngine != nil {
		response.ArchiveDB = getDbConnectionStats(mysqlEngine, true)
	}

	// 获取监控数据库状态
	if monitorMysqlEngine != nil {
		response.MonitorDB = getDbConnectionStats(monitorMysqlEngine, false)
	}

	returnJson(r, w, nil, response)
}

func handleDbHealth(w http.ResponseWriter, r *http.Request) {
	// 执行健康检查
	checkConnectionPoolHealth()

	response := map[string]string{
		"message":   "Database health check completed, check logs for details",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	returnJson(r, w, nil, response)
}
