package prom

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetEndpointData(param models.QueryPrometheusMetricParam) (error, []string) {
	var strList []string
	httpClient := http.Client{Timeout: 3 * time.Second}
	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%s/metrics", param.Ip, param.Port))
	if err != nil {
		var tmpErr error
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
			resp, tmpErr = httpClient.Get(fmt.Sprintf("http://%s:%s/metrics", param.Ip, param.Port))
			if tmpErr == nil {
				break
			}
		}
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "Get agent metric data fail,retry 3 times", zap.Error(tmpErr))
			return tmpErr, strList
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get agent metric response body fail", zap.Error(err))
		return err, strList
	}
	if resp.StatusCode/100 != 2 {
		log.Error(nil, log.LOGGER_APP, "Get agent metric response code error", zap.Int("code", resp.StatusCode))
		return fmt.Errorf("Get agent metric response code:%d ", resp.StatusCode), strList
	}
	for _, v := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if strings.Contains(v, ` `) {
			v = v[:strings.LastIndex(v, ` `)]
		}
		if isInBlacklist(v) {
			continue
		}
		if len(param.Prefix) == 0 && len(param.Keyword) == 0 {
			strList = append(strList, v)
			continue
		}
		tmpStrLower := strings.ToLower(v)
		tmpFlag := false
		for _, vv := range param.Prefix {
			if strings.HasPrefix(tmpStrLower, vv+"_") {
				strList = append(strList, v)
				tmpFlag = true
				break
			}
		}
		if tmpFlag {
			continue
		}
		for _, vv := range param.Keyword {
			if strings.Contains(tmpStrLower, vv) {
				strList = append(strList, v)
				break
			}
		}
	}
	log.Info(nil, log.LOGGER_APP, "Get agent metric success", zap.Int("num", len(strList)))
	return nil, strList
}

func isInBlacklist(metric string) bool {
	if len(models.Config().TagBlacklist) == 0 {
		return false
	}
	illegal := false
	for _, v := range models.Config().TagBlacklist {
		if strings.Contains(metric, "=\""+v) {
			illegal = true
			break
		}
	}
	return illegal
}

func GetSnmpMetricList(address, target string) (metricList []string, err error) {
	resp, respErr := http.Get(fmt.Sprintf("http://%s/snmp?target=%s", address, target))
	if respErr != nil {
		err = fmt.Errorf("Request snmp metric fail,%s ", respErr.Error())
		return
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if readErr != nil {
		err = fmt.Errorf("Read snmp response body fail,%s ", readErr.Error())
		return
	}
	if resp.StatusCode/100 != 2 {
		err = fmt.Errorf("Get snmp metric response code:%d ", resp.StatusCode)
		return
	}
	for _, v := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if strings.Contains(v, ` `) {
			v = v[:strings.LastIndex(v, ` `)]
		}
		metricList = append(metricList, v)
	}
	return
}
