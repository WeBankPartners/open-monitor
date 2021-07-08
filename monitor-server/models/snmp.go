package models

import "time"

type SnmpExporterTable struct {
	Id  string  `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
	Modules string `json:"modules"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
}

type SnmpEndpointRelTable struct {
	Id  int  `json:"id"`
	SnmpExporter string `json:"snmp_exporter"`
	EndpointGuid string `json:"endpoint_guid"`
	Target string `json:"target"`
}
