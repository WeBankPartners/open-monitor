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
	Guid         string `json:"guid" xorm:"guid"`
	Endpoint     string `json:"endpoint" xorm:"endpoint"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
}

type ServiceGroupLinkNode struct {
	Guid     string                  `json:"guid"`
	Parent   *ServiceGroupLinkNode   `json:"parent"`
	Children []*ServiceGroupLinkNode `json:"children"`
}

func (s *ServiceGroupLinkNode) FetchChildGuid() []string {
	result := []string{s.Guid}
	for _, v := range s.Children {
		result = append(result, v.FetchChildGuid()...)
	}
	return result
}

func (s *ServiceGroupLinkNode) FetchParentGuid() []string {
	result := []string{s.Guid}
	if s.Parent != nil {
		result = append(result, s.Parent.FetchParentGuid()...)
	}
	return result
}

type ServiceGroupEndpointListObj struct {
	Guid        string `json:"guid"`
	DisplayName string `json:"display_name"`
}

type ServiceGroupRoleRelTable struct {
	Guid         string `json:"guid" xorm:"guid"`
	ServiceGroup string `json:"service_group" xorm:"service_group"`
	Role         string `json:"role" xorm:"role"`
}

type PluginUpdateServicePathRequest struct {
	RequestId      string                        `json:"requestId"`
	DueDate        string                        `json:"dueDate"`
	AllowedOptions []string                      `json:"allowedOptions"`
	Inputs         []*PluginUpdateServicePathRequestObj `json:"inputs"`
}

type PluginUpdateServicePathRequestObj struct {
	CallbackParameter string `json:"callbackParameter"`
	ProcInstId        string `json:"procInstId"`
	CallbackUrl       string `json:"callbackUrl"`
	Reporter          string `json:"reporter"`
	Handler           string `json:"handler"`
	RoleName          string `json:"roleName"`
	TaskName          string `json:"taskName"`
	TaskDescription   string `json:"taskDescription"`
	TaskFormInput     string `json:"taskFormInput"`
}

type PluginUpdateServicePathResp struct {
	ResultCode    string                 `json:"resultCode"`
	ResultMessage string                 `json:"resultMessage"`
	Results       PluginUpdateServicePathOutput `json:"results"`
}

type PluginUpdateServicePathOutput struct {
	RequestId      string                       `json:"requestId"`
	AllowedOptions []string                     `json:"allowedOptions,omitempty"`
	Outputs        []*PluginUpdateServicePathOutputObj `json:"outputs"`
}

type PluginUpdateServicePathOutputObj struct {
	CallbackParameter string `json:"callbackParameter"`
	Comment           string `json:"comment"`
	TaskFormOutput    string `json:"taskFormOutput"`
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
	ErrorDetail       string `json:"errorDetail,omitempty"`
}