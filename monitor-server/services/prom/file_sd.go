package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"io/ioutil"
	"strings"
)

var consumeSdConfigChan = make(chan m.SdLocalConfigJob, 100)

func StartConsumeSdConfig() {
	if strings.HasSuffix(m.Config().Prometheus.SdConfigPath, "/") {
		m.Config().Prometheus.SdConfigPath = m.Config().Prometheus.SdConfigPath[:len(m.Config().Prometheus.SdConfigPath)-1]
	}
	output, err := execCommand("mkdir -p " + m.Config().Prometheus.SdConfigPath)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Run make sd dir command fail", zap.String("output", output), zap.Error(err))
	}
	log.Info(nil, log.LOGGER_APP, "Start consume prometheus service discover config job")
	for {
		sdConfig := <-consumeSdConfigChan
		consumeSdConfig(sdConfig.Configs)
		//if !sdConfig.FromPeer && tmpErr == nil && len(steps) > 0 {
		//	go other.SyncPeerConfig(0, m.SyncSdConfigDto{StepList: steps})
		//}
	}
}

func consumeSdConfig(params []*m.SdConfigSyncObj) {
	var err error
	log.Info(nil, log.LOGGER_APP, "start consume sd config")
	for _, param := range params {
		configFile := fmt.Sprintf("%s/sd_file_%d.json", m.Config().Prometheus.SdConfigPath, param.Step)
		writeErr := ioutil.WriteFile(configFile, []byte(param.Content), 0644)
		if writeErr != nil {
			err = fmt.Errorf("Try to write sd file fail,%s ", writeErr.Error())
			break
		}
	}
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Consume sd config fail", zap.Error(err))
	} else {
		ReloadConfig()
	}
	return
}

func SyncLocalSdConfig(param m.SdLocalConfigJob) {
	if len(param.Configs) > 0 {
		consumeSdConfigChan <- param
	}
}
