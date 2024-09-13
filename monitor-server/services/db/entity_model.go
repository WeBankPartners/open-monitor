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
