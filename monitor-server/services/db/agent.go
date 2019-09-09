package db

import (
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	mid "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
	"fmt"
)

func UpdateEndpoint(endpoint *m.EndpointTable) error {
	err,host := GetEndpoint(endpoint.Guid)
	if err != nil {
		return err
	}else if host.Id == 0 {
		insert := fmt.Sprintf("INSERT INTO endpoint(guid,NAME,ip,export_type,os_ip,os_type) VALUE ('%s','%s','%s','%s','%s','%s')",endpoint.Guid,endpoint.Name,endpoint.Ip,endpoint.ExportType,endpoint.OsIp,endpoint.OsType)
		_,ierr := x.Exec(insert)
		if ierr != nil {
			mid.LogError("insert endpoint fail ", ierr)
			return ierr
		}
		_,host = GetEndpoint(endpoint.Guid)
		endpoint.Id = host.Id
	}else{
		update := fmt.Sprintf("UPDATE endpoint SET name='%s',ip='%s',export_type='%s',os_ip='%s',os_type='%s' WHERE id=%d", endpoint.Name,endpoint.Ip,endpoint.ExportType,endpoint.OsIp,endpoint.OsType,endpoint.Id)
		_,uerr := x.Exec(update)
		if uerr != nil {
			mid.LogError("update endpoint fail ", uerr)
			return uerr
		}
	}
	return nil
}
