package monitor

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/gin-gonic/gin"
	"strings"
)

// systemMonitorTypeMap 系统类型配置
var systemMonitorTypeList = []string{"host", "mysql", "redis", "java", "tomcat", "nginx", "ping", "pingext",
	"telnet", "telnetext", "http", "httpext", "windows", "snmp", "process", "pod"}

func QueryTypeConfigList(c *gin.Context) {
	var err error
	var list []*models.TypeConfig
	name := c.Query("name")
	if list, err = db.GetTypeConfigList(name); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, list)
}

func BatchGetTypeConfigList(c *gin.Context) {
	var param models.CommonNameParam
	var err error
	var list []*models.TypeConfig
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(param.Names) == 0 {
		middleware.ReturnSuccess(c)
		return
	}
	if list, err = db.GetTypeConfigListByNames(param.Names); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccessData(c, list)
}

func AddTypeConfig(c *gin.Context) {
	var param models.TypeConfig
	var typeConfigList []*models.TypeConfig
	var err error
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	param.Guid = param.DisplayName
	param.CreateUser = middleware.GetOperateUser(c)
	if typeConfigList, err = db.QueryTypeConfigByName(param.DisplayName); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(typeConfigList) > 0 {
		middleware.ReturnServerHandleError(c, models.GetMessageMap(c).TypeConfigNameRepeatError)
		return
	}
	if err = db.AddTypeConfig(param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func BatchAddTypeConfig(c *gin.Context) {
	var err error
	var param models.BatchAddTypeConfigParam
	var typeConfigList []*models.TypeConfig
	if err = c.ShouldBindJSON(&param); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	var newMonitorTypeList []string
	systemMonitorTypeMap := ConvertArr2Map(systemMonitorTypeList)
	for _, s := range param.DisplayNameList {
		// 过滤掉系统默认类型配置
		if systemMonitorTypeMap[s] {
			continue
		}
		newMonitorTypeList = append(newMonitorTypeList, s)
	}
	// 如果 role_new表还未初始化,需要先同步数据
	if !db.ExistRoles() {
		db.SyncCoreRole()
		db.SyncCoreRoleList()
	}
	for _, monitorType := range newMonitorTypeList {
		typeConfig := models.TypeConfig{Guid: monitorType, DisplayName: monitorType, CreateUser: middleware.GetOperateUser(c)}
		if typeConfigList, err = db.QueryTypeConfigByName(monitorType); err != nil {
			middleware.ReturnServerHandleError(c, err)
			return
		}
		if len(typeConfigList) > 0 {
			middleware.ReturnServerHandleError(c, models.GetMessageMap(c).TypeConfigNameRepeatError)
			return
		}
		if err = db.AddTypeConfig(typeConfig); err != nil {
			err = fmt.Errorf("monitorType:%s,fail:%v", monitorType, err.Error())
			middleware.ReturnServerHandleError(c, err)
			return
		}
	}
	middleware.ReturnSuccess(c)
}

func DeleteTypeConfig(c *gin.Context) {
	var typeConfig *models.TypeConfig
	var endpointList []*models.EndpointNewTable
	var endpointGroupList []*models.EndpointGroupTable
	var err error
	id := c.Query("id")
	if strings.TrimSpace(id) == "" {
		middleware.ReturnParamEmptyError(c, "id")
		return
	}
	if typeConfig, err = db.GetTypeConfig(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if typeConfig == nil {
		middleware.ReturnValidateError(c, "id invalid")
		return
	}
	if typeConfig.SystemType == 1 {
		middleware.ReturnValidateError(c, "system config not allow delete")
		return
	}
	if endpointList, err = db.GetEndpointByMonitorType(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(endpointList) > 0 {
		middleware.ReturnServerHandleError(c, models.GetMessageMap(c).TypeConfigNameAssociationObjectError)
		return
	}
	if endpointGroupList, err = db.GetEndpointGroupByMonitorType(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	if len(endpointGroupList) > 0 {
		middleware.ReturnServerHandleError(c, models.GetMessageMap(c).TypeConfigNameAssociationObjectGroupError)
		return
	}
	if err = db.DeleteTypeConfig(id); err != nil {
		middleware.ReturnServerHandleError(c, err)
		return
	}
	middleware.ReturnSuccess(c)
}

func ConvertArr2Map(list []string) map[string]bool {
	var hashMap = make(map[string]bool)
	for _, s := range list {
		hashMap[s] = true
	}
	return hashMap
}

func ConvertMap2Arr(hashMap map[string]bool) []string {
	var list []string
	for s, _ := range hashMap {
		list = append(list, s)
	}
	return list
}
