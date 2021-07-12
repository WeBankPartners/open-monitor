package models

type ClusterTable struct {
	Id                 string `json:"id"`
	DisplayName        string `json:"display_name"`
	RemoteAgentAddress string `json:"remote_agent_address"`
	PromAddress        string `json:"prom_address"`
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
	IsGrp string    `json:"is_grp"`
	Rules []*RFRule `json:"rules"`
}
