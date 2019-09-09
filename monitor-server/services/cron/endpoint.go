package cron

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
)

func InitCronJob()  {

}

func syncEndpointMeta()  {

}

// old -> wait for delete
func SyncEndpointMetric(endpoint string) error {
	err,host := db.GetEndpoint(endpoint)
	if err!=nil {
		return err
	}else if host.Id == 0 {
		return fmt.Errorf("endpoint is not in table")
	}
	resp,err := http.Get(fmt.Sprintf("http://%s/metrics", host.OsIp))
	if err != nil {
		fmt.Printf("http get error %v \n", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("read body error %v \n", err)
		return err
	}
	if resp.StatusCode/100 != 2 {
		fmt.Printf("response http code %v \n", resp.StatusCode)
		return err
	}
	var endpointMetrics []string
	tmpMetric := ""
	for _,v := range strings.Split(string(body), ` `) {
		if strings.Contains(v, "#") {
			continue
		}
		if strings.Contains(v, "node_") {
			tmpStr := v[strings.Index(v, "node"):]
			if tmpStr != tmpMetric{
				//fmt.Printf("--%s--\n", tmpStr)
				endpointMetrics = append(endpointMetrics, tmpStr)
				tmpMetric = tmpStr
			}
		}
	}
	err = db.RegisterEndpoint(host.Id, endpointMetrics)
	if err != nil {
		mid.LogError("sync endpoint metric error ", err)
	}
	return err
}

func GetEndpointData(ip,port,prefix string) (error, []string) {
	var strList []string
	resp,err := http.Get(fmt.Sprintf("http://%s:%s/metrics", ip, port))
	if err != nil {
		fmt.Printf("http get error %v \n", err)
		return err,strList
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("read body error %v \n", err)
		return err,strList
	}
	if resp.StatusCode/100 != 2 {
		fmt.Printf("response http code %v \n", resp.StatusCode)
		return err,strList
	}
	tmpMetric := ""
	for _,v := range strings.Split(string(body), ` `) {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if strings.Contains(v, prefix) {
			tmpStr := v[strings.Index(v, "node"):]
			if tmpStr != tmpMetric{
				fmt.Printf("--%s--\n", tmpStr)
				strList = append(strList, tmpStr)
				tmpMetric = tmpStr
			}
		}
	}
	return nil,strList
}

func RegisteConsul(guid,ip,port string, tags []string, interval int)  {
	var param m.RegisterConsulParam
	param.Id = guid
	param.Name = guid
	param.Address = ip
	param.Port = port
	param.Tags = tags
	checks := []*m.RegisterConsulCheck{}
	checks = append(checks, &m.RegisterConsulCheck{Http:fmt.Sprintf("http://%s:%s/", ip, port), Interval:fmt.Sprintf("%ds", interval)})
	param.Checks = checks

}