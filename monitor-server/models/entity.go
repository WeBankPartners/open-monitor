package models

import (
	"time"
)

type CoreNotifyRequest struct{
	EventSeqNo  string  `json:"eventSeqNo"`
	EventType  string  `json:"eventType"`
	SourceSubSystem  string  `json:"sourceSubSystem"`
	OperationKey  string  `json:"operationKey"`
	OperationData  string  `json:"operationData"`
	NotifyRequired  string  `json:"notifyRequired"`
	NotifyEndpoint  string  `json:"notifyEndpoint"`
	OperationUser  string  `json:"operationUser"`
}

type CoreProcessResult struct {
	Status  string  `json:"status"`
	Message  string  `json:"message"`
	Data  CoreProcessResultData  `json:"data"`
}

type CoreProcessDataObj struct {
	ProcDefId  string  `json:"procDefId"`
	ProcDefKey  string  `json:"procDefKey"`
	ProcDefName  string  `json:"procDefName"`
	ProcDefVersion  string  `json:"procDefVersion"`
	ProcDefData  string  `json:"procDefData"`
	RootEntity  string  `json:"rootEntity"`
	Status  string  `json:"status"`
	CreatedTime  string  `json:"createdTime"`
}

type CoreProcessResultData []*CoreProcessDataObj

func (s CoreProcessResultData) Len() int {
	return len(s)
}

func (s CoreProcessResultData) Swap(i,j int)  {
	s[i], s[j] = s[j], s[i]
}

func (s CoreProcessResultData) Less(i,j int) bool {
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
	iTime,_ := time.Parse(DatetimeFormat, s[i].CreatedTime)
	jTime,_ := time.Parse(DatetimeFormat, s[j].CreatedTime)
	if iTime.Unix() > jTime.Unix() {
		tmpBool = true
	}
	return tmpBool
}

type CoreNotifyResult struct {
	Status  string  `json:"status"`
	Message  string  `json:"message"`
}

type AlarmEntity  struct {
	Status  string  `json:"status"`
	Message  string  `json:"message"`
	Data  []*AlarmEntityObj  `json:"data"`
}

type AlarmEntityObj struct {
	Id  string  `json:"id"`
	Status  string  `json:"status"`
	Subject  string  `json:"subject"`
	Content  string  `json:"content"`
	To  string  `json:"to"`
	ToMail  string  `json:"toMail"`
	ToPhone  string  `json:"toPhone"`
	ToRole  string  `json:"toRole"`
}