package main

import (
	"flag"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/api"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/db"
	ds "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/datasource"
	"github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/services/prom"
)

// @title Monitor Server API
// @version 1.0
// @description 监控门户后台服务.
// @termsOfService http://swagger.io/terms/

// @host monitor.test.com
// @BasePath /v1
func main() {
	cfgFile := flag.String("c", "conf/default.json", "config file")
	flag.Parse()
	m.InitConfig(*cfgFile)
	db.InitDbConn()
	ds.InitPrometheusDatasource()
	prom.InitPrometheusConfigFile()
	api.InitHttpServer()
}
