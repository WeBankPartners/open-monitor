package db

import "github.com/WeBankPartners/open-monitor/monitor-server/models"

func GetLogKeywordByServiceGroup(serviceGroupGuid string) (result []*models.LogKeywordServiceGroupObj,err error) {
	return
}

func GetLogKeywordByEndpoint(endpointGuid string) (result []*models.LogKeywordServiceGroupObj,err error) {
	return
}

func AddLogKeywordMonitor(param *models.LogKeywordMonitorObj) (err error) {
	return
}

func UpdateLogKeywordMonitor(param *models.LogKeywordMonitorObj) (err error) {
	return
}

func DeleteLogKeywordMonitor(logKeywordMonitorGuid string) (err error) {
	return
}

func AddLogKeyword(param *models.LogKeywordConfigTable) (err error) {
	return
}

func UpdateLogKeyword(param *models.LogKeywordConfigTable) (err error) {
	return
}

func DeleteLogKeyword(logKeywordConfigGuid string) (err error) {
	return
}
