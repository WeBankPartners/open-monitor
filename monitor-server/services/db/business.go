package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
)

func GetBusinessList(endpointId int) (err error, pathList []*m.BusinessMonitorTable) {
	err = x.SQL("SELECT * FROM business_monitor WHERE endpoint_id=?", endpointId).Find(&pathList)
	return err,pathList
}

func UpdateBusiness(param m.BusinessUpdateDto) error {
	var sqls []string
	sqls = append(sqls, fmt.Sprintf("DELETE FROM business_monitor WHERE endpoint_id=%d", param.EndpointId))
	for _,v := range param.PathList {
		sqls = append(sqls, fmt.Sprintf("INSERT INTO business_monitor(endpoint_id,path) VALUE (%d,'%s')", param.EndpointId, v))
	}
	return ExecuteTransactionSql(sqls)
}
