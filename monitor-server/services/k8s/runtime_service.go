package k8s

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

type MetricSourceFactory func(Registration) (MetricSource, func(), error)
type MetricService struct {
	Index    *Index
	Exporter *Exporter
	Factory  MetricSourceFactory
}

func (s MetricService) Cycle(ctx context.Context, rows []NamespaceState, credentials []Registration) error {
	groups := map[string][]Registration{}
	for _, row := range rows {
		if row.Enabled {
			groups[row.ClusterID] = append(groups[row.ClusterID], row.Registration)
		}
	}
	sources := map[string]Registration{}
	for _, r := range credentials {
		if r.Enabled {
			sources[r.ClusterID] = r
		}
	}
	var first error
	for cluster, registrations := range groups {
		credential, ok := sources[cluster]
		if !ok {
			for _, r := range registrations {
				s.Exporter.Fail(cluster, r.Namespace)
			}
			first = errors.New("cluster metrics credential unavailable")
			continue
		}
		source, closeFn, err := s.Factory(credential)
		if err != nil {
			for _, r := range registrations {
				s.Exporter.Fail(cluster, r.Namespace)
			}
			first = err
			continue
		}
		// A cycle-local response cache avoids scraping a shared node once per namespace.
		cached := &nodeResponseCache{source: source, payloads: map[string]string{}, failures: map[string]error{}}
		for _, r := range registrations {
			runtime := MetricRuntime{Index: s.Index, Exporter: s.Exporter, Collector: Collector{Source: cached}, Freshness: time.Minute}
			if err = runtime.Collect(ctx, cluster, r.Namespace); err != nil {
				first = err
			}
		}
		closeFn()
	}
	return first
}

type nodeResponseCache struct {
	source   MetricSource
	payloads map[string]string
	failures map[string]error
}

func (c *nodeResponseCache) Fetch(ctx context.Context, node string) (string, error) {
	if p, ok := c.payloads[node]; ok {
		return p, nil
	}
	if err, ok := c.failures[node]; ok {
		return "", err
	}
	p, err := c.source.Fetch(ctx, node)
	if err != nil {
		c.failures[node] = err
		return "", err
	}
	c.payloads[node] = p
	return p, nil
}

// RunMetrics is serialized: a slow cycle never overlaps the next one.
func RunMetrics(ctx context.Context, policy Policy, store Store, index *Index, exporter *Exporter, report func(error)) {
	service := MetricService{Index: index, Exporter: exporter, Factory: func(r Registration) (MetricSource, func(), error) {
		token, err := r.Credentials.PlainToken(store.Seed)
		if err != nil {
			return nil, nil, err
		}
		client, err := NewClient(r.Credentials.APIServer, []byte(r.Credentials.CA), token, "default")
		if err != nil {
			return nil, nil, err
		}
		return CAdvisorSource{Client: client}, client.Close, nil
	}}
	interval := policy.MetricsInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		func() {
			defer func() {
				if p := recover(); p != nil {
					exporter.Clear()
					report(fmt.Errorf("metrics worker panic (%T)\n%s", p, debug.Stack()))
				}
			}()
			timeout := interval - time.Second
			if timeout < time.Second {
				timeout = time.Second
			}
			cycleCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			rows, err := store.List(cycleCtx)
			if err != nil {
				exporter.Clear()
				report(errors.New("cannot load metric scopes"))
				return
			}
			filtered := []NamespaceState{}
			for _, row := range rows {
				if row.Enabled && policy.Validate(row.Registration) == nil {
					filtered = append(filtered, row)
				}
			}
			clusters, err := (ClusterStore{Store: store}).List(cycleCtx)
			if err != nil {
				exporter.Clear()
				report(errors.New("cannot load cluster metric credentials"))
				return
			}
			credentials := []Registration{}
			for _, r := range clusters {
				if policy.ValidateCluster(r.ClusterID) == nil {
					credentials = append(credentials, r.Registration)
				}
			}
			exporter.Prune(filtered)
			if err = service.Cycle(cycleCtx, filtered, credentials); err != nil {
				report(err)
			}
		}()
		select {
		case <-ctx.Done():
			exporter.Clear()
			return
		case <-ticker.C:
		}
	}
}
