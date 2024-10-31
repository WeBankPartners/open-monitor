package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strconv"
	"strings"
	"time"
)

func GetOrganizationList(nameText, endpointText string) (result []*m.OrganizationPanel, err error) {
	nameText = strings.ToLower(nameText)
	endpointText = strings.ToLower(endpointText)
	var data []*m.PanelRecursiveTable
	err = x.SQL("SELECT * FROM panel_recursive").Find(&data)
	if err != nil {
		log.Logger.Error("Get panel_recursive table error", log.Error(err))
		return result, err
	}
	if len(data) == 0 {
		return result, err
	}
	tmpMap := make(map[string]string)
	objTypeMap := make(map[string]string)
	endpointMap := make(map[string]string)
	for _, v := range data {
		tmpMap[v.Guid] = v.DisplayName
		objTypeMap[v.Guid] = v.ObjType
		endpointMap[v.Guid] = v.Endpoint
	}
	var headers []string
	for _, v := range data {
		if v.Parent == "" {
			headers = append(headers, v.Guid)
		} else {
			tmpFlag := true
			for _, vv := range strings.Split(v.Parent, "^") {
				if tmpMap[vv] != "" {
					tmpFlag = false
					break
				}
			}
			if tmpFlag {
				headers = append(headers, v.Guid)
			}
		}
	}
	for _, v := range headers {
		tmpHeaderObj := m.OrganizationPanel{Guid: v, DisplayName: tmpMap[v], Type: objTypeMap[v]}
		if nameText != "" {
			if strings.Contains(strings.ToLower(tmpMap[v]), nameText) {
				tmpHeaderObj.FetchSearch = true
				tmpHeaderObj.FetchOriginFlag = true
			}
		}
		if endpointText != "" {
			if strings.Contains(strings.ToLower(endpointMap[v]), endpointText) {
				tmpHeaderObj.FetchSearch = true
				tmpHeaderObj.FetchOriginFlag = true
			}
		}
		tmpNodeList := recursiveOrganization(data, v, tmpHeaderObj, nameText, endpointText)
		if nameText != "" || endpointText != "" {
			if tmpNodeList.FetchOriginFlag == false {
				continue
			}
		}
		result = append(result, &tmpNodeList)
	}
	return result, nil
}

func recursiveOrganization(data []*m.PanelRecursiveTable, parent string, tmpNode m.OrganizationPanel, nameText, endpointText string) m.OrganizationPanel {
	for _, v := range data {
		if v.Parent == "" || v.Guid == parent {
			continue
		}
		tmpFlag := false
		for _, vv := range strings.Split(v.Parent, "^") {
			if vv == parent {
				tmpFlag = true
				break
			}
		}
		if tmpFlag {
			tmpOrganizationObj := m.OrganizationPanel{Guid: v.Guid, DisplayName: v.DisplayName, Type: v.ObjType}
			if endpointText != "" {
				if strings.Contains(strings.ToLower(v.Endpoint), endpointText) {
					tmpOrganizationObj.FetchSearch = true
					tmpOrganizationObj.FetchOriginFlag = true
				}
			}
			if nameText != "" {
				if strings.Contains(strings.ToLower(v.DisplayName), nameText) {
					tmpOrganizationObj.FetchSearch = true
					tmpOrganizationObj.FetchOriginFlag = true
				}
			}
			tn := recursiveOrganization(data, v.Guid, tmpOrganizationObj, nameText, endpointText)
			if nameText != "" || endpointText != "" {
				//if tn.FetchOriginFlag == false {
				//	continue
				//}
				tmpNode.Children = append(tmpNode.Children, &tn)
				for _, tmpChildren := range tmpNode.Children {
					if tmpChildren.FetchOriginFlag {
						tmpNode.FetchOriginFlag = true
						break
					}
				}
			} else {
				tmpNode.Children = append(tmpNode.Children, &tn)
			}
		}
	}
	return tmpNode
}

func UpdateOrganization(operation string, param m.UpdateOrgPanelParam) (err error) {
	log.Logger.Info("start UpdateOrganization", log.String("operation", operation), log.String("guid", param.Guid))
	var tableData []*m.PanelRecursiveTable
	var dbMetricMonitorList []*m.DbMetricMonitorObj
	var actions, delLogMetricMonitorActions []*Action
	var endpointGroup, affectHost, subHost, affectEndpointGroup []string
	var logMetricMonitorList []*m.LogMetricMonitorTable
	nowTime := time.Now().Format(m.DatetimeFormat)
	if operation == "add" {
		if param.Guid == "" || param.DisplayName == "" {
			return fmt.Errorf("param guid and display_name cat not be empty")
		}
		x.SQL("SELECT guid,display_name,parent FROM panel_recursive WHERE guid=?", param.Guid).Find(&tableData)
		if len(tableData) > 0 {
			return fmt.Errorf("guid already exist")
		}
		//_, err = x.Exec("INSERT INTO panel_recursive(guid,display_name,parent,obj_type) VALUE (?,?,?,?)", param.Guid, param.DisplayName, param.Parent, param.Type)
		actions = append(actions, &Action{Sql: "INSERT INTO panel_recursive(guid,display_name,parent,obj_type) VALUE (?,?,?,?)", Param: []interface{}{param.Guid, param.DisplayName, param.Parent, param.Type}})
		actions = append(actions, getCreateServiceGroupAction(&m.ServiceGroupTable{Guid: param.Guid, DisplayName: param.DisplayName, Description: "", Parent: param.Parent, ServiceType: param.Type, UpdateTime: nowTime}, operation)...)
		err = Transaction(actions)
		if err == nil {
			addGlobalServiceGroupNode(m.ServiceGroupTable{Guid: param.Guid, Parent: param.Parent, DisplayName: param.DisplayName})
		}
	} else if operation == "edit" {
		if param.Guid == "" || param.DisplayName == "" {
			return fmt.Errorf("param guid and display_name cat not be empty")
		}
		x.SQL("SELECT guid,display_name,parent FROM panel_recursive WHERE guid=?", param.Guid).Find(&tableData)
		if len(tableData) == 0 {
			return fmt.Errorf("guid: %s can not find any record", param.Guid)
		}
		actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET display_name=?,obj_type=? WHERE guid=?", Param: []interface{}{param.DisplayName, param.Type, param.Guid}})
		actions = append(actions, &Action{Sql: "update service_group set display_name=?,service_type=?,update_user=? where guid=?", Param: []interface{}{param.DisplayName, param.Type, operation, param.Guid}})
		err = Transaction(actions)
		if err == nil {
			m.GlobalSGDisplayNameMap[param.Guid] = param.DisplayName
		}
	} else if operation == "delete" {
		if param.Guid == "" {
			return fmt.Errorf("param guid cat not be empty")
		}
		x.SQL("SELECT guid,display_name,parent FROM panel_recursive").Find(&tableData)
		if len(tableData) == 0 {
			return fmt.Errorf("guid:%s can not find any record", param.Guid)
		}
		guidList := getNodeFromParent(tableData, []string{param.Guid}, param.Guid)
		tmpMap := make(map[string]bool)
		for _, v := range guidList {
			tmpMap[v] = true
		}
		guidList = []string{}
		for k, _ := range tmpMap {
			if k != "" {
				guidList = append(guidList, k)
			}
		}
		actions = append(actions, &Action{Sql: fmt.Sprintf("DELETE FROM panel_recursive WHERE guid in ('%s')", strings.Join(guidList, "','"))})
		// 删除业务配置列表-数据库
		if dbMetricMonitorList, err = GetDbMetricByServiceGroup(param.Guid, ""); err != nil {
			return err
		}
		if len(dbMetricMonitorList) > 0 {
			for _, obj := range dbMetricMonitorList {
				if subDbMetricList, subEndpointGroup := GetDeleteDbMetricActions(obj.Guid); len(subDbMetricList) > 0 {
					actions = append(actions, subDbMetricList...)
					if len(subEndpointGroup) > 0 {
						endpointGroup = append(endpointGroup, subEndpointGroup...)
					}
				}
			}
		}
		// 删除业务配置列表-日志文件
		if err = x.SQL("select * from log_metric_monitor where service_group=?", param.Guid).Find(&logMetricMonitorList); err != nil {
			return
		}
		if len(logMetricMonitorList) > 0 {
			for _, logMetricMonitor := range logMetricMonitorList {
				delLogMetricMonitorActions, subHost, affectEndpointGroup = getDeleteLogMetricMonitor(logMetricMonitor.Guid)
				if len(delLogMetricMonitorActions) > 0 {
					actions = append(actions, delLogMetricMonitorActions...)
				}
				if len(affectEndpointGroup) > 0 {
					endpointGroup = append(endpointGroup, affectEndpointGroup...)
				}
				if len(subHost) > 0 {
					affectHost = append(affectHost, subHost...)
				}
			}
		}

		// 删除 层级对象下面的指标列表
		var metricList []*m.MetricTable
		if err = x.SQL("select * from metric  where service_group = ?", param.Guid).Find(&metricList); err != nil {
			return
		}
		if len(metricList) > 0 {
			// 删除同环比 指标
			for _, metric := range metricList {
				tmpActions, tmpEndpointGroup := getMetricComparisonDeleteAction(metric.Guid)
				actions = append(actions, tmpActions...)
				affectEndpointGroup = append(tmpEndpointGroup, tmpEndpointGroup...)
				tmpActions, tmpEndpointGroup = getDeleteMetricActions(metric.Guid)
				actions = append(actions, tmpActions...)
				affectEndpointGroup = append(tmpEndpointGroup, tmpEndpointGroup...)
				tmpActions = getDeleteEndpointDashboardChartMetricAction(metric.Metric, metric.MetricType)
				actions = append(actions, tmpActions...)
				if len(affectEndpointGroup) > 0 {
					endpointGroup = append(endpointGroup, affectEndpointGroup...)
				}
			}
		}
		actions = append(actions, getDeleteServiceGroupAction(param.Guid, guidList)...)
		err = Transaction(actions)
		if err == nil {
			DeleteServiceWithChildConfig(param.Guid)
			deleteGlobalServiceGroupNode(param.Guid)
			if len(affectHost) > 0 {
				err = SyncLogMetricExporterConfig(affectHost)
				if err != nil {
					log.Logger.Error("SyncLogMetricExporterConfig fail", log.Error(err))
				}
			}
			if len(endpointGroup) > 0 {
				for _, v := range endpointGroup {
					SyncPrometheusRuleFile(v, false)
				}
			}
		}
	}
	return err
}

func getNodeFromParent(data []*m.PanelRecursiveTable, input []string, guid string) []string {
	tmpInput := input
	for _, v := range data {
		tmpFlag := false
		for _, vv := range strings.Split(v.Parent, "^") {
			if vv == guid {
				tmpFlag = true
				break
			}
		}
		if tmpFlag {
			tmpInput = append(tmpInput, v.Guid)
			tmpResult := getNodeFromParent(data, tmpInput, v.Guid)
			for _, vv := range tmpResult {
				tmpInput = append(tmpInput, vv)
			}
		}
	}
	return tmpInput
}

func GetOrgRoleNew(guid string) (result []*m.OptionModel, err error) {
	var serviceGroupRoleRows []*m.ServiceGroupRoleRelTable
	err = x.SQL("select * from service_group_role_rel where service_group=?", guid).Find(&serviceGroupRoleRows)
	if err != nil {
		err = fmt.Errorf("query service group role rel table fail,%s ", err.Error())
		return
	}
	var roleData []*m.RoleTable
	x.SQL("SELECT id,name,display_name FROM role").Find(&roleData)
	if len(roleData) == 0 {
		return result, nil
	}
	roleIdMap := make(map[string]int)
	for _, row := range roleData {
		roleIdMap[row.Name] = row.Id
	}
	for _, row := range serviceGroupRoleRows {
		result = append(result, &m.OptionModel{OptionText: row.Role, OptionValue: fmt.Sprintf("%d", roleIdMap[row.Role]), Id: roleIdMap[row.Role], OptionName: row.Role})
	}
	return
}

func GetOrgRole(guid string) (result []*m.OptionModel, err error) {
	var tableData []*m.PanelRecursiveTable
	err = x.SQL("SELECT role FROM panel_recursive WHERE guid=?", guid).Find(&tableData)
	if err != nil {
		return result, err
	}
	if len(tableData) == 0 {
		return result, fmt.Errorf("guid:%s can not find any record", guid)
	}
	if tableData[0].Role == "" {
		return result, nil
	}
	var roleData []*m.RoleTable
	x.SQL("SELECT id,name,display_name FROM role").Find(&roleData)
	if len(roleData) == 0 {
		return result, nil
	}
	for _, v := range strings.Split(tableData[0].Role, ",") {
		tmpId, _ := strconv.Atoi(v)
		if tmpId <= 0 {
			continue
		}
		for _, vv := range roleData {
			if vv.Id == tmpId {
				tmpName := vv.DisplayName
				if tmpName == "" {
					tmpName = vv.Name
				}
				result = append(result, &m.OptionModel{OptionText: tmpName, OptionValue: fmt.Sprintf("%d", vv.Id), Id: vv.Id, OptionName: vv.Name})
				break
			}
		}
	}
	return result, nil
}

func UpdateOrgRole(param m.UpdateOrgPanelRoleParam) (err error) {
	var actions []*Action
	var idString string
	var idStringList, roleStringList []string
	for _, v := range param.RoleId {
		idString += fmt.Sprintf("%d,", v)
		idStringList = append(idStringList, fmt.Sprintf("%d", v))
	}
	if idString != "" {
		idString = idString[:len(idString)-1]
	}
	var roleTable []*m.RoleTable
	x.SQL("select name from role where id in ('" + strings.Join(idStringList, "','") + "')").Find(&roleTable)
	for _, v := range roleTable {
		roleStringList = append(roleStringList, v.Name)
	}
	actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET role=? WHERE guid=?", Param: []interface{}{idString, param.Guid}})
	//var roleTable []*m.RoleTable
	actions = append(actions, getUpdateServiceGroupNotifyRoles(param.Guid, roleStringList)...)
	err = Transaction(actions)
	if err != nil {
		log.Logger.Error("Update organization role error", log.Error(err))
	}
	return err
}

func GetOrgEndpoint(guid string) (result []*m.OptionModel, err error) {
	var tableData []*m.PanelRecursiveTable
	err = x.SQL("SELECT endpoint FROM panel_recursive WHERE guid=?", guid).Find(&tableData)
	if err != nil {
		return result, err
	}
	if len(tableData) == 0 {
		return result, fmt.Errorf("guid:%s can not find any record", guid)
	}
	if tableData[0].Endpoint == "" {
		return result, nil
	}
	endpointString := strings.Replace(tableData[0].Endpoint, "^", "','", -1)
	var endpointData []*m.EndpointTable
	x.SQL(fmt.Sprintf("SELECT guid,name,ip,export_type FROM endpoint WHERE guid IN ('%s')", endpointString)).Find(&endpointData)
	if len(endpointData) == 0 {
		return result, nil
	}
	for _, v := range strings.Split(tableData[0].Endpoint, "^") {
		for _, vv := range endpointData {
			if vv.Guid == v {
				result = append(result, &m.OptionModel{OptionText: fmt.Sprintf("%s:%s", vv.Name, vv.Ip), OptionValue: vv.Guid, OptionType: vv.ExportType})
				break
			}
		}
	}
	return result, nil
}

func UpdateOrgEndpoint(param m.UpdateOrgPanelEndpointParam, operator string) error {
	var actions []*Action
	var endpointString string
	nowTime := time.Now().Format(m.DatetimeFormat)
	endpointString = strings.Join(param.Endpoint, "^")
	actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET endpoint=? WHERE guid=?", Param: []interface{}{endpointString, param.Guid}})
	actions = append(actions, getUpdateServiceEndpointAction(param.Guid, nowTime, operator, param.Endpoint)...)
	err := Transaction(actions)
	if err == nil {
		var endpointGroup []*m.EndpointGroupTable
		parentGuidList, _ := fetchGlobalServiceGroupParentGuidList(param.Guid)
		x.SQL("select guid from endpoint_group where service_group in ('" + strings.Join(parentGuidList, "','") + "')").Find(&endpointGroup)
		for _, v := range endpointGroup {
			err = SyncPrometheusRuleFile(v.Guid, false)
			if err != nil {
				break
			}
		}
		if err == nil {
			UpdateServiceConfigWithParent(param.Guid)
		}
	}
	return err
}

func GetOrgCallback(guid string) (result m.PanelRecursiveTable, err error) {
	var tableData []*m.PanelRecursiveTable
	err = x.SQL("SELECT firing_callback_name,firing_callback_key,recover_callback_name,recover_callback_key FROM panel_recursive WHERE guid=?", guid).Find(&tableData)
	if err != nil {
		return result, err
	}
	if len(tableData) == 0 {
		return result, fmt.Errorf("guid:%s can not find any record", guid)
	}
	return *tableData[0], nil
}

func UpdateOrgCallback(param m.UpdateOrgPanelEventParam) (err error) {
	var actions []*Action
	var roleList []string
	var serviceRoleTable []*m.ServiceGroupRoleRelTable
	x.SQL("select * from service_group_role_rel where service_group=?", param.Guid).Find(&serviceRoleTable)
	for _, v := range serviceRoleTable {
		roleList = append(roleList, v.Role)
	}
	actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET firing_callback_name=?,firing_callback_key=?,recover_callback_name=?,recover_callback_key=? WHERE guid=?", Param: []interface{}{param.FiringCallbackName, param.FiringCallbackKey, param.RecoverCallbackName, param.RecoverCallbackKey, param.Guid}})
	actions = append(actions, getUpdateServiceGroupNotifyActions(param.Guid, param.FiringCallbackKey, param.RecoverCallbackKey, roleList)...)
	err = Transaction(actions)
	if err != nil {
		log.Logger.Error("Update organization callback error", log.Error(err))
	}
	return err
}

func UpdateOrgConnect(param m.UpdateOrgConnectParam) error {
	_, err := x.Exec("UPDATE panel_recursive SET email=?,phone=? WHERE guid=?", strings.Join(param.Mail, ","), strings.Join(param.Phone, ","), param.Guid)
	if err != nil {
		log.Logger.Error("Update organization connection error", log.Error(err))
	}
	return err
}

func GetOrgConnect(guid string) (result m.UpdateOrgConnectParam, err error) {
	result.Mail = []string{}
	result.Phone = []string{}
	var tableData []*m.PanelRecursiveTable
	err = x.SQL("SELECT email,phone FROM panel_recursive WHERE guid=?", guid).Find(&tableData)
	if err != nil {
		return result, err
	}
	if len(tableData) == 0 {
		return result, fmt.Errorf("guid:%s can not find any record", guid)
	}
	result.Mail = strings.Split(tableData[0].Email, ",")
	result.Phone = strings.Split(tableData[0].Phone, ",")
	return result, nil
}

func SearchPanelByName(name, endpoint string) []m.OptionModel {
	name = "%" + name + "%"
	var result []m.OptionModel
	var panelRecursiveTables []*m.PanelRecursiveTable
	var err error
	if endpoint == "" {
		err = x.SQL("SELECT guid,display_name,obj_type FROM panel_recursive WHERE display_name LIKE ?", name).Find(&panelRecursiveTables)
	} else {
		endpoint = "%" + endpoint + "%"
		err = x.SQL("SELECT guid,display_name,obj_type FROM panel_recursive WHERE display_name LIKE ? AND endpoint LIKE ?", name, endpoint).Find(&panelRecursiveTables)
	}
	if err != nil {
		log.Logger.Error("Get panel_recursive table data fail", log.Error(err))
	}
	for _, v := range panelRecursiveTables {
		result = append(result, m.OptionModel{OptionText: fmt.Sprintf("%s(%s)", v.DisplayName, v.ObjType), OptionValue: v.Guid})
	}
	return result
}

func GetPanelRecursiveEndpoints(guid, endpointType string) (result []*m.EndpointTable, err error) {
	var panelRecursiveTable []*m.PanelRecursiveTable
	err = x.SQL("select endpoint from panel_recursive where guid=?", guid).Find(&panelRecursiveTable)
	if err != nil {
		return
	}
	if len(panelRecursiveTable) == 0 {
		err = fmt.Errorf("Can not find recursive object with guid:%s ", guid)
		return
	}
	result = []*m.EndpointTable{}
	err = x.SQL("select * from endpoint where guid in ('" + strings.ReplaceAll(panelRecursiveTable[0].Endpoint, "^", "','") + "')").Find(&result)
	return
}

func BatchGetServiceGroupNameByIds(ids []string) (list []string, err error) {
	err = x.SQL(fmt.Sprintf("select display_name from service_group where  guid in ('%s')", strings.Join(ids, "','"))).Find(&list)
	return
}
