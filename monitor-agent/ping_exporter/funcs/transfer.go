package funcs

import (
	"log"
	"sync"
	"time"
	"fmt"
)

var (
	TransferClientsLock = new(sync.RWMutex)
	TClients  *SingleConnRpcClient
	localUuid  string
	Hosts = make(map[string]string)   // 存储主机IP和UUID对应关系
)

func InitTransfer()  {
	localUuid = Uuid()
}

func HandleTransferResult(result map[string]PingResultObj,successCount int)  {
	sendData := []*MetricValue{}
	metric := Config().Metrics.Ping
	interval := Config().Interval
	now := time.Now().Unix()
	for k,v := range result{
		endpoint := Hosts[k]
		if endpoint!="" {
			metricData := MetricValue{Endpoint: endpoint, Metric: metric, Value: v.UpDown, Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
			sendData = append(sendData, &metricData)
		}else{
			metricData := MetricValue{Endpoint: localUuid, Metric: metric, Value: v.UpDown, Step: int64(interval), Type: "GAUGE", Tags: fmt.Sprintf("ip=%s",k), Timestamp: now}
			sendData = append(sendData, &metricData)
		}
	}
	if localUuid!=""{
		metricOk := MetricValue{Endpoint: localUuid, Metric: "ip_ping_ok_num", Value: successCount, Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
		metricFa := MetricValue{Endpoint: localUuid, Metric: "ip_ping_fa_num", Value: len(result)-successCount, Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
		metricAll := MetricValue{Endpoint: localUuid, Metric: "ip_ping_all_num", Value: len(result), Step: int64(interval), Type: "GAUGE", Tags: "", Timestamp: now}
		sendData = append(sendData, &metricOk)
		sendData = append(sendData, &metricFa)
		sendData = append(sendData, &metricAll)
	}
	length := len(sendData)
	sn := Config().OpenFalcon.Transfer.Sn
	if sn<=0{
		sn = 500
	}
	if length > 0 {
		log.Printf("=> <Total=%d> %v\n", length, sendData[0])
	}
	if length>sn{
		var resps TransferResponse
		cut := length/sn
		if cut*sn<length{
			cut = cut+1
		}
		for i:=0;i<cut;i++{
			s := i*sn
			e := (i+1)*sn
			if i==cut-1{
				e = length
			}
			SendMetrics(sendData[s:e], &resps)
			log.Printf("s: %d , e : %d <= %v \n", s, e, &resps)
		}
	}else {
		var resp TransferResponse
		SendMetrics(sendData, &resp)
		log.Println("<=", &resp)
	}
}

func SendMetrics(metrics []*MetricValue, resp *TransferResponse) {
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

func updateMetrics(c *SingleConnRpcClient, metrics []*MetricValue, resp *TransferResponse) bool {
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

func updateRetry(metrics []*MetricValue, resp *TransferResponse, n int) bool {
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

type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

func (this *MetricValue) String() string {
	return fmt.Sprintf(
		"<Endpoint:%s, Metric:%s, Type:%s, Tags:%s, Step:%d, Time:%d, Value:%v>",
		this.Endpoint,
		this.Metric,
		this.Type,
		this.Tags,
		this.Step,
		this.Timestamp,
		this.Value,
	)
}

type TransferResponse struct {
	Message string
	Total   int
	Invalid int
	Latency int64
}

func (this *TransferResponse) String() string {
	return fmt.Sprintf(
		"<Total=%v, Invalid:%v, Latency=%vms, Message:%s>",
		this.Total,
		this.Invalid,
		this.Latency,
		this.Message,
	)
}

type Response struct {
	Code    int  	`json:"code"`
	Data    []string  	`json:"data"`
	Msg     string  	`json:"msg"`
}
