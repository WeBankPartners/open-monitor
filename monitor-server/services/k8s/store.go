package k8s

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"time"
)

type Store struct {
	DB   *sql.DB
	Seed string
}
type NamespaceState struct {
	Registration
	APIServer    string     `json:"apiServer,omitempty"`
	CAConfigured bool       `json:"caConfigured"`
	State        string     `json:"state"`
	Snapshot     *Snapshot  `json:"snapshot,omitempty"`
	ObservedAt   *time.Time `json:"observedAt,omitempty"`
}

func (s Store) ListAlertRules(ctx context.Context, cluster, namespace string) ([]AlertRule, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT rule_id,workload_type,workload_name,pod_name,container_name,metric,operator,threshold,for_seconds,recover_after,notify_strategy_id,enabled FROM k8s_alert_rule WHERE cluster_id=? AND namespace=? ORDER BY rule_id`, cluster, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AlertRule{}
	for rows.Next() {
		var r AlertRule
		if err = rows.Scan(&r.ID, &r.WorkloadType, &r.WorkloadName, &r.PodName, &r.Container, &r.Metric, &r.Operator, &r.Threshold, &r.ForSeconds, &r.RecoverAfter, &r.NotifyStrategyID, &r.Enabled); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s Store) Save(ctx context.Context, r Registration) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var result sql.Result
	state := "PENDING"
	if !r.Enabled {
		state = "DISABLED"
	}
	if r.Version == 0 {
		_, err = tx.ExecContext(ctx, "INSERT INTO k8s_namespace_source(cluster_id,namespace,enabled,version,state,credential_guid,token) VALUES(?,?,?,1,?,?,?)", r.ClusterID, r.Namespace, r.Enabled, state, r.Credentials.Guid, r.Credentials.EncryptedToken)
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			_ = tx.Rollback()
			return s.checkRetry(ctx, r)
		}
		if err != nil {
			return err
		}
		if err = s.ensureDefaultAlertRulesTx(ctx, tx, r.ClusterID, r.Namespace); err != nil {
			return err
		}
		return tx.Commit()
	}
	result, err = tx.ExecContext(ctx, "UPDATE k8s_namespace_source SET enabled=?,version=version+1,state=?,snapshot=NULL,observed_at=NULL,lease_token=NULL,lease_until=NULL,credential_guid=IF(?='',credential_guid,?),token=IF(?='',token,?) WHERE cluster_id=? AND namespace=? AND version=?", r.Enabled, state, r.Credentials.Guid, r.Credentials.Guid, r.Credentials.Guid, r.Credentials.EncryptedToken, r.ClusterID, r.Namespace, r.Version)
	if err = changed(result, err); err == ErrConflict {
		_ = tx.Rollback()
		return s.checkRetry(ctx, r)
	}
	if err != nil {
		return err
	}
	if err = s.ensureDefaultAlertRulesTx(ctx, tx, r.ClusterID, r.Namespace); err != nil {
		return err
	}
	return tx.Commit()
}

func (s Store) ensureDefaultAlertRules(ctx context.Context, cluster, namespace string) error {
	return s.ensureDefaultAlertRulesTx(ctx, s.DB, cluster, namespace)
}

type sqlExecutor interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

func (s Store) ensureDefaultAlertRulesTx(ctx context.Context, executor sqlExecutor, cluster, namespace string) error {
	rules := []struct {
		id, metric, op string
		threshold      float64
	}{
		{"default-log-error", "log_matches", ">", 0},
		{"default-not-ready", "ready", "<", 1},
		{"default-oom", "oom_events", ">", 0},
		{"default-cpu-throttled", "cpu_throttled", ">", 0},
	}
	for _, rule := range rules {
		if _, err := executor.ExecContext(ctx, `INSERT INTO k8s_alert_rule(rule_id,cluster_id,namespace,metric,operator,threshold,for_seconds,recover_after,notify_strategy_id,enabled,version,updated_at) VALUES(?,?,?,?,?,?,0,3,0,1,1,CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE rule_id=rule_id`, rule.id, cluster, namespace, rule.metric, rule.op, rule.threshold); err != nil {
			return err
		}
	}
	return nil
}
func changed(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrConflict
	}
	return nil
}
func (s Store) List(ctx context.Context) ([]NamespaceState, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT cluster_id,namespace,enabled,version,state,snapshot,observed_at,credential_guid,token FROM k8s_namespace_source ORDER BY cluster_id,namespace")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []NamespaceState{}
	for rows.Next() {
		var row NamespaceState
		var raw []byte
		var observed sql.NullTime
		if err = rows.Scan(&row.ClusterID, &row.Namespace, &row.Enabled, &row.Version, &row.State, &raw, &observed, &row.Credentials.Guid, &row.Credentials.EncryptedToken); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			if err = json.Unmarshal(raw, &row.Snapshot); err != nil {
				return nil, err
			}
		}
		if observed.Valid {
			row.ObservedAt = &observed.Time
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
func (s Store) Claim(ctx context.Context, r Registration, duration time.Duration) (string, error) {
	token := uuid.NewString()
	seconds := int64(duration / time.Second)
	if seconds < 1 {
		return "", errors.New("invalid lease duration")
	}
	result, err := s.DB.ExecContext(ctx, "UPDATE k8s_namespace_source SET lease_token=?,lease_until=TIMESTAMPADD(SECOND,?,CURRENT_TIMESTAMP(6)) WHERE cluster_id=? AND namespace=? AND version=? AND enabled=1 AND (lease_until IS NULL OR lease_until<CURRENT_TIMESTAMP(6))", token, seconds, r.ClusterID, r.Namespace, r.Version)
	if err = changed(result, err); err != nil {
		return "", err
	}
	return token, nil
}
func (s Store) Publish(ctx context.Context, r Registration, ticket string, snapshot *Snapshot, scanErr error) error {
	if scanErr != nil {
		result, err := s.DB.ExecContext(ctx, "UPDATE k8s_namespace_source SET state='DEGRADED',lease_token=NULL,lease_until=NULL WHERE cluster_id=? AND namespace=? AND version=? AND enabled=1 AND lease_token=? AND lease_until>CURRENT_TIMESTAMP(6)", r.ClusterID, r.Namespace, r.Version, ticket)
		return changed(result, err)
	}
	if snapshot == nil {
		return errors.New("cannot publish empty snapshot")
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	result, err := s.DB.ExecContext(ctx, "UPDATE k8s_namespace_source SET snapshot=?,state='ACTIVE',observed_at=CURRENT_TIMESTAMP(6),lease_token=NULL,lease_until=NULL WHERE cluster_id=? AND namespace=? AND version=? AND enabled=1 AND lease_token=? AND lease_until>CURRENT_TIMESTAMP(6)", string(raw), r.ClusterID, r.Namespace, r.Version, ticket)
	return changed(result, err)
}

// checkRetry recognizes a completed write without replaying its side effects.
func (s Store) checkRetry(ctx context.Context, input Registration) error {
	current := Registration{ClusterID: input.ClusterID, Namespace: input.Namespace}
	err := s.DB.QueryRowContext(ctx, "SELECT enabled,version,credential_guid,token FROM k8s_namespace_source WHERE cluster_id=? AND namespace=?", input.ClusterID, input.Namespace).Scan(&current.Enabled, &current.Version, &current.Credentials.Guid, &current.Credentials.EncryptedToken)
	if err == sql.ErrNoRows {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if sameAppliedRegistration(current, input, s.Seed) {
		return nil
	}
	return ErrConflict
}
func sameAppliedRegistration(current, input Registration, seed string) bool {
	if input.Version == ^uint64(0) || current.Version != input.Version+1 || current.ClusterID != input.ClusterID || current.Namespace != input.Namespace || current.Enabled != input.Enabled {
		return false
	}
	if !input.Enabled {
		return true
	}
	a, b := current.Credentials, input.Credentials
	if a.Guid != b.Guid || a.APIServer != b.APIServer || a.CA != b.CA {
		return false
	}
	if a.EncryptedToken == "" && b.EncryptedToken == "" {
		return input.Namespace != ""
	}
	plainA, err := a.PlainToken(seed)
	if err != nil || plainA == "" {
		return false
	}
	plainB, err := b.PlainToken(seed)
	return err == nil && plainA == plainB
}

// UpdateLogAgentConfigs serializes all writers for a cluster using its registration
// row. The callback always sees the latest committed node configurations; all
// node moves and their version changes are committed together.
func (s Store) UpdateLogAgentConfigs(ctx context.Context, cluster string, update func(map[string]LogAgentConfig) (map[string]LogAgentConfig, error)) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var enabled bool
	if err = tx.QueryRowContext(ctx, "SELECT enabled FROM k8s_metric_source WHERE cluster_id=? FOR UPDATE", cluster).Scan(&enabled); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT namespace FROM k8s_namespace_source WHERE cluster_id=? AND enabled=1 ORDER BY namespace FOR SHARE", cluster)
	if err != nil {
		return err
	}
	allowed := map[string]bool{}
	for rows.Next() {
		var namespace string
		if err = rows.Scan(&namespace); err != nil {
			rows.Close()
			return err
		}
		allowed[namespace] = enabled
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, "SELECT node_name,version,config FROM k8s_log_agent_config WHERE cluster_id=? ORDER BY node_name FOR UPDATE", cluster)
	if err != nil {
		return err
	}
	original := map[string]LogAgentConfig{}
	for rows.Next() {
		var node string
		var config LogAgentConfig
		var raw []byte
		if err = rows.Scan(&node, &config.Version, &raw); err != nil {
			rows.Close()
			return err
		}
		version := config.Version
		if err = json.Unmarshal(raw, &config); err != nil || config.Version != version {
			rows.Close()
			return errors.New("invalid stored log agent configuration")
		}
		original[node] = config
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	updated, err := update(cloneLogAgentConfigs(original))
	if err != nil {
		return err
	}
	for node, config := range updated {
		targets := make([]LogTarget, 0, len(config.Targets))
		for _, target := range config.Targets {
			if allowed[target.Namespace] {
				targets = append(targets, target)
			}
		}
		config.Targets = targets
		updated[node] = config
	}
	updated, _ = applyLogAgentVersions(original, updated)
	for node, config := range updated {
		before, exists := original[node]
		if (exists && before.Version == config.Version) || (!exists && len(config.Targets) == 0) {
			continue
		}
		raw, encodeErr := EncodeLogAgentConfig(config)
		if encodeErr != nil {
			return encodeErr
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO k8s_log_agent_config(cluster_id,node_name,version,config,updated_at) VALUES(?,?,?,?,CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE version=VALUES(version),config=VALUES(config),updated_at=VALUES(updated_at)", cluster, node, config.Version, string(raw)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s Store) SaveLogAgentConfig(ctx context.Context, cluster, node string, c LogAgentConfig) error {
	c = versionLogRules(LogAgentConfig{}, c)
	if err := ValidateLogAgentConfig(c); err != nil {
		return err
	}
	return s.UpdateLogAgentConfigs(ctx, cluster, func(configs map[string]LogAgentConfig) (map[string]LogAgentConfig, error) {
		current := configs[node]
		c = versionLogRules(current, c)
		if sameLogTargets(current.Targets, c.Targets) {
			return configs, nil
		}
		if c.Version != current.Version+1 {
			return nil, ErrConflict
		}
		configs[node] = c
		return configs, nil
	})
}
func (s Store) GetLogAgentConfig(ctx context.Context, cluster, node string) (LogAgentConfig, error) {
	var configs map[string]LogAgentConfig
	err := s.UpdateLogAgentConfigs(ctx, cluster, func(current map[string]LogAgentConfig) (map[string]LogAgentConfig, error) {
		configs = current
		return current, nil
	})
	if err != nil {
		return LogAgentConfig{}, err
	}
	result, found := configs[node]
	if !found {
		return result, sql.ErrNoRows
	}
	return result, nil
}
func (s Store) ListLogAgentNodes(ctx context.Context, cluster string) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT node_name FROM k8s_log_agent_config WHERE cluster_id=? ORDER BY node_name", cluster)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var n string
		if err = rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s Store) UpsertLogAgentEndpoint(ctx context.Context, e LogAgentEndpoint) error {
	if err := ValidateLogAgentEndpoint(e); err != nil {
		return err
	}
	_, err := s.DB.ExecContext(ctx, "INSERT INTO k8s_log_agent_endpoint(cluster_id,node_name,address,port,reported_address,reported_port,pod_uid,last_seen,enabled,manual_override) VALUES(?,?,?,?,?,?,?,CURRENT_TIMESTAMP(6),1,0) ON DUPLICATE KEY UPDATE address=IF(manual_override=1,address,VALUES(address)),port=IF(manual_override=1,port,VALUES(port)),reported_address=VALUES(reported_address),reported_port=VALUES(reported_port),pod_uid=VALUES(pod_uid),last_seen=VALUES(last_seen)", e.ClusterID, e.Node, e.Address, e.Port, e.Address, e.Port, e.PodUID)
	return err
}
func (s Store) SetLogAgentEndpoint(ctx context.Context, e LogAgentEndpoint, manual bool) error {
	if !manual {
		return s.UpsertLogAgentEndpoint(ctx, e)
	}
	if err := ValidateLogAgentEndpoint(e); err != nil {
		return err
	}
	_, err := s.DB.ExecContext(ctx, "INSERT INTO k8s_log_agent_endpoint(cluster_id,node_name,address,port,last_seen,enabled,manual_override) VALUES(?,?,?,?,NULL,1,1) ON DUPLICATE KEY UPDATE address=VALUES(address),port=VALUES(port),manual_override=1,enabled=1", e.ClusterID, e.Node, e.Address, e.Port)
	return err
}
func (s Store) ListLiveLogAgentEndpoints(ctx context.Context, cluster string, ttl time.Duration) ([]LogAgentEndpoint, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT cluster_id,node_name,address,port,pod_uid,COALESCE(FLOOR(UNIX_TIMESTAMP(last_seen)),0),enabled,manual_override FROM k8s_log_agent_endpoint WHERE cluster_id=? AND enabled=1 AND last_seen>=TIMESTAMPADD(SECOND,?,CURRENT_TIMESTAMP(6)) ORDER BY node_name", cluster, -int64(ttl/time.Second))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LogAgentEndpoint{}
	for rows.Next() {
		var e LogAgentEndpoint
		if err = rows.Scan(&e.ClusterID, &e.Node, &e.Address, &e.Port, &e.PodUID, &e.LastSeen, &e.Enabled, &e.ManualOverride); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s Store) ListLiveLogAgentEndpointsAll(ctx context.Context, ttl time.Duration) ([]LogAgentEndpoint, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT cluster_id,node_name,address,port,pod_uid,COALESCE(FLOOR(UNIX_TIMESTAMP(last_seen)),0),enabled,manual_override FROM k8s_log_agent_endpoint WHERE enabled=1 AND last_seen>=TIMESTAMPADD(SECOND,?,CURRENT_TIMESTAMP(6)) ORDER BY cluster_id,node_name", -int64(ttl/time.Second))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LogAgentEndpoint{}
	for rows.Next() {
		var e LogAgentEndpoint
		if err = rows.Scan(&e.ClusterID, &e.Node, &e.Address, &e.Port, &e.PodUID, &e.LastSeen, &e.Enabled, &e.ManualOverride); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s Store) ListLogAgentEndpoints(ctx context.Context, cluster string) ([]LogAgentEndpoint, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT cluster_id,node_name,address,port,pod_uid,COALESCE(FLOOR(UNIX_TIMESTAMP(last_seen)),0),enabled,manual_override,reported_address,reported_port FROM k8s_log_agent_endpoint WHERE cluster_id=? ORDER BY node_name", cluster)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LogAgentEndpoint{}
	for rows.Next() {
		var e LogAgentEndpoint
		if err = rows.Scan(&e.ClusterID, &e.Node, &e.Address, &e.Port, &e.PodUID, &e.LastSeen, &e.Enabled, &e.ManualOverride, &e.ReportedAddress, &e.ReportedPort); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s Store) ReleaseLogAgentEndpoint(ctx context.Context, cluster, node string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE k8s_log_agent_endpoint SET address=IF(reported_address='',address,reported_address),port=IF(reported_port=0,port,reported_port),manual_override=0 WHERE cluster_id=? AND node_name=?", cluster, node)
	return err
}

func (s Store) DisableLogAgentEndpoint(ctx context.Context, cluster, node string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE k8s_log_agent_endpoint SET enabled=0 WHERE cluster_id=? AND node_name=?", cluster, node)
	return err
}
