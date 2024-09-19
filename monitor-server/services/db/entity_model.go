package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func QueryEntityEndpoint(param *models.EntityQueryParam) (result []*models.EndpointEntityObj, err error) {
	queryParam := param.TransToQueryParam()
	filterSql, _, filterParam := transFiltersToSQL(queryParam, &models.TransFiltersParam{IsStruct: true, StructObj: models.EndpointEntityObj{}, PrimaryKey: "guid"})
	result = []*models.EndpointEntityObj{}
	err = x.SQL("SELECT * FROM endpoint_new WHERE 1=1 "+filterSql, filterParam...).Find(&result)
	if err != nil {
		err = fmt.Errorf("query endpoint new table fail,%s ", err.Error())
	}
	for _, row := range result {
		row.DisplayName = row.Id
	}
	return
}

func QueryEntityServiceGroup(param *models.EntityQueryParam) (result []*models.ServiceGroupEntityObj, err error) {
	queryParam := param.TransToQueryParam()
	filterSql, _, filterParam := transFiltersToSQL(queryParam, &models.TransFiltersParam{IsStruct: true, StructObj: models.ServiceGroupEntityObj{}, PrimaryKey: "guid"})
	result = []*models.ServiceGroupEntityObj{}
	err = x.SQL("SELECT * FROM service_group WHERE 1=1 "+filterSql, filterParam...).Find(&result)
	if err != nil {
		err = fmt.Errorf("query endpoint group table fail,%s ", err.Error())
	}
	return
}

func QueryEntityEndpointGroup(param *models.EntityQueryParam) (result []*models.EndpointGroupEntityObj, err error) {
	queryParam := param.TransToQueryParam()
	filterSql, _, filterParam := transFiltersToSQL(queryParam, &models.TransFiltersParam{IsStruct: true, StructObj: models.EndpointGroupEntityObj{}, PrimaryKey: "guid"})
	result = []*models.EndpointGroupEntityObj{}
	err = x.SQL("SELECT * FROM endpoint_group WHERE 1=1 "+filterSql, filterParam...).Find(&result)
	if err != nil {
		err = fmt.Errorf("query service group table fail,%s ", err.Error())
	}
	return
}

func QueryEntityMonitorType(param *models.EntityQueryParam) (result []*models.MonitorTypeEntityObj, err error) {
	queryParam := param.TransToQueryParam()
	filterSql, _, filterParam := transFiltersToSQL(queryParam, &models.TransFiltersParam{IsStruct: true, StructObj: models.MonitorTypeEntityObj{}, PrimaryKey: "guid"})
	result = []*models.MonitorTypeEntityObj{}
	err = x.SQL("SELECT * FROM monitor_type WHERE 1=1 "+filterSql, filterParam...).Find(&result)
	if err != nil {
		err = fmt.Errorf("query monitor type table fail,%s ", err.Error())
	}
	return
}

func QueryEntityLogMonitorTemplate(param *models.EntityQueryParam) (result []*models.LogMonitorTemplateEntityObj, err error) {
	queryParam := param.TransToQueryParam()
	filterSql, _, filterParam := transFiltersToSQL(queryParam, &models.TransFiltersParam{IsStruct: true, StructObj: models.LogMonitorTemplateEntityObj{}, PrimaryKey: "guid"})
	result = []*models.LogMonitorTemplateEntityObj{}
	err = x.SQL("SELECT * FROM log_monitor_template WHERE 1=1 "+filterSql, filterParam...).Find(&result)
	if err != nil {
		err = fmt.Errorf("query log monitor template table fail,%s ", err.Error())
	}
	return
}

func AnalyzeTransExportData(param *models.AnalyzeTransParam) (result *models.AnalyzeTransData, err error) {
	result = &models.AnalyzeTransData{MonitorType: []string{}, EndpointGroup: []string{}, CustomMetricMonitorType: []string{}, CustomMetricEndpointGroup: []string{}, CustomMetricServiceGroup: []string{}, LogMonitorServiceGroup: []string{}, LogMonitorTemplate: []string{}, StrategyEndpointGroup: []string{}, StrategyServiceGroup: []string{}, LogKeywordServiceGroup: []string{}, DashboardIdList: []string{}}
	if len(param.EndpointList) == 0 && len(param.ServiceGroupList) == 0 {
		return
	}
	endpointFilterSql, endpointFilterParams := createListParams(param.EndpointList, "")
	serviceGroupFilterSql, serviceGroupFilterParams := createListParams(param.ServiceGroupList, "")
	var monitorTypeRows []*models.MonitorTypeTable
	err = x.SQL("select guid,display_name from monitor_type").Find(&monitorTypeRows)
	if err != nil {
		err = fmt.Errorf("query monitor type table fail,%s ", err.Error())
		return
	}
	for _, row := range monitorTypeRows {
		result.MonitorType = append(result.MonitorType, row.Guid)
	}
	customDashboardDistinctMap := make(map[int]int)
	if len(param.EndpointList) > 0 {
		var endpointGroupRows []*models.EndpointGroupTable
		err = x.SQL("select guid,display_name from endpoint_group where service_group is null and guid in (select endpoint_group from endpoint_group_rel where endpoint in ("+endpointFilterSql+"))", endpointFilterParams...).Find(&endpointGroupRows)
		if err != nil {
			err = fmt.Errorf("query endpointGroup table fail,%s ", err.Error())
			return
		}
		for _, row := range endpointGroupRows {
			result.EndpointGroup = append(result.EndpointGroup, row.Guid)
		}
		if len(result.EndpointGroup) > 0 {
			endpointGroupFilterSql, endpointGroupFilterParams := createListParams(result.EndpointGroup, "")
			var alarmStrategyRows []*models.AlarmStrategyTable
			err = x.SQL("select distinct endpoint_group from alarm_strategy where endpoint_group in ("+endpointGroupFilterSql+")", endpointGroupFilterParams...).Find(&alarmStrategyRows)
			if err != nil {
				err = fmt.Errorf("query alarm strategy table fail,%s ", err.Error())
				return
			}
			for _, row := range alarmStrategyRows {
				result.StrategyEndpointGroup = append(result.StrategyEndpointGroup, row.EndpointGroup)
			}
			var metricRows []*models.MetricTable
			err = x.SQL("select distinct endpoint_group from metric where endpoint_group in ("+endpointGroupFilterSql+") and log_metric_group is null", endpointGroupFilterParams...).Find(&metricRows)
			if err != nil {
				err = fmt.Errorf("query metric table fail,%s ", err.Error())
				return
			}
			for _, row := range metricRows {
				result.CustomMetricEndpointGroup = append(result.CustomMetricEndpointGroup, row.EndpointGroup)
			}
		}
		var customDashboardRows []*models.CustomDashboardChartRel
		err = x.SQL("select distinct custom_dashboard from custom_dashboard_chart_rel where dashboard_chart in (select dashboard_chart from custom_chart_series where endpoint in ("+endpointFilterSql+"))", endpointFilterParams...).Find(&customDashboardRows)
		if err != nil {
			err = fmt.Errorf("query custom dashboard with endpoint fail,%s ", err.Error())
			return
		}
		for _, row := range customDashboardRows {
			result.DashboardIdList = append(result.DashboardIdList, fmt.Sprintf("%d", *row.CustomDashboard))
			customDashboardDistinctMap[*row.CustomDashboard] = 1
		}
	}
	if len(param.ServiceGroupList) > 0 {
		var logMetricMonitorRows []*models.LogMetricMonitorTable
		err = x.SQL("select t1.service_group from (select service_group from log_metric_monitor union select service_group from db_metric_monitor) t1 where t1.service_group in ("+serviceGroupFilterSql+")", serviceGroupFilterParams...).Find(&logMetricMonitorRows)
		if err != nil {
			err = fmt.Errorf("query log metric monitor table fail,%s ", err.Error())
			return
		}
		for _, row := range logMetricMonitorRows {
			result.LogMonitorServiceGroup = append(result.LogMonitorServiceGroup, row.ServiceGroup)
		}
		var logMetricGroupRows []*models.LogMetricGroup
		err = x.SQL("select distinct log_monitor_template from log_metric_group where log_monitor_template<>'' and log_metric_monitor in (select guid from log_metric_monitor where service_group in ("+serviceGroupFilterSql+"))", serviceGroupFilterParams...).Find(&logMetricGroupRows)
		if err != nil {
			err = fmt.Errorf("query log metric group table fail,%s ", err.Error())
			return
		}
		for _, row := range logMetricGroupRows {
			result.LogMonitorTemplate = append(result.LogMonitorTemplate, row.LogMonitorTemplate)
		}
		var logKeywordMonitorRows []*models.LogKeywordMonitorTable
		err = x.SQL("select t1.service_group from (select service_group from log_keyword_monitor union select service_group from db_keyword_monitor) t1 where t1.service_group in ("+serviceGroupFilterSql+")", serviceGroupFilterParams...).Find(&logKeywordMonitorRows)
		if err != nil {
			err = fmt.Errorf("query log keyword monitor table fail,%s ", err.Error())
			return
		}
		for _, row := range logKeywordMonitorRows {
			result.LogKeywordServiceGroup = append(result.LogKeywordServiceGroup, row.ServiceGroup)
		}
		var endpointGroupRows []*models.EndpointGroupTable
		err = x.SQL("select distinct t2.service_group from alarm_strategy t1 left join endpoint_group t2 on t1.endpoint_group=t2.guid where t2.service_group in ("+serviceGroupFilterSql+")", serviceGroupFilterParams...).Find(&endpointGroupRows)
		if err != nil {
			err = fmt.Errorf("query alarm strategy with service group table fail,%s ", err.Error())
			return
		}
		for _, row := range endpointGroupRows {
			result.StrategyServiceGroup = append(result.StrategyServiceGroup, row.ServiceGroup)
		}
		var customDashboardRows []*models.CustomDashboardChartRel
		err = x.SQL("select distinct custom_dashboard from custom_dashboard_chart_rel where dashboard_chart in (select dashboard_chart from custom_chart_series where service_group in ("+serviceGroupFilterSql+"))", serviceGroupFilterParams...).Find(&customDashboardRows)
		if err != nil {
			err = fmt.Errorf("query custom dashboard with service group fail,%s ", err.Error())
			return
		}
		for _, row := range customDashboardRows {
			if _, existFlag := customDashboardDistinctMap[*row.CustomDashboard]; existFlag {
				continue
			}
			result.DashboardIdList = append(result.DashboardIdList, fmt.Sprintf("%d", *row.CustomDashboard))
			customDashboardDistinctMap[*row.CustomDashboard] = 1
		}
		var metricRows []*models.MetricTable
		err = x.SQL("select distinct service_group from metric where service_group in ("+serviceGroupFilterSql+") and log_metric_group is null", serviceGroupFilterParams...).Find(&metricRows)
		if err != nil {
			err = fmt.Errorf("query metric table fail,%s ", err.Error())
			return
		}
		for _, row := range metricRows {
			result.CustomMetricServiceGroup = append(result.CustomMetricServiceGroup, row.ServiceGroup)
		}
	}
	var metricRows []*models.MetricTable
	err = x.SQL("select distinct monitor_type from metric where service_group is null and endpoint_group is null").Find(&metricRows)
	if err != nil {
		err = fmt.Errorf("query metric table fail,%s ", err.Error())
		return
	}
	for _, row := range metricRows {
		result.CustomMetricMonitorType = append(result.CustomMetricMonitorType, row.MonitorType)
	}

	return
}
