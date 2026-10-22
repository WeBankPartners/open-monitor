package k8smonitor

import (
	"database/sql"
	"encoding/json"
	"errors"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

// GetLogAgentSource exposes the node configuration to the authenticated Web UI.
func (h Handler) GetLogAgentSource(c *gin.Context) {
	cluster, node := strings.TrimSpace(c.Query("clusterId")), strings.TrimSpace(c.Query("node"))
	if h.LogStore == nil || h.Policy.ValidateCluster(cluster) != nil || node == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	config, err := h.LogStore.GetLogAgentConfig(c.Request.Context(), cluster, node)
	if errors.Is(err, sql.ErrNoRows) {
		mid.ReturnSuccessData(c, gin.H{"clusterId": cluster, "node": node, "version": 0, "targets": []k8s.LogTarget{}})
		return
	}
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	mid.ReturnSuccessData(c, gin.H{"clusterId": cluster, "node": node, "version": config.Version, "targets": config.Targets})
}

func (h Handler) PutLogAgentSource(c *gin.Context) {
	var input struct {
		ClusterID string          `json:"clusterId"`
		Node      string          `json:"node"`
		Targets   []k8s.LogTarget `json:"targets"`
	}
	if json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 4<<20)).Decode(&input) != nil || h.LogStore == nil || h.Policy.ValidateCluster(strings.TrimSpace(input.ClusterID)) != nil || strings.TrimSpace(input.Node) == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	current, err := h.LogStore.GetLogAgentConfig(c.Request.Context(), input.ClusterID, input.Node)
	if errors.Is(err, sql.ErrNoRows) {
		current.Version = 0
	} else if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	config := k8s.LogAgentConfig{Version: current.Version + 1, Targets: input.Targets}
	if err = h.LogStore.SaveLogAgentConfig(c.Request.Context(), input.ClusterID, input.Node, config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mid.ReturnSuccessData(c, gin.H{"version": config.Version})
}

func (h Handler) DeleteLogAgentSource(c *gin.Context) {
	var input struct {
		ClusterID string `json:"clusterId"`
		Node      string `json:"node"`
	}
	if c.Request.ContentLength > 0 {
		if json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)).Decode(&input) != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		input.ClusterID, input.Node = c.Query("clusterId"), c.Query("node")
	}
	if h.LogStore == nil || h.Policy.ValidateCluster(strings.TrimSpace(input.ClusterID)) != nil || strings.TrimSpace(input.Node) == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	current, err := h.LogStore.GetLogAgentConfig(c.Request.Context(), input.ClusterID, input.Node)
	if errors.Is(err, sql.ErrNoRows) {
		mid.ReturnSuccessData(c, gin.H{"status": "ok"})
		return
	}
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	current.Version++
	current.Targets = nil
	if err = h.LogStore.SaveLogAgentConfig(c.Request.Context(), input.ClusterID, input.Node, current); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	mid.ReturnSuccessData(c, gin.H{"status": "ok"})
}

type logAgentInput struct {
	Guid              string          `json:"guid"`
	CallbackParameter string          `json:"callbackParameter"`
	ClusterID         string          `json:"clusterId"`
	Node              string          `json:"node"`
	Version           string          `json:"version"`
	Config            json.RawMessage `json:"config"`
}

// LogAgentPlugin follows the WeCube batch envelope and stores complete,
// node-scoped configuration snapshots. The caller supplies the already
// resolved directory and glob; the server still validates path safety.
func (h Handler) LogAgentPlugin(c *gin.Context) {
	result := struct {
		ResultCode    string `json:"resultCode"`
		ResultMessage string `json:"resultMessage"`
		Results       struct {
			Outputs []map[string]string `json:"outputs"`
		} `json:"results"`
	}{ResultCode: "0", ResultMessage: "OK"}
	defer func() { c.JSON(http.StatusOK, result) }()
	var request struct {
		RequestID string          `json:"requestId"`
		Inputs    []logAgentInput `json:"inputs"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 4<<20))
	if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF || request.RequestID == "" || len(request.Inputs) == 0 || len(request.Inputs) > 100 {
		result.ResultCode, result.ResultMessage = "1", "invalid plugin request"
		return
	}
	action := c.Param("action")
	if action != "add" && action != "update" && action != "delete" {
		result.ResultCode, result.ResultMessage = "1", "unsupported action"
		return
	}
	for _, in := range request.Inputs {
		out := map[string]string{"guid": in.Guid, "callbackParameter": in.CallbackParameter, "errorCode": "0"}
		version, err := strconv.ParseUint(in.Version, 10, 64)
		if err != nil || in.Guid == "" || in.Node == "" || h.Policy.ValidateCluster(in.ClusterID) != nil || (action == "add" && version != 0) || (action != "add" && version == 0) {
			err = errors.New("invalid log agent scope or version")
		} else if action == "delete" {
			var current k8s.LogAgentConfig
			current, err = h.LogStore.GetLogAgentConfig(c.Request.Context(), in.ClusterID, in.Node)
			if err == nil {
				current.Version++
				current.Targets = nil
				err = h.LogStore.SaveLogAgentConfig(c.Request.Context(), in.ClusterID, in.Node, current)
			}
		} else {
			var config k8s.LogAgentConfig
			err = json.Unmarshal(in.Config, &config)
			if err == nil {
				config.Version = version + 1
				err = h.LogStore.SaveLogAgentConfig(c.Request.Context(), in.ClusterID, in.Node, config)
			}
		}
		if err != nil {
			out["errorCode"], out["errorMessage"] = "1", "log agent configuration failed"
			result.ResultCode, result.ResultMessage = "1", "one or more inputs failed"
		} else {
			out["version"] = strconv.FormatUint(version+1, 10)
		}
		result.Results.Outputs = append(result.Results.Outputs, out)
	}
}
