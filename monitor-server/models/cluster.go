package models

type ClusterTable struct {
	Id                 string `json:"id" binding:"required"`
	DisplayName        string `json:"display_name" binding:"required"`
	RemoteAgentAddress string `json:"remote_agent_address" binding:"required"`
	PromAddress        string `json:"prom_address" binding:"required"`
}

type SdLocalConfigJob struct {
	FromPeer bool
	Configs  []*SdConfigSyncObj
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
