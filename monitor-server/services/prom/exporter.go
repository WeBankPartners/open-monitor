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
	tmpMap := make(map[string]int)
	for _,v := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(v, "#") {
			continue
		}
		if strings.Contains(v, ` `) {
			v = v[:strings.LastIndex(v, ` `)]
		}
		for _,vv := range prefix {
			if strings.HasPrefix(v, vv+"_") {
				tmpStr := v[strings.Index(v, vv+"_"):]
				tmpMap[tmpStr] = 1
			}
		}
		for _,vv := range keyword {
			if strings.Contains(v, vv) {
				tmpStr := v
				tmpMap[tmpStr] = 1
			}
		}
	}
	for k,_ := range tmpMap {
		strList = append(strList, k)
	}
	fmt.Printf("metric num : %d \n", len(tmpMap))
	return nil,strList
}