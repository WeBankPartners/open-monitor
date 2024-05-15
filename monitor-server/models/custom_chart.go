package models

type CustomChart struct {
	Guid            string `json:"id" xorm:"'guid' pk"`
	SourceDashboard string `json:"sourceDashboard" xorm:"source_dashboard"` // 源看板
	Public          int    `json:"public" xorm:"public"`                    // 是否公共
	Name            string `json:"name" xorm:"name"`                        // 图表名称
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
	SourceDashboard string                  `json:"sourceDashboard"` // 源看板
	Name            string                  `json:"name"`            // 图表名称
	Unit            string                  `json:"unit"`            // 单位
	ChartType       string                  `json:"chartType"`       // 曲线图/饼图,line/pie
	LineType        string                  `json:"lineType"`        // 折线/柱状/面积,line/bar/area
	Aggregate       string                  `json:"aggregate"`       // 聚合类型
	AggStep         int                     `json:"aggStep"`         // 聚合间隔
	Query           []*CustomChartSeriesDto `json:"query"`
	DisplayConfig   interface{}             `json:"displayConfig"`
	Group           string                  `json:"group"` // 所属分组
}

type ChartSharedDto struct {
	Id              string `json:"id"`
	SourceDashboard string `json:"sourceDashboard"` // 源看板
	Name            string `json:"name"`            // 图表名称
}
