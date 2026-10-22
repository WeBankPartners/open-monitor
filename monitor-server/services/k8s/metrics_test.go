package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestParseCAdvisorFiltersRegisteredContainersAndNormalizesCPUAndMemory(t *testing.T) {
	snapshot := &Snapshot{Containers: []Container{{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-new", ContainerID: "runtime-new", Node: "node-a"}}}
	payload := "# HELP container_cpu_usage_seconds_total CPU\n" +
		"container_cpu_usage_seconds_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-new\"} 12 1700000000000\n" +
		"container_cpu_usage_seconds_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/old\"} 99 1700000000000\n" +
		"container_memory_working_set_bytes{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-new\"} 2048 1700000000000\n" +
		"container_memory_working_set_bytes{namespace=\"sit-a\",pod=\"other-0\",container=\"app\",id=\"/kubepods/other\"} 4096 1700000000000\n"
	samples, err := ParseCAdvisor(payload, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("got %d samples", len(samples))
	}
	if samples[0].Name != "k8s_pod_cpu_usage_seconds_total" || samples[1].Name != "k8s_pod_memory_working_set_bytes" {
		t.Fatalf("unexpected names: %+v", samples)
	}
	if samples[0].Labels["pod_uid"] != "pod-new" || samples[0].Labels["container_id"] != "runtime-new" {
		t.Fatalf("missing runtime labels: %+v", samples[0].Labels)
	}
}

func TestParseCAdvisorRejectsMalformedSampleAndReportsNoPartialData(t *testing.T) {
	_, err := ParseCAdvisor(`container_cpu_usage_seconds_total{namespace="sit-a" 1`, &Snapshot{})
	if err == nil {
		t.Fatal("malformed exposition accepted")
	}
}

func TestRenderMetricSamplesEscapesLabels(t *testing.T) {
	out, err := RenderMetricSamples([]MetricSample{{Name: "k8s_pod_ready", Value: 1, Labels: map[string]string{"namespace": "sit-a", "pod": "orders-0", "quote": "a\\\"b"}}})
	if err != nil {
		t.Fatal(err)
	}
	want := "k8s_pod_ready{namespace=\"sit-a\",pod=\"orders-0\",quote=\"a\\\\\\\"b\"} 1\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

type metricSourceFunc func(context.Context, string) (string, error)

func (f metricSourceFunc) Fetch(ctx context.Context, node string) (string, error) {
	return f(ctx, node)
}

func TestCollectorFetchesEachRegisteredNodeOnceAndKeepsOnlyCurrentRuntime(t *testing.T) {
	snapshot := &Snapshot{Containers: []Container{
		{Key: ObjectKey{Namespace: "sit-a", PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-a", ContainerID: "runtime-a", Node: "node-a"},
		{Key: ObjectKey{Namespace: "sit-a", PodName: "orders-1", ContainerName: "app"}, PodUID: "pod-b", ContainerID: "runtime-b", Node: "node-a"},
	}}
	calls := 0
	collector := Collector{Source: metricSourceFunc(func(_ context.Context, node string) (string, error) {
		calls++
		if node != "node-a" {
			t.Fatalf("unexpected node %s", node)
		}
		return "container_memory_working_set_bytes{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"runtime-a\"} 4\n", nil
	})}
	samples, err := collector.Collect(context.Background(), snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(samples) != 1 || samples[0].Value != 4 {
		t.Fatalf("unexpected collection: calls=%d samples=%+v", calls, samples)
	}
}

func TestParseCAdvisorFiltersNetworkByPodAndInterfaceAndBlockIOByDeviceOperation(t *testing.T) {
	snapshot := &Snapshot{Containers: []Container{
		{Key: ObjectKey{ClusterID: "cluster-a", Namespace: "sit-a", WorkloadType: "StatefulSet", WorkloadName: "orders", PodName: "orders-0", ContainerName: "app"}, PodUID: "pod-a", ContainerID: "runtime-a", Node: "node-a"},
	}}
	payload := "container_network_receive_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"\",interface=\"eth0\",id=\"/kubepods.slice/podpod-a.slice\"} 10\n" +
		"container_network_receive_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"\",interface=\"eth0\",id=\"/kubepods.slice/podpod-a.slice\"} 11\n" +
		"container_network_transmit_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"\",interface=\"eth1\",id=\"/kubepods.slice/podpod-a.slice\"} 12\n" +
		"container_network_receive_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"\",interface=\"eth0\",id=\"/kubepods.slice/podold.slice\"} 99\n" +
		"container_blkio_device_usage_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",device=\"/dev/vda\",operation=\"Read\",id=\"/kubepods/runtime-a\"} 20\n" +
		"container_blkio_device_usage_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",device=\"/dev/vda\",operation=\"Write\",id=\"/kubepods/runtime-a\"} 21\n" +
		"container_blkio_device_usage_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",device=\"\",operation=\"Read\",id=\"/kubepods/runtime-a\"} 22\n"
	samples, err := ParseCAdvisor(payload, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 4 {
		t.Fatalf("got %d samples: %+v", len(samples), samples)
	}
	seen := map[string]bool{}
	for _, sample := range samples {
		seen[sample.Name+"/"+sample.Labels["interface"]+"/"+sample.Labels["device"]+"/"+sample.Labels["operation"]] = true
	}
	for _, want := range []string{"k8s_pod_network_receive_bytes_total/eth0//", "k8s_pod_network_transmit_bytes_total/eth1//", "k8s_pod_block_read_bytes_total///dev/vda/Read", "k8s_pod_block_write_bytes_total///dev/vda/Write"} {
		if !seen[want] {
			t.Fatalf("missing %s in %+v", want, seen)
		}
	}
}

// A substring UID match can attach a deleted Pod's counters to its replacement.
func TestNetworkRejectsAmbiguousPodIdentity(t *testing.T) {
	for _, id := range []string{"/kubepods/podpod-a-old/sandbox", "/unrelated/pod-a", "/kubepods/podold/runtime-a"} {
		payload := "container_network_receive_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",interface=\"eth0\",id=" + strconv.Quote(id) + "} 99\n"
		samples, err := ParseCAdvisor(payload, metricSnapshot())
		if err != nil || len(samples) != 0 {
			t.Fatalf("accepted ambiguous cgroup %s: %+v %v", id, samples, err)
		}
	}
}

func TestBlockIOKeepsEveryContainerAndDevice(t *testing.T) {
	s := metricSnapshot()
	other := s.Containers[0]
	other.Key.ContainerName = "sidecar"
	other.ContainerID = "runtime-b"
	s.Containers = append(s.Containers, other)
	payload := ""
	for _, container := range []struct{ name, id string }{{"app", "runtime-a"}, {"sidecar", "runtime-b"}} {
		for _, device := range []string{"/dev/vda", "/dev/vdb"} {
			payload += fmt.Sprintf("container_blkio_device_usage_total{namespace=\"sit-a\",pod=\"orders-0\",container=%q,device=%q,operation=\"Read\",id=%q} 20\n", container.name, device, container.id)
		}
	}
	samples, err := ParseCAdvisor(payload, s)
	if err != nil || len(samples) != 4 {
		t.Fatalf("lost container/device series: %+v %v", samples, err)
	}
}

func TestNetworkIsPodScopedAndSelectsOneSourceDeterministically(t *testing.T) {
	s := metricSnapshot()
	a := "container_network_receive_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"POD\",interface=\"eth0\",id=\"/kubepods/podpod-a/sandbox-a\"} 10\n"
	b := "container_network_receive_bytes_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",interface=\"eth0\",id=\"/kubepods/podpod-a/runtime-a\"} 10\n"
	for _, payload := range []string{a + b, b + a} {
		samples, err := ParseCAdvisor(payload, s)
		if err != nil || len(samples) != 1 {
			t.Fatalf("network duplicated or dropped: %+v %v", samples, err)
		}
		if samples[0].Labels["container"] != "" || samples[0].Labels["container_id"] != "" || samples[0].Labels["network_source"] != "/kubepods/podpod-a/sandbox-a" {
			t.Fatalf("network assigned to business container or unstable source: %+v", samples[0])
		}
	}
}

func TestParseCAdvisorIncludesThrottlingOOMAndMemoryFailureMetrics(t *testing.T) {
	s := metricSnapshot()
	payload := "container_cpu_cfs_periods_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 100\n" +
		"container_cpu_cfs_throttled_periods_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 4\n" +
		"container_cpu_cfs_throttled_seconds_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 2\n" +
		"container_oom_events_total{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 1\n" +
		"container_memory_failcnt{namespace=\"sit-a\",pod=\"orders-0\",container=\"app\",id=\"/kubepods/runtime-a\"} 3\n"
	samples, err := ParseCAdvisor(payload, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 5 {
		t.Fatalf("got %d samples: %+v", len(samples), samples)
	}
}

func TestSnapshotMetricSamplesExposeResourcesAndPodState(t *testing.T) {
	s := metricSnapshot()
	s.Containers[0].Resources.Requests = map[string]string{"cpu": "250m", "memory": "64Mi"}
	s.Containers[0].Resources.Limits = map[string]string{"cpu": "1", "memory": "128Mi"}
	s.Containers[0].RestartCount = 3
	s.Containers[0].StatusKnown = true
	s.Containers[0].PodReady = "True"
	s.Containers[0].Ready = true
	s.Containers[0].Phase = "Running"
	samples := SnapshotMetricSamples(s)
	if len(samples) == 0 {
		t.Fatal("no snapshot samples")
	}
	out, err := RenderMetricSamples(samples)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"k8s_pod_cpu_request_cores", "k8s_pod_memory_limit_bytes", "k8s_pod_restart_count", "k8s_pod_ready", "k8s_pod_phase"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s: %s", want, out)
		}
	}
}
