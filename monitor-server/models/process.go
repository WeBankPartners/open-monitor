package models

type ProcessMonitorTable struct {
	Id  int  `json:"id"`
	EndpointId  int  `json:"endpoint_id"`
	ProcessName  string  `json:"process_name"`
	DisplayName  string  `json:"display_name"`
	Tags         string  `json:"tags"`
}

type ProcessUpdateDto struct {
	EndpointId  int  `json:"endpoint_id" binding:"required"`
	ProcessList  []ProcessMonitorTable  `json:"process_list"`
	Check       bool  `json:"check"`
}

type AliveCheckQueueTable struct {
	Id  int  `json:"id"`
	Message  string  `json:"message"`
}

type ProcessUpdateDtoNew struct {
	EndpointId  int  `json:"endpoint_id" binding:"required"`
	ProcessList  []ProcessMonitorTable  `json:"process_list"`
	Check       bool  `json:"check"`
}

type SyncProcessObj struct {
	ProcessGuid string `json:"process_guid"`
	ProcessName string `json:"process_name"`
	ProcessTags string `json:"process_tags"`
}

type SyncProcessDto struct {
	Check int `json:"check"`
	Process []*SyncProcessObj `json:"process"`
}