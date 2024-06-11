package db

import (
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"time"
)

func ListEndpointGroup(param *models.QueryRequestParam) (pageInfo models.PageInfo, rowData []*models.EndpointGroupTable, err error) {
	rowData = []*models.EndpointGroupTable{}
	pageInfo = models.PageInfo{}
	filterSql, queryColumn, queryParam := transFiltersToSQL(param, &models.TransFiltersParam{IsStruct: true, StructObj: models.EndpointGroupTable{}, PrimaryKey: "guid"})
	baseSql := fmt.Sprintf("SELECT %s FROM endpoint_group WHERE service_group is null %s ", queryColumn, filterSql)
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

func CreateEndpointGroup(param *models.EndpointGroupTable) error {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	if param.DisplayName == "" {
		param.DisplayName = param.Guid
	}
	if param.ServiceGroup == "" {
		actions = append(actions, &Action{Sql: "insert into endpoint_group(guid,display_name,description,monitor_type,update_time) value (?,?,?,?,?)", Param: []interface{}{param.Guid, param.DisplayName, param.Description, param.MonitorType, nowTime}})
	} else {
		actions = append(actions, &Action{Sql: "insert into endpoint_group(guid,display_name,description,monitor_type,service_group,update_time) value (?,?,?,?,?,?)", Param: []interface{}{param.Guid, param.DisplayName, param.Description, param.MonitorType, param.ServiceGroup, nowTime}})
	}
	actions = append(actions, &Action{Sql: "insert into grp(name,description,endpoint_type,create_at) value (?,?,?,NOW())", Param: []interface{}{param.Guid, param.Description, param.MonitorType}})
	return Transaction(actions)
}

func UpdateEndpointGroup(param *models.EndpointGroupTable) error {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	if param.DisplayName == "" {
		param.DisplayName = param.Guid
	}
	if param.ServiceGroup == "" {
		actions = append(actions, &Action{Sql: "update endpoint_group set display_name=?,description=?,monitor_type=?,service_group=NULL,update_time=? where guid=?", Param: []interface{}{param.DisplayName, param.Description, param.MonitorType, nowTime, param.Guid}})
	} else {
		actions = append(actions, &Action{Sql: "update endpoint_group set display_name=?,description=?,monitor_type=?,service_group=?,update_time=? where guid=?", Param: []interface{}{param.DisplayName, param.Description, param.MonitorType, param.ServiceGroup, nowTime, param.Guid}})
	}
	actions = append(actions, &Action{Sql: "update grp set description=? where name=?", Param: []interface{}{param.Description, param.Guid}})
	return Transaction(actions)
}

func DeleteEndpointGroup(endpointGroupGuid string) (err error) {
	err = Transaction(getDeleteEndpointGroupAction(endpointGroupGuid))
	if err == nil {
		RemovePrometheusRuleFile(endpointGroupGuid, false)
	}
	return err
}

func GetGroupEndpointRel(endpointGroupGuid string) (result []*models.EndpointGroupRelTable, err error) {
	result = []*models.EndpointGroupRelTable{}
	err = x.SQL("select endpoint from endpoint_group_rel where endpoint_group=?", endpointGroupGuid).Find(&result)
	return
}

func UpdateGroupEndpoint(param *models.UpdateGroupEndpointParam, appendFlag bool) error {
	var actions []*Action
	if !appendFlag {
		actions = append(actions, &Action{Sql: "delete from endpoint_group_rel where endpoint_group=?", Param: []interface{}{param.GroupGuid}})
	}
	guidList := guid.CreateGuidList(len(param.EndpointGuidList))
	nowTime := time.Now().Format(models.DatetimeFormat)
	for i, v := range param.EndpointGuidList {
		actions = append(actions, &Action{Sql: "insert into endpoint_group_rel(guid,endpoint,endpoint_group) value (?,?,?)", Param: []interface{}{guidList[i], v, param.GroupGuid}})
	}
	actions = append(actions, &Action{Sql: "update endpoint_group set update_time=? where guid=?", Param: []interface{}{nowTime, param.GroupGuid}})
	err := Transaction(actions)
	return err
}

func GetGroupEndpointNotify(endpointGroupGuid string) (result []*models.NotifyObj, err error) {
	result = getNotifyList("", endpointGroupGuid, "")
	return result, nil
}

func UpdateGroupEndpointNotify(endpointGroupGuid string, param []*models.NotifyObj) error {
	for _, v := range param {
		v.EndpointGroup = endpointGroupGuid
		if v.ProcCallbackKey != "" && v.ProcCallbackName == "" {
			return fmt.Errorf("procCallbackName can not empty with key:%s ", v.ProcCallbackKey)
		}
	}
	//actions := getNotifyListDeleteAction("", endpointGroupGuid, "")
	//actions = append(actions, getNotifyListInsertAction(param)...)
	actions := getNotifyListUpdateAction(param)
	return Transaction(actions)
}

func ListEndpointGroupOptions(searchText string) (result []*models.OptionModel, err error) {
	result = []*models.OptionModel{}
	if searchText == "." {
		searchText = ""
	}
	searchText = "%" + searchText + "%"
	var endpointGroupTable []*models.EndpointGroupTable
	err = x.SQL("select guid,monitor_type from endpoint_group where service_group is null and (guid like ?)", searchText).Find(&endpointGroupTable)
	if err != nil {
		return
	}
	for _, v := range endpointGroupTable {
		result = append(result, &models.OptionModel{OptionValue: v.Guid, OptionText: v.Guid, OptionType: v.MonitorType, OptionTypeName: v.MonitorType})
	}
	return
}

func GetSimpleEndpointGroup(guid string) (result *models.EndpointGroupTable, err error) {
	var endpointGroup []*models.EndpointGroupTable
	err = x.SQL("select * from endpoint_group where guid=?", guid).Find(&endpointGroup)
	if err != nil {
		return
	}
	if len(endpointGroup) == 0 {
		return result, fmt.Errorf("Can not find endpointGroup with guid:%s ", guid)
	}
	result = endpointGroup[0]
	return
}

func getCreateEndpointGroupAction(serviceGroupGuid, monitorType, nowTime string) (actions []*Action) {
	endpointGroupGuid := fmt.Sprintf("service_%s_%s", serviceGroupGuid, monitorType)
	var endpointGroup []*models.EndpointGroupTable
	x.SQL("select guid from endpoint_group where guid=?", endpointGroupGuid).Find(&endpointGroup)
	if len(endpointGroup) > 0 {
		return actions
	}
	actions = append(actions, &Action{Sql: "insert into endpoint_group(guid,display_name,monitor_type,service_group,update_time) value (?,?,?,?,?)", Param: []interface{}{endpointGroupGuid, endpointGroupGuid, monitorType, serviceGroupGuid, nowTime}})
	return actions
}

func getDeleteEndpointGroupAction(endpointGroupGuid string) (actions []*Action) {
	actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify in (select guid from notify where endpoint_group=? or alarm_strategy in (select guid from alarm_strategy where endpoint_group=?))", Param: []interface{}{endpointGroupGuid, endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from notify where endpoint_group=? or alarm_strategy in (select guid from alarm_strategy where endpoint_group=?)", Param: []interface{}{endpointGroupGuid, endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where endpoint_group=?", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from endpoint_group_rel where endpoint_group=?", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from endpoint_group where guid=?", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from grp where name=?", Param: []interface{}{endpointGroupGuid}})
	return actions
}
