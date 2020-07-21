package main

import (
	"flag"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/archive_mysql_tool/funcs"
)

func main() {
	cfgFile := flag.String("c", "default.json", "config file")
	flag.Parse()
	err := funcs.InitConfig(*cfgFile)
	if err != nil {
		log.Printf("init config fail : %v \n", err)
		return
	}
	err = funcs.InitDbEngine()
	if err != nil {
		log.Printf("init mysql connect fail : %v \n", err)
		return
	}
	err = funcs.InitMonitorDbEngine()
	if err != nil {
		log.Printf("init monitor mysql connect fail : %v \n", err)
		return
	}
	err = funcs.InitMonitorMetricMap()
	if err != nil {
		log.Printf("init monitor metric data fail : %v \n", err)
		return
	}
	err = funcs.InitHttpTransport()
	if err != nil {
		log.Printf("init prometheus http transport pool fail : %v \n", err)
		return
	}
	go funcs.InitHttpHandles()
	funcs.StartCronJob()
}
