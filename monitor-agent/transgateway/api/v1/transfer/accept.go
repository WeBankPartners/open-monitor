package transfer

import (
	"encoding/json"
	"fmt"
	m "github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/models"
	"github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/util"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func AcceptPostData(c *gin.Context) {
	var param m.TransRequest
	if err := c.ShouldBindJSON(&param); err == nil {
		tmpIndex := -1
		for i, v := range m.DataCache {
			if v.Token == param.UserAuthKey {
				tmpIndex = i
				break
			}
		}
		if tmpIndex == -1 {
			if endpointName, dcErr := util.Dncrypt(param.UserAuthKey); dcErr != nil {
				util.ReturnMessage(c, util.RespJson{Code: 1, Msg: "Please register,token validate fail!"})
				return
			} else {
				var member m.Member
				member.Lock = *new(sync.RWMutex)
				member.Name = endpointName
				member.Token = param.UserAuthKey
				member.LastUpdate = time.Now()
				m.DataCache = append(m.DataCache, &member)
				m.TokenCache[param.UserAuthKey] = endpointName
				for i, v := range m.DataCache {
					if v.Token == param.UserAuthKey {
						tmpIndex = i
						break
					}
				}
			}
		}
		m.DataCache[tmpIndex].Lock.Lock()
		tNow := time.Now()
		m.DataCache[tmpIndex].LastUpdate = tNow
		m.DataCache[tmpIndex].Active = true
		for _, v := range param.MetricDataList {
			v.AttrName = strings.ReplaceAll(v.AttrName, ".", "_")
			objectString := ""
			if v.Object != nil {
				objectString = fmt.Sprintf("%s", v.Object)
			}
			attrId := v.AttrName + "__" + v.InterfaceName + "__" + objectString
			tmpFlag := false
			for ii, vv := range m.DataCache[tmpIndex].Metrics {
				if vv.Id == attrId {
					tmpFlag = true
					m.DataCache[tmpIndex].Metrics[ii].Metric = v.AttrName
					m.DataCache[tmpIndex].Metrics[ii].AttrName = v.AttrName
					m.DataCache[tmpIndex].Metrics[ii].Value = formatMetricValueData(v.MetricValue)
					m.DataCache[tmpIndex].Metrics[ii].HostIp = v.HostIp
					m.DataCache[tmpIndex].Metrics[ii].InterfaceName = v.InterfaceName
					m.DataCache[tmpIndex].Metrics[ii].Object = objectString
					m.DataCache[tmpIndex].Metrics[ii].LastUpdate = tNow
					m.DataCache[tmpIndex].Metrics[ii].Active = true
					break
				}
			}
			if !tmpFlag {
				m.DataCache[tmpIndex].Metrics = append(m.DataCache[tmpIndex].Metrics, &m.MetricObj{Id: attrId, Metric: v.AttrName, AttrName: v.AttrName, Value: formatMetricValueData(v.MetricValue), HostIp: v.HostIp, InterfaceName: v.InterfaceName, Object: objectString, LastUpdate: tNow, Active: true})
			}
		}
		m.DataCache[tmpIndex].Lock.Unlock()
		util.ReturnMessage(c, util.RespJson{Code: 0, Msg: "Success"})
	} else {
		util.ReturnMessage(c, util.RespJson{Code: 1, Msg: fmt.Sprintf("fail : %v", err)})
	}
}

func formatMetricValueData(input interface{}) (output float64) {
	rn := reflect.TypeOf(input).Name()
	if rn == "string" {
		output, _ = strconv.ParseFloat(input.(string), 64)
	} else if rn == "float64" {
		output = input.(float64)
	} else if rn == "int64" {
		output = float64(input.(int64))
	}
	return
}

func AddMember(c *gin.Context) {
	sysName := c.Query("name")
	sysIp := c.Query("ip")
	if sysName == "" {
		util.ReturnMessage(c, util.RespJson{Code: 1, Msg: "Param name cat not be null"})
		return
	}
	for _, v := range m.DataCache {
		if v.Name == sysName {
			util.ReturnMessage(c, util.RespJson{Code: 1, Msg: "Param name already exist"})
			return
		}
	}
	var member m.Member
	member.Lock = *new(sync.RWMutex)
	member.Name = sysName
	token, err := util.Encrypt([]byte(sysName))
	if err != nil {
		util.ReturnMessage(c, util.RespJson{Code: 2, Msg: fmt.Sprintf("Create token fail %v", err)})
		return
	}
	if m.MonitorUrl != "" {
		var param m.TransGatewayRequestDto
		param.Name = sysName
		if sysIp != "" {
			param.HostIp = sysIp
		} else {
			if m.LocalIp != "" {
				param.HostIp = m.LocalIp
			} else {
				param.HostIp = "127.0.0.1"
			}
		}
		param.Address = fmt.Sprintf("%s:%s", m.LocalIp, m.LocalPort)
		b, _ := json.Marshal(param)
		client := &http.Client{}
		requestObj, _ := http.NewRequest(http.MethodPost, m.MonitorUrl, strings.NewReader(string(b)))
		requestObj.Header.Set("X-Auth-Token", "default-token-used-in-server-side")
		resp, err := client.Do(requestObj)
		//resp,err := http.Post(m.MonitorUrl, "application/json", strings.NewReader(string(b)))
		if err != nil {
			util.ReturnMessage(c, util.RespJson{Code: 2, Msg: fmt.Sprintf("Update remote monitor endpoint fail %v", err)})
			return
		}
		if resp.StatusCode >= 300 {
			body, _ := ioutil.ReadAll(resp.Body)
			util.ReturnMessage(c, util.RespJson{Code: 2, Msg: fmt.Sprintf("Update remote monitor endpoint fail with code %d body: %s", resp.StatusCode, string(body))})
			resp.Body.Close()
			return
		}
	}
	member.Token = token
	member.LastUpdate = time.Now()
	m.DataCache = append(m.DataCache, &member)
	m.TokenCache[token] = sysName
	util.ReturnMessage(c, util.RespJson{Code: 0, Msg: fmt.Sprintf("Token : %s", token)})
}
