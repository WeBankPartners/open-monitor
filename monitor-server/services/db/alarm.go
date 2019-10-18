package db

import (
	"reflect"
	"strconv"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"fmt"
	"time"
	"strings"
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

func GetSingleGrp(id int) (error,m.GrpTable) {
	var result []*m.GrpTable
	err := x.SQL("SELECT * FROM grp WHERE id=?", id).Find(&result)
	if err != nil || len(result) <= 0 {
		mid.LogError("get single grp fail", err)
		return err,m.GrpTable{}
	}
	return nil,*result[0]
}

func SearchGrp(search string) (error,[]*m.OptionModel) {
	var result []*m.OptionModel
	var grps []*m.GrpTable
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
            SELECT t4.id,t4.guid,GROUP_CONCAT(t4.name, ',') groups_name FROM (
			SELECT t1.id,t1.guid,t3.name FROM endpoint t1 
			LEFT JOIN grp_endpoint t2 ON t1.id=t2.endpoint_id 
			LEFT JOIN grp t3 ON t2.grp_id=t3.id 
			WHERE 1=1 `+whereSql+`
			) t4 GROUP BY t4.guid
			) t5 ORDER BY t5.guid LIMIT ?,?`
	countSql := `SELECT COUNT(1) num FROM (
			SELECT t4.guid,GROUP_CONCAT(t4.name, ',') groups_name FROM (
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
	var sqls []string
	for _,grp := range obj.Groups {
		grp.UpdateUser = obj.OperateUser
		if obj.Operation == "insert" {
			grp.CreateUser = obj.OperateUser
			//grp.CreateAt = time.Now()
		}
		sql := Classify(*grp, obj.Operation, "grp", true)
		if sql != "" {
			sqls = append(sqls, sql)
		}
	}
	err := ExecuteTransactionSql(sqls)
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

func Classify(obj interface{}, operation string, table string, force bool) string {
	sql := ``
	if operation == "insert" {
		sql = insert(obj, table)
	}else if operation == "update" {
		sql = update(obj, table, force)
	}else if operation == "delete" {
		sql = delete(obj, table)
	}
	return sql
}

func insert(obj interface{}, table string) string {
	column := `(`
	value := ` value (`
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	length := t.NumField()
	for i := 0; i < length; i++ {
		if t.Field(i).Name == "Id" {
			if v.Field(i).Int() == 0 {
				continue
			}
		}
		f := v.Field(i).Type().String()
		switch f {
		case "string":
			if i == length-1 {
				column = column + transColumn(t.Field(i).Name) + `)`
				value = value + `'` + v.Field(i).String() + `')`
			} else {
				column = column + transColumn(t.Field(i).Name) + `,`
				value = value + `'` + v.Field(i).String() + `',`
			}
		case "int":
			if i == length-1 {
				column = column + transColumn(t.Field(i).Name) + `)`
				value = value + `'` + strconv.FormatInt(v.Field(i).Int(), 10) + `')`
			} else {
				column = column + transColumn(t.Field(i).Name) + `,`
				value = value + `'` + strconv.FormatInt(v.Field(i).Int(), 10) + `',`
			}
		case "int64":
			if i == length-1 {
				column = column + transColumn(t.Field(i).Name) + `)`
				value = value + `'` + strconv.FormatInt(v.Field(i).Int(), 10) + `')`
			} else {
				column = column + transColumn(t.Field(i).Name) + `,`
				value = value + `'` + strconv.FormatInt(v.Field(i).Int(), 10) + `',`
			}
		case "time.Time":
			if i == length-1 {
				column = column + transColumn(t.Field(i).Name) + `)`
				value = value + `'` + time.Now().Format(m.DatetimeFormat) + `')`
			} else {
				column = column + transColumn(t.Field(i).Name) + `,`
				value = value + `'` + time.Now().Format(m.DatetimeFormat) + `',`
			}
		default:
			if i == length-1 {
				column = column + transColumn(t.Field(i).Name) + `)`
				value = value + `'` + v.Field(i).String() + `')`
			} else {
				column = column + transColumn(t.Field(i).Name) + `,`
				value = value + `'` + v.Field(i).String() + `',`
			}
		}
	}
	sql := `insert into ` + table + column + value
	return sql
}

func update(obj interface{}, table string, force bool) string {
	value := ``
	where := ` where id=`
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	length := t.NumField()
	flag := false
	for i := 0; i < length; i++ {
		if t.Field(i).Name == "Id" {
			where = where + strconv.FormatInt(v.Field(i).Int(), 10)
		}else {
			f := v.Field(i).Kind()
			switch f {
			case reflect.String:
				if v.Field(i).String() != "" || force {
					value = value + transColumn(t.Field(i).Name) + `='` + v.Field(i).String() + `',`
					flag = true
				}
			case reflect.Int:
				if v.Field(i).Int() > 0 || force {
					value = value + transColumn(t.Field(i).Name) + `=` + strconv.FormatInt(v.Field(i).Int(), 10) + `,`
					flag = true
				}
			}
		}
	}
	if flag==true{
		value = value[0:len(value)-1]
		sql := `update ` + table + ` set ` + value + where
		return sql
	}else{
		return ``
	}
}

func delete(obj interface{}, table string) string {
	where := ` where id=`
	where_sec := ` where `
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	length := t.NumField()
	flag := 0
	for i := 0; i < length; i++ {
		if t.Field(i).Name == "Id" {
			where = where + strconv.FormatInt(v.Field(i).Int(), 10)
			flag = 1
			break
		}
		f := v.Field(i).Kind()
		switch f {
		case reflect.String:
			if v.Field(i).String() != "" {
				if flag == 2 {
					where_sec = where_sec + ` and `
				}
				where_sec = where_sec + transColumn(t.Field(i).Name) + `='` + v.Field(i).String() + `'`
				flag = 2
			}
		case reflect.Int:
			if v.Field(i).Int() > 0 {
				if flag == 2 {
					where_sec = where_sec + ` and `
				}
				where_sec = where_sec + transColumn(t.Field(i).Name) + `=` + strconv.FormatInt(v.Field(i).Int(), 10)
				flag = 2
			}
		}
	}
	if flag == 1 {
		sql := `delete from ` + table + where
		return sql
	}else if flag == 2 {
		sql := `delete from ` + table + where_sec
		return sql
	}else{
		return ``
	}
}

func transColumn(s string) string {
	r := []byte(s)
	var v []byte
	for i := 0; i < len(r); i++ {
		rr := r[i]
		if 'A' <= rr && rr <= 'Z' {
			rr += 'a' - 'A'
			if i != 0 {
				v = append(v, '_')
			}
		}
		v = append(v, rr)
	}
	return string(v)
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

func GetStrategys(query *m.TplQuery) error {
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
		var tpls []*m.TplStrategyTable
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE `+grpIds+` endpoint_id=? order by t1.endpoint_id,t1.id,t2.id`
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
		var tpls []*m.TplStrategyTable
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.metric,t2.expr,t2.cond,t2.last,t2.priority,t2.content 
				FROM tpl t1 LEFT JOIN strategy t2 ON t1.id=t2.tpl_id WHERE t1.grp_id=? ORDER BY t2.id`
		err = x.SQL(sql, query.SearchId).Find(&tpls)
		if err != nil {
			mid.LogError("get strategy fail", err)
			return err
		}
		if len(tpls) > 0 {
			tmpStrategys := []*m.StrategyTable{}
			for _, v := range tpls {
				if v.StrategyId > 0 {
					tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId: v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content})
				}
			}
			result = append(result, &m.TplObj{TplId:tpls[0].TplId, ObjId:tpls[0].GrpId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, Strategy:tmpStrategys})
		}else{
			result = append(result, &m.TplObj{TplId:0, ObjId:query.SearchId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, Strategy:[]*m.StrategyTable{}})
		}
	}
	query.Tpl = result
	return nil
}

func UpdateStrategy(obj *m.UpdateStrategy) error {
	var sqls []string
	for _,v := range obj.Strategy {
		sql := Classify(*v, obj.Operation, "strategy", false)
		if sql != "" {
			sqls = append(sqls, sql)
		}
	}
	err := ExecuteTransactionSql(sqls)
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

func AddTpl(grpId,endpointId int, operateUser string) (error,m.TplTable) {
	_,tpl := GetTpl(0, grpId, endpointId)
	if tpl.Id > 0 {
		return nil,tpl
	}
	insertSql := fmt.Sprintf("INSERT INTO tpl(grp_id,endpoint_id,create_user,update_user,create_at) VALUE (%d,%d,'%s','%s',NOW())", grpId, endpointId, operateUser, operateUser)
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

func GetAlarms(query m.AlarmTable) (error,[]*m.AlarmProblemQuery) {
	var result []*m.AlarmProblemQuery
	var whereSql,extWhereSql string
	var params []interface{}
	var extParams []interface{}
	if query.Id > 0 {
		whereSql += " and t1.id=? "
		extWhereSql += " and t1.id=? "
		params = append(params, query.Id)
		extParams = append(extParams, query.Id)
	}
	if query.StrategyId > 0 {
		whereSql += " and t1.strategy_id=? "
		extWhereSql += " and t1.strategy_id=? "
		params = append(params, query.StrategyId)
		extParams = append(extParams, query.StrategyId)
	}
	if query.Endpoint != "" {
		whereSql += " and t1.endpoint=? "
		extWhereSql += " and t1.endpoint=? "
		params = append(params, query.Endpoint)
		extParams = append(extParams, query.Endpoint)
	}
	if query.Status != "" {
		whereSql += " and t1.status=? "
		params = append(params, query.Status)
		if query.Status == "firing" {
			extWhereSql += "and t1.status!='closed' "
		}
	}
	if query.SMetric != "" {
		whereSql += " and t1.s_metric=? "
		params = append(params, query.SMetric)
	}
	if query.SPriority != "" {
		whereSql += " and t1.s_priority=? "
		extWhereSql += " and t1.s_priority=? "
		params = append(params, query.SPriority)
		extParams = append(extParams, query.SPriority)
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
	fmt.Println(sql)
	fmt.Println(params)
	err := x.SQL(sql,params...).Find(&result)
	if err != nil {
		mid.LogError("get alarms fail", err)
	}
	for _,v := range result {
		v.StartString = v.Start.Format(m.DatetimeFormat)
		if v.Path != "" {
			v.IsLogMonitor = true
		}
	}
	return err,result
}

func UpdateAlarms(alarms []*m.AlarmTable) error {
	var sqls []string
	for _,v := range alarms {
		if v.Id > 0 {
			sqls = append(sqls, fmt.Sprintf("UPDATE alarm SET status='%s',end_value=%f,end='%s' WHERE id=%d", v.Status, v.EndValue, v.End.Format(m.DatetimeFormat), v.Id))
		}else{
			if !judgeExist(*v) {
				sqls = append(sqls, fmt.Sprintf("INSERT INTO alarm(strategy_id,endpoint,status,s_metric,s_expr,s_cond,s_last,s_priority,content,start_value,start) VALUE (%d,'%s','%s','%s','%s','%s','%s','%s','%s',%f,'%s')",
					v.StrategyId, v.Endpoint, v.Status, v.SMetric, v.SExpr, v.SCond, v.SLast, v.SPriority, v.Content, v.StartValue, v.Start.Format(m.DatetimeFormat)))
			}
		}
	}
	return ExecuteTransactionSql(sqls)
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
	var sqls []string
	for _,v := range obj.LogMonitor {
		sql := Classify(*v, obj.Operation, "log_monitor", false)
		if sql != "" {
			sqls = append(sqls, sql)
		}
	}
	err := ExecuteTransactionSql(sqls)
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
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.expr,t2.cond,t2.last,t2.priority,t3.path,t3.keyword FROM tpl t1 
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
					keywordMap[key] = []*m.LogMonitorStrategyDto{&m.LogMonitorStrategyDto{StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority}}
				}else{
					keywordMap[key] = append(vv, &m.LogMonitorStrategyDto{StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority})
				}
			}
			for i, v := range tpls {
				if i == 0 {
					tmpTplId = v.TplId
					if v.StrategyId > 0 {
						tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id:v.StrategyId, TplId:v.TplId, Path:v.Path, Strategy:keywordMap[fmt.Sprintf("%d^%s", v.TplId, v.Path)]})
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
						tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id:v.StrategyId, TplId:v.TplId, Path:v.Path, Strategy:keywordMap[fmt.Sprintf("%d^%s", v.TplId, v.Path)]})
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
		sql := `SELECT t1.id tpl_id,t1.grp_id,t1.endpoint_id,t2.id strategy_id,t2.expr,t2.cond,t2.last,t2.priority,t3.path,t3.keyword FROM tpl t1 
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
				key := fmt.Sprintf("%d^%s", v.TplId, v.Path)
				if vv,b := keywordMap[key];!b {
					keywordMap[key] = []*m.LogMonitorStrategyDto{&m.LogMonitorStrategyDto{StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority}}
				}else{
					keywordMap[key] = append(vv, &m.LogMonitorStrategyDto{StrategyId:v.StrategyId, Keyword:v.Keyword, Cond:v.Cond, Last:getLastFromExpr(v.Expr), Priority:v.Priority})
				}
			}
			tmpLogMonitor := []*m.LogMonitorDto{}
			for _, v := range tpls {
				if v.StrategyId > 0 {
					tmpLogMonitor = append(tmpLogMonitor, &m.LogMonitorDto{Id:v.StrategyId, TplId:v.TplId, Path:v.Path, Strategy:keywordMap[fmt.Sprintf("%d^%s", v.TplId, v.Path)]})
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
	_,err := x.Exec("UPDATE alarm SET STATUS='closed' WHERE id=?", id)
	return err
}