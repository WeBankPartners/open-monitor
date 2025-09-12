package monitor

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/gin-gonic/gin"
	"strconv"
)

// PrometheusRaw 提供对 Prometheus 的简单透传查询
// 入参：query 必填；time 可选（存在则调用 /query）；start、end 同时存在则调用 /query_range，step 固定 10
func PrometheusRaw(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		middleware.ReturnParamEmptyError(c, "query")
		return
	}

	timeParam := c.Query("time")
	startStr := c.Query("start")
	endStr := c.Query("end")

	// 优先判断 instant query（传了 time 即认为是瞬时查询）
	if timeParam != "" {
		result, err := ds.QueryPrometheusInstantRaw(query, timeParam)
		if err != nil {
			middleware.ReturnHandleError(c, "prometheus query", err)
			return
		}
		middleware.ReturnSuccessData(c, result)
		return
	}

	// 其次判断 range query（start/end 同时存在）
	if startStr != "" && endStr != "" {
		startVal, err1 := strconv.ParseInt(startStr, 10, 64)
		endVal, err2 := strconv.ParseInt(endStr, 10, 64)
		if err1 != nil || err2 != nil {
			middleware.ReturnHandleError(c, "invalid start or end", fmt.Errorf("parse start/end error"))
			return
		}
		// step 固定 10
		result, err := ds.QueryPrometheusRangeRaw(query, startVal, endVal, 10)
		if err != nil {
			middleware.ReturnHandleError(c, "prometheus query_range", err)
			return
		}
		middleware.ReturnSuccessData(c, result)
		return
	}

	// 参数不完整
	middleware.ReturnValidateError(c, "must provide time or start&end")
}
