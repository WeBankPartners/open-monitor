package models

type BusinessMonitorTable struct {
	Id            int    `json:"id"`
	EndpointId    int    `json:"endpoint_id"`
	Path          string `json:"path"`
	OwnerEndpoint string `json:"owner_endpoint"`
}

type BusinessUpdatePathObj struct {
	Id            int                         `json:"id"`
	Path          string                      `json:"path"`
	OwnerEndpoint string                      `json:"owner_endpoint"`
	Rules         []*BusinessMonitorCfgObj    `json:"rules"`
	CustomMetrics []*BusinessMonitorCustomObj `json:"custom_metrics"`
}

type BusinessUpdateDto struct {
	EndpointId int                      `json:"endpoint_id" binding:"required"`
	PathList   []*BusinessUpdatePathObj `json:"path_list"`
}

type BusinessMonitorCfgTable struct {
	Id                int    `json:"id"`
	BusinessMonitorId int    `json:"business_monitor_id"`
	Regular           string `json:"regular"`
	Tags              string `json:"tags"`
	StringMap         string `json:"string_map"`
	MetricConfig      string `json:"metric_config"`
	AggType           string `json:"agg_type"`
	ConfigType        string `json:"config_type"`
}

type BusinessStringMapObj struct {
	Key         string  `json:"key"`
	Regulation  string  `json:"regulation"`
	StringValue string  `json:"string_value"`
	IntValue    float64 `json:"int_value"`
}

type BusinessMetricObj struct {
	Key     string `json:"key"`
	Metric  string `json:"metric"`
	Title   string `json:"title"`
	AggType string `json:"agg_type"`
}

type BusinessMonitorCfgObj struct {
	Id           int                     `json:"id"`
	Regular      string                  `json:"regular"`
	Tags         string                  `json:"tags"`
	StringMap    []*BusinessStringMapObj `json:"string_map"`
	MetricConfig []*BusinessMetricObj    `json:"metric_config"`
}

type BusinessMonitorCustomObj struct {
	Id           int                     `json:"id"`
	Metric       string                  `json:"metric"`
	ValueRegular string                  `json:"value_regular"`
	AggType      string                  `json:"agg_type"`
	StringMap    []*BusinessStringMapObj `json:"string_map"`
}

type BusinessAgentDto struct {
	Path   string                   `json:"path"`
	Config []*BusinessMonitorCfgObj `json:"config"`
	Custom []*BusinessMonitorCustomObj `json:"custom"`
}

type PluginBusinessRequest struct {
	RequestId string                           `json:"requestId"`
	Inputs    []*PluginBusinessValueRequestObj `json:"inputs"`
}

type PluginBusinessValueRequestObj struct {
	CallbackParameter string                     `json:"callbackParameter"`
	HostIp            string                     `json:"hostIp"`
	RefMonitorObj     string                     `json:"refMonitorObj"`
	PathPrefix        string                     `json:"pathPrefix"`
	Config            []*PluginBusinessConfigObj `json:"config"`
}

type PluginBusinessResp struct {
	ResultCode    string               `json:"resultCode"`
	ResultMessage string               `json:"resultMessage"`
	Results       PluginBusinessOutput `json:"results"`
}

type PluginBusinessOutput struct {
	Outputs []*PluginBusinessOutputObj `json:"outputs"`
}

type PluginBusinessOutputObj struct {
	CallbackParameter string `json:"callbackParameter"`
	HostIp            string `json:"hostIp"`
	LogPath           string `json:"logPath"`
	ErrorCode         string `json:"errorCode"`
	ErrorMessage      string `json:"errorMessage"`
	ErrorDetail       string `json:"errorDetail,omitempty"`
}

type PluginBusinessConfigObj struct {
	Path  string                   `json:"path"`
	Rules []*PluginBusinessRuleObj `json:"rules"`
}

type PluginBusinessRuleObj struct {
	Regular      string                     `json:"regular"`
	Tags         string                     `json:"tags"`
	MetricConfig []*PluginBusinessMetricObj `json:"metricConfig"`
}

type PluginBusinessMetricObj struct {
	Key      string                       `json:"key"`
	Metric   string                       `json:"metric"`
	Title    string                       `json:"title"`
	AggType  string                       `json:"aggType"`
	ValueMap []*PluginBusinessValueMapObj `json:"valueMap"`
}

type PluginBusinessValueMapObj struct {
	IsRegular   string  `json:"isRegular"`
	StringValue string  `json:"stringValue"`
	IntValue    float64 `json:"intValue"`
}
