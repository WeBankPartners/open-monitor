package k8s

import (
	"errors"
	"math"
	"time"
)

type AlertStatus string

const (
	AlertNormal    AlertStatus = "normal"
	AlertPending   AlertStatus = "pending"
	AlertFiring    AlertStatus = "firing"
	AlertRecovered AlertStatus = "recovered"
	AlertUnknown   AlertStatus = "unknown"
)

type AlertEvent string

const (
	AlertEventNone      AlertEvent = ""
	AlertEventFiring    AlertEvent = "firing"
	AlertEventRecovered AlertEvent = "recovered"
)

type AlertRule struct {
	WorkloadType     string  `json:"workloadType"`
	WorkloadName     string  `json:"workloadName"`
	PodName          string  `json:"podName"`
	Container        string  `json:"container"`
	ID               string  `json:"ruleId"`
	Metric           string  `json:"metric"`
	Operator         string  `json:"operator"`
	Threshold        float64 `json:"threshold"`
	Enabled          bool    `json:"enabled"`
	ForSeconds       int     `json:"forSeconds"`
	RecoverAfter     int     `json:"recoverAfter"`
	NotifyStrategyID int     `json:"notifyStrategyId"`
}

type AlertRuleRecord struct {
	AlertRule
	ClusterID string    `json:"clusterId"`
	Namespace string    `json:"namespace"`
	Version   uint64    `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ValidateAlertRule(r AlertRule) error {
	if r.ID == "" || len(r.ID) > 128 || r.Metric == "" || r.Operator == "" {
		return errors.New("rule id, metric and operator are required")
	}
	if r.ForSeconds < 0 || r.RecoverAfter < 0 || r.NotifyStrategyID < 0 {
		return errors.New("rule durations and strategy must be non-negative")
	}
	if math.IsNaN(r.Threshold) || math.IsInf(r.Threshold, 0) {
		return errors.New("threshold must be finite")
	}
	for _, field := range []struct {
		value string
		limit int
	}{{r.ID, 128}, {r.WorkloadType, 32}, {r.WorkloadName, 253}, {r.PodName, 253}, {r.Container, 253}} {
		if len(field.value) > field.limit {
			return errors.New("alert selector exceeds field length")
		}
		for _, ch := range field.value {
			if ch < 32 || ch > 126 {
				return errors.New("alert identifiers must be printable ASCII")
			}
		}
	}
	if r.Metric == "phase" || r.Metric == "quality" {
		return errors.New("metric cannot be used as a container threshold; use ready or container_ready for health alerts")
	}
	if r.Operator != ">" && r.Operator != ">=" && r.Operator != "<" && r.Operator != "<=" && r.Operator != "=" && r.Operator != "==" && r.Operator != "!=" {
		return errors.New("unsupported alert operator")
	}
	if _, err := MetricQuery("cluster-a", "namespace-a", r.Metric); err != nil {
		return err
	}
	return nil
}

type Observation struct {
	Valid    bool
	Violated bool
	PodUID   string
	Value    *float64
}

type AlertState struct {
	Status        AlertStatus
	UnknownFrom   AlertStatus
	PodUID        string
	ObservedAt    time.Time
	StartedAt     time.Time
	RecoveryCount int
}

type Transition struct {
	State AlertState
	Event AlertEvent
}

func Evaluate(previous AlertState, observation Observation, rule AlertRule, now time.Time) Transition {
	previous.ObservedAt = now
	if !observation.Valid {
		if previous.Status != AlertUnknown {
			previous.UnknownFrom = previous.Status
		}
		previous.Status = AlertUnknown
		previous.RecoveryCount = 0
		return Transition{State: previous}
	}
	wasUnknown := previous.Status == AlertUnknown
	if wasUnknown {
		previous.Status = previous.UnknownFrom
		previous.UnknownFrom = ""
		previous.RecoveryCount = 0
		if previous.Status == AlertFiring && !observation.Violated {
			return Transition{State: previous}
		}
	}
	previous.PodUID = observation.PodUID
	if observation.Violated {
		previous.RecoveryCount = 0
		if previous.Status == AlertFiring {
			return Transition{State: previous}
		}
		if previous.Status != AlertPending || previous.StartedAt.IsZero() {
			previous.Status = AlertPending
			previous.StartedAt = now
		}
		if rule.ForSeconds <= 0 || now.Sub(previous.StartedAt) >= time.Duration(rule.ForSeconds)*time.Second {
			previous.Status = AlertFiring
			return Transition{State: previous, Event: AlertEventFiring}
		}
		return Transition{State: previous}
	}
	if previous.Status != AlertFiring {
		previous.Status = AlertNormal
		previous.StartedAt = time.Time{}
		return Transition{State: previous}
	}
	previous.RecoveryCount++
	recoverAfter := rule.RecoverAfter
	if recoverAfter <= 0 {
		recoverAfter = 1
	}
	if previous.RecoveryCount >= recoverAfter {
		previous.Status = AlertRecovered
		return Transition{State: previous, Event: AlertEventRecovered}
	}
	return Transition{State: previous}
}
