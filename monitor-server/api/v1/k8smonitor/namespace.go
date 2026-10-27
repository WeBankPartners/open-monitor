package k8smonitor

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Store interface {
	Save(context.Context, k8s.Registration) error
	List(context.Context) ([]k8s.NamespaceState, error)
}
type LogAgentStore interface {
	GetLogAgentConfig(context.Context, string, string) (k8s.LogAgentConfig, error)
	SaveLogAgentConfig(context.Context, string, string, k8s.LogAgentConfig) error
	UpsertLogAgentEndpoint(context.Context, k8s.LogAgentEndpoint) error
	SetLogAgentEndpoint(context.Context, k8s.LogAgentEndpoint, bool) error
	ListLiveLogAgentEndpoints(context.Context, string, time.Duration) ([]k8s.LogAgentEndpoint, error)
}

type LogAgentInventoryStore interface {
	ListLogAgentEndpoints(context.Context, string) ([]k8s.LogAgentEndpoint, error)
	ReleaseLogAgentEndpoint(context.Context, string, string) error
}
type AlertStateStore interface {
	ListStates(context.Context, string, string) ([]k8s.AlertStateView, error)
}
type Handler struct {
	ClusterMode  bool
	Seed         string
	Policy       k8s.Policy
	Store        Store
	Exporter     *k8s.Exporter
	LogStore     LogAgentStore
	AlertStore   AlertStateStore
	Rules        AlertRuleStore
	QueryRange   func(context.Context, string, int64, int64, int64) (*models.PrometheusData, error)
	QueryInstant func(context.Context, string, string) (map[string]interface{}, error)
}

var logAgentClusterID = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9]{0,62}[a-z0-9])?$`)
var logAgentNodeName = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$`)

// LogAgentConfig serves a node-scoped immutable configuration snapshot to the
// daemon. Authentication is deliberately separate from the user session API.
func (h Handler) LogAgentConfig(c *gin.Context) {
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	want := "Bearer " + strings.TrimSpace(h.Policy.LogAgentToken)
	if h.Policy.LogAgentToken == "" || subtle.ConstantTimeCompare([]byte(auth), []byte(want)) != 1 {
		log.Warn(c, log.LOGGER_APP, "Log agent config authentication failed", zap.String("cluster", c.Query("clusterId")), zap.String("node", c.Query("node")))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	cluster, node := strings.TrimSpace(c.Query("clusterId")), strings.TrimSpace(c.Query("node"))
	if !logAgentClusterID.MatchString(cluster) || !logAgentNodeName.MatchString(node) || len(node) > 253 || h.Policy.ValidateCluster(cluster) != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if h.LogStore == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	config, err := h.LogStore.GetLogAgentConfig(c.Request.Context(), cluster, node)
	if errors.Is(err, sql.ErrNoRows) {
		log.Debug(c, log.LOGGER_APP, "Log agent config is not published", zap.String("cluster", cluster), zap.String("node", node))
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	if err != nil {
		log.Error(c, log.LOGGER_APP, "Log agent config query failed", zap.String("cluster", cluster), zap.String("node", node), zap.Error(err))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": config})
	log.Debug(c, log.LOGGER_APP, "Log agent config served", zap.String("cluster", cluster), zap.String("node", node), zap.Uint64("version", config.Version), zap.Int("target_count", len(config.Targets)))
}

func (h Handler) List(c *gin.Context) {
	rows, err := h.Store.List(c)
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("namespace list failed"))
		return
	}
	filtered := []map[string]interface{}{}
	for _, row := range rows {
		if (!h.ClusterMode && h.Policy.Validate(row.Registration) == nil) || (h.ClusterMode && h.Policy.ValidateCluster(row.ClusterID) == nil) {
			view := map[string]interface{}{"clusterId": row.ClusterID, "namespace": row.Namespace, "enabled": row.Enabled, "state": row.State, "hasToken": row.Credentials.EncryptedToken != ""}
			if row.ObservedAt != nil {
				view["observedAt"] = row.ObservedAt.Unix()
			}
			if row.Snapshot != nil {
				view["containerCount"] = len(row.Snapshot.Containers)
			}
			if h.ClusterMode {
				view["apiServer"] = row.Credentials.APIServer
				view["ca"] = row.Credentials.CA
			}
			filtered = append(filtered, view)
		}
	}
	mid.ReturnSuccessData(c, filtered)
}

func (h Handler) Metrics(c *gin.Context) {
	if h.Exporter == nil {
		c.String(503, "metrics unavailable\n")
		return
	}
	h.Exporter.ServeHTTP(c.Writer, c.Request)
}

func (h Handler) AlertStates(c *gin.Context) {
	cluster, namespace := strings.TrimSpace(c.Query("clusterId")), strings.TrimSpace(c.Query("namespace"))
	scope := k8s.Registration{ClusterID: cluster, Namespace: namespace}
	if h.Policy.Validate(scope) != nil || h.AlertStore == nil {
		mid.ReturnValidateError(c, "invalid alert scope")
		return
	}
	rows, err := h.Store.List(c)
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("alert scope lookup failed"))
		return
	}
	allowed := false
	for _, row := range rows {
		if row.Enabled && row.ClusterID == cluster && row.Namespace == namespace {
			allowed = true
			break
		}
	}
	if !allowed {
		mid.ReturnValidateError(c, "namespace is not enabled")
		return
	}
	states, err := h.AlertStore.ListStates(c, cluster, namespace)
	if err != nil {
		mid.ReturnServerHandleError(c, errors.New("alert state query failed"))
		return
	}
	mid.ReturnSuccessData(c, states)
}

// InternalMetrics trusts only the socket peer, never forwarded headers.
func (h Handler) InternalMetrics(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || !ip.IsLoopback() {
		c.AbortWithStatus(403)
		return
	}
	h.Metrics(c)
}
