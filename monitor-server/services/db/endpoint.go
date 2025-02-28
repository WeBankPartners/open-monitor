package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strings"
	"time"
)

func GetSimpleEndpointTypeList() (result []string, err error) {
	result = []string{}
	queryRows, queryErr := x.QueryString("select guid from monitor_type order by create_time desc")
	if queryErr != nil {
		err = queryErr
		return
	}
	for _, row := range queryRows {
		result = append(result, row["guid"])
	}
	return
}

func GetEndpointTypeList() (result []*models.MonitorTypeTable, err error) {
	result = []*models.MonitorTypeTable{}
	err = x.SQL("select guid,system_type from monitor_type order by create_time desc").Find(&result)
	return
}

func GetEndpointByType(endpointType, serviceGroup, endpointGroup, workspace string) (result []*models.EndpointNewTable, err error) {
	result = []*models.EndpointNewTable{}
	if serviceGroup != "" {
		serviceGroupObj, getSGErr := getSimpleServiceGroup(serviceGroup)
		if getSGErr != nil {
			err = getSGErr
			return
		}
		serviceGroupList, _ := fetchGlobalServiceGroupChildGuidList(serviceGroup)
		err = x.SQL("select guid from endpoint_new where monitor_type=? and guid in (select endpoint from endpoint_service_rel where service_group in ('"+strings.Join(serviceGroupList, "','")+"'))", endpointType).Find(&result)
		if err == nil {
			if workspace == "all_object" {
				result = append([]*models.EndpointNewTable{{Guid: serviceGroup, Name: serviceGroupObj.DisplayName}}, result...)
			}
		}
	} else if endpointGroup != "" {
		err = x.SQL("select guid from endpoint_new where guid in (select endpoint from endpoint_group_rel where endpoint_group=?)", endpointGroup).Find(&result)
		//if err == nil {
		//	if workspace == "all_object" {
		//		result = append([]*models.EndpointNewTable{{Guid: endpointGroup, Name: endpointGroup}}, result...)
		//	}
		//}
	} else {
		err = x.SQL("select guid from endpoint_new where monitor_type=?", endpointType).Find(&result)
	}
	for _, v := range result {
		if v.Name == "" {
			v.Name = v.Guid
		}
	}
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
				log.Info(nil, log.LOGGER_APP, "Use business alarm endpoint", zap.Int("from", endpointId), zap.String("to", endpoint.Guid))
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
				log.Info(nil, log.LOGGER_APP, "Change alarm endpoint", zap.Int("from", endpointId), zap.String("to", endpointTables[0].Guid))
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
		if param.MonitorType != "" {
			err = x.SQL("select * from endpoint_new where agent_address=? and monitor_type=?", param.AgentAddress, param.MonitorType).Find(&endpointNew)
			filterMessage = fmt.Sprintf("agent_address=%s and monitor_type=%s ", param.AgentAddress, param.MonitorType)
		} else {
			err = x.SQL("select * from endpoint_new where agent_address=?", param.AgentAddress).Find(&endpointNew)
			filterMessage = fmt.Sprintf("agent_address=%s", param.AgentAddress)
		}
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

func QueryEndpointList(endpoint string) (list []models.AlarmEndpoint, err error) {
	var endpointList []string
	if endpoint == "" {
		if err = x.SQL("select  guid from endpoint_new order by update_time desc limit 20").Find(&endpointList); err != nil {
			return
		}
	} else {
		if err = x.SQL("select  guid from endpoint_new where guid like '%" + endpoint + "%' order by update_time desc limit 20").Find(&endpointList); err != nil {
			return
		}
	}

	if len(endpointList) > 0 {
		for _, endpoint := range endpointList {
			list = append(list, models.AlarmEndpoint{
				Name:        endpoint,
				DisplayName: endpoint,
			})
		}
	}
	var serviceGroupList []*models.ServiceGroupTable
	if endpoint == "" {
		if err = x.SQL("select  * from service_group order by update_time desc limit 20").Find(&serviceGroupList); err != nil {
			return
		}
	} else {
		if err = x.SQL("select  * from service_group where display_name like '%" + endpoint + "%' order by update_time desc limit 20").Find(&serviceGroupList); err != nil {
			return
		}
	}

	if len(serviceGroupList) > 0 {
		for _, serviceGroup := range serviceGroupList {
			list = append(list, models.AlarmEndpoint{
				Name:        "sg__" + serviceGroup.Guid,
				DisplayName: serviceGroup.DisplayName,
			})
		}
	}
	if len(list) > 0 {
		list = list[:min(20, len(list))]
	}
	return
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func CheckEndpointInAgentManager(guid string) bool {
	queryRows, _ := x.QueryString("select endpoint_guid from agent_manager where endpoint_guid=?", guid)
	if len(queryRows) > 0 {
		return true
	}
	return false
}

func UpdateAgentManager(param *models.AgentManagerTable) error {
	var agentRemotePort string
	if splitIndex := strings.Index(param.AgentAddress, ":"); splitIndex >= 0 {
		agentRemotePort = param.AgentAddress[splitIndex+1:]
	}
	_, err := x.Exec("update agent_manager set `user`=?,`password`=?,instance_address=?,agent_address=?,agent_remote_port=? where endpoint_guid=?", param.User, param.Password, param.InstanceAddress, param.AgentAddress, agentRemotePort, param.EndpointGuid)
	if err != nil {
		err = fmt.Errorf("Update agent manager fail,%s ", err.Error())
	}
	return err
}

func UpdateEndpointData(oldEndpoint, endpoint *models.EndpointNewTable, operator string) (err error) {
	nowTimeString := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	if endpoint.Ip != "" {
		actions = append(actions, &Action{Sql: "update endpoint_new set ip=?,agent_address=?,step=?,endpoint_address=?,extend_param=?,update_time=?,update_user=? where guid=?", Param: []interface{}{
			endpoint.Ip, endpoint.AgentAddress, endpoint.Step, endpoint.EndpointAddress, endpoint.ExtendParam, nowTimeString, operator, endpoint.Guid,
		}})
		actions = append(actions, &Action{Sql: "update endpoint set ip=? where guid=?", Param: []interface{}{endpoint.Ip, endpoint.Guid}})
		//_, err = x.Exec("update endpoint_new set ip=?,agent_address=?,step=?,endpoint_address=?,extend_param=?,update_time=?,update_user=? where guid=?", endpoint.Ip, endpoint.AgentAddress, endpoint.Step, endpoint.EndpointAddress, endpoint.ExtendParam, nowTimeString, operator, endpoint.Guid)
		//x.Exec("update endpoint set ip=? where guid=?", endpoint.Ip, endpoint.Guid)
	} else {
		actions = append(actions, &Action{Sql: "update endpoint_new set agent_address=?,step=?,endpoint_address=?,extend_param=?,update_time=?,update_user=? where guid=?", Param: []interface{}{endpoint.AgentAddress, endpoint.Step, endpoint.EndpointAddress, endpoint.ExtendParam, nowTimeString, operator, endpoint.Guid}})
		//_, err = x.Exec("update endpoint_new set agent_address=?,step=?,endpoint_address=?,extend_param=?,update_time=?,update_user=? where guid=?", endpoint.AgentAddress, endpoint.Step, endpoint.EndpointAddress, endpoint.ExtendParam, nowTimeString, operator, endpoint.Guid)
	}
	if endpoint.AgentAddress != endpoint.EndpointAddress {
		actions = append(actions, &Action{Sql: "update endpoint set address_agent=?,address=?,step=? where guid=?", Param: []interface{}{endpoint.AgentAddress, endpoint.EndpointAddress, endpoint.Step, endpoint.Guid}})
	} else {
		actions = append(actions, &Action{Sql: "update endpoint set address=?,step=? where guid=?", Param: []interface{}{endpoint.EndpointAddress, endpoint.Step, endpoint.Guid}})
	}
	if oldEndpoint.MonitorType == "host" && (oldEndpoint.AgentAddress != endpoint.AgentAddress) {
		actions = append(actions, &Action{Sql: "update endpoint_new set agent_address=? where agent_address=?", Param: []interface{}{endpoint.AgentAddress, oldEndpoint.AgentAddress}})
		actions = append(actions, &Action{Sql: "update endpoint set address=? where address=?", Param: []interface{}{endpoint.AgentAddress, oldEndpoint.AgentAddress}})
	}
	err = Transaction(actions)
	if err != nil {
		err = fmt.Errorf("Update endpoint table failj,%s ", err.Error())
	}
	//if err != nil {
	//	err = fmt.Errorf("Update endpoint table failj,%s ", err.Error())
	//} else {
	//	if endpoint.AgentAddress != endpoint.EndpointAddress {
	//		x.Exec("update endpoint set address_agent=?,address=?,step=? where guid=?", endpoint.AgentAddress, endpoint.EndpointAddress, endpoint.Step, endpoint.Guid)
	//	} else {
	//		x.Exec("update endpoint set address=?,step=? where guid=?", endpoint.EndpointAddress, endpoint.Step, endpoint.Guid)
	//	}
	//}
	return
}

func GetProcessByHostEndpoint(hostIp string) (processEndpoints []*models.EndpointNewTable, err error) {
	processEndpoints = []*models.EndpointNewTable{}
	err = x.SQL("select * from endpoint_new where monitor_type='process' and ip=?", hostIp).Find(&processEndpoints)
	if err != nil {
		err = fmt.Errorf("query endpoint with process ip fail,%s ", err.Error())
	}
	return
}
