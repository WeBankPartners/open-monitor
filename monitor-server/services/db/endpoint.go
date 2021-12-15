package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"time"
)

func GetEndpointTypeList() (result []string, err error) {
	result = []string{}
	queryRows, queryErr := x.QueryString("select distinct t1.export_type from (select export_type from endpoint union select dashboard_type as export_type from dashboard) t1 order by t1.export_type")
	if queryErr != nil {
		err = queryErr
		return
	}
	for _, row := range queryRows {
		result = append(result, row["export_type"])
	}
	return
}

func GetEndpointByType(endpointType string) (result []*models.EndpointTable, err error) {
	result = []*models.EndpointTable{}
	err = x.SQL("select id,guid from endpoint where export_type=?", endpointType).Find(&result)
	return
}

func GetAlarmRealEndpoint(endpointId, strategyId int, endpointType, expr string) (isReal bool, endpoint models.EndpointTable) {
	isReal = true
	endpoint = models.EndpointTable{}
	if endpointType == "host" && strings.HasPrefix(expr, "node_business_monitor_value") {
		var businessMonitorTable []*models.BusinessMonitorTable
		x.SQL("select owner_endpoint from business_monitor where endpoint_id=?", endpointId).Find(&businessMonitorTable)
		if len(businessMonitorTable) > 0 {
			for _, v := range businessMonitorTable {
				if strings.Contains(expr, v.Path) {
					endpoint.Guid = v.OwnerEndpoint
					break
				}
			}
			if endpoint.Guid != "" {
				GetEndpoint(&endpoint)
				log.Logger.Info("Use business alarm endpoint", log.Int("from", endpointId), log.String("to", endpoint.Guid))
				return false, endpoint
			}
		}
	}
	var tplTables []*models.TplTable
	x.SQL("select * from tpl where id in (select tpl_id from strategy where id=?)", strategyId).Find(&tplTables)
	if len(tplTables) > 0 {
		if tplTables[0].EndpointId > 0 {
			if tplTables[0].EndpointId == endpointId {
				return true, endpoint
			}
			endpoint.Id = tplTables[0].EndpointId
			GetEndpoint(&endpoint)
			return false, endpoint
		} else {
			var grpEndpointTables []*models.GrpEndpointTable
			x.SQL("select * from grp_endpoint where grp_id=? and endpoint_id=?", tplTables[0].GrpId, endpointId).Find(&grpEndpointTables)
			if len(grpEndpointTables) > 0 {
				return true, endpoint
			}
			var endpointTables []*models.EndpointTable
			x.SQL("select * from endpoint where guid in (select owner_endpoint from business_monitor where endpoint_id=?) and id in (select endpoint_id from grp_endpoint where grp_id=?)", endpointId, tplTables[0].GrpId).Find(&endpointTables)
			if len(endpointTables) > 0 {
				log.Logger.Info("Change alarm endpoint", log.Int("from", endpointId), log.String("to", endpointTables[0].Guid))
				return false, *endpointTables[0]
			}
		}
	}
	return true, endpoint
}

func GetEndpointNew(param *models.EndpointNewTable) (result models.EndpointNewTable, err error) {
	var endpointNew []*models.EndpointNewTable
	var filterMessage string
	result = models.EndpointNewTable{}
	if param.Guid != "" {
		err = x.SQL("select * from endpoint_new where guid=?", param.Guid).Find(&endpointNew)
		filterMessage = fmt.Sprintf("guid=%s", param.Guid)
	} else if param.Ip != "" && param.MonitorType != "" {
		err = x.SQL("select * from endpoint_new where ip=? and monitor_type=?", param.Ip, param.MonitorType).Find(&endpointNew)
		filterMessage = fmt.Sprintf("ip=%s and monitor_type=%s", param.Ip, param.MonitorType)
	} else if param.AgentAddress != "" {
		err = x.SQL("select * from endpoint_new where agent_address=?", param.AgentAddress).Find(&endpointNew)
		filterMessage = fmt.Sprintf("agent_address=%s", param.AgentAddress)
	} else {
		err = fmt.Errorf("param illegal ")
	}
	if err != nil {
		return result, fmt.Errorf("Query endpoint fail,%s ", err.Error())
	}
	if len(endpointNew) == 0 {
		return result, fmt.Errorf("Can not find endpoint %s ", filterMessage)
	}
	result = *endpointNew[0]
	return result, nil
}

func ListEndpoint(param *models.QueryRequestParam) (pageInfo models.PageInfo, rowData []*models.EndpointNewTable, err error) {
	rowData = []*models.EndpointNewTable{}
	pageInfo = models.PageInfo{}
	filterSql, queryColumn, queryParam := transFiltersToSQL(param, &models.TransFiltersParam{IsStruct: true, StructObj: models.EndpointNewTable{}, PrimaryKey: "guid"})
	baseSql := fmt.Sprintf("SELECT %s FROM endpoint_new WHERE 1=1 %s ", queryColumn, filterSql)
	if param.Paging {
		pageInfo.StartIndex = param.Pageable.StartIndex
		pageInfo.PageSize = param.Pageable.PageSize
		pageInfo.TotalRows = queryCount(baseSql, queryParam...)
		pageSql, pageParam := transPageInfoToSQL(*param.Pageable)
		baseSql += pageSql
		queryParam = append(queryParam, pageParam...)
	}
	err = x.SQL(baseSql, queryParam...).Find(&rowData)
	return
}

func ListEndpointOptions(searchText string) (result []*models.OptionModel, err error) {
	result = []*models.OptionModel{}
	if searchText == "." {
		searchText = ""
	}
	searchText = "%" + searchText + "%"
	var endpointTable []*models.EndpointNewTable
	err = x.SQL("select guid,monitor_type from endpoint_new where guid like ?", searchText).Find(&endpointTable)
	if err != nil {
		return
	}
	for _, v := range endpointTable {
		result = append(result, &models.OptionModel{OptionValue: v.Guid, OptionText: v.Guid, OptionType: v.MonitorType, OptionTypeName: v.MonitorType})
	}
	return
}

func CheckEndpointInAgentManager(guid string) bool {
	queryRows,_ := x.QueryString("select endpoint_guid from agent_manager where endpoint_guid=?", guid)
	if len(queryRows) > 0 {
		return true
	}
	return false
}

func UpdateAgentManager(param *models.AgentManagerTable) error {
	_,err := x.Exec("update agent_manager set `user`=?,`password`=?,instance_address=?,agent_address=? where endpoint_guid=?", param.User, param.Password, param.InstanceAddress, param.AgentAddress, param.EndpointGuid)
	if err != nil {
		err = fmt.Errorf("Update agent manager fail,%s ", err.Error())
	}
	return err
}

func UpdateEndpointData(endpoint *models.EndpointNewTable) (err error) {
	_,err = x.Exec("update endpoint_new set agent_address=?,step=?,endpoint_address=?,extend_param=?,update_time=? where guid=?", endpoint.AgentAddress,endpoint.Step,endpoint.EndpointAddress,endpoint.ExtendParam,time.Now().Format(models.DatetimeFormat),endpoint.Guid)
	if err != nil {
		err = fmt.Errorf("Update endpoint table failj,%s ", err.Error())
	}
	return
}