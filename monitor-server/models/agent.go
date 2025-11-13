package models

import "time"

type RegisterParam struct {
	Type         string `json:"type" form:"type" binding:"required"`
	ExporterIp   string `json:"exporter_ip" form:"exporter_ip" binding:"required"`
	ExporterPort string `json:"exporter_port" form:"exporter_port" binding:"required"`
	Instance     string `json:"instance" form:"instance"`
	User         string `json:"user"`
	Password     string `json:"password"`
}

type RegisterParamNew struct {
	Guid              string `json:"guid"`
	Type              string `json:"type"`
	Name              string `json:"name"`
	Ip                string `json:"ip"`
	Port              string `json:"port"`
	User              string `json:"user"`
	Password          string `json:"password"`
	Method            string `json:"method"`
	Url               string `json:"url"`
	AddDefaultGroup   bool   `json:"add_default_group"`
	DefaultGroupName  string `json:"default_group_name"`
	AgentManager      bool   `json:"agent_manager"`
	FetchMetric       bool   `json:"fetch_metric"`
	Step              int    `json:"step"`
	ExportAddress     string `json:"export_address"`
	Cluster           string `json:"cluster"`
	ProxyExporter     string `json:"proxy_exporter"`
	ProcessName       string `json:"process_name"`
	Tags              string `json:"tags"`
	KubernetesCluster string `json:"kubernetes_cluster"`
	NodeIp            string `json:"node_ip"`
}

type RegisterConsulParam struct {
	Id      string                 `json:"id"`
	Name    string                 `json:"name"`
	Address string                 `json:"address"`
	Port    int                    `json:"port"`
	Tags    []string               `json:"tags"`
	Checks  []*RegisterConsulCheck `json:"checks"`
}

type RegisterConsulCheck struct {
	Http     string `json:"http"`
	Interval string `json:"interval"`
}

type PanelRecursiveTable struct {
	Guid                string    `json:"guid"`
	DisplayName         string    `json:"display_name"`
	Parent              string    `json:"parent"`
	Endpoint            string    `json:"endpoint"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	Role                string    `json:"role"`
	FiringCallbackName  string    `json:"firing_callback_name"`
	FiringCallbackKey   string    `json:"firing_callback_key"`
	RecoverCallbackName string    `json:"recover_callback_name"`
	RecoverCallbackKey  string    `json:"recover_callback_key"`
	ObjType             string    `json:"obj_type"`
	UpdateAt            time.Time `json:"update_at"`
}

type RecursivePanelObj struct {
	DisplayName string               `json:"display_name"`
	Charts      []*ChartModel        `json:"charts"`
	Children    []*RecursivePanelObj `json:"children"`
}

type TransGatewayRequestDto struct {
	Name    string   `json:"name"`
	HostIp  string   `json:"host_ip"`
	Address string   `json:"address"`
	Metrics []string `json:"metrics"`
}

type TransGatewayMetricDto struct {
	Params []*TransGatewayRequestDto `json:"params"`
}

type UpdateEndpointTelnetParam struct {
	Guid   string               `json:"guid" binding:"required"`
	Config []*EndpointTelnetObj `json:"config"`
}

type EndpointTelnetObj struct {
	Note string `json:"note"`
	Port string `json:"port"`
}

type EndpointTelnetTable struct {
	Id           int    `json:"id"`
	EndpointGuid string `json:"endpoint_guid"`
	Port         string `json:"port"`
	Note         string `json:"note"`
}

type PingExporterSourceDto struct {
	Config []*PingExportSourceObj `json:"config"`
}

type PingExportSourceObj struct {
	Ip   string `json:"ip"`
	Guid string `json:"guid"`
}

type TelnetSourceQuery struct {
	Guid string `json:"guid"`
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type DefaultSortObj struct {
	Key   string
	Value string
}

type DefaultSortList []*DefaultSortObj

func (s DefaultSortList) Len() int {
	return len(s)
}

func (s DefaultSortList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DefaultSortList) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

type AgentManagerTable struct {
	EndpointGuid    string `json:"endpoint_guid"`
	Name            string `json:"name"`
	User            string `json:"user"`
	Password        string `json:"password"`
	InstanceAddress string `json:"instance_address"`
	AgentAddress    string `json:"agent_address"`
	ConfigFile      string `json:"config_file"`
	BinPath         string `json:"bin_path"`
	AgentRemotePort string `json:"agent_remote_port"`
}

type InitDeployParam struct {
	AgentManagerRemoteIp string               `json:"agentManagerRemoteIp"`
	Config               []*AgentManagerTable `json:"config"`
}

type QueryPrometheusMetricParam struct {
	Ip            string   `json:"ip"`
	Port          string   `json:"port"`
	Cluster       string   `json:"cluster"`
	Prefix        []string `json:"prefix"`
	Keyword       []string `json:"keyword"`
	TargetGuid    string   `json:"target_guid"`
	EndpointGuid  string   `json:"endpoint_guid"`
	IsConfigQuery bool     `json:"is_config_query"`
	ServiceGroup  string   `json:"service_group"`
}
