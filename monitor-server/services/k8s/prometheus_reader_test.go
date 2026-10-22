package k8s

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPrometheusReaderQueriesEachMetricOncePerEvaluation(t *testing.T) {
	snapshot := &Snapshot{}
	rows := []interface{}{}
	for n := 0; n < 400; n++ {
		pod, uid := fmt.Sprintf("app-%d", n), fmt.Sprintf("uid-%d", n)
		snapshot.Containers = append(snapshot.Containers, Container{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", PodName: pod, ContainerName: "app"}, PodUID: uid})
		rows = append(rows, map[string]interface{}{"metric": map[string]interface{}{"pod": pod, "pod_uid": uid, "container": "app"}, "value": []interface{}{1.0, "2"}})
	}
	calls := 0
	reader := PrometheusMetricReader{Interval: 30 * time.Second, Query: func(ctx context.Context, query string) (map[string]interface{}, error) {
		calls++
		return map[string]interface{}{"status": "success", "data": map[string]interface{}{"result": rows}}, nil
	}}
	rules := []AlertRule{{ID: "memory", Metric: "memory", Enabled: true, Operator: ">", Threshold: 1}, {ID: "other-memory", Metric: "memory", Enabled: true, Operator: ">", Threshold: 3}, {ID: "ready", Metric: "ready", Enabled: true, Operator: "<", Threshold: 1}}
	for cycle := 1; cycle <= 2; cycle++ {
		sink := &testAlertSink{}
		if err := EvaluateSnapshot(context.Background(), snapshot, rules, reader, sink, time.Now()); err != nil {
			t.Fatal(err)
		}
		if calls != 2*cycle {
			t.Fatalf("expected two queries per evaluation, got %d after cycle %d", calls, cycle)
		}
		if len(sink.inputs) != 1200 || !sink.inputs[0].Observation.Violated || sink.inputs[1].Observation.Violated {
			t.Fatalf("invalid evaluations: %d", len(sink.inputs))
		}
	}
}

func TestPrometheusReaderSharesFailuresOnlyWithinCurrentEvaluation(t *testing.T) {
	calls := 0
	reader := PrometheusMetricReader{Query: func(ctx context.Context, query string) (map[string]interface{}, error) {
		calls++
		return nil, errors.New("unavailable")
	}}
	snapshot := metricSnapshot()
	rules := []AlertRule{{ID: "one", Metric: "memory", Enabled: true}, {ID: "two", Metric: "memory", Enabled: true}}
	for n := 1; n <= 2; n++ {
		sink := &testAlertSink{}
		if err := EvaluateSnapshot(context.Background(), snapshot, rules, reader, sink, time.Now()); err != nil {
			t.Fatal(err)
		}
		if calls != n || len(sink.inputs) != 2 || sink.inputs[0].Observation.Valid {
			t.Fatalf("invalid failure handling: calls=%d inputs=%+v", calls, sink.inputs)
		}
	}
}
