package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"io/ioutil"
	"net/http"
	"strconv"
)

func SyncLogMetricExporterConfig(endpoints []string) error {
	log.Logger.Info("UpdateNodeExportConfig", log.StringList("endpoints", endpoints))
	var err error
	existMap := make(map[string]int)
	for _, v := range endpoints {
		if _, b := existMap[v]; b {
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
	logMetricConfig, err := GetLogMetricByEndpoint(endpointGuid)
	if err != nil {
		return fmt.Errorf("Query endpoint:%s log metric config fail,%s ", endpointGuid, err.Error())
	}
	syncParam := transLogMetricConfigToJob(logMetricConfig, endpointGuid)
	endpointObj := models.EndpointNewTable{Guid: endpointGuid}
	endpointObj, err = GetEndpointNew(&endpointObj)
	log.Logger.Info("get endpoint new", log.JsonObj("endpoint", endpointObj), log.Error(err))
	if err != nil || endpointObj.AgentAddress == "" {
		return err
	}
	b, _ := json.Marshal(syncParam)
	log.Logger.Info("sync log metric data", log.String("body", string(b)))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/log_metric/config", endpointObj.AgentAddress), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("Do http request to %s fail,%s ", endpointObj.AgentAddress, respErr.Error())
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Do http request to %s fail,status code:%d ", endpointObj.AgentAddress, resp.StatusCode)
	}
	b, _ = ioutil.ReadAll(resp.Body)
	var response models.LogMetricNodeExporterResponse
	err = json.Unmarshal(b, &response)
	log.Logger.Info("response", log.String("body", string(b)))
	if err == nil {
		if response.Status == "OK" {
			return nil
		} else {
			return fmt.Errorf(response.Message)
		}
	}
	return fmt.Errorf("json unmarhsal reponse body fail,%s ", err.Error())
}

func transLogMetricConfigToJob(logMetricConfig []*models.LogMetricQueryObj, endpointGuid string) (syncParam []*models.LogMetricMonitorNeObj) {
	syncParam = []*models.LogMetricMonitorNeObj{}
	for _, serviceGroupConfig := range logMetricConfig {
		for _, lmMonitorObj := range serviceGroupConfig.Config {
			tmpMonitorJob := models.LogMetricMonitorNeObj{Path: lmMonitorObj.LogPath, JsonConfig: []*models.LogMetricJsonNeObj{}, MetricConfig: []*models.LogMetricNeObj{}}
			for _, v := range lmMonitorObj.EndpointRel {
				if v.SourceEndpoint == endpointGuid {
					tmpMonitorJob.TargetEndpoint = v.TargetEndpoint
					break
				}
			}
			for _, v := range lmMonitorObj.JsonConfigList {
				tmpJsonJob := models.LogMetricJsonNeObj{Regular: v.JsonRegular, Tags: v.Tags, MetricConfig: []*models.LogMetricNeObj{}}
				for _, vv := range v.MetricList {
					tmpMetricJob := models.LogMetricNeObj{Metric: vv.Metric, Key: vv.JsonKey, AggType: vv.AggType, Step: vv.Step, StringMap: []*models.LogMetricStringMapNeObj{}}
					for _, vvv := range vv.StringMap {
						targetFloatValue, _ := strconv.ParseFloat(vvv.TargetValue, 64)
						tmpStringMapJob := models.LogMetricStringMapNeObj{StringValue: vvv.SourceValue, IntValue: targetFloatValue, RegEnable: false}
						if vvv.Regulative > 0 {
							tmpStringMapJob.RegEnable = true
							tmpStringMapJob.Regulation = vvv.SourceValue
						}
						tmpMetricJob.StringMap = append(tmpMetricJob.StringMap, &tmpStringMapJob)
					}
					tmpJsonJob.MetricConfig = append(tmpJsonJob.MetricConfig, &tmpMetricJob)
				}
				tmpMonitorJob.JsonConfig = append(tmpMonitorJob.JsonConfig, &tmpJsonJob)
			}
			for _, v := range lmMonitorObj.MetricConfigList {
				tmpMetricJob := models.LogMetricNeObj{Metric: v.Metric, ValueRegular: v.Regular, AggType: v.AggType, Step: v.Step, StringMap: []*models.LogMetricStringMapNeObj{}}
				for _, vv := range v.StringMap {
					targetFloatValue, _ := strconv.ParseFloat(vv.TargetValue, 64)
					tmpStringMapJob := models.LogMetricStringMapNeObj{StringValue: vv.SourceValue, IntValue: targetFloatValue, RegEnable: false}
					if vv.Regulative > 0 {
						tmpStringMapJob.RegEnable = true
						tmpStringMapJob.Regulation = vv.SourceValue
					}
					tmpMetricJob.StringMap = append(tmpMetricJob.StringMap, &tmpStringMapJob)
				}
				tmpMonitorJob.MetricConfig = append(tmpMonitorJob.MetricConfig, &tmpMetricJob)
			}
			syncParam = append(syncParam, &tmpMonitorJob)
		}
	}
	return syncParam
}

func SyncLogKeywordExporterConfig(endpoints []string) error {
	return nil
}