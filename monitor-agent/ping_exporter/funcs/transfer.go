package funcs

import (
	"log"
	"sync"
	"time"
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/model"
)

var (
	TransferClientsLock *sync.RWMutex                   = new(sync.RWMutex)
	//TransferClients     map[string]*SingleConnRpcClient = map[string]*SingleConnRpcClient{}
	TClients  *SingleConnRpcClient
)

func SendMetrics(metrics []*model.MetricValue, resp *model.TransferResponse) {
	addr := Config().OpenFalcon.Transfer.Addrs[0]
	c := getTransferClient(addr)
	if c == nil {
		c = initTransferClient(addr)
	}
	updateMetrics(c, metrics, resp)
}

func initTransferClient(addr string) *SingleConnRpcClient {
	var c *SingleConnRpcClient = &SingleConnRpcClient{
		RpcServer: addr,
		Timeout:   time.Duration(Config().OpenFalcon.Transfer.Timeout) * time.Millisecond,
	}
	TransferClientsLock.Lock()
	defer TransferClientsLock.Unlock()
	TClients = c
	return c
}

func updateMetrics(c *SingleConnRpcClient, metrics []*model.MetricValue, resp *model.TransferResponse) bool {
	err := c.Call("Transfer.Update", metrics, resp)
	re := false
	if err != nil {
		log.Println("call Transfer.Update fail:", c, err)
		re = updateRetry(metrics, resp, 1)
	}else{
		re = true
	}
	return re
}

func updateRetry(metrics []*model.MetricValue, resp *model.TransferResponse, n int) bool {
	re := false
	if n<4 {
		log.Printf("call transfer retry time : %d \n", n)
		cc := initTransferClient(Config().OpenFalcon.Transfer.Addrs[0])
		err := cc.Call("Transfer.Update", metrics, resp)
		if err!=nil {
			log.Println("call Transfer.Update fail:", cc, err)
			j := n+1
			re = updateRetry(metrics, resp, j)
		}else {
			re = true
		}
	}
	return re
}

func getTransferClient(addr string) *SingleConnRpcClient {
	TransferClientsLock.RLock()
	defer TransferClientsLock.RUnlock()
	return TClients
}