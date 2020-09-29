package main

import (
	"github.com/WeBankPartners/open-monitor/monitor-agent/db_data_exporter/funcs"
	"flag"
)

func main() {
	port := flag.Int("p", 9192, "http listen port")
	flag.Parse()
	go funcs.StartHttpServer(*port)
	funcs.StartCronTask()
}
