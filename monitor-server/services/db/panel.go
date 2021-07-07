package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
)

func PanelList(id,groupId int) (result []*models.PanelTable,err error) {
	result = []*models.PanelTable{}
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
	err = x.SQL(baseSql, params...).Find(&result)
	return
}

func PanelCreate(endpointType string,param []*models.PanelTable) error {
	groupId,err := getPanelGroupId(endpointType)
	if err != nil {
		return err
	}
	maxGroupId,queryErr := getMaxChartGroupId()
	if queryErr != nil {
		return queryErr
	}
	var actions []*Action
	for _,panel := range param {
		maxGroupId = maxGroupId + 1
		if panel.TagsKey != "" {
			panel.TagsEnable = true
			panel.TagsUrl = "/dashboard/tags"
		}
		actions = append(actions, &Action{Sql: "insert into panel(group_id,title,tags_enable,tags_url,tags_key,chart_group) value (?,?,?,?,?,?)", Param: []interface{}{groupId,panel.Title,panel.TagsEnable,panel.TagsUrl,panel.TagsKey,maxGroupId}})
	}
	return Transaction(actions)
}

func getPanelGroupId(endpointType string) (groupId int,err error) {
	var dashboardTable []*models.DashboardTable
	err = x.SQL("select * from dashboard where dashboard_type=?", endpointType).Find(&dashboardTable)
	if err != nil {
		err = fmt.Errorf("Try to query dashboard table fail,%s ", err.Error())
		return
	}
	if len(dashboardTable) == 0 {
		maxIdQuery,queryErr := x.QueryString("select max(id) as id from dashboard")
		if queryErr != nil {
			err = fmt.Errorf("Try to get max dashboard id fail,%s ", queryErr.Error())
			return
		}
		maxId,_ := strconv.Atoi(maxIdQuery[0]["id"])
		maxId = maxId + 1
		_,err = x.Exec("insert into dashboard(dashboard_type,search_enable,search_id,button_enable,button_group,panels_group,panels_param) value (?,1,1,1,1,?,'endpoint={endpoint}')", endpointType, maxId)
		if err != nil {
			err = fmt.Errorf("Try to add dashboard data fail,%s ", err.Error())
			return
		}
		groupId = maxId
	}else{
		groupId = dashboardTable[0].PanelsGroup
	}
	return
}

func PanelUpdate(param []*models.PanelTable) error {
	var actions []*Action
	for _,panel := range param {
		if panel.TagsKey != "" {
			panel.TagsEnable = true
			panel.TagsUrl = "/dashboard/tags"
		}
		actions = append(actions, &Action{Sql: "update panel set title=?,tags_enable=?,tags_url=?,tags_key=? where id=?", Param: []interface{}{panel.Title,panel.TagsEnable,panel.TagsUrl,panel.TagsKey,panel.Id}})
	}
	return Transaction(actions)
}

func PanelDelete(param []*models.PanelTable) error {
	var actions []*Action
	for _,panel := range param {
		actions = append(actions, &Action{Sql: "delete from panel where id=?", Param: []interface{}{panel.Id}})
	}
	return Transaction(actions)
}
