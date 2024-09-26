package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"golang.org/x/net/context/ctxhttp"
	"io"
	"net/http"
	"strings"
	"time"
)

func SyncCoreRoleList() {
	if models.CoreUrl == "" {
		return
	}
	log.Logger.Debug("Start sync role list")
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/roles/retrieve", models.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Logger.Error("Get core role key new request fail", log.Error(err))
		return
	}
	request.Header.Set("Authorization", models.GetCoreToken())
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Logger.Error("Get core role key ctxhttp request fail", log.Error(err))
		return
	}
	b, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	var result models.CoreRoleDto
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Logger.Error("Get core role key json unmarshal result", log.Error(err))
		return
	}
	if len(result.Data) == 0 {
		log.Logger.Warn("Get core role key fail with no data")
		return
	}
	var tableData, insertData, updateData []*models.RoleNewTable
	x.SQL("SELECT * FROM role_new").Find(&tableData)
	for _, v := range result.Data {
		var existFlag, updateFlag bool
		for _, vv := range tableData {
			if vv.Guid == v.Name {
				existFlag = true
				if vv.DisplayName != v.DisplayName || vv.Email != v.Email {
					updateFlag = true
				}
				break
			}
		}
		if !existFlag {
			insertData = append(insertData, &models.RoleNewTable{Guid: v.Name, DisplayName: v.DisplayName, Email: v.Email})
			continue
		}
		if updateFlag {
			updateData = append(updateData, &models.RoleNewTable{Guid: v.Name, DisplayName: v.DisplayName, Email: v.Email})
		}
	}
	nowTime := time.Now().Format(models.DatetimeFormat)
	var actions []*Action
	for _, v := range insertData {
		actions = append(actions, &Action{Sql: "insert into role_new(guid,display_name,email,update_time) value (?,?,?,?)", Param: []interface{}{v.Guid, v.DisplayName, v.Email, nowTime}})
	}
	for _, v := range updateData {
		actions = append(actions, &Action{Sql: "update role_new set display_name=?,email=?,update_time=? where guid=?", Param: []interface{}{v.DisplayName, v.Email, nowTime, v.Guid}})
	}
	if len(actions) > 0 {
		err = Transaction(actions)
		if err != nil {
			log.Logger.Error("Sync core role fail", log.Error(err))
		}
	}
}

func ExistRoles() bool {
	var tableData []*models.RoleNewTable
	x.SQL("SELECT * FROM role_new").Find(&tableData)
	if len(tableData) > 0 {
		return true
	}
	return false
}
