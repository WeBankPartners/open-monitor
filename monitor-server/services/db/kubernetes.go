package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
)

func ListKubernetesCluster(clusterName string) (result []*m.KubernetesClusterTable, err error) {
	if clusterName != "" {
		err = x.SQL("select * from kubernetes_cluster where cluster_name=?", clusterName).Find(&result)
	} else {
		err = x.SQL("select * from kubernetes_cluster").Find(&result)
	}
	return result, err
}

func AddKubernetesCluster(param m.KubernetesClusterParam) error {
	_, err := x.Exec("insert into kubernetes_cluster(cluster_name,api_server,token,create_at) value (?,?,?,'"+time.Now().Format(m.DatetimeFormat)+"')", param.ClusterName, fmt.Sprintf("%s:%s", param.Ip, param.Port), param.Token)
	if err != nil {
		err = fmt.Errorf("Insert into db fail,%s ", err.Error())
		return err
	}
	err = SyncKubernetesConfig()
	if err != nil {
		err = fmt.Errorf("Update prometheus config fail,%s ", err.Error())
	}
	return err
}

func UpdateKubernetesCluster(param m.KubernetesClusterParam) error {
	_, err := x.Exec("update kubernetes_cluster set cluster_name=?,api_server=?,token=? where id=?", param.ClusterName, fmt.Sprintf("%s:%s", param.Ip, param.Port), param.Token, param.Id)
	if err != nil {
		err = fmt.Errorf("Update db table fail,%s ", err.Error())
		return err
	}
	err = SyncKubernetesConfig()
	if err != nil {
		err = fmt.Errorf("Update prometheus config fail,%s ", err.Error())
	}
	return err
}

func DeleteKubernetesCluster(id int, clusterName string) error {
	if id <= 0 {
		kubernetesTables, _ := ListKubernetesCluster(clusterName)
		if len(kubernetesTables) == 0 {
			log.Warn(nil, log.LOGGER_APP, "Delete kubernetes cluster break,can not find fetch data", zap.String("cluster_name", clusterName))
			return nil
		}
		id = kubernetesTables[0].Id
	}
	_, err := x.Exec("delete from kubernetes_cluster where id=?", id)
	if err != nil {
		err = fmt.Errorf("Delete db data fail,%s ", err.Error())
		return err
	}
	x.Exec("delete from kubernetes_endpoint_rel where kubernete_id=?", id)
	SyncKubernetesConfig()
	return err
}

func DeleteKubernetesEndpointRelByEndpointId(endpointGuid string) error {
	_, err := x.Exec("delete from kubernetes_endpoint_rel where endpoint_guid=?", endpointGuid)
	return err
}

func InitPrometheusConfigFile() {
	err := SyncKubernetesConfig()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Init kubernetes config fail", zap.Error(err))
	}
	err = SyncSnmpPrometheusConfig()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Init Snmp config fail", zap.Error(err))
	}
}

func SyncKubernetesConfig() error {
	var kubernetesTables []*m.KubernetesClusterTable
	err := x.SQL("select * from kubernetes_cluster").Find(&kubernetesTables)
	if len(kubernetesTables) == 0 {
		if err == nil {
			x.Exec("delete from kubernetes_endpoint_rel")
		}
		return fmt.Errorf("kubernetes config empty")
	}
	promBytes, err := ioutil.ReadFile("/app/monitor/prometheus/prometheus.yml")
	if err != nil {
		err = fmt.Errorf("Read prometheus tpl file fail,%s ", err.Error())
		return err
	}
	backupConfigName, backupErr := backupPrometheusConfig()
	if backupErr != nil {
		return backupErr
	}
	cleanTokenOutput, err := exec.Command("/bin/sh", "-c", "rm -f /app/monitor/prometheus/token/*").Output()
	if err != nil {
		err = fmt.Errorf("Clean token fail,output:%s,err:%s ", string(cleanTokenOutput), err.Error())
		return err
	}
	for _, v := range kubernetesTables {
		err = ioutil.WriteFile(fmt.Sprintf("/app/monitor/prometheus/token/%s", v.ClusterName), []byte(v.Token), 0644)
		if err != nil {
			err = fmt.Errorf("Write cluster %s token file fail,%s ", v.ClusterName, err.Error())
			break
		}
	}
	if err != nil {
		return err
	}
	promString := string(promBytes)
	startIndex := strings.Index(promString, "#Kubernetes_start")
	endIndex := strings.Index(promString, "#Kubernetes_end")
	tplBytes, err := ioutil.ReadFile("/app/monitor/prometheus/kubernetes_prometheus.tpl")
	if err != nil {
		err = fmt.Errorf("Read kubernetes prometheus config template file fail,%s ", err.Error())
		return err
	}
	kubernetesConfigString := ""
	tplString := string(tplBytes)
	for _, kube := range kubernetesTables {
		tmpKPConfig := tplString + "\n"
		tmpIpSplit := strings.Split(kube.ApiServer, ":")
		tmpKPConfig = strings.ReplaceAll(tmpKPConfig, "{{cluster_name}}", kube.ClusterName)
		tmpKPConfig = strings.ReplaceAll(tmpKPConfig, "{{api_server_ip}}", tmpIpSplit[0])
		tmpKPConfig = strings.ReplaceAll(tmpKPConfig, "{{api_server_port}}", tmpIpSplit[1])
		kubernetesConfigString += tmpKPConfig + "\n"
	}
	promString = promString[:startIndex+17] + "\n" + kubernetesConfigString + promString[endIndex:]
	err = ioutil.WriteFile("/app/monitor/prometheus/prometheus.yml", []byte(promString), 0644)
	if err != nil {
		err = fmt.Errorf("Write kubenetes config to prometheus fail,%s ", err.Error())
		recoverPrometheusConfig(backupConfigName)
		return err
	}
	err = prom.ReloadConfig()
	if err != nil {
		err = fmt.Errorf("Reload prometheus config fail,%s ", err.Error())
		recoverPrometheusConfig(backupConfigName)
		prom.ReloadConfig()
		return err
	}
	return nil
}

func backupPrometheusConfig() (name string, err error) {
	name = fmt.Sprintf("prometheus_%d.yml", time.Now().Unix())
	backupOutput, err := exec.Command("/bin/sh", "-c", "cp /app/monitor/prometheus/prometheus.yml /tmp/"+name).Output()
	if err != nil {
		err = fmt.Errorf("Backup prometheus config file fail,output:%s,err:%s ", string(backupOutput), err.Error())
		log.Error(nil, log.LOGGER_APP, "Backup prometheus config fail", zap.Error(err))
	}
	return
}

func recoverPrometheusConfig(name string) {
	recoverOutput, recoverError := exec.Command("/bin/sh", "-c", "rm -f /app/monitor/prometheus/prometheus.yml && cp /tmp/"+name+" /app/monitor/prometheus/prometheus.yml").Output()
	if recoverError != nil {
		log.Error(nil, log.LOGGER_APP, "Recover prometheus config fail", zap.String("output", string(recoverOutput)), zap.Error(recoverError))
	}
}

func StartCronSyncKubernetesPod(interval int) {
	t := time.NewTicker(time.Duration(120) * time.Second).C
	for {
		<-t
		go SyncPodToEndpoint()
	}
}

func SyncPodToEndpoint() bool {
	log.Info(nil, log.LOGGER_APP, "start to sync kubernetes pod")
	var kubernetesTables []*m.KubernetesClusterTable
	result := false
	x.SQL("select * from kubernetes_cluster").Find(&kubernetesTables)
	if len(kubernetesTables) == 0 {
		return result
	}
	var endpointGroup []*m.GrpTable
	x.SQL("select * from grp where name='default_pod_group'").Find(&endpointGroup)
	var endpointTables []*m.EndpointTable
	var kubernetesEndpointTables []*m.KubernetesEndpointRelTable
	for _, v := range kubernetesTables {
		queryParam := m.QueryMonitorData{Legend: "$pod", SameEndpoint: true, ChartType: "line", PromQ: fmt.Sprintf("container_processes{pod=~\".*-.*\",job=\"k8s-cadvisor-%s\"}", v.ClusterName), Start: time.Now().Unix() - 600, End: time.Now().Unix()}
		series := datasource.PrometheusData(&queryParam)
		tmpGuidMap := make(map[string]int)
		var tmpKubernetesEndpointTables []*m.KubernetesEndpointRelTable
		x.SQL("select * from kubernetes_endpoint_rel where kubernete_id=?", v.Id).Find(&tmpKubernetesEndpointTables)
		tmpApiServerIp := v.ApiServer[:strings.Index(v.ApiServer, ":")]
		log.Debug(nil, log.LOGGER_APP, "kubernetes series", log.JsonObj("series", series))
		for _, vv := range series {
			tmpPodName := vv.Name
			if strings.HasPrefix(tmpPodName, "pod=") {
				tmpPodName = tmpPodName[4:]
			}
			tmpEndpointGuid := fmt.Sprintf("%s_%s_pod", tmpPodName, tmpApiServerIp)
			existsFlag := false
			for _, ke := range tmpKubernetesEndpointTables {
				if ke.EndpointGuid == tmpEndpointGuid {
					existsFlag = true
					break
				}
			}
			if !existsFlag {
				if _, b := tmpGuidMap[tmpEndpointGuid]; b {
					continue
				} else {
					tmpGuidMap[tmpEndpointGuid] = 1
				}
				endpointTables = append(endpointTables, &m.EndpointTable{Guid: tmpEndpointGuid, Name: tmpPodName, Ip: tmpApiServerIp, ExportType: "pod", Step: 10, OsType: v.ClusterName})
				kubernetesEndpointTables = append(kubernetesEndpointTables, &m.KubernetesEndpointRelTable{KuberneteId: v.Id, EndpointGuid: tmpEndpointGuid, PodGuid: tmpPodName})
			}
		}
	}
	log.Debug(nil, log.LOGGER_APP, "kubernetes endpointTables", log.JsonObj("endpointTables", endpointTables))
	if len(endpointTables) > 0 {
		result = true
		var tmpGuidList []string
		endpointSql := "insert into endpoint(guid,name,ip,export_type,step,os_type) values "
		newEndpointSql := "insert into endpoint_new(guid,name,ip,monitor_type,step,tags,create_user,update_user,cluster) values "
		for i, v := range endpointTables {
			tmpGuidList = append(tmpGuidList, v.Guid)
			log.Info(nil, log.LOGGER_APP, "add kubernetes pod endpoint", zap.String("guid", v.Guid))
			endpointSql += fmt.Sprintf("('%s','%s','%s','%s',%d,'%s')", v.Guid, v.Name, v.Ip, v.ExportType, v.Step, v.OsType)
			newEndpointSql += fmt.Sprintf("('%s','%s','%s','%s',%d,'%s','system','system','default')", v.Guid, v.Name, v.Ip, v.ExportType, v.Step, v.OsType)
			if i < len(endpointTables)-1 {
				endpointSql += ","
				newEndpointSql += ","
			}
		}
		_, err := x.Exec(endpointSql)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update kubernetes pod to endpoint table fail", zap.String("sql", endpointSql), zap.Error(err))
		}
		x.Exec(newEndpointSql)
		if len(endpointGroup) > 0 {
			var tmpEndpointTables []*m.EndpointTable
			x.SQL("select id from endpoint where guid in ('" + strings.Join(tmpGuidList, "','") + "')").Find(&tmpEndpointTables)
			if len(tmpEndpointTables) > 0 {
				insertEndpointGrpSql := "insert into grp_endpoint(grp_id,endpoint_id) values "
				for _, v := range tmpEndpointTables {
					insertEndpointGrpSql += fmt.Sprintf("(%d,%d),", endpointGroup[0].Id, v.Id)
				}
				_, err = x.Exec(insertEndpointGrpSql[:len(insertEndpointGrpSql)-1])
				if err != nil {
					log.Error(nil, log.LOGGER_APP, "Try to update endpoint group fail", zap.String("sql", insertEndpointGrpSql), zap.Error(err))
				}
			}
			if len(tmpGuidList) > 0 {
				for _, tmpGuid := range tmpGuidList {
					_, err = x.Exec("INSERT INTO endpoint_group_rel (guid,endpoint,endpoint_group) values (?,?,'default_pod_group')", guid.CreateGuid(), tmpGuid)
					if err != nil {
						log.Error(nil, log.LOGGER_APP, "Try to insert endpoint group fail", zap.String("endpointGuid", tmpGuid), zap.Error(err))
					}
				}
			}
		}
	}
	log.Debug(nil, log.LOGGER_APP, "kubernetes series", log.JsonObj("kubernetesEndpointTables", kubernetesEndpointTables))
	if len(kubernetesEndpointTables) > 0 {
		result = true
		keRelSql := "insert into kubernetes_endpoint_rel(kubernete_id,endpoint_guid,pod_guid,namespace) values "
		for i, v := range kubernetesEndpointTables {
			keRelSql += fmt.Sprintf("(%d,'%s','%s','default')", v.KuberneteId, v.EndpointGuid, v.PodGuid)
			if i < len(kubernetesEndpointTables)-1 {
				keRelSql += ","
			}
		}
		_, err := x.Exec(keRelSql)
		if err != nil {
			log.Error(nil, log.LOGGER_APP, "Update kubernetes endpoint rel table fail", zap.String("sql", keRelSql), zap.Error(err))
		}
	}
	return result
}

func AddKubernetesPod(cluster *m.KubernetesClusterTable, podGuid, podName, namespace, serviceIp, nodeIp string) (err error, id int64, endpointGuid string) {
	apiServerIp := cluster.ApiServer[:strings.Index(cluster.ApiServer, ":")]
	endpointGuid = fmt.Sprintf("%s_%s_pod", podName, serviceIp)
	endpointObj := m.EndpointTable{Guid: endpointGuid}
	endpointObjNew := m.EndpointNewTable{Guid: endpointGuid}
	GetEndpoint(&endpointObj)
	result, _ := GetEndpointNew(&endpointObjNew)
	if endpointObj.Id <= 0 {
		execResult, err := x.Exec("insert into endpoint(guid,name,ip,export_type,step,export_version,os_type) value (?,?,?,'pod',10,?,?)", endpointGuid, podName, apiServerIp, namespace, cluster.ClusterName)
		if err != nil {
			return err, id, endpointGuid
		}
		lastId, _ := execResult.LastInsertId()
		if lastId <= 0 {
			err = fmt.Errorf("Insert endpoint table fail,last insert id illegal,please check server log ")
			return err, id, endpointGuid
		}
		id = lastId
	}
	// 插入endpoint_new表
	if result.Guid == "" {
		result.Guid = endpointGuid
		result.Name = podName
		result.Ip = serviceIp
		result.MonitorType = "pod"
		result.Step = 10
		result.Cluster = "default"
		nowTime := time.Now().Format(m.DatetimeFormat)
		extendString := ""
		if nodeIp != "" {
			extendParam := m.EndpointExtendParamObj{}
			extendParam.NodeIp = nodeIp
			tmpExtendBytes, _ := json.Marshal(extendParam)
			extendString = string(tmpExtendBytes)
		}
		_, err := x.Exec("insert into endpoint_new(guid,name,ip,monitor_type,agent_version,agent_address,step,endpoint_version,endpoint_address,cluster,extend_param,update_time,create_user,update_user) "+
			"value (?,?,?,?,?,?,?,?,?,?,?,?,?,?)", result.Guid, result.Name, result.Ip, result.MonitorType, "", "", result.Step, "", "", result.Cluster, extendString, nowTime, "system", "system")
		if err != nil {
			return err, id, endpointGuid
		}
	}
	var kubernetesEndpointTables []*m.KubernetesEndpointRelTable
	x.SQL("select * from kubernetes_endpoint_rel where kubernete_id=? and endpoint_guid=?", cluster.Id, endpointGuid).Find(&kubernetesEndpointTables)
	if len(kubernetesEndpointTables) <= 0 {
		_, err = x.Exec("insert into kubernetes_endpoint_rel(kubernete_id,endpoint_guid,pod_guid,namespace) value (?,?,?,?)", cluster.Id, endpointGuid, podGuid, namespace)
	}
	return err, id, endpointGuid
}

func GetKubernetesEndpointRelByPodGuid(podGuid string) (*m.KubernetesEndpointRelTable, error) {
	var kubernetesEndpointTables []*m.KubernetesEndpointRelTable
	x.SQL("select * from kubernetes_endpoint_rel where pod_guid=?", podGuid).Find(&kubernetesEndpointTables)
	if len(kubernetesEndpointTables) <= 0 {
		return nil, nil
	}
	return kubernetesEndpointTables[0], nil
}

func AddKubernetesEndpointRel(kubernetesId int, endpointGuid, podGuid string) (err error) {
	_, err = x.Exec("insert into kubernetes_endpoint_rel(kubernete_id,endpoint_guid,pod_guid,namespace) value (?,?,?,?)", kubernetesId, endpointGuid, podGuid, "default")
	return
}

func DeleteKubernetesEndpointRel(endpointGuid, podGuid string) (err error) {
	_, err = x.Exec("delete from kubernetes_endpoint_rel where endpoint_guid =? and pod_guid=?", endpointGuid, podGuid)
	return
}

func GetKubernetesByName(clusterName string) (cluster *m.KubernetesClusterTable, err error) {
	cluster = &m.KubernetesClusterTable{}
	var clusterList []*m.KubernetesClusterTable
	if err = x.SQL("select * from kubernetes_cluster where cluster_name=?", clusterName).Find(&clusterList); err != nil {
		return
	}
	if len(clusterList) == 0 {
		return nil, errors.New("kubernetes cluster not exist")
	}
	cluster = clusterList[0]
	return
}

func GetKubernetesClusterByEndpointGuid(guid string) (clusterName string, err error) {
	var kubernetesEndpointRelList []*m.KubernetesEndpointRelTable
	var kubernetesId int
	if err = x.SQL("select kubernete_id from kubernetes_endpoint_rel where endpoint_guid=?", guid).Find(&kubernetesEndpointRelList); err != nil {
		return
	}
	if len(kubernetesEndpointRelList) == 0 {
		return "", errors.New("kubernetes cluster not exist")
	}
	kubernetesId = kubernetesEndpointRelList[0].KuberneteId
	x.SQL("select cluster_name from kubernetes_cluster where id=?", kubernetesId).Find(&clusterName)
	return
}

func DeleteKubernetesPod(podGuid, endpointGuid string) (err error, id int64) {
	if endpointGuid == "" {
		var kubernetesEndpointTables []*m.KubernetesEndpointRelTable
		x.SQL("select * from kubernetes_endpoint_rel where pod_guid=?", podGuid).Find(&kubernetesEndpointTables)
		if len(kubernetesEndpointTables) <= 0 {
			return err, id
		}
		_, err = x.Exec("delete from kubernetes_endpoint_rel where pod_guid=?", podGuid)
		if err != nil {
			return err, id
		}
		endpointGuid = kubernetesEndpointTables[0].EndpointGuid
	} else {
		_, err = x.Exec("delete from kubernetes_endpoint_rel where pod_guid=? and endpoint_guid=?", podGuid, endpointGuid)
		if err != nil {
			return err, id
		}
	}
	err = DeleteEndpoint(endpointGuid, "system")
	log.Info(nil, log.LOGGER_APP, "DeleteKubernetesPod success", zap.String("podGuid", podGuid), zap.String("endpointGuid", endpointGuid))
	return err, id
}

func UpdateKubernetesPodGroup(endpointId int64, group, operation string) (err error, tplId int) {
	var groupTables []*m.GrpTable
	x.SQL("select * from grp where name=?", group).Find(&groupTables)
	if len(groupTables) == 0 {
		err = fmt.Errorf("Group:%s can not find,please check group config ", group)
		return
	}
	var groupEndpointTable []*m.GrpEndpointTable
	x.SQL("select * from grp_endpoint where grp_id=? and endpoint_id=?", groupTables[0].Id, endpointId).Find(&groupEndpointTable)
	if len(groupEndpointTable) > 0 {
		if operation == "add" {
			return
		}
		_, err = x.Exec("delete from grp_endpoint where grp_id=? and endpoint_id=?", groupTables[0].Id, endpointId)
		if err != nil {
			return
		}
	} else {
		if operation == "delete" {
			return
		}
		_, err = x.Exec("insert into grp_endpoint(grp_id,endpoint_id) value (?,?)", groupTables[0].Id, endpointId)
		if err != nil {
			return
		}
	}
	_, tplObj := GetTpl(0, groupTables[0].Id, 0)
	tplId = tplObj.Id
	return
}
