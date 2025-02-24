package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strings"
)

func GetBusinessList(endpointId int, ownerEndpoint string) (err error, pathList []*m.BusinessMonitorTable) {
	if endpointId > 0 {
		err = x.SQL("SELECT * FROM business_monitor WHERE endpoint_id=?", endpointId).Find(&pathList)
	} else if ownerEndpoint != "" {
		err = x.SQL("SELECT * FROM business_monitor WHERE owner_endpoint=?", ownerEndpoint).Find(&pathList)
	}
	return err, pathList
}

func GetBusinessListNew(endpointId int, ownerEndpoint string) (err error, result m.BusinessUpdateDto) {
	result.EndpointId = endpointId
	result.PathList = []*m.BusinessUpdatePathObj{}
	var businessMonitorTable []*m.BusinessMonitorTable
	if endpointId > 0 {
		err = x.SQL("SELECT * FROM business_monitor WHERE endpoint_id=?", endpointId).Find(&businessMonitorTable)
	} else if ownerEndpoint != "" {
		err = x.SQL("SELECT * FROM business_monitor WHERE owner_endpoint=?", ownerEndpoint).Find(&businessMonitorTable)
	}
	if err != nil {
		return err, result
	}
	var tmpErr error
	for _, v := range businessMonitorTable {
		var businessMonitorConfigTable []*m.BusinessMonitorCfgTable
		tmpBup := m.BusinessUpdatePathObj{Id: v.Id, Path: v.Path, OwnerEndpoint: v.OwnerEndpoint, Rules: []*m.BusinessMonitorCfgObj{}, CustomMetrics: []*m.BusinessMonitorCustomObj{}}
		x.SQL("select * from business_monitor_cfg where business_monitor_id=?", v.Id).Find(&businessMonitorConfigTable)
		for _, vv := range businessMonitorConfigTable {
			if vv.ConfigType == "json" {
				tmpBmc := m.BusinessMonitorCfgObj{Id: vv.Id, Regular: vv.Regular, Tags: vv.Tags}
				tmpStringMap := []*m.BusinessStringMapObj{}
				tmpErr = json.Unmarshal([]byte(vv.StringMap), &tmpStringMap)
				if tmpErr != nil {
					log.Error(nil, log.LOGGER_APP, "json unmarshal string map obj fail", zap.String("string_map", vv.StringMap), zap.Error(tmpErr))
				} else {
					tmpBmc.StringMap = tmpStringMap
				}
				tmpMetricConfig := []*m.BusinessMetricObj{}
				tmpErr = json.Unmarshal([]byte(vv.MetricConfig), &tmpMetricConfig)
				if tmpErr != nil {
					log.Error(nil, log.LOGGER_APP, "json unmarshal metric config obj fail", zap.String("metric_config", vv.MetricConfig), zap.Error(tmpErr))
				} else {
					tmpBmc.MetricConfig = tmpMetricConfig
				}
				tmpBup.Rules = append(tmpBup.Rules, &tmpBmc)
			} else {
				tmpCustom := m.BusinessMonitorCustomObj{Id: vv.Id, Metric: vv.MetricConfig, AggType: vv.AggType, ValueRegular: vv.Regular}
				tmpStringMap := []*m.BusinessStringMapObj{}
				tmpErr = json.Unmarshal([]byte(vv.StringMap), &tmpStringMap)
				if tmpErr != nil {
					log.Error(nil, log.LOGGER_APP, "json unmarshal string map obj fail", zap.String("string_map", vv.StringMap), zap.Error(tmpErr))
				} else {
					tmpCustom.StringMap = tmpStringMap
				}
				tmpBup.CustomMetrics = append(tmpBup.CustomMetrics, &tmpCustom)
			}
		}
		result.PathList = append(result.PathList, &tmpBup)
	}
	return nil, result
}

func GetBusinessRealEndpoint(endpoint string) string {
	var endpointTable []*m.EndpointTable
	x.SQL("select t2.guid from business_monitor t1 left join endpoint t2 on t1.endpoint_id=t2.id where t1.owner_endpoint=?", endpoint).Find(&endpointTable)
	if len(endpointTable) > 0 {
		return endpointTable[0].Guid
	} else {
		return endpoint
	}
}

func UpdateBusiness(param m.BusinessUpdateDto) error {
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM business_monitor WHERE endpoint_id=?", Param: []interface{}{param.EndpointId}})
	for _, v := range param.PathList {
		var action Action
		params := make([]interface{}, 0)
		action.Sql = "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)"
		params = append(params, param.EndpointId)
		params = append(params, v.Path)
		params = append(params, v.OwnerEndpoint)
		action.Param = params
		actions = append(actions, &action)
	}
	return Transaction(actions)
}

func AddBusinessTable(param m.BusinessUpdateDto) error {
	var actions []*Action
	var err error
	var businessMonitorTable []*m.BusinessMonitorTable
	x.SQL("select * from business_monitor where endpoint_id=?", param.EndpointId).Find(&businessMonitorTable)
	for _, v := range param.PathList {
		for _, vv := range businessMonitorTable {
			if v.Path == vv.Path {
				err = fmt.Errorf("Path %s already exists ", v.Path)
				break
			}
		}
		actions = append(actions, &Action{Sql: "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)", Param: []interface{}{param.EndpointId, v.Path, v.OwnerEndpoint}})
	}
	if err != nil {
		return err
	}
	return Transaction(actions)
}

func UpdateBusinessNew(param m.BusinessUpdateDto) error {
	var actions []*Action
	var businessMonitorTable []*m.BusinessMonitorTable
	x.SQL("select * from business_monitor where endpoint_id=?", param.EndpointId).Find(&businessMonitorTable)
	if len(param.PathList) == 0 {
		if len(businessMonitorTable) == 0 {
			return nil
		}
		actions = append(actions, &Action{Sql: "delete from business_monitor where endpoint_id=?", Param: []interface{}{param.EndpointId}})
		for _, v := range businessMonitorTable {
			actions = append(actions, &Action{Sql: "delete from business_monitor_cfg where business_monitor_id=?", Param: []interface{}{v.Id}})
		}
		return Transaction(actions)
	}
	cfgIdMap := make(map[int]int)
	for _, v := range param.PathList {
		for _, rule := range v.Rules {
			if rule.Id > 0 {
				cfgIdMap[rule.Id] = 1
			}
		}
		for _, customCfg := range v.CustomMetrics {
			if customCfg.Id > 0 {
				cfgIdMap[customCfg.Id] = 1
			}
		}
		if v.Id == 0 {
			actions = append(actions, &Action{Sql: "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)", Param: []interface{}{param.EndpointId, v.Path, v.OwnerEndpoint}})
			continue
		}
		for _, vv := range businessMonitorTable {
			if v.Id == vv.Id {
				actions = append(actions, &Action{Sql: "update business_monitor set path=?,owner_endpoint=? where id=?", Param: []interface{}{v.Path, v.OwnerEndpoint, v.Id}})
				break
			}
		}
	}
	for _, v := range businessMonitorTable {
		delFlag := true
		for _, vv := range param.PathList {
			if vv.Id == v.Id {
				delFlag = false
				break
			}
		}
		if delFlag {
			actions = append(actions, &Action{Sql: "delete from business_monitor where id=?", Param: []interface{}{v.Id}})
		}
	}
	err := Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update business_monitor table fail,%s ", err.Error())
		return err
	}
	actions = []*Action{}
	businessMonitorTable = []*m.BusinessMonitorTable{}
	x.SQL("select * from business_monitor where endpoint_id=?", param.EndpointId).Find(&businessMonitorTable)
	pathIdMap := make(map[string]int)
	for _, v := range businessMonitorTable {
		pathIdMap[v.Path] = v.Id
	}
	var businessMonitorCfgTable []*m.BusinessMonitorCfgTable
	x.SQL("select id from business_monitor_cfg where business_monitor_id in (select id from business_monitor where endpoint_id=?)", param.EndpointId).Find(&businessMonitorCfgTable)
	for _, v := range businessMonitorCfgTable {
		if _, b := cfgIdMap[v.Id]; !b {
			actions = append(actions, &Action{Sql: "delete from business_monitor_cfg where id=?", Param: []interface{}{v.Id}})
		}
	}
	configActions := getUpdateBusinessConfigAction(pathIdMap, param)
	if len(configActions) > 0 {
		actions = append(actions, configActions...)
	}
	customActions := getUpdateBusinessCustomAction(pathIdMap, param)
	if len(customActions) > 0 {
		actions = append(actions, customActions...)
	}
	return Transaction(actions)
}

func getUpdateBusinessConfigAction(pathIdMap map[string]int, param m.BusinessUpdateDto) (actions []*Action) {
	actions = []*Action{}
	for _, v := range param.PathList {
		for _, rule := range v.Rules {
			stringMapBytes, _ := json.Marshal(rule.StringMap)
			metricConfigBytes, _ := json.Marshal(rule.MetricConfig)
			if rule.Id > 0 {
				actions = append(actions, &Action{Sql: "update business_monitor_cfg set regular=?,tags=?,string_map=?,metric_config=? where id=?", Param: []interface{}{rule.Regular, rule.Tags, string(stringMapBytes), string(metricConfigBytes), rule.Id}})
			} else {
				actions = append(actions, &Action{Sql: "insert into business_monitor_cfg(business_monitor_id,regular,tags,string_map,metric_config) value (?,?,?,?,?)", Param: []interface{}{pathIdMap[v.Path], rule.Regular, rule.Tags, string(stringMapBytes), string(metricConfigBytes)}})
			}
		}
	}
	return actions
}

func getUpdateBusinessCustomAction(pathIdMap map[string]int, param m.BusinessUpdateDto) (actions []*Action) {
	actions = []*Action{}
	for _, v := range param.PathList {
		for _, custom := range v.CustomMetrics {
			stringMapBytes, _ := json.Marshal(custom.StringMap)
			if custom.Id > 0 {
				actions = append(actions, &Action{Sql: "update business_monitor_cfg set regular=?,string_map=?,metric_config=?,agg_type=? where id=?", Param: []interface{}{custom.ValueRegular, string(stringMapBytes), custom.Metric, custom.AggType, custom.Id}})
			} else {
				actions = append(actions, &Action{Sql: "insert into business_monitor_cfg(business_monitor_id,regular,string_map,metric_config,agg_type,config_type) value (?,?,?,?,?,?)", Param: []interface{}{pathIdMap[v.Path], custom.ValueRegular, string(stringMapBytes), custom.Metric, custom.AggType, "custom"}})
			}
		}
	}
	return actions
}

func UpdateAppendBusiness(param m.BusinessUpdateDto) error {
	var actions []*Action
	var tmpOwnerList []string
	for _, v := range param.PathList {
		exist := false
		for _, vv := range tmpOwnerList {
			if v.OwnerEndpoint == vv {
				exist = true
				break
			}
		}
		if !exist {
			tmpOwnerList = append(tmpOwnerList, v.OwnerEndpoint)
		}
	}
	for _, v := range tmpOwnerList {
		actions = append(actions, &Action{Sql: "DELETE FROM business_monitor WHERE endpoint_id=? AND owner_endpoint=?", Param: []interface{}{param.EndpointId, v}})
	}
	for _, v := range param.PathList {
		var action Action
		params := make([]interface{}, 0)
		action.Sql = "INSERT INTO business_monitor(endpoint_id,path,owner_endpoint) VALUE (?,?,?)"
		params = append(params, param.EndpointId)
		params = append(params, v.Path)
		params = append(params, v.OwnerEndpoint)
		action.Param = params
		actions = append(actions, &action)
	}
	return Transaction(actions)
}

func CheckEndpointBusiness(endpoint string) bool {
	result := true
	var businessMonitorTables []*m.BusinessMonitorTable
	x.SQL("SELECT t2.* FROM endpoint t1 JOIN business_monitor t2 ON t1.id=t2.`endpoint_id` WHERE t1.`guid`=?", endpoint).Find(&businessMonitorTables)
	for _, v := range businessMonitorTables {
		if v.OwnerEndpoint != "" {
			result = false
			break
		}
	}
	return result
}

func GetBusinessPanelChart() (charts []*m.ChartTable, panels []*m.PanelTable) {
	x.SQL("SELECT t2.* FROM dashboard t1 LEFT JOIN panel t2 ON t1.panels_group=t2.group_id WHERE t1.dashboard_type='host' AND t2.auto_display=1").Find(&panels)
	x.SQL("SELECT t3.* FROM dashboard t1 LEFT JOIN panel t2 ON t1.panels_group=t2.group_id LEFT JOIN chart t3 ON t2.chart_group=t3.group_id WHERE t1.dashboard_type='host' AND t2.auto_display=1").Find(&charts)
	return charts, panels
}

func GetBusinessPromMetric(keys []string) (err error, result []*m.PromMetricTable) {
	if len(keys) == 0 {
		return err, result
	}
	sql := "SELECT * FROM monitor.prom_metric where "
	for i, v := range keys {
		if v == "" {
			continue
		}
		sql += " (prom_ql like '%key=\"" + v + "\"%') "
		if i < len(keys)-1 {
			sql += " OR "
		}
	}
	err = x.SQL(sql).Find(&result)
	return err, result
}

func CheckHostLogFileExist(hostIp, logPath string) error {
	return nil
}

func PluginBusinessAction(input *m.PluginBusinessValueRequestObj) (result *m.PluginBusinessOutputObj, endpointId int, err error) {
	result = &m.PluginBusinessOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: ""}
	if input.HostIp == "" {
		err = fmt.Errorf("Param validate fail,hostIp can not empty ")
		return
	}
	endpointObj := m.EndpointTable{Ip: input.HostIp, ExportType: "host"}
	GetEndpoint(&endpointObj)
	endpointId = endpointObj.Id
	refEndpointObj := m.EndpointTable{Guid: input.RefMonitorObj}
	GetEndpoint(&refEndpointObj)
	if input.RefMonitorObj != "" && refEndpointObj.Id <= 0 {
		err = fmt.Errorf("Param refMonitorObj can not fetch any data ")
		return
	}
	var actions []*Action
	for _, pathConfig := range input.Config {
		logPath := input.PathPrefix + pathConfig.Path
		tmpErr := CheckHostLogFileExist(input.HostIp, logPath)
		if tmpErr != nil {
			err = fmt.Errorf("LogFile:%s check illegal in host:%s,%s ", logPath, input.HostIp, tmpErr.Error())
			break
		}
		actions = append(actions, getPluginBusinessUpdateActions(endpointObj.Id, logPath, refEndpointObj, pathConfig.Rules)...)
	}
	if err != nil {
		return
	}
	err = Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update database fail,%s ", err.Error())
	}
	return
}

func getPluginBusinessUpdateActions(endpointId int, path string, refEndpointObj m.EndpointTable, rules []*m.PluginBusinessRuleObj) []*Action {
	var actions []*Action
	var businessMonitorTable []*m.BusinessMonitorTable
	var businessMonitorConfigs []*m.BusinessMonitorCfgTable
	var metricConfigList []*m.BusinessMetricObj
	for _, rule := range rules {
		tmpConfigRow := m.BusinessMonitorCfgTable{Regular: rule.Regular, Tags: rule.Tags}
		tmpMetricConfigs := []*m.BusinessMetricObj{}
		tmpStringMap := []*m.BusinessStringMapObj{}
		for _, metric := range rule.MetricConfig {
			tmpMetricConfigs = append(tmpMetricConfigs, &m.BusinessMetricObj{Key: metric.Key, Metric: metric.Metric, Title: metric.Title, AggType: metric.AggType})
			for _, valueMap := range metric.ValueMap {
				tmpReguslar := "regexp"
				if strings.ToLower(valueMap.IsRegular) == "no" {
					tmpReguslar = "!regexp"
				}
				tmpStringMap = append(tmpStringMap, &m.BusinessStringMapObj{Key: metric.Key, StringValue: valueMap.StringValue, IntValue: valueMap.IntValue, Regulation: tmpReguslar})
			}
		}
		metricConfigList = append(metricConfigList, tmpMetricConfigs...)
		metricConfigBytes, _ := json.Marshal(tmpMetricConfigs)
		tmpConfigRow.MetricConfig = string(metricConfigBytes)
		stringMapBytes, _ := json.Marshal(tmpStringMap)
		tmpConfigRow.StringMap = string(stringMapBytes)
		businessMonitorConfigs = append(businessMonitorConfigs, &tmpConfigRow)
	}
	x.SQL("select * from business_monitor where endpoint_id=? and `path`=?", endpointId, path).Find(&businessMonitorTable)
	if len(businessMonitorTable) == 0 {
		actions = append(actions, &Action{Sql: "insert into business_monitor(endpoint_id,`path`,owner_endpoint) value (?,?,?)", Param: []interface{}{endpointId, path, refEndpointObj.Guid}})
		for _, v := range businessMonitorConfigs {
			actions = append(actions, &Action{Sql: "insert into business_monitor_cfg(business_monitor_id,regular,tags,string_map,metric_config) select id as business_monitor_id,? as regular,? as tags,? as string_map,? as metric_config from business_monitor where endpoint_id=? and `path`=?", Param: []interface{}{v.Regular, v.Tags, v.StringMap, v.MetricConfig, endpointId, path}})
		}
	} else {
		if businessMonitorTable[0].OwnerEndpoint != refEndpointObj.Guid {
			actions = append(actions, &Action{Sql: "update business_monitor set owner_endpoint=? where id=?", Param: []interface{}{refEndpointObj.Guid, businessMonitorTable[0].Id}})
			actions = append(actions, &Action{Sql: "delete from business_monitor_cfg where business_monitor_id=?", Param: []interface{}{businessMonitorTable[0].Id}})
			for _, v := range businessMonitorConfigs {
				actions = append(actions, &Action{Sql: "insert into business_monitor_cfg(business_monitor_id,regular,tags,string_map,metric_config) value (?,?,?,?,?)", Param: []interface{}{businessMonitorTable[0].Id, v.Regular, v.Tags, v.StringMap, v.MetricConfig}})
			}
		}
	}

	return actions
}
