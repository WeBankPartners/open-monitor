package db

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/common/smtp"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
)

var (
	checkEventKey    string
	checkEventToMail string
	monitorSelfIp    string
	intervalMin      int
	CronJobList      []*m.CornJobObj
)

func StartCallCronJob() {
	if len(CronJobList) == 0 {
		return
	}
	for _, cron := range CronJobList {
		go callCronJob(cron)
	}
	select {}
}

func callCronJob(param *m.CornJobObj) {
	if param.Interval == 0 {
		return
	}
	t := time.NewTicker(time.Duration(param.Interval) * time.Second).C
	for {
		<-t
		go param.Func()
	}
}

func StartCheckCron() {
	checkEventKey = os.Getenv("MONITOR_CHECK_EVENT_KEY")
	if checkEventKey == "" {
		log.Info(nil, log.LOGGER_APP, "Start check cron fail,event key is empty,please check env MONITOR_CHECK_EVENT_KEY")
		return
	}
	checkEventToMail = os.Getenv("MONITOR_CHECK_EVENT_TO_MAIL")
	if checkEventToMail == "" {
		log.Info(nil, log.LOGGER_APP, "Start check cron fail,to mail is empty,please check env MONITOR_CHECK_EVENT_TO_MAIL")
		return
	}
	intervalMin, _ = strconv.Atoi(os.Getenv("MONITOR_CHECK_EVENT_INTERVAL_MIN"))
	if intervalMin < 1 {
		log.Info(nil, log.LOGGER_APP, "Start check cron fail,interval min is validate fail,please check env MONITOR_CHECK_EVENT_INTERVAL_MIN")
		return
	}
	monitorSelfIp = os.Getenv("MONITOR_HOST_IP")
	var timeStartValue string
	var timeSubValue, sleepWaitTime int64
	switch intervalMin {
	case 1:
		timeStartValue = fmt.Sprintf("%s:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15:04"))
		timeSubValue = 60
	case 10:
		tmpTimeString := time.Now().Format("2006-01-02 15:04")
		timeStartValue = fmt.Sprintf("%s0:00 "+m.DefaultLocalTimeZone, tmpTimeString[:len(tmpTimeString)-1])
		timeSubValue = 600
	case 30:
		timeStartValue = fmt.Sprintf("%s:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15"))
		timeSubValue = 1800
	case 60:
		timeStartValue = fmt.Sprintf("%s:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15"))
		timeSubValue = 3600
	default:
		if intervalMin%60 == 0 && intervalMin/60 > 1 {
			timeStartValue = fmt.Sprintf("%s:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 15"))
			timeSubValue = 3600
		} else {
			timeSubValue = 0
		}
	}
	if timeSubValue == 0 {
		log.Warn(nil, log.LOGGER_APP, "Invalidate interval setting,must like 1、10、30、60、120、180...60*n")
		return
	}
	log.Info(nil, log.LOGGER_APP, "Start check cron with event", zap.String("key", checkEventKey), zap.String("to", checkEventToMail), zap.Int("interval_min", intervalMin), zap.String("monitor_ip", monitorSelfIp))
	t, _ := time.Parse("2006-01-02 15:04:05 MST", timeStartValue)
	if timeSubValue == 1800 {
		if time.Now().Unix() > t.Unix()+timeSubValue {
			sleepWaitTime = t.Unix() + 3600 - time.Now().Unix()
		} else {
			sleepWaitTime = t.Unix() + 1800 - time.Now().Unix()
		}
	} else {
		sleepWaitTime = t.Unix() + timeSubValue - time.Now().Unix()
	}
	time.Sleep(time.Duration(sleepWaitTime) * time.Second)
	c := time.NewTicker(time.Duration(intervalMin) * time.Minute).C
	for {
		log.Info(nil, log.LOGGER_APP, "Monitor check --> active")
		go DoCheckProgress()
		<-c
	}
}

func DoCheckProgress() error {
	err := UpdateAliveCheckQueue(monitorSelfIp)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Update alive check queue fail", zap.Error(err))
		return err
	}
	var requestParam m.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("monitor-auto-check-%s-%d", strings.Replace(monitorSelfIp, ".", "-", -1), time.Now().Unix())
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "SYS_MONITOR"
	requestParam.OperationKey = checkEventKey
	requestParam.OperationData = fmt.Sprintf("monitor-check-%s", monitorSelfIp)
	requestParam.OperationUser = ""
	log.Info(nil, log.LOGGER_APP, "Notify request data", zap.String("eventSeqNo", requestParam.EventSeqNo), zap.String("operationKey", requestParam.OperationKey), zap.String("operationData", requestParam.OperationData))
	b, _ := json.Marshal(requestParam)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", m.CoreUrl), strings.NewReader(string(b)))
	request.Header.Set("Authorization", m.GetCoreToken())
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Notify core event new request fail", zap.Error(err))
		return err
	}
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Notify core event ctxhttp request fail", zap.Error(err))
		return err
	}
	resultBody, _ := ioutil.ReadAll(res.Body)
	var resultObj m.CoreNotifyResult
	err = json.Unmarshal(resultBody, &resultObj)
	res.Body.Close()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Notify core event unmarshal json body fail", zap.Error(err))
		return err
	}
	log.Info(nil, log.LOGGER_APP, "Request core operation-events result", zap.String("status", resultObj.Status), zap.String("message", resultObj.Message))
	return nil
}

func GetCheckProgressContent(param string) m.AlarmEntityObj {
	var result m.AlarmEntityObj
	requestMessageIp := strings.Split(param, "-")
	if len(requestMessageIp) != 3 {
		log.Warn(nil, log.LOGGER_APP, "Get check progress content param validate error", zap.String("data", param))
		return result
	}
	err, aliveQueueTable := GetAliveCheckQueue(requestMessageIp[2])
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get check alive queue fail", zap.Error(err))
		return result
	}
	result.Id = "monitor-check"
	result.Status = "OK"
	result.To = checkEventToMail
	result.ToMail = checkEventToMail
	result.Subject = "Monitor Check - " + aliveQueueTable[0].Message
	result.Content = fmt.Sprintf("Monitor Self Check Message From %s \r\nTime:%s ", aliveQueueTable[0].Message, time.Now().Format(m.DatetimeFormat))
	log.Info(nil, log.LOGGER_APP, "get check progress content", zap.String("toMail", result.ToMail), zap.String("subject", result.Subject), zap.String("content", result.Content))
	return result
}

func StartCleanAlarmTable() {
	if m.Config().AlarmAliveMaxDay == "" {
		return
	}
	t, err := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+m.DefaultLocalTimeZone, time.Now().Format("2006-01-02 ")))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Start clean alarm table job init fail", zap.Error(err))
		return
	}
	sleepTime := t.Unix() + 86400 - time.Now().Unix()
	if sleepTime < 0 {
		log.Warn(nil, log.LOGGER_APP, "Start clean alarm table job fail,calc sleep time fail", zap.Int64("sleep time", sleepTime))
		return
	}
	time.Sleep(time.Duration(sleepTime) * time.Second)
	tc := time.NewTicker(86400 * time.Second).C
	for {
		go cleanAlarmTableJob()
		<-tc
	}
}

func cleanAlarmTableJob() {
	log.Info(nil, log.LOGGER_APP, "Start to clean alarm table")
	aliveInt, _ := strconv.Atoi(m.Config().AlarmAliveMaxDay)
	if aliveInt <= 0 {
		return
	}
	maxDay := int64(aliveInt)
	lastDayString := time.Unix(time.Now().Unix()-maxDay*86400, 0).Format("2006-01-02")
	execResult, err := x.Exec(fmt.Sprintf("delete from alarm where status='ok' and start<='%s 00:00:00'", lastDayString))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Clean alarm table job fail", zap.Error(err))
		return
	}
	rowAffected, _ := execResult.RowsAffected()
	log.Info(nil, log.LOGGER_APP, "Clean alarm table job done", zap.String("last day", lastDayString), zap.Int64("delete row num", rowAffected))
}

// StartCheckBusinessConfigMatchCodeCount /** 检查业务配置 匹配的code数是否超过阈值配置,超过最大值的话,则认为配置错误,直接禁用该条业务配置
func StartCheckBusinessConfigMatchCodeCount() {
	t := time.NewTicker(time.Minute * 2).C
	for {
		go checkBusinessConfigMatchCodeCount()
		<-t
	}
}

func checkBusinessConfigMatchCodeCount() {
	var metricKeyMap = make(map[string]int)
	var metricServiceGroupMap = make(map[string]string)
	var logMetricGroups, subLogMetricGroups, logMetricMonitorGuids []string
	var logMetricMonitorGuidMap = make(map[string]bool)
	var logMetricGroupWarnDtoList []*m.LogMetricGroupWarnDto
	var logMetricGroupWarnDto *m.LogMetricGroupWarnDto
	var err error
	maxCountStr := os.Getenv("MONITOR_BUSINESS_CONFIG_MAX_CODE_COUNT")
	maxCount, _ := strconv.Atoi(maxCountStr)
	if maxCount == 0 {
		log.Warn(nil, log.LOGGER_APP, "system variable MONITOR_BUSINESS_CONFIG_MAX_CODE_COUNT is null")
		maxCount = 1000
	}
	if metricKeyMap, err = getRecentCustomMetricCodeStats(maxCount); err != nil {
		log.Error(nil, log.LOGGER_APP, "Get recent custom metric code fail", zap.Error(err))
		return
	}
	if len(metricKeyMap) == 0 {
		return
	}
	for metric, _ := range metricKeyMap {
		if subLogMetricGroups, err = GetLogMetricGroupByMetric(metric); err != nil {
			log.Error(nil, log.LOGGER_APP, "Get log metric fail", zap.Error(err))
			continue
		}
		if len(subLogMetricGroups) > 0 {
			metricServiceGroupMap[metric] = subLogMetricGroups[0]
		}
		logMetricGroups = append(logMetricGroups, subLogMetricGroups...)
	}
	if len(logMetricGroups) == 0 {
		return
	}
	for _, logMetricGroup := range logMetricGroups {
		if logMetricGroupWarnDto, err = GetLogMetricGroupDto(logMetricGroup); err != nil {
			log.Error(nil, log.LOGGER_APP, "Get log metric fail", zap.Error(err))
			continue
		}
		if logMetricGroupWarnDto == nil {
			continue
		}
		logMetricGroupWarnDto.ServiceGroupDisplayName = GetServiceGroupDisplayName(logMetricGroupWarnDto.ServiceGroup)
		logMetricGroupWarnDtoList = append(logMetricGroupWarnDtoList, logMetricGroupWarnDto)
		logMetricMonitorGuidMap[logMetricGroupWarnDto.LogMetricMonitorGuid] = true
	}
	// 过滤掉已经禁用的数据(兜底逻辑)
	var disableIdsMap = make(map[string]bool)
	if disableIdsMap, err = BatchQueryDisabledLogMetricGroupStatus(logMetricGroups); err != nil {
		log.Error(nil, log.LOGGER_APP, "BatchQueryDisabledLogMetricGroupStatus fail", zap.Error(err))
	}
	if len(disableIdsMap) > 0 {
		var tempLogMetricGroups []string
		for _, group := range logMetricGroups {
			if _, ok := disableIdsMap[group]; !ok {
				tempLogMetricGroups = append(tempLogMetricGroups, group)
			}
		}
		logMetricGroups = tempLogMetricGroups
	}
	if len(logMetricGroups) == 0 {
		return
	}
	log.Info(nil, log.LOGGER_APP, "start do disable log_metric_group", zap.Strings("logMetricGroups", logMetricGroups))
	// 更新状态
	if err = BatchDisableLogMetricGroupStatus(logMetricGroups); err != nil {
		log.Error(nil, log.LOGGER_APP, "Batch disable log metric fail", zap.Error(err))
		return
	}
	// 同步 prometheus数据
	if len(logMetricMonitorGuidMap) > 0 {
		logMetricMonitorGuids = ConvertMap2Arr(logMetricMonitorGuidMap)
		for _, guid := range logMetricMonitorGuids {
			var endpointList []string
			endpointRel := ListLogMetricEndpointRel(guid)
			for _, v := range endpointRel {
				if v.SourceEndpoint != "" {
					endpointList = append(endpointList, v.SourceEndpoint)
				}
			}
			if err = SyncLogMetricExporterConfig(endpointList); err != nil {
				log.Error(nil, log.LOGGER_APP, "Sync log metric  exporter config fail", zap.Error(err))
			}
		}
	}
	// 发送邮件
	var mailSender *smtp.MailSender
	var getMailSenderErr error
	if mailSender, getMailSenderErr = GetMailSender(); getMailSenderErr != nil {
		log.Error(nil, log.LOGGER_APP, "Try to send custom init fail", zap.Error(getMailSenderErr))
		return
	}
	toMail := []string{os.Getenv("MONITOR_CHECK_EVENT_TO_MAIL"), os.Getenv("MONITOR_MAIL_DEFAULT_RECEIVER")}
	log.Info(nil, log.LOGGER_APP, "send mail", zap.Strings("receivers", toMail))
	for _, dto := range logMetricGroupWarnDtoList {
		subject := fmt.Sprintf("业务配置【%s】自动关闭通知", dto.LogMetricGroupName)
		content := fmt.Sprintf("【层级对象%s】【%s】【%s】服务码code识别超过%d条，可能出现大量异常告警，系统已自动关闭，请先修复告警配置之后再打开告警", dto.ServiceGroupDisplayName, dto.LogMetricGroupName, dto.Metric, maxCount)
		if sendErr := mailSender.Send(subject, content, toMail); sendErr != nil {
			log.Error(nil, log.LOGGER_APP, "Try to send custom alarm mail fail", zap.Error(sendErr))
		}
	}
}

// getRecentCustomMetricCodeStats 查询最近60秒所有自定义指标内容，统计每个key下不同code的数量，返回数量大于1000的key及其code列表
func getRecentCustomMetricCodeStats(maxCount int) (map[string]int, error) {
	// 1. 查询所有自定义 service_group
	serviceGroups := GetLogMetricMonitorServiceGroups()
	result := make(map[string]map[string]struct{}) // key -> set of code
	end := time.Now().Unix()
	start := end - 60
	step := int64(10)
	for _, sg := range serviceGroups {
		promQL := fmt.Sprintf(`node_log_metric_monitor_value{service_group="%s", key=~".*_req_count$"}`, sg)
		data, err := datasource.QueryPrometheusRange(promQL, start, end, step)
		if err != nil || data == nil {
			continue
		}
		for _, r := range data.Result {
			key := r.Metric["key"]
			code := r.Metric["code"]
			if key == "" || code == "" {
				continue
			}
			if _, ok := result[key]; !ok {
				result[key] = make(map[string]struct{})
			}
			result[key][code] = struct{}{}
		}
	}
	filtered := make(map[string]int)
	for key, codeSet := range result {
		if len(codeSet) > maxCount {
			filtered[key] = len(codeSet)
		}
	}
	return filtered, nil
}

func ConvertMap2Arr(hashMap map[string]bool) []string {
	var list []string
	for s, _ := range hashMap {
		list = append(list, s)
	}
	return list
}
