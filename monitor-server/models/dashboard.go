package models

type SearchModel struct {
	Id  int  `json:"id"`
	Enable  bool  `json:"enable"`
	Name    string  `json:"name"`
	SearchUrl  string  `json:"search_url"`
	SearchCol  string  `json:"search_col"`
	RefreshPanels  bool  `json:"refresh_panels"`
	RefreshMessage bool  `json:"refresh_message"`
}

type OptionModel struct {
	OptionValue  string  `json:"option_value"`
	OptionText   string  `json:"option_text"`
	Active  bool  `json:"active"`
}

type ButtonModel struct {
	Id  int  `json:"id"`
	GroupId  int  `json:"group_id"`
	Name  string  `json:"name"`
	BType  string  `json:"b_type"`
	BText  string  `json:"b_text"`
	RefreshPanels  bool  `json:"refresh_panels"`
	RefreshCharts  bool  `json:"refresh_charts"`
	OptionGroup  int  `json:"option_group"`
	RefreshButton  int  `json:"refresh_button"`
	RefreshButtonUrl  string  `json:"refresh_button_url"`
	Options  []*OptionModel   `json:"option"`
}

type MessageModel struct {
	Enable  bool  `json:"enable"`
	Url   string  `json:"url"`
}

type PanelsModel struct {
	Enable  bool  `json:"enable"`
	Type  string  `json:"type"`
	Url   string  `json:"url"`
}

type Dashboard struct {
	Search  SearchModel  `json:"search"`
	Buttons  []*ButtonModel  `json:"buttons"`
	Message  MessageModel  `json:"message"`
	Panels  PanelsModel  `json:"panels"`
}

type ChartModel struct {
	Id  int  `json:"id"`
	Col  int  `json:"col"`
	Endpoint  []string  `json:"endpoint"`
	Metric  []string  `json:"metric"`
	Url  string  `json:"url"`
	Aggregate  string  `json:"aggregate"`
}

type PanelModel struct {
	Title  string  `json:"title"`
	Tags  TagsModel  `json:"tags"`
	Other  bool  `json:"other"`
	Charts  []*ChartModel  `json:"charts"`
}

type TagsModel struct {
	Enable  bool  `json:"enable"`
	Url  string  `json:"url"`
	Option  []*OptionModel  `json:"option"`
}

type PanelTag struct {
	Col  int   `json:"col"`
	Endpoint  []string  `json:"endpoint"`
	Metric  []string  `json:"metric"`
	Url  string  `json:"url"`
}

type YaxisModel struct {
	Unit  string  `json:"unit"`
}

type SerialModel struct {
	Type  string  `json:"type"`
	Name  string  `json:"name"`
	Data  [][]float64  `json:"data"`
}

type EChartOption struct {
	Id     int     `json:"id"`
	Title  string  `json:"title"`
	Legend  []string  `json:"legend"`
	Xaxis  interface{}  `json:"xaxis"`
	Yaxis  YaxisModel  `json:"yaxis"`
	Series []*SerialModel  `json:"series"`
}

type Chart struct {
	Endpoint  []string  `json:"endpoint"`
	Metric  []string  `json:"metric"`
	Option  EChartOption  `json:"option"`
}

// DB Map struct

type DashboardTable struct {
	Id  int  `json:"id"`
	DashboardType  string  `json:"dashboard_type"`
	SearchEnable  bool  `json:"search_enable"`
	SearchId  int  `json:"search_id"`
	ButtonEnable  bool  `json:"button_enable"`
	ButtonGroup  int  `json:"button_group"`
	MessageEnable  bool  `json:"message_enable"`
	MessageGroup  int  `json:"message_group"`
	MessageUrl  string  `json:"message_url"`
	PanelsEnable  bool  `json:"panels_enable"`
	PanelsType  string  `json:"panels_type"`
	PanelsGroup  int  `json:"panels_group"`
	PanelsParam  string  `json:"panels_param"`
}

type MessageTable struct {
	Id  int  `json:"id"`
	GroupId  int  `json:"group_id"`
	K    string  `json:"k"`
	Rename  string  `json:"rename"`
	Col    string  `json:"col"`
	Href  bool  `json:"href"`
	Url    string  `json:"url"`
}

type PanelTable struct {
	Id  int  `json:"id"`
	GroupId  int  `json:"group_id"`
	Title  string  `json:"title"`
	TagsEnable  bool  `json:"tags_enable"`
	TagsUrl  string  `json:"tags_url"`
	TagsKey  string  `json:"tags_key"`
	ChartGroup  int  `json:"chart_group"`
	ExIsPhy  bool
}

type ChartTable struct {
	Id  int  `json:"id"`
	GroupId  int  `json:"group_id"`
	Endpoint  string  `json:"endpoint"`
	Metric  string  `json:"metric"`
	Col  int  `json:"col"`
	Url  string  `json:"url"`
	Unit  string  `json:"unit"`
	Title  string  `json:"title"`
	GridType  string  `json:"grid_type"`
	SeriesName  string  `json:"series_name"`
	Rate   bool   `json:"rate"`
	AggType  string  `json:"agg_type"`
	Legend   string  `json:"legend"`
}

// POST DATA
type GetChartDto struct {
	Id   int    `form:"id" json:"id"`
	Endpoint   []string    `form:"endpoint" json:"endpoint"`
	Ip  []string  `form:"ip" json:"ip"`
	Metric   []string    `form:"metric" json:"metric"`
	Start  string  `form:"start" json:"start"`
	End  string  `form:"end" json:"end"`
	Time  string  `form:"time" json:"time"`
	Aggregate  string  `form:"aggregate" json:"aggregate"`
	Extend  bool  `form:"extend" json:"extend"`
	Rate  bool  `form:"rate" json:"rate"`
}

type PromMetricTable struct {
	Id  int  `json:"id"`
	Metric  string  `json:"metric"`
	MetricType  string  `json:"metric_type"`
	PromQl  string  `json:"prom_ql"`
	PromMain  string  `json:"prom_main"`
}

type EndpointTable struct {
	Id  int  `json:"id"`
	Guid  string  `json:"guid"`
	Name  string  `json:"name"`
	Ip  string  `json:"ip"`
	EndpointVersion  string  `json:"endpoint_version"`
	ExportType  string  `json:"export_type"`
	ExportVersion  string  `json:"export_version"`
	Step  int  `json:"step"`
	OsIp  string  `json:"os_ip"`
	OsType  string  `json:"os_type"`
	CreateAt  string  `json:"create_at"`
}

type EndpointMetricTable struct {
	Id  int  `json:"id"`
	EndpointId  int  `json:"endpoint_id"`
	Metric  string  `json:"metric"`
}