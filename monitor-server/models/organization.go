package models

type OrganizationPanel struct {
	Guid  string  `json:"guid"`
	DisplayName  string  `json:"display_name"`
	Type  string  `json:"type"`
	Children  []*OrganizationPanel  `json:"children"`
}

type UpdateOrgPanelParam struct {
	Guid  string  `json:"guid"`
	DisplayName  string  `json:"display_name"`
	Parent  string  `json:"parent"`
	Type  string  `json:"type"`
}

type UpdateOrgPanelRoleParam struct {
	Guid  string  `json:"guid" binding:"required"`
	RoleId  []int  `json:"role_id"`
}

type UpdateOrgPanelEndpointParam struct {
	Guid  string  `json:"guid" binding:"required"`
	Endpoint  []string  `json:"endpoint"`
}

type UpdateOrgConnectParam struct {
	Guid  string  `json:"guid" binding:"required"`
	Mail  []string  `json:"mail"`
	Phone  []string  `json:"phone"`
}

type GetOrgPanelCallbackData struct {
	FiringCallback  []*OptionModel  `json:"firing_callback"`
	RecoverCallback  []*OptionModel  `json:"recover_callback"`
}

type UpdateOrgPanelEventParam struct {
	Guid  string  `json:"guid" binding:"required"`
	FiringCallbackName  string  `json:"firing_callback_name"`
	FiringCallbackKey  string  `json:"firing_callback_key"`
	RecoverCallbackName  string  `json:"recover_callback_name"`
	RecoverCallbackKey  string  `json:"recover_callback_key"`
}

type IsPluginModeResult struct {
	IsPlugin  bool  `json:"is_plugin"`
}

type EntityQueryParam struct {
	Criteria  EntityQueryObj  `json:"criteria"`
	AdditionalFilters  []*EntityQueryObj  `json:"additionalFilters"`
}

type EntityQueryObj struct {
	AttrName  string  `json:"attrName"`
	Op  string  `json:"op"`
	Condition  string  `json:"condition"`
}