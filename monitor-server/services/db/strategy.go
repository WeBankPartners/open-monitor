package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
	"strings"
)

var ruleConfigIgnoreType = []string{}

// 更新告警规则文件rule file
func SyncRuleConfigFile(tplId int, moveOutEndpoints []string, fromPeer bool) error {
	// 获取tpl对象
	tplObj, err := GetTemplateObject(tplId, 0, 0)
	if err != nil {
		return err
	}
	// 获取文件名
	var ruleFileName string
	var endpointList, moveOutEndpointList []*models.EndpointTable
	if tplObj.GrpId > 0 {
		getGrpErr, groupObj := GetSingleGrp(tplObj.GrpId, "")
		if getGrpErr != nil {
			return getGrpErr
		}
		ruleFileName = "g_" + groupObj.Name
		err, endpointList = GetEndpointsByGrp(tplObj.GrpId)
		if err != nil {
			return err
		}
	} else {
		endpointObj := models.EndpointTable{Id: tplObj.EndpointId}
		err = GetEndpoint(&endpointObj)
		if err != nil {
			return err
		}
		ruleFileName = "e_" + endpointObj.Guid
		endpointList = append(endpointList, &endpointObj)
	}
	// 获取strategy
	strategyList, getStrategyErr := GetStrategyList(0, tplId)
	if getStrategyErr != nil {
		return getStrategyErr
	}
	// 区分cluster，分别下发
	var clusterList []string
	var clusterEndpointMap = make(map[string][]*models.EndpointTable)
	if len(moveOutEndpoints) > 0 {
		err = x.SQL("select guid,cluster from endpoint where guid in ('" + strings.Join(moveOutEndpoints, "','") + "')").Find(&moveOutEndpointList)
		if err != nil {
			return fmt.Errorf("Try to query move out endpoints fail,%s ", err.Error())
		}
		for _, endpoint := range moveOutEndpointList {
			clusterList = append(clusterList, endpoint.Cluster)
			if _, b := clusterEndpointMap[endpoint.Cluster]; !b {
				clusterEndpointMap[endpoint.Cluster] = []*models.EndpointTable{}
			}
		}
	}
	if len(endpointList) > 0 {
		for _, endpoint := range endpointList {
			if len(moveOutEndpoints) > 0 {
				ignoreFlag := false
				for _, moveOutGuid := range moveOutEndpoints {
					if endpoint.Guid == moveOutGuid {
						ignoreFlag = true
						break
					}
				}
				if ignoreFlag {
					continue
				}
			}
			if _, b := clusterEndpointMap[endpoint.Cluster]; !b {
				clusterList = append(clusterList, endpoint.Cluster)
				clusterEndpointMap[endpoint.Cluster] = []*models.EndpointTable{endpoint}
			} else {
				clusterEndpointMap[endpoint.Cluster] = append(clusterEndpointMap[endpoint.Cluster], endpoint)
			}
		}
	}
	for _, cluster := range clusterList {
		guidExpr, addressExpr, ipExpr := buildRuleReplaceExpr(clusterEndpointMap[cluster])
		ruleFileConfig := buildRuleFileContent(ruleFileName, guidExpr, addressExpr, ipExpr, copyStrategyList(strategyList))
		if cluster == "default" || cluster == "" {
			prom.SyncLocalRuleConfig(models.RuleLocalConfigJob{FromPeer: fromPeer, TplId: tplId, Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
		} else {
			tmpErr := SyncRemoteRuleConfigFile(cluster, models.RFClusterRequestObj{Name: ruleFileConfig.Name, Rules: ruleFileConfig.Rules})
			if tmpErr != nil {
				err = fmt.Errorf("Update remote cluster:%s rule file fail,%s ", cluster, tmpErr.Error())
				log.Error(nil, log.LOGGER_APP, "Update remote cluster rule file fail", zap.String("cluster", cluster), zap.Error(tmpErr))
			}
		}
	}
	return err
}

func buildRuleReplaceExpr(endpointList []*models.EndpointTable) (guidExpr, addressExpr, ipExpr string) {
	for _, endpoint := range endpointList {
		ignoreType := false
		for _, tmpType := range ruleConfigIgnoreType {
			if endpoint.ExportType == tmpType {
				ignoreType = true
				break
			}
		}
		if ignoreType {
			continue
		}
		if endpoint.AddressAgent != "" {
			addressExpr += endpoint.AddressAgent + "|"
		} else {
			addressExpr += endpoint.Address + "|"
		}
		guidExpr += endpoint.Guid + "|"
		ipExpr += endpoint.Ip + "|"
	}
	if addressExpr != "" {
		addressExpr = addressExpr[:len(addressExpr)-1]
	}
	if guidExpr != "" {
		guidExpr = guidExpr[:len(guidExpr)-1]
	}
	if ipExpr != "" {
		ipExpr = ipExpr[:len(ipExpr)-1]
	}
	return
}

func buildRuleFileContent(ruleFileName, guidExpr, addressExpr, ipExpr string, strategyList []*models.StrategyTable) models.RFGroup {
	result := models.RFGroup{Name: ruleFileName}
	if len(strategyList) == 0 {
		return result
	}
	for _, strategy := range strategyList {
		tmpRfu := models.RFRule{}
		tmpRfu.Alert = fmt.Sprintf("%s_%d", strategy.Metric, strategy.Id)
		if !strings.Contains(strategy.Cond, " ") && strategy.Cond != "" {
			if strings.Contains(strategy.Cond, "=") {
				strategy.Cond = strategy.Cond[:2] + " " + strategy.Cond[2:]
			} else {
				strategy.Cond = strategy.Cond[:1] + " " + strategy.Cond[1:]
			}
		}
		if strings.Contains(strategy.Expr, "$address") {
			if strings.Contains(addressExpr, "|") {
				strategy.Expr = strings.Replace(strategy.Expr, "=\"$address\"", "=~\""+addressExpr+"\"", -1)
			} else {
				strategy.Expr = strings.Replace(strategy.Expr, "=\"$address\"", "=\""+addressExpr+"\"", -1)
			}
		}
		if strings.Contains(strategy.Expr, "$guid") {
			if strings.Contains(guidExpr, "|") {
				strategy.Expr = strings.Replace(strategy.Expr, "=\"$guid\"", "=~\""+guidExpr+"\"", -1)
			} else {
				strategy.Expr = strings.Replace(strategy.Expr, "=\"$guid\"", "=\""+guidExpr+"\"", -1)
			}
		}
		if strings.Contains(strategy.Expr, "$ip") {
			if strings.Contains(ipExpr, "|") {
				tmpStr := strings.Split(strategy.Expr, "$ip")[1]
				tmpStr = tmpStr[:strings.Index(tmpStr, "\"")]
				newList := []string{}
				for _, v := range strings.Split(ipExpr, "|") {
					newList = append(newList, v+tmpStr)
				}
				strategy.Expr = strings.Replace(strategy.Expr, "=\"$ip"+tmpStr+"\"", "=~\""+strings.Join(newList, "|")+"\"", -1)
			} else {
				strategy.Expr = strings.ReplaceAll(strategy.Expr, "$ip", ipExpr)
			}
		}
		tmpRfu.Expr = fmt.Sprintf("%s %s", strategy.Expr, strategy.Cond)
		tmpRfu.For = strategy.Last
		tmpRfu.Labels = make(map[string]string)
		tmpRfu.Labels["strategy_id"] = fmt.Sprintf("%d", strategy.Id)
		tmpRfu.Annotations = models.RFAnnotation{Summary: fmt.Sprintf("{{$labels.instance}}__%s__%s__{{$value}}", strategy.Priority, strategy.Metric), Description: strategy.Content}
		result.Rules = append(result.Rules, &tmpRfu)
	}
	return result
}

func copyStrategyList(inputs []*models.StrategyTable) []*models.StrategyTable {
	result := []*models.StrategyTable{}
	for _, strategy := range inputs {
		tmpStrategy := models.StrategyTable{Id: strategy.Id, TplId: strategy.TplId, Metric: strategy.Metric, Expr: strategy.Expr, Cond: strategy.Cond, Last: strategy.Last, Priority: strategy.Priority, Content: strategy.Content, ConfigType: strategy.ConfigType}
		result = append(result, &tmpStrategy)
	}
	return result
}

func GetTemplateObject(id, grpId, endpointId int) (result *models.TplTable, err error) {
	var queryResult []*models.TplTable
	var queryParams = make([]interface{}, 0)
	baseSql := "select * from tpl where 1=1"
	if id > 0 {
		baseSql += " and id=? "
		queryParams = append(queryParams, id)
	}
	if endpointId > 0 {
		baseSql += " and endpoint_id=? "
		queryParams = append(queryParams, endpointId)
	}
	if grpId > 0 {
		baseSql += " and grp_id=? "
		queryParams = append(queryParams, grpId)
	}
	err = x.SQL(baseSql, queryParams...).Find(&queryResult)
	if err != nil {
		err = fmt.Errorf("Try to query tpl table fail,%s ", err.Error())
	}
	if len(queryResult) == 0 {
		result = &models.TplTable{Id: 0}
		err = fmt.Errorf("Can not find any tpl row by id=%d or endpoint_id=%d or grp_id=%d ", id, endpointId, grpId)
	} else {
		result = queryResult[0]
	}
	return
}

func GetStrategyList(id, tplId int) (result []*models.StrategyTable, err error) {
	result = []*models.StrategyTable{}
	baseSql := "select * from strategy where 1=1"
	var queryParams = make([]interface{}, 0)
	if id > 0 {
		baseSql += " and id=? "
		queryParams = append(queryParams, id)
	}
	if tplId > 0 {
		baseSql += " and tpl_id=? "
		queryParams = append(queryParams, tplId)
	}
	err = x.SQL(baseSql, queryParams...).Find(&result)
	if err != nil {
		err = fmt.Errorf("Try to query strategy table fail,%s ", err.Error())
	}
	return
}

func UpdateStrategy(obj *models.UpdateStrategy) error {
	var actions []*Action
	for _, v := range obj.Strategy {
		action := Classify(*v, obj.Operation, "strategy", true)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	err := Transaction(actions)
	return err
}
