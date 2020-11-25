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

func DeleteKubernetesCluster(id int) error {
	_,err := x.Exec("delete from kubernetes_cluster where id=?", id)
	if err != nil {
		err = fmt.Errorf("Delete db data fail,%s ", err.Error())
		return err
	}
	err = SyncKubernetesConfig()
	if err != nil {
		err = fmt.Errorf("Update prometheus config fail,%s ", err.Error())
	}
	return err
}

func SyncKubernetesConfig() error {
	var kubernetesTables []*m.KubernetesClusterTable
	x.SQL("select * from kubernetes_cluster").Find(&kubernetesTables)
	if len(kubernetesTables) == 0 {
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

func SyncKubernetesPod()  {
	for i:=0;i<3;i++ {
		time.Sleep(time.Second*30)
		tmpResult := syncPodToEndpoint()
		if tmpResult {
			break
		}
	}
}

func syncPodToEndpoint() bool {
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
