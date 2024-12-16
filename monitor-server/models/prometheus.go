package models

import "encoding/json"

type RuleFile struct {
	Groups []*RFGroup `yaml:"groups"`
}

// RF -> RuleFile
type RFGroup struct {
	Name  string    `yaml:"name"`
	Rules []*RFRule `yaml:"rules"`
}

type RuleLocalConfigJob struct {
	EndpointGroup       string
	TplId               int
	FromPeer            bool
	Name                string
	Rules               []*RFRule
	WithoutReloadConfig bool
}

type RFRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for"`
	Labels      map[string]string `yaml:"labels"`
	Annotations RFAnnotation      `yaml:"annotations"`
}

type RFAnnotation struct {
	Summary     string `yaml:"summary"`
	Description string `yaml:"description"`
}

type ConsulServicesDto struct {
	Name string `json:"name"`
}

type FileSdConfig []*FileSdObj

type ServiceDiscoverFileList []*ServiceDiscoverFileObj

type ServiceDiscoverFileObj struct {
	Guid    string `json:"guid"`
	Address string `json:"address"`
	Step    int    `json:"step"`
	Cluster string `json:"cluster"`
}

func (s ServiceDiscoverFileList) TurnToFileSdConfigByte(step int) []byte {
	var result []*FileSdObj
	for _, v := range s {
		if v.Step == step {
			result = append(result, &FileSdObj{Targets: []string{v.Address}, Labels: FileSdLabel{EGuid: v.Guid}})
		}
	}
	b, _ := json.Marshal(result)
	return b
}

type FileSdObj struct {
	Targets []string    `json:"targets"`
	Labels  FileSdLabel `json:"labels"`
}

type FileSdLabel struct {
	EGuid string `json:"e_guid"`
}
