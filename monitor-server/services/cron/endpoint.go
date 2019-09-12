package cron

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"encoding/json"
	"golang.org/x/net/context/ctxhttp"
	"context"
	"strconv"
)

func GetEndpointData(ip,port string,prefix,keyword []string) (error, []string) {
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
	tmpMap := make(map[string]int)
	for _,v := range strings.Split(string(body), ` `) {
		if strings.HasPrefix(v, "#") {
			continue
		}
		for _,vv := range prefix {
			if strings.HasPrefix(v, vv+"_") {
				tmpStr := v[strings.Index(v, vv):]
				tmpMap[tmpStr] = 1
			}
		}
		for _,vv := range keyword {
			if strings.Contains(v, vv) {
				tmpStr := v
				tmpMap[tmpStr] = 1
			}
		}
	}
	for k,_ := range tmpMap {
		strList = append(strList, k)
	}
	return nil,strList
}

func RegisteConsul(guid,ip,port string, tags []string, interval int) error {
	var consulUrl string
	for _,v := range m.Config().Dependence {
		if v.Name == "consul" {
			consulUrl = v.Server
		}
	}
	if consulUrl == "" {
		return fmt.Errorf("cat't find consul url")
	}
	var param m.RegisterConsulParam
	param.Id = guid
	param.Name = guid
	param.Address = ip
	param.Port,_ = strconv.Atoi(port)
	param.Tags = tags
	checks := []*m.RegisterConsulCheck{}
	checks = append(checks, &m.RegisterConsulCheck{Http:fmt.Sprintf("http://%s:%s/", ip, port), Interval:fmt.Sprintf("%ds", interval)})
	param.Checks = checks
	putData,err := json.Marshal(param)
	if err != nil {
		mid.LogError("Failed marshalling data", err)
		return err
	}
	req,err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/agent/service/register", consulUrl), strings.NewReader(string(putData)))
	if err != nil {
		mid.LogError("curl consul http request error ", err)
		return err
	}
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		mid.LogError("curl consul http response error ", err)
		return err
	}
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	mid.LogInfo(fmt.Sprintf("guid: %s, curl register consul response : %s ", guid, string(body)))
	if string(body) != "" {
		return fmt.Errorf("consul response %s", string(body))
	}
	return nil
}

func DeregisteConsul(guid string) error {
	var consulUrl string
	for _,v := range m.Config().Dependence {
		if v.Name == "consul" {
			consulUrl = v.Server
		}
	}
	if consulUrl == "" {
		return fmt.Errorf("cat't find consul url")
	}
	req,_ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/agent/service/deregister/%s", consulUrl, guid), strings.NewReader(""))
	res,_ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	mid.LogInfo(fmt.Sprintf("guid: %s, curl deregister consul response : %s ", guid, string(body)))
	if string(body) != "" {
		return fmt.Errorf("consul response %s", string(body))
	}
	return nil
}