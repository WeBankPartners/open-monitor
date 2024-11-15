package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"time"
)

func GetTypeConfigList(name string) (list []*models.TypeConfig, err error) {
	var endpointList []*models.EndpointNewTable
	if name == "" {
		err = x.SQL("select * from monitor_type order by create_time desc").Find(&list)
	} else {
		err = x.SQL("select * from monitor_type where display_name like ? order by create_time desc", fmt.Sprintf("%%%s%%", name)).Find(&list)
	}
	if err != nil {
		return
	}
	if len(list) > 0 {
		if err = x.SQL("select * from endpoint_new ").Find(&endpointList); err != nil {
			return
		}
		if len(endpointList) > 0 {
			for _, typeConf := range list {
				for _, endpoint := range endpointList {
					if endpoint.MonitorType == typeConf.DisplayName && endpoint.MonitorType != "" {
						typeConf.ObjectCount++
					}
				}
			}
		}
	}
	return
}

func GetTypeConfigListByNames(names []string) (list []*models.TypeConfig, err error) {
	var endpointList []*models.EndpointNewTable
	err = x.SQL(fmt.Sprintf("select * from monitor_type where display_name in ('%s') order by create_time desc", strings.Join(names, "','"))).Find(&list)
	if err != nil {
		return
	}
	if len(list) > 0 {
		if err = x.SQL("select * from endpoint_new ").Find(&endpointList); err != nil {
			return
		}
		if len(endpointList) > 0 {
			for _, typeConf := range list {
				for _, endpoint := range endpointList {
					if endpoint.MonitorType == typeConf.DisplayName && endpoint.MonitorType != "" {
						typeConf.ObjectCount++
					}
				}
			}
		}
	}
	return
}

func AddTypeConfig(param models.TypeConfig) (err error) {
	_, err = x.Exec("insert into monitor_type(guid,display_name,system_type,create_user,create_time) values(?,?,?,?,?)",
		param.Guid, param.DisplayName, param.SystemType, param.CreateUser, time.Now().Format(models.DatetimeFormat))
	return
}

func DeleteTypeConfig(id string) (err error) {
	_, err = x.Exec("delete from monitor_type where guid=?", id)
	return
}

func GetEndpointByMonitorType(id string) (list []*models.EndpointNewTable, err error) {
	err = x.SQL("select * from endpoint_new where monitor_type=?", id).Find(&list)
	return
}
func GetEndpointGroupByMonitorType(id string) (list []*models.EndpointGroupTable, err error) {
	err = x.SQL("select * from endpoint_group where monitor_type=?", id).Find(&list)
	return
}

func QueryTypeConfigByName(name string) (typeConfigList []*models.TypeConfig, err error) {
	err = x.SQL("select * from monitor_type where display_name=?", name).Find(&typeConfigList)
	return
}

func GetTypeConfig(id string) (typeConfig *models.TypeConfig, err error) {
	typeConfig = &models.TypeConfig{}
	_, err = x.SQL("select * from monitor_type where guid=?", id).Get(typeConfig)
	return
}
