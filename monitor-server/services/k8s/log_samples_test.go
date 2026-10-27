package k8s

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLogSampleQueryKeepsHistoricalSamplesWithoutChartStepSampling(t *testing.T) {
	query, err := LogSampleQuery("cluster-a", "ns-a", MetricObjectFilter{PodName: "app-0", Container: "main"}, 1000, 1900)
	// Collapse duplicate scrape instances before topk, so they cannot consume
	// slots that should contain older distinct samples.
	want := `topk(201, (max without (job, instance) (max_over_time(k8s_log_match_sample_timestamp_seconds{job="k8s_log_agent",cluster_id="cluster-a",namespace="ns-a",pod="app-0",container="main"}[901s])) >= 1000) <= 1900)`
	if err != nil || query != want {
		t.Fatalf("query=%s error=%v", query, err)
	}
	for _, object := range []MetricObjectFilter{{Container: "main"}, {PodName: "app-0"}, {PodName: `x"} or up{`, Container: "main"}, {PodName: "app-0", Container: "main", WorkloadName: strings.Repeat("a", 254)}} {
		if _, err := LogSampleQuery("cluster-a", "ns-a", object, 1000, 1900); err == nil {
			t.Fatalf("accepted invalid selector: %+v", object)
		}
	}
}

func sampleResponse(t *testing.T, body string) map[string]interface{} {
	t.Helper()
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestLogSamplesUseObservationTimeDeduplicateScrapesAndKeepByteOffsets(t *testing.T) {
	raw := sampleResponse(t, `{"status":"success","data":{"resultType":"vector","result":[
		{"metric":{"pod_uid":"pod-old","rule":"error","file":"app.log","file_id":"device:inode","byte_offset":"9007199254740993","preview":"ERROR old","instance":"agent-a"},"value":[1900,"1100"]},
		{"metric":{"pod_uid":"pod-new","rule":"error","file":"app.log","file_id":"device:inode","byte_offset":"2","preview":"ERROR <script>text</script>"},"value":[1900,"1800"]},
		{"metric":{"pod_uid":"pod-old","rule":"error","file":"app.log","file_id":"device:inode","byte_offset":"9007199254740993","preview":"ERROR old","instance":"agent-b"},"value":[1900,"1100"]},
		{"metric":{"preview":"outside"},"value":[1900,"999"]}
	]}}`)
	page, err := ReadLogSamples(raw, 1000, 1900)
	if err != nil || len(page.Samples) != 2 || page.Limited {
		t.Fatalf("page=%+v error=%v", page, err)
	}
	if page.Samples[0].ObservedAt != 1800 || page.Samples[1].Labels["byte_offset"] != "9007199254740993" || page.Samples[1].Labels["pod_uid"] != "pod-old" {
		t.Fatalf("lost observation time, byte position, or incarnation: %+v", page)
	}
}

func TestLogSamplesReportEmptyFailureAndResultLimitSeparately(t *testing.T) {
	page, err := ReadLogSamples(sampleResponse(t, `{"status":"success","data":{"resultType":"vector","result":[]}}`), 1000, 1900)
	if err != nil || page.Samples == nil || len(page.Samples) != 0 {
		t.Fatalf("empty query failed: %+v %v", page, err)
	}
	for _, body := range []string{`{"status":"error","error":"query failed"}`, `{"status":"success","data":{"resultType":"matrix","result":[]}}`, `{"status":"success","data":{"resultType":"vector","result":[{"value":[1900,"NaN"]}]}}`} {
		if _, err = ReadLogSamples(sampleResponse(t, body), 1000, 1900); err == nil {
			t.Fatalf("upstream failure appeared as an empty success: %s", body)
		}
	}
	rows := []interface{}{}
	for i := 0; i < 201; i++ {
		rows = append(rows, map[string]interface{}{"metric": map[string]string{"rule": "error", "preview": strings.Repeat("x", i+1)}, "value": []interface{}{1900, "1800"}})
	}
	page, err = ReadLogSamples(map[string]interface{}{"status": "success", "data": map[string]interface{}{"resultType": "vector", "result": rows}}, 1000, 1900)
	if err != nil || len(page.Samples) != 200 || !page.Limited || page.Limit != 200 {
		t.Fatalf("unbounded sample response: %+v %v", page, err)
	}
}
