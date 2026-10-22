package k8smonitor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/k8s"
	"github.com/gin-gonic/gin"
)

type AlertRuleStore interface {
	ListAlertRuleRecords(context.Context, string, string) ([]k8s.AlertRuleRecord, error)
	CreateAlertRule(context.Context, k8s.AlertRuleRecord) error
	UpdateAlertRule(context.Context, k8s.AlertRuleRecord, uint64) error
	DeleteAlertRule(context.Context, string, string, string, uint64) error
}

type alertRuleInput struct {
	ClusterID string `json:"clusterId"`
	Namespace string `json:"namespace"`
	k8s.AlertRule
	Version uint64 `json:"version"`
}

func (h Handler) ListAlertRules(c *gin.Context) {
	if h.Rules == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	cluster, namespace := strings.TrimSpace(c.Query("clusterId")), strings.TrimSpace(c.Query("namespace"))
	if err := h.validateAlertScope(c, cluster, namespace); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	rows, err := h.Rules.ListAlertRuleRecords(c.Request.Context(), cluster, namespace)
	if err != nil {
		middleware.ReturnServerHandleError(c, errors.New("alert rule list failed"))
		return
	}
	middleware.ReturnSuccessData(c, rows)
}

func (h Handler) CreateAlertRule(c *gin.Context) {
	var input alertRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.ReturnValidateError(c, "invalid alert rule")
		return
	}
	if err := h.validateAlertScope(c, input.ClusterID, input.Namespace); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := k8s.ValidateAlertRule(input.AlertRule); err != nil || input.Version != 0 {
		if err == nil {
			err = errors.New("new alert rule version must be zero")
		}
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := h.Rules.CreateAlertRule(c.Request.Context(), k8s.AlertRuleRecord{ClusterID: input.ClusterID, Namespace: input.Namespace, AlertRule: input.AlertRule, Version: 1}); err != nil {
		if errors.Is(err, k8s.ErrConflict) {
			middleware.ReturnValidateError(c, "alert rule already exists")
			return
		}
		middleware.ReturnServerHandleError(c, errors.New("alert rule create failed"))
		return
	}
	middleware.ReturnSuccessData(c, map[string]any{"ruleId": input.ID, "version": 1})
}

func (h Handler) UpdateAlertRule(c *gin.Context) {
	var input alertRuleInput
	if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.Version == 0 {
		middleware.ReturnValidateError(c, "rule id and version are required")
		return
	}
	if routeID := strings.TrimSpace(c.Param("ruleId")); routeID != "" && routeID != input.ID {
		middleware.ReturnValidateError(c, "rule id does not match path")
		return
	}
	if err := h.validateAlertScope(c, input.ClusterID, input.Namespace); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := k8s.ValidateAlertRule(input.AlertRule); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := h.Rules.UpdateAlertRule(c.Request.Context(), k8s.AlertRuleRecord{ClusterID: input.ClusterID, Namespace: input.Namespace, AlertRule: input.AlertRule}, input.Version); err != nil {
		if errors.Is(err, k8s.ErrConflict) {
			middleware.ReturnValidateError(c, "alert rule version conflict")
			return
		}
		middleware.ReturnServerHandleError(c, errors.New("alert rule update failed"))
		return
	}
	middleware.ReturnSuccessData(c, map[string]any{"ruleId": input.ID, "version": input.Version + 1})
}

func (h Handler) DeleteAlertRule(c *gin.Context) {
	var input struct {
		ClusterID string `json:"clusterId"`
		Namespace string `json:"namespace"`
		Version   uint64 `json:"version"`
	}
	if c.Request.ContentLength == 0 {
		input.ClusterID, input.Namespace = c.Query("clusterId"), c.Query("namespace")
		input.Version, _ = strconv.ParseUint(c.Query("version"), 10, 64)
	} else if err := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)).Decode(&input); err != nil {
		middleware.ReturnValidateError(c, "invalid delete request")
		return
	}
	if input.Version == 0 {
		middleware.ReturnValidateError(c, "scope and version are required")
		return
	}
	if err := h.validateAlertScope(c, input.ClusterID, input.Namespace); err != nil {
		middleware.ReturnValidateError(c, err.Error())
		return
	}
	if err := h.Rules.DeleteAlertRule(c.Request.Context(), input.ClusterID, input.Namespace, c.Param("ruleId"), input.Version); err != nil {
		if errors.Is(err, k8s.ErrConflict) {
			middleware.ReturnValidateError(c, "alert rule version conflict")
			return
		}
		middleware.ReturnServerHandleError(c, errors.New("alert rule delete failed"))
		return
	}
	middleware.ReturnSuccess(c)
}

func (h Handler) validateAlertScope(c *gin.Context, cluster, namespace string) error {
	scope := k8s.Registration{ClusterID: strings.TrimSpace(cluster), Namespace: strings.TrimSpace(namespace)}
	if h.Policy.Validate(scope) != nil {
		return errors.New("invalid alert scope")
	}
	rows, err := h.Store.List(c.Request.Context())
	if err != nil {
		return errors.New("alert scope lookup failed")
	}
	for _, row := range rows {
		if row.Enabled && row.ClusterID == scope.ClusterID && row.Namespace == scope.Namespace {
			return nil
		}
	}
	return errors.New("namespace is not enabled")
}
