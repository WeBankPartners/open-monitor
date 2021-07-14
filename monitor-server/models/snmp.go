package models

import "time"

type SnmpExporterTable struct {
	Id  string  `json:"id" binding:"required"`
	ScrapeInterval  int  `json:"scrape_interval"`
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

type PluginSnmpExporterRequest struct {
	RequestId string                             `json:"requestId"`
	Inputs    []*PluginSnmpExporterRequestObj `json:"inputs"`
}

type PluginSnmpExporterRequestObj struct {
	CallbackParameter string `json:"callbackParameter"`
	Id                string `json:"id"`
	Address           string `json:"address"`
	ScrapeInterval    string `json:"scrapeInterval"`
}

type PluginSnmpExporterResp struct {
	ResultCode    string                      `json:"resultCode"`
	ResultMessage string                      `json:"resultMessage"`
	Results       PluginSnmpExporterOutput `json:"results"`
}

type PluginSnmpExporterOutput struct {
	Outputs []*PluginSnmpExporterOutputObj `json:"outputs"`
}

type PluginSnmpExporterOutputObj struct {
	CallbackParameter string `json:"callbackParameter"`
	Id                string `json:"id"`
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
	ErrorDetail       string `json:"errorDetail,omitempty"`
}