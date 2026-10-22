package k8s

import (
	"context"
	"strconv"
	"sync"
	"time"
)

type PrometheusMetricReader struct {
	Query    func(context.Context, string) (map[string]interface{}, error)
	Interval time.Duration
}

type metricQueryCacheKey struct{}
type metricQueryCache struct {
	mu      sync.Mutex
	results map[string]map[string]interface{}
}

func (r PrometheusMetricReader) Read(ctx context.Context, rule AlertRule, container Container) (*float64, string, bool) {
	query, err := MetricQuery(container.Key.ClusterID, container.Key.Namespace, rule.Metric, r.Interval)
	if err != nil || r.Query == nil || ctx.Err() != nil {
		return nil, "", false
	}
	var result map[string]interface{}
	if cache, ok := ctx.Value(metricQueryCacheKey{}).(*metricQueryCache); ok {
		cache.mu.Lock()
		var found bool
		result, found = cache.results[query]
		if !found {
			result, err = r.Query(ctx, query)
			if err != nil {
				result = nil
			}
			cache.results[query] = result
		}
		cache.mu.Unlock()
	} else {
		result, err = r.Query(ctx, query)
	}
	if err != nil || result["status"] != "success" {
		return nil, "", false
	}
	data, _ := result["data"].(map[string]interface{})
	rows, _ := data["result"].([]interface{})
	for _, raw := range rows {
		row, _ := raw.(map[string]interface{})
		labels, _ := row["metric"].(map[string]interface{})
		pod, _ := labels["pod"].(string)
		uid, _ := labels["pod_uid"].(string)
		name, _ := labels["container"].(string)
		if !MetricLabelsMatchContainer(map[string]string{"pod": pod, "pod_uid": uid, "container": name}, container) {
			continue
		}
		value, _ := row["value"].([]interface{})
		if len(value) != 2 {
			continue
		}
		text, _ := value[1].(string)
		v, parseErr := strconv.ParseFloat(text, 64)
		if parseErr == nil {
			return &v, uid, true
		}
	}
	return nil, "", false
}
