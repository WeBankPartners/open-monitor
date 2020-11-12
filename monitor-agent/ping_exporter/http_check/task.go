package http_check

import (
	"github.com/WeBankPartners/open-monitor/monitor-agent/ping_exporter/funcs"
	"sync"
	"log"
	"time"
	"net/http"
	"strings"
	"net/url"
)

var (
	httpCheckResultList  []*funcs.HttpCheckObj
	resultLock = new(sync.RWMutex)
)

func StartHttpCheckTask()  {
	interval := funcs.Config().Interval
	if interval < 30 {
		log.Println("http_check interval refresh to 30s")
		interval = 30
	}
	t := time.NewTicker(time.Second*time.Duration(interval)).C
	for {
		go httpCheckTask()
		<- t
	}
}

func buildHtppClient() *http.Client {
	var proxy func(*http.Request) (*url.URL, error) = nil
	if funcs.Config().HttpProxyEnable {
		proxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse(funcs.Config().HttpProxyAddress)
		}
	}
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport}
	return client
}

func httpCheckTask()  {
	clearHttpCheckResult()
	startTime := time.Now()
	httpCheckList := funcs.GetHttpCheckList()

	httpClient := buildHtppClient()
	httpClient.CloseIdleConnections()
	wg := sync.WaitGroup{}
	//var successCounter int
	for _,v := range httpCheckList {
		wg.Add(1)
		go func(method string,url string) {
			//b := doHttpCheck(method, url)
			b := doHttpCheckNew(method, url, httpClient)
			writeHttpCheckResult(method, url, b)
			funcs.DebugLog("http check %s:%s result %d ", method, url, b)
			wg.Done()
		}(v.Method,v.Url)
	}
	wg.Wait()
	endTime := time.Now()
	useTime := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6
	resultList,successCount := getHttpCheckResult()
	log.Printf("end http check, success num %d, fail num %d, use time %.3f ms \n", successCount, len(resultList)-successCount, useTime)
	funcs.UpdateHttpCheckExportMetric(resultList, successCount)
}

func doHttpCheck(method,url string) int  {
	reqHttpMethod := http.MethodGet
	if method == "post" {
		reqHttpMethod = http.MethodPost
	}
	req,err := http.NewRequest(reqHttpMethod, url, strings.NewReader(""))
	if err != nil {
		log.Printf("do http check -> method:%s url:%s new request error: %v \n", method, url, err)
		return 1
	}
	resp,err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("do http check -> method:%s url:%s response error: %v \n", method, url, err)
		return 2
	}
	return resp.StatusCode
}

func doHttpCheckNew(method,url string,httpClient *http.Client) int {
	var resp *http.Response
	var err error
	body:=strings.NewReader("")
	var httpMethods map[string]string = map[string]string{"POST": "POST","GET":"GET","OPTIONS":"OPTIONS","HEAD":"HEAD","PUT":"PUT","DELETE":"DELETE","TRACE":"TRACE","CONNECT":"CONNECT"}
	if _, ok := httpMethods[method]; !ok {
		log.Printf("do http check -> Not support method:%s \n", method)
		return 2
	}
	
	req, err := http.NewRequest(strings.ToUpper(method), url, body)
	if err != nil {
		log.Printf("do http check -> method:%s url:%s response error: %v \n", method, url, err)
		return 2
	}
	req.Header.Set("Content-Type", "application/json")
	resp , err = httpClient.Do(req)

	if err != nil {
		log.Printf("do http check -> method:%s url:%s response error: %v \n", method, url, err)
		return 2
	}
	if resp.Body != nil {
		resp.Body.Close()
	}
	return resp.StatusCode
}

func writeHttpCheckResult(method,url string,statusCode int)  {
	resultLock.Lock()
	httpCheckResultList = append(httpCheckResultList, &funcs.HttpCheckObj{Method:method, Url:url, StatusCode:statusCode})
	resultLock.Unlock()
}

func clearHttpCheckResult()  {
	resultLock.Lock()
	httpCheckResultList = []*funcs.HttpCheckObj{}
	resultLock.Unlock()
}

func getHttpCheckResult() (result []*funcs.HttpCheckObj,successCount int) {
	resultLock.RLock()
	for _,v := range httpCheckResultList {
		if v.StatusCode >= 200 && v.StatusCode < 300 {
			successCount += 1
		}
		result = append(result, &funcs.HttpCheckObj{Method:v.Method, Url:v.Url, StatusCode:v.StatusCode})
	}
	resultLock.RUnlock()
	return result,successCount
}
