package k8smonitor

import (
	"context"
	"errors"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"time"
)

func (h Handler) History(c *gin.Context) {
	var input struct {
		k8s.MetricObjectFilter
		ClusterID string `json:"clusterId"`
		Namespace string `json:"namespace"`
		Metric    string `json:"metric"`
		Start     int64  `json:"start"`
		End       int64  `json:"end"`
		Step      int64  `json:"step"`
	}
	if c.ShouldBindJSON(&input) != nil || input.Start < 0 || input.End <= input.Start || input.End > time.Now().Unix()+60 || input.End-input.Start > 30*86400 || input.Step < 10 || (input.End-input.Start)/input.Step > 11000 {
		mid.ReturnValidateError(c, "invalid metric range")
		return
	}
	scope := k8s.Registration{ClusterID: input.ClusterID, Namespace: input.Namespace}
	if h.Policy.Validate(scope) != nil {
		mid.ReturnValidateError(c, "invalid metric scope")
		return
	}
	rows, err := h.Store.List(c.Request.Context())
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("cannot validate metric scope"))
		return
	}
	allowed := false
	for _, row := range rows {
		if row.Enabled && row.ClusterID == input.ClusterID && row.Namespace == input.Namespace {
			allowed = true
			break
		}
	}
	if !allowed {
		mid.ReturnValidateError(c, "namespace is not enabled")
		return
	}
	query, err := k8s.ObjectMetricQuery(input.ClusterID, input.Namespace, input.Metric, input.MetricObjectFilter, h.Policy.MetricsInterval())
	if err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	queryRange := h.QueryRange
	if queryRange == nil {
		queryRange = datasource.QueryPrometheusRangeContext
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	data, err := queryRange(ctx, query, input.Start, input.End, input.Step)
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("metric history unavailable"))
		return
	}
	mid.ReturnSuccessData(c, data)
}

func (h Handler) MetricCatalog(c *gin.Context) {
	seconds := int(h.Policy.MetricsInterval() / time.Second)
	window := 4 * seconds
	if window < 300 {
		window = 300
	}
	mid.ReturnSuccessData(c, gin.H{"metrics": k8s.MetricCatalog(), "scrapeIntervalSeconds": seconds, "countWindowSeconds": window})
}
