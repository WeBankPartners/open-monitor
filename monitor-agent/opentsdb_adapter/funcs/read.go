package funcs

import (
	"net/http"
	"io/ioutil"
	"log"
	"encoding/json"
	"fmt"
	"strings"
	"strconv"
	"sort"
)

func read(w http.ResponseWriter,r *http.Request)  {
	var respMessage string
	var respError error
	var param QueryMonitorData
	serials := []*SerialModel{}
	b,err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	defer func() {
		var respData QueryResponseDto
		w.Header().Set("Content-Type", "application/json")
		if respMessage != "" {
			if respError == nil {
				log.Printf(respMessage)
				respData.Code = 1
				respData.Message = respMessage
				w.WriteHeader(http.StatusBadRequest)
			}else{
				log.Printf("%s -> %v", respMessage, respError)
				respData.Code = 2
				respData.Message = fmt.Sprintf("%s : %s", respMessage, respError)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}else{
			respData.Code = 0
			respData.Message = "success"
			w.WriteHeader(http.StatusOK)
		}
		respData.Data = serials
		rb,_ := json.Marshal(respData)
		w.Write(rb)
	}()
	if err != nil {
		respMessage = "read opentsdb check param read body error"
		respError = err
		return
	}
	err = json.Unmarshal(b, &param)
	if err != nil {
		respMessage = "read opentsdb check param unmarshal json error"
		respError = err
		return
	}
	validateMessage,err := checkParam(param)
	if err != nil {
		respMessage = validateMessage
		return
	}
	var tmpEndpointList,tmpMetricList []string
	for _,v := range param.Endpoint {
		tmpEndpointList = append(tmpEndpointList, transSpecial(v, false))
	}
	for _,v := range param.Metric {
		tmpMetricList = append(tmpMetricList, transSpecial(v, false))
	}
	param.Endpoint = tmpEndpointList
	param.Metric = tmpMetricList
	// Request openTSDB server
	var tsdbQuery OpenTsdbQuery
	recordMetric := param.Metric
	tsdbQuery.Start = param.Start
	tsdbQuery.End = param.End
	tsdbQuery.Queries = buildMetric(param)
	postData, err := json.Marshal(tsdbQuery)
	if err != nil {
		respMessage = "request opentsdb param json marshal error"
		respError = err
		return
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/query", OpenTSDBUrl), strings.NewReader(string(postData)))
	if err != nil {
		respMessage = "request opentsdb http new request error"
		respError = err
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp,err := http.DefaultClient.Do(req)
	if err != nil {
		respMessage = "request opentsdb response error"
		respError = err
		return
	}
	respBody,err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		respMessage = "request opentsdb read body error"
		respError = err
		return
	}
	if resp.StatusCode/100 != 2 {
		respMessage = fmt.Sprintf("request opentsdb error with response code %d", resp.StatusCode)
		respError = fmt.Errorf(respMessage)
		return
	}
	var tsdbResponse []OpenTsdbResponse
	err = json.Unmarshal(respBody, &tsdbResponse)
	if err != nil {
		respMessage = "request opentsdb json unmarshal response error"
		respError = err
		return
	}
	for _,otr := range tsdbResponse {
		var serial SerialModel
		if endpoint,ok := otr.Tags["instance"];ok{
			if len(otr.Tags) > 1 {
				var tmpList []string
				var tmpListString string
				for k,v := range otr.Tags {
					if k != "instance" {
						tmpList = append(tmpList, fmt.Sprintf("%s=%s", k, v))
						tmpListString += fmt.Sprintf("%s=%s,", k, v)
					}
				}
				newMetric := otr.Metric
				for _,v := range recordMetric {
					count := 0
					for _,vv := range tmpList {
						if strings.Contains(v, vv) {
							count = count + 1
						}
					}
					if count == len(tmpList) {
						newMetric = v
					}
				}
				if tmpListString != "" {
					tmpListString = tmpListString[:len(tmpListString)-1]
				}
				serial.Name = fmt.Sprintf("%s:%s[%s]", endpoint, newMetric, tmpListString)
			}else {
				serial.Name = fmt.Sprintf("%s:%s", endpoint, otr.Metric)
			}
		}else {
			serial.Name = otr.Metric
		}
		serial.Name = transSpecial(serial.Name, true)
		serial.Type = "line"
		sdata := DataSort{}
		for c,v := range otr.DataPoints {
			cfloat,err := strconv.ParseFloat(c, 64)
			if err==nil {
				vv, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", v), 64)
				clist := []float64{cfloat*1000, vv}
				sdata = append(sdata, clist)
			}
		}
		sort.Sort(sdata)
		serial.Data = sdata
		serials = append(serials, &serial)
	}
}

func checkParam(param QueryMonitorData) (string,error) {
	msg := "Param validate fail -> "
	if param.Start >= param.End {
		msg += "start >= end ? "
		return msg, fmt.Errorf(msg)
	}
	if len(param.Endpoint) == 0 {
		msg += "endpoint is empty "
		return msg, fmt.Errorf(msg)
	}
	if len(param.Metric) == 0 {
		msg += "metric is empty "
		return msg, fmt.Errorf(msg)
	}
	return "",nil
}

func buildMetric(query QueryMonitorData) []map[string]interface{} {
	metricsMap := make([]map[string]interface{}, 0)
	for _,metricTags := range query.Metric {
		metric := metricTags
		isTags := false
		if strings.Contains(metricTags, "/") {
			metric = strings.Split(metricTags, "/")[0]
			isTags = true
		}
		for _,endpoint := range query.Endpoint {
			tags := make(map[string]string)
			tags["instance"] = endpoint
			if isTags {
				tagString := strings.Replace(metricTags, metric+"/", "", -1)
				for _,tag := range strings.Split(tagString, ",") {
					tagList := strings.Split(tag, "=")
					tags[strings.ToLower(tagList[0])] = tagList[1]
				}
			}
			metricMap := make(map[string]interface{})
			metricMap["metric"] = metric
			metricMap["aggregator"] = "none"
			metricMap["tags"] = tags
			// 计算网络流量需要用到这种rate
			if query.ComputeRate {
				metricMap["rate"] = true
				rateOptions := make(map[string]interface{})
				rateOptions["counter"] = true
				rateOptions["dropResets"] = true
				metricMap["rateOptions"] = rateOptions
			}
			metricsMap = append(metricsMap, metricMap)
		}
	}
	return metricsMap
}

func transSpecial(input string, back bool) string {
	var output string
	if back {
		output = strings.Replace(input, "__", "_", -1)
		output = strings.Replace(output, "_.", ":", -1)
	}else{
		output = strings.Replace(input, "_", "__", -1)
		output = strings.Replace(output, ":", "_.", -1)
	}
	return output
}