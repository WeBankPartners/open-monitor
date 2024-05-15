package db

import "github.com/WeBankPartners/open-monitor/monitor-server/models"

func QueryCustomChartListByDashboard(customDashboard int) (list []*models.CustomChartExtend, err error) {
	err = x.SQL("select c.*,r.`group`,r.display_config from custom_dashboard_chart_rel  r join custom_chart  c "+
		"on r.dashboard_chart = c.guid where r.custom_dashboard = ?", customDashboard).Find(&list)
	return
}

func QueryCustomChartSeriesByChart(dashboardChart string) (list []*models.CustomChartSeries, err error) {
	err = x.SQL("select * from custom_chart_series where dashboard_chart  = ?", dashboardChart).Find(&list)
	return
}

func QueryAllChartSeriesConfig() (configMap map[string][]*models.CustomChartSeriesConfig, err error) {
	var list []*models.CustomChartSeriesConfig
	configMap = make(map[string][]*models.CustomChartSeriesConfig)
	if err = x.SQL("select * from custom_chart_series_config").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, config := range list {
			if arr, ok := configMap[*config.DashboardChartConfig]; ok {
				arr = append(arr, config)
			} else {
				configMap[*config.DashboardChartConfig] = []*models.CustomChartSeriesConfig{}
				configMap[*config.DashboardChartConfig] = append(configMap[*config.DashboardChartConfig], config)
			}
		}
	}
	return
}

func QueryAllChartSeriesTag() (tagMap map[string][]*models.CustomChartSeriesTag, err error) {
	var list []*models.CustomChartSeriesTag
	tagMap = make(map[string][]*models.CustomChartSeriesTag)
	if err = x.SQL("select * from custom_chart_series_tag").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, config := range list {
			if arr, ok := tagMap[*config.DashboardChartConfig]; ok {
				arr = append(arr, config)
			} else {
				tagMap[*config.DashboardChartConfig] = []*models.CustomChartSeriesTag{}
				tagMap[*config.DashboardChartConfig] = append(tagMap[*config.DashboardChartConfig], config)
			}
		}
	}
	return
}

func QueryAllChartSeriesTagValue() (tagValueMap map[string][]*models.CustomChartSeriesTagValue, err error) {
	var list []*models.CustomChartSeriesTagValue
	tagValueMap = make(map[string][]*models.CustomChartSeriesTagValue)
	if err = x.SQL("select * from custom_chart_series_tagvalue").Find(&list); err != nil {
		return
	}
	if len(list) > 0 {
		for _, tagValue := range list {
			if arr, ok := tagValueMap[*tagValue.DashboardChartTag]; ok {
				arr = append(arr, tagValue)
			} else {
				tagValueMap[*tagValue.DashboardChartTag] = []*models.CustomChartSeriesTagValue{}
				tagValueMap[*tagValue.DashboardChartTag] = append(tagValueMap[*tagValue.DashboardChartTag], tagValue)
			}
		}
	}
	return
}
