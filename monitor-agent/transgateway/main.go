package main

import (
	"flag"
	"os"
	"syscall"
	"os/signal"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/api"
	"github.com/WeBankPartners/open-monitor/monitor-agent/transgateway/models"
)

func main() {
	port := flag.String("p", "19091", "listen port")
	timeout := flag.Int64("t", 120, "data timeout")
	dataDir := flag.String("d", "", "data save path")
	monitorUrl := flag.String("m", "", "monitor endpoint register url")
	flag.Parse()
	models.InitMonitorUrl(*monitorUrl, *port)
	models.LoadCacheData(*dataDir)
	go models.CleanTimeoutData(*timeout)
	go api.InitHttpServer(*port)
	startSignal(os.Getpid())
	select{}
}

func startSignal(pid int) {
	sigs := make(chan os.Signal, 1)
	log.Println(pid, "register signal notify")
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		s := <-sigs
		log.Println("recv", s)
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("shutdown , start save file")
			models.SaveCacheData()
			log.Println("shutdown , done")
			log.Println(pid, "exit")
			os.Exit(0)
		}
	}
}