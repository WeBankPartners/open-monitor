package main

import (
	"flag"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/api"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
)

// @title Monitor Server API
// @version 1.0
// @description 监控门户后台服务.
// @termsOfService http://swagger.io/terms/

// @host monitor.test.com
// @BasePath /v1
func main() {
	cfgFile := flag.String("c", "conf/default.json", "config file")
	//exportAgent := flag.Bool("export_agent", false, "true or false to choice export agent")
	flag.Parse()
	m.InitConfig(*cfgFile)
	if err := log.InitLogger(); err != nil {
		fmt.Printf("Server  init loggers failed, err: %v\n", err)
		return
	}
	defer log.SyncLoggers()
	db.InitDatabase()
	if m.Config().Http.Session.Enable == "true" {
		middleware.InitSession()
	}
	ds.InitPrometheusDatasource()
	if m.Config().Alert.Enable {
		other.InitSmtpMail()
	}
	go api.InitClusterApi()
	go db.InitPrometheusConfig()
	go db.InitSysParameter()
	go db.InitServiceGroup()
	go db.StartCronJob()
	go db.StartCheckCron()
	go db.StartLogKeywordMonitorCronJob()
	go db.StartDbKeywordMonitorCronJob()
	go alarm.StartAlarmEngineCron()
	go db.SyncDbMetric(true)
	go db.StartCallCronJob()
	go db.StartNotifyPingExport()
	go api.InitDependenceParam()
	go db.StartInitAlarmUniqueTags()
	go db.SyncMetricComparison()
	api.InitHttpServer()
}
