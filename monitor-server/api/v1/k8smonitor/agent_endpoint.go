package k8smonitor

import (
	"context"
	"encoding/json"
	"errors"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var agentTargetSync sync.Mutex

func (h Handler) RegisterLogAgent(c *gin.Context) {
	if strings.TrimSpace(c.GetHeader("Authorization")) != "Bearer "+strings.TrimSpace(h.Policy.LogAgentToken) || h.Policy.LogAgentToken == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var endpoint k8s.LogAgentEndpoint
	if json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)).Decode(&endpoint) != nil || h.Policy.ValidateCluster(endpoint.ClusterID) != nil || h.LogStore == nil || k8s.ValidateLogAgentEndpoint(endpoint) != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// Agent registration is authoritative for automatic discovery. A manual
	// endpoint override is applied only through the administrator API below.
	if err := h.LogStore.SetLogAgentEndpoint(c.Request.Context(), endpoint, false); err != nil {
		log.Error(c, log.LOGGER_APP, "Log agent registration persistence failed", zap.String("cluster", endpoint.ClusterID), zap.String("node", endpoint.Node), zap.Error(err))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	if err := h.syncLogAgentTargets(c, endpoint.ClusterID); err != nil {
		log.Error(c, log.LOGGER_APP, "Log agent target file sync failed", zap.String("cluster", endpoint.ClusterID), zap.String("node", endpoint.Node), zap.Error(err))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
	log.Debug(c, log.LOGGER_APP, "Log agent heartbeat registered", zap.String("cluster", endpoint.ClusterID), zap.String("node", endpoint.Node), zap.String("address", endpoint.Address), zap.Int("port", endpoint.Port))
}

func (h Handler) ListLogAgents(c *gin.Context) {
	cluster := strings.TrimSpace(c.Query("clusterId"))
	if h.Policy.ValidateCluster(cluster) != nil || h.LogStore == nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var rows []k8s.LogAgentEndpoint
	var err error
	if inventory, ok := h.LogStore.(LogAgentInventoryStore); ok {
		rows, err = inventory.ListLogAgentEndpoints(c.Request.Context(), cluster)
		now := time.Now().Unix()
		for i := range rows {
			rows[i].Online = rows[i].Enabled && rows[i].LastSeen > 0 && now-rows[i].LastSeen <= int64((2*time.Minute)/time.Second)
			if config, configErr := h.LogStore.GetLogAgentConfig(c.Request.Context(), cluster, rows[i].Node); configErr == nil {
				rows[i].ConfigVersion = config.Version
				rows[i].TargetCount = len(config.Targets)
			}
		}
	} else {
		rows, err = h.LogStore.ListLiveLogAgentEndpoints(c.Request.Context(), cluster, 2*time.Minute)
		for i := range rows {
			rows[i].Online = true
		}
	}
	if err != nil {
		log.Error(c, log.LOGGER_APP, "Log agent inventory query failed", zap.String("cluster", cluster), zap.Error(err))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	mid.ReturnSuccessData(c, gin.H{"agents": rows})
}

func (h Handler) ReleaseLogAgent(c *gin.Context) {
	cluster, node := strings.TrimSpace(c.Query("clusterId")), strings.TrimSpace(c.Query("node"))
	inventory, ok := h.LogStore.(LogAgentInventoryStore)
	if !ok || h.Policy.ValidateCluster(cluster) != nil || !logAgentNodeName.MatchString(node) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := inventory.ReleaseLogAgentEndpoint(c.Request.Context(), cluster, node); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	if err := h.syncLogAgentTargets(c, cluster); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	mid.ReturnSuccessData(c, gin.H{"status": "ok"})
}

func (h Handler) UpdateLogAgent(c *gin.Context) {
	if h.LogStore == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var input struct {
		k8s.LogAgentEndpoint
		ManualOverride *bool `json:"manualOverride"`
	}
	if json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)).Decode(&input) != nil || h.Policy.ValidateCluster(input.ClusterID) != nil || k8s.ValidateLogAgentEndpoint(input.LogAgentEndpoint) != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	manual := true
	if input.ManualOverride != nil {
		manual = *input.ManualOverride
	}
	var updateErr error
	if !manual {
		if inventory, ok := h.LogStore.(LogAgentInventoryStore); ok {
			updateErr = inventory.ReleaseLogAgentEndpoint(c.Request.Context(), input.ClusterID, input.Node)
		} else {
			updateErr = h.LogStore.SetLogAgentEndpoint(c.Request.Context(), input.LogAgentEndpoint, false)
		}
	} else {
		updateErr = h.LogStore.SetLogAgentEndpoint(c.Request.Context(), input.LogAgentEndpoint, true)
	}
	if updateErr != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	if err := h.syncLogAgentTargets(c, input.ClusterID); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	mid.ReturnSuccessData(c, gin.H{"status": "ok"})
}

func (h Handler) syncLogAgentTargets(c *gin.Context, cluster string) error {
	if h.LogStore == nil {
		return errors.New("log agent store is unavailable")
	}
	agentTargetSync.Lock()
	defer agentTargetSync.Unlock()
	if !h.Policy.LocalCollectionActive() {
		return prom.WriteLogAgentTargets(models.Config().Prometheus.SdConfigPath, nil)
	}
	var rows []k8s.LogAgentEndpoint
	var err error
	if all, ok := h.LogStore.(interface {
		ListLiveLogAgentEndpointsAll(context.Context, time.Duration) ([]k8s.LogAgentEndpoint, error)
	}); ok {
		rows, err = all.ListLiveLogAgentEndpointsAll(c.Request.Context(), 2*time.Minute)
	} else {
		rows, err = h.LogStore.ListLiveLogAgentEndpoints(c.Request.Context(), cluster, 2*time.Minute)
	}
	if err != nil {
		log.Error(c, log.LOGGER_APP, "Log agent target file sync failed", zap.String("cluster", cluster), zap.Error(err))
		return err
	}
	addresses := make([]string, 0, len(rows))
	for _, row := range rows {
		addresses = append(addresses, net.JoinHostPort(row.Address, strconv.Itoa(row.Port)))
	}
	if err = prom.WriteLogAgentTargets(models.Config().Prometheus.SdConfigPath, addresses); err != nil {
		return err
	}
	// file_sd refreshes automatically; agent heartbeats do not reload Prometheus.
	return nil
}

func (h Handler) DeleteLogAgent(c *gin.Context) {
	cluster, node := c.Query("clusterId"), c.Query("node")
	store, ok := h.LogStore.(interface {
		DisableLogAgentEndpoint(context.Context, string, string) error
	})
	if !ok || h.Policy.ValidateCluster(cluster) != nil || !logAgentNodeName.MatchString(node) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := store.DisableLogAgentEndpoint(c.Request.Context(), cluster, node); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	if err := h.syncLogAgentTargets(c, cluster); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	mid.ReturnSuccess(c)
}
