package k8smonitor

import (
	"context"
	"errors"
	"strconv"
	"time"

	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h Handler) LogSamples(c *gin.Context) {
	var input struct {
		k8s.MetricObjectFilter
		ClusterID string `json:"clusterId"`
		Namespace string `json:"namespace"`
		Start     int64  `json:"start"`
		End       int64  `json:"end"`
	}
	if c.ShouldBindJSON(&input) != nil || input.End > time.Now().Unix()+60 {
		mid.ReturnValidateError(c, "invalid log sample range")
		return
	}
	query, err := k8s.LogSampleQuery(input.ClusterID, input.Namespace, input.MetricObjectFilter, input.Start, input.End)
	if err != nil {
		mid.ReturnValidateError(c, err.Error())
		return
	}
	if h.Policy.Validate(k8s.Registration{ClusterID: input.ClusterID, Namespace: input.Namespace}) != nil {
		mid.ReturnValidateError(c, "invalid log sample scope")
		return
	}
	rows, err := h.Store.List(c.Request.Context())
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("cannot validate log sample scope"))
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
	queryInstant := h.QueryInstant
	if queryInstant == nil {
		queryInstant = datasource.QueryPrometheusInstantRawContext
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	raw, err := queryInstant(ctx, query, strconv.FormatInt(input.End, 10))
	if err != nil {
		log.Error(c, log.LOGGER_APP, "Query Kubernetes log samples failed", zap.String("cluster", input.ClusterID), zap.String("namespace", input.Namespace), zap.String("pod", input.PodName), zap.String("container", input.Container), zap.Error(err))
		mid.ReturnServerHandleError(c, errors.New("log samples unavailable"))
		return
	}
	page, err := k8s.ReadLogSamples(raw, input.Start, input.End)
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("invalid log sample response"))
		return
	}
	mid.ReturnSuccessData(c, page)
}
