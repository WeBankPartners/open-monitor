package models

type OrganizationPanel struct {
	Guid            string               `json:"guid"`
	DisplayName     string               `json:"display_name"`
	Type            string               `json:"type"`
	FetchSearch     bool                 `json:"fetch_search"`
	FetchOriginFlag bool                 `json:"fetch_origin_flag"`
	Children        []*OrganizationPanel `json:"children"`
}

type UpdateOrgPanelParam struct {
	Guid        string `json:"guid"`
	DisplayName string `json:"display_name"`
	Parent      string `json:"parent"`
	Type        string `json:"type"`
	Force       string `json:"force"`
}

type UpdateOrgPanelRoleParam struct {
	Guid   string `json:"guid" binding:"required"`
	RoleId []int  `json:"role_id"`
}

type UpdateOrgPanelEndpointParam struct {
	Guid     string   `json:"guid" binding:"required"`
	Endpoint []string `json:"endpoint"`
}

type UpdateOrgConnectParam struct {
	Guid  string   `json:"guid" binding:"required"`
	Mail  []string `json:"mail"`
	Phone []string `json:"phone"`
}

type GetOrgPanelCallbackData struct {
	FiringCallback  []*OptionModel `json:"firing_callback"`
	RecoverCallback []*OptionModel `json:"recover_callback"`
}

type UpdateOrgPanelEventParam struct {
	Guid                string `json:"guid" binding:"required"`
	FiringCallbackName  string `json:"firing_callback_name"`
	FiringCallbackKey   string `json:"firing_callback_key"`
	RecoverCallbackName string `json:"recover_callback_name"`
	RecoverCallbackKey  string `json:"recover_callback_key"`
}

type IsPluginModeResult struct {
	IsPlugin bool `json:"is_plugin"`
}

type EntityQueryParam struct {
	Criteria          EntityQueryObj    `json:"criteria"`
	AdditionalFilters []*EntityQueryObj `json:"additionalFilters"`
}

type EntityQueryObj struct {
	AttrName  string `json:"attrName"`
	Op        string `json:"op"`
	Condition string `json:"condition"`
}

func (ep *EntityQueryParam) TransToQueryParam() (queryParam *QueryRequestParam) {
	queryParam = &QueryRequestParam{}
	if ep.Criteria.AttrName != "" {
		if ep.Criteria.Op == "" {
			ep.Criteria.Op = "eq"
		}
		queryParam.Filters = append(queryParam.Filters, &QueryRequestFilterObj{Name: ep.Criteria.AttrName, Operator: ep.Criteria.Op, Value: ep.Criteria.Condition})
	}
	for _, filter := range ep.AdditionalFilters {
		if filter.Op == "" {
			filter.Op = "eq"
		}
		queryParam.Filters = append(queryParam.Filters, &QueryRequestFilterObj{Name: filter.AttrName, Operator: filter.Op, Value: filter.Condition})
	}
	return
}
