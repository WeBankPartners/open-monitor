package icmpping

func IcmpType(rt uint8, rc uint8) string {
	var message string
	if rt == 3 {
		switch rc {
		case 0: message = `Network Unreachable——网络不可达`
		case 1: message = `Host Unreachable——主机不可达`
		case 2: message = `Protocol Unreachable——协议不可达`
		case 3: message = `Port Unreachable——端口不可达`
		case 4: message = `Fragmentation needed but no frag. bit set——需要进行分片但设置不分片比特`
		case 5: message = `Source routing failed——源站选路失败`
		case 6: message = `Destination network unknown——目的网络未知`
		case 7: message = `Destination host unknown——目的主机未知`
		case 8: message = `Source host isolated (obsolete)——源主机被隔离`
		case 9: message = `Destination network administratively prohibited——目的网络被强制禁止`
		case 10: message = `Destination host administratively prohibited——目的主机被强制禁止`
		case 11: message = `Network unreachable for TOS——由于服务类型TOS，网络不可达`
		case 12: message = `Host unreachable for TOS——由于服务类型TOS，主机不可达`
		case 13: message = `Communication administratively prohibited by filtering——由于过滤，通信被强制禁止`
		case 14: message = `Host precedence violation——主机越权`
		case 15: message = `Precedence cutoff in effect——优先中止生效`
		}
	}else if rt == 4 {
		message = `Source quench——源端被关闭（基本流控制）`
	}else if rt == 5 {
		switch rc {
		case 0: message = `Redirect for network——对网络重定向`
		case 1: message = `Redirect for host——对主机重定向`
		case 2: message = `Redirect for TOS and network——对服务类型和网络重定向`
		case 3: message = `Redirect for TOS and host——对服务类型和主机重定向`
		}
	}else if rt == 11 {
		switch rc {
		case 0: message = `TTL equals 0 during transit——传输期间生存时间为0`
		case 1: message = `TTL equals 0 during reassembly——在数据报组装期间生存时间为0`
		}
	}else if rt == 12 {
		switch rc {
		case 0: message = `IP header bad (catchall error)——坏的IP首部（包括各种差错）`
		case 1: message = `Required options missing——缺少必需的选项`
		}
	}
	return message
}
