package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"strings"
	"time"
)

func InitPrometheusConfig() {
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

func InitPrometheusServiceDiscoverConfig() {
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
	for _, endpoint := range endpointTable {
		if _, b := clusterStepMap[endpoint.Cluster]; !b {
			clusterStepMap[endpoint.Cluster] = []int{endpoint.Step}
		} else {
			clusterStepMap[endpoint.Cluster] = append(clusterStepMap[endpoint.Cluster], endpoint.Step)
		}
	}
	for k, v := range clusterStepMap {
		tmpErr := SyncSdEndpointNew(v, k, true)
		if tmpErr != nil {
			log.Logger.Error("Init sd config fail", log.String("cluster", k), log.Error(tmpErr))
		}
	}
}

func initPrometheusRuleConfig() {
	// Sync rule file
	//var tplList []*models.TplTable
	//err := x.SQL("SELECT * FROM tpl").Find(&tplList)
	//if err != nil {
	//	log.Logger.Error("init prometheus rule config fail,query tpl table fail", log.Error(err))
	//	return
	//}
	//for _, tpl := range tplList {
	//	tmpErr := SyncRuleConfigFile(tpl.Id, []string{}, false)
	//	if tmpErr != nil {
	//		log.Logger.Error("init prometheus rule config fail", log.Error(tmpErr))
	//	}
	//}

	var endpointGroup []*models.EndpointGroupTable
	err := x.SQL("select guid from endpoint_group").Find(&endpointGroup)
	if err != nil {
		log.Logger.Error("Init prometheus rule config fail,query endpoint group fail", log.Error(err))
		return
	}
	for _,v := range endpointGroup {
		tmpErr := SyncPrometheusRuleFile(v.Guid, false)
		if tmpErr != nil {
			log.Logger.Error("init prometheus rule config fail", log.String("endpointGroup", v.Guid), log.Error(tmpErr))
		}
	}
}

func QueryExporterMetric(param models.QueryPrometheusMetricParam) (err error, result []string) {
	if !param.IsConfigQuery {
		if param.Cluster == "" || param.Cluster == "default" {
			err, result = prom.GetEndpointData(param)
			return
		}
	}
	clusterAddress := GetClusterAddress(param.Cluster)
	if clusterAddress == "" {
		err = fmt.Errorf("Can not find cluster address with cluster:%s ", param.Cluster)
		return
	}
	queryPromQl := fmt.Sprintf("{instance=\"%s:%s\"}", param.Ip, param.Port)
	if param.ProcessGuid != "" {
		queryPromQl = fmt.Sprintf("{process_guid=\"%s\"}", param.ProcessGuid)
	}else if param.EndpointGuid != "" {
		queryPromQl = fmt.Sprintf("{e_guid=\"%s\"}", param.EndpointGuid)
	}
	nowTime := time.Now().Unix()
	metricList, queryErr := datasource.QueryPromQLMetric(queryPromQl, clusterAddress, nowTime-120, nowTime)
	if queryErr != nil {
		err = fmt.Errorf("Try to query remote cluster data fail,%s ", queryErr.Error())
		return
	}
	metricMap := make(map[string]int)
	prefixLen := len(param.Prefix)
	keywordLen := len(param.Keyword)
	for _, metric := range metricList {
		if prefixLen == 0 && keywordLen == 0 {
			if _, b := metricMap[metric]; !b {
				result = append(result, metric)
				metricMap[metric] = 1
			}
			continue
		}
		if prefixLen > 0 {
			prefixFlag := false
			for _, prefix := range param.Prefix {
				if strings.HasPrefix(metric, prefix) {
					prefixFlag = true
					break
				}
			}
			if prefixFlag {
				if _, b := metricMap[metric]; !b {
					result = append(result, metric)
					metricMap[metric] = 1
				}
				continue
			}
		}
		if keywordLen > 0 {
			keywordFlag := false
			for _, keyword := range param.Keyword {
				if strings.Contains(metric, keyword) {
					keywordFlag = true
					break
				}
			}
			if keywordFlag {
				if _, b := metricMap[metric]; !b {
					result = append(result, metric)
					metricMap[metric] = 1
				}
			}
		}
	}
	return
}
