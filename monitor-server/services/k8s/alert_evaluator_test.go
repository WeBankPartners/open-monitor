package k8s

import (
	"context"
	"testing"
	"time"
)

type testMetricReader struct {
	values map[string]float64
	valid  bool
	uid    string
}

func (r testMetricReader) Read(context.Context, AlertRule, Container) (*float64, string, bool) {
	v, ok := r.values["value"]
	if !ok {
		return nil, r.uid, r.valid
	}
	return &v, r.uid, r.valid
}

type testAlertSink struct{ inputs []AlertEvaluation }

func (s *testAlertSink) EvaluateAndPersist(_ context.Context, in AlertEvaluation, _ time.Time) (Transition, error) {
	s.inputs = append(s.inputs, in)
	return Transition{}, nil
}

func TestEvaluateSnapshotBuildsThresholdObservation(t *testing.T) {
	rule := AlertRule{ID: "cpu", Operator: ">", Threshold: 0.8, Enabled: true}
	v := 0.9
	sink := &testAlertSink{}
	err := EvaluateSnapshot(context.Background(), &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "c", Namespace: "n", WorkloadType: "StatefulSet", WorkloadName: "w", PodName: "w-0", ContainerName: "app"}, PodUID: "p"}}}, []AlertRule{rule}, testMetricReader{values: map[string]float64{"value": v}, valid: true, uid: "p"}, sink, time.Unix(1, 0))
	if err != nil || len(sink.inputs) != 1 || !sink.inputs[0].Observation.Violated || sink.inputs[0].Observation.Value == nil || *sink.inputs[0].Observation.Value != v {
		t.Fatalf("unexpected evaluation: err=%v inputs=%+v", err, sink.inputs)
	}
}

func TestEvaluateSnapshotMarksUIDMismatchUnknown(t *testing.T) {
	rule := AlertRule{ID: "cpu", Operator: ">", Threshold: 0.8, Enabled: true}
	sink := &testAlertSink{}
	err := EvaluateSnapshot(context.Background(), &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "c", Namespace: "n", WorkloadType: "StatefulSet", WorkloadName: "w", PodName: "w-0", ContainerName: "app"}, PodUID: "p"}}}, []AlertRule{rule}, testMetricReader{valid: true, uid: "other"}, sink, time.Unix(1, 0))
	if err != nil || len(sink.inputs) != 1 || sink.inputs[0].Observation.Valid {
		t.Fatalf("expected unknown observation: err=%v inputs=%+v", err, sink.inputs)
	}
}

func TestMetricLabelsMatchPodScopedHealthMetric(t *testing.T) {
	c := Container{Key: ObjectKey{PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-1"}
	if !MetricLabelsMatchContainer(map[string]string{"pod": "orders-0", "pod_uid": "pod-1"}, c) {
		t.Fatal("pod-scoped metric should match without a container label")
	}
	if MetricLabelsMatchContainer(map[string]string{"pod": "orders-0", "pod_uid": "pod-2"}, c) {
		t.Fatal("metric from another pod uid must not match")
	}
}
