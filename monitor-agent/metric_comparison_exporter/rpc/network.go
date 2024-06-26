package rpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// HttpGet  Get请求
func HttpGet(url string) (byteArr []byte, err error) {
	req, newReqErr := http.NewRequest(http.MethodGet, url, strings.NewReader(""))
	if newReqErr != nil {
		err = fmt.Errorf("try to new http request fail,%s ", newReqErr.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = fmt.Errorf("try to do http request fail,%s ", respErr.Error())
		return
	}
	byteArr, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}

// HttpPost Post请求
func HttpPost(url string, postBytes []byte) (byteArr []byte, err error) {
	req, reqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(postBytes))
	if reqErr != nil {
		err = fmt.Errorf("new http reqeust fail,%s ", reqErr.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = fmt.Errorf("do http reqeust fail,%s ", respErr.Error())
		return
	}
	byteArr, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}
