package agent

import (
	"encoding/json"
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func UpdateKubernetesCluster(c *gin.Context) {
	var param m.KubernetesClusterParam
	operation := c.Param("operation")
	var err error
	if operation == "get" || operation == "list" {
		result, err := db.ListKubernetesCluster("")
		if err != nil {
			mid.ReturnHandleError(c, err.Error(), err)
		} else {
			mid.ReturnSuccessData(c, result)
		}
		return
	}
	if operation == "delete" {
		var tmpParam m.KubernetesClusterTable
		if err = c.ShouldBindJSON(&tmpParam); err == nil {
			if tmpParam.Id <= 0 {
				mid.ReturnParamEmptyError(c, "id")
				return
			}
			err = db.DeleteKubernetesCluster(tmpParam.Id, "")
		} else {
			mid.ReturnValidateError(c, err.Error())
			return
		}
	} else {
		if err = c.ShouldBindJSON(&param); err == nil {
			if mid.IsIllegalIp(param.Ip) {
				mid.ReturnValidateError(c, "param ip is illegal")
				return
			}
			portInt, _ := strconv.Atoi(param.Port)
			if portInt <= 0 {
				mid.ReturnValidateError(c, "param port is illegal")
				return
			}
			param.ClusterName = strings.TrimSpace(param.ClusterName)
			if !mid.IsIllegalNormalInput(param.ClusterName) {
				mid.ReturnValidateError(c, "param cluster_name is illegal")
				return
			}
			if operation == "update" {
				if param.Id <= 0 {
					mid.ReturnValidateError(c, "param id is empty")
					return
				}
				err = db.UpdateKubernetesCluster(param)
			} else {
				err = db.AddKubernetesCluster(param)
			}
		} else {
			mid.ReturnValidateError(c, err.Error())
			return
		}
	}
	if err != nil {
		mid.ReturnHandleError(c, err.Error(), err)
	} else {
		time.Sleep(20 * time.Second)
		db.SyncPodToEndpoint()
		mid.ReturnSuccess(c)
	}
}

// Kubernetes plugin interface
type k8sClusterResultObj struct {
	ResultCode    string                 `json:"resultCode"`
	ResultMessage string                 `json:"resultMessage"`
	Results       k8sClusterResultOutput `json:"results"`
}

type k8sClusterResultOutput struct {
	Outputs []k8sClusterResultOutputObj `json:"outputs"`
}

type k8sClusterResultOutputObj struct {
	CallbackParameter string `json:"callbackParameter"`
	Guid              string `json:"guid"`
	MonitorKey        string `json:"monitorKey"`
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
	ErrorDetail       string `json:"errorDetail,omitempty"`
}

type k8sClusterRequestObj struct {
	RequestId string                      `json:"requestId"`
	Inputs    []k8sClusterRequestInputObj `json:"inputs"`
}

type k8sClusterRequestInputObj struct {
	Guid              string `json:"guid"`
	CallbackParameter string `json:"callbackParameter"`
	ClusterName       string `json:"clusterName"`
	Namespace         string `json:"namespace"`
	Ip                string `json:"ip"`
	Port              string `json:"port"`
	Token             string `json:"token"`
	PodName           string `json:"podName"`
	PodGroup          string `json:"podGroup"`
	PodMonitorKey     string `json:"podMonitorKey"`
}

func PluginKubernetesCluster(c *gin.Context) {
	// Normal handle func for plugin API
	logFuncMessage := "Plugin k8s cluster interface request"
	// Action -> add | delete
	action := c.Param("action")
	var resultCode, resultMessage string
	resultCode = "0"
	resultData := k8sClusterResultOutput{}
	defer func() {
		log.Info(nil, log.LOGGER_APP, logFuncMessage, log.JsonObj("result", resultData))
		c.JSON(http.StatusOK, k8sClusterResultObj{ResultCode: resultCode, ResultMessage: resultMessage, Results: resultData})
	}()
	data, _ := ioutil.ReadAll(c.Request.Body)
	log.Debug(nil, log.LOGGER_APP, logFuncMessage, zap.String("action", action), zap.String("param", string(data)))
	var param k8sClusterRequestObj
	err := json.Unmarshal(data, &param)
	if err != nil {
		resultCode = "1"
		resultMessage = m.GetMessageMap(c).RequestJsonUnmarshalError.Error()
		return
	}
	if len(param.Inputs) == 0 {
		resultCode = "0"
		resultMessage = fmt.Sprintf(m.GetMessageMap(c).ParamEmptyError.Error(), "inputs")
		return
	}
	for _, input := range param.Inputs {
		var tmpErr error
		if action == "add" {
			tmpErr = handleAddKubernetesCluster(input)
		} else {
			tmpErr = handleDeleteKubernetesCluster(input)
		}
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, logFuncMessage, zap.String("guid", input.Guid), zap.Error(tmpErr))
			resultMessage = tmpErr.Error()
			resultData.Outputs = append(resultData.Outputs, k8sClusterResultOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpErr.Error(), Guid: input.Guid})
		} else {
			resultData.Outputs = append(resultData.Outputs, k8sClusterResultOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", Guid: input.Guid})
		}
	}
}

func handleAddKubernetesCluster(input k8sClusterRequestInputObj) error {
	var err error
	// Validate input param
	if mid.IsIllegalIp(input.Ip) {
		err = fmt.Errorf("Param ip is illegal ")
		return err
	}
	portInt, _ := strconv.Atoi(input.Port)
	if portInt <= 0 {
		err = fmt.Errorf("Param port is illegal ")
		return err
	}
	input.ClusterName = strings.TrimSpace(input.ClusterName)
	if !mid.IsIllegalNormalInput(input.ClusterName) {
		err = fmt.Errorf("Param clusterName is illegal ")
		return err
	}
	currentData, _ := db.ListKubernetesCluster(input.ClusterName)
	if len(currentData) > 0 {
		if currentData[0].ClusterName == input.ClusterName && currentData[0].Token == input.Token && fmt.Sprintf("%s:%s", input.Ip, input.Port) == currentData[0].ApiServer {
			log.Warn(nil, log.LOGGER_APP, "Plugin k8s cluster add break with same data ")
			return nil
		}
		err = db.UpdateKubernetesCluster(m.KubernetesClusterParam{Id: currentData[0].Id, ClusterName: input.ClusterName, Ip: input.Ip, Port: input.Port, Token: input.Token})
	} else {
		err = db.AddKubernetesCluster(m.KubernetesClusterParam{ClusterName: input.ClusterName, Ip: input.Ip, Port: input.Port, Token: input.Token})
	}
	db.SyncPodToEndpoint()
	return err
}

func handleDeleteKubernetesCluster(input k8sClusterRequestInputObj) error {
	input.ClusterName = strings.TrimSpace(input.ClusterName)
	if !mid.IsIllegalNormalInput(input.ClusterName) {
		return fmt.Errorf("Param clusterName is illegal ")
	}
	return db.DeleteKubernetesCluster(0, input.ClusterName)
}

func PluginKubernetesPod(c *gin.Context) {
	// Normal handle func for plugin API
	logFuncMessage := "Plugin k8s pod interface request"
	// Action -> add | delete
	action := c.Param("action")
	var resultCode, resultMessage string
	resultCode = "0"
	resultData := k8sClusterResultOutput{}
	defer func() {
		log.Info(nil, log.LOGGER_APP, logFuncMessage, log.JsonObj("result", resultData))
		c.JSON(http.StatusOK, k8sClusterResultObj{ResultCode: resultCode, ResultMessage: resultMessage, Results: resultData})
	}()
	data, _ := ioutil.ReadAll(c.Request.Body)
	log.Debug(nil, log.LOGGER_APP, logFuncMessage, zap.String("action", action), zap.String("param", string(data)))
	var param k8sClusterRequestObj
	err := json.Unmarshal(data, &param)
	if err != nil {
		resultCode = "1"
		resultMessage = m.GetMessageMap(c).RequestJsonUnmarshalError.Error()
		return
	}
	if len(param.Inputs) == 0 {
		resultCode = "0"
		resultMessage = fmt.Sprintf(m.GetMessageMap(c).ParamEmptyError.Error(), "inputs")
		return
	}
	for _, input := range param.Inputs {
		var tmpErr error
		tmpMonitorGuidKey := ""
		if action == "add" {
			tmpErr, tmpMonitorGuidKey = handleAddKubernetesPod(input)
		} else {
			tmpErr = handleDeleteKubernetesPod(input)
		}
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, logFuncMessage, zap.String("guid", input.Guid), zap.Error(tmpErr))
			resultMessage = tmpErr.Error()
			resultData.Outputs = append(resultData.Outputs, k8sClusterResultOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpErr.Error(), Guid: input.Guid})
		} else {
			resultData.Outputs = append(resultData.Outputs, k8sClusterResultOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", Guid: input.Guid, MonitorKey: tmpMonitorGuidKey})
		}
	}
}

func handleAddKubernetesPod(input k8sClusterRequestInputObj) (err error, endpointGuid string) {
	input.Guid = strings.TrimSpace(input.Guid)
	if input.Guid == "" {
		err = fmt.Errorf("Pod guid can not empty ")
		return err, endpointGuid
	}
	clusterList, err := db.ListKubernetesCluster(input.ClusterName)
	if err != nil {
		return err, endpointGuid
	}
	if len(clusterList) == 0 {
		err = fmt.Errorf("Cluster_name: %s can not find ", input.ClusterName)
		return err, endpointGuid
	}
	input.Namespace = strings.TrimSpace(input.Namespace)
	if input.Namespace == "" {
		err = fmt.Errorf("Namespace can not empty ")
		return err, endpointGuid
	}
	input.PodName = strings.TrimSpace(input.PodName)
	if input.PodName == "" {
		err = fmt.Errorf("Pod name can not empty ")
		return err, endpointGuid
	}
	var insertId int64
	err, insertId, endpointGuid = db.AddKubernetesPod(clusterList[0], input.Guid, input.PodName, input.Namespace)
	if err != nil {
		return err, endpointGuid
	}
	if input.PodGroup != "" {
		err, tplId := db.UpdateKubernetesPodGroup(insertId, input.PodGroup, "add")
		if err != nil {
			return err, endpointGuid
		}
		if tplId > 0 {
			//err = alarm.SaveConfigFile(tplId, false)
			err = db.SyncRuleConfigFile(tplId, []string{}, false)
		}
	}
	return err, endpointGuid
}

func handleDeleteKubernetesPod(input k8sClusterRequestInputObj) error {
	var err error
	if input.PodMonitorKey == "" {
		input.Guid = strings.TrimSpace(input.Guid)
		if input.Guid == "" {
			return fmt.Errorf("Pod guid can not empty ")
		}
	}
	err, endpointId := db.DeleteKubernetesPod(input.Guid, input.PodMonitorKey)
	if err != nil {
		return err
	}
	if input.PodGroup != "" {
		err, tplId := db.UpdateKubernetesPodGroup(endpointId, input.PodGroup, "delete")
		if err != nil {
			return err
		}
		if tplId > 0 {
			//err = alarm.SaveConfigFile(tplId, false)
			err = db.SyncRuleConfigFile(tplId, []string{}, false)
		}
	}
	return err
}
