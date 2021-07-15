package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-server/api"
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
	exportAgent := flag.Bool("export_agent", false, "true or false to choice export agent")
	flag.Parse()
	m.InitConfig(*cfgFile)
	//middleware.InitMonitorLog()
	log.InitLogger()
	db.InitDbConn()
	if m.Config().Http.Session.Enable {
		middleware.InitSession()
	}
	ds.InitPrometheusDatasource()
	//prom.InitPrometheusRuleFile()
	if m.Config().Alert.Enable {
		other.InitSmtpMail()
	}
	go api.InitClusterApi()
	go db.InitPrometheusConfig()
	//go db.InitPrometheusConfigFile()
	go db.StartCronJob()
	go db.StartCheckCron()
	go db.StartCheckLogKeyword()
	go db.SendConfigToDbManager()
	go db.StartNotifyPingExport()
	//alarm.SyncInitSdFile()
	//alarm.SyncInitConfigFile()
	api.InitDependenceParam()
	middleware.InitErrorMessageList()
	api.InitHttpServer(*exportAgent)
}
