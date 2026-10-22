package k8s

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AlertKey struct {
	ClusterID, Namespace, WorkloadType, WorkloadName, PodName, Container, RuleID string
}

// Endpoint retains the complete logical object identity in the legacy alarm
// table. Very long Kubernetes names use a bounded hash for its VARCHAR(255).
func (k AlertKey) Endpoint() string {
	endpoint := strings.Join([]string{"k8s", k.ClusterID, k.Namespace, k.WorkloadType, k.WorkloadName, k.PodName, k.Container}, "/")
	if len(endpoint) <= 255 {
		return endpoint
	}
	k.RuleID = ""
	raw, _ := json.Marshal(k)
	sum := sha256.Sum256(raw)
	return "k8s-" + hex.EncodeToString(sum[:])
}

func OutboxIdempotencyKey(key AlertKey, stateVersion uint64, event AlertEvent) string {
	payload := struct {
		Key          AlertKey   `json:"key"`
		StateVersion uint64     `json:"state_version"`
		Event        AlertEvent `json:"event"`
	}{Key: key, StateVersion: stateVersion, Event: event}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

type AlertEvaluation struct {
	Key         AlertKey
	Rule        AlertRule
	Observation Observation
	Value       *float64
}

type OutboxEvent struct {
	EventID        string
	IdempotencyKey string
	Key            AlertKey
	Event          AlertEvent
	Payload        map[string]any
	Attempts       uint
	LeaseToken     string
}

type AlertStore struct{ DB *sql.DB }

func (s AlertStore) ListAlertRuleRecords(ctx context.Context, cluster, namespace string) ([]AlertRuleRecord, error) {
	if s.DB == nil {
		return nil, errors.New("alert store database is unavailable")
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT rule_id,cluster_id,namespace,workload_type,workload_name,pod_name,container_name,metric,operator,threshold,for_seconds,recover_after,notify_strategy_id,enabled,version,updated_at FROM k8s_alert_rule WHERE cluster_id=? AND namespace=? ORDER BY rule_id`, cluster, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AlertRuleRecord{}
	for rows.Next() {
		var r AlertRuleRecord
		if err = rows.Scan(&r.ID, &r.ClusterID, &r.Namespace, &r.WorkloadType, &r.WorkloadName, &r.PodName, &r.Container, &r.Metric, &r.Operator, &r.Threshold, &r.ForSeconds, &r.RecoverAfter, &r.NotifyStrategyID, &r.Enabled, &r.Version, &r.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s AlertStore) CreateAlertRule(ctx context.Context, r AlertRuleRecord) error {
	if s.DB == nil {
		return errors.New("alert store database is unavailable")
	}
	if err := ValidateAlertRule(r.AlertRule); err != nil {
		return err
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO k8s_alert_rule(rule_id,cluster_id,namespace,workload_type,workload_name,pod_name,container_name,metric,operator,threshold,for_seconds,recover_after,notify_strategy_id,enabled,version,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,CURRENT_TIMESTAMP(6))`, r.ID, r.ClusterID, r.Namespace, r.WorkloadType, r.WorkloadName, r.PodName, r.Container, r.Metric, r.Operator, r.Threshold, r.ForSeconds, r.RecoverAfter, r.NotifyStrategyID, r.Enabled)
	if err != nil {
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			return ErrConflict
		}
	}
	return err
}

func (s AlertStore) UpdateAlertRule(ctx context.Context, r AlertRuleRecord, version uint64) error {
	if s.DB == nil {
		return errors.New("alert store database is unavailable")
	}
	if err := ValidateAlertRule(r.AlertRule); err != nil {
		return err
	}
	result, err := s.DB.ExecContext(ctx, `UPDATE k8s_alert_rule SET workload_type=?,workload_name=?,pod_name=?,container_name=?,metric=?,operator=?,threshold=?,for_seconds=?,recover_after=?,notify_strategy_id=?,enabled=?,version=version+1,updated_at=CURRENT_TIMESTAMP(6) WHERE rule_id=? AND cluster_id=? AND namespace=? AND version=?`, r.WorkloadType, r.WorkloadName, r.PodName, r.Container, r.Metric, r.Operator, r.Threshold, r.ForSeconds, r.RecoverAfter, r.NotifyStrategyID, r.Enabled, r.ID, r.ClusterID, r.Namespace, version)
	return changed(result, err)
}

func (s AlertStore) DeleteAlertRule(ctx context.Context, cluster, namespace, ruleID string, version uint64) error {
	if s.DB == nil {
		return errors.New("alert store database is unavailable")
	}
	result, err := s.DB.ExecContext(ctx, `DELETE FROM k8s_alert_rule WHERE rule_id=? AND cluster_id=? AND namespace=? AND version=?`, ruleID, cluster, namespace, version)
	return changed(result, err)
}

type AlertStateView struct {
	Key           AlertKey
	Status        AlertStatus
	PodUID        string
	Value         *float64
	ObservedAt    time.Time
	StartedAt     time.Time
	RecoveryCount int
	UpdatedAt     time.Time
}

func (s AlertStore) ListStates(ctx context.Context, cluster, namespace string) ([]AlertStateView, error) {
	if s.DB == nil {
		return nil, errors.New("alert store database is unavailable")
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT rule_id,cluster_id,namespace,workload_type,workload_name,pod_name,container_name,status,pod_uid,value,observed_at,started_at,recovery_count,updated_at FROM k8s_alert_state WHERE cluster_id=? AND namespace=? ORDER BY updated_at DESC`, cluster, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AlertStateView{}
	for rows.Next() {
		var row AlertStateView
		var status string
		var value sql.NullFloat64
		var observed, started sql.NullTime
		if err = rows.Scan(&row.Key.RuleID, &row.Key.ClusterID, &row.Key.Namespace, &row.Key.WorkloadType, &row.Key.WorkloadName, &row.Key.PodName, &row.Key.Container, &status, &row.PodUID, &value, &observed, &started, &row.RecoveryCount, &row.UpdatedAt); err != nil {
			return nil, err
		}
		row.Status = AlertStatus(status)
		if value.Valid {
			v := value.Float64
			row.Value = &v
		}
		if observed.Valid {
			row.ObservedAt = observed.Time
		}
		if started.Valid {
			row.StartedAt = started.Time
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

var ErrLeaseLost = errors.New("alert outbox lease lost")

func (s AlertStore) EvaluateAndPersist(ctx context.Context, input AlertEvaluation, now time.Time) (Transition, error) {
	return s.evaluateAndPersist(ctx, input, now, true)
}

func (s AlertStore) evaluateAndPersist(ctx context.Context, input AlertEvaluation, now time.Time, retry bool) (Transition, error) {
	if s.DB == nil {
		return Transition{}, errors.New("alert store database is unavailable")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Transition{}, err
	}
	defer func() { _ = tx.Rollback() }()
	state := AlertState{}
	var status string
	var unknownFrom string
	var podUID string
	var started, observed sql.NullTime
	var recovery uint
	var version uint64
	err = tx.QueryRowContext(ctx, `SELECT status,unknown_from,pod_uid,started_at,observed_at,recovery_count,state_version
FROM k8s_alert_state WHERE rule_id=? AND cluster_id=? AND namespace=? AND workload_type=? AND workload_name=? AND pod_name=? AND container_name=? FOR UPDATE`,
		input.Key.RuleID, input.Key.ClusterID, input.Key.Namespace, input.Key.WorkloadType, input.Key.WorkloadName, input.Key.PodName, input.Key.Container).Scan(&status, &unknownFrom, &podUID, &started, &observed, &recovery, &version)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Transition{}, err
	}
	if err == nil {
		state.Status, state.UnknownFrom, state.PodUID, state.RecoveryCount = AlertStatus(status), AlertStatus(unknownFrom), podUID, int(recovery)
		state.StartedAt, state.ObservedAt = nullableTime(started), nullableTime(observed)
	}
	transition := Evaluate(state, input.Observation, input.Rule, now)
	newVersion := version + 1
	var value any
	if input.Value != nil {
		value = *input.Value
	}
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `INSERT INTO k8s_alert_state(rule_id,cluster_id,namespace,workload_type,workload_name,pod_name,container_name,status,unknown_from,pod_uid,value,observed_at,started_at,recovery_count,state_version,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			input.Key.RuleID, input.Key.ClusterID, input.Key.Namespace, input.Key.WorkloadType, input.Key.WorkloadName, input.Key.PodName, input.Key.Container, transition.State.Status, transition.State.UnknownFrom, transition.State.PodUID, value, nullableValueTime(transition.State.ObservedAt), nullableValueTime(transition.State.StartedAt), transition.State.RecoveryCount, newVersion, now)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE k8s_alert_state SET status=?,unknown_from=?,pod_uid=?,value=?,observed_at=?,started_at=?,recovery_count=?,state_version=?,updated_at=? WHERE rule_id=? AND cluster_id=? AND namespace=? AND workload_type=? AND workload_name=? AND pod_name=? AND container_name=?`,
			transition.State.Status, transition.State.UnknownFrom, transition.State.PodUID, value, nullableValueTime(transition.State.ObservedAt), nullableValueTime(transition.State.StartedAt), transition.State.RecoveryCount, newVersion, now,
			input.Key.RuleID, input.Key.ClusterID, input.Key.Namespace, input.Key.WorkloadType, input.Key.WorkloadName, input.Key.PodName, input.Key.Container)
	}
	if err != nil {
		var duplicate *mysql.MySQLError
		if retry && errors.As(err, &duplicate) && duplicate.Number == 1062 {
			_ = tx.Rollback()
			return s.evaluateAndPersist(ctx, input, now, false)
		}
		return Transition{}, err
	}
	if transition.Event != AlertEventNone {
		payload, marshalErr := json.Marshal(map[string]any{"key": input.Key, "state": transition.State, "value": input.Value, "notify_strategy_id": input.Rule.NotifyStrategyID})
		if marshalErr != nil {
			return Transition{}, marshalErr
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO k8s_alert_outbox(event_id,idempotency_key,rule_id,cluster_id,namespace,event_type,payload,status,available_at,created_at) VALUES(?,?,?,?,?,?,?,'pending',?,?) ON DUPLICATE KEY UPDATE idempotency_key=idempotency_key`,
			uuid.NewString(), OutboxIdempotencyKey(input.Key, newVersion, transition.Event), input.Key.RuleID, input.Key.ClusterID, input.Key.Namespace, transition.Event, payload, now, now)
		if err != nil {
			return Transition{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Transition{}, err
	}
	return transition, nil
}

func (s AlertStore) ClaimOutbox(ctx context.Context, owner string, limit int, lease time.Duration, now time.Time) ([]OutboxEvent, error) {
	if s.DB == nil || owner == "" || limit <= 0 || lease <= 0 {
		return nil, errors.New("invalid alert outbox claim")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT event_id,idempotency_key,rule_id,cluster_id,namespace,event_type,payload,attempts
FROM k8s_alert_outbox WHERE status='pending' AND available_at<=? AND (lease_until IS NULL OR lease_until<?)
ORDER BY available_at,event_id LIMIT ? FOR UPDATE`, now, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []OutboxEvent{}
	leaseToken := uuid.NewString()
	leaseUntil := now.Add(lease)
	for rows.Next() {
		var event OutboxEvent
		var payload []byte
		if err = rows.Scan(&event.EventID, &event.IdempotencyKey, &event.Key.RuleID, &event.Key.ClusterID, &event.Key.Namespace, &event.Event, &payload, &event.Attempts); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(payload, &event.Payload); err != nil {
			return nil, errors.New("invalid alert outbox payload")
		}
		if err = decodeOutboxKey(event.Payload, &event.Key); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	claimed := make([]OutboxEvent, 0, len(events))
	for _, event := range events {
		result, updateErr := tx.ExecContext(ctx, `UPDATE k8s_alert_outbox SET lease_token=?,lease_until=?,attempts=attempts+1 WHERE event_id=? AND status='pending' AND (lease_until IS NULL OR lease_until<?)`, leaseToken, leaseUntil, event.EventID, now)
		if updateErr != nil {
			return nil, updateErr
		}
		if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
			continue
		}
		event.LeaseToken = leaseToken
		event.Attempts++
		claimed = append(claimed, event)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return claimed, nil
}

func decodeOutboxKey(payload map[string]any, key *AlertKey) error {
	if key == nil {
		return errors.New("alert outbox key is required")
	}
	raw, ok := payload["key"]
	if !ok {
		return errors.New("alert outbox payload key is missing")
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return errors.New("invalid alert outbox payload key")
	}
	if err = json.Unmarshal(encoded, key); err != nil || key.RuleID == "" || key.ClusterID == "" || key.Namespace == "" {
		return errors.New("invalid alert outbox payload key")
	}
	return nil
}

func (s AlertStore) MarkOutboxSent(ctx context.Context, eventID, leaseToken string, now time.Time) error {
	result, err := s.DB.ExecContext(ctx, `UPDATE k8s_alert_outbox SET status='sent',sent_at=?,lease_token=NULL,lease_until=NULL WHERE event_id=? AND status='pending' AND lease_token=?`, now, eventID, leaseToken)
	if err != nil {
		return err
	}
	return requireOneRow(result, ErrLeaseLost)
}

func (s AlertStore) RetryOutbox(ctx context.Context, eventID, leaseToken string, availableAt time.Time, sendErr error) error {
	message := "notification failed"
	if sendErr != nil {
		message = sendErr.Error()
	}
	if len(message) > 1024 {
		message = message[:1024]
	}
	result, err := s.DB.ExecContext(ctx, `UPDATE k8s_alert_outbox SET available_at=?,last_error=?,lease_token=NULL,lease_until=NULL WHERE event_id=? AND status='pending' AND lease_token=?`, availableAt, message, eventID, leaseToken)
	if err != nil {
		return err
	}
	return requireOneRow(result, ErrLeaseLost)
}

func (s AlertStore) FailOutbox(ctx context.Context, eventID, leaseToken string, now time.Time, sendErr error) error {
	message := "notification failed permanently"
	if sendErr != nil {
		message = sendErr.Error()
	}
	if len(message) > 1024 {
		message = message[:1024]
	}
	result, err := s.DB.ExecContext(ctx, `UPDATE k8s_alert_outbox SET status='failed',last_error=?,sent_at=?,lease_token=NULL,lease_until=NULL WHERE event_id=? AND status='pending' AND lease_token=?`, message, now, eventID, leaseToken)
	if err != nil {
		return err
	}
	return requireOneRow(result, ErrLeaseLost)
}

func requireOneRow(result sql.Result, zero error) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return zero
	}
	return nil
}

func nullableTime(value sql.NullTime) time.Time {
	if value.Valid {
		return value.Time
	}
	return time.Time{}
}

func nullableValueTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
