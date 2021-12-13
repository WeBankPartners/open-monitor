package db

import (
	"os"
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
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
)

var (
	checkEventKey string
	checkEventToMail string
	monitorSelfIp  string
	intervalMin int
	CronJobList []*m.CornJobObj
)

func StartCallCronJob()  {
	if len(CronJobList) == 0 {
		return
	}
	for _,cron := range CronJobList {
		go callCronJob(cron)
	}
	select{}
}

func callCronJob(param *m.CornJobObj)  {
	if param.Interval == 0 {
		return
	}
	t := time.NewTicker(time.Duration(param.Interval)*time.Second).C
	for {
		<- t
		go param.Func()
	}
}

func StartCheckCron()  {
	checkEventKey = os.Getenv("MONITOR_CHECK_EVENT_KEY")
	if checkEventKey == "" {
		log.Logger.Info("Start check cron fail,event key is empty,please check env MONITOR_CHECK_EVENT_KEY")
		return
	}
	checkEventToMail = os.Getenv("MONITOR_CHECK_EVENT_TO_MAIL")
	if checkEventToMail == "" {
		log.Logger.Info("Start check cron fail,to mail is empty,please check env MONITOR_CHECK_EVENT_TO_MAIL")
		return
	}
	intervalMin,_ = strconv.Atoi(os.Getenv("MONITOR_CHECK_EVENT_INTERVAL_MIN"))
	if intervalMin < 1 {
		log.Logger.Info("Start check cron fail,interval min is validate fail,please check env MONITOR_CHECK_EVENT_INTERVAL_MIN")
		return
	}
	monitorSelfIp = os.Getenv("MONITOR_HOST_IP")
	var timeStartValue string
	var timeSubValue,sleepWaitTime int64
	switch intervalMin {
	case 1:
		timeStartValue = fmt.Sprintf("%s:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15:04"))
		timeSubValue=60
	case 10:
		tmpTimeString := time.Now().Format("2006-01-02 15:04")
		timeStartValue = fmt.Sprintf("%s0:00 "+m.DefaultLocalTimeZone, tmpTimeString[:len(tmpTimeString)-1])
		timeSubValue=600
	case 30:
		timeStartValue = fmt.Sprintf("%s:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15"))
		timeSubValue=1800
	case 60:
		timeStartValue = fmt.Sprintf("%s:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15"))
		timeSubValue=3600
	default:
		if intervalMin%60==0 && intervalMin/60>1 {
			timeStartValue = fmt.Sprintf("%s:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15"))
			timeSubValue=3600
		}else{
			timeSubValue = 0
		}
	}
	if timeSubValue == 0 {
		log.Logger.Warn("Invalidate interval setting,must like 1、10、30、60、120、180...60*n")
		return
	}
	log.Logger.Info("Start check cron with event", log.String("key",checkEventKey),log.String("to",checkEventToMail),log.Int("interval_min",intervalMin),log.String("monitor_ip",monitorSelfIp))
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
		log.Logger.Info("Monitor check --> active")
		go DoCheckProgress()
		<- c
	}
}

func DoCheckProgress() error {
	err := UpdateAliveCheckQueue(monitorSelfIp)
	if err != nil {
		log.Logger.Error("Update alive check queue fail", log.Error(err))
		return err
	}
	var requestParam m.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("monitor-auto-check-%s-%d", strings.Replace(monitorSelfIp, ".", "-", -1), time.Now().Unix())
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "SYS_MONITOR"
	requestParam.OperationKey = checkEventKey
	requestParam.OperationData = fmt.Sprintf("monitor-check-%s", monitorSelfIp)
	requestParam.OperationUser = ""
	log.Logger.Info("Notify request data", log.String("eventSeqNo",requestParam.EventSeqNo),log.String("operationKey",requestParam.OperationKey),log.String("operationData",requestParam.OperationData))
	b, _ := json.Marshal(requestParam)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", m.CoreUrl), strings.NewReader(string(b)))
	request.Header.Set("Authorization", m.GetCoreToken())
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Logger.Error("Notify core event new request fail", log.Error(err))
		return err
	}
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Logger.Error("Notify core event ctxhttp request fail", log.Error(err))
		return err
	}
	resultBody, _ := ioutil.ReadAll(res.Body)
	var resultObj m.CoreNotifyResult
	err = json.Unmarshal(resultBody, &resultObj)
	res.Body.Close()
	if err != nil {
		log.Logger.Error("Notify core event unmarshal json body fail", log.Error(err))
		return err
	}
	log.Logger.Info("Request core operation-events result", log.String("status",resultObj.Status),log.String("message",resultObj.Message))
	return nil
}

func GetCheckProgressContent(param string) m.AlarmEntityObj {
	var result m.AlarmEntityObj
	requestMessageIp := strings.Split(param, "-")
	if len(requestMessageIp) != 3 {
		log.Logger.Warn("Get check progress content param validate error", log.String("data",param))
		return result
	}
	err,aliveQueueTable := GetAliveCheckQueue(requestMessageIp[2])
	if err != nil {
		log.Logger.Error("Get check alive queue fail", log.Error(err))
		return result
	}
	result.Id = "monitor-check"
	result.Status = "OK"
	result.To = checkEventToMail
	result.ToMail = checkEventToMail
	result.Subject = "Monitor Check - "+aliveQueueTable[0].Message
	result.Content = fmt.Sprintf("Monitor Self Check Message From %s \r\nTime:%s ", aliveQueueTable[0].Message, time.Now().Format(m.DatetimeFormat))
	log.Logger.Info("get check progress content", log.String("toMail",result.ToMail),log.String("subject",result.Subject),log.String("content",result.Content))
	return result
}

func StartCheckLogKeyword()  {
	t := time.NewTicker(10*time.Second).C
	for {
		<- t
		go CheckLogKeyword()
	}
}

func CheckLogKeyword()  {
	nowTime := time.Now()
	var queryParam m.QueryMonitorData
	queryParam.Start = nowTime.Unix() - 10
	queryParam.End = nowTime.Unix()
	queryParam.Step = 10
	queryParam.PromQ = "node_log_monitor_count_total"
	queryParam.Legend = "$custom_all"
	queryParam.Endpoint = []string{"endpoint"}
	queryParam.Metric = []string{"metric"}
	dataSerials := datasource.PrometheusData(&queryParam)
	if len(dataSerials) == 0 {
		log.Logger.Debug("Check log keyword data empty")
		return
	}
	var logMonitorTable []*m.LogMonitorTable
	err := x.SQL("SELECT * FROM log_monitor").Find(&logMonitorTable)
	if err != nil {
		log.Logger.Error("Check log keyword,get log_monitor data fail", log.Error(err))
		return
	}
	if len(logMonitorTable) == 0 {
		return
	}
	var alarmTable []*m.AlarmTable
	err = x.SQL("SELECT * FROM alarm WHERE s_metric='log_monitor' and status='firing' ORDER BY id DESC").Find(&alarmTable)
	if err != nil {
		log.Logger.Error("Check log keyword,get alarm data fail", log.Error(err))
		return
	}
	var addAlarmRows []*m.AlarmTable
	notifyMap := make(map[string]int)
	for _,v := range logMonitorTable {
		tmpEndpointObj := m.EndpointTable{Id:v.StrategyId}
		GetEndpoint(&tmpEndpointObj)
		if tmpEndpointObj.Guid == "" {
			log.Logger.Warn("Check log keyword,endpoint not find with id="+strconv.Itoa(v.StrategyId))
			continue
		}
		for _,vv := range dataSerials {
			lmt := getLogMonitorAlarmTags(vv.Name)
			if lmt.Endpoint == tmpEndpointObj.Guid && lmt.FilePath == v.Path && lmt.Keyword == v.Keyword {
				lastValue := vv.Data[len(vv.Data)-1][1]
				var oldValue float64
				if lastValue > 0 {
					needAdd := true
					var fetchAlarm m.AlarmTable
					for _,tmpAlarm := range alarmTable {
						if tmpAlarm.Tags == lmt.Tags {
							oldValue = tmpAlarm.StartValue
							if tmpAlarm.EndValue > 0 {
								oldValue = tmpAlarm.EndValue
							}
							if lastValue <= oldValue {
								needAdd = false
							}else{
								fetchAlarm = *tmpAlarm
							}
							break
						}
					}
					if needAdd {
						tmpContent := getLogMonitorRows(tmpEndpointObj.Ip, v.Path, v.Keyword,lastValue,oldValue)
						if len(tmpContent) > 240 {
							tmpContent = tmpContent[:240]
						}
						notifyMap[lmt.Tags] = v.NotifyEnable
						rowAlarmEndpoint := lmt.Endpoint
						if v.OwnerEndpoint != "" {
							rowAlarmEndpoint = v.OwnerEndpoint
						}
						if fetchAlarm.Id > 0 && fetchAlarm.Status == "firing" {
							tmpContent = strings.Split(fetchAlarm.Content, "^^")[0] + "^^" + tmpContent
							addAlarmRows = append(addAlarmRows, &m.AlarmTable{Id: fetchAlarm.Id, StrategyId: 0, Endpoint: rowAlarmEndpoint, Status: "firing", SMetric: "log_monitor", SExpr: "node_log_monitor_count_total", SCond: ">0", SLast: "10s", SPriority: v.Priority, Content: tmpContent, Tags: lmt.Tags, StartValue: fetchAlarm.StartValue, EndValue: lastValue, Start: fetchAlarm.Start, End: nowTime})
						}else {
							tmpContent = tmpContent + "^^"
							addAlarmRows = append(addAlarmRows, &m.AlarmTable{StrategyId: 0, Endpoint: rowAlarmEndpoint, Status: "firing", SMetric: "log_monitor", SExpr: "node_log_monitor_count_total", SCond: ">0", SLast: "10s", SPriority: v.Priority, Content: tmpContent, Tags: lmt.Tags, StartValue: lastValue, Start: nowTime})
						}
					}
				}
			}
		}
	}
	if len(addAlarmRows) > 0 {
		var actions []*Action
		for _,v := range addAlarmRows {
			tmpAction := Action{}
			if v.Id > 0 {
				tmpAction.Sql = "UPDATE alarm SET content=?,end_value=?,end=? WHERE id=?"
				tmpAction.Param = []interface{}{v.Content, v.EndValue, v.End.Format(m.DatetimeFormat), v.Id}
			}else{
				tmpAction.Sql = "INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start,tags) VALUE (?,?,?,?,?,?,?,?,?,?,?,?)"
				tmpAction.Param = []interface{}{v.StrategyId,v.Endpoint,v.Status,v.SMetric,v.SExpr,v.SCond,v.SLast,v.SPriority,v.Content,v.StartValue,v.Start.Format(m.DatetimeFormat),v.Tags}
			}
			actions = append(actions, &tmpAction)
		}
		err = Transaction(actions)
		//err = UpdateAlarms(addAlarmRows)
		if err != nil {
			log.Logger.Error("Update alarm table fail", log.Error(err))
		}
		for _,v := range addAlarmRows {
			if v.Id > 0 {
				continue
			}
			if notifyMap[v.Tags] == 0 {
				log.Logger.Warn("Log monitor notify disable,ignore", log.String("tags", v.Tags))
				continue
			}
			var tmpAlarmTable []*m.AlarmTable
			x.SQL("SELECT id FROM alarm WHERE status='firing' AND tags=?", v.Tags).Find(&tmpAlarmTable)
			if len(tmpAlarmTable) > 0 {
				notifyErr := NotifyCoreEvent("", 0 , tmpAlarmTable[0].Id, 0)
				if notifyErr != nil {
					log.Logger.Error("Try to notify log monitor alarm fail", log.String("tags", v.Tags), log.Error(notifyErr))
				}
			}
		}
	}
}

func getLogMonitorAlarmTags(name string) m.LogMonitorTags {
	var result m.LogMonitorTags
	tmpEndpoint := strings.Split(name, "e_guid=")[1]
	result.Endpoint = strings.Split(tmpEndpoint, ",")[0]
	result.Tags = fmt.Sprintf("e_guid:%s", result.Endpoint)
	fileName := strings.Split(name, "file=")
	if len(fileName) > 1 {
		result.FilePath = strings.Split(fileName[1], ",instance")[0]
	}else{
		return result
	}
	keyWord := strings.Split(name, "keyword=")
	if len(keyWord) > 1 {
		tmpKeyword := keyWord[1]
		result.Keyword = tmpKeyword[:strings.LastIndex(tmpKeyword, "}")]
	}else{
		return result
	}
	result.Tags = fmt.Sprintf("e_guid:%s^file:%s^keyword:%s", result.Endpoint, result.FilePath, result.Keyword)
	return result
}

type logRowsHttpDto struct {
	Path  string  `json:"path"`
	Keyword  string  `json:"keyword"`
	Value  float64  `json:"value"`
	LastValue float64  `json:"last_value"`
}

type logKeywordFetchObj struct {
	Content    string `json:"content"`
}

type logRowsHttpResult struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Data  []logKeywordFetchObj  `json:"data"`
}

func getLogMonitorRows(ip,path,keyword string,lastValue,oldValue float64) string {
	var result string
	if ip == "" || path == "" || keyword == "" {
		return result
	}
	param := logRowsHttpDto{Path: path, Keyword: keyword, Value: lastValue, LastValue: oldValue}
	postData,_ := json.Marshal(param)
	http.DefaultClient.CloseIdleConnections()
	req,err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:9100/log/rows/query", ip), strings.NewReader(string(postData)))
	if err != nil {
		log.Logger.Error("Get log monitor rows fail,new request error", log.Error(err))
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	resp,err := http.DefaultClient.Do(req)
	if err != nil {
		log.Logger.Error("Get log monitor rows fail,response error", log.Error(err))
		return result
	}
	var responseData logRowsHttpResult
	respBytes,_ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBytes, &responseData)
	if err != nil {
		log.Logger.Error("Get log monitor rows fail,response data json unmarshal error", log.Error(err))
		return result
	}
	if responseData.Status != "ok" {
		log.Logger.Error("Get log monitor rows fail,response status error", log.String("status", responseData.Status), log.String("message", responseData.Message))
		return result
	}
	for _,v := range responseData.Data {
		result += fmt.Sprintf("%s\n", v.Content)
	}
	return result
}

func StartCleanAlarmTable()  {
	if m.Config().AlarmAliveMaxDay <= 0 {
		return
	}
	t,err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 ")))
	if err != nil {
		log.Logger.Error("Start clean alarm table job init fail", log.Error(err))
		return
	}
	sleepTime := t.Unix() + 86400 - time.Now().Unix()
	if sleepTime < 0 {
		log.Logger.Warn("Start clean alarm table job fail,calc sleep time fail", log.Int64("sleep time", sleepTime))
		return
	}
	time.Sleep(time.Duration(sleepTime)*time.Second)
	tc := time.NewTicker(86400*time.Second).C
	for {
		go cleanAlarmTableJob()
		<- tc
	}
}

func cleanAlarmTableJob()  {
	log.Logger.Info("Start to clean alarm table")
	maxDay := int64(m.Config().AlarmAliveMaxDay)
	lastDayString := time.Unix(time.Now().Unix() - maxDay*86400, 0).Format("2006-01-02")
	execResult, err := x.Exec(fmt.Sprintf("delete from alarm where (status='ok' or status='closed') and start<='%s 00:00:00'", lastDayString))
	if err != nil {
		log.Logger.Error("Clean alarm table job fail", log.Error(err))
		return
	}
	rowAffected,_ := execResult.RowsAffected()
	log.Logger.Info("Clean alarm table job done", log.String("last day", lastDayString), log.Int64("delete row num", rowAffected))
}