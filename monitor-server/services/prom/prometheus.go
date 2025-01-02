package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"os"
	"os/exec"
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
	_, err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
	if err != nil {
		log.Logger.Error("Reload prometheus config fail", log.Error(err))
	} else {
		log.Logger.Info("Reload prometheus config done")
	}
}

func ReloadConfig() error {
	//_, err := http.Post(m.Config().Prometheus.ConfigReload, "application/json", strings.NewReader(""))
	//if err != nil {
	//	log.Logger.Error("Reload prometheus config fail", log.Error(err))
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
	_, err := http.Get(fmt.Sprintf("http://%s", address))
	if err != nil {
		log.Logger.Error("Prometheus alive check: error", log.Error(err))
		restartPrometheus()
	}
}

func restartPrometheus() {
	log.Logger.Info("Try to start prometheus . . . . . .")
	lastLog, _ := execCommand("tail -n 30 /app/monitor/prometheus/logs/prometheus.log")
	if lastLog != "" {
		for _, v := range strings.Split(lastLog, "\n") {
			if strings.Contains(v, "err=\"/app/monitor/prometheus/rules/") {
				errorFile := strings.Split(strings.Split(v, "err=\"/app/monitor/prometheus/rules/")[1], ":")[0]
				err := os.Remove(fmt.Sprintf("/app/monitor/prometheus/rules/%s", errorFile))
				if err != nil {
					log.Logger.Error(fmt.Sprintf("Remove problem file %s error ", errorFile), log.Error(err))
				} else {
					log.Logger.Info(fmt.Sprintf("Remove problem file %s success", errorFile))
				}
			}
		}
	} else {
		log.Logger.Info("Prometheus last log is empty ??")
	}
	startCommand, _ := execCommand("cat /app/monitor/start.sh |grep prometheus.yml")
	if startCommand != "" {
		startCommand = strings.Replace(startCommand, "\n", " && ", -1)
		startCommand = strings.ReplaceAll(startCommand, "${archive_day}", m.PrometheusArchiveDay)
		startCommand = startCommand[:len(startCommand)-3]
		log.Logger.Debug("restartPrometheus", log.String("cmd", startCommand))
		_, err := execCommand(startCommand)
		if err != nil {
			log.Logger.Error("Start prometheus fail,error", log.Error(err))
		} else {
			log.Logger.Info("Start prometheus success")
		}
	} else {
		log.Logger.Warn("Start prometheus fail, the start command is empty!!")
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
					log.Logger.Error("Start sub process fail,error", log.String("process", name), log.String("command", startCommand), log.Error(err))
				} else {
					log.Logger.Info("Start sub process success", log.String("process", name))
				}
			}
		}
	}
}

func execCommand(str string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", str).Output()
	if err != nil {
		log.Logger.Error(fmt.Sprintf("Exec command %s fail,error", str), log.Error(err))
	}
	return string(b), err
}
