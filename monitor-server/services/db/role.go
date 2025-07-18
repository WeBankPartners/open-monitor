package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
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
	log.Debug(nil, log.LOGGER_APP, "Start sync role list")
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/platform/v1/roles/retrieve", models.CoreUrl), strings.NewReader(""))
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core role key new request fail", zap.Error(err))
		return
	}
	request.Header.Set("Authorization", models.GetCoreToken())
	res, err := ctxhttp.Do(context.Background(), http.DefaultClient, request)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core role key ctxhttp request fail", zap.Error(err))
		return
	}
	b, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	var result models.CoreRoleDto
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get core role key json unmarshal result", zap.Error(err))
		return
	}
	if len(result.Data) == 0 {
		log.Warn(nil, log.LOGGER_APP, "Get core role key fail with no data")
		return
	}
	var tableData []*models.RoleNewTable
	err = x.SQL("SELECT guid,display_name,email,disable FROM role_new").Find(&tableData)
	if err != nil {
		log.Error(nil, log.LOGGER_APP, "Get role new table data fail", zap.Error(err))
		return
	}
	// 1. 先禁用本地所有不在平台返回列表中的角色
	platformGuids := make(map[string]struct{})
	for _, v := range result.Data {
		platformGuids[v.Name] = struct{}{}
	}
	for _, local := range tableData {
		if _, ok := platformGuids[local.Guid]; !ok && local.Disable == 0 {
			_, err = x.SQL("UPDATE role_new SET disable=1,update_time=? WHERE guid=?", time.Now().Format("2006-01-02 15:04:05"), local.Guid)
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "Disable role new fail", zap.Error(err), zap.String("guid", local.Guid))
			}
		}
	}
	// 2. 插入/更新平台返回的角色（disable=0）
	for _, v := range result.Data {
		var existFlag bool
		for _, local := range tableData {
			if local.Guid == v.Name {
				existFlag = true
				if local.DisplayName != v.DisplayName || local.Email != v.Email || local.Disable == 1 {
					_, err = x.SQL("UPDATE role_new SET display_name=?,email=?,disable=0,update_time=? WHERE guid=?", v.DisplayName, v.Email, time.Now().Format("2006-01-02 15:04:05"), v.Name)
					if err != nil {
						log.Error(nil, log.LOGGER_APP, "Update role new fail", zap.Error(err), zap.String("guid", v.Name))
					}
				}
				break
			}
		}
		if !existFlag {
			_, err = x.SQL("INSERT INTO role_new (guid,display_name,email,disable,update_time) VALUES (?,?,?,?,?)", v.Name, v.DisplayName, v.Email, 0, time.Now().Format("2006-01-02 15:04:05"))
			if err != nil {
				log.Error(nil, log.LOGGER_APP, "Insert role new fail", zap.Error(err), zap.String("guid", v.Name))
			}
		}
	}
	log.Debug(nil, log.LOGGER_APP, "Sync role list end")
}

func ExistRoles() bool {
	var tableData []*models.RoleNewTable
	x.SQL("SELECT * FROM role_new").Find(&tableData)
	if len(tableData) > 0 {
		return true
	}
	return false
}
