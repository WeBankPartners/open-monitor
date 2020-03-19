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
	"sort"
)

var coreProcessKey string
const tmpCoreToken = ***REMOVED***

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

func GetCoreEventList() (result m.CoreProcessResult,err error) {
	if m.CoreUrl == "" {
		mid.LogInfo("get core process key fail, core url is null")
		return result,fmt.Errorf("get core process key fail, core url is null")
	}
	request,err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/process/definitions", m.CoreUrl), strings.NewReader(""))
	if err != nil {
		mid.LogError("get core process key new request fail", err)
		return result,err
	}
	request.Header.Set("Authorization", tmpCoreToken)
	res,err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		mid.LogError("get core process key ctxhttp request fail", err)
		return result,err
	}
	defer res.Body.Close()
	b,_ := ioutil.ReadAll(res.Body)
	//testResult := `{"status":"OK","message":"Success","data":[{"procDefId":"rRqRf87S2Ev","procDefKey":"wecube1581455678621","procDefName":"L3_数据库首次部署_V0.2","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:subsys","createdTime":"2020-02-24 23:31:15","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShPf61z2Bv","procDefKey":"wecube1583340833459","procDefName":null,"procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 00:55:44","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqRgG6S2F1","procDefKey":"wecube1581455678621","procDefName":"L3_子系统首次部署[CLB+APP]_V0.1","procDefVersion":"2","status":"deployed","procDefData":null,"rootEntity":"wecmdb:subsys","createdTime":"2020-02-24 23:31:21","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShRwAZz2BH","procDefKey":"wecube1583340833459","procDefName":"创建主机2","procDefVersion":"3","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 01:04:48","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShNC9mz2Bj","procDefKey":"wecube1583340191630","procDefName":"ITSM-创建主机表单","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 00:49:16","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rTnNXdiW2Bo","procDefKey":"wecube1581437988379","procDefName":"L1_应用服务均衡配置V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:business_app_instance","createdTime":"2020-03-16 15:57:07","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rSZ3YAr42Cp","procDefKey":"wecube1583979798296","procDefName":"aaa","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-12 10:23:58","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRMtOU6S2Bi","procDefKey":"wecube1581626244634","procDefName":"数据库包解析替换","procDefVersion":"null","status":"draft","procDefData":null,"rootEntity":"wecmdb:subsys","createdTime":"2020-02-28 16:16:12","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShTWibz2BT","procDefKey":"wecube1583340833459","procDefName":"创建主机3","procDefVersion":"5","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 01:14:23","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rSOH7zF42Br","procDefKey":"wecube1581454682287","procDefName":"L3_监控告警处理V0.2","procDefVersion":"2","status":"deployed","procDefData":null,"rootEntity":"service-mgmt:task","createdTime":"2020-03-10 15:50:27","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqR6DPS2BG","procDefKey":"wecube1576991964758","procDefName":"L2_应用资源集合CVM安装及初始化_V0.2","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_set","createdTime":"2020-02-24 23:30:42","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rSr9I81z2Cc","procDefKey":"wecube1581454682287","procDefName":"L3_监控告警处理V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"service-mgmt:task","createdTime":"2020-03-06 15:13:29","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqR95US2D5","procDefKey":"wecube1577352660224","procDefName":"L2_业务区域子网网络初始化_V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:business_zone","createdTime":"2020-02-24 23:30:51","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqRcpGS2DY","procDefKey":"wecube1578296432413","procDefName":"L2_APP应用首次部署_V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:unit","createdTime":"2020-02-24 23:31:04","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqRa1yS2Dt","procDefKey":"wecube1577437368910","procDefName":"L1_DB资源集合MYSQL资源实例创建_V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_set","createdTime":"2020-02-24 23:30:55","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShPizzz2BB","procDefKey":"wecube1583340833459","procDefName":"创建主机2","procDefVersion":"2","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 00:55:57","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqR7RSS2Cs","procDefKey":"wecube1577351326050","procDefName":"L2_地域数据中心VPC级网络初始化_V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:data_center","createdTime":"2020-02-24 23:30:47","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqRbgTS2DI","procDefKey":"wecube1581437316084","procDefName":"L1_应用LB资源集合CLB资源实例创建_V0.1","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_set","createdTime":"2020-02-24 23:31:00","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShNTe8z2Bp","procDefKey":"wecube1583340191630","procDefName":"ITSM-创建主机表单","procDefVersion":"2","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 00:50:21","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rShSG34z2BN","procDefKey":"wecube1583340833459","procDefName":"创建主机3","procDefVersion":"4","status":"deployed","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-05 01:09:22","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rT6Miau42CN","procDefKey":"wecube1583340833459","procDefName":"创建主机3","procDefVersion":"null","status":"draft","procDefData":null,"rootEntity":"wecmdb:resource_instance","createdTime":"2020-03-13 18:03:54","permissionToRole":null,"taskNodeInfos":[]},{"procDefId":"rRqRhwZS2FC","procDefKey":"wecube1581626244634","procDefName":"L3_子系统升级部署[DB+APP]_V0.3","procDefVersion":"1","status":"deployed","procDefData":null,"rootEntity":"wecmdb:subsys","createdTime":"2020-02-24 23:31:24","permissionToRole":null,"taskNodeInfos":[]}]}`
	//b := []byte(testResult)
	err = json.Unmarshal(b, &result)
	if err != nil {
		mid.LogError("get core process key json unmarshal result ", err)
		return result,err
	}
	sort.Sort(result.Data)
	mid.LogInfo(fmt.Sprintf("get core process, resultObj status:%s  message:%s  data length:%d", result.Status, result.Message, len(result.Data)))
	return result,nil
}

func getCoreEventKey(status,endpoint string) []string {
	var result []string
	firingMap := make(map[string]bool)
	recoverMap := make(map[string]bool)
	var recursiveData []*m.PanelRecursiveTable
	x.SQL("SELECT * FROM panel_recursive").Find(&recursiveData)
	if len(recursiveData) > 0 {
		for _,v := range recursiveData {
			if strings.Contains(v.Endpoint, endpoint) {
				if v.FiringCallbackKey != "" {
					firingMap[v.FiringCallbackKey] = true
				}
				if v.RecoverCallbackKey != "" {
					recoverMap[v.RecoverCallbackKey] = true
				}
				for _,vv := range strings.Split(v.Parent, "^") {
					_, _, _, tmpFiring, tmpRecover := searchRecursiveParent(recursiveData, []string{}, []string{}, []string{}, []string{}, []string{}, vv)
					for _,vvv := range tmpFiring {
						firingMap[vvv] = true
					}
					for _,vvv := range tmpRecover {
						recoverMap[vvv] = true
					}
				}
			}
		}
	}else{
		return result
	}
	if status == "firing" {
		for k,_ := range firingMap {
			if k == "" {
				continue
			}
			result = append(result, k)
		}
	}else{
		for k,_ := range recoverMap {
			if k == "" {
				continue
			}
			result = append(result, k)
		}
	}
	return result
}

func NotifyCoreEvent(endpoint string,strategyId int) error {
	var alarms []*m.AlarmTable
	x.SQL("SELECT id,status FROM alarm WHERE endpoint=? AND strategy_id=? ORDER BY id DESC", endpoint, strategyId).Find(&alarms)
	if len(alarms) == 0 {
		return fmt.Errorf("can not find any alarm with endpoint:%s startegy_id:%d", endpoint, strategyId)
	}
	eventKeys := getCoreEventKey(alarms[0].Status, endpoint)
	if len(eventKeys) == 0 {
		return fmt.Errorf("notify core event fail, event key is null")
	}
	for i,coreKey := range eventKeys {
		var requestParam m.CoreNotifyRequest
		requestParam.EventSeqNo = fmt.Sprintf("%d-%s-%d-%d", alarms[0].Id, alarms[0].Status, time.Now().Unix(), i)
		requestParam.EventType = "alarm"
		requestParam.SourceSubSystem = "monitor"
		requestParam.OperationKey = coreKey
		requestParam.OperationData = fmt.Sprintf("%d", alarms[0].Id)
		requestParam.OperationUser = "wds_system"
		mid.LogInfo(fmt.Sprintf("notify request data --> eventSeqNo:%s operationKey:%s operationData:%s", requestParam.EventSeqNo, coreKey, requestParam.OperationData))
		b, _ := json.Marshal(requestParam)
		request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/operation-events", m.CoreUrl), strings.NewReader(string(b)))
		request.Header.Set("Authorization", tmpCoreToken)
		request.Header.Set("Content-Type", "application/json")
		if err != nil {
			mid.LogError("notify core event new request fail", err)
			return err
		}
		res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
		if err != nil {
			mid.LogError("notify core event ctxhttp request fail", err)
			return err
		}
		resultBody, _ := ioutil.ReadAll(res.Body)
		var resultObj m.CoreNotifyResult
		err = json.Unmarshal(resultBody, &resultObj)
		res.Body.Close()
		if err != nil {
			mid.LogError("notify core event unmarshal json body fail", err)
			return err
		}
		mid.LogInfo(fmt.Sprintf("result -> status:%s  message:%s", resultObj.Status, resultObj.Message))
	}
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
		phoneMap := make(map[string]bool)
		roleMap := make(map[string]bool)
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
					for _,vv := range strings.Split(v.Phone, ",") {
						phoneMap[vv] = true
					}
					for _,vv := range strings.Split(v.Role, ",") {
						roleMap[vv] = true
					}
					for _,vv := range strings.Split(v.Parent, "^") {
						tmpToRecursiveMail,tmpToRecursivePhone,tmpToRecursiveRole,_,_ := searchRecursiveParent(recursiveData,[]string{},[]string{},[]string{},[]string{},[]string{},vv)
						mid.LogInfo(fmt.Sprintf("parent: %s  mail: %v phone: %v role: %v", vv, tmpToRecursiveMail, tmpToRecursivePhone, tmpToRecursiveRole))
						for _,vvv := range tmpToRecursiveMail {
							mailMap[vvv] = true
						}
						for _,vvv := range tmpToRecursivePhone {
							phoneMap[vvv] = true
						}
						for _,vvv := range tmpToRecursiveRole {
							roleMap[vvv] = true
						}
					}
				}
			}
		}
		var toMail,toPhone,toRole []string
		for k,_ := range mailMap {
			if k == "" {
				continue
			}
			toMail = append(toMail, k)
		}
		for k,_ := range phoneMap {
			if k == "" {
				continue
			}
			toPhone = append(toPhone, k)
		}
		for k,_ := range roleMap {
			if k == "" {
				continue
			}
			toRole = append(toRole, k)
		}
		if len(toRole) > 0 {
			var roleTable []*m.RoleTable
			x.SQL(fmt.Sprintf("SELECT * FROM role WHERE id IN (%s)", strings.Join(toRole, ","))).Find(&roleTable)
			toRole = []string{}
			for _,v := range roleTable {
				toRole = append(toRole, v.DisplayName)
			}
		}
		result.To = strings.Join(toMail, ",")
		result.ToMail = strings.Join(toMail, ",")
		result.ToPhone = strings.Join(toPhone, ",")
		result.ToRole = strings.Join(toRole, ",")
		result.Subject = fmt.Sprintf("[%s][%s] Endpoint:%s Metric:%s", alarms[0].Status, alarms[0].SPriority, alarms[0].Endpoint, alarms[0].SMetric)
		result.Content = fmt.Sprintf("Endpoint:%s \r\nStatus:%s\r\nMetric:%s\r\nEvent:%.3f%s\r\nLast:%s\r\nPriority:%s\r\nNote:%s\r\nTime:%s",alarms[0].Endpoint,alarms[0].Status,alarms[0].SMetric,alarms[0].StartValue,alarms[0].SCond,alarms[0].SLast,alarms[0].SPriority,alarms[0].Content,alarms[0].Start.Format(m.DatetimeFormat))
		mid.LogInfo(fmt.Sprintf("alarm event --> id:%s status:%s to:%s subejct:%s content:%s", result.Id, result.Status, result.To, result.Subject, result.Content))
	}
	return result,err
}

func searchRecursiveParent(data []*m.PanelRecursiveTable,tmpEmail,tmpPhone,tmpRole,tmpFiringKey,tmpRecoverKey []string,tmpParent string) (email,phone,role,firing,recover []string) {
	var parent []string
	email = tmpEmail
	phone = tmpPhone
	role = tmpRole
	firing = tmpFiringKey
	recover = tmpRecoverKey
	for _,v := range data {
		if v.Guid == tmpParent {
			parent = strings.Split(v.Parent, "^")
			for _,vv := range strings.Split(v.Email, ",") {
				email = append(email, vv)
			}
			for _,vv := range strings.Split(v.Phone, ",") {
				phone = append(phone, vv)
			}
			for _,vv := range strings.Split(v.Role, ",") {
				role = append(role, vv)
			}
			firing = append(firing, v.FiringCallbackKey)
			recover = append(recover, v.RecoverCallbackKey)
			break
		}
	}
	if len(parent) > 0 {
		for _,v := range parent {
			tEmail,tPhone,tRole,tFiring,tRecover := searchRecursiveParent(data,email,phone,role,firing,recover,v)
			for _,vv := range tEmail {
				email = append(email, vv)
			}
			for _,vv := range tPhone {
				phone = append(phone, vv)
			}
			for _,vv := range tRole {
				role = append(role, vv)
			}
			for _,vv := range tFiring {
				firing = append(firing, vv)
			}
			for _,vv := range tRecover {
				recover = append(recover, vv)
			}
		}
	}
	return email,phone,role,firing,recover
}