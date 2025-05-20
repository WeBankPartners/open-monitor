package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SyncLogMetricExporterConfig(endpoints []string) error {
	log.Info(nil, log.LOGGER_APP, "UpdateNodeExportConfig", zap.Strings("endpoints", endpoints))
	var err error
	existMap := make(map[string]int)
	for _, v := range endpoints {
		if _, b := existMap[v]; b {
			continue
		}
		err = updateEndpointLogMetric(v)
		if err != nil {
			err = fmt.Errorf("Sync endpoint:%s log metric config fail,%s ", v, err.Error())
			log.Error(nil, log.LOGGER_APP, "sync log metric data error", zap.String("endpoint", v), zap.Error(err))
			break
		}
		log.Info(nil, log.LOGGER_APP, "sync log metric data done", zap.String("endpoint", v))
		existMap[v] = 1
	}
	return err
}

func updateEndpointLogMetric(endpointGuid string) error {
	logMetricConfig, err := GetLogMetricByEndpoint(endpointGuid, "", true)
	if err != nil {
		return fmt.Errorf("Query endpoint:%s log metric config fail,%s ", endpointGuid, err.Error())
	}
	log.Debug(nil, log.LOGGER_APP, "sync log metric config data", zap.String("endpoint", endpointGuid), log.JsonObj("logMetricConfig", logMetricConfig))
	syncParam := transLogMetricConfigToJobNew(logMetricConfig, endpointGuid)
	endpointObj := models.EndpointNewTable{Guid: endpointGuid}
	endpointObj, err = GetEndpointNew(&endpointObj)
	if err != nil || endpointObj.AgentAddress == "" {
		return err
	}
	b, _ := json.Marshal(syncParam)
	log.Info(nil, log.LOGGER_APP, "sync log metric data", zap.String("endpoint", endpointGuid), zap.String("body", string(b)))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/log_metric/config", endpointObj.AgentAddress), bytes.NewReader(b))
	timeOutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	req.WithContext(timeOutCtx)
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("Do http request to %s fail,%s ", endpointObj.AgentAddress, respErr.Error())
	}
	b, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Do http request to %s fail,status code:%d ", endpointObj.AgentAddress, resp.StatusCode)
	}
	var response models.LogMetricNodeExporterResponse
	err = json.Unmarshal(b, &response)
	log.Debug(nil, log.LOGGER_APP, "response", zap.String("body", string(b)))
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
			tmpMonitorJob := models.LogMetricMonitorNeObj{Path: lmMonitorObj.LogPath, JsonConfig: []*models.LogMetricJsonNeObj{}, MetricConfig: []*models.LogMetricNeObj{}, ServiceGroup: serviceGroupConfig.Guid}
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
						tmpStringMapJob := models.LogMetricStringMapNeObj{StringValue: vvv.SourceValue, IntValue: targetFloatValue, RegEnable: false, TargetStringValue: vvv.TargetValue}
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
					tmpStringMapJob := models.LogMetricStringMapNeObj{StringValue: vv.SourceValue, IntValue: targetFloatValue, RegEnable: false, TargetStringValue: vv.TargetValue}
					if vv.Regulative > 0 {
						tmpStringMapJob.RegEnable = true
						tmpStringMapJob.Regulation = vv.SourceValue
					}
					tmpMetricJob.StringMap = append(tmpMetricJob.StringMap, &tmpStringMapJob)
				}
				tmpMetricJob.TagConfig = v.TagConfig
				tmpMonitorJob.MetricConfig = append(tmpMonitorJob.MetricConfig, &tmpMetricJob)
			}
			syncParam = append(syncParam, &tmpMonitorJob)
		}
	}
	return syncParam
}

func transLogMetricConfigToJobNew(logMetricConfig []*models.LogMetricQueryObj, endpointGuid string) (syncParam []*models.LogMetricMonitorNeObj) {
	syncParam = []*models.LogMetricMonitorNeObj{}
	for _, serviceGroupConfig := range logMetricConfig {
		for _, lmMonitorObj := range serviceGroupConfig.Config {
			tmpMonitorJob := models.LogMetricMonitorNeObj{Path: lmMonitorObj.LogPath, JsonConfig: []*models.LogMetricJsonNeObj{}, MetricConfig: []*models.LogMetricNeObj{}, MetricGroupConfig: []*models.LogMetricGroupNeObj{}, ServiceGroup: serviceGroupConfig.Guid}
			for _, v := range lmMonitorObj.EndpointRel {
				if v.SourceEndpoint == endpointGuid {
					tmpMonitorJob.TargetEndpoint = v.TargetEndpoint
					break
				}
			}
			for _, v := range lmMonitorObj.MetricGroups {
				if v.UpdateUser == "old_data" {
					continue
				}
				// 禁用,直接跳过
				if strings.ToLower(v.Status) == "disabled" {
					continue
				}
				tmpGroupJob := models.LogMetricGroupNeObj{LogMetricGroup: v.Guid, LogType: v.LogType, JsonRegular: v.JsonRegular, ParamList: []*models.LogMetricParamNeObj{}, MetricConfig: []*models.LogMetricNeObj{}}
				for _, groupParam := range v.ParamList {
					tmpGroupParamObj := models.LogMetricParamNeObj{Name: groupParam.Name, JsonKey: groupParam.JsonKey, Regular: groupParam.Regular, StringMap: []*models.LogMetricStringMapNeObj{}}
					for _, vv := range groupParam.StringMap {
						targetFloatValue, _ := strconv.ParseFloat(vv.TargetValue, 64)
						tmpStringMapJob := models.LogMetricStringMapNeObj{StringValue: vv.SourceValue, IntValue: targetFloatValue, RegEnable: false, TargetStringValue: vv.TargetValue}
						if vv.Regulative > 0 {
							tmpStringMapJob.RegEnable = true
							tmpStringMapJob.Regulation = vv.SourceValue
						}
						tmpGroupParamObj.StringMap = append(tmpGroupParamObj.StringMap, &tmpStringMapJob)
					}
					tmpGroupJob.ParamList = append(tmpGroupJob.ParamList, &tmpGroupParamObj)
				}
				for _, groupMetric := range v.MetricList {
					if groupMetric.AggType != "avg" && groupMetric.AggType != "count" && groupMetric.AggType != "sum" && groupMetric.AggType != "max" && groupMetric.AggType != "min" {
						continue
					}
					tmpMetric := groupMetric.Metric
					if v.MetricPrefixCode != "" {
						tmpMetric = v.MetricPrefixCode + "_" + groupMetric.Metric
					}
					tmpGroupMetricObj := models.LogMetricNeObj{Metric: tmpMetric, LogParamName: groupMetric.LogParamName, AggType: groupMetric.AggType, Step: groupMetric.Step, TagConfig: []*models.LogMetricConfigTag{}}
					for _, vv := range groupMetric.TagConfigList {
						tmpGroupMetricObj.TagConfig = append(tmpGroupMetricObj.TagConfig, &models.LogMetricConfigTag{LogParamName: vv})
					}
					tmpGroupJob.MetricConfig = append(tmpGroupJob.MetricConfig, &tmpGroupMetricObj)
				}
				tmpMonitorJob.MetricGroupConfig = append(tmpMonitorJob.MetricGroupConfig, &tmpGroupJob)
			}
			for _, v := range lmMonitorObj.MetricConfigList {
				tmpMetricJob := models.LogMetricNeObj{Metric: v.Metric, ValueRegular: v.Regular, AggType: v.AggType, Step: v.Step, StringMap: []*models.LogMetricStringMapNeObj{}}
				for _, vv := range v.StringMap {
					targetFloatValue, _ := strconv.ParseFloat(vv.TargetValue, 64)
					tmpStringMapJob := models.LogMetricStringMapNeObj{StringValue: vv.SourceValue, IntValue: targetFloatValue, RegEnable: false, TargetStringValue: vv.TargetValue}
					if vv.Regulative > 0 {
						tmpStringMapJob.RegEnable = true
						tmpStringMapJob.Regulation = vv.SourceValue
					}
					tmpMetricJob.StringMap = append(tmpMetricJob.StringMap, &tmpStringMapJob)
				}
				tmpMetricJob.TagConfig = v.TagConfig
				tmpMonitorJob.MetricConfig = append(tmpMonitorJob.MetricConfig, &tmpMetricJob)
			}
			syncParam = append(syncParam, &tmpMonitorJob)
		}
	}
	return syncParam
}

func SyncLogKeywordExporterConfig(endpoints []string) error {
	log.Info(nil, log.LOGGER_APP, "UpdateNodeExportConfig", zap.Strings("endpoints", endpoints))
	var err error
	existMap := make(map[string]int)
	for _, v := range endpoints {
		if _, b := existMap[v]; b {
			continue
		}
		err = updateEndpointLogKeyword(v)
		if err != nil {
			err = fmt.Errorf("Sync endpoint:%s log keyword config fail,%s ", v, err.Error())
			break
		}
		existMap[v] = 1
	}
	return err
}

func updateEndpointLogKeyword(endpoint string) error {
	syncParam, err := getLogKeywordExporterConfig(endpoint)
	if err != nil {
		return err
	}
	endpointObj := models.EndpointNewTable{Guid: endpoint}
	endpointObj, err = GetEndpointNew(&endpointObj)
	if err != nil || endpointObj.AgentAddress == "" {
		return err
	}
	b, _ := json.Marshal(syncParam)
	log.Info(nil, log.LOGGER_APP, "sync log keyword data", zap.String("endpoint", endpoint), zap.String("body", string(b)))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/log_keyword/config", endpointObj.AgentAddress), bytes.NewReader(b))
	timeOutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	req.WithContext(timeOutCtx)
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("Do http request to %s fail,%s ", endpointObj.AgentAddress, respErr.Error())
	}
	b, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Do http request to %s fail,status code:%d ", endpointObj.AgentAddress, resp.StatusCode)
	}
	var response models.LogKeywordHttpResult
	err = json.Unmarshal(b, &response)
	log.Debug(nil, log.LOGGER_APP, "response", zap.String("body", string(b)))
	if err == nil {
		if response.Status == "OK" {
			return nil
		} else {
			return fmt.Errorf(response.Message)
		}
	}
	return fmt.Errorf("json unmarhsal reponse body fail,%s ", err.Error())
}

func getLogKeywordExporterConfig(endpoint string) (result []*models.LogKeywordHttpDto, err error) {
	serviceGroupKeywordList, queryConfigErr := GetLogKeywordByEndpoint(endpoint, "", true)
	if queryConfigErr != nil {
		return result, queryConfigErr
	}
	result = []*models.LogKeywordHttpDto{}
	var pathList []string
	pathMap := make(map[string][]*models.LogKeywordHttpRuleObj)
	for _, serviceGroupConfig := range serviceGroupKeywordList {
		for _, logKeywordMonitor := range serviceGroupConfig.Config {
			targetEndpoint := ""
			for _, endpointRel := range logKeywordMonitor.EndpointRel {
				if endpointRel.SourceEndpoint == endpoint {
					targetEndpoint = endpointRel.TargetEndpoint
					break
				}
			}
			if existKeywordList, b := pathMap[logKeywordMonitor.LogPath]; b {
				existKeywordMap := make(map[string]string)
				for _, existKeywordObj := range existKeywordList {
					existKeywordMap[existKeywordObj.Keyword] = existKeywordObj.TargetEndpoint
				}
				tmpKeywordList := existKeywordList
				for _, logKeywordConfig := range logKeywordMonitor.KeywordList {
					if existTarget, sameFlag := existKeywordMap[logKeywordConfig.Keyword]; sameFlag {
						if existTarget == targetEndpoint {
							// path keyword target is same
							//err = fmt.Errorf("Endpint:%s Path:%s keyword:%s duplicated ", endpoint, logKeywordMonitor.LogPath, logKeywordConfig.Keyword)
							//return
							continue
						}
					}
					tmpKeywordObj := models.LogKeywordHttpRuleObj{Keyword: logKeywordConfig.Keyword, TargetEndpoint: targetEndpoint, RegularEnable: false}
					if logKeywordConfig.Regulative > 0 {
						tmpKeywordObj.RegularEnable = true
					}
					tmpKeywordList = append(tmpKeywordList, &tmpKeywordObj)
				}
				pathMap[logKeywordMonitor.LogPath] = tmpKeywordList
			} else {
				pathList = append(pathList, logKeywordMonitor.LogPath)
				tmpKeywordList := []*models.LogKeywordHttpRuleObj{}
				for _, logKeywordConfig := range logKeywordMonitor.KeywordList {
					tmpKeywordObj := models.LogKeywordHttpRuleObj{Keyword: logKeywordConfig.Keyword, TargetEndpoint: targetEndpoint, RegularEnable: false}
					if logKeywordConfig.Regulative > 0 {
						tmpKeywordObj.RegularEnable = true
					}
					tmpKeywordList = append(tmpKeywordList, &tmpKeywordObj)
				}
				pathMap[logKeywordMonitor.LogPath] = tmpKeywordList
			}
		}
	}
	for _, path := range pathList {
		result = append(result, &models.LogKeywordHttpDto{Path: path, Keywords: pathMap[path]})
	}
	return
}
