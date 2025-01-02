package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func AddUser(user m.UserTable) error {
	_, err := x.Exec("INSERT INTO user(name,passwd,display_name,email,phone,created) VALUE (?,?,?,?,?,NOW())", user.Name, user.Passwd, user.DisplayName, user.Email, user.Phone)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("Add user %s fail", user.Name), log.Error(err))
	}
	return err
}

func GetUser(username string) (err error, user m.UserQuery) {
	var users []*m.UserQuery
	err = x.SQL("SELECT * FROM user WHERE name=?", username).Find(&users)
	if len(users) == 0 {
		return err, m.UserQuery{}
	} else {
		users[0].CreatedString = users[0].Created.Format(m.DatetimeFormat)
	}
	return nil, *users[0]
}

func UpdateUser(user m.UserTable) error {
	param := make([]interface{}, 0)
	sql := "UPDATE user SET "
	if user.Passwd != "" {
		sql += "passwd=?,"
		param = append(param, user.Passwd)
	}
	if user.DisplayName != "" {
		sql += "display_name=?,"
		param = append(param, user.DisplayName)
	}
	if user.Email != "" {
		sql += "email=?,"
		param = append(param, user.Email)
	}
	if user.Phone != "" {
		sql += "phone=?,"
		param = append(param, user.Phone)
	}
	updateSql := sql[:len(sql)-1] + " WHERE name=?"
	param = append(param, user.Name)
	newParam := make([]interface{}, 0)
	newParam = append(newParam, updateSql)
	for _, v := range param {
		newParam = append(newParam, v)
	}
	_, err := x.Exec(newParam...)
	if err != nil {
		log.Logger.Error("Update user error", log.Error(err))
	}
	return err
}

func SearchUserRole(search string, searchType string) (err error, options []*m.OptionModel) {
	likeString := "%" + search + "%"
	var result []*m.RoleTable
	err = x.SQL(fmt.Sprintf("SELECT id,name,display_name FROM %s WHERE name LIKE '%s' OR display_name LIKE '%s' ORDER BY id LIMIT 15", searchType, likeString, likeString)).Find(&result)
	if err != nil {
		return err, options
	}
	tmpActive := false
	if searchType == "role" {
		tmpActive = true
	}
	for _, v := range result {
		tmpText := v.Name
		if v.DisplayName != "" {
			tmpText = tmpText + "(" + v.DisplayName + ")"
		}
		options = append(options, &m.OptionModel{Id: v.Id, OptionText: tmpText, OptionValue: fmt.Sprintf("%d", v.Id), Active: tmpActive, OptionType: fmt.Sprintf("%s_%d", searchType, v.Id)})
	}
	return nil, options
}

func GetMailByStrategy(strategyId int) []string {
	result := []string{}
	resultMap := make(map[string]int)
	var tpls []*m.TplTable
	x.SQL("SELECT DISTINCT t2.* FROM strategy t1 LEFT JOIN tpl t2 ON t1.tpl_id=t2.id WHERE t1.id=?", strategyId).Find(&tpls)
	if len(tpls) == 0 {
		log.Logger.Warn(fmt.Sprintf("can not find tpl with strategy %d", strategyId))
		return result
	}
	userIds := tpls[0].ActionUser
	if tpls[0].ActionRole != "" {
		var tmpRel []*m.RelRoleUserTable
		x.SQL(fmt.Sprintf("SELECT user_id FROM rel_role_user WHERE role_id IN (%s)", tpls[0].ActionRole)).Find(&tmpRel)
		for _, v := range tmpRel {
			userIds = userIds + fmt.Sprintf(",%d", v.UserId)
		}
		if strings.HasPrefix(userIds, ",") {
			userIds = userIds[1:]
		}
		var roleTable []*m.RoleTable
		x.SQL(fmt.Sprintf("SELECT * FROM role WHERE id IN (%s)", tpls[0].ActionRole)).Find(&roleTable)
		for _, v := range roleTable {
			if v.Email != "" {
				if _, b := resultMap[v.Email]; !b {
					result = append(result, v.Email)
					resultMap[v.Email] = 1
				}
			}
		}
	}
	if userIds != "" {
		var users []*m.UserTable
		x.SQL(fmt.Sprintf("SELECT DISTINCT email FROM user WHERE id IN (%s)", userIds)).Find(&users)
		for _, v := range users {
			if _, b := resultMap[v.Email]; !b {
				result = append(result, v.Email)
				resultMap[v.Email] = 1
			}
		}
	}
	if tpls[0].ExtraMail != "" {
		for _, v := range strings.Split(tpls[0].ExtraMail, ",") {
			if _, b := resultMap[v]; !b {
				result = append(result, v)
				resultMap[v] = 1
			}
		}
	}
	return result
}

func GetMailByEndpointGroup(guid string) []string {
	result := []string{}
	resultMap := make(map[string]int)
	var tpls []*m.TplTable
	x.SQL("SELECT t1.* FROM tpl t1 LEFT JOIN grp_endpoint t2 ON t1.grp_id=t2.grp_id LEFT JOIN endpoint t3 ON t2.endpoint_id=t3.id WHERE t3.guid=?", guid).Find(&tpls)
	if len(tpls) == 0 {
		log.Logger.Warn(fmt.Sprintf("can not find group with endpoint %s", guid))
		return result
	}
	for _, tpl := range tpls {
		userIds := tpl.ActionUser
		if tpl.ActionRole != "" {
			var tmpRel []*m.RelRoleUserTable
			x.SQL(fmt.Sprintf("SELECT user_id FROM rel_role_user WHERE role_id IN (%s)", tpl.ActionRole)).Find(&tmpRel)
			for _, v := range tmpRel {
				userIds = userIds + fmt.Sprintf(",%d", v.UserId)
			}
			if strings.HasPrefix(userIds, ",") {
				userIds = userIds[1:]
			}
			var roleTable []*m.RoleTable
			x.SQL(fmt.Sprintf("SELECT * FROM role WHERE id IN (%s)", tpl.ActionRole)).Find(&roleTable)
			for _, v := range roleTable {
				if v.Email != "" {
					if _, b := resultMap[v.Email]; !b {
						result = append(result, v.Email)
						resultMap[v.Email] = 1
					}
				}
			}
		}
		if userIds != "" {
			var users []*m.UserTable
			x.SQL(fmt.Sprintf("SELECT DISTINCT email FROM user WHERE id IN (%s)", userIds)).Find(&users)
			for _, v := range users {
				if _, b := resultMap[v.Email]; !b {
					result = append(result, v.Email)
					resultMap[v.Email] = 1
				}
			}
		}
		if tpl.ExtraMail != "" {
			for _, v := range strings.Split(tpl.ExtraMail, ",") {
				if _, b := resultMap[v]; !b {
					result = append(result, v)
					resultMap[v] = 1
				}
			}
		}
	}
	return result
}

func ListUser(search string, role, page, size int) (err error, data m.TableData) {
	var users []*m.UserQuery
	var count []int
	var whereSql string
	if role > 0 {
		whereSql = fmt.Sprintf(" AND t1.id IN (SELECT user_id FROM rel_role_user WHERE role_id=%d) ", role)
	}
	if search != "" {
		whereSql = " AND t1.name LIKE '%" + search + "%' OR display_name LIKE '%" + search + "%'"
	}
	sql := `SELECT t5.* FROM (
	SELECT t4.id,t4.name,t4.display_name,t4.email,t4.phone,t4.created,GROUP_CONCAT(role) role FROM (
	SELECT t1.id,t1.name,t1.display_name,t1.email,t1.phone,t1.created,CONCAT(t3.name,':',t3.display_name) role FROM user t1
	LEFT JOIN rel_role_user t2 ON t1.id=t2.user_id
	LEFT JOIN role t3 ON t2.role_id=t3.id
	WHERE 1=1 ` + whereSql + `
	) t4 GROUP BY t4.id
	) t5`
	err = x.SQL(sql + fmt.Sprintf(" ORDER BY t5.id LIMIT %d,%d", (page-1)*size, size)).Find(&users)
	x.SQL(sql).Find(&count)
	if len(users) > 0 {
		for _, v := range users {
			v.CreatedString = v.Created.Format(m.DatetimeFormat)
		}
		data.Data = users
	} else {
		data.Data = []*m.UserQuery{}
	}
	data.Size = size
	data.Page = page
	if len(count) > 0 {
		data.Num = count[0]
	} else {
		data.Num = len(users)
	}
	return err, data
}

func ListRole(search string, page, size int) (err error, data m.TableData) {
	var roles []*m.RoleQuery
	var count []int
	var whereSql string
	if search != "" {
		whereSql = "where (name LIKE '%" + search + "%' OR display_name LIKE '%" + search + "%') AND disable=0 "
	} else {
		whereSql = "where disable=0 "
	}
	err = x.SQL("SELECT id,name,display_name,email FROM role " + whereSql + fmt.Sprintf(" ORDER BY id LIMIT %d,%d", (page-1)*size, size)).Find(&roles)
	x.SQL("SELECT count(1) num FROM role " + whereSql).Find(&count)
	if len(roles) > 0 {
		for _, v := range roles {
			v.CreatedString = v.Created.Format(m.DatetimeFormat)
		}
		data.Data = roles
	} else {
		data.Data = []*m.RoleQuery{}
	}
	data.Size = size
	data.Page = page
	if len(count) > 0 {
		data.Num = count[0]
	} else {
		data.Num = len(roles)
	}
	return err, data
}

func ListManageRole(roles []string) (result []*m.RoleTable, err error) {
	var roleList []*m.RoleTable
	result = []*m.RoleTable{}
	if err = x.SQL("SELECT id,name,display_name FROM role where disable=0").Find(&roleList); err != nil {
		return
	}
	if len(roles) > 0 && len(roleList) > 0 {
		for _, roleTable := range roleList {
			for _, role := range roles {
				if roleTable.Name == role {
					result = append(result, roleTable)
				}
			}
		}
	}
	return
}

func StartCronJob() {
	go func() {
		SyncCoreRole()
		SyncCoreRoleList()
	}()
	//if !m.Config().CronJob.Enable {
	//	return
	//}
	intervalSec := 60
	if m.Config().CronJob.Interval > 30 {
		intervalSec = m.Config().CronJob.Interval
	}
	go StartSyncCoreJob(intervalSec)
	go prom.StartConsumeReloadConfig()
	go prom.StartCheckPrometheusJob(intervalSec)
	go prom.StartCheckProcessList(intervalSec)
	//go StartCronSyncKubernetesPod(intervalSec)
	go StartCleanAlarmTable()
}

func StartSyncCoreJob(interval int) {
	// Sync core role
	t := time.NewTicker(time.Minute * 1).C
	for {
		//go SyncCoreRole()
		go SyncCoreSystemVariable()
		<-t
	}
}

func SyncCoreRole() {
	if m.CoreUrl == "" {
		return
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/roles/retrieve", m.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Logger.Error("Get core role key new request fail", log.Error(err))
		return
	}
	request.Header.Set("Authorization", m.GetCoreToken())
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Logger.Error("Get core role key ctxhttp request fail", log.Error(err))
		return
	}
	b, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	var result m.CoreRoleDto
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Logger.Error("Get core role key json unmarshal result", log.Error(err))
		return
	}
	if len(result.Data) == 0 {
		log.Logger.Warn("Get core role key fail with no data")
		return
	}
	var tableData, insertData, updateData, deleteData []*m.RoleTable
	x.SQL("SELECT id,name,display_name FROM role").Find(&tableData)
	for _, v := range result.Data {
		var existFlag bool
		var updateName string
		for _, vv := range tableData {
			if vv.Name == v.Name {
				existFlag = true
				if vv.Disable == 1 {
					updateName = v.DisplayName
				}
				if vv.DisplayName != v.DisplayName {
					updateName = v.DisplayName
				}
				break
			}
		}
		if !existFlag {
			insertData = append(insertData, &m.RoleTable{Name: v.Name, DisplayName: v.DisplayName, Email: v.Email})
		}
		if updateName != "" {
			updateData = append(updateData, &m.RoleTable{Name: v.Name, DisplayName: v.DisplayName, Email: v.Email})
		}
	}
	for _, v := range tableData {
		var existFlag bool
		for _, vv := range result.Data {
			if vv.Name == v.Name {
				existFlag = true
				break
			}
		}
		if !existFlag {
			deleteData = append(deleteData, &m.RoleTable{Name: v.Name})
		}
	}
	var actions []*Action
	for _, v := range insertData {
		actions = append(actions, &Action{Sql: fmt.Sprintf("INSERT INTO role(name,display_name) VALUE ('%s','%s')", v.Name, v.DisplayName)})
	}
	for _, v := range updateData {
		actions = append(actions, &Action{Sql: fmt.Sprintf("UPDATE role SET display_name='%s',disable=0 WHERE name='%s'", v.DisplayName, v.Name)})
	}
	for _, v := range deleteData {
		actions = append(actions, &Action{Sql: fmt.Sprintf("UPDATE role SET disable=1 WHERE name='%s'", v.Name)})
	}
	if len(actions) > 0 {
		err = Transaction(actions)
		if err != nil {
			log.Logger.Error("Sync core role fail", log.Error(err))
		}
	}
}

func SyncCoreSystemVariable() {
	if m.CoreUrl == "" {
		return
	}
	var param m.RequestCoreVariableDto
	var filters []*m.CoreVariableFilter
	param.Paging = true
	param.Pageable = m.CoreVariablePage{PageSize: 10, StartIndex: 0}
	filters = append(filters, &m.CoreVariableFilter{Name: "name", Operator: "contains", Value: "MONITOR_MAIL_DEFAULT_RECEIVER"})
	filters = append(filters, &m.CoreVariableFilter{Name: "status", Operator: "eq", Value: "active"})
	param.Filters = filters
	postBytes, _ := json.Marshal(param)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/platform/v1/system-variables/retrieve", m.CoreUrl), bytes.NewReader(postBytes))
	if err != nil {
		log.Logger.Error("Get core system variable new request fail", log.Error(err))
		return
	}
	request.Header.Set("Authorization", m.GetCoreToken())
	request.Header.Set("Content-Type", "application/json")
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Logger.Error("Get core system variable ctxhttp request fail", log.Error(err))
		return
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	var result m.RequestCoreVariableResult
	log.Logger.Info("Get core system variable response", log.String("body", string(b)))
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Logger.Error("Get core system variable json unmarshal result", log.Error(err))
		return
	}
	if result.Status != "OK" {
		log.Logger.Error("Get core system variable fail", log.JsonObj("response", result))
	} else {
		if len(result.Data.Contents) > 0 {
			tmpValue := result.Data.Contents[0].Value
			if tmpValue == "" {
				tmpValue = result.Data.Contents[0].DefaultValue
			}
			if strings.Join(m.DefaultMailReceiver, ",") != tmpValue {
				log.Logger.Info("Get core system variable success", log.String("name", "MONITOR_MAIL_DEFAULT_RECEIVER"), log.String("value", tmpValue))
			}
			m.DefaultMailReceiver = []string{}
			for _, v := range strings.Split(tmpValue, ",") {
				if v != "" {
					m.DefaultMailReceiver = append(m.DefaultMailReceiver, v)
				}
			}
		}
	}
}

func UpdateRoleUser(param m.UpdateRoleUserDto) error {
	var roleUserTable []*m.RelRoleUserTable
	err := x.SQL("SELECT user_id FROM rel_role_user WHERE role_id=?", param.RoleId).Find(&roleUserTable)
	if err != nil {
		return err
	}
	isSame := true
	if len(roleUserTable) != len(param.UserId) {
		isSame = false
	} else {
		for _, v := range roleUserTable {
			tmp := false
			for _, vv := range param.UserId {
				if v.UserId == vv {
					tmp = true
					break
				}
			}
			if !tmp {
				isSame = false
				break
			}
		}
	}
	if isSame {
		return nil
	}
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM rel_role_user WHERE role_id=?", Param: []interface{}{param.RoleId}})
	for _, v := range param.UserId {
		actions = append(actions, &Action{Sql: "INSERT INTO rel_role_user(role_id,user_id) VALUE (?,?)", Param: []interface{}{param.RoleId, v}})
	}
	err = Transaction(actions)
	return err
}

func UpdateGrpRole(param m.RoleGrpDto) error {
	var roleGrpTable []*m.RelRoleGrpTable
	err := x.SQL("SELECT role_id FROM rel_role_grp WHERE grp_id=?", param.GrpIdInt).Find(&roleGrpTable)
	if err != nil {
		return err
	}
	isSame := true
	if len(roleGrpTable) != len(param.RoleId) {
		isSame = false
	} else {
		for _, v := range roleGrpTable {
			tmp := false
			for _, vv := range param.RoleId {
				if v.RoleId == vv {
					tmp = true
					break
				}
			}
			if !tmp {
				isSame = false
				break
			}
		}
	}
	if isSame {
		return nil
	}
	var actions []*Action
	actions = append(actions, &Action{Sql: "DELETE FROM rel_role_grp WHERE grp_id=?", Param: []interface{}{param.GrpIdInt}})
	for _, v := range param.RoleId {
		actions = append(actions, &Action{Sql: "INSERT INTO rel_role_grp(role_id,grp_id) VALUE (?,?)", Param: []interface{}{v, param.GrpIdInt}})
	}
	err = Transaction(actions)
	return err
}

func UpdateRoleNew(param m.UpdateRoleDto) error {
	var rowData []*m.RoleTable
	var err error
	var actions []*Action
	if param.Operation == "add" {
		if param.Name == "" {
			return fmt.Errorf("Role name can not emtpy")
		}
		x.SQL("select * from role where name=?", param.Name).Find(&rowData)
		if len(rowData) > 0 {
			actions = append(actions, &Action{Sql: "update role set disable=0 where name=?", Param: []interface{}{param.Name}})
		} else {
			actions = append(actions, &Action{Sql: "insert into role(name,display_name,email,creator,created) value (?,?,?,?,?)", Param: []interface{}{param.Name, param.DisplayName, param.Email, param.Operator, time.Now()}})
			actions = append(actions, &Action{Sql: "insert into role_new(guid,display_name,email) value (?,?,?)", Param: []interface{}{param.Name, param.DisplayName, param.Email}})
		}
		err = Transaction(actions)
	} else if param.Operation == "update" {
		if param.RoleId <= 0 {
			return fmt.Errorf("Role id can not emtpy ")
		}
		if param.Name == "" {
			return fmt.Errorf("Role name can not empty ")
		}
		_, err = x.Exec("update role set name=?,display_name=?,email=? where id=?", param.Name, param.DisplayName, param.Email, param.RoleId)
	} else {
		if param.RoleId <= 0 {
			return fmt.Errorf("Role id can not emtpy ")
		}
		x.SQL("select * from role where name=?", param.Name).Find(&rowData)
		if len(rowData) > 0 {
			_, err = x.Exec("update role set disable=1 where id=?", param.RoleId)
		}
	}
	return err
}

func UpdateRole(param m.UpdateRoleDto) error {
	var role m.RoleTable
	force := false
	if param.Operation == "add" {
		if param.Name == "" {
			return fmt.Errorf("role name is null")
		}
		role.Name = param.Name
		role.DisplayName = param.DisplayName
		role.Email = param.Email
		role.Creator = param.Operator
		role.Created = time.Now()
		param.Operation = "insert"
	}
	if param.Operation == "update" {
		if param.RoleId <= 0 {
			return fmt.Errorf("role id is null")
		}
		if param.Name == "" {
			return fmt.Errorf("role name is null")
		}
		role.Id = param.RoleId
		role.Name = param.Name
		role.DisplayName = param.DisplayName
		role.Email = param.Email
		force = true
	}
	if param.Operation == "delete" {
		if param.RoleId <= 0 {
			return fmt.Errorf("role id is null")
		}
		role.Id = param.RoleId
	}
	action := Classify(role, param.Operation, "role", force)
	return Transaction([]*Action{&action})
}

func GetGrpRole(grpId int) (err error, result []*m.OptionModel) {
	var queryData []*m.RoleTable
	err = x.SQL("SELECT t2.id,t2.name,t2.display_name FROM rel_role_grp t1 LEFT JOIN role t2 ON t1.role_id=t2.id WHERE t1.grp_id=?", grpId).Find(&queryData)
	if err != nil {
		log.Logger.Error("Get grp role fail", log.Error(err))
		return err, result
	}
	for _, v := range queryData {
		tmpName := v.DisplayName
		if tmpName == "" {
			tmpName = v.Name
		}
		result = append(result, &m.OptionModel{Id: v.Id, OptionValue: fmt.Sprintf("%d", v.Id), OptionText: tmpName})
	}
	return nil, result
}

func GetRoleMap() (roleMap map[string]string) {
	SyncCoreRole()
	SyncCoreRoleList()
	roleMap = make(map[string]string)
	var roleTable []*m.RoleNewTable
	x.SQL("select * from role_new").Find(&roleTable)
	for _, v := range roleTable {
		roleMap[v.Guid] = v.Email
	}
	return roleMap
}

func CheckRoleIllegal(roleList []string, roleMap map[string]string) (err error) {
	for _, v := range roleList {
		if _, b := roleMap[v]; !b {
			err = fmt.Errorf("role:%s illegal")
			break
		}
	}
	return err
}

func CheckRoleList(param string) string {
	if param == "" {
		return ""
	}
	tmpMap := make(map[string]int)
	for _, v := range strings.Split(param, ",") {
		tmpMap[v] = 0
	}
	var roleTable []*m.RoleNewTable
	x.SQL("select guid from role_new").Find(&roleTable)
	for k, _ := range tmpMap {
		var tableData []*m.RoleTable
		x.SQL("SELECT id FROM role WHERE name=?", k).Find(&tableData)
		if len(tableData) > 0 {
			tmpMap[k] = tableData[0].Id
		} else {
			_, err := x.Exec("INSERT INTO role(name,display_name) VALUE (?,?)", k, k)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("check role list,insert table with name:%s error", k), log.Error(err))
			} else {
				x.SQL("SELECT id FROM role WHERE name=?", k).Find(&tableData)
				if len(tableData) > 0 {
					tmpMap[k] = tableData[0].Id
				}
			}
		}
	}
	var result string
	for _, v := range tmpMap {
		result += fmt.Sprintf("%d,", v)
	}
	if result != "" {
		result = result[:len(result)-1]
	}
	return result
}

func GetUserRole(user string) (err error, result []*m.RoleTable) {
	err = x.SQL("SELECT DISTINCT t3.* FROM user t1 LEFT JOIN rel_role_user t2 ON t1.id=t2.user_id LEFT JOIN role t3 ON t2.role_id=t3.id WHERE t1.name=?", user).Find(&result)
	return err, result
}
