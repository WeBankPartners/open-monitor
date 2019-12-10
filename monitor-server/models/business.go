package models

type BusinessMonitorTable struct {
	Id  int  `json:"id"`
	EndpointId  int  `json:"endpoint_id"`
	Path  string  `json:"path"`
}

type BusinessUpdateDto struct {
	EndpointId  int  `json:"endpoint_id" binding:"required"`
	PathList  []string  `json:"path_list"`
}
