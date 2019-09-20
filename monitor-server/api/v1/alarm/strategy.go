package alarm

import (
	"github.com/gin-gonic/gin"
	"strconv"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	"net/http"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/prom"
	"fmt"
	"strings"
)

func ListTpl(c *gin.Context)  {
	searchType := c.Query("type")
	id,_ := strconv.Atoi(c.Query("id"))
	if searchType == "" || id <= 0 {
		mid.ReturnValidateFail(c, "type or id is null")
		return
	}
	if !(searchType == "endpoint" || searchType == "grp") {
		mid.ReturnValidateFail(c, "type mast be endpoint or grp")
		return
	}
	var query m.TplQuery
	query.SearchType = searchType
	query.SearchId = id
	err := db.GetStrategys(&query)
	if err != nil {
		mid.ReturnError(c, "query strategy error", err)
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
				mid.ReturnValidateFail(c, "endpoint and group id is null")
				return
			}
			if param.GrpId > 0 && param.EndpointId > 0 {
				mid.ReturnValidateFail(c, "endpoint and group id is both not null")
				return
			}
			err,tplObj := db.AddTpl(param.GrpId, param.EndpointId, "")
			if err != nil {
				mid.ReturnError(c, "add strategy fail", err)
				return
			}
			param.TplId = tplObj.Id
		}
		strategyObj := m.StrategyTable{TplId:param.TplId,Metric:param.Metric,Expr:param.Expr,Cond:param.Cond,Last:param.Last,Priority:param.Priority,Content:param.Content}
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"insert"})
		if err != nil {
			mid.ReturnError(c, "Insert strategy fail", err)
			return
		}
		err = saveConfigFile(param.TplId)
		if err != nil {
			mid.ReturnError(c, "save prometheus rule file fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.Return(c, mid.RespJson{Msg:"Param validate fail", Code:http.StatusBadRequest})
	}
}

func EditStrategy(c *gin.Context)  {
	var param m.TplStrategyTable
	if err := c.ShouldBindJSON(&param); err==nil {
		if param.StrategyId <= 0 {
			mid.ReturnValidateFail(c, "strategyId must not null")
			return
		}
		strategyObj := m.StrategyTable{Id:param.StrategyId,Metric:param.Metric,Expr:param.Expr,Cond:param.Cond,Last:param.Last,Priority:param.Priority,Content:param.Content}
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"update"})
		if err != nil {
			mid.ReturnError(c, "Update strategy fail", err)
			return
		}
		_,strategy := db.GetStrategyTable(param.StrategyId)
		db.UpdateTpl(strategy.TplId, "")
		err = saveConfigFile(strategy.TplId)
		if err != nil {
			mid.ReturnError(c, "save prometheus rule file fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.Return(c, mid.RespJson{Msg:"Param validate fail", Code:http.StatusBadRequest})
	}
}

func DeleteStrategy(c *gin.Context)  {
	strategyId,_ := strconv.Atoi(c.Query("id"))
	if strategyId <= 0 {
		mid.ReturnValidateFail(c, "id is null")
		return
	}
	_,strategy := db.GetStrategyTable(strategyId)
	if strategy.Id <= 0 {
		mid.ReturnValidateFail(c, "this id is not in used")
		return
	}
	err := db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&m.StrategyTable{Id:strategyId}}, Operation:"delete"})
	if err != nil {
		mid.ReturnError(c, "Delete strategy fail", err)
		return
	}
	db.UpdateTpl(strategy.TplId, "")
	err = saveConfigFile(strategy.TplId)
	if err != nil {
		mid.ReturnError(c, "save prometheus rule file fail", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}

func SearchObjOption(c *gin.Context)  {
	searchType := c.Query("type")
	searchMsg := c.Query("search")
	if searchType == "" || searchMsg == "" {
		mid.ReturnValidateFail(c, "type or search is null")
		return
	}
	var err error
	var data []*m.OptionModel
	if searchType == "endpoint" {
		err,data = db.SearchHost(searchMsg, true)
	}else{
		err,data = db.SearchGrp(searchMsg)
	}
	if err != nil {
		mid.ReturnError(c, "search fail", err)
		return
	}
	mid.ReturnData(c, data)
}

func saveConfigFile(tplId int) error {
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
	err = db.GetStrategys(&query)
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
			_,grpObj := db.GetSingleGrp(tplObj.GrpId)
			fileName = grpObj.Name
		}else{
			_,endpointObj := db.GetEndpoint(tplObj.EndpointId, "")
			fileName = endpointObj.Guid
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
		for _,v := range query.Tpl[len(query.Tpl)-1].Strategy {
			tmpRfu := m.RFRule{}
			tmpRfu.Alert = v.Metric
			if !strings.Contains(v.Cond, " ") {
				if strings.Contains(v.Cond, "=") {
					v.Cond = v.Cond[:2] + " " + v.Cond[2:]
				}else{
					v.Cond = v.Cond[:1] + " " + v.Cond[1:]
				}
			}
			if strings.Contains(v.Expr, " ") {
				v.Expr = strings.Replace(v.Expr, " ", "", -1)
			}
			if isGrp && strings.Contains(v.Expr, "$address") {
				v.Expr = strings.Replace(v.Expr, "=\"$address\"", "=~\""+endpointExpr+"\"", -1)
			}
			tmpRfu.Expr = fmt.Sprintf("%s %s", v.Expr, v.Cond)
			tmpRfu.For = v.Last
			tmpRfu.Labels = make(map[string]string)
			tmpRfu.Labels["strategy_id"] = fmt.Sprintf("%d", v.Id)
			tmpRfu.Annotations = m.RFAnnotation{Summary:fmt.Sprintf("{{$labels.instance}} %s %s", v.Priority, v.Metric), Description:v.Content}
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