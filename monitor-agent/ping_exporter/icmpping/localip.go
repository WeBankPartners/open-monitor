package icmpping

func IsLocalIp(ip string, ips []string) bool {
	re := false
	for _,v := range ips {
		if ip == v {
			re = true
			break
		}
	}
	return re
}
