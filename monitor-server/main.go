package main

import (
	"flag"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/api"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	ds "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/datasource"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/prom"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/middleware"
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
	ds.InitPrometheusDatasource()
	prom.InitPrometheusConfigFile()
	api.InitHttpServer(*exportAgent)
}
