package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-agent/archive_mysql_tool/funcs"
	"log"
	"strings"
)

func main() {
	// Initialize daily rotating logger first
	funcs.InitLogger()

	cfgFile := flag.String("c", "default.json", "config file")
	flag.Parse()
	err := funcs.InitConfig(*cfgFile)
	if err != nil {
		log.Printf("init config fail : %v \n", err)
		return
	}
	enable := strings.ToLower(funcs.Config().Enable)
	if enable == "y" || enable == "yes" || enable == "true" {
		log.Println("enable flag true,start... ")
	} else {
		log.Println("enable flag false,stop... ")
		return
	}
	err = funcs.InitDbEngine("")
	if err != nil {
		log.Printf("init mysql connect fail : %v \n", err)
		return
	}
	err = funcs.ChangeDatabase("")
	if err != nil {
		log.Printf("change mysql database connect fail : %v \n", err)
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
