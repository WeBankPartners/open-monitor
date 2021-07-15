package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"io/ioutil"
	"strings"
)

var consumeSdConfigChan = make(chan m.SdLocalConfigJob, 50)

func StartConsumeSdConfig()  {
	if strings.HasSuffix(m.Config().Prometheus.SdConfigPath, "/") {
		m.Config().Prometheus.SdConfigPath = m.Config().Prometheus.SdConfigPath[:len(m.Config().Prometheus.SdConfigPath)-1]
	}
	output,err := execCommand("mkdir -p " + m.Config().Prometheus.SdConfigPath)
	if err != nil {
		log.Logger.Error("Run make sd dir command fail", log.String("output", output), log.Error(err))
	}
	log.Logger.Info("Start consume prometheus service discover config job")
	for {
		sdConfig := <- consumeSdConfigChan
		steps,tmpErr := consumeSdConfig(sdConfig.Configs)
		if !sdConfig.FromPeer && tmpErr == nil && len(steps) > 0 {
			go other.SyncPeerConfig(0, m.SyncSdConfigDto{StepList: steps})
		}
	}
}

func consumeSdConfig(params []*m.SdConfigSyncObj) (steps []int,err error) {
	log.Logger.Info("start consume sd config")
	for _, param := range params {
		steps = append(steps, param.Step)
		configFile := fmt.Sprintf("%s/sd_file_%d.json", m.Config().Prometheus.SdConfigPath, param.Step)
		writeErr := ioutil.WriteFile(configFile, []byte(param.Content), 0644)
		if writeErr != nil {
			err = fmt.Errorf("Try to write sd file fail,%s ", writeErr.Error())
			break
		}
	}
	if err != nil {
		log.Logger.Error("Consume sd config fail", log.Error(err))
	}else{
		err = ReloadConfig()
	}
	return
}

func SyncLocalSdConfig(param m.SdLocalConfigJob)  {
	if len(param.Configs) > 0 {
		consumeSdConfigChan <- param
	}
}