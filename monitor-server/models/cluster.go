package models

type ClusterTable struct {
	Id                 string `json:"id"`
	DisplayName        string `json:"display_name"`
	RemoteAgentAddress string `json:"remote_agent_address"`
	PromAddress        string `json:"prom_address"`
}

type SdLocalConfigJob struct {
	FromPeer  bool
	Configs   []*SdConfigSyncObj
}

type SdConfigSyncObj struct {
	Step    int    `json:"step"`
	Content string `json:"content"`
}

type ClusterAgentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RFClusterRequestObj struct {
	Name  string    `json:"name"`
	Rules []*RFRule `json:"rules"`
}
