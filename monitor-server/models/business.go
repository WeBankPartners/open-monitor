package models

type BusinessMonitorTable struct {
	Id  int  `json:"id"`
	EndpointId  int  `json:"endpoint_id"`
	Path  string  `json:"path"`
	OwnerEndpoint  string  `json:"owner_endpoint"`
}

type BusinessUpdateDto struct {
	EndpointId  int  `json:"endpoint_id" binding:"required"`
	PathList  []*BusinessMonitorTable  `json:"path_list"`
}
