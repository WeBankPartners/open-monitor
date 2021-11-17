package collector

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	businessCollectorName   = "business_monitor"
	businessMonitorFilePath = "data/business_monitor_cache.data"
)

var (
	businessMonitorJobs       []*businessMonitorObj
	businessMonitorLock       = new(sync.RWMutex)
	businessMonitorMetrics    []*businessRuleMetricObj
	businessMonitorMetricLock = new(sync.RWMutex)
	newLogger                 log.Logger
)

type businessMonitorCollector struct {
	businessMonitor *prometheus.Desc
	logger          log.Logger
}

func InitNewLogger(logger log.Logger) {
	newLogger = logger
}

func init() {
	registerCollector(businessCollectorName, defaultEnabled, BusinessMonitorCollector)
}

func BusinessMonitorCollector(logger log.Logger) (Collector, error) {
	return &businessMonitorCollector{
		businessMonitor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, businessCollectorName, "value"),
			"Show business data from log file.",
			[]string{"key", "tags", "path", "agg"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *businessMonitorCollector) Update(ch chan<- prometheus.Metric) error {
	businessMonitorMetricLock.RLock()
	for _, v := range businessMonitorMetrics {
		ch <- prometheus.MustNewConstMetric(c.businessMonitor,
			prometheus.GaugeValue,
			v.Value, v.Metric, v.TagsString, v.Path, v.Agg)
	}
	businessMonitorMetricLock.RUnlock()
	return nil
}

type businessStoreMonitorObj struct {
	Path  string                    `json:"path"`
	Rules []*businessStoreMetricObj `json:"rules"`
}

type businessStoreMetricObj struct {
	Regular      string                     `json:"regular"`
	StringMap    []*businessStringMapObj    `json:"string_map"`
	TagsString   string                     `json:"tags_string"`
	TagsKey      []string                   `json:"tags_key"`
	TagsValue    []string                   `json:"tags_value"`
	MetricConfig []*businessMetricConfigObj `json:"metric_config"`
}

type businessRuleObj struct {
	Regular      string                      `json:"regular"`
	RegExp       *regexp.Regexp              `json:"-"`
	StringMap    []*businessStringMapObj     `json:"string_map"`
	TagsString   string                      `json:"tags_string"`
	TagsKey      []string                    `json:"tags_key"`
	TagsValue    []string                    `json:"tags_value"`
	MetricConfig []*businessMetricConfigObj  `json:"metric_config"`
	DataChannel  chan map[string]interface{} `json:"-"`
}

type businessCustomObj struct {
	RegExp       *regexp.Regexp              `json:"-"`
	Metric string `json:"metric"`
	ValueRegular string `json:"value_regular"`
	AggType string `json:"agg_type"`
	StringMap    []*businessStringMapObj     `json:"string_map"`
	DataChannel  chan string `json:"-"`
}

type businessRuleMetricObj struct {
	Metric     string   `json:"metric"`
	Path       string   `json:"path"`
	Agg        string   `json:"agg"`
	Tags       []string `json:"tags"`
	TagsString string   `json:"tags_string"`
	Value      float64  `json:"value"`
}

type businessMonitorObj struct {
	Path        string             `json:"path"`
	TailSession *tail.Tail         `json:"-"`
	Lock        *sync.RWMutex      `json:"-"`
	Rules       []*businessRuleObj `json:"rules"`
	Custom      []*businessCustomObj `json:"custom"`
}

type businessStringMapObj struct {
	Key         string  `json:"key"`
	Regulation  string  `json:"regulation"`
	StringValue string  `json:"string_value"`
	IntValue    float64 `json:"int_value"`
}

type businessStringMapRegexpObj struct {
	Key         string  `json:"key"`
	Regulation  string  `json:"regulation"`
	StringValue string  `json:"string_value"`
	IntValue    float64 `json:"int_value"`
	RegEnable   bool
	Regexp      pcre.Regexp
}

type businessMetricConfigObj struct {
	Key     string `json:"key"`
	Metric  string `json:"metric"`
	AggType string `json:"agg_type"`
}

type businessMonitorCfgObj struct {
	Regular      string                     `json:"regular"`
	Tags         string                     `json:"tags"`
	StringMap    []*businessStringMapObj    `json:"string_map"`
	MetricConfig []*businessMetricConfigObj `json:"metric_config"`
}

type BusinessMonitorCustomObj struct {
	Id           int                     `json:"id"`
	Metric       string                  `json:"metric"`
	ValueRegular string                  `json:"value_regular"`
	AggType      string                  `json:"agg_type"`
	StringMap    []*businessStringMapObj `json:"string_map"`
}

type businessAgentDto struct {
	Path   string                      `json:"path"`
	Config []*businessMonitorCfgObj    `json:"config"`
	Custom []*BusinessMonitorCustomObj `json:"custom"`
}

func (c *businessMonitorObj) start() {
	var err error
	c.TailSession, err = tail.TailFile(c.Path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		level.Error(newLogger).Log("msg", fmt.Sprintf("start business collector fail, path: %s, error: %v", c.Path, err))
		return
	}
	firstFlag := true
	timeNow := time.Now()
	for line := range c.TailSession.Lines {
		if firstFlag {
			if time.Now().Sub(timeNow).Seconds() >= 5 {
				firstFlag = false
			} else {
				continue
			}
		}
		c.Lock.RLock()
		for _, rule := range c.Rules {
			fetchList := rule.RegExp.FindStringSubmatch(line.Text)
			if len(fetchList) > 1 {
				fetchKeyMap := make(map[string]interface{})
				for i, v := range fetchList {
					if i == 0 {
						continue
					}
					tmpKeyMap := make(map[string]interface{})
					tmpErr := json.Unmarshal([]byte(v), &tmpKeyMap)
					if tmpErr != nil {
						level.Error(newLogger).Log("line fetch regexp fail", fmt.Sprintf("line:%s error:%s", v, tmpErr.Error()))
					} else {
						for tmpKeyMapKey, tmpKeyMapValue := range tmpKeyMap {
							fetchKeyMap[tmpKeyMapKey] = tmpKeyMapValue
						}
					}
				}
				if len(fetchKeyMap) > 0 {
					rule.DataChannel <- fetchKeyMap
				}
			}
		}
		for _, custom := range c.Custom {
			fetchList := custom.RegExp.FindStringSubmatch(line.Text)
			if len(fetchList) > 0 {
				level.Info(newLogger).Log("line fetch custom regexp", fmt.Sprintf("find string:%s ", fetchList))
				if fetchList[0] != "" {
					custom.DataChannel <- fetchList[0]
				}
			}
		}
		c.Lock.RUnlock()
	}
}

func (c *businessMonitorObj) destroy() {
	c.TailSession.Stop()
	c.Rules = []*businessRuleObj{}
}

func BusinessMonitorHttpHandle(w http.ResponseWriter, r *http.Request) {
	buff, err := ioutil.ReadAll(r.Body)
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Handel business monitor http request fail,read body error: %v \n", err)
		level.Error(newLogger).Log("msg", errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	var param []*businessAgentDto
	level.Info(newLogger).Log("http_param", string(buff))
	err = json.Unmarshal(buff, &param)
	if err != nil {
		errorMsg = fmt.Sprintf("Handel business monitor http request fail,json unmarshal error: %v \n", err)
		level.Error(newLogger).Log("msg", errorMsg)
		w.Write([]byte(errorMsg))
		return
	}
	businessMonitorLock.Lock()
	var newBmj []*businessMonitorObj
	for _, v := range businessMonitorJobs {
		delFlag := true
		for _, vv := range param {
			if vv.Path == v.Path {
				delFlag = false
				v.Lock.Lock()
				updateBusinessRules(v, vv)
				updateBusinessCustomConfig(v, vv)
				v.Lock.Unlock()
				break
			}
		}
		if delFlag {
			v.destroy()
		} else {
			newBmj = append(newBmj, v)
		}
	}
	businessMonitorJobs = newBmj
	for _, v := range param {
		addFlag := true
		for _, vv := range businessMonitorJobs {
			if vv.Path == v.Path {
				addFlag = false
				break
			}
		}
		if addFlag {
			newBmo := businessMonitorObj{}
			newBmo.Path = v.Path
			newBmo.Lock = new(sync.RWMutex)
			for _, vv := range v.Config {
				tmpRuleObj := businessRuleObj{}
				tmpRuleObj.StringMap = vv.StringMap
				tmpRuleObj.MetricConfig = vv.MetricConfig
				tmpRuleObj.Regular = vv.Regular
				tmpRuleObj.RegExp = regexp.MustCompile(vv.Regular)
				tmpRuleObj.TagsString = vv.Tags
				var tmpTagsKey, tmpTagsValue []string
				for _, tmpKey := range strings.Split(vv.Tags, ",") {
					tmpTagsKey = append(tmpTagsKey, tmpKey)
					tmpTagsValue = append(tmpTagsValue, "")
				}
				tmpRuleObj.TagsKey = tmpTagsKey
				tmpRuleObj.TagsValue = tmpTagsValue
				tmpRuleObj.DataChannel = make(chan map[string]interface{}, 10000)
				newBmo.Rules = append(newBmo.Rules, &tmpRuleObj)
			}
			for _,vv := range v.Custom {
				tmpCustomObj := businessCustomObj{Metric: vv.Metric,AggType: vv.AggType,ValueRegular: vv.ValueRegular}
				tmpCustomObj.RegExp = regexp.MustCompile(vv.ValueRegular)
				tmpCustomObj.StringMap = vv.StringMap
				tmpCustomObj.DataChannel = make(chan string, 10000)
				newBmo.Custom = append(newBmo.Custom, &tmpCustomObj)
			}
			go newBmo.start()
			businessMonitorJobs = append(businessMonitorJobs, &newBmo)
		}
	}
	businessMonitorLock.Unlock()
	level.Info(newLogger).Log("msg", "success")
	w.Write([]byte("success"))
}

func updateBusinessRules(bmo *businessMonitorObj, config *businessAgentDto) {
	var newRules []*businessRuleObj
	for _, v := range bmo.Rules {
		delFlag := true
		for _, vv := range config.Config {
			if vv.Regular == v.Regular && vv.Tags == v.TagsString {
				delFlag = false
				v.StringMap = vv.StringMap
				v.MetricConfig = vv.MetricConfig
				var newTagsKey, newTagsValue []string
				for _, cfgKey := range strings.Split(vv.Tags, ",") {
					newTagsKey = append(newTagsKey, cfgKey)
					keyExistFlag := false
					for existKeyIndex, existKey := range v.TagsKey {
						if existKey == cfgKey {
							keyExistFlag = true
							newTagsValue = append(newTagsValue, v.TagsValue[existKeyIndex])
							break
						}
					}
					if !keyExistFlag {
						newTagsValue = append(newTagsValue, "")
					}
				}
				v.TagsKey = newTagsKey
				v.TagsValue = newTagsValue
				break
			}
		}
		if !delFlag {
			newRules = append(newRules, v)
		}
	}
	for _, v := range config.Config {
		addFlag := true
		for _, vv := range newRules {
			if v.Regular == vv.Regular && v.Tags == vv.TagsString {
				addFlag = false
				break
			}
		}
		if addFlag {
			tmpRuleObj := businessRuleObj{}
			tmpRuleObj.StringMap = v.StringMap
			tmpRuleObj.MetricConfig = v.MetricConfig
			tmpRuleObj.Regular = v.Regular
			tmpRuleObj.RegExp = regexp.MustCompile(v.Regular)
			tmpRuleObj.TagsString = v.Tags
			var tmpTagsKey, tmpTagsValue []string
			for _, tmpKey := range strings.Split(v.Tags, ",") {
				tmpTagsKey = append(tmpTagsKey, tmpKey)
				tmpTagsValue = append(tmpTagsValue, "")
			}
			tmpRuleObj.TagsKey = tmpTagsKey
			tmpRuleObj.TagsValue = tmpTagsValue
			tmpRuleObj.DataChannel = make(chan map[string]interface{}, 10000)
			newRules = append(newRules, &tmpRuleObj)
		}
	}
	printByte, _ := json.Marshal(newRules)
	level.Info(newLogger).Log("updateBusinessRules", string(printByte))
	bmo.Rules = newRules
}

func updateBusinessCustomConfig(bmo *businessMonitorObj, config *businessAgentDto)  {
	level.Info(newLogger).Log("StartUpdateBusiness", "custom")
	var newCustom []*businessCustomObj
	for _,v := range bmo.Custom {
		delFlag := true
		for _,vv := range config.Custom {
			if v.Metric == vv.Metric {
				delFlag = false
				v.StringMap = vv.StringMap
				v.AggType = vv.AggType
				if v.ValueRegular != vv.ValueRegular {
					v.RegExp = regexp.MustCompile(vv.ValueRegular)
					v.ValueRegular = vv.ValueRegular
				}
				break
			}
		}
		if !delFlag {
			newCustom = append(newCustom, v)
		}
	}
	for _,v := range config.Custom {
		addFlag := true
		for _,vv := range newCustom {
			if v.Metric == vv.Metric {
				addFlag = false
				break
			}
		}
		if addFlag {
			tmpCustomObj := businessCustomObj{Metric: v.Metric,AggType: v.AggType,ValueRegular: v.ValueRegular}
			tmpCustomObj.RegExp = regexp.MustCompile(v.ValueRegular)
			tmpCustomObj.StringMap = v.StringMap
			tmpCustomObj.DataChannel = make(chan string, 10000)
			newCustom = append(newCustom, &tmpCustomObj)
		}
	}
	printByte, _ := json.Marshal(newCustom)
	level.Info(newLogger).Log("updateBusinessCustoms", string(printByte))
	bmo.Custom = newCustom
}

type businessCollectorStore struct {
	Data []*businessStoreMonitorObj `json:"data"`
}

var BusinessCollectorStore businessCollectorStore

func (c *businessCollectorStore) Save() {
	for _, v := range businessMonitorJobs {
		var newStoreRules []*businessStoreMetricObj
		for _, vv := range v.Rules {
			newStoreRules = append(newStoreRules, &businessStoreMetricObj{Regular: vv.Regular, StringMap: vv.StringMap, MetricConfig: vv.MetricConfig, TagsKey: vv.TagsKey, TagsValue: vv.TagsValue, TagsString: vv.TagsString})
		}
		c.Data = append(c.Data, &businessStoreMonitorObj{Path: v.Path, Rules: newStoreRules})
	}
	var tmpBuffer bytes.Buffer
	enc := gob.NewEncoder(&tmpBuffer)
	err := enc.Encode(c.Data)
	if err != nil {
		level.Error(newLogger).Log("msg", fmt.Sprintf("gob encode business monitor error : %v ", err))
	} else {
		ioutil.WriteFile(businessMonitorFilePath, tmpBuffer.Bytes(), 0644)
		level.Info(newLogger).Log("msg", fmt.Sprintf("write %s succeed ", businessMonitorFilePath))
	}
}

func (c *businessCollectorStore) Load() {
	file, err := os.Open(businessMonitorFilePath)
	if err != nil {
		level.Info(newLogger).Log("msg", fmt.Sprintf("read %s file error %v ", businessMonitorFilePath, err))
	} else {
		dec := gob.NewDecoder(file)
		err = dec.Decode(&c.Data)
		if err != nil {
			level.Error(newLogger).Log("msg", fmt.Sprintf("gob decode %s error %v ", businessMonitorFilePath, err))
		} else {
			level.Info(newLogger).Log("msg", fmt.Sprintf("load %s file succeed ", businessMonitorFilePath))
		}
	}
	businessMonitorLock.Lock()
	businessMonitorJobs = []*businessMonitorObj{}
	for _, v := range c.Data {
		if v.Path != "" {
			newBusinessMonitorObj := businessMonitorObj{Path: v.Path}
			newBusinessMonitorObj.Lock = new(sync.RWMutex)
			for _, vv := range v.Rules {
				tmpRuleObj := businessRuleObj{Regular: vv.Regular, MetricConfig: vv.MetricConfig, StringMap: vv.StringMap, TagsKey: vv.TagsKey, TagsValue: vv.TagsValue, TagsString: vv.TagsString}
				tmpRuleObj.RegExp = regexp.MustCompile(vv.Regular)
				tmpRuleObj.DataChannel = make(chan map[string]interface{}, 10000)
				newBusinessMonitorObj.Rules = append(newBusinessMonitorObj.Rules, &tmpRuleObj)
			}
			businessMonitorJobs = append(businessMonitorJobs, &newBusinessMonitorObj)
		}
	}
	for _, v := range businessMonitorJobs {
		go v.start()
	}
	businessMonitorLock.Unlock()
}

//func transBusinessRegular(regRuleString string) *regexp.Regexp {
//	regRuleString = strings.ReplaceAll(regRuleString, "[", "\\[")
//	regRuleString = strings.ReplaceAll(regRuleString, "]", "\\]")
//	regRuleString = strings.ReplaceAll(regRuleString, "${json_content}", "(.*)")
//	return regexp.MustCompile(regRuleString)
//}

func StartBusinessAggCron() {
	t := time.NewTicker(10 * time.Second).C
	for {
		<-t
		go calcBusinessAggData()
	}
}

type businessValueObj struct {
	Sum   float64
	Avg   float64
	Count float64
}

func calcBusinessAggData() {
	var newRuleData []*businessRuleMetricObj
	businessMonitorLock.RLock()
	for _, v := range businessMonitorJobs {
		for _, rule := range v.Rules {
			dataLength := len(rule.DataChannel)
			if dataLength == 0 {
				continue
			}
			valueCountMap := make(map[string]*businessValueObj)
			for i := 0; i < dataLength; i++ {
				tmpMapData := <-rule.DataChannel
				// Try to change value to float64
				changedMapData, tmpTagString := changeValueByStringMap(tmpMapData, rule.TagsKey, rule.StringMap)
				for _, metricConfig := range rule.MetricConfig {
					if metricValueFloat, b := changedMapData[metricConfig.Key]; b {
						tmpMapKey := fmt.Sprintf("%s^%s^%s^%s", metricConfig.Key, metricConfig.AggType, metricConfig.Metric, tmpTagString)
						if _, keyExist := valueCountMap[tmpMapKey]; keyExist {
							valueCountMap[tmpMapKey].Sum += metricValueFloat
							valueCountMap[tmpMapKey].Count++
						} else {
							valueCountMap[tmpMapKey] = &businessValueObj{Sum: metricValueFloat, Count: 1, Avg: metricValueFloat}
						}
					}
				}
			}
			for mapKey, mapValue := range valueCountMap {
				mapValue.Avg = mapValue.Sum / mapValue.Count
				keySplitList := strings.Split(mapKey, "^")
				tmpMetricObj := businessRuleMetricObj{Path: v.Path, Agg: keySplitList[1], TagsString: keySplitList[3], Metric: keySplitList[2]}
				if keySplitList[1] == "sum" {
					tmpMetricObj.Value = mapValue.Sum
				} else if keySplitList[1] == "avg" {
					tmpMetricObj.Value = mapValue.Avg
				} else if keySplitList[1] == "count" {
					tmpMetricObj.Value = mapValue.Count
				}
				newRuleData = append(newRuleData, &tmpMetricObj)
			}
		}
		for _,custom := range v.Custom {
			dataLength := len(custom.DataChannel)
			if dataLength == 0 {
				continue
			}
			tmpValueObj := businessValueObj{}
			for i := 0; i < dataLength; i++ {
				customFetchString := <-custom.DataChannel
				// Try to change value to float64
				tmpMapData := make(map[string]interface{})
				tmpMapData[custom.Metric] = customFetchString
				changedMapData, _ := changeValueByStringMap(tmpMapData, []string{}, custom.StringMap)
				tmpValueObj.Sum += changedMapData[custom.Metric]
				tmpValueObj.Count ++
			}
			if tmpValueObj.Count > 0 {
				tmpValueObj.Avg = tmpValueObj.Sum / tmpValueObj.Count
				tmpMetricObj := businessRuleMetricObj{Path: v.Path, Agg: custom.AggType, Metric: custom.Metric, Tags: []string{}}
				if custom.AggType == "sum" {
					tmpMetricObj.Value = tmpValueObj.Sum
				}else if custom.AggType == "avg" {
					tmpMetricObj.Value = tmpValueObj.Avg
				}else if custom.AggType == "count" {
					tmpMetricObj.Value = tmpValueObj.Count
				}
				newRuleData = append(newRuleData, &tmpMetricObj)
			}
		}
	}
	businessMonitorLock.RUnlock()
	businessMonitorMetricLock.Lock()
	businessMonitorMetrics = newRuleData
	businessMonitorMetricLock.Unlock()
}

func printReflectString(input interface{}) string {
	if input == nil {
		return ""
	}
	outputString := ""
	typeString := reflect.TypeOf(input).String()
	if strings.Contains(typeString, "string") {
		outputString = fmt.Sprintf("%s", input)
	} else if strings.Contains(typeString, "int") {
		outputString = fmt.Sprintf("%d", input)
	} else if strings.Contains(typeString, "float") {
		outputString = fmt.Sprintf("%.6f", input)
		for i := 0; i < 6; i++ {
			if outputString[len(outputString)-1:] == "0" {
				outputString = outputString[:len(outputString)-1]
			} else {
				break
			}
		}
		if outputString[len(outputString)-1:] == "." {
			outputString = outputString[:len(outputString)-1]
		}
	}
	return outputString
}

func changeValueByStringMap(input map[string]interface{}, tagKey []string, mapConfig []*businessStringMapObj) (output map[string]float64, tagString string) {
	output = make(map[string]float64)
	inputTagMap := make(map[string]string)
	var newMapConfigList []*businessStringMapRegexpObj
	for _, v := range mapConfig {
		tmpExp, tmpErr := pcre.Compile(v.StringValue, 0)
		tmpRegEnable := false
		if tmpErr == nil {
			tmpRegEnable = true
		}
		newMapConfigList = append(newMapConfigList, &businessStringMapRegexpObj{Key: v.Key, Regulation: v.Regulation, StringValue: v.StringValue, IntValue: v.IntValue, RegEnable: tmpRegEnable, Regexp: tmpExp})
	}
	for k, v := range input {
		if v == nil || reflect.TypeOf(v) == nil {
			continue
		}
		var newValue float64 = 0
		typeString := reflect.TypeOf(v).String()
		if strings.Contains(typeString, "string") {
			valueString := fmt.Sprintf("%s", v)
			fetchReg := false
			for _, stringMapObj := range newMapConfigList {
				if stringMapObj.Key == k {
					if strings.Contains(stringMapObj.Regulation, "regexp") && stringMapObj.RegEnable {
						if len(stringMapObj.Regexp.FindIndex([]byte(valueString), 0)) > 0 {
							if !strings.HasPrefix(stringMapObj.Regulation, "!") {
								newValue = stringMapObj.IntValue
								fetchReg = true
							}
						} else {
							if strings.HasPrefix(stringMapObj.Regulation, "!") {
								newValue = stringMapObj.IntValue
								fetchReg = true
							}
						}
					}
					if fetchReg {
						break
					}
				}
			}
			if !fetchReg {
				metricValueFloat, parseError := strconv.ParseFloat(valueString, 64)
				if parseError == nil {
					newValue = metricValueFloat
				}
				inputTagMap[k] = valueString
			}
		} else if strings.Contains(typeString, "int") {
			newValue, _ = strconv.ParseFloat(fmt.Sprintf("%d", v), 64)
		} else if strings.Contains(typeString, "float") {
			fmt.Sprintf("%.6f", input)
			newValue, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", v), 64)
		}
		output[k] = newValue
	}
	for _, v := range tagKey {
		if tmpTagValue, b := input[v]; b {
			outputString := fmt.Sprintf("%s", tmpTagValue)
			if _, bb := inputTagMap[v]; !bb {
				outputString = fmt.Sprintf("%.6f", output[v])
				for i := 0; i < 6; i++ {
					if outputString[len(outputString)-1:] == "0" {
						outputString = outputString[:len(outputString)-1]
					} else {
						break
					}
				}
				if outputString[len(outputString)-1:] == "." {
					outputString = outputString[:len(outputString)-1]
				}
			}
			tagString += fmt.Sprintf("%s=%s,", v, outputString)
		}
	}
	if tagString != "" {
		tagString = tagString[:len(tagString)-1]
	}
	return output, tagString
}
