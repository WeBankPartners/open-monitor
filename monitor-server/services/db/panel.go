package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"strconv"
)

func PanelList(id int, endpointType, serviceGroup string) (result []*models.PanelTable, err error) {
	groupId := 0
	result = []*models.PanelTable{}
	if endpointType != "" {
		var dashboardTable []*models.DashboardTable
		err = x.SQL("select * from dashboard where dashboard_type=?", endpointType).Find(&dashboardTable)
		if err != nil {
			err = fmt.Errorf("Try to query dashboard table fail,%s ", err.Error())
			return
		}
		if len(dashboardTable) == 0 {
			return
		}
		groupId = dashboardTable[0].PanelsGroup
	}
	params := []interface{}{}
	baseSql := "select * from panel where 1=1 "
	if id > 0 {
		baseSql += " and id=? "
		params = append(params, id)
	}
	if groupId > 0 {
		baseSql += " and group_id=? "
		params = append(params, groupId)
	}
	if endpointType == "host" {
		baseSql += " and title<>'Business' "
	}
	if serviceGroup != "" {
		baseSql += " and service_group=? "
		params = append(params, serviceGroup)
	} else {
		baseSql += " and service_group is null "
	}
	err = x.SQL(baseSql, params...).Find(&result)
	return
}

func PanelCreate(endpointType string, param []*models.PanelTable) error {
	groupId, err := getPanelGroupId(endpointType)
	if err != nil {
		return err
	}
	maxGroupId, queryErr := getMaxChartGroupId()
	if queryErr != nil {
		return queryErr
	}
	var actions []*Action
	for _, panel := range param {
		maxGroupId = maxGroupId + 1
		if panel.TagsKey != "" {
			panel.TagsEnable = true
			panel.TagsUrl = "/dashboard/tags"
		}
		if panel.ServiceGroup != "" {
			actions = append(actions, &Action{Sql: "insert into panel(group_id,title,tags_enable,tags_url,tags_key,chart_group,service_group) value (?,?,?,?,?,?,?)", Param: []interface{}{groupId, panel.Title, panel.TagsEnable, panel.TagsUrl, panel.TagsKey, maxGroupId, panel.ServiceGroup}})
		} else {
			actions = append(actions, &Action{Sql: "insert into panel(group_id,title,tags_enable,tags_url,tags_key,chart_group) value (?,?,?,?,?,?)", Param: []interface{}{groupId, panel.Title, panel.TagsEnable, panel.TagsUrl, panel.TagsKey, maxGroupId}})
		}
	}
	return Transaction(actions)
}

func getPanelGroupId(endpointType string) (groupId int, err error) {
	var dashboardTable []*models.DashboardTable
	err = x.SQL("select * from dashboard where dashboard_type=?", endpointType).Find(&dashboardTable)
	if err != nil {
		err = fmt.Errorf("Try to query dashboard table fail,%s ", err.Error())
		return
	}
	if len(dashboardTable) == 0 {
		maxIdQuery, queryErr := x.QueryString("select max(id) as id from dashboard")
		if queryErr != nil {
			err = fmt.Errorf("Try to get max dashboard id fail,%s ", queryErr.Error())
			return
		}
		maxId, _ := strconv.Atoi(maxIdQuery[0]["id"])
		maxId = maxId + 1
		_, err = x.Exec("insert into dashboard(dashboard_type,search_enable,search_id,button_enable,button_group,panels_group,panels_param) value (?,1,1,1,1,?,'endpoint={endpoint}')", endpointType, maxId)
		if err != nil {
			err = fmt.Errorf("Try to add dashboard data fail,%s ", err.Error())
			return
		}
		groupId = maxId
	} else {
		groupId = dashboardTable[0].PanelsGroup
	}
	return
}

func PanelUpdate(param []*models.PanelTable) error {
	var actions []*Action
	for _, panel := range param {
		if panel.TagsKey != "" {
			panel.TagsEnable = true
			panel.TagsUrl = "/dashboard/tags"
		}
		if panel.ServiceGroup != "" {
			actions = append(actions, &Action{Sql: "update panel set title=?,tags_enable=?,tags_url=?,tags_key=?,service_group=? where id=?", Param: []interface{}{panel.Title, panel.TagsEnable, panel.TagsUrl, panel.TagsKey, panel.ServiceGroup, panel.Id}})
		} else {
			actions = append(actions, &Action{Sql: "update panel set title=?,tags_enable=?,tags_url=?,tags_key=? where id=?", Param: []interface{}{panel.Title, panel.TagsEnable, panel.TagsUrl, panel.TagsKey, panel.Id}})
		}
	}
	return Transaction(actions)
}

func PanelDelete(ids []string) error {
	var actions []*Action
	for _, id := range ids {
		idInt, tmpErr := strconv.Atoi(id)
		if tmpErr != nil {
			log.Error(nil, log.LOGGER_APP, "Try to trans id to int fail", zap.Error(tmpErr))
			continue
		}
		actions = append(actions, &Action{Sql: "delete from panel where id=?", Param: []interface{}{idInt}})
	}
	return Transaction(actions)
}
