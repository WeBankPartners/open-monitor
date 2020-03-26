package other

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"net/http"
	"golang.org/x/net/context/ctxhttp"
	"context"
	"io/ioutil"
	"encoding/json"
	"time"
)

var clusterList []string
var selfIp string
var timeoutCheck int64

func SyncConfig(tplId int, param m.SyncConsulDto) {
	mid.LogInfo(fmt.Sprintf("start sync config: id->%d param.guid->%s param.is_register->%v", tplId, param.Guid, param.IsRegister))
	if !m.Config().Cluster.Enable {
		return
	}
	if m.CoreUrl == "" {
		clusterList = m.Config().Cluster.ServerList
		if len(clusterList) == 0 {
			return
		}
	}else{
		if len(m.Config().Cluster.ServerList) == 0 {
			mid.LogInfo("config cluster server list is empty, return")
			return
		}
		if selfIp == "" {
			selfIp = m.Config().Cluster.ServerList[0]
		}
		if timeoutCheck < time.Now().Unix() {
			chd,err := getCoreContainerHost()
			if err != nil {
				return
			}
			clusterList = []string{}
			for _,v := range chd.Data {
				if v != selfIp {
					clusterList = append(clusterList, v)
				}
			}
			timeoutCheck = time.Now().Unix() + 300
		}
	}
	for _,v := range clusterList {
		if v == "" || strings.Contains(v, "127.0.0.1") || strings.Contains(v, "localhost") {
			continue
		}
		mid.LogInfo(fmt.Sprintf("cluster : %s", v))
		address := strings.Replace(v, "http://", "", -1)
		if m.Config().Cluster.HttpPort != "" {
			address = fmt.Sprintf("%s:%s", address, m.Config().Cluster.HttpPort)
		}
		tmpFlag := false
		for i:=0;i<3;i++ {
			if requestClusterSync(tplId, address, param) {
				tmpFlag = true
				break
			}
			time.Sleep(3*time.Second)
		}
		if !tmpFlag {
			mid.LogInfo(fmt.Sprintf("sync cluster:%s config fail!!", v))
		}
	}
}

func requestClusterSync(tplId int,address string,param m.SyncConsulDto) bool {
	mid.LogInfo(fmt.Sprintf("request sync: tplid->%d address->%s", tplId, address))
	url := fmt.Sprintf("http://%s", address)
	var req *http.Request
	if tplId > 0 {
		req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/sync/config?id=%d", url, tplId), strings.NewReader(""))
	}else{
		postData,_ := json.Marshal(param)
		req,_ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sync/consul", url), strings.NewReader(string(postData)))
	}
	req.Header.Set("X-Auth-Token", "default-token-used-in-server-side")
	resp,err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		mid.LogInfo(fmt.Sprintf("sync cluster:%s error:%v", address, err))
		return false
	}
	var result mid.RespJson
	b,_ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &result)
	resp.Body.Close()
	if result.Code >= 400 {
		mid.LogInfo(fmt.Sprintf("sync cluster:%s fail,response code:%d message:%s error:%v", address, result.Code, result.Msg, result.Data))
		return false
	}
	return true
}

type coreHostDto struct {
	Status  string  `json:"status"`
	Message  string  `json:"message"`
	Data  []string  `json:"data"`
}

func getCoreContainerHost() (result coreHostDto,err error) {
	if m.CoreUrl == "" {
		mid.LogInfo("get core hosts key fail, core url is null")
		return result,fmt.Errorf("get core hosts key fail, core url is null")
	}
	request,err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/available-container-hosts", m.CoreUrl), strings.NewReader(""))
	if err != nil {
		mid.LogError("get core hosts key new request fail", err)
		return result,err
	}
	request.Header.Set("Authorization", m.TmpCoreToken)
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		mid.LogError("get core hosts key ctxhttp request fail", err)
		return result,err
	}
	defer res.Body.Close()
	b,_ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(b, &result)
	if err != nil {
		mid.LogError("get core hosts key json unmarshal result ", err)
		return result,err
	}
	mid.LogInfo(fmt.Sprintf("get core hosts, resultObj status:%s  message:%s  data:%v", result.Status, result.Message, result.Data))
	return result,nil
}