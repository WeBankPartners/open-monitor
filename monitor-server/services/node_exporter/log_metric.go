package node_exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"io/ioutil"
	"net/http"
)

func UpdateNodeExportConfig(endpoints []string) error {
	return nil
	var err error
	existMap := make(map[string]int)
	for _,v := range endpoints {
		if _,b:=existMap[v];b {
			continue
		}
		err = updateEndpointLogMetric(v)
		if err != nil {
			err = fmt.Errorf("Sync endpoint:%s log metric config fail,%s ", v, err.Error())
			break
		}
		existMap[v] = 1
	}
	return err
}

func updateEndpointLogMetric(endpointGuid string) error {
	logMetricConfig,err := db.GetLogMetricByEndpoint(endpointGuid)
	if err != nil {
		return fmt.Errorf("Query endpoint:%s log metric config fail,%s ", endpointGuid, err.Error())
	}
	endpointObj := models.EndpointNewTable{Guid: endpointGuid}
	err = db.GetEndpointNew(&endpointObj)
	if err != nil || endpointObj.AgentAddress == "" {
		return err
	}
	b,_ := json.Marshal(logMetricConfig)
	req,_ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/business/config", endpointObj.AgentAddress), bytes.NewReader(b))
	resp,respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("Do http request to %s fail,%s ", endpointObj.AgentAddress, respErr.Error())
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Do http request to %s fail,status code:%d ", endpointObj.AgentAddress, resp.StatusCode)
	}
	b,_ = ioutil.ReadAll(resp.Body)
	var response models.LogMetricNodeExporterResponse
	err = json.Unmarshal(b, &response)
	if err == nil {
		if response.Status == "OK" {
			return nil
		}else{
			return fmt.Errorf(response.Message)
		}
	}
	return fmt.Errorf("json unmarhsal reponse body fail,%s ", err.Error())
}