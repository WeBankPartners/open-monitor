package k8s

import (
	"context"
	"errors"

	"regexp"
	"strings"
	"time"
)

var clusterPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var ErrConflict = errors.New("namespace version conflict")

const (
	DefaultMetricsIntervalSeconds = 30
	MinMetricsIntervalSeconds     = 15
	MaxMetricsIntervalSeconds     = 300
)

// Policy controls the deployment environment and optional namespace allowlist.
type Policy struct {
	Enabled                bool   `json:"enabled"`
	LocalCollectionEnabled bool   `json:"-"`
	Environment            string `json:"environment"`
	MetricsIntervalSeconds int    `json:"-"`
	// LogAgentToken authenticates node-local log agents. It is injected through
	// the deployment secret and is never returned by an API.
	LogAgentToken string                   `json:"logAgentToken"`
	Clusters      map[string]ClusterConfig `json:"clusters"`
}

// LocalCollectionActive controls only built-in collection and alert delivery.
// Discovery and Agent configuration continue under the existing master switch.
func (p Policy) LocalCollectionActive() bool {
	return p.Enabled && p.LocalCollectionEnabled
}

func (p Policy) MetricsInterval() time.Duration {
	seconds := p.MetricsIntervalSeconds
	if seconds < MinMetricsIntervalSeconds || seconds > MaxMetricsIntervalSeconds {
		seconds = DefaultMetricsIntervalSeconds
	}
	return time.Duration(seconds) * time.Second
}

type ClusterConfig struct {
	Namespaces map[string]NamespaceScope `json:"namespaces"`
}
type NamespaceScope struct{}

type Registration struct {
	Credentials Credentials `json:"-"`
	ClusterID   string      `json:"clusterId"`
	Namespace   string      `json:"namespace"`
	Enabled     bool        `json:"enabled"`
	Version     uint64      `json:"version"`
}
type RegistrationStore interface {
	Save(context.Context, Registration) error
}
type Registry struct {
	Policy     Policy
	Repository RegistrationStore
}

func (p Policy) Validate(r Registration) error {
	if err := p.ValidateCluster(r.ClusterID); err != nil {
		return err
	}
	if len(r.Namespace) > 63 || !namespacePattern.MatchString(r.Namespace) || len(r.ClusterID) > 64 || !clusterPattern.MatchString(r.ClusterID) {
		return errors.New("invalid monitoring scope")
	}
	if len(p.Clusters) == 0 {
		return nil
	}
	cluster, ok := p.Clusters[r.ClusterID]
	if !ok {
		return errors.New("cluster is not allowed")
	}
	if _, ok = cluster.Namespaces[r.Namespace]; !ok {
		return errors.New("namespace is not allowed")
	}
	return nil
}
func (r Registry) Save(ctx context.Context, input Registration) error {
	if err := r.Policy.Validate(input); err != nil {
		return err
	}
	if r.Repository == nil {
		return errors.New("namespace repository unavailable")
	}
	return r.Repository.Save(ctx, input)
}
func (p Policy) Client(r Registration, seed string) (*Client, error) {
	return p.ClientWithCluster(r, r, seed)
}

func (p Policy) ClientWithCluster(namespace, cluster Registration, seed string) (*Client, error) {
	if err := p.Validate(namespace); err != nil {
		return nil, err
	}
	clusterCredentials := Resolve(namespace.Credentials, cluster.Credentials)
	token, err := clusterCredentials.PlainToken(seed)
	if err != nil {
		return nil, err
	}
	return NewClient(clusterCredentials.APIServer, []byte(clusterCredentials.CA), strings.TrimSpace(token), namespace.Namespace)
}

func (p Policy) ValidateCluster(cluster string) error {
	if len(cluster) > 64 || !clusterPattern.MatchString(cluster) {
		return errors.New("invalid monitoring cluster")
	}
	if len(p.Clusters) > 0 {
		if _, ok := p.Clusters[cluster]; !ok {
			return errors.New("cluster is not allowed")
		}
	}
	return nil
}
