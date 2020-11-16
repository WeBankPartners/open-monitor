package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/icmpping"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/telnet"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/http_check"
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
	if !funcs.Config().PingEnable && !funcs.Config().TelnetEnable && !funcs.Config().HttpCheckEnable {
		return
	}
	icmpping.TestModel = *isTest
	funcs.InitSourceList()
	go icmpping.StartHttpServer()
	if funcs.Config().PingEnable {
		go icmpping.StartTask()
	}
	if funcs.Config().TelnetEnable {
		go telnet.StartTelnetTask()
	}
	if funcs.Config().HttpCheckEnable {
		go http_check.StartHttpCheckTask()
	}
	select {}
}
