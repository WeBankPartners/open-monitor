package k8s

import (
	"context"
	"errors"
	"time"
)

type MetricRuntime struct {
	Index     *Index
	Collector Collector
	Exporter  *Exporter
	Freshness time.Duration
}

func (r MetricRuntime) Collect(ctx context.Context, cluster, namespace string) error {
	if r.Index == nil || r.Exporter == nil {
		return errors.New("metric runtime is not configured")
	}
	snapshot, ok := r.Index.Current(cluster, namespace)
	if !ok {
		return errors.New("namespace inventory unavailable")
	}
	if snapshot.DiscoveryFailed {
		r.Exporter.Fail(cluster, namespace)
		return errors.New("namespace discovery unavailable")
	}
	samples, err := r.Collector.Collect(ctx, snapshot)
	if err != nil {
		r.Exporter.Fail(cluster, namespace)
		return err
	}
	if r.Freshness <= 0 {
		r.Freshness = time.Minute
	}
	now := r.Exporter.now().UnixMilli()
	fresh := samples[:0]
	for _, sample := range samples {
		if sample.Timestamp == 0 || (sample.Timestamp <= now+5000 && now-sample.Timestamp <= r.Freshness.Milliseconds()) {
			fresh = append(fresh, sample)
		}
	}
	if len(snapshot.Containers) > 0 && !hasFreshRuntimeSample(fresh) {
		r.Exporter.Fail(cluster, namespace)
		return errors.New("no fresh container metrics")
	}
	return r.Exporter.Set(cluster, namespace, fresh)
}

func hasFreshRuntimeSample(samples []MetricSample) bool {
	for _, sample := range samples {
		if sample.Labels["container_id"] != "" || sample.Labels["network_source"] != "" {
			return true
		}
	}
	return false
}
