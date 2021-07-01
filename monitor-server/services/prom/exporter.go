package prom

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
)

func GetEndpointData(ip,port string,prefix,keyword []string) (error, []string) {
	var strList []string
	resp,err := http.Get(fmt.Sprintf("http://%s:%s/metrics", ip, port))
	if err != nil {
		var tmpErr error
		for i:=0;i<6;i++ {
			time.Sleep(10*time.Second)
			resp,tmpErr = http.Get(fmt.Sprintf("http://%s:%s/metrics", ip, port))
			if tmpErr == nil {
				break
			}
		}
		if tmpErr != nil {
			log.Logger.Error("Get agent metric data fail,retry 6 times", log.Error(tmpErr))
			return tmpErr, strList
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Logger.Error("Get agent metric response body fail", log.Error(err))
		return err,strList
	}
	if resp.StatusCode/100 != 2 {
		log.Logger.Error("Get agent metric response code error", log.Int("code", resp.StatusCode))
		return err,strList
	}
	for _,v := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if strings.Contains(v, ` `) {
			v = v[:strings.LastIndex(v, ` `)]
		}
		if isInBlacklist(v) {
			continue
		}
		if len(prefix) == 0 && len(keyword) == 0 {
			strList = append(strList, v)
			continue
		}
		tmpStrLower := strings.ToLower(v)
		tmpFlag := false
		for _,vv := range prefix {
			if strings.HasPrefix(tmpStrLower, vv+"_") {
				strList = append(strList, v)
				tmpFlag = true
				break
			}
		}
		if tmpFlag {
			continue
		}
		for _,vv := range keyword {
			if strings.Contains(tmpStrLower, vv) {
				strList = append(strList, v)
				break
			}
		}
	}
	log.Logger.Info("Get agent metric success", log.Int("num", len(strList)))
	return nil,strList
}

func isInBlacklist(metric string) bool {
	if len(models.Config().TagBlacklist) == 0 {
		return false
	}
	illegal := false
	for _,v := range models.Config().TagBlacklist {
		if strings.Contains(metric, "=\""+v) {
			illegal = true
			break
		}
	}
	return illegal
}