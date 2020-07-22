package db

import (
	"os"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
	"encoding/json"
	"net/http"
	"strings"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"fmt"
	"context"
	"strconv"
)

var (
	checkEventKey string
	checkEventToMail string
	monitorSelfIp  string
	intervalMin int
)

func StartCheckCron()  {
	checkEventKey = os.Getenv("MONITOR_CHECK_EVENT_KEY")
	if checkEventKey == "" {
		mid.LogInfo("start check cron fail,event key is empty,please check env MONITOR_CHECK_EVENT_KEY")
		return
	}
	checkEventToMail = os.Getenv("MONITOR_CHECK_EVENT_TO_MAIL")
	if checkEventToMail == "" {
		mid.LogInfo("start check cron fail,to mail is empty,please check env MONITOR_CHECK_EVENT_TO_MAIL")
		return
	}
	intervalMin,_ = strconv.Atoi(os.Getenv("MONITOR_CHECK_EVENT_INTERVAL_MIN"))
	if intervalMin < 1 {
		mid.LogInfo("start check cron fail,interval min is validate fail,please check env MONITOR_CHECK_EVENT_INTERVAL_MIN")
		return
	}
	monitorSelfIp = os.Getenv("MONITOR_HOST_IP")
	var timeStartValue string
	var timeSubValue,sleepWaitTime int64
	switch intervalMin {
	case 1:
		timeStartValue = fmt.Sprintf("%s:00 CST", time.Now().Format("2006-01-02 15:04"))
		timeSubValue=60
	case 10:
		tmpTimeString := time.Now().Format("2006-01-02 15:04")
		timeStartValue = fmt.Sprintf("%s0:00 CST", tmpTimeString[:len(tmpTimeString)-1])
		timeSubValue=600
	case 30:
		timeStartValue = fmt.Sprintf("%s:00:00 CST", time.Now().Format("2006-01-02 15"))
		timeSubValue=1800
	case 60:
		timeStartValue = fmt.Sprintf("%s:00:00 CST", time.Now().Format("2006-01-02 15"))
		timeSubValue=3600
	default:
		if intervalMin%60==0 && intervalMin/60>1 {
			timeStartValue = fmt.Sprintf("%s:00:00 CST", time.Now().Format("2006-01-02 15"))
			timeSubValue=3600
		}else{
			timeSubValue = 0
		}
	}
	if timeSubValue == 0 {
		mid.LogInfo("invalidate interval setting,must like 1、10、30、60、120、180...60*n \n")
		return
	}
	mid.LogInfo(fmt.Sprintf("start check cron with event key=%s to=%s interval_min=%d monitor_ip=%s", checkEventKey, checkEventToMail, intervalMin, monitorSelfIp))
	t,_ := time.Parse("2006-01-02 15:04:05 MST", timeStartValue)
	if timeSubValue == 1800 {
		if time.Now().Unix() > t.Unix()+timeSubValue {
			sleepWaitTime = t.Unix()+3600-time.Now().Unix()
		}else{
			sleepWaitTime = t.Unix()+1800-time.Now().Unix()
		}
	}else{
		sleepWaitTime = t.Unix()+timeSubValue-time.Now().Unix()
	}
	time.Sleep(time.Duration(sleepWaitTime)*time.Second)
	c := time.NewTicker(time.Duration(intervalMin)*time.Minute).C
	for {
		mid.LogInfo("monitor check --> active \n")
		go DoCheckProgress()
		<- c
	}
}

func DoCheckProgress() error {
	err := UpdateAliveCheckQueue(monitorSelfIp)
	if err != nil {
		mid.LogError("update alive check queue fail", err)
		return err
	}
	var requestParam m.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("monitor-auto-check-%s-%d", strings.Replace(monitorSelfIp, ".", "-", -1), time.Now().Unix())
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "SYS_MONITOR"
	requestParam.OperationKey = checkEventKey
	requestParam.OperationData = fmt.Sprintf("monitor-check-%s", monitorSelfIp)
	requestParam.OperationUser = ""
	mid.LogInfo(fmt.Sprintf("notify request data --> eventSeqNo:%s operationKey:%s operationData:%s", requestParam.EventSeqNo, requestParam.OperationKey, requestParam.OperationData))
	b, _ := json.Marshal(requestParam)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", m.CoreUrl), strings.NewReader(string(b)))
	request.Header.Set("Authorization", m.TmpCoreToken)
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		mid.LogError("notify core event new request fail", err)
		return err
	}
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		mid.LogError("notify core event ctxhttp request fail", err)
		return err
	}
	resultBody, _ := ioutil.ReadAll(res.Body)
	var resultObj m.CoreNotifyResult
	err = json.Unmarshal(resultBody, &resultObj)
	res.Body.Close()
	if err != nil {
		mid.LogError("notify core event unmarshal json body fail", err)
		return err
	}
	mid.LogInfo(fmt.Sprintf("result -> status:%s  message:%s", resultObj.Status, resultObj.Message))
	return nil
}

func GetCheckProgressContent(param string) m.AlarmEntityObj {
	var result m.AlarmEntityObj
	requestMessageIp := strings.Split(param, "-")
	if len(requestMessageIp) != 3 {
		mid.LogError(fmt.Sprintf("get check progress content param validate error with data=%s ", param), nil)
		return result
	}
	err,aliveQueueTable := GetAliveCheckQueue(requestMessageIp[2])
	if err != nil {
		mid.LogError("get check alive queue fail", err)
		return result
	}
	result.Id = "monitor-check"
	result.Status = "OK"
	result.To = checkEventToMail
	result.ToMail = checkEventToMail
	result.Subject = "Monitor Check - "+aliveQueueTable[0].Message
	result.Content = fmt.Sprintf("Monitor Self Check Message From %s \r\nTime:%s ", aliveQueueTable[0].Message, time.Now().Format(m.DatetimeFormat))
	mid.LogInfo(fmt.Sprintf("get check progress content: %s, %s, %s", result.ToMail, result.Subject, result.Content))
	return result
}