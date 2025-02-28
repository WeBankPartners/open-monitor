package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"time"
)

func ListGrp(query *m.GrpQuery) error {
	var querySql = `SELECT * FROM grp WHERE 1=1 `
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
		whereSql += ` AND (name like '%` + query.Search + `%' or description like '%` + query.Search + `%') `
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
	err = x.SQL(countSql, cParams...).Find(&count)
	if len(result) > 0 {
		query.Result = result
		query.ResultNum = count[0]
	} else {
		query.Result = []*m.GrpTable{}
		query.ResultNum = 0
	}
	return err
}

func GetSingleGrp(id int, name string) (error, m.GrpTable) {
	var result []*m.GrpTable
	err := x.SQL("SELECT * FROM grp WHERE id=? or name=?", id, name).Find(&result)
	if err != nil {
		return fmt.Errorf("Get grp table fail,%s ", err.Error()), m.GrpTable{}
	}
	if len(result) == 0 {
		return fmt.Errorf("Can not find grp data with id=%d or name=%s ", id, name), m.GrpTable{}
	}
	return nil, *result[0]
}

func SearchGrp(search string) (error, []*m.OptionModel) {
	var result []*m.OptionModel
	var grps []*m.GrpTable
	if search == "." {
		search = ""
	}
	search = "%" + search + "%"
	err := x.SQL(`SELECT * FROM grp WHERE name LIKE ?`, search).Find(&grps)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Search grp fail", zap.Error(err))
		return err, result
	}
	for _, v := range grps {
		result = append(result, &m.OptionModel{OptionValue: fmt.Sprintf("%d", v.Id), OptionText: v.Name, Id: v.Id, OptionType: v.EndpointType, OptionTypeName: v.EndpointType})
	}
	return nil, result
}

func UpdateGrp(obj *m.UpdateGrp) error {
	var actions []*Action
	for _, grp := range obj.Groups {
		grp.UpdateUser = obj.OperateUser
		if obj.Operation == "insert" {
			grp.CreateUser = obj.OperateUser
			grp.CreateAt = time.Now()
			grp.UpdateAt = time.Now()
		}
		action := Classify(*grp, obj.Operation, "grp", true)
		if action.Sql != "" {
			actions = append(actions, &action)
		}
	}
	err := Transaction(actions)
	return err
}

func UpdateEndpointGrp(param m.EndpointGrpParam) (err error, affectGroupIds []int) {
	var grpEndpoints []*m.GrpEndpointTable
	x.SQL("select * from grp_endpoint where endpoint_id=?", param.EndpointId).Find(&grpEndpoints)
	if len(grpEndpoints) > 0 {
		for _, v := range grpEndpoints {
			existFlag := false
			for _, vv := range param.GroupIds {
				if vv == v.GrpId {
					existFlag = true
					break
				}
			}
			if !existFlag {
				// need delete
				affectGroupIds = append(affectGroupIds, v.GrpId)
			}
		}
		for _, v := range param.GroupIds {
			existFlag := false
			for _, vv := range grpEndpoints {
				if vv.GrpId == v {
					existFlag = true
					break
				}
			}
			if !existFlag {
				affectGroupIds = append(affectGroupIds, v)
			}
		}
	} else {
		affectGroupIds = param.GroupIds
	}
	var actions []*Action
	actions = append(actions, &Action{Sql: "delete from grp_endpoint where endpoint_id=?", Param: []interface{}{param.EndpointId}})
	if len(param.GroupIds) > 0 {
		insertSql := "INSERT INTO grp_endpoint(endpoint_id,grp_id) VALUES "
		for _, v := range param.GroupIds {
			insertSql += fmt.Sprintf("(%d,%d),", param.EndpointId, v)
		}
		actions = append(actions, &Action{Sql: insertSql[:len(insertSql)-1]})
	}
	return Transaction(actions), affectGroupIds
}

func UpdateGrpEndpoint(param m.GrpEndpointParamNew) (error, bool) {
	if param.Operation == "update" {
		var actions []*Action
		actions = append(actions, &Action{Sql: "delete from grp_endpoint where grp_id=?", Param: []interface{}{param.Grp}})
		if len(param.Endpoints) > 0 {
			insertSql := "INSERT INTO grp_endpoint VALUES "
			for _, v := range param.Endpoints {
				insertSql += fmt.Sprintf("(%d,%d),", param.Grp, v)
			}
			actions = append(actions, &Action{Sql: insertSql[:len(insertSql)-1]})
		}
		updateError := Transaction(actions)
		if updateError != nil {
			return updateError, false
		} else {
			return updateError, true
		}
	}
	if len(param.Endpoints) == 0 {
		return nil, false
	}
	var ids string
	for _, v := range param.Endpoints {
		ids += fmt.Sprintf("%d,", v)
	}
	if param.Operation == "add" {
		var grpEndpoints []*m.GrpEndpointTable
		err := x.SQL(fmt.Sprintf("SELECT * FROM grp_endpoint WHERE grp_id=%d AND endpoint_id IN (%s)", param.Grp, ids[:len(ids)-1])).Find(&grpEndpoints)
		if err != nil {
			return err, false
		}
		var needAdd = true
		var needInsert = false
		insertSql := "INSERT INTO grp_endpoint VALUES "
		for _, v := range param.Endpoints {
			needAdd = true
			for _, vv := range grpEndpoints {
				if v == vv.EndpointId {
					needAdd = false
					break
				}
			}
			if needAdd {
				insertSql += fmt.Sprintf("(%d,%d),", param.Grp, v)
				needInsert = true
			}
		}
		if needInsert {
			_, err = x.Exec(insertSql[:len(insertSql)-1])
			return err, needInsert
		} else {
			return nil, needInsert
		}
	}
	if param.Operation == "delete" {
		_, err := x.Exec(fmt.Sprintf("DELETE FROM grp_endpoint WHERE grp_id=%d AND endpoint_id IN (%s)", param.Grp, ids[:len(ids)-1]))
		return err, true
	}
	return fmt.Errorf("operation is not add or delete"), false
}

func getGrpParent(grpId int) m.GrpTable {
	var grp []*m.GrpTable
	x.SQL("SELECT id,name,parent FROM grp WHERE id=?", grpId).Find(&grp)
	if len(grp) > 0 {
		return *grp[0]
	}
	return m.GrpTable{}
}

func DeleteEndpointFromGroup(endpointId int) (tplList []int, err error) {
	var tplTable []*m.TplTable
	err = x.SQL("select * from tpl where grp_id in (select grp_id from grp_endpoint where endpoint_id=?)", endpointId).Find(&tplTable)
	if err != nil {
		err = fmt.Errorf("Get delete endpoint affect tpl list fail,%s ", err.Error())
		return
	}
	for _, tpl := range tplTable {
		tplList = append(tplList, tpl.Id)
	}
	_, err = x.Exec("delete from grp_endpoint where endpoint_id=?", endpointId)
	if err != nil {
		err = fmt.Errorf("Delete group endpoint relation fail,%s ", err.Error())
	}
	return
}
