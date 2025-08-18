package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strings"
	"time"
)

func ListEndpointGroupMonitoryType() (result []string, err error) {
	err = x.SQL("select distinct monitor_type from endpoint_group").Find(&result)
	return
}

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

func CreateEndpointGroup(param *models.EndpointGroupTable, operator string) error {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	if param.DisplayName == "" {
		param.DisplayName = param.Guid
	}
	if param.ServiceGroup == "" {
		actions = append(actions, &Action{Sql: "insert into endpoint_group(guid,display_name,description,monitor_type,update_time,create_user,update_user) value (?,?,?,?,?,?,?)",
			Param: []interface{}{param.Guid, param.DisplayName, param.Description, param.MonitorType, nowTime, operator, operator}})
	} else {
		actions = append(actions, &Action{Sql: "insert into endpoint_group(guid,display_name,description,monitor_type,service_group,update_time,create_user,update_user) value (?,?,?,?,?,?,?,?)",
			Param: []interface{}{param.Guid, param.DisplayName, param.Description, param.MonitorType, param.ServiceGroup, nowTime, operator, operator}})
	}
	actions = append(actions, &Action{Sql: "insert into grp(name,description,endpoint_type,create_at) value (?,?,?,NOW())", Param: []interface{}{param.Guid, param.Description, param.MonitorType}})
	return Transaction(actions)
}

func UpdateEndpointGroup(param *models.EndpointGroupTable, operator string) error {
	var actions []*Action
	nowTime := time.Now().Format(models.DatetimeFormat)
	if param.DisplayName == "" {
		param.DisplayName = param.Guid
	}
	if param.ServiceGroup == "" {
		actions = append(actions, &Action{Sql: "update endpoint_group set display_name=?,description=?,monitor_type=?,service_group=NULL,update_time=?,update_user=? where guid=?", Param: []interface{}{param.DisplayName, param.Description, param.MonitorType, nowTime, operator, param.Guid}})
	} else {
		actions = append(actions, &Action{Sql: "update endpoint_group set display_name=?,description=?,monitor_type=?,service_group=?,update_time=?,update_user=? where guid=?", Param: []interface{}{param.DisplayName, param.Description, param.MonitorType, param.ServiceGroup, nowTime, operator, param.Guid}})
	}
	actions = append(actions, &Action{Sql: "update grp set description=? where name=?", Param: []interface{}{param.Description, param.Guid}})
	return Transaction(actions)
}

func DeleteEndpointGroup(endpointGroupGuid string) (err error) {
	log.Info(nil, log.LOGGER_APP, "DeleteEndpointGroup start", zap.String("endpointGroupGuid", endpointGroupGuid))
	err = Transaction(getDeleteEndpointGroupAction(endpointGroupGuid))
	if err == nil {
		log.Info(nil, log.LOGGER_APP, "DeleteEndpointGroup - Transaction success, calling RemovePrometheusRuleFile", zap.String("endpointGroupGuid", endpointGroupGuid))
		RemovePrometheusRuleFile(endpointGroupGuid, false)
		log.Info(nil, log.LOGGER_APP, "DeleteEndpointGroup - RemovePrometheusRuleFile completed", zap.String("endpointGroupGuid", endpointGroupGuid))
	} else {
		log.Error(nil, log.LOGGER_APP, "DeleteEndpointGroup - Transaction failed", zap.String("endpointGroupGuid", endpointGroupGuid), zap.Error(err))
	}
	return err
}

func GetGroupEndpointRel(endpointGroupGuid string) (result []*models.EndpointGroupRelTable, err error) {
	result = []*models.EndpointGroupRelTable{}
	err = x.SQL("select endpoint from endpoint_group_rel where endpoint_group=?", endpointGroupGuid).Find(&result)
	return
}

func UpdateGroupEndpoint(param *models.UpdateGroupEndpointParam, operator string, appendFlag bool) error {
	var actions []*Action
	if !appendFlag {
		actions = append(actions, &Action{Sql: "delete from endpoint_group_rel where endpoint_group=?", Param: []interface{}{param.GroupGuid}})
	}
	guidList := guid.CreateGuidList(len(param.EndpointGuidList))
	nowTime := time.Now().Format(models.DatetimeFormat)
	for i, v := range param.EndpointGuidList {
		actions = append(actions, &Action{Sql: "insert into endpoint_group_rel(guid,endpoint,endpoint_group) value (?,?,?)", Param: []interface{}{guidList[i], v, param.GroupGuid}})
	}
	actions = append(actions, &Action{Sql: "update endpoint_group set update_time=?,update_user=? where guid=?", Param: []interface{}{nowTime, operator, param.GroupGuid}})
	err := Transaction(actions)
	return err
}

func GetGroupEndpointNotify(endpointGroupGuid string) (result []*models.NotifyObj, err error) {
	result = getNotifyList("", endpointGroupGuid, "")
	return result, nil
}

func UpdateGroupEndpointNotify(endpointGroupGuid string, param []*models.NotifyObj) error {
	endpointGroupRow, getEndpointGroupErr := GetSimpleEndpointGroup(endpointGroupGuid)
	if getEndpointGroupErr != nil {
		return getEndpointGroupErr
	}
	for _, v := range param {
		v.EndpointGroup = endpointGroupGuid
		v.ServiceGroup = endpointGroupRow.ServiceGroup
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

func getDeleteEndpointGroupAction(endpointGroupGuid string) (actions []*Action) {
	log.Info(nil, log.LOGGER_APP, "getDeleteEndpointGroupAction start", zap.String("endpointGroupGuid", endpointGroupGuid))
	actions = append(actions, &Action{Sql: "delete from notify_role_rel where notify in (select guid from notify where endpoint_group=? or alarm_strategy in (select guid from alarm_strategy where endpoint_group=?))", Param: []interface{}{endpointGroupGuid, endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from notify where endpoint_group=? or alarm_strategy in (select guid from alarm_strategy where endpoint_group=?)", Param: []interface{}{endpointGroupGuid, endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy_tag_value where alarm_strategy_tag in (select guid from  alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy in (select guid from alarm_strategy where endpoint_group=?)))", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy_tag where alarm_strategy_metric in (select guid from alarm_strategy_metric where alarm_strategy in (select guid from alarm_strategy where endpoint_group=?))", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy_metric where alarm_strategy in (select guid from alarm_strategy where endpoint_group=?)", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from alarm_strategy where endpoint_group=?", Param: []interface{}{endpointGroupGuid}})
	log.Info(nil, log.LOGGER_APP, "getDeleteEndpointGroupAction - will delete alarm_strategy", zap.String("endpointGroupGuid", endpointGroupGuid))
	actions = append(actions, &Action{Sql: "delete from endpoint_group_rel where endpoint_group=?", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from endpoint_group where guid=?", Param: []interface{}{endpointGroupGuid}})
	actions = append(actions, &Action{Sql: "delete from grp where name=?", Param: []interface{}{endpointGroupGuid}})
	log.Info(nil, log.LOGGER_APP, "getDeleteEndpointGroupAction end", zap.String("endpointGroupGuid", endpointGroupGuid), zap.Int("actionsCount", len(actions)))
	return actions
}

func GetSimpleGrp(name string) (result *models.GrpTable, err error) {
	var grpRows []*models.GrpTable
	err = x.SQL("select * from grp where name=?", name).Find(&grpRows)
	if err != nil {
		return
	}
	if len(grpRows) > 0 {
		result = grpRows[0]
	}
	return
}
