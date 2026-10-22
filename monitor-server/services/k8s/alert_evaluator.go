package k8s

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
)

type MetricReader interface {
	Read(context.Context, AlertRule, Container) (*float64, string, bool)
}
type MetricReadFunc func(context.Context, AlertRule, Container) (*float64, string, bool)

func (f MetricReadFunc) Read(ctx context.Context, r AlertRule, c Container) (*float64, string, bool) {
	return f(ctx, r, c)
}

type AlertSink interface {
	EvaluateAndPersist(context.Context, AlertEvaluation, time.Time) (Transition, error)
}

func EvaluateSnapshot(ctx context.Context, snapshot *Snapshot, rules []AlertRule, reader MetricReader, sink AlertSink, now time.Time) error {
	if snapshot == nil || reader == nil || sink == nil {
		return errors.New("alert evaluation is not configured")
	}
	// Keep one namespace query per metric for this evaluation only. A new
	// evaluation must retry failures and observe new Prometheus samples.
	ctx = context.WithValue(ctx, metricQueryCacheKey{}, &metricQueryCache{results: map[string]map[string]interface{}{}})
	for _, container := range snapshot.Containers {
		for _, rule := range rules {
			if !rule.Enabled || rule.ID == "" || !ruleMatches(rule, container) {
				continue
			}
			var value *float64
			var uid string
			var valid bool
			if !snapshot.DiscoveryFailed {
				value, uid, valid = reader.Read(ctx, rule, container)
			}
			observation := Observation{Valid: valid && uid == container.PodUID, PodUID: container.PodUID, Value: value}
			if observation.Valid && value != nil && !math.IsNaN(*value) && !math.IsInf(*value, 0) {
				observation.Violated = compare(rule.Operator, *value, rule.Threshold)
			} else if observation.Valid {
				observation.Valid = false
			}
			if _, err := sink.EvaluateAndPersist(ctx, AlertEvaluation{Key: AlertKey{ClusterID: container.Key.ClusterID, Namespace: container.Key.Namespace, WorkloadType: container.Key.WorkloadType, WorkloadName: container.Key.WorkloadName, PodName: container.Key.PodName, Container: container.Key.ContainerName, RuleID: rule.ID}, Rule: rule, Observation: observation, Value: value}, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func ruleMatches(rule AlertRule, c Container) bool {
	return (rule.WorkloadType == "" || rule.WorkloadType == c.Key.WorkloadType) && (rule.WorkloadName == "" || rule.WorkloadName == c.Key.WorkloadName) && (rule.PodName == "" || rule.PodName == c.Key.PodName) && (rule.Container == "" || rule.Container == c.Key.ContainerName)
}

// MetricLabelsMatchContainer accepts both container-scoped samples and
// pod-scoped samples such as Ready and Phase. Pod UID is always required so a
// recreated Pod cannot reuse the previous instance's value.
func MetricLabelsMatchContainer(labels map[string]string, c Container) bool {
	if labels["pod"] != c.Key.PodName || labels["pod_uid"] != c.PodUID {
		return false
	}
	container := strings.TrimSpace(labels["container"])
	return container == "" || container == c.Key.ContainerName
}
func compare(op string, value, threshold float64) bool {
	switch op {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "=", "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}
