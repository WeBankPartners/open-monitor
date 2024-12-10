package db

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func ListAlarmEndpoints(query *m.AlarmEndpointQuery) error {
	var queryParams, countParams []interface{}
	whereSql := ""
	if query.Search != "" {
		whereSql += ` AND t1.guid LIKE ? `
		countParams = append(countParams, "%"+query.Search+"%")
	}
	if query.Grp > 0 {
		whereSql += " AND t2.grp_id=? "
		countParams = append(countParams, query.Grp)
	}
	if len(query.EndpointGroup) > 0 {
		whereSql += " AND t2.endpoint_group in ('" + strings.Join(query.EndpointGroup, "','") + "') "
	}
	if len(query.BasicType) > 0 {
		whereSql += " AND t1.monitor_type in ('" + strings.Join(query.BasicType, "','") + "') "
	}
	querySql := `SELECT t5.* FROM (
			SELECT t4.guid,GROUP_CONCAT(t4.endpoint_group) groups_ids,t4.type,t4.tags,t4.create_user,t4.update_user,t4.update_time FROM (
			SELECT t1.guid,t2.endpoint_group,t1.monitor_type as type,t1.tags,t1.create_user,t1.update_user,t1.update_time FROM endpoint_new t1 
			LEFT JOIN endpoint_group_rel t2 ON t1.guid=t2.endpoint 
			WHERE 1=1 ` + whereSql + `
			) t4 GROUP BY t4.guid
			) t5 ORDER BY t5.update_time DESC LIMIT ?,?`
	countSql := `SELECT COUNT(1) num FROM (
			SELECT t4.guid,GROUP_CONCAT(t4.endpoint_group) groups_ids,t4.type,t4.tags FROM (
			SELECT t1.guid,t2.endpoint_group,t1.monitor_type as type,t1.tags FROM endpoint_new t1 
			LEFT JOIN endpoint_group_rel t2 ON t1.guid=t2.endpoint
			WHERE 1=1 ` + whereSql + `
			) t4 GROUP BY t4.guid
			) t5`
	var result []*m.AlarmEndpointObj
	var count []int
	queryParams = append(queryParams, countParams...)
	queryParams = append(queryParams, (query.Page-1)*query.Size, query.Size)
	err := x.SQL(querySql, queryParams...).Find(&result)
	err = x.SQL(countSql, countParams...).Find(&count)
	if len(result) > 0 {
		var groupTableData []*m.EndpointGroupTable
		x.SQL("select * from endpoint_group").Find(&groupTableData)
		for _, v := range result {
			if v.GroupsIds != "" {
				tmpExistMap := make(map[string]int)
				for _, tmpGroupId := range strings.Split(v.GroupsIds, ",") {
					if tmpGroupId == "" {
						continue
					}
					if _, b := tmpExistMap[tmpGroupId]; b {
						continue
					}
					tmpExistMap[tmpGroupId] = 1
					for _, groupObj := range groupTableData {
						if groupObj.Guid == tmpGroupId {
							v.Groups = append(v.Groups, &m.GrpTable{Name: groupObj.Guid})
							break
						}
					}
				}
			}
		}
		query.Result = result
		query.ResultNum = count[0]
	} else {
		query.Result = []*m.AlarmEndpointObj{}
		query.ResultNum = 0
	}
	return err
}

func ListGrpEndpointOptions() (options *m.EndpointOptions, err error) {
	var monitorTypeList []*m.MonitorTypeTable
	var endpointGroupList []*m.EndpointGroupTable
	if err = x.SQL("select display_name from monitor_type where display_name in (select monitor_type from endpoint_new) order by create_time desc ").Find(&monitorTypeList); err != nil {
		return
	}
	if err = x.SQL("select guid from endpoint_group where guid in (select endpoint_group from endpoint_group_rel) order by update_time desc").Find(&endpointGroupList); err != nil {
		return
	}
	options = &m.EndpointOptions{EndpointGroup: []string{}, BasicType: []string{}}
	if len(monitorTypeList) > 0 {
		for _, monitorType := range monitorTypeList {
			options.BasicType = append(options.BasicType, monitorType.DisplayName)
		}
	}
	if len(endpointGroupList) > 0 {
		for _, endpointGroup := range endpointGroupList {
			options.EndpointGroup = append(options.EndpointGroup, endpointGroup.Guid)
		}
	}
	return
}

func GetTplStrategy(query *m.TplQuery, ignoreLogMonitor bool) error {
	var result []*m.TplObj
	if query.SearchType == "endpoint" {
		var grps []*m.GrpTable
		err := x.SQL("SELECT * FROM grp where id in (select grp_id from grp_endpoint WHERE endpoint_id=?)", query.SearchId).Find(&grps)
		if err != nil {
			log.Logger.Error("Get strategy fail", log.Error(err))
			return err
		}
		var grpIds string
		grpMap := make(map[int]string)
		if len(grps) > 0 {
			grpIds = "t1.grp_id IN ("
			for _, v := range grps {
				grpIds += fmt.Sprintf("%d,", v.Id)
				grpMap[v.Id] = v.Name
				if v.Parent > 0 {
					tmpGrpId := v.Id
					tmpParentId := v.Parent
					tmpGrpName := v.Name
					// 查找父模板,最多递归10级
					for i := 0; i < 10; i++ {
						parentGrp := getGrpParent(tmpParentId)
						if parentGrp.Id > 0 {
							grpIds += fmt.Sprintf("%d,", parentGrp.Id)
							grpMap[tmpGrpId] = fmt.Sprintf("%s [%s]", tmpGrpName, parentGrp.Name)
							if parentGrp.Parent <= 0 {
								grpMap[parentGrp.Id] = parentGrp.Name
								break
							} else {
								tmpGrpId = parentGrp.Id
								tmpParentId = parentGrp.Parent
								tmpGrpName = parentGrp.Name
							}
						} else {
							grpMap[tmpGrpId] = tmpGrpName
							break
						}
					}
				}
			}
			grpIds = grpIds[:len(grpIds)-1]
			grpIds += ") OR"
		}
		var tpls []*m.TplStrategyTable
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content,t2.notify_enable,t2.notify_delay 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE (` + grpIds + ` endpoint_id=?)  order by t1.endpoint_id,t1.id,t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			log.Logger.Error("Get strategy fail", log.Error(err))
			return err
		}
		if len(tpls) == 0 {
			endpointObj := m.EndpointTable{Id: query.SearchId}
			GetEndpoint(&endpointObj)
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}})
		} else {
			var tmpTplId int
			tmpStrategys := []*m.StrategyTable{}
			for i, v := range tpls {
				if ignoreLogMonitor && v.Metric == "log_monitor" {
					continue
				}
				if i == 0 {
					tmpTplId = v.TplId
					if v.StrategyId > 0 {
						tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelay: v.NotifyDelay})
					}
				} else {
					if v.TplId != tmpTplId {
						tmpTplObj := m.TplObj{TplId: tpls[i-1].TplId}
						if tpls[i-1].GrpId > 0 {
							tmpTplObj.ObjId = tpls[i-1].GrpId
							tmpTplObj.ObjName = grpMap[tpls[i-1].GrpId]
							tmpTplObj.ObjType = "grp"
							tmpTplObj.Operation = false
						} else {
							tmpTplObj.ObjId = tpls[i-1].EndpointId
							endpointObj := m.EndpointTable{Id: tpls[i-1].EndpointId}
							GetEndpoint(&endpointObj)
							tmpTplObj.ObjName = endpointObj.Guid
							tmpTplObj.ObjType = "endpoint"
							tmpTplObj.Operation = true
						}
						tmpTplObj.Strategy = tmpStrategys
						result = append(result, &tmpTplObj)
						tmpTplId = v.TplId
						tmpStrategys = []*m.StrategyTable{}
					}
					if v.StrategyId > 0 {
						tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelay: v.NotifyDelay})
					}
				}
			}
			if tpls[len(tpls)-1].EndpointId > 0 {
				endpointObj := m.EndpointTable{Id: tpls[len(tpls)-1].EndpointId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: tpls[len(tpls)-1].TplId, ObjId: tpls[len(tpls)-1].EndpointId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: tmpStrategys})
			} else {
				result = append(result, &m.TplObj{TplId: tpls[len(tpls)-1].TplId, ObjId: tpls[len(tpls)-1].GrpId, ObjName: grpMap[tpls[len(tpls)-1].GrpId], ObjType: "grp", Operation: false, Strategy: tmpStrategys})
				endpointObj := m.EndpointTable{Id: query.SearchId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}})
			}
		}
	} else {
		var grps []*m.GrpTable
		err := x.SQL("SELECT * FROM grp WHERE id=?", query.SearchId).Find(&grps)
		if err != nil {
			log.Logger.Error("Get group fail", log.Error(err))
			return err
		}
		if len(grps) <= 0 {
			log.Logger.Warn("Can't find this grp")
			return fmt.Errorf("can't find this grp")
		}
		var parentTpls []*m.TplStrategyTable
		var tpls []*m.TplStrategyTable
		if grps[0].Parent > 0 {
			tmpParentId := grps[0].Parent
			for i := 0; i < 10; i++ {
				parentGrp := getGrpParent(tmpParentId)
				sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content,t2.notify_enable,t2.notify_delay 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE t1.grp_id=?  ORDER BY t2.id`
				parentTpls = []*m.TplStrategyTable{}
				x.SQL(sql, parentGrp.Id).Find(&parentTpls)
				if len(parentTpls) > 0 {
					tmpStrategys := []*m.StrategyTable{}
					for _, v := range parentTpls {
						if v.StrategyId > 0 {
							if ignoreLogMonitor && v.Metric == "log_monitor" {
								continue
							}
							tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelay: v.NotifyDelay})
						}
					}
					result = append(result, &m.TplObj{TplId: parentTpls[0].TplId, ObjId: parentGrp.Id, ObjName: parentGrp.Name, ObjType: "grp", Operation: false, Strategy: tmpStrategys})
				} else {
					result = append(result, &m.TplObj{TplId: 0, ObjId: parentGrp.Id, ObjName: parentGrp.Name, ObjType: "grp", Operation: false, Strategy: []*m.StrategyTable{}})
				}
				if parentGrp.Parent <= 0 {
					break
				} else {
					tmpParentId = parentGrp.Parent
				}
			}
			var newResult []*m.TplObj
			var tmpParentName, tmpObjName string
			for i := len(result); i > 0; i-- {
				tmpObjName = result[i-1].ObjName
				if tmpParentName != "" {
					result[i-1].ObjName = fmt.Sprintf("%s [%s]", tmpObjName, tmpParentName)
				}
				tmpParentName = tmpObjName
				newResult = append(newResult, result[i-1])
			}
			result = newResult
		}
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content,t2.notify_enable,t2.notify_delay 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE t1.grp_id=?  ORDER BY t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			log.Logger.Error("Get strategy fail", log.Error(err))
			return err
		}
		if len(tpls) > 0 {
			tmpStrategys := []*m.StrategyTable{}
			for _, v := range tpls {
				if v.StrategyId > 0 {
					if ignoreLogMonitor && v.Metric == "log_monitor" {
						continue
					}
					tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content, NotifyEnable: v.NotifyEnable, NotifyDelay: v.NotifyDelay})
				}
			}
			result = append(result, &m.TplObj{TplId: tpls[0].TplId, ObjId: tpls[0].GrpId, ObjName: grps[0].Name, ObjType: "grp", Operation: true, Strategy: tmpStrategys})
		} else {
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: grps[0].Name, ObjType: "grp", Operation: true, Strategy: []*m.StrategyTable{}})
		}
	}
	for i, v := range result {
		result[i].Accept = getActionOptions(v.TplId)
	}
	query.Tpl = result
	return nil
}

func GetTpl(tplId, grpId, endpointId int) (error, m.TplTable) {
	param := make([]interface{}, 0)
	sql := `SELECT id,grp_id,endpoint_id,notify_url FROM tpl WHERE 1=1 `
	if tplId > 0 {
		sql += " and id=?"
		param = append(param, tplId)
	}
	if grpId > 0 || endpointId > 0 {
		sql += " and grp_id=? and endpoint_id=?"
		param = append(param, grpId)
		param = append(param, endpointId)
	}
	var result []*m.TplTable
	err := x.SQL(sql, param...).Find(&result)
	if err != nil {
		return fmt.Errorf("Query tpl table fail,%s ", err.Error()), m.TplTable{}
	}
	if len(result) == 0 {
		return fmt.Errorf("Can not find any tpl obj with id=%d or grp_id=%d and endpoint_id=%d ", tplId, grpId, endpointId), m.TplTable{}
	}
	return nil, *result[0]
}

func ListTpl() []*m.TplTable {
	var result []*m.TplTable
	x.SQL("SELECT * FROM tpl").Find(&result)
	return result
}

func GetParentTpl(tplId int) []int {
	type tplGrpParent struct {
		Id     int
		GrpId  int
		Parent int
	}
	var result []*tplGrpParent
	x.SQL("SELECT t1.id,t1.grp_id,t2.parent FROM tpl t1 LEFT JOIN grp t2 ON t1.grp_id=t2.id").Find(&result)
	if len(result) == 0 {
		return []int{}
	}
	tmpGrptId := 0
	for _, v := range result {
		if v.Id == tplId {
			tmpGrptId = v.GrpId
			break
		}
	}
	var tplIdList []int
	tmpGrpMap := make(map[int]int)
	for i := 0; i < 10; i++ {
		endFlag := true
		for _, v := range result {
			for kk, vv := range tmpGrpMap {
				if vv == 2 {
					continue
				}
				if v.Parent == kk {
					endFlag = false
					tmpGrpMap[v.GrpId] = 1
					tplIdList = append(tplIdList, v.Id)
					tmpGrpMap[kk] = 2
				}
			}
			if v.Parent == tmpGrptId {
				if _, b := tmpGrpMap[v.GrpId]; !b {
					endFlag = false
					tmpGrpMap[v.GrpId] = 1
					tplIdList = append(tplIdList, v.Id)
				}
			}
		}
		if endFlag {
			break
		}
	}
	return tplIdList
}

func AddTpl(grpId, endpointId int, operateUser string) (error, m.TplTable) {
	_, tpl := GetTpl(0, grpId, endpointId)
	if tpl.Id > 0 {
		return nil, tpl
	}
	insertSql := fmt.Sprintf("INSERT INTO tpl(grp_id,endpoint_id,create_user,update_user,create_at,update_at) VALUE (%d,%d,'%s','%s',NOW(),NOW())", grpId, endpointId, operateUser, operateUser)
	_, err := x.Exec(insertSql)
	if err != nil {
		log.Logger.Error("Add tpl fail", log.Error(err))
		return err, tpl
	}
	_, tpl = GetTpl(0, grpId, endpointId)
	if tpl.Id <= 0 {
		err = fmt.Errorf("cat't find the new one")
		log.Logger.Error("Add tpl fail", log.Error(err))
		return err, tpl
	}
	return nil, tpl
}

func UpdateTpl(tplId int, operateUser string) error {
	_, err := x.Exec("UPDATE tpl SET update_user=?,update_at=NOW() WHERE id=?", operateUser, tplId)
	if err != nil {
		log.Logger.Error("Update tpl fail", log.Error(err))
	}
	return err
}

func DeleteTpl(tplId int) error {
	_, err := x.Exec("DELETE from tpl where id=?", tplId)
	if err != nil {
		log.Logger.Error("Delete tpl fail", log.Error(err))
	}
	return err
}

func GetStrategyTable(id int) (error, m.StrategyTable) {
	var result []*m.StrategyTable
	err := x.SQL("SELECT * FROM strategy WHERE id=?", id).Find(&result)
	if err != nil || len(result) <= 0 {
		log.Logger.Error("Get strategy table fail", log.Error(err))
		return err, m.StrategyTable{}
	}
	return nil, *result[0]
}

func GetEndpointsByGrp(grpId int) (error, []*m.EndpointTable) {
	var result []*m.EndpointTable
	err := x.SQL("SELECT * FROM endpoint WHERE id IN (SELECT endpoint_id FROM grp_endpoint WHERE grp_id=?)", grpId).Find(&result)
	if err != nil {
		err = fmt.Errorf("Get endpoint by grp fail,%s ", err.Error())
	}
	return err, result
}

func GetAlarms(cond m.QueryAlarmCondition) (error, m.AlarmProblemList) {
	var result []*m.AlarmProblemQuery
	var whereSql string
	var params []interface{}
	if cond.AlarmTable.Id > 0 {
		whereSql += " and id=? "
		params = append(params, cond.AlarmTable.Id)
	}
	if cond.AlarmTable.StrategyId > 0 {
		whereSql += " and strategy_id=? "
		params = append(params, cond.AlarmTable.StrategyId)
	}
	if cond.AlarmTable.Endpoint != "" {
		cond.EndpointFilterList = append(cond.EndpointFilterList, cond.AlarmTable.Endpoint)
	}
	if len(cond.EndpointFilterList) > 0 {
		endpointFilterSql, endpointFilterParam := createListParams(cond.EndpointFilterList, "")
		whereSql += " and endpoint in (" + endpointFilterSql + ") "
		params = append(params, endpointFilterParam...)
	}
	if cond.AlarmTable.SMetric != "" {
		cond.MetricFilterList = append(cond.MetricFilterList, cond.AlarmTable.SMetric)
	}
	if len(cond.MetricFilterList) > 0 {
		metricFilterSql, metricFilterParam := createListParams(cond.MetricFilterList, "")
		whereSql += " and s_metric in (" + metricFilterSql + ") "
		params = append(params, metricFilterParam...)
	}
	if cond.AlarmTable.AlarmName != "" {
		cond.AlarmNameFilterList = append(cond.AlarmNameFilterList, cond.AlarmTable.AlarmName)
	}
	if len(cond.AlarmNameFilterList) > 0 {
		alarmNameFilterSql, alarmNameFilterParam := createListParams(cond.AlarmNameFilterList, "")
		whereSql += " and ( alarm_name in (" + alarmNameFilterSql + ")  or  content in (" + alarmNameFilterSql + "))"
		params = append(append(params, alarmNameFilterParam...), alarmNameFilterParam...)
	}
	if cond.AlarmTable.SCond != "" {
		whereSql += " and s_cond=? "
		params = append(params, cond.AlarmTable.SCond)
	}
	if cond.AlarmTable.SLast != "" {
		whereSql += " and s_last=? "
		params = append(params, cond.AlarmTable.SLast)
	}
	if cond.AlarmTable.SPriority != "" {
		cond.PriorityList = append(cond.PriorityList, cond.AlarmTable.SPriority)
	}
	if len(cond.PriorityList) > 0 {
		priorityFilterSql, priorityFilterParam := createListParams(cond.PriorityList, "")
		whereSql += " and s_priority in (" + priorityFilterSql + ") "
		params = append(params, priorityFilterParam...)
	}
	if cond.AlarmTable.Tags != "" {
		whereSql += " and tags=? "
		params = append(params, cond.AlarmTable.Tags)
	}
	if cond.AlarmTable.Status != "" {
		whereSql += " and status=? "
		params = append(params, cond.AlarmTable.Status)
	}
	if !cond.AlarmTable.Start.IsZero() {
		whereSql += fmt.Sprintf(" and start>='%s' ", cond.AlarmTable.Start.Format(m.DatetimeFormat))
	}
	if !cond.AlarmTable.End.IsZero() {
		whereSql += fmt.Sprintf(" and end<='%s' ", cond.AlarmTable.End.Format(m.DatetimeFormat))
	}
	//  支持 告警任意搜索,但是只能搜索不关闭的告警
	if strings.TrimSpace(cond.Query) != "" {
		whereSql += " and status <> 'closed' and (endpoint like ? or s_metric like ? or alarm_name like ? or content like ? or s_priority like ?)"
		params = append(params, []interface{}{fmt.Sprintf("%%%s%%", cond.Query), fmt.Sprintf("%%%s%%", cond.Query),
			fmt.Sprintf("%%%s%%", cond.Query), fmt.Sprintf("%%%s%%", cond.Query), fmt.Sprintf("%%%s%%", cond.Query)}...)
	}

	sql := "SELECT * FROM alarm where 1=1 " + whereSql + " ORDER BY id DESC "
	if cond.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", cond.Limit)
	}
	err := x.SQL(sql, params...).Find(&result)
	if err != nil {
		log.Logger.Error("Get alarms fail", log.Error(err))
		return err, result
	}
	var notifyIdList, alarmIdList, alarmStrategyList, endpointList, logKeywordConfigList, dbKeywordMonitorList []string
	for _, v := range result {
		v.StartString = v.Start.Format(m.DatetimeFormat)
		v.EndString = v.End.Format(m.DatetimeFormat)
		if v.AlarmName == "" {
			v.AlarmName = v.Content
		}
		alarmStrategyList = append(alarmStrategyList, v.AlarmStrategy)
		endpointList = append(endpointList, v.Endpoint)
		if v.SMetric == "log_monitor" || v.SMetric == "db_keyword_monitor" {
			if v.SMetric == "log_monitor" {
				logKeywordConfigList = append(logKeywordConfigList, v.AlarmStrategy)
			} else {
				dbKeywordMonitorList = append(dbKeywordMonitorList, v.AlarmStrategy)
			}
			v.IsLogMonitor = true
			v.Log = v.Content
			if v.EndValue > 0 {
				//v.Start, v.End = v.End, v.Start
				if v.EndValue < v.StartValue {
					v.StartValue = v.EndValue
				} else {
					v.StartValue = v.EndValue - v.StartValue + 1
				}
				if strings.Contains(v.Log, "^^") {
					if brIndex := strings.Index(v.Log, "<br/>"); brIndex > 0 {
						v.Content = v.Log[:brIndex+5]
						v.Log = v.Log[brIndex+5:]
					} else {
						v.Content = ""
					}
					v.Log = fmt.Sprintf("%s: %s <br/>%s: %s", v.StartString, v.Log[:strings.Index(v.Log, "^^")], v.EndString, v.Log[strings.Index(v.Log, "^^")+2:])
				}
				//v.StartString = v.EndString
			} else {
				v.StartValue = 1
				if brIndex := strings.Index(v.Log, "<br/>"); brIndex > 0 {
					v.Content = v.Log[:brIndex+5]
					v.Log = v.Log[brIndex+5:]
				} else {
					v.Content = ""
				}
				if strings.HasSuffix(v.Log, "^^") {
					v.Log = v.StartString + ": " + v.Log[:len(v.Log)-2]
				} else {
					v.Log = v.StartString + ": " + v.Log
				}
			}
		}
		if strings.Contains(v.Log, "\n") {
			v.Log = strings.ReplaceAll(v.Log, "\n", "<br/>")
		}
		// 显示 endpointGuid
		v.EndpointGuid = v.Endpoint
		if strings.HasPrefix(v.Endpoint, "sg__") {
			v.Endpoint = v.Endpoint[4:]
			if serviceGroupName, b := m.GlobalSGDisplayNameMap[v.Endpoint]; b {
				v.Endpoint = serviceGroupName
			}
		}
		if strings.HasPrefix(v.Endpoint, "eg__") {
			v.Endpoint = v.Endpoint[4:]
		}
		if v.NotifyId != "" {
			notifyIdList = append(notifyIdList, v.NotifyId)
		}
		if v.Id > 0 {
			alarmIdList = append(alarmIdList, fmt.Sprintf("%d", v.Id))
		}
		var alarmDetailList []*m.AlarmDetailData
		if strings.HasPrefix(v.EndpointTags, "ac_") {
			alarmDetailList, err = GetAlarmDetailList(v.Id)
			if err != nil {
				return err, result
			}
			for _, alarmDetail := range alarmDetailList {
				v.AlarmMetricList = append(v.AlarmMetricList, alarmDetail.Metric)
			}
		} else {
			alarmDetailList = append(alarmDetailList, &m.AlarmDetailData{Metric: v.SMetric, Cond: v.SCond, Last: v.SLast, Start: v.Start, StartValue: v.StartValue, End: v.End, EndValue: v.EndValue, Tags: v.Tags})
			v.AlarmMetricList = []string{v.SMetric}
		}
		v.AlarmDetail = buildAlarmDetailData(alarmDetailList, "<br/>")
	}
	if !cond.ExtOpenAlarm {
		if cond.AlarmTable.Endpoint == "" && len(cond.EndpointFilterList) == 0 && cond.AlarmTable.SMetric == "" && len(cond.MetricFilterList) == 0 {
			cond.ExtOpenAlarm = true
		} else {
			for _, v := range cond.EndpointFilterList {
				if v == "custom_alarm" {
					cond.ExtOpenAlarm = true
					break
				}
			}
			for _, v := range cond.MetricFilterList {
				if v == "custom" {
					cond.ExtOpenAlarm = true
					break
				}
			}
		}
	}
	if cond.ExtOpenAlarm && len(cond.MetricFilterList) == 0 && len(cond.EndpointFilterList) == 0 {
		for _, v := range GetOpenAlarm(m.CustomAlarmQueryParam{Enable: true, Status: "problem", Start: "", End: "", Level: cond.PriorityList, AlterTitleList: cond.AlarmNameFilterList, Query: cond.Query}) {
			result = append(result, v)
		}
	}
	//}
	var sortResult m.AlarmProblemList
	sortResult = result
	if len(result) > 1 {
		sort.Sort(sortResult)
	}
	if len(result) == 0 {
		sortResult = []*m.AlarmProblemQuery{}
	}

	if len(notifyIdList) > 0 {
		var notifyRows []*m.NotifyTable
		filterSql, filterParams := createListParams(notifyIdList, "")
		err = x.SQL("select guid,proc_callback_name,proc_callback_key,description from notify where guid in ("+filterSql+")", filterParams...).Find(&notifyRows)
		if err != nil {
			err = fmt.Errorf("query notify table fail,%s ", err.Error())
			return err, result
		}
		notifyMsgMap := make(map[string]*m.NotifyTable)
		for _, row := range notifyRows {
			notifyMsgMap[row.Guid] = row
		}
		alarmNotifyMap := make(map[int]*m.AlarmNotifyTable)
		if len(alarmIdList) > 0 {
			var alarmNotifyRows []*m.AlarmNotifyTable
			err = x.SQL("select id,alarm_id,status,proc_def_name from alarm_notify where alarm_id in (" + strings.Join(alarmIdList, ",") + ")").Find(&alarmNotifyRows)
			if err != nil {
				err = fmt.Errorf("query alarm_notify table fail,%s ", err.Error())
				return err, result
			}
			for _, v := range alarmNotifyRows {
				alarmNotifyMap[v.AlarmId] = v
			}
		}
		for _, v := range sortResult {
			if notifyRowObj, ok := notifyMsgMap[v.NotifyId]; ok {
				v.NotifyMessage = notifyRowObj.Description
				v.NotifyCallbackName = notifyRowObj.ProcCallbackName
				if alarmNotify, alarmNotifyExists := alarmNotifyMap[v.Id]; alarmNotifyExists {
					v.NotifyStatus = "started"
					if checkHasProcDefUsePermission(alarmNotify, convertString2Map(cond.UserRoles), cond.Token) {
						v.NotifyPermission = "yes"
					}
				} else {
					if notifyRowObj.ProcCallbackName != "" && checkHasProcDefUsePermission(&m.AlarmNotifyTable{ProcDefName: notifyRowObj.ProcCallbackName}, convertString2Map(cond.UserRoles), cond.Token) {
						v.NotifyPermission = "yes"
					}
					v.NotifyStatus = "notStart"
				}
			} else {
				v.NotifyStatus = "notStart"
			}
			v.AlarmObjName = fmt.Sprintf("%d-%s-%s", v.Id, v.Endpoint, v.SMetric)
		}
	}
	if len(alarmStrategyList) > 0 || len(logKeywordConfigList) > 0 || len(dbKeywordMonitorList) > 0 {
		logKeywordConfigMap, dbKeywordMonitorMap, matchKeywordStrategyErr := getAlarmKeywordServiceGroup(logKeywordConfigList, dbKeywordMonitorList)
		if matchKeywordStrategyErr != nil {
			log.Logger.Error("try to match alarm keyword strategy fail", log.Error(matchKeywordStrategyErr))
		}
		strategyGroupMap, endpointServiceMap, matchErr := matchAlarmGroups(alarmStrategyList, endpointList)
		if matchErr != nil {
			log.Logger.Error("try to match alarm groups fail", log.Error(matchErr))
		} else {
			for _, v := range sortResult {
				var tmpStrategyGroups []*m.AlarmStrategyGroup
				if v.SMetric == "log_monitor" {
					if serviceGroup, ok := logKeywordConfigMap[v.AlarmStrategy]; ok {
						tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: serviceGroup, Type: "serviceGroup"})
					}
				} else if v.SMetric == "db_keyword_monitor" {
					if serviceGroup, ok := dbKeywordMonitorMap[v.AlarmStrategy]; ok {
						tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: serviceGroup, Type: "serviceGroup"})
					}
				} else {
					if strategyRow, ok := strategyGroupMap[v.AlarmStrategy]; ok {
						if strategyRow.ServiceGroup == "" {
							tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: strategyRow.EndpointGroup, Type: "endpointGroup"})
							if endpointServiceList, endpointOk := endpointServiceMap[v.Endpoint]; endpointOk {
								for _, endpointServiceRelRow := range endpointServiceList {
									tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: endpointServiceRelRow.ServiceGroup, Type: "serviceGroup"})
								}
							}
						} else {
							tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: strategyRow.ServiceGroup, Type: "serviceGroup"})
						}
					}
				}
				v.StrategyGroups = tmpStrategyGroups
			}
		}
	}
	return err, sortResult
}

func UpdateAlarms(alarms []*m.AlarmHandleObj) []*m.AlarmHandleObj {
	var successAlarms []*m.AlarmHandleObj
	if len(alarms) == 0 {
		return alarms
	}
	//var rowAffected int64
	for _, v := range alarms {
		//rowAffected = 0
		//var action Action
		//var cErr error
		//var execResult sql.Result
		calcAlarmUniqueFlag(&v.AlarmTable)
		if v.MultipleConditionFlag {
			alarmObj, updateConditionAlarmErr := UpdateAlarmWithConditions(v)
			if updateConditionAlarmErr != nil {
				log.Logger.Error("Update alarm condition fail", log.JsonObj("alarm", v), log.Error(updateConditionAlarmErr))
			} else if alarmObj != nil {
				successAlarms = append(successAlarms, alarmObj)
			}
		} else {
			if tmpErr := doInsertOrUpdateAlarm(v); tmpErr != nil {
				log.Logger.Warn("doInsertOrUpdateAlarm fail", log.JsonObj("alarm", v), log.Error(tmpErr))
			} else {
				successAlarms = append(successAlarms, v)
			}
			//if v.Id > 0 {
			//	action.Sql = "UPDATE alarm SET status=?,end_value=?,end=? WHERE id=? AND status='firing'"
			//	execResult, cErr = x.Exec(action.Sql, v.Status, v.EndValue, v.End.Format(m.DatetimeFormat), v.Id)
			//} else {
			//	action.Sql = "INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start,tags,endpoint_tags,alarm_strategy,alarm_name) VALUE (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
			//	execResult, cErr = x.Exec(action.Sql, v.StrategyId, v.Endpoint, v.Status, v.SMetric, v.SExpr, v.SCond, v.SLast, v.SPriority, v.Content, v.StartValue, v.Start.Format(m.DatetimeFormat), v.Tags, v.EndpointTags, v.AlarmStrategy, v.AlarmName)
			//}
			//if cErr != nil {
			//	log.Logger.Error("Update alarm fail", log.JsonObj("alarm", v), log.Error(cErr))
			//} else {
			//	rowAffected, _ = execResult.RowsAffected()
			//	if rowAffected > 0 {
			//		if v.Id <= 0 {
			//			lastInsertId, _ := execResult.LastInsertId()
			//			if lastInsertId > 0 {
			//				v.Id = int(lastInsertId)
			//			}
			//		}
			//		successAlarms = append(successAlarms, v)
			//	} else {
			//		log.Logger.Warn("Update alarm done but not any rows affected", log.JsonObj("alarm", v))
			//	}
			//}
		}
	}
	return successAlarms
}

func judgeExist(alarm m.AlarmTable) bool {
	var result []*m.AlarmTable
	x.SQL("SELECT * FROM alarm WHERE strategy_id=? AND endpoint=? AND status=? AND s_metric=? AND s_expr=? AND s_cond=? AND s_last=? AND s_priority=?",
		alarm.StrategyId, alarm.Endpoint, alarm.Status, alarm.SMetric, alarm.SExpr, alarm.SCond, alarm.SLast, alarm.SPriority).Find(&result)
	if len(result) > 0 {
		return true
	}
	return false
}

func UpdateLogMonitor(obj *m.UpdateLogMonitor) error {
	var actions []*Action
	for _, v := range obj.LogMonitor {
		action := Classify(*v, obj.Operation, "log_monitor", true)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	err := Transaction(actions)
	return err
}

func AutoUpdateLogMonitor(obj *m.UpdateLogMonitor) error {
	if len(obj.LogMonitor) == 0 {
		return fmt.Errorf("update log monitor fail,data empty")
	}
	var err error
	if obj.Operation == "add" {
		var logMonitorTable []*m.LogMonitorTable
		x.SQL("SELECT * FROM log_monitor WHERE strategy_id=? AND path=? AND keyword=?", obj.LogMonitor[0].StrategyId, obj.LogMonitor[0].Path, obj.LogMonitor[0].Keyword).Find(&logMonitorTable)
		if len(logMonitorTable) == 0 {
			_, err = x.Exec("INSERT INTO log_monitor(strategy_id,path,keyword,priority) VALUE (?,?,?,?)", obj.LogMonitor[0].StrategyId, obj.LogMonitor[0].Path, obj.LogMonitor[0].Keyword, obj.LogMonitor[0].Priority)
		}
	}
	if obj.Operation == "delete" {
		_, err = x.Exec("DELETE FROM log_monitor WHERE strategy_id=? AND path=? AND keyword=?", obj.LogMonitor[0].StrategyId, obj.LogMonitor[0].Path, obj.LogMonitor[0].Keyword)
	}
	return err
}

func GetLogMonitorTable(id, strategyId, tplId int, path string) (err error, result []*m.LogMonitorTable) {
	if id > 0 {
		err = x.SQL("SELECT * FROM log_monitor WHERE id=?", id).Find(&result)
	}
	if path != "" && strategyId > 0 {
		err = x.SQL("SELECT * FROM log_monitor WHERE strategy_id=? and path=?", strategyId, path).Find(&result)
	} else {
		if path != "" {
			err = x.SQL("SELECT * FROM log_monitor WHERE path=?", path).Find(&result)
		}
		if strategyId > 0 {
			err = x.SQL("SELECT * FROM log_monitor WHERE strategy_id=?", strategyId).Find(&result)
		}
	}
	if tplId > 0 {
		err = x.SQL("SELECT * FROM log_monitor WHERE strategy_id IN (SELECT id FROM strategy WHERE tpl_id=?) ORDER BY path", tplId).Find(&result)
	}
	return err, result
}

func GetLogMonitorByEndpoint(endpointId int) (err error, result []*m.LogMonitorTable) {
	sql := `SELECT DISTINCT t1.* FROM log_monitor t1 
			LEFT JOIN strategy t2 ON t1.strategy_id=t2.id 
			LEFT JOIN tpl t3 ON t2.tpl_id=t3.id 
			WHERE t3.endpoint_id=? 
			OR t3.grp_id IN (SELECT grp_id FROM grp_endpoint WHERE endpoint_id=?) 
			ORDER BY t1.path`
	err = x.SQL(sql, endpointId, endpointId).Find(&result)
	return err, result
}

func GetLogMonitorByEndpointNew(endpointId int) (err error, result []*m.LogMonitorTable) {
	err = x.SQL("SELECT * FROM log_monitor WHERE strategy_id=? order by path", endpointId).Find(&result)
	return err, result
}

func ListLogMonitorNew(query *m.TplQuery) error {
	var result []*m.TplObj
	if query.SearchType == "endpoint" {
		var logMonitorTable []*m.LogMonitorTable
		err := x.SQL("SELECT * FROM log_monitor where strategy_id=? order by path,keyword", query.SearchId).Find(&logMonitorTable)
		if err != nil {
			return err
		}
		endpointObj := m.EndpointTable{Id: query.SearchId}
		GetEndpoint(&endpointObj)
		if len(logMonitorTable) == 0 {
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}, LogMonitor: []*m.LogMonitorDto{}})
			query.Tpl = result
			return nil
		}
		var lms []*m.LogMonitorDto
		var tmpKeywords []*m.LogMonitorStrategyDto
		tmpPath := logMonitorTable[0].Path
		for i, v := range logMonitorTable {
			if v.Path != tmpPath {
				lms = append(lms, &m.LogMonitorDto{Id: logMonitorTable[i-1].Id, EndpointId: v.StrategyId, Path: tmpPath, Strategy: tmpKeywords, OwnerEndpoint: logMonitorTable[i-1].OwnerEndpoint})
				tmpPath = v.Path
				tmpKeywords = []*m.LogMonitorStrategyDto{}
			}
			tmpKeywords = append(tmpKeywords, &m.LogMonitorStrategyDto{Id: v.Id, Keyword: v.Keyword, Priority: v.Priority, NotifyEnable: v.NotifyEnable})
		}
		if len(tmpKeywords) > 0 {
			lms = append(lms, &m.LogMonitorDto{Id: logMonitorTable[len(logMonitorTable)-1].Id, EndpointId: logMonitorTable[len(logMonitorTable)-1].StrategyId, Path: logMonitorTable[len(logMonitorTable)-1].Path, Strategy: tmpKeywords, OwnerEndpoint: logMonitorTable[len(logMonitorTable)-1].OwnerEndpoint})
		}
		result = append(result, &m.TplObj{Operation: true, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", LogMonitor: lms})
	}
	query.Tpl = result
	return nil
}

func ListLogMonitor(query *m.TplQuery) error {
	var result []*m.TplObj
	if query.SearchType == "endpoint" {
		var grps []*m.GrpTable
		err := x.SQL("SELECT id,name FROM grp where id in (select grp_id from grp_endpoint WHERE endpoint_id=?)", query.SearchId).Find(&grps)
		if err != nil {
			log.Logger.Error("Get strategy fail", log.Error(err))
			return err
		}
		var grpIds string
		grpMap := make(map[int]string)
		if len(grps) > 0 {
			grpIds = "t1.grp_id IN ("
			for _, v := range grps {
				grpIds += fmt.Sprintf("%d,", v.Id)
				grpMap[v.Id] = v.Name
			}
			grpIds = grpIds[:len(grpIds)-1]
			grpIds += ") OR"
		}
		var tpls []*m.TplStrategyLogMonitorTable
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.expr,t2.cond,t2.last,t2.priority,t3.id log_monitor_id,t3.path,t3.keyword FROM tpl t1 
				LEFT JOIN strategy t2 ON t1.id=t2.tpl_id 
				LEFT JOIN log_monitor t3 ON t2.id=t3.strategy_id 
				WHERE (` + grpIds + ` t1.endpoint_id=?) and t2.config_type='log_monitor' ORDER BY t1.endpoint_id,t1.id,t3.path`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			log.Logger.Error("Get log monitor strategy fail", log.Error(err))
			return err
		}
		if len(tpls) == 0 {
			endpointObj := m.EndpointTable{Id: query.SearchId}
			GetEndpoint(&endpointObj)
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}, LogMonitor: []*m.LogMonitorDto{}})
		} else {
			var tmpTplId int
			var tmpLogMonitor []*m.LogMonitorDto
			keywordMap := make(map[string][]*m.LogMonitorStrategyDto)
			for _, v := range tpls {
				key := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if vv, b := keywordMap[key]; !b {
					keywordMap[key] = []*m.LogMonitorStrategyDto{&m.LogMonitorStrategyDto{Id: v.LogMonitorId, StrategyId: v.StrategyId, Keyword: v.Keyword, Cond: v.Cond, Last: getLastFromExpr(v.Expr), Priority: v.Priority}}
				} else {
					keywordMap[key] = append(vv, &m.LogMonitorStrategyDto{Id: v.LogMonitorId, StrategyId: v.StrategyId, Keyword: v.Keyword, Cond: v.Cond, Last: getLastFromExpr(v.Expr), Priority: v.Priority})
				}
			}
			existFlag := make(map[string]int)
			for i, v := range tpls {
				tmpMapKey := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if i == 0 {
					tmpTplId = v.TplId
					if v.StrategyId > 0 {
						if _, b := existFlag[tmpMapKey]; !b {
							tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id: v.StrategyId, TplId: v.TplId, Path: v.Path, Strategy: keywordMap[tmpMapKey]})
							existFlag[tmpMapKey] = 1
						}
					}
				} else {
					if v.TplId != tmpTplId {
						tmpTplObj := m.TplObj{TplId: tpls[i-1].TplId}
						if tpls[i-1].GrpId > 0 {
							tmpTplObj.ObjId = tpls[i-1].GrpId
							tmpTplObj.ObjName = grpMap[tpls[i-1].GrpId]
							tmpTplObj.ObjType = "grp"
							tmpTplObj.Operation = false
						} else {
							tmpTplObj.ObjId = tpls[i-1].EndpointId
							endpointObj := m.EndpointTable{Id: tpls[i-1].EndpointId}
							GetEndpoint(&endpointObj)
							tmpTplObj.ObjName = endpointObj.Guid
							tmpTplObj.ObjType = "endpoint"
							tmpTplObj.Operation = true
						}
						tmpTplObj.LogMonitor = tmpLogMonitor
						result = append(result, &tmpTplObj)
						tmpTplId = v.TplId
						tmpLogMonitor = []*m.LogMonitorDto{}
					}
					if v.StrategyId > 0 {
						if _, b := existFlag[tmpMapKey]; !b {
							tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id: v.StrategyId, TplId: v.TplId, Path: v.Path, Strategy: keywordMap[tmpMapKey]})
							existFlag[tmpMapKey] = 1
						}
					}
				}
			}
			if tpls[len(tpls)-1].EndpointId > 0 {
				endpointObj := m.EndpointTable{Id: tpls[len(tpls)-1].EndpointId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: tpls[len(tpls)-1].TplId, ObjId: tpls[len(tpls)-1].EndpointId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, LogMonitor: tmpLogMonitor})
			} else {
				result = append(result, &m.TplObj{TplId: tpls[len(tpls)-1].TplId, ObjId: tpls[len(tpls)-1].GrpId, ObjName: grpMap[tpls[len(tpls)-1].GrpId], ObjType: "grp", Operation: false, LogMonitor: tmpLogMonitor})
				endpointObj := m.EndpointTable{Id: query.SearchId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}, LogMonitor: []*m.LogMonitorDto{}})
			}
		}
	} else {
		var grps []*m.GrpTable
		err := x.SQL("SELECT * FROM grp WHERE id=?", query.SearchId).Find(&grps)
		if err != nil {
			log.Logger.Error("Get group fail", log.Error(err))
			return err
		}
		if len(grps) <= 0 {
			log.Logger.Warn("Can't find this grp", log.Int("grpId", query.SearchId))
			return fmt.Errorf("can't find this grp")
		}
		var tpls []*m.TplStrategyLogMonitorTable
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.expr,t2.cond,t2.last,t2.priority,t3.id log_monitor_id,t3.path,t3.keyword FROM tpl t1 
			LEFT JOIN strategy t2 ON t1.id=t2.tpl_id 
			LEFT JOIN log_monitor t3 ON t2.id=t3.strategy_id 
			WHERE t1.grp_id=? and t2.config_type='log_monitor' ORDER BY t1.endpoint_id,t1.id,t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			log.Logger.Error("Get log monitor strategy fail", log.Error(err))
			return err
		}
		if len(tpls) > 0 {
			keywordMap := make(map[string][]*m.LogMonitorStrategyDto)
			for _, v := range tpls {
				tmpMapKey := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if vv, b := keywordMap[tmpMapKey]; !b {
					keywordMap[tmpMapKey] = []*m.LogMonitorStrategyDto{&m.LogMonitorStrategyDto{StrategyId: v.StrategyId, Keyword: v.Keyword, Cond: v.Cond, Last: getLastFromExpr(v.Expr), Priority: v.Priority}}
				} else {
					keywordMap[tmpMapKey] = append(vv, &m.LogMonitorStrategyDto{StrategyId: v.StrategyId, Keyword: v.Keyword, Cond: v.Cond, Last: getLastFromExpr(v.Expr), Priority: v.Priority})
				}
			}
			tmpLogMonitor := []*m.LogMonitorDto{}
			existFlag := make(map[string]int)
			for _, v := range tpls {
				tmpMapKey := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if v.StrategyId > 0 {
					if _, b := existFlag[tmpMapKey]; !b {
						tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id: v.StrategyId, TplId: v.TplId, Path: v.Path, Strategy: keywordMap[fmt.Sprintf("%d^%s", v.TplId, v.Path)]})
						existFlag[tmpMapKey] = 1
					}
				}
			}
			result = append(result, &m.TplObj{TplId: tpls[0].TplId, ObjId: tpls[0].GrpId, ObjName: grps[0].Name, ObjType: "grp", Operation: true, LogMonitor: tmpLogMonitor})
		} else {
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: grps[0].Name, ObjType: "grp", Operation: true, LogMonitor: []*m.LogMonitorDto{}})
		}
	}
	query.Tpl = result
	return nil
}

func getLastFromExpr(expr string) string {
	var last string
	if strings.Contains(expr, "[") {
		last = strings.Split(strings.Split(expr, "[")[1], "]")[0]
	} else {
		last = "10s"
	}
	return last
}

func CloseAlarm(param m.AlarmCloseParam) (actions []*Action, err error) {
	var alarmRows []*m.AlarmTable
	var queryFilterList []string
	var queryFilterParam []interface{}
	if param.Id > 0 {
		queryFilterList = append(queryFilterList, "id=?")
		queryFilterParam = append(queryFilterParam, param.Id)
	} else {
		if len(param.Priority) > 0 {
			filterSql, filterParam := createListParams(param.Priority, "")
			queryFilterList = append(queryFilterList, "s_priority in ("+filterSql+")")
			queryFilterParam = append(queryFilterParam, filterParam...)
		}
		if len(param.Metric) > 0 {
			filterSql, filterParam := createListParams(param.Metric, "")
			queryFilterList = append(queryFilterList, "s_metric in ("+filterSql+")")
			queryFilterParam = append(queryFilterParam, filterParam...)
		}
		if len(param.Endpoint) > 0 {
			filterSql, filterParam := createListParams(param.Endpoint, "")
			queryFilterList = append(queryFilterList, "endpoint in ("+filterSql+")")
			queryFilterParam = append(queryFilterParam, filterParam...)
		}
		if len(param.AlarmName) > 0 {
			filterSql, filterParam := createListParams(param.AlarmName, "")
			queryFilterList = append(queryFilterList, "alarm_name in ("+filterSql+")")
			queryFilterParam = append(queryFilterParam, filterParam...)
		}
	}
	if len(queryFilterList) == 0 {
		err = fmt.Errorf("close filter can not empty")
		return
	}
	err = x.SQL("select id,s_metric,endpoint_tags from alarm WHERE status='firing' and "+strings.Join(queryFilterList, " and "), queryFilterParam...).Find(&alarmRows)
	if err != nil {
		err = fmt.Errorf("query alarm table fail,%s ", err.Error())
		return
	}
	for _, v := range alarmRows {
		actions = append(actions, &Action{Sql: "UPDATE alarm SET STATUS='closed',end=NOW() WHERE id=?", Param: []interface{}{v.Id}})
		if v.SMetric == "log_monitor" {
			actions = append(actions, &Action{Sql: "update log_keyword_alarm set status='closed',updated_time=NOW() WHERE alarm_id=?", Param: []interface{}{v.Id}})
		}
		if v.SMetric == "db_keyword_monitor" {
			actions = append(actions, &Action{Sql: "update db_keyword_alarm set status='closed',updated_time=NOW() WHERE alarm_id=?", Param: []interface{}{v.Id}})
		}
		if strings.HasPrefix(v.EndpointTags, "ac_") {
			actions = append(actions, &Action{Sql: "UPDATE alarm_condition SET STATUS='closed',end=NOW() WHERE guid in (select alarm_condition from alarm_condition_rel where alarm=?)", Param: []interface{}{v.Id}})
		}
		actions = append(actions, &Action{Sql: "delete from alarm_firing where alarm_id=?", Param: []interface{}{v.Id}})
	}
	return
}

func UpdateAlarmCustomMessage(param m.UpdateAlarmCustomMessageDto) error {
	var err error
	if param.IsCustom {
		_, err = x.Exec("UPDATE alarm_custom SET custom_message=? WHERE id=?", param.Message, param.Id)
	} else {
		_, err = x.Exec("UPDATE alarm SET custom_message=? WHERE id=?", param.Message, param.Id)
	}
	return err
}

func GetGrpStrategy(idList []string) (err error, result []*m.GrpStrategyExportObj) {
	sql := `SELECT t1.name,t1.description,t3.metric,t3.expr,t3.cond,t3.last,t3.priority,t3.content,t3.config_type 
		FROM grp t1 
		LEFT JOIN tpl t2 ON t1.id=t2.grp_id 
		LEFT JOIN strategy t3 ON t2.id=t3.tpl_id 
		WHERE t1.id IN `
	sql = sql + fmt.Sprintf("(%s)", strings.Join(idList, ",")) + " ORDER BY t1.name"
	var queryResult []*m.GrpStrategyQuery
	err = x.SQL(sql).Find(&queryResult)
	if err != nil {
		return err, result
	}
	if len(queryResult) == 0 {
		return nil, result
	}
	var tmpStrategyList []m.StrategyTable
	tmpName := queryResult[0].Name
	for i, v := range queryResult {
		if v.Name != tmpName {
			tmpObj := m.GrpStrategyExportObj{GrpName: tmpName, Description: queryResult[i-1].Description, Strategy: tmpStrategyList}
			result = append(result, &tmpObj)
			tmpStrategyList = []m.StrategyTable{}
			tmpName = v.Name
		}
		tmpStrategyList = append(tmpStrategyList, m.StrategyTable{Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content, ConfigType: v.ConfigType})
	}
	tmpObj := m.GrpStrategyExportObj{GrpName: tmpName, Description: queryResult[len(queryResult)-1].Description, Strategy: tmpStrategyList}
	result = append(result, &tmpObj)
	return nil, result
}

func SetGrpStrategy(paramObj []*m.GrpStrategyExportObj) error {
	if len(paramObj) == 0 {
		return nil
	}
	var existGrp []*m.GrpTable
	err := x.SQL("SELECT * FROM grp order by name").Find(&existGrp)
	if err != nil {
		return err
	}
	for _, v := range paramObj {
		tmpName := takeGrpName(v.GrpName, existGrp)
		err := UpdateGrp(&m.UpdateGrp{Operation: "insert", Groups: []*m.GrpTable{&m.GrpTable{Name: tmpName, Description: v.Description}}})
		if err != nil {
			log.Logger.Error("Set group strategy, insert group fail", log.Error(err))
			return err
		}
		_, grpObj := GetSingleGrp(0, tmpName)
		err, tplObj := AddTpl(grpObj.Id, 0, "")
		if err != nil {
			log.Logger.Error("Set group strategy, insert tpl fail", log.Error(err))
			return err
		}
		for _, vv := range v.Strategy {
			strategyObj := m.StrategyTable{TplId: tplObj.Id, Metric: vv.Metric, Expr: vv.Expr, Cond: vv.Cond, Last: vv.Last, Priority: vv.Priority, Content: vv.Content, ConfigType: vv.ConfigType}
			UpdateStrategy(&m.UpdateStrategy{Strategy: []*m.StrategyTable{&strategyObj}, Operation: "insert"})
		}
	}
	return nil
}

func takeGrpName(name string, grpList []*m.GrpTable) string {
	exist := false
	tmpIndex := 0
	for _, v := range grpList {
		if v.Name == name {
			exist = true
		}
		if strings.HasPrefix(v.Name, name) && strings.Contains(v.Name, "_") {
			tmpList := strings.Split(v.Name, "_")
			ii, _ := strconv.Atoi(tmpList[len(tmpList)-1])
			if ii > tmpIndex {
				tmpIndex = ii
			}
		}
	}
	if !exist {
		return name
	} else {
		if tmpIndex > 0 {
			name = strings.Replace(name, fmt.Sprintf("_%d", tmpIndex), "", -1)
		}
		return fmt.Sprintf("%s_%d", name, tmpIndex+1)
	}
}

func DeleteStrategyByGrp(grpId int, tplId int) error {
	var action Action
	params := make([]interface{}, 0)
	if grpId > 0 {
		action.Sql = "DELETE FROM grp_endpoint WHERE grp_id=?"
		params = append(params, grpId)
	}
	if tplId > 0 {
		action.Sql = "DELETE FROM strategy WHERE tpl_id=?"
		params = append(params, tplId)
	}
	if action.Sql == "" {
		return nil
	}
	return Transaction([]*Action{&action})
}

func SaveOpenAlarm(param m.OpenAlarmRequest) error {
	var err error
	var alertLevel, subSystemId int
	for _, v := range param.AlertList {
		if len(v.AlertInfo) > 1024 {
			v.AlertInfo = v.AlertInfo[:1024]
		}
		var customAlarmId int
		var query []*m.OpenAlarmObj
		x.SQL("SELECT * FROM alarm_custom WHERE alert_title=? AND closed=0", v.AlertTitle).Find(&query)
		if v.AlertLevel == "0" {
			if len(query) > 0 {
				tmpIds := ""
				for _, vv := range query {
					tmpIds += fmt.Sprintf("%d,", vv.Id)
					customAlarmId = vv.Id
				}
				tmpIds = tmpIds[:len(tmpIds)-1]
				_, cErr := x.Exec(fmt.Sprintf("UPDATE alarm_custom SET closed=1,closed_at=NOW() WHERE id in (%s)", tmpIds))
				if cErr != nil {
					log.Logger.Error("Update custom alarm close fail", log.String("ids", tmpIds), log.Error(cErr))
				}
			} else {
				log.Logger.Warn("Get recover custom alarm,but not found in table", log.JsonObj("input", v))
				continue
			}
		} else {
			if len(query) > 0 {
				continue
			}
			alertLevel, _ = strconv.Atoi(v.AlertLevel)
			subSystemId, _ = strconv.Atoi(v.SubSystemId)
			_, err = x.Exec("INSERT INTO alarm_custom(alert_info,alert_ip,alert_level,alert_obj,alert_title,alert_reciver,remark_info,sub_system_id,use_umg_policy,alert_way) VALUE (?,?,?,?,?,?,?,?,?,?)", v.AlertInfo, v.AlertIp, alertLevel, v.AlertObj, v.AlertTitle, v.AlertReciver, v.RemarkInfo, subSystemId, v.UseUmgPolicy, v.AlertWay)
			if err != nil {
				log.Logger.Error("Save open alarm error", log.Error(err))
				err = fmt.Errorf("Update database fail,%s ", err.Error())
				break
			}
			x.SQL("SELECT * FROM alarm_custom WHERE alert_title=? AND closed=0", v.AlertTitle).Find(&query)
			for _, vv := range query {
				customAlarmId = vv.Id
			}
		}
		if v.UseUmgPolicy != "1" && customAlarmId > 0 {
			if mailAlarmObj, buildMailAlarmErr := getCustomAlarmEvent(customAlarmId); buildMailAlarmErr != nil {
				log.Logger.Error("Build custom alarm mail fail", log.Int("Id", customAlarmId), log.Error(buildMailAlarmErr))
			} else {
				if mailSender, getMailSenderErr := GetMailSender(); getMailSenderErr != nil {
					log.Logger.Error("Try to send custom alarm mail fail", log.Error(getMailSenderErr))
				} else {
					if sendErr := mailSender.Send(mailAlarmObj.Subject, mailAlarmObj.Content, strings.Split(mailAlarmObj.ToMail, ",")); sendErr != nil {
						log.Logger.Error("Try to send custom alarm mail fail", log.Error(sendErr))
					}
				}
			}
			sendMailErr := NotifyCoreEvent("", 0, 0, customAlarmId)
			if sendMailErr != nil {
				log.Logger.Error("Send custom alarm mail event fail", log.Error(sendMailErr))
			}
		}
	}
	return err
}

func GetOpenAlarm(param m.CustomAlarmQueryParam) []*m.AlarmProblemQuery {
	var query []*m.OpenAlarmObj
	var sql string
	result := []*m.AlarmProblemQuery{}
	//sql := fmt.Sprintf("SELECT * FROM alarm_custom WHERE closed<>1 and update_at>'%s' ORDER BY id ASC", time.Unix(time.Now().Unix()-300,0).Format(m.DatetimeFormat))
	if param.Status == "problem" {
		sql = "SELECT * FROM alarm_custom WHERE closed=0 "
	} else {
		if param.Start != "" && param.End != "" {
			sql = fmt.Sprintf("SELECT * FROM alarm_custom WHERE update_at<='%s' AND update_at>'%s' ", param.End, param.Start)
		}
	}
	levelFilterSql := getLevelSQL(convertString2Map(param.Level))
	if len(levelFilterSql) > 0 {
		sql += levelFilterSql
	}
	if len(param.AlterTitleList) > 0 {
		var alterTitleFilterList []string
		for _, v := range param.AlterTitleList {
			alterTitleFilterList = append(alterTitleFilterList, fmt.Sprintf("alert_title like '%%%s%%'", v))
		}
		sql += " AND (" + strings.Join(alterTitleFilterList, " OR ") + ")"
	}
	if param.Query != "" {
		sql += " AND (alert_info like '%" + param.Query + "%' OR alert_title like '%" + param.Query + "%')"
	}
	sql += " ORDER BY id DESC"
	x.SQL(sql).Find(&query)
	if len(query) == 0 {
		return result
	}
	tmpFlag := fmt.Sprintf("%d_%s_%s_%d", query[0].SubSystemId, query[0].AlertTitle, query[0].AlertIp, query[0].AlertLevel)
	for i, v := range query {
		if tmpFlag != fmt.Sprintf("%d_%s_%s_%d", v.SubSystemId, v.AlertTitle, v.AlertIp, v.AlertLevel) {
			priority := "high"
			tmpAlertLevel, _ := strconv.Atoi(query[i-1].AlertLevel)
			if tmpAlertLevel > 4 {
				priority = "low"
			} else if tmpAlertLevel > 2 {
				priority = "medium"
			}
			query[i-1].AlertInfo = strings.Replace(query[i-1].AlertInfo, "\n", " <br/> ", -1)
			//tmpDisplayEndpoint := fmt.Sprintf("%s(%s)", query[i-1].AlertObj, query[i-1].AlertIp)
			//if query[i-1].AlertObj == "" && query[i-1].AlertIp == "" {
			//	tmpDisplayEndpoint = "custom_alarm"
			//}
			tmpDisplayEndpoint := "custom_alarm"
			result = append(result, &m.AlarmProblemQuery{IsCustom: true, Id: query[i-1].Id, Endpoint: tmpDisplayEndpoint, Status: "firing", Content: query[i-1].AlertInfo, Start: query[i-1].UpdateAt, StartString: query[i-1].UpdateAt.Format(m.DatetimeFormat), SPriority: priority, SMetric: "custom", CustomMessage: query[i-1].CustomMessage, Title: query[i-1].AlertTitle, SystemId: query[i-1].SubSystemId})
		}
	}
	priority := "high"
	lastIndex := len(query) - 1
	tmpAlertLevel, _ := strconv.Atoi(query[lastIndex].AlertLevel)
	if tmpAlertLevel > 4 {
		priority = "low"
	} else if tmpAlertLevel > 2 {
		priority = "medium"
	}
	query[lastIndex].AlertInfo = strings.Replace(query[lastIndex].AlertInfo, "\n", " <br/> ", -1)
	//tmpDisplayEndpoint := fmt.Sprintf("%s(%s)", query[lastIndex].AlertObj, query[lastIndex].AlertIp)
	//if query[lastIndex].AlertObj == "" && query[lastIndex].AlertIp == "" {
	//	tmpDisplayEndpoint = "custom_alarm"
	//}
	tmpDisplayEndpoint := "custom_alarm"
	result = append(result, &m.AlarmProblemQuery{IsCustom: true, Id: query[lastIndex].Id, Endpoint: tmpDisplayEndpoint, Status: "firing", IsLogMonitor: false, Content: query[lastIndex].AlertInfo, Start: query[lastIndex].UpdateAt, StartString: query[lastIndex].UpdateAt.Format(m.DatetimeFormat), SPriority: priority, SMetric: "custom", CustomMessage: query[lastIndex].CustomMessage, Title: query[lastIndex].AlertTitle, SystemId: query[lastIndex].SubSystemId})
	return result
}

func CloseOpenAlarm(param m.AlarmCloseParam) (actions []*Action, err error) {
	var query []*m.OpenAlarmObj
	containsCustomMetric := false
	for _, v := range param.Metric {
		if strings.ToLower(v) == "custom" {
			containsCustomMetric = true
			break
		}
	}
	priorityMap := make(map[string][]string)
	priorityMap["high"] = []string{"1", "2"}
	priorityMap["medium"] = []string{"3", "4"}
	priorityMap["low"] = []string{"5"}
	if containsCustomMetric && param.Id == 0 {
		x.SQL("SELECT * FROM alarm_custom WHERE closed=0").Find(&query)
	} else if len(param.Priority) > 0 {
		levelList := []string{}
		for _, v := range param.Priority {
			if levelValue, ok := priorityMap[v]; ok {
				levelList = append(levelList, levelValue...)
			}
		}
		x.SQL("SELECT * FROM alarm_custom WHERE closed=0 AND alert_level in (" + strings.Join(levelList, ",") + ")").Find(&query)
	} else {
		x.SQL("SELECT * FROM alarm_custom WHERE id=?", param.Id).Find(&query)
	}
	if len(query) == 0 {
		return
	}
	for _, v := range query {
		subQueryList := []*m.OpenAlarmObj{}
		err = x.SQL("SELECT id FROM alarm_custom WHERE alert_ip=? AND alert_title=? AND alert_obj=?", v.AlertIp, v.AlertTitle, v.AlertObj).Find(&subQueryList)
		if len(subQueryList) > 0 {
			tmpIds := ""
			for _, vv := range subQueryList {
				tmpIds += fmt.Sprintf("%d,", vv.Id)
			}
			tmpIds = tmpIds[:len(tmpIds)-1]
			actions = append(actions, &Action{Sql: fmt.Sprintf("UPDATE alarm_custom SET closed=1,closed_at=NOW() WHERE id in (%s)", tmpIds), Param: []interface{}{}})
			//_, err = x.Exec(fmt.Sprintf("UPDATE alarm_custom SET closed=1,closed_at=NOW() WHERE id in (%s)", tmpIds))
			//if err != nil {
			//	log.Logger.Error("Update custom alarm close fail", log.String("ids", tmpIds), log.Error(err))
			//}
		}
		if err != nil {
			break
		}
	}
	return
}

func UpdateTplAction(tplId int, user, role []int, extraMail, extraPhone []string) error {
	var userString, roleString, mailString, phoneString string
	if len(user) > 0 {
		for _, v := range user {
			userString += fmt.Sprintf("%d,", v)
		}
		userString = userString[:len(userString)-1]
	}
	if len(role) > 0 {
		for _, v := range role {
			roleString += fmt.Sprintf("%d,", v)
		}
		roleString = roleString[:len(roleString)-1]
	}
	if len(extraMail) > 0 {
		mailString = strings.Join(extraMail, ",")
	}
	if len(extraPhone) > 0 {
		phoneString = strings.Join(extraPhone, ",")
	}
	_, err := x.Exec(fmt.Sprintf("UPDATE tpl SET action_user='%s',action_role='%s',extra_mail='%s',extra_phone='%s' WHERE id=%d", userString, roleString, mailString, phoneString, tplId))
	if err != nil {
		log.Logger.Error("Update tpl action error", log.Error(err))
	}
	return err
}

func getActionOptions(tplId int) []*m.OptionModel {
	var tpls []*m.TplTable
	result := []*m.OptionModel{}
	x.SQL("SELECT * FROM tpl WHERE id=?", tplId).Find(&tpls)
	if len(tpls) == 0 {
		return result
	}
	if tpls[0].ActionRole != "" {
		var roles []*m.RoleTable
		x.SQL(fmt.Sprintf("SELECT id,name,display_name FROM role WHERE id IN (%s)", tpls[0].ActionRole)).Find(&roles)
		for _, v := range roles {
			tmpText := v.Name
			if v.DisplayName != "" {
				tmpText = tmpText + "(" + v.DisplayName + ")"
			}
			result = append(result, &m.OptionModel{Id: v.Id, OptionText: tmpText, OptionValue: fmt.Sprintf("%d", v.Id), Active: false, OptionType: fmt.Sprintf("role_%d", v.Id)})
		}
	}
	if tpls[0].ActionUser != "" {
		var users []*m.UserTable
		x.SQL(fmt.Sprintf("SELECT id,NAME,display_name FROM user WHERE id IN (%s)", tpls[0].ActionUser)).Find(&users)
		for _, v := range users {
			tmpText := v.Name
			if v.DisplayName != "" {
				tmpText = tmpText + "(" + v.DisplayName + ")"
			}
			result = append(result, &m.OptionModel{Id: v.Id, OptionText: tmpText, OptionValue: fmt.Sprintf("%d", v.Id), Active: false, OptionType: fmt.Sprintf("user_%d", v.Id)})
		}
	}
	if tpls[0].ExtraMail != "" {
		for _, v := range strings.Split(tpls[0].ExtraMail, ",") {
			result = append(result, &m.OptionModel{Id: 0, OptionText: v, OptionValue: v, Active: false, OptionType: "mail"})
		}
	}
	if tpls[0].ExtraPhone != "" {
		for _, v := range strings.Split(tpls[0].ExtraPhone, ",") {
			result = append(result, &m.OptionModel{Id: 0, OptionText: v, OptionValue: v, Active: false, OptionType: "phone"})
		}
	}
	return result
}

func GetServiceGroupHistoryAlarm(param m.ServiceGroupAlarmParam) (data m.AlarmProblemList, count int, err error) {
	startIndex := (param.Page - 1) * param.PageSize
	var whereSql = " where endpoint = ?"
	if !param.StartTime.IsZero() {
		whereSql += fmt.Sprintf(" and start>='%s' ", param.StartTime.Format(m.DatetimeFormat))
	}
	if !param.EndTime.IsZero() {
		whereSql += fmt.Sprintf(" and end<='%s' ", param.EndTime.Format(m.DatetimeFormat))
	}
	serviceGroupEndpoint := "sg__" + param.ServiceGroup
	countSql := "select count(1) FROM alarm " + whereSql
	_, err = x.SQL(countSql, serviceGroupEndpoint).Get(&count)
	if err != nil {
		return
	}
	sql := "SELECT * FROM alarm " + whereSql + " ORDER BY id DESC limit ?,?"
	err = x.SQL(sql, serviceGroupEndpoint, startIndex, param.PageSize).Find(&data)
	return
}

func QueryAlarmBySql(sql string, params []interface{}, customQueryParam m.CustomAlarmQueryParam, page *m.PageInfo) (err error, result m.AlarmProblemQueryResult) {
	result = m.AlarmProblemQueryResult{High: 0, Mid: 0, Low: 0, Data: []*m.AlarmProblemQuery{}, Page: &m.PageInfo{}}
	var alarmQuery []*m.AlarmProblemQuery
	var logKeywordConfigList, dbKeywordMonitorList []string
	err = x.SQL(sql, params...).Find(&alarmQuery)
	if len(alarmQuery) > 0 {
		//var logMonitorStrategyIds []string
		for _, v := range alarmQuery {
			v.StartString = v.Start.Format(m.DatetimeFormat)
			v.EndString = v.End.Format(m.DatetimeFormat)
			if v.SMetric == "log_monitor" || v.SMetric == "db_keyword_monitor" {
				if v.SMetric == "log_monitor" {
					logKeywordConfigList = append(logKeywordConfigList, v.AlarmStrategy)
				} else {
					dbKeywordMonitorList = append(dbKeywordMonitorList, v.AlarmStrategy)
				}
				v.IsLogMonitor = true
				v.Log = v.Content
				if v.EndValue > 0 {
					//v.Start, v.End = v.End, v.Start
					if v.EndValue < v.StartValue {
						v.StartValue = v.EndValue
					} else {
						v.StartValue = v.EndValue - v.StartValue + 1
					}
					if strings.Contains(v.Log, "^^") {
						if brIndex := strings.Index(v.Log, "<br/>"); brIndex > 0 {
							v.Content = v.Log[:brIndex+5]
							v.Log = v.Log[brIndex+5:]
						} else {
							v.Content = ""
						}
						v.Log = fmt.Sprintf("%s: %s <br/>%s: %s", v.StartString, v.Log[:strings.Index(v.Log, "^^")], v.EndString, v.Log[strings.Index(v.Log, "^^")+2:])
					}
					//v.StartString = v.EndString
				} else {
					v.StartValue = 1
					if brIndex := strings.Index(v.Log, "<br/>"); brIndex > 0 {
						v.Content = v.Log[:brIndex+5]
						v.Log = v.Log[brIndex+5:]
					} else {
						v.Content = ""
					}
					if strings.HasSuffix(v.Log, "^^") {
						v.Log = v.StartString + ": " + v.Log[:len(v.Log)-2]
					} else {
						v.Log = v.StartString + ": " + v.Log
					}
				}
			}
			if strings.Contains(v.Log, "\n") {
				v.Log = strings.ReplaceAll(v.Log, "\n", "<br/>")
			}
			//if v.SMetric == "log_monitor" || v.SMetric == "db_keyword_monitor" {
			//	v.IsLogMonitor = true
			//	if v.EndValue > 0 {
			//		v.Start, v.End = v.End, v.Start
			//		v.StartValue = v.EndValue - v.StartValue + 1
			//		if strings.Contains(v.Content, "^^") {
			//			v.Content = fmt.Sprintf("%s: %s <br/>%s: %s", v.StartString, v.Content[:strings.Index(v.Content, "^^")], v.EndString, v.Content[strings.Index(v.Content, "^^")+2:])
			//		}
			//		v.StartString = v.EndString
			//	} else {
			//		v.StartValue = 1
			//		if strings.HasSuffix(v.Content, "^^") {
			//			v.Content = v.StartString + ": " + v.Content[:len(v.Content)-2]
			//		}
			//	}
			//}
			//if strings.Contains(v.Content, "\n") {
			//	v.Content = strings.ReplaceAll(v.Content, "\n", "<br/>")
			//}
		}
	}
	if customQueryParam.Enable {
		for _, v := range GetOpenAlarm(customQueryParam) {
			alarmQuery = append(alarmQuery, v)
		}
	}
	var alarmSortQuery m.AlarmProblemList
	alarmSortQuery = alarmQuery
	sort.Sort(alarmSortQuery)
	alarmQuery = alarmSortQuery
	metricMap := make(map[string]int)
	for _, v := range alarmQuery {
		if v.SPriority == "high" {
			result.High += 1
		} else if v.SPriority == "medium" {
			result.Mid += 1
		} else if v.SPriority == "low" {
			result.Low += 1
		}
		tmpMetricLevel := fmt.Sprintf("%s^%s", v.SMetric, v.SPriority)
		if _, b := metricMap[tmpMetricLevel]; b {
			metricMap[tmpMetricLevel] += 1
		} else {
			metricMap[tmpMetricLevel] = 1
		}
	}
	var resultCount m.AlarmProblemCountList
	for k, v := range metricMap {
		tmpSplit := strings.Split(k, "^")
		resultCount = append(resultCount, &m.AlarmProblemCountObj{Name: tmpSplit[0], Type: tmpSplit[1], Value: v, FilterType: "metric"})
	}
	sort.Sort(resultCount)
	result.Count = resultCount
	if page != nil && page.PageSize > 0 {
		si := (page.StartIndex - 1) * page.PageSize
		ei := page.StartIndex*page.PageSize - 1
		var pageResult []*m.AlarmProblemQuery
		for i, v := range alarmQuery {
			if i >= si && i <= ei {
				pageResult = append(pageResult, v)
			}
		}
		result.Data = pageResult
		result.Page.StartIndex = page.StartIndex
		result.Page.PageSize = page.PageSize
		result.Page.TotalRows = len(alarmQuery)
	} else {
		result.Data = alarmQuery
	}
	var alarmStrategyList, endpointList []string
	for _, v := range result.Data {
		if v.AlarmName == "" {
			v.AlarmName = v.Content
		}
		alarmStrategyList = append(alarmStrategyList, v.AlarmStrategy)
		endpointList = append(endpointList, v.Endpoint)
		alarmDetailList := []*m.AlarmDetailData{}
		if strings.HasPrefix(v.EndpointTags, "ac_") {
			alarmDetailList, err = GetAlarmDetailList(v.Id)
			if err != nil {
				return err, result
			}
			for _, alarmDetail := range alarmDetailList {
				v.AlarmMetricList = append(v.AlarmMetricList, alarmDetail.Metric)
			}
		} else {
			alarmDetailList = append(alarmDetailList, &m.AlarmDetailData{Metric: v.SMetric, Cond: v.SCond, Last: v.SLast, Start: v.Start, StartValue: v.StartValue, End: v.End, EndValue: v.EndValue, Tags: v.Tags})
			v.AlarmMetricList = []string{v.SMetric}
		}
		v.AlarmDetail = buildAlarmDetailData(alarmDetailList, "<br/>")
		v.EndpointGuid = v.Endpoint
		if strings.HasPrefix(v.Endpoint, "sg__") {
			v.Endpoint = v.Endpoint[4:]
			if serviceGroupName, b := m.GlobalSGDisplayNameMap[v.Endpoint]; b {
				v.Endpoint = serviceGroupName
			}
		}
	}
	if len(alarmStrategyList) > 0 || len(logKeywordConfigList) > 0 || len(dbKeywordMonitorList) > 0 {
		logKeywordConfigMap, dbKeywordMonitorMap, matchKeywordStrategyErr := getAlarmKeywordServiceGroup(logKeywordConfigList, dbKeywordMonitorList)
		if matchKeywordStrategyErr != nil {
			log.Logger.Error("try to match alarm keyword strategy fail", log.Error(matchKeywordStrategyErr))
		}
		strategyGroupMap, endpointServiceMap, matchErr := matchAlarmGroups(alarmStrategyList, endpointList)
		if matchErr != nil {
			log.Logger.Error("try to match alarm groups fail", log.Error(matchErr))
		} else {
			for _, v := range result.Data {
				var tmpStrategyGroups []*m.AlarmStrategyGroup
				if v.SMetric == "log_monitor" {
					if serviceGroup, ok := logKeywordConfigMap[v.AlarmStrategy]; ok {
						tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: serviceGroup, Type: "serviceGroup"})
					}
				} else if v.SMetric == "db_keyword_monitor" {
					if serviceGroup, ok := dbKeywordMonitorMap[v.AlarmStrategy]; ok {
						tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: serviceGroup, Type: "serviceGroup"})
					}
				} else {
					if strategyRow, ok := strategyGroupMap[v.AlarmStrategy]; ok {
						if strategyRow.ServiceGroup == "" {
							tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: strategyRow.EndpointGroup, Type: "endpointGroup"})
							if endpointServiceList, endpointOk := endpointServiceMap[v.Endpoint]; endpointOk {
								for _, endpointServiceRelRow := range endpointServiceList {
									tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: endpointServiceRelRow.ServiceGroup, Type: "serviceGroup"})
								}
							}
						} else {
							tmpStrategyGroups = append(tmpStrategyGroups, &m.AlarmStrategyGroup{Name: strategyRow.ServiceGroup, Type: "serviceGroup"})
						}
					}
				}
				v.StrategyGroups = tmpStrategyGroups
			}
		}
	}
	return err, result
}

func QueryHistoryAlarm(param m.QueryHistoryAlarmParam) (err error, result m.AlarmProblemQueryResult) {
	result = m.AlarmProblemQueryResult{High: 0, Mid: 0, Low: 0, Data: []*m.AlarmProblemQuery{}}
	startString := time.Unix(param.Start, 0).Format(m.DatetimeFormat)
	endString := time.Unix(param.End, 0).Format(m.DatetimeFormat)
	if startString == "" || endString == "" {
		return fmt.Errorf("param start or end format fail"), result
	}
	var sql, whereSql string
	var params []interface{}
	if len(param.Endpoint) > 0 {
		whereSql += fmt.Sprintf(" AND endpoint in ('" + strings.Join(param.Endpoint, "','") + "') ")
	}
	if len(param.Priority) > 0 {
		whereSql += fmt.Sprintf(" AND s_priority in ('" + strings.Join(param.Priority, "','") + "') ")
	}
	if len(param.Metric) > 0 {
		whereSql += fmt.Sprintf(" AND s_metric in ('" + strings.Join(param.Metric, "','") + "') ")
	}
	if len(param.AlarmName) > 0 {
		alarmNameSql, alarmNameParam := createListParams(param.AlarmName, "")
		whereSql += " AND alarm_name  in (" + alarmNameSql + ")"
		params = append(params, alarmNameParam...)
	}
	//  支持 告警任意搜索,但是只能搜索不关闭的告警
	if strings.TrimSpace(param.Query) != "" {
		whereSql += " and status <> 'closed' and (endpoint like ? or s_metric like ? or alarm_name like ? or content like ? or s_priority like ?)"
		params = append(params, []interface{}{fmt.Sprintf("%%%s%%", param.Query), fmt.Sprintf("%%%s%%", param.Query),
			fmt.Sprintf("%%%s%%", param.Query), fmt.Sprintf("%%%s%%", param.Query), fmt.Sprintf("%%%s%%", param.Query)}...)
	}
	if param.Filter == "all" {
		sql = "SELECT * FROM alarm WHERE (start<='" + endString + "' OR end>='" + startString + "') " + whereSql + " ORDER BY id DESC"
	}
	if param.Filter == "start" {
		sql = "SELECT * FROM alarm WHERE start>='" + startString + "' AND start<'" + endString + "' " + whereSql + " ORDER BY id DESC"
	}
	if param.Filter == "end" {
		sql = "SELECT * FROM alarm WHERE end>='" + startString + "' AND end<'" + endString + "' " + whereSql + " ORDER BY id DESC"
	}
	if param.Page == nil {
		param.Page = &m.PageInfo{}
	}
	customQueryParam := m.CustomAlarmQueryParam{Enable: true, Level: param.Priority, Start: startString, End: endString, Status: "all"}
	if param.Query != "" {
		customQueryParam.Enable = true
		customQueryParam.Query = param.Query
	}
	if len(param.Metric) > 0 {
		for _, s := range param.Metric {
			if s != "custom" {
				customQueryParam.Enable = false
				break
			}
		}
	} else {
		customQueryParam.Enable = true
	}
	err, result = QueryAlarmBySql(sql, params, customQueryParam, param.Page)
	return err, result
}

func NotifyAlarm(alarmObj *m.AlarmHandleObj) {
	if alarmObj.NotifyDelay > 0 {
		var alarmRows []*m.AlarmTable
		abortNotify := false
		if alarmObj.Status == "ok" {
			x.SQL("select * from alarm where id=?", alarmObj.Id).Find(&alarmRows)
			if len(alarmRows) > 0 {
				if (alarmRows[0].End.Unix() - alarmRows[0].Start.Unix()) <= int64(alarmObj.NotifyDelay) {
					log.Logger.Info("Abort recover alarm notify", log.Int("id", alarmObj.Id), log.String("start", alarmRows[0].Start.Format(m.DatetimeFormat)), log.String("end", alarmRows[0].End.Format(m.DatetimeFormat)))
					abortNotify = true
				}
			}
		} else {
			time.Sleep(time.Duration(alarmObj.NotifyDelay) * time.Second)
			x.SQL("select * from alarm where strategy_id=? and endpoint=? and start=?", alarmObj.StrategyId, alarmObj.Endpoint, alarmObj.Start.Format(m.DatetimeFormat)).Find(&alarmRows)
			if len(alarmRows) > 0 {
				if alarmRows[0].Status == "ok" {
					log.Logger.Info("Abort firing alarm notify", log.String("endpoint", alarmObj.Endpoint), log.Int("strategyId", alarmObj.StrategyId), log.String("start", alarmObj.Start.Format(m.DatetimeFormat)))
					abortNotify = true
				}
			}
		}
		if abortNotify {
			return
		}
	}
	if m.CoreUrl != "" {
		notifyErr := NotifyCoreEvent(alarmObj.Endpoint, alarmObj.StrategyId, alarmObj.Id, 0)
		if notifyErr != nil {
			log.Logger.Error("notify core event fail", log.Error(notifyErr))
		}
	} else {
		if m.Config().Alert.Enable {
			var sao m.SendAlertObj
			accept := GetMailByStrategy(alarmObj.StrategyId)
			if len(accept) == 0 {
				return
			}
			timeString := alarmObj.Start.Format(m.DatetimeFormat)
			if alarmObj.Status == "ok" {
				timeString = alarmObj.End.Format(m.DatetimeFormat)
			}
			sao.Accept = accept
			sao.Subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", alarmObj.Status, alarmObj.SPriority, alarmObj.Endpoint, alarmObj.SMetric)
			sao.Content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s", alarmObj.Endpoint, alarmObj.Status, alarmObj.SMetric, alarmObj.StartValue, alarmObj.SCond, alarmObj.SLast, alarmObj.SPriority, alarmObj.Content, timeString)
			other.SendSmtpMail(sao)
		}
	}

}

func NotifyTreevent(param m.EventTreeventNotifyDto) {
	if m.CoreUrl == "" || len(param.Data) == 0 {
		return
	}
	postBytes, _ := json.Marshal(param)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/treevent/api/v1/event/send", m.CoreUrl), strings.NewReader(string(postBytes)))
	request.Header.Set("Authorization", m.GetCoreToken())
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Logger.Error("Notify treevent event new request fail", log.Error(err))
		return
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Logger.Error("Notify treevent event ctxhttp request fail", log.Error(err))
		return
	}
	resultBody, _ := ioutil.ReadAll(res.Body)
	var responseData m.EventTreeventResponse
	err = json.Unmarshal(resultBody, &responseData)
	res.Body.Close()
	if err != nil {
		log.Logger.Error("Notify treevent event unmarshal json body fail", log.Error(err))
		return
	}
	if responseData.Code > 0 {
		log.Logger.Error("Notify treevent fail", log.String("message", responseData.Msg))
	} else {
		log.Logger.Info("Notify treevent success", log.Int("event length", len(param.Data)))
	}
}

func StartInitAlarmUniqueTags() {
	var alarmRows []*m.AlarmTable
	err := x.SQL("select * from alarm where status='firing'").Find(&alarmRows)
	if err != nil {
		log.Logger.Error("init alarm unique tags fail,query alarm table error", log.Error(err))
		return
	}
	for _, row := range alarmRows {
		if row.EndpointTags != "" {
			continue
		}
		calcAlarmUniqueFlag(row)
		x.Exec("update alarm set endpoint_tags=? where id=?", row.EndpointTags, row.Id)
	}
}

func ManualNotifyAlarm(alarmId int, operator string) (err error) {
	var alarmRows []*m.AlarmTable
	err = x.SQL("select * from alarm where id=?", alarmId).Find(&alarmRows)
	if err != nil {
		err = fmt.Errorf("query alarm table with id:%d error:%s ", alarmId, err.Error())
		return
	}
	if len(alarmRows) == 0 {
		err = fmt.Errorf("can not find alarm with id:%d ", alarmId)
		return
	}
	alarmObj := m.AlarmHandleObj{AlarmTable: *alarmRows[0]}
	if alarmObj.NotifyId == "" {
		err = fmt.Errorf("alarm notify id is empty")
		return
	}
	var notifyRows []*m.NotifyTable
	err = x.SQL("select * from notify where guid=?", alarmObj.NotifyId).Find(&notifyRows)
	if err != nil {
		err = fmt.Errorf("query alarm table with id:%d error:%s ", alarmId, err.Error())
		return
	}
	if len(notifyRows) == 0 {
		err = fmt.Errorf("can not find notify with guid:%s ", alarmObj.NotifyId)
		return
	}
	notifyObj := notifyRows[0]
	if _, err = notifyEventAction(notifyObj, &alarmObj, false, operator); err != nil {
		err = fmt.Errorf("notify event action fail:%s ", err.Error())
	} else {
		_, err = x.Exec("insert into alarm_notify(alarm_id,notify_id,endpoint,metric,status,proc_def_key,proc_def_name,notify_description,created_user,created_time) values (?,?,?,?,?,?,?,?,?,?)",
			alarmObj.Id, notifyObj.Guid, alarmObj.Endpoint, alarmObj.SMetric, "created", notifyObj.ProcCallbackKey, notifyObj.ProcCallbackName, notifyObj.Description, operator, time.Now())
		if err != nil {
			err = fmt.Errorf("notify event write db alarm notify record fail,%s ", err.Error())
		}
	}
	return
}

func UpdateAlarmWithConditions(alarmConditionObj *m.AlarmHandleObj) (alarmRow *m.AlarmHandleObj, err error) {
	var actions []*Action
	var strategyMetricRows []*m.AlarmStrategyMetric
	err = x.SQL("select guid,alarm_strategy,metric,`condition`,`last`,crc_hash from alarm_strategy_metric where alarm_strategy=?", alarmConditionObj.AlarmStrategy).Find(&strategyMetricRows)
	if err != nil {
		err = fmt.Errorf("query alarm strategy metric table fail,%s ", err.Error())
		return
	}
	if len(strategyMetricRows) == 0 {
		err = fmt.Errorf("alarm strategy:%s condition metric query with empty data", alarmConditionObj.AlarmStrategy)
		return
	}
	var configCrcList, conditionGuidList, conditionMetricList []string
	var crcIndex int
	for i, row := range strategyMetricRows {
		configCrcList = append(configCrcList, row.CrcHash)
		if row.CrcHash == alarmConditionObj.AlarmConditionCrcHash {
			crcIndex = i
		}
	}
	if crcIndex > 0 {
		time.Sleep(time.Duration(crcIndex) * time.Second)
	}
	alarmCrcMap := make(map[string]int)
	alarmCrcMap[alarmConditionObj.AlarmConditionCrcHash] = 1
	var alarmConditionRows []*m.AlarmCondition
	err = x.SQL("select guid,status,metric,crc_hash,tags from alarm_condition where crc_hash in ('"+strings.Join(configCrcList, "','")+"') and alarm_strategy=? and endpoint=? and status='firing'", alarmConditionObj.AlarmStrategy, alarmConditionObj.Endpoint).Find(&alarmConditionRows)
	if err != nil {
		err = fmt.Errorf("query alarm condition table fail,%s ", err.Error())
		return
	}
	for _, row := range alarmConditionRows {
		if row.CrcHash == alarmConditionObj.AlarmConditionCrcHash {
			// 相同crc策略
			if row.Tags != alarmConditionObj.Tags {
				// 不同标签
				err = fmt.Errorf("same crc alarm:%s is firing,ignore this:%s ", row.Tags, alarmConditionObj.Tags)
				return
			}
		}
		alarmCrcMap[row.CrcHash] = 1
		conditionGuidList = append(conditionGuidList, row.Guid)
		conditionMetricList = append(conditionMetricList, row.Metric)
	}
	if alarmConditionObj.AlarmConditionGuid != "" {
		var alarmConditionRelRows []*m.AlarmConditionRel
		err = x.SQL("select alarm,alarm_condition from alarm_condition_rel where alarm_condition=? and alarm in (select id from alarm where status='firing' and alarm_strategy=?)", alarmConditionObj.AlarmConditionGuid, alarmConditionObj.AlarmStrategy).Find(&alarmConditionRelRows)
		if err != nil {
			err = fmt.Errorf("query alarm condition ref fail,%s ", err.Error())
			return
		}
		actions = append(actions, &Action{Sql: "UPDATE alarm_condition SET status=?,end_value=?,end=? WHERE guid=? AND status='firing'", Param: []interface{}{
			alarmConditionObj.Status, alarmConditionObj.EndValue, alarmConditionObj.End.Format(m.DatetimeFormat), alarmConditionObj.AlarmConditionGuid,
		}})
		if len(alarmConditionRelRows) > 0 {
			// 有要一起关闭的alarm
			for _, alarmConditionRef := range alarmConditionRelRows {
				actions = append(actions, &Action{Sql: "UPDATE alarm SET status=?,end_value=?,end=? WHERE id=? AND status='firing'", Param: []interface{}{
					alarmConditionObj.Status, alarmConditionObj.EndValue, alarmConditionObj.End.Format(m.DatetimeFormat), alarmConditionRef.Alarm,
				}})
			}
		}
		err = Transaction(actions)
	} else {
		alarmConditionObj.AlarmConditionGuid = "ac_" + guid.CreateGuid()
		conditionGuidList = append(conditionGuidList, alarmConditionObj.AlarmConditionGuid)
		conditionMetricList = append(conditionMetricList, alarmConditionObj.SMetric)
		sort.Strings(conditionGuidList)
		sort.Strings(conditionMetricList)
		session := x.NewSession()
		session.Begin()
		defer func() {
			if err != nil {
				session.Rollback()
			} else {
				session.Commit()
			}
			session.Close()
		}()
		_, err = session.Exec("INSERT INTO alarm_condition(guid,alarm_strategy,endpoint,status,metric,expr,cond,`last`,priority,crc_hash,tags,start_value,`start`,unique_hash) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
			alarmConditionObj.AlarmConditionGuid, alarmConditionObj.AlarmStrategy, alarmConditionObj.Endpoint, alarmConditionObj.Status, alarmConditionObj.SMetric, alarmConditionObj.SExpr, alarmConditionObj.SCond, alarmConditionObj.SLast, alarmConditionObj.SPriority, alarmConditionObj.AlarmConditionCrcHash, alarmConditionObj.Tags, alarmConditionObj.StartValue, time.Now().Format(m.DatetimeFormat), alarmConditionObj.EndpointTags)
		if err != nil {
			err = fmt.Errorf("insert alarm_condition fail,%s ", err.Error())
			return
		}
		log.Logger.Debug("UpdateAlarmWithConditions", log.JsonObj("alarmCrcMap", alarmCrcMap), log.StringList("configCrcList", configCrcList))
		if len(alarmCrcMap) >= len(configCrcList) {
			// 如果条件都满足
			alarmStrategyObj, getStrategyErr := GetSimpleAlarmStrategy(alarmConditionObj.AlarmStrategy)
			if getStrategyErr != nil {
				err = getStrategyErr
				return
			}
			alarmRow = &m.AlarmHandleObj{}
			alarmRow.Endpoint = alarmConditionObj.Endpoint
			alarmRow.Status = alarmConditionObj.Status
			alarmRow.SMetric = strings.Join(conditionMetricList, ",")
			alarmRow.Tags = strings.Join(conditionMetricList, ",")
			alarmRow.SCond = strategyMetricRows[0].Condition
			alarmRow.SLast = strategyMetricRows[0].Last
			alarmRow.SPriority = alarmStrategyObj.Priority
			alarmRow.Content = alarmStrategyObj.Content
			alarmRow.StartValue = alarmConditionObj.StartValue
			alarmRow.Start = alarmConditionObj.Start
			alarmRow.AlarmStrategy = alarmConditionObj.AlarmStrategy
			alarmRow.AlarmName = alarmStrategyObj.Name
			alarmRow.EndpointTags = strings.Join(conditionGuidList, ",")
			alarmRow.EndpointTags = fmt.Sprintf("ac_%x", sha256.Sum256([]byte(strings.Join(conditionGuidList, ","))))
			alarmRow.NotifyEnable = alarmConditionObj.NotifyEnable
			alarmRow.NotifyId = alarmConditionObj.NotifyId
			alarmRow.NotifyDelay = alarmConditionObj.NotifyDelay
			insertAlarmResult, insertErr := session.Exec("INSERT INTO alarm (endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,tags,start_value,`start`,endpoint_tags,alarm_strategy,alarm_name) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
				alarmRow.Endpoint, alarmRow.Status, alarmRow.SMetric, alarmRow.SExpr, alarmRow.SCond, alarmRow.SLast, alarmRow.SPriority, alarmRow.Content, alarmRow.Tags, alarmRow.StartValue, alarmRow.Start, alarmRow.EndpointTags, alarmRow.AlarmStrategy, alarmRow.AlarmName)
			if insertErr != nil {
				err = fmt.Errorf("insert alarm fail,%s ", insertErr.Error())
				return
			}
			alarmInsertId, _ := insertAlarmResult.LastInsertId()
			if alarmInsertId <= 0 {
				err = fmt.Errorf("insert alarm fail with get last insert id fail")
				return
			}
			alarmRow.Id = int(alarmInsertId)
			for _, conditionGuid := range conditionGuidList {
				_, err = session.Exec("insert into alarm_condition_rel(alarm,alarm_condition) values (?,?)", alarmInsertId, conditionGuid)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func GetAlarmDetailList(alarmId int) (alarmDetailList []*m.AlarmDetailData, err error) {
	alarmDetailList = []*m.AlarmDetailData{}
	//filterSql, filterParam := createListParams(alarmConditionGuidList, "")
	err = x.SQL("select t1.metric,t1.cond,t1.`last`,t1.`start`,t1.start_value,t1.`end`,t1.end_value,t1.tags,t2.metric as 'metric_name' from alarm_condition t1 left join metric t2 on t1.metric=t2.guid where t1.guid in (select alarm_condition from alarm_condition_rel where alarm=?)", alarmId).Find(&alarmDetailList)
	if err != nil {
		err = fmt.Errorf("GetAlarmDetailList -> query alarm condition table fail,%s ", err.Error())
		return
	}
	return
}

func buildAlarmDetailData(inputList []*m.AlarmDetailData, splitChar string) string {
	stringList := []string{}
	for _, v := range inputList {
		if v != nil {
			tagList := []string{}
			for _, tagV := range strings.Split(v.Tags, "^") {
				if strings.HasPrefix(tagV, "e_guid:") || strings.HasPrefix(tagV, "guid:") || strings.HasPrefix(tagV, "agg:") || strings.HasPrefix(tagV, "key:") || strings.HasPrefix(tagV, "condition_crc:") {
					continue
				}
				if firstSplitIndex := strings.Index(tagV, ":"); firstSplitIndex > 0 {
					tagV = tagV[:firstSplitIndex] + "=" + tagV[firstSplitIndex+1:]
					tagList = append(tagList, tagV)
				}
			}
			if len(tagList) > 0 {
				stringList = append(stringList, fmt.Sprintf("Metric:%s Tag:[%s] %s Value:%.3f Duration:%s", v.Metric, strings.Join(tagList, ","), v.Cond, v.StartValue, v.Last))
			} else {
				stringList = append(stringList, fmt.Sprintf("Metric:%s %s Value:%.3f Duration:%s", v.Metric, v.Cond, v.StartValue, v.Last))
			}
		}
	}
	return strings.Join(stringList, splitChar)
}

func matchAlarmGroups(alarmStrategyList, endpointList []string) (strategyGroupMap map[string]*m.StrategyGroupRow, endpointServiceMap map[string][]*m.EndpointServiceRelTable, err error) {
	strategyGroupMap = make(map[string]*m.StrategyGroupRow)
	endpointServiceMap = make(map[string][]*m.EndpointServiceRelTable)
	if len(alarmStrategyList) == 0 {
		return
	}
	var strategyGroupRows []*m.StrategyGroupRow
	strategyFilter, strategyParams := createListParams(alarmStrategyList, "")
	err = x.SQL("select t1.guid,t1.endpoint_group,t3.display_name as service_group from alarm_strategy t1 left join endpoint_group t2 on t1.endpoint_group=t2.guid left join service_group t3 on t2.service_group=t3.guid where t1.guid in ("+strategyFilter+")", strategyParams...).Find(&strategyGroupRows)
	if err != nil {
		err = fmt.Errorf("matchAlarmGroups -> query alarm strategy table fail,%s ", err.Error())
		return
	}
	endpointServiceRelRows := []*m.EndpointServiceRelTable{}
	if len(endpointList) > 0 {
		endpointFilter, endpointParams := createListParams(endpointList, "")
		err = x.SQL("select endpoint,t2.display_name as service_group from endpoint_service_rel t1 left join service_group t2 on t1.service_group=t2.guid where t1.endpoint in ("+endpointFilter+")", endpointParams...).Find(&endpointServiceRelRows)
		if err != nil {
			err = fmt.Errorf("matchAlarmGroups -> query alarm endpoint service rel table fail,%s ", err.Error())
			return
		}
	}
	for _, row := range strategyGroupRows {
		strategyGroupMap[row.Guid] = row
	}
	for _, row := range endpointServiceRelRows {
		if existList, ok := endpointServiceMap[row.Endpoint]; ok {
			endpointServiceMap[row.Endpoint] = append(existList, row)
		} else {
			endpointServiceMap[row.Endpoint] = []*m.EndpointServiceRelTable{row}
		}
	}
	return
}

func GetAlarmNameList(status, alarmName string) (list []string, err error) {
	var alarmList, customAlarmList []*m.SimpleAlarm
	var closed = 1
	if status == "firing" {
		closed = 0
	}
	alarmSQL := "select alarm_name as alarm_name,start as update_at from alarm  where 1=1"
	var params []interface{}
	if status != "" {
		alarmSQL = alarmSQL + " and status=?"
		params = append(params, status)
	}
	if strings.TrimSpace(alarmName) != "" {
		alarmSQL = alarmSQL + " and alarm_name like '%" + alarmName + "%'"
	}
	alarmSQL = alarmSQL + " order by start desc limit 100"
	if err = x.SQL(alarmSQL, params...).Find(&alarmList); err != nil {
		return
	}

	var customAlarmParams []interface{}
	customAlarmSQL := "select alert_title as alarm_name,update_at from alarm_custom where 1=1"
	if status != "" {
		customAlarmSQL = customAlarmSQL + " and closed=?"
		customAlarmParams = append(customAlarmParams, closed)
	}
	if strings.TrimSpace(alarmName) != "" {
		customAlarmSQL = customAlarmSQL + " and alert_title like '%" + alarmName + "%'"
	}
	customAlarmSQL = customAlarmSQL + "  order by update_at desc limit 100"

	if err = x.SQL(customAlarmSQL, customAlarmParams...).Find(&customAlarmList); err != nil {
		return
	}

	alarmList = append(alarmList, customAlarmList...)
	// 数据去重复
	alarmList = filterRepeatAndEmptyNameAlarm(alarmList)
	// 按最新时间在前面
	sort.Sort(m.SimpleAlarmSort(alarmList))

	for _, alarm := range alarmList {
		if strings.TrimSpace(alarm.AlarmName) == "" {
			continue
		}
		list = append(list, alarm.AlarmName)
		// 最长20条数据
		if len(list) >= 20 {
			break
		}
	}
	return
}

func convertString2Map(list []string) map[string]bool {
	var hashMap = make(map[string]bool)
	if len(list) == 0 {
		return hashMap
	}
	for _, s := range list {
		hashMap[s] = true
	}
	return hashMap
}

func filterRepeatAndEmptyNameAlarm(list []*m.SimpleAlarm) []*m.SimpleAlarm {
	var newList []*m.SimpleAlarm
	hashMap := make(map[string]*m.SimpleAlarm)
	for _, alarm := range list {
		if alarm.AlarmName == "" {
			continue
		}
		if _, ok := hashMap[alarm.AlarmName]; ok {
			continue
		}
		hashMap[alarm.AlarmName] = alarm
	}
	for _, value := range hashMap {
		newList = append(newList, value)
	}
	return newList
}

func getLevelSQL(levelMap map[string]bool) string {
	var levelFilterSql string
	switch len(levelMap) {
	case 1:
		if levelMap["high"] {
			levelFilterSql = " AND alert_level in (1,2) "
		} else if levelMap["medium"] {
			levelFilterSql = " AND alert_level in (3,4) "
		} else {
			levelFilterSql = " AND alert_level>=5 "
		}
	case 2:
		if levelMap["high"] && levelMap["medium"] {
			levelFilterSql = " AND alert_level<5 "
		} else if levelMap["high"] && levelMap["low"] {
			levelFilterSql = " ( AND alert_level in (1,2) or alert_level>=5) "
		} else {
			levelFilterSql = " AND alert_level>=3 "
		}
	default:
	}
	return levelFilterSql
}

// 校验是否有编排使用权限
func checkHasProcDefUsePermission(alarmNotify *m.AlarmNotifyTable, hasRoleMap map[string]bool, token string) (result bool) {
	var name = alarmNotify.ProcDefName
	var version string
	var resByteArr []byte
	var response m.QueryProcessDefinitionPublicResponse
	var err error
	if strings.TrimSpace(alarmNotify.ProcDefName) != "" {
		index := strings.LastIndex(alarmNotify.ProcDefName, "[")
		if index < 0 {
			return
		}
		name = alarmNotify.ProcDefName[:index]
		version = alarmNotify.ProcDefName[index+1 : len(alarmNotify.ProcDefName)-1]
		if resByteArr, err = HttpGet(m.CoreUrl+"/platform/v1/public/process/definitions/detail?name="+name+"&version="+version, token); err != nil {
			log.Logger.Error("checkHasProcDefUsePermission HttpPost err", log.Error(err))
			return
		}
		if err = json.Unmarshal(resByteArr, &response); err != nil {
			log.Logger.Error("checkHasProcDefUsePermission Unmarshal err", log.Error(err))
			return
		}
		if response.Status != "OK" {
			err = fmt.Errorf(response.Message)
			log.Logger.Error("checkHasProcDefUsePermission response err", log.Error(err))
			return
		}
		if response.Data != nil && len(response.Data.UseRoles) > 0 {
			for _, role := range response.Data.UseRoles {
				if hasRoleMap[role] {
					return true
				}
			}
		}
	}
	return
}

func getAlarmKeywordServiceGroup(logKeywordConfigList, dbKeywordMonitorList []string) (logKeywordConfigMap, dbKeywordMonitorMap map[string]string, err error) {
	logKeywordConfigMap = make(map[string]string)
	dbKeywordMonitorMap = make(map[string]string)
	if len(logKeywordConfigList) > 0 {
		var logKeywordRows []*m.DbKeywordMonitor
		err = x.SQL("select t1.guid,t3.display_name as `service_group` from log_keyword_config t1 left join log_keyword_monitor t2 on t1.log_keyword_monitor=t2.guid left join service_group t3 on t2.service_group=t3.guid where t1.guid in ('" + strings.Join(logKeywordConfigList, "','") + "')").Find(&logKeywordRows)
		if err != nil {
			return
		}
		for _, row := range logKeywordRows {
			logKeywordConfigMap[row.Guid] = row.ServiceGroup
		}
	}
	if len(dbKeywordMonitorList) > 0 {
		var dbKeywordRows []*m.DbKeywordMonitor
		err = x.SQL("select t1.guid,t2.display_name as `service_group` from db_keyword_monitor t1 left join service_group t2 on t1.service_group=t2.guid where t1.guid in ('" + strings.Join(dbKeywordMonitorList, "','") + "')").Find(&dbKeywordRows)
		if err != nil {
			return
		}
		for _, row := range dbKeywordRows {
			dbKeywordMonitorMap[row.Guid] = row.ServiceGroup
		}
	}
	return
}

// HttpGet  Get请求
func HttpGet(url, userToken string) (byteArr []byte, err error) {
	req, newReqErr := http.NewRequest(http.MethodGet, url, strings.NewReader(""))
	if newReqErr != nil {
		err = fmt.Errorf("try to new http request fail,%s ", newReqErr.Error())
		return
	}
	req.Header.Set("Authorization", userToken)
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = fmt.Errorf("try to do http request fail,%s ", respErr.Error())
		return
	}
	byteArr, _ = io.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}
