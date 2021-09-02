package db

import (
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

func GetEndpointTypeList() (result []string,err error) {
	result = []string{}
	queryRows,queryErr := x.QueryString("select distinct t1.export_type from (select export_type from endpoint union select dashboard_type as export_type from dashboard) t1 order by t1.export_type")
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

func GetAlarmRealEndpoint(endpointId,strategyId int) (isReal bool,endpoint models.EndpointTable) {
	isReal = true
	endpoint = models.EndpointTable{}
	var tplTables []*models.TplTable
	x.SQL("select * from tpl where id in (select tpl_id from strategy where id=?)", strategyId).Find(&tplTables)
	if len(tplTables) > 0 {
		if tplTables[0].EndpointId > 0 {
			if tplTables[0].EndpointId == endpointId {
				return true,endpoint
			}
			endpoint.Id = tplTables[0].EndpointId
			GetEndpoint(&endpoint)
			return false,endpoint
		}else{
			var grpEndpointTables []*models.GrpEndpointTable
			x.SQL("select * from grp_endpoint where grp_id=? and endpoint_id=?", tplTables[0].GrpId, endpointId).Find(&grpEndpointTables)
			if len(grpEndpointTables) > 0 {
				return true,endpoint
			}
			var endpointTables []*models.EndpointTable
			x.SQL("select * from endpoint where guid in (select owner_endpoint from business_monitor where endpoint_id=?) and id in (select endpoint_id from grp_endpoint where grp_id=?)",endpointId,tplTables[0].GrpId).Find(&endpointTables)
			if len(endpointTables) > 0 {
				log.Logger.Info("Change alarm endpoint", log.Int("from", endpointId), log.String("to", endpointTables[0].Guid))
				return false,*endpointTables[0]
			}
		}
	}
	return true,endpoint
}