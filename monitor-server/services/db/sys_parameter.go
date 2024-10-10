package db

import (
	"encoding/json"
	"fmt"
	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"os"
)

func InitSysParameter() {
	if models.PluginRunningMode {
		alertMailObj := models.SysAlertMailParameter{SenderMail: os.Getenv("MONITOR_MAIL_SENDER_USER"), AuthServer: os.Getenv("MONITOR_MAIL_SENDER_SERVER"), AuthPassword: os.Getenv("MONITOR_MAIL_SENDER_PASSWORD"), SSL: os.Getenv("MONITOR_MAIL_SENDER_SSL"), AuthUser: os.Getenv("MONITOR_MAIL_AUTH_USER")}
		UpdateSysAlertMailConfig(&alertMailObj)
	}
}

func GetSysAlertMailConfig() (result models.SysAlertMailParameter, err error) {
	queryRow, tmpErr := getSysParameterTableData(models.SPAlertMailKey)
	if tmpErr != nil {
		return result, tmpErr
	}
	if len(queryRow) == 0 {
		return result, fmt.Errorf("Can not find sys parameter with key:%s ", models.SPAlertMailKey)
	}
	if queryRow[0].ParamValue == "" {
		return result, fmt.Errorf("Sys parameter %s value is empty ", models.SPAlertMailKey)
	}
	err = json.Unmarshal([]byte(queryRow[0].ParamValue), &result)
	if err != nil {
		err = fmt.Errorf("Sys parameter %s value illegal ", models.SPAlertMailKey)
	}
	return
}

func UpdateSysAlertMailConfig(param *models.SysAlertMailParameter) {
	if param.AuthServer == "" || param.SenderMail == "" {
		return
	}
	param.SenderName = param.SenderMail
	b, _ := json.Marshal(param)
	existRow, err := getSysParameterTableData(models.SPAlertMailKey)
	if err != nil {
		return
	}
	if len(existRow) == 0 {
		_, err = x.Exec("insert into sys_parameter(guid,param_key,param_value) value (?,?,?)", guid.CreateGuid(), models.SPAlertMailKey, string(b))
	} else {
		_, err = x.Exec("update sys_parameter set param_value=? where param_key=?", string(b), models.SPAlertMailKey)
	}
	if err != nil {
		log.Logger.Error("Update sys alert mail fail", log.Error(err))
	}
}

func GetSysMetricTemplateConfig(workspace string) (result []*models.SysMetricTemplateParameter, err error) {
	result = []*models.SysMetricTemplateParameter{}
	paramKey := models.SPMetricTemplate
	if workspace == models.MetricWorkspaceService {
		paramKey = models.SPServiceMetricTemplate
	}
	queryRow, tmpErr := getSysParameterTableData(paramKey)
	if tmpErr != nil {
		return result, tmpErr
	}
	if len(queryRow) == 0 {
		return result, fmt.Errorf("Can not find sys parameter with key:%s ", models.SPMetricTemplate)
	}
	for _, v := range queryRow {
		tmpObj := models.SysMetricTemplateParameter{}
		tmpErr := json.Unmarshal([]byte(v.ParamValue), &tmpObj)
		if tmpErr != nil {
			err = tmpErr
			log.Logger.Error("Json unmarshal value fail", log.String("key", paramKey), log.String("value", v.ParamValue), log.Error(tmpErr))
			break
		} else {
			result = append(result, &tmpObj)
		}
	}
	return
}

func getSysParameterTableData(key string) (result []*models.SysParameterTable, err error) {
	err = x.SQL("select * from sys_parameter where param_key=?", key).Find(&result)
	if err != nil {
		log.Logger.Error("Query sys parameter key fail", log.String("key", key), log.Error(err))
	}
	return
}
