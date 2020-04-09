package db

import (
	"strconv"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"time"
	"strings"
	"sort"
)

func ListGrp(query *m.GrpQuery) error {
	var querySql = `SELECT id,name,description FROM grp WHERE 1=1 `
	var countSql = `SELECT count(1) num FROM grp WHERE 1=1 `
	var whereSql string
	qParams := make([]interface{}, 0)
	if query.Id > 0 {
		whereSql += ` AND id=? `
		qParams = append(qParams, query.Id)
	}
	if query.Search != "" {
		if query.Search == "." {
			query.Search = ""
		}
		whereSql += ` AND (name like '%`+query.Search+`%' or description like '%`+query.Search+`%') `
	}
	if query.Name != "" {
		whereSql += ` AND name=? `
		qParams = append(qParams, query.Name)
	}
	querySql += whereSql
	countSql += whereSql
	cParams := qParams
	if query.Size > 0 && query.Page > 0 {
		querySql += ` ORDER BY create_at DESC limit ?,?`
		qParams = append(qParams, (query.Page-1)*query.Size)
		qParams = append(qParams, query.Size)
	}
	var result []*m.GrpTable
	var count []int
	err := x.SQL(querySql, qParams...).Find(&result)
	err = x.SQL(countSql,cParams...).Find(&count)
	if len(result) > 0 {
		query.Result = result
		query.ResultNum = count[0]
	}else{
		query.Result = []*m.GrpTable{}
		query.ResultNum = 0
	}
	return err
}

func GetSingleGrp(id int,name string) (error,m.GrpTable) {
	var result []*m.GrpTable
	err := x.SQL("SELECT * FROM grp WHERE id=? or name=?", id, name).Find(&result)
	if err != nil {
		mid.LogError("get single grp fail", err)
		return err,m.GrpTable{}
	}
	if len(result) == 0 {
		return nil,m.GrpTable{}
	}
	return nil,*result[0]
}

func SearchGrp(search string) (error,[]*m.OptionModel) {
	var result []*m.OptionModel
	var grps []*m.GrpTable
	if search == "." {
		search = ""
	}
	err := x.SQL(`SELECT * FROM grp WHERE name LIKE '%`+search+`%'`).Find(&grps)
	if err != nil {
		mid.LogError("search grp fail", err)
		return err,result
	}
	for _,v := range grps {
		result = append(result, &m.OptionModel{OptionValue:fmt.Sprintf("%d", v.Id), OptionText:v.Name, Id:v.Id})
	}
	return nil,result
}

func ListAlarmEndpoints(query *m.AlarmEndpointQuery) error {
	whereSql := ""
	if query.Search != "" {
		whereSql += ` AND t1.guid LIKE '%`+query.Search+`%' `
	}
	if query.Grp > 0 {
		whereSql += fmt.Sprintf(" AND t3.id=%d ", query.Grp)
	}
	querySql := `SELECT t5.* FROM (
            SELECT t4.id,t4.guid,GROUP_CONCAT(t4.name) groups_name,t4.type FROM (
			SELECT t1.id,t1.guid,t3.name,t1.export_type as type FROM endpoint t1 
			LEFT JOIN grp_endpoint t2 ON t1.id=t2.endpoint_id 
			LEFT JOIN grp t3 ON t2.grp_id=t3.id 
			WHERE 1=1 `+whereSql+`
			) t4 GROUP BY t4.guid
			) t5 ORDER BY t5.guid LIMIT ?,?`
	countSql := `SELECT COUNT(1) num FROM (
			SELECT t4.guid,GROUP_CONCAT(t4.name) groups_name FROM (
			SELECT t1.guid,t3.name FROM endpoint t1 
			LEFT JOIN grp_endpoint t2 ON t1.id=t2.endpoint_id 
			LEFT JOIN grp t3 ON t2.grp_id=t3.id
			WHERE 1=1 `+whereSql+`
			) t4 GROUP BY t4.guid
			) t5`
	var result []*m.AlarmEndpointObj
	var count []int
	err := x.SQL(querySql, (query.Page-1)*query.Size, query.Size).Find(&result)
	err = x.SQL(countSql).Find(&count)
	if len(result) > 0 {
		for _,v := range result {
			if v.GroupsName != "" {
				v.GroupsName = v.GroupsName[:len(v.GroupsName)-1]
			}
		}
		query.Result = result
		query.ResultNum = count[0]
	}else{
		query.Result = []*m.AlarmEndpointObj{}
		query.ResultNum = 0
	}
	return err
}

func UpdateGrp(obj *m.UpdateGrp) error {
	var actions []*Action
	for _,grp := range obj.Groups {
		grp.UpdateUser = obj.OperateUser
		if obj.Operation == "insert" {
			grp.CreateUser = obj.OperateUser
			grp.CreateAt = time.Now()
			grp.UpdateAt = time.Now()
		}
		action := Classify(*grp, obj.Operation, "grp", true)
		mid.LogInfo(fmt.Sprintf("action: sql-> %s params->%v", action.Sql, action.Param))
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	err := Transaction(actions)
	return err
}

func UpdateGrpEndpoint(param m.GrpEndpointParamNew) (error,bool) {
	if len(param.Endpoints) == 0 {
		return nil,false
	}
	var ids string
	for _,v := range param.Endpoints {
		ids += fmt.Sprintf("%d,", v)
	}
	if param.Operation == "add" {
		var grpEndpoints []*m.GrpEndpointTable
		err := x.SQL(fmt.Sprintf("SELECT * FROM grp_endpoint WHERE grp_id=%d AND endpoint_id IN (%s)", param.Grp, ids[:len(ids)-1])).Find(&grpEndpoints)
		if err != nil {
			return err,false
		}
		var needAdd = true
		var needInsert = false
		insertSql := "INSERT INTO grp_endpoint VALUES "
		for _,v := range param.Endpoints {
			needAdd = true
			for _,vv := range grpEndpoints {
				if v == vv.EndpointId {}
				needAdd = false
				break
			}
			if needAdd {
				insertSql += fmt.Sprintf("(%d,%d),", param.Grp, v)
				needInsert = true
			}
		}
		if needInsert {
			_, err = x.Exec(insertSql[:len(insertSql)-1])
			return err,needInsert
		}else{
			return nil,needInsert
		}
	}
	if param.Operation == "delete" {
		_,err := x.Exec(fmt.Sprintf("DELETE FROM grp_endpoint WHERE grp_id=%d AND endpoint_id IN (%s)", param.Grp, ids[:len(ids)-1]))
		return err,true
	}
	return fmt.Errorf("operation is not add or delete"),false
}



func GetStrategy(param m.StrategyTable) (error,m.StrategyTable) {
	var result []*m.StrategyTable
	var err error
	if param.Id > 0 {
		err = x.SQL("SELECT * FROM strategy WHERE id=?", param.Id).Find(&result)
	}else if param.Expr != "" {
		err = x.SQL("SELECT * FROM strategy WHERE expr=? order by id desc", param.Expr).Find(&result)
	}
	if err == nil && len(result) == 0 {
		err = fmt.Errorf("no data")
	}
	if err != nil {
		return err,m.StrategyTable{}
	}
	return err,*result[0]
}

func getGrpParent(grpId int) m.GrpTable {
	var grp []*m.GrpTable
	x.SQL("SELECT id,name,parent FROM grp WHERE id=?", grpId).Find(&grp)
	if len(grp) > 0 {
		return *grp[0]
	}
	return m.GrpTable{}
}

func GetStrategys(query *m.TplQuery, ignoreLogMonitor bool) error {
	var result []*m.TplObj
	//ignoreLogMonitorSql := "and t2.metric<>'log_monitor'"
	//if !ignoreLogMonitor {
	//	ignoreLogMonitorSql = ""
	//}
	if query.SearchType == "endpoint" {
		var grps []*m.GrpTable
		err := x.SQL("SELECT * FROM grp where id in (select grp_id from grp_endpoint WHERE endpoint_id=?)", query.SearchId).Find(&grps)
		if err != nil {
			mid.LogError("get strategy fail", err)
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
					for i:=0;i<10;i++ {
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
						}else{
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
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE (`+grpIds+` endpoint_id=?)  order by t1.endpoint_id,t1.id,t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			mid.LogError("get strategy fail", err)
			return err
		}
		if len(tpls) == 0 {
			endpointObj := m.EndpointTable{Id:query.SearchId}
			GetEndpoint(&endpointObj)
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}})
		}else {
			var tmpTplId int
			var tmpStrategys []*m.StrategyTable
			for i, v := range tpls {
				if ignoreLogMonitor && v.Metric == "log_monitor" {
					continue
				}
				if i == 0 {
					tmpTplId = v.TplId
					if v.StrategyId > 0 {
						tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content})
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
						tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content})
					}
				}
			}
			if tpls[len(tpls)-1].EndpointId > 0 {
				endpointObj := m.EndpointTable{Id: tpls[len(tpls)-1].EndpointId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: tpls[len(tpls)-1].TplId, ObjId: tpls[len(tpls)-1].EndpointId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: tmpStrategys})
			}else{
				result = append(result, &m.TplObj{TplId: tpls[len(tpls)-1].TplId, ObjId:tpls[len(tpls)-1].GrpId, ObjName:grpMap[tpls[len(tpls)-1].GrpId], ObjType:"grp", Operation:false, Strategy:tmpStrategys})
				endpointObj := m.EndpointTable{Id:query.SearchId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}})
			}
		}
	}else{
		var grps []*m.GrpTable
		err := x.SQL("SELECT * FROM grp WHERE id=?", query.SearchId).Find(&grps)
		if err != nil {
			mid.LogError("get grp fail", err)
			return err
		}
		if len(grps) <= 0 {
			mid.LogInfo("can't find this grp")
			return fmt.Errorf("can't find this grp")
		}
		var parentTpls []*m.TplStrategyTable
		var tpls []*m.TplStrategyTable
		if grps[0].Parent > 0 {
			tmpParentId := grps[0].Parent
			for i:=0;i<10;i++ {
				parentGrp := getGrpParent(tmpParentId)
				sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content 
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
							tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content})
						}
					}
					result = append(result, &m.TplObj{TplId:parentTpls[0].TplId, ObjId:parentGrp.Id, ObjName:parentGrp.Name, ObjType:"grp", Operation:false, Strategy:tmpStrategys})
				}else{
					result = append(result, &m.TplObj{TplId:0, ObjId:parentGrp.Id, ObjName:parentGrp.Name, ObjType:"grp", Operation:false, Strategy:[]*m.StrategyTable{}})
				}
				if parentGrp.Parent <= 0 {
					break
				}else{
					tmpParentId = parentGrp.Parent
				}
			}
			var newResult []*m.TplObj
			var tmpParentName,tmpObjName string
			for i:=len(result);i>0;i-- {
				tmpObjName = result[i-1].ObjName
				if tmpParentName != "" {
					result[i-1].ObjName = fmt.Sprintf("%s [%s]", tmpObjName, tmpParentName)
				}
				tmpParentName = tmpObjName
				newResult = append(newResult, result[i-1])
			}
			result = newResult
		}
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE t1.grp_id=?  ORDER BY t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			mid.LogError("get strategy fail", err)
			return err
		}
		if len(tpls) > 0 {
			tmpStrategys := []*m.StrategyTable{}
			for _, v := range tpls {
				if v.StrategyId > 0 {
					if ignoreLogMonitor && v.Metric == "log_monitor" {
						continue
					}
					tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content})
				}
			}
			result = append(result, &m.TplObj{TplId:tpls[0].TplId, ObjId:tpls[0].GrpId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, Strategy:tmpStrategys})
		}else{
			result = append(result, &m.TplObj{TplId:0, ObjId:query.SearchId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, Strategy:[]*m.StrategyTable{}})
		}
	}
	for i,v := range result {
		result[i].Accept = getActionOptions(v.TplId)
	}
	query.Tpl = result
	return nil
}

func UpdateStrategy(obj *m.UpdateStrategy) error {
	var actions []*Action
	for _,v := range obj.Strategy {
		action := Classify(*v, obj.Operation, "strategy", false)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	err := Transaction(actions)
	return err
}

func GetTpl(tplId,grpId,endpointId int) (error,m.TplTable) {
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
	err := x.SQL(sql,param...).Find(&result)
	if err != nil || len(result) <= 0 {
		return err,m.TplTable{}
	}
	return nil,*result[0]
}

func ListTpl() []*m.TplTable {
	var result []*m.TplTable
	x.SQL("SELECT * FROM tpl").Find(&result)
	return result
}

func GetParentTpl(tplId int) []int {
	type tplGrpParent struct {
		Id  int
		GrpId  int
		Parent  int
	}
	var result []*tplGrpParent
	x.SQL("SELECT t1.id,t1.grp_id,t2.parent FROM tpl t1 LEFT JOIN grp t2 ON t1.grp_id=t2.id").Find(&result)
	if len(result) == 0 {
		return []int{}
	}
	tmpGrptId := 0
	for _,v := range result {
		if v.Id == tplId {
			tmpGrptId = v.GrpId
			break
		}
	}
	var tplIdList []int
	tmpGrpMap := make(map[int]int)
	for i:=0;i<10;i++ {
		endFlag := true
		for _,v := range result {
			for kk,vv := range tmpGrpMap {
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
				if _,b := tmpGrpMap[v.GrpId];!b {
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

func AddTpl(grpId,endpointId int, operateUser string) (error,m.TplTable) {
	_,tpl := GetTpl(0, grpId, endpointId)
	if tpl.Id > 0 {
		return nil,tpl
	}
	insertSql := fmt.Sprintf("INSERT INTO tpl(grp_id,endpoint_id,create_user,update_user,create_at,update_at) VALUE (%d,%d,'%s','%s',NOW(),NOW())", grpId, endpointId, operateUser, operateUser)
	_,err := x.Exec(insertSql)
	if err != nil {
		mid.LogError("add tpl fail", err)
		return err,tpl
	}
	_,tpl = GetTpl(0, grpId, endpointId)
	if tpl.Id <= 0 {
		err = fmt.Errorf("cat't find the new one")
		mid.LogError("add tpl fail", err)
		return err,tpl
	}
	return nil,tpl
}

func UpdateTpl(tplId int, operateUser string) error {
	_,err := x.Exec("UPDATE tpl SET update_user=?,update_at=NOW() WHERE id=?", operateUser, tplId)
	if err != nil {
		mid.LogError("update tpl fail", err)
	}
	return err
}

func DeleteTpl(tplId int) error {
	_,err := x.Exec("DELETE from tpl where id=?", tplId)
	if err != nil {
		mid.LogError("delete tpl fail", err)
	}
	return err
}

func GetStrategyTable(id int) (error,m.StrategyTable) {
	var result []*m.StrategyTable
	err := x.SQL("SELECT * FROM strategy WHERE id=?", id).Find(&result)
	if err != nil || len(result) <= 0 {
		mid.LogError("get strategy table fail", err)
		return err,m.StrategyTable{}
	}
	return nil,*result[0]
}

func GetEndpointsByGrp(grpId int) (error,[]*m.EndpointTable) {
	var result []*m.EndpointTable
	err := x.SQL("SELECT * FROM endpoint WHERE id IN (SELECT endpoint_id FROM grp_endpoint WHERE grp_id=?)", grpId).Find(&result)
	if err != nil {
		mid.LogError("get endpoint by grp fail", err)
	}
	return err,result
}

func GetAlarms(query m.AlarmTable) (error,m.AlarmProblemList) {
	var result []*m.AlarmProblemQuery
	var whereSql,extWhereSql string
	var params,extParams []interface{}
	if query.Id > 0 {
		whereSql += " and t1.id=? "
		params = append(params, query.Id)
	}
	if query.StrategyId > 0 {
		whereSql += " and t1.strategy_id=? "
		params = append(params, query.StrategyId)
	}
	if query.Endpoint != "" {
		whereSql += " and t1.endpoint=? "
		params = append(params, query.Endpoint)
	}
	if query.SMetric != "" {
		whereSql += " and t1.s_metric=? "
		params = append(params, query.SMetric)
	}
	if query.SPriority != "" {
		whereSql += " and t1.s_priority=? "
		params = append(params, query.SPriority)
	}
	if query.Tags != "" {
		whereSql += " and t1.tags=? "
		params = append(params, query.Tags)
	}
	extWhereSql = whereSql
	extParams = params
	if query.Status != "" {
		whereSql += " and t1.status=? "
		params = append(params, query.Status)
		if query.Status == "firing" {
			extWhereSql += "and t1.status!='closed' "
		}
	}
	if !query.Start.IsZero() {
		whereSql += fmt.Sprintf(" and t1.start>='%s' ", query.Start.Format(m.DatetimeFormat))
	}
	if !query.End.IsZero() {
		whereSql += fmt.Sprintf(" and t1.end<='%s' ", query.End.Format(m.DatetimeFormat))
	}
	for _,v := range extParams {
		params = append(params, v)
	}
	//err := x.SQL("SELECT t1.*,t2.path,t2.keyword FROM alarm t1 LEFT JOIN log_monitor t2 ON t1.strategy_id=t2.strategy_id where 1=1 " + whereSql + " order by t1.id desc", params...).Find(&result)
	sql := `SELECT t3.* FROM (
			SELECT t1.*,'' path,'' keyword FROM alarm t1 WHERE t1.s_metric<>'log_monitor' `+whereSql+`
			UNION 
			SELECT t1.*,t2.path,t2.keyword FROM alarm t1 LEFT JOIN log_monitor t2 ON t1.strategy_id=t2.strategy_id WHERE t1.s_metric='log_monitor' `+extWhereSql+`
			) t3 ORDER BY t3.id DESC`
	err := x.SQL(sql,params...).Find(&result)
	if err != nil {
		mid.LogError("get alarms fail", err)
	}
	for _,v := range result {
		v.StartString = v.Start.Format(m.DatetimeFormat)
		v.EndString = v.End.Format(m.DatetimeFormat)
		if v.Path != "" {
			v.IsLogMonitor = true
		}
	}
	for _,v := range GetOpenAlarm() {
		result = append(result, v)
	}
	var sortResult m.AlarmProblemList
	sortResult = result
	sort.Sort(sortResult)
	return err,sortResult
}

func UpdateAlarms(alarms []*m.AlarmTable) error {
	if len(alarms) == 0 {
		return nil
	}
	var actions []*Action
	for _,v := range alarms {
		var action Action
		var cErr error
		params := make([]interface{}, 0)
		if v.Id > 0 {
			action.Sql = "UPDATE alarm SET status=?,end_value=?,end=? WHERE id=?"
			params = append(params, v.Status)
			params = append(params, v.EndValue)
			params = append(params, v.End.Format(m.DatetimeFormat))
			params = append(params, v.Id)
			_,cErr = x.Exec(action.Sql, v.Status, v.EndValue, v.End.Format(m.DatetimeFormat), v.Id)
		}else{
			action.Sql = "INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start,tags) VALUE (?,?,?,?,?,?,?,?,?,?,NOW(),?)"
			params = append(params, v.StrategyId)
			params = append(params, v.Endpoint)
			params = append(params, v.Status)
			params = append(params, v.SMetric)
			params = append(params, v.SExpr)
			params = append(params, v.SCond)
			params = append(params, v.SLast)
			params = append(params, v.SPriority)
			params = append(params, v.Content)
			params = append(params, v.StartValue)
			//params = append(params, v.Start.Format(m.DatetimeFormat))
			params = append(params, v.Tags)
			_,cErr = x.Exec(action.Sql, v.StrategyId, v.Endpoint, v.Status, v.SMetric, v.SExpr, v.SCond, v.SLast, v.SPriority, v.Content, v.StartValue, v.Tags)
		}
		action.Param = params
		if action.Sql != "" {
			actions = append(actions, &action)
		}
		if cErr != nil {
			mid.LogInfo(fmt.Sprintf("update alarm:%s param:%v fail with:%v", action.Sql, action.Param, cErr))
		}
	}
	//if len(actions) > 0 {
	//	return Transaction(actions)
	//}
	return nil
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
	for _,v := range obj.LogMonitor {
		action := Classify(*v, obj.Operation, "log_monitor", false)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	err := Transaction(actions)
	return err
}

func GetLogMonitorTable(id,strategyId,tplId int, path string) (err error,result []*m.LogMonitorTable) {
	if id > 0 {
		err = x.SQL("SELECT * FROM log_monitor WHERE id=?", id).Find(&result)
	}
	if path != "" {
		err = x.SQL("SELECT * FROM log_monitor WHERE path=?", path).Find(&result)
	}
	if strategyId > 0 {
		err = x.SQL("SELECT * FROM log_monitor WHERE strategy_id=?", strategyId).Find(&result)
	}
	if tplId > 0 {
		err = x.SQL("SELECT * FROM log_monitor WHERE strategy_id IN (SELECT id FROM strategy WHERE tpl_id=?) ORDER BY path", tplId).Find(&result)
	}
	return err,result
}

func GetLogMonitorByEndpoint(endpointId int) (err error,result []*m.LogMonitorTable) {
	sql := `SELECT DISTINCT t1.* FROM log_monitor t1 
			LEFT JOIN strategy t2 ON t1.strategy_id=t2.id 
			LEFT JOIN tpl t3 ON t2.tpl_id=t3.id 
			WHERE t3.endpoint_id=? 
			OR t3.grp_id IN (SELECT grp_id FROM grp_endpoint WHERE endpoint_id=?) 
			ORDER BY t1.path`
	err = x.SQL(sql, endpointId, endpointId).Find(&result)
	return err,result
}

func ListLogMonitor(query *m.TplQuery) error {
	var result []*m.TplObj
	if query.SearchType == "endpoint" {
		var grps []*m.GrpTable
		err := x.SQL("SELECT id,name FROM grp where id in (select grp_id from grp_endpoint WHERE endpoint_id=?)", query.SearchId).Find(&grps)
		if err != nil {
			mid.LogError("get strategy fail", err)
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
				WHERE (`+grpIds+` t1.endpoint_id=?) and t2.config_type='log_monitor' ORDER BY t1.endpoint_id,t1.id,t3.path`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			mid.LogError("get log monitor strategy fail", err)
			return err
		}
		if len(tpls) == 0 {
			endpointObj := m.EndpointTable{Id:query.SearchId}
			GetEndpoint(&endpointObj)
			result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}, LogMonitor: []*m.LogMonitorDto{}})
		}else {
			var tmpTplId int
			var tmpLogMonitor []*m.LogMonitorDto
			keywordMap := make(map[string][]*m.LogMonitorStrategyDto)
			for _,v := range tpls {
				key := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if vv,b := keywordMap[key];!b {
					keywordMap[key] = []*m.LogMonitorStrategyDto{&m.LogMonitorStrategyDto{Id:v.LogMonitorId, StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority}}
				}else{
					keywordMap[key] = append(vv, &m.LogMonitorStrategyDto{Id:v.LogMonitorId, StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority})
				}
			}
			existFlag := make(map[string]int)
			for i, v := range tpls {
				tmpMapKey := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if i == 0 {
					tmpTplId = v.TplId
					if v.StrategyId > 0 {
						if _,b := existFlag[tmpMapKey];!b {
							tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id:v.StrategyId, TplId:v.TplId, Path:v.Path, Strategy:keywordMap[tmpMapKey]})
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
						if _,b := existFlag[tmpMapKey];!b {
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
			}else{
				result = append(result, &m.TplObj{TplId:tpls[len(tpls)-1].TplId, ObjId:tpls[len(tpls)-1].GrpId, ObjName:grpMap[tpls[len(tpls)-1].GrpId], ObjType:"grp", Operation:false, LogMonitor:tmpLogMonitor})
				endpointObj := m.EndpointTable{Id:query.SearchId}
				GetEndpoint(&endpointObj)
				result = append(result, &m.TplObj{TplId: 0, ObjId: query.SearchId, ObjName: endpointObj.Guid, ObjType: "endpoint", Operation: true, Strategy: []*m.StrategyTable{}, LogMonitor: []*m.LogMonitorDto{}})
			}
		}
	}else{
		var grps []*m.GrpTable
		err := x.SQL("SELECT * FROM grp WHERE id=?", query.SearchId).Find(&grps)
		if err != nil {
			mid.LogError("get grp fail", err)
			return err
		}
		if len(grps) <= 0 {
			mid.LogInfo("can't find this grp")
			return fmt.Errorf("can't find this grp")
		}
		var tpls []*m.TplStrategyLogMonitorTable
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.expr,t2.cond,t2.last,t2.priority,t3.id log_monitor_id,t3.path,t3.keyword FROM tpl t1 
			LEFT JOIN strategy t2 ON t1.id=t2.tpl_id 
			LEFT JOIN log_monitor t3 ON t2.id=t3.strategy_id 
			WHERE t1.grp_id=? and t2.config_type='log_monitor' ORDER BY t1.endpoint_id,t1.id,t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			mid.LogError("get log monitor strategy fail", err)
			return err
		}
		if len(tpls) > 0 {
			keywordMap := make(map[string][]*m.LogMonitorStrategyDto)
			for _,v := range tpls {
				tmpMapKey := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if vv,b := keywordMap[tmpMapKey];!b {
					keywordMap[tmpMapKey] = []*m.LogMonitorStrategyDto{&m.LogMonitorStrategyDto{StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority}}
				}else{
					keywordMap[tmpMapKey] = append(vv, &m.LogMonitorStrategyDto{StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority})
				}
			}
			tmpLogMonitor := []*m.LogMonitorDto{}
			existFlag := make(map[string]int)
			for _, v := range tpls {
				tmpMapKey := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if v.StrategyId > 0 {
					if _,b := existFlag[tmpMapKey];!b {
						tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id: v.StrategyId, TplId: v.TplId, Path: v.Path, Strategy: keywordMap[fmt.Sprintf("%d^%s", v.TplId, v.Path)]})
						existFlag[tmpMapKey] = 1
					}
				}
			}
			result = append(result, &m.TplObj{TplId:tpls[0].TplId, ObjId:tpls[0].GrpId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, LogMonitor:tmpLogMonitor})
		}else{
			result = append(result, &m.TplObj{TplId:0, ObjId:query.SearchId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, LogMonitor:[]*m.LogMonitorDto{}})
		}
	}
	query.Tpl = result
	return nil
}

func getLastFromExpr(expr string) string {
	var last string
	if strings.Contains(expr, "[") {
		last = strings.Split(strings.Split(expr,"[")[1],"]")[0]
	}else{
		last = "10s"
	}
	return last
}

func CloseAlarm(id int) error {
	_,err := x.Exec("UPDATE alarm SET STATUS='closed',end=NOW() WHERE id=?", id)
	return err
}

func GetGrpStrategy(idList []string) (err error,result []*m.GrpStrategyExportObj) {
	sql := `SELECT t1.name,t1.description,t3.metric,t3.expr,t3.cond,t3.last,t3.priority,t3.content,t3.config_type 
		FROM grp t1 
		LEFT JOIN tpl t2 ON t1.id=t2.grp_id 
		LEFT JOIN strategy t3 ON t2.id=t3.tpl_id 
		WHERE t1.id IN `
	sql = sql + fmt.Sprintf("(%s)", strings.Join(idList, ",")) + " ORDER BY t1.name"
	var queryResult []*m.GrpStrategyQuery
	err = x.SQL(sql).Find(&queryResult)
	if err != nil {
		return err,result
	}
	if len(queryResult) == 0 {
		return nil,result
	}
	var tmpStrategyList []m.StrategyTable
	tmpName := queryResult[0].Name
	for i,v := range queryResult {
		if v.Name != tmpName {
			tmpObj := m.GrpStrategyExportObj{GrpName:tmpName, Description:queryResult[i-1].Description, Strategy:tmpStrategyList}
			result = append(result, &tmpObj)
			tmpStrategyList = []m.StrategyTable{}
			tmpName = v.Name
		}
		tmpStrategyList = append(tmpStrategyList, m.StrategyTable{Metric:v.Metric,Expr:v.Expr,Cond:v.Cond,Last:v.Last,Priority:v.Priority,Content:v.Content,ConfigType:v.ConfigType})
	}
	tmpObj := m.GrpStrategyExportObj{GrpName:tmpName, Description:queryResult[len(queryResult)-1].Description, Strategy:tmpStrategyList}
	result = append(result, &tmpObj)
	return nil,result
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
	for _,v := range paramObj {
		tmpName := takeGrpName(v.GrpName, existGrp)
		err := UpdateGrp(&m.UpdateGrp{Operation:"insert", Groups:[]*m.GrpTable{&m.GrpTable{Name:tmpName, Description:v.Description}}})
		if err != nil {
			mid.LogError("Set group strategy, insert group fail", err)
			return err
		}
		_,grpObj := GetSingleGrp(0, tmpName)
		err,tplObj := AddTpl(grpObj.Id, 0, "")
		if err != nil {
			mid.LogError("Set group strategy, insert tpl fail", err)
			return err
		}
		for _,vv := range v.Strategy {
			strategyObj := m.StrategyTable{TplId:tplObj.Id,Metric:vv.Metric,Expr:vv.Expr,Cond:vv.Cond,Last:vv.Last,Priority:vv.Priority,Content:vv.Content,ConfigType:vv.ConfigType}
			UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"insert"})
		}
	}
	return nil
}

func takeGrpName(name string,grpList []*m.GrpTable) string {
	exist := false
	tmpIndex := 0
	for _,v := range grpList {
		if v.Name == name {
			exist = true
		}
		if strings.HasPrefix(v.Name, name) && strings.Contains(v.Name, "_") {
			tmpList := strings.Split(v.Name, "_")
			ii,_ := strconv.Atoi(tmpList[len(tmpList)-1])
			if ii > tmpIndex {
				tmpIndex = ii
			}
		}
	}
	if !exist {
		return name
	}else{
		if tmpIndex > 0 {
			name = strings.Replace(name, fmt.Sprintf("_%d", tmpIndex), "", -1)
		}
		return fmt.Sprintf("%s_%d", name, tmpIndex+1)
	}
}

func DeleteStrategyByGrp(grpId int,tplId int) error {
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
	var alertLevel,subSystemId int
	for _, v := range param.AlertList {
		if v.AlertLevel == "0" {
			var query []*m.OpenAlarmObj
			x.SQL("SELECT id FROM alarm_custom WHERE alert_ip=? AND alert_title=? AND alert_obj=?", v.AlertIp, v.AlertTitle, v.AlertObj).Find(&query)
			if len(query) > 0 {
				tmpIds := ""
				for _,vv := range query {
					tmpIds += fmt.Sprintf("%d,", vv.Id)
				}
				tmpIds = tmpIds[:len(tmpIds)-1]
				_,cErr := x.Exec(fmt.Sprintf("UPDATE alarm_custom SET closed=1,closed_at=NOW() WHERE id in (%s)", tmpIds))
				if cErr != nil {
					mid.LogError(fmt.Sprintf("update custom alarm %s close fail", tmpIds), cErr)
				}
			}
			continue
		}
		alertLevel,_ = strconv.Atoi(v.AlertLevel)
		subSystemId,_ = strconv.Atoi(v.SubSystemId)
		_,err = x.Exec("INSERT INTO alarm_custom(alert_info,alert_ip,alert_level,alert_obj,alert_title,alert_reciver,remark_info,sub_system_id) VALUE (?,?,?,?,?,?,?,?)", v.AlertInfo, v.AlertIp, alertLevel,v.AlertObj,v.AlertTitle,v.AlertReciver,v.RemarkInfo, subSystemId)
		if err != nil {
			mid.LogError("save open alarm error", err)
		}
	}
	return err
}

func GetOpenAlarm() []*m.AlarmProblemQuery {
	var query []*m.OpenAlarmObj
	result := []*m.AlarmProblemQuery{}
	//sql := fmt.Sprintf("SELECT * FROM alarm_custom WHERE closed<>1 and update_at>'%s' ORDER BY id ASC", time.Unix(time.Now().Unix()-300,0).Format(m.DatetimeFormat))
	sql := fmt.Sprintf("SELECT * FROM alarm_custom WHERE closed<>1 ORDER BY id ASC")
	x.SQL(sql).Find(&query)
	if len(query) == 0 {
		return result
	}
	tmpFlag := fmt.Sprintf("%d_%s_%s_%d", query[0].SubSystemId,query[0].AlertTitle,query[0].AlertIp,query[0].AlertLevel)
	for i,v := range query {
		if tmpFlag != fmt.Sprintf("%d_%s_%s_%d", v.SubSystemId,v.AlertTitle,v.AlertIp,v.AlertLevel) {
			priority := "high"
			tmpAlertLevel,_ := strconv.Atoi(query[i-1].AlertLevel)
			if tmpAlertLevel > 4 {
				priority = "low"
			}else if tmpAlertLevel > 2 {
				priority = "medium"
			}
			query[i-1].AlertInfo = strings.Replace(query[i-1].AlertInfo, "\n", " <br/> ", -1)
			result = append(result, &m.AlarmProblemQuery{IsCustom:true,Id:query[i-1].Id,Endpoint:query[i-1].AlertIp,Status:"firing",Content:fmt.Sprintf("system_id:%s <br/> title:%s <br/> %s info: <br/> %s ",query[i-1].SubSystemId,query[i-1].AlertTitle,query[i-1].AlertObj,query[i-1].AlertInfo),Start:query[i-1].UpdateAt,StartString:query[i-1].UpdateAt.Format(m.DatetimeFormat),SPriority:priority})
		}
	}
	priority := "high"
	lastIndex := len(query)-1
	tmpAlertLevel,_ := strconv.Atoi(query[lastIndex].AlertLevel)
	if tmpAlertLevel > 4 {
		priority = "low"
	}else if tmpAlertLevel > 2 {
		priority = "medium"
	}
	query[lastIndex].AlertInfo = strings.Replace(query[lastIndex].AlertInfo, "\n", " <br/> ", -1)
	result = append(result, &m.AlarmProblemQuery{IsCustom:true,Id:query[lastIndex].Id,Endpoint:query[lastIndex].AlertIp,Status:"firing",IsLogMonitor:false,Content:fmt.Sprintf("system_id:%s <br/> title:%s <br/> %s info: <br/> %s ",query[lastIndex].SubSystemId,query[lastIndex].AlertTitle,query[lastIndex].AlertObj,query[lastIndex].AlertInfo),Start:query[lastIndex].UpdateAt,StartString:query[lastIndex].UpdateAt.Format(m.DatetimeFormat),SPriority:priority})
	return result
}

func CloseOpenAlarm(id int) error {
	var query,secQuery []*m.OpenAlarmObj
	x.SQL("SELECT * FROM alarm_custom WHERE id=?", id).Find(&query)
	if len(query) == 0 {
		return fmt.Errorf("alarm id %d cat not find", id)
	}
	err := x.SQL(fmt.Sprintf("SELECT id FROM alarm_custom WHERE alert_ip='%s' AND alert_title='%s' AND alert_obj='%s'", query[0].AlertIp, query[0].AlertTitle, query[0].AlertObj)).Find(&secQuery)
	if len(secQuery) > 0 {
		tmpIds := ""
		for _,vv := range secQuery {
			tmpIds += fmt.Sprintf("%d,", vv.Id)
		}
		tmpIds = tmpIds[:len(tmpIds)-1]
		_,err = x.Exec(fmt.Sprintf("UPDATE alarm_custom SET closed=1,closed_at=NOW() WHERE id in (%s)", tmpIds))
		if err != nil {
			mid.LogError(fmt.Sprintf("update custom alarm %s close fail", tmpIds), err)
		}
	}
	return err
}

func UpdateTplAction(tplId int,user,role []int) error {
	var userString,roleString string
	if len(user) > 0 {
		for _,v := range user {
			userString += fmt.Sprintf("%d,", v)
		}
		userString = userString[:len(userString)-1]
	}
	if len(role) > 0 {
		for _,v := range role {
			roleString += fmt.Sprintf("%d,", v)
		}
		roleString = roleString[:len(roleString)-1]
	}
	_,err := x.Exec(fmt.Sprintf("UPDATE tpl SET action_user='%s',action_role='%s' WHERE id=%d", userString, roleString, tplId))
	if err != nil {
		mid.LogError("Update tpl action error", err)
	}
	return err
}

func getActionOptions(tplId int) []*m.OptionModel {
	var tpls []*m.TplTable
	result := []*m.OptionModel{}
	x.SQL("SELECT action_user,action_role FROM tpl WHERE id=?", tplId).Find(&tpls)
	if len(tpls) == 0 {
		return result
	}
	if tpls[0].ActionRole != "" {
		var roles []*m.RoleTable
		x.SQL(fmt.Sprintf("SELECT id,name,display_name FROM role WHERE id IN (%s)", tpls[0].ActionRole)).Find(&roles)
		for _,v := range roles {
			tmpText := v.Name
			if v.DisplayName != "" {
				tmpText = tmpText + "(" + v.DisplayName + ")"
			}
			result = append(result, &m.OptionModel{Id:v.Id, OptionText:tmpText, OptionValue:fmt.Sprintf("role_%d", v.Id), Active:true})
		}
	}
	if tpls[0].ActionUser != "" {
		var users []*m.UserTable
		x.SQL(fmt.Sprintf("SELECT id,NAME,display_name FROM user WHERE id IN (%s)", tpls[0].ActionUser)).Find(&users)
		for _,v := range users {
			tmpText := v.Name
			if v.DisplayName != "" {
				tmpText = tmpText + "(" + v.DisplayName + ")"
			}
			result = append(result, &m.OptionModel{Id:v.Id, OptionText:tmpText, OptionValue:fmt.Sprintf("user_%d", v.Id), Active:false})
		}
	}
	return result
}