package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"io/ioutil"
	"net/http"
	"strings"
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

// Service discover functions
func SyncSdEndpointNew(steps []int, cluster string) error {
	log.Logger.Info("Add sd endpoint", log.String("steps", fmt.Sprintf("%v", steps)), log.String("cluster",cluster))
	var syncList []*m.SdConfigSyncObj
	var err error
	for _, step := range steps {
		tmpSdFileList, tmpErr := GetSdFileListByStep(step, cluster)
		if tmpErr != nil {
			err = tmpErr
			break
		}
		log.Logger.Info("Get sd file list content", log.Int("step", step), log.JsonObj("sdFileList", tmpSdFileList))
		syncList = append(syncList, &m.SdConfigSyncObj{Step: step, Content: string(tmpSdFileList.TurnToFileSdConfigByte(step))})
	}
	if err != nil {
		return err
	}
	if cluster == "" || cluster == "default" {
		err = SyncLocalSdConfigFile(syncList)
		if err != nil {
			log.Logger.Error("Sync local sd config file fail", log.Error(err))
		}
	} else {
		err = SyncRemoteSdConfigFile(cluster, syncList)
		if err != nil {
			log.Logger.Error("Sync remote sd config file fail", log.Error(err))
		}
	}
	return err
}

func DeleteSdEndpointNew(step int, cluster string) error {
	sdFileList, err := GetSdFileListByStep(step, cluster)
	if err != nil {
		return err
	}
	syncList := []*m.SdConfigSyncObj{&m.SdConfigSyncObj{Step: step, Content: string(sdFileList.TurnToFileSdConfigByte(step))}}
	if cluster == "" || cluster == "default" {
		err = SyncLocalSdConfigFile(syncList)
	} else {
		err = SyncRemoteSdConfigFile(cluster, syncList)
	}
	return err
}

func SyncLocalSdConfigFile(params []*m.SdConfigSyncObj) error {
	var err error
	fileSdPath := m.Config().SdFile.Path
	if fileSdPath != "" {
		if fileSdPath[len(fileSdPath)-1:] != "/" {
			fileSdPath = fileSdPath + "/"
		}
	}
	for _, param := range params {
		configFile := fmt.Sprintf("%ssd_file_%d.json", fileSdPath, param.Step)
		writeErr := ioutil.WriteFile(configFile, []byte(param.Content), 0644)
		if writeErr != nil {
			err = fmt.Errorf("Try to write sd file fail,%s ", writeErr.Error())
			break
		}
	}
	if err != nil {
		return err
	}
	err = prom.ReloadConfig()
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
	if bodyErr != nil {
		return fmt.Errorf("Try to read response body fail,%s ", bodyErr.Error())
	}
	resp.Body.Close()
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
	var endpointTables []*m.EndpointTable
	err = x.SQL("select guid,address,address_agent,step,cluster from endpoint where step=? and cluster=?", step, cluster).Find(&endpointTables)
	if err != nil {
		err = fmt.Errorf("Try to query endpoint table fail,%s ", err.Error())
		return
	}
	for _, v := range endpointTables {
		tmpSdFileObj := m.ServiceDiscoverFileObj{Guid: v.Guid, Step: v.Step, Cluster: v.Cluster, Address: v.Address}
		if v.AddressAgent != "" {
			tmpSdFileObj.Address = v.AddressAgent
		}
		log.Logger.Info("add endpoint", log.String("guid", v.Guid))
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
	if bodyErr != nil {
		return fmt.Errorf("Try to read response body fail,%s ", bodyErr.Error())
	}
	resp.Body.Close()
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
