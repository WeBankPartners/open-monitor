package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/funcs"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/api"
)

func main() {
	cfgFile := flag.String("c", "conf.json", "config file")
	flag.Parse()
	err := funcs.InitConfig(*cfgFile)
	if err != nil {
		log.Println("config file init fail, stop...")
		return
	}
	funcs.InitDeploy()
	go funcs.StartManager()
	api.InitHttpServer()
}
