package main

import (
	"github.com/WeBankPartners/open-monitor/monitor-agent/opentsdb_adapter/funcs"
	"flag"
)

func main()  {
	openTsdbUrl := flag.String("u", "http://127.0.0.1:4242", "openTsdb url")
	flag.Parse()
	funcs.OpenTSDBUrl = *openTsdbUrl
	funcs.InitHttpServer()
}