package models

type RegisterParam struct {
	Type  string  `json:"type" form:"type" binding:"required"`
	ExporterIp  string  `json:"exporter_ip" form:"exporter_ip" binding:"required"`
	ExporterPort  string  `json:"exporter_port" form:"exporter_port" binding:"required"`
	Instance  string  `json:"instance" form:"instance"`
	User  string  `json:"user"`
	Password  string  `json:"password"`
}

type RegisterConsulParam struct {
	Id  string  `json:"id"`
	Name  string  `json:"name"`
	Address  string  `json:"address"`
	Port  int  `json:"port"`
	Tags  []string  `json:"tags"`
	Checks  []*RegisterConsulCheck  `json:"checks"`
}

type RegisterConsulCheck struct {
	Http  string  `json:"http"`
	Interval  string  `json:"interval"`
}

type PanelRecursiveTable struct {
	Guid  string  `json:"guid"`
	DisplayName  string  `json:"display_name"`
	Parent  string  `json:"parent"`
	Endpoint  string  `json:"endpoint"`
}

type RecursivePanelObj struct {
	DisplayName  string  `json:"display_name"`
	Charts  []*ChartModel  `json:"charts"`
	Children  []*RecursivePanelObj  `json:"children"`
}