package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"fmt"
)

func GetProcessList(endpointId int) (err error, processList []*m.ProcessMonitorTable) {
	err = x.SQL("SELECT * FROM process_monitor WHERE endpoint_id=?", endpointId).Find(&processList)
	return err,processList
}

func UpdateProcess(param m.ProcessUpdateDto) error {
	var sqls []string
	sqls = append(sqls, fmt.Sprintf("DELETE FROM process_monitor WHERE endpoint_id=%d", param.EndpointId))
	for _,v := range param.ProcessList {
		sqls = append(sqls, fmt.Sprintf("INSERT INTO process_monitor(endpoint_id,NAME) VALUE (%d,'%s')", param.EndpointId, v))
	}
	return ExecuteTransactionSql(sqls)
}