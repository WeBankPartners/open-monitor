package alarm

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/pcre"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// @Summary 日志告警配置接口 : 获取列表
// @Description 获取配置好的对象或组的日志告警列表
// @Produce  json
// @Param type query string true "类型，区分是单个对象还是组，枚举endpoint、grp"
// @Param id query int true "对象或组的id"
// @Router /api/v1/alarm/log/monitor/list [get]
func ListLogTpl(c *gin.Context) {
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
	err := db.ListLogMonitorNew(&query)
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
		return
	}
	mid.ReturnSuccessData(c, query.Tpl)
}

// @Summary 日志告警配置接口 : 新增
// @Produce  json
// @Param grp_id query int false "组Id，和对象id二选一"
// @Param endpoint_id query int false "对象Id，和组id二选一"
// @Param path query string true "表单输入的日志路径"
// @Param strategy query string true "对象数组类型[{'keyword':'关键字','cond':'条件,如 >1','last':'时间范围,如 5min','priority':'优先级,如 high'}]"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/add [post]
func AddLogStrategy(c *gin.Context) {
	var param m.LogMonitorDto
	if err := c.ShouldBindJSON(&param); err == nil {
		if len(param.Strategy) == 0 {
			mid.ReturnParamEmptyError(c, "strategy")
			return
		}
		if param.EndpointId <= 0 {
			mid.ReturnParamEmptyError(c, "endpoint_id")
			return
		}
		if !mid.IsIllegalPath(param.Path) {
			mid.ReturnValidateError(c, "path illegal")
			return
		}
		_, regErr := pcre.Compile(param.Strategy[0].Keyword, 0)
		if regErr != nil {
			mid.ReturnValidateError(c, "keyword validate fail")
			return
		}
		var logMonitorObj m.LogMonitorTable
		logMonitorObj.Path = param.Path
		logMonitorObj.Keyword = param.Strategy[0].Keyword
		logMonitorObj.Priority = param.Strategy[0].Priority
		if logMonitorObj.Priority == "" {
			mid.ReturnParamEmptyError(c, "priority")
			return
		}
		logMonitorObj.NotifyEnable = param.Strategy[0].NotifyEnable
		logMonitorObj.OwnerEndpoint = param.OwnerEndpoint
		if param.Id <= 0 {
			_, lms := db.GetLogMonitorTable(0, param.EndpointId, 0, "")
			for _, v := range lms {
				if v.Path == param.Path {
					mid.ReturnValidateError(c, "path exists")
					return
				}
			}
		}
		logMonitorObj.StrategyId = param.EndpointId
		err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor: []*m.LogMonitorTable{&logMonitorObj}, Operation: "insert"})
		if err != nil {
			mid.ReturnUpdateTableError(c, "log_monitor", err)
			return
		}
		// Call endpoint node exporter
		err = db.SendLogConfig(param.EndpointId, param.GrpId, param.TplId)
		if err != nil {
			mid.ReturnHandleError(c, "send log config to endpoint failed", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

// @Summary 日志告警配置接口 : 修改日志路径
// @Produce  json
// @Param id query int true "列表获取中的id"
// @Param tpl_id query int true "列表获取中的tpl_id"
// @Param path query string true "新的日志路径"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/update_path [post]
func EditLogPath(c *gin.Context) {
	var param m.LogMonitorDto
	if err := c.ShouldBindJSON(&param); err == nil {
		if param.Id <= 0 {
			mid.ReturnParamEmptyError(c, "id")
			return
		}
		if !mid.IsIllegalPath(param.Path) {
			mid.ReturnValidateError(c, "path illegal")
			return
		}
		err, lms := db.GetLogMonitorTable(param.Id, 0, 0, "")
		if err != nil || len(lms) == 0 {
			mid.ReturnFetchDataError(c, "log_monitor", "id", strconv.Itoa(param.Id))
			return
		}
		oldPath := lms[0].Path
		_, lmsGrp := db.GetLogMonitorTable(0, lms[0].StrategyId, 0, oldPath)
		//var strategyObjs []*m.StrategyTable
		// Update log_monitor
		for _, v := range lmsGrp {
			//strategyObjs = append(strategyObjs, &m.StrategyTable{Id:v.StrategyId})
			logMonitorObj := m.LogMonitorTable{Id: v.Id, StrategyId: v.StrategyId, Path: param.Path, Keyword: v.Keyword, NotifyEnable: v.NotifyEnable, OwnerEndpoint: param.OwnerEndpoint, Priority: v.Priority}
			err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor: []*m.LogMonitorTable{&logMonitorObj}, Operation: "update"})
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "Update log monitor alert failed", zap.Error(err))
			}
		}
		err = db.SendLogConfig(lms[0].StrategyId, param.GrpId, param.TplId)
		if err != nil {
			mid.ReturnHandleError(c, "send log config to endpoint fail", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

// @Summary 日志告警配置接口 : 修改
// @Produce  json
// @Param tpl_id query int true "列表获取中的tpl_id"
// @Param path query string true "表单输入的日志路径"
// @Param strategy query string true "对象数组类型[{'id':int类型, 'strategy_id':int类型,'keyword':'关键字','cond':'条件,如 >1','last':'时间范围,如 5min','priority':'优先级,如 high'}]"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/update [post]
func EditLogStrategy(c *gin.Context) {
	var param m.LogMonitorDto
	if err := c.ShouldBindJSON(&param); err == nil {
		if len(param.Strategy) == 0 {
			mid.ReturnParamEmptyError(c, "strategy")
			return
		}
		if param.Strategy[0].Priority == "" {
			mid.ReturnParamEmptyError(c, "priority")
			return
		}
		_, regErr := pcre.Compile(param.Strategy[0].Keyword, 0)
		if regErr != nil {
			mid.ReturnValidateError(c, "keyword is illegal")
			return
		}
		// Update log_monitor
		logMonitorObj := m.LogMonitorTable{Id: param.Strategy[0].Id, StrategyId: param.EndpointId, Path: param.Path, Keyword: param.Strategy[0].Keyword, Priority: param.Strategy[0].Priority, NotifyEnable: param.Strategy[0].NotifyEnable, OwnerEndpoint: param.OwnerEndpoint}
		err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor: []*m.LogMonitorTable{&logMonitorObj}, Operation: "update"})
		if err != nil {
			mid.ReturnUpdateTableError(c, "log_monitor", err)
			return
		}
		// Call endpoint node exporter
		err, tplObj := db.GetTpl(param.TplId, 0, 0)
		if err != nil {
			mid.ReturnFetchDataError(c, "tpl", "id", strconv.Itoa(param.TplId))
			return
		}
		param.EndpointId = tplObj.EndpointId
		param.GrpId = tplObj.GrpId
		err = db.SendLogConfig(param.EndpointId, param.GrpId, param.TplId)
		if err != nil {
			mid.ReturnHandleError(c, "send log config to endpoint failed", err)
			return
		}
		mid.ReturnSuccess(c)
	} else {
		mid.ReturnValidateError(c, err.Error())
	}
}

// @Summary 日志告警配置接口 : 删除
// @Produce  json
// @Param id query int true "strategy_id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/delete_path [get]
func DeleteLogPath(c *gin.Context) {
	logMonitorId, err := strconv.Atoi(c.Query("id"))
	if err != nil || logMonitorId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	//err,strategyObj := db.GetStrategy(m.StrategyTable{Id:strategyId})
	//if err != nil || strategyObj.TplId <= 0 {
	//	mid.ReturnFetchDataError(c, "strategy", "id", strconv.Itoa(strategyId))
	//	return
	//}
	err, lms := db.GetLogMonitorTable(logMonitorId, 0, 0, "")
	if err != nil || len(lms) == 0 {
		mid.ReturnFetchDataError(c, "log_monitor", "id", strconv.Itoa(logMonitorId))
		return
	}
	oldPath := lms[0].Path
	_, lmsGrp := db.GetLogMonitorTable(0, lms[0].StrategyId, 0, oldPath)
	//var strategyObjs []*m.StrategyTable
	// Delete log monitor
	for _, v := range lmsGrp {
		//strategyObjs = append(strategyObjs, &m.StrategyTable{Id:v.StrategyId})
		err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor: []*m.LogMonitorTable{&m.LogMonitorTable{Id: v.Id}}, Operation: "delete"})
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Delete log monitor alert failed", zap.Error(err))
		}
	}
	// Delete strategy
	//for _,v := range strategyObjs {
	//	err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&m.StrategyTable{Id:v.Id}}, Operation:"delete"})
	//	if err != nil {
	//		log.Error(nil, log.LOGGER_APP, "Delete strategy failed", zap.Error(err))
	//	}
	//}
	// Call endpoint node exporter
	//err,tplObj := db.GetTpl(strategyObj.TplId, 0, 0)
	//if err != nil {
	//	mid.ReturnFetchDataError(c, "tpl", "id", strconv.Itoa(strategyObj.TplId))
	//	return
	//}
	err = db.SendLogConfig(lms[0].StrategyId, 0, 0)
	if err != nil {
		mid.ReturnHandleError(c, "send log config to endpoint failed", err)
		return
	}
	// Save Prometheus rule file
	//err = SaveConfigFile(tplObj.Id, false)
	//if err != nil {
	//	mid.ReturnHandleError(c, "save prometheus rule file failed", err)
	//	return
	//}
	mid.ReturnSuccess(c)
}

// @Summary 日志告警配置接口 : 删除
// @Produce  json
// @Param id query int true "id"
// @Success 200 {string} json "{"message": "Success"}"
// @Router /api/v1/alarm/log/monitor/delete [get]
func DeleteLogStrategy(c *gin.Context) {
	logMonitorId, err := strconv.Atoi(c.Query("id"))
	if err != nil || logMonitorId <= 0 {
		mid.ReturnParamTypeError(c, "id", "int")
		return
	}
	err, lms := db.GetLogMonitorTable(logMonitorId, 0, 0, "")
	if err != nil || len(lms) == 0 {
		mid.ReturnFetchDataError(c, "log_monitor", "id", strconv.Itoa(logMonitorId))
		return
	}
	//err,strategyObj := db.GetStrategy(m.StrategyTable{Id:lms[0].StrategyId})
	//if err != nil || strategyObj.TplId <= 0 {
	//	mid.ReturnFetchDataError(c, "strategy", "id", strconv.Itoa(lms[0].StrategyId))
	//	return
	//}
	// Delete log monitor
	err = db.UpdateLogMonitor(&m.UpdateLogMonitor{LogMonitor: []*m.LogMonitorTable{&m.LogMonitorTable{Id: logMonitorId}}, Operation: "delete"})
	if err != nil {
		mid.ReturnUpdateTableError(c, "log_monitor", err)
		return
	}
	// Delete strategy
	//err = db.UpdateStrategy(&m.UpdateStrategy{Strategy:[]*m.StrategyTable{&m.StrategyTable{Id:strategyObj.Id}}, Operation:"delete"})
	//if err != nil {
	//	mid.ReturnUpdateTableError(c, "strategy", err)
	//	return
	//}
	// Call endpoint node exporter
	//err,tplObj := db.GetTpl(strategyObj.TplId, 0, 0)
	//if err != nil {
	//	mid.ReturnFetchDataError(c, "tpl", "id", strconv.Itoa(strategyObj.TplId))
	//	return
	//}
	err = db.SendLogConfig(lms[0].StrategyId, 0, 0)
	if err != nil {
		mid.ReturnHandleError(c, "send log config to endpoint failed", err)
		return
	}
	// Save Prometheus rule file
	//err = SaveConfigFile(tplObj.Id, false)
	//if err != nil {
	//	mid.ReturnHandleError(c, "save prometheus rule file failed", err)
	//	return
	//}
	mid.ReturnSuccess(c)
}

func makeStrategyMsg(path, keyword, cond, last string) (metric, expr, content string) {
	metric = "log_monitor"
	expr = fmt.Sprintf("increase(node_log_monitor_count_total{file=\"%s\",keyword=\"%s\",instance=\"$address\"}[%s])", path, keyword, last)
	content = fmt.Sprintf("{{$labels.instance}} log alarm , log file: %s, keyword: %s , appear {{$value}} times in past %s", path, keyword, cond)
	return metric, expr, content
}
