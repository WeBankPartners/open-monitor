package alarm

import (
	"encoding/json"
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func ListTpl(c *gin.Context) {
	searchType := c.Query("type")
	id, _ := strconv.Atoi(c.Query("id"))
	if searchType == "" || id <= 0 {
		mid.ReturnParamEmptyError(c, "type or id")
		return
	}
	if !(searchType == "endpoint" || searchType == "grp") {
		mid.ReturnValidateError(c, "type must be \"endpoint\" or \"grp\"")
		return
	}
	var query m.TplQuery
	query.SearchType = searchType
	query.SearchId = id
	err := db.GetTplStrategy(&query, true)
	if err != nil {
		mid.ReturnQueryTableError(c, "strategy", err)
		return
	}
	mid.ReturnSuccessData(c, query.Tpl)
}

func AddStrategy(c *gin.Context) {
	var param m.TplStrategyTable
	if err := c.ShouldBindJSON(&param); err == nil {
		// check param
		param.Expr = strings.Replace(param.Expr, "'", "", -1)
		param.Content = strings.Replace(param.Content, "'", "", -1)
		param.Content = strings.Replace(param.Content, "\"", "", -1)
		if !mid.IsIllegalCond(param.Cond) || !mid.IsIllegalLast(param.Last) {
			mid.ReturnValidateError(c, "cond or last illegal")
			return
		}
		// check tpl
		if param.TplId <= 0 {
			if param.GrpId+param.EndpointId <= 0 {
				mid.ReturnValidateError(c, "grp_id and endpoint_id is empty")
				return
			}
			if param.GrpId > 0 && param.EndpointId > 0 {
				mid.ReturnValidateError(c, "grp_id and endpoint_id can not be provided at the same time")
				return
			}
			err, tplObj := db.AddTpl(param.GrpId, param.EndpointId, mid.GetOperateUser(c))
			if err != nil {
				mid.ReturnUpdateTableError(c, "tpl", err)
				return
			}
			param.TplId = tplObj.Id
		}
		strategyObj := m.StrategyTable{TplId: param.TplId, Metric: param.Metric, Expr: param.Expr, Cond: param.Cond, Last: param.Last, Priority: param.Priority, Content: param.Content}
		strategyObj.NotifyEnable = param.NotifyEnable
		strategyObj.NotifyDelay = param.NotifyDelay
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy: []*m.StrategyTable{&strategyObj}, Operation: "insert"})
		if err != nil {
			mid.ReturnUpdateTableError(c, "strategy", err)
			return
		}
		//err = SaveConfigFile(param.TplId, false)
		err = db.SyncRuleConfigFile(param.TplId, []string{}, false)
		if err != nil {
			mid.ReturnHandleError(c, "save alert rules file failed", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func EditStrategy(c *gin.Context) {
	var param m.TplStrategyTable
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.StrategyId <= 0 {
			mid.ReturnParamEmptyError(c, "strategy_id")
			return
		}
		_, strategy := db.GetStrategyTable(param.StrategyId)
		if strategy.TplId <= 0 {
			mid.ReturnHandleError(c, "template for this strategy is empty", nil)
			return
		}
		// check param
		param.Expr = strings.Replace(param.Expr, "'", "", -1)
		param.Content = strings.Replace(param.Content, "'", "", -1)
		param.Content = strings.Replace(param.Content, "\"", "", -1)
		if !mid.IsIllegalCond(param.Cond) || !mid.IsIllegalLast(param.Last) {
			mid.ReturnValidateError(c, "cond or last illegal")
			return
		}
		strategyObj := m.StrategyTable{Id: param.StrategyId, TplId: strategy.TplId, Metric: param.Metric, Expr: param.Expr, Cond: param.Cond, Last: param.Last, Priority: param.Priority, Content: param.Content, NotifyEnable: param.NotifyEnable, NotifyDelay: param.NotifyDelay}
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy: []*m.StrategyTable{&strategyObj}, Operation: "update"})
		if err != nil {
			mid.ReturnUpdateTableError(c, "strategy", err)
			return
		}
		db.UpdateTpl(strategy.TplId, mid.GetOperateUser(c))
		err = db.SyncRuleConfigFile(strategy.TplId, []string{}, false)
		//err = SaveConfigFile(strategy.TplId, false)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Sync rule config file fail", zap.Error(err))
			mid.ReturnHandleError(c, "save alert rules file failed", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func DeleteStrategy(c *gin.Context) {
	strategyId, _ := strconv.Atoi(c.Query("id"))
	if strategyId <= 0 {
		mid.ReturnParamEmptyError(c, "id")
		return
	}
	_, strategy := db.GetStrategyTable(strategyId)
	if strategy.Id <= 0 {
		mid.ReturnFetchDataError(c, "strategy", "id", strconv.Itoa(strategyId))
		return
	}
	err := db.UpdateStrategy(&m.UpdateStrategy{Strategy: []*m.StrategyTable{&m.StrategyTable{Id: strategyId}}, Operation: "delete"})
	if err != nil {
		mid.ReturnUpdateTableError(c, "strategy", err)
		return
	}
	db.UpdateTpl(strategy.TplId, "")
	//err = SaveConfigFile(strategy.TplId, false)
	err = db.SyncRuleConfigFile(strategy.TplId, []string{}, false)
	if err != nil {
		mid.ReturnHandleError(c, "save prometheus rule file failed", err)
		return
	}
	mid.ReturnSuccess(c)
}

func SearchObjOption(c *gin.Context) {
	searchType := c.Query("type")
	searchMsg := c.Query("search")
	if searchType == "" || searchMsg == "" {
		mid.ReturnParamEmptyError(c, "type and search")
		return
	}
	var err error
	var data []*m.OptionModel
	if searchType == "endpoint" {
		err, data = db.SearchHost(searchMsg)
	} else {
		err, data = db.SearchGrp(searchMsg)
	}
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	for _, v := range data {
		v.OptionTypeName = v.OptionType
	}
	mid.ReturnSuccessData(c, data)
}

func SearchUserRole(c *gin.Context) {
	search := c.Query("search")
	err, roles := db.SearchUserRole(search, "role")
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Search role error", zap.Error(err))
	}
	if len(roles) < 15 {
		err, users := db.SearchUserRole(search, "user")
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Search user error", zap.Error(err))
		}
		for _, v := range users {
			if len(roles) >= 15 {
				break
			}
			roles = append(roles, v)
		}
	}
	mid.ReturnSuccessData(c, roles)
}

func UpdateTplAction(c *gin.Context) {
	var param m.UpdateActionDto
	if err := c.ShouldBindJSON(&param); err == nil {
		var userIds, roleIds []int
		var extraMail, extraPhone []string
		for _, v := range param.Accept {
			tmpFlag := false
			if strings.HasPrefix(v.OptionType, "user_") {
				tmpId, _ := strconv.Atoi(strings.Split(v.OptionType, "_")[1])
				for _, vv := range userIds {
					if vv == tmpId {
						tmpFlag = true
						break
					}
				}
				if !tmpFlag {
					userIds = append(userIds, tmpId)
				}
			}
			if strings.HasPrefix(v.OptionType, "role_") {
				tmpId, _ := strconv.Atoi(strings.Split(v.OptionType, "_")[1])
				for _, vv := range roleIds {
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
				for _, vv := range extraMail {
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
				for _, vv := range extraPhone {
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
			mid.ReturnUpdateTableError(c, "tpl", err)
		} else {
			mid.ReturnSuccess(c)
		}
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

func SyncConfigHandle(w http.ResponseWriter, r *http.Request) {
	log.Debug(nil, log.LOGGER_APP, "Start sync config")
	var response mid.RespJson
	w.Header().Set("Content-Type", "application/json")
	defer w.Write([]byte(fmt.Sprintf("{\"Code\":%d,\"Message\":\"%s\",\"Data\":\"%v\"}", response.Code, response.Message, response.Data)))
	tplId, _ := strconv.Atoi(r.FormValue("id"))
	if tplId <= 0 {
		response.Code = 401
		response.Message = "Parameter id is empty"
		return
	}
	//err := SaveConfigFile(tplId, true)
	err := db.SyncRuleConfigFile(tplId, []string{}, true)
	if err != nil {
		response.Code = 500
		response.Message = "Sync save config file fail"
		response.Data = err
		return
	}
	response.Code = 200
	response.Message = "Success"
}

func AcceptPeerSdConfigHandle(w http.ResponseWriter, r *http.Request) {
	log.Info(nil, log.LOGGER_APP, "start to sync sd config from peer request")
	var response mid.RespJson
	w.Header().Set("Content-Type", "application/json")
	defer w.Write([]byte(fmt.Sprintf("{\"Code\":%d,\"Message\":\"%s\",\"Data\":\"%v\"}", response.Code, response.Message, response.Data)))
	var param m.SyncSdConfigDto
	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &param)
	if err != nil {
		response.Code = 401
		response.Message = "Param json format fail"
		response.Data = err
		return
	}
	if param.Guid == "" {
		response.Code = 401
		response.Message = "Guid is empty"
		return
	}
	err = db.SyncSdEndpointNew(param.StepList, "default", true)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Handle sd config from peer fail", zap.Error(err))
		response.Code = 500
		response.Message = "Sync peer sd config fail"
		response.Data = err.Error()
		return
	}
	response.Code = 200
	response.Message = "Success"
}
