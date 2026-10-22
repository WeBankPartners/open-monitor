package k8s

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Exporter struct {
	index     *Index
	mu        sync.RWMutex
	samples   map[scope][]MetricSample
	updated   map[scope]time.Time
	failures  map[scope]int
	now       func() time.Time
	freshness time.Duration
}

func NewExporter(index *Index) *Exporter {
	return &Exporter{index: index, samples: map[scope][]MetricSample{}, updated: map[scope]time.Time{}, failures: map[scope]int{}, now: time.Now, freshness: time.Minute}
}

func (e *Exporter) SetFreshness(freshness time.Duration) {
	if e == nil || freshness <= 0 {
		return
	}
	e.mu.Lock()
	e.freshness = freshness
	e.mu.Unlock()
}
func (e *Exporter) Set(cluster, namespace string, samples []MetricSample) error {
	if e == nil || e.index == nil {
		return errors.New("metric exporter unavailable")
	}
	if _, err := RenderMetricSamples(samples); err != nil {
		return err
	}
	cp := append([]MetricSample(nil), samples...)
	for i := range cp {
		labels := map[string]string{}
		for k, v := range cp[i].Labels {
			labels[k] = v
		}
		cp[i].Labels = labels
	}
	key := scope{cluster, namespace}
	e.mu.Lock()
	e.samples[key] = cp
	e.updated[key] = e.now()
	e.failures[key] = 0
	e.mu.Unlock()
	return nil
}
func (e *Exporter) Fail(cluster, namespace string) {
	key := scope{cluster, namespace}
	e.mu.Lock()
	e.samples[key] = nil
	e.updated[key] = e.now()
	e.failures[key]++
	e.mu.Unlock()
}
func (e *Exporter) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	e.mu.RLock()
	all := []MetricSample{}
	states := []MetricSample{}
	now := e.now()
	keys := make([]scope, 0, len(e.samples))
	for key := range e.samples {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].cluster+keys[i].namespace < keys[j].cluster+keys[j].namespace })
	for _, key := range keys {
		current, ok := e.index.Current(key.cluster, key.namespace)
		if !ok {
			continue
		}
		if !current.DiscoveryFailed {
			all = append(all, SnapshotMetricSamples(current)...)
		}

		good := float64(1)
		if current.DiscoveryFailed || now.Sub(e.updated[key]) > e.freshness || e.failures[key] > 0 {
			good = 0
		}
		states = append(states, MetricSample{Name: "k8s_source_good", Value: good, Labels: map[string]string{"cluster_id": key.cluster, "namespace": key.namespace}})
		alarm := float64(0)
		if e.failures[key] >= 3 {
			alarm = 1
		}
		states = append(states, MetricSample{Name: "k8s_source_failure_threshold", Value: alarm, Labels: map[string]string{"cluster_id": key.cluster, "namespace": key.namespace}})
		if good == 0 {
			states = append(states, MetricSample{Name: "k8s_source_failed", Value: float64(e.failures[key]), Labels: map[string]string{"cluster_id": key.cluster, "namespace": key.namespace}})
			continue
		}

		allowed := map[string]bool{}
		for _, c := range current.Containers {
			allowed[c.PodUID+"\x00"+c.ContainerID] = true
		}
		for _, sample := range e.samples[key] {
			podAllowed := false
			if sample.Labels["container"] == "" && sample.Labels["container_id"] == "" {
				for identity := range allowed {
					if strings.HasPrefix(identity, sample.Labels["pod_uid"]+"\x00") {
						podAllowed = true
						break
					}
				}
			} else {
				podAllowed = allowed[sample.Labels["pod_uid"]+"\x00"+sample.Labels["container_id"]]
			}
			if podAllowed {
				if sample.Timestamp == 0 || (sample.Timestamp <= now.UnixMilli()+5000 && now.UnixMilli()-sample.Timestamp <= e.freshness.Milliseconds()) {
					sample.Timestamp = 0
					all = append(all, sample)
				}
			}
		}
	}
	e.mu.RUnlock()
	all = append(all, states...)
	body, err := RenderMetricSamples(all)
	if err != nil {
		http.Error(w, "metrics unavailable", 500)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(body))
}

func (e *Exporter) Clear() {
	e.mu.Lock()
	e.samples = map[scope][]MetricSample{}
	e.updated = map[scope]time.Time{}
	e.failures = map[scope]int{}
	e.mu.Unlock()
}
func (e *Exporter) Prune(rows []NamespaceState) {
	allowed := map[scope]bool{}
	for _, r := range rows {
		allowed[scope{r.ClusterID, r.Namespace}] = true
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for key := range e.samples {
		if !allowed[key] {
			delete(e.samples, key)
			delete(e.updated, key)
			delete(e.failures, key)
		}
	}
}
