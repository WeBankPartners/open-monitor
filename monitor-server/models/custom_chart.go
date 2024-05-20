package models

type CustomChart struct {
	Guid            string `json:"id" xorm:"'guid' pk"`
	SourceDashboard int    `json:"sourceDashboard" xorm:"source_dashboard"` // 源看板
	Public          int    `json:"public" xorm:"public"`                    // 是否公共
	Name            string `json:"name" xorm:"name"`                        // 图表名称
	ChartTemplate   string `json:"chartTemplate" xorm:"chart_template"`     // 图表模板
	ChartType       string `json:"chartType" xorm:"chart_type"`             // 曲线图/饼图,line//pie
	LineType        string `json:"lineType" xorm:"line_type"`               // 折线/柱状/面积,line/bar/area
	Aggregate       string `json:"aggregate" xorm:"aggregate"`              // 聚合类型
	AggStep         int    `json:"aggStep" xorm:"agg_step"`                 // 聚合间隔
	Unit            string `json:"unit" xorm:"unit"`                        // 单位
	CreateUser      string `json:"createUser" xorm:"create_user"`           // 创建人
	UpdateUser      string `json:"updateUser" xorm:"update_user"`           // 更新人
	CreateTime      string `json:"createTime" xorm:"create_time"`           // 创建时间
	UpdateTime      string `json:"updateTime" xorm:"update_time"`           // 更新时间
}

type CustomChartExtend struct {
	CustomChart   *CustomChart
	Group         string `json:"group" xorm:"group"`                  // 所属分组
	DisplayConfig string `json:"displayConfig" xorm:"display_config"` // 视图位置与长宽
}

type CustomChartDto struct {
	Id              string                  `json:"id"`
	Public          bool                    `json:"public"`
	SourceDashboard int                     `json:"sourceDashboard"` // 源看板
	Name            string                  `json:"name"`            // 图表名称
	ChartTemplate   string                  `json:"chartTemplate"`   // 图表模板
	Unit            string                  `json:"unit"`            // 单位
	ChartType       string                  `json:"chartType"`       // 曲线图/饼图,line/pie
	LineType        string                  `json:"lineType"`        // 折线/柱状/面积,line/bar/area
	Aggregate       string                  `json:"aggregate"`       // 聚合类型
	AggStep         int                     `json:"aggStep"`         // 聚合间隔
	ChartSeries     []*CustomChartSeriesDto `json:"chartSeries"`
	DisplayConfig   interface{}             `json:"displayConfig"`
	Group           string                  `json:"group"` // 所属分组
}

type ChartSharedDto struct {
	Id              string `json:"id"`
	SourceDashboard int    `json:"sourceDashboard"` // 源看板
	Name            string `json:"name"`            // 图表名称
}

type ChartSharedParam struct {
	ChartId   string   `json:"chartId"`
	UseRoles  []string `json:"useRoles"`
	MgmtRoles []string `json:"mgmtRoles"`
}

type QueryChartParam struct {
	ChartId          string   `json:"chartId"`
	ChartName        string   `json:"chartName"`
	ChartType        string   `json:"chartType"`
	SourceDashboard  int      `json:"sourceDashboard"`  // 源看板
	UseDashboard     []string `json:"useDashboard"`     // 应用看板
	MgmtRoles        []string `json:"mgmtRoles"`        // 管理角色
	UseRoles         []string `json:"useRoles"`         // 使用角色
	UpdateUser       string   `json:"updateUser"`       // 更新人
	UpdatedTimeStart string   `json:"updatedTimeStart"` // 更新时间开始
	UpdatedTimeEnd   string   `json:"updatedTimeEnd"`   // 更新时间结束
	Permission       string   `json:"permission"`       // 等于 MGMT表示可编辑
	StartIndex       int      `json:"startIndex"`
	PageSize         int      `json:"pageSize"`
}

type QueryChartResultDto struct {
	ChartId         string   `json:"chartId"`
	ChartName       string   `json:"chartName"`
	ChartType       string   `json:"chartType"`
	SourceDashboard string   `json:"sourceDashboard"` // 源看板名称
	UseDashboard    []string `json:"useDashboard"`    // 应用看板
	MgmtRoles       []string `json:"mgmtRoles"`       // 管理角色
	UseRoles        []string `json:"useRoles"`        // 使用角色
	UpdateUser      string   `json:"updateUser"`      // 更新人
	UpdatedTime     string   `json:"updatedTime"`     // 更新时间
	Permission      string   `json:"permission"`      // MGMT表示可编辑,USE可使用
}

type AddCustomChartParam struct {
	DashboardId   int         `json:"dashboardId"`   // 源看板
	Name          string      `json:"name"`          // 图表名称
	ChartTemplate string      `json:"chartTemplate"` // 图表模板
	ChartType     string      `json:"chartType"`     // 曲线图/饼图,line/pie
	LineType      string      `json:"lineType"`      // 折线/柱状/面积,line/bar/area
	Aggregate     string      `json:"aggregate"`     // 聚合类型
	AggStep       int         `json:"aggStep"`       // 聚合间隔
	Unit          string      `json:"unit"`          // 单位
	Group         string      `json:"group"`         // 所属分组
	DisplayConfig interface{} `json:"displayConfig"` // 视图位置与长宽
}

type CopyCustomChartParam struct {
	DashboardId   int         `json:"dashboardId"`   // 源看板
	Ref           bool        `json:"ref"`           // 是否引用
	OriginChartId string      `json:"originChartId"` // 原图表ID
	Group         string      `json:"group"`         // 所属分组
	DisplayConfig interface{} `json:"displayConfig"` // 视图位置与长宽
}
