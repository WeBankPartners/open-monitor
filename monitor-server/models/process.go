package models

type ProcessMonitorTable struct {
	Id  int  `json:"id"`
	EndpointId  int  `json:"endpoint_id"`
	Name  string  `json:"name"`
}

type ProcessUpdateDto struct {
	EndpointId  int  `json:"endpoint_id" binding:"required"`
	ProcessList  []string  `json:"process_list"`
	Force       bool  `json:"force"`
}

type AliveCheckQueueTable struct {
	Id  int  `json:"id"`
	Message  string  `json:"message"`
}