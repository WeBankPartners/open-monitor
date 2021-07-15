package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

var consumeRuleConfigChan = make(chan models.RFGroup, 50)

func StartConsumeRuleConfig()  {
	if strings.HasSuffix(models.Config().Prometheus.RuleConfigPath, "/") {
		models.Config().Prometheus.RuleConfigPath = models.Config().Prometheus.RuleConfigPath[:len(models.Config().Prometheus.RuleConfigPath)-1]
	}
	output,err := execCommand("mkdir -p " + models.Config().Prometheus.RuleConfigPath)
	if err != nil {
		log.Logger.Error("Run make rule dir command fail", log.String("output", output), log.Error(err))
	}
	log.Logger.Info("Start consume prometheus rule config job")
	for {
		ruleConfig := <- consumeRuleConfigChan
		consumeRuleConfig(ruleConfig)
	}
}

func consumeRuleConfig(input models.RFGroup)  {
	log.Logger.Info("Start consume rule config", log.String("name", input.Name))
	path := fmt.Sprintf("%s/%s.yml", models.Config().Prometheus.RuleConfigPath, input.Name)
	if len(input.Rules) == 0 {
		err := os.Remove(path)
		if err == nil {
			ReloadConfig()
			return
		}
	}
	rf := models.RuleFile{Groups:[]*models.RFGroup{&input}}
	data,err := yaml.Marshal(&rf)
	if err != nil {
		log.Logger.Error("Set prometheus rule,marshal fail", log.Error(err))
		return
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		log.Logger.Error("Set prometheus rule,write file fail", log.Error(err))
	}else{
		ReloadConfig()
	}
}

func SyncLocalRuleConfig(input models.RFGroup)  {
	consumeRuleConfigChan <- input
}