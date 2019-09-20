package db

import (
	"reflect"
	"strconv"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"fmt"
	"time"
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
		result = append(result, &m.OptionModel{OptionValue:fmt.Sprintf("%d", v.Id), OptionText:v.Name})
	}
	return nil,result
}

func ListEndopint(query *m.AlarmEndpointQuery) error {
	whereSql := ""
	if query.Search != "" {
		whereSql += ` AND t1.guid LIKE '%`+query.Search+`%' `
	}
	if query.Grp > 0 {
		whereSql += fmt.Sprintf(" AND t3.id=%d ", query.Grp)
	}
	querySql := `SELECT t5.* FROM (
            SELECT t4.id,t4.guid,GROUP_CONCAT(t4.name, ',') groups FROM (
			SELECT t1.id,t1.guid,t3.name FROM endpoint t1 
			LEFT JOIN grp_endpoint t2 ON t1.id=t2.endpoint_id 
			LEFT JOIN grp t3 ON t2.grp_id=t3.id 
			WHERE 1=1 `+whereSql+`
			) t4 GROUP BY t4.guid
			) t5 ORDER BY t5.guid LIMIT ?,?`
	countSql := `SELECT COUNT(1) num FROM (
			SELECT t4.guid,GROUP_CONCAT(t4.name, ',') groups FROM (
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

func UpdateGrpEndpoint(param m.GrpEndpointParamNew) error {
	if len(param.Endpoints) == 0 {
		return nil
	}
	var ids string
	for _,v := range param.Endpoints {
		ids += fmt.Sprintf("%d,", v)
	}
	if param.Operation == "add" {
		var grpEndpoints []*m.GrpEndpointTable
		err := x.SQL(fmt.Sprintf("SELECT * FROM grp_endpoint WHERE grp_id=%d AND endpoint_id IN (%s)", param.Grp, ids[:len(ids)-1])).Find(&grpEndpoints)
		if err != nil {
			return err
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
			return err
		}else{
			return nil
		}
	}
	if param.Operation == "delete" {
		_,err := x.Exec(fmt.Sprintf("DELETE FROM grp_endpoint WHERE grp_id=%d AND endpoint_id IN (%s)", param.Grp, ids[:len(ids)-1]))
		return err
	}
	return fmt.Errorf("operation is not add or delete")
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
		var tmpTplId int
		var tmpStrategys []*m.StrategyTable
		for i,v := range tpls {
			if i == 0 {
				tmpTplId = v.TplId
				tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id:v.StrategyId, TplId:v.TplId, Metric:v.Metric, Expr:v.Expr, Cond:v.Cond, Last:v.Last, Priority:v.Priority, Content:v.Content})
			}else{
				if v.TplId != tmpTplId {
					tmpTplObj := m.TplObj{TplId:tpls[i-1].TplId}
					if tpls[i-1].GrpId > 0 {
						tmpTplObj.ObjId = tpls[i-1].GrpId
						tmpTplObj.ObjName = grpMap[tpls[i-1].GrpId]
						tmpTplObj.ObjType = "grp"
						tmpTplObj.Operation = false
					}else{
						tmpTplObj.ObjId = tpls[i-1].EndpointId
						_,endpointObj := GetEndpoint(tpls[i-1].EndpointId, "")
						tmpTplObj.ObjName = endpointObj.Guid
						tmpTplObj.ObjType = "endpoint"
						tmpTplObj.Operation = true
					}
					tmpTplObj.Strategy = tmpStrategys
					result = append(result, &tmpTplObj)
					tmpTplId = v.TplId
					tmpStrategys = []*m.StrategyTable{}
				}
				tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id:v.StrategyId, TplId:v.TplId, Metric:v.Metric, Expr:v.Expr, Cond:v.Cond, Last:v.Last, Priority:v.Priority, Content:v.Content})
			}
		}
		if len(tmpStrategys) > 0 {
			_,endpointObj := GetEndpoint(tpls[len(tpls)-1].EndpointId, "")
			result = append(result, &m.TplObj{TplId:tpls[len(tpls)-1].TplId, ObjId:tpls[len(tpls)-1].EndpointId, ObjName:endpointObj.Guid, ObjType:"endpoint", Operation:true, Strategy:tmpStrategys})
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
				tmpStrategys = append(tmpStrategys, &m.StrategyTable{Id: v.StrategyId, TplId:v.TplId, Metric: v.Metric, Expr: v.Expr, Cond: v.Cond, Last: v.Last, Priority: v.Priority, Content: v.Content})
			}
			result = append(result, &m.TplObj{TplId:tpls[0].TplId, ObjId:tpls[0].GrpId, ObjName:grps[0].Name, ObjType:"grp", Operation:true, Strategy:tmpStrategys})
		}
	}
	if len(result) > 0 {
		query.Tpl = result
	}else{
		query.Tpl = []*m.TplObj{}
	}
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