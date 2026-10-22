package k8s

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
)

// ClusterStore keeps the node-proxy identity separate from namespace credentials.
type ClusterStore struct{ Store }

func (s ClusterStore) Save(ctx context.Context, r Registration) error {
	state := "PENDING"
	if !r.Enabled {
		state = "DISABLED"
	}
	var result sql.Result
	var err error
	if r.Version == 0 {
		_, err = s.DB.ExecContext(ctx, "INSERT INTO k8s_metric_source(cluster_id,enabled,version,state,credential_guid,api_server,ca,token) VALUES(?,?,1,?,?,?,?,?)", r.ClusterID, r.Enabled, state, r.Credentials.Guid, r.Credentials.APIServer, r.Credentials.CA, r.Credentials.EncryptedToken)
		var duplicate *mysql.MySQLError
		if !errors.As(err, &duplicate) || duplicate.Number != 1062 {
			return err
		}
	} else {
		result, err = s.DB.ExecContext(ctx, "UPDATE k8s_metric_source SET enabled=?,version=version+1,state=?,credential_guid=IF(?='',credential_guid,?),api_server=IF(?='',api_server,?),ca=IF(?='',ca,?),token=IF(?='',token,?) WHERE cluster_id=? AND version=?", r.Enabled, state, r.Credentials.Guid, r.Credentials.Guid, r.Credentials.Guid, r.Credentials.APIServer, r.Credentials.Guid, r.Credentials.CA, r.Credentials.Guid, r.Credentials.EncryptedToken, r.ClusterID, r.Version)
		if err = changed(result, err); err != ErrConflict {
			return err
		}
	}
	current := Registration{ClusterID: r.ClusterID, Namespace: r.Namespace}
	err = s.DB.QueryRowContext(ctx, "SELECT enabled,version,credential_guid,api_server,ca,token FROM k8s_metric_source WHERE cluster_id=?", r.ClusterID).Scan(&current.Enabled, &current.Version, &current.Credentials.Guid, &current.Credentials.APIServer, &current.Credentials.CA, &current.Credentials.EncryptedToken)
	if err == sql.ErrNoRows {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if sameAppliedRegistration(current, r, s.Seed) {
		return nil
	}
	return ErrConflict
}
func (s ClusterStore) List(ctx context.Context) ([]NamespaceState, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT cluster_id,enabled,version,state,credential_guid,api_server,ca,token FROM k8s_metric_source ORDER BY cluster_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []NamespaceState{}
	for rows.Next() {
		var r NamespaceState
		if err = rows.Scan(&r.ClusterID, &r.Enabled, &r.Version, &r.State, &r.Credentials.Guid, &r.Credentials.APIServer, &r.Credentials.CA, &r.Credentials.EncryptedToken); err != nil {
			return nil, err
		}
		r.APIServer = r.Credentials.APIServer
		r.CAConfigured = r.Credentials.CA != ""
		out = append(out, r)
	}
	return out, rows.Err()
}
