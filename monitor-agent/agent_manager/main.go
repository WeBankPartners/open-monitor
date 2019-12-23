package main

import (
	"flag"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/funcs"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/agent_manager/api"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfgFile := flag.String("c", "conf.json", "config file")
	flag.Parse()
	err := funcs.InitConfig(*cfgFile)
	if err != nil {
		log.Println("config file init fail, stop...")
		return
	}
	if !funcs.InitLocalIp() {
		log.Println("init local ip fail, stop...")
		return
	}
	funcs.InitDeploy()
	funcs.LoadDeployProcess()
	go funcs.StartManager()
	go api.InitHttpServer()
	startSignal(os.Getpid())
}

func startSignal(pid int) {
	sigs := make(chan os.Signal, 1)
	log.Println(pid, "register signal notify")
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		s := <-sigs
		log.Println("receive signal ", s)
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("shutdown , start save file")
			// do something
			funcs.SaveDeployProcess()
			funcs.StopDeployProcess()
			log.Println("shutdown , done")
			log.Println(pid, "exit")
			os.Exit(0)
		}
	}
}