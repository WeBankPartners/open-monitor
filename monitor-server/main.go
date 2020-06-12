package main

import (
	"flag"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/api"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/db"
	ds "github.com/WeBankPartners/open-monitor/monitor-server/services/datasource"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/other"
	"github.com/WeBankPartners/open-monitor/monitor-server/api/v1/alarm"
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
	middleware.InitMonitorLog()
	db.InitDbConn()
	if m.Config().Http.Session.Enable {
		middleware.InitSession()
	}
	ds.InitPrometheusDatasource()
	prom.InitPrometheusConfigFile()
	if m.Config().Alert.Enable {
		other.InitSmtpMail()
	}
	go api.InitClusterApi()
	go db.StartCronJob()
	alarm.SyncInitSdFile()
	alarm.SyncInitConfigFile()
	api.InitDependenceParam()
	api.InitHttpServer(*exportAgent)
}
