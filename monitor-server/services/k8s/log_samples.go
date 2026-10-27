package k8s

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
)

const logSampleResultLimit = 200

type LogSample struct {
	Labels     map[string]string `json:"labels"`
	ObservedAt int64             `json:"observedAt"`
}

type LogSamplePage struct {
	Samples []LogSample `json:"samples"`
	Limited bool        `json:"limited"`
	Limit   int         `json:"limit"`
}

func LogSampleQuery(cluster, namespace string, object MetricObjectFilter, start, end int64) (string, error) {
	if len(cluster) > 64 || !clusterPattern.MatchString(cluster) || len(namespace) > 63 || !namespacePattern.MatchString(namespace) || object.PodName == "" || object.Container == "" || start < 0 || end <= start || end-start > 30*86400 {
		return "", errors.New("invalid log sample scope or range")
	}
	selector := "{job=\"k8s_log_agent\",cluster_id=" + strconv.Quote(cluster) + ",namespace=" + strconv.Quote(namespace)
	for _, field := range []struct{ label, value string }{
		{"workload_type", object.WorkloadType}, {"workload_name", object.WorkloadName}, {"pod", object.PodName}, {"container", object.Container},
	} {
		if field.value == "" {
			continue
		}
		if len(field.value) > 253 || !metricObjectName.MatchString(field.value) {
			return "", errors.New("invalid log sample object selector")
		}
		selector += "," + field.label + "=" + strconv.Quote(field.value)
	}
	// Evaluate at end, using observation time rather than the scrape timestamp.
	// A range function preserves short-lived series even for wide chart ranges.
	// Collapse duplicate scrape instances before limiting distinct samples.
	return fmt.Sprintf("topk(%d, (max without (job, instance) (max_over_time(k8s_log_match_sample_timestamp_seconds%s}[%ds])) >= %d) <= %d)", logSampleResultLimit+1, selector, end-start+1, start, end), nil
}

func ReadLogSamples(raw map[string]interface{}, start, end int64) (LogSamplePage, error) {
	page := LogSamplePage{Samples: []LogSample{}, Limit: logSampleResultLimit}
	var response struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []json.RawMessage `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return page, errors.New("invalid log sample response")
	}
	if json.Unmarshal(encoded, &response) != nil || response.Status != "success" || response.Data.ResultType != "vector" || len(response.Data.Result) > logSampleResultLimit+1 {
		return page, errors.New("invalid log sample response")
	}
	page.Limited = len(response.Data.Result) > logSampleResultLimit
	seen := map[string]bool{}
	for _, row := range response.Data.Result {
		var text string
		if len(row.Value) != 2 || json.Unmarshal(row.Value[1], &text) != nil {
			return page, errors.New("invalid log sample value")
		}
		value, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value != math.Trunc(value) {
			return page, errors.New("invalid log sample observation time")
		}
		if value < float64(start) || value > float64(end) {
			continue
		}
		if row.Metric == nil {
			return page, errors.New("invalid log sample labels")
		}
		delete(row.Metric, "__name__")
		sample := LogSample{Labels: row.Metric, ObservedAt: int64(value)}
		identity := map[string]string{}
		for key, label := range row.Metric {
			if key != "job" && key != "instance" {
				identity[key] = label
			}
		}
		key, _ := json.Marshal(LogSample{Labels: identity, ObservedAt: sample.ObservedAt})
		if !seen[string(key)] {
			seen[string(key)] = true
			page.Samples = append(page.Samples, sample)
		}
	}
	sort.SliceStable(page.Samples, func(i, j int) bool { return page.Samples[i].ObservedAt > page.Samples[j].ObservedAt })
	if len(page.Samples) > logSampleResultLimit {
		page.Samples = page.Samples[:logSampleResultLimit]
	}
	return page, nil
}
