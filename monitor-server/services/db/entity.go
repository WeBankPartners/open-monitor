package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
	"net/http"
	"strings"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"golang.org/x/net/context/ctxhttp"
	"context"
	"io/ioutil"
	"encoding/json"
	"time"
)

var coreProcessKey string
const tmpCoreToken = `Bearer eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJ3ZHNfc3lzdGVtIiwiaWF0IjoxNTgyMDkyODYyLCJ0eXBlIjoiYWNjZXNzVG9rZW4iLCJjbGllbnRUeXBlIjoiVVNFUiIsImV4cCI6MzQ3NDI1Mjg2MiwiYXV0aG9yaXR5IjoiW2FkbWluLElNUExFTUVOVEFUSU9OX1dPUktGTE9XX0VYRUNVVElPTixJTVBMRU1FTlRBVElPTl9CQVRDSF9FWEVDVVRJT04sSU1QTEVNRU5UQVRJT05fQVJUSUZBQ1RfTUFOQUdFTUVOVCxNT05JVE9SX01BSU5fREFTSEJPQVJELE1PTklUT1JfTUVUUklDX0NPTkZJRyxNT05JVE9SX0NVU1RPTV9EQVNIQk9BUkQsTU9OSVRPUl9BTEFSTV9DT05GSUcsTU9OSVRPUl9BTEFSTV9NQU5BR0VNRU5ULENPTExBQk9SQVRJT05fUExVR0lOX01BTkFHRU1FTlQsQ09MTEFCT1JBVElPTl9XT1JLRkxPV19PUkNIRVNUUkFUSU9OLEFETUlOX1NZU1RFTV9QQVJBTVMsQURNSU5fUkVTT1VSQ0VTX01BTkFHRU1FTlQsQURNSU5fVVNFUl9ST0xFX01BTkFHRU1FTlQsQURNSU5fQ01EQl9NT0RFTF9NQU5BR0VNRU5ULENNREJfQURNSU5fQkFTRV9EQVRBX01BTkFHRU1FTlQsQURNSU5fUVVFUllfTE9HLE1FTlVfQURNSU5fUEVSTUlTU0lPTl9NQU5BR0VNRU5ULE1FTlVfREVTSUdOSU5HX0NJX0RBVEFfRU5RVUlSWSxNRU5VX0RFU0lHTklOR19DSV9JTlRFR1JBVEVEX1FVRVJZX0VYRUNVVElPTixNRU5VX0NNREJfREVTSUdOSU5HX0VOVU1fRU5RVUlSWSxNRU5VX0RFU0lHTklOR19DSV9EQVRBX01BTkFHRU1FTlQsTUVOVV9ERVNJR05JTkdfQ0lfSU5URUdSQVRFRF9RVUVSWV9NQU5BR0VNRU5ULE1FTlVfQ01EQl9ERVNJR05JTkdfRU5VTV9NQU5BR0VNRU5ULE1FTlVfSURDX1BMQU5OSU5HX0RFU0lHTixNRU5VX0lEQ19SRVNPVVJDRV9QTEFOTklORyxNRU5VX0FQUExJQ0FUSU9OX0FSQ0hJVEVDVFVSRV9ERVNJR04sTUVOVV9BUFBMSUNBVElPTl9ERVBMT1lNRU5UX0RFU0lHTixNRU5VX0FETUlOX0NNREJfTU9ERUxfTUFOQUdFTUVOVCxNRU5VX0NNREJfQURNSU5fQkFTRV9EQVRBX01BTkFHRU1FTlQsTUVOVV9BRE1JTl9RVUVSWV9MT0csSk9CU19TRVJWSUNFX0NBVEFMT0dfTUFOQUdFTUVOVCxKT0JTX1RBU0tfTUFOQUdFTUVOVF0ifQ.XbPpkiS6AG7zSLHYxFacU3gnyQMWcIvxqXbI3MSlxTGQqJDWrdPUCyyvE0lfJrPoG69GC2gI25Ys_WyGA71E8A`

func getCoreProcessKey() string {
	if coreProcessKey != "" {
		return coreProcessKey
	}
	if m.CoreUrl == "" {
		mid.LogInfo("get core process key fail, core url is null")
		return ""
	}
	request,err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/process/definitions", m.CoreUrl), strings.NewReader(""))
	if err != nil {
		mid.LogError("get core process key new request fail", err)
		return ""
	}
	request.Header.Set("Authorization", tmpCoreToken)
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		mid.LogError("get core process key ctxhttp request fail", err)
		return ""
	}
	defer res.Body.Close()
	b,_ := ioutil.ReadAll(res.Body)
	var resultObj m.CoreProcessResult
	err = json.Unmarshal(b, &resultObj)
	if err != nil {
		mid.LogError("get core process key json unmarshal result ", err)
		return ""
	}
	mid.LogInfo(fmt.Sprintf("get core process, resultObj status:%s  message:%s  data length:%d", resultObj.Status, resultObj.Message, len(resultObj.Data)))
	for _,v := range resultObj.Data {
		mid.LogInfo(fmt.Sprintf("process result name:%s", v.ProcDefName))
		if strings.Contains(v.ProcDefName, "监控告警处理") {
			coreProcessKey = v.ProcDefKey
		}
	}
	return coreProcessKey
}

func NotifyCoreEvent(endpoint string,strategyId int) error {
	var alarms []*m.AlarmTable
	x.SQL("SELECT id,status FROM alarm WHERE endpoint=? AND strategy_id=? ORDER BY id DESC", endpoint, strategyId).Find(&alarms)
	if len(alarms) == 0 {
		return fmt.Errorf("can not find any alarm with endpoint:%s startegy_id:%d", endpoint, strategyId)
	}

	coreKey := getCoreProcessKey()
	if coreKey == "" {
		return fmt.Errorf("notify core event fail, core key is null")
	}
	var requestParam m.CoreNotifyRequest
	requestParam.EventSeqNo = fmt.Sprintf("%d-%s-%d", alarms[0].Id, alarms[0].Status, time.Now().Unix())
	requestParam.EventType = "alarm"
	requestParam.SourceSubSystem = "monitor"
	requestParam.OperationKey = coreKey
	requestParam.OperationData = fmt.Sprintf("alarm_%d", alarms[0].Id)
	requestParam.OperationUser = "wds_system"
	mid.LogInfo(fmt.Sprintf("notify request data --> eventSeqNo:%s operationKey:%s operationData:%s", requestParam.EventSeqNo, coreKey, requestParam.OperationData))
	b,_ := json.Marshal(requestParam)
	request,err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", m.CoreUrl), strings.NewReader(string(b)))
	request.Header.Set("Authorization", tmpCoreToken)
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		mid.LogError("notify core event new request fail", err)
		return err
	}
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		mid.LogError("notify core event ctxhttp request fail", err)
		return err
	}
	defer res.Body.Close()
	resultBody,_ := ioutil.ReadAll(res.Body)
	var resultObj m.CoreNotifyResult
	err = json.Unmarshal(resultBody, &resultObj)
	if err != nil {
		mid.LogError("notify core event unmarshal json body fail", err)
		return err
	}
	mid.LogInfo(fmt.Sprintf("result -> status:%s  message:%s", resultObj.Status, resultObj.Message))
	return nil
}

func GetAlarmEvent(alarmType string,id int) (result m.AlarmEntityObj,err error) {
	result.Id = fmt.Sprintf("%s_%d", alarmType, id)
	if alarmType == "alarm" {
		var alarms []*m.AlarmTable
		err = x.SQL("SELECT * FROM alarm WHERE id=?", id).Find(&alarms)
		if err != nil {
			return result,err
		}
		if len(alarms) == 0 {
			return result,fmt.Errorf("%s %d can not fetch any data", alarmType, id)
		}
		result.Status = alarms[0].Status
		mailMap := make(map[string]bool)
		for _,v := range GetMailByStrategy(alarms[0].StrategyId) {
			mailMap[v] = true
		}

		var recursiveData []*m.PanelRecursiveTable
		x.SQL("SELECT * FROM panel_recursive").Find(&recursiveData)
		if len(recursiveData) > 0 {
			for _,v := range recursiveData {
				if strings.Contains(v.Endpoint, alarms[0].Endpoint) {
					for _,vv := range strings.Split(v.Email, ",") {
						mailMap[vv] = true
					}
					for _,vv := range strings.Split(v.Parent, "^") {
						tmpToRecursiveMail,_ := searchRecursiveParent(recursiveData,[]string{},[]string{},vv)
						mid.LogInfo(fmt.Sprintf("parent: %s  mail: %v", vv, tmpToRecursiveMail))
						for _,vvv := range tmpToRecursiveMail {
							mailMap[vvv] = true
						}
					}
				}
			}
		}
		var toMail []string
		for k,_ := range mailMap {
			toMail = append(toMail, k)
		}
		result.To = strings.Join(toMail, ",")
		result.Subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", alarms[0].Status, alarms[0].SPriority, alarms[0].Endpoint, alarms[0].SMetric)
		result.Content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s",alarms[0].Endpoint,alarms[0].Status,alarms[0].SMetric,alarms[0].StartValue,alarms[0].SCond,alarms[0].SLast,alarms[0].SPriority,alarms[0].Content,alarms[0].Start.Format(m.DatetimeFormat))
		mid.LogInfo(fmt.Sprintf("alarm event --> id:%s status:%s to:%s subejct:%s content:%s", result.Id, result.Status, result.To, result.Subject, result.Content))
	}
	return result,err
}

func searchRecursiveParent(data []*m.PanelRecursiveTable,tmpEmail,tmpPhone []string,tmpParent string) (email,phone []string) {
	var parent []string
	email = tmpEmail
	phone = tmpPhone
	for _,v := range data {
		if v.Guid == tmpParent {
			parent = strings.Split(v.Parent, "^")
			for _,vv := range strings.Split(v.Email, ",") {
				email = append(email, vv)
			}
			for _,vv := range strings.Split(v.Phone, ",") {
				phone = append(phone, vv)
			}
			break
		}
	}
	if len(parent) > 0 {
		for _,v := range parent {
			tEmail,tPhone := searchRecursiveParent(data,email,phone,v)
			for _,vv := range tEmail {
				email = append(email, vv)
			}
			for _,vv := range tPhone {
				phone = append(phone, vv)
			}
		}
	}
	return email,phone
}