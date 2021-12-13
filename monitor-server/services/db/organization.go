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
			if strings.Contains(tmpMap[v], nameText) {
				tmpHeaderObj.FetchSearch = true
				tmpHeaderObj.FetchOriginFlag = true
			}
		}
		if endpointText != "" {
			if strings.Contains(endpointMap[v], endpointText) {
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
				if strings.Contains(v.Endpoint, endpointText) {
					tmpOrganizationObj.FetchSearch = true
					tmpOrganizationObj.FetchOriginFlag = true
				}
			}
			if nameText != "" {
				if strings.Contains(v.DisplayName, nameText) {
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
	var actions []*Action
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
		actions = append(actions, getCreateServiceGroupAction(&m.ServiceGroupTable{Guid: param.Guid, DisplayName: param.DisplayName, Description: "", Parent: param.Parent, ServiceType: param.Type, UpdateTime: nowTime})...)
		err = Transaction(actions)
		if err == nil {
			addGlobalServiceGroupNode(m.ServiceGroupTable{Guid: param.Guid, Parent: param.Parent})
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
		actions = append(actions, &Action{Sql: "update service_group set display_name=?,service_type=? where guid=?", Param: []interface{}{param.DisplayName, param.Type, param.Guid}})
		err = Transaction(actions)
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
		//TODO delete fore dep
		actions = append(actions, getDeleteServiceGroupAction(param.Guid)...)
		actions = append(actions, &Action{Sql: fmt.Sprintf("DELETE FROM service_group WHERE guid in ('%s')", strings.Join(guidList, "','"))})
		err = Transaction(actions)
		if err == nil {
			DeleteServiceWithChildConfig(param.Guid)
			deleteGlobalServiceGroupNode(param.Guid)
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
				result = append(result, &m.OptionModel{OptionText: tmpName, OptionValue: fmt.Sprintf("%d", vv.Id), Id: vv.Id})
				break
			}
		}
	}
	return result, nil
}

func UpdateOrgRole(param m.UpdateOrgPanelRoleParam) error {
	var idString string
	for _, v := range param.RoleId {
		idString += fmt.Sprintf("%d,", v)
	}
	if idString != "" {
		idString = idString[:len(idString)-1]
	}
	_, err := x.Exec("UPDATE panel_recursive SET role=? WHERE guid=?", idString, param.Guid)
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

func UpdateOrgEndpoint(param m.UpdateOrgPanelEndpointParam) error {
	var actions []*Action
	var endpointString string
	nowTime := time.Now().Format(m.DatetimeFormat)
	endpointString = strings.Join(param.Endpoint, "^")
	actions = append(actions, &Action{Sql: "UPDATE panel_recursive SET endpoint=? WHERE guid=?", Param: []interface{}{endpointString, param.Guid}})
	actions = append(actions, getUpdateServiceEndpointAction(param.Guid, nowTime, param.Endpoint)...)
	err := Transaction(actions)
	if err == nil {
		var endpointGroup []*m.EndpointGroupTable
		x.SQL("select guid from endpoint_group where service_group=?", param.Guid).Find(&endpointGroup)
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

func UpdateOrgCallback(param m.UpdateOrgPanelEventParam) error {
	_, err := x.Exec("UPDATE panel_recursive SET firing_callback_name=?,firing_callback_key=?,recover_callback_name=?,recover_callback_key=? WHERE guid=?", param.FiringCallbackName, param.FiringCallbackKey, param.RecoverCallbackName, param.RecoverCallbackKey, param.Guid)
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
