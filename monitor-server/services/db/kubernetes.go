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

