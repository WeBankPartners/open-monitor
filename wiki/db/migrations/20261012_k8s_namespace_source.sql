#@v3.8.0-begin@;

-- MySQL 8. Apply to each environment's independent Monitor database.
-- Additive and safe to re-run. Tokens are encrypted using the existing WeCube credential mechanism.
CREATE TABLE IF NOT EXISTS k8s_namespace_source (
  cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  namespace VARCHAR(63) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  credential_guid VARCHAR(128) NOT NULL DEFAULT '',
  token TEXT NOT NULL DEFAULT (''),
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  version BIGINT UNSIGNED NOT NULL,
  state VARCHAR(16) NOT NULL,
  snapshot JSON NULL,
  observed_at DATETIME(6) NULL,
  lease_token VARCHAR(36) CHARACTER SET ascii COLLATE ascii_bin NULL,
  lease_until DATETIME(6) NULL,
  PRIMARY KEY (cluster_id, namespace)
) ENGINE=InnoDB;

-- Independent encrypted cluster-level credential for cAdvisor node proxy access.
CREATE TABLE IF NOT EXISTS k8s_metric_source (
 cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 enabled BOOLEAN NOT NULL DEFAULT FALSE,
 version BIGINT UNSIGNED NOT NULL,
 state VARCHAR(16) NOT NULL,
 credential_guid VARCHAR(128) NOT NULL DEFAULT '',
 api_server VARCHAR(512) NOT NULL DEFAULT '',
 ca TEXT NOT NULL DEFAULT (''),
 token TEXT NOT NULL DEFAULT (''),
 PRIMARY KEY (cluster_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS k8s_log_agent_config (
 cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 node_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 version BIGINT UNSIGNED NOT NULL,
 config JSON NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY (cluster_id,node_name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS k8s_log_agent_endpoint (
  cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  node_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  address VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  port INT UNSIGNED NOT NULL,
  pod_uid VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  reported_address VARCHAR(253) NOT NULL DEFAULT '',
  reported_port INT UNSIGNED NOT NULL DEFAULT 0,
  last_seen DATETIME(6) NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  manual_override BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (cluster_id,node_name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS k8s_alert_rule (
  rule_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  namespace VARCHAR(63) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  workload_type VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  workload_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  pod_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  container_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  metric VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  operator VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  threshold DOUBLE NOT NULL,
  for_seconds INT UNSIGNED NOT NULL DEFAULT 0,
  recover_after INT UNSIGNED NOT NULL DEFAULT 3,
  notify_strategy_id INT NOT NULL DEFAULT 0,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  version BIGINT UNSIGNED NOT NULL DEFAULT 1,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (cluster_id, namespace, rule_id),
  KEY idx_k8s_alert_rule_scope (cluster_id, namespace, enabled)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS k8s_alert_state (
  rule_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  namespace VARCHAR(63) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  workload_type VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  workload_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  pod_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  container_name VARCHAR(253) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  status VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  unknown_from VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  pod_uid VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  value DOUBLE NULL,
  observed_at DATETIME(6) NULL,
  started_at DATETIME(6) NULL,
  recovery_count INT UNSIGNED NOT NULL DEFAULT 0,
  state_version BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (rule_id, cluster_id, namespace, workload_type, workload_name, pod_name, container_name),
  KEY idx_k8s_alert_state_status (cluster_id, namespace, status, updated_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS k8s_alert_outbox (
  event_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  -- Application values are fixed 64-character SHA-256 hex keys; the wider column preserves upgrade compatibility.
  idempotency_key VARCHAR(384) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  rule_id VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  cluster_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  namespace VARCHAR(63) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  event_type VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  payload JSON NOT NULL,
  status VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'pending',
  attempts INT UNSIGNED NOT NULL DEFAULT 0,
  available_at DATETIME(6) NOT NULL,
  lease_token CHAR(36) CHARACTER SET ascii COLLATE ascii_bin NULL,
  lease_until DATETIME(6) NULL,
  last_error VARCHAR(1024) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL,
  sent_at DATETIME(6) NULL,
  PRIMARY KEY (event_id),
  UNIQUE KEY uk_k8s_alert_outbox_idempotency (idempotency_key),
  KEY idx_k8s_alert_outbox_claim (status, available_at, lease_until)
) ENGINE=InnoDB;

#@v3.8.0-end@;
