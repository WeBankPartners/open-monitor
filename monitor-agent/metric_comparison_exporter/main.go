package main

import (
	"flag"
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-agent/metric_comparison/funcs"
	"net/http"
)

func main() {
	port := flag.Int("p", 8181, "http listen port")
	flag.Parse()
	go StartHttpServer(*port)
	go funcs.StartCalcMetricComparisonCron()
}

func StartHttpServer(port int) {
	http.Handle("/metrics", http.HandlerFunc(funcs.HandlePrometheus))
	http.Handle("/receive", http.HandlerFunc(funcs.ReceiveMetricComparisonData))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
