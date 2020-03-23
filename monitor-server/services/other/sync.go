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

func SyncConfig(tplId int) {
	clusterList := m.Config().Cluster
	if len(clusterList) == 0 {
		return
	}
	for _,v := range clusterList {
		if v == "" || strings.Contains(v, "127.0.0.1") {
			continue
		}
		mid.LogInfo(fmt.Sprintf("cluster : %s", v))
		address := strings.Replace(v, "http://", "", -1)
		tmpFlag := false
		for i:=0;i<3;i++ {
			if requestClusterSync(tplId, address) {
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

func requestClusterSync(tplId int,address string) bool {
	req,_ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/wecube-monitor/api/v1/alarm/sync/config?id=%d", address, tplId), strings.NewReader(""))
	req.Header.Set("X-Auth-Token", "default-token-used-in-server-side")
	resp,err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		mid.LogInfo(fmt.Sprintf("sync cluster:%s error:%v", address, err))
		return false
	}
	if resp.StatusCode >= 400 {
		var result mid.RespJson
		b,_ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(b, &result)
		resp.Body.Close()
		mid.LogInfo(fmt.Sprintf("sync cluster:%s fail,response code:%d message:%s error:%v", address, resp.StatusCode, result.Msg, result.Data))
		return false
	}
	return true
}