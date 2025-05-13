package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v2/monitor"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v2/service"
	"github.com/WeBankPartners/open-monitor/monitor-server/common/smtp"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	reloadConfigChan = make(chan int, 10000)
)

func StartConsumeReloadConfig() {
	t := time.NewTicker(5 * time.Second).C
	for {
		<-t
		consumeLength := len(reloadConfigChan)
		if consumeLength == 0 {
			continue
		}
		for i := 0; i < consumeLength; i++ {
			<-reloadConfigChan
		}
		consumeReloadConfig()
	}
}

func consumeReloadConfig() {
	resp, err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Reload prometheus config fail", zap.Error(err))
	} else {
		log.Info(nil, log.LOGGER_APP, "Reload prometheus config done")
	}
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

func ReloadConfig() error {
	//_, err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
	//if err != nil {
	//	log.Error(nil, log.LOGGER_APP, "Reload prometheus config fail", zap.Error(err))
	//}
	//return err
	reloadConfigChan <- 1
	return nil
}

func StartCheckPrometheusJob(interval int) {
	// Check prometheus
	var prometheusAddress string
	for _, v := range m.Config().Datasource.Servers {
		if v.Type == "prometheus" {
			prometheusAddress = v.Host
			break
		}
	}
	if prometheusAddress == "" {
		return
	}
	t := time.NewTicker(time.Second * time.Duration(interval)).C
	for {
		go checkPrometheusAlive(prometheusAddress)
		<-t
	}
}

func checkPrometheusAlive(address string) {
	resp, err := http.Get(fmt.Sprintf("http://%s", address))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Prometheus alive check: error", zap.Error(err))
		restartPrometheus()
	} else {
		if resp != nil {
			if resp.Body != nil {
				resp.Body.Close()
			}
		}
	}
}

func restartPrometheus() {
	log.Info(nil, log.LOGGER_APP, "Try to start prometheus . . . . . .")
	lastLog, _ := execCommand("tail -n 30 /app/monitor/prometheus/logs/prometheus.log")
	if lastLog != "" {
		for _, v := range strings.Split(lastLog, "\n") {
			if strings.Contains(v, "err=\"/app/monitor/prometheus/rules/") {
				errorFile := strings.Split(strings.Split(v, "err=\"/app/monitor/prometheus/rules/")[1], ":")[0]
				err := os.Remove(fmt.Sprintf("/app/monitor/prometheus/rules/%s", errorFile))
				if err != nil {
					log.Error(nil, log.LOGGER_APP, fmt.Sprintf("Remove problem file %s error ", errorFile), zap.Error(err))
				} else {
					log.Info(nil, log.LOGGER_APP, fmt.Sprintf("Remove problem file %s success", errorFile))
				}
			}
		}
	} else {
		log.Info(nil, log.LOGGER_APP, "Prometheus last log is empty ??")
	}
	startCommand, _ := execCommand("cat /app/monitor/start.sh |grep prometheus.yml")
	if startCommand != "" {
		startCommand = strings.Replace(startCommand, "\n", " && ", -1)
		startCommand = strings.ReplaceAll(startCommand, "${archive_day}", m.PrometheusArchiveDay)
		startCommand = startCommand[:len(startCommand)-3]
		log.Debug(nil, log.LOGGER_APP, "restartPrometheus", zap.String("cmd", startCommand))
		_, err := execCommand(startCommand)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Start prometheus fail,error", zap.Error(err))
		} else {
			log.Info(nil, log.LOGGER_APP, "Start prometheus success")
		}
	} else {
		log.Warn(nil, log.LOGGER_APP, "Start prometheus fail, the start command is empty!!")
	}
}

func StartCheckProcessList(interval int) {
	if len(m.Config().ProcessCheckList) == 0 {
		return
	}
	t := time.NewTicker(time.Second * time.Duration(interval)).C
	for {
		<-t
		for _, v := range m.Config().ProcessCheckList {
			go checkSubProcessAlive(v)
		}
	}
}

// StartCheckBusinessConfigMatchCodeCount /** 检查业务配置 匹配的code数是否超过阈值配置,超过最大值的话,则认为配置错误,直接禁用该条业务配置
func StartCheckBusinessConfigMatchCodeCount() {
	t := time.NewTicker(time.Minute * 10).C
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
		if subLogMetricGroups, err = db.GetLogMetricGroupByMetric(metric); err != nil {
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
		if logMetricGroupWarnDto, err = db.GetLogMetricGroupDto(logMetricGroup); err != nil {
			log.Error(nil, log.LOGGER_APP, "Get log metric fail", zap.Error(err))
			continue
		}
		if logMetricGroupWarnDto == nil {
			continue
		}
		logMetricGroupWarnDto.ServiceGroupDisplayName = db.GetServiceGroupDisplayName(logMetricGroupWarnDto.ServiceGroup)
		logMetricGroupWarnDtoList = append(logMetricGroupWarnDtoList, logMetricGroupWarnDto)
		logMetricMonitorGuidMap[logMetricGroupWarnDto.LogMetricMonitorGuid] = true
	}
	log.Info(nil, log.LOGGER_APP, "start do disable log_metric_group", zap.Strings("logMetricGroups", logMetricGroups))
	// 更新状态
	if err = db.BatchDisableLogMetricGroupStatus(logMetricGroups); err != nil {
		log.Error(nil, log.LOGGER_APP, "Batch disable log metric fail", zap.Error(err))
		return
	}
	// 同步 prometheus数据
	if len(logMetricMonitorGuidMap) > 0 {
		logMetricMonitorGuids = monitor.ConvertMap2Arr(logMetricMonitorGuidMap)
		for _, guid := range logMetricMonitorGuids {
			if err = service.SyncLogMetricMonitorConfig(guid); err != nil {
				log.Error(nil, log.LOGGER_APP, "Sync log metric fail", zap.Error(err))
			}
		}
	}
	// 发送邮件
	var mailSender *smtp.MailSender
	var getMailSenderErr error
	if mailSender, getMailSenderErr = db.GetMailSender(); getMailSenderErr != nil {
		log.Error(nil, log.LOGGER_APP, "Try to send custom alarm mail fail", zap.Error(getMailSenderErr))
		return
	}
	toMail := []string{os.Getenv("MONITOR_CHECK_EVENT_TO_MAIL"), os.Getenv("MONITOR_MAIL_DEFAULT_RECEIVER")}
	for _, dto := range logMetricGroupWarnDtoList {
		subject := fmt.Sprintf("业务配置【%s】自动关闭通知", dto.LogMetricGroupName)
		content := fmt.Sprintf("【层级对象%s】【%s】【%s】服务码code识别超过n条，可能出现大量异常告警%d条，系统已自动关闭，请先修复告警配置之后再打开告警", dto.ServiceGroupDisplayName, dto.LogMetricGroupName, dto.Metric, maxCount)
		if sendErr := mailSender.Send(subject, content, toMail); sendErr != nil {
			log.Error(nil, log.LOGGER_APP, "Try to send custom alarm mail fail", zap.Error(sendErr))
		}
	}
}

func checkSubProcessAlive(name string) {
	if name == "" {
		return
	}
	_, err := execCommand("ps aux|grep -v grep|grep " + name)
	if err != nil {
		if strings.Contains(err.Error(), "status 1") {
			startCommand, _ := execCommand("cat /app/monitor/start.sh |grep " + name)
			if startCommand != "" {
				startCommand = strings.Replace(startCommand, "\n", " && ", -1)
				startCommand = startCommand[:len(startCommand)-3]
				_, err := execCommand(startCommand)
				if err != nil {
					log.Error(nil, log.LOGGER_APP, "Start sub process fail,error", zap.String("process", name), zap.String("command", startCommand), zap.Error(err))
				} else {
					log.Info(nil, log.LOGGER_APP, "Start sub process success", zap.String("process", name))
				}
			}
		}
	}
}

func execCommand(str string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", str).Output()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, fmt.Sprintf("Exec command %s fail,error", str), zap.Error(err))
	}
	return string(b), err
}

// getRecentCustomMetricCodeStats 查询最近60秒所有自定义指标内容，统计每个key下不同code的数量，返回数量大于1000的key及其code列表
func getRecentCustomMetricCodeStats(maxCount int) (map[string]int, error) {
	// 1. 查询所有自定义 service_group
	serviceGroups := db.GetLogMetricMonitorServiceGroups()
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
