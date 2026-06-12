package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
)

func QueryClusterConfig(id string) (result []*m.ClusterTable, err error) {
	var filterSql = ""
	var filterParams []interface{}
	if id != "" {
		filterSql += " and id=? "
		filterParams = append(filterParams, id)
	}
	result = []*m.ClusterTable{}
	err = x.SQL("select * from cluster where 1=1 "+filterSql, filterParams...).Find(&result)
	if err != nil {
		err = fmt.Errorf("Try to query cluster table fail,%s ", err.Error())
	}
	return
}

func GetClusterAddress(cluster string) string {
	if cluster == "default" || cluster == "" {
		return "127.0.0.1:9090"
	}
	var clusterTable []*m.ClusterTable
	x.SQL("select * from cluster where id=?", cluster).Find(&clusterTable)
	if len(clusterTable) > 0 {
		return clusterTable[0].PromAddress
	}
	return ""
}

// Service discover functions
func SyncSdEndpointNew(steps []int, cluster string, fromPeer bool) error {
	log.Info(nil, log.LOGGER_APP, "Start sync sd endpoint", zap.String("steps", fmt.Sprintf("%v", steps)), zap.String("cluster", cluster))
	var syncList []*m.SdConfigSyncObj
	var err error
	for _, step := range steps {
		tmpSdFileList, tmpErr := GetSdFileListByStep(step, cluster)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		log.Info(nil, log.LOGGER_APP, "Get sd file list content", zap.Int("step", step), log.JsonObj("sdFileList", tmpSdFileList))
		if len(tmpSdFileList) <= 0 {
			continue
		}
		syncList = append(syncList, &m.SdConfigSyncObj{Step: step, Content: string(tmpSdFileList.TurnToFileSdConfigByte(step))})
	}
	if err != nil {
		return err
	}
	if len(syncList) == 0 {
		log.Warn(nil, log.LOGGER_APP, "Sync sd endpoint break,sync list is empty")
		return nil
	}
	if cluster == "" || cluster == "default" {
		prom.SyncLocalSdConfig(m.SdLocalConfigJob{Configs: syncList, FromPeer: fromPeer})
	} else {
		err = SyncRemoteSdConfigFile(cluster, syncList)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Sync remote sd config file fail", zap.Error(err))
		}
	}
	return err
}

func SyncRemoteSdConfigFile(cluster string, params []*m.SdConfigSyncObj) error {
	clusterObj, err := QueryClusterConfig(cluster)
	if err != nil {
		return err
	}
	if len(clusterObj) == 0 {
		return fmt.Errorf("Can not find cluster:%s in database ", cluster)
	}
	bodyBytes, _ := json.Marshal(params)
	req, reqErr := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/cluster-agent/accept/service-discover", clusterObj[0].RemoteAgentAddress), strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return fmt.Errorf("Try to new http request fail,%s ", reqErr.Error())
	}
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("Try to do request fail,%s ", respErr.Error())
	}
	respBodyBytes, bodyErr := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		resp.Body.Close()
	}
	if bodyErr != nil {
		return fmt.Errorf("Try to read response body fail,%s ", bodyErr.Error())
	}
	var respObj m.ClusterAgentResponse
	unmarshalErr := json.Unmarshal(respBodyBytes, &respObj)
	if unmarshalErr != nil {
		return fmt.Errorf("Try to json unmarshal response json fail,%s ", unmarshalErr.Error())
	}
	if respObj.Status != "OK" {
		return fmt.Errorf(respObj.Message)
	}
	return nil
}

func GetSdFileListByStep(step int, cluster string) (result m.ServiceDiscoverFileList, err error) {
	if cluster == "" {
		cluster = "default"
	}
	var endpointTables []*m.EndpointNewTable
	err = x.SQL("select * from endpoint_new where step=? and cluster=?", step, cluster).Find(&endpointTables)
	if err != nil {
		err = fmt.Errorf("Try to query endpoint table fail,%s ", err.Error())
		return
	}
	result = m.ServiceDiscoverFileList{}
	for _, v := range endpointTables {
		if v.MonitorType == "snmp" || v.MonitorType == "process" || v.MonitorType == "custom" || v.MonitorType == "pod" {
			continue
		}
		// 对于所有类型，如果 AgentAddress 为空则跳过，避免生成空 targets
		if v.AgentAddress == "" {
			continue
		}
		if v.MonitorType == "ping" || v.MonitorType == "telnet" || v.MonitorType == "http" {
			if v.Ip != "" && strings.Contains(v.AgentAddress, v.Ip) {
				continue
			}
		}
		tmpSdFileObj := m.ServiceDiscoverFileObj{Guid: v.Guid, Step: v.Step, Cluster: v.Cluster, Address: v.AgentAddress}
		log.Info(nil, log.LOGGER_APP, "add endpoint", zap.String("guid", v.Guid))
		result = append(result, &tmpSdFileObj)
	}
	return
}

func SyncRemoteRuleConfigFile(cluster string, param m.RFClusterRequestObj) error {
	clusterObj, err := QueryClusterConfig(cluster)
	if err != nil {
		return err
	}
	if len(clusterObj) == 0 {
		return fmt.Errorf("Can not find cluster:%s in database ", cluster)
	}
	bodyBytes, _ := json.Marshal(param)
	req, reqErr := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/cluster-agent/accept/rule", clusterObj[0].RemoteAgentAddress), strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return fmt.Errorf("Try to new http request fail,%s ", reqErr.Error())
	}
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("Try to do request fail,%s ", respErr.Error())
	}
	respBodyBytes, bodyErr := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		resp.Body.Close()
	}
	if bodyErr != nil {
		return fmt.Errorf("Try to read response body fail,%s ", bodyErr.Error())
	}
	var respObj m.ClusterAgentResponse
	unmarshalErr := json.Unmarshal(respBodyBytes, &respObj)
	if unmarshalErr != nil {
		return fmt.Errorf("Try to json unmarshal response json fail,%s ", unmarshalErr.Error())
	}
	if respObj.Status != "OK" {
		return fmt.Errorf(respObj.Message)
	}
	return nil
}
