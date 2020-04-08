package telnet

import (
	"net"
	"fmt"
	"log"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"time"
	"sync"
)

var (
	telnetTaskMap = make(map[string][]int)
	telnetResultList []*funcs.TelnetObj
)

func StartTelnetTask()  {
	interval := funcs.Config().Interval
	if interval < 30 {
		log.Println("telnet interval refresh to 30s")
		interval = 30
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		go telnetTask()
		<- t
	}
}

func telnetTask()  {
	telnetList := funcs.GetTelnetList()
	wg := sync.WaitGroup{}
	//var successCounter int
	for _,v := range telnetList {
		wg.Add(1)
		go func(ip string,port int) {
			defer wg.Done()
			d := doTelnet(ip, port)
			funcs.DebugLog("ping %s result %d ", ip, d)
		}(v.Ip,v.Port)
	}
}

func doTelnet(ip string,port int) bool {
	conn,err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return false
	}else{
		conn.Close()
		return true
	}
}
