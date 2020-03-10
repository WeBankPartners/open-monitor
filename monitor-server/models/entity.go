package models

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
	Data  []*CoreProcessDataObj  `json:"data"`
}

type CoreProcessDataObj struct {
	ProcDefId  string  `json:"procDefId"`
	ProcDefKey  string  `json:"procDefKey"`
	ProcDefName  string  `json:"procDefName"`
	ProcDefVersion  string  `json:"procDefVersion"`
	ProcDefData  string  `json:"procDefData"`
	RootEntity  string  `json:"rootEntity"`
	Status  string  `json:"status"`
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
}