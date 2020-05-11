package prom

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
)

func GetEndpointData(ip,port string,prefix,keyword []string) (error, []string) {
	var strList []string
	resp,err := http.Get(fmt.Sprintf("http://%s:%s/metrics", ip, port))
	if err != nil {
		fmt.Printf("http get error %v \n", err)
		return err,strList
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("read body error %v \n", err)
		return err,strList
	}
	if resp.StatusCode/100 != 2 {
		fmt.Printf("response http code %v \n", resp.StatusCode)
		return err,strList
	}
	for _,v := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if strings.Contains(v, ` `) {
			v = v[:strings.LastIndex(v, ` `)]
		}
		tmpStr := strings.ToLower(v)
		if len(prefix) == 0 && len(keyword) == 0 {
			strList = append(strList, tmpStr)
			continue
		}
		tmpFlag := false
		for _,vv := range prefix {
			if strings.HasPrefix(tmpStr, vv+"_") {
				strList = append(strList, tmpStr)
				tmpFlag = true
				break
			}
		}
		if tmpFlag {
			continue
		}
		for _,vv := range keyword {
			if strings.Contains(tmpStr, vv) {
				strList = append(strList, tmpStr)
				break
			}
		}
	}
	fmt.Printf("metric num : %d \n", len(strList))
	return nil,strList
}