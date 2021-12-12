package models

type ServiceGroupTable struct {
	Guid        string `json:"guid" xorm:"guid"`
	DisplayName string `json:"display_name" xorm:"display_name"`
	Description string `json:"description" xorm:"description"`
	Parent      string `json:"parent" xorm:"parent"`
	ServiceType string `json:"service_type" xorm:"service_type"`
	UpdateTime  string `json:"update_time" xorm:"update_time"`
}

type EndpointServiceRelTable struct {
	Guid string `json:"guid" xorm:"guid"`
	Endpoint string `json:"endpoint" xorm:"endpoint"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
}

type ServiceGroupLinkNode struct {
	Guid string `json:"guid"`
	Parent *ServiceGroupLinkNode `json:"parent"`
	Children []*ServiceGroupLinkNode `json:"children"`
}

func (s ServiceGroupLinkNode) FetchChildGuid() []string {
	result := []string{s.Guid}
	for _,v := range s.Children {
		result = append(result, v.FetchChildGuid()...)
	}
	return result
}

type ServiceGroupEndpointListObj struct {
	Guid string `json:"guid"`
	DisplayName string `json:"display_name"`
}