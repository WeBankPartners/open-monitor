package models

type ClusterTableNew struct {
	Guid               string `json:"guid" xorm:"guid" binding:"required"`
	DisplayName        string `json:"display_name" xorm:"display_name" binding:"required"`
	RemoteAgentAddress string `json:"remote_agent_address" xorm:"remote_agent_address" binding:"required"`
	PrometheusAddress  string `json:"prometheus_address" xorm:"prometheus_address" binding:"required"`
}

type MonitorTypeTable struct {
	Guid        string `json:"guid" xorm:"guid" binding:"required"`
	DisplayName string `json:"display_name" xorm:"display_name"`
	Description string `json:"description" xorm:"description"`
}

type EndpointNew struct {
	Guid            string `json:"guid" xorm:"guid"`
	Name            string `json:"name" xorm:"name"`
	Ip              string `json:"ip" xorm:"ip"`
	MonitorType     string `json:"monitor_type" xorm:"monitor_type"`
	AgentVersion    string `json:"agent_version" xorm:"agent_version"`
	AgentAddress    string `json:"agent_address" xorm:"agent_address"`
	Step            int    `json:"step" xorm:"step"`
	EndpointVersion string `json:"endpoint_version" xorm:"endpoint_version"`
	EndpointAddress string `json:"endpoint_address" xorm:"endpoint_address"`
	Cluster         string `json:"cluster" xorm:"cluster"`
	AlarmEnable     int    `json:"alarm_enable" xorm:"alarm_enable"`
	Tags            string `json:"tags" xorm:"tags"`
	ExtendParam     string `json:"extend_param" xorm:"extend_param"`
	Description     string `json:"description" xorm:"description"`
	UpdateTime      string `json:"update_time" xorm:"update_time"`
}

type EndpointExtendParamObj struct {
	Enable bool `json:"-"`
	Ip  string `json:"ip,omitempty"`
	Port string `json:"port,omitempty"`
	User string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	BinPath string `json:"bin_path,omitempty"`
	ConfigPath string `json:"config_path,omitempty"`
	HttpMethod string `json:"http_method,omitempty"`
	HttpUrl string `json:"http_url,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
	ProcessTags string `json:"process_tags,omitempty"`
}