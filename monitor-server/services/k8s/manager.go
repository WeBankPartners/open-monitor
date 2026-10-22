package k8s

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"
)

var autoLogResolver LogTargetResolver

type LogTargetResolver interface {
	Resolve(context.Context, Container) (LogTarget, error)
}

func sameLogTargetObject(a, b LogTarget) bool {
	return a.Namespace == b.Namespace && a.WorkloadType == b.WorkloadType && a.WorkloadName == b.WorkloadName && a.PodName == b.PodName && a.Container == b.Container
}

func ConfigureLogTargetResolver(r LogTargetResolver) { autoLogResolver = r }

type Persistence interface {
	Claim(context.Context, Registration, time.Duration) (string, error)
	Publish(context.Context, Registration, string, *Snapshot, error) error
}
type LogAgentPersistence interface {
	GetLogAgentConfig(context.Context, string, string) (LogAgentConfig, error)
	SaveLogAgentConfig(context.Context, string, string, LogAgentConfig) error
	ListLogAgentNodes(context.Context, string) ([]string, error)
}

type LogAgentEndpointPersistence interface {
	ListLogAgentEndpoints(context.Context, string) ([]LogAgentEndpoint, error)
}
type persistentSource struct {
	DiscoverySource
	Persistence  Persistence
	Registration Registration
}

func (s persistentSource) Snapshot(ctx context.Context, cluster string) (*Snapshot, error) {
	ticket, err := s.Persistence.Claim(ctx, s.Registration, 2*time.Minute)
	if err != nil {
		return nil, err
	}
	scanCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	snapshot, scanErr := s.DiscoverySource.Snapshot(scanCtx, cluster)
	if err = s.Persistence.Publish(ctx, s.Registration, ticket, snapshot, scanErr); err != nil {
		return nil, err
	}
	if scanErr != nil {
		return nil, scanErr
	}
	if err = reconcileLogAgents(ctx, s.Persistence, cluster, s.Registration.Namespace, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

type AtomicLogAgentPersistence interface {
	UpdateLogAgentConfigs(context.Context, string, func(map[string]LogAgentConfig) (map[string]LogAgentConfig, error)) error
}
type logConfigSnapshot struct {
	Persistence
	configs map[string]LogAgentConfig
}

func (s *logConfigSnapshot) GetLogAgentConfig(_ context.Context, _, node string) (LogAgentConfig, error) {
	return s.configs[node], nil
}
func (s *logConfigSnapshot) SaveLogAgentConfig(_ context.Context, _, node string, config LogAgentConfig) error {
	s.configs[node] = config
	return nil
}
func (s *logConfigSnapshot) ListLogAgentNodes(context.Context, string) ([]string, error) {
	nodes := make([]string, 0, len(s.configs))
	for node := range s.configs {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

type resolvedLogTargets map[ObjectKey]LogTarget

func (r resolvedLogTargets) Resolve(_ context.Context, c Container) (LogTarget, error) {
	target, ok := r[c.Key]
	if !ok {
		return LogTarget{}, errors.New("log target resolution unavailable")
	}
	return target, nil
}
func reconcileLogAgents(ctx context.Context, persistence Persistence, cluster, namespace string, snapshot *Snapshot) error {
	if snapshot == nil {
		return nil
	}
	// Resolve external CMDB data before taking database locks.
	var resolved resolvedLogTargets
	if autoLogResolver != nil {
		resolved = resolvedLogTargets{}
		for _, container := range snapshot.Containers {
			if container.InstanceID == "" || container.Key.Namespace != namespace {
				continue
			}
			if target, err := autoLogResolver.Resolve(ctx, container); err == nil {
				resolved[container.Key] = target
			}
		}
	}
	if atomic, ok := persistence.(AtomicLogAgentPersistence); ok {
		return atomic.UpdateLogAgentConfigs(ctx, cluster, func(configs map[string]LogAgentConfig) (map[string]LogAgentConfig, error) {
			memory := &logConfigSnapshot{Persistence: persistence, configs: configs}
			err := reconcileResolvedLogAgents(ctx, memory, cluster, namespace, snapshot, resolved)
			return memory.configs, err
		})
	}
	return reconcileResolvedLogAgents(ctx, persistence, cluster, namespace, snapshot, resolved)
}
func reconcileResolvedLogAgents(ctx context.Context, persistence Persistence, cluster, namespace string, snapshot *Snapshot, resolved resolvedLogTargets) error {

	store, ok := persistence.(LogAgentPersistence)
	if !ok || snapshot == nil {
		return nil
	}
	nodes, err := store.ListLogAgentNodes(ctx, cluster)
	if err != nil {
		return err
	}
	if endpoints, ok := persistence.(LogAgentEndpointPersistence); ok {
		rows, endpointErr := endpoints.ListLogAgentEndpoints(ctx, cluster)
		if endpointErr != nil {
			return endpointErr
		}
		seen := make(map[string]bool, len(nodes))
		for _, node := range nodes {
			seen[node] = true
		}
		for _, endpoint := range rows {
			if endpoint.Enabled && !seen[endpoint.Node] {
				nodes = append(nodes, endpoint.Node)
				seen[endpoint.Node] = true
			}
		}
	}
	configs := make(map[string]LogAgentConfig, len(nodes))
	for _, node := range nodes {
		config, err := store.GetLogAgentConfig(ctx, cluster, node)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		configs[node] = cloneLogAgentConfig(config)
	}
	original := cloneLogAgentConfigs(configs)
	if resolved != nil {
		autoChanged := false
		for _, container := range snapshot.Containers {
			if container.InstanceID == "" {
				continue
			}
			target, resolveErr := resolved.Resolve(ctx, container)
			if resolveErr != nil {
				continue
			}
			metricTarget := target
			metricTarget.ID = target.ID + "-metric"
			metricTarget.Glob = "*_metric.log"
			metricTarget.Format = "otlp_metric"
			metricTarget.Rules = nil
			for _, candidate := range []LogTarget{target, metricTarget} {
				target = candidate
				for node, existingConfig := range configs {
					filtered := existingConfig.Targets[:0]
					for _, existing := range existingConfig.Targets {
						if existing.Namespace == namespace && strings.HasPrefix(existing.ID, "cmdb-") && sameLogTargetObject(existing, target) && existing.Format == target.Format {
							target.Rules = existing.Rules
							continue
						}
						filtered = append(filtered, existing)
					}
					existingConfig.Targets = filtered
					configs[node] = existingConfig
				}
				config := configs[container.Node]
				beforeTargets := append([]LogTarget(nil), config.Targets...)
				found := false
				for i := range config.Targets {
					if config.Targets[i].ID == target.ID {
						if len(config.Targets[i].Rules) > 0 {
							target.Rules = config.Targets[i].Rules
						}
						config.Targets[i] = target
						found = true
						break
					}
				}
				if !found {
					config.Targets = append(config.Targets, target)
				}
				if config.Version == 0 {
					config.Version = 1
				}
				if !sameLogTargets(beforeTargets, config.Targets) {
					autoChanged = true
				}
				configs[container.Node] = config
			}
		}
		if autoChanged {
			updated, changed, err := RebindLogAgentConfigsForNamespace(configs, *snapshot, namespace)
			if err != nil {
				return err
			}
			updated, changed = applyLogAgentVersions(original, updated)
			return saveLogAgentConfigs(ctx, store, cluster, original, updated, changed)
		}
	}
	updated, changed, err := RebindLogAgentConfigsForNamespace(configs, *snapshot, namespace)
	if err != nil {
		return err
	}
	updated, changed = applyLogAgentVersions(original, updated)
	if !changed {
		return nil
	}
	return saveLogAgentConfigs(ctx, store, cluster, original, updated, changed)
}

func cloneLogAgentConfig(config LogAgentConfig) LogAgentConfig {
	config.Targets = append([]LogTarget(nil), config.Targets...)
	for i := range config.Targets {
		config.Targets[i].Rules = append([]LogRule(nil), config.Targets[i].Rules...)
		for j := range config.Targets[i].Rules {
			config.Targets[i].Rules[j].Exclude = append([]LogExclusion(nil), config.Targets[i].Rules[j].Exclude...)
		}
	}
	return config
}

func cloneLogAgentConfigs(configs map[string]LogAgentConfig) map[string]LogAgentConfig {
	out := make(map[string]LogAgentConfig, len(configs))
	for node, config := range configs {
		out[node] = cloneLogAgentConfig(config)
	}
	return out
}

func applyLogAgentVersions(original, updated map[string]LogAgentConfig) (map[string]LogAgentConfig, bool) {
	changed := false
	for node, config := range updated {
		before, exists := original[node]
		if !exists {
			if len(config.Targets) > 0 {
				config.Version = 1
				changed = true
			}
		} else if !sameLogTargets(before.Targets, config.Targets) {
			config.Version = before.Version + 1
			changed = true
		} else {
			config = cloneLogAgentConfig(before)
		}
		updated[node] = config
	}
	return updated, changed
}

func saveLogAgentConfigs(ctx context.Context, store LogAgentPersistence, cluster string, configs, updated map[string]LogAgentConfig, changed bool) error {
	if !changed {
		return nil
	}
	for node, config := range updated {
		if previous, exists := configs[node]; exists && previous.Version == config.Version {
			continue
		}
		if _, existed := configs[node]; !existed && len(config.Targets) == 0 {
			continue
		}
		if err := store.SaveLogAgentConfig(ctx, cluster, node, config); err != nil {
			return err
		}
	}
	return nil
}

// RunDiscovery restores registrations at startup and supervises per-namespace workers.
func RunDiscovery(ctx context.Context, policy Policy, store Store, index *Index, report func(error)) {
	if !policy.Enabled {
		return
	}
	type running struct {
		registration discoveryRegistration
		cancel       context.CancelFunc
		done         chan struct{}
	}
	active := map[scope]running{}
	var workers sync.WaitGroup
	defer func() {
		for _, job := range active {
			job.cancel()
		}
		workers.Wait()
	}()
	for {
		readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		rows, err := store.List(readCtx)
		clusters, clusterErr := (ClusterStore{Store: store}).List(readCtx)
		cancel()
		if err != nil || clusterErr != nil {
			report(errors.New("cannot load Kubernetes namespace registrations"))
		} else {
			desired := desiredDiscoveryScopes(policy, rows, clusters)
			for key, job := range active {
				r, ok := desired[key]
				if !ok || !r.sameVersion(job.registration) {
					index.Remove(key.cluster, key.namespace)
					job.cancel()
					<-job.done
					delete(active, key)
				}
			}
			for key, r := range desired {
				if _, ok := active[key]; ok {
					continue
				}
				client, err := policy.ClientWithCluster(r.Registration, r.cluster, store.Seed)
				if err != nil {
					report(err)
					continue
				}
				index.Configure(r.ClusterID, r.Namespace, r.Version, true)
				workCtx, stop := context.WithCancel(ctx)
				done := make(chan struct{})
				active[key] = running{r, stop, done}
				workers.Add(1)
				go func(registration Registration, client *Client, workCtx context.Context, done chan struct{}) {
					defer workers.Done()
					defer close(done)
					defer client.Close()
					source := persistentSource{DiscoverySource: client, Persistence: store, Registration: registration}
					worker := Worker{Index: index, Source: source, Cluster: registration.ClusterID, Namespace: registration.Namespace, Version: registration.Version, Report: report}
					if err := worker.Run(workCtx); err != nil && !errors.Is(err, context.Canceled) {
						report(err)
					}
				}(r.Registration, client, workCtx, done)
			}
		}
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

type discoveryRegistration struct {
	Registration
	cluster Registration
}

func (r discoveryRegistration) sameVersion(other discoveryRegistration) bool {
	return r.Version == other.Version && r.cluster.Version == other.cluster.Version
}
func desiredDiscoveryScopes(policy Policy, namespaces, clusters []NamespaceState) map[scope]discoveryRegistration {
	enabledClusters := map[string]Registration{}
	for _, cluster := range clusters {
		if cluster.Enabled {
			enabledClusters[cluster.ClusterID] = cluster.Registration
		}
	}
	desired := map[scope]discoveryRegistration{}
	for _, row := range namespaces {
		if row.Enabled && policy.Validate(row.Registration) == nil {
			if cluster, ok := enabledClusters[row.ClusterID]; ok {
				desired[scope{row.ClusterID, row.Namespace}] = discoveryRegistration{row.Registration, cluster}
			}
		}
	}
	return desired
}
