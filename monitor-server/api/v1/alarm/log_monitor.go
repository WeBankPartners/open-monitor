package alarm

import (
	"github.com/gin-gonic/gin"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"net/http"
	"encoding/json"
	"strings"
	"strconv"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"io/ioutil"
)

// @Summary 日志告警配置接口 : 获取列表
// @Description 获取配置好的对象或组的日志告警列表
// @Produce  json
// @Param type query string true "类型，区分是单个对象还是组，枚举endpoint、grp"
// @Param id query int true "对象或组的id"
// @Router /api/v1/alarm/log/monitor/list [get]
func ListLogTpl(c *gin.Context)  {
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
	err := db.ListLogMonitor(&query)
	if err != nil {
		mid.ReturnError(c, "Query strategy failed", err)
		return
	}
	mid.ReturnData(c, query.Tpl)
}

// @Summary 日志告警配置接口 : 新增
// @Produce  json
// @Param grp_id query int false "组Id，和对象id二选一"
// @Param endpoint_id query int false "对象Id，和组id二选一"
// @Param path query string true "表单输入的日志路径"
// @Param strategy query string true "对象数组类型[{'keyword':'关键字','cond':'条件,如 >1','last':'时间范围,如 5min','priority':'优先级,如 high'}]"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/add [post]
func AddLogStrategy(c *gin.Context)  {
	var param m.LogMonitorDto
	if err := c.ShouldBindJSON(&param);err == nil {
		if len(param.Strategy) == 0 {
			mid.ReturnValidateFail(c, "Parameter must contain a strategy at least")
			return
		}
		if !mid.IsIllegalCond(param.Strategy[0].Cond) || !mid.IsIllegalLast(param.Strategy[0].Last) {
			mid.ReturnValidateFail(c, "Parameter validate fail, cond or last illegal")
			return
		}
		if !mid.IsIllegalPath(param.Path) {
			mid.ReturnValidateFail(c, "Parameter validate fail, path is illegal")
			return
		}
		if !mid.IsIllegalNormalInput(param.Strategy[0].Keyword) {
			mid.ReturnValidateFail(c, "Parameter validate fail, keyword is illegal")
			return
		}
		var logMonitorObj m.LogMonitorTable
		logMonitorObj.Path = param.Path
		logMonitorObj.Keyword = param.Strategy[0].Keyword
		// Add strategy
		if param.TplId <= 0 {
			if param.GrpId+param.EndpointId <= 0 {
				mid.ReturnValidateFail(c, "Both endpoint and group id are missing")
				return
			}
			if param.GrpId > 0 && param.EndpointId > 0 {
				mid.ReturnValidateFail(c, "Endpoint and group id can not be provided at the same time")
				return
			}
			needAdd := true
			if param.Id > 0 {
				_,strategyObj := db.GetStrategy(m.StrategyTable{Id:param.Id})
				if strategyObj.TplId > 0 {
					needAdd = false
					param.TplId = strategyObj.TplId
				}
			}
			if needAdd {
				err, tplObj := db.AddTpl(param.GrpId, param.EndpointId, "")
				if err != nil {
					mid.ReturnError(c, "Add strategy failed", err)
					return
				}
				param.TplId = tplObj.Id
			}
		}
		if param.Id <= 0 {
			_,lms := db.GetLogMonitorTable(0,0, param.TplId, "")
			for _, v := range lms {
				if v.Path == param.Path {
					mid.ReturnValidateFail(c, "Path already exists")
					return
				}
			}
		}
		tmpMetric,tmpExpr,tmpContent := makeStrategyMsg(param.Path, param.Strategy[0].Keyword, param.Strategy[0].Cond, param.Strategy[0].Last)
		strategyObj := m.StrategyTable{TplId:param.TplId,Metric:tmpMetric,Expr:tmpExpr,Cond:param.Strategy[0].Cond,Last:"10s",Priority:param.Strategy[0].Priority,Content:tmpContent}
		strategyObj.ConfigType = "log_monitor"
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"insert"})
		if err != nil {
			mid.ReturnError(c, "Insert strategy failed", err)
			return
		}
		_,strategyObj = db.GetStrategy(m.StrategyTable{Expr:tmpExpr})
		// Add log_monitor
		logMonitorObj.StrategyId = strategyObj.Id
		err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor:[]*m.LogMonitorTable{&logMonitorObj}, Operation:"insert"})
		if err != nil {
			mid.ReturnError(c, "Insert log monitor alert failed", err)
			return
		}
		// Call endpoint node exporter
		err = sendLogConfig(param.EndpointId, param.GrpId, param.TplId)
		if err != nil {
			mid.ReturnError(c, "Send log config to endpoint failed", err)
			return
		}
		// Save Prometheus rule file
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

// @Summary 日志告警配置接口 : 修改日志路径
// @Produce  json
// @Param id query int true "列表获取中的id"
// @Param tpl_id query int true "列表获取中的tpl_id"
// @Param path query string true "新的日志路径"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/update_path [post]
func EditLogPath(c *gin.Context)  {
	var param m.LogMonitorDto
	if err := c.ShouldBindJSON(&param);err == nil {
		if param.TplId <= 0 || param.Id <= 0 {
			mid.ReturnValidateFail(c, "Parameter id or template id can not be empty")
			return
		}
		if !mid.IsIllegalPath(param.Path) {
			mid.ReturnValidateFail(c, "Parameter validate fail, path is illegal")
			return
		}
		err,lms := db.GetLogMonitorTable(0, param.Id, 0, "")
		if err != nil || len(lms) == 0 {
			mid.ReturnError(c, "Get log monitor alert failed", err)
			return
		}
		oldPath := lms[0].Path
		_,lmsGrp := db.GetLogMonitorTable(0, 0, 0, oldPath)
		var strategyObjs []*m.StrategyTable
		// Update log_monitor
		for _,v := range lmsGrp {
			strategyObjs = append(strategyObjs, &m.StrategyTable{Id:v.StrategyId})
			logMonitorObj := m.LogMonitorTable{Id:v.Id, StrategyId:v.StrategyId, Path:param.Path, Keyword:v.Keyword}
			err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor:[]*m.LogMonitorTable{&logMonitorObj}, Operation:"update"})
			if err != nil {
				log.Logger.Error("Update log monitor alert failed", log.Error(err))
			}
		}
		// Update strategy
		for _,v := range strategyObjs {
			err,tmpObj := db.GetStrategyTable(v.Id)
			if err != nil {
				log.Logger.Error("Get strategy failed", log.Int("id", v.Id), log.Error(err))
				continue
			}
			tmpObj.Expr = strings.Replace(tmpObj.Expr, oldPath, param.Path, -1)
			tmpObj.Content = strings.Replace(tmpObj.Content, oldPath, param.Path, -1)
			err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&tmpObj}, Operation:"update"})
			if err != nil {
				log.Logger.Error("Update strategy fail", log.Error(err))
			}
		}
		// Call endpoint node exporter
		err,tplObj := db.GetTpl(param.TplId, 0, 0)
		if err != nil {
			mid.ReturnError(c, "Update log monitor alert failed for getting template error", err)
			return
		}
		param.EndpointId = tplObj.EndpointId
		param.GrpId = tplObj.GrpId
		err = sendLogConfig(param.EndpointId, param.GrpId, param.TplId)
		if err != nil {
			mid.ReturnError(c, "Send log config to endpoint fail", err)
			return
		}
		// Save Prometheus rule file
		err = SaveConfigFile(param.TplId, false)
		if err != nil {
			mid.ReturnError(c, "Save prometheus rule file fail", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

// @Summary 日志告警配置接口 : 修改
// @Produce  json
// @Param tpl_id query int true "列表获取中的tpl_id"
// @Param path query string true "表单输入的日志路径"
// @Param strategy query string true "对象数组类型[{'id':int类型, 'strategy_id':int类型,'keyword':'关键字','cond':'条件,如 >1','last':'时间范围,如 5min','priority':'优先级,如 high'}]"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/update [post]
func EditLogStrategy(c *gin.Context)  {
	var param m.LogMonitorDto
	if err := c.ShouldBindJSON(&param);err == nil {
		if len(param.Strategy) == 0 {
			mid.ReturnValidateFail(c, "Param strategy must contain a strategy at least")
			return
		}
		if param.Strategy[0].Id <= 0 {
			mid.ReturnValidateFail(c, "Param id in strategy can not be empty")
			return
		}
		if param.TplId <= 0 {
			mid.ReturnValidateFail(c, "Param tplId can not be empty")
			return
		}
		if !mid.IsIllegalCond(param.Strategy[0].Cond) || !mid.IsIllegalLast(param.Strategy[0].Last) {
			mid.ReturnValidateFail(c, "Parameter validate fail, cond or last illegal")
			return
		}
		if !mid.IsIllegalNormalInput(param.Strategy[0].Keyword) {
			mid.ReturnValidateFail(c, "Parameter validate fail, keyword is illegal")
			return
		}
		// Update strategy
		err,lms := db.GetLogMonitorTable(param.Strategy[0].Id,0,0,"")
		if err != nil || len(lms) == 0 {
			mid.ReturnError(c, "Update strategy failed for getting log monitor by strategy error", err)
			return
		}
		tmpMetric,tmpExpr,tmpContent := makeStrategyMsg(param.Path, param.Strategy[0].Keyword, param.Strategy[0].Cond, param.Strategy[0].Last)
		strategyObj := m.StrategyTable{Id:lms[0].StrategyId,TplId:param.TplId,Metric:tmpMetric,Expr:tmpExpr,Cond:param.Strategy[0].Cond,Priority:param.Strategy[0].Priority,Content:tmpContent}
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&strategyObj}, Operation:"update"})
		if err != nil {
			mid.ReturnError(c, "Update strategy failed", err)
			return
		}
		// Update log_monitor
		logMonitorObj := m.LogMonitorTable{Id:param.Strategy[0].Id, StrategyId:lms[0].StrategyId, Path:param.Path, Keyword:param.Strategy[0].Keyword}
		err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor:[]*m.LogMonitorTable{&logMonitorObj}, Operation:"update"})
		if err != nil {
			mid.ReturnError(c, "Update log monitor failed", err)
			return
		}
		// Call endpoint node exporter
		err,tplObj := db.GetTpl(param.TplId, 0, 0)
		if err != nil {
			mid.ReturnError(c, "Update log monitor failed for getting tpl error", err)
			return
		}
		param.EndpointId = tplObj.EndpointId
		param.GrpId = tplObj.GrpId
		err = sendLogConfig(param.EndpointId, param.GrpId, param.TplId)
		if err != nil {
			mid.ReturnError(c, "Send log config to endpoint failed", err)
			return
		}
		// Save Prometheus rule file
		err = SaveConfigFile(param.TplId, false)
		if err != nil {
			mid.ReturnError(c, "Save prometheus rule file failed", err)
			return
		}
		mid.ReturnSuccess(c, "Success")
	}else{
		mid.ReturnValidateFail(c, fmt.Sprintf("Parameter validation failed %v", err))
	}
}

// @Summary 日志告警配置接口 : 删除
// @Produce  json
// @Param id query int true "strategy_id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/delete_path [get]
func DeleteLogPath(c *gin.Context)  {
	strategyId,err := strconv.Atoi(c.Query("id"))
	if err != nil || strategyId <= 0 {
		mid.ReturnValidateFail(c, fmt.Sprintf("Param validate failed:%v", err))
		return
	}
	err,strategyObj := db.GetStrategy(m.StrategyTable{Id:strategyId})
	if err != nil || strategyObj.TplId <= 0 {
		mid.ReturnError(c, "Delete strategy failed for getting strategy by id error", err)
		return
	}
	err,lms := db.GetLogMonitorTable(0, strategyId, 0, "")
	if err != nil || len(lms) == 0 {
		mid.ReturnError(c, "Get log monitor alert failed", err)
		return
	}
	oldPath := lms[0].Path
	_,lmsGrp := db.GetLogMonitorTable(0, 0, 0, oldPath)
	var strategyObjs []*m.StrategyTable
	// Delete log monitor
	for _,v := range lmsGrp {
		strategyObjs = append(strategyObjs, &m.StrategyTable{Id:v.StrategyId})
		err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor:[]*m.LogMonitorTable{&m.LogMonitorTable{Id:v.Id}}, Operation:"delete"})
		if err != nil {
			log.Logger.Error("Delete log monitor alert failed", log.Error(err))
		}
	}
	// Delete strategy
	for _,v := range strategyObjs {
		err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&m.StrategyTable{Id:v.Id}}, Operation:"delete"})
		if err != nil {
			log.Logger.Error("Delete strategy failed", log.Error(err))
		}
	}
	// Call endpoint node exporter
	err,tplObj := db.GetTpl(strategyObj.TplId, 0, 0)
	if err != nil {
		mid.ReturnError(c, "Delete log monitor alert failed for getting tpl error", err)
		return
	}
	err = sendLogConfig(tplObj.EndpointId, tplObj.GrpId, tplObj.Id)
	if err != nil {
		mid.ReturnError(c, "Send log config to endpoint failed", err)
		return
	}
	// Save Prometheus rule file
	err = SaveConfigFile(tplObj.Id, false)
	if err != nil {
		mid.ReturnError(c, "Save prometheus rule file failed", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}

// @Summary 日志告警配置接口 : 删除
// @Produce  json
// @Param id query int true "id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/delete [get]
func DeleteLogStrategy(c *gin.Context)  {
	logMonitorId,err := strconv.Atoi(c.Query("id"))
	if err != nil || logMonitorId <= 0 {
		mid.ReturnValidateFail(c, fmt.Sprintf("Param validate failed:%v", err))
		return
	}
	err,lms := db.GetLogMonitorTable(logMonitorId,0,0,"")
	if err != nil || len(lms) == 0 {
		mid.ReturnError(c, "Update strategy failed for getting log monitor by strategy error", err)
		return
	}
	err,strategyObj := db.GetStrategy(m.StrategyTable{Id:lms[0].StrategyId})
	if err != nil || strategyObj.TplId <= 0 {
		mid.ReturnError(c, "Delete strategy failed for getting strategy by id error", err)
		return
	}
	// Delete log monitor
	err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor:[]*m.LogMonitorTable{&m.LogMonitorTable{Id:logMonitorId}}, Operation:"delete"})
	if err != nil {
		mid.ReturnError(c, "Delete log monitor alert failed", err)
		return
	}
	// Delete strategy
	err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&m.StrategyTable{Id:strategyObj.Id}}, Operation:"delete"})
	if err != nil {
		mid.ReturnError(c, "Delete strategy failed", err)
		return
	}
	// Call endpoint node exporter
	err,tplObj := db.GetTpl(strategyObj.TplId, 0, 0)
	if err != nil {
		mid.ReturnError(c, "Delete log monitor alert failed for getting template error", err)
		return
	}
	err = sendLogConfig(tplObj.EndpointId, tplObj.GrpId, tplObj.Id)
	if err != nil {
		mid.ReturnError(c, "Send log config to endpoint failed", err)
		return
	}
	// Save Prometheus rule file
	err = SaveConfigFile(tplObj.Id, false)
	if err != nil {
		mid.ReturnError(c, "Save prometheus rule file failed", err)
		return
	}
	mid.ReturnSuccess(c, "Success")
}

type logHttpDto struct {
	Path  string  `json:"path"`
	Keywords  []string  `json:"keywords"`
}

func sendLogConfig(endpointId,grpId,tplId int) error {
	var endpoints []*m.EndpointTable
	var err error
	if grpId > 0 {
		err,endpoints = db.GetEndpointsByGrp(grpId)
		if err != nil {
			return err
		}
	}
	if endpointId > 0 {
		endpointQuery := m.EndpointTable{Id:endpointId}
		err = db.GetEndpoint(&endpointQuery)
		if err != nil {
			return err
		}
		endpoints = append(endpoints, &endpointQuery)
	}
	var postParam []logHttpDto
	var tmpList []string
	var tmpPath string
	for _,v := range endpoints {
		err,logMonitors := db.GetLogMonitorByEndpoint(v.Id)
		if err != nil {
			log.Logger.Error("Send log config with endpoint failed", log.String("endpoint", v.Guid), log.Error(err))
			continue
		}
		if len(logMonitors) == 0 {
			continue
		}
		postParam = []logHttpDto{}
		tmpList = []string{}
		tmpPath = logMonitors[0].Path
		for _,v := range logMonitors {
			if v.Path != tmpPath {
				postParam = append(postParam, logHttpDto{Path:tmpPath, Keywords:tmpList})
				tmpPath = v.Path
				tmpList = []string{}
			}
			tmpList = append(tmpList, v.Keyword)
		}
		postParam = append(postParam, logHttpDto{Path:logMonitors[len(logMonitors)-1].Path, Keywords:tmpList})
		postData,err := json.Marshal(postParam)
		if err == nil {
			url := fmt.Sprintf("http://%s/log/config", v.Address)
			resp,err := http.Post(url, "application/json", strings.NewReader(string(postData)))
			if err != nil {
				log.Logger.Error("curl "+url+" error ", log.Error(err))
			}else{
				responseBody,_ := ioutil.ReadAll(resp.Body)
				log.Logger.Info("curl " + url, log.String("response", string(responseBody)))
				resp.Body.Close()
			}
		}
	}
	return nil
}

func makeStrategyMsg(path,keyword,cond,last string) (metric,expr,content string) {
	metric = "log_monitor"
	expr = fmt.Sprintf("increase(node_log_monitor_count_total{file=\"%s\",keyword=\"%s\",instance=\"$address\"}[%s])", path, keyword, last)
	content = fmt.Sprintf("{{$labels.instance}} log alarm , log file: %s, keyword: %s , appear {{$value}} times in past %s", path, keyword, cond)
	return metric,expr,content
}


