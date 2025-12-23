package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/WeBankPartners/go-common-lib/cipher"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListKubernetesCluster 查询所有 Kubernetes 集群信息
func ListKubernetesCluster(c *gin.Context) {
	clusterName := strings.TrimSpace(c.Query("clusterName"))
	result, err := db.ListKubernetesCluster(clusterName)
	if err != nil {
		mid.ReturnServerHandleError(c, err)
		return
	}
	mid.ReturnSuccessData(c, result)
}

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
				// 新增时检查集群名称是否已存在
				existingClusters, checkErr := db.ListKubernetesCluster(param.ClusterName)
				if checkErr != nil {
					mid.ReturnHandleError(c, checkErr.Error(), checkErr)
					return
				}
				if len(existingClusters) > 0 {
					mid.ReturnValidateError(c, fmt.Sprintf("kubernetes cluster name '%s' already exists", param.ClusterName))
					return
				}
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
	Guid              string `json:"guid"` // 实例名
	CallbackParameter string `json:"callbackParameter"`
	ClusterName       string `json:"clusterName"`
	Namespace         string `json:"namespace"`
	ApiServer         string `json:"apiServer"`
	Token             string `json:"token"`
	PodName           string `json:"podName"`
	PodGroup          string `json:"podGroup"`
	PodMonitorKey     string `json:"podMonitorKey"` // endpointGuid,删除时候和更新时候用
	Ip                string `json:"ip"`            // serviceIP
	NodeIp            string `json:"nodeIp"`        // 真实ip,更新时候需要传递过来
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
	log.Info(nil, log.LOGGER_APP, logFuncMessage, zap.String("action", action), zap.String("param", string(data)))
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
			resultCode = "1"
			resultData.Outputs = append(resultData.Outputs, k8sClusterResultOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "1", ErrorMessage: tmpErr.Error(), Guid: input.Guid})
		} else {
			resultData.Outputs = append(resultData.Outputs, k8sClusterResultOutputObj{CallbackParameter: input.CallbackParameter, ErrorCode: "0", ErrorMessage: "", Guid: input.Guid})
		}
	}
}

func handleAddKubernetesCluster(input k8sClusterRequestInputObj) error {
	var err error
	if strings.TrimSpace(input.ApiServer) == "" {
		err = fmt.Errorf("Param ApiServer is empty ")
		return err
	}
	strArr := strings.Split(input.ApiServer, ":")
	if len(strArr) != 2 {
		err = fmt.Errorf("Param ApiServer is invalid")
		return err
	}
	ip := strArr[0]
	port := strArr[1]
	// Validate input param
	if mid.IsIllegalIp(ip) {
		err = fmt.Errorf("Param ip is illegal ")
		return err
	}
	portInt, _ := strconv.Atoi(port)
	if portInt <= 0 {
		err = fmt.Errorf("Param port is illegal ")
		return err
	}
	if strings.TrimSpace(input.Token) == "" {
		err = fmt.Errorf("token is empty")
		return err
	}
	input.ClusterName = strings.TrimSpace(input.ClusterName)
	if !mid.IsIllegalNormalInput(input.ClusterName) {
		err = fmt.Errorf("Param clusterName is illegal ")
		return err
	}
	// 新增时检查集群名称是否已存在
	currentData, checkErr := db.ListKubernetesCluster(input.ClusterName)
	if checkErr != nil {
		err = fmt.Errorf("Check kubernetes cluster name fail: %s", checkErr.Error())
		return err
	}
	if len(currentData) > 0 {
		// 如果集群名称已存在，返回错误
		err = fmt.Errorf("kubernetes cluster name '%s' already exists", input.ClusterName)
		return err
	}
	// 名称不存在，执行新增操作
	err = db.AddKubernetesCluster(m.KubernetesClusterParam{ClusterName: input.ClusterName, Ip: ip, Port: port, Token: input.Token, Guid: input.Guid})
	return err
}

func isSameKubernetesToken(cluster *m.KubernetesClusterTable, targetToken, targetGuid string) bool {
	if cluster == nil {
		return false
	}
	normalizeToken := strings.TrimSpace(targetToken)
	if normalizeToken == "" {
		return false
	}
	if cluster.Token == normalizeToken {
		return true
	}
	clusterGuid := strings.TrimSpace(cluster.Guid)
	if clusterGuid == "" {
		clusterGuid = strings.TrimSpace(targetGuid)
	}
	if clusterGuid == "" {
		return false
	}
	encToken, err := cipher.AesEnPasswordByGuid(clusterGuid, m.Config().EncryptSeed, normalizeToken, "")
	if err != nil {
		return false
	}
	return encToken == cluster.Token
}

func handleDeleteKubernetesCluster(input k8sClusterRequestInputObj) error {
	input.ClusterName = strings.TrimSpace(input.ClusterName)
	if !mid.IsIllegalNormalInput(input.ClusterName) {
		return fmt.Errorf("Param clusterName is illegal ")
	}
	clusterList, err := db.ListKubernetesCluster(input.ClusterName)
	if err != nil {
		return err
	}
	if len(clusterList) == 0 {
		return fmt.Errorf("kubernetes cluster %s not found", input.ClusterName)
	}
	hasPods, err := db.HasKubernetesClusterPods(clusterList[0].Id)
	if err != nil {
		return err
	}
	if hasPods {
		return fmt.Errorf("kubernetes cluster %s still has pod endpoints, please remove pod objects first", input.ClusterName)
	}
	return db.DeleteKubernetesCluster(clusterList[0].Id, input.ClusterName)
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
			// 处理 pod新赠&更新,可能会出现 pod的ip漂移。业务配置和关键字映射变更 直接修改映射host就好了
			tmpErr, tmpMonitorGuidKey = handleAddKubernetesPod(input)
		} else if action == "delete" {
			tmpErr = handleDeleteKubernetesPod(input)
		}
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, logFuncMessage, zap.String("guid", input.Guid), zap.Error(tmpErr))
			resultMessage = tmpErr.Error()
			resultCode = "1"
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
		input.Namespace = "default"
	}
	input.PodName = strings.TrimSpace(input.PodName)
	if input.PodName == "" {
		err = fmt.Errorf("Pod name can not empty ")
		return err, endpointGuid
	}
	// 根据 PodName 和集群 ID 查询是否存在
	k8sEndpointRel, err := db.GetKubernetesEndpointRelByPodName(input.PodName, clusterList[0].Id)
	if err != nil {
		return err, endpointGuid
	}
	// 如果存在，则更新pod
	if k8sEndpointRel != nil && k8sEndpointRel.EndpointGuid != "" {
		var result m.EndpointNewTable
		result, err = db.GetEndpointNew(&m.EndpointNewTable{Guid: k8sEndpointRel.EndpointGuid})
		if err != nil {
			return err, endpointGuid
		}
		if result.Guid != "" {
			endpointGuid = result.Guid
			var oldExtendObj m.EndpointExtendParamObj
			if result.ExtendParam != "" {
				err = json.Unmarshal([]byte(result.ExtendParam), &oldExtendObj)
				if err != nil {
					return err, endpointGuid
				}
			}
			// 构建新的 extend_param
			newExtendObj := oldExtendObj
			if input.NodeIp != "" {
				newExtendObj.NodeIp = input.NodeIp
			}
			// 检查 node_ip 是否有变化
			nodeIpChanged := oldExtendObj.NodeIp != newExtendObj.NodeIp
			// 如果 node_ip 有变化，才执行复杂的更新操作
			if nodeIpChanged && oldExtendObj.NodeIp != "" && input.NodeIp != "" {
				err = handleUpdateKubernetesPod(endpointGuid, input.NodeIp, oldExtendObj.NodeIp)
				if err != nil {
					return err, endpointGuid
				}
				// 更新 endpoint_new 表（包括 extend_param）
				newExtendParamBytes, _ := json.Marshal(newExtendObj)
				newExtendParamString := string(newExtendParamBytes)
				err = db.UpdateKubernetesPodEndpointNew(endpointGuid, input.Ip, newExtendParamString)
				if err != nil {
					return err, endpointGuid
				}
			} else {
				// 更新 endpoint_new 表（包括 extend_param），即使 node_ip 没有变化
				newExtendParamBytes, _ := json.Marshal(newExtendObj)
				newExtendParamString := string(newExtendParamBytes)
				err = db.UpdateKubernetesPodEndpointNew(endpointGuid, input.Ip, newExtendParamString)
				if err != nil {
					return err, endpointGuid
				}
			}
			log.Info(nil, log.LOGGER_APP, "Update kubernetes pod success", zap.String("podName", input.PodName), zap.String("endpointGuid", endpointGuid), zap.String("namespace", input.Namespace), zap.String("clusterName", input.ClusterName))
			return err, endpointGuid
		}
	}
	// 新增pod
	var insertId int64
	err, insertId, endpointGuid = db.AddKubernetesPod(clusterList[0], input.Guid, input.PodName, input.Namespace, input.Ip, input.NodeIp)
	if err != nil {
		return err, endpointGuid
	}
	if input.PodGroup != "" {
		err, tplId := db.UpdateKubernetesPodGroup(insertId, input.PodGroup, "add")
		if err != nil {
			return err, endpointGuid
		}
		if tplId > 0 {
			if err = db.SyncRuleConfigFile(tplId, []string{}, false); err != nil {
				return err, endpointGuid
			}
		}
		// 设置对象和对象组关系
		if err = db.AddGroupEndpointRel(endpointGuid, input.PodGroup); err != nil {
			return err, endpointGuid
		}
	}
	log.Info(nil, log.LOGGER_APP, "Add kubernetes pod success", zap.String("podName", input.PodName), zap.String("endpointGuid", endpointGuid), zap.String("namespace", input.Namespace), zap.String("clusterName", input.ClusterName), zap.String("podGuid", input.Guid))
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
			err = db.SyncRuleConfigFile(tplId, []string{}, false)
		}
	}
	return err
}

// handleUpdateKubernetesPod  目前更新业务配置、指标阈值映射 ip映射就好了
func handleUpdateKubernetesPod(sourceEndpointGuid, targetNodeIp, sourceRealIp string) (err error) {
	var endpointList []string
	if sourceEndpointGuid == "" {
		return fmt.Errorf("sourceEndpointGuid can not empty ")
	}
	if targetNodeIp == "" {
		return fmt.Errorf("targetNodeIp can not empty ")
	}
	if sourceRealIp == "" {
		return fmt.Errorf("sourceRealIp can not empty ")
	}
	var targetEndpoint *m.EndpointNewTable
	if targetEndpoint, err = db.GetEndpointByIpAndType(targetNodeIp, "host"); err != nil {
		log.Warn(nil, log.LOGGER_APP, "GetEndpointByIpAndType for targetNodeIp failed", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.Error(err))
		return
	}
	if targetEndpoint == nil {
		log.Warn(nil, log.LOGGER_APP, "targetNodeIp mapping pod host endpoint not found, skip update operation", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("monitorType", "host"))
		return nil
	}
	var sourceHostEndpoint *m.EndpointNewTable
	if sourceHostEndpoint, err = db.GetEndpointByIpAndType(sourceRealIp, "host"); err != nil {
		log.Warn(nil, log.LOGGER_APP, "GetEndpointByIpAndType for sourceRealIp failed", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.Error(err))
		return
	}
	if sourceHostEndpoint == nil {
		log.Warn(nil, log.LOGGER_APP, "sourceRealIp mapping pod host endpoint not found, skip update operation", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("monitorType", "host"))
		return nil
	}
	endpointList = append(endpointList, targetEndpoint.Guid, sourceHostEndpoint.Guid)
	// 更新 日志文件业务配置
	if err = db.UpdateLogMetricSourceEndpoint(targetEndpoint.Guid, sourceEndpointGuid); err != nil {
		log.Error(nil, log.LOGGER_APP, "UpdateLogMetricSourceEndpoint fail", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip), zap.Error(err))
		return
	}
	log.Info(nil, log.LOGGER_APP, "UpdateLogMetricSourceEndpoint success", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip))
	// 更新 数据库业务配置
	if err = db.UpdateDbMetricSourceEndpoint(targetEndpoint.Guid, sourceEndpointGuid); err != nil {
		log.Error(nil, log.LOGGER_APP, "UpdateDbMetricSourceEndpoint fail", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip), zap.Error(err))
		return
	}
	if err = db.SyncLogMetricExporterConfig(endpointList); err != nil {
		log.Error(nil, log.LOGGER_APP, "SyncLogMetricExporterConfig fail", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.Strings("endpointList", endpointList), zap.Error(err))
		return
	}
	log.Info(nil, log.LOGGER_APP, "UpdateDbMetricSourceEndpoint success", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip))

	// 更新 日志文件关键字配置
	if err = db.UpdateLogKeywordSourceEndpoint(targetEndpoint.Guid, sourceEndpointGuid); err != nil {
		log.Error(nil, log.LOGGER_APP, "UpdateLogKeywordSourceEndpoint fail", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip), zap.Error(err))
		return
	}
	log.Info(nil, log.LOGGER_APP, "UpdateLogKeywordSourceEndpoint success", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip))

	// 更新 数据库关键字配置
	if err = db.UpdateDbKeywordSourceEndpoint(targetEndpoint.Guid, sourceEndpointGuid); err != nil {
		log.Error(nil, log.LOGGER_APP, "UpdateDbKeywordSourceEndpoint fail", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip), zap.Error(err))
	}
	if err = db.SyncLogKeywordExporterConfig(endpointList); err != nil {
		log.Error(nil, log.LOGGER_APP, "SyncLogKeywordExporterConfig fail", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.Strings("endpointList", endpointList), zap.Error(err))
		return
	}
	log.Info(nil, log.LOGGER_APP, "UpdateDbKeywordSourceEndpoint success", zap.String("sourceEndpointGuid", sourceEndpointGuid), zap.String("targetNodeIp", targetNodeIp), zap.String("sourceRealIp", sourceRealIp), zap.String("targetEndpointGuid", targetEndpoint.Guid), zap.String("targetEndpointIp", targetEndpoint.Ip), zap.String("sourceHostEndpointGuid", sourceHostEndpoint.Guid), zap.String("sourceHostEndpointIp", sourceHostEndpoint.Ip))

	return
}
