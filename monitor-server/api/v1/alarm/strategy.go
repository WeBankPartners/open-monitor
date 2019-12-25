package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"fmt"
	"strings"
)

func ListTpl(c *gin.Context)  {
	searchType := c.Query("type")
	id,_ := strconv.Atoi(c.Query("id"))
	if searchType == "" || id <= 0 {
		mid.ReturnValidateFail(c, "Type or id can not be empty")
		return
	}
	if !(searchType == "endpoint" || searchType == "grp") {
		mid.ReturnValidateFail(c, "Type must be \"endpoint\" or \"grp\"")
		return
	}
	var query m.TplQuery
	query.SearchType = searchType
	query.SearchId = id
	err := db.GetStrategys(&query, true)
	if err != nil {
		mid.ReturnError(c, "Query strategy failed", err)
		return
	}
	mid.ReturnData(c, query.Tpl)
}

func AddStrategy(c *gin.Context)  {
	var param m.TplStrategyTable
	if err := c.ShouldBindJSON(&param); err==nil {
		// check tpl
		if param.TplId <= 0 {
			if param.GrpId + param.EndpointId <= 0 {
				mid.ReturnValidateFail(c, "Both endpoint and group id are missing")
				return
			}
			if param.GrpId > 0 && param.EndpointId > 0 {
				mid.ReturnValidateFail(c, "Endpoint and group id can not be provided at the same time")
				return
			}
			err,tplObj := db.AddTpl(param.GrpId, param.EndpointId, "")
			if err != nil {
				mid.ReturnError(c, "Add strategy failed", err)
				return
			}
			param.TplId = tplObj.Id
		}
		strategyObj := m.StrategyTable{TplId:param.TplId,Metric:param.Metric,Expr:param.Expr,Cond:param.Cond,Last:param.Last,Priority:param.Priority,Content:param.Content}
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"insert"})
		if err != nil {
			mid.ReturnError(c, "Insert strategy failed", err)
			return
		}
		err = SaveConfigFile(param.TplId)
		if err != nil {
			mid.ReturnError(c, "Save alert rules file failed", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func EditStrategy(c *gin.Context)  {
	var param m.TplStrategyTable
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.StrategyId <= 0 {
			mid.ReturnValidateFail(c, "Strategy id can not be empty")
			return
		}
		strategyObj := m.StrategyTable{Id:param.StrategyId,Metric:param.Metric,Expr:param.Expr,Cond:param.Cond,Last:param.Last,Priority:param.Priority,Content:param.Content}
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"update"})
		if err != nil {
			mid.ReturnError(c, "Update strategy failed", err)
			return
		}
		_,strategy := db.GetStrategyTable(param.StrategyId)
		db.UpdateTpl(strategy.TplId, "")
		err = SaveConfigFile(strategy.TplId)
		if err != nil {
			mid.ReturnError(c, "Save alert rules file failed", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func DeleteStrategy(c *gin.Context)  {
	strategyId,_ := strconv.Atoi(c.Query("id"))
	if strategyId <= 0 {
		mid.ReturnValidateFail(c, "Id can not be empty")
		return
	}
	_,strategy := db.GetStrategyTable(strategyId)
	if strategy.Id <= 0 {
		mid.ReturnValidateFail(c, "The strategy id is not in use")
		return
	}
	err := db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&m.StrategyTable{Id:strategyId}}, Operation:"delete"})
	if err != nil {
		mid.ReturnError(c, "Delete strategy failed", err)
		return
	}
	db.UpdateTpl(strategy.TplId, "")
	err = SaveConfigFile(strategy.TplId)
	if err != nil {
		mid.ReturnError(c, "Save prometheus rule file failed", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}

func SearchObjOption(c *gin.Context)  {
	searchType := c.Query("type")
	searchMsg := c.Query("search")
	if searchType == "" || searchMsg == "" {
		mid.ReturnValidateFail(c, "Type or search content can not be empty")
		return
	}
	var err error
	var data []*m.OptionModel
	if searchType == "endpoint" {
		err,data = db.SearchHost(searchMsg)
	}else{
		err,data = db.SearchGrp(searchMsg)
	}
	if err != nil {
		mid.ReturnError(c, "Search failed", err)
		return
	}
	mid.ReturnData(c, data)
}

func SaveConfigFile(tplId int) error {
	err,tplObj := db.GetTpl(tplId,0 ,0)
	if err != nil {
		return err
	}
	var query m.TplQuery
	var isGrp bool
	if tplObj.GrpId > 0 {
		isGrp = true
		query.SearchType = "grp"
		query.SearchId = tplObj.GrpId
	}else{
		isGrp = false
		query.SearchType = "endpoint"
		query.SearchId = tplObj.EndpointId
	}
	err = db.GetStrategys(&query, false)
	if err != nil {
		return err
	}
	var fileName string
	var endpointExpr string
	if len(query.Tpl) > 0 {
		fileName = query.Tpl[len(query.Tpl)-1].ObjName
	}else{
		fmt.Printf("is grp : %b \n", isGrp)
		if isGrp {
			_,grpObj := db.GetSingleGrp(tplObj.GrpId, "")
			fileName = grpObj.Name
		}else{
			endpointObj := m.EndpointTable{Id:tplObj.EndpointId}
			db.GetEndpoint(&endpointObj)
			fileName = endpointObj.Guid
			endpointExpr = endpointObj.Address
		}
	}
	if isGrp {
		_,endpointObjs := db.GetEndpointsByGrp(tplObj.GrpId)
		if len(endpointObjs) > 0 {
			for _, v := range endpointObjs {
				endpointExpr += fmt.Sprintf("%s|", v.Address)
			}
			endpointExpr = endpointExpr[:len(endpointExpr)-1]
		}
		fmt.Printf("expr : %s , objs : %v \n", endpointExpr, endpointObjs)
	}
	err,isExist,cObj := prom.GetConfig(fileName, isGrp)
	if err != nil {
		return err
	}
	rfu := []*m.RFRule{}
	if !isExist {
		cObj.Name = fileName
	}
	if len(query.Tpl) > 0 {
		if !isGrp && endpointExpr == "" && query.Tpl[len(query.Tpl)-1].ObjType == "endpoint" {
			endpointObj := m.EndpointTable{Guid:query.Tpl[len(query.Tpl)-1].ObjName}
			db.GetEndpoint(&endpointObj)
			endpointExpr = endpointObj.Address
		}
		for _,v := range query.Tpl[len(query.Tpl)-1].Strategy {
			tmpRfu := m.RFRule{}
			tmpRfu.Alert = v.Metric
			if !strings.Contains(v.Cond, " ") && v.Cond != "" {
				if strings.Contains(v.Cond, "=") {
					v.Cond = v.Cond[:2] + " " + v.Cond[2:]
				}else{
					v.Cond = v.Cond[:1] + " " + v.Cond[1:]
				}
			}
			//if strings.Contains(v.Expr, " ") {
			//	v.Expr = strings.Replace(v.Expr, " ", "", -1)
			//}
			if strings.Contains(v.Expr, "$address") {
				if isGrp {
					v.Expr = strings.Replace(v.Expr, "=\"$address\"", "=~\""+endpointExpr+"\"", -1)
				}else{
					v.Expr = strings.Replace(v.Expr, "=\"$address\"", "=\""+endpointExpr+"\"", -1)
				}
			}
			tmpRfu.Expr = fmt.Sprintf("%s %s", v.Expr, v.Cond)
			tmpRfu.For = v.Last
			tmpRfu.Labels = make(map[string]string)
			tmpRfu.Labels["strategy_id"] = fmt.Sprintf("%d", v.Id)
			tmpRfu.Annotations = m.RFAnnotation{Summary:fmt.Sprintf("{{$labels.instance}}__%s__%s__{{$value}}", v.Priority, v.Metric), Description:v.Content}
			rfu = append(rfu, &tmpRfu)
		}
		if len(query.Tpl[len(query.Tpl)-1].Strategy) == 0 {
			rfu = []*m.RFRule{}
		}
	}
	cObj.Rules = rfu
	err = prom.SetConfig(fileName, isGrp, cObj, isExist)
	return err
}

func SearchUserRole(c *gin.Context)  {
	search := c.Query("search")
	err,roles := db.SearchUserRole(search, "role")
	if err != nil {
		mid.LogError("search role error", err)
	}
	if len(roles) < 15 {
		err,users := db.SearchUserRole(search, "user")
		if err != nil {
			mid.LogError("search user error", err)
		}
		for _,v := range users {
			if len(roles) >= 15 {
				break
			}
			roles = append(roles, v)
		}
	}
	mid.ReturnData(c, roles)
}

func UpdateTplAction(c *gin.Context)  {
	var param m.UpdateActionDto
	if err := c.ShouldBindJSON(&param); err==nil {
		var userIds,roleIds []int
		for _,v := range param.Accept {
			if strings.Contains(v, "user_") {
				tmpId,_ := strconv.Atoi(strings.Split(v, "user_")[1])
				if tmpId > 0 {
					userIds = append(userIds, tmpId)
				}
			}
			if strings.Contains(v, "role_") {
				tmpId,_ := strconv.Atoi(strings.Split(v, "role_")[1])
				if tmpId > 0 {
					roleIds = append(roleIds, tmpId)
				}
			}
		}
		err = db.UpdateTplAction(param.TplId, userIds, roleIds)
		if err != nil {
			mid.ReturnError(c, "Update tpl action fail ", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}