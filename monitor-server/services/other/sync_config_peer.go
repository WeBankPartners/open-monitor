package other

import (
	"context"
	"encoding/json"
	"fmt"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var peerList []string
var timeoutCheck int64

// 配置同步给兄弟实例
func SyncPeerConfig(tplId int, param m.SyncSdConfigDto) {
	if !m.Config().Peer.Enable {
		return
	}
	if !m.PluginRunningMode {
		peerList = m.Config().Peer.OtherServerList
	} else {
		log.Info(nil, log.LOGGER_APP, fmt.Sprintf("Start sync config: id->%d param.guid->%s param.is_register->%v", tplId, param.Guid, param.IsRegister))
		//if len(m.Config().Peer.OtherServerList) == 0 {
		//	log.Warn(nil, log.LOGGER_APP, "Config peer server list is empty, return")
		//	return
		//}
		if timeoutCheck < time.Now().Unix() {
			chd, err := getCoreContainerHost()
			if err != nil {
				return
			}
			peerList = []string{}
			for _, v := range chd.Data {
				if v != m.Config().Peer.InstanceHostIp {
					peerList = append(peerList, v)
				}
			}
			timeoutCheck = time.Now().Unix() + 300
		}
	}
	if len(peerList) == 0 {
		return
	}
	for _, v := range peerList {
		if v == "" || strings.Contains(v, "127.0.0.1") || strings.Contains(v, "localhost") {
			continue
		}
		address := strings.Replace(v, "http://", "", -1)
		if m.Config().Peer.HttpPort != "" {
			address = fmt.Sprintf("%s:%s", address, m.Config().Peer.HttpPort)
		}
		tmpFlag := false
		for i := 0; i < 3; i++ {
			if requestPeerSync(tplId, address, param) {
				tmpFlag = true
				break
			}
			time.Sleep(3 * time.Second)
		}
		if !tmpFlag {
			log.Warn(nil, log.LOGGER_APP, fmt.Sprintf("Sync peer:%s config fail!!", v))
		}
	}
}

func requestPeerSync(tplId int, address string, param m.SyncSdConfigDto) bool {
	log.Info(nil, log.LOGGER_APP, fmt.Sprintf("Request sync: tplid->%d address->%s", tplId, address))
	url := fmt.Sprintf("http://%s", address)
	var req *http.Request
	if tplId > 0 {
		req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/sync/config?id=%d", url, tplId), strings.NewReader(""))
	} else {
		postData, _ := json.Marshal(param)
		req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sync/sd", url), strings.NewReader(string(postData)))
	}
	req.Header.Set("X-Auth-Token", "default-token-used-in-server-side")
	resp, err := ctxhttp.Do(context.Background(), http.DefaultClient, req)
	if err != nil {
		log.Warn(nil, log.LOGGER_APP, fmt.Sprintf("Sync peer:%s error:%v", address, err))
		return false
	}
	var result mid.RespJson
	b, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &result)
	resp.Body.Close()
	if result.Code >= 400 {
		log.Warn(nil, log.LOGGER_APP, fmt.Sprintf("sync peer:%s fail,response code:%d message:%s error:%v", address, result.Code, result.Message, result.Data))
		return false
	}
	return true
}

type coreHostDto struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

func getCoreContainerHost() (result coreHostDto, err error) {
	if m.CoreUrl == "" {
		log.Warn(nil, log.LOGGER_APP, "Get core hosts key fail, core url is null")
		return result, fmt.Errorf("get core hosts key fail, core url is null")
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/available-container-hosts", m.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core hosts key new request fail", zap.Error(err))
		return result, err
	}
	request.Header.Set("Authorization", m.GetCoreToken())
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core hosts key ctxhttp request fail", zap.Error(err))
		return result, err
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core hosts key json unmarshal result", zap.Error(err))
		return result, err
	}
	log.Debug(nil, log.LOGGER_APP, fmt.Sprintf("get core hosts, resultObj status:%s  message:%s  data:%v", result.Status, result.Message, result.Data))
	return result, nil
}
