package prom

import (
	"fmt"
	"strconv"
	"encoding/json"
	"net/http"
	"strings"
	"context"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
)

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
