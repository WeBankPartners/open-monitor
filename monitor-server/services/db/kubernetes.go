package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
	"fmt"
	"io/ioutil"
	"os/exec"
	"bytes"
	"strings"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
)

func ListKubernetesCluster(clusterName string) (result []*m.KubernetesClusterTable,err error) {
	if clusterName != "" {
		err = x.SQL("select * from kubernetes_cluster where cluster_name=?", clusterName).Find(&result)
	}else {
		err = x.SQL("select * from kubernetes_cluster").Find(&result)
	}
	return result,err
}

func AddKubernetesCluster(param m.KubernetesClusterParam) error {
	_,err := x.Exec("insert into kubernetes_cluster(cluster_name,api_server,token,create_at) value (?,?,?,'"+ time.Now().Format(m.DatetimeFormat) +"')", param.ClusterName, fmt.Sprintf("%s:%s", param.Ip, param.Port), param.Token)
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
	_,err := x.Exec("update kubernetes_cluster set cluster_name=?,api_server=?,token=? where id=?", param.ClusterName, fmt.Sprintf("%s:%s", param.Ip, param.Port), param.Token, param.Id)
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
		kubernetesTables,_ := ListKubernetesCluster(clusterName)
		if len(kubernetesTables) == 0 {
			log.Logger.Warn("Delete kubernetes cluster break,can not find fetch data", log.String("cluster_name", clusterName))
			return nil
		}
		id = kubernetesTables[0].Id
	}
	_,err := x.Exec("delete from kubernetes_cluster where id=?", id)
	if err != nil {
		err = fmt.Errorf("Delete db data fail,%s ", err.Error())
		return err
	}
	x.Exec("delete from kubernetes_endpoint_rel where kubernete_id=?", id)
	SyncKubernetesConfig()
	return err
}

func InitKubernetesConfig()  {
	err := SyncKubernetesConfig()
	if err != nil {
		log.Logger.Error("Init kubernetes config fail", log.Error(err))
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
	tplBytes,err := ioutil.ReadFile("/app/monitor/prometheus/prometheus_tpl.yml")
	if err != nil {
		err = fmt.Errorf("Read prometheus tpl file fail,%s ", err.Error())
		return err
	}
	backupConfigName := fmt.Sprintf("prometheus_%d.yml", time.Now().Unix())
	backupBytes,err := exec.Command("/bin/sh", "-c", "cp /app/monitor/prometheus/prometheus.yml /app/monitor/prometheus/"+backupConfigName).Output()
	if err != nil {
		err = fmt.Errorf("Backup prometheus config file fail,output:%s,err:%s ", string(backupBytes), err.Error())
		return err
	}
	cleanTokenOutput,err := exec.Command("/bin/sh", "-c", "rm -f /app/monitor/prometheus/token/*").Output()
	if err != nil {
		err = fmt.Errorf("Clean token fail,output:%s,err:%s ", string(cleanTokenOutput), err.Error())
		return err
	}
	for _,v := range kubernetesTables {
		err = ioutil.WriteFile(fmt.Sprintf("/app/monitor/prometheus/token/%s", v.ClusterName), []byte(v.Token), 0644)
		if err != nil {
			err = fmt.Errorf("Write cluster %s token file fail,%s ", v.ClusterName, err.Error())
			break
		}
	}
	if err != nil {
		return err
	}
	kubernetesPrometheusConfig,err := ioutil.ReadFile("/app/monitor/prometheus/kubernetes_prometheus.tpl")
	if err != nil {
		err = fmt.Errorf("Read kubernetes prometheus config template file fail,%s ", err.Error())
		return err
	}
	var kubernetesConfigBuffer bytes.Buffer
	kubernetesConfigBuffer.Write(tplBytes)
	kubernetesConfigBuffer.WriteString("\n")
	for _,v := range kubernetesTables {
		tmpKPConfig := string(kubernetesPrometheusConfig)
		tmpIpSplit := strings.Split(v.ApiServer, ":")
		tmpKPConfig = strings.ReplaceAll(tmpKPConfig, "{{cluster_name}}", v.ClusterName)
		tmpKPConfig = strings.ReplaceAll(tmpKPConfig, "{{api_server_ip}}", tmpIpSplit[0])
		tmpKPConfig = strings.ReplaceAll(tmpKPConfig, "{{api_server_port}}", tmpIpSplit[1])
		tmpKPConfig += "\n\n"
		kubernetesConfigBuffer.WriteString(tmpKPConfig)
	}
	err = ioutil.WriteFile("/app/monitor/prometheus/prometheus.yml", kubernetesConfigBuffer.Bytes(), 0644)
	if err != nil {
		err = fmt.Errorf("Update prometheus file fail,%s ", err.Error())
		recoverOutput,recoverError := exec.Command("/bin/sh", "-c", "rm -f /app/monitor/prometheus/prometheus.yml && cp /app/monitor/prometheus/"+backupConfigName+" /app/monitor/prometheus/prometheus.yml").Output()
		if recoverError != nil {
			log.Logger.Error("Try to rebuild prometheus config file fail", log.String("output", string(recoverOutput)), log.Error(recoverError))
		}
		return err
	}
	err = prom.ReloadConfig()
	if err != nil {
		err = fmt.Errorf("Reload prometheus config fail,%s ", err.Error())
		return err
	}
	return nil
}

func StartCronSyncKubernetesPod(interval int)  {
	t := time.NewTicker(time.Duration(interval*10)*time.Second).C
	for {
		<- t
		go syncPodToEndpoint()
	}
}

func syncPodToEndpoint() bool {
	log.Logger.Info("start to sync kubernetes pod")
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
	for _,v := range kubernetesTables {
		queryParam := m.QueryMonitorData{Legend: "$pod", SameEndpoint: true, ChartType: "line", PromQ: fmt.Sprintf("container_processes{pod=~\".*-.*\",job=\"k8s-cadvisor-%s\"}", v.ClusterName), Start: time.Now().Unix() - 600, End: time.Now().Unix()}
		series := datasource.PrometheusData(&queryParam)
		tmpGuidMap := make(map[string]int)
		var tmpKubernetesEndpointTables []*m.KubernetesEndpointRelTable
		x.SQL("select * from kubernetes_endpoint_rel where kubernete_id=?", v.Id).Find(&tmpKubernetesEndpointTables)
		tmpApiServerIp := v.ApiServer[:strings.Index(v.ApiServer, ":")]
		for _,vv := range series {
			tmpEndpointGuid := fmt.Sprintf("%s_%s_pod", vv.Name, tmpApiServerIp)
			existsFlag := false
			for _,ke := range tmpKubernetesEndpointTables {
				if ke.EndpointGuid == tmpEndpointGuid {
					existsFlag = true
					break
				}
			}
			if !existsFlag {
				if _,b := tmpGuidMap[tmpEndpointGuid]; b {
					continue
				}else{
					tmpGuidMap[tmpEndpointGuid] = 1
				}
				endpointTables = append(endpointTables, &m.EndpointTable{Guid:tmpEndpointGuid, Name:vv.Name, Ip:tmpApiServerIp, ExportType:"pod", Step:10, OsType:v.ClusterName})
				kubernetesEndpointTables = append(kubernetesEndpointTables, &m.KubernetesEndpointRelTable{KuberneteId:v.Id, EndpointGuid:tmpEndpointGuid})
			}
		}
	}
	if len(endpointTables) > 0 {
		result = true
		var tmpGuidList []string
		endpointSql := "insert into endpoint(guid,name,ip,export_type,step,os_type) values "
		for i,v := range endpointTables {
			tmpGuidList = append(tmpGuidList, v.Guid)
			log.Logger.Info("add kubernetes pod endpoint", log.String("guid", v.Guid))
			endpointSql += fmt.Sprintf("('%s','%s','%s','%s',%d,'%s')", v.Guid, v.Name, v.Ip, v.ExportType, v.Step, v.OsType)
			if i < len(endpointTables)-1 {
				endpointSql += ","
			}
		}
		_,err := x.Exec(endpointSql)
		if err != nil {
			log.Logger.Error("Update kubernetes pod to endpoint table fail", log.String("sql", endpointSql), log.Error(err))
		}
		if len(endpointGroup) > 0 {
			var tmpEndpointTables []*m.EndpointTable
			x.SQL("select id from endpoint where guid in ('"+strings.Join(tmpGuidList, "','")+"')").Find(&tmpEndpointTables)
			if len(tmpEndpointTables) > 0 {
				insertEndpointGrpSql := "insert into grp_endpoint(grp_id,endpoint_id) values "
				for _,v := range tmpEndpointTables {
					insertEndpointGrpSql += fmt.Sprintf("(%d,%d),", endpointGroup[0].Id, v.Id)
				}
				_,err = x.Exec(insertEndpointGrpSql[:len(insertEndpointGrpSql)-1])
				if err != nil {
					log.Logger.Error("Try to update endpoint group fail", log.String("sql", insertEndpointGrpSql), log.Error(err))
				}
			}
		}
	}
	if len(kubernetesEndpointTables) > 0 {
		result = true
		keRelSql := "insert into kubernetes_endpoint_rel(kubernete_id,endpoint_guid) values "
		for i,v := range kubernetesEndpointTables {
			keRelSql += fmt.Sprintf("(%d,'%s')", v.KuberneteId, v.EndpointGuid)
			if i < len(kubernetesEndpointTables)-1 {
				keRelSql += ","
			}
		}
		_,err := x.Exec(keRelSql)
		if err != nil {
			log.Logger.Error("Update kubernetes endpoint rel table fail", log.String("sql", keRelSql), log.Error(err))
		}
	}
	return result
}

func AddKubernetesPod(cluster *m.KubernetesClusterTable,podGuid,podName,namespace string) (err error,id int64,endpointGuid string) {
	apiServerIp := cluster.ApiServer[:strings.Index(cluster.ApiServer, ":")]
	endpointName := fmt.Sprintf("%s-%s", namespace, podName)
	endpointGuid = fmt.Sprintf("%s_%s_pod", endpointName, apiServerIp)
	endpointObj := m.EndpointTable{Guid: endpointGuid}
	GetEndpoint(&endpointObj)
	if endpointObj.Id <= 0 {
		execResult,err := x.Exec("insert into endpoint(guid,name,ip,export_type,step,export_version,os_type) value (?,?,?,'pod',10,?,?)", endpointGuid,endpointName,apiServerIp, namespace, cluster.ClusterName)
		if err != nil {
			return err,id,endpointGuid
		}
		lastId,_ := execResult.LastInsertId()
		if lastId <= 0 {
			err = fmt.Errorf("Insert endpoint table fail,last insert id illegal,please check server log ")
			return err,id,endpointGuid
		}
		id = lastId
	}
	var kubernetesEndpointTables []*m.KubernetesEndpointRelTable
	x.SQL("select * from kubernetes_endpoint_rel where kubernete_id=? and endpoint_guid=?", cluster.Id, endpointGuid).Find(&kubernetesEndpointTables)
	if len(kubernetesEndpointTables) <= 0 {
		_,err = x.Exec("insert into kubernetes_endpoint_rel(kubernete_id,endpoint_guid,pod_guid,namespace) value (?,?,?,?)", cluster.Id, endpointGuid, podGuid, namespace)
	}
	return err,id,endpointGuid
}

func DeleteKubernetesPod(podGuid,endpointGuid string) (err error,id int64) {
	if endpointGuid == "" {
		var kubernetesEndpointTables []*m.KubernetesEndpointRelTable
		x.SQL("select * from kubernetes_endpoint_rel where pod_guid=?", podGuid).Find(&kubernetesEndpointTables)
		if len(kubernetesEndpointTables) <= 0 {
			return err,id
		}
		_,err = x.Exec("delete from kubernetes_endpoint_rel where pod_guid=?", podGuid)
		if err != nil {
			return err,id
		}
		endpointGuid = kubernetesEndpointTables[0].EndpointGuid
	}
	endpointObj := m.EndpointTable{Guid: endpointGuid}
	GetEndpoint(&endpointObj)
	if endpointObj.Id > 0 {
		id = int64(endpointObj.Id)
		_,err = x.Exec("delete from endpoint where id=?", endpointObj.Id)
		if err != nil {
			return err,id
		}
	}
	return err,id
}

func UpdateKubernetesPodGroup(endpointId int64,group,operation string) (err error,tplId int) {
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
		_,err = x.Exec("delete from grp_endpoint where grp_id=? and endpoint_id=?", groupTables[0].Id, endpointId)
		if err != nil {
			return
		}
	}else{
		if operation == "delete" {
			return
		}
		_,err = x.Exec("insert into grp_endpoint(grp_id,endpoint_id) value (?,?)", groupTables[0].Id, endpointId)
		if err != nil {
			return
		}
	}
	_,tplObj := GetTpl(0, groupTables[0].Id, 0)
	tplId = tplObj.Id
	return
}