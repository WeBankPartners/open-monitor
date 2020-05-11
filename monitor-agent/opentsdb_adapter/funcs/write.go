package funcs

import (
	"net/http"
	"math"
	"net/url"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"log"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"context"
	"time"
)

const (
	putEndpoint     = "/api/put"
	contentTypeJSON = "application/json"
)

var OpenTSDBUrl string

func write(w http.ResponseWriter,r *http.Request)  {
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read error %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		log.Printf("Decode error %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req WriteRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		log.Printf("Unmarshal error %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	samples := protoToSamples(&req)

	reqs := make([]StoreSamplesRequest, 0, len(samples))
	for _, s := range samples {
		v := float64(s.Value)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			log.Printf("cannot send value to OpenTSDB, skipping sample %v \n", s)
			continue
		}
		metric := TagValue(s.Metric["__name__"])
		reqs = append(reqs, StoreSamplesRequest{
			Metric:    metric,
			Timestamp: s.Timestamp.Unix(),
			Value:     v,
			Tags:      tagsFromMetric(s.Metric),
		})
	}

	u, err := url.Parse(OpenTSDBUrl)
	if err != nil {
		log.Printf("opentsdb url parse error %v", err)
		return
	}

	u.Path = putEndpoint

	buf, err := json.Marshal(reqs)
	if err != nil {
		log.Printf("request opentsdb json marshal error %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
	defer cancel()

	reqTsdb, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("request opentsdb http new request error %v", err)
		return
	}
	reqTsdb.Header.Set("Content-Type", contentTypeJSON)
	respTsdb, err := http.DefaultClient.Do(reqTsdb.WithContext(ctx))
	if err != nil {
		log.Printf("request opentsdb response error %v",err)
		return
	}
	if respTsdb.StatusCode == http.StatusNoContent {
		return
	}else{
		log.Printf("request opentsdb response status code -> %d", respTsdb.StatusCode)
	}
}

func protoToSamples(req *WriteRequest) Samples {
	var samples Samples
	for _, ts := range req.Timeseries {
		metric := make(Metric, len(ts.Labels))
		for _, l := range ts.Labels {
			metric[LabelName(l.Name)] = LabelValue(l.Value)
		}

		for _, s := range ts.Samples {
			samples = append(samples, &PrometheusSample{
				Metric:    metric,
				Value:     SampleValue(s.Value),
				Timestamp: Time(s.Timestamp),
			})
		}
	}
	return samples
}

func tagsFromMetric(m Metric) map[string]TagValue {
	tags := make(map[string]TagValue, len(m)-1)
	for l, v := range m {
		if l == "__name__" {
			continue
		}
		tags[string(l)] = TagValue(v)
	}
	return tags
}