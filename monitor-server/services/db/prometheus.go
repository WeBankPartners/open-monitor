package db

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
)

func InitPrometheusConfig()  {
	var err error
	// init sd config
	go prom.StartConsumeSdConfig()
	InitPrometheusServiceDiscoverConfig()
	// init rule config
	go prom.StartConsumeRuleConfig()
	initPrometheusRuleConfig()
	// init kubernetes config
	err = SyncKubernetesConfig()
	if err != nil {
		log.Logger.Error("Start sync kubernetes config fail", log.Error(err))
	}
	// init snmp config
	err = SyncSnmpPrometheusConfig()
	if err != nil {
		log.Logger.Error("Start sync snmp config fail", log.Error(err))
	}
	select {}
}

func InitPrometheusServiceDiscoverConfig()  {
	var endpointTable []*models.EndpointTable
	err := x.SQL("select step,cluster from endpoint group by step,cluster order by cluster,step").Find(&endpointTable)
	if err != nil {
		log.Logger.Error("Init prometheus sd config fail,query endpoint table fail", log.Error(err))
		return
	}
	if len(endpointTable) == 0 {
		return
	}
	var clusterStepMap = make(map[string][]int)
	for _,endpoint := range endpointTable {
		if _,b := clusterStepMap[endpoint.Cluster]; !b {
			clusterStepMap[endpoint.Cluster] = []int{endpoint.Step}
		}else{
			clusterStepMap[endpoint.Cluster] = append(clusterStepMap[endpoint.Cluster], endpoint.Step)
		}
	}
	for k,v := range clusterStepMap {
		tmpErr := SyncSdEndpointNew(v, k)
		if tmpErr != nil {
			log.Logger.Error("Init sd config fail", log.String("cluster",k), log.Error(tmpErr))
		}
	}
}

func initPrometheusRuleConfig() {
	// Sync rule file
	var tplList []*models.TplTable
	err := x.SQL("SELECT * FROM tpl").Find(&tplList)
	if err != nil {
		log.Logger.Error("init prometheus rule config fail,query tpl table fail", log.Error(err))
		return
	}
	for _,tpl := range tplList {
		tmpErr := SyncRuleConfigFile(tpl.Id, []string{}, false)
		if tmpErr != nil {
			log.Logger.Error("init prometheus rule config fail", log.Error(tmpErr))
		}
	}
}
