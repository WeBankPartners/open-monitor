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
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

var consulUrl string

func RegisteConsul(guid,ip,port string, tags []string, interval int, fromCluster bool) error {
	if consulUrl == "" {
		for _, v := range m.Config().Dependence {
			if v.Name == "consul" {
				consulUrl = v.Server
				break
			}
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
		log.Logger.Error("Failed marshalling data", log.Error(err))
		return err
	}
	req,err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/agent/service/register", consulUrl), strings.NewReader(string(putData)))
	if err != nil {
		log.Logger.Error("Curl consul http request error", log.Error(err))
		return err
	}
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		log.Logger.Error("Curl consul http response error", log.Error(err))
		return err
	}
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	if string(body) != "" {
		return fmt.Errorf("consul response %s", string(body))
	}
	if !fromCluster {
		go other.SyncConfig(0, m.SyncConsulDto{Guid:guid, Ip:ip, Port:port, Tags:tags, Interval:interval, IsRegister:true})
	}
	return nil
}

func DeregisteConsul(guid string, fromCluster bool) error {
	if consulUrl == "" {
		for _, v := range m.Config().Dependence {
			if v.Name == "consul" {
				consulUrl = v.Server
				break
			}
		}
	}
	if consulUrl == "" {
		return fmt.Errorf("cat't find consul url")
	}
	req,err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/agent/service/deregister/%s", consulUrl, guid), strings.NewReader(""))
	if err != nil {
		log.Logger.Error(fmt.Sprintf("deregister %s consul error", guid), log.Error(err))
		return err
	}
	res,_ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body,_ := ioutil.ReadAll(res.Body)
	if string(body) != "" {
		return fmt.Errorf("consul response %s", string(body))
	}
	if !fromCluster {
		go other.SyncConfig(0, m.SyncConsulDto{Guid:guid, IsRegister:false})
	}
	return nil
}
