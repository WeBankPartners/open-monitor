package models

type DbMonitorTable struct {
	Id  int  `json:"id"`
	EndpointGuid  string  `json:"endpoint_guid"`
	Name  string  `json:"name"`
	Sql   string  `json:"sql"`
	SysPanel string `json:"sys_panel"`
}

type DbMonitorUpdateDto struct {
	Id  int  `json:"id"`
	EndpointId  int  `json:"endpoint_id" binding:"required"`
	Name  string  `json:"name" binding:"required"`
	Sql   string  `json:"sql" binding:"required"`
	SysPanel string `json:"sys_panel"`
}

type DbMonitorTaskObj struct {
	DbType   string `json:"db_type"`
	Endpoint string `json:"endpoint"`
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Sql      string `json:"sql"`
}

type DbMonitorConfigQuery struct {
	EndpointGuid  string  `json:"endpoint_guid"`
	Name  string  `json:"name"`
	Sql   string  `json:"sql"`
	User  string  `json:"user"`
	Password  string  `json:"password"`
	InstanceAddress  string  `json:"instance_address"`
}

type DbMonitorListObj struct {
	SysPanel  string  `json:"sys_panel"`
	Data  []*DbMonitorTable  `json:"data"`
}

type DbMonitorSysNameDto struct {
	OldName  string  `json:"old_name" binding:"required"`
	NewName  string  `json:"new_name" binding:"required"`
	EndpointId  int  `json:"endpoint_id" binding:"required"`
}