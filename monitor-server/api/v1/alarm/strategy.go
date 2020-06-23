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
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"net/http"
	"io/ioutil"
	"encoding/json"
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
		// check param
		param.Expr = strings.Replace(param.Expr, "'", "", -1)
		param.Content = strings.Replace(param.Content, "'", "", -1)
		param.Content = strings.Replace(param.Content, "\"", "", -1)
		if !mid.IsIllegalCond(param.Cond) || !mid.IsIllegalLast(param.Last) {
			mid.ReturnValidateFail(c, "cond or last param validate fail")
			return
		}
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
		err = SaveConfigFile(param.TplId, false)
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
		// check param
		param.Expr = strings.Replace(param.Expr, "'", "", -1)
		param.Content = strings.Replace(param.Content, "'", "", -1)
		param.Content = strings.Replace(param.Content, "\"", "", -1)
		if !mid.IsIllegalCond(param.Cond) || !mid.IsIllegalLast(param.Last) {
			mid.ReturnValidateFail(c, "cond or last param validate fail")
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
		err = SaveConfigFile(strategy.TplId, false)
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
	err = SaveConfigFile(strategy.TplId, false)
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
	for _,v := range data {
		v.OptionTypeName = v.OptionType
	}
	mid.ReturnData(c, data)
}

func SaveConfigFile(tplId int, fromCluster bool) error  {
	var err error
	idList := db.GetParentTpl(tplId)
	err = updateConfigFile(tplId)
	if err != nil {
		mid.LogError("update prometheus rule file error", err)
		return err
	}
	if len(idList) > 0 {
		for _,v := range idList {
			err = updateConfigFile(v)
			if err != nil {
				mid.LogError(fmt.Sprintf("update prometheus rule tpl id %d error", v), err)
			}
		}
	}
	if err != nil {
		return err
	}
	err = prom.ReloadConfig()
	if err != nil {
		mid.LogError("reload prometheus config error", err)
		return err
	}
	if !fromCluster {
		go other.SyncConfig(tplId, m.SyncConsulDto{})
	}
	return nil
}

func updateConfigFile(tplId int) error {
	err,tplObj := db.GetTpl(tplId,0 ,0)
	if err != nil {
		mid.LogError("get tpl error", err)
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
		mid.LogError("get strategy error", err)
		return err
	}
	var fileName string
	var endpointExpr,guidExpr string
	if len(query.Tpl) > 0 {
		fileName = query.Tpl[len(query.Tpl)-1].ObjName
		if isGrp {
			tmpStrategy := []*m.StrategyTable{}
			tmpStrategyMap := make(map[string]*m.StrategyTable)
			for _,v := range query.Tpl {
				for _,vv := range v.Strategy {
					tmpStrategyMap[vv.Metric] = vv
				}
			}
			for _,v := range tmpStrategyMap {
				tmpStrategy = append(tmpStrategy, v)
			}
			query.Tpl[len(query.Tpl)-1].Strategy = tmpStrategy
		}
	}else{
		if isGrp {
			_,grpObj := db.GetSingleGrp(tplObj.GrpId, "")
			fileName = grpObj.Name
		}else{
			endpointObj := m.EndpointTable{Id:tplObj.EndpointId}
			db.GetEndpoint(&endpointObj)
			fileName = endpointObj.Guid
			if endpointObj.AddressAgent != "" {
				endpointExpr = endpointObj.AddressAgent
			}else {
				endpointExpr = endpointObj.Address
			}
			guidExpr = endpointObj.Guid
		}
	}
	if isGrp {
		_,endpointObjs := db.GetEndpointsByGrp(tplObj.GrpId)
		if len(endpointObjs) > 0 {
			for _, v := range endpointObjs {
				if v.AddressAgent != "" {
					endpointExpr += fmt.Sprintf("%s|", v.AddressAgent)
				}else {
					endpointExpr += fmt.Sprintf("%s|", v.Address)
				}
				guidExpr += fmt.Sprintf("%s|", v.Guid)
			}
			endpointExpr = endpointExpr[:len(endpointExpr)-1]
			guidExpr = guidExpr[:len(guidExpr)-1]
		}
	}
	if fileName == "" {
		return nil
	}
	err,isExist,cObj := prom.GetConfig(fileName, isGrp)
	if err != nil {
		mid.LogError("get prom get config error", err)
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
			if endpointObj.AddressAgent != "" {
				endpointExpr = endpointObj.AddressAgent
			}else {
				endpointExpr = endpointObj.Address
			}
			guidExpr = endpointObj.Guid
		}
		for _,v := range query.Tpl[len(query.Tpl)-1].Strategy {
			tmpRfu := m.RFRule{}
			tmpRfu.Alert = fmt.Sprintf("%s_%d", v.Metric, v.Id)
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
			if strings.Contains(v.Expr, "$guid") {
				if isGrp {
					v.Expr = strings.Replace(v.Expr, "=\"$guid\"", "=~\""+guidExpr+"\"", -1)
				}else{
					v.Expr = strings.Replace(v.Expr, "=\"$guid\"", "=\""+guidExpr+"\"", -1)
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
	if err != nil {
		mid.LogError("prom set config error", err)
	}
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
		var extraMail,extraPhone []string
		for _,v := range param.Accept {
			tmpFlag := false
			if strings.HasPrefix(v.OptionType, "user_") {
				tmpId,_ := strconv.Atoi(strings.Split(v.OptionType, "_")[1])
				for _,vv := range userIds {
					if vv == tmpId {
						tmpFlag = true
						break
					}
				}
				if !tmpFlag {
					userIds = append(userIds, tmpId)
				}
			}
			if strings.HasPrefix(v.OptionType,"role_") {
				tmpId,_ := strconv.Atoi(strings.Split(v.OptionType, "_")[1])
				for _,vv := range roleIds {
					if vv == tmpId {
						tmpFlag = true
						break
					}
				}
				if !tmpFlag {
					roleIds = append(roleIds, tmpId)
				}
			}
			if v.OptionType == "mail" {
				for _,vv := range extraMail {
					if vv == v.OptionValue {
						tmpFlag = true
						break
					}
				}
				if !tmpFlag {
					extraMail = append(extraMail, v.OptionValue)
				}
			}
			if v.OptionType == "phone" {
				for _,vv := range extraPhone {
					if vv == v.OptionValue {
						tmpFlag = true
						break
					}
				}
				if !tmpFlag {
					extraPhone = append(extraPhone, v.OptionValue)
				}
			}
		}
		err = db.UpdateTplAction(param.TplId, userIds, roleIds, extraMail, extraPhone)
		if err != nil {
			mid.ReturnError(c, "Update tpl action fail ", err)
		}else{
			mid.ReturnSuccess(c, "Success")
		}
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

func SyncConfigHandle(w http.ResponseWriter,r *http.Request)  {
	mid.LogInfo("start sync config")
	var response mid.RespJson
	w.Header().Set("Content-Type", "application/json")
	defer w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"msg\":\"%s\",\"data\":\"%v\"}", response.Code,response.Msg,response.Data)))
	tplId,_ := strconv.Atoi(r.FormValue("id"))
	if tplId <= 0 {
		response.Code = 401
		response.Msg = "Parameter id is empty"
		return
	}
	err := SaveConfigFile(tplId, true)
	if err != nil {
		response.Code = 500
		response.Msg = "Sync save config file fail"
		response.Data = err
		return
	}
	response.Code = 200
	response.Msg = "Success"
}

func SyncConsulHandle(w http.ResponseWriter,r *http.Request)  {
	mid.LogInfo("start sync consul")
	var response mid.RespJson
	w.Header().Set("Content-Type", "application/json")
	defer w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"msg\":\"%s\",\"data\":\"%v\"}", response.Code,response.Msg,response.Data)))
	var param m.SyncConsulDto
	b,_ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &param)
	if err != nil {
		response.Code = 401
		response.Msg = "Param json format fail"
		response.Data = err
		return
	}
	if param.Guid == "" {
		response.Code = 401
		response.Msg = "Guid is empty"
		return
	}
	if param.IsRegister {
		err = prom.RegisteConsul(param.Guid,param.Ip,param.Port,param.Tags,param.Interval,true)
	}else{
		err = prom.DeregisteConsul(param.Guid,true)
	}
	if err != nil {
		response.Code = 500
		response.Msg = "Sync consul fail"
		response.Data = err
		return
	}
	response.Code = 200
	response.Msg = "Success"
}