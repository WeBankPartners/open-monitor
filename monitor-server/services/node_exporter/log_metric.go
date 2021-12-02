package node_exporter

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
)

func UpdateNodeExportConfig(endpoints []string) error {
	var err error
	existMap := make(map[string]int)
	for _,v := range endpoints {
		if _,b:=existMap[v];b {
			continue
		}
		err = updateEndpointLogMetric(v)
		if err != nil {
			err = fmt.Errorf("Sync endpoint:%s log metric config fail,%s ", v, err.Error())
			break
		}
		existMap[v] = 1
	}
	return err
}

func updateEndpointLogMetric(endpointGuid string) error {
	logMetricConfig,err := db.GetLogMetricByEndpoint(endpointGuid)
	if err != nil {
		return fmt.Errorf("Query endpoint:%s log metric config fail,%s ", endpointGuid, err.Error())
	}
	endpointObj := models.EndpointTable{Guid: endpointGuid}
	db.GetEndpoint(&endpointObj)
	log.Logger.Debug("sync log metric config", log.String("endpoint", endpointGuid), log.JsonObj("config", logMetricConfig))
	return nil
}