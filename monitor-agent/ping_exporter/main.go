package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/icmpping"
	"log"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	isTest := flag.Bool("test", false, "run once and ignore send/export metric")
	flag.Parse()
	err := funcs.ParseConfig(*cfg)
	if err != nil {
		log.Println("parse config fail,stop now...")
		return
	}
	icmpping.TestModel = *isTest
	funcs.InitIpList()
	go icmpping.StartHttpServer()
	icmpping.StartTask()
}
