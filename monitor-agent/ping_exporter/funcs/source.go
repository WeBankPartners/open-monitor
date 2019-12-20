package funcs

import (
	"net"
	"log"
)

func GetIntranetIp() []string {
	addrs, err := net.InterfaceAddrs()
	re := []string{}
	if err != nil {
		log.Println(err)
		return re
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				re = append(re, ipNet.IP.String())
			}
		}
	}
	return re
}
