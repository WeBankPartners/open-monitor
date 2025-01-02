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

var (
	sdConfigSyncTime   int64
	ruleConfigSyncTime int64
)

func InitPrometheusConfig() {
	var err error
	// init sd config
	go prom.StartConsumeSdConfig()
	sdConfigSyncTime = time.Now().Unix()
	checkSdConfigTime(true)
	// init rule config
	go prom.StartConsumeRuleConfig()
	ruleConfigSyncTime = time.Now().Unix()
	checkRuleConfigTime(true)
	// start cron check
	go startCheckPrometheusConfig()
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
	// init snmp config
	err = SyncRemoteWritePrometheusConfig()
	if err != nil {
		log.Logger.Error("Start sync remote write config fail", log.Error(err))
	}
	select {}
}

func startCheckPrometheusConfig() {
	tMin, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s00", time.Now().Format("2006-01-02 15:04:")), time.Local)
	if err != nil {
		log.Logger.Error("Start check prometheus config job init fail", log.Error(err))
		return
	}
	sleepTime := tMin.Unix() + 60 - time.Now().Unix()
	if sleepTime < 0 {
		log.Logger.Warn("Start check prometheus config job fail,calc sleep time fail", log.Int64("sleep time", sleepTime))
		return
	}
	time.Sleep(time.Duration(sleepTime) * time.Second)
	jobCount := 0
	refreshAll := false
	t := time.NewTicker(time.Second * time.Duration(60)).C
	for {
		<-t
		if jobCount == 1440 {
			refreshAll = true
			jobCount = 0
		} else {
			refreshAll = false
		}
		go checkSdConfigTime(refreshAll)
		go checkRuleConfigTime(refreshAll)
		jobCount = jobCount + 1
	}
}

func checkSdConfigTime(refreshAll bool) {
	querySql := "select step,cluster from endpoint_new group by step,cluster order by cluster,step"
	if !refreshAll {
		querySql = "select step,cluster from endpoint_new where update_time>'" + time.Unix(sdConfigSyncTime, 0).Format(models.DatetimeFormat) + "' group by step,cluster order by cluster,step"
	}
	var endpointTable []*models.EndpointNewTable
	err := x.SQL(querySql).Find(&endpointTable)
	if err != nil {
		log.Logger.Error("checkSdConfigTime query endpoint fail", log.Error(err))
		return
	}
	if len(endpointTable) == 0 {
		sdConfigSyncTime = time.Now().Unix()
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
			log.Logger.Error("Sync sd config fail", log.String("cluster", k), log.String("steps", fmt.Sprintf("%v", v)), log.Error(tmpErr))
		}
	}
	sdConfigSyncTime = time.Now().Unix()
}

func checkRuleConfigTime(refreshAll bool) {
	var endpointGroup []*models.EndpointGroupTable
	querySql := "select guid from endpoint_group"
	if !refreshAll {
		nowTime := time.Unix(ruleConfigSyncTime, 0).Format(models.DatetimeFormat)
		querySql = fmt.Sprintf("select endpoint_group as 'guid' from alarm_strategy where update_time>'%s' union select guid from endpoint_group where update_time>'%s' union select guid from endpoint_group where service_group in (select guid from service_group where update_time>'%s')", nowTime, nowTime, nowTime)
	}
	err := x.SQL(querySql).Find(&endpointGroup)
	if err != nil {
		log.Logger.Error("Init prometheus rule config fail,query endpoint group fail", log.Error(err))
		return
	}
	existMap := make(map[string]int)
	for _, v := range endpointGroup {
		if _, b := existMap[v.Guid]; b {
			continue
		}
		existMap[v.Guid] = 1
		tmpErr := SyncPrometheusRuleFile(v.Guid, true)
		if tmpErr != nil {
			log.Logger.Error("init prometheus rule config fail", log.String("endpointGroup", v.Guid), log.Error(tmpErr))
		}
	}
	if len(endpointGroup) > 0 {
		prom.ReloadConfig()
	}
	ruleConfigSyncTime = time.Now().Unix()
}

func QueryExporterMetric(param models.QueryPrometheusMetricParam) (err error, result []string) {
	log.Logger.Info("QueryExporterMetric", log.JsonObj("param", param))
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
	var metricList, tmpMetricList []string
	var queryPromQl string
	nowTime := time.Now().Unix()
	if param.ServiceGroup != "" {
		queryPromQl = fmt.Sprintf("{service_group=\"%s\"}", param.ServiceGroup)
		tmpMetricList, err = datasource.QueryPromQLMetric(queryPromQl, clusterAddress, nowTime-120, nowTime)
		if err != nil {
			log.Logger.Error("Try go get tGuid fail", log.String("service_group", param.ServiceGroup), log.Error(err))
		} else {
			log.Logger.Info("tGuid tmpMetricList", log.StringList("tmpMetricList", tmpMetricList))
		}
	}
	queryPromQl = fmt.Sprintf("{instance=\"%s:%s\"}", param.Ip, param.Port)
	metricList, err = datasource.QueryPromQLMetric(queryPromQl, clusterAddress, nowTime-120, nowTime)
	if err != nil {
		err = fmt.Errorf("Try to query remote cluster data fail,%s ", err.Error())
		return
	}
	if len(tmpMetricList) > 0 {
		metricList = append(metricList, tmpMetricList...)
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
