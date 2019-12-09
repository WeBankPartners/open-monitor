package db

import (
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	mid "github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"fmt"
	"time"
)

func UpdateEndpoint(endpoint *m.EndpointTable) error {
	host := m.EndpointTable{Guid:endpoint.Guid}
	GetEndpoint(&host)
	if host.Id == 0 {
		insert := fmt.Sprintf("INSERT INTO endpoint(guid,name,ip,endpoint_version,export_type,export_version,step,address,os_type,create_at) VALUE ('%s','%s','%s','%s','%s','%s','%d','%s','%s','%s')",
			endpoint.Guid,endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,time.Now().Format(m.DatetimeFormat))
		_,err := x.Exec(insert)
		if err != nil {
			mid.LogError("insert endpoint fail ", err)
			return err
		}
		host := m.EndpointTable{Guid:endpoint.Guid}
		GetEndpoint(&host)
		endpoint.Id = host.Id
	}else{
		update := fmt.Sprintf("UPDATE endpoint SET name='%s',ip='%s',endpoint_version='%s',export_type='%s',export_version='%s',step=%d,address='%s',os_type='%s' WHERE id=%d",
			endpoint.Name,endpoint.Ip,endpoint.EndpointVersion,endpoint.ExportType,endpoint.ExportVersion,endpoint.Step,endpoint.Address,endpoint.OsType,endpoint.Id)
		_,err := x.Exec(update)
		if err != nil {
			mid.LogError("update endpoint fail ", err)
			return err
		}
		endpoint.Id = host.Id
	}
	return nil
}

func DeleteEndpoint(guid string) error {
	_,err := x.Exec(fmt.Sprintf("DELETE FROM endpoint_metric WHERE endpoint_id IN (SELECT id FROM endpoint WHERE guid='%s')", guid))
	_,err = x.Exec(fmt.Sprintf("DELETE FROM endpoint WHERE guid='%s'", guid))
	if err != nil {
		mid.LogError("delete endpoint fail", err)
		return err
	}
	return nil
}