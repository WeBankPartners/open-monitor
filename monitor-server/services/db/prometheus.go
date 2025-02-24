package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
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
		log.Error(nil, log.LOGGER_APP, "Start sync kubernetes config fail", zap.Error(err))
	}
	// init snmp config
	err = SyncSnmpPrometheusConfig()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Start sync snmp config fail", zap.Error(err))
	}
	// init snmp config
	err = SyncRemoteWritePrometheusConfig()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Start sync remote write config fail", zap.Error(err))
	}
	select {}
}

func startCheckPrometheusConfig() {
	tMin, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s00", time.Now().Format("2006-01-02 15:04:")), time.Local)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Start check prometheus config job init fail", zap.Error(err))
		return
	}
	sleepTime := tMin.Unix() + 60 - time.Now().Unix()
	if sleepTime < 0 {
		log.Warn(nil, log.LOGGER_APP, "Start check prometheus config job fail,calc sleep time fail", zap.Int64("sleep time", sleepTime))
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
		log.Error(nil, log.LOGGER_APP, "checkSdConfigTime query endpoint fail", zap.Error(err))
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
			log.Error(nil, log.LOGGER_APP, "Sync sd config fail", zap.String("cluster", k), zap.String("steps", fmt.Sprintf("%v", v)), zap.Error(tmpErr))
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
		log.Error(nil, log.LOGGER_APP, "Init prometheus rule config fail,query endpoint group fail", zap.Error(err))
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
			log.Error(nil, log.LOGGER_APP, "init prometheus rule config fail", zap.String("endpointGroup", v.Guid), zap.Error(tmpErr))
		}
	}
	if len(endpointGroup) > 0 {
		prom.ReloadConfig()
	}
	ruleConfigSyncTime = time.Now().Unix()
}

func QueryExporterMetric(param models.QueryPrometheusMetricParam) (err error, result []string) {
	log.Info(nil, log.LOGGER_APP, "QueryExporterMetric", log.JsonObj("param", param))
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
			log.Error(nil, log.LOGGER_APP, "Try go get tGuid fail", zap.String("service_group", param.ServiceGroup), zap.Error(err))
		} else {
			log.Info(nil, log.LOGGER_APP, "tGuid tmpMetricList", zap.Strings("tmpMetricList", tmpMetricList))
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
