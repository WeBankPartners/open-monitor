package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/icmpping"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/telnet"
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
	if !funcs.Config().PingEnable && !funcs.Config().TelnetEnable {
		return
	}
	icmpping.TestModel = *isTest
	funcs.InitSourceList()
	go icmpping.StartHttpServer()
	if !funcs.Config().PingEnable {
		telnet.StartTelnetTask()
		return
	}
	if !funcs.Config().TelnetEnable {
		icmpping.StartTask()
		return
	}
	if funcs.Config().PingEnable && funcs.Config().TelnetEnable {
		go icmpping.StartTask()
		telnet.StartTelnetTask()
	}
}
