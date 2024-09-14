package models

import (
	"time"
)

type CoreNotifyRequest struct {
	EventSeqNo      string `json:"eventSeqNo"`
	EventType       string `json:"eventType"`
	SourceSubSystem string `json:"sourceSubSystem"`
	OperationKey    string `json:"operationKey"`
	OperationData   string `json:"operationData"`
	NotifyRequired  string `json:"notifyRequired"`
	NotifyEndpoint  string `json:"notifyEndpoint"`
	OperationUser   string `json:"operationUser"`
}

type CoreProcessResult struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    CoreProcessResultData `json:"data"`
}

type CoreProcessDataObj struct {
	ProcDefId            string      `json:"procDefId"`
	ProcDefKey           string      `json:"procDefKey"`
	ProcDefName          string      `json:"procDefName"`
	ProcDefVersion       string      `json:"procDefVersion"`
	ProcDefData          string      `json:"procDefData"`
	RootEntity           interface{} `json:"rootEntity"`
	Status               string      `json:"status"`
	CreatedTime          string      `json:"createdTime"`
	RootEntityExpression string      `json:"rootEntityExpression"`
}

type CoreProcessResultData []*CoreProcessDataObj

func (s CoreProcessResultData) Len() int {
	return len(s)
}

func (s CoreProcessResultData) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s CoreProcessResultData) Less(i, j int) bool {
	tmpBool := false
	//iVersion,_ := strconv.Atoi(s[i].ProcDefVersion)
	//jVersion,_ := strconv.Atoi(s[j].ProcDefVersion)
	//if iVersion > jVersion {
	//	tmpBool = true
	//}else if iVersion == jVersion {
	//	iTime,_ := time.Parse(DatetimeFormat, s[i].CreatedTime)
	//	jTime,_ := time.Parse(DatetimeFormat, s[j].CreatedTime)
	//	if iTime.Unix() > jTime.Unix() {
	//		tmpBool = true
	//	}
	//}
	iTime, _ := time.Parse(DatetimeFormat, s[i].CreatedTime)
	jTime, _ := time.Parse(DatetimeFormat, s[j].CreatedTime)
	if iTime.Unix() > jTime.Unix() {
		tmpBool = true
	}
	return tmpBool
}

type CoreNotifyResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AlarmEntity struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    []*AlarmEntityObj `json:"data"`
}

type AlarmEntityObj struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	SmsContent  string `json:"smsContent"`
	To          string `json:"to"`
	ToMail      string `json:"toMail"`
	ToPhone     string `json:"toPhone"`
	ToRole      string `json:"toRole"`
}

type AlarmEventEntity struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    []*AlarmEventEntityObj `json:"data"`
}

type AlarmEventEntityObj struct {
	Id            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Handler       string `json:"handler"`
	HandleRole    string `json:"handleRole"`
	Content       string `json:"content"`
	Priority      string `json:"priority"`
	Message       string `json:"message"`
	StartTime     string `json:"startTime"`
	Endpoint      string `json:"endpoint"`
	Detail        string `json:"detail"`
	NotifyEventId string `json:"notifyEventId"`
}

type AlarmEventUpdateResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

type EndpointEntityResp struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    []*EndpointEntityObj `json:"data"`
}

type EndpointEntityObj struct {
	Id          string `json:"id" xorm:"guid"`
	DisplayName string `json:"displayName" xorm:"guid"`
	Name        string `json:"name" xorm:"name"`
	Ip          string `json:"ip" xorm:"ip"`
	MonitorType string `json:"monitorType" xorm:"monitor_type"`
}

type EndpointGroupEntityResp struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Data    []*EndpointGroupEntityObj `json:"data"`
}

type EndpointGroupEntityObj struct {
	Id          string `json:"id" xorm:"guid"`
	DisplayName string `json:"displayName" xorm:"display_name"`
	Description string `json:"description" xorm:"description"`
	MonitorType string `json:"monitorType" xorm:"monitor_type"`
}

type ServiceGroupEntityResp struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    []*ServiceGroupEntityObj `json:"data"`
}

type ServiceGroupEntityObj struct {
	Id          string `json:"id" xorm:"guid"`
	DisplayName string `json:"displayName" xorm:"display_name"`
	Description string `json:"description" xorm:"description"`
	Parent      string `json:"parent" xorm:"parent"`
	ServiceType string `json:"serviceType" xorm:"service_type"`
}

type MonitorTypeEntityResp struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    []*MonitorTypeEntityObj `json:"data"`
}

type MonitorTypeEntityObj struct {
	Id          string `json:"id" xorm:"guid"`
	DisplayName string `json:"displayName" xorm:"display_name"`
	Description string `json:"description" xorm:"description"`
}

type LogMonitorTemplateEntityResp struct {
	Status  string                         `json:"status"`
	Message string                         `json:"message"`
	Data    []*LogMonitorTemplateEntityObj `json:"data"`
}

type LogMonitorTemplateEntityObj struct {
	Id          string `json:"id" xorm:"guid"`
	DisplayName string `json:"displayName" xorm:"name"`
	LogType     string `json:"logType" xorm:"log_type"`
}

type AnalyzeTransParam struct {
	EndpointList     []string `json:"endpointList"`
	ServiceGroupList []string `json:"serviceGroupList"`
}

type AnalyzeTransData struct {
	MonitorType               []string `json:"monitorType"`
	EndpointGroup             []string `json:"endpointGroup"`
	CustomMetricServiceGroup  []string `json:"customMetricServiceGroup"`
	CustomMetricEndpointGroup []string `json:"customMetricEndpointGroup"`
	CustomMetricMonitorType   []string `json:"CustomMetricMonitorType"`
	LogMonitorServiceGroup    []string `json:"logMonitorServiceGroup"`
	LogMonitorTemplate        []string `json:"logMonitorTemplate"`
	StrategyServiceGroup      []string `json:"strategyServiceGroup"`
	StrategyEndpointGroup     []string `json:"strategyEndpointGroup"`
	LogKeywordServiceGroup    []string `json:"logKeywordServiceGroup"`
	DashboardIdList           []string `json:"dashboardIdList"`
}
