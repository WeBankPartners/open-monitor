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

func SyncConfig(tplId int, param m.SyncConsulDto) {
	if !m.Config().Cluster.Enable {
		return
	}
	clusterList := m.Config().Cluster.ServerList
	if len(clusterList) == 0 {
		return
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
	url := fmt.Sprintf("http://%s/wecube-monitor/api/v1/alarm/sync/config", address)
	var req *http.Request
	if tplId > 0 {
		req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s?id=%d", url, tplId), strings.NewReader(""))
	}else{
		postData,_ := json.Marshal(param)
		req,_ = http.NewRequest(http.MethodPost, url, strings.NewReader(string(postData)))
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