package db

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func GetEndpointTypeList() (result []string,err error) {
	result = []string{}
	queryRows,queryErr := x.QueryString("select distinct export_type from endpoint")
	if queryErr != nil {
		err = queryErr
		return
	}
	for _,row := range queryRows {
		result = append(result, row["export_type"])
	}
	return
}

func GetEndpointByType(endpointType string) (result []*models.EndpointTable,err error) {
	result = []*models.EndpointTable{}
	err = x.SQL("select id,guid from endpoint where export_type=?", endpointType).Find(&result)
	return
}
